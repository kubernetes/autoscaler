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
	"slices"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	resourceapi "k8s.io/api/resource/v1"
	v1 "k8s.io/api/resource/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/store"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/testsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/comparator"
	drasnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	utils "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

type noOpMetricsEmitter struct{}

func (m noOpMetricsEmitter) SetNodeTemplateResourcesMismatch(driver string, mismatchType metrics.ResourceMismatchType, value uint32) {
}

type mockTemplateNodeInfoRegistry struct {
	nodeInfos map[string]*framework.NodeInfo
}

func newMockTemplateNodeInfoRegistry(nodeInfos map[string]*framework.NodeInfo) *mockTemplateNodeInfoRegistry {
	return &mockTemplateNodeInfoRegistry{
		nodeInfos: nodeInfos,
	}
}

func (m *mockTemplateNodeInfoRegistry) GetNodeInfo(id string) (*framework.NodeInfo, bool) {
	nodeInfo, found := m.nodeInfos[id]
	return nodeInfo, found
}

func (m *mockTemplateNodeInfoRegistry) GetNodeInfos() map[string]*framework.NodeInfo {
	return m.nodeInfos
}

func (m *mockTemplateNodeInfoRegistry) Recompute(_ *ca_context.AutoscalingContext, _ []*apiv1.Node, _ []*appsv1.DaemonSet, _ taints.TaintConfig, _ time.Time) errors.AutoscalerError {
	return nil
}

func TestFilterOutNodesWithUnreadyDRAResources(t *testing.T) {
	testCases := map[string]struct {
		nodeGroupsAllNodes        map[string][]*apiv1.Node
		nodeGroupsTemplatesSlices map[string][]*resourceapi.ResourceSlice
		nodesSlices               map[string][]*resourceapi.ResourceSlice
		expectedNodesReadiness    map[string]bool
		registryNodeInfos         map[string]*framework.NodeInfo
	}{
		"NonDraNodes_Unaffected": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1", true),
					buildTestNode("node_2", false),
				},
			},
			registryNodeInfos: map[string]*framework.NodeInfo{
				"ng1": createTemplateNodeInfo("ng1_template", []*v1.ResourceSlice{}),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{},
			expectedNodesReadiness: map[string]bool{
				"node_1": true,
				"node_2": false,
			},
		},
		"UnreadyNodeWithPools_NotBecomingReady": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1", false),
				},
			},
			registryNodeInfos: map[string]*framework.NodeInfo{
				"ng1": createTemplateNodeInfo("ng1_template", buildResourceSlices("ng1_template", "driver", 1)),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1": buildResourceSlices("node_1", "driver", 1),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1": false,
			},
		},
		"PoolComplete_Unaffected": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1", true),
				},
			},
			registryNodeInfos: map[string]*framework.NodeInfo{
				"ng1": createTemplateNodeInfo("ng1_template", buildResourceSlices("ng1_template", "driver", 1)),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1": buildResourceSlices("node_1", "driver", 1),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1": true,
			},
		},
		"MultipleDriversPoolsComplete_Unaffected": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1", true),
				},
			},
			registryNodeInfos: map[string]*framework.NodeInfo{
				"ng1": createTemplateNodeInfo(
					"ng1_template",
					slices.Concat(
						buildResourceSlices("ng1_template", "driver1", 5),
						buildResourceSlices("ng1_template", "driver2", 9),
					),
				),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1": slices.Concat(
					buildResourceSlices("node_1", "driver1", 5),
					buildResourceSlices("node_1", "driver2", 9),
				),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1": true,
			},
		},
		"MoreReasourcePoolsAvailableThanInTemplate_Unaffected": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1", true),
				},
			},
			registryNodeInfos: map[string]*framework.NodeInfo{
				"ng1": createTemplateNodeInfo(
					"ng1_template",
					slices.Concat(
						buildResourceSlices("ng1_template", "driver1", 1),
						buildResourceSlices("ng1_template", "driver2", 1),
					),
				),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1": slices.Concat(
					buildResourceSlices("node_1", "driver1", 10),
					buildResourceSlices("node_1", "driver2", 10),
				),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1": true,
			},
		},
		"IncompletePoolFromUntrackedDriver_Unaffected": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1", true),
				},
			},
			registryNodeInfos: map[string]*framework.NodeInfo{
				"ng1": createTemplateNodeInfo(
					"ng1_template",
					slices.Concat(
						buildResourceSlices("ng1_template", "driver1", 1),
						buildResourceSlices("ng1_template", "driver2", 1),
					),
				),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1": slices.Concat(
					buildResourceSlices("node_1", "driver1", 10),
					buildResourceSlices("node_1", "driver2", 10),
					buildIncompleteResourceSlices("node_1", "driver3", 1),
				),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1": true,
			},
		},
		"CompletePoolFromUntrackedDriver_Unaffected": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1", true),
				},
			},
			registryNodeInfos: map[string]*framework.NodeInfo{
				"ng1": createTemplateNodeInfo(
					"ng1_template",
					slices.Concat(
						buildResourceSlices("ng1_template", "driver1", 1),
						buildResourceSlices("ng1_template", "driver2", 1),
					),
				),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1": slices.Concat(
					buildResourceSlices("node_1", "driver1", 10),
					buildResourceSlices("node_1", "driver2", 10),
					buildResourceSlices("node_1", "driver3", 1),
				),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1": true,
			},
		},
		"ResourcePoolIncomplete_MarkedUnready": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1", true),
				},
			},
			registryNodeInfos: map[string]*framework.NodeInfo{
				"ng1": createTemplateNodeInfo(
					"ng1_template",
					buildResourceSlices("ng1_template", "driver1", 1),
				),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1": buildIncompleteResourceSlices("node_1", "driver1", 1),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1": false,
			},
		},
		"MultipleDriversPoolsIncomplete_MarkedUnready": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1", true),
				},
			},
			registryNodeInfos: map[string]*framework.NodeInfo{
				"ng1": createTemplateNodeInfo(
					"ng1_template",
					slices.Concat(
						buildResourceSlices("ng1_template", "driver1", 5),
						buildResourceSlices("ng1_template", "driver2", 3),
					),
				),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1": slices.Concat(
					buildIncompleteResourceSlices("node_1", "driver1", 5),
					buildIncompleteResourceSlices("node_1", "driver2", 3),
				),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1": false,
			},
		},
		"OneDriverPoolIncomplete_MarkedUnready": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1", true),
				},
			},
			registryNodeInfos: map[string]*framework.NodeInfo{
				"ng1": createTemplateNodeInfo(
					"ng1_template",
					slices.Concat(
						buildResourceSlices("ng1_template", "driver1", 5),
						buildResourceSlices("ng1_template", "driver2", 3),
					),
				),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1": slices.Concat(
					buildResourceSlices("node_1", "driver1", 5),
					buildIncompleteResourceSlices("node_1", "driver2", 3),
				),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1": false,
			},
		},
		"ResourcePoolBothGenerationsNotReady_MarkedUnready": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1", true),
				},
			},
			registryNodeInfos: map[string]*framework.NodeInfo{
				"ng1": createTemplateNodeInfo(
					"ng1_template",
					buildResourceSlices("ng1_template", "driver1", 1),
				),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1": slices.Concat(
					buildResourcePoolWithSplitGeneration("node_1", "driver1", 1, map[int]int{1: 1, 2: 1}),
				),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1": false,
			},
		},
		"ResourcePoolOldGenerationReady_MarkedUnready": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1", true),
				},
			},
			registryNodeInfos: map[string]*framework.NodeInfo{
				"ng1": createTemplateNodeInfo(
					"ng1_template",
					buildResourceSlices("ng1_template", "driver1", 1),
				),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1": slices.Concat(
					buildResourcePoolWithSplitGeneration("node_1", "driver1", 1, map[int]int{1: 0, 2: 1}),
				),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1": false,
			},
		},
		"ResourcePoolNewGenerationComplete_Unaffected": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1", true),
				},
			},
			registryNodeInfos: map[string]*framework.NodeInfo{
				"ng1": createTemplateNodeInfo(
					"ng1_template",
					buildResourceSlices("ng1_template", "driver1", 1),
				),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1": slices.Concat(
					buildResourcePoolWithSplitGeneration("node_1", "driver1", 1, map[int]int{1: 1, 2: 0}),
				),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1": true,
			},
		},
		"ResourcePoolBothGenerationsReady_Unaffected": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1", true),
				},
			},
			registryNodeInfos: map[string]*framework.NodeInfo{
				"ng1": createTemplateNodeInfo(
					"ng1_template",
					buildResourceSlices("ng1_template", "driver1", 1),
				),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1": slices.Concat(
					buildResourcePoolWithSplitGeneration("node_1", "driver1", 1, map[int]int{1: 0, 2: 0}),
				),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1": true,
			},
		},
		"MissingResourcePool_MarkedUnready": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1", true),
				},
			},
			registryNodeInfos: map[string]*framework.NodeInfo{
				"ng1": createTemplateNodeInfo(
					"ng1_template",
					buildResourceSlices("ng1_template", "driver1", 5),
				),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{},
			expectedNodesReadiness: map[string]bool{
				"node_1": false,
			},
		},
		"FallBackToNodeGroupTemplateWhenRegistryIsMissing": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1", true),
				},
			},
			nodeGroupsTemplatesSlices: map[string][]*resourceapi.ResourceSlice{
				"ng1": buildResourceSlices("ng1_template", "driver", 1),
			},
			registryNodeInfos: map[string]*framework.NodeInfo{},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1": buildResourceSlices("node_1", "driver", 1),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1": true,
			},
		},
		"RegistryPreferredOverNodeGroupTemplate": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1", true),
				},
			},
			nodeGroupsTemplatesSlices: map[string][]*resourceapi.ResourceSlice{
				"ng1": buildResourceSlices("ng1_template", "driver", 10),
			},
			registryNodeInfos: map[string]*framework.NodeInfo{
				"ng1": createTemplateNodeInfo("ng1_template", buildResourceSlices("ng1_template", "driver", 9)),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1": buildResourceSlices("node_1", "driver", 9),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1": true,
			},
		},
		"MultipleNodeGroups_AllNodesReady": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1", true),
					buildTestNode("node_11", true),
				},
				"ng2": {
					buildTestNode("node_2", true),
					buildTestNode("node_22", true),
				},
			},
			registryNodeInfos: map[string]*framework.NodeInfo{
				"ng1": createTemplateNodeInfo("ng1_template", buildResourceSlices("ng1_template", "driver", 1)),
				"ng2": createTemplateNodeInfo("ng2_template", buildResourceSlices("ng2_template", "driver", 1)),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1":  buildResourceSlices("node_1", "driver", 1),
				"node_11": buildResourceSlices("node_11", "driver", 2),
				"node_2":  buildResourceSlices("node_2", "driver", 1),
				"node_22": buildResourceSlices("node_22", "driver", 2),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1":  true,
				"node_11": true,
				"node_2":  true,
				"node_22": true,
			},
		},
		"MultipleNodeGroups_OneNodeResourcePoolUnready": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"ng1": {
					buildTestNode("node_1", true),
				},
				"ng2": {
					buildTestNode("node_2", true),
				},
				"ng3": {
					buildTestNode("node_3", true),
				},
			},
			registryNodeInfos: map[string]*framework.NodeInfo{
				"ng1": createTemplateNodeInfo("ng1_template", buildResourceSlices("ng1_template", "driver", 1)),
				"ng2": createTemplateNodeInfo("ng2_template", buildResourceSlices("ng2_template", "driver", 1)),
				"ng3": createTemplateNodeInfo("ng3_template", buildResourceSlices("ng3_template", "driver", 1)),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"node_1": buildResourceSlices("node_1", "driver", 1),
				"node_2": buildResourceSlices("node_2", "driver", 1),
			},
			expectedNodesReadiness: map[string]bool{
				"node_1": true,
				"node_2": true,
				"node_3": false,
			},
		},
		"AllInOne": {
			nodeGroupsAllNodes: map[string][]*apiv1.Node{
				"unready_ng":              {buildTestNode("unready", false)},
				"split_generation_ng":     {buildTestNode("split_generation", true)},
				"pools_ready_ng":          {buildTestNode("pools_ready", true)},
				"pool_incomplete_ng":      {buildTestNode("pool_incomplete", true)},
				"missing_pool_ng":         {buildTestNode("missing_pool", true)},
				"multiple_drivers_ng":     {buildTestNode("multiple_drivers", true)},
				"extra_resource_pools_ng": {buildTestNode("extra_resource_pools", true)},
			},
			registryNodeInfos: map[string]*framework.NodeInfo{
				"unready_ng":          createTemplateNodeInfo("unready_ng_template", buildResourceSlices("unready_ng_template", "driver", 1)),
				"split_generation_ng": createTemplateNodeInfo("split_generation_ng_template", buildResourceSlices("split_generation_ng_template", "driver", 1)),
				"pools_ready_ng":      createTemplateNodeInfo("pools_ready_ng_template", buildResourceSlices("pools_ready_ng_template", "driver", 1)),
				"pool_incomplete_ng":  createTemplateNodeInfo("pool_incomplete_ng_template", buildResourceSlices("pool_incomplete_ng_template", "driver", 1)),
				"missing_pool_ng":     createTemplateNodeInfo("missing_pool_ng_template", buildResourceSlices("missing_pool_ng_template", "driver", 1)),
				"multiple_drivers_ng": createTemplateNodeInfo("multiple_drivers_ng_template", slices.Concat(
					buildResourceSlices("multiple_drivers_ng_template", "driver", 1),
					buildResourceSlices("multiple_drivers_ng_template", "other_driver", 1),
				)),
				"extra_resource_pools_ng": createTemplateNodeInfo("extra_resource_pools_ng_template", buildResourceSlices("extra_resource_pools_ng_template", "driver", 1)),
			},
			nodesSlices: map[string][]*resourceapi.ResourceSlice{
				"unready":          buildResourceSlices("unready", "driver", 1),
				"split_generation": buildResourcePoolWithSplitGeneration("split_generation", "driver", 1, map[int]int{1: 1, 2: 7}),
				"pools_ready":      buildResourceSlices("pools_ready", "driver", 1),
				"pool_incomplete":  buildIncompleteResourceSlices("pool_incomplete", "driver", 1),
				"multiple_drivers": slices.Concat(
					buildResourceSlices("multiple_drivers", "driver", 1),
					buildResourceSlices("multiple_drivers", "other_driver", 1),
				),
				"extra_resource_pools": slices.Concat(
					buildResourceSlices("extra_resource_pools", "driver", 1),
					buildResourceSlices("extra_resource_pools", "other_driver", 1),
				),
			},
			expectedNodesReadiness: map[string]bool{
				"unready":              false,
				"split_generation":     false,
				"pools_ready":          true,
				"pool_incomplete":      false,
				"missing_pool":         false,
				"multiple_drivers":     true,
				"extra_resource_pools": true,
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
			clusterSnapshot, _, _ := testsnapshot.NewCustomTestSnapshotAndHandle(clusterSnapshotStore)
			clusterSnapshot.SetClusterState([]*apiv1.Node{}, []*apiv1.Pod{}, draSnapshot, nil)

			autoscalingCtx := &ca_context.AutoscalingContext{
				CloudProvider:            provider,
				ClusterSnapshot:          clusterSnapshot,
				TemplateNodeInfoRegistry: newMockTemplateNodeInfoRegistry(tc.registryNodeInfos),
			}
			processor := DraCustomResourcesProcessor{resourcesComparator: comparator.NewNodeResourcesComparator(noOpMetricsEmitter{})}
			newAllNodes, newReadyNodes := processor.FilterOutNodesWithUnreadyResources(autoscalingCtx, initialAllNodes, initialReadyNodes, draSnapshot, nil)

			readyNodes := make(map[string]bool)
			for _, node := range newReadyNodes {
				readyNodes[node.Name] = true
			}

			assert.True(t, len(newAllNodes) == len(initialAllNodes), "Total number of nodes should not change")
			for _, node := range newAllNodes {
				gotReadiness := getNodeReadiness(node)
				assert.Equal(t, tc.expectedNodesReadiness[node.Name], gotReadiness, "Node %v readiness doesn't match expected readiness %v != %v", node.Name, gotReadiness, tc.expectedNodesReadiness[node.Name])
				assert.Equal(t, gotReadiness, readyNodes[node.Name], "Node %v is marked as ready, but not categorized as one", node.Name)
			}
		})
	}
}

func createTemplateNodeInfo(nodeName string, slices []*resourceapi.ResourceSlice) *framework.NodeInfo {
	return framework.NewNodeInfo(buildTestNode(nodeName, true), slices)
}

// buildResourceSlices builds a slice of resource slices with the given node name, driver name and number of ready pools.
// All resource pools are referencing a single generation and target one resource slice count to be counted as complete pool.
func buildResourceSlices(nodeName, driverName string, readyPoolsCount int) []*resourceapi.ResourceSlice {
	resourceSlices := make([]*resourceapi.ResourceSlice, readyPoolsCount)
	for i := 0; i < readyPoolsCount; i++ {
		pool := resourceapi.ResourcePool{
			Name:               fmt.Sprintf("%s_pool_%d", nodeName, i),
			ResourceSliceCount: 1,
			Generation:         1,
		}
		spec := resourceapi.ResourceSliceSpec{
			NodeName: &nodeName,
			Driver:   driverName,
			Pool:     pool,
			Devices:  []resourceapi.Device{{Name: fmt.Sprintf("%d_%d", i, 0)}},
		}
		resourceSlices[i] = &resourceapi.ResourceSlice{ObjectMeta: metav1.ObjectMeta{Name: nodeName}, Spec: spec}
	}
	return resourceSlices
}

func buildIncompleteResourceSlices(nodeName, driverName string, poolsCount int) []*resourceapi.ResourceSlice {
	rs := buildResourceSlices(nodeName, driverName, poolsCount)
	for i := range rs {
		rs[i].Spec.Pool.ResourceSliceCount = 2
		rs[i].Spec.Pool.Name = fmt.Sprintf("%s_incomplete_pool_%d", nodeName, i)
	}
	return rs
}

// buildResourcePoolWithSplitGeneration builds a list of resource slices for the requested resource pool
// The resource pool is split across multiple generations with different number of missing resource slices.
// If the number of missing resource slices is 0, the resource slice is considered complete.
// Resource slices are presented in random order depending on the map iteration order, it may lead to
// test flakiness, but results in a better test coverage without any additional complexity.
func buildResourcePoolWithSplitGeneration(nodeName, driverName string, id int, slicesMissingPerGeneration map[int]int) []*resourceapi.ResourceSlice {
	rs := make([]*resourceapi.ResourceSlice, 0, len(slicesMissingPerGeneration))
	for generation, countMissing := range slicesMissingPerGeneration {
		rs = append(rs, &resourceapi.ResourceSlice{
			ObjectMeta: metav1.ObjectMeta{Name: nodeName},
			Spec: resourceapi.ResourceSliceSpec{
				Driver: driverName,
				Pool: resourceapi.ResourcePool{
					Name:               fmt.Sprintf("%s_split_pool_%d", nodeName, id),
					ResourceSliceCount: int64(countMissing + 1),
					Generation:         int64(generation),
				},
			},
		})
	}

	return rs
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

func TestGetCompleteResourcePools(t *testing.T) {
	tests := map[string]struct {
		slices   []*v1.ResourceSlice
		expected map[string]int
	}{
		"EmptySlices": {
			slices:   []*v1.ResourceSlice{},
			expected: map[string]int{},
		},
		"SingleSlice": {
			slices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_1",
							ResourceSliceCount: 1,
							Generation:         1,
						},
					},
				},
			},
			expected: map[string]int{"driver": 1},
		},
		"MultipleSlicesSamePool": {
			slices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_1",
							ResourceSliceCount: 2,
							Generation:         1,
						},
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_1",
							ResourceSliceCount: 2,
							Generation:         1,
						},
					},
				},
			},
			expected: map[string]int{"driver": 1},
		},
		"MultipleSlicesDifferentPools": {
			slices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_1",
							ResourceSliceCount: 1,
							Generation:         1,
						},
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_2",
							ResourceSliceCount: 1,
							Generation:         1,
						},
					},
				},
			},
			expected: map[string]int{"driver": 2},
		},
		"MultipleSlicesDifferentDrivers": {
			slices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_1",
							ResourceSliceCount: 1,
							Generation:         1,
						},
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver2",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_1",
							ResourceSliceCount: 1,
							Generation:         1,
						},
					},
				},
			},
			expected: map[string]int{"driver1": 1, "driver2": 1},
		},
		"MultipleSlicesDifferentDriversSamePool_NotCompatible": {
			slices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_1",
							ResourceSliceCount: 2,
							Generation:         1,
						},
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver2",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_1",
							ResourceSliceCount: 2,
							Generation:         1,
						},
					},
				},
			},
			expected: map[string]int{},
		},
		"PoolWithMultipleGenerations": {
			slices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_1",
							ResourceSliceCount: 1,
							Generation:         1,
						},
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_1",
							ResourceSliceCount: 2,
							Generation:         2,
						},
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_1",
							ResourceSliceCount: 2,
							Generation:         2,
						},
					},
				},
			},
			expected: map[string]int{"driver": 1},
		},
		"PoolWithMultipleGenerationsDifferentDrivers": {
			slices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_1",
							ResourceSliceCount: 1,
							Generation:         1,
						},
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_1",
							ResourceSliceCount: 2,
							Generation:         1,
						},
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_1",
							ResourceSliceCount: 2,
							Generation:         2,
						},
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver2",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_1",
							ResourceSliceCount: 2,
							Generation:         2,
						},
					},
				},
			},
			expected: map[string]int{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := getCompleteResourcePools(test.slices)
			if diff := cmp.Diff(test.expected, result); diff != "" {
				t.Errorf("getCompleteResourcePools(): unexpected result (-want +got): %s", diff)
			}
		})
	}
}

func TestAreResourcePoolsReady(t *testing.T) {
	tests := map[string]struct {
		realSlices     []*v1.ResourceSlice
		templateSlices []*v1.ResourceSlice
		expected       bool
	}{
		"EmptyTemplatesAndRealSlices": {
			realSlices:     []*v1.ResourceSlice{},
			templateSlices: []*v1.ResourceSlice{},
			expected:       true,
		},
		"EmptyTemplatesWithRealSlices": {
			realSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_real",
							ResourceSliceCount: 1,
							Generation:         1,
						},
					},
				},
			},
			templateSlices: []*v1.ResourceSlice{},
			expected:       true,
		},
		"TemplateRequiresOneRealHasOne": {
			realSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_real",
							ResourceSliceCount: 1,
							Generation:         1,
						},
					},
				},
			},
			templateSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_template",
							ResourceSliceCount: 1,
							Generation:         1,
						},
					},
				},
			},
			expected: true,
		},
		"TemplateRequiresOneRealHasNone": {
			realSlices: []*v1.ResourceSlice{},
			templateSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_template",
							ResourceSliceCount: 1,
							Generation:         1,
						},
					},
				},
			},
			expected: false,
		},
		"TemplateRequiresOneRealHasTwo": {
			realSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_real_1",
							ResourceSliceCount: 1,
							Generation:         1,
						},
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_real_2",
							ResourceSliceCount: 1,
							Generation:         1,
						},
					},
				},
			},
			templateSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_template",
							ResourceSliceCount: 1,
							Generation:         1,
						},
					},
				},
			},
			expected: true,
		},
		"TemplateRequiresMultipleDriversRealMatches": {
			realSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
						Pool:   resourceapi.ResourcePool{Name: "real_1", ResourceSliceCount: 1, Generation: 1},
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver2",
						Pool:   resourceapi.ResourcePool{Name: "real_2", ResourceSliceCount: 1, Generation: 1},
					},
				},
			},
			templateSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
						Pool:   resourceapi.ResourcePool{Name: "tpl_1", ResourceSliceCount: 1, Generation: 1},
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver2",
						Pool:   resourceapi.ResourcePool{Name: "tpl_2", ResourceSliceCount: 1, Generation: 1},
					},
				},
			},
			expected: true,
		},
		"TemplateRequiresMultipleDriversRealMissingOne": {
			realSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
						Pool:   resourceapi.ResourcePool{Name: "real_1", ResourceSliceCount: 1, Generation: 1},
					},
				},
			},
			templateSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
						Pool:   resourceapi.ResourcePool{Name: "tpl_1", ResourceSliceCount: 1, Generation: 1},
					},
				},
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver2",
						Pool:   resourceapi.ResourcePool{Name: "tpl_2", ResourceSliceCount: 1, Generation: 1},
					},
				},
			},
			expected: false,
		},
		"RealHasIncompletePool": {
			realSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_real",
							ResourceSliceCount: 2,
							Generation:         1,
						},
					},
				},
			},
			templateSlices: []*v1.ResourceSlice{
				{
					Spec: v1.ResourceSliceSpec{
						Driver: "driver1",
						Pool: resourceapi.ResourcePool{
							Name:               "pool_template",
							ResourceSliceCount: 1,
							Generation:         1,
						},
					},
				},
			},
			expected: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			result := areResourcePoolsReady(test.realSlices, test.templateSlices)
			if result != test.expected {
				t.Errorf("areResourcePoolsReady(): unexpected result, got %v, want %v", result, test.expected)
			}
		})
	}
}

func BenchmarkGetCompleteResourcePools(b *testing.B) {
	slices := []*v1.ResourceSlice{
		{
			Spec: v1.ResourceSliceSpec{
				Pool: resourceapi.ResourcePool{
					Name:               "pool_1",
					ResourceSliceCount: 1,
					Generation:         2,
				},
			},
		},
		{
			Spec: v1.ResourceSliceSpec{
				Pool: resourceapi.ResourcePool{
					Name:               "pool_1",
					ResourceSliceCount: 2,
					Generation:         1,
				},
			},
		},
		{
			Spec: v1.ResourceSliceSpec{
				Pool: resourceapi.ResourcePool{
					Name:               "pool_1",
					ResourceSliceCount: 2,
					Generation:         1,
				},
			},
		},
		{
			Spec: v1.ResourceSliceSpec{
				Pool: resourceapi.ResourcePool{
					Name:               "pool_1",
					ResourceSliceCount: 2,
					Generation:         2,
				},
			},
		},
	}

	for b.Loop() {
		getCompleteResourcePools(slices)
	}
}

type fakeResourceDiscrepancyReporter struct {
	reportedNodeNames      []string
	reportedTemplateSlices [][]*resourceapi.ResourceSlice
	reportedNodeSlices     [][]*resourceapi.ResourceSlice
}

func (m *fakeResourceDiscrepancyReporter) ReportResourceDiscrepancies(nodeNames []string, templateSlices [][]*resourceapi.ResourceSlice, nodeSlices [][]*resourceapi.ResourceSlice) {
	m.reportedNodeNames = nodeNames
	m.reportedTemplateSlices = templateSlices
	m.reportedNodeSlices = nodeSlices
}

func TestDraProcessorResourceComparator(t *testing.T) {
	templateSlices := buildResourceSlices("template", "driver", 1)
	slicesReady := buildResourceSlices("node", "driver", 1)
	slicesUnready := buildIncompleteResourceSlices("node", "driver", 1)

	testCases := map[string]struct {
		nodeSlices             []*resourceapi.ResourceSlice
		templateSlices         []*resourceapi.ResourceSlice
		expectedTemplateSlices [][]*resourceapi.ResourceSlice
		expectedNodeSlices     [][]*resourceapi.ResourceSlice
		expectReported         bool
	}{
		"Ready": {
			nodeSlices:             slicesReady,
			templateSlices:         templateSlices,
			expectReported:         true,
			expectedTemplateSlices: [][]*resourceapi.ResourceSlice{templateSlices},
			expectedNodeSlices:     [][]*resourceapi.ResourceSlice{slicesReady},
		},
		"Unready": {
			nodeSlices:     slicesUnready,
			templateSlices: templateSlices,
			expectReported: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			node := buildTestNode("node1", true)
			nodeGroup := "ng1"

			provider := testprovider.NewTestCloudProviderBuilder().Build()
			provider.AddAutoprovisionedNodeGroup(nodeGroup, 0, 10, 1, "template")
			provider.AddNode(nodeGroup, node)

			nodeResourceSlices := map[string][]*resourceapi.ResourceSlice{node.Name: tc.nodeSlices}
			draSnapshot := drasnapshot.NewSnapshot(nil, nodeResourceSlices, nil, nil)
			clusterSnapshotStore := store.NewBasicSnapshotStore()
			clusterSnapshot, _, _ := testsnapshot.NewCustomTestSnapshotAndHandle(clusterSnapshotStore)
			clusterSnapshot.SetClusterState([]*apiv1.Node{}, []*apiv1.Pod{}, draSnapshot, nil)

			autoscalingCtx := &ca_context.AutoscalingContext{
				CloudProvider:   provider,
				ClusterSnapshot: clusterSnapshot,
				TemplateNodeInfoRegistry: newMockTemplateNodeInfoRegistry(map[string]*framework.NodeInfo{
					nodeGroup: createTemplateNodeInfo("template", tc.templateSlices),
				}),
			}

			mockComparator := &fakeResourceDiscrepancyReporter{}
			processor := DraCustomResourcesProcessor{resourcesComparator: mockComparator}

			processor.FilterOutNodesWithUnreadyResources(autoscalingCtx, []*apiv1.Node{node}, []*apiv1.Node{node}, draSnapshot, nil)

			if tc.expectReported {
				assert.Equal(t, []string{node.Name}, mockComparator.reportedNodeNames)
				assert.Equal(t, tc.expectedTemplateSlices, mockComparator.reportedTemplateSlices)
				assert.Equal(t, tc.expectedNodeSlices, mockComparator.reportedNodeSlices)
			} else {
				assert.Empty(t, mockComparator.reportedNodeNames)
				assert.Empty(t, mockComparator.reportedTemplateSlices)
				assert.Empty(t, mockComparator.reportedNodeSlices)
			}
		})
	}
}
