package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gtoc",
	Short: "Generate a table of contents for markdown files",
	Long: `gtoc is a CLI tool that generates a table of contents based on the headings
in a markdown file and updates the file with the generated table of contents.`,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(generateCmd)
}
