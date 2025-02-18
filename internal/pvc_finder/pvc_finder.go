package pvc_finder

import (
	"bkch/internal/args"
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

// Interface is an interface for finding PVCs
type Interface interface {
	FindPvcs() ([]string, error)
	Delete(pvcName string) error
}

// pvcFinderImpl is an implementation of Interface
type pvcFinderImpl struct {
	Args   args.DiskMaintenanceArgs
	Client kubernetes.Interface
}

// enforce interface on PvcFinderImpl
var _ Interface = &pvcFinderImpl{}

// NewPvcFinder returns a new Interface
func NewPvcFinder(args args.DiskMaintenanceArgs, client kubernetes.Interface) Interface {
	return &pvcFinderImpl{
		Args:   args,
		Client: client,
	}
}

// FindPvcs returns a list of PVCs
func (p pvcFinderImpl) FindPvcs() (pvcs []string, err error) {
	ctx := context.Background()
	list, err := p.Client.
		CoreV1().
		PersistentVolumeClaims(p.Args.KubernetesNamespace).
		List(ctx, metav1.ListOptions{
			LabelSelector: p.Args.SelectorLabel,
		})
	if err != nil {
		return
	}
	for _, pvc := range list.Items {
		pvcs = append(pvcs, pvc.Name)
	}

	return
}

func (p pvcFinderImpl) Delete(pvcName string) (err error) {
	ctx := context.Background()
	err = p.Client.
		CoreV1().
		PersistentVolumeClaims(p.Args.KubernetesNamespace).
		Delete(ctx, pvcName, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
