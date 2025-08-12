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

package fake

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestIsFake(t *testing.T) {
	testCases := []struct {
		name string
		pod  *apiv1.Pod
		want bool
	}{
		{
			name: "real pod",
			pod:  test.BuildTestPod("real", 10, 10),
			want: false,
		},
		{
			name: "fake pod",
			pod:  WithFakePodAnnotation(test.BuildTestPod("fake", 10, 10)),
			want: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsFake(tc.pod)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestWithFakePodAnnotation(t *testing.T) {
	pod := test.BuildTestPod("pod", 10, 10)
	assert.Equal(t, map[string]string{}, pod.Annotations)
	pod = WithFakePodAnnotation(pod)
	assert.NotNil(t, pod.Annotations)
	assert.Equal(t, FakePodAnnotationValue, pod.Annotations[FakePodAnnotationKey])
}
