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
	"github.com/google/go-cmp/cmp/cmpopts"
	apiv1 "k8s.io/api/core/v1"
	cptest "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
)

func TestCheckDelta(t *testing.T) {
	testCases := []struct {
		name         string
		tracker      *Tracker
		node         *apiv1.Node
		nodeDelta    int
		wantResult   *CheckDeltaResult
		wantExceeded bool
	}{
		{
			name: "delta fits within limits",
			tracker: newTracker(&fakeCustomResourcesProcessor{}, []*quotaStatus{
				{
					quota:      &fakeQuota{id: "limiter1", appliesToFn: func(*apiv1.Node) bool { return true }},
					limitsLeft: resourceList{"cpu": 10, "memory": 1000, "nodes": 5},
				},
			}),
			node:      test.BuildTestNode("n1", 1000, 200),
			nodeDelta: 2,
			wantResult: &CheckDeltaResult{
				AllowedDelta: 2,
			},
		},
		{
			name: "delta exceeds one resource limit",
			tracker: newTracker(&fakeCustomResourcesProcessor{}, []*quotaStatus{
				{
					quota:      &fakeQuota{id: "limiter1", appliesToFn: func(*apiv1.Node) bool { return true }},
					limitsLeft: resourceList{"cpu": 1, "memory": 1000, "nodes": 5},
				},
			}),
			node:      test.BuildTestNode("n1", 1000, 200),
			nodeDelta: 2,
			wantResult: &CheckDeltaResult{
				AllowedDelta: 1,
				ExceededQuotas: []ExceededQuota{
					{ID: "limiter1", ExceededResources: []string{"cpu"}},
				},
			},
			wantExceeded: true,
		},
		{
			name: "delta exceeds multiple resource limits",
			tracker: newTracker(&fakeCustomResourcesProcessor{}, []*quotaStatus{
				{
					quota:      &fakeQuota{id: "limiter1", appliesToFn: func(*apiv1.Node) bool { return true }},
					limitsLeft: resourceList{"cpu": 1, "memory": 300, "nodes": 5},
				},
			}),
			node:      test.BuildTestNode("n1", 1000, 200),
			nodeDelta: 2,
			wantResult: &CheckDeltaResult{
				AllowedDelta: 1,
				ExceededQuotas: []ExceededQuota{
					{ID: "limiter1", ExceededResources: []string{"cpu", "memory"}},
				},
			},
			wantExceeded: true,
		},
		{
			name: "no matching quotas",
			tracker: newTracker(&fakeCustomResourcesProcessor{}, []*quotaStatus{
				{
					quota:      &fakeQuota{id: "limiter1", appliesToFn: func(*apiv1.Node) bool { return false }},
					limitsLeft: resourceList{"cpu": 1, "memory": 100, "nodes": 1},
				},
			}),
			node:      test.BuildTestNode("n1", 1000, 200),
			nodeDelta: 2,
			wantResult: &CheckDeltaResult{
				AllowedDelta: 2,
			},
		},
		{
			name: "resource in limits but not in the node",
			tracker: newTracker(&fakeCustomResourcesProcessor{}, []*quotaStatus{
				{
					quota:      &fakeQuota{id: "limiter1", appliesToFn: func(*apiv1.Node) bool { return true }},
					limitsLeft: resourceList{"cpu": 4, "memory": 32 * units.GiB, "gpu": 2},
				},
			}),
			node:      test.BuildTestNode("n1", 1000, 2000),
			nodeDelta: 2,
			wantResult: &CheckDeltaResult{
				AllowedDelta: 2,
			},
		},
		{
			name: "resource in the node but not in the limits",
			tracker: newTracker(&fakeCustomResourcesProcessor{NodeResourceTargets: func(node *apiv1.Node) []customresources.CustomResourceTarget {
				return []customresources.CustomResourceTarget{
					{
						ResourceType:  "gpu",
						ResourceCount: 1,
					},
				}
			}}, []*quotaStatus{
				{
					quota:      &fakeQuota{id: "limiter1", appliesToFn: func(*apiv1.Node) bool { return true }},
					limitsLeft: resourceList{"cpu": 4, "memory": 32 * units.GiB},
				},
			}),
			node:      test.BuildTestNode("n1", 1000, 2000),
			nodeDelta: 2,
			wantResult: &CheckDeltaResult{
				AllowedDelta: 2,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := cptest.NewTestCloudProviderBuilder().Build()
			ctx := &context.AutoscalingContext{CloudProvider: provider}
			gotResult, err := tc.tracker.CheckDelta(ctx, nil, tc.node, tc.nodeDelta)
			if err != nil {
				t.Fatalf("CheckDelta() returned an unexpected error: %v", err)
			}

			if diff := cmp.Diff(tc.wantResult, gotResult, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("CheckDelta() mismatch (-want +got):\n%s", diff)
			}
			if gotResult.Exceeded() != tc.wantExceeded {
				t.Errorf("Exceeded() mismatch, want: %v, got: %v", tc.wantExceeded, gotResult.Exceeded())
			}
		})
	}
}

func TestApplyDelta(t *testing.T) {
	testCases := []struct {
		name           string
		tracker        *Tracker
		node           *apiv1.Node
		nodeDelta      int
		wantResult     *CheckDeltaResult
		wantLimitsLeft map[string]resourceList
	}{
		{
			name: "delta applied successfully",
			tracker: newTracker(&fakeCustomResourcesProcessor{}, []*quotaStatus{
				{
					quota:      &fakeQuota{id: "limiter1", appliesToFn: func(*apiv1.Node) bool { return true }},
					limitsLeft: resourceList{"cpu": 10, "memory": 1000, "nodes": 5},
				},
			}),
			node:      test.BuildTestNode("n1", 1000, 200),
			nodeDelta: 2,
			wantResult: &CheckDeltaResult{
				AllowedDelta: 2,
			},
			wantLimitsLeft: map[string]resourceList{
				"limiter1": {"cpu": 8, "memory": 600, "nodes": 3},
			},
		},
		{
			name: "partial delta calculated, nothing applied",
			tracker: newTracker(&fakeCustomResourcesProcessor{}, []*quotaStatus{
				{
					quota:      &fakeQuota{id: "limiter1", appliesToFn: func(*apiv1.Node) bool { return true }},
					limitsLeft: resourceList{"cpu": 3, "memory": 1000, "nodes": 5},
				},
			}),
			node:      test.BuildTestNode("n1", 2000, 200),
			nodeDelta: 2,
			wantResult: &CheckDeltaResult{
				AllowedDelta: 1,
				ExceededQuotas: []ExceededQuota{
					{ID: "limiter1", ExceededResources: []string{"cpu"}},
				},
			},
			wantLimitsLeft: map[string]resourceList{
				"limiter1": {"cpu": 3, "memory": 1000, "nodes": 5},
			},
		},
		{
			name: "delta not applied because it exceeds limits",
			tracker: newTracker(&fakeCustomResourcesProcessor{}, []*quotaStatus{
				{
					quota:      &fakeQuota{id: "limiter1", appliesToFn: func(*apiv1.Node) bool { return true }},
					limitsLeft: resourceList{"cpu": 1, "memory": 100, "nodes": 5},
				},
			}),
			node:      test.BuildTestNode("n1", 2000, 200),
			nodeDelta: 1,
			wantResult: &CheckDeltaResult{
				AllowedDelta: 0,
				ExceededQuotas: []ExceededQuota{
					{ID: "limiter1", ExceededResources: []string{"cpu", "memory"}},
				},
			},
			wantLimitsLeft: map[string]resourceList{
				"limiter1": {"cpu": 1, "memory": 100, "nodes": 5},
			},
		},
		{
			name: "applied delta results in zero limit",
			tracker: newTracker(&fakeCustomResourcesProcessor{}, []*quotaStatus{
				{
					quota:      &fakeQuota{id: "limiter1", appliesToFn: func(*apiv1.Node) bool { return true }},
					limitsLeft: resourceList{"cpu": 2, "memory": 500, "nodes": 10},
				},
			}),
			node:      test.BuildTestNode("n1", 1000, 200),
			nodeDelta: 2,
			wantResult: &CheckDeltaResult{
				AllowedDelta: 2,
			},
			wantLimitsLeft: map[string]resourceList{
				"limiter1": {"cpu": 0, "memory": 100, "nodes": 8},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := cptest.NewTestCloudProviderBuilder().Build()
			ctx := &context.AutoscalingContext{CloudProvider: provider}
			gotResult, err := tc.tracker.ApplyDelta(ctx, nil, tc.node, tc.nodeDelta)
			if err != nil {
				t.Fatalf("ApplyDelta() returned an unexpected error: %v", err)
			}

			if diff := cmp.Diff(tc.wantResult, gotResult, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ApplyDelta() result mismatch (-want +got):\n%s", diff)
			}

			gotLimitsLeft := make(map[string]resourceList)
			for _, ls := range tc.tracker.quotaStatuses {
				gotLimitsLeft[ls.quota.ID()] = ls.limitsLeft
			}
			if diff := cmp.Diff(tc.wantLimitsLeft, gotLimitsLeft, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("ApplyDelta() limitsLeft mismatch (-want +got):\n%s", diff)
			}

		})
	}
}

func TestDeltaForNode(t *testing.T) {
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
			delta, err := deltaForNode(ctx, tc.crp, tc.node, nil)
			if err != nil {
				t.Errorf("deltaForNode: unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantDelta, delta); diff != "" {
				t.Errorf("delta mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
