/*
Copyright The Kubernetes Authors.

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

package coreweave

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ca_context "sigs.k8s.io/cluster-autoscaler/pkg/context"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors/customresources"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/gpu"
)

// A GPU node can register Ready before its device plugin has advertised
// nvidia.com/gpu in allocatable. During that window the node must be
// classified as unready (ResourceUnready): otherwise the core autoscaler
// treats it as a usable node with zero GPUs, still considers the pending GPU
// pods unschedulable, and scales the node group up again, over-provisioning
// the pool. The GPU processor keys that classification off the presence of
// the presence of the provider's GPULabel() key — see the GPULabel comment
// for why that key is the driver-version label rather than
// gpu.nvidia.com/class.
func TestFilterOutNodesWithUnreadyGpu(t *testing.T) {
	readyCondition := []apiv1.NodeCondition{
		{Type: apiv1.NodeReady, Status: apiv1.ConditionTrue},
	}

	bootingGpuNode := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gpu-node-booting",
			Labels: map[string]string{
				GPULabel:               "595.71.05",
				"gpu.nvidia.com/class": "H100_NVLINK_80GB",
			},
		},
		Status: apiv1.NodeStatus{
			// Device plugin not up yet: no nvidia.com/gpu in allocatable.
			Conditions: readyCondition,
		},
	}

	runningGpuNode := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gpu-node-running",
			Labels: map[string]string{
				GPULabel:               "595.71.05",
				"gpu.nvidia.com/class": "H100_NVLINK_80GB",
			},
		},
		Status: apiv1.NodeStatus{
			Allocatable: apiv1.ResourceList{
				gpu.ResourceNvidiaGPU: *resource.NewQuantity(8, resource.DecimalSI),
			},
			Conditions: readyCondition,
		},
	}

	// CPU nodes carry empty-valued gpu.nvidia.com/* keys but not the
	// driver-version label; they must not be classified GPU-unready.
	cpuNode := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cpu-node",
			Labels: map[string]string{
				"gpu.nvidia.com/class": "",
			},
		},
		Status: apiv1.NodeStatus{
			Conditions: readyCondition,
		},
	}

	allNodes := []*apiv1.Node{bootingGpuNode, runningGpuNode, cpuNode}
	autoscalingCtx := &ca_context.AutoscalingContext{CloudProvider: &CoreWeaveCloudProvider{}}
	processor := customresources.GpuCustomResourcesProcessor{}

	_, readyNodes := processor.FilterOutNodesWithUnreadyResources(
		context.Background(), autoscalingCtx, allNodes, allNodes, nil, nil)

	readyNames := make([]string, 0, len(readyNodes))
	for _, n := range readyNodes {
		readyNames = append(readyNames, n.Name)
	}
	assert.NotContains(t, readyNames, bootingGpuNode.Name,
		"GPU node without allocatable nvidia.com/gpu must be filtered out of ready nodes until the device plugin reports in")
	assert.Contains(t, readyNames, runningGpuNode.Name)
	assert.Contains(t, readyNames, cpuNode.Name,
		"CPU node with an empty-valued GPU label key must stay ready")
}

func TestGetNodeGpuConfig(t *testing.T) {
	provider := &CoreWeaveCloudProvider{}

	gpuNode := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gpu-node",
			Labels: map[string]string{
				GPULabel:               "595.71.05",
				"gpu.nvidia.com/class": "H100_NVLINK_80GB",
			},
		},
	}
	config := provider.GetNodeGpuConfig(context.Background(), gpuNode)
	assert.NotNil(t, config)
	assert.Equal(t, GPULabel, config.Label)
	assert.Equal(t, apiv1.ResourceName(gpu.ResourceNvidiaGPU), config.ExtendedResourceName)

	cpuNode := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "cpu-node",
			Labels: map[string]string{"gpu.nvidia.com/class": ""},
		},
	}
	assert.Nil(t, provider.GetNodeGpuConfig(context.Background(), cpuNode))
}
