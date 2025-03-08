package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lpsm-dev/gtoc/internal/generator"
	"github.com/lpsm-dev/gtoc/internal/git"
	"github.com/spf13/cobra"
)

var (
	filePath     string
	depth        int
	pattern      string
	excludePaths string
	dryRun       bool
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate [file]",
	Short: "Generate a markdown index and update the specified file",
	Long: `Generate a hierarchical index of markdown files in your Git repository
and update the specified markdown file with the generated index.

Example:
  gtoc generate README.md
  gtoc generate --file README.md
  gtoc generate docs/index.md --depth 2 --pattern "docs/**/*.md"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if file path is provided as a positional argument
		if len(args) > 0 {
			filePath = args[0]
		}

		if filePath == "" {
			return fmt.Errorf("file path is required")
		}

		// Convert file path to absolute path
		absFilePath, err := filepath.Abs(filePath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}

		// Check if file exists
		if _, err := os.Stat(absFilePath); os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", absFilePath)
		}

		// Get repository root
		repoRoot, err := git.GetRepositoryRoot()
		if err != nil {
			return fmt.Errorf("failed to get repository root: %w", err)
		}

		// Parse exclude paths
		excludeList := []string{}
		if excludePaths != "" {
			excludeList = strings.Split(excludePaths, ",")
			for i, path := range excludeList {
				excludeList[i] = strings.TrimSpace(path)
			}
		}

		// Generate index
		gen := generator.NewGenerator(
			repoRoot,
			absFilePath,
			depth,
			pattern,
			excludeList,
		)

		index, err := gen.Generate()
		if err != nil {
			return fmt.Errorf("failed to generate index: %w", err)
		}

		// Update file
		if dryRun {
			fmt.Println("Dry run mode. The following index would be generated:")
			fmt.Println(index)
		} else {
			if err := gen.UpdateFile(index); err != nil {
				return fmt.Errorf("failed to update file: %w", err)
			}
			fmt.Printf("Successfully updated %s with the generated index\n", filePath)
		}

		return nil
	},
}

func init() {
	generateCmd.Flags().StringVar(&filePath, "file", "", "Path to the markdown file to update (required if not provided as positional argument)")
	generateCmd.Flags().IntVar(&depth, "depth", 0, "Maximum directory depth (0 for unlimited)")
	generateCmd.Flags().StringVar(&pattern, "pattern", "**/*.md", "Glob pattern to filter markdown files")
	generateCmd.Flags().StringVar(&excludePaths, "exclude", "", "Comma-separated list of paths to exclude")
	generateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing")
}
