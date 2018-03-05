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

package logic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

func TestUpdateResourceRequests(t *testing.T) {
	type testCase struct {
		pod            *apiv1.Pod
		vpas           []*vpa_types.VerticalPodAutoscaler
		expectedAction bool
		expectedMem    string
		expectedCPU    string
	}
	containerName := "container1"
	labels := map[string]string{"app": "testingApp"}
	vpa := test.BuildTestVerticalPodAutoscaler(containerName, "2", "1", "3", "200M", "100M", "1G", "app = testingApp")

	uninitialized := test.BuildTestPod("test_uninitialized", containerName, "", "", nil, nil)
	uninitialized.ObjectMeta.Labels = labels

	initialized := test.BuildTestPod("test_initialized", containerName, "1", "100M", nil, nil)
	initialized.ObjectMeta.Labels = labels

	mismatchedVPA := test.BuildTestVerticalPodAutoscaler(containerName, "2", "1", "3", "200M", "100M", "1G", "app = differentApp")
	offVPA := test.BuildTestVerticalPodAutoscaler(containerName, "2.5", "1", "3", "250M", "100M", "1G", "app = testingApp")
	offVPA.Spec.UpdatePolicy.UpdateMode = vpa_types.UpdateModeOff

	targetBelowMinVPA := test.BuildTestVerticalPodAutoscaler(containerName, "3", "4", "5", "150M", "300M", "1G", "app = testingApp")

	targetAboveMaxVPA := test.BuildTestVerticalPodAutoscaler(containerName, "7", "4", "5", "2G", "300M", "1G", "app = testingApp")

	testCases := []testCase{{
		pod:            uninitialized,
		vpas:           []*vpa_types.VerticalPodAutoscaler{vpa},
		expectedAction: true,
		expectedMem:    "200M",
		expectedCPU:    "2",
	}, {
		pod:            uninitialized,
		vpas:           []*vpa_types.VerticalPodAutoscaler{targetBelowMinVPA},
		expectedAction: true,
		expectedMem:    "300M", // MinMemory is expected to be used
		expectedCPU:    "4",    // MinCpu is expected to be used
	}, {
		pod:            uninitialized,
		vpas:           []*vpa_types.VerticalPodAutoscaler{targetAboveMaxVPA},
		expectedAction: true,
		expectedMem:    "1G", // MaxMemory is expected to be used
		expectedCPU:    "5",  // MaxCpu is expected to be used
	}, {
		pod:            initialized,
		vpas:           []*vpa_types.VerticalPodAutoscaler{vpa},
		expectedAction: true,
		expectedMem:    "200M",
		expectedCPU:    "2",
	}, {
		pod:            uninitialized,
		vpas:           []*vpa_types.VerticalPodAutoscaler{mismatchedVPA},
		expectedAction: false,
	}, {
		pod:            uninitialized,
		vpas:           []*vpa_types.VerticalPodAutoscaler{offVPA},
		expectedAction: false,
	}, {
		pod:            uninitialized,
		vpas:           []*vpa_types.VerticalPodAutoscaler{offVPA, vpa},
		expectedAction: true,
		expectedMem:    "200M",
		expectedCPU:    "2",
	}}
	for _, tc := range testCases {
		vpaNamespaceLister := &test.VerticalPodAutoscalerListerMock{}
		vpaNamespaceLister.On("List").Return(tc.vpas, nil)

		vpaLister := &test.VerticalPodAutoscalerListerMock{}
		vpaLister.On("VerticalPodAutoscalers", "default").Return(vpaNamespaceLister)

		recommendationProvider := &recommendationProvider{
			vpaLister: vpaLister,
		}

		requests, err := recommendationProvider.GetRequestForPod(tc.pod)

		if tc.expectedAction {
			assert.Nil(t, err)
			assert.Equal(t, len(requests), 1)
			cpu, err := resource.ParseQuantity(tc.expectedCPU)
			assert.NoError(t, err)
			assert.Equal(t, cpu, requests[0][apiv1.ResourceCPU])
			memory, err := resource.ParseQuantity(tc.expectedMem)
			assert.NoError(t, err)
			assert.Equal(t, memory, requests[0][apiv1.ResourceMemory])

		} else {
			assert.Equal(t, len(requests), 0)
		}
	}
}
