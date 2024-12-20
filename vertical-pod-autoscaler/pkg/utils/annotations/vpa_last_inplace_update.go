/*
Copyright 2020 The Kubernetes Authors.

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

package annotations

import (
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
)

const (
	// VpaLastInPlaceUpdateTime is a label used by the vpa last in place update annotation.
	VpaLastInPlaceUpdateTime = "vpaLastInPlaceUpdateTime"
)

// GetVpaLastInPlaceUpdatedTimeValue creates an annotation value for a given pod.
func GetVpaLastInPlaceUpdatedTimeValue(lastUpdated time.Time) string {
	// return string version of v1.Time
	return lastUpdated.Format(time.RFC3339)
}

// ParseVpaLastInPlaceUpdateTimeValue returns the last in-place update time from a pod annotation.
func ParseVpaLastInPlaceUpdateTimeValue(pod *apiv1.Pod) (time.Time, error) {
	if pod == nil {
		return time.Time{}, fmt.Errorf("pod is nil")
	}

	if pod.Annotations == nil {
		return time.Time{}, fmt.Errorf("pod.Annotations is nil")
	}

	lastInPlaceUpdateTime, ok := pod.Annotations[VpaLastInPlaceUpdateTime]
	if !ok {
		return time.Time{}, fmt.Errorf("pod.Annotations[VpaLastInPlaceUpdateTime] is nil")
	}

	return time.Parse(time.RFC3339, lastInPlaceUpdateTime)
}
