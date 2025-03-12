package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lpsm-dev/gtoc/internal/generator"
	"github.com/spf13/cobra"
)

var (
	filePath     string
	depth        int
	excludePaths string
	dryRun       bool
	language     string
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate [file]",
	Short: "Generate a table of contents for a markdown file",
	Long: `Generate a table of contents based on the headings in a markdown file
and update the file with the generated table of contents.

Example:
  gtoc generate README.md
  gtoc generate --file docs/index.md
  gtoc generate docs/index.md --depth 3`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if file is provided as positional argument
		if len(args) > 0 && filePath == "" {
			filePath = args[0]
		}

		if filePath == "" {
			return fmt.Errorf("file path is required (provide it as an argument or with --file flag)")
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

		// Parse exclude paths
		excludeList := []string{}
		if excludePaths != "" {
			excludeList = strings.Split(excludePaths, ",")
			for i, path := range excludeList {
				excludeList[i] = strings.TrimSpace(path)
			}
		}

		// Generate table of contents
		gen := generator.NewGenerator(
			absFilePath,
			depth,
			excludeList,
			language,
		)

		toc, err := gen.Generate()
		if err != nil {
			return fmt.Errorf("failed to generate table of contents: %w", err)
		}

		// Update file
		if dryRun {
			fmt.Println("Dry run mode. The following table of contents would be generated:")
			fmt.Println(toc)
		} else {
			if err := gen.UpdateFile(toc); err != nil {
				return fmt.Errorf("failed to update file: %w", err)
			}
			fmt.Printf("Successfully updated %s with the generated table of contents\n", filePath)
		}

		return nil
	},
}

func init() {
	generateCmd.Flags().StringVar(&filePath, "file", "", "Path to the markdown file to update")
	generateCmd.Flags().IntVar(&depth, "depth", 0, "Maximum heading depth (0 for unlimited)")
	generateCmd.Flags().StringVar(&excludePaths, "exclude", "", "Comma-separated list of heading patterns to exclude")
	generateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing")
	generateCmd.Flags().StringVar(&language, "language", "pt", "Language for the table of contents title (pt or en)")
}
