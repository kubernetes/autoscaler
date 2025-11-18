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

package resourcequotas

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	cptest "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
	drasnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

type mockCustomResourcesProcessor struct {
	mock.Mock
}

func (m *mockCustomResourcesProcessor) FilterOutNodesWithUnreadyResources(_ *context.AutoscalingContext, allNodes, readyNodes []*apiv1.Node, _ *drasnapshot.Snapshot) ([]*apiv1.Node, []*apiv1.Node) {
	return allNodes, readyNodes
}

func (m *mockCustomResourcesProcessor) GetNodeResourceTargets(autoscalingCtx *context.AutoscalingContext, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup) ([]customresources.CustomResourceTarget, errors.AutoscalerError) {
	args := m.Called(autoscalingCtx, node, nodeGroup)
	return args.Get(0).([]customresources.CustomResourceTarget), nil
}

func (m *mockCustomResourcesProcessor) CleanUp() {
	return
}

func TestNodeResourcesCache(t *testing.T) {
	node := test.BuildTestNode("n1", 1000, 2000)
	autoscalingCtx := &context.AutoscalingContext{}
	ng1 := cptest.NewTestNodeGroup("ng1", 1, 10, 1, true, false, "n1-template", nil, nil)
	ng2 := cptest.NewTestNodeGroup("ng2", 1, 10, 1, true, false, "n2-template", nil, nil)
	resourceTargets := []customresources.CustomResourceTarget{
		{ResourceType: "gpu", ResourceCount: 1},
	}
	wantResources := resourceList{"cpu": 1, "memory": 2000, "nodes": 1, "gpu": 1}

	type nodeResourcesCall struct {
		node      *apiv1.Node
		nodeGroup cloudprovider.NodeGroup
	}

	testCases := []struct {
		name                 string
		calls                []nodeResourcesCall
		setupCRPExpectations func(*mock.Mock)
	}{
		{
			name: "cache hit",
			calls: []nodeResourcesCall{
				{node: node, nodeGroup: ng1},
				{node: node, nodeGroup: ng1},
			},
			setupCRPExpectations: func(m *mock.Mock) {
				m.On("GetNodeResourceTargets", autoscalingCtx, node, ng1).Return(resourceTargets, nil).Once()
			},
		},
		{
			name: "cache miss on different node group",
			calls: []nodeResourcesCall{
				{node: node, nodeGroup: ng1},
				{node: node, nodeGroup: ng2},
			},
			setupCRPExpectations: func(m *mock.Mock) {
				m.On("GetNodeResourceTargets", autoscalingCtx, node, ng1).Return(resourceTargets, nil).Once().
					On("GetNodeResourceTargets", autoscalingCtx, node, ng2).Return(resourceTargets, nil).Once()
			},
		},
		{
			name: "no node group bypasses cache",
			calls: []nodeResourcesCall{
				{node: node, nodeGroup: nil},
				{node: node, nodeGroup: nil},
			},
			setupCRPExpectations: func(m *mock.Mock) {
				m.On("GetNodeResourceTargets", autoscalingCtx, node, nil).Return(resourceTargets, nil).Twice()
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCRP := &mockCustomResourcesProcessor{}
			tc.setupCRPExpectations(&mockCRP.Mock)
			nc := newNodeResourcesCache(mockCRP)
			for _, call := range tc.calls {
				resources, err := nc.totalNodeResources(autoscalingCtx, call.node, call.nodeGroup)
				if err != nil {
					t.Fatalf("totalNodeResources unexpected error: %v", err)
				}
				if diff := cmp.Diff(wantResources, resources); diff != "" {
					t.Errorf("totalNodeResources mismatch (-want, +got):\n%s", diff)
				}
			}
		})
	}
}

func TestNodeResources(t *testing.T) {
	testCases := []struct {
		name      string
		node      *apiv1.Node
		crp       customresources.CustomResourcesProcessor
		wantDelta resourceList
	}{
		{
			name: "node just with CPU and memory",
			node: test.BuildTestNode("test", 1000, 2048),
			crp:  &fakeCustomResourcesProcessor{},
			wantDelta: resourceList{
				"cpu":    1,
				"memory": 2048,
				"nodes":  1,
			},
		},
		{
			// nodes should not have milliCPUs in the capacity, so we round it up
			// to the nearest integer.
			name: "node just with CPU and memory, milli cores rounded up",
			node: test.BuildTestNode("test", 2500, 4096),
			crp:  &fakeCustomResourcesProcessor{},
			wantDelta: resourceList{
				"cpu":    3,
				"memory": 4096,
				"nodes":  1,
			},
		},
		{
			name: "node with custom resources",
			node: test.BuildTestNode("test", 1000, 2048),
			crp: &fakeCustomResourcesProcessor{NodeResourceTargets: func(node *apiv1.Node) []customresources.CustomResourceTarget {
				return []customresources.CustomResourceTarget{
					{
						ResourceType:  "gpu",
						ResourceCount: 1,
					},
				}
			}},
			wantDelta: resourceList{
				"cpu":    1,
				"memory": 2048,
				"gpu":    1,
				"nodes":  1,
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &context.AutoscalingContext{}
			delta, err := totalNodeResources(ctx, tc.crp, tc.node, nil)
			if err != nil {
				t.Errorf("totalNodeResources: unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantDelta, delta); diff != "" {
				t.Errorf("delta mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
