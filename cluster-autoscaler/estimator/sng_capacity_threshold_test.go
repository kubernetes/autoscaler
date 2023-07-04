/*
Copyright 2023 The Kubernetes Authors.

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

package estimator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
)

func TestSngCapacityThreshold(t *testing.T) {
	type nodeGroupConfig struct {
		name       string
		maxNodes   int
		nodesCount int
	}
	tests := []struct {
		name             string
		nodeGroupsConfig []nodeGroupConfig
		currentNodeGroup nodeGroupConfig
		wantThreshold    int
	}{
		{
			name: "returns available capacity",
			nodeGroupsConfig: []nodeGroupConfig{
				{name: "ng1", maxNodes: 10, nodesCount: 5},
				{name: "ng2", maxNodes: 100, nodesCount: 50},
				{name: "ng3", maxNodes: 5, nodesCount: 3},
			},
			currentNodeGroup: nodeGroupConfig{name: "main-ng", maxNodes: 20, nodesCount: 10},
			wantThreshold:    67,
		},
		{
			name: "returns available capacity and skips over-provisioned groups",
			nodeGroupsConfig: []nodeGroupConfig{
				{name: "ng1", maxNodes: 10, nodesCount: 5},
				{name: "ng3", maxNodes: 10, nodesCount: 11},
				{name: "ng3", maxNodes: 0, nodesCount: 5},
			},
			currentNodeGroup: nodeGroupConfig{name: "main-ng", maxNodes: 5, nodesCount: 10},
			wantThreshold:    5,
		},
		{
			name: "threshold is negative if cluster has no capacity",
			nodeGroupsConfig: []nodeGroupConfig{
				{name: "ng1", maxNodes: 10, nodesCount: 10},
				{name: "ng2", maxNodes: 100, nodesCount: 100},
			},
			currentNodeGroup: nodeGroupConfig{name: "main-ng", maxNodes: 5, nodesCount: 5},
			wantThreshold:    -1,
		},
		{
			name: "threshold is negative if all groups are over-provisioned",
			nodeGroupsConfig: []nodeGroupConfig{
				{name: "ng1", maxNodes: 10, nodesCount: 11},
				{name: "ng3", maxNodes: 100, nodesCount: 111},
				{name: "ng3", maxNodes: 0, nodesCount: 5},
			},
			currentNodeGroup: nodeGroupConfig{name: "main-ng", maxNodes: 5, nodesCount: 10},
			wantThreshold:    -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := testprovider.NewTestCloudProvider(func(string, int) error { return nil }, nil)
			for _, ng := range tt.nodeGroupsConfig {
				provider.AddNodeGroup(ng.name, 0, ng.maxNodes, ng.nodesCount)
			}
			// Context must be constructed first to exclude current node group passed from orchestrator
			context := estimationContext{similarNodeGroups: provider.NodeGroups()}
			provider.AddNodeGroup(tt.currentNodeGroup.name, 0, tt.currentNodeGroup.maxNodes, tt.currentNodeGroup.nodesCount)
			currentNodeGroup := provider.GetNodeGroup(tt.currentNodeGroup.name)
			assert.Equalf(t, tt.wantThreshold, NewSngCapacityThreshold().NodeLimit(currentNodeGroup, &context), "NewSngCapacityThreshold()")
			assert.True(t, NewClusterCapacityThreshold().DurationLimit(currentNodeGroup, &context) == 0)
		})
	}
}
