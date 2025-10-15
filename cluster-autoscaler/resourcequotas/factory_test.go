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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	cptest "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
)

type nodeExcludeFn func(node *apiv1.Node) bool

func (n nodeExcludeFn) ExcludeFromTracking(node *apiv1.Node) bool {
	return n(node)
}

func TestNewQuotasTracker(t *testing.T) {
	testCases := []struct {
		name       string
		crp        customresources.CustomResourcesProcessor
		nodeFilter NodeFilter
		nodes      []*apiv1.Node
		limits     map[string]int64
		newNode    *apiv1.Node
		nodeDelta  int
		wantResult *CheckDeltaResult
	}{
		{
			name: "default config allowed operation",
			nodes: []*apiv1.Node{
				test.BuildTestNode("n1", 1000, 2*units.GiB),
				test.BuildTestNode("n2", 2000, 4*units.GiB),
				test.BuildTestNode("n3", 3000, 8*units.GiB),
			},
			limits: map[string]int64{
				"cpu":    12,
				"memory": 32 * units.GiB,
			},
			newNode:   test.BuildTestNode("n4", 2000, 4*units.GiB),
			nodeDelta: 2,
			wantResult: &CheckDeltaResult{
				AllowedDelta: 2,
			},
		},
		{
			name: "default config exceeded operation",
			nodes: []*apiv1.Node{
				test.BuildTestNode("n1", 1000, 2*units.GiB),
				test.BuildTestNode("n2", 2000, 4*units.GiB),
				test.BuildTestNode("n3", 3000, 8*units.GiB),
			},
			limits: map[string]int64{
				"cpu":    6,
				"memory": 16 * units.GiB,
			},
			newNode:   test.BuildTestNode("n4", 2000, 4*units.GiB),
			nodeDelta: 2,
			wantResult: &CheckDeltaResult{
				AllowedDelta: 0,
				ExceededQuotas: []ExceededQuota{
					{ID: "cluster-wide", ExceededResources: []string{"cpu", "memory"}},
				},
			},
		},
		{
			name: "default config partially allowed operation",
			nodes: []*apiv1.Node{
				test.BuildTestNode("n1", 1000, 2*units.GiB),
				test.BuildTestNode("n2", 2000, 4*units.GiB),
				test.BuildTestNode("n3", 3000, 8*units.GiB),
			},
			limits: map[string]int64{
				"cpu":    7,
				"memory": 16 * units.GiB,
			},
			newNode:   test.BuildTestNode("n4", 2000, 4*units.GiB),
			nodeDelta: 2,
			wantResult: &CheckDeltaResult{
				AllowedDelta: 0,
				ExceededQuotas: []ExceededQuota{
					{ID: "cluster-wide", ExceededResources: []string{"cpu", "memory"}},
				},
			},
		},
		{
			name: "custom resource config allowed operation",
			crp: &fakeCustomResourcesProcessor{
				NodeResourceTargets: func(n *apiv1.Node) []customresources.CustomResourceTarget {
					if n.Name == "n1" {
						return []customresources.CustomResourceTarget{
							{
								ResourceType:  "gpu",
								ResourceCount: 1,
							},
						}
					}
					return nil
				},
			},
			nodes: []*apiv1.Node{
				test.BuildTestNode("n1", 1000, 2*units.GiB),
				test.BuildTestNode("n2", 2000, 4*units.GiB),
				test.BuildTestNode("n3", 3000, 8*units.GiB),
			},
			limits: map[string]int64{
				"cpu":    12,
				"memory": 32 * units.GiB,
				"gpu":    6,
			},
			newNode:   test.BuildTestNode("n4", 2000, 4*units.GiB),
			nodeDelta: 2,
			wantResult: &CheckDeltaResult{
				AllowedDelta: 2,
			},
		},
		{
			name: "custom resource config exceeded operation",
			crp: &fakeCustomResourcesProcessor{
				NodeResourceTargets: func(n *apiv1.Node) []customresources.CustomResourceTarget {
					if n.Name == "n1" || n.Name == "n4" {
						return []customresources.CustomResourceTarget{
							{
								ResourceType:  "gpu",
								ResourceCount: 1,
							},
						}
					}
					return nil
				},
			},
			nodes: []*apiv1.Node{
				test.BuildTestNode("n1", 1000, 2*units.GiB),
				test.BuildTestNode("n2", 2000, 4*units.GiB),
				test.BuildTestNode("n3", 3000, 8*units.GiB),
			},
			limits: map[string]int64{
				"cpu":    12,
				"memory": 32 * units.GiB,
				"gpu":    1,
			},
			newNode:   test.BuildTestNode("n4", 2000, 4*units.GiB),
			nodeDelta: 2,
			wantResult: &CheckDeltaResult{
				AllowedDelta: 0,
				ExceededQuotas: []ExceededQuota{
					{ID: "cluster-wide", ExceededResources: []string{"gpu"}},
				},
			},
		},
		{
			name: "node filter config allowed operation",
			nodeFilter: nodeExcludeFn(func(node *apiv1.Node) bool {
				return node.Name == "n3"
			}),
			nodes: []*apiv1.Node{
				test.BuildTestNode("n1", 1000, 2*units.GiB),
				test.BuildTestNode("n2", 2000, 4*units.GiB),
				test.BuildTestNode("n3", 3000, 8*units.GiB),
			},
			limits: map[string]int64{
				"cpu":    4,
				"memory": 8 * units.GiB,
			},
			newNode:   test.BuildTestNode("n4", 1000, 2*units.GiB),
			nodeDelta: 1,
			wantResult: &CheckDeltaResult{
				AllowedDelta: 1,
			},
		},
		{
			name: "node filter config exceeded operation",
			nodeFilter: nodeExcludeFn(func(node *apiv1.Node) bool {
				return node.Name == "n3"
			}),
			nodes: []*apiv1.Node{
				test.BuildTestNode("n1", 1000, 2*units.GiB),
				test.BuildTestNode("n2", 2000, 4*units.GiB),
				test.BuildTestNode("n3", 3000, 8*units.GiB),
			},
			limits: map[string]int64{
				"cpu":    4,
				"memory": 8 * units.GiB,
			},
			newNode:   test.BuildTestNode("n4", 2000, 4*units.GiB),
			nodeDelta: 1,
			wantResult: &CheckDeltaResult{
				AllowedDelta: 0,
				ExceededQuotas: []ExceededQuota{
					{ID: "cluster-wide", ExceededResources: []string{"cpu", "memory"}},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cloudProvider := cptest.NewTestCloudProviderBuilder().Build()
			resourceLimiter := cloudprovider.NewResourceLimiter(nil, tc.limits)
			cloudProvider.SetResourceLimiter(resourceLimiter)
			ctx := &context.AutoscalingContext{CloudProvider: cloudProvider}
			crp := tc.crp
			if crp == nil {
				crp = &fakeCustomResourcesProcessor{}
			}
			factory := NewTrackerFactory(TrackerOptions{
				CustomResourcesProcessor: crp,
				QuotaProvider:            NewCloudQuotasProvider(cloudProvider),
				NodeFilter:               tc.nodeFilter,
			})
			tracker, err := factory.NewQuotasTracker(ctx, tc.nodes)
			if err != nil {
				t.Errorf("failed to create tracker: %v", err)
			}
			var ng cloudprovider.NodeGroup
			result, err := tracker.CheckDelta(ctx, ng, tc.newNode, tc.nodeDelta)
			if err != nil {
				t.Errorf("failed to check delta: %v", err)
			}
			opts := []cmp.Option{
				cmpopts.SortSlices(func(a, b string) bool { return a < b }),
				cmpopts.EquateEmpty(),
			}
			if diff := cmp.Diff(tc.wantResult, result, opts...); diff != "" {
				t.Errorf("CheckDelta() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
