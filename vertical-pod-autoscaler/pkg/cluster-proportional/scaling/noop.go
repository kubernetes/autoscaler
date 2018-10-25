package scaling

//import (
//	"sync"
//
//	"github.com/golang/glog"
//	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/cluster-proportional/apis/scalingpolicy/v1alpha1"
//	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/cluster-proportional/debug"
//	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/cluster-proportional/factors"
//	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/cluster-proportional/http"
//	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/cluster-proportional/scaling"
//	"k8s.io/api/core/v1"
//)
//
////type NoopSmoothing struct {
//	mutex  sync.Mutex
//	target *v1.PodSpec
//}
//
//func NewNoop() *NoopSmoothing {
//	return &NoopSmoothing{}
//}
//
//func (e *NoopSmoothing) UpdateTarget(snapshot factors.Snapshot, policy *v1alpha1.ScalingPolicySpec) error {
//	e.mutex.Lock()
//	defer e.mutex.Unlock()
//
//	podSpec, err := scaling.ComputeResources(snapshot, policy)
//	if err != nil {
//		return err
//	}
//
//	glog.V(4).Infof("computed target values: %s", debug.Print(podSpec))
//
//	e.target = podSpec
//	return nil
//}
//
//func (e *NoopSmoothing) ComputeChange(parentPath string, current *v1.PodSpec) (bool, *v1.PodSpec) {
//	e.mutex.Lock()
//	defer e.mutex.Unlock()
//
//	podChanged := false
//	podChanges := new(v1.PodSpec)
//
//	if e.target == nil {
//		glog.V(2).Infof("target value %s not computed", parentPath)
//		return false, nil
//	}
//
//	for i := range e.target.Containers {
//		targetContainer := &e.target.Containers[i]
//		containerName := targetContainer.Name
//
//		path := parentPath + "." + containerName
//
//		var currentContainer *v1.Container
//		for i := range current.Containers {
//			c := &current.Containers[i]
//			if c.Name == containerName {
//				currentContainer = c
//				break
//			}
//		}
//
//		if currentContainer == nil {
//			glog.Warningf("ignoring policy for non-existent container %q", path)
//			continue
//		}
//
//		if changed, changes := e.updateContainer(path, currentContainer, targetContainer); changed {
//			podChanges.Containers = append(podChanges.Containers, *changes)
//			podChanged = true
//		}
//	}
//
//	return podChanged, podChanges
//
//}
//
//
//
//func (e *NoopSmoothing) Query() *http.Info {
//	e.mutex.Lock()
//	defer e.mutex.Unlock()
//
//	info := &http.Info{
//		LatestTarget: e.target,
//	}
//	return info
//}
