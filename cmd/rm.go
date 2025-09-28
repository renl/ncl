/*
Copyright © 2025 NAME HERE
*/
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/renl/ncl/internal/cmdutil"
	"github.com/spf13/cobra"
)

var (
	rmRecursive bool
	rmForce     bool
)

// rmCmd represents the rm command
var rmCmd = &cobra.Command{
	Use:   "rm [flags] <path>...",
	Short: "Remove files and directories",
	Long:  "Remove files. With -r, remove directories and their contents recursively. With -f, ignore nonexistent targets and never prompt.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		failCount := 0

		cmdutil.VPrintf(cmd, "rm start: recursive=%v, force=%v, targets=%d", rmRecursive, rmForce, len(args))

		for _, p := range args {
			info, err := os.Lstat(p)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					if rmForce {
						cmdutil.VPrintf(cmd, "skip nonexistent (force): %s", p)
						continue
					}
					fmt.Fprintf(cmd.ErrOrStderr(), "rm: cannot remove '%s': no such file or directory\n", p)
				} else {
					fmt.Fprintf(cmd.ErrOrStderr(), "rm: cannot access '%s': %v\n", p, err)
				}
				failCount++
				continue
			}

			if info.IsDir() {
				if !rmRecursive {
					// GNU rm -f still fails on directories without -r; suppress message if -f.
					if !rmForce {
						fmt.Fprintf(cmd.ErrOrStderr(), "rm: cannot remove '%s': is a directory (use -r to remove directories)\n", p)
					} else {
						cmdutil.VPrintf(cmd, "skip directory without -r (force): %s", p)
					}
					failCount++
					continue
				}
				cmdutil.VPrintf(cmd, "removing directory recursively: %s", p)
				if err := os.RemoveAll(p); err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "rm: failed to remove '%s': %v\n", p, err)
					failCount++
				} else {
					cmdutil.VPrintf(cmd, "removed directory: %s", p)
				}
				continue
			}

			// file or symlink
			cmdutil.VPrintf(cmd, "removing file: %s", p)
			if err := os.Remove(p); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "rm: failed to remove '%s': %v\n", p, err)
				failCount++
			} else {
				cmdutil.VPrintf(cmd, "removed file: %s", p)
			}
		}

		if failCount > 0 {
			return fmt.Errorf("rm: %d path(s) failed", failCount)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rmCmd)
	// Flags
	rmCmd.Flags().BoolVarP(&rmRecursive, "recursive", "r", false, "remove directories and their contents recursively")
	rmCmd.Flags().BoolVarP(&rmForce, "force", "f", false, "ignore nonexistent files and arguments; never prompt")
}
