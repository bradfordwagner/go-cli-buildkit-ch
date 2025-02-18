package disk_maintenance_cmd_test

import (
	mock_buildkit_client "bkch/gen/mocks/buildkit_client"
	mock_pod_component "bkch/gen/mocks/pod_component"
	mock_pvc_finder "bkch/gen/mocks/pvc_finder"
	"bkch/internal/args"
	"bkch/internal/cmds/disk_maintenance_cmd"
	"errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	"time"
)

var _ = Describe("Run", func() {
	type expect struct {
		err string
	}

	type findpvcs struct {
		pvcs []string
		err  error
	}

	type prune struct {
		addr string
		err  error
	}

	type deletepvcs struct {
		name string
		err  error
	}

	type deletepods struct {
		name string
		err  error
	}

	type testargs struct {
		expect     expect
		findpvcs   findpvcs
		prune      []prune
		deletepvcs []deletepvcs
		deletepods []deletepods
	}

	var ctrl *gomock.Controller
	BeforeEach(func() { ctrl = gomock.NewController(GinkgoT()) })
	AfterEach(func() { ctrl.Finish() })

	var test = func(a testargs) {
		maintenanceArgs := args.DiskMaintenanceArgs{
			Kubeconfig:          "",
			KubernetesNamespace: "buildkit",
			SelectorLabel:       "app=buildkit",
			DnsFormatInCluster:  "buildkit-%d.buildkit.buildkit.svc.cluster.local",
			KeepDuration:        time.Second * 30,
		}
		pvcFinder := mock_pvc_finder.NewMockInterface(ctrl)
		prune := mock_buildkit_client.NewMockPruneInterface(ctrl)
		pod := mock_pod_component.NewMockInterface(ctrl)

		// setup assertions
		func() {
			pvcFinder.EXPECT().FindPvcs().Return(a.findpvcs.pvcs, a.findpvcs.err)
			if a.findpvcs.err != nil {
				return
			}

			var failedPvcs []int
			for i, p := range a.prune {
				prune.EXPECT().Prune(p.addr, maintenanceArgs.KeepDuration).Return(p.err)
				if p.err != nil {
					failedPvcs = append(failedPvcs, i)
				}
			}

			var deletePodsIndex int
			for _, pvc := range a.deletepvcs {
				pvcFinder.EXPECT().Delete(pvc.name).Return(pvc.err)
				if pvc.err != nil {
					return
				}
				deletepo := a.deletepods[deletePodsIndex]
				pod.EXPECT().Delete(maintenanceArgs.KubernetesNamespace, deletepo.name).Return(deletepo.err)
				if deletepo.err != nil {
					return
				}
				deletePodsIndex++
			}
		}()

		// invoke
		context := disk_maintenance_cmd.Context{
			Args:      maintenanceArgs,
			PvcFinder: pvcFinder,
			Prune:     prune,
			Pod:       pod,
		}
		resErr := disk_maintenance_cmd.Run(context)

		// assert
		if a.expect.err != "" {
			Expect(resErr.Error()).To(ContainSubstring(a.expect.err))
		} else {
			Expect(resErr).ShouldNot(HaveOccurred())
		}
	}

	It("Succeed", func() {
		test(testargs{
			expect: expect{
				err: "",
			},
			findpvcs: findpvcs{
				pvcs: []string{"pvc1", "pvc2"},
				err:  nil,
			},
			prune: []prune{
				{addr: "buildkit-0.buildkit.buildkit.svc.cluster.local:1234", err: nil},
				{addr: "buildkit-1.buildkit.buildkit.svc.cluster.local:1234", err: nil},
			},
			deletepvcs: nil,
			deletepods: nil,
		})
	})
	It("Fails to find pvcs", func() {
		test(testargs{
			expect: expect{
				err: "could not find pvcs",
			},
			findpvcs: findpvcs{
				pvcs: []string{"pvc1", "pvc2"},
				err:  errors.New("could not find pvcs"),
			},
			prune:      []prune{},
			deletepvcs: nil,
			deletepods: nil,
		})
	})
	It("Fails to prune buildkit-0, succeeds on others, and succeeds on deleting pvc+pod", func() {
		test(testargs{
			expect: expect{
				err: "",
			},
			findpvcs: findpvcs{
				pvcs: []string{"data-buildkit-0", "data-buildkit-1"},
				err:  nil,
			},
			prune: []prune{
				{addr: "buildkit-0.buildkit.buildkit.svc.cluster.local:1234", err: errors.New("couldnt prune")},
				{addr: "buildkit-1.buildkit.buildkit.svc.cluster.local:1234", err: nil},
			},
			deletepvcs: []deletepvcs{
				{name: "data-buildkit-0", err: nil},
			},
			deletepods: []deletepods{
				{name: "buildkit-0", err: nil},
			},
		})
	})
	It("Fails to prune buildkit-0, succeeds on others, and fails to delete pvc", func() {
		test(testargs{
			expect: expect{
				err: "failed to delete pvc",
			},
			findpvcs: findpvcs{
				pvcs: []string{"data-buildkit-0", "data-buildkit-1"},
				err:  nil,
			},
			prune: []prune{
				{addr: "buildkit-0.buildkit.buildkit.svc.cluster.local:1234", err: errors.New("couldnt prune")},
				{addr: "buildkit-1.buildkit.buildkit.svc.cluster.local:1234", err: nil},
			},
			deletepvcs: []deletepvcs{
				{name: "data-buildkit-0", err: errors.New("failed to delete pvc")},
			},
			deletepods: []deletepods{},
		})
	})
	It("Fails to prune buildkit-0, succeeds on others, and succeeds on deleting pvc, fails to delete pod", func() {
		test(testargs{
			expect: expect{
				err: "failed to delete pod",
			},
			findpvcs: findpvcs{
				pvcs: []string{"data-buildkit-0", "data-buildkit-1"},
				err:  nil,
			},
			prune: []prune{
				{addr: "buildkit-0.buildkit.buildkit.svc.cluster.local:1234", err: errors.New("couldnt prune")},
				{addr: "buildkit-1.buildkit.buildkit.svc.cluster.local:1234", err: nil},
			},
			deletepvcs: []deletepvcs{
				{name: "data-buildkit-0", err: nil},
			},
			deletepods: []deletepods{
				{name: "buildkit-0", err: errors.New("failed to delete pod")},
			},
		})
	})
})
