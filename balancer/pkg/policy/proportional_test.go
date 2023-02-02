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

package policy

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaling/balancer/pkg/pods"
)

func TestDistributeByProportions(t *testing.T) {
	tests := []struct {
		name     string
		replicas int32
		infos    map[string]*targetInfo
		expected ReplicaPlacement
		problems PlacementProblems
	}{
		{
			name:     "1 replica, 50/50",
			replicas: 1,
			infos: map[string]*targetInfo{
				"a": {proportion: 50, max: maxReplicas},
				"b": {proportion: 50, max: maxReplicas},
			},
			expected: ReplicaPlacement{"a": 1, "b": 0},
		},
		{
			name:     "1 replica, 33/33/33",
			replicas: 1,
			infos: map[string]*targetInfo{
				"a": {proportion: 50, max: maxReplicas},
				"b": {proportion: 50, max: maxReplicas},
				"c": {proportion: 50, max: maxReplicas},
			},
			expected: ReplicaPlacement{"a": 1, "b": 0, "c": 0},
		},
		{
			name:     "2 replica, 33/33/33",
			replicas: 2,
			infos: map[string]*targetInfo{
				"a": {proportion: 50, max: maxReplicas},
				"b": {proportion: 50, max: maxReplicas},
				"c": {proportion: 50, max: maxReplicas},
			},
			expected: ReplicaPlacement{"a": 1, "b": 1, "c": 0},
		},
		{
			name:     "10 replicas, 50/50",
			replicas: 10,
			infos: map[string]*targetInfo{
				"a": {proportion: 50, max: maxReplicas},
				"b": {proportion: 50, max: maxReplicas},
			},
			expected: ReplicaPlacement{"a": 5, "b": 5},
		},
		{
			name:     "10 replicas, 70/30",
			replicas: 10,
			infos: map[string]*targetInfo{
				"a": {proportion: 70, max: maxReplicas},
				"b": {proportion: 30, max: maxReplicas},
			},
			expected: ReplicaPlacement{"a": 7, "b": 3},
		},
		{
			name:     "100 replicas, 70/30",
			replicas: 100,
			infos: map[string]*targetInfo{
				"a": {proportion: 70, max: maxReplicas},
				"b": {proportion: 30, max: maxReplicas},
			},
			expected: ReplicaPlacement{"a": 70, "b": 30},
		},
		{
			name:     "11 replicas, 50/50, stability test",
			replicas: 11,
			infos: map[string]*targetInfo{
				"a": {proportion: 50, max: maxReplicas},
				"b": {proportion: 50, max: maxReplicas},
			},
			expected: ReplicaPlacement{"a": 6, "b": 5},
		},
		{
			name:     "10 replicas, 50/50, one max",
			replicas: 10,
			infos: map[string]*targetInfo{
				"a": {proportion: 50, max: 3},
				"b": {proportion: 50, max: maxReplicas},
			},
			expected: ReplicaPlacement{"a": 3, "b": 7},
		},
		{
			name:     "10 replicas, 50/50, two max",
			replicas: 10,
			infos: map[string]*targetInfo{
				"a": {proportion: 50, max: 3},
				"b": {proportion: 50, max: 2},
			},
			expected: ReplicaPlacement{"a": 3, "b": 2},
			problems: PlacementProblems{OverflowReplicas: 5},
		},
		{
			name:     "10 replicas, 50/50, one small min",
			replicas: 10,
			infos: map[string]*targetInfo{
				"a": {proportion: 50, min: 3, max: maxReplicas},
				"b": {proportion: 50, max: maxReplicas},
			},
			expected: ReplicaPlacement{"a": 5, "b": 5},
		},
		{
			name:     "10 replicas, 50/50, one big min",
			replicas: 10,
			infos: map[string]*targetInfo{
				"a": {proportion: 50, min: 7, max: maxReplicas},
				"b": {proportion: 50, max: maxReplicas},
			},
			expected: ReplicaPlacement{"a": 7, "b": 3},
		},
		{
			name:     "10 replicas, 50/50, one huge min",
			replicas: 10,
			infos: map[string]*targetInfo{
				"a": {proportion: 50, min: 20, max: maxReplicas},
				"b": {proportion: 50, max: maxReplicas},
			},
			expected: ReplicaPlacement{"a": 20, "b": 0},
			problems: PlacementProblems{MissingReplicas: 10},
		},
		{
			name:     "10 replicas, 50/50, two small min",
			replicas: 10,
			infos: map[string]*targetInfo{
				"a": {proportion: 50, min: 3, max: maxReplicas},
				"b": {proportion: 50, min: 4, max: maxReplicas},
			},
			expected: ReplicaPlacement{"a": 5, "b": 5},
		},
		{
			name:     "10 replicas, 50/50, two small min, one max",
			replicas: 10,
			infos: map[string]*targetInfo{
				"a": {proportion: 50, min: 3, max: 3},
				"b": {proportion: 50, min: 4, max: maxReplicas},
			},
			expected: ReplicaPlacement{"a": 3, "b": 7},
		},
		{
			name:     "10 replicas, 50/50, two small min, two max",
			replicas: 10,
			infos: map[string]*targetInfo{
				"a": {proportion: 50, min: 3, max: 3},
				"b": {proportion: 50, min: 4, max: 6},
			},
			expected: ReplicaPlacement{"a": 3, "b": 6},
			problems: PlacementProblems{OverflowReplicas: 1},
		},
		{
			name:     "20 replicas, 70/20/10, one max",
			replicas: 20,
			infos: map[string]*targetInfo{
				"a": {proportion: 70, max: 5},
				"b": {proportion: 20, max: maxReplicas},
				"c": {proportion: 10, max: maxReplicas},
			},
			expected: ReplicaPlacement{"a": 5, "b": 10, "c": 5},
		},
		{
			name:     "20 replicas, 48/48/4, two small min, two max",
			replicas: 20,
			infos: map[string]*targetInfo{
				"a": {proportion: 48, min: 2, max: 5},
				"b": {proportion: 48, min: 2, max: 5},
				"c": {proportion: 4, min: 2, max: maxReplicas},
			},
			expected: ReplicaPlacement{"a": 5, "b": 5, "c": 10},
		},
		{
			name:     "10 replicas, 50/50, fallback of 2/5 ",
			replicas: 10,
			infos: map[string]*targetInfo{
				"a": {proportion: 50, min: 0, max: maxReplicas, summary: pods.Summary{
					Total:                    5,
					NotStartedWithinDeadline: 2,
				}},
				"b": {proportion: 50, min: 0, max: maxReplicas},
			},
			expected: ReplicaPlacement{"a": 5, "b": 7},
		},
		{
			name:     "6 replicas, 50/50, fallback in min ",
			replicas: 6,
			infos: map[string]*targetInfo{
				"a": {proportion: 50, min: 3, max: maxReplicas, summary: pods.Summary{
					Total:                    3,
					NotStartedWithinDeadline: 3,
				}},
				"b": {proportion: 50, min: 3, max: maxReplicas},
			},
			expected: ReplicaPlacement{"a": 3, "b": 6},
		},
		{
			name:     "10 replicas, 50/50, fallback of 2/3 + 2 that are likely to fail ",
			replicas: 10,
			infos: map[string]*targetInfo{
				"a": {proportion: 50, min: 0, max: maxReplicas, summary: pods.Summary{
					Total:                    3,
					NotStartedWithinDeadline: 2,
				}},
				"b": {proportion: 50, min: 0, max: maxReplicas},
			},
			expected: ReplicaPlacement{"a": 5, "b": 9},
		},
		{
			name:     "18 replicas, 50/50, with max, fallback of 2/9",
			replicas: 18,
			infos: map[string]*targetInfo{
				"a": {proportion: 50, min: 0, max: 10, summary: pods.Summary{
					Total:                    9,
					NotStartedWithinDeadline: 2,
				}},
				"b": {proportion: 50, min: 0, max: 10},
			},
			expected: ReplicaPlacement{"a": 9, "b": 10},
			problems: PlacementProblems{OverflowReplicas: 1},
		},
		{
			name:     "20 replicas, 50/50, with imbalance and fallback",
			replicas: 20,
			infos: map[string]*targetInfo{
				"a": {proportion: 50, min: 0, max: maxReplicas, summary: pods.Summary{
					Total:                    15,
					NotStartedWithinDeadline: 3,
				}},
				"b": {proportion: 50, min: 0, max: maxReplicas},
			},
			expected: ReplicaPlacement{"a": 10, "b": 10},
		},
		{
			name:     "20 replicas, 50/50, with too few and fallback",
			replicas: 20,
			infos: map[string]*targetInfo{
				"a": {proportion: 50, min: 0, max: maxReplicas, summary: pods.Summary{
					Total:                    3,
					NotStartedWithinDeadline: 3,
				}},
				"b": {proportion: 50, min: 0, max: maxReplicas},
			},
			expected: ReplicaPlacement{"a": 10, "b": 20},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d: %s", i, tc.name), func(t *testing.T) {
			result, problems := distributeByProportions(tc.replicas, tc.infos)
			assert.Equal(t, tc.expected, result)
			assert.Equal(t, tc.problems, problems)
		})
	}
}
