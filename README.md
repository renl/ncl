# ncl (Neo Command Line)

A lightweight, batteries-included grab‑bag of tiny developer convenience commands written in Go and powered by [Cobra](https://github.com/spf13/cobra).

Current commands:
- `grep`  – Regex search across files (with optional recursive directory walk + colored highlights)
- `touch` – Create an empty file (like the Unix utility)
- `rm`    – Remove files and directories (with optional recursive and force modes)
- `ps`    – List running processes with their IDs
- `techlead gendoc` – Start an interactive AI-assisted tech lead document session

Global flag:
- `-v, --verbose` – Extra diagnostic / tracing output for any command

---
## Installation

### Using `go install` (recommended)
Requires Go toolchain (module file declares **go 1.25**).

```powershell
# Installs the develop branch
go install github.com/renl/ncl@develop
```
The resulting executable will be placed in your `$GOBIN` (or `$GOPATH/bin` if `GOBIN` unset). Ensure that directory is on your `PATH`.

Verify:
```powershell
ncl --help
```

### Building from source (local clone)
```powershell
git clone -b develop https://github.com/renl/ncl.git
cd ncl
# Standard build
go build -o ncl.exe ./...
# Or install into GOBIN
go install ./...
```

### Updating
```powershell
go install github.com/renl/ncl@develop
```
(Go will rebuild only if there are changes.)

---
## Usage Overview
Run `ncl --help` for the root help, or `ncl <command> --help` for any subcommand.

### Global Flags
| Flag | Description |
|------|-------------|
| `-v, --verbose` | Emit additional progress / trace messages (prefixed with `[verbose]`). |

---
## Commands

### `grep`
Search for a (RE2) regular expression in files or standard input. Matches are color‑highlighted in red using ANSI escape codes.

Synopsis:
```text
ncl grep [-r] <pattern> [files_or_dirs_or_-...]
```

Key behaviors:
- If no file arguments are provided (and `-r` is not used), input is read from stdin.
- Pass `-` explicitly among other file names to also read stdin.
- With `-r/--recursive`, you may pass directories (or nothing, which defaults to `.`).
- Skips VCS directories: `.git`, `.hg`, `.svn`.
- Prints lines in the format: `path:line_number: highlighted line` (stdin is labeled `(stdin)`).

Examples:
```powershell
# Search a single file
ncl grep "TODO" main.go

# Recursive search from current directory
ncl grep -r "(?i)error" .

# Multiple targets
ncl grep "^package" cmd/*.go

# Verbose recursive search
ncl -v grep -r "Compile" ./cmd

# Pipe from another command (stdin auto-detected)
git diff | ncl grep "TODO"

# Explicit stdin among other files
type somefile.txt | ncl grep "pattern" - main.go
```

Exit codes:
- `0` on success (even if no matches)
- Non-zero on invalid regex, unreadable files, or traversal errors

### `touch`
Create an empty file (or truncate if it already exists) similar to Unix `touch`.

Synopsis:
```text
ncl touch <filename>
```

Example:
```powershell
ncl touch demo.txt
```

Verbose mode prints file metadata after creation.

### `rm`
Remove files and directories, with support for recursive removal and force mode.

Synopsis:
```text
ncl rm [-r] [-f] <path>...
```

Key behaviors:
- By default, removes only files (not directories)
- With `-r/--recursive`, removes directories and their contents recursively
- With `-f/--force`, ignores nonexistent files and never prompts
- Follows standard Unix `rm` behavior for error handling

Examples:
```powershell
# Remove a file
ncl rm unwanted.txt

# Remove multiple files
ncl rm file1.txt file2.txt

# Remove directory recursively
ncl rm -r unwanted_dir

# Force remove (ignore missing files)
ncl rm -f maybe_missing.txt

# Remove directory recursively, ignoring errors
ncl rm -rf unwanted_dir
```

Exit codes:
- `0` on success
- Non-zero when any removal fails

### `ps`
List running processes with their process IDs.

Synopsis:
```text
ncl ps
```

Key behaviors:
- Shows a list of currently running processes
- Displays process ID (PID) and process name
- Works across Windows, macOS, and Linux systems
- Uses native OS APIs for optimal performance when available

Examples:
```powershell
# List all running processes
ncl ps

# List processes with verbose output
ncl -v ps
```

Exit codes:
- `0` on success
- Non-zero when process listing fails

### `techlead gendoc`
Launch an interactive workshop that gathers project context and co-writes a Tech Lead planning document with an LLM via LangChainGo.

Synopsis:
```text
ncl techlead gendoc [flags]
```

Workflow highlights:
- Guided intake collects project name, context, goals, architecture direction, risks, timeline, and stakeholders.
- Generates a Markdown tech lead plan (Summary, Goals & Non-Goals, Architecture, Implementation Plan, Risks, Stakeholders, Timeline, Open Questions, Next Steps).
- Iterative refinement loop — type feedback to refresh the draft, or `:help` to discover session commands.
- `:save [path]` saves the current draft at any point; providing `--out <file>` auto-saves on exit.
- Supports OpenAI models via LangChainGo (`OPENAI_API_KEY` env or `--api-key`).

Examples:
```powershell
# Start a session using the default OpenAI model
ncl techlead gendoc

# Specify a model and auto-save to a file on exit
ncl techlead gendoc --model gpt-4o-mini --out .\techlead.md

# Override the API key and base URL (Azure / custom proxy)
ncl techlead gendoc --api-key "$env:OPENAI_API_KEY" --base-url https://example.openai.azure.com/v1
```

Exit codes:
- `0` on normal completion or if the user exits during the briefing stage
- Non-zero when provider configuration fails (e.g., missing API key)

---
## Shell Completion
Cobra's auto-generated `completion` command is enabled. Generate completions for your shell to get tab completion of commands & flags.

List available shells:
```powershell
ncl completion --help
```

### PowerShell (Windows)
Temporary for current session:
```powershell
ncl completion powershell | Out-String | Invoke-Expression
```
Persist across sessions (adds to profile):
```powershell
# Ensure profile exists
if (!(Test-Path -Path $PROFILE)) { New-Item -Type File -Path $PROFILE -Force | Out-Null }
Add-Content $PROFILE "`ncl completion powershell | Out-String | Invoke-Expression"
```
Restart PowerShell (or `.& $PROFILE`).

### Bash
```bash
# One-off (Linux/macOS)
source <(ncl completion bash)
# Persist (Linux example)
ncl completion bash > /etc/bash_completion.d/ncl
```

### Zsh
```zsh
ncl completion zsh > ~/.zfunc/_ncl
print -l 'fpath+=(~/.zfunc)' >> ~/.zshrc
autoload -Uz compinit && compinit
```

### Fish
```fish
ncl completion fish | source
ncl completion fish > ~/.config/fish/completions/ncl.fish
```

---
## Color Output
Match highlighting uses ANSI escape codes. On Windows 10+ (modern terminals / VS Code integrated terminal / Windows Terminal) this is supported by default. If you see raw sequences like `\x1b[31m`, enable virtual terminal processing or use a compatible terminal emulator.

---
## Verbose Diagnostics
Adding `-v` prints:
- Target resolution & defaults (e.g., when grep falls back to `.`)
- Each directory walked (recursive grep)
- Each file opened & summary of match counts
- Skipped directories & reasons
- Details about file/directory removal operations
- Platform information for process listing

Example snippet:
```text
[verbose] grep pattern="TODO" recursive=true targets=[]
[verbose] no targets provided; defaulting to current directory '.'
[verbose] walking: .
[verbose] scanning file: cmd/grep.go
...
[verbose] rm start: recursive=true, force=false, targets=1
[verbose] removing directory recursively: test_dir
[verbose] removed directory: test_dir
...
[verbose] ps command called on windows/amd64
```

---
## Development Notes
Project structure:
```
cmd/
  root.go      # Root command & global flags
  grep.go      # grep implementation
  touch.go     # touch implementation
  rm.go        # rm implementation
  ps.go        # ps command interface
internal/
  cmdutil/
    logging.go  # Shared utility functions for command output
  process/     # Process listing functionality
    process.go      # Process struct definition
    ps_windows.go   # Windows-specific process listing
    ps_unix.go      # Unix-specific process listing
    ps_fallback.go  # Fallback implementation using external commands
main.go     # Entry point calling cmd.Execute()
```

Add a new command:
1. Create a new file in `cmd/` (e.g. `foo.go`).
2. Define a `cobra.Command` and register in its `init()` with `rootCmd.AddCommand(fooCmd)`.
3. Rebuild / reinstall.

Run with race detector while developing:
```powershell
go run -race ./...
```

Run vet & tidy:
```powershell
go vet ./...
go mod tidy
```

---
## License
See [LICENSE](LICENSE) for details.

---
## Troubleshooting
| Issue | Resolution |
|-------|------------|
| `ncl: command not found` | Ensure `$GOBIN` / `$GOPATH/bin` is on PATH. Run `go env GOBIN GOPATH`. |
| Colors not showing | Use Windows Terminal / VS Code terminal or enable ANSI support. |
| PowerShell completion not loading | Reopen shell or `.& $PROFILE`; confirm the profile line was appended. |
| Invalid regex error | Verify your pattern is valid RE2 (Go regexp). Test with `go test` or simplify pattern. |
| Process listing fails | Ensure you have appropriate permissions to list processes on your system. |

---
## Quick Reference
```text
Install:    go install github.com/renl/ncl@develop
Help:       ncl --help
Grep:       ncl grep -r "pattern" .
Touch:      ncl touch newfile.txt
Remove:     ncl rm unwanted.txt
Processes:  ncl ps
Verbose:    ncl -v ps
Completion: ncl completion powershell | Out-String | Invoke-Expression
```

Happy hacking!