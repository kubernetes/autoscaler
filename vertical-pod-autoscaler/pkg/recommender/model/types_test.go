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
				apiv1.ResourceMemory: *resource.NewQuantity(1000, resource.DecimalSI),
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
				apiv1.ResourceMemory: resource.MustParse("250.00Mi"),
			},
		},
		{
			name: "large memory value with humanize",
			resources: Resources{
				ResourceCPU:    1000,
				ResourceMemory: 839500000, // 800.61Mi
			},
			humanize: true,
			resourceList: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewMilliQuantity(1000, resource.DecimalSI),
				apiv1.ResourceMemory: resource.MustParse("800.61Mi"),
			},
		},
		{
			name: "zero values without humanize",
			resources: Resources{
				ResourceCPU:    0,
				ResourceMemory: 0,
			},
			humanize: false,
			resourceList: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewMilliQuantity(0, resource.DecimalSI),
				apiv1.ResourceMemory: *resource.NewQuantity(0, resource.DecimalSI),
			},
		},
		{
			name: "large memory value without humanize",
			resources: Resources{
				ResourceCPU:    1000,
				ResourceMemory: 839500000,
			},
			humanize: false,
			resourceList: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewMilliQuantity(1000, resource.DecimalSI),
				apiv1.ResourceMemory: *resource.NewQuantity(839500000, resource.DecimalSI),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ResourcesAsResourceList(tc.resources, tc.humanize)
			if !result[apiv1.ResourceCPU].Equal(tc.resourceList[apiv1.ResourceCPU]) {
				t.Errorf("expected %v, got %v", tc.resourceList[apiv1.ResourceCPU], result[apiv1.ResourceCPU])
			}
			if !result[apiv1.ResourceMemory].Equal(tc.resourceList[apiv1.ResourceMemory]) {
				t.Errorf("expected %v, got %v", tc.resourceList[apiv1.ResourceMemory], result[apiv1.ResourceMemory])
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
			name:   "1.00Ki",
			value:  1024,
			wanted: "1.00Ki",
		},
		{
			name:   "1.00Mi",
			value:  1024 * 1024,
			wanted: "1.00Mi",
		},
		{
			name:   "1.00Gi",
			value:  1024 * 1024 * 1024,
			wanted: "1.00Gi",
		},
		{
			name:   "1.00Ti",
			value:  1024 * 1024 * 1024 * 1024,
			wanted: "1.00Ti",
		},
		{
			name:   "256.00Mi",
			value:  256 * 1024 * 1024,
			wanted: "256.00Mi",
		},
		{
			name:   "1.50Gi",
			value:  1.5 * 1024 * 1024 * 1024,
			wanted: "1.50Gi",
		},
		{
			name:   "1Mi in bytes",
			value:  1050000,
			wanted: "1.00Mi",
		},
		{
			name:   "1.5Ki in bytes",
			value:  1537,
			wanted: "1.50Ki",
		},
		{
			name:   "4.65Gi",
			value:  4992073454,
			wanted: "4.65Gi",
		},
		{
			name:   "6.05Gi",
			value:  6499152537,
			wanted: "6.05Gi",
		},
		{
			name:   "15.23Gi",
			value:  16357476492,
			wanted: "15.23Gi",
		},
		{
			name:   "3.75Gi",
			value:  4022251530,
			wanted: "3.75Gi",
		},
		{
			name:   "12.65Gi",
			value:  13580968030,
			wanted: "12.65Gi",
		},
		{
			name:   "14.46Gi",
			value:  15530468536,
			wanted: "14.46Gi",
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
