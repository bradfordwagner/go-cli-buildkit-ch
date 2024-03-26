package main

import (
	"fmt"
	"os"

	"github.com/bradfordwagner/go-util/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "testcli",
}

var myVerb = &cobra.Command{
	Use:   "myVerb",
	Short: "myVerb does something",
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{}, cobra.ShellCompDirectiveDefault
	},
	Run: func(cmd *cobra.Command, args []string) {
		l := log.Log().With("cmd", "myVerb")
		l.Info("hi friends")
	},
}

func init() {
	rootCmd.AddCommand(myVerb)
}

func main() {
	// cobra
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
