/*
Copyright 2025 The Kubernetes Authors.

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

package provreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1"
	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/pods"
	testutils "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestProvisioningRequestScaleUpEnforcer(t *testing.T) {
	prPod1 := testutils.BuildTestPod("pr-pod-1", 500, 10)
	prPod1.Annotations[v1.ProvisioningRequestPodAnnotationKey] = "pr-class"

	prPod2 := testutils.BuildTestPod("pr-pod-2", 500, 10)
	prPod2.Annotations[pods.DeprecatedProvisioningRequestPodAnnotationKey] = "pr-class-2"

	pod1 := testutils.BuildTestPod("pod-1", 500, 10)
	pod2 := testutils.BuildTestPod("pod-2", 500, 10)

	testCases := map[string]struct {
		unschedulablePods []*apiv1.Pod
		want              bool
	}{
		"Any pod with ProvisioningRequest annotation key forces scale up": {
			unschedulablePods: []*corev1.Pod{prPod1, pod1},
			want:              true,
		},
		"Any pod with ProvisioningRequest deprecated annotation key forces scale up": {
			unschedulablePods: []*corev1.Pod{prPod2, pod1},
			want:              true,
		},
		"Pod without ProvisioningRequest annotation key don't force scale up": {
			unschedulablePods: []*corev1.Pod{pod1, pod2},
			want:              false,
		},
		"No pods don't force scale up": {
			unschedulablePods: []*corev1.Pod{},
			want:              false,
		},
	}
	for _, test := range testCases {
		scaleUpEnforcer := NewProvisioningRequestScaleUpEnforcer()
		got := scaleUpEnforcer.ShouldForceScaleUp(test.unschedulablePods)
		assert.Equal(t, got, test.want)
	}
}
