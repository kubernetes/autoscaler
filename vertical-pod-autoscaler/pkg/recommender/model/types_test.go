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

	"github.com/stretchr/testify/assert"
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
			assert.Equal(t, tc.wanted, result)
		})
	}
}

type CPUAmountFromCoresTestCase struct {
	name  string
	cores float64
	want  ResourceAmount
}

func TestCPUAmountFromCores(t *testing.T) {
	testCases := []CPUAmountFromCoresTestCase{
		{
			name:  "should get 69",
			cores: 0.069946,
			want:  69,
		},
		{
			name:  "should get 615",
			cores: 0.61535112,
			want:  615,
		},
		{
			name:  "should get 17",
			cores: 0.0172071,
			want:  17,
		},
		{
			name:  "should get 4",
			cores: 0.00455,
			want:  4,
		},
		{
			name:  "should get 12",
			cores: 0.0123456789,
			want:  12,
		},
		{
			name:  "should get 1",
			cores: 0.00123456789,
			want:  1,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := CPUAmountFromCores(tc.cores)
			assert.Equal(t, tc.want, result)
		})
	}
}

type CoresFromCPUAmountTestCase struct {
	name      string
	cpuAmount ResourceAmount
	want      float64
}

func TestCoresFromCPUAmount(t *testing.T) {
	tc := []CoresFromCPUAmountTestCase{
		{
			name:      "should get 0.069",
			cpuAmount: 69,
			want:      0.069,
		},
		{
			name:      "should get 0.615",
			cpuAmount: 615,
			want:      0.615,
		},
		{
			name:      "should get 0.017",
			cpuAmount: 17,
			want:      0.017,
		},
		{
			name:      "should get 0.004",
			cpuAmount: 4,
			want:      0.004,
		},
		{
			name:      "should get 0.012",
			cpuAmount: 12,
			want:      0.012,
		},
		{
			name:      "should get 0.001",
			cpuAmount: 1,
			want:      0.001,
		},
	}
	for _, tc := range tc {
		t.Run(tc.name, func(t *testing.T) {
			result := CoresFromCPUAmount(tc.cpuAmount)
			assert.Equal(t, tc.want, result)
		})
	}
}

type QuantityFromCPUAmountTestCase struct {
	name      string
	cpuAmount ResourceAmount
	want      resource.Quantity
}

func TestQuantityFromCPUAmount(t *testing.T) {
	tc := []QuantityFromCPUAmountTestCase{
		{
			name:      "should get 69",
			cpuAmount: 69,
			want:      *resource.NewScaledQuantity(69, -3),
		},
		{
			name:      "should get 615",
			cpuAmount: 615,
			want:      *resource.NewScaledQuantity(615, -3),
		},
		{
			name:      "should get 17",
			cpuAmount: 17,
			want:      *resource.NewScaledQuantity(17, -3),
		},
		{
			name:      "should get 4",
			cpuAmount: 4,
			want:      *resource.NewScaledQuantity(4, -3),
		},
		{
			name:      "should get 12",
			cpuAmount: 12,
			want:      *resource.NewScaledQuantity(12, -3),
		},
		{
			name:      "should get 1",
			cpuAmount: 1,
			want:      *resource.NewScaledQuantity(1, -3),
		},
	}
	for _, tc := range tc {
		t.Run(tc.name, func(t *testing.T) {
			result := QuantityFromCPUAmount(tc.cpuAmount)
			assert.Equal(t, tc.want, result)
		})
	}
}

type MemoryAmountFromBytesTestCase struct {
	name  string
	bytes float64
	want  ResourceAmount
}

func TestMemoryAmountFromBytes(t *testing.T) {
	tc := []MemoryAmountFromBytesTestCase{
		{
			name:  "should get 69",
			bytes: 69.669,
			want:  69,
		},
		{
			name:  "should get 12",
			bytes: 12.333,
			want:  12,
		},
		{
			name:  "should get 17",
			bytes: 17.357,
			want:  17,
		},
		{
			name:  "should get 4",
			bytes: 4.000,
			want:  4,
		},
		{
			name:  "should get 12",
			bytes: 12.123456,
			want:  12,
		},
		{
			name:  "should get 1",
			bytes: 1,
			want:  1,
		},
	}
	for _, tc := range tc {
		t.Run(tc.name, func(t *testing.T) {
			result := MemoryAmountFromBytes(tc.bytes)
			assert.Equal(t, tc.want, result)
		})
	}
}

type BytesFromMemoryAmountTestCase struct {
	name         string
	memoryAmount ResourceAmount
	want         float64
}

func TestBytesFromMemoryAmount(t *testing.T) {
	tc := []BytesFromMemoryAmountTestCase{
		{
			name:         "should get 69",
			memoryAmount: 69,
			want:         69,
		},
		{
			name:         "should get 12",
			memoryAmount: 12,
			want:         12,
		},
		{
			name:         "should get 17",
			memoryAmount: 17,
			want:         17,
		},
		{
			name:         "should get 4",
			memoryAmount: 4,
			want:         4,
		},
		{
			name:         "should get 12",
			memoryAmount: 12,
			want:         12,
		},
		{
			name:         "should get 1",
			memoryAmount: 1,
			want:         1,
		},
	}
	for _, tc := range tc {
		t.Run(tc.name, func(t *testing.T) {
			result := BytesFromMemoryAmount(tc.memoryAmount)
			assert.Equal(t, tc.want, result)
		})
	}
}

type QuantityFromMemoryAmountTestCase struct {
	name         string
	memoryAmount ResourceAmount
	want         resource.Quantity
}

func TestQuantityFromMemoryAmount(t *testing.T) {
	tc := []QuantityFromMemoryAmountTestCase{
		{
			name:         "should get 69",
			memoryAmount: 69,
			want:         *resource.NewQuantity(69, resource.DecimalSI),
		},
		{
			name:         "should get 12",
			memoryAmount: 12,
			want:         *resource.NewQuantity(12, resource.DecimalSI),
		},
		{
			name:         "should get 17",
			memoryAmount: 17,
			want:         *resource.NewQuantity(17, resource.DecimalSI),
		},
		{
			name:         "should get 4",
			memoryAmount: 4,
			want:         *resource.NewQuantity(4, resource.DecimalSI),
		},
		{
			name:         "should get 12",
			memoryAmount: 12,
			want:         *resource.NewQuantity(12, resource.DecimalSI),
		},
		{
			name:         "should get 1",
			memoryAmount: 1,
			want:         *resource.NewQuantity(1, resource.DecimalSI),
		},
		{
			name:         "should get 0",
			memoryAmount: 0,
			want:         *resource.NewQuantity(0, resource.DecimalSI),
		},
		{
			name:         "should get 123456789",
			memoryAmount: 123456789,
			want:         *resource.NewQuantity(123456789, resource.DecimalSI),
		},
	}
	for _, tc := range tc {
		t.Run(tc.name, func(t *testing.T) {
			result := QuantityFromMemoryAmount(tc.memoryAmount)
			assert.Equal(t, tc.want, result)
		})
	}
}

type ScaleResourceTestCase struct {
	name   string
	amount ResourceAmount
	factor float64
	want   ResourceAmount
}

func TestScaleResource(t *testing.T) {
	tc := []ScaleResourceTestCase{
		{
			name:   "should get 69",
			amount: 69,
			factor: 1.0,
			want:   69,
		},
		{
			name:   "should get 615",
			amount: 615,
			factor: 1.0,
			want:   615,
		},
		{
			name:   "should get 17",
			amount: 17,
			factor: 1.0,
			want:   17,
		},
		{
			name:   "should get 8",
			amount: 4,
			factor: 2.124,
			want:   8,
		},
		{
			name:   "should get 13",
			amount: 12,
			factor: 1.1,
			want:   13,
		},
		{
			name:   "should get 1505",
			amount: 1213,
			factor: 1.2414,
			want:   1505,
		},
		{
			name:   "should get 0",
			amount: 0,
			factor: 1.0,
			want:   0,
		},
		{
			name:   "should get 0",
			amount: 11,
			factor: 0.0,
			want:   0,
		},
		{
			name:   "should get 0",
			amount: 41208,
			factor: 0.000001,
			want:   0,
		},
	}
	for _, tc := range tc {
		t.Run(tc.name, func(t *testing.T) {
			result := ScaleResource(tc.amount, tc.factor)
			assert.Equal(t, tc.want, result)
		})
	}
}

type ResourceNamesApiToModelTestCase struct {
	name           string
	apiResources   []apiv1.ResourceName
	modelResources []ResourceName
}

func TestResourceNamesApiToModel(t *testing.T) {
	tc := []ResourceNamesApiToModelTestCase{
		{
			name: "should get cpu and memory",
			apiResources: []apiv1.ResourceName{
				apiv1.ResourceCPU,
				apiv1.ResourceMemory,
			},
			modelResources: []ResourceName{
				ResourceCPU,
				ResourceMemory,
			},
		},
		{
			name: "should get cpu",
			apiResources: []apiv1.ResourceName{
				apiv1.ResourceCPU,
			},
			modelResources: []ResourceName{
				ResourceCPU,
			},
		},
		{
			name: "should get memory",
			apiResources: []apiv1.ResourceName{
				apiv1.ResourceMemory,
			},
			modelResources: []ResourceName{
				ResourceMemory,
			},
		},
		{
			name:           "should get empty",
			apiResources:   []apiv1.ResourceName{},
			modelResources: []ResourceName{},
		},
	}
	for _, tc := range tc {
		t.Run(tc.name, func(t *testing.T) {
			result := ResourceNamesApiToModel(tc.apiResources)
			assert.Equal(t, tc.modelResources, *result) // Dereference the result here
		})
	}
}

type RoundResourceAmountsTestCase struct {
	name      string
	resources Resources
	want      Resources
}

func TestRoundResourceAmounts(t *testing.T) {
	tc := []RoundResourceAmountsTestCase{
		{
			name: "should round down CPU and memory",
			resources: Resources{
				ResourceCPU:    ResourceAmount(4500), // 4.5 CPU cores
				ResourceMemory: ResourceAmount(7168), // 7168MB memory
			},
			want: Resources{
				ResourceCPU:    ResourceAmount(4000), // Round down to 4.0 CPU cores
				ResourceMemory: ResourceAmount(7000), // Round down to 7000MB
			},
		},
		{
			name: "exact multiples should remain unchanged",
			resources: Resources{
				ResourceCPU:    ResourceAmount(2000), // 2.0 CPU cores
				ResourceMemory: ResourceAmount(4000), // 4000MB memory
			},
			want: Resources{
				ResourceCPU:    ResourceAmount(2000), // Should remain 2.0
				ResourceMemory: ResourceAmount(4000), // Should remain 4000
			},
		},
		{
			name: "small values should round to zero",
			resources: Resources{
				ResourceCPU:    ResourceAmount(499), // 0.499 CPU cores
				ResourceMemory: ResourceAmount(999), // 999MB memory
			},
			want: Resources{
				ResourceCPU:    ResourceAmount(0), // Round down to 0
				ResourceMemory: ResourceAmount(0), // Round down to 0
			},
		},
		{
			name:      "empty resources should remain empty",
			resources: Resources{},
			want:      Resources{},
		},
	}

	for _, tc := range tc {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.resources
			for resource, amount := range result {
				// Assuming unit of 1000 for both CPU and Memory
				result[resource] = RoundResourceAmount(amount, 1000)
			}
			assert.Equal(t, tc.want, result)
		})
	}
}

type ResourceAmountFromFloatTestCase struct {
	name   string
	amount float64
	want   ResourceAmount
}

func TestResourceAmountFromFloat(t *testing.T) {
	tc := []ResourceAmountFromFloatTestCase{
		{
			name:   "regular positive number",
			amount: 100.5,
			want:   ResourceAmount(100),
		},
		{
			name:   "zero",
			amount: 0.0,
			want:   ResourceAmount(0),
		},
		{
			name:   "negative number should return zero",
			amount: -50.5,
			want:   ResourceAmount(0),
		},
		{
			name:   "number larger than MaxResourceAmount should return MaxResourceAmount",
			amount: float64(MaxResourceAmount) + 100.0,
			want:   MaxResourceAmount,
		},
		{
			name:   "number equal to MaxResourceAmount should return MaxResourceAmount",
			amount: float64(MaxResourceAmount),
			want:   MaxResourceAmount,
		},
		{
			name:   "small decimal number",
			amount: 0.125,
			want:   ResourceAmount(0),
		},
		{
			name:   "very small negative number should return zero",
			amount: -0.0001,
			want:   ResourceAmount(0),
		},
	}

	for _, tc := range tc {
		t.Run(tc.name, func(t *testing.T) {
			result := resourceAmountFromFloat(tc.amount)
			assert.Equal(t, tc.want, result)
		})
	}
}
