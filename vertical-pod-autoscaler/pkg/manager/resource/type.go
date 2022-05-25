package resource

var (
	resourcesControlKey string = "autoscaling/vpa"
	podUpdateModeKey    string = `vpa/update-mode`
)

// Resource is runnable interface
type Resource interface {
	//WaitForCacheSyncOrDir()
	Run(stopCh <-chan struct{})
}
