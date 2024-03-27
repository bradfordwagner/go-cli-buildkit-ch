package main

import (
	"bkch/internal/args"
	"bkch/internal/cmds/server_cmd"

	"github.com/bradfordwagner/go-util/flag_helper"
	"github.com/spf13/cobra"
)

func init() {
	fs := serverCmd.Flags()
	flag_helper.CreateFlag(fs, &serverArgs.Port, "port", "p", 8888, "env.PORT,default=8888: port to open http server on")
}

var serverArgs args.ServerArgs

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "starts a server to compute buildkit consistent hash",
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{}, cobra.ShellCompDirectiveDefault
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		flag_helper.Load(&serverArgs)
		return server_cmd.Run(serverArgs)
	},
}
