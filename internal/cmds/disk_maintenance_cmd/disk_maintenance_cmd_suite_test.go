package disk_maintenance_cmd_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDiskMaintenanceCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DiskMaintenanceCmd Suite")
}
