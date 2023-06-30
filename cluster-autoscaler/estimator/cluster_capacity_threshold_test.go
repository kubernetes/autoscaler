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
			name:                "computes available capacity",
			wantThreshold:       5,
			contextMaxNodes:     10,
			contextCurrentNodes: 5,
		},
		{
			name:                "for not defined cluster max nodes returns 0",
			wantThreshold:       0,
			contextMaxNodes:     0,
			contextCurrentNodes: 10,
		},
		{
			name:                "for negative total capacity returns 0",
			wantThreshold:       0,
			contextMaxNodes:     5,
			contextCurrentNodes: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := &EstimationContext{
				similarNodeGroups:   nil,
				currentNodeCount:    tt.contextCurrentNodes,
				nodeGroupCapacity:   0,
				clusterMaxNodeLimit: tt.contextMaxNodes,
			}
			assert.Equal(t, tt.wantThreshold, NewClusterCapacityThreshold().GetNodeLimit(nil, context))
			assert.True(t, NewClusterCapacityThreshold().GetDurationLimit() == 0)
		})
	}
}
