/*
Copyright 2026 The Kubernetes Authors.

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

package skipreschedule

import (
	"testing"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
)

func TestDrainable(t *testing.T) {
	testCases := []struct {
		name    string
		pod     *apiv1.Pod
		wantOut drainability.OutcomeType
	}{
		{
			name: "pod with skip-reschedule annotation",
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod",
					Namespace: "ns",
					Annotations: map[string]string{
						SkipRescheduleAnnotationKey: "true",
					},
				},
			},
			wantOut: drainability.SkipDrain,
		},
		{
			name: "pod with skip-reschedule annotation set to false",
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod",
					Namespace: "ns",
					Annotations: map[string]string{
						SkipRescheduleAnnotationKey: "false",
					},
				},
			},
			wantOut: drainability.UndefinedOutcome,
		},
		{
			name: "pod without skip-reschedule annotation",
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod",
					Namespace: "ns",
				},
			},
			wantOut: drainability.UndefinedOutcome,
		},
		{
			name: "pod with other annotations but not skip-reschedule",
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod",
					Namespace: "ns",
					Annotations: map[string]string{
						"some-other-annotation": "value",
					},
				},
			},
			wantOut: drainability.UndefinedOutcome,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rule := New()
			got := rule.Drainable(nil, tc.pod, nil)
			if got.Outcome != tc.wantOut {
				t.Errorf("Drainable() = %v, want %v", got.Outcome, tc.wantOut)
			}
		})
	}
}

func TestHasSkipRescheduleAnnotation(t *testing.T) {
	testCases := []struct {
		name string
		pod  *apiv1.Pod
		want bool
	}{
		{
			name: "has annotation set to true",
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						SkipRescheduleAnnotationKey: "true",
					},
				},
			},
			want: true,
		},
		{
			name: "has annotation set to false",
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						SkipRescheduleAnnotationKey: "false",
					},
				},
			},
			want: false,
		},
		{
			name: "no annotations",
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{},
			},
			want: false,
		},
		{
			name: "nil annotations",
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: nil,
				},
			},
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := HasSkipRescheduleAnnotation(tc.pod)
			if got != tc.want {
				t.Errorf("HasSkipRescheduleAnnotation() = %v, want %v", got, tc.want)
			}
		})
	}
}
