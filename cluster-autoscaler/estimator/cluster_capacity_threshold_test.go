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
)

func TestNewClusterCapacityThreshold(t *testing.T) {
	tests := []struct {
		name                string
		wantThreshold       int
		contextMaxNodes     int
		contextCurrentNodes int
	}{
		{
			name:                "returns available capacity",
			contextMaxNodes:     10,
			contextCurrentNodes: 5,
			wantThreshold:       5,
		},
		{
			name:                "no threshold is set if cluster capacity is unlimited",
			contextMaxNodes:     0,
			contextCurrentNodes: 10,
			wantThreshold:       0,
		},
		{
			name:                "threshold is negative if cluster has no capacity",
			contextMaxNodes:     5,
			contextCurrentNodes: 10,
			wantThreshold:       -1,
		},
		{
			name:                "threshold is negative if cluster node limit is negative",
			contextMaxNodes:     -5,
			contextCurrentNodes: 0,
			wantThreshold:       -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := &estimationContext{
				similarNodeGroups:   nil,
				currentNodeCount:    tt.contextCurrentNodes,
				clusterMaxNodeLimit: tt.contextMaxNodes,
			}
			assert.Equal(t, tt.wantThreshold, NewClusterCapacityThreshold().NodeLimit(nil, context))
			assert.True(t, NewClusterCapacityThreshold().DurationLimit(nil, nil) == 0)
		})
	}
}
