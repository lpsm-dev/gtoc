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
- <p align="right">(<a href="#readme-top">back to top</a>)</p> at the end of each main heading (#) section`,
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

		// Add BEGIN_DOCS marker if not present
		if !strings.Contains(contentStr, "<!-- BEGIN_DOCS -->") {
			contentStr = "<!-- BEGIN_DOCS -->\n<a name=\"readme-top\"></a>\n\n" + contentStr
		}

		// Add END_DOCS marker if not present
		if !strings.Contains(contentStr, "<!-- END_DOCS -->") {
			contentStr = contentStr + "\n<!-- END_DOCS -->"
		}

		// Find H1 headings and add "back to top" links
		h1Regex := regexp.MustCompile(`(?m)^#\s+(.+)$`)
		backToTopLink := "<p align=\"right\">(<a href=\"#readme-top\">back to top</a>)</p>"

		// Find all H1 sections (from one H1 to the next H1 or end)
		sections := h1Regex.FindAllStringIndex(contentStr, -1)

		if len(sections) > 0 {
			// Create a new content string with back to top links
			var newContentBuilder strings.Builder
			lastPos := 0

			for i, match := range sections {
				start := match[0]
				end := len(contentStr)

				// If not the last section, end is the start of the next section
				if i < len(sections)-1 {
					end = sections[i+1][0]
				}

				sectionText := contentStr[start:end]

				// Add content before this section if not the first section
				if i > 0 {
					// Only add the "back to top" link if it's not already there
					prevSectionText := contentStr[lastPos:start]
					if !strings.Contains(prevSectionText, backToTopLink) {
						// Ensure there's a blank line before the link
						if !strings.HasSuffix(prevSectionText, "\n\n") && !strings.HasSuffix(prevSectionText, "\n\r\n") {
							if strings.HasSuffix(prevSectionText, "\n") {
								newContentBuilder.WriteString("\n")
							} else {
								newContentBuilder.WriteString("\n\n")
							}
						}
						newContentBuilder.WriteString(backToTopLink + "\n\n")
					}
				}

				// Add the current section
				newContentBuilder.WriteString(sectionText)
				lastPos = end
			}

			// Add back to top link after the last section if needed
			lastSection := contentStr[lastPos:]
			if !strings.Contains(lastSection, backToTopLink) && !strings.HasSuffix(lastSection, "<!-- END_DOCS -->") {
				if !strings.HasSuffix(lastSection, "\n\n") && !strings.HasSuffix(lastSection, "\n\r\n") {
					if strings.HasSuffix(lastSection, "\n") {
						newContentBuilder.WriteString("\n")
					} else {
						newContentBuilder.WriteString("\n\n")
					}
				}
				newContentBuilder.WriteString(backToTopLink + "\n\n")
			}

			// Update the content string
			contentStr = newContentBuilder.String()
		}

		// Write the updated content back to the file
		if err := os.WriteFile(absFilePath, []byte(contentStr), 0644); err != nil {
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
