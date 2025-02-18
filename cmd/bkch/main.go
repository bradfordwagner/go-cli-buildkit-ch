package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "bkch",
}

func init() {
	rootCmd.AddCommand(
		//serverCmd,
		diskMaintenanceCmd,
	)
}

func main() {
	// cobra
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
