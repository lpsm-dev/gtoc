package cmd

import (
	"fmt"
	"runtime"

	"github.com/lpsm-dev/gtoc/internal/logger"
	"github.com/spf13/cobra"
)

// Version is set during build using ldflags.
var Version = "dev"

// versionCmd displays version and build information.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("Displaying version information")

		info := fmt.Sprintf("gtoc version %s\n", Version)
		info += fmt.Sprintf("  Go version: %s\n", runtime.Version())
		info += fmt.Sprintf("  OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)

		logger.Info("Version information", "version", Version, "go_version", runtime.Version())
		fmt.Print(info)
	},
}
