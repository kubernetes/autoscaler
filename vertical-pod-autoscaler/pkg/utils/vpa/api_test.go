/*
Copyright 2018 The Kubernetes Authors.

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

package api

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

const (
	containerName = "container1"
)

func TestPodMatchesVPA(t *testing.T) {
	type testCase struct {
		pod    *apiv1.Pod
		vpa    *vpa_types.VerticalPodAutoscaler
		result bool
	}
	selector := "app = testingApp"

	pod := test.BuildTestPod("test-pod", containerName, "1", "100M", nil, nil)
	pod.Labels = map[string]string{"app": "testingApp"}

	vpa := test.BuildTestVerticalPodAutoscaler(containerName, "2", "1", "3", "200M", "100M", "1G", selector)

	otherNamespaceVPA := test.BuildTestVerticalPodAutoscaler(containerName, "2", "1", "3", "200M", "100M", "1G", selector)
	otherNamespaceVPA.Namespace = "other"

	otherSelectorVPA := test.BuildTestVerticalPodAutoscaler(containerName, "2", "1", "3", "200M", "100M", "1G", "app = other")

	testCases := []testCase{
		{pod, vpa, true},
		{pod, otherNamespaceVPA, false},
		{pod, otherSelectorVPA, false}}

	for _, tc := range testCases {
		actual := PodMatchesVPA(tc.pod, tc.vpa)
		assert.Equal(t, tc.result, actual)
	}
}

func TestGetControllingVPAForPod(t *testing.T) {
	selector := "app = testingApp"

	pod := test.BuildTestPod("test-pod", containerName, "1", "100M", nil, nil)
	pod.Labels = map[string]string{"app": "testingApp"}

	vpaA := test.BuildTestVerticalPodAutoscaler(containerName, "2", "1", "3", "200M", "100M", "1G", selector)
	vpaA.CreationTimestamp = metav1.NewTime(time.Unix(5, 0))
	vpaB := test.BuildTestVerticalPodAutoscaler(containerName, "20", "10", "30", "200M", "100M", "1G", selector)
	vpaB.CreationTimestamp = metav1.NewTime(time.Unix(10, 0))
	nonMatchingVPA := test.BuildTestVerticalPodAutoscaler(containerName, "2", "1", "3", "200M", "100M", "1G", "app = other")
	nonMatchingVPA.CreationTimestamp = metav1.NewTime(time.Unix(2, 0))

	chosen := GetControllingVPAForPod(pod, []*vpa_types.VerticalPodAutoscaler{vpaB, vpaA, nonMatchingVPA})
	assert.Equal(t, vpaA, chosen)
}
