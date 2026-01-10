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

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/store"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/testsnapshot"
	csisnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/csi/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

func TestFilterOutNodesWithUnreadyCSIResources(t *testing.T) {
	testCases := map[string]struct {
		nodeGroupsAllNodes         map[string][]*apiv1.Node
		nodeGroupsTemplatesCSINode map[string]*storagev1.CSINode
		nodesCSINode               map[string]*storagev1.CSINode
		csiSnapshot                *csisnapshot.Snapshot
		expectedNodesReadiness     map[string]bool
	}{
		"1 CSI node group all totally ready": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1_CSI_Ready", true),
					buildTestNode("node_2_CSI_Ready", true),
				},
			},
			nodeGroupsTemplatesCSINode: map[string]*storagev1.CSINode{
				"ng1": createCSINode("ng1_template", []string{"driver1", "driver2"}),
			},
			nodesCSINode: map[string]*storagev1.CSINode{
				"node_1_CSI_Ready": createCSINode("node_1_CSI_Ready", []string{"driver1", "driver2"}),
				"node_2_CSI_Ready": createCSINode("node_2_CSI_Ready", []string{"driver1", "driver2"}),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1_CSI_Ready": true,
				"node_2_CSI_Ready": true,
			},
		},
		"1 CSI node group, one initially unready": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1_CSI_Ready", true),
					buildTestNode("node_2_CSI_Ready", false),
				},
			},
			nodeGroupsTemplatesCSINode: map[string]*storagev1.CSINode{
				"ng1": createCSINode("ng1_template", []string{"driver1", "driver2"}),
			},
			nodesCSINode: map[string]*storagev1.CSINode{
				"node_1_CSI_Ready": createCSINode("node_1_CSI_Ready", []string{"driver1", "driver2"}),
				"node_2_CSI_Ready": createCSINode("node_2_CSI_Ready", []string{"driver1", "driver2"}),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1_CSI_Ready": true,
				"node_2_CSI_Ready": false,
			},
		},
		"1 CSI node group, one initially ready with missing driver": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1_CSI_Ready", true),
					buildTestNode("node_2_CSI_Unready", true),
				},
			},
			nodeGroupsTemplatesCSINode: map[string]*storagev1.CSINode{
				"ng1": createCSINode("ng1_template", []string{"driver1", "driver2"}),
			},
			nodesCSINode: map[string]*storagev1.CSINode{
				"node_1_CSI_Ready":   createCSINode("node_1_CSI_Ready", []string{"driver1", "driver2"}),
				"node_2_CSI_Unready": createCSINode("node_2_CSI_Unready", []string{"driver1"}),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1_CSI_Ready":   true,
				"node_2_CSI_Unready": false,
			},
		},
		"1 CSI node group, one initially ready with no drivers": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1_CSI_Ready", true),
					buildTestNode("node_2_CSI_Unready", true),
				},
			},
			nodeGroupsTemplatesCSINode: map[string]*storagev1.CSINode{
				"ng1": createCSINode("ng1_template", []string{"driver1", "driver2"}),
			},
			nodesCSINode: map[string]*storagev1.CSINode{
				"node_1_CSI_Ready":   createCSINode("node_1_CSI_Ready", []string{"driver1", "driver2"}),
				"node_2_CSI_Unready": createCSINode("node_2_CSI_Unready", []string{}),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1_CSI_Ready":   true,
				"node_2_CSI_Unready": false,
			},
		},
		"1 CSI node group, one initially ready with extra drivers": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1_CSI_Ready", true),
					buildTestNode("node_2_CSI_Ready", true),
				},
			},
			nodeGroupsTemplatesCSINode: map[string]*storagev1.CSINode{
				"ng1": createCSINode("ng1_template", []string{"driver1", "driver2"}),
			},
			nodesCSINode: map[string]*storagev1.CSINode{
				"node_1_CSI_Ready": createCSINode("node_1_CSI_Ready", []string{"driver1", "driver2", "driver3"}),
				"node_2_CSI_Ready": createCSINode("node_2_CSI_Ready", []string{"driver1", "driver2"}),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1_CSI_Ready": true,
				"node_2_CSI_Ready": true,
			},
		},
		"2 node groups, one CSI with 1 driver unready node": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1_CSI_Ready", true),
					buildTestNode("node_2_CSI_Ready", true),
					buildTestNode("node_3_CSI_Unready", true),
				},
				"ng2": {
					buildTestNode("node_4_NonCSI_Ready", true),
					buildTestNode("node_5_NonCSI_Unready", false),
				},
			},
			nodeGroupsTemplatesCSINode: map[string]*storagev1.CSINode{
				"ng1": createCSINode("ng1_template", []string{"driver1", "driver2"}),
			},
			nodesCSINode: map[string]*storagev1.CSINode{
				"node_1_CSI_Ready":   createCSINode("node_1_CSI_Ready", []string{"driver1", "driver2"}),
				"node_2_CSI_Ready":   createCSINode("node_2_CSI_Ready", []string{"driver1", "driver2"}),
				"node_3_CSI_Unready": createCSINode("node_3_CSI_Unready", []string{"driver1"}),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1_CSI_Ready":      true,
				"node_2_CSI_Ready":      true,
				"node_3_CSI_Unready":    false,
				"node_4_NonCSI_Ready":   true,
				"node_5_NonCSI_Unready": false,
			},
		},
		"2 CSI node groups, each with 1 driver unready node": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1_CSI_Ready", true),
					buildTestNode("node_2_CSI_Ready", true),
					buildTestNode("node_3_CSI_Unready", true),
				},
				"ng2": {
					buildTestNode("node_4_CSI_Ready", true),
					buildTestNode("node_5_CSI_Unready", true),
				},
			},
			nodeGroupsTemplatesCSINode: map[string]*storagev1.CSINode{
				"ng1": createCSINode("ng1_template", []string{"driver1", "driver2"}),
				"ng2": createCSINode("ng2_template", []string{"driver3", "driver4"}),
			},
			nodesCSINode: map[string]*storagev1.CSINode{
				"node_1_CSI_Ready":   createCSINode("node_1_CSI_Ready", []string{"driver1", "driver2"}),
				"node_2_CSI_Ready":   createCSINode("node_2_CSI_Ready", []string{"driver1", "driver2"}),
				"node_3_CSI_Unready": createCSINode("node_3_CSI_Unready", []string{"driver1"}),
				"node_4_CSI_Ready":   createCSINode("node_4_CSI_Ready", []string{"driver3", "driver4"}),
				"node_5_CSI_Unready": createCSINode("node_5_CSI_Unready", []string{"driver3"}),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1_CSI_Ready":   true,
				"node_2_CSI_Ready":   true,
				"node_3_CSI_Unready": false,
				"node_4_CSI_Ready":   true,
				"node_5_CSI_Unready": false,
			},
		},
		"nil CSI snapshot returns original nodes": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1_CSI_Ready", true),
					buildTestNode("node_2_CSI_Ready", true),
				},
			},
			nodeGroupsTemplatesCSINode: map[string]*storagev1.CSINode{
				"ng1": createCSINode("ng1_template", []string{"driver1", "driver2"}),
			},
			nodesCSINode: map[string]*storagev1.CSINode{},
			csiSnapshot:  nil,
			expectedNodesReadiness: map[string]bool{
				"node_1_CSI_Ready": true,
				"node_2_CSI_Ready": true,
			},
		},
		"node missing from CSI snapshot stays ready": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1_CSI_Ready", true),
					buildTestNode("node_2_CSI_Ready", true),
				},
			},
			nodeGroupsTemplatesCSINode: map[string]*storagev1.CSINode{
				"ng1": createCSINode("ng1_template", []string{"driver1", "driver2"}),
			},
			nodesCSINode: map[string]*storagev1.CSINode{
				"node_1_CSI_Ready": createCSINode("node_1_CSI_Ready", []string{"driver1", "driver2"}),
				// node_2_CSI_Ready is missing from snapshot
			},
			expectedNodesReadiness: map[string]bool{
				"node_1_CSI_Ready": true,
				"node_2_CSI_Ready": true, // stays ready because error getting CSI node keeps it in ready list
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
			nodeGroupsTemplatesCSINode: map[string]*storagev1.CSINode{
				"ng1": createCSINode("ng1_template", []string{"driver1", "driver2"}),
				"ng2": createCSINode("ng2_template", []string{"driver3"}),
			},
			nodesCSINode: map[string]*storagev1.CSINode{
				"node_1": createCSINode("node_1", []string{"driver1", "driver2"}),
				"node_2": createCSINode("node_2", []string{"driver1"}),
				"node_3": createCSINode("node_3", []string{"driver1", "driver2", "driver3"}),
				"node_4": createCSINode("node_4", []string{"driver3"}),
				"node_5": createCSINode("node_5", []string{}),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1": true,
				"node_2": false,
				"node_3": true,
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
				if csiNode, found := tc.nodeGroupsTemplatesCSINode[ng]; found {
					machineTemplates[machineName] = framework.NewNodeInfo(buildTestNode(fmt.Sprintf("%s_template", ng), true), nil).
						SetCSINode(csiNode)
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

			var csiSnapshot *csisnapshot.Snapshot
			if tc.csiSnapshot == nil && len(tc.nodesCSINode) > 0 {
				csiSnapshot = csisnapshot.NewSnapshot(tc.nodesCSINode)
			} else {
				csiSnapshot = tc.csiSnapshot
			}

			clusterSnapshotStore := store.NewBasicSnapshotStore()
			clusterSnapshotStore.SetClusterState([]*apiv1.Node{}, []*apiv1.Pod{}, nil, csiSnapshot)
			clusterSnapshot, _, _ := testsnapshot.NewCustomTestSnapshotAndHandle(clusterSnapshotStore)

			autoscalingCtx := &ca_context.AutoscalingContext{CloudProvider: provider, ClusterSnapshot: clusterSnapshot}
			processor := CSICustomResourcesProcessor{}
			newAllNodes, newReadyNodes := processor.FilterOutNodesWithUnreadyResources(autoscalingCtx, initialAllNodes, initialReadyNodes, nil, csiSnapshot)

			readyNodes := make(map[string]bool)
			for _, node := range newReadyNodes {
				readyNodes[node.Name] = true
			}

			assert.True(t, len(newAllNodes) == len(initialAllNodes), "Total number of nodes should not change")
			for _, node := range newAllNodes {
				gotReadiness := getNodeReadiness(node)
				assert.Equal(t, tc.expectedNodesReadiness[node.Name], gotReadiness, "Node %s readiness mismatch", node.Name)
				assert.Equal(t, gotReadiness, readyNodes[node.Name], "Node %s readiness consistency check", node.Name)
			}
		})
	}
}

func createCSINode(nodeName string, driverNames []string) *storagev1.CSINode {
	drivers := make([]storagev1.CSINodeDriver, 0, len(driverNames))
	for _, driverName := range driverNames {
		drivers = append(drivers, storagev1.CSINodeDriver{
			Name:   driverName,
			NodeID: fmt.Sprintf("%s-%s", nodeName, driverName),
		})
	}
	return &storagev1.CSINode{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
		},
		Spec: storagev1.CSINodeSpec{
			Drivers: drivers,
		},
	}
}
