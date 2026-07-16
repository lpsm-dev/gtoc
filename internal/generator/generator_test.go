package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTempFile creates a markdown file with the given content in a fresh
// temp directory and returns its path.
func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	return path
}

func TestGenerateSimpleDocument(t *testing.T) {
	content := `# First Heading
This is some content under the first heading.

## First Sub-heading
More content here.

# Second Heading
Content for the second heading.

# Third Heading
Content for the third heading.
`
	path := writeTempFile(t, content)

	gen := NewGenerator(path, 0, nil)
	toc, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if strings.Contains(toc, "# Summary") || strings.Contains(toc, "# Sumário") {
		t.Error("TOC should not contain a '# Summary' or '# Sumário' title")
	}
	if !strings.HasPrefix(toc, tocStartMarker) {
		t.Error("TOC should start with the start marker")
	}
	if !strings.HasSuffix(toc, tocEndMarker) {
		t.Error("TOC should end with the end marker")
	}

	expected := tocStartMarker + "\n\n" +
		"- [First Heading](#first-heading)\n" +
		"  - [First Sub-heading](#first-sub-heading)\n" +
		"- [Second Heading](#second-heading)\n" +
		"- [Third Heading](#third-heading)\n" +
		"\n" + backToTopLink + "\n" +
		"\n" + tocEndMarker

	if toc != expected {
		t.Errorf("Generate() = %q, want %q", toc, expected)
	}
}

func TestCreateAnchor(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{"simple heading", "Simple Heading", "simple-heading"},
		{"accented portuguese - visao geral", "Visão Geral", "visão-geral"},
		{"accented portuguese - instalacao", "Instalação", "instalação"},
		{"special characters", "Heading with $pecial Ch@racters!!!", "heading-with-pecial-chracters"},
		{"multiple spaces not collapsed", "Heading   with   multiple   spaces", "heading---with---multiple---spaces"},
		{"leading and trailing hyphens kept", "-Heading with hyphens-", "-heading-with-hyphens-"},
		{"underscores kept", "Heading_with_underscores", "heading_with_underscores"},
		{"numbers kept", "Heading 123", "heading-123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createAnchor(tt.text)
			if result != tt.expected {
				t.Errorf("createAnchor(%q) = %q, want %q", tt.text, result, tt.expected)
			}
		})
	}
}

func TestGenerateDuplicateHeadings(t *testing.T) {
	content := `# Setup
First setup section.

# Setup
Second setup section.
`
	path := writeTempFile(t, content)

	gen := NewGenerator(path, 0, nil)
	toc, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expectedEntries := []string{
		"- [Setup](#setup)\n",
		"- [Setup](#setup-1)\n",
	}
	for _, entry := range expectedEntries {
		if !strings.Contains(toc, entry) {
			t.Errorf("TOC should contain entry %q, got: %s", entry, toc)
		}
	}
}

func TestExtractHeadingsSkipsCodeFences(t *testing.T) {
	content := "# Heading One\n\n" +
		"```\n" +
		"# Not A Heading\n" +
		"```\n\n" +
		"## Heading Two\n\n" +
		"~~~\n" +
		"### Also Not A Heading\n" +
		"~~~\n\n" +
		"# Heading Three\n"
	path := writeTempFile(t, content)

	gen := NewGenerator(path, 0, nil)
	toc, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if strings.Contains(toc, "Not A Heading") {
		t.Error("TOC should not contain headings found inside ``` fenced code blocks")
	}
	if strings.Contains(toc, "Also Not A Heading") {
		t.Error("TOC should not contain headings found inside ~~~ fenced code blocks")
	}

	expectedEntries := []string{
		"- [Heading One](#heading-one)",
		"- [Heading Two](#heading-two)",
		"- [Heading Three](#heading-three)",
	}
	for _, entry := range expectedEntries {
		if !strings.Contains(toc, entry) {
			t.Errorf("TOC should contain entry: %s", entry)
		}
	}
}

func TestExtractHeadingsSkipsFencedCodeWithInfoString(t *testing.T) {
	content := "# Real Heading\n\n" +
		"```go\n" +
		"# Also Not A Heading\n" +
		"```\n"
	path := writeTempFile(t, content)

	gen := NewGenerator(path, 0, nil)
	toc, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if strings.Contains(toc, "Also Not A Heading") {
		t.Error("TOC should not contain headings inside a fence with an info string")
	}
	if !strings.Contains(toc, "- [Real Heading](#real-heading)") {
		t.Error("TOC should contain the real heading")
	}
}

func TestGenerateAndUpdateFileIdempotent(t *testing.T) {
	original := tocStartMarker + "\n\n" +
		"- [Old Entry](#old-entry)\n" +
		"# Fake Heading Inside TOC\n" +
		"\n" + backToTopLink + "\n" +
		"\n" + tocEndMarker + "\n\n" +
		"# Real Heading\n" +
		"Some content.\n"
	path := writeTempFile(t, original)

	toc1, contentAfterFirst := generateAndUpdate(t, path)

	if strings.Contains(toc1, "Fake Heading Inside TOC") {
		t.Error("TOC must not include headings found inside an existing TOC block")
	}
	if strings.Contains(toc1, "Old Entry") {
		t.Error("TOC must not include entries found inside an existing TOC block")
	}
	if !strings.Contains(toc1, "- [Real Heading](#real-heading)") {
		t.Error("TOC should contain the real heading outside the TOC block")
	}

	toc2, contentAfterSecond := generateAndUpdate(t, path)
	if toc1 != toc2 {
		t.Errorf("regenerating TOC should be idempotent: first = %q, second = %q", toc1, toc2)
	}
	if contentAfterFirst != contentAfterSecond {
		t.Error("running Generate+UpdateFile twice should yield the same file content as running it once")
	}
}

// generateAndUpdate runs Generate and UpdateFile on path and returns the
// generated TOC along with the resulting file content.
func generateAndUpdate(t *testing.T, path string) (string, string) {
	t.Helper()
	gen := NewGenerator(path, 0, nil)
	toc, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if err := gen.UpdateFile(toc); err != nil {
		t.Fatalf("UpdateFile failed: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file after update: %v", err)
	}
	return toc, string(content)
}

func TestGenerateMaxDepthFiltering(t *testing.T) {
	content := `# Level One

## Level Two

### Level Three
`
	path := writeTempFile(t, content)

	gen := NewGenerator(path, 2, nil)
	toc, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(toc, "- [Level One](#level-one)") {
		t.Error("TOC should contain level 1 heading")
	}
	if !strings.Contains(toc, "- [Level Two](#level-two)") {
		t.Error("TOC should contain level 2 heading")
	}
	if strings.Contains(toc, "Level Three") {
		t.Error("TOC should not contain level 3 heading when maxDepth is 2")
	}
}

func TestGenerateExcludePatterns(t *testing.T) {
	content := `# Public Section

# Internal Notes

# Another Public Section
`
	path := writeTempFile(t, content)

	gen := NewGenerator(path, 0, []string{"internal"})
	toc, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if strings.Contains(toc, "Internal Notes") {
		t.Error("TOC should exclude headings matching an exclude pattern, case-insensitively")
	}
	if !strings.Contains(toc, "- [Public Section](#public-section)") {
		t.Error("TOC should contain non-matching headings")
	}
	if !strings.Contains(toc, "- [Another Public Section](#another-public-section)") {
		t.Error("TOC should contain non-matching headings")
	}
}

func TestGenerateIndentNormalization(t *testing.T) {
	content := `## Section One

### Subsection

## Section Two
`
	path := writeTempFile(t, content)

	gen := NewGenerator(path, 0, nil)
	toc, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expectedEntries := []string{
		"- [Section One](#section-one)\n",
		"  - [Subsection](#subsection)\n",
		"- [Section Two](#section-two)\n",
	}
	for _, entry := range expectedEntries {
		if !strings.Contains(toc, entry) {
			t.Errorf("TOC should contain entry %q when document starts at level 2, got: %s", entry, toc)
		}
	}
	if strings.Contains(toc, "    - [Subsection]") {
		t.Error("indentation should be normalized relative to the document's minimum heading level")
	}
}

func TestUpdateFileReplacesExistingTOC(t *testing.T) {
	original := "# Title\n\n" +
		tocStartMarker + "\n\n" +
		"- [Old Entry](#old-entry)\n" +
		"\n" + tocEndMarker + "\n\n" +
		"## Body\nSome text.\n"
	path := writeTempFile(t, original)

	newTOC := tocStartMarker + "\n\n- [New Entry](#new-entry)\n\n" + tocEndMarker

	gen := NewGenerator(path, 0, nil)
	if err := gen.UpdateFile(newTOC); err != nil {
		t.Fatalf("UpdateFile failed: %v", err)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}

	got := string(updated)
	if strings.Contains(got, "Old Entry") {
		t.Error("UpdateFile should remove the old TOC block content")
	}
	if !strings.Contains(got, "New Entry") {
		t.Error("UpdateFile should insert the new TOC content")
	}
	if !strings.Contains(got, "# Title") || !strings.Contains(got, "## Body") {
		t.Error("UpdateFile should preserve content surrounding the TOC block")
	}
}

func TestUpdateFilePrependsWhenNoExistingTOC(t *testing.T) {
	original := "# Title\n\nSome content without a TOC.\n"
	path := writeTempFile(t, original)

	newTOC := tocStartMarker + "\n\n- [Title](#title)\n\n" + tocEndMarker

	gen := NewGenerator(path, 0, nil)
	if err := gen.UpdateFile(newTOC); err != nil {
		t.Fatalf("UpdateFile failed: %v", err)
	}

	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}

	expected := newTOC + "\n" + original
	if string(updated) != expected {
		t.Errorf("UpdateFile() content = %q, want %q", string(updated), expected)
	}
}

func TestUpdateFilePreservesPermissions(t *testing.T) {
	path := writeTempFile(t, "# Title\n\nSome content.\n")
	if err := os.Chmod(path, 0600); err != nil {
		t.Fatalf("failed to chmod test file: %v", err)
	}

	gen := NewGenerator(path, 0, nil)
	toc := tocStartMarker + "\n\n- [Title](#title)\n\n" + tocEndMarker
	if err := gen.UpdateFile(toc); err != nil {
		t.Fatalf("UpdateFile failed: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("UpdateFile should preserve original file permissions, got %o, want %o", perm, 0600)
	}
}

func TestGetFileWithUpdatedTOC(t *testing.T) {
	gen := NewGenerator("unused.md", 0, nil)

	t.Run("replaces existing block", func(t *testing.T) {
		original := "before\n" + tocStartMarker + "\nold\n" + tocEndMarker + "\nafter"
		newTOC := tocStartMarker + "\nnew\n" + tocEndMarker
		result := gen.GetFileWithUpdatedTOC(original, newTOC)
		expected := "before\n" + newTOC + "\nafter"
		if result != expected {
			t.Errorf("GetFileWithUpdatedTOC() = %q, want %q", result, expected)
		}
	})

	t.Run("prepends when no block present", func(t *testing.T) {
		original := "just content\n"
		newTOC := tocStartMarker + "\nnew\n" + tocEndMarker
		result := gen.GetFileWithUpdatedTOC(original, newTOC)
		expected := newTOC + "\n" + original
		if result != expected {
			t.Errorf("GetFileWithUpdatedTOC() = %q, want %q", result, expected)
		}
	})
}
