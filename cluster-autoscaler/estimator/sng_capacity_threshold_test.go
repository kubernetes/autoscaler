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
		wantThreshold    int
	}{
		{
			"Computes capacity correctly",
			[]nodeGroupConfig{
				{"ng1", 10, 5},
				{"ng2", 100, 50},
				{"ng3", 5, 3},
			},
			57,
		},
		{
			"For empty capacity returns 0",
			[]nodeGroupConfig{
				{"ng1", 10, 10},
				{"ng2", 100, 100},
			},
			0,
		},
		{
			"Skips over-provisioned groups",
			[]nodeGroupConfig{
				{"ng1", 10, 5},
				{"ng3", 10, 11},
				{"ng3", 0, 5},
			},
			5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := testprovider.NewTestCloudProvider(func(string, int) error { return nil }, nil)
			for _, ng := range tt.nodeGroupsConfig {
				provider.AddNodeGroup(ng.name, 0, ng.maxNodes, ng.nodesCount)
			}
			context := EstimationContext{similarNodeGroups: provider.NodeGroups()}
			assert.Equalf(t, tt.wantThreshold, NewSngCapacityThreshold().GetNodeLimit(nil, &context), "NewSngCapacityThreshold()")
			assert.True(t, NewClusterCapacityThreshold().GetDurationLimit() == 0)
		})
	}
}
