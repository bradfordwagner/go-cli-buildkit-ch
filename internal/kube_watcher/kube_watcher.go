package kube_watcher

import (
	"bkch/internal/args"
	"bkch/internal/cache"
	"bkch/internal/constants"
	"context"
	"k8s.io/apimachinery/pkg/watch"
	"strconv"
	"strings"

	bwutil "github.com/bradfordwagner/go-util"
	"github.com/bradfordwagner/go-util/log"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"go.uber.org/zap"
)

var watchTimeoutSeconds = bwutil.Pointer(int64(86400)) // 1 day in seconds

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

	clientset, err := w.auth()
	if err != nil {
		w.l.With("error", err).Error("failed to authenticate")
		w.cancel()
		return
	}

	w.updatePodCache(clientset)
	go w.watchPods(clientset)
	go w.watchStatefulset(clientset)
}

func (w *Watcher) auth() (clientset *kubernetes.Clientset, err error) {
	// in cluster
	config, err := rest.InClusterConfig()
	if err != nil {
		w.l.With("error", err).Warn("failed to create in cluster config - in cluster")
	} else {
		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			w.l.With("error", err).Warn("failed to create kubernetes client, trying file based")
		} else {
			return
		}
	}

	// kubeconfig / file based
	config, err = clientcmd.BuildConfigFromFlags("", w.a.Kubeconfig)
	if err != nil {
		w.l.With("error", err).Error("failed to create in cluster config - file based")
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		w.l.With("error", err).Error("failed to create kubernetes client")
	}

	return
}

// watchPods watches for pod changes
func (w *Watcher) watchPods(clientset *kubernetes.Clientset) {
	watcher, err := w.createPodWatcher(clientset)
	if err != nil {
		w.l.With("error", err).Error("failed to watch pods")
		w.cancel()
		return
	}
	for {
		select {
		case event, open := <-watcher.ResultChan():
			// handle closed watch channel
			if !open {
				watcher, err = w.createPodWatcher(clientset)
				if err != nil {
					w.l.With("error", err).Error("failed to watch pods")
					w.cancel()
					return
				}
				// ensure we stop processing
				continue
			}

			po := event.Object.(*v1.Pod)
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

func (w *Watcher) createPodWatcher(clientset *kubernetes.Clientset) (watch.Interface, error) {
	watcher, err := clientset.CoreV1().Pods(w.a.KubernetesNamespace).Watch(context.TODO(), metav1.ListOptions{
		LabelSelector:  w.a.SelectorLabel,
		TimeoutSeconds: watchTimeoutSeconds,
	})
	return watcher, err
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
	watcher, err := w.createStatefulsetWatcher(clientset)
	// stop the server on error
	if err != nil {
		w.l.With("error", err).Error("failed to watch statefulset")
		w.cancel()
		return
	}

	for {
		select {
		case event, open := <-watcher.ResultChan():
			if !open {
				watcher, err = w.createStatefulsetWatcher(clientset)
				if err != nil {
					w.l.With("error", err).Error("failed to watch statefulset")
					w.cancel()
					return
				}
				// ensure we stop processing
				continue
			}

			// cast to StatefulSet
			statefulset := event.Object.(*appsv1.StatefulSet)
			// update replicas
			_ = w.c.SetF(func(v *cache.Cache) (*cache.Cache, error) {
				replicas := int(*statefulset.Spec.Replicas)
				// let us know if replicas have changed
				if replicas != v.Replicas {
					v.Replicas = replicas
					w.l.With("replicas", replicas).Info("replicas updated")
				}
				// extract annotations for dns format to use in CH requests
				apiGatewayFormat := statefulset.Spec.Template.Annotations[constants.DnsFormatAnnotationApiGateway.String()]
				inClusterFormat := statefulset.Spec.Template.Annotations[constants.DnsFormatAnnotationInCluster.String()]
				v.DnsFormatApiGateway, v.DnsFormatInCluster = apiGatewayFormat, inClusterFormat
				return v, nil
			})
			// update pod cache
			_ = w.updatePodCache(clientset)
		case <-w.ctx.Done():
			return
		}
	}
}

func (w *Watcher) createStatefulsetWatcher(clientset *kubernetes.Clientset) (watch.Interface, error) {
	watcher, err := clientset.AppsV1().StatefulSets(w.a.KubernetesNamespace).Watch(w.ctx, metav1.ListOptions{
		LabelSelector:  w.a.SelectorLabel,
		TimeoutSeconds: watchTimeoutSeconds,
	})
	return watcher, err
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

			// compute index
			// use pod name last value split by '-'
			parts := strings.Split(po.Name, "-")
			index, _ := strconv.Atoi(parts[len(parts)-1])

			// update cache
			cachePods[po.Name] = &cache.Pod{
				IsAvailable: isReady,
				Index:       index,
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
