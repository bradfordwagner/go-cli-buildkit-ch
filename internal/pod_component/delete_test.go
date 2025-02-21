package pod_component_test

import (
	"bkch/internal/pod_component"
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

var _ = Describe("Delete", func() {
	type pod struct {
		name      string
		namespace string
	}

	type expect struct {
		err string
	}

	type args struct {
		namespace string
		podName   string
		pods      []pod
		expect    expect
	}

	var test = func(a args) {
		clientset := fake.NewClientset()

		// create pods
		for _, p := range a.pods {
			_, err := clientset.CoreV1().Pods(p.namespace).Create(context.Background(), &v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: p.name,
				},
			}, metav1.CreateOptions{})
			Expect(err).ShouldNot(HaveOccurred())
		}

		pc := pod_component.New(clientset)
		err := pc.Delete(a.namespace, a.podName)
		if a.expect.err == "" {
			Expect(err).Should(BeNil())
		} else {
			Expect(err.Error()).To(ContainSubstring(a.expect.err))
		}
	}

	It("deletes a pod that exists", func() {
		test(args{
			namespace: "namespace",
			podName:   "podName",
			pods: []pod{
				{name: "podName", namespace: "namespace"},
			},
			expect: expect{
				err: "",
			},
		})
	})

	It("deletes a pod that does not exist", func() {
		test(args{
			namespace: "namespace",
			podName:   "dne",
			pods: []pod{
				{name: "podName", namespace: "namespace"},
			},
			expect: expect{
				err: `pods "dne" not found`,
			},
		})
	})
})
