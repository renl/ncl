/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// touchCmd represents the touch command
var touchCmd = &cobra.Command{
	Use:   "touch <filename>",
	Short: "Create an empty file",
	Long:  `Create an empty file with the given filename`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]
		if verbose {
			fmt.Printf("[verbose] creating file: %s\n", filename)
		}
		f, err := os.Create(filename)
		if err != nil {
			fmt.Printf("Failed to create file '%s': %v\n", filename, err)
			return
		}
		defer f.Close()
		if verbose {
			fi, _ := f.Stat()
			fmt.Printf("[verbose] file created; size=%d perms=%#o\n", fi.Size(), fi.Mode().Perm())
		}
		fmt.Printf("Created file: %s\n", filename)
	},
}

func init() {
	rootCmd.AddCommand(touchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// touchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// touchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
