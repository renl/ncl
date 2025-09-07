# ncl (Neo Command Line)

A lightweight, batteries-included grab‑bag of tiny developer convenience commands written in Go and powered by [Cobra](https://github.com/spf13/cobra).

Current commands:
- `grep`  – Regex search across files (with optional recursive directory walk + colored highlights)
- `touch` – Create an empty file (like the Unix utility)

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
Add-Content $PROFILE "`ncl completion powershell | Out-String | Invoke-Expression`"
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

Example snippet:
```text
[verbose] grep pattern="TODO" recursive=true targets=[]
[verbose] no targets provided; defaulting to current directory '.'
[verbose] walking: .
[verbose] scanning file: cmd/grep.go
...
```

---
## Development Notes
Project structure:
```
cmd/
  root.go   # Root command & global flags
  grep.go   # grep implementation
  touch.go  # touch implementation
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

---
## Quick Reference
```text
Install:    go install github.com/renl/ncl@develop
Help:       ncl --help
Grep:       ncl grep -r "pattern" .
Touch:      ncl touch newfile.txt
Verbose:    ncl -v grep "TODO" main.go
Completion: ncl completion powershell | Out-String | Invoke-Expression
```

Happy hacking!
