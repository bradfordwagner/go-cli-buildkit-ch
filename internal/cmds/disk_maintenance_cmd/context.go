package disk_maintenance_cmd

import (
	"bkch/internal/args"
	"bkch/internal/buildkit_client"
	"bkch/internal/pod_component"
	"bkch/internal/pvc_finder"
	"k8s.io/client-go/kubernetes"
)

// Context is the context for the disk maintenance command
type Context struct {
	Args      args.DiskMaintenanceArgs
	PvcFinder pvc_finder.Interface
	Prune     buildkit_client.PruneInterface
	Pod       pod_component.Interface
}

// NewContext creates a new Context
func NewContext(
	args args.DiskMaintenanceArgs,
	client kubernetes.Interface,
) Context {
	return Context{
		Args:      args,
		PvcFinder: pvc_finder.NewPvcFinder(args, client),
		Prune:     buildkit_client.NewPrune(),
		Pod:       pod_component.New(client),
	}
}
