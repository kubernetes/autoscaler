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
	"fmt"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
)

func TestFilterOutNodesWithUnreadyGpus(t *testing.T) {
	start := time.Now()
	later := start.Add(10 * time.Minute)
	expectedReadiness := make(map[string]bool)
	gpuLabels := map[string]string{
		GPULabel: "nvidia-tesla-k80",
	}
	readyCondition := apiv1.NodeCondition{
		Type:               apiv1.NodeReady,
		Status:             apiv1.ConditionTrue,
		LastTransitionTime: metav1.NewTime(later),
	}
	unreadyCondition := apiv1.NodeCondition{
		Type:               apiv1.NodeReady,
		Status:             apiv1.ConditionFalse,
		LastTransitionTime: metav1.NewTime(later),
	}

	nodeGpuReady := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "nodeGpuReady",
			Labels:            gpuLabels,
			CreationTimestamp: metav1.NewTime(start),
		},
		Status: apiv1.NodeStatus{
			Capacity:    apiv1.ResourceList{},
			Allocatable: apiv1.ResourceList{},
			Conditions:  []apiv1.NodeCondition{readyCondition},
		},
	}
	nodeGpuReady.Status.Allocatable[ResourceNvidiaGPU] = *resource.NewQuantity(1, resource.DecimalSI)
	nodeGpuReady.Status.Capacity[ResourceNvidiaGPU] = *resource.NewQuantity(1, resource.DecimalSI)
	expectedReadiness[nodeGpuReady.Name] = true

	nodeGpuUnready := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "nodeGpuUnready",
			Labels:            gpuLabels,
			CreationTimestamp: metav1.NewTime(start),
		},
		Status: apiv1.NodeStatus{
			Capacity:    apiv1.ResourceList{},
			Allocatable: apiv1.ResourceList{},
			Conditions:  []apiv1.NodeCondition{readyCondition},
		},
	}
	nodeGpuUnready.Status.Allocatable[ResourceNvidiaGPU] = *resource.NewQuantity(0, resource.DecimalSI)
	nodeGpuUnready.Status.Capacity[ResourceNvidiaGPU] = *resource.NewQuantity(0, resource.DecimalSI)
	expectedReadiness[nodeGpuUnready.Name] = false

	nodeGpuUnready2 := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "nodeGpuUnready2",
			Labels:            gpuLabels,
			CreationTimestamp: metav1.NewTime(start),
		},
		Status: apiv1.NodeStatus{
			Conditions: []apiv1.NodeCondition{readyCondition},
		},
	}
	expectedReadiness[nodeGpuUnready2.Name] = false

	nodeNoGpuReady := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "nodeNoGpuReady",
			Labels:            make(map[string]string),
			CreationTimestamp: metav1.NewTime(start),
		},
		Status: apiv1.NodeStatus{
			Conditions: []apiv1.NodeCondition{readyCondition},
		},
	}
	expectedReadiness[nodeNoGpuReady.Name] = true

	nodeNoGpuUnready := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "nodeNoGpuUnready",
			Labels:            make(map[string]string),
			CreationTimestamp: metav1.NewTime(start),
		},
		Status: apiv1.NodeStatus{
			Conditions: []apiv1.NodeCondition{unreadyCondition},
		},
	}
	expectedReadiness[nodeNoGpuUnready.Name] = false

	initialReadyNodes := []*apiv1.Node{
		nodeGpuReady,
		nodeGpuUnready,
		nodeGpuUnready2,
		nodeNoGpuReady,
	}
	initialAllNodes := []*apiv1.Node{
		nodeGpuReady,
		nodeGpuUnready,
		nodeGpuUnready2,
		nodeNoGpuReady,
		nodeNoGpuUnready,
	}

	newAllNodes, newReadyNodes := FilterOutNodesWithUnreadyGpus(initialAllNodes, initialReadyNodes)

	foundInReady := make(map[string]bool)
	for _, node := range newReadyNodes {
		foundInReady[node.Name] = true
		assert.True(t, expectedReadiness[node.Name], fmt.Sprintf("Node %s found in ready nodes list (it shouldn't be there)", node.Name))
	}
	for nodeName, expected := range expectedReadiness {
		if expected {
			assert.True(t, foundInReady[nodeName], fmt.Sprintf("Node %s expected ready, but not found in ready nodes list", nodeName))
		}
	}
	for _, node := range newAllNodes {
		assert.Equal(t, len(node.Status.Conditions), 1)
		if expectedReadiness[node.Name] {
			assert.Equal(t, node.Status.Conditions[0].Status, apiv1.ConditionTrue, fmt.Sprintf("Unexpected ready condition value for node %s", node.Name))
		} else {
			assert.Equal(t, node.Status.Conditions[0].Status, apiv1.ConditionFalse, fmt.Sprintf("Unexpected ready condition value for node %s", node.Name))
		}
	}
}

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
	assert.True(t, NodeHasGpu(nodeGpuReady))

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
	assert.True(t, NodeHasGpu(nodeGpuUnready))

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
	assert.False(t, NodeHasGpu(nodeNoGpu))
}
