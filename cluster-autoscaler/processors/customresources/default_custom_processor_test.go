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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	csisnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/csi/snapshot"
	drasnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"

	apiv1 "k8s.io/api/core/v1"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	utils "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestDefaultProcessorFilterOut(t *testing.T) {
	processor := DefaultCustomResourcesProcessor{[]CustomResourcesProcessor{
		&mockCustomResourcesProcessor{nodeMark: "p1"},
		&mockCustomResourcesProcessor{nodeMark: "p2"},
		&mockCustomResourcesProcessor{nodeMark: "p3"},
	}}

	testCases := map[string]struct {
		allNodes              []*apiv1.Node
		nodesInitialReadiness map[string]bool
		expectedReadyNodes    map[string]bool
	}{
		"filtering one node by one processor": {
			allNodes: []*apiv1.Node{
				utils.BuildTestNode("p1_node_1", 500, 100),
				utils.BuildTestNode("node_2", 500, 100),
			},
			nodesInitialReadiness: map[string]bool{
				"p1_node_1": true,
				"node_2":    true,
			},
			expectedReadyNodes: map[string]bool{
				"node_2": true,
			},
		},
		"filtering multiple nodes by one processor": {
			allNodes: []*apiv1.Node{
				utils.BuildTestNode("p1_node_1", 500, 100),
				utils.BuildTestNode("p1_node_2", 500, 100),
				utils.BuildTestNode("node_3", 500, 100),
			},
			nodesInitialReadiness: map[string]bool{
				"p1_node_1": true,
				"p1_node_2": true,
				"node_3":    false,
			},
			expectedReadyNodes: map[string]bool{},
		},
		"filtering one node by multiple processors": {
			allNodes: []*apiv1.Node{
				utils.BuildTestNode("p1_p3_node_1", 500, 100),
				utils.BuildTestNode("p1_node_2", 500, 100),
				utils.BuildTestNode("node_3", 500, 100),
			},
			nodesInitialReadiness: map[string]bool{
				"p1_node_1": true,
				"p1_node_2": false,
				"node_3":    false,
			},
			expectedReadyNodes: map[string]bool{},
		},
		"filtering multiple nodes by multiple processor": {
			allNodes: []*apiv1.Node{
				utils.BuildTestNode("p1_node_1", 500, 100),
				utils.BuildTestNode("p1_node_2", 500, 100),
				utils.BuildTestNode("node_3", 500, 100),
				utils.BuildTestNode("node_4", 500, 100),
				utils.BuildTestNode("p2_node_5", 500, 100),
				utils.BuildTestNode("p3_node_6", 500, 100),
			},
			nodesInitialReadiness: map[string]bool{
				"p1_node_1": false,
				"p1_node_2": true,
				"node_3":    false,
				"node_4":    true,
				"p2_node_5": true,
				"p3_node_6": true,
			},
			expectedReadyNodes: map[string]bool{
				"node_4": true,
			},
		},
	}
	for tcName, tc := range testCases {
		t.Run(tcName, func(t *testing.T) {
			readyNodes := []*apiv1.Node{}
			for _, node := range tc.allNodes {
				if tc.nodesInitialReadiness[node.Name] {
					readyNodes = append(readyNodes, node)
				}
			}
			resultedAllNodes, resultedReadyNodes := processor.FilterOutNodesWithUnreadyResources(nil, tc.allNodes, readyNodes, nil, nil)
			assert.ElementsMatch(t, tc.allNodes, resultedAllNodes)
			assert.True(t, len(resultedReadyNodes) == len(tc.expectedReadyNodes))
			for _, node := range resultedReadyNodes {
				assert.True(t, tc.expectedReadyNodes[node.Name])
			}

		})
	}

}

func TestDefaultProcessorGetNodeResourceTargets(t *testing.T) {
	processor := DefaultCustomResourcesProcessor{[]CustomResourcesProcessor{
		&mockCustomResourcesProcessor{nodeMark: "p1", customResourceTargetsToAdd: []string{"p1_R1", "p1_R2"}, customResourceTargetsQuantity: 1},
		&mockCustomResourcesProcessor{nodeMark: "p2", customResourceTargetsToAdd: []string{"p2_R1"}, customResourceTargetsQuantity: 2},
		&mockCustomResourcesProcessor{nodeMark: "p3", customResourceTargetsToAdd: []string{"p3_R1"}, customResourceTargetsQuantity: 3},
	}}

	testCases := map[string]struct {
		node              *apiv1.Node
		expectedResources []CustomResourceTarget
	}{
		"single processor": {
			node: utils.BuildTestNode("p1", 500, 100),
			expectedResources: []CustomResourceTarget{
				{ResourceType: "p1_R1", ResourceCount: 1},
				{ResourceType: "p1_R2", ResourceCount: 1},
			},
		},
		"many processors": {
			node: utils.BuildTestNode("p1_p3", 500, 100),
			expectedResources: []CustomResourceTarget{
				{ResourceType: "p1_R1", ResourceCount: 1},
				{ResourceType: "p1_R2", ResourceCount: 1},
				{ResourceType: "p3_R1", ResourceCount: 3},
			},
		},
		"all processors": {
			node: utils.BuildTestNode("p1_p2_p3", 500, 100),
			expectedResources: []CustomResourceTarget{
				{ResourceType: "p1_R1", ResourceCount: 1},
				{ResourceType: "p1_R2", ResourceCount: 1},
				{ResourceType: "p2_R1", ResourceCount: 2},
				{ResourceType: "p3_R1", ResourceCount: 3},
			},
		},
	}
	for tcName, tc := range testCases {
		t.Run(tcName, func(t *testing.T) {
			customResourceTarget, _ := processor.GetNodeResourceTargets(nil, tc.node, nil)
			assert.ElementsMatch(t, customResourceTarget, tc.expectedResources)
		})
	}
}

type mockCustomResourcesProcessor struct {
	nodeMark                      string
	customResourceTargetsToAdd    []string
	customResourceTargetsQuantity int64
}

func (m *mockCustomResourcesProcessor) FilterOutNodesWithUnreadyResources(_ *ca_context.AutoscalingContext, allNodes, readyNodes []*apiv1.Node, _ *drasnapshot.Snapshot, _ *csisnapshot.Snapshot) ([]*apiv1.Node, []*apiv1.Node) {
	filteredReadyNodes := []*apiv1.Node{}
	for _, node := range readyNodes {
		if !strings.Contains(node.Name, m.nodeMark) {
			filteredReadyNodes = append(filteredReadyNodes, node)
		}
	}
	return allNodes, filteredReadyNodes
}

func (m *mockCustomResourcesProcessor) GetNodeResourceTargets(_ *ca_context.AutoscalingContext, node *apiv1.Node, _ cloudprovider.NodeGroup) ([]CustomResourceTarget, errors.AutoscalerError) {
	result := []CustomResourceTarget{}
	if strings.Contains(node.Name, m.nodeMark) {
		for _, rt := range m.customResourceTargetsToAdd {
			result = append(result, CustomResourceTarget{ResourceType: rt, ResourceCount: m.customResourceTargetsQuantity})
		}
	}
	return result, nil
}

func (m *mockCustomResourcesProcessor) CleanUp() {
}
