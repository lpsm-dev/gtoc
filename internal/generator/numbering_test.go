package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeNumberingFile writes content to a temp markdown file and returns its path.
func writeNumberingFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "doc.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	return path
}

func TestGenerateNumberedFile(t *testing.T) {
	content := "# First\n\n## Sub A\n\n## Sub B\n\n# Second\n\n## Sub C\n"
	path := writeNumberingFile(t, content)

	got, err := NewGenerator(path, 0, nil).GenerateNumberedFile()
	if err != nil {
		t.Fatalf("GenerateNumberedFile failed: %v", err)
	}

	wantHeadings := []string{
		"# 1. First",
		"## 1.1. Sub A",
		"## 1.2. Sub B",
		"# 2. Second",
		"## 2.1. Sub C",
	}
	for _, h := range wantHeadings {
		if !strings.Contains(got, h+"\n") {
			t.Errorf("body should contain numbered heading %q, got:\n%s", h, got)
		}
	}

	wantTOC := []string{
		"[1. First](#1-first)<br>",
		"&nbsp;&nbsp;&nbsp;[1.1. Sub A](#11-sub-a)<br>",
		"[2. Second](#2-second)<br>",
		"&nbsp;&nbsp;&nbsp;[2.1. Sub C](#21-sub-c)<br>",
	}
	for _, entry := range wantTOC {
		if !strings.Contains(got, entry) {
			t.Errorf("TOC should link to numbered heading %q, got:\n%s", entry, got)
		}
	}
}

func TestGenerateNumberedFileIsIdempotent(t *testing.T) {
	path := writeNumberingFile(t, "# First\n\n## Sub A\n\n# Second\n")

	gen := NewGenerator(path, 0, nil)
	first, err := gen.GenerateNumberedFile()
	if err != nil {
		t.Fatalf("first run failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(first), 0644); err != nil {
		t.Fatalf("failed to write first result: %v", err)
	}

	second, err := gen.GenerateNumberedFile()
	if err != nil {
		t.Fatalf("second run failed: %v", err)
	}
	if first != second {
		t.Errorf("numbering should be idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
	if strings.Contains(second, "1. 1. First") || strings.Contains(second, "# 1. 1.") {
		t.Error("re-numbering must strip the existing number instead of stacking a new one")
	}
}
