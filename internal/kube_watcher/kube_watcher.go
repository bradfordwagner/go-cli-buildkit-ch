package kube_watcher

import (
	"bkch/internal/args"
	"context"
	"fmt"

	"github.com/bradfordwagner/go-util/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"go.uber.org/zap"
)

type Watcher struct {
	a   args.ServerArgs
	ctx context.Context
	l   *zap.SugaredLogger
}

// NewWatcher creates a new Watcher
func NewWatcher(ctx context.Context, cancel context.CancelFunc, a args.ServerArgs) *Watcher {
	l := log.Log().With("module", "kube_watcher")
	return &Watcher{
		l:   l,
		a:   a,
		ctx: ctx,
	}
}

func (w *Watcher) Start() {
	w.l.Info("starting")
	config, err := clientcmd.BuildConfigFromFlags("", w.a.Kubeconfig)
	if err != nil {
		w.l.Errorw("failed to build kubeconfig", "error", err)
		return
	}
	// this should fallback to in cluster eventually
	// if not try: https://github.com/kubernetes/client-go/tree/master/examples/in-cluster-client-configuration
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	i, err2 := clientset.AppsV1().StatefulSets("buildkit").Watch(w.ctx, metav1.ListOptions{})
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	// watch kubernetes
}
