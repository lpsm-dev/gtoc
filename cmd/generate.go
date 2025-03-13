package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/lpsm-dev/gtoc/internal/generator"
	"github.com/lpsm-dev/gtoc/internal/logger"
	"github.com/spf13/cobra"
)

var (
	filePath      string
	depth         int
	excludePaths  string
	dryRun        bool
	prettyOutput  bool
)

// generateCmd handles TOC generation for markdown files
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
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get file path from args or flag
		if len(args) > 0 && filePath == "" {
			filePath = args[0]
		}

		if filePath == "" {
			logger.Error("File path is required", "error", "no file path provided")
			return fmt.Errorf("file path is required (provide it as an argument or with --file flag)")
		}

		logger.Debug("Processing file", "path", filePath, "depth", depth)

		// Validate file exists
		absFilePath, err := filepath.Abs(filePath)
		if err != nil {
			logger.Error("Failed to get absolute path", "error", err)
			return fmt.Errorf("failed to get absolute path: %w", err)
		}

		if _, err := os.Stat(absFilePath); os.IsNotExist(err) {
			logger.Error("File does not exist", "path", absFilePath)
			return fmt.Errorf("file does not exist: %s", absFilePath)
		}

		// Process exclude paths
		excludeList := []string{}
		if excludePaths != "" {
			for _, path := range strings.Split(excludePaths, ",") {
				excludeList = append(excludeList, strings.TrimSpace(path))
			}
			logger.Debug("Using exclude paths", "paths", excludeList)
		}

		// Generate TOC
		logger.Info("Generating table of contents", "file", absFilePath)
		gen := generator.NewGenerator(absFilePath, depth, excludeList)
		toc, err := gen.Generate()
		if err != nil {
			logger.Error("Failed to generate table of contents", "error", err)
			return fmt.Errorf("failed to generate table of contents: %w", err)
		}

		// Update file or display in dry run mode
		if dryRun {
			logger.Info("Dry run mode - not updating file")
			
			// Using glamour for pretty output when --pretty flag is enabled
			if prettyOutput {
				logger.Info("Pretty output enabled", "render", "glamour")
				
				// Configurar o renderizador do glamour
				rendererOpts := []glamour.TermRendererOption{
					glamour.WithWordWrap(100),
					glamour.WithStandardStyle("light"),
				}
				
				// Criar o renderizador
				r, err := glamour.NewTermRenderer(rendererOpts...)
				
				if err != nil {
					logger.Error("Failed to create markdown renderer", "error", err)
					// Em caso de erro, mostramos o toc não formatado
					outputMarkdown(gen, toc)
					return nil
				}

				// Com --pretty, sempre mostramos o arquivo completo (similar a --view-full anteriormente)
				// Ler o conteúdo do arquivo
				fileContent, err := os.ReadFile(absFilePath)
				if err != nil {
					logger.Error("Failed to read file", "error", err)
					return fmt.Errorf("failed to read file: %w", err)
				}
				fileContentStr := string(fileContent)
				
				// Atualizar o conteúdo com o novo TOC
				updatedContent := gen.GetFileWithUpdatedTOC(fileContentStr, toc)
				
				// Renderizar o conteúdo atualizado
				fmt.Println("Dry run mode. The following is how the file would look with the updated TOC:")
				renderedContent, err := r.Render(updatedContent)
				if err != nil {
					logger.Error("Failed to render content", "error", err)
					outputMarkdown(gen, toc)
					return nil
				}
				
				fmt.Println(renderedContent)
			} else {
				// Output plain markdown in dry-run mode (default)
				outputMarkdown(gen, toc)
			}
		} else {
			logger.Info("Updating file with generated table of contents")
			if err := gen.UpdateFile(toc); err != nil {
				logger.Error("Failed to update file", "error", err)
				return fmt.Errorf("failed to update file: %w", err)
			}
			logger.Info("File updated successfully", "path", filePath)
			fmt.Printf("Successfully updated %s with the generated table of contents\n", filePath)
		}

		return nil
	},
}

// outputMarkdown outputs raw markdown in dry-run mode
func outputMarkdown(gen *generator.Generator, toc string) {
	fmt.Println("Dry run mode. The following table of contents would be generated:")
	fmt.Println("\n" + toc + "\n")
}

func init() {
	generateCmd.Flags().StringVar(&filePath, "file", "", "Path to the markdown file to update")
	generateCmd.Flags().IntVar(&depth, "depth", 0, "Maximum heading depth (0 for unlimited)")
	generateCmd.Flags().StringVar(&excludePaths, "exclude", "", "Comma-separated list of heading patterns to exclude")
	generateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing")
	generateCmd.Flags().BoolVar(&prettyOutput, "pretty", false, "Render output with formatting and show full file in dry-run mode")
}
