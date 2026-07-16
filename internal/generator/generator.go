package generator

import (
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

// headingPattern matches ATX-style markdown headings (# through ######).
var headingPattern = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)

// fenceOpenPattern matches the leading run of backticks or tildes that opens
// or closes a fenced code block, optionally followed by an info string.
var fenceOpenPattern = regexp.MustCompile("^(`{3,}|~{3,})")

// disallowedAnchorChars matches every rune that GitHub strips out of a
// heading before slugifying it: anything that is not a unicode letter,
// unicode number, space, hyphen, or underscore.
var disallowedAnchorChars = regexp.MustCompile(`[^\p{L}\p{N} _-]`)

// Generator handles the TOC generation for markdown files.
type Generator struct {
	targetFile      string
	maxDepth        int
	excludePatterns []string
}

// Heading represents a markdown heading discovered in the document.
type Heading struct {
	Level  int
	Text   string
	Anchor string
	Line   int
}

// NewGenerator creates a new Generator.
func NewGenerator(targetFile string, maxDepth int, excludePatterns []string) *Generator {
	return &Generator{
		targetFile:      targetFile,
		maxDepth:        maxDepth,
		excludePatterns: excludePatterns,
	}
}

// Generate creates a markdown table of contents from the target file's headings.
func (g *Generator) Generate() (string, error) {
	headings, err := g.extractHeadings()
	if err != nil {
		return "", err
	}

	minLevel := minHeadingLevel(headings)

	var sb strings.Builder
	sb.WriteString(tocStartMarker + "\n\n")

	for _, heading := range headings {
		indent := strings.Repeat("  ", heading.Level-minLevel)
		sb.WriteString(fmt.Sprintf("%s- [%s](#%s)\n", indent, heading.Text, heading.Anchor))
	}

	sb.WriteString("\n" + backToTopLink + "\n")
	sb.WriteString("\n" + tocEndMarker)
	return sb.String(), nil
}

// minHeadingLevel returns the smallest heading level present in headings, or
// 1 when there are no headings. This is used to normalize indentation so a
// document that starts at ## renders its top-level entries with no indent.
func minHeadingLevel(headings []*Heading) int {
	min := 0
	for _, h := range headings {
		if min == 0 || h.Level < min {
			min = h.Level
		}
	}
	if min == 0 {
		return 1
	}
	return min
}

// lineFilter tracks scanning state so that lines inside fenced code blocks or
// an existing TOC block are excluded from heading extraction.
type lineFilter struct {
	inCodeFence bool
	fenceMarker string
	inTOCBlock  bool
}

// skip reports whether the current line should be ignored when looking for
// headings, updating the filter's internal state as it scans.
func (f *lineFilter) skip(line string) bool {
	if f.inTOCBlock {
		if strings.Contains(line, tocEndMarker) {
			f.inTOCBlock = false
		}
		return true
	}
	if strings.Contains(line, tocStartMarker) {
		f.inTOCBlock = true
		return true
	}

	if marker := fenceMarker(line); marker != "" {
		f.toggleFence(marker)
		return true
	}
	return f.inCodeFence
}

// toggleFence opens or closes a fenced code block based on the marker found
// on the current line.
func (f *lineFilter) toggleFence(marker string) {
	switch {
	case f.inCodeFence && marker == f.fenceMarker:
		f.inCodeFence = false
		f.fenceMarker = ""
	case !f.inCodeFence:
		f.inCodeFence = true
		f.fenceMarker = marker
	}
}

// fenceMarker returns "`" or "~" if line opens or closes a fenced code block,
// or "" if it does not.
func fenceMarker(line string) string {
	trimmed := strings.TrimLeft(line, " \t")
	match := fenceOpenPattern.FindString(trimmed)
	if match == "" {
		return ""
	}
	return match[:1]
}

// extractHeadings parses the target markdown file and returns its headings,
// skipping fenced code blocks and any existing TOC block.
func (g *Generator) extractHeadings() ([]*Heading, error) {
	content, err := os.ReadFile(g.targetFile)
	if err != nil {
		return nil, err
	}

	headings := []*Heading{}
	anchorCounts := map[string]int{}
	filter := &lineFilter{}

	for i, line := range strings.Split(string(content), "\n") {
		if filter.skip(line) {
			continue
		}

		if heading := g.parseHeadingLine(line, i+1, anchorCounts); heading != nil {
			headings = append(headings, heading)
		}
	}

	return headings, nil
}

// parseHeadingLine attempts to parse a single line as a markdown heading,
// applying depth filtering, exclusion patterns, and anchor deduplication.
// It returns nil when the line is not a qualifying heading.
func (g *Generator) parseHeadingLine(line string, lineNum int, anchorCounts map[string]int) *Heading {
	matches := headingPattern.FindStringSubmatch(line)
	if len(matches) <= 2 {
		return nil
	}

	level := len(matches[1])
	if g.maxDepth > 0 && level > g.maxDepth {
		return nil
	}

	text := strings.TrimSpace(matches[2])
	if g.isExcluded(text) {
		return nil
	}

	return &Heading{
		Level:  level,
		Text:   text,
		Anchor: uniqueAnchor(createAnchor(text), anchorCounts),
		Line:   lineNum,
	}
}

// isExcluded reports whether heading text matches any exclude pattern via a
// case-insensitive substring match. A nil or empty pattern list excludes nothing.
func (g *Generator) isExcluded(text string) bool {
	lowerText := strings.ToLower(text)
	for _, pattern := range g.excludePatterns {
		if pattern == "" {
			continue
		}
		if strings.Contains(lowerText, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// uniqueAnchor returns slug, or slug with a "-1", "-2", ... suffix if it has
// already been used earlier in the same Generate() call, matching GitHub's
// duplicate-heading anchor behavior. counts is mutated in place.
func uniqueAnchor(slug string, counts map[string]int) string {
	count := counts[slug]
	counts[slug] = count + 1
	if count == 0 {
		return slug
	}
	return fmt.Sprintf("%s-%d", slug, count)
}

// createAnchor generates a GitHub-compatible anchor slug from heading text,
// matching the behavior of github-slugger: lowercase, strip everything that
// isn't a unicode letter, unicode number, space, hyphen, or underscore, then
// turn each space into a hyphen. Consecutive hyphens are not collapsed and
// leading/trailing hyphens are not trimmed, since GitHub does neither.
func createAnchor(text string) string {
	trimmed := strings.TrimSpace(text)
	lowered := strings.ToLower(trimmed)
	stripped := disallowedAnchorChars.ReplaceAllString(lowered, "")
	return strings.ReplaceAll(stripped, " ", "-")
}

// UpdateFile writes the file with the TOC replaced or prepended, preserving
// the original file's permission bits.
func (g *Generator) UpdateFile(toc string) error {
	content, err := os.ReadFile(g.targetFile)
	if err != nil {
		return err
	}

	mode := os.FileMode(0644)
	if info, statErr := os.Stat(g.targetFile); statErr == nil {
		mode = info.Mode().Perm()
	}

	newContent := g.GetFileWithUpdatedTOC(string(content), toc)
	return os.WriteFile(g.targetFile, []byte(newContent), mode)
}

// GetFileWithUpdatedTOC returns the file content with the TOC block replaced
// in place, or the TOC prepended when no existing block is found. It does
// not write to disk, which makes it useful for dry-run previews.
func (g *Generator) GetFileWithUpdatedTOC(fileContent, toc string) string {
	startIdx := strings.Index(fileContent, tocStartMarker)
	endIdx := strings.Index(fileContent, tocEndMarker)

	if startIdx != -1 && endIdx != -1 && startIdx < endIdx {
		return fileContent[:startIdx] + toc + fileContent[endIdx+len(tocEndMarker):]
	}
	return toc + "\n" + fileContent
}
