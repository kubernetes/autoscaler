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

package pods

import (
	"time"

	v1 "k8s.io/api/core/v1"
)

// Summary contains information about the total observed number of pods within a
// group, number of running, as well as the count of those who failed to
// start within the deadline.
type Summary struct {
	// Total is the total observed number of pods that are either running or
	// about to be started.
	Total int32
	// Running is the number of running pods.
	Running int32
	// NotStartedWithinDeadline is the number of pods that not only has not
	// fully stared (not scheduled or not fully started, in phase PodPending)
	// but also has been in the not started phase for a while.
	NotStartedWithinDeadline int32
}

// CalculateSummary calculates the Summary structure for the given set of pod
// and startup constraint.
func CalculateSummary(podList []*v1.Pod, now time.Time, startupTimeout time.Duration) Summary {
	result := Summary{}
	for _, p := range podList {
		switch p.Status.Phase {
		case v1.PodRunning:
			result.Total++
			result.Running++
			break
		case v1.PodPending:
			result.Total++
			if p.CreationTimestamp.Add(startupTimeout).Before(now) {
				result.NotStartedWithinDeadline++
			}
			break
		}
	}
	return result
}
