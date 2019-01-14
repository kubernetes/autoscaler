package gpumemory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNodeHasGpuMemory(t *testing.T) {
	gpuLabels := map[string]string{
		GPULabel: "nvidia-tesla-k80",
	}
	nodeGpuReady := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "nodeGpuReady",
			Labels: gpuLabels,
		},
		Status: apiv1.NodeStatus{
			Capacity:    apiv1.ResourceList{},
			Allocatable: apiv1.ResourceList{},
		},
	}
	nodeGpuReady.Status.Allocatable[ResourceVisenzeGPUMemory] = *resource.NewQuantity(8e9, resource.DecimalSI)
	nodeGpuReady.Status.Capacity[ResourceVisenzeGPUMemory] = *resource.NewQuantity(8e9, resource.DecimalSI)
	assert.True(t, NodeHasGpuMemory(nodeGpuReady))

	nodeGpuUnready := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "nodeGpuUnready",
			Labels: gpuLabels,
		},
		Status: apiv1.NodeStatus{
			Capacity:    apiv1.ResourceList{},
			Allocatable: apiv1.ResourceList{},
		},
	}
	assert.True(t, NodeHasGpuMemory(nodeGpuUnready))

	nodeNoGpu := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "nodeNoGpu",
			Labels: map[string]string{},
		},
		Status: apiv1.NodeStatus{
			Capacity:    apiv1.ResourceList{},
			Allocatable: apiv1.ResourceList{},
		},
	}
	assert.False(t, NodeHasGpuMemory(nodeNoGpu))
}

func TestPodRequestsGpuMemory(t *testing.T) {
	podNoGpu := &apiv1.Pod{
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				apiv1.Container{
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceCPU: *resource.NewQuantity(1, resource.DecimalSI),
						},
					},
				},
			},
		},
	}
	podWithGpu := &apiv1.Pod{Spec: apiv1.PodSpec{Containers: []apiv1.Container{
		apiv1.Container{
			Resources: apiv1.ResourceRequirements{
				Requests: apiv1.ResourceList{
					apiv1.ResourceCPU:        *resource.NewQuantity(1, resource.DecimalSI),
					ResourceVisenzeGPUMemory: *resource.NewQuantity(1, resource.DecimalSI),
				},
			},
		},
	}}}
	podWithGpu.Spec.Containers[0].Resources.Requests[ResourceVisenzeGPUMemory] = *resource.NewQuantity(1, resource.DecimalSI)

	assert.False(t, PodRequestsGpuMemory(podNoGpu))
	assert.True(t, PodRequestsGpuMemory(podWithGpu))
}
