package pod_component_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPodComponent(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PodComponent Suite")
}
