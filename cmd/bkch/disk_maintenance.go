package main

import (
	"bkch/internal/args"
	"bkch/internal/cmds/disk_maintenance_cmd"
	"fmt"
	"github.com/bradfordwagner/go-kubeclient/kube"
	"os"
	"time"

	"github.com/bradfordwagner/go-util/flag_helper"
	"github.com/spf13/cobra"
)

func init() {
	fs := diskMaintenanceCmd.Flags()
	home, _ := os.UserHomeDir()
	flag_helper.CreateFlag(fs, &diskMaintenanceArgs.Kubeconfig, "kubeconfig", "k", fmt.Sprintf("%s/.kube/config", home), "env.KUBECONFIG,default=home/.kube/config or sa token: this is to help setup kubernetes access using sa token or kubeconfig")
	flag_helper.CreateFlag(fs, &diskMaintenanceArgs.SelectorLabel, "selectorlabel", "s", "app=buildkit", "env.SELECTOR_LABEL,default='app=buildkit' selector to find statefulset+pods to watch for changes")
	flag_helper.CreateFlag(fs, &diskMaintenanceArgs.KubernetesNamespace, "kubernetes_namespace", "n", "buildkit", "env.KUBERNETES_NAMESPACE,default='buildkit' namespace to watch for statefulset+pods")
	flag_helper.CreateFlag(fs, &diskMaintenanceArgs.DnsFormatInCluster, "dns_format_in_cluster", "d", "buildkit-%d.buildkit.buildkit.svc.cluster.local:1234", "env.DNS_FORMAT_IN_CLUSTER,default='buildkit-%d.buildkit.buildkit.svc.cluster.local:1234' format for dns in cluster")
	flag_helper.CreateFlag(fs, &diskMaintenanceArgs.KeepDuration, "keep_duration", "p", time.Hour*24, "env.KEEP_DURATION,default=24h duration to keep disk data")
	flag_helper.CreateFlag(fs, &diskMaintenanceArgs.PruneTimeout, "prune_timeout", "t", time.Second*30, "env.PRUNE_TIMEOUT,default=30s timeout for prune command")
}

var diskMaintenanceArgs args.DiskMaintenanceArgs

var diskMaintenanceCmd = &cobra.Command{
	Use:   "disk_maintenance",
	Short: "runs disk maintenance 'buildctl prune' or 'kubectl delete pvc'",
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{}, cobra.ShellCompDirectiveDefault
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		flag_helper.Load(&diskMaintenanceArgs)

		// get kube client
		client, err := kube.Client()
		if err != nil {
			return err
		}

		// run disk maintenance
		return disk_maintenance_cmd.Run(disk_maintenance_cmd.NewContext(diskMaintenanceArgs, client))
	},
}
