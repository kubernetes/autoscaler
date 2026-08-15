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
	"sort"
)

// Returns a sorted subset of the given keys that has some available
// capacity (placement < max). Sorting is required for the stability
// of the algorithm - every time we should process targets in the same
// order so replicas are not flapping between targets.
func sortedKeysWithCapacity(keys []string, placement ReplicaPlacement,
	infos map[string]*targetInfo) []string {

	result := make([]string, 0)
	for _, k := range keys {
		allocated := placement[k]
		if infos[k].max <= allocated {
			continue
		}
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

// Main algorithm of the proportional policy. The function returns the
// desired replica placement and information about problems.
func distributeByProportions(replicas int32,
	infos map[string]*targetInfo) (ReplicaPlacement, PlacementProblems) {

	placement := make(ReplicaPlacement)
	problems := PlacementProblems{}
	keys := make([]string, 0)
	for k := range infos {
		keys = append(keys, k)
	}

	// Place target minimums.
	for _, k := range keys {
		info := infos[k]
		placement[k] = info.min
		replicas -= placement[k]
	}
	// continue computations as there still may be fallbacks.
	if replicas < 0 {
		problems.MissingReplicas = -replicas
		replicas = 0
	}

	// Do the fist pass with proportions, ignoring not started replicas.
	replicas = distributeGroupProportionally(replicas, keys, infos, placement)

	// All targets are full. Not possible to place more, so no fallback.
	if replicas > 0 {
		problems.OverflowReplicas = replicas
		return placement, problems
	}

	// Checks which targets have troubles and calculates the number
	// of replicas that need to be duplicated in other targets until
	// the base targets recover.
	notBlockedKeys := make([]string, 0)
	for _, key := range keys {
		info := infos[key]
		if info.summary.NotStartedWithinDeadline > 0 {
			fallback := info.summary.NotStartedWithinDeadline + placement[key] - info.summary.Total
			if fallback > 0 {
				replicas += fallback
			}
		} else {
			notBlockedKeys = append(notBlockedKeys, key)
		}
	}

	// If there are some replicas that need to be duplicated, distribute them,
	// but only among non-problematic targets.
	if replicas > 0 {
		replicas = distributeGroupProportionally(replicas, notBlockedKeys, infos, placement)
	}
	problems.OverflowReplicas = replicas
	return placement, problems
}

// Distributes replicas among the given target keys, using the provided proportions.
// The passed replicaPlacement is updated and the return value is the number
// of replicas that could not be placed. It uses D'Hondt method for distributing
// replicas. See https://en.wikipedia.org/wiki/D%27Hondt_method.
func distributeGroupProportionally(replicas int32, keys []string,
	infos map[string]*targetInfo, placement ReplicaPlacement) int32 {
	okKeys := sortedKeysWithCapacity(keys, placement, infos)

	for ; replicas > 0; replicas-- {
		bestKey := ""
		bestValue := 0.0
		for _, k := range okKeys {
			if placement[k] >= infos[k].max {
				continue
			}
			rank := float64(infos[k].proportion) / float64((1 + placement[k]))
			if rank > bestValue {
				bestKey = k
				bestValue = rank
			}
		}
		if bestKey == "" {
			break
		}
		placement[bestKey] += 1
	}
	return replicas
}
