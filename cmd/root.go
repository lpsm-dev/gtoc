package cmd

import (
	"fmt"
	"os"

	"charm.land/log/v2"
	"github.com/lpsm-dev/gtoc/internal/logger"
	"github.com/spf13/cobra"
)

var (
	// Logger flags shared by every subcommand.
	logLevel    string
	logFormat   string
	logNoColors bool
)

// RootCmd is the base command for the CLI application.
var RootCmd = &cobra.Command{
	Use:   "gtoc",
	Short: "Generate a table of contents for markdown files",
	Long: `gtoc is a CLI tool that generates a table of contents based on the headings
in a markdown file and updates the file with the generated table of contents.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Configure the global logger before any command runs.
		setupLogger()
	},
}

// Execute runs the root command and reports any resulting error exactly
// once, on stderr, before exiting with a non-zero status.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// setupLogger configures the global logger based on the persistent flags.
func setupLogger() {
	opts := &logger.Options{
		Level:      logLevelFromFlag(logLevel),
		TimeFormat: "15:04:05",
		ShowCaller: false,
	}
	logger.Init(opts)

	if logFormat == "json" {
		logger.GetLogger().SetFormatter(log.JSONFormatter)
	} else {
		logger.GetLogger().SetFormatter(log.TextFormatter)
	}

	if logNoColors {
		logger.GetLogger().SetStyles(&log.Styles{})
	}
}

// logLevelFromFlag converts the --log-level flag value into a log.Level,
// defaulting to log.WarnLevel for empty or unrecognized values.
func logLevelFromFlag(level string) log.Level {
	switch level {
	case "debug":
		return log.DebugLevel
	case "info":
		return log.InfoLevel
	case "warn":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	case "fatal":
		return log.FatalLevel
	default:
		return log.WarnLevel
	}
}

func init() {
	RootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "warn", "Set log level (debug, info, warn, error, fatal)")
	RootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "Log format (text, json)")
	RootCmd.PersistentFlags().BoolVar(&logNoColors, "log-no-colors", false, "Disable colors in logs")

	// Register every subcommand here so command registration lives in a
	// single, predictable place.
	RootCmd.AddCommand(generateCmd)
	RootCmd.AddCommand(analyzeCmd)
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(upgradeCmd)
}
