package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/glamour/v2"
	"github.com/lpsm-dev/gtoc/internal/generator"
	"github.com/lpsm-dev/gtoc/internal/logger"
	"github.com/spf13/cobra"
)

var (
	filePath     string
	depth        int
	excludePaths string
	dryRun       bool
	prettyOutput bool
)

// generateCmd handles TOC generation for markdown files.
var generateCmd = &cobra.Command{
	Use:     "generate [file]",
	Aliases: []string{"gen"},
	Short:   "Generate a table of contents for a markdown file",
	Long: `Generate a table of contents based on the headings in a markdown file
and update the file with the generated table of contents.

Example:
  gtoc generate README.md
  gtoc generate --file docs/index.md
  gtoc generate docs/index.md --depth 3`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGenerate,
}

// runGenerate resolves the target file, generates a TOC, and either previews
// it (--dry-run) or writes it back to the file.
func runGenerate(cmd *cobra.Command, args []string) error {
	path, err := resolveFilePath(args)
	if err != nil {
		return err
	}

	logger.Debug("Processing file", "path", path, "depth", depth)

	absFilePath, err := validateFileExists(path)
	if err != nil {
		return err
	}

	excludeList := parseExcludeList(excludePaths)
	if len(excludeList) > 0 {
		logger.Debug("Using exclude paths", "paths", excludeList)
	}

	logger.Info("Generating table of contents", "file", absFilePath)
	gen := generator.NewGenerator(absFilePath, depth, excludeList)
	toc, err := gen.Generate()
	if err != nil {
		return fmt.Errorf("failed to generate table of contents: %w", err)
	}

	if dryRun {
		return previewTOC(gen, absFilePath, toc)
	}

	return writeTOC(gen, path, toc)
}

// resolveFilePath returns the target file path from the positional argument
// or the --file flag, in that priority order.
func resolveFilePath(args []string) (string, error) {
	if len(args) > 0 && filePath == "" {
		filePath = args[0]
	}

	if filePath == "" {
		return "", fmt.Errorf("file path is required (provide it as an argument or with --file flag)")
	}

	return filePath, nil
}

// validateFileExists resolves path to an absolute path and confirms the file
// exists on disk.
func validateFileExists(path string) (string, error) {
	absFilePath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	if _, err := os.Stat(absFilePath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", absFilePath)
	}

	return absFilePath, nil
}

// parseExcludeList splits the comma-separated --exclude flag value into a
// trimmed slice of heading text patterns.
func parseExcludeList(raw string) []string {
	if raw == "" {
		return []string{}
	}

	excludeList := []string{}
	for _, pattern := range strings.Split(raw, ",") {
		excludeList = append(excludeList, strings.TrimSpace(pattern))
	}
	return excludeList
}

// previewTOC prints the generated TOC (or, with --pretty, the fully
// rendered file) without writing any changes to disk.
func previewTOC(gen *generator.Generator, absFilePath, toc string) error {
	logger.Info("Dry run mode - not updating file")

	if !prettyOutput {
		outputMarkdown(toc)
		return nil
	}

	fileContent, err := os.ReadFile(absFilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	updatedContent := gen.GetFileWithUpdatedTOC(string(fileContent), toc)
	renderPretty(updatedContent, toc)
	return nil
}

// renderPretty renders markdown content with glamour and prints it. If
// rendering fails for any reason, it logs a warning and falls back to plain
// TOC output instead of failing the command.
func renderPretty(content, toc string) {
	logger.Info("Pretty output enabled", "render", "glamour")

	// glamour v2 removed WithAutoStyle; WithEnvironmentConfig honors the
	// GLAMOUR_STYLE env var and defaults to the dark theme.
	r, err := glamour.NewTermRenderer(
		glamour.WithWordWrap(100),
		glamour.WithEnvironmentConfig(),
	)
	if err != nil {
		logger.Warn("Failed to create markdown renderer, falling back to plain output", "error", err)
		outputMarkdown(toc)
		return
	}

	fmt.Println("Dry run mode. The following is how the file would look with the updated TOC:")
	rendered, err := r.Render(content)
	if err != nil {
		logger.Warn("Failed to render content, falling back to plain output", "error", err)
		outputMarkdown(toc)
		return
	}

	fmt.Println(rendered)
}

// writeTOC updates the file on disk with the generated TOC.
func writeTOC(gen *generator.Generator, path, toc string) error {
	logger.Info("Updating file with generated table of contents")
	if err := gen.UpdateFile(toc); err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}

	logger.Info("File updated successfully", "path", path)
	fmt.Printf("Successfully updated %s with the generated table of contents\n", path)
	return nil
}

// outputMarkdown prints the raw generated TOC in dry-run mode.
func outputMarkdown(toc string) {
	fmt.Println("Dry run mode. The following table of contents would be generated:")
	fmt.Println("\n" + toc + "\n")
}

func init() {
	generateCmd.Flags().StringVar(&filePath, "file", "", "Path to the markdown file to update")
	generateCmd.Flags().IntVar(&depth, "depth", 0, "Maximum heading depth (0 for unlimited)")
	generateCmd.Flags().StringVar(&excludePaths, "exclude", "", "Comma-separated heading texts to exclude from the TOC (case-insensitive substring match)")
	generateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing")
	generateCmd.Flags().BoolVar(&prettyOutput, "pretty", false, "Render output with formatting and show full file in dry-run mode")
}
