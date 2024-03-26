package main

import (
	"fmt"
	"os"
	"template_cli/internal/args"

	"github.com/bradfordwagner/go-util/flag_helper"
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
		l.With("args", defaultArgs).Info("hi friends")
	},
}

var defaultArgs args.Args

func init() {
	rootCmd.AddCommand(myVerb)
	fs := myVerb.Flags()
	flag_helper.CreateFlag(fs, &defaultArgs.HelloWorld, "hello_world", "w", "default_value", "hello world")
}

func main() {
	flag_helper.Load(&defaultArgs)
	// cobra
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
