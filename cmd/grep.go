/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var recursive bool

// grepCmd represents the grep command
var grepCmd = &cobra.Command{
	Use:   "grep [-r] <pattern> [files_or_dirs_or_-...]",
	Short: "Search for a regex pattern in files or stdin",
	Long: `Search for a regular expression pattern in files or standard input.

Examples:
  ncl grep TODO main.go
  echo "hello world" | ncl grep "hello"
  someCommand | ncl grep - "error"   # (also supported but pattern must remain first argument; prefer: someCommand | ncl grep "error")

When no file arguments are provided (and -r is not set), input is read from stdin.
Use -r to recurse into directories. Pass '-' explicitly to also read from stdin amidst other files.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pattern := args[0]
		re, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}

		targets := args[1:]
		if verbose {
			fmt.Printf("[verbose] grep pattern=%q recursive=%v targets=%v\n", pattern, recursive, targets)
		}

		if recursive {
			if len(targets) == 0 {
				targets = []string{"."}
				if verbose {
					fmt.Println("[verbose] no targets provided; defaulting to current directory '.'")
				}
			}
			for _, t := range targets {
				if verbose {
					fmt.Printf("[verbose] walking: %s\n", t)
				}
				if err := grepRecursive(t, re); err != nil {
					return err
				}
			}
			return nil
		}

		// If no targets and not recursive, read from stdin
		if len(targets) == 0 {
			if verbose {
				fmt.Println("[verbose] reading from stdin (no targets provided)")
			}
			return grepReader("(stdin)", os.Stdin, re)
		}

		for _, f := range targets {
			if f == "-" { // explicit stdin
				if verbose {
					fmt.Println("[verbose] reading from explicit '-' (stdin)")
				}
				if err := grepReader("(stdin)", os.Stdin, re); err != nil {
					return err
				}
				continue
			}
			if err := grepFile(f, re); err != nil {
				return err
			}
		}
		return nil
	},
}

func grepRecursive(root string, re *regexp.Regexp) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Skip VCS directories
			base := filepath.Base(path)
			if base == ".git" || base == ".hg" || base == ".svn" {
				if verbose {
					fmt.Printf("[verbose] skipping VCS directory: %s\n", path)
				}
				return filepath.SkipDir
			}
			return nil
		}
		if verbose {
			fmt.Printf("[verbose] scanning file: %s\n", path)
		}
		return grepFile(path, re)
	})
}

func grepFile(path string, re *regexp.Regexp) error {
	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if fi.IsDir() {
		// Skip directories in non-recursive mode
		if verbose {
			fmt.Printf("[verbose] skipping directory (non-recursive): %s\n", path)
		}
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	if verbose {
		fmt.Printf("[verbose] opened file: %s (size=%d bytes)\n", path, fi.Size())
	}

	s := bufio.NewScanner(f)
	lineNo := 0
	matches := 0
	for s.Scan() {
		lineNo++
		line := s.Text()
		highlighted, count := highlightMatches(line, re)
		if count > 0 {
			matches += count
			fmt.Printf("%s:%d: %s\n", path, lineNo, highlighted)
		}
	}
	if err := s.Err(); err != nil {
		return fmt.Errorf("scan %s: %w", path, err)
	}
	if verbose {
		fmt.Printf("[verbose] file done: %s (matches=%d)\n", path, matches)
	}
	return nil
}

// grepReader scans an arbitrary io.Reader (typically stdin) and applies the regex.
// It mimics grepFile output formatting using the provided name label.
func grepReader(name string, r io.Reader, re *regexp.Regexp) error {
	if verbose {
		fmt.Printf("[verbose] scanning stream: %s\n", name)
	}
	s := bufio.NewScanner(r)
	lineNo := 0
	matches := 0
	for s.Scan() {
		lineNo++
		line := s.Text()
		highlighted, count := highlightMatches(line, re)
		if count > 0 {
			matches += count
			fmt.Printf("%s:%d: %s\n", name, lineNo, highlighted)
		}
	}
	if err := s.Err(); err != nil {
		return fmt.Errorf("scan %s: %w", name, err)
	}
	if verbose {
		fmt.Printf("[verbose] stream done: %s (matches=%d)\n", name, matches)
	}
	return nil
}

const (
	ansiRed   = "\x1b[31m"
	ansiReset = "\x1b[0m"
)

func highlightMatches(line string, re *regexp.Regexp) (string, int) {
	idxs := re.FindAllStringIndex(line, -1)
	if len(idxs) == 0 {
		return line, 0
	}
	var b strings.Builder
	last := 0
	for _, idx := range idxs {
		start, end := idx[0], idx[1]
		b.WriteString(line[last:start])
		b.WriteString(ansiRed)
		b.WriteString(line[start:end])
		b.WriteString(ansiReset)
		last = end
	}
	b.WriteString(line[last:])
	return b.String(), len(idxs)
}

func init() {
	rootCmd.AddCommand(grepCmd)

	grepCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recurse into directories")
}
