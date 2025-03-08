package main

import (
	"fmt"
	"os"

	"github.com/lpsm-dev/mdtoc/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
