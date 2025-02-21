package pvc_finder_test

import (
	"bkch/internal/args"
	"bkch/internal/pvc_finder"
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

type testargs struct {
	args     args.DiskMaintenanceArgs
	disks    []disk
	expected expected
}

type expected struct {
	pvcs []string
	err  error
}

type disk struct {
	name      string
	namespace string
	appLabel  string
}

func test(args testargs) {
	clientset := fake.NewClientset()
	pvcFinder := pvc_finder.NewPvcFinder(args.args, clientset)

	// create each disk
	for _, d := range args.disks {
		clientset.CoreV1().PersistentVolumeClaims(d.namespace).Create(context.Background(), &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:   d.name,
				Labels: map[string]string{"app": d.appLabel},
			},
		}, metav1.CreateOptions{})
	}

	// find the disks
	pvcs, err := pvcFinder.FindPvcs()

	// check the results
	if err == nil {
		Expect(err).Should(BeNil())
	} else {
		Expect(err).Should(Equal(args.expected.err))
	}
	Expect(pvcs).Should(ConsistOf(args.expected.pvcs))
}

var _ = Describe("Interface", func() {
	It("should find no pvcs", func() {
		test(testargs{
			args: args.DiskMaintenanceArgs{
				KubernetesNamespace: "buildkit",
				SelectorLabel:       "app=buildkit",
			},
			disks: nil,
			expected: expected{
				pvcs: []string{},
				err:  nil,
			},
		})
	})
	It("should find pvc in the buildkit namespace, but none elsewhere", func() {
		test(testargs{
			args: args.DiskMaintenanceArgs{
				KubernetesNamespace: "buildkit",
				SelectorLabel:       "app=buildkit",
			},
			disks: []disk{
				{name: "buildkit-disk", namespace: "buildkit", appLabel: "buildkit"},
				{name: "other-disk1", namespace: "other", appLabel: "buildkit"},
				{name: "other-disk2", namespace: "other", appLabel: "other"},
			},
			expected: expected{
				pvcs: []string{"buildkit-disk"},
				err:  nil,
			},
		})
	})
})
