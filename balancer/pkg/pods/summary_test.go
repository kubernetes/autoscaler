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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newPod(name string, phase v1.PodPhase, createTime time.Time) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			Namespace:         "default",
			CreationTimestamp: metav1.NewTime(createTime),
		},
		Status: v1.PodStatus{
			Phase: phase,
		},
	}
}

func TestCalculateSummary(t *testing.T) {
	now := time.Now()
	timeout := time.Minute * 5

	tests := []struct {
		name     string
		pods     []*v1.Pod
		expected Summary
	}{
		{
			name: "basic running/pending",
			pods: []*v1.Pod{
				newPod("a", v1.PodRunning, now.Add(-time.Hour)),
				newPod("a2", v1.PodRunning, now.Add(-time.Hour)),
				newPod("b", v1.PodPending, now.Add(-time.Minute)),
			},
			expected: Summary{
				Total:   3,
				Running: 2,
			},
		},
		{
			name: "skip completed",
			pods: []*v1.Pod{
				newPod("a", v1.PodSucceeded, now.Add(-time.Hour)),
				newPod("c", v1.PodFailed, now.Add(-time.Hour)),
			},
			expected: Summary{
				Total: 0,
			},
		},
		{
			name: "deadline",
			pods: []*v1.Pod{
				newPod("a", v1.PodPending, now.Add(-time.Hour)),
			},
			expected: Summary{
				Total:                    1,
				NotStartedWithinDeadline: 1,
			},
		},
		{
			name: "mix",
			pods: []*v1.Pod{
				newPod("problem", v1.PodPending, now.Add(-time.Hour)),
				newPod("b", v1.PodRunning, now.Add(-time.Hour)),
				newPod("b2", v1.PodRunning, now.Add(-time.Hour)),
				newPod("c", v1.PodFailed, now.Add(-time.Hour)),
			},
			expected: Summary{
				Total:                    3,
				Running:                  2,
				NotStartedWithinDeadline: 1,
			},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d: %s", i, tc.name), func(t *testing.T) {
			result := CalculateSummary(tc.pods, now, timeout)
			assert.Equal(t, tc.expected, result)
		})
	}
}
