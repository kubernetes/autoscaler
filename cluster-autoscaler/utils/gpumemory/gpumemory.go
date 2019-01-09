package gpumemory

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	// ResourceVisenzeGPUMemory is the name of the GPU Memory resource
	ResourceVisenzeGPUMemory = "visenze.com/nvidia-gpu-memory"
	// GPULabel is the label added to nodes with GPU resource by Visenze.
	// If you're not scaling - this is probably the problem!
	GPULabel = "accelerator"
)

// NodeHasGpuMemory returns true if a given node has GPU hardware
func NodeHasGpuMemory(node *apiv1.Node) bool {
	_, hasGpuLabel := node.Labels[GPULabel]
	gpuAllocatable, hasGpuAllocatable := node.Status.Allocatable[ResourceVisenzeGPUMemory]
	return hasGpuLabel || (hasGpuAllocatable && !gpuAllocatable.IsZero())
}

// PodRequestsGpuMemory returns true if a given pod has GPU Memory request
func PodRequestsGpuMemory(pod *apiv1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if container.Resources.Requests != nil {
			_, gpuMemoryFound := container.Resources.Requests[ResourceVisenzeGPUMemory]
			if gpuMemoryFound {
				return true
			}
		}
	}
	return false
}

// RequestInfo gives some information about hwo much GPU memory is needed
type RequestInfo struct {
	MaximumMemory resource.Quantity
	TotalMemory   resource.Quantity
	Pods          []*apiv1.Pod
}

// GetGPUMemoryRequests returns information about how much memory is needed for a set of pods
func GetGPUMemoryRequests(pods []*apiv1.Pod) *RequestInfo {
	ri := &RequestInfo{}
	for _, pod := range pods {
		var podGpuMemory resource.Quantity
		for _, container := range pod.Spec.Containers {
			if container.Resources.Requests != nil {
				containerGpuMemory := container.Resources.Requests[ResourceVisenzeGPUMemory]
				podGpuMemory.Add(containerGpuMemory)
			}
		}
		if podGpuMemory.Value() == 0 {
			continue
		}

		if podGpuMemory.Cmp(ri.MaximumMemory) > 0 {
			ri.MaximumMemory = podGpuMemory
		}
		ri.TotalMemory.Add(podGpuMemory)
		ri.Pods = append(ri.Pods, pod)
	}
	return ri
}
