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
	apiv1 "k8s.io/api/core/v1"
	cptest "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestCalculateUsages(t *testing.T) {
	testCases := []struct {
		name          string
		nodes         []*apiv1.Node
		quotas        []Quota
		nodeFilter    func(node *apiv1.Node) bool
		customTargets map[string][]customresources.CustomResourceTarget
		wantUsages    map[string]resourceList
	}{
		{
			name: "cluster-wide limiter, no node filter",
			nodes: []*apiv1.Node{
				test.BuildTestNode("n1", 1000, 2000),
				test.BuildTestNode("n2", 2000, 4000),
				test.BuildTestNode("n3", 3000, 8000),
			},
			quotas: []Quota{
				&fakeQuota{
					id:          "cluster-wide",
					appliesToFn: includeAll,
				},
			},
			wantUsages: map[string]resourceList{
				"cluster-wide": {
					"cpu":    6,
					"memory": 14000,
					"nodes":  3,
				},
			},
		},
		{
			name: "multiple quotas",
			nodes: []*apiv1.Node{
				addLabel(test.BuildTestNode("n1", 1000, 2000), "pool", "a"),
				addLabel(test.BuildTestNode("n2", 2000, 4000), "pool", "b"),
				addLabel(test.BuildTestNode("n3", 3000, 8000), "pool", "a"),
			},
			quotas: []Quota{
				&fakeQuota{
					id:          "pool-a",
					appliesToFn: func(node *apiv1.Node) bool { return node.Labels["pool"] == "a" },
				},
				&fakeQuota{
					id:          "pool-b",
					appliesToFn: func(node *apiv1.Node) bool { return node.Labels["pool"] == "b" },
				},
			},
			wantUsages: map[string]resourceList{
				"pool-a": {
					"cpu":    4,
					"memory": 10000,
					"nodes":  2,
				},
				"pool-b": {
					"cpu":    2,
					"memory": 4000,
					"nodes":  1,
				},
			},
		},
		{
			name: "with node filter",
			nodes: []*apiv1.Node{
				test.BuildTestNode("n1", 1000, 2000),
				test.BuildTestNode("n2", 2000, 4000),
				test.BuildTestNode("n3", 3000, 8000),
			},
			quotas: []Quota{
				&fakeQuota{
					id:          "cluster-wide",
					appliesToFn: includeAll,
				},
			},
			nodeFilter: func(node *apiv1.Node) bool { return node.Name == "n2" },
			wantUsages: map[string]resourceList{
				"cluster-wide": {
					"cpu":    4,
					"memory": 10000,
					"nodes":  2,
				},
			},
		},
		{
			name: "limiter doesn't match any node",
			nodes: []*apiv1.Node{
				test.BuildTestNode("n1", 1000, 2000),
			},
			quotas: []Quota{
				&fakeQuota{
					id:          "no-match",
					appliesToFn: func(node *apiv1.Node) bool { return false },
				},
			},
			wantUsages: map[string]resourceList{
				"no-match": {},
			},
		},
		{
			name: "with custom resources",
			nodes: []*apiv1.Node{
				test.BuildTestNode("n1", 1000, 2000),
				test.BuildTestNode("n2", 2000, 4000),
			},
			quotas: []Quota{
				&fakeQuota{
					id:          "cluster-wide",
					appliesToFn: includeAll,
				},
			},
			customTargets: map[string][]customresources.CustomResourceTarget{
				"n1": {
					{ResourceType: "gpu", ResourceCount: 2},
				},
			},
			wantUsages: map[string]resourceList{
				"cluster-wide": {
					"cpu":    3,
					"memory": 6000,
					"gpu":    2,
					"nodes":  2,
				},
			},
		},
		{
			name: "multiple quotas and node filter",
			nodes: []*apiv1.Node{
				addLabel(test.BuildTestNode("n1", 1000, 2000), "pool", "a"),
				addLabel(test.BuildTestNode("n2", 2000, 4000), "pool", "b"),
				addLabel(test.BuildTestNode("n3", 3000, 8000), "pool", "a"),
			},
			quotas: []Quota{
				&fakeQuota{
					id:          "pool-a",
					appliesToFn: func(node *apiv1.Node) bool { return node.Labels["pool"] == "a" },
				},
				&fakeQuota{
					id:          "pool-b",
					appliesToFn: func(node *apiv1.Node) bool { return node.Labels["pool"] == "b" },
				},
			},
			nodeFilter: func(node *apiv1.Node) bool { return node.Name == "n3" },
			wantUsages: map[string]resourceList{
				"pool-a": {
					"cpu":    1,
					"memory": 2000,
					"nodes":  1,
				},
				"pool-b": {
					"cpu":    2,
					"memory": 4000,
					"nodes":  1,
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := cptest.NewTestCloudProviderBuilder().Build()
			ctx := &context.AutoscalingContext{CloudProvider: provider}
			crp := &fakeCustomResourcesProcessor{
				NodeResourceTargets: func(node *apiv1.Node) []customresources.CustomResourceTarget {
					if tc.customTargets == nil {
						return nil
					}
					return tc.customTargets[node.Name]
				},
			}
			var nf NodeFilter
			if tc.nodeFilter != nil {
				nf = &fakeNodeFilter{NodeFilterFn: tc.nodeFilter}
			}
			calculator := newUsageCalculator(crp, nf)
			usages, err := calculator.calculateUsages(ctx, tc.nodes, tc.quotas)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.wantUsages, usages); diff != "" {
				t.Errorf("calculateUsages() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func includeAll(node *apiv1.Node) bool {
	return true
}

func addLabel(node *apiv1.Node, key, value string) *apiv1.Node {
	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}
	node.Labels[key] = value
	return node
}
