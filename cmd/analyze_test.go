package cmd

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// analyzeTestCase describes one analyze command scenario: the README
// content before running analyze, and what the result must look like.
type analyzeTestCase struct {
	name            string
	initialContent  string
	expectedMarkers []string
	expectedLinks   int // Number of "back to top" links expected.
	shouldContain   []string
}

// analyzeTestCases returns the table of scenarios exercised by
// TestAnalyzeCommand.
func analyzeTestCases() []analyzeTestCase {
	return []analyzeTestCase{
		{
			name:            "Empty README",
			initialContent:  "",
			expectedMarkers: []string{"<!-- BEGIN_DOCS -->", "<!-- END_DOCS -->", "<a name=\"readme-top\"></a>"},
			expectedLinks:   0,
			shouldContain:   []string{"<!-- BEGIN_DOCS -->", "<!-- END_DOCS -->", "<a name=\"readme-top\"></a>"},
		},
		{
			name:            "README with no headings",
			initialContent:  "This is a simple README file with no headings.",
			expectedMarkers: []string{"<!-- BEGIN_DOCS -->", "<!-- END_DOCS -->", "<a name=\"readme-top\"></a>"},
			expectedLinks:   0,
			shouldContain:   []string{"<!-- BEGIN_DOCS -->", "<!-- END_DOCS -->", "<a name=\"readme-top\"></a>", "This is a simple README file with no headings."},
		},
		{
			name:            "README with one heading",
			initialContent:  "# Heading 1\nThis is content under heading 1.",
			expectedMarkers: []string{"<!-- BEGIN_DOCS -->", "<!-- END_DOCS -->", "<a name=\"readme-top\"></a>"},
			expectedLinks:   1,
			shouldContain:   []string{"<!-- BEGIN_DOCS -->", "<!-- END_DOCS -->", "<a name=\"readme-top\"></a>", "# Heading 1", "<p align=\"right\">(<a href=\"#readme-top\">back to top</a>)</p>"},
		},
		{
			name:            "README with multiple headings",
			initialContent:  "# Heading 1\nContent 1\n\n# Heading 2\nContent 2\n\n# Heading 3\nContent 3",
			expectedMarkers: []string{"<!-- BEGIN_DOCS -->", "<!-- END_DOCS -->", "<a name=\"readme-top\"></a>"},
			expectedLinks:   3,
			shouldContain:   []string{"<!-- BEGIN_DOCS -->", "<!-- END_DOCS -->", "<a name=\"readme-top\"></a>", "# Heading 1", "# Heading 2", "# Heading 3"},
		},
		{
			name:            "README with existing BEGIN_DOCS marker",
			initialContent:  "<!-- BEGIN_DOCS -->\n<a name=\"readme-top\"></a>\n# Heading 1\nContent 1\n\n# Heading 2\nContent 2",
			expectedMarkers: []string{"<!-- BEGIN_DOCS -->", "<!-- END_DOCS -->", "<a name=\"readme-top\"></a>"},
			expectedLinks:   2,
			shouldContain:   []string{"<!-- BEGIN_DOCS -->", "<!-- END_DOCS -->", "<a name=\"readme-top\"></a>", "# Heading 1", "# Heading 2"},
		},
		{
			name:            "README with existing back to top links",
			initialContent:  "# Heading 1\nContent 1\n<p align=\"right\">(<a href=\"#readme-top\">back to top</a>)</p>\n\n# Heading 2\nContent 2",
			expectedMarkers: []string{"<!-- BEGIN_DOCS -->", "<!-- END_DOCS -->", "<a name=\"readme-top\"></a>"},
			expectedLinks:   2,
			shouldContain:   []string{"<!-- BEGIN_DOCS -->", "<!-- END_DOCS -->", "<a name=\"readme-top\"></a>", "# Heading 1", "# Heading 2"},
		},
		{
			name:            "README with subheadings",
			initialContent:  "# Heading 1\nContent 1\n\n## Subheading 1.1\nSubcontent 1.1\n\n# Heading 2\nContent 2\n\n## Subheading 2.1\nSubcontent 2.1",
			expectedMarkers: []string{"<!-- BEGIN_DOCS -->", "<!-- END_DOCS -->", "<a name=\"readme-top\"></a>"},
			expectedLinks:   2,
			shouldContain:   []string{"<!-- BEGIN_DOCS -->", "<!-- END_DOCS -->", "<a name=\"readme-top\"></a>", "# Heading 1", "# Heading 2", "## Subheading 1.1", "## Subheading 2.1"},
		},
		{
			name:            "README with existing markers and links",
			initialContent:  "<!-- BEGIN_DOCS -->\n<a name=\"readme-top\"></a>\n\n# Heading 1\nContent 1\n<p align=\"right\">(<a href=\"#readme-top\">back to top</a>)</p>\n\n# Heading 2\nContent 2\n<p align=\"right\">(<a href=\"#readme-top\">back to top</a>)</p>\n\n<!-- END_DOCS -->",
			expectedMarkers: []string{"<!-- BEGIN_DOCS -->", "<!-- END_DOCS -->", "<a name=\"readme-top\"></a>"},
			expectedLinks:   2,
			shouldContain:   []string{"<!-- BEGIN_DOCS -->", "<!-- END_DOCS -->", "<a name=\"readme-top\"></a>", "# Heading 1", "# Heading 2"},
		},
		{
			// Regression test for the data-loss bug where addBackToTopLinks
			// cut the content at the END_DOCS marker and never re-appended
			// what followed it.
			name:            "README preserves content after END_DOCS marker",
			initialContent:  "<!-- BEGIN_DOCS -->\n<a name=\"readme-top\"></a>\n\n# Heading 1\nContent 1\n<p align=\"right\">(<a href=\"#readme-top\">back to top</a>)</p>\n\n<!-- END_DOCS -->\n\n## Footer\nThis footer must survive analyze.\n",
			expectedMarkers: []string{"<!-- BEGIN_DOCS -->", "<!-- END_DOCS -->", "<a name=\"readme-top\"></a>"},
			expectedLinks:   1,
			shouldContain:   []string{"## Footer", "This footer must survive analyze."},
		},
		{
			// Regression test: a literal "# " inside a fenced code block
			// must not be treated as a heading.
			name:            "README with heading-like text inside a code fence",
			initialContent:  "# Heading 1\nContent 1\n\n```\n# Not a heading\n```\n",
			expectedMarkers: []string{"<!-- BEGIN_DOCS -->", "<!-- END_DOCS -->", "<a name=\"readme-top\"></a>"},
			expectedLinks:   1,
			shouldContain:   []string{"# Not a heading"},
		},
	}
}

func TestAnalyzeCommand(t *testing.T) {
	tempDir := t.TempDir()

	for _, tt := range analyzeTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			runAnalyzeTestCase(t, tempDir, tt)
		})
	}
}

// runAnalyzeTestCase writes tt's initial content to a scratch file, runs the
// analyze command against it, and checks the result against tt's
// expectations.
func runAnalyzeTestCase(t *testing.T, tempDir string, tt analyzeTestCase) {
	t.Helper()

	testFile := filepath.Join(tempDir, "test_readme.md")
	if err := os.WriteFile(testFile, []byte(tt.initialContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	resetRootCmd()
	readmePath = ""
	RootCmd.SetArgs([]string{"analyze", "--file", testFile})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("analyze command failed: %v", err)
	}

	updated, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}
	content := string(updated)

	for _, marker := range tt.expectedMarkers {
		if !strings.Contains(content, marker) {
			t.Errorf("updated content doesn't contain expected marker %q", marker)
		}
	}

	backToTopRegex := regexp.MustCompile(`<p align="right">\(<a href="#readme-top">back to top</a>\)</p>`)
	matches := backToTopRegex.FindAllStringIndex(content, -1)
	if len(matches) != tt.expectedLinks {
		t.Errorf("expected %d 'back to top' links, found %d", tt.expectedLinks, len(matches))
	}

	for _, s := range tt.shouldContain {
		if !strings.Contains(content, s) {
			t.Errorf("updated content doesn't contain expected string %q", s)
		}
	}
}

func TestAnalyzeCommandErrors(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Non-existent file", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "non_existent.md")

		resetRootCmd()
		readmePath = ""
		RootCmd.SetArgs([]string{"analyze", "--file", nonExistentFile})

		if err := RootCmd.Execute(); err == nil {
			t.Error("expected error for non-existent file, got nil")
		}
	})
}
