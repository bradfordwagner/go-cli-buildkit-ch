package pod_component

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Delete deletes a pod in a given namespace
func (i impl) Delete(namespace, podName string) error {
	ctx := context.Background()
	return i.client.CoreV1().
		Pods(namespace).
		Delete(ctx, podName, metav1.DeleteOptions{})
}
