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

package utils

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

// CalculateDynamicResourceUtilization calculates a map of ResourceSlice pool utilization grouped by the driver and pool. Returns
// an error if the NodeInfo doesn't have all ResourceSlices from a pool.
func CalculateDynamicResourceUtilization(nodeInfo *framework.NodeInfo) (map[string]map[string]float64, error) {
	result := map[string]map[string]float64{}
	claims := nodeInfo.ResourceClaims()
	allocatedDevices, err := groupAllocatedDevices(claims)
	if err != nil {
		return nil, err
	}
	for driverName, slicesByPool := range groupSlices(nodeInfo.LocalResourceSlices) {
		result[driverName] = map[string]float64{}
		for poolName, poolSlices := range slicesByPool {
			currentSlices, err := allCurrentGenSlices(poolSlices)
			if err != nil {
				return nil, fmt.Errorf("pool %q error: %v", poolName, err)
			}
			poolDevices := getAllDevices(currentSlices)
			allocatedDeviceNames := allocatedDevices[driverName][poolName]
			unallocated, allocated := splitDevicesByAllocation(poolDevices, allocatedDeviceNames)
			result[driverName][poolName] = calculatePoolUtil(unallocated, allocated, currentSlices)
		}
	}
	return result, nil
}

// HighestDynamicResourceUtilization returns the ResourceSlice driver and pool with the highest utilization.
func HighestDynamicResourceUtilization(nodeInfo *framework.NodeInfo) (v1.ResourceName, float64, error) {
	utils, err := CalculateDynamicResourceUtilization(nodeInfo)
	if err != nil {
		return "", 0, err
	}

	highestUtil := 0.0
	var highestResourceName v1.ResourceName
	for driverName, utilsByPool := range utils {
		for poolName, util := range utilsByPool {
			if util >= highestUtil {
				highestUtil = util
				highestResourceName = v1.ResourceName(driverName + "/" + poolName)
			}
		}
	}
	return highestResourceName, highestUtil, nil
}

func calculatePoolUtil(unallocated, allocated []resourceapi.Device, resourceSlices []*resourceapi.ResourceSlice) float64 {
	TotalConsumedCounters := map[string]map[string]resource.Quantity{}
	for _, resourceSlice := range resourceSlices {
		for _, sharedCounter := range resourceSlice.Spec.SharedCounters {
			if _, ok := TotalConsumedCounters[sharedCounter.Name]; !ok {
				TotalConsumedCounters[sharedCounter.Name] = map[string]resource.Quantity{}
			}
			for counter, value := range sharedCounter.Counters {
				TotalConsumedCounters[sharedCounter.Name][counter] = value.Value
			}
		}
	}
	allocatedConsumedCounters := calculateConsumedCounters(allocated)

	// not all devices are partitionable, so fallback to the ratio of non-partionable devices
	allocatedDevicesWithoutCounters := 0
	devicesWithoutCounters := 0

	for _, device := range allocated {
		if device.ConsumesCounters == nil {
			devicesWithoutCounters++
			allocatedDevicesWithoutCounters++
		}
	}
	for _, device := range unallocated {
		if device.ConsumesCounters == nil {
			devicesWithoutCounters++
		}
	}

	// we want to find the counter that is most utilized, since it is the "bottleneck" of the pool
	var partitionableUtilization float64 = 0
	var atomicDevicesUtilization float64 = 0
	if devicesWithoutCounters != 0 {
		atomicDevicesUtilization = float64(allocatedDevicesWithoutCounters) / float64(devicesWithoutCounters)
	}
	if len(TotalConsumedCounters) == 0 {
		return atomicDevicesUtilization
	}
	for counterSet, counters := range TotalConsumedCounters {
		for counterName, totalValue := range counters {
			if totalValue.IsZero() {
				continue
			}
			if allocatedSet, exists := allocatedConsumedCounters[counterSet]; exists {
				if allocatedValue, exists := allocatedSet[counterName]; exists {
					utilization := float64(allocatedValue.Value()) / float64(totalValue.Value())
					if utilization > partitionableUtilization {
						partitionableUtilization = utilization
					}
				}
			}
		}
	}
	var uniquePartitionableDevicesCount float64 = float64(getUniquePartitionableDevicesCount(allocated))
	var totalUniqueDevices float64 = uniquePartitionableDevicesCount + float64(devicesWithoutCounters)
	var partitionableDevicesUtilizationWeight float64 = uniquePartitionableDevicesCount / totalUniqueDevices
	var nonPartitionableDevicesUtilizationWeight float64 = 1 - partitionableDevicesUtilizationWeight
	// when a pool has both atomic and partitionable devices, we sum their utilizations since they are mutually exclusive
	return partitionableUtilization * partitionableDevicesUtilizationWeight + atomicDevicesUtilization * nonPartitionableDevicesUtilizationWeight
}

// calculateConsumedCounters calculates the total counters consumed by a list of devices
func calculateConsumedCounters(devices []resourceapi.Device) map[string]map[string]resource.Quantity {
	countersConsumed := map[string]map[string]resource.Quantity{}
	for _, device := range devices {
		if device.ConsumesCounters == nil {
			continue
		}
		for _, consumedCounter := range device.ConsumesCounters {
			if _, ok := countersConsumed[consumedCounter.CounterSet]; !ok {
				countersConsumed[consumedCounter.CounterSet] = map[string]resource.Quantity{}
			}
			for counter, value := range consumedCounter.Counters {
				if _, ok := countersConsumed[consumedCounter.CounterSet][counter]; !ok {
					countersConsumed[consumedCounter.CounterSet][counter] = resource.Quantity{}
				}
				v := countersConsumed[consumedCounter.CounterSet][counter]
				v.Add(value.Value)
				countersConsumed[consumedCounter.CounterSet][counter] = v
			}
		}
	}
	return countersConsumed
}

// getUniquePartitionableDevicesCount returns the count of unique partitionable devices in the provided list.
// a partitionable device can be represented by multiple devices with different names and properties, and for utilization purposes we'd like to count the hardware and not the software.
func getUniquePartitionableDevicesCount(devices []resourceapi.Device) int {
	var deviceCount int = 0
	var counted bool
	consumedCounters := map[string]bool{}
	for _, device := range devices {
		// the assumption here is that a partitionable device will consume the actual resources from the hardware, which will be represented by consumedCounters.
		// if a device consumes multiple counters of the same device, we count them both at the same time in order to not "overcount" devices with multiple counters, the assumption here is that a device will always consume some of every resource in a device. (f.e. a GPU DRA request cannot use VRAM without using GPU cycles and vice versa)
		if device.ConsumesCounters != nil {
			counted = false
			for _, consumedCounter := range device.ConsumesCounters {
				if _, exists := consumedCounters[consumedCounter.CounterSet]; !exists {
					consumedCounters[consumedCounter.CounterSet] = true
				} else {
					counted = true
				}
			}
			if !counted {
				deviceCount++
			}
		}
	}
	return deviceCount
}

func splitDevicesByAllocation(devices []resourceapi.Device, allocatedNames []string) (unallocated, allocated []resourceapi.Device) {
	allocatedNamesSet := map[string]bool{}
	for _, allocatedName := range allocatedNames {
		allocatedNamesSet[allocatedName] = true
	}
	for _, device := range devices {
		if allocatedNamesSet[device.Name] {
			allocated = append(allocated, device)
		} else {
			unallocated = append(unallocated, device)
		}
	}
	return unallocated, allocated
}

// getAllDevices aggregates all Devices from the provided ResourceSlices into one list.
func getAllDevices(slices []*resourceapi.ResourceSlice) []resourceapi.Device {
	var devices []resourceapi.Device
	for _, slice := range slices {
		devices = append(devices, slice.Spec.Devices...)
	}
	return devices
}

// groupAllocatedDevices groups the devices from claim allocations by their driver and pool. Returns an error
// if any of the claims isn't allocated.
func groupAllocatedDevices(claims []*resourceapi.ResourceClaim) (map[string]map[string][]string, error) {
	result := map[string]map[string][]string{}
	for _, claim := range claims {
		alloc := claim.Status.Allocation
		if alloc == nil {
			return nil, fmt.Errorf("claim %s/%s not allocated", claim.Namespace, claim.Name)
		}

		for _, deviceAlloc := range alloc.Devices.Results {
			if result[deviceAlloc.Driver] == nil {
				result[deviceAlloc.Driver] = map[string][]string{}
			}
			result[deviceAlloc.Driver][deviceAlloc.Pool] = append(result[deviceAlloc.Driver][deviceAlloc.Pool], deviceAlloc.Device)
		}
	}
	return result, nil
}

// groupSlices groups the provided slices by their driver and pool name.
func groupSlices(slices []*resourceapi.ResourceSlice) map[string]map[string][]*resourceapi.ResourceSlice {
	result := map[string]map[string][]*resourceapi.ResourceSlice{}
	for _, slice := range slices {
		driver := slice.Spec.Driver
		pool := slice.Spec.Pool.Name
		if result[driver] == nil {
			result[driver] = map[string][]*resourceapi.ResourceSlice{}
		}
		result[driver][pool] = append(result[driver][pool], slice)
	}
	return result
}

// allCurrentGenSlices filters out slices that aren't from the newest pool generation, and returns an error
// if not all slices from the newest generations are provided.
func allCurrentGenSlices(slices []*resourceapi.ResourceSlice) ([]*resourceapi.ResourceSlice, error) {
	var maxGenSlices []*resourceapi.ResourceSlice
	maxGen := int64(0)
	for _, slice := range slices {
		gen := slice.Spec.Pool.Generation
		if gen > maxGen {
			maxGen = gen
			maxGenSlices = []*resourceapi.ResourceSlice{slice}
			continue
		}
		if gen == maxGen {
			maxGenSlices = append(maxGenSlices, slice)
		}
	}

	foundCurrentSlices := len(maxGenSlices)
	if foundCurrentSlices == 0 {
		return nil, nil
	}

	if wantCurrentSlices := maxGenSlices[0].Spec.Pool.ResourceSliceCount; int64(foundCurrentSlices) != wantCurrentSlices {
		return nil, fmt.Errorf("newest generation: %d, slice count: %d - found only %d slices", maxGen, wantCurrentSlices, foundCurrentSlices)
	}

	return maxGenSlices, nil
}
