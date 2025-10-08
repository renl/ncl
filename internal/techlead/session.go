package techlead

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
)

var (
	// ErrAborted is returned when the user aborts the session intentionally (e.g. via :quit).
	ErrAborted = errors.New("session aborted by user")

	errEmptyResponse = errors.New("language model returned an empty response")
)

const defaultSystemPrompt = "" +
	"You are an experienced software engineering tech lead collaborating with another tech lead to " +
	"create a crisp, actionable planning document.\n" +
	"Always produce well-structured Markdown with clear section headings, tables when appropriate, " +
	"and concise bullet points.\n" +
	"Focus on clarity, decision-making, risks, trade-offs, and stakeholder alignment."

// SessionConfig defines runtime configuration for the gendoc session.
type SessionConfig struct {
	// DefaultOutputPath optionally stores the path that should receive the final document when the
	// session exits. If empty, no automatic write happens.
	DefaultOutputPath string
	// SystemPrompt allows overriding the default system prompt used for the model.
	SystemPrompt string
	// CallOptions are forwarded to every LLM invocation (e.g. temperature, top_p).
	CallOptions []llms.CallOption
}

// Session orchestrates an interactive conversation with an LLM to co-author a tech lead document.
type Session struct {
	llm     llms.Model
	reader  *bufio.Reader
	out     io.Writer
	errOut  io.Writer
	history []llms.MessageContent
	config  SessionConfig

	lastDoc    string
	savedPaths map[string]struct{}
}

// NewSession constructs a session that reads from in, writes to out/errOut, and talks to llm.
func NewSession(model llms.Model, in io.Reader, out io.Writer, errOut io.Writer, cfg SessionConfig) *Session {
	systemPrompt := strings.TrimSpace(cfg.SystemPrompt)
	if systemPrompt == "" {
		systemPrompt = defaultSystemPrompt
	}

	s := &Session{
		llm:        model,
		reader:     bufio.NewReader(in),
		out:        out,
		errOut:     errOut,
		config:     cfg,
		history:    []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt)},
		savedPaths: make(map[string]struct{}),
	}

	return s
}

// Run executes the interactive workflow. It blocks until the user exits or an error occurs.
func (s *Session) Run(ctx context.Context) error {
	s.printIntro()

	brief, err := s.collectBrief(ctx)
	if err != nil {
		if errors.Is(err, ErrAborted) {
			fmt.Fprintln(s.out, "Session cancelled before drafting any content.")
			return nil
		}
		return err
	}

	if err := s.generateInitialDraft(ctx, brief); err != nil {
		return err
	}

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		fmt.Fprint(s.out, "\nRefine (:help for commands)> ")
		line, err := s.readLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return s.finish()
			}
			return err
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, ":") {
			exit, cmdErr := s.handleCommand(ctx, trimmed)
			if cmdErr != nil {
				fmt.Fprintf(s.errOut, "command error: %v\n", cmdErr)
				continue
			}
			if exit {
				return s.finish()
			}
			continue
		}

		if err := s.generateRevision(ctx, trimmed); err != nil {
			fmt.Fprintf(s.errOut, "revision failed: %v\n", err)
		}
	}
}

func (s *Session) printIntro() {
	fmt.Fprintf(s.out, "Tech Lead Document Co-authoring Session\n")
	fmt.Fprintln(s.out, strings.Repeat("=", 42))
	fmt.Fprintln(s.out, "We'll gather a quick project brief and draft a document together.")
	fmt.Fprintln(s.out, "You can type :quit at any time to exit.")
	fmt.Fprintln(s.out, "")
}

type briefField struct {
	id       string
	prompt   string
	hint     string
	optional bool
}

type sessionBrief struct {
	Project      string
	Context      string
	Goals        string
	Architecture string
	Risks        string
	Timeline     string
	Stakeholders string
}

func (s sessionBrief) toPrompt() string {
	var b strings.Builder
	b.WriteString("Using the following project context, create a comprehensive tech lead document in Markdown.\n")
	b.WriteString("Ensure the output contains these sections: Summary, Goals & Non-Goals, Architecture Overview, Implementation Plan, Risks & Mitigations, Stakeholders, Timeline & Milestones, Open Questions, and Next Steps.\n")
	b.WriteString("Respond with the complete document only.\n\n")

	if s.Project != "" {
		b.WriteString("Project Name: ")
		b.WriteString(s.Project)
		b.WriteString("\n")
	}
	if s.Context != "" {
		b.WriteString("Context: ")
		b.WriteString(s.Context)
		b.WriteString("\n")
	}
	if s.Goals != "" {
		b.WriteString("Goals: ")
		b.WriteString(s.Goals)
		b.WriteString("\n")
	}
	if s.Architecture != "" {
		b.WriteString("Architecture Direction: ")
		b.WriteString(s.Architecture)
		b.WriteString("\n")
	}
	if s.Risks != "" {
		b.WriteString("Risks & Constraints: ")
		b.WriteString(s.Risks)
		b.WriteString("\n")
	}
	if s.Timeline != "" {
		b.WriteString("Timeline: ")
		b.WriteString(s.Timeline)
		b.WriteString("\n")
	}
	if s.Stakeholders != "" {
		b.WriteString("Stakeholders: ")
		b.WriteString(s.Stakeholders)
		b.WriteString("\n")
	}

	return b.String()
}

var introFields = []briefField{
	{id: "project", prompt: "Project or initiative name", hint: "e.g. Phoenix Payments Revamp", optional: false},
	{id: "context", prompt: "Context / business problem", hint: "Why this work matters", optional: false},
	{id: "goals", prompt: "Goals and non-goals", hint: "Key outcomes and explicit exclusions", optional: false},
	{id: "architecture", prompt: "Architecture direction", hint: "Key components, patterns, integrations", optional: true},
	{id: "risks", prompt: "Risks and constraints", hint: "Technical, org, or delivery risks", optional: true},
	{id: "timeline", prompt: "Timeline / milestones", hint: "Critical dates or target launch", optional: true},
	{id: "stakeholders", prompt: "Stakeholders / partners", hint: "Teams, execs, external groups", optional: true},
}

func (s *Session) collectBrief(ctx context.Context) (sessionBrief, error) {
	var brief sessionBrief

	for _, field := range introFields {
		if err := ctx.Err(); err != nil {
			return sessionBrief{}, err
		}

		label := field.prompt
		if field.hint != "" {
			label = fmt.Sprintf("%s (%s)", field.prompt, field.hint)
		}
		fmt.Fprintf(s.out, "%s: ", label)

		answer, err := s.readLine()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return sessionBrief{}, ErrAborted
			}
			return sessionBrief{}, err
		}
		trimmed := strings.TrimSpace(answer)
		if strings.EqualFold(trimmed, ":quit") {
			return sessionBrief{}, ErrAborted
		}
		if trimmed == "" && !field.optional {
			fmt.Fprintln(s.out, "  (Please provide a response; empty values are allowed for optional prompts.)")
			// repeat same field
			for {
				fmt.Fprintf(s.out, "%s: ", label)
				answer, err = s.readLine()
				if err != nil {
					if errors.Is(err, io.EOF) {
						return sessionBrief{}, ErrAborted
					}
					return sessionBrief{}, err
				}
				trimmed = strings.TrimSpace(answer)
				if strings.EqualFold(trimmed, ":quit") {
					return sessionBrief{}, ErrAborted
				}
				if trimmed != "" {
					break
				}
				fmt.Fprintln(s.out, "  (This field is required; describe as best as you can.)")
			}
		}

		switch field.id {
		case "project":
			brief.Project = trimmed
		case "context":
			brief.Context = trimmed
		case "goals":
			brief.Goals = trimmed
		case "architecture":
			brief.Architecture = trimmed
		case "risks":
			brief.Risks = trimmed
		case "timeline":
			brief.Timeline = trimmed
		case "stakeholders":
			brief.Stakeholders = trimmed
		}
	}

	return brief, nil
}

func (s *Session) generateInitialDraft(ctx context.Context, brief sessionBrief) error {
	prompt := brief.toPrompt()
	s.history = append(s.history, llms.TextParts(llms.ChatMessageTypeHuman, prompt))

	content, err := s.invokeModel(ctx)
	if err != nil {
		return fmt.Errorf("initial draft: %w", err)
	}

	s.lastDoc = content
	s.history = append(s.history, llms.TextParts(llms.ChatMessageTypeAI, content))

	fmt.Fprintln(s.out)
	fmt.Fprintln(s.out, "--- Draft v1 ---")
	fmt.Fprintln(s.out, content)

	return nil
}

func (s *Session) generateRevision(ctx context.Context, instruction string) error {
	if s.lastDoc == "" {
		return errors.New("no draft available yet")
	}

	s.history = append(s.history, llms.TextParts(llms.ChatMessageTypeHuman, instruction))

	content, err := s.invokeModel(ctx)
	if err != nil {
		// rollback history entry to keep conversation clean
		s.history = s.history[:len(s.history)-1]
		return err
	}

	s.lastDoc = content
	s.history = append(s.history, llms.TextParts(llms.ChatMessageTypeAI, content))

	fmt.Fprintln(s.out)
	fmt.Fprintln(s.out, "--- Updated Draft ---")
	fmt.Fprintln(s.out, content)

	return nil
}

func (s *Session) invokeModel(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	resp, err := s.llm.GenerateContent(ctx, s.history, s.config.CallOptions...)
	if err != nil {
		return "", err
	}
	if resp == nil || len(resp.Choices) == 0 {
		return "", errEmptyResponse
	}

	content := strings.TrimSpace(resp.Choices[0].Content)
	if content == "" {
		return "", errEmptyResponse
	}

	return content, nil
}

func (s *Session) handleCommand(ctx context.Context, input string) (bool, error) {
	lower := strings.ToLower(strings.TrimSpace(input))
	switch {
	case lower == ":quit" || lower == ":q" || lower == ":exit":
		return true, nil
	case lower == ":help":
		s.printHelp()
		return false, nil
	case lower == ":show":
		if s.lastDoc == "" {
			fmt.Fprintln(s.out, "No draft yet. Complete the initial briefing first.")
			return false, nil
		}
		fmt.Fprintln(s.out, "\n--- Current Draft ---")
		fmt.Fprintln(s.out, s.lastDoc)
		return false, nil
	case strings.HasPrefix(lower, ":save"):
		path := strings.TrimSpace(input[len(":save"):])
		if path == "" {
			path = s.config.DefaultOutputPath
		}
		if path == "" {
			return false, errors.New("provide a path, e.g. :save techlead.md or pass --out")
		}
		if err := s.saveDocument(path, true); err != nil {
			return false, err
		}
		return false, nil
	default:
		return false, fmt.Errorf("unknown command %q", input)
	}
}

func (s *Session) finish() error {
	if s.config.DefaultOutputPath != "" && s.lastDoc != "" {
		if err := s.saveDocument(s.config.DefaultOutputPath, false); err != nil {
			return err
		}
		fmt.Fprintf(s.out, "\nFinal document saved to %s\n", s.config.DefaultOutputPath)
	} else {
		fmt.Fprintln(s.out, "\nSession ended.")
	}
	return nil
}

func (s *Session) saveDocument(path string, announce bool) error {
	if s.lastDoc == "" {
		return errors.New("there is no document to save yet")
	}

	cleaned := strings.TrimSpace(path)
	if cleaned == "" {
		return errors.New("invalid path")
	}

	abs, err := filepath.Abs(cleaned)
	if err != nil {
		return err
	}

	dir := filepath.Dir(abs)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data := []byte(s.lastDoc)
	if err := os.WriteFile(abs, data, 0o644); err != nil {
		return err
	}

	s.savedPaths[abs] = struct{}{}
	if announce {
		fmt.Fprintf(s.out, "Saved draft to %s (%s)\n", abs, time.Now().Format(time.RFC3339))
	}
	return nil
}

func (s *Session) printHelp() {
	fmt.Fprintln(s.out, "\nCommands:")
	fmt.Fprintln(s.out, "  :help           Show this help message")
	fmt.Fprintln(s.out, "  :show           Print the latest draft")
	fmt.Fprintln(s.out, "  :save [path]    Save the current draft to [path] or default --out path")
	fmt.Fprintln(s.out, "  :quit           Exit the session (auto-saves if --out provided)")
}

func (s *Session) readLine() (string, error) {
	line, err := s.reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) && len(line) > 0 {
			// Allow EOF after partial line.
			return line, nil
		}
		return line, err
	}
	return line, nil
}

// LatestDocument returns the most recent document snapshot.
func (s *Session) LatestDocument() string {
	return s.lastDoc
}
