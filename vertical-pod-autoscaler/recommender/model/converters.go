/*
Copyright 2017 The Kubernetes Authors.

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
	"fmt"
)

// CPUAmountFromCores converts CPU cores to a ResourceAmount.
func CPUAmountFromCores(cores float64) ResourceAmount {
	return ResourceAmount(cores * 1000.0)
}

// CoresFromCPUAmount converts ResourceAmount to number of cores expressed as float64
func CoresFromCPUAmount(cpuAmunt ResourceAmount) float64 {
	return float64(cpuAmunt) / 1000.0
}

// MemoryAmountFromBytes converts memory bytes to a ResourceAmount.
func MemoryAmountFromBytes(bytes float64) ResourceAmount {
	return ResourceAmount(bytes)
}

// BytesFromMemoryAmount converts ResourceAmount to number byts expressed as float64
func BytesFromMemoryAmount(memoryAmount ResourceAmount) float64 {
	return float64(memoryAmount)
}

// UsageFromResourceAmount converts given ResourceAmount to usage expressed in floaf64, based on given MetricName.
func UsageFromResourceAmount(metric MetricName, ammount ResourceAmount) (float64, error) {
	switch metric {
	case ResourceCPU:
		return CoresFromCPUAmount(ammount), nil
	case ResourceMemory:
		return BytesFromMemoryAmount(ammount), nil
	default:
		return 0, fmt.Errorf("Type conversion for MetricName '+%v' is not defined", metric)
	}
}
