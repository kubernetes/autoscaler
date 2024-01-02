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

package longterminating

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
)

func TestDrainable(t *testing.T) {
	var (
		testTime            = time.Date(2020, time.December, 18, 17, 0, 0, 0, time.UTC)
		zeroGracePeriod     = int64(0)
		extendedGracePeriod = int64(6 * 60) // 6 minutes
	)

	for desc, tc := range map[string]struct {
		pod  *apiv1.Pod
		want drainability.Status
	}{
		"regular pod": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod",
					Namespace: "ns",
				},
			},
			want: drainability.NewUndefinedStatus(),
		},
		"long terminating pod with 0 grace period": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "bar",
					Namespace:         "default",
					DeletionTimestamp: &metav1.Time{Time: testTime.Add(drain.PodLongTerminatingExtraThreshold / 2)},
				},
				Spec: apiv1.PodSpec{
					RestartPolicy:                 apiv1.RestartPolicyOnFailure,
					TerminationGracePeriodSeconds: &zeroGracePeriod,
				},
				Status: apiv1.PodStatus{
					Phase: apiv1.PodUnknown,
				},
			},
			want: drainability.NewUndefinedStatus(),
		},
		"expired long terminating pod with 0 grace period": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "bar",
					Namespace:         "default",
					DeletionTimestamp: &metav1.Time{Time: testTime.Add(-2 * drain.PodLongTerminatingExtraThreshold)},
				},
				Spec: apiv1.PodSpec{
					RestartPolicy:                 apiv1.RestartPolicyOnFailure,
					TerminationGracePeriodSeconds: &zeroGracePeriod,
				},
				Status: apiv1.PodStatus{
					Phase: apiv1.PodUnknown,
				},
			},
			want: drainability.NewSkipStatus(),
		},
		"long terminating pod with extended grace period": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "bar",
					Namespace:         "default",
					DeletionTimestamp: &metav1.Time{Time: testTime.Add(time.Duration(extendedGracePeriod) / 2 * time.Second)},
				},
				Spec: apiv1.PodSpec{
					RestartPolicy:                 apiv1.RestartPolicyOnFailure,
					TerminationGracePeriodSeconds: &extendedGracePeriod,
				},
				Status: apiv1.PodStatus{
					Phase: apiv1.PodUnknown,
				},
			},
			want: drainability.NewUndefinedStatus(),
		},
		"expired long terminating pod with extended grace period": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "bar",
					Namespace:         "default",
					DeletionTimestamp: &metav1.Time{Time: testTime.Add(-2 * time.Duration(extendedGracePeriod) * time.Second)},
				},
				Spec: apiv1.PodSpec{
					RestartPolicy:                 apiv1.RestartPolicyOnFailure,
					TerminationGracePeriodSeconds: &extendedGracePeriod,
				},
				Status: apiv1.PodStatus{
					Phase: apiv1.PodUnknown,
				},
			},
			want: drainability.NewSkipStatus(),
		},
	} {
		t.Run(desc, func(t *testing.T) {
			drainCtx := &drainability.DrainContext{
				Timestamp: testTime,
			}
			got := New().Drainable(drainCtx, tc.pod)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Rule.Drainable(%v): got status diff (-want +got):\n%s", tc.pod.Name, diff)
			}
		})
	}
}
