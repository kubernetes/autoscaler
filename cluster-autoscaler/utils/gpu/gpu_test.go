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

package gpu_test

import (
	"testing"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
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
	nodeGpuReady.Status.Allocatable[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(1, resource.DecimalSI)
	nodeGpuReady.Status.Capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(1, resource.DecimalSI)
	assert.True(t, gpu.NodeHasGpu(GPULabel, nodeGpuReady))

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
	assert.True(t, gpu.NodeHasGpu(GPULabel, nodeGpuUnready))

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
	assert.False(t, gpu.NodeHasGpu(GPULabel, nodeNoGpu))
}

func TestPodRequestsGpu(t *testing.T) {
	podNoGpu := test.BuildTestPod("podNoGpu", 0, 1000)
	podWithGpu := test.BuildTestPod("pod1AnyGpu", 0, 1000)
	podWithGpu.Spec.Containers[0].Resources.Requests[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(1, resource.DecimalSI)

	assert.False(t, gpu.PodRequestsGpu(podNoGpu))
	assert.True(t, gpu.PodRequestsGpu(podWithGpu))
}

func TestGetGpuInfoForMetrics(t *testing.T) {
	knownGpu := "nvidia-tesla-k80"
	unknownGpu := "unknown-gpu"
	availableGPUTypes := map[string]struct{}{
		knownGpu: {},
	}
	resourceName := apiv1.ResourceName(gpu.ResourceNvidiaGPU)

	// Basic node
	node := test.BuildTestNode("node", 1000, 1000)

	// Node with GPU capacity
	nodeWithGpu := test.BuildTestNode("node-with-gpu", 1000, 1000)
	nodeWithGpu.Status.Capacity[resourceName] = *resource.NewQuantity(1, resource.DecimalSI)

	// Node without GPU capacity
	nodeWithoutGpu := test.BuildTestNode("node-without-gpu", 1000, 1000)

	// Node group with GPU in template
	provider := testprovider.TestCloudProvider{}
	templateWithGpu := test.BuildTestNode("template-with-gpu", 1000, 1000)
	templateWithGpu.Status.Capacity[resourceName] = *resource.NewQuantity(1, resource.DecimalSI)
	nodeGroupWithGpu := provider.BuildNodeGroup("ng-with-gpu", 1, 10, 1, false, false, "n1-standard-1", nil)

	// Node group without GPU in template
	templateWithoutGpu := test.BuildTestNode("template-without-gpu", 1000, 1000)
	nodeGroupWithoutGpu := provider.BuildNodeGroup("ng-without-gpu", 1, 10, 1, false, false, "n1-standard-1", nil)

	templates := map[string]*framework.NodeInfo{
		nodeGroupWithoutGpu.Id(): framework.NewNodeInfo(templateWithoutGpu, nil),
		nodeGroupWithGpu.Id():    framework.NewNodeInfo(templateWithGpu, nil),
	}

	provider.SetMachineTemplates(templates)

	testCases := []struct {
		name                string
		gpuConfig           *cloudprovider.GpuConfig
		node                *apiv1.Node
		nodeGroup           cloudprovider.NodeGroup
		expectedGpuResource string
		expectedGpuType     string
	}{
		{
			name:                "no gpu config",
			gpuConfig:           nil,
			node:                node,
			nodeGroup:           nil,
			expectedGpuResource: "",
			expectedGpuType:     gpu.MetricsNoGPU,
		},
		{
			name: "generic gpu",
			gpuConfig: &cloudprovider.GpuConfig{
				Type:                 "",
				ExtendedResourceName: resourceName,
			},
			node:                nodeWithGpu,
			nodeGroup:           nil,
			expectedGpuResource: gpu.ResourceNvidiaGPU,
			expectedGpuType:     gpu.MetricsGenericGPU,
		},
		{
			name: "dra gpu, known type",
			gpuConfig: &cloudprovider.GpuConfig{
				Type:          knownGpu,
				DraDriverName: "test-driver",
			},
			node:                nodeWithoutGpu,
			nodeGroup:           nil,
			expectedGpuResource: "dra_test-driver",
			expectedGpuType:     knownGpu,
		},
		{
			name: "dra gpu, unknown type",
			gpuConfig: &cloudprovider.GpuConfig{
				Type:          unknownGpu,
				DraDriverName: "test-driver",
			},
			node:                nodeWithoutGpu,
			nodeGroup:           nil,
			expectedGpuResource: "dra_test-driver",
			expectedGpuType:     gpu.MetricsUnknownGPU,
		},
		{
			name: "capacity present, known type",
			gpuConfig: &cloudprovider.GpuConfig{
				Type:                 knownGpu,
				ExtendedResourceName: resourceName,
			},
			node:                nodeWithGpu,
			nodeGroup:           nil,
			expectedGpuResource: gpu.ResourceNvidiaGPU,
			expectedGpuType:     knownGpu,
		},
		{
			name: "capacity present, unknown type",
			gpuConfig: &cloudprovider.GpuConfig{
				Type:                 unknownGpu,
				ExtendedResourceName: resourceName,
			},
			node:                nodeWithGpu,
			nodeGroup:           nil,
			expectedGpuResource: gpu.ResourceNvidiaGPU,
			expectedGpuType:     gpu.MetricsUnknownGPU,
		},
		{
			name: "no capacity, template has gpu",
			gpuConfig: &cloudprovider.GpuConfig{
				Type:                 knownGpu,
				ExtendedResourceName: resourceName,
			},
			node:                nodeWithoutGpu,
			nodeGroup:           nodeGroupWithGpu,
			expectedGpuResource: gpu.ResourceNvidiaGPU,
			expectedGpuType:     gpu.MetricsMissingGPU,
		},
		{
			name: "no capacity, template has no gpu",
			gpuConfig: &cloudprovider.GpuConfig{
				Type:                 knownGpu,
				ExtendedResourceName: resourceName,
			},
			node:                nodeWithoutGpu,
			nodeGroup:           nodeGroupWithoutGpu,
			expectedGpuResource: gpu.ResourceNvidiaGPU,
			expectedGpuType:     gpu.MetricsUnexpectedLabelGPU,
		},
		{
			name: "no capacity, no node group",
			gpuConfig: &cloudprovider.GpuConfig{
				Type:                 knownGpu,
				ExtendedResourceName: resourceName,
			},
			node:                nodeWithoutGpu,
			nodeGroup:           nil,
			expectedGpuResource: gpu.ResourceNvidiaGPU,
			expectedGpuType:     gpu.MetricsUnexpectedLabelGPU,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gpuResource, gpuType := gpu.GetGpuInfoForMetrics(tc.gpuConfig, availableGPUTypes, tc.node, tc.nodeGroup)
			assert.Equal(t, tc.expectedGpuResource, gpuResource)
			assert.Equal(t, tc.expectedGpuType, gpuType)
		})
	}
}
