/*
Copyright 2025 The Kubernetes Authors.

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

	resourceapi "k8s.io/api/resource/v1beta1"
	drasnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/context"
)

func TestFilterOutNodesForGPULabledAndDra(t *testing.T) {
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

	nodeDraReady := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "node_1_Dra_Ready",
			CreationTimestamp: metav1.NewTime(start),
		},
		Status: apiv1.NodeStatus{
			Capacity:    apiv1.ResourceList{},
			Allocatable: apiv1.ResourceList{},
			Conditions:  []apiv1.NodeCondition{readyCondition},
		},
	}
	nodeDraReady.Status.Allocatable[NvidiaDriverName] = *resource.NewQuantity(1, resource.DecimalSI)
	nodeDraReady.Status.Allocatable[SomeOtherDriverName] = *resource.NewQuantity(1, resource.DecimalSI)
	nodeDraReady.Status.Capacity[NvidiaDriverName] = *resource.NewQuantity(1, resource.DecimalSI)
	nodeDraReady.Status.Capacity[SomeOtherDriverName] = *resource.NewQuantity(1, resource.DecimalSI)
	expectedReadiness[nodeDraReady.Name] = true

	nodeDraUnready := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "node_2_GPU_Unready",
			Labels:            gpuLabels,
			CreationTimestamp: metav1.NewTime(start),
		},
		Status: apiv1.NodeStatus{
			Capacity:    apiv1.ResourceList{},
			Allocatable: apiv1.ResourceList{},
			Conditions:  []apiv1.NodeCondition{readyCondition},
		},
	}

	nodeDraUnready.Status.Allocatable[NvidiaDriverName] = *resource.NewQuantity(0, resource.DecimalSI)
	nodeDraUnready.Status.Capacity[NvidiaDriverName] = *resource.NewQuantity(0, resource.DecimalSI)

	expectedReadiness[nodeDraUnready.Name] = false

	nodeDraUnready2 := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "node_3_Dra_Unready2",
			CreationTimestamp: metav1.NewTime(start),
		},
		Status: apiv1.NodeStatus{
			Conditions: []apiv1.NodeCondition{readyCondition},
		},
	}
	expectedReadiness[nodeDraUnready2.Name] = false

	nodeNoDraReady := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "node_4_NonDra_Ready",
			CreationTimestamp: metav1.NewTime(start),
		},
		Status: apiv1.NodeStatus{
			Conditions: []apiv1.NodeCondition{readyCondition},
		},
	}
	expectedReadiness[nodeNoDraReady.Name] = true

	nodeNoDraUnready := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "node_5_NonDra_Unready",
			CreationTimestamp: metav1.NewTime(start),
		},
		Status: apiv1.NodeStatus{
			Conditions: []apiv1.NodeCondition{unreadyCondition},
		},
	}
	expectedReadiness[nodeNoDraUnready.Name] = false

	initialReadyNodes := []*apiv1.Node{
		nodeDraReady,
		nodeDraUnready,
		nodeDraUnready2,
		nodeNoDraReady,
	}
	initialAllNodes := []*apiv1.Node{
		nodeDraReady,
		nodeDraUnready,
		nodeDraUnready2,
		nodeNoDraReady,
		nodeNoDraUnready,
	}

	localSlices := map[string][]*resourceapi.ResourceSlice{
		"node_1_Dra_Ready":    {node1Slice1, node1Slice2},
		"node_3_Dra_Unready2": {node3Slice1, node3Slice2},
	}

	processor := NewDefaultCustomResourcesProcessor(true)
	provider := testprovider.NewTestCloudProvider(nil, nil)
	snapshot := drasnapshot.NewSnapshot(nil, localSlices, nil, nil)

	ctx := &context.AutoscalingContext{CloudProvider: provider, DraSnapShot: snapshot}
	newAllNodes, newReadyNodes := processor.FilterOutNodesWithUnreadyResources(ctx, initialAllNodes, initialReadyNodes)

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
