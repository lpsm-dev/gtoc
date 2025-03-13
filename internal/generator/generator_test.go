package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateTOCWithoutTitle(t *testing.T) {
	// Create a temporary test file
	tempFile := filepath.Join(t.TempDir(), "test.md")
	content := `# First Heading
This is some content under the first heading.

## First Sub-heading
More content here.

# Second Heading
Content for the second heading.

# Third Heading
Content for the third heading.
`
	err := os.WriteFile(tempFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	gen := NewGenerator(tempFile, 0, nil, "pt")
	toc, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verificar que o TOC não contém o título "# Sumário" ou "# Summary"
	if strings.Contains(toc, "# Sumário") || strings.Contains(toc, "# Summary") {
		t.Error("TOC should not contain title '# Sumário' or '# Summary'")
	}

	// Verificar que o TOC contém os marcadores corretos
	if !strings.Contains(toc, tocStartMarker) || !strings.Contains(toc, tocEndMarker) {
		t.Error("TOC should contain start and end markers")
	}

	// Verificar que as entradas de TOC estão presentes
	expectedEntries := []string{
		"- [First Heading](#first-heading)",
		"  - [First Sub-heading](#first-sub-heading)",
		"- [Second Heading](#second-heading)",
		"- [Third Heading](#third-heading)",
	}

	for _, entry := range expectedEntries {
		if !strings.Contains(toc, entry) {
			t.Errorf("TOC should contain entry: %s", entry)
		}
	}
}

func TestGenerateTOCWithLanguageOption(t *testing.T) {
	// Create a temporary test file
	tempFile := filepath.Join(t.TempDir(), "test.md")
	content := `# First Heading
This is some content under the first heading.

## First Sub-heading
More content here.

# Second Heading
Content for the second heading.

# Third Heading
Content for the third heading.
`
	err := os.WriteFile(tempFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		language string
	}{
		{
			name:     "With Portuguese language option",
			language: "pt",
		},
		{
			name:     "With English language option",
			language: "en",
		},
		{
			name:     "Default to Portuguese when empty",
			language: "",
		},
		{
			name:     "Default to Portuguese when invalid",
			language: "fr",
		},
		{
			name:     "Case insensitive - uppercase EN",
			language: "EN",
		},
		{
			name:     "Case insensitive - mixed case En",
			language: "En",
		},
		{
			name:     "Case insensitive - uppercase PT",
			language: "PT",
		},
		{
			name:     "Case insensitive - mixed case Pt",
			language: "Pt",
		},
		{
			name:     "Whitespace - language with spaces",
			language: "  en  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator(tempFile, 0, nil, tt.language)
			toc, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			// Verificar que o TOC não contém o título "# Sumário" ou "# Summary"
			if strings.Contains(toc, "# Sumário") || strings.Contains(toc, "# Summary") {
				t.Error("TOC should not contain title '# Sumário' or '# Summary'")
			}

			// Verificar que o TOC contém os marcadores corretos
			if !strings.Contains(toc, tocStartMarker) || !strings.Contains(toc, tocEndMarker) {
				t.Error("TOC should contain start and end markers")
			}

			// Verificar que as entradas de TOC estão presentes
			expectedEntries := []string{
				"- [First Heading](#first-heading)",
				"  - [First Sub-heading](#first-sub-heading)",
				"- [Second Heading](#second-heading)",
				"- [Third Heading](#third-heading)",
			}

			for _, entry := range expectedEntries {
				if !strings.Contains(toc, entry) {
					t.Errorf("TOC should contain entry: %s", entry)
				}
			}
		})
	}
}

func TestGenerateTOCEntries(t *testing.T) {
	// Create a temporary test file
	tempFile := filepath.Join(t.TempDir(), "test.md")
	content := `# First Heading
This is some content under the first heading.

## First Sub-heading
More content here.

# Second Heading
Content for the second heading.

# Third Heading
Content for the third heading.
`
	err := os.WriteFile(tempFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test that entries are correctly generated regardless of language
	gen := NewGenerator(tempFile, 0, nil, "en")
	toc, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check for expected entries
	expectedEntries := []string{
		"- [First Heading](#first-heading)",
		"  - [First Sub-heading](#first-sub-heading)",
		"- [Second Heading](#second-heading)",
		"- [Third Heading](#third-heading)",
	}

	for _, entry := range expectedEntries {
		if !strings.Contains(toc, entry) {
			t.Errorf("Generated TOC doesn't contain expected entry: %s", entry)
		}
	}
}

func TestGenerateAnchor(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "Simple text",
			text:     "Simple Heading",
			expected: "simple-heading",
		},
		{
			name:     "Text with special characters",
			text:     "Heading with $pecial Ch@racters!!!",
			expected: "heading-with-pecial-chracters",
		},
		{
			name:     "Text with multiple spaces",
			text:     "Heading   with   multiple   spaces",
			expected: "heading-with-multiple-spaces",
		},
		{
			name:     "Text with punctuation",
			text:     "Heading, with: punctuation; marks!",
			expected: "heading-with-punctuation-marks",
		},
		{
			name:     "Text with leading and trailing spaces",
			text:     "  Heading with spaces  ",
			expected: "heading-with-spaces",
		},
		{
			name:     "Text with numbers",
			text:     "Heading 123",
			expected: "heading-123",
		},
		{
			name:     "Text with hyphens",
			text:     "Heading-with-hyphens",
			expected: "heading-with-hyphens",
		},
		{
			name:     "Text with multiple hyphens",
			text:     "Heading--with--multiple--hyphens",
			expected: "heading-with-multiple-hyphens",
		},
		{
			name:     "Text with leading and trailing hyphens",
			text:     "-Heading with hyphens-",
			expected: "heading-with-hyphens",
		},
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