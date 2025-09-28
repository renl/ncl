/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"runtime"

	"github.com/renl/ncl/internal/cmdutil"
	"github.com/renl/ncl/internal/process"
	"github.com/spf13/cobra"
)

// psCmd represents the ps command
var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List running processes with their IDs",
	Long: `List running processes with their IDs.
	
This command displays a list of currently running processes along with their process IDs (PIDs).
The output format may vary slightly between operating systems but will generally include the PID
and process name/command line.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmdutil.VPrintf(cmd, "ps command called on %s/%s", runtime.GOOS, runtime.GOARCH)
		
		procs, err := process.ListProcesses()
		if err != nil {
			return fmt.Errorf("failed to list processes: %w", err)
		}
		
		// Print header
		fmt.Fprintf(cmd.OutOrStdout(), "%-10s %s\n", "PID", "NAME")
		fmt.Fprintf(cmd.OutOrStdout(), "%-10s %s\n", "---", "----")
		
		// Print processes
		for _, proc := range procs {
			fmt.Fprintf(cmd.OutOrStdout(), "%-10d %s\n", proc.PID, proc.Name)
		}
		
		return nil
	},
}

func init() {
	rootCmd.AddCommand(psCmd)
}