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

func TestDistributeByPriority(t *testing.T) {
	tests := []struct {
		name       string
		replicas   int32
		infos      map[string]*targetInfo
		priorities []string
		expected   ReplicaPlacement
		problems   PlacementProblems
	}{
		{
			name:     "10 replicas, no max",
			replicas: 10,
			infos: map[string]*targetInfo{
				"a": {max: maxReplicas},
				"b": {max: maxReplicas},
			},
			priorities: []string{"a", "b"},
			expected:   ReplicaPlacement{"a": 10, "b": 0},
		},
		{
			name:     "10 replicas, one max",
			replicas: 10,
			infos: map[string]*targetInfo{
				"a": {max: 3},
				"b": {max: maxReplicas},
			},
			priorities: []string{"a", "b"},
			expected:   ReplicaPlacement{"a": 3, "b": 7},
		},
		{
			name:     "10 replicas, two max",
			replicas: 10,
			infos: map[string]*targetInfo{
				"a": {max: 3},
				"b": {max: 4},
			},
			priorities: []string{"a", "b"},
			expected:   ReplicaPlacement{"a": 3, "b": 4},
			problems:   PlacementProblems{OverflowReplicas: 3},
		},
		{
			name:     "10 replicas, two mins",
			replicas: 10,
			infos: map[string]*targetInfo{
				"a": {min: 2, max: maxReplicas},
				"b": {min: 3, max: maxReplicas},
			},
			priorities: []string{"a", "b"},
			expected:   ReplicaPlacement{"a": 7, "b": 3},
		},
		{
			name:     "1 replica, missing 4",
			replicas: 1,
			infos: map[string]*targetInfo{
				"a": {min: 2, max: maxReplicas},
				"b": {min: 3, max: maxReplicas},
			},
			priorities: []string{"a", "b"},
			expected:   ReplicaPlacement{"a": 2, "b": 3},
			problems:   PlacementProblems{MissingReplicas: 4},
		},
		{
			name:     "10 replicas, two mins, two max",
			replicas: 10,
			infos: map[string]*targetInfo{
				"a": {min: 2, max: 4},
				"b": {min: 3, max: 5},
			},
			priorities: []string{"a", "b"},
			expected:   ReplicaPlacement{"a": 4, "b": 5},
			problems:   PlacementProblems{OverflowReplicas: 1},
		},
		{
			name:     "10 replicas, fallback",
			replicas: 10,
			infos: map[string]*targetInfo{
				"a": {max: maxReplicas,
					summary: pods.Summary{
						Total: 3, NotStartedWithinDeadline: 2}},
				"b": {max: maxReplicas},
			},
			priorities: []string{"a", "b"},
			expected:   ReplicaPlacement{"a": 10, "b": 9},
		},
		{
			name:     "6 replica, fallback in min",
			replicas: 6,
			infos: map[string]*targetInfo{
				"a": {min: 3, max: maxReplicas,
					summary: pods.Summary{
						Total: 3, NotStartedWithinDeadline: 3}},
				"b": {min: 3, max: maxReplicas},
			},
			priorities: []string{"a", "b"},
			expected:   ReplicaPlacement{"a": 3, "b": 6},
		},
		{
			name:     "10 replicas, double fallback",
			replicas: 10,
			infos: map[string]*targetInfo{
				"a": {max: maxReplicas,
					summary: pods.Summary{
						Total: 3, NotStartedWithinDeadline: 2}},
				"b": {max: maxReplicas,
					summary: pods.Summary{
						Total: 3, NotStartedWithinDeadline: 2}},
			},
			priorities: []string{"a", "b"},
			expected:   ReplicaPlacement{"a": 10, "b": 9},
			// one is running in both "a" and "b", rest is problematic.
			problems: PlacementProblems{OverflowReplicas: 8},
		},
		{
			name:     "20 replica, double fallback hitting max",
			replicas: 20,
			infos: map[string]*targetInfo{
				"a": {max: 10,
					summary: pods.Summary{
						Total: 3, NotStartedWithinDeadline: 2}},
				"b": {max: 10,
					summary: pods.Summary{
						Total: 3, NotStartedWithinDeadline: 2}},
			},
			priorities: []string{"a", "b"},
			expected:   ReplicaPlacement{"a": 10, "b": 10},
			// one is running in both "a" and "b", rest is problematic.
			problems: PlacementProblems{OverflowReplicas: 18},
		},
		{
			name:     "10 replicas, double problems (policy change) but hopefully problematic pods are dropped first",
			replicas: 10,
			infos: map[string]*targetInfo{
				"a": {max: 10,
					summary: pods.Summary{
						Total: 15, NotStartedWithinDeadline: 2}},
				"b": {max: 10,
					summary: pods.Summary{
						Total: 15, NotStartedWithinDeadline: 2}},
			},
			priorities: []string{"a", "b"},
			expected:   ReplicaPlacement{"a": 10, "b": 0},
		},
		{
			name:     "10 replicas, single fallback",
			replicas: 10,
			infos: map[string]*targetInfo{
				"a": {max: maxReplicas,
					summary: pods.Summary{
						Total: 3, NotStartedWithinDeadline: 2}},
				"b": {max: maxReplicas,
					summary: pods.Summary{
						Total: 3, NotStartedWithinDeadline: 0}},
			},
			priorities: []string{"a", "b"},
			expected:   ReplicaPlacement{"a": 10, "b": 9},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d: %s", i, tc.name), func(t *testing.T) {
			result, problems := distributeByPriority(tc.replicas, tc.priorities, tc.infos)
			assert.Equal(t, tc.expected, result)
			assert.Equal(t, tc.problems, problems)
		})
	}
}
