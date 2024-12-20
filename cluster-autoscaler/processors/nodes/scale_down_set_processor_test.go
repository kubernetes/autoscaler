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

func TestAtomicResizeFilterUnremovableNodes(t *testing.T) {
	testCases := []struct {
		name       string
		nodeGroups []struct {
			nodeGroupName        string
			nodeGroupTargetSize  int
			zeroOrMaxNodeScaling bool
		}
		removableCandidates []struct {
			candidate simulator.NodeToBeRemoved
			nodeGroup string
		}
		scaleDownContext    *ScaleDownContext
		expectedToBeRemoved []simulator.NodeToBeRemoved
		expectedUnremovable []simulator.UnremovableNode
	}{
		{
			name: "Atomic removal",
			nodeGroups: []struct {
				nodeGroupName        string
				nodeGroupTargetSize  int
				zeroOrMaxNodeScaling bool
			}{
				{
					nodeGroupName:        "ng1",
					nodeGroupTargetSize:  3,
					zeroOrMaxNodeScaling: true,
				},
				{
					nodeGroupName:        "ng2",
					nodeGroupTargetSize:  4,
					zeroOrMaxNodeScaling: true,
				},
			},
			removableCandidates: []struct {
				candidate simulator.NodeToBeRemoved
				nodeGroup string
			}{
				{
					candidate: buildRemovableNode("ng1-node-1"),
					nodeGroup: "ng1",
				},
				{
					candidate: buildRemovableNode("ng1-node-2"),
					nodeGroup: "ng1",
				},
				{
					candidate: buildRemovableNode("ng1-node-3"),
					nodeGroup: "ng1",
				},
				{
					candidate: buildRemovableNode("ng2-node-1"),
					nodeGroup: "ng2",
				},
				{
					candidate: buildRemovableNode("ng2-node-2"),
					nodeGroup: "ng2",
				},
			},
			scaleDownContext: NewDefaultScaleDownContext(),
			expectedToBeRemoved: []simulator.NodeToBeRemoved{
				buildRemovableNode("ng1-node-1"),
				buildRemovableNode("ng1-node-2"),
				buildRemovableNode("ng1-node-3"),
			},
			expectedUnremovable: []simulator.UnremovableNode{
				buildUnremovableNode("ng2-node-1", simulator.AtomicScaleDownFailed),
				buildUnremovableNode("ng2-node-2", simulator.AtomicScaleDownFailed),
			},
		},
		{
			name: "Mixed Groups",
			nodeGroups: []struct {
				nodeGroupName        string
				nodeGroupTargetSize  int
				zeroOrMaxNodeScaling bool
			}{
				{
					nodeGroupName:        "ng1",
					nodeGroupTargetSize:  3,
					zeroOrMaxNodeScaling: false,
				},
				{
					nodeGroupName:        "ng2",
					nodeGroupTargetSize:  4,
					zeroOrMaxNodeScaling: true,
				},
			},
			removableCandidates: []struct {
				candidate simulator.NodeToBeRemoved
				nodeGroup string
			}{
				{
					candidate: buildRemovableNode("ng1-node-1"),
					nodeGroup: "ng1",
				},
				{
					candidate: buildRemovableNode("ng1-node-2"),
					nodeGroup: "ng1",
				},
				{
					candidate: buildRemovableNode("ng1-node-3"),
					nodeGroup: "ng1",
				},
				{
					candidate: buildRemovableNode("ng2-node-1"),
					nodeGroup: "ng2",
				},
				{
					candidate: buildRemovableNode("ng2-node-2"),
					nodeGroup: "ng2",
				},
			},
			scaleDownContext: NewDefaultScaleDownContext(),
			expectedToBeRemoved: []simulator.NodeToBeRemoved{
				buildRemovableNode("ng1-node-1"),
				buildRemovableNode("ng1-node-2"),
				buildRemovableNode("ng1-node-3"),
			},
			expectedUnremovable: []simulator.UnremovableNode{
				buildUnremovableNode("ng2-node-1", simulator.AtomicScaleDownFailed),
				buildUnremovableNode("ng2-node-2", simulator.AtomicScaleDownFailed),
			},
		},
		{
			name: "No atomic groups",
			nodeGroups: []struct {
				nodeGroupName        string
				nodeGroupTargetSize  int
				zeroOrMaxNodeScaling bool
			}{
				{
					nodeGroupName:        "ng1",
					nodeGroupTargetSize:  3,
					zeroOrMaxNodeScaling: false,
				},
				{
					nodeGroupName:        "ng2",
					nodeGroupTargetSize:  4,
					zeroOrMaxNodeScaling: false,
				},
			},
			removableCandidates: []struct {
				candidate simulator.NodeToBeRemoved
				nodeGroup string
			}{
				{
					candidate: buildRemovableNode("ng1-node-1"),
					nodeGroup: "ng1",
				},
				{
					candidate: buildRemovableNode("ng1-node-2"),
					nodeGroup: "ng1",
				},
				{
					candidate: buildRemovableNode("ng1-node-3"),
					nodeGroup: "ng1",
				},
				{
					candidate: buildRemovableNode("ng2-node-1"),
					nodeGroup: "ng2",
				},
				{
					candidate: buildRemovableNode("ng2-node-2"),
					nodeGroup: "ng2",
				},
			},
			scaleDownContext: NewDefaultScaleDownContext(),
			expectedToBeRemoved: []simulator.NodeToBeRemoved{
				buildRemovableNode("ng1-node-1"),
				buildRemovableNode("ng1-node-2"),
				buildRemovableNode("ng1-node-3"),
				buildRemovableNode("ng2-node-1"),
				buildRemovableNode("ng2-node-2"),
			},
			expectedUnremovable: []simulator.UnremovableNode{},
		},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			processor := NewAtomicResizeFilteringProcessor()
			provider := testprovider.NewTestCloudProvider(nil, nil)
			for _, ng := range tc.nodeGroups {
				provider.AddNodeGroupWithCustomOptions(ng.nodeGroupName, 0, 100, ng.nodeGroupTargetSize, &config.NodeGroupAutoscalingOptions{
					ZeroOrMaxNodeScaling: ng.zeroOrMaxNodeScaling,
				})
			}
			candidates := []simulator.NodeToBeRemoved{}
			for _, node := range tc.removableCandidates {
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
