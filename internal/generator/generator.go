package generator

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lpsm-dev/mdtoc/internal/git"
)

const (
	indexStartMarker = "<!-- START_MDTOC -->"
	indexEndMarker   = "<!-- END_MDTOC -->"
)

// Generator represents a markdown index generator
type Generator struct {
	repoRoot     string
	targetFile   string
	maxDepth     int
	pattern      string
	excludePaths []string
}

// NewGenerator creates a new Generator instance
func NewGenerator(repoRoot, targetFile string, maxDepth int, pattern string, excludePaths []string) *Generator {
	return &Generator{
		repoRoot:     repoRoot,
		targetFile:   targetFile,
		maxDepth:     maxDepth,
		pattern:      pattern,
		excludePaths: excludePaths,
	}
}

// Generate generates a markdown index
func (g *Generator) Generate() (string, error) {
	// Get all markdown files
	files, err := git.ListMarkdownFiles(g.repoRoot, g.pattern, g.excludePaths)
	if err != nil {
		return "", err
	}

	// Filter out the target file
	var filteredFiles []string
	for _, file := range files {
		if file != g.targetFile {
			filteredFiles = append(filteredFiles, file)
		}
	}

	// Build the index
	var sb strings.Builder
	sb.WriteString(indexStartMarker + "\n\n")
	sb.WriteString("# Documentation Index\n\n")

	// Group files by directory
	filesByDir := make(map[string][]string)
	for _, file := range filteredFiles {
		relPath, err := filepath.Rel(g.repoRoot, file)
		if err != nil {
			return "", err
		}

		dir := filepath.Dir(relPath)
		if dir == "." {
			dir = ""
		}

		// Check depth
		if g.maxDepth > 0 {
			depth := len(strings.Split(dir, string(os.PathSeparator)))
			if depth > g.maxDepth {
				continue
			}
		}

		filesByDir[dir] = append(filesByDir[dir], file)
	}

	// Sort directories
	dirs := make([]string, 0, len(filesByDir))
	for dir := range filesByDir {
		dirs = append(dirs, dir)
	}
	// Sort dirs here if needed

	// Generate index
	for _, dir := range dirs {
		if dir != "" {
			sb.WriteString(fmt.Sprintf("## %s\n\n", dir))
		}

		for _, file := range filesByDir[dir] {
			relPath, _ := filepath.Rel(g.repoRoot, file)
			title, err := extractTitle(file)
			if err != nil {
				return "", err
			}
			if title == "" {
				title = filepath.Base(file)
			}

			sb.WriteString(fmt.Sprintf("- [%s](%s)\n", title, relPath))
		}

		sb.WriteString("\n")
	}

	sb.WriteString(indexEndMarker + "\n")
	return sb.String(), nil
}

// UpdateFile updates the target file with the generated index
func (g *Generator) UpdateFile(index string) error {
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
		newContent = contentStr[:startIdx] + index + contentStr[endIdx+len(indexEndMarker):]
	} else {
		// Add index at the beginning of the file
		newContent = index + "\n" + contentStr
	}

	// Write the file
	return ioutil.WriteFile(g.targetFile, []byte(newContent), 0644)
}

// extractTitle extracts the title (H1) from a markdown file
func extractTitle(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	h1Regex := regexp.MustCompile(`^#\s+(.+)$`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := h1Regex.FindStringSubmatch(line)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}

	return "", scanner.Err()
}
