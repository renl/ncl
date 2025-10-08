package techlead

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

type fakeModel struct {
	responses []string
	callErr   error
	calls     int
}

func (f *fakeModel) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	if f.callErr != nil {
		return nil, f.callErr
	}
	if f.calls >= len(f.responses) {
		return nil, errors.New("no more responses")
	}
	resp := f.responses[f.calls]
	f.calls++
	return &llms.ContentResponse{Choices: []*llms.ContentChoice{{Content: resp}}}, nil
}

func (f *fakeModel) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return "", errors.New("not implemented")
}

func TestGenerateInitialDraft(t *testing.T) {
	model := &fakeModel{responses: []string{"# Draft\ncontent"}}
	session := NewSession(model, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, SessionConfig{})

	brief := sessionBrief{Project: "Project X", Context: "Improve reliability", Goals: "99.99% uptime"}
	if err := session.generateInitialDraft(context.Background(), brief); err != nil {
		t.Fatalf("generateInitialDraft: %v", err)
	}

	if got, want := session.lastDoc, "# Draft\ncontent"; got != want {
		t.Fatalf("unexpected draft contents\nwant: %q\n got: %q", want, got)
	}

	if len(session.history) != 3 {
		t.Fatalf("expected 3 history entries (system+prompt+response), got %d", len(session.history))
	}
}

func TestGenerateRevision(t *testing.T) {
	model := &fakeModel{responses: []string{"# Draft\ncontent", "# Draft v2\nupdated"}}
	session := NewSession(model, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, SessionConfig{})

	brief := sessionBrief{Project: "Project X"}
	if err := session.generateInitialDraft(context.Background(), brief); err != nil {
		t.Fatalf("initial draft: %v", err)
	}

	if err := session.generateRevision(context.Background(), "Tighten the risk section"); err != nil {
		t.Fatalf("generateRevision: %v", err)
	}

	if got, want := session.lastDoc, "# Draft v2\nupdated"; got != want {
		t.Fatalf("unexpected last document\nwant: %q\n got: %q", want, got)
	}

	if len(session.history) != 5 {
		t.Fatalf("expected 5 history entries after revision, got %d", len(session.history))
	}
}

func TestCollectBriefAbort(t *testing.T) {
	model := &fakeModel{responses: []string{"ignored"}}
	session := NewSession(model, strings.NewReader(":quit\n"), &bytes.Buffer{}, &bytes.Buffer{}, SessionConfig{})

	if _, err := session.collectBrief(context.Background()); !errors.Is(err, ErrAborted) {
		t.Fatalf("expected ErrAborted, got %v", err)
	}
}

func TestSaveDocumentWritesFile(t *testing.T) {
	model := &fakeModel{}
	session := NewSession(model, strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, SessionConfig{})
	session.lastDoc = "hello world"

	dir := t.TempDir()
	target := filepath.Join(dir, "doc.md")
	if err := session.saveDocument(target, true); err != nil {
		t.Fatalf("saveDocument: %v", err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("reading saved file: %v", err)
	}
	if got, want := string(data), "hello world"; got != want {
		t.Fatalf("unexpected file contents\nwant: %q\n got: %q", want, got)
	}
}
