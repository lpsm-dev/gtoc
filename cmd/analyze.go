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

		// Check if END_DOCS marker is present
		hasEndDocsMarker := strings.Contains(contentStr, "<!-- END_DOCS -->")

		// Find H1 headings and add "back to top" links
		h1Regex := regexp.MustCompile(`(?m)^#\s+(.+)$`)
		backToTopLink := "<p align=\"right\">(<a href=\"#readme-top\">back to top</a>)</p>"

		// Split content by H1 headings to process each section separately
		matches := h1Regex.FindAllStringIndex(contentStr, -1)

		if len(matches) > 0 {
			var newContent strings.Builder

			// Remove END_DOCS marker temporarily if it exists
			contentWithoutEndDocs := contentStr
			var endDocsPosition int
			if hasEndDocsMarker {
				endDocsPosition = strings.LastIndex(contentStr, "<!-- END_DOCS -->")
				contentWithoutEndDocs = contentStr[:endDocsPosition]
			}

			// Add content before the first H1 heading
			if matches[0][0] > 0 {
				newContent.WriteString(contentWithoutEndDocs[:matches[0][0]])
			}

			// Process each H1 section
			for i, match := range matches {
				start := match[0]
				end := len(contentWithoutEndDocs)

				// If not the last section, end is the start of the next section
				if i < len(matches)-1 {
					end = matches[i+1][0]
				}

				// Extract the section content
				sectionContent := contentWithoutEndDocs[start:end]

				// Check if section already has a back to top link
				if !strings.Contains(sectionContent, backToTopLink) {
					// Get the heading line by finding the first newline after the heading
					firstNewline := strings.Index(sectionContent, "\n")
					if firstNewline == -1 {
						firstNewline = len(sectionContent)
					}

					headingLine := sectionContent[:firstNewline+1]
					contentAfterHeading := ""
					if firstNewline+1 < len(sectionContent) {
						contentAfterHeading = sectionContent[firstNewline+1:]
					}

					// Add the heading followed by content and then back to top link
					newContent.WriteString(headingLine)
					if len(contentAfterHeading) > 0 {
						newContent.WriteString(contentAfterHeading)
						// Ensure there's a blank line before adding the link
						if !strings.HasSuffix(contentAfterHeading, "\n\n") {
							if strings.HasSuffix(contentAfterHeading, "\n") {
								newContent.WriteString("\n")
							} else {
								newContent.WriteString("\n\n")
							}
						}
					}

					// Add the back to top link with a newline after
					newContent.WriteString(backToTopLink)
					newContent.WriteString("\n\n")
				} else {
					// Section already has the link, keep it as is
					newContent.WriteString(sectionContent)
				}
			}

			// Add END_DOCS marker at the end
			newContent.WriteString("<!-- END_DOCS -->\n")

			contentStr = newContent.String()
		} else {
			// If no H1 headings found, just ensure END_DOCS marker is present
			if !hasEndDocsMarker {
				contentStr = contentStr + "\n<!-- END_DOCS -->\n"
			}
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
