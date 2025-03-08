package cmd

import (
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gtoc",
	Short: "Generate a markdown index for your Git repository's documentation",
	Long: `gtoc is a CLI tool that generates a hierarchical index of markdown files
in your Git repository and updates a specified markdown file with the generated index.

It respects .gitignore rules and provides various customization options.`,
}

func init() {
	RootCmd.AddCommand(generateCmd)
}
