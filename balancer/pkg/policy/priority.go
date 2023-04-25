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

// Main algorithm of the priority policy. The function returns the
// desired replica placement and information about problems that
// possibly happened during placement.
func distributeByPriority(replicas int32,
	priorities []string, infos map[string]*targetInfo) (ReplicaPlacement, PlacementProblems) {

	placement := make(ReplicaPlacement)
	problems := PlacementProblems{}

	// Place target minimums
	for k, info := range infos {
		placement[k] = info.min
		replicas -= placement[k]
	}
	// continue computations as there still may be fallbacks.
	if replicas < 0 {
		problems.MissingReplicas = -replicas
		replicas = 0
	}

	for _, key := range priorities {
		info := infos[key]
		free := info.max - placement[key]
		if replicas < free {
			placement[key] += replicas
			replicas = 0
		} else {
			placement[key] += free
			replicas -= free
		}
		// calculate how many may need to fall back to the other target = all new plus
		// and those that are past deadline.
		if infos[key].summary.NotStartedWithinDeadline > 0 {
			fallback := info.summary.NotStartedWithinDeadline + placement[key] - info.summary.Total
			if fallback > 0 {
				replicas += fallback
			}
		}
	}
	if replicas > 0 {
		problems.OverflowReplicas = replicas
	}
	return placement, problems
}
