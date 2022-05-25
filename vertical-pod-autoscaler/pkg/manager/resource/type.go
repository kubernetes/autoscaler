package resource

var (
	ResourcesControlKey string = "autoscaling/vpa"
	PodUpdateModeKey    string = `vpa/update-mode`
)

type VPAAction string

var (
	VPACreate VPAAction = "create"
	VPADelete VPAAction = "delete"
)

type Resource interface {
	//WaitForCacheSyncOrDir()
	Run(stopCh <-chan struct{})
}
