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

package oncompletion

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestDrainable(t *testing.T) {
	tests := []struct {
		name string
		pod  *apiv1.Pod
		want drainability.Status
	}{
		{
			name: "regular pod",
			pod:  BuildTestPod("pod1", 100, 0),
			want: drainability.NewUndefinedStatus(),
		},
		{
			name: "safe to evict on-completion pod",
			pod: func() *apiv1.Pod {
				p := BuildTestPod("pod2", 100, 0)
				p.Annotations = map[string]string{
					drain.PodSafeToEvictKey: drain.PodSafeToEvictOnCompletionValue,
				}
				return p
			}(),
			want: drainability.NewWaitStatus(),
		},
		{
			name: "pod with true safe-to-evict annotation",
			pod: func() *apiv1.Pod {
				p := BuildTestPod("pod3", 100, 0)
				p.Annotations = map[string]string{
					drain.PodSafeToEvictKey: "true",
				}
				return p
			}(),
			want: drainability.NewUndefinedStatus(),
		},
		{
			name: "pod with false safe-to-evict annotation",
			pod: func() *apiv1.Pod {
				p := BuildTestPod("pod4", 100, 0)
				p.Annotations = map[string]string{
					drain.PodSafeToEvictKey: "false",
				}
				return p
			}(),
			want: drainability.NewUndefinedStatus(),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := New().Drainable(nil, tc.pod, nil)
			assert.Equal(t, tc.want, got)
		})
	}
}
