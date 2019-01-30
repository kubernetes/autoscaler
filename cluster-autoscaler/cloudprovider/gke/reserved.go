/*
Copyright 2016 The Kubernetes Authors.

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

package gke

// There should be no imports as it is used standalone in e2e tests

const (
	// MiB - MebiByte size (2^20)
	MiB = 1024 * 1024
	// Duplicating an upstream bug treating GB as 1000*MiB (we need to predict the end result accurately).
	mbPerGB           = 1000
	millicoresPerCore = 1000
)

// PredictKubeReservedMemory calculates kube-reserved memory based on physical memory
func PredictKubeReservedMemory(physicalMemory int64) int64 {
	return memoryReservedMiB(physicalMemory/MiB) * MiB
}

// PredictKubeReservedCpuMillicores calculates kube-reserved cpu based on physical cpu
func PredictKubeReservedCpuMillicores(physicalCpuMillicores int64) int64 {
	return cpuReservedMillicores(physicalCpuMillicores)
}

type allocatableBracket struct {
	threshold            int64
	marginalReservedRate float64
}

func memoryReservedMiB(memoryCapacityMiB int64) int64 {
	if memoryCapacityMiB <= mbPerGB {
		if memoryCapacityMiB <= 0 {
			return 0
		}
		// The minimum reservation required for proper node operation is 255 MiB.
		// For any node with less than 1 GB of memory use the minimum. Nodes with
		// more memory will use the existing reservation thresholds.
		return 255
	}
	return calculateReserved(memoryCapacityMiB, []allocatableBracket{
		{
			threshold:            0,
			marginalReservedRate: 0.25,
		},
		{
			threshold:            4 * mbPerGB,
			marginalReservedRate: 0.2,
		},
		{
			threshold:            8 * mbPerGB,
			marginalReservedRate: 0.1,
		},
		{
			threshold:            16 * mbPerGB,
			marginalReservedRate: 0.06,
		},
		{
			threshold:            128 * mbPerGB,
			marginalReservedRate: 0.02,
		},
	})
}

func cpuReservedMillicores(cpuCapacityMillicores int64) int64 {
	return calculateReserved(cpuCapacityMillicores, []allocatableBracket{
		{
			threshold:            0,
			marginalReservedRate: 0.06,
		},
		{
			threshold:            1 * millicoresPerCore,
			marginalReservedRate: 0.01,
		},
		{
			threshold:            2 * millicoresPerCore,
			marginalReservedRate: 0.005,
		},
		{
			threshold:            4 * millicoresPerCore,
			marginalReservedRate: 0.0025,
		},
	})
}

// calculateReserved calculates reserved using capacity and a series of
// brackets as follows:  the marginalReservedRate applies to all capacity
// greater than the bracket, but less than the next bracket.  For example, if
// the first bracket is threshold: 0, rate:0.1, and the second bracket has
// threshold: 100, rate: 0.4, a capacity of 100 results in a reserved of
// 100*0.1 = 10, but a capacity of 200 results in a reserved of
// 10 + (200-100)*.4 = 50.  Using brackets with marginal rates ensures that as
// capacity increases, reserved always increases, and never decreases.
func calculateReserved(capacity int64, brackets []allocatableBracket) int64 {
	var reserved float64
	for i, bracket := range brackets {
		c := capacity
		if i < len(brackets)-1 && brackets[i+1].threshold < capacity {
			c = brackets[i+1].threshold
		}
		additionalReserved := float64(c-bracket.threshold) * bracket.marginalReservedRate
		if additionalReserved > 0 {
			reserved += additionalReserved
		}
	}
	return int64(reserved)
}
