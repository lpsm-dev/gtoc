package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
		wantTitle   string
		wantErr     bool
	}{
		{
			name:      "With Portuguese language flag",
			args:      []string{"generate", "--file", testFile, "--language", "pt"},
			wantTitle: "# Sumário",
			wantErr:   false,
		},
		{
			name:      "With English language flag",
			args:      []string{"generate", "--file", testFile, "--language", "en"},
			wantTitle: "# Summary",
			wantErr:   false,
		},
		{
			name:      "Default to Portuguese when no language flag",
			args:      []string{"generate", "--file", testFile},
			wantTitle: "# Sumário",
			wantErr:   false,
		},
		{
			name:      "With uppercase English language flag",
			args:      []string{"generate", "--file", testFile, "--language", "EN"},
			wantTitle: "# Summary",
			wantErr:   false,
		},
		{
			name:      "With mixed case English language flag",
			args:      []string{"generate", "--file", testFile, "--language", "En"},
			wantTitle: "# Summary",
			wantErr:   false,
		},
		{
			name:      "With uppercase Portuguese language flag",
			args:      []string{"generate", "--file", testFile, "--language", "PT"},
			wantTitle: "# Sumário",
			wantErr:   false,
		},
		{
			name:      "With invalid language flag (defaults to Portuguese)",
			args:      []string{"generate", "--file", testFile, "--language", "fr"},
			wantTitle: "# Sumário",
			wantErr:   false,
		},
		{
			name:      "With whitespace in language flag",
			args:      []string{"generate", "--file", testFile, "--language", "  en  "},
			wantTitle: "# Summary",
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
			
			// Check if the updated file contains the expected title
			if !strings.Contains(string(updatedContent), tt.wantTitle) {
				t.Errorf("Updated file doesn't contain expected title %q", tt.wantTitle)
			}
			
			// Reset the file for the next test
			err = os.WriteFile(testFile, []byte(content), 0644)
			if err != nil {
				t.Fatalf("Failed to reset test file: %v", err)
			}
		})
	}
} 