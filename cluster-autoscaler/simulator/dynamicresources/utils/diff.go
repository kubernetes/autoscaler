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

package utils

import (
	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

type resourceSliceSpecs struct {
	driver string
	pool   string
}

// MissingResourcePool represents a missing resource pool found during resource slices
// comparison.
type MissingResourcePool struct {
	Driver  string
	Pool    string
	Devices []string
}

// WalkMissingResourcePools walks the resource slices and calls the onMissingResourcePool
// callback for each missing resource pool. Walker stops when the callback returns false.
// Comparison is strictly based on driver names, device names and resource pool shapes.
// No other properties of the resource slices / devices are compared.
//
// Superset naming should be treated like one in regards to resource pools, not
// in terms of devices themselves.
func WalkMissingResourcePools(
	subset []*resourceapi.ResourceSlice,
	superset []*resourceapi.ResourceSlice,
	onMissingResourcePool func(MissingResourcePool) bool,
) {
	supersetDevicesBySpec := getDevicesBySpecs(superset)
	subsetDevicesBySpec := getDevicesBySpecs(subset)

	for targetSpec, targetDevices := range supersetDevicesBySpec {
		foundResourcePool := false
		for consideredSpec, consideredDevices := range subsetDevicesBySpec {
			if targetSpec.driver != consideredSpec.driver {
				continue
			}

			// Amount of devices in the subset pool is smaller than in the target pool.
			// There's no way to find the missing devices, so we skip this pool.
			if len(targetDevices) > len(consideredDevices) {
				continue
			}

			diff := targetDevices.SymmetricDifference(consideredDevices)
			if len(diff) == 0 {
				foundResourcePool = true
				delete(subsetDevicesBySpec, consideredSpec)
				break
			}
		}

		if !foundResourcePool {
			missing := MissingResourcePool{Driver: targetSpec.driver, Pool: targetSpec.pool, Devices: targetDevices.UnsortedList()}
			if !onMissingResourcePool(missing) {
				return
			}
		}
	}
}

// GetMissingSliceDevices returns the set of resource pool which are missing from the subset resource slices list.
// For comparison purposes, the resource pool shape is considered to be the same if the their device names exactly match.
func GetMissingSliceDevices(subset []*resourceapi.ResourceSlice, superset []*resourceapi.ResourceSlice) []MissingResourcePool {
	missingDevices := make([]MissingResourcePool, 0)
	WalkMissingResourcePools(subset, superset, func(missingDevice MissingResourcePool) bool {
		missingDevices = append(missingDevices, missingDevice)
		return true
	})
	return missingDevices
}

// HasSliceDevices checks if the node has all resource poos defined in the superset are present in the subset.
// For comparison purposes, the resource pool shape is considered to be the same if the their device names exactly match.
func HasSliceDevices(subset []*resourceapi.ResourceSlice, superset []*resourceapi.ResourceSlice) bool {
	matching := true
	WalkMissingResourcePools(subset, superset, func(missingDevice MissingResourcePool) bool {
		matching = false
		return false
	})
	return matching
}

// getDevicesBySpecs groups the resource slices by their specs (pool, driver) and returns a map of device names.
func getDevicesBySpecs(resourceSlices []*resourceapi.ResourceSlice) map[resourceSliceSpecs]sets.Set[string] {
	groupedDevices := make(map[resourceSliceSpecs]sets.Set[string])
	for _, rs := range resourceSlices {
		spec := resourceSliceSpecs{pool: rs.Spec.Pool.Name, driver: rs.Spec.Driver}

		if _, found := groupedDevices[spec]; !found {
			groupedDevices[spec] = make(sets.Set[string], len(rs.Spec.Devices))
		}

		for _, device := range rs.Spec.Devices {
			groupedDevices[spec].Insert(device.Name)
		}
	}

	return groupedDevices
}
