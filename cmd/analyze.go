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

// Constant markers used to delimit and format the generated README sections.
const (
	beginDocsMarker = "<!-- BEGIN_DOCS -->"
	endDocsMarker   = "<!-- END_DOCS -->"
	readmeAnchor    = "<a name=\"readme-top\"></a>"
	backToTopLink   = "<p align=\"right\">(<a href=\"#readme-top\">back to top</a>)</p>"
)

// h1HeadingPattern matches a top-level (H1) markdown heading line.
var h1HeadingPattern = regexp.MustCompile(`^#\s+(.+)$`)

// codeFencePattern matches the leading run of backticks or tildes that opens
// or closes a fenced code block.
var codeFencePattern = regexp.MustCompile("^(`{3,}|~{3,})")

// analyzeCmd adds best practices formatting to README files.
var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze and update README.md with best practices",
	Long: `Analyze the README.md file and add best practices elements such as:
- <!-- BEGIN_DOCS --> and <a name="readme-top"></a> in the header
- <p align="right">(<a href="#readme-top">back to top</a>)</p> at the end of each main heading (#) section`,
	RunE: runAnalyze,
}

// runAnalyze reads the target README, adds any missing best-practice
// markers, and writes the updated content back to disk.
func runAnalyze(cmd *cobra.Command, args []string) error {
	if readmePath == "" {
		readmePath = "README.md"
	}

	logger.Debug("Analyzing README file", "path", readmePath)

	absFilePath, err := filepath.Abs(readmePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if _, err := os.Stat(absFilePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", absFilePath)
	}

	logger.Debug("Reading file", "path", absFilePath)
	content, err := os.ReadFile(absFilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	contentStr := string(content)

	if !strings.Contains(contentStr, beginDocsMarker) {
		logger.Info("Adding header markers", "marker", beginDocsMarker)
		contentStr = beginDocsMarker + "\n" + readmeAnchor + "\n\n" + contentStr
	}

	logger.Info("Processing headings and adding back-to-top links")
	contentStr = addBackToTopLinks(contentStr)

	logger.Debug("Writing updated content to file", "path", absFilePath)
	if err := os.WriteFile(absFilePath, []byte(contentStr), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	logger.Info("File updated successfully", "path", readmePath)
	fmt.Printf("Successfully updated %s with best practices\n", readmePath)
	return nil
}

// addBackToTopLinks ensures every top-level (H1) section ends with a
// back-to-top link and that the content is terminated with the END_DOCS
// marker. Any content that already follows an existing END_DOCS marker is
// preserved verbatim after the marker instead of being discarded.
func addBackToTopLinks(content string) string {
	hasEndDocsMarker := strings.Contains(content, endDocsMarker)
	logger.Debug("Checking for END_DOCS marker", "present", hasEndDocsMarker)

	body, trailing := splitAtEndDocsMarker(content)

	starts := findH1LineStarts(body)
	logger.Debug("Found headings", "count", len(starts))

	if len(starts) == 0 {
		if hasEndDocsMarker {
			// Nothing to add; the file is unchanged.
			return content
		}
		logger.Info("No headings found, adding END_DOCS marker")
		return body + "\n" + endDocsMarker + "\n"
	}

	var sb strings.Builder
	sb.WriteString(body[:starts[0]])

	for i, start := range starts {
		end := len(body)
		if i < len(starts)-1 {
			end = starts[i+1]
		}
		sb.WriteString(processSection(body[start:end]))
	}

	logger.Debug("Adding END_DOCS marker")
	sb.WriteString(endDocsMarker)
	sb.WriteString(trailing)
	return sb.String()
}

// splitAtEndDocsMarker splits content around an existing END_DOCS marker. It
// returns the body preceding the marker and the trailing content that must
// be preserved verbatim after it. When no marker is present, the whole
// content is treated as the body and a leading newline is returned as the
// separator for the marker that will be appended.
func splitAtEndDocsMarker(content string) (body, trailing string) {
	if !strings.Contains(content, endDocsMarker) {
		return content, "\n"
	}

	endPos := strings.LastIndex(content, endDocsMarker)
	return content[:endPos], content[endPos+len(endDocsMarker):]
}

// findH1LineStarts returns the byte offsets, within body, of every top-level
// (H1) heading line. Lines inside fenced code blocks (``` or ~~~) are
// skipped so a literal "# " inside a code sample is never treated as a
// heading.
func findH1LineStarts(body string) []int {
	var starts []int
	inFence := false
	fenceMarker := ""
	offset := 0

	for _, line := range strings.Split(body, "\n") {
		if marker := codeFenceRune(line); marker != "" {
			inFence, fenceMarker = toggleFence(inFence, fenceMarker, marker)
		} else if !inFence && h1HeadingPattern.MatchString(line) {
			starts = append(starts, offset)
		}

		offset += len(line) + 1
	}

	return starts
}

// codeFenceRune returns "`" or "~" when line opens or closes a fenced code
// block, or "" otherwise.
func codeFenceRune(line string) string {
	trimmed := strings.TrimLeft(line, " \t")
	match := codeFencePattern.FindString(trimmed)
	if match == "" {
		return ""
	}
	return match[:1]
}

// toggleFence updates fence-tracking state given the marker found on the
// current line.
func toggleFence(inFence bool, currentMarker, marker string) (bool, string) {
	switch {
	case inFence && marker == currentMarker:
		return false, ""
	case !inFence:
		return true, marker
	default:
		return inFence, currentMarker
	}
}

// processSection appends a back-to-top link to a single H1 section (from its
// heading line up to, but not including, the next H1 heading) unless the
// section already contains one.
func processSection(section string) string {
	if strings.Contains(section, backToTopLink) {
		return section
	}

	if !strings.HasSuffix(section, "\n") {
		section += "\n"
	}
	if !strings.HasSuffix(section, "\n\n") {
		section += "\n"
	}

	return section + backToTopLink + "\n\n"
}

func init() {
	analyzeCmd.Flags().StringVar(&readmePath, "file", "README.md", "Path to the README.md file to analyze")
}
