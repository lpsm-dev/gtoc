package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetRepositoryRoot returns the absolute path to the Git repository root
func GetRepositoryRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// IsIgnored checks if a file is ignored by Git
func IsIgnored(path string) (bool, error) {
	cmd := exec.Command("git", "check-ignore", "-q", path)
	err := cmd.Run()

	// Exit code 0 means the file is ignored
	// Exit code 1 means the file is not ignored
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ListMarkdownFiles finds all markdown files in the repository
func ListMarkdownFiles(repoRoot string, pattern string, excludePaths []string) ([]string, error) {
	var files []string

	err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		// Check if file matches pattern
		matched, err := filepath.Match(pattern, path)
		if err != nil || !matched {
			return err
		}

		// Skip files ignored by Git
		ignored, err := IsIgnored(path)
		if err != nil || ignored {
			return err
		}

		// Skip files in exclude paths
		relPath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}

		for _, excludePath := range excludePaths {
			if matched, err := filepath.Match(excludePath, relPath); err != nil || matched {
				return err
			}
		}

		files = append(files, path)
		return nil
	})

	return files, err
}
