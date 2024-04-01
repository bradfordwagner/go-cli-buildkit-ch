package main

import (
	"bkch/internal/args"
	"bkch/internal/cmds/server_cmd"
	"fmt"
	"os"

	"github.com/bradfordwagner/go-util/flag_helper"
	"github.com/spf13/cobra"
)

func init() {
	fs := serverCmd.Flags()
	flag_helper.CreateFlag(fs, &serverArgs.Port, "port", "p", 8888, "env.PORT,default=8888: port to open http server on")
	home, _ := os.UserHomeDir()
	flag_helper.CreateFlag(fs, &serverArgs.Kubeconfig, "kubeconfig", "k", fmt.Sprintf("%s/.kube/config", home), "env.KUBECONFIG,default=home/.kube/config or sa token: run server locally, this is to help setup kubernetes access using sa token or kubeconfig")
	flag_helper.CreateFlag(fs, &serverArgs.SelectorLabel, "selectorlabel", "s", "app=buildkit", "env.SELECTOR_LABEL,default='app=buildkit' selector to find statefulset+pods to watch for changes")
	flag_helper.CreateFlag(fs, &serverArgs.KubernetesNamespace, "kubernetes_namespace", "n", "buildkit", "env.KUBERNETES_NAMESPACE,default='buildkit' namespace to watch for statefulset+pods")
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
