/*
Copyright 2017 The Kubernetes Authors.

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

package pod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/kubernetes/pkg/kubelet/types"
)

func TestIsDaemonSetPod(t *testing.T) {
	tests := []struct {
		name string
		pod  *apiv1.Pod
		want bool
	}{
		{
			name: "Pod with empty OwnerRef",
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
			},
			want: false,
		},
		{
			name: "Pod with empty OwnerRef.Kind == DaemonSet",
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
					OwnerReferences: []metav1.OwnerReference{
						{
							Controller: newBool(true),
							Kind:       "DaemonSet",
						},
					},
				},
			},
			want: true,
		},
		{
			name: "Pod with annotation but not with `DaemonSetPodAnnotationKey`",
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "foo",
					Namespace:   "bar",
					Annotations: make(map[string]string),
				},
			},
			want: false,
		},
		{
			name: "Pod with `DaemonSetPodAnnotationKey` annotation but wrong value",
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
					Annotations: map[string]string{
						DaemonSetPodAnnotationKey: "bad value",
					},
				},
			},
			want: false,
		},
		{
			name: "Pod with `DaemonSetPodAnnotationKey:true` annotation",
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
					Annotations: map[string]string{
						DaemonSetPodAnnotationKey: "true",
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDaemonSetPod(tt.pod); got != tt.want {
				t.Errorf("IsDaemonSetPod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newBool(b bool) *bool {
	return &b
}

func TestIsMirrorPod(t *testing.T) {
	tests := []struct {
		name string
		pod  *apiv1.Pod
		want bool
	}{
		{
			name: "pod with ConfigMirrorAnnotationKey",
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
					Annotations: map[string]string{
						types.ConfigMirrorAnnotationKey: "",
					},
				},
			},
			want: true,
		},
		{
			name: "pod with without ConfigMirrorAnnotationKey",
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
					Annotations: map[string]string{
						types.ConfigMirrorAnnotationKey: "",
					},
				},
			},
			want: true,
		},
		{
			name: "pod with nil Annotations map",
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsMirrorPod(tt.pod); got != tt.want {
				t.Errorf("IsMirrorPod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsStaticPod(t *testing.T) {
	tests := []struct {
		name string
		pod  *apiv1.Pod
		want bool
	}{
		{
			name: "not a static pod",
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
				},
			},
			want: false,
		},
		{
			name: "is a static pod",
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "bar",
					Annotations: map[string]string{
						types.ConfigSourceAnnotationKey: types.FileSource,
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsStaticPod(tt.pod); got != tt.want {
				t.Errorf("IsStaticPod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterRecreatablePods(t *testing.T) {
	testCases := []struct {
		name     string
		pods     []*apiv1.Pod
		wantPods []*apiv1.Pod
	}{
		{
			name: "no pods",
		},
		{
			name:     "keep single pod",
			pods:     []*apiv1.Pod{BuildTestPod("p", 100, 1)},
			wantPods: []*apiv1.Pod{BuildTestPod("p", 100, 1)},
		},
		{
			name:     "keep single RS pod",
			pods:     []*apiv1.Pod{SetRSPodSpec(BuildTestPod("p", 100, 1), "rs")},
			wantPods: []*apiv1.Pod{SetRSPodSpec(BuildTestPod("p", 100, 1), "rs")},
		},
		{
			name: "filter-out single DS pod",
			pods: []*apiv1.Pod{SetDSPodSpec(BuildTestPod("p", 100, 1))},
		},
		{
			name: "filter-out single mirror pod",
			pods: []*apiv1.Pod{SetMirrorPodSpec(BuildTestPod("p", 100, 1))},
		},
		{
			name: "filter-out single static pod",
			pods: []*apiv1.Pod{SetStaticPodSpec(BuildTestPod("p", 100, 1))},
		},
		{
			name: "all pods together",
			pods: []*apiv1.Pod{
				BuildTestPod("p1", 100, 1),
				SetRSPodSpec(BuildTestPod("p2", 100, 1), "rs"),
				SetDSPodSpec(BuildTestPod("p3", 100, 1)),
				SetMirrorPodSpec(BuildTestPod("p4", 100, 1)),
				SetStaticPodSpec(BuildTestPod("p5", 100, 1)),
			},
			wantPods: []*apiv1.Pod{
				BuildTestPod("p1", 100, 1),
				SetRSPodSpec(BuildTestPod("p2", 100, 1), "rs"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.ElementsMatch(t, tc.wantPods, FilterRecreatablePods(tc.pods))
		})
	}
}

func TestClearPodNodeNames(t *testing.T) {
	testCases := []struct {
		name string
		pods []*apiv1.Pod
	}{
		{
			name: "no pods",
		},
		{
			name: "single scheduled pod",
			pods: []*apiv1.Pod{BuildScheduledTestPod("p", 100, 1, "n")},
		},
		{
			name: "single not scheduled pod",
			pods: []*apiv1.Pod{BuildTestPod("p", 100, 1)},
		},
		{
			name: "mixed scheduled and not scheduled pod",
			pods: []*apiv1.Pod{
				BuildScheduledTestPod("p", 100, 1, "n"),
				BuildTestPod("p", 100, 1),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i, cleanedPod := range ClearPodNodeNames(tc.pods) {
				assert.Equal(t, "", cleanedPod.Spec.NodeName)
				// check if pods are otherwise the same
				cleanedPod.Spec.NodeName = tc.pods[i].Spec.NodeName
				assert.Equal(t, tc.pods[i], cleanedPod)
			}
		})
	}
}
