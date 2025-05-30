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
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/store"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/testsnapshot"
	drasnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/context"
)

var (
	node1Slice1 = createNodeResourceSlice("node1", "local-slice-1", "driver_X", "pool", []string{"device_A", "device_B"})
	node1Slice2 = createNodeResourceSlice("node1", "local-slice-1", "driver_Y", "pool", []string{"device_A", "device_C"})
	node2Slice1 = createNodeResourceSlice("node2", "local-slice-2", "driver_Y", "pool", []string{"device_A", "device_C"})
	node2Slice2 = createNodeResourceSlice("node2", "local-slice-2", "driver_X", "pool", []string{"device_A", "device_B"})
	node3Slice1 = createNodeResourceSlice("node3", "local-slice-3", "driver_X", "pool", []string{"device_A", "device_B"})
	node3Slice2 = createNodeResourceSlice("node3", "local-slice-3", "driver_Y", "pool", []string{"device_A", "device_B"})
)

func createNodeResourceSlice(nodeName, id, driverName, poolName string, devicesNames []string) *resourceapi.ResourceSlice {
	devices := []resourceapi.Device{}
	for _, name := range devicesNames {
		devices = append(devices, resourceapi.Device{Name: name})
	}
	spec := resourceapi.ResourceSliceSpec{NodeName: nodeName, Driver: driverName, Pool: resourceapi.ResourcePool{Name: poolName}, Devices: devices}
	return &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: id}, Spec: spec}
}
func TestFilterOutNodesWithUnreadyDRAResources(t *testing.T) {
	start := time.Now()
	later := start.Add(10 * time.Minute)
	expectedReadiness := make(map[string]bool)

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

	node1DraReady := &apiv1.Node{
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
	expectedReadiness[node1DraReady.Name] = true

	node2DraReady := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "node_2_Dra_Ready",
			CreationTimestamp: metav1.NewTime(start),
		},
		Status: apiv1.NodeStatus{
			Capacity:    apiv1.ResourceList{},
			Allocatable: apiv1.ResourceList{},
			Conditions:  []apiv1.NodeCondition{readyCondition},
		},
	}
	expectedReadiness[node2DraReady.Name] = true

	node3DraUnready := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "node_3_Dra_Unready",
			CreationTimestamp: metav1.NewTime(start),
		},
		Status: apiv1.NodeStatus{
			Capacity:    apiv1.ResourceList{},
			Allocatable: apiv1.ResourceList{},
			Conditions:  []apiv1.NodeCondition{readyCondition},
		},
	}
	expectedReadiness[node3DraUnready.Name] = false

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
		node1DraReady,
		node2DraReady,
		node3DraUnready,
		nodeNoDraReady,
	}
	initialAllNodes := []*apiv1.Node{
		node1DraReady,
		node2DraReady,
		node3DraUnready,
		nodeNoDraReady,
		nodeNoDraUnready,
	}

	localSlices := map[string][]*resourceapi.ResourceSlice{
		"node_1_Dra_Ready":   {node1Slice1, node1Slice2},
		"node_2_Dra_Ready":   {node2Slice1, node2Slice2},
		"node_3_Dra_Unready": {node3Slice1, node3Slice2},
	}

	provider := testprovider.NewTestCloudProviderBuilder().Build()
	provider.SetMachineTemplates(map[string]*framework.NodeInfo{"m1": framework.NewNodeInfo(node1DraReady, []*resourceapi.ResourceSlice{node1Slice1, node1Slice2}), "m2": framework.NewTestNodeInfo(nodeNoDraReady)})
	provider.AddAutoprovisionedNodeGroup("ng1", 0, 10, 2, "m1")
	provider.AddNode("ng1", node1DraReady)
	provider.AddNode("ng1", node2DraReady)
	provider.AddNode("ng1", node3DraUnready)
	provider.AddAutoprovisionedNodeGroup("ng2", 0, 10, 2, "m2")
	provider.AddNode("ng2", nodeNoDraReady)
	provider.AddNode("ng2", nodeNoDraUnready)

	draSnapshot := drasnapshot.NewSnapshot(nil, localSlices, nil, nil)
	clusterSnapshotStore := store.NewBasicSnapshotStore()
	clusterSnapshotStore.SetClusterState([]*apiv1.Node{}, []*apiv1.Pod{}, draSnapshot)
	clusterSnapshot, _, _ := testsnapshot.NewCustomTestSnapshotAndHandle(clusterSnapshotStore)

	ctx := &context.AutoscalingContext{CloudProvider: provider, ClusterSnapshot: clusterSnapshot}
	processor := DraCustomResourcesProcessor{}
	newAllNodes, newReadyNodes := processor.FilterOutNodesWithUnreadyResources(ctx, initialAllNodes, initialReadyNodes)

	readyNodes := make(map[string]bool)
	for _, node := range newReadyNodes {
		readyNodes[node.Name] = true
	}

	assert.True(t, len(newAllNodes) == len(initialAllNodes), "Total number of nodes should not chane")
	assert.True(t, len(newReadyNodes) <= len(initialReadyNodes), "Processor should not add ready nodes")
	for _, node := range newAllNodes {
		assert.Equal(t, len(node.Status.Conditions), 1)
		if expectedReadiness[node.Name] {
			assert.True(t, readyNodes[node.Name], fmt.Sprintf("Node %s should be listed in ready nodes", node.Name))
			assert.Equal(t, node.Status.Conditions[0].Status, apiv1.ConditionTrue, fmt.Sprintf("Unexpected ready condition value for node %s", node.Name))
		} else {
			assert.False(t, readyNodes[node.Name], fmt.Sprintf("Node %s should not be listed in ready nodes", node.Name))
			assert.Equal(t, node.Status.Conditions[0].Status, apiv1.ConditionFalse, fmt.Sprintf("Unexpected ready condition value for node %s", node.Name))
		}
	}
}
