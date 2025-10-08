/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// techleadCmd represents the techlead command
var techleadCmd = &cobra.Command{
	Use:   "techlead",
	Short: "Leadership helpers for planning and delivery",
	Long: `The techlead command houses collaborative tooling that helps engineering
leaders run faster. Use its subcommands to generate tech lead documents,
structure plans, and partner with AI to explore delivery scenarios.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(techleadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// techleadCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// techleadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
