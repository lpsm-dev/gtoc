package generator

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	tocStartMarker = "<!-- START_TABLE_OF_CONTENTS -->"
	tocEndMarker   = "<!-- END_TABLE_OF_CONTENTS -->"
	backToTopLink  = "<p align=\"right\">(<a href=\"#readme-top\">back to top</a>)</p>"
)

// Generator handles the TOC generation for markdown files
type Generator struct {
	targetFile   string
	maxDepth     int
	excludePaths []string
}

// Heading represents a markdown heading
type Heading struct {
	Level  int
	Text   string
	Anchor string
	Line   int
}

// NewGenerator creates a new Generator
func NewGenerator(targetFile string, maxDepth int, excludePaths []string) *Generator {
	return &Generator{
		targetFile:   targetFile,
		maxDepth:     maxDepth,
		excludePaths: excludePaths,
	}
}

// Generate creates a markdown table of contents
func (g *Generator) Generate() (string, error) {
	headings, err := g.extractHeadings()
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(tocStartMarker + "\n\n")
	
	// Add title
	sb.WriteString("# Summary\n\n")
	
	// Build TOC entries
	for _, heading := range headings {
		indent := strings.Repeat("  ", heading.Level-1)
		sb.WriteString(fmt.Sprintf("%s- [%s](#%s)\n", indent, heading.Text, heading.Anchor))
	}

	sb.WriteString("\n" + backToTopLink + "\n")
	sb.WriteString("\n" + tocEndMarker)
	return sb.String(), nil
}

// extractHeadings parses markdown file to find headings
func (g *Generator) extractHeadings() ([]*Heading, error) {
	file, err := os.Open(g.targetFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	headings := []*Heading{}
	scanner := bufio.NewScanner(file)
	lineNum := 0
	headingPattern := regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		matches := headingPattern.FindStringSubmatch(line)

		if len(matches) > 2 {
			level := len(matches[1])
			text := matches[2]

			// Skip headings deeper than maxDepth
			if g.maxDepth > 0 && level > g.maxDepth {
				continue
			}

			headings = append(headings, &Heading{
				Level:  level,
				Text:   text,
				Anchor: createAnchor(text),
				Line:   lineNum,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return headings, nil
}

// UpdateFile adds or replaces TOC in the markdown file
func (g *Generator) UpdateFile(toc string) error {
	content, err := os.ReadFile(g.targetFile)
	if err != nil {
		return err
	}

	fileContent := string(content)
	startIdx := strings.Index(fileContent, tocStartMarker)
	endIdx := strings.Index(fileContent, tocEndMarker)
	
	var newContent string
	if startIdx != -1 && endIdx != -1 && startIdx < endIdx {
		// Replace existing TOC
		newContent = fileContent[:startIdx] + toc + fileContent[endIdx+len(tocEndMarker):]
	} else {
		// Add TOC at beginning
		newContent = toc + "\n" + fileContent
	}

	return os.WriteFile(g.targetFile, []byte(newContent), 0644)
}

// createAnchor generates a GitHub-compatible anchor from heading text
func createAnchor(text string) string {
	// Convert to lowercase and replace spaces with hyphens
	anchor := strings.ToLower(text)
	anchor = strings.ReplaceAll(anchor, " ", "-")
	
	// Remove special characters
	nonAlphanumeric := regexp.MustCompile(`[^a-z0-9-]`)
	anchor = nonAlphanumeric.ReplaceAllString(anchor, "")
	
	// Replace multiple hyphens with a single one
	multipleHyphens := regexp.MustCompile(`-+`)
	anchor = multipleHyphens.ReplaceAllString(anchor, "-")
	
	// Remove leading and trailing hyphens
	return strings.Trim(anchor, "-")
}

// GetFileWithUpdatedTOC retorna o conteúdo do arquivo com o TOC atualizado
// sem realmente escrever no arquivo (útil para preview)
func (g *Generator) GetFileWithUpdatedTOC(fileContent, toc string) string {
	startIdx := strings.Index(fileContent, tocStartMarker)
	endIdx := strings.Index(fileContent, tocEndMarker)
	
	var newContent string
	if startIdx != -1 && endIdx != -1 && startIdx < endIdx {
		// Replace existing TOC
		newContent = fileContent[:startIdx] + toc + fileContent[endIdx+len(tocEndMarker):]
	} else {
		// Add TOC at beginning
		newContent = toc + "\n" + fileContent
	}

	return newContent
}
