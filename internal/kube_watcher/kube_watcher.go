package kube_watcher

import (
	"bkch/internal/args"
	"bkch/internal/cache"
	"context"

	bwutil "github.com/bradfordwagner/go-util"
	"github.com/bradfordwagner/go-util/log"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"go.uber.org/zap"
)

type Watcher struct {
	a      args.ServerArgs
	c      *bwutil.Lockable[*cache.Cache]
	ctx    context.Context
	cancel context.CancelFunc
	l      *zap.SugaredLogger
}

// NewWatcher creates a new Watcher
func NewWatcher(
	ctx context.Context,
	cancel context.CancelFunc,
	a args.ServerArgs,
	c *bwutil.Lockable[*cache.Cache],
) *Watcher {
	l := log.Log().With("module", "kube_watcher")
	return &Watcher{
		a:      a,
		c:      c,
		cancel: cancel,
		ctx:    ctx,
		l:      l,
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

	w.updatePodCache(clientset)
	go w.watchPods(clientset)
	go w.watchStatefulset(clientset)

	// watcher, err := clientset.AppsV1().StatefulSets(w.a.KubernetesNamespace).Watch(w.ctx, metav1.ListOptions{
	// 	LabelSelector: "app=buildkit",
	// })
	// if err != nil {
	// 	panic(err.Error())
	// }
	// event := <-watcher.ResultChan()
	// w.l.With("event", event).Info("event")

	// pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	// if err != nil {
	// 	panic(err.Error())
	// }
	// fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
	// watch kubernetes
}

// watchPods watches for pod changes
func (w *Watcher) watchPods(clientset *kubernetes.Clientset) {
	watcher, err := clientset.CoreV1().Pods("").Watch(context.TODO(), metav1.ListOptions{
		LabelSelector: w.a.SelectorLabel,
	})
	if err != nil {
		w.l.With("error", err).Error("failed to watch pods")
		w.cancel()
		return
	}
	for {
		select {
		case event := <-watcher.ResultChan():
			po, ok := event.Object.(*v1.Pod)
			if !ok {
				break
			}

			isReady := isPodReady(po)
			w.c.SetF(func(v *cache.Cache) (*cache.Cache, error) {
				podMeta, ok := v.Pods[po.Name]
				if !ok || podMeta.IsAvailable != isReady {
					w.l.With("pod", po.Name, "is_available", isReady).Info("pod availability changed")
				}
				cachedPod := v.GetPod(po.Name)
				cachedPod.IsAvailable = isReady
				return v, nil
			})
		}
	}

}

func isPodReady(pod *v1.Pod) bool {
	for i, _ := range pod.Status.ContainerStatuses {
		status := pod.Status.ContainerStatuses[i]
		if status.Ready &&
			*status.Started &&
			status.State.Running != nil &&
			status.Name == "main" {
			return true
		}
	}
	return false
}

// watchStatefulset watches for statefulset changes
func (w *Watcher) watchStatefulset(clientset *kubernetes.Clientset) {
	// watch for statefulset changes
	watcher, err := clientset.AppsV1().StatefulSets(w.a.KubernetesNamespace).Watch(w.ctx, metav1.ListOptions{
		LabelSelector: w.a.SelectorLabel,
	})

	// stop the server
	if err != nil {
		w.l.With("error", err).Error("failed to watch statefulset")
		w.cancel()
		return
	}

	for {
		select {
		case event := <-watcher.ResultChan():
			// cast to StatefulSet
			statefulset, ok := event.Object.(*appsv1.StatefulSet)
			if !ok {
				break
			}
			// update replicas
			_ = w.c.SetF(func(v *cache.Cache) (*cache.Cache, error) {
				replicas := int(*statefulset.Spec.Replicas)
				// let us know if replicas have changed
				if replicas != v.Replicas {
					v.Replicas = replicas
					w.l.With("replicas", replicas).Info("replicas updated")
				}
				return v, nil
			})
			// update pod cache
			_ = w.updatePodCache(clientset)
		case <-w.ctx.Done():
			return
		}
	}
}

func (w *Watcher) updatePodCache(clientset *kubernetes.Clientset) (err error) {
	list, err := clientset.CoreV1().Pods(w.a.KubernetesNamespace).List(w.ctx, metav1.ListOptions{
		LabelSelector: w.a.SelectorLabel,
	})
	if err != nil {
		w.l.With("error", err).Error("failed to list pods")
		return
	}

	pods := list.Items
	_ = w.c.SetF(func(v *cache.Cache) (*cache.Cache, error) {
		cachePods := make(map[string]*cache.Pod)
		for i, _ := range pods {
			po := &pods[i]
			isReady := isPodReady(po)
			cachePods[po.Name] = &cache.Pod{
				IsAvailable: isReady,
			}

			// check for changes
			orig, ok := v.Pods[po.Name]
			if len(cachePods) == 0 || (ok && orig.IsAvailable != cachePods[po.Name].IsAvailable) {
				w.l.With("pod", po.Name, "is_available", isReady).Info("pod availability changed")
			}
			v.Pods = cachePods
		}

		return v, nil
	})
	return
}
