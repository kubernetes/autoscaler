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

package translator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	"k8s.io/utils/ptr"
)

func TestGetBufferNumberOfPods(t *testing.T) {
	podTemplate := corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU: resource.MustParse("1"),
						},
					},
				},
			},
		},
	}

	testCases := []struct {
		name             string
		replicas         *int32
		percentage       *int32
		limits           *v1.ResourceList
		scalableReplicas *int32
		expectedReplicas int32
		expectError      bool
	}{
		{
			name:             "Only Replicas",
			replicas:         ptr.To[int32](5),
			expectedReplicas: 5,
		},
		{
			name:             "Only Percentage",
			percentage:       ptr.To[int32](200), // 200% * 10 = 20
			scalableReplicas: ptr.To[int32](10),
			expectedReplicas: 20,
		},
		{
			name:             "Only Percentage Rounding Up",
			percentage:       ptr.To[int32](33), // 33% of 10 = 3.3 -> ceil(3.3) = 4
			scalableReplicas: ptr.To[int32](10),
			expectedReplicas: 4,
		},
		{
			name:             "Replicas and Percentage (Replicas larger)",
			replicas:         ptr.To[int32](25),
			percentage:       ptr.To[int32](200), // 200% * 10 = 20
			scalableReplicas: ptr.To[int32](10),
			expectedReplicas: 25,
		},
		{
			name:             "Replicas and Percentage (Percentage larger)",
			replicas:         ptr.To[int32](15),
			percentage:       ptr.To[int32](200), // 200% * 10 = 20
			scalableReplicas: ptr.To[int32](10),
			expectedReplicas: 20,
		},
		{
			name:             "Replicas and Limits (Limits smaller)",
			replicas:         ptr.To[int32](25),
			limits:           &v1.ResourceList{v1.ResourceName(corev1.ResourceCPU): resource.MustParse("20")},
			expectedReplicas: 20,
		},
		{
			name:             "Replicas and Limits (Replicas smaller)",
			replicas:         ptr.To[int32](15),
			limits:           &v1.ResourceList{v1.ResourceName(corev1.ResourceCPU): resource.MustParse("20")},
			expectedReplicas: 15,
		},
		{
			name:             "Percentage and Limits (Limits smaller)",
			percentage:       ptr.To[int32](250), // 250% * 10 = 25
			scalableReplicas: ptr.To[int32](10),
			limits:           &v1.ResourceList{v1.ResourceName(corev1.ResourceCPU): resource.MustParse("20")},
			expectedReplicas: 20,
		},
		{
			name:             "Only Limits",
			limits:           &v1.ResourceList{v1.ResourceName(corev1.ResourceCPU): resource.MustParse("20")},
			expectedReplicas: 20,
		},
		{
			name:        "Nothing specified",
			expectError: true,
		},
		{
			name:        "Percentage present but scalableReplicas nil",
			percentage:  ptr.To[int32](200),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buffer := &v1.CapacityBuffer{
				Spec: v1.CapacityBufferSpec{
					Replicas:   tc.replicas,
					Percentage: tc.percentage,
					Limits:     tc.limits,
				},
			}
			actual, err := getBufferNumberOfPods(buffer, podTemplate, tc.scalableReplicas)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedReplicas, actual)
			}
		})
	}
}

func TestReplicasFromPercentage(t *testing.T) {
	testCases := []struct {
		name             string
		percentage       int32
		scalableReplicas int32
		expected         int32
	}{
		{
			name:             "200% of 10",
			percentage:       200,
			scalableReplicas: 10,
			expected:         20,
		},
		{
			name:             "50% of 10",
			percentage:       50,
			scalableReplicas: 10,
			expected:         5,
		},
		{
			name:             "0% of 10",
			percentage:       0,
			scalableReplicas: 10,
			expected:         0,
		},
		{
			name:             "50% of 0",
			percentage:       50,
			scalableReplicas: 0,
			expected:         0,
		},
		{
			name:             "Rounding up (3.3 -> 4)",
			percentage:       33,
			scalableReplicas: 10,
			expected:         4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := replicasFromPercentage(tc.percentage, tc.scalableReplicas)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestCalculateMaxPodsFromLimits(t *testing.T) {
	podRequests := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse("100m"),
		corev1.ResourceMemory: resource.MustParse("200Mi"),
	}

	limits := v1.ResourceList{
		v1.ResourceName(corev1.ResourceCPU):    resource.MustParse("2"),
		v1.ResourceName(corev1.ResourceMemory): resource.MustParse("2000Mi"),
	}

	maxPods, err := calculateMaxPodsFromLimits(podRequests, limits)
	assert.NoError(t, err)
	// CPU allows 2 / 0.1 = 20
	// Memory allows 2000 / 200 = 10
	// Should be the min of these -> 10
	assert.Equal(t, int32(10), maxPods)
}
