/*
Copyright The Kubernetes Authors.

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

package scheduling

import (
	"sort"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/karpenter/pkg/utils/resources"
)

// Queue is a queue of pods that is scheduled.  It's used to attempt to schedule pods as long as we are making progress
// in scheduling. This is sometimes required to maintain zonal topology spreads with constrained pods, and can satisfy
// pod affinities that occur in a batch of pods if there are enough constraints provided.
type Queue struct {
	pods    []*v1.Pod
	lastLen map[types.UID]int
}

// NewQueue constructs a new queue given the input pods, sorting them to optimize for bin-packing into nodes.
func NewQueue(pods []*v1.Pod, podData map[types.UID]*PodData) *Queue {
	sort.Slice(pods, byCPUAndMemoryDescending(pods, podData))
	return &Queue{
		pods:    pods,
		lastLen: map[types.UID]int{},
	}
}

// Pop returns the next pod or false if no longer making progress
func (q *Queue) Pop() (*v1.Pod, bool) {
	if len(q.pods) == 0 {
		return nil, false
	}
	p := q.pods[0]

	// If we are about to pop a pod when it was last pushed with the same number of pods in the queue, then
	// we've cycled through all pods in the queue without making progress and can stop
	if q.lastLen[p.UID] == len(q.pods) {
		return nil, false
	}

	q.pods = q.pods[1:]
	return p, true
}

// Push a pod onto the queue, counting each time a pod is immediately requeued. This is used to detect staleness.
func (q *Queue) Push(pod *v1.Pod) {
	q.pods = append(q.pods, pod)
	q.lastLen[pod.UID] = len(q.pods)
}

func (q *Queue) List() []*v1.Pod {
	return q.pods
}

func byCPUAndMemoryDescending(pods []*v1.Pod, podData map[types.UID]*PodData) func(i int, j int) bool {
	return func(i, j int) bool {
		lhsPod := pods[i]
		rhsPod := pods[j]

		lhs := podData[lhsPod.UID].Requests
		rhs := podData[rhsPod.UID].Requests

		cpuCmp := resources.Cmp(lhs[v1.ResourceCPU], rhs[v1.ResourceCPU])
		if cpuCmp < 0 {
			// LHS has less CPU, so it should be sorted after
			return false
		} else if cpuCmp > 0 {
			return true
		}
		memCmp := resources.Cmp(lhs[v1.ResourceMemory], rhs[v1.ResourceMemory])

		if memCmp < 0 {
			return false
		} else if memCmp > 0 {
			return true
		}

		// If all else is equal, give a consistent ordering. This reduces the number of NominatePod events as we
		// de-duplicate those based on identical content.

		// unfortunately creation timestamp only has a 1-second resolution, so we would still re-order pods created
		// during a deployment scale-up if we only looked at creation time
		if lhsPod.CreationTimestamp != rhsPod.CreationTimestamp {
			return lhsPod.CreationTimestamp.Before(&rhsPod.CreationTimestamp)
		}

		// pod UIDs aren't in any order, but since we first sort by creation time this only serves to consistently order
		// pods created within the same second
		return lhsPod.UID < rhsPod.UID
	}
}
