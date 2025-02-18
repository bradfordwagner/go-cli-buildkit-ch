package pod_component

import "k8s.io/client-go/kubernetes"

// Interface is an interface for deleting pods
type Interface interface {
	Delete(namespace, podName string) error
}

type impl struct {
	client kubernetes.Interface
}

// enforce interface
var _ Interface = &impl{}

// New returns a new pod_util
func New(client kubernetes.Interface) Interface {
	return &impl{
		client: client,
	}
}
