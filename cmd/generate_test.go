package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateCommandWithoutTitle(t *testing.T) {
	// Create a temporary test directory
	tempDir := t.TempDir()
	
	// Create a test markdown file
	testFile := filepath.Join(tempDir, "test.md")
	content := `# First Heading
This is some content under the first heading.

## First Sub-heading
More content here.

# Second Heading
Content for the second heading.

# Third Heading
Content for the third heading.
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Reset RootCmd
	RootCmd.ResetFlags()
	RootCmd.ResetCommands()
	RootCmd.AddCommand(generateCmd)
	
	// Create a buffer to capture output
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	
	// Set args and execute command
	RootCmd.SetArgs([]string{"generate", "--file", testFile})
	
	// Reset global variables
	filePath = ""
	language = "pt" // Default value
	
	err = RootCmd.Execute()
	if err != nil {
		t.Errorf("Execute() error = %v", err)
		return
	}
	
	// Read the updated file
	updatedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}
	
	updatedContentStr := string(updatedContent)
	
	// Verificar que o TOC não contém o título "# Sumário" ou "# Summary"
	if strings.Contains(updatedContentStr, "# Sumário") || strings.Contains(updatedContentStr, "# Summary") {
		t.Error("Generated TOC should not contain title '# Sumário' or '# Summary'")
	}
	
	// Verificar que o TOC contém os marcadores corretos
	if !strings.Contains(updatedContentStr, "<!-- START_TABLE_OF_CONTENTS -->") || 
	   !strings.Contains(updatedContentStr, "<!-- END_TABLE_OF_CONTENTS -->") {
		t.Error("Generated TOC should contain start and end markers")
	}
	
	// Verificar que as entradas de TOC estão presentes
	expectedEntries := []string{
		"- [First Heading](#first-heading)",
		"  - [First Sub-heading](#first-sub-heading)",
		"- [Second Heading](#second-heading)",
		"- [Third Heading](#third-heading)",
	}

	for _, entry := range expectedEntries {
		if !strings.Contains(updatedContentStr, entry) {
			t.Errorf("Generated TOC should contain entry: %s", entry)
		}
	}
}

func TestGenerateCommandWithLanguageOption(t *testing.T) {
	// Create a temporary test directory
	tempDir := t.TempDir()
	
	// Create a test markdown file
	testFile := filepath.Join(tempDir, "test.md")
	content := `# First Heading
This is some content under the first heading.

## First Sub-heading
More content here.

# Second Heading
Content for the second heading.

# Third Heading
Content for the third heading.
`
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name        string
		args        []string
		wantErr     bool
	}{
		{
			name:      "With Portuguese language flag",
			args:      []string{"generate", "--file", testFile, "--language", "pt"},
			wantErr:   false,
		},
		{
			name:      "With English language flag",
			args:      []string{"generate", "--file", testFile, "--language", "en"},
			wantErr:   false,
		},
		{
			name:      "Default to Portuguese when no language flag",
			args:      []string{"generate", "--file", testFile},
			wantErr:   false,
		},
		{
			name:      "With uppercase English language flag",
			args:      []string{"generate", "--file", testFile, "--language", "EN"},
			wantErr:   false,
		},
		{
			name:      "With mixed case English language flag",
			args:      []string{"generate", "--file", testFile, "--language", "En"},
			wantErr:   false,
		},
		{
			name:      "With uppercase Portuguese language flag",
			args:      []string{"generate", "--file", testFile, "--language", "PT"},
			wantErr:   false,
		},
		{
			name:      "With invalid language flag (defaults to Portuguese)",
			args:      []string{"generate", "--file", testFile, "--language", "fr"},
			wantErr:   false,
		},
		{
			name:      "With whitespace in language flag",
			args:      []string{"generate", "--file", testFile, "--language", "  en  "},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset RootCmd for each test
			RootCmd.ResetFlags()
			RootCmd.ResetCommands()
			RootCmd.AddCommand(generateCmd)
			
			// Create a buffer to capture output
			buf := new(bytes.Buffer)
			RootCmd.SetOut(buf)
			RootCmd.SetErr(buf)
			
			// Set args and execute command
			RootCmd.SetArgs(tt.args)
			
			// Reset global variables
			filePath = ""
			language = "pt" // Default value
			
			err := RootCmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			// Read the updated file
			updatedContent, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("Failed to read updated file: %v", err)
			}
			
			updatedContentStr := string(updatedContent)
			
			// Verificar que o TOC não contém o título "# Sumário" ou "# Summary"
			if strings.Contains(updatedContentStr, "# Sumário") || strings.Contains(updatedContentStr, "# Summary") {
				t.Error("Generated TOC should not contain title '# Sumário' or '# Summary'")
			}
			
			// Verificar que o TOC contém os marcadores corretos
			if !strings.Contains(updatedContentStr, "<!-- START_TABLE_OF_CONTENTS -->") || 
			   !strings.Contains(updatedContentStr, "<!-- END_TABLE_OF_CONTENTS -->") {
				t.Error("Generated TOC should contain start and end markers")
			}
			
			// Verificar que as entradas de TOC estão presentes
			expectedEntries := []string{
				"- [First Heading](#first-heading)",
				"  - [First Sub-heading](#first-sub-heading)",
				"- [Second Heading](#second-heading)",
				"- [Third Heading](#third-heading)",
			}

			for _, entry := range expectedEntries {
				if !strings.Contains(updatedContentStr, entry) {
					t.Errorf("Generated TOC should contain entry: %s", entry)
				}
			}
			
			// Reset the file for the next test
			err = os.WriteFile(testFile, []byte(content), 0644)
			if err != nil {
				t.Fatalf("Failed to reset test file: %v", err)
			}
		})
	}
} 