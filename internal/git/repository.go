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
	
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Exit code 1 means the file is not ignored
			if exitError.ExitCode() == 1 {
				return false, nil
			}
		}
		return false, err
	}
	
	// Exit code 0 means the file is ignored
	return true, nil
}

// ListMarkdownFiles returns a list of markdown files in the repository
func ListMarkdownFiles(repoRoot string, pattern string, excludePaths []string) ([]string, error) {
	var files []string

	// Check if the pattern contains a directory
	var searchDir string
	if strings.Contains(pattern, "/") {
		parts := strings.Split(pattern, "/")
		searchDir = filepath.Join(repoRoot, parts[0])
		pattern = strings.Join(parts[1:], "/")
	} else {
		searchDir = repoRoot
	}

	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file matches pattern
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if !matched {
			return nil
		}

		// Check if file is ignored by Git
		ignored, err := IsIgnored(path)
		if err != nil {
			return err
		}
		if ignored {
			return nil
		}

		// Check if file is in exclude paths
		relPath, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}

		for _, excludePath := range excludePaths {
			if matched, err := filepath.Match(excludePath, relPath); err != nil {
				return err
			} else if matched {
				return nil
			}
		}

		files = append(files, path)
		return nil
	})

	return files, err
}

