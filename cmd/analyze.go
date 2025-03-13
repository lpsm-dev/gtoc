package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lpsm-dev/gtoc/internal/logger"
	"github.com/spf13/cobra"
)

var readmePath string

// Constant markers
const (
	beginDocsMarker = "<!-- BEGIN_DOCS -->"
	endDocsMarker   = "<!-- END_DOCS -->"
	readmeAnchor    = "<a name=\"readme-top\"></a>"
	backToTopLink   = "<p align=\"right\">(<a href=\"#readme-top\">back to top</a>)</p>"
)

// analyzeCmd adds best practices formatting to README files
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze and update README.md with best practices",
	Long: `Analyze the README.md file and add best practices elements such as:
- <!-- BEGIN_DOCS --> and <a name="readme-top"></a> in the header
- <p align="right">(<a href="#readme-top">back to top</a>)</p> at the end of each main heading (#) section`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Use default README.md if no path provided
		if readmePath == "" {
			readmePath = "README.md"
		}

		logger.Debug("Analyzing README file", "path", readmePath)

		// Validate file exists
		absFilePath, err := filepath.Abs(readmePath)
		if err != nil {
			logger.Error("Failed to get absolute path", "error", err)
			return fmt.Errorf("failed to get absolute path: %w", err)
		}

		if _, err := os.Stat(absFilePath); os.IsNotExist(err) {
			logger.Error("File does not exist", "path", absFilePath)
			return fmt.Errorf("file does not exist: %s", absFilePath)
		}

		// Read file content
		logger.Debug("Reading file", "path", absFilePath)
		content, err := os.ReadFile(absFilePath)
		if err != nil {
			logger.Error("Failed to read file", "error", err)
			return fmt.Errorf("failed to read file: %w", err)
		}
		contentStr := string(content)

		// Add header markers if needed
		if !strings.Contains(contentStr, beginDocsMarker) {
			logger.Info("Adding header markers", "marker", beginDocsMarker)
			contentStr = beginDocsMarker + "\n" + readmeAnchor + "\n\n" + contentStr
		}

		// Process headings and add back-to-top links
		logger.Info("Processing headings and adding back-to-top links")
		contentStr = addBackToTopLinks(contentStr)

		// Write updated content
		logger.Debug("Writing updated content to file", "path", absFilePath)
		if err := os.WriteFile(absFilePath, []byte(contentStr), 0644); err != nil {
			logger.Error("Failed to write file", "error", err)
			return fmt.Errorf("failed to write file: %w", err)
		}

		logger.Info("File updated successfully", "path", readmePath)
		fmt.Printf("Successfully updated %s with best practices\n", readmePath)
		return nil
	},
}

// addBackToTopLinks adds "back to top" links after each h1 heading
func addBackToTopLinks(content string) string {
	// Check if END_DOCS marker is present
	hasEndDocsMarker := strings.Contains(content, endDocsMarker)
	logger.Debug("Checking for END_DOCS marker", "present", hasEndDocsMarker)

	// Find H1 headings
	h1Regex := regexp.MustCompile(`(?m)^#\s+(.+)$`)
	matches := h1Regex.FindAllStringIndex(content, -1)
	logger.Debug("Found headings", "count", len(matches))

	if len(matches) == 0 {
		// No headings found, just add end marker if needed
		if !hasEndDocsMarker {
			logger.Info("No headings found, adding END_DOCS marker")
			content = content + "\n" + endDocsMarker + "\n"
		}
		return content
	}

	// Process content with headings
	var newContent strings.Builder
	contentWithoutEnd := content
	
	// Remove end marker temporarily if it exists
	if hasEndDocsMarker {
		endPos := strings.LastIndex(content, endDocsMarker)
		contentWithoutEnd = content[:endPos]
		logger.Debug("Temporarily removed END_DOCS marker")
	}

	// Add content before first heading
	if matches[0][0] > 0 {
		newContent.WriteString(contentWithoutEnd[:matches[0][0]])
	}

	// Process each heading section
	for i, match := range matches {
		start := match[0]
		end := len(contentWithoutEnd)
		
		// If not the last heading, end at next heading
		if i < len(matches)-1 {
			end = matches[i+1][0]
		}

		section := contentWithoutEnd[start:end]
		heading := h1Regex.FindStringSubmatch(section)[1]
		logger.Debug("Processing heading", "heading", heading, "index", i+1)
		
		// Add back-to-top link if not already present
		if !strings.Contains(section, backToTopLink) {
			logger.Debug("Adding back-to-top link to heading", "heading", heading)
			headingEnd := strings.Index(section, "\n")
			if headingEnd == -1 {
				headingEnd = len(section)
			}
			
			headingLine := section[:headingEnd+1]
			contentAfter := ""
			
			if headingEnd+1 < len(section) {
				contentAfter = section[headingEnd+1:]
			}
			
			newContent.WriteString(headingLine)
			newContent.WriteString(contentAfter)
			
			// Ensure proper spacing before link
			if !strings.HasSuffix(contentAfter, "\n\n") {
				if strings.HasSuffix(contentAfter, "\n") {
					newContent.WriteString("\n")
				} else {
					newContent.WriteString("\n\n")
				}
			}
			
			newContent.WriteString(backToTopLink + "\n\n")
		} else {
			logger.Debug("Back-to-top link already exists for heading", "heading", heading)
			newContent.WriteString(section)
		}
	}

	// Add end marker
	logger.Debug("Adding END_DOCS marker")
	newContent.WriteString(endDocsMarker + "\n")
	
	return newContent.String()
}

func init() {
	analyzeCmd.Flags().StringVar(&readmePath, "file", "README.md", "Path to the README.md file to analyze")
	RootCmd.AddCommand(analyzeCmd)
}
