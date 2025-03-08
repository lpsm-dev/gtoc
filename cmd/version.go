package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is set during build using ldflags
var Version = "dev"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gtoc",
	Long:  `All software has versions. This is gtoc's.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("gtoc version %s\n", Version)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
