package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	readmePath string
)

// analyzeCmd represents the analyze command
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze and update README.md with best practices",
	Long: `Analyze the README.md file and add best practices elements such as:
- <!-- BEGIN_DOCS --> and <a name="readme-top"></a> in the header
- <p align="right">(<a href="#readme-top">back to top</a>)</p> at the end of each main heading (#)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if readmePath == "" {
			readmePath = "README.md"
		}

		// Convert file path to absolute path
		absFilePath, err := filepath.Abs(readmePath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}

		// Check if file exists
		if _, err := os.Stat(absFilePath); os.IsNotExist(err) {
			return fmt.Errorf("file does not exist: %s", absFilePath)
		}

		// Read the file
		content, err := os.ReadFile(absFilePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		contentStr := string(content)
		lines := strings.Split(contentStr, "\n")

		// Check and add header elements
		if !strings.Contains(contentStr, "<!-- BEGIN_DOCS -->") {
			lines = append([]string{"<!-- BEGIN_DOCS -->", "<a name=\"readme-top\"></a>", ""}, lines...)
		}

		// Add back to top link at the end of each main heading
		headingRegex := regexp.MustCompile(`^#\s+(.+)$`)
		for i, line := range lines {
			if headingRegex.MatchString(line) {
				headingText := headingRegex.FindStringSubmatch(line)[1]
				// Skip the "Table of Contents" heading
				if strings.ToLower(headingText) == "table of contents" {
					continue
				}
				// Check if the next line is the back to top link
				if i+1 < len(lines) && strings.Contains(lines[i+1], "<p align=\"right\">(<a href=\"#readme-top\">back to top</a>)</p>") {
					continue
				}
				lines = append(lines[:i+1], append([]string{"<p align=\"right\">(<a href=\"#readme-top\">back to top</a>)</p>"}, lines[i+1:]...)...)
			}
		}

		// Write the updated content back to the file
		newContent := strings.Join(lines, "\n")
		if err := os.WriteFile(absFilePath, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}

		fmt.Printf("Successfully updated %s with best practices\n", readmePath)
		return nil
	},
}

func init() {
	analyzeCmd.Flags().StringVar(&readmePath, "file", "README.md", "Path to the README.md file to analyze")
	RootCmd.AddCommand(analyzeCmd)
}
