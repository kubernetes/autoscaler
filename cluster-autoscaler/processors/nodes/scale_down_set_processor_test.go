/*
Copyright 2024 The Kubernetes Authors.

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

package nodes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/kubernetes/fake"
)

func TestMaxNodesFilterUnremovableNodes(t *testing.T) {

	testCases := []struct {
		name                string
		candidates          []simulator.NodeToBeRemoved
		scaleDownContext    ScaleDownContext
		expectedToBeRemoved []simulator.NodeToBeRemoved
		expectedUnremovable []simulator.UnremovableNode
	}{
		{
			name: "Max nodes with simple limit",
			candidates: []simulator.NodeToBeRemoved{
				buildRemovableNode("node-1"),
				buildRemovableNode("node-2"),
				buildRemovableNode("node-3"),
				buildRemovableNode("node-4"),
				buildRemovableNode("node-5"),
			},
			scaleDownContext: *NewDefaultScaleDownSetContext(3),
			expectedToBeRemoved: []simulator.NodeToBeRemoved{
				buildRemovableNode("node-1"),
				buildRemovableNode("node-2"),
				buildRemovableNode("node-3"),
			},
			expectedUnremovable: []simulator.UnremovableNode{
				buildUnremovableNode("node-4", simulator.NodeGroupMaxDeletionCountReached),
				buildUnremovableNode("node-5", simulator.NodeGroupMaxDeletionCountReached),
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			processor := NewMaxNodesProcessor()
			toBeRemoved, unRemovable := processor.FilterUnremovableNodes(nil, tc.scaleDownContext, tc.candidates)

			assert.ElementsMatch(t, tc.expectedToBeRemoved, toBeRemoved)
			assert.ElementsMatch(t, tc.expectedUnremovable, unRemovable)
		})
	}
}

func TestAtomicResizeFilterUnremovableNodes(t *testing.T) {
	testCases := []struct {
		name       string
		nodeGroups []struct {
			nodeGroupName string
			nodeGroupSize int
		}
		nodes []struct {
			candidate simulator.NodeToBeRemoved
			nodeGroup string
		}
		scaleDownContext    ScaleDownContext
		expectedToBeRemoved []simulator.NodeToBeRemoved
		expectedUnremovable []simulator.UnremovableNode
	}{
		{
			name: "Atomic removal",
			nodeGroups: []struct {
				nodeGroupName string
				nodeGroupSize int
			}{
				{
					nodeGroupName: "ng1",
					nodeGroupSize: 3,
				},
				{
					nodeGroupName: "ng2",
					nodeGroupSize: 4,
				},
			},
			nodes: []struct {
				candidate simulator.NodeToBeRemoved
				nodeGroup string
			}{
				{
					candidate: buildRemovableNode("node-1"),
					nodeGroup: "ng1",
				},
				{
					candidate: buildRemovableNode("node-2"),
					nodeGroup: "ng1",
				},
				{
					candidate: buildRemovableNode("node-3"),
					nodeGroup: "ng1",
				},
				{
					candidate: buildRemovableNode("node-4"),
					nodeGroup: "ng2",
				},
				{
					candidate: buildRemovableNode("node-5"),
					nodeGroup: "ng2",
				},
			},
			scaleDownContext: *NewDefaultScaleDownSetContext(10),
			expectedToBeRemoved: []simulator.NodeToBeRemoved{
				buildRemovableNode("node-1"),
				buildRemovableNode("node-2"),
				buildRemovableNode("node-3"),
			},
			expectedUnremovable: []simulator.UnremovableNode{
				buildUnremovableNode("node-4", simulator.AtomicScaleDownFailed),
				buildUnremovableNode("node-5", simulator.AtomicScaleDownFailed),
			},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			processor := NewAtomicResizeFilteringProcessor()
			provider := testprovider.NewTestCloudProvider(nil, nil)
			for _, ng := range tc.nodeGroups {
				provider.AddNodeGroupWithCustomOptions(ng.nodeGroupName, 0, 100, ng.nodeGroupSize, &config.NodeGroupAutoscalingOptions{
					ZeroOrMaxNodeScaling: true,
				})
			}
			candidates := []simulator.NodeToBeRemoved{}
			for _, node := range tc.nodes {
				provider.AddNode(node.nodeGroup, node.candidate.Node)
				candidates = append(candidates, node.candidate)
			}
			context, _ := NewScaleTestAutoscalingContext(config.AutoscalingOptions{
				NodeGroupDefaults: config.NodeGroupAutoscalingOptions{},
			}, &fake.Clientset{}, nil, provider, nil, nil)

			toBeRemoved, unRemovable := processor.FilterUnremovableNodes(&context, tc.scaleDownContext, candidates)

			assert.ElementsMatch(t, tc.expectedToBeRemoved, toBeRemoved)
			assert.ElementsMatch(t, tc.expectedUnremovable, unRemovable)
		})
	}
}

func buildRemovableNode(name string) simulator.NodeToBeRemoved {

	return simulator.NodeToBeRemoved{
		Node: BuildTestNode(name, 1000, 10),
	}
}

func buildUnremovableNode(name string, reason simulator.UnremovableReason) simulator.UnremovableNode {
	return simulator.UnremovableNode{
		Node:   BuildTestNode(name, 1000, 10),
		Reason: reason,
	}
}
