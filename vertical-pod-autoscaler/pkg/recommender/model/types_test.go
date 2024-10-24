/*
Copyright 2024 The Kubernetes Authors.

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

package model

import (
	"testing"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type ResourcesAsResourceListTestCase struct {
	name         string
	resources    Resources
	humanize     bool
	resourceList apiv1.ResourceList
}

func TestResourcesAsResourceList(t *testing.T) {
	testCases := []ResourcesAsResourceListTestCase{
		{
			name: "basic resources without humanize",
			resources: Resources{
				ResourceCPU:    1000,
				ResourceMemory: 1000,
			},
			humanize: false,
			resourceList: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewMilliQuantity(1000, resource.DecimalSI),
				apiv1.ResourceMemory: *resource.NewQuantity(1000, resource.BinarySI),
			},
		},
		{
			name: "basic resources with humanize",
			resources: Resources{
				ResourceCPU:    1000,
				ResourceMemory: 262144000, // 250Mi
			},
			humanize: true,
			resourceList: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewMilliQuantity(1000, resource.DecimalSI),
				apiv1.ResourceMemory: *resource.NewQuantity(250, resource.BinarySI),
			},
		},
		{
			name: "large memory value with humanize",
			resources: Resources{
				ResourceCPU:    1000,
				ResourceMemory: 839500000, // 800Mi
			},
			resourceList: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewMilliQuantity(1000, resource.DecimalSI),
				apiv1.ResourceMemory: *resource.NewQuantity(800, resource.BinarySI),
			},
		},
		{
			name: "zero values",
			resources: Resources{
				ResourceCPU:    0,
				ResourceMemory: 0,
			},
			resourceList: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewMilliQuantity(0, resource.DecimalSI),
				apiv1.ResourceMemory: *resource.NewQuantity(0, resource.BinarySI),
			},
		},
		{
			name: "large memory value without humanize",
			resources: Resources{
				ResourceCPU:    1000,
				ResourceMemory: 839500000,
			},
			resourceList: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewMilliQuantity(1000, resource.DecimalSI),
				apiv1.ResourceMemory: *resource.NewQuantity(839500000, resource.BinarySI),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ResourcesAsResourceList(tc.resources, tc.humanize)
			if len(result) != len(tc.resourceList) {
				t.Errorf("expected %v, got %v", tc.resourceList, result)
			}
			if result[apiv1.ResourceCPU] != tc.resourceList[apiv1.ResourceCPU] {
				t.Errorf("expected %v, got %v", tc.resourceList[apiv1.ResourceCPU], result[apiv1.ResourceCPU])
			}
		})
	}
}

type HumanizeMemoryQuantityTestCase struct {
	name   string
	value  int64
	wanted string
}

func TestHumanizeMemoryQuantity(t *testing.T) {
	testCases := []HumanizeMemoryQuantityTestCase{
		{
			name:   "1Ki",
			value:  1024,
			wanted: "1Ki",
		},
		{
			name:   "1Mi",
			value:  1024 * 1024,
			wanted: "1Mi",
		},
		{
			name:   "1Gi",
			value:  1024 * 1024 * 1024,
			wanted: "1Gi",
		},
		{
			name:   "1Ti",
			value:  1024 * 1024 * 1024 * 1024,
			wanted: "1Ti",
		},
		{
			name:   "256Mi",
			value:  256 * 1024 * 1024,
			wanted: "256Mi",
		},
		{
			name:   "2Gi",
			value:  1.5 * 1024 * 1024 * 1024,
			wanted: "2Gi",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			result := HumanizeMemoryQuantity(tc.value)
			if result != tc.wanted {
				t.Errorf("expected %v, got %v", tc.wanted, result)
			}
		})
	}
}
