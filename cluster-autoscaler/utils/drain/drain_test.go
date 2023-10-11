/*
Copyright 2016 The Kubernetes Authors.

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

package drain

import (
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsPodLongTerminating(t *testing.T) {
	testTime := time.Date(2020, time.December, 18, 17, 0, 0, 0, time.UTC)
	twoMinGracePeriod := int64(2 * 60)
	zeroGracePeriod := int64(0)

	tests := []struct {
		name string
		pod  apiv1.Pod
		want bool
	}{
		{
			name: "No deletion timestamp",
			pod: apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: nil,
				},
				Spec: apiv1.PodSpec{
					TerminationGracePeriodSeconds: &zeroGracePeriod,
				},
			},
			want: false,
		},
		{
			name: "Just deleted no grace period defined",
			pod: apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{Time: testTime}, // default Grace Period is 30s so this pod can still be terminating
				},
				Spec: apiv1.PodSpec{
					TerminationGracePeriodSeconds: nil,
				},
			},
			want: false,
		},
		{
			name: "Deleted for longer than PodLongTerminatingExtraThreshold with no grace period",
			pod: apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{Time: testTime.Add(-3 * PodLongTerminatingExtraThreshold)},
				},
				Spec: apiv1.PodSpec{
					TerminationGracePeriodSeconds: nil,
				},
			},
			want: true,
		},
		{
			name: "Just deleted with grace period defined to 0",
			pod: apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{Time: testTime},
				},
				Spec: apiv1.PodSpec{
					TerminationGracePeriodSeconds: &zeroGracePeriod,
				},
			},
			want: false,
		},
		{
			name: "Deleted for longer than PodLongTerminatingExtraThreshold with grace period defined to 0",
			pod: apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{Time: testTime.Add(-2 * PodLongTerminatingExtraThreshold)},
				},
				Spec: apiv1.PodSpec{
					TerminationGracePeriodSeconds: &zeroGracePeriod,
				},
			},
			want: true,
		},
		{
			name: "Deleted for longer than PodLongTerminatingExtraThreshold but not longer than grace period (2 min)",
			pod: apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{Time: testTime.Add(-2 * PodLongTerminatingExtraThreshold)},
				},
				Spec: apiv1.PodSpec{
					TerminationGracePeriodSeconds: &twoMinGracePeriod,
				},
			},
			want: false,
		},
		{
			name: "Deleted for longer than grace period (2 min) and PodLongTerminatingExtraThreshold",
			pod: apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{Time: testTime.Add(-2*PodLongTerminatingExtraThreshold - time.Duration(twoMinGracePeriod)*time.Second)},
				},
				Spec: apiv1.PodSpec{
					TerminationGracePeriodSeconds: &twoMinGracePeriod,
				},
			},
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsPodLongTerminating(&tc.pod, testTime); got != tc.want {
				t.Errorf("IsPodLongTerminating() = %v, want %v", got, tc.want)
			}
		})
	}
}
