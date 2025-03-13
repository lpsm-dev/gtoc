package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/lpsm-dev/gtoc/internal/logger"
	"github.com/spf13/cobra"
)

// Version is set during build using ldflags
var Version = "dev"

// versionCmd displays version and build information
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("Displaying version information")
		
		info := fmt.Sprintf("gtoc version %s\n", Version)
		info += fmt.Sprintf("  Go version: %s\n", runtime.Version())
		
		// Adicionar informações do ambiente
		hostname, err := os.Hostname()
		if err == nil {
			info += fmt.Sprintf("  Hostname: %s\n", hostname)
		}
		
		// Adicionar informações sobre o AWS EKS, se disponível
		if eksCluster := os.Getenv("EKS_CLUSTER_NAME"); eksCluster != "" {
			logger.Debug("AWS EKS information found", "cluster", eksCluster)
			info += fmt.Sprintf("  AWS EKS: %s\n", eksCluster)
		}
		
		logger.Info("Version information", "version", Version, "go_version", runtime.Version())
		fmt.Print(info)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
