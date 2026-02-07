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

package annotations

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetOriginalResourcesAnnotationValue(t *testing.T) {
	testCases := []struct {
		name      string
		container *core.Container
		expected  *OriginalResources
		expectErr bool
	}{
		{
			name: "full resources",
			container: &core.Container{
				Resources: core.ResourceRequirements{
					Requests: core.ResourceList{
						core.ResourceCPU:    resource.MustParse("1"),
						core.ResourceMemory: resource.MustParse("1Gi"),
					},
					Limits: core.ResourceList{
						core.ResourceCPU:    resource.MustParse("2"),
						core.ResourceMemory: resource.MustParse("2Gi"),
					},
				},
			},
			expected: &OriginalResources{
				Requests: core.ResourceList{
					core.ResourceCPU:    resource.MustParse("1"),
					core.ResourceMemory: resource.MustParse("1Gi"),
				},
				Limits: core.ResourceList{
					core.ResourceCPU:    resource.MustParse("2"),
					core.ResourceMemory: resource.MustParse("2Gi"),
				},
			},
			expectErr: false,
		},
		{
			name: "only requests",
			container: &core.Container{
				Resources: core.ResourceRequirements{
					Requests: core.ResourceList{
						core.ResourceCPU:    resource.MustParse("1"),
						core.ResourceMemory: resource.MustParse("1Gi"),
					},
				},
			},
			expected: &OriginalResources{
				Requests: core.ResourceList{
					core.ResourceCPU:    resource.MustParse("1"),
					core.ResourceMemory: resource.MustParse("1Gi"),
				},
				Limits: core.ResourceList{},
			},
			expectErr: false,
		},
		{
			name: "no resources",
			container: &core.Container{
				Resources: core.ResourceRequirements{},
			},
			expected: &OriginalResources{
				Requests: core.ResourceList{},
				Limits:   core.ResourceList{},
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := GetOriginalResourcesAnnotationValue(tc.container)
			if tc.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			var got OriginalResources
			err = json.Unmarshal([]byte(val), &got)
			assert.NoError(t, err)
			assert.True(t, tc.expected.Requests.Cpu().Equal(*got.Requests.Cpu()), "CPU requests do not match")
			assert.True(t, tc.expected.Requests.Memory().Equal(*got.Requests.Memory()), "Memory requests do not match")
			assert.True(t, tc.expected.Limits.Cpu().Equal(*got.Limits.Cpu()), "CPU limits do not match")
			assert.True(t, tc.expected.Limits.Memory().Equal(*got.Limits.Memory()), "Memory limits do not match")
		})
	}
}

func TestGetOriginalResourcesFromAnnotation(t *testing.T) {
	testCases := []struct {
		name      string
		pod       *core.Pod
		expected  *OriginalResources
		expectErr bool
	}{
		{
			name: "valid annotation",
			pod: &core.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						StartupCPUBoostAnnotation: `{"requests":{"cpu":"1","memory":"1Gi"},"limits":{"cpu":"2","memory":"2Gi"}}`,
					},
				},
			},
			expected: &OriginalResources{
				Requests: core.ResourceList{
					core.ResourceCPU:    resource.MustParse("1"),
					core.ResourceMemory: resource.MustParse("1Gi"),
				},
				Limits: core.ResourceList{
					core.ResourceCPU:    resource.MustParse("2"),
					core.ResourceMemory: resource.MustParse("2Gi"),
				},
			},
			expectErr: false,
		},
		{
			name: "no annotation",
			pod: &core.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
			},
			expected:  nil,
			expectErr: false,
		},
		{
			name: "invalid json",
			pod: &core.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						StartupCPUBoostAnnotation: "invalid-json",
					},
				},
			},
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GetOriginalResourcesFromAnnotation(tc.pod)
			if tc.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tc.expected == nil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
				assert.True(t, tc.expected.Requests.Cpu().Equal(*got.Requests.Cpu()), "CPU requests do not match")
				assert.True(t, tc.expected.Requests.Memory().Equal(*got.Requests.Memory()), "Memory requests do not match")
				assert.True(t, tc.expected.Limits.Cpu().Equal(*got.Limits.Cpu()), "CPU limits do not match")
				assert.True(t, tc.expected.Limits.Memory().Equal(*got.Limits.Memory()), "Memory limits do not match")
			}
		})
	}
}
