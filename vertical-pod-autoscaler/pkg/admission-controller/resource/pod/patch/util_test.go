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

package patch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestGetAddResourceRequirementValuePatch(t *testing.T) {
	tests := []struct {
		name          string
		index         int
		kind          string
		resource      corev1.ResourceName
		quantity      resource.Quantity
		expectedPath  string
		expectedValue string
	}{
		{
			name:          "Standard resource cpu",
			index:         0,
			kind:          "requests",
			resource:      corev1.ResourceCPU,
			quantity:      resource.MustParse("500m"),
			expectedPath:  "/spec/containers/0/resources/requests/cpu",
			expectedValue: "500m",
		},
		{
			name:          "Standard resource memory",
			index:         1,
			kind:          "limits",
			resource:      corev1.ResourceMemory,
			quantity:      resource.MustParse("1Gi"),
			expectedPath:  "/spec/containers/1/resources/limits/memory",
			expectedValue: "1Gi",
		},
		{
			name:          "Extended resource with slash",
			index:         2,
			kind:          "limits",
			resource:      corev1.ResourceName("nvidia.com/gpu"),
			quantity:      resource.MustParse("1"),
			expectedPath:  "/spec/containers/2/resources/limits/nvidia.com~1gpu",
			expectedValue: "1",
		},
		{
			name:          "Extended resource with tilde and slash",
			index:         0,
			kind:          "requests",
			resource:      corev1.ResourceName("custom~domain.com/resource/name"),
			quantity:      resource.MustParse("10"),
			expectedPath:  "/spec/containers/0/resources/requests/custom~0domain.com~1resource~1name",
			expectedValue: "10",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			patch := GetAddResourceRequirementValuePatch(tc.index, tc.kind, tc.resource, tc.quantity)
			assert.Equal(t, "add", patch.Op)
			assert.Equal(t, tc.expectedPath, patch.Path)
			assert.Equal(t, tc.expectedValue, patch.Value)
		})
	}
}

func TestGetAddAnnotationPatch(t *testing.T) {
	tests := []struct {
		name           string
		annotationName string
		value          string
		expectedPath   string
	}{
		{
			name:           "Simple annotation",
			annotationName: "vpaObservedContainers",
			value:          "container1",
			expectedPath:   "/metadata/annotations/vpaObservedContainers",
		},
		{
			name:           "Annotation with slash",
			annotationName: "example.com/annotation",
			value:          "value1",
			expectedPath:   "/metadata/annotations/example.com~1annotation",
		},
		{
			name:           "Annotation with tilde and slash",
			annotationName: "custom~domain.com/annotation/name",
			value:          "value2",
			expectedPath:   "/metadata/annotations/custom~0domain.com~1annotation~1name",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			patch := GetAddAnnotationPatch(tc.annotationName, tc.value)
			assert.Equal(t, "add", patch.Op)
			assert.Equal(t, tc.expectedPath, patch.Path)
			assert.Equal(t, tc.value, patch.Value)
		})
	}
}

func TestGetRemoveAnnotationPatch(t *testing.T) {
	tests := []struct {
		name           string
		annotationName string
		expectedPath   string
	}{
		{
			name:           "Simple annotation",
			annotationName: "vpaObservedContainers",
			expectedPath:   "/metadata/annotations/vpaObservedContainers",
		},
		{
			name:           "Annotation with slash",
			annotationName: "example.com/annotation",
			expectedPath:   "/metadata/annotations/example.com~1annotation",
		},
		{
			name:           "Annotation with tilde and slash",
			annotationName: "custom~domain.com/annotation/name",
			expectedPath:   "/metadata/annotations/custom~0domain.com~1annotation~1name",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			patch := GetRemoveAnnotationPatch(tc.annotationName)
			assert.Equal(t, "remove", patch.Op)
			assert.Equal(t, tc.expectedPath, patch.Path)
			assert.Nil(t, patch.Value)
		})
	}
}
