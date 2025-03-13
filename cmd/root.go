package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/lpsm-dev/gtoc/internal/logger"
	"github.com/spf13/cobra"
)

var (
	// Opções de log
	logLevel    string
	logFormat   string
	logNoColors bool
)

// RootCmd is the base command for the CLI application
var RootCmd = &cobra.Command{
	Use:   "gtoc",
	Short: "Generate a table of contents for markdown files",
	Long: `gtoc is a CLI tool that generates a table of contents based on the headings
in a markdown file and updates the file with the generated table of contents.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Configurar o logger global antes de qualquer comando ser executado
		setupLogger()
	},
}

// Execute runs the root command
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

// setupLogger configura o logger global com base nas flags
func setupLogger() {
	// Converter string de nível para log.Level
	var level log.Level
	switch logLevel {
	case "debug":
		level = log.DebugLevel
	case "info":
		level = log.InfoLevel
	case "warn":
		level = log.WarnLevel
	case "error":
		level = log.ErrorLevel
	case "fatal":
		level = log.FatalLevel
	default:
		level = log.InfoLevel
	}

	// Configurar o logger com valores simplificados
	opts := &logger.Options{
		Level:      level,
		TimeFormat: "15:04:05", // Formato de hora padrão
		ShowCaller: false,      // Não mostrar o caller por padrão
	}

	// Definir formato de saída (JSON, texto, etc.)
	if logFormat == "json" {
		logger.Init(opts)
		logger.GetLogger().SetFormatter(log.JSONFormatter)
	} else {
		logger.Init(opts)
		logger.GetLogger().SetFormatter(log.TextFormatter)
	}

	// Desativar cores se solicitado
	if logNoColors {
		emptyStyles := &log.Styles{}
		logger.GetLogger().SetStyles(emptyStyles)
	}
}

func init() {
	// Adicionar flags persistentes para configuração do logger
	RootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set log level (debug, info, warn, error, fatal)")
	RootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "Log format (text, json)")
	RootCmd.PersistentFlags().BoolVar(&logNoColors, "log-no-colors", false, "Disable colors in logs")

	// Adicionar comandos
	RootCmd.AddCommand(generateCmd)
	RootCmd.AddCommand(analyzeCmd)
	RootCmd.AddCommand(versionCmd)
}
