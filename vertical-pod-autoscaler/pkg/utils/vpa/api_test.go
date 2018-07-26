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
	"k8s.io/apimachinery/pkg/api/resource"
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

	pod := test.Pod().WithName("test-pod").AddContainer(test.BuildTestContainer(containerName, "1", "100M")).Get()
	pod.Labels = map[string]string{"app": "testingApp"}

	vpaBuilder := test.VerticalPodAutoscaler().
		WithContainer(containerName).
		WithTarget("2", "200M").
		WithMinAllowed("1", "100M").
		WithMaxAllowed("3", "1G").
		WithSelector(selector)

	vpa := vpaBuilder.Get()
	otherNamespaceVPA := vpaBuilder.WithNamespace("other").Get()
	otherSelectorVPA := vpaBuilder.WithSelector("app = other").Get()

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

	pod := test.Pod().WithName("test-pod").AddContainer(test.BuildTestContainer(containerName, "1", "100M")).Get()
	pod.Labels = map[string]string{"app": "testingApp"}

	vpaBuilder := test.VerticalPodAutoscaler().
		WithContainer(containerName).
		WithTarget("2", "200M").
		WithMinAllowed("1", "100M").
		WithMaxAllowed("3", "1G").
		WithSelector(selector)
	vpaA := vpaBuilder.WithCreationTimestamp(time.Unix(5, 0)).Get()
	vpaB := vpaBuilder.WithCreationTimestamp(time.Unix(10, 0)).Get()
	nonMatchingVPA := vpaBuilder.WithCreationTimestamp(time.Unix(2, 0)).WithSelector("app = other").Get()

	chosen := GetControllingVPAForPod(pod, []*vpa_types.VerticalPodAutoscaler{vpaB, vpaA, nonMatchingVPA})
	assert.Equal(t, vpaA, chosen)
}

func TestGetContainerResourcePolicy(t *testing.T) {
	containerPolicy1 := vpa_types.ContainerResourcePolicy{
		ContainerName: "container1",
		MinAllowed: apiv1.ResourceList{
			apiv1.ResourceCPU: *resource.NewScaledQuantity(10, 1),
		},
	}
	containerPolicy2 := vpa_types.ContainerResourcePolicy{
		ContainerName: "container2",
		MaxAllowed: apiv1.ResourceList{
			apiv1.ResourceMemory: *resource.NewScaledQuantity(100, 1),
		},
	}
	policy := vpa_types.PodResourcePolicy{
		ContainerPolicies: []vpa_types.ContainerResourcePolicy{
			containerPolicy1, containerPolicy2,
		},
	}
	assert.Equal(t, &containerPolicy1, GetContainerResourcePolicy("container1", &policy))
	assert.Equal(t, &containerPolicy2, GetContainerResourcePolicy("container2", &policy))
	assert.Nil(t, GetContainerResourcePolicy("container3", &policy))

	// Add the wildcard ("*") policy.
	defaultPolicy := vpa_types.ContainerResourcePolicy{
		ContainerName: "*",
		MinAllowed: apiv1.ResourceList{
			apiv1.ResourceCPU: *resource.NewScaledQuantity(20, 1),
		},
	}
	policy = vpa_types.PodResourcePolicy{
		ContainerPolicies: []vpa_types.ContainerResourcePolicy{
			containerPolicy1, defaultPolicy, containerPolicy2,
		},
	}
	assert.Equal(t, &containerPolicy1, GetContainerResourcePolicy("container1", &policy))
	assert.Equal(t, &containerPolicy2, GetContainerResourcePolicy("container2", &policy))
	assert.Equal(t, &defaultPolicy, GetContainerResourcePolicy("container3", &policy))
}
