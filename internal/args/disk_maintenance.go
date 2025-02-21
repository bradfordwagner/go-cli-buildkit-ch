package args

import "time"

// DiskMaintenanceArgs is a struct that holds the arguments for the CLI
type DiskMaintenanceArgs struct {
	Kubeconfig          string        `mapstructure:"KUBECONFIG"`
	KubernetesNamespace string        `mapstructure:"KUBERNETES_NAMESPACE"`
	SelectorLabel       string        `mapstructure:"SELECTOR_LABEL"`
	DnsFormatInCluster  string        `mapstructure:"DNS_FORMAT_IN_CLUSTER"`
	KeepDuration        time.Duration `mapstructure:"KEEP_DURATION"`
	PruneTimeout        time.Duration `mapstructure:"PRUNE_TIMEOUT"`
}
