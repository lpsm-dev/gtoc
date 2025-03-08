package generator

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

const (
	indexStartMarker = "<!-- START_GTOC -->"
	indexEndMarker   = "<!-- END_GTOC -->"
)

// Generator represents a markdown table of contents generator
type Generator struct {
	targetFile   string
	maxDepth     int
	excludePaths []string
}

// Heading represents a markdown heading
type Heading struct {
	Level       int
	Text        string
	Anchor      string
	Line        int
	SubHeadings []*Heading
}

// NewGenerator creates a new Generator instance
func NewGenerator(targetFile string, maxDepth int, excludePaths []string) *Generator {
	return &Generator{
		targetFile:   targetFile,
		maxDepth:     maxDepth,
		excludePaths: excludePaths,
	}
}

// Generate generates a markdown table of contents
func (g *Generator) Generate() (string, error) {
	// Parse the markdown file and extract headings
	headings, err := g.parseHeadings()
	if err != nil {
		return "", err
	}

	// Build the table of contents
	var sb strings.Builder
	sb.WriteString(indexStartMarker + "\n\n")
	sb.WriteString("## Table of Contents\n\n")

	// Generate TOC from headings
	g.generateTOC(&sb, headings, 0)

	sb.WriteString("\n" + indexEndMarker + "\n")
	return sb.String(), nil
}

// parseHeadings parses the markdown file and extracts headings
func (g *Generator) parseHeadings() ([]*Heading, error) {
	file, err := os.Open(g.targetFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	headingRegex := regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

	var headings []*Heading
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		matches := headingRegex.FindStringSubmatch(line)

		if len(matches) > 2 {
			level := len(matches[1])
			text := matches[2]

			// Skip headings deeper than maxDepth if specified
			if g.maxDepth > 0 && level > g.maxDepth {
				continue
			}

			heading := &Heading{
				Level:  level,
				Text:   text,
				Anchor: generateAnchor(text),
				Line:   lineNum,
			}

			headings = append(headings, heading)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return headings, nil
}

// generateTOC generates a table of contents from headings
func (g *Generator) generateTOC(sb *strings.Builder, headings []*Heading, level int) {
	for _, heading := range headings {
		// Indent based on heading level
		indent := strings.Repeat("  ", heading.Level-1)

		// Generate TOC entry
		sb.WriteString(fmt.Sprintf("%s- [%s](#%s)\n", indent, heading.Text, heading.Anchor))

		// Process subheadings if any
		if len(heading.SubHeadings) > 0 {
			g.generateTOC(sb, heading.SubHeadings, level+1)
		}
	}
}

// UpdateFile updates the target file with the generated table of contents
func (g *Generator) UpdateFile(toc string) error {
	// Read the file
	content, err := ioutil.ReadFile(g.targetFile)
	if err != nil {
		return err
	}

	// Check if the file already has index markers
	contentStr := string(content)
	startIdx := strings.Index(contentStr, indexStartMarker)
	endIdx := strings.Index(contentStr, indexEndMarker)

	var newContent string
	if startIdx != -1 && endIdx != -1 && startIdx < endIdx {
		// Replace existing index
		newContent = contentStr[:startIdx] + toc + contentStr[endIdx+len(indexEndMarker):]
	} else {
		// Add index after the first heading or at the beginning if no heading
		firstHeadingRegex := regexp.MustCompile(`(?m)^#.*\n`)
		loc := firstHeadingRegex.FindStringIndex(contentStr)

		if loc != nil {
			// Insert after the first heading and its newline
			insertPos := loc[1]
			newContent = contentStr[:insertPos] + "\n" + toc + contentStr[insertPos:]
		} else {
			// Add at the beginning
			newContent = toc + "\n" + contentStr
		}
	}

	// Write the file
	return ioutil.WriteFile(g.targetFile, []byte(newContent), 0644)
}

// generateAnchor generates a GitHub-compatible anchor from heading text
func generateAnchor(text string) string {
	// Convert to lowercase
	anchor := strings.ToLower(text)

	// Replace spaces with hyphens
	anchor = strings.ReplaceAll(anchor, " ", "-")

	// Remove any non-alphanumeric characters except hyphens
	nonAlphanumericRegex := regexp.MustCompile(`[^a-z0-9-]`)
	anchor = nonAlphanumericRegex.ReplaceAllString(anchor, "")

	// Replace multiple hyphens with a single hyphen
	multipleHyphensRegex := regexp.MustCompile(`-+`)
	anchor = multipleHyphensRegex.ReplaceAllString(anchor, "-")

	// Remove leading and trailing hyphens
	anchor = strings.Trim(anchor, "-")

	return anchor
}
