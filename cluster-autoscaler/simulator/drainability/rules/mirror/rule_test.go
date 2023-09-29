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

package mirror

import (
	"testing"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
	"k8s.io/kubernetes/pkg/kubelet/types"
)

func TestRule(t *testing.T) {
	testCases := []struct {
		desc string
		pod  *apiv1.Pod
		want drainability.Status
	}{
		{
			desc: "non mirror pod",
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "regularPod",
					Namespace: "ns",
				},
			},
			want: drainability.NewUndefinedStatus(),
		},
		{
			desc: "mirror pod",
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "manifestPod",
					Namespace: "kube-system",
					Annotations: map[string]string{
						types.ConfigMirrorAnnotationKey: "something",
					},
				},
			},
			want: drainability.NewSkipStatus(),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			got := New().Drainable(nil, tc.pod)
			if tc.want != got {
				t.Errorf("Rule.Drainable(%v) = %v, want %v", tc.pod.Name, got, tc.want)
			}
		})
	}
}
