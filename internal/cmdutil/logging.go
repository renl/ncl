package cmdutil

import (
	"fmt"

	"github.com/spf13/cobra"
)

// VPrintf prints a verbose-formatted line if the inherited --verbose flag is true.
// It uses the command's configured stdout so redirections/tests work.
func VPrintf(cmd *cobra.Command, format string, a ...any) {
	if cmd == nil {
		return
	}
	v, _ := cmd.InheritedFlags().GetBool("verbose")
	if !v {
		return
	}
	fmt.Fprintf(cmd.OutOrStdout(), "[verbose] "+format+"\n", a...)
}

// EPrintf prints an error line to the command's stderr.
func EPrintf(cmd *cobra.Command, format string, a ...any) {
	if cmd == nil {
		return
	}
	fmt.Fprintf(cmd.ErrOrStderr(), format+"\n", a...)
}
