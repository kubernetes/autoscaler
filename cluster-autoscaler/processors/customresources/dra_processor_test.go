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

	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/store"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/testsnapshot"
	drasnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	utils "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestFilterOutNodesWithUnreadyDRAResources(t *testing.T) {
	testCases := map[string]struct {
		nodeGroupsAllNodes        map[string][]*apiv1.Node
		nodeGroupsTemplatesSlices map[string][]*resourceapi.ResourceSlice
		nodesSlices               map[string][]*resourceapi.ResourceSlice
		expectedNodesReadiness    map[string]bool
	}{
		"1 DRA node group all totally ready": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1_Dra_Ready", true),
					buildTestNode("node_2_Dra_Ready", true),
				},
			},
			nodeGroupsTemplatesSlices: map[string][]*resourceapi.ResourceSlice{
				"ng1": createNodeResourceSlices("ng1_template", []int{1, 1}),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1_Dra_Ready": createNodeResourceSlices("node_1_Dra_Ready", []int{1, 1}),
				"node_2_Dra_Ready": createNodeResourceSlices("node_2_Dra_Ready", []int{1, 1}),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1_Dra_Ready": true,
				"node_2_Dra_Ready": true,
			},
		},
		"1 DRA node group, one initially unready": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1_Dra_Ready", true),
					buildTestNode("node_2_Dra_Ready", false),
				},
			},
			nodeGroupsTemplatesSlices: map[string][]*resourceapi.ResourceSlice{
				"ng1": createNodeResourceSlices("ng1_template", []int{1, 1}),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1_Dra_Ready": createNodeResourceSlices("node_1_Dra_Ready", []int{1, 1}),
				"node_2_Dra_Ready": createNodeResourceSlices("node_2_Dra_Ready", []int{1, 1}),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1_Dra_Ready": true,
				"node_2_Dra_Ready": false,
			},
		},
		"1 DRA node group, one initially ready with unready reasource": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1_Dra_Ready", true),
					buildTestNode("node_2_Dra_Ready", true),
				},
			},
			nodeGroupsTemplatesSlices: map[string][]*resourceapi.ResourceSlice{
				"ng1": createNodeResourceSlices("ng1_template", []int{1, 1}),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1_Dra_Ready": createNodeResourceSlices("node_1_Dra_Ready", []int{1, 1}),
				"node_2_Dra_Ready": createNodeResourceSlices("node_2_Dra_Ready", []int{1, 0}),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1_Dra_Ready": true,
				"node_2_Dra_Ready": false,
			},
		},
		"1 DRA node group, one initially ready with more reasources than expected": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1_Dra_Ready", true),
					buildTestNode("node_2_Dra_Ready", true),
				},
			},
			nodeGroupsTemplatesSlices: map[string][]*resourceapi.ResourceSlice{
				"ng1": createNodeResourceSlices("ng1_template", []int{1, 1}),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1_Dra_Ready": createNodeResourceSlices("node_1_Dra_Ready", []int{1, 1}),
				"node_2_Dra_Ready": createNodeResourceSlices("node_2_Dra_Ready", []int{1, 3}),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1_Dra_Ready": true,
				"node_2_Dra_Ready": false,
			},
		},
		"1 DRA node group, one initially ready with no slices": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1_Dra_Ready", true),
					buildTestNode("node_2_Dra_Ready", true),
				},
			},
			nodeGroupsTemplatesSlices: map[string][]*resourceapi.ResourceSlice{
				"ng1": createNodeResourceSlices("ng1_template", []int{1, 1}),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1_Dra_Ready": {},
				"node_2_Dra_Ready": createNodeResourceSlices("node_2_Dra_Ready", []int{1, 3}),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1_Dra_Ready": false,
				"node_2_Dra_Ready": false,
			},
		},
		"1 DRA node group, single driver multiple pools, only one published": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1_Dra_Ready", true),
				},
			},
			nodeGroupsTemplatesSlices: map[string][]*resourceapi.ResourceSlice{
				"ng1": buildNodeResourceSlices("ng1_template", "driver", []int{2, 2, 2}),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1_Dra_Ready": buildNodeResourceSlices("node_2_Dra_Ready", "driver", []int{2}),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1_Dra_Ready": false,
			},
		},
		"1 DRA node group, single driver multiple pools, more pools published including template pools": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_2_Dra_Ready", true),
				},
			},
			nodeGroupsTemplatesSlices: map[string][]*resourceapi.ResourceSlice{
				"ng1": buildNodeResourceSlices("ng1_template", "driver", []int{2, 2, 2}),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_2_Dra_Ready": buildNodeResourceSlices("node_2_Dra_Ready", "driver", []int{2, 2, 2, 2}),
			},
			expectedNodesReadiness: map[string]bool{
				"node_2_Dra_Ready": true,
			},
		},
		"1 DRA node group, single driver multiple pools, more pools published not including template pools": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1_Dra_Ready", true),
				},
			},
			nodeGroupsTemplatesSlices: map[string][]*resourceapi.ResourceSlice{
				"ng1": buildNodeResourceSlices("ng1_template", "driver", []int{2, 2, 2}),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1_Dra_Ready": buildNodeResourceSlices("node_1_Dra_Ready", "driver", []int{2, 2, 1, 2}),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1_Dra_Ready": false,
			},
		},
		"2 node groups, one DRA with 1 reasource unready node": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1_Dra_Ready", true),
					buildTestNode("node_2_Dra_Ready", true),
					buildTestNode("node_3_Dra_Unready", true),
				},
				"ng2": {
					buildTestNode("node_4_NonDra_Ready", true),
					buildTestNode("node_5_NonDra_Unready", false),
				},
			},
			nodeGroupsTemplatesSlices: map[string][]*resourceapi.ResourceSlice{
				"ng1": createNodeResourceSlices("ng1_template", []int{2, 2}),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1_Dra_Ready":   createNodeResourceSlices("node_1_Dra_Ready", []int{2, 2}),
				"node_2_Dra_Ready":   createNodeResourceSlices("node_2_Dra_Ready", []int{2, 2}),
				"node_3_Dra_Unready": createNodeResourceSlices("node_3_Dra_Unready", []int{2, 1}),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1_Dra_Ready":      true,
				"node_2_Dra_Ready":      true,
				"node_3_Dra_Unready":    false,
				"node_4_NonDra_Ready":   true,
				"node_5_NonDra_Unready": false,
			},
		},
		"2 DRA node groups, each with 1 reasource unready node": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1_Dra_Ready", true),
					buildTestNode("node_2_Dra_Ready", true),
					buildTestNode("node_3_Dra_Unready", true),
				},
				"ng2": {
					buildTestNode("node_4_Dra_Ready", true),
					buildTestNode("node_5_Dra_Unready", true),
				},
			},
			nodeGroupsTemplatesSlices: map[string][]*resourceapi.ResourceSlice{
				"ng1": createNodeResourceSlices("ng1_template", []int{2, 2}),
				"ng2": createNodeResourceSlices("ng2_template", []int{3, 3}),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1_Dra_Ready":   createNodeResourceSlices("node_1_Dra_Ready", []int{2, 2}),
				"node_2_Dra_Ready":   createNodeResourceSlices("node_2_Dra_Ready", []int{2, 2}),
				"node_3_Dra_Unready": createNodeResourceSlices("node_3_Dra_Unready", []int{2, 1}),
				"node_4_Dra_Ready":   createNodeResourceSlices("node_4_Dra_Ready", []int{3, 3}),
				"node_5_Dra_Unready": createNodeResourceSlices("node_5_Dra_Unready", []int{2, 1}),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1_Dra_Ready":   true,
				"node_2_Dra_Ready":   true,
				"node_3_Dra_Unready": false,
				"node_4_Dra_Ready":   true,
				"node_5_Dra_Unready": false,
			},
		},
		"2 DRA node group, single driver multiple pools, more pools published including template pools": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1_Dra_Ready", true),
					buildTestNode("node_2_Dra_Ready", true),
				},
				"ng2": {
					buildTestNode("node_3_Dra_Ready", true),
				},
			},
			nodeGroupsTemplatesSlices: map[string][]*resourceapi.ResourceSlice{
				"ng1": buildNodeResourceSlices("ng1_template", "driver", []int{2, 2, 2}),
				"ng2": buildNodeResourceSlices("ng2_template", "driver", []int{1, 1}),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1_Dra_Ready": buildNodeResourceSlices("node_1_Dra_Ready", "driver", []int{2, 2, 2, 2}),
				"node_2_Dra_Ready": buildNodeResourceSlices("node_2_Dra_Ready", "driver", []int{2, 2, 2}),
				"node_3_Dra_Ready": buildNodeResourceSlices("node_3_Dra_Ready", "driver", []int{1, 1, 1}),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1_Dra_Ready": true,
				"node_2_Dra_Ready": true,
				"node_3_Dra_Ready": true,
			},
		},
		"All together": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1", true),
					buildTestNode("node_2", true),
					buildTestNode("node_3", true),
				},
				"ng2": {
					buildTestNode("node_4", false),
					buildTestNode("node_5", true),
				},
				"ng3": {
					buildTestNode("node_6", false),
					buildTestNode("node_7", true),
				},
			},
			nodeGroupsTemplatesSlices: map[string][]*resourceapi.ResourceSlice{
				"ng1": createNodeResourceSlices("ng1_template", []int{2, 2}),
				"ng2": createNodeResourceSlices("ng2_template", []int{3, 3}),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1": createNodeResourceSlices("node_1", []int{2, 2, 2}),
				"node_2": createNodeResourceSlices("node_2", []int{1}),
				"node_3": createNodeResourceSlices("node_3", []int{1, 2}),
				"node_4": createNodeResourceSlices("node_4", []int{3, 3}),
				"node_5": {},
			},
			expectedNodesReadiness: map[string]bool{
				"node_1": true,
				"node_2": false,
				"node_3": false,
				"node_4": false,
				"node_5": false,
				"node_6": false,
				"node_7": true,
			},
		},
	}

	for tcName, tc := range testCases {
		t.Run(tcName, func(t *testing.T) {
			provider := testprovider.NewTestCloudProviderBuilder().Build()
			machineTemplates := map[string]*framework.NodeInfo{}
			initialAllNodes := []*apiv1.Node{}
			initialReadyNodes := []*apiv1.Node{}
			for ng, nodes := range tc.nodeGroupsAllNodes {
				machineName := fmt.Sprintf("%s_machine_template", ng)
				if rs, found := tc.nodeGroupsTemplatesSlices[ng]; found {
					machineTemplates[machineName] = framework.NewNodeInfo(buildTestNode(fmt.Sprintf("%s_template", ng), true), rs)
				} else {
					machineTemplates[machineName] = framework.NewTestNodeInfo(buildTestNode(fmt.Sprintf("%s_template", ng), true))
				}
				provider.AddAutoprovisionedNodeGroup(ng, 0, 20, len(nodes), machineName)
				for _, node := range nodes {
					initialAllNodes = append(initialAllNodes, node)
					if getNodeReadiness(node) {
						initialReadyNodes = append(initialReadyNodes, node)
					}
					provider.AddNode(ng, node)
				}
			}
			provider.SetMachineTemplates(machineTemplates)
			draSnapshot := drasnapshot.NewSnapshot(nil, tc.nodesSlices, nil, nil)
			clusterSnapshotStore := store.NewBasicSnapshotStore()
			clusterSnapshotStore.SetClusterState([]*apiv1.Node{}, []*apiv1.Pod{}, draSnapshot)
			clusterSnapshot, _, _ := testsnapshot.NewCustomTestSnapshotAndHandle(clusterSnapshotStore)

			ctx := &context.AutoscalingContext{CloudProvider: provider, ClusterSnapshot: clusterSnapshot}
			processor := DraCustomResourcesProcessor{}
			newAllNodes, newReadyNodes := processor.FilterOutNodesWithUnreadyResources(ctx, initialAllNodes, initialReadyNodes, draSnapshot)

			readyNodes := make(map[string]bool)
			for _, node := range newReadyNodes {
				readyNodes[node.Name] = true
			}

			assert.True(t, len(newAllNodes) == len(initialAllNodes), "Total number of nodes should not change")
			for _, node := range newAllNodes {
				gotReadiness := getNodeReadiness(node)
				assert.Equal(t, tc.expectedNodesReadiness[node.Name], gotReadiness)
				assert.Equal(t, gotReadiness, readyNodes[node.Name])
			}

		})
	}

}

func createNodeResourceSlices(nodeName string, numberOfDevicesInSlices []int) []*resourceapi.ResourceSlice {
	return buildNodeResourceSlices(nodeName, "", numberOfDevicesInSlices)
}

func buildNodeResourceSlices(nodeName, driverName string, numberOfDevicesInSlices []int) []*resourceapi.ResourceSlice {
	numberOfSlices := len(numberOfDevicesInSlices)
	resourceSlices := []*resourceapi.ResourceSlice{}
	for sliceIndex := range numberOfSlices {
		devices := []resourceapi.Device{}
		for deviceIndex := range numberOfDevicesInSlices[sliceIndex] {
			devices = append(devices, resourceapi.Device{Name: fmt.Sprintf("%d_%d", sliceIndex, deviceIndex)})
		}
		if driverName == "" {
			driverName = fmt.Sprintf("driver_%d", sliceIndex)
		}
		spec := resourceapi.ResourceSliceSpec{
			NodeName: &nodeName,
			Driver:   driverName,
			Pool:     resourceapi.ResourcePool{Name: fmt.Sprintf("%s_pool_%d", nodeName, sliceIndex)},
			Devices:  devices,
		}
		resourceSlices = append(resourceSlices, &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: nodeName}, Spec: spec})
	}
	return resourceSlices
}

func buildTestNode(nodeName string, ready bool) *apiv1.Node {
	node := utils.BuildTestNode(nodeName, 500, 100)
	utils.SetNodeReadyState(node, ready, time.Now().Add(-5*time.Minute))
	return node
}

func getNodeReadiness(node *apiv1.Node) bool {
	for i := range node.Status.Conditions {
		if node.Status.Conditions[i].Type == apiv1.NodeReady {
			return node.Status.Conditions[i].Status == apiv1.ConditionTrue
		}
	}
	return false
}
