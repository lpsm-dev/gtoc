package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const generateTestContent = `# First Heading
This is some content under the first heading.

## First Sub-heading
More content here.

# Second Heading
Content for the second heading.

# Third Heading
Content for the third heading.
`

// setupGenerateTest resets RootCmd and the generate command's package-level
// flag variables to a known state before each test.
func setupGenerateTest() {
	resetRootCmd()
	filePath = ""
	depth = 0
	excludePaths = ""
	dryRun = false
	prettyOutput = false
}

func TestGenerateCommandUpdatesFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.md")
	if err := os.WriteFile(testFile, []byte(generateTestContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	setupGenerateTest()
	RootCmd.SetArgs([]string{"generate", "--file", testFile})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	updated, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}
	content := string(updated)

	// The generator no longer emits a title heading, only the markers and
	// entries.
	if strings.Contains(content, "# Sumário") || strings.Contains(content, "# Summary") {
		t.Error("generated TOC should not contain a title heading")
	}

	if !strings.Contains(content, "<!-- START_TABLE_OF_CONTENTS -->") ||
		!strings.Contains(content, "<!-- END_TABLE_OF_CONTENTS -->") {
		t.Error("generated TOC should contain start and end markers")
	}

	expectedEntries := []string{
		"1\\. [First Heading](#first-heading)",
		"&nbsp;&nbsp;&nbsp;1\\.1. [First Sub-heading](#first-sub-heading)",
		"2\\. [Second Heading](#second-heading)",
		"3\\. [Third Heading](#third-heading)",
	}
	for _, entry := range expectedEntries {
		if !strings.Contains(content, entry) {
			t.Errorf("generated TOC should contain entry: %s", entry)
		}
	}
}

func TestGenerateCommandPositionalArg(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.md")
	if err := os.WriteFile(testFile, []byte(generateTestContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	setupGenerateTest()
	RootCmd.SetArgs([]string{"generate", testFile})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	updated, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}

	if !strings.Contains(string(updated), "<!-- START_TABLE_OF_CONTENTS -->") {
		t.Error("generate with a positional file argument should update the file like --file")
	}
}

func TestGenerateCommandDryRun(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.md")
	if err := os.WriteFile(testFile, []byte(generateTestContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	setupGenerateTest()
	RootCmd.SetArgs([]string{"generate", "--file", testFile, "--dry-run"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	updated, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(updated) != generateTestContent {
		t.Error("--dry-run should not modify the file")
	}
}

func TestGenerateCommandExclude(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.md")
	if err := os.WriteFile(testFile, []byte(generateTestContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	setupGenerateTest()
	RootCmd.SetArgs([]string{"generate", "--file", testFile, "--exclude", "Second Heading"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	updated, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read updated file: %v", err)
	}
	content := string(updated)

	if strings.Contains(content, "[Second Heading]") {
		t.Error("--exclude should remove the matching heading from the TOC")
	}
	if !strings.Contains(content, "[First Heading]") || !strings.Contains(content, "[Third Heading]") {
		t.Error("--exclude should not remove non-matching headings")
	}
}

func TestGenerateCommandMissingFile(t *testing.T) {
	tempDir := t.TempDir()
	missingFile := filepath.Join(tempDir, "does-not-exist.md")

	setupGenerateTest()
	RootCmd.SetArgs([]string{"generate", "--file", missingFile})

	if err := RootCmd.Execute(); err == nil {
		t.Error("expected an error for a missing file, got nil")
	}
}
