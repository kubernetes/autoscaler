/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gpu

import (
	"testing"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
)

const (
	GPULabel = "TestGPULabel/accelerator"
)

func TestNodeHasGpu(t *testing.T) {
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
	nodeGpuReady.Status.Allocatable[ResourceNvidiaGPU] = *resource.NewQuantity(1, resource.DecimalSI)
	nodeGpuReady.Status.Capacity[ResourceNvidiaGPU] = *resource.NewQuantity(1, resource.DecimalSI)
	assert.True(t, NodeHasGpu(GPULabel, nodeGpuReady))

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
	assert.True(t, NodeHasGpu(GPULabel, nodeGpuUnready))

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
	assert.False(t, NodeHasGpu(GPULabel, nodeNoGpu))
}

func TestPodRequestsGpu(t *testing.T) {
	podNoGpu := test.BuildTestPod("podNoGpu", 0, 1000)
	podWithGpu := test.BuildTestPod("pod1AnyGpu", 0, 1000)
	podWithGpu.Spec.Containers[0].Resources.Requests[ResourceNvidiaGPU] = *resource.NewQuantity(1, resource.DecimalSI)

	assert.False(t, PodRequestsGpu(podNoGpu))
	assert.True(t, PodRequestsGpu(podWithGpu))
}
