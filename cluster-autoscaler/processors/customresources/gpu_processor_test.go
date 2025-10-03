/*
Copyright 2021 The Kubernetes Authors.

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

package customresources

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
)

const (
	GPULabel = "TestGPULabel/accelerator"
)

var (
	gpuLabels = map[string]string{
		GPULabel: "nvidia-tesla-k80",
	}
)

func TestFilterOutNodesWithUnreadyResources(t *testing.T) {
	start := time.Now()
	later := start.Add(10 * time.Minute)
	expectedReadiness := make(map[string]bool)
	originalConditions := make(map[string]int)
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
	// Any non-zero resource quantity value on a Ready node indicates GPU is ready.
	resourceQuantityOne := *resource.NewQuantity(1, resource.DecimalSI)
	// A zero resource quantity value on a Ready node indicates GPU is unready.
	resourceQuantityZero := *resource.NewQuantity(0, resource.DecimalSI)

	// Here is a set of Ready nodes that do not have any status conditions
	// that would indicate they should be filtered out for GPU (or other device) allocation.
	readyNodesForResourceAllocation := []*apiv1.Node{
		// This node has the sufficient GPU labels and GPU status conditions to
		// be considered for device allocation.
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "nodeGpuReady",
				Labels:            gpuLabels,
				CreationTimestamp: metav1.NewTime(start),
			},
			Status: apiv1.NodeStatus{
				Capacity: apiv1.ResourceList{
					gpu.ResourceNvidiaGPU: resourceQuantityOne,
				},
				Allocatable: apiv1.ResourceList{
					gpu.ResourceNvidiaGPU: resourceQuantityOne,
				},
				Conditions: []apiv1.NodeCondition{readyCondition},
			},
		},
		// This node has the sufficient GPU labels and DirectX status conditions to
		// be considered for device allocation.
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "nodeDirectXReady",
				Labels:            gpuLabels,
				CreationTimestamp: metav1.NewTime(start),
			},
			Status: apiv1.NodeStatus{
				Capacity: apiv1.ResourceList{
					gpu.ResourceDirectX: resourceQuantityOne,
				},
				Allocatable: apiv1.ResourceList{
					gpu.ResourceDirectX: resourceQuantityOne,
				},
				Conditions: []apiv1.NodeCondition{readyCondition},
			},
		},
	}
	// Here we add a vanilla Ready node (no GPU or other device labels or status conditions)
	// to ensure that our filter function *does not* filter out Ready nodes.
	readyNodesForResourceAllocation = append(readyNodesForResourceAllocation, &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "nodeVanillaReady",
			Labels:            make(map[string]string),
			CreationTimestamp: metav1.NewTime(start),
		},
		Status: apiv1.NodeStatus{
			Conditions: []apiv1.NodeCondition{readyCondition},
		},
	})

	// Here is a set of nodes that have status conditions that *do not* satisfy
	// GPU (or other device) allocation requirements. They should be filtered out
	// of the list of nodes considered for scheduling devices.
	notReadyNodesForResourceAllocation := []*apiv1.Node{
		// This node has zero allocatable + capacity resource values for GPU,
		// so it should be filtered out of the list of nodes considered for resource
		// allocation, even though it is a Ready node.
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "nodeGpuUnready",
				Labels:            gpuLabels,
				CreationTimestamp: metav1.NewTime(start),
			},
			Status: apiv1.NodeStatus{
				Capacity: apiv1.ResourceList{
					gpu.ResourceNvidiaGPU: resourceQuantityZero,
				},
				Allocatable: apiv1.ResourceList{
					gpu.ResourceNvidiaGPU: resourceQuantityZero,
				},
				Conditions: []apiv1.NodeCondition{readyCondition},
			},
		},
		// This node has zero allocatable + capacity resource values for DirectX,
		// so it should be filtered out of the list of nodes considered for resource
		// allocation, even though it is a Ready node.
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "nodeDirectXUnready",
				Labels:            gpuLabels,
				CreationTimestamp: metav1.NewTime(start),
			},
			Status: apiv1.NodeStatus{
				Capacity: apiv1.ResourceList{
					gpu.ResourceDirectX: resourceQuantityZero,
				},
				Allocatable: apiv1.ResourceList{
					gpu.ResourceDirectX: resourceQuantityZero,
				},
				Conditions: []apiv1.NodeCondition{readyCondition},
			},
		},
		// This node does not have any allocatable + capacity status reported,
		// so it should be filtered out of the list of nodes considered for resource
		// allocation, even though it is a Ready node.
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "nodeGpuUnready2",
				Labels:            gpuLabels,
				CreationTimestamp: metav1.NewTime(start),
			},
			Status: apiv1.NodeStatus{
				Conditions: []apiv1.NodeCondition{readyCondition},
			},
		},
	}
	// Here we add a vanilla NotReady node (no GPU or other device labels or status conditions)
	// to ensure that our filter function *does* filter out NotReady nodes.
	notReadyNodesForResourceAllocation = append(notReadyNodesForResourceAllocation, &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "nodeVanillaNotReady",
			Labels:            make(map[string]string),
			CreationTimestamp: metav1.NewTime(start),
		},
		Status: apiv1.NodeStatus{
			Conditions: []apiv1.NodeCondition{unreadyCondition},
		},
	})

	for _, n := range readyNodesForResourceAllocation {
		expectedReadiness[n.Name] = true
		numConditions := 0
		for _, condition := range n.Status.Conditions {
			if condition.Type != apiv1.NodeReady {
				numConditions++
			}
		}
		originalConditions[n.Name] = numConditions
	}
	for _, n := range notReadyNodesForResourceAllocation {
		expectedReadiness[n.Name] = false
		numConditions := 0
		for _, condition := range n.Status.Conditions {
			if condition.Type != apiv1.NodeReady {
				numConditions++
			}
		}
		originalConditions[n.Name] = numConditions
	}

	var initialAllNodes, initialReadyNodes []*apiv1.Node
	for _, n := range readyNodesForResourceAllocation {
		initialAllNodes = append(initialAllNodes, n)
		initialReadyNodes = append(initialReadyNodes, n)
	}
	for _, n := range notReadyNodesForResourceAllocation {
		initialAllNodes = append(initialAllNodes, n)
		for _, condition := range n.Status.Conditions {
			if condition.Type == apiv1.NodeReady && condition.Status == apiv1.ConditionTrue {
				initialReadyNodes = append(initialReadyNodes, n)
			}
		}
	}

	processor := GpuCustomResourcesProcessor{}
	provider := testprovider.NewTestCloudProviderBuilder().Build()
	autoscalingCtx := &ca_context.AutoscalingContext{CloudProvider: provider}
	newAllNodes, newReadyNodes := processor.FilterOutNodesWithUnreadyResources(autoscalingCtx, initialAllNodes, initialReadyNodes, nil)

	foundInReady := make(map[string]bool)
	for _, n := range newReadyNodes {
		foundInReady[n.Name] = true
		assert.True(t, expectedReadiness[n.Name], fmt.Sprintf("Node %s found in ready nodes list (it shouldn't be there)", n.Name))
	}
	for nodeName, expected := range expectedReadiness {
		if expected {
			assert.True(t, foundInReady[nodeName], fmt.Sprintf("Node %s expected ready, but not found in ready nodes list", nodeName))
		}
	}
	for _, n := range newAllNodes {
		// GetUnreadyNodeCopy called by FilterOutNodesWithUnreadyResources always adds
		// one Ready=false condition if it's not already set.
		assert.Equal(t, len(n.Status.Conditions), originalConditions[n.Name]+1)
		if expectedReadiness[n.Name] {
			for _, condition := range n.Status.Conditions {
				if condition.Type != apiv1.NodeReady {
					assert.Equal(t, condition.Status, apiv1.ConditionTrue, fmt.Sprintf("Unexpected ready condition value for node %s", n.Name))
				}
			}
		} else {
			for _, condition := range n.Status.Conditions {
				if condition.Type != apiv1.NodeReady {
					assert.Equal(t, n.Status.Conditions[0].Status, apiv1.ConditionFalse, fmt.Sprintf("Unexpected ready condition value for node %s", n.Name))
				}
			}
		}
	}
}

// TestFilterOutNodesWithUnreadyResourcesDRA tests that FilterOutNodesWithUnreadyResources
// does the right thing based on DRA configuration present in the node.
func TestFilterOutNodesWithUnreadyResourcesDRA(t *testing.T) {
	start := time.Now()
	later := start.Add(10 * time.Minute)
	expectedReadinessDRA := make(map[string]bool)
	originalConditionsDRA := make(map[string]int)
	expectedReadiness := make(map[string]bool)
	originalConditions := make(map[string]int)

	readyCondition := apiv1.NodeCondition{
		Type:               apiv1.NodeReady,
		Status:             apiv1.ConditionTrue,
		LastTransitionTime: metav1.NewTime(later),
	}

	// Here is a set of Ready Nodes
	readyNodes := []*apiv1.Node{
		// This is a GPU-labeled Ready node with no GPU allocatable or capacity.
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "nodeGPUReady",
				Labels:            gpuLabels,
				CreationTimestamp: metav1.NewTime(start),
			},
			Status: apiv1.NodeStatus{
				Conditions: []apiv1.NodeCondition{readyCondition},
			},
		},
		// This is a vanilla ready node
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "nodeVanillaReady",
				Labels:            make(map[string]string),
				CreationTimestamp: metav1.NewTime(start),
			},
			Status: apiv1.NodeStatus{
				Conditions: []apiv1.NodeCondition{readyCondition},
			},
		},
	}

	for _, n := range readyNodes {
		// We are going to wire all of our ready nodes up w/ a scenario that includes a DRA driver
		expectedReadinessDRA[n.Name] = true
		// We are also going to create a non-DRA driver scenario
		if n.Labels[GPULabel] == "" {
			expectedReadiness[n.Name] = true
		} else {
			expectedReadiness[n.Name] = false
		}
		numConditions := 0
		for _, condition := range n.Status.Conditions {
			if condition.Type != apiv1.NodeReady {
				numConditions++
			}
		}
		originalConditionsDRA[n.Name] = numConditions
		originalConditions[n.Name] = numConditions
	}

	var initialAllNodes, initialReadyNodes []*apiv1.Node
	for _, n := range readyNodes {
		initialAllNodes = append(initialAllNodes, n)
		initialReadyNodes = append(initialReadyNodes, n)
	}

	// DRA-driver node GPU config
	nodeGpuConfigDRA := func(node *apiv1.Node) *cloudprovider.GpuConfig {
		return &cloudprovider.GpuConfig{
			ExtendedResourceName: "",
			DraDriverName:        "gpu.nvidia.com",
		}
	}
	// non-DRA-driver node GPU config
	nodeGpuConfig := func(node *apiv1.Node) *cloudprovider.GpuConfig {
		return &cloudprovider.GpuConfig{
			DraDriverName: "",
		}
	}
	providerDRA := testprovider.NewTestCloudProviderBuilder().WithNodeGpuConfig(nodeGpuConfigDRA).Build()
	provider := testprovider.NewTestCloudProviderBuilder().WithNodeGpuConfig(nodeGpuConfig).Build()
	autoscalingContextDRANodes := &context.AutoscalingContext{CloudProvider: providerDRA}
	autoscalingContext := &context.AutoscalingContext{CloudProvider: provider}
	processor := GpuCustomResourcesProcessor{}
	newAllNodesDRA, newReadyNodesDRA := processor.FilterOutNodesWithUnreadyResources(autoscalingContextDRANodes, initialAllNodes, initialReadyNodes, nil)
	newAllNodes, newReadyNodes := processor.FilterOutNodesWithUnreadyResources(autoscalingContext, initialAllNodes, initialReadyNodes, nil)

	foundInReadyDRA := make(map[string]bool)
	for _, n := range newReadyNodesDRA {
		foundInReadyDRA[n.Name] = true
		assert.True(t, expectedReadinessDRA[n.Name], fmt.Sprintf("Node %s found in ready nodes list (it shouldn't be there)", n.Name))
	}
	for nodeName, expected := range expectedReadinessDRA {
		if expected {
			assert.True(t, foundInReadyDRA[nodeName], fmt.Sprintf("Node %s expected ready, but not found in ready nodes list", nodeName))
		}
	}
	for _, n := range newAllNodesDRA {
		// GetUnreadyNodeCopy called by FilterOutNodesWithUnreadyResources always adds
		// one Ready=false condition if it's not already set.
		assert.Equal(t, len(n.Status.Conditions), originalConditionsDRA[n.Name]+1)
		if expectedReadinessDRA[n.Name] {
			for _, condition := range n.Status.Conditions {
				if condition.Type != apiv1.NodeReady {
					assert.Equal(t, condition.Status, apiv1.ConditionTrue, fmt.Sprintf("Unexpected ready condition value for node %s", n.Name))
				}
			}
		} else {
			for _, condition := range n.Status.Conditions {
				if condition.Type != apiv1.NodeReady {
					assert.Equal(t, n.Status.Conditions[0].Status, apiv1.ConditionFalse, fmt.Sprintf("Unexpected ready condition value for node %s", n.Name))
				}
			}
		}
	}

	foundInReady := make(map[string]bool)
	for _, n := range newReadyNodes {
		foundInReady[n.Name] = true
		assert.True(t, expectedReadiness[n.Name], fmt.Sprintf("Node %s found in ready nodes list (it shouldn't be there)", n.Name))
	}
	for nodeName, expected := range expectedReadiness {
		if expected {
			assert.True(t, foundInReady[nodeName], fmt.Sprintf("Node %s expected ready, but not found in ready nodes list", nodeName))
		}
	}
	for _, n := range newAllNodes {
		// GetUnreadyNodeCopy called by FilterOutNodesWithUnreadyResources always adds
		// one Ready=false condition if it's not already set.
		assert.Equal(t, len(n.Status.Conditions), originalConditions[n.Name]+1)
		if expectedReadiness[n.Name] {
			for _, condition := range n.Status.Conditions {
				if condition.Type != apiv1.NodeReady {
					assert.Equal(t, condition.Status, apiv1.ConditionTrue, fmt.Sprintf("Unexpected ready condition value for node %s", n.Name))
				}
			}
		} else {
			for _, condition := range n.Status.Conditions {
				if condition.Type != apiv1.NodeReady {
					assert.Equal(t, n.Status.Conditions[0].Status, apiv1.ConditionFalse, fmt.Sprintf("Unexpected ready condition value for node %s", n.Name))
				}
			}
		}
	}
}
