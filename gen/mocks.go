package mocks

//go:generate mockgen -destination=mocks/pvc_finder/module.go -package=mock_pvc_finder bkch/internal/pvc_finder Interface
//go:generate mockgen -destination=mocks/buildkit_client/module.go -package=mock_buildkit_client bkch/internal/buildkit_client PruneInterface
//go:generate mockgen -destination=mocks/pod_component/module.go -package=mock_pod_component bkch/internal/pod_component Interface
