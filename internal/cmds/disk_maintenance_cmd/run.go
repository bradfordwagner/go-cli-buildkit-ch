package disk_maintenance_cmd

import (
	"fmt"
	"github.com/bradfordwagner/go-util/log"
	"strings"
)

// Run is the main function for the serverCmd command
func Run(maintenanceContext Context) (err error) {
	l := log.Log().With("cmd", "disk_maintenance")
	l.With("args", maintenanceContext.Args).
		With("keep_duration", maintenanceContext.Args.KeepDuration.String()).
		Infof("disk_maintenance")

	// find disks matching selector label
	pvcs, err := maintenanceContext.PvcFinder.FindPvcs()
	if err != nil {
		return
	}
	numPvcs := len(pvcs)

	// associate disks with service
	failedPvcs := []string{}
	for i := 0; i < numPvcs; i++ {
		addr := fmt.Sprintf(maintenanceContext.Args.DnsFormatInCluster, i)

		// this should recover rather than return
		err := maintenanceContext.Prune.Prune(addr, maintenanceContext.Args.KeepDuration, maintenanceContext.Args.PruneTimeout)
		if err != nil {
			pvcName := pvcs[i]
			l.With("error", err, "pvc_name", pvcName).Warn("prune failed - will delete pvc")
			failedPvcs = append(failedPvcs, pvcName)
		}
	}

	// run buildctl prune against services
	for _, pvc := range failedPvcs {
		// delete pvc
		err = maintenanceContext.PvcFinder.Delete(pvc)
		if err != nil {
			l.With("error", err, "pvc_name", pvc).Error("failed to delete pvc")
			return
		}
		l.With("pvc_name", pvc).Info("deleted pvc")
		// delete associated pod to unbind pvc
		podName, _ := strings.CutPrefix(pvc, "data-")
		err = maintenanceContext.Pod.Delete(maintenanceContext.Args.KubernetesNamespace, podName)
		if err != nil {
			l.With("error", err, "pod_name", podName).Error("failed to delete pod")
			return
		}
		l.With("pod_name", podName).Info("deleted pod")
	}

	return
}
