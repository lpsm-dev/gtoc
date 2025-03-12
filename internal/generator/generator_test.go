package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
		wantTOC  string
	}{
		{
			name:     "Portuguese title",
			language: "pt",
			wantTOC:  "# Sumário",
		},
		{
			name:     "English title",
			language: "en",
			wantTOC:  "# Summary",
		},
		{
			name:     "Default to Portuguese when empty",
			language: "",
			wantTOC:  "# Sumário",
		},
		{
			name:     "Default to Portuguese when invalid",
			language: "fr",
			wantTOC:  "# Sumário",
		},
		{
			name:     "Case insensitive - uppercase EN",
			language: "EN",
			wantTOC:  "# Summary",
		},
		{
			name:     "Case insensitive - mixed case En",
			language: "En",
			wantTOC:  "# Summary",
		},
		{
			name:     "Case insensitive - uppercase PT",
			language: "PT",
			wantTOC:  "# Sumário",
		},
		{
			name:     "Case insensitive - mixed case Pt",
			language: "Pt",
			wantTOC:  "# Sumário",
		},
		{
			name:     "Whitespace - language with spaces",
			language: "  en  ",
			wantTOC:  "# Summary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator(tempFile, 0, nil, tt.language)
			toc, err := gen.Generate()
			if err != nil {
				t.Fatalf("Generate failed: %v", err)
			}

			// Check if the generated TOC contains the expected title
			if !strings.Contains(toc, tt.wantTOC) {
				t.Errorf("Generate() = %v, want to contain %v", toc, tt.wantTOC)
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