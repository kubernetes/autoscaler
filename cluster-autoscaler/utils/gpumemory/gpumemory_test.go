package gpumemory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
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
	podNoGpu := test.BuildTestPod("podNoGpu", 0, 1000)
	podWithGpu := test.BuildTestPod("pod1AnyGpu", 0, 1000)
	podWithGpu.Spec.Containers[0].Resources.Requests[ResourceVisenzeGPUMemory] = *resource.NewQuantity(1, resource.DecimalSI)

	assert.False(t, PodRequestsGpuMemory(podNoGpu))
	assert.True(t, PodRequestsGpuMemory(podWithGpu))
}

func TestGetGPUMemoryRequests(t *testing.T) {
	podNoGpu := test.BuildTestPod("podNoGpu", 0, 1000)
	podNoGpu.Spec.NodeSelector = map[string]string{}

	pod1AnyGpu := test.BuildTestPod("pod1AnyGpu", 0, 1000)
	pod1AnyGpu.Spec.NodeSelector = map[string]string{}
	pod1AnyGpu.Spec.Containers[0].Resources.Requests[ResourceVisenzeGPUMemory] = *resource.NewQuantity(4e9, resource.DecimalSI)

	pod2DefaultGpu := test.BuildTestPod("pod2DefaultGpu", 0, 1000)
	pod2DefaultGpu.Spec.NodeSelector = map[string]string{GPULabel: "visenze.com/gpu"}
	pod2DefaultGpu.Spec.Containers[0].Resources.Requests[ResourceVisenzeGPUMemory] = *resource.NewQuantity(8e9, resource.DecimalSI)

	requestInfo := GetGPUMemoryRequests([]*apiv1.Pod{podNoGpu})
	assert.Equal(t, int64(0), requestInfo.TotalMemory.Value(), "non empty gpu request calculated for pods without gpu")

	requestInfo = GetGPUMemoryRequests([]*apiv1.Pod{podNoGpu, pod1AnyGpu})
	assert.Equal(t, int64(4e9), requestInfo.TotalMemory.Value(), "expected a single gpu request for a single pod requesting gpu")
	assert.Equal(t, int64(4e9), requestInfo.MaximumMemory.Value(), "expected a single gpu request for a single pod requesting gpu")
	assert.Equal(t, len(requestInfo.Pods), 1)
	assert.Equal(t, requestInfo.Pods[0], pod1AnyGpu)

	requestInfo = GetGPUMemoryRequests([]*apiv1.Pod{podNoGpu, pod1AnyGpu, pod2DefaultGpu})
	assert.Equal(t, int64(12e9), requestInfo.TotalMemory.Value(), "Expected total memory request to be 12GB")
	assert.Equal(t, int64(8e9), requestInfo.MaximumMemory.Value())
	assert.Equal(t, len(requestInfo.Pods), 2)
}
