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

package comparator

import (
	"fmt"
	"hash/maphash"
	"math"
	"slices"
	"strings"

	v1 "k8s.io/api/resource/v1"
)

const (
	// countOfDriversEstimate is an estimate of the number of DRA drivers in the cluster.
	// In case this assumption doesn't hold true - the performance of the algorithm would
	// degrade by increasing amount of allocations.
	countOfDriversEstimate = 10
)

var (
	mapHashSeed = maphash.MakeSeed()
)

type resourceDeltaType uint8

const (
	resourceDeltaTypeUnknown  resourceDeltaType = 0
	resourceDeltaTypeMissing  resourceDeltaType = 1
	resourceDeltaTypeExtra    resourceDeltaType = 2
	resourceDeltaTypeMismatch resourceDeltaType = 3
)

type signatureMap = map[v1.QualifiedName]v1.DeviceAttribute

// resourceDelta represents a discrepancy between expected and actual DRA resource topologies.
type resourceDelta struct {
	// DeviceCountDelta is the difference in the number of devices between the template and the node.
	DeviceCountDelta int64
	// Driver is the name of the driver.
	Driver string
	// TemplateResourcePool is the name of the resource pool in the template,
	// is empty if there's no matching node resource pool
	TemplateResourcePool string
	// NodeResourcePool is the name of the resource pool in the node,
	// is empty if there's no matching template resource pool
	NodeResourcePool string
	// TemplateSignatureMap is the signature of the attributes of the resource pool.
	TemplateSignatureMap signatureMap
	// NodeSignatureMap is the signature of the attributes of the resource pool.
	NodeSignatureMap signatureMap
}

// MissingAttributes returns the list of attributes that are present in the template but not in the node.
func (d *resourceDelta) MissingAttributes() []string {
	return mapKeyDifference(d.TemplateSignatureMap, d.NodeSignatureMap)
}

// ExtraAttributes returns the list of attributes that are present in the node but not in the template.
func (d *resourceDelta) ExtraAttributes() []string {
	return mapKeyDifference(d.NodeSignatureMap, d.TemplateSignatureMap)
}

// TemplateSignature returns the list of attributes of the resource pool in the template.
func (d *resourceDelta) TemplateSignature() []string {
	return mapKeys(d.TemplateSignatureMap)
}

// NodeSignature returns the list of attributes of the resource pool in the node.
func (d *resourceDelta) NodeSignature() []string {
	return mapKeys(d.NodeSignatureMap)
}

func (d *resourceDelta) Type() resourceDeltaType {
	if d.TemplateResourcePool == "" && d.NodeResourcePool == "" {
		return resourceDeltaTypeUnknown
	}
	if d.TemplateResourcePool == "" {
		return resourceDeltaTypeExtra
	}
	if d.NodeResourcePool == "" {
		return resourceDeltaTypeMissing
	}

	return resourceDeltaTypeMismatch
}

// Summary returns a human-readable summary of the discrepancy.
func (d *resourceDelta) Summary() string {
	switch d.Type() {
	case resourceDeltaTypeMissing:
		return fmt.Sprintf("MissingResource{Driver=%q, ResourcePool=%q, DeviceCount=%q, AttributeSignature=%q}",
			d.Driver, d.TemplateResourcePool, fmt.Sprintf("%d", d.DeviceCountDelta), strings.Join(d.TemplateSignature(), ";"))
	case resourceDeltaTypeExtra:
		return fmt.Sprintf("ExtraResource{Driver=%q, ResourcePool=%q, DeviceCount=%q, AttributeSignature=%q}",
			d.Driver, d.NodeResourcePool, fmt.Sprintf("%d", -d.DeviceCountDelta), strings.Join(d.NodeSignature(), ";"))
	case resourceDeltaTypeMismatch:
		return fmt.Sprintf("MismatchResource{Driver=%q, TemplatePool=%q, NodePool=%q, DeviceCountDelta=%q, MissingAttributes=%q, ExtraAttributes=%q}",
			d.Driver, d.TemplateResourcePool, d.NodeResourcePool, fmt.Sprintf("%d", d.DeviceCountDelta),
			strings.Join(d.MissingAttributes(), ";"), strings.Join(d.ExtraAttributes(), ";"))
	default:
		return fmt.Sprintf("UnknownResourceDelta{%+v}", d)
	}
}

// resourceGroup represents a group of devices with the same attributes, driver and resource pool.
type resourceGroup struct {
	driver      string
	pool        string
	deviceCount uint64
	signature   uint64
	attrs       map[v1.QualifiedName]v1.DeviceAttribute
}

// resourceGroupIdentifier is used to identify a group of devices with the same attributes and driver.
type resourceGroupIdentifier struct {
	driver    string
	signature uint64
}

// compareDraResources compares the resource topologies of the template and the node and
// returns a list of ResourceDelta that represent the differences between the two.
// Function uses a three-phase approach to find matching resource groups:
//
// Phase 1: Find and eliminate resource groups that match exactly
// Phase 2: Find and eliminate resource groups that match with the best overlap and report discrepancies
// Phase 3: Report the differences not reported by other phases (e.g. completely different pools, missing attributes)
//
// Warnings:
//   - Resource pool exposed for a driver X in the real node is only checked if the template
//     exposes at least one resource pool for the same driver, otherwise - it's ignored.
//   - Drivers which have at least a single resource pool with multiple generation numbers
//     or incomplete resource pools are not compared.
func compareDraResources(
	templateSlices []*v1.ResourceSlice,
	nodeSlices []*v1.ResourceSlice,
	deltasBuffer []resourceDelta,
) []resourceDelta {
	driversToCompare := findComparedDrivers(templateSlices, nodeSlices)
	templateResources := buildResourceGroups(templateSlices, driversToCompare)
	nodeResources := buildResourceGroups(nodeSlices, driversToCompare)
	// For node resources we maintain a map of identifier to index in the nodeResources slice.
	// This is used to quickly find a matching resource group in the node resources,
	// in case of hash collision we iterate over the node resources to find the matching resource group instead
	nodeResourceAddresses := make(map[resourceGroupIdentifier]int, len(nodeResources))
	for i, n := range nodeResources {
		identifier := resourceGroupIdentifier{driver: n.driver, signature: n.signature}
		nodeResourceAddresses[identifier] = i
	}

	// Phase 1: Exact Match
	for i := len(templateResources) - 1; i >= 0; i-- {
		t := templateResources[i]
		identifier := resourceGroupIdentifier{driver: t.driver, signature: t.signature}
		index, ok := nodeResourceAddresses[identifier]
		if !ok {
			continue
		}

		// Fast path: check if the resource group at the found index matches exactly
		if attributesMatch(t.attrs, nodeResources[index].attrs) && t.deviceCount == nodeResources[index].deviceCount {
			templateResources = swapDelete(templateResources, i)
			nodeResources = swapDelete(nodeResources, index)

			// Update address map with a new moved element and clear old entry
			delete(nodeResourceAddresses, identifier)
			if index < len(nodeResources) {
				movedIdentifier := resourceGroupIdentifier{
					driver:    nodeResources[index].driver,
					signature: nodeResources[index].signature,
				}
				nodeResourceAddresses[movedIdentifier] = index
			}

			continue
		}

		// Slow path: in case of a hash collision, iterate over the node resources to find the matching resource group
		for j, n := range nodeResources {
			if n.signature != t.signature || n.driver != t.driver {
				continue
			}

			if attributesMatch(t.attrs, n.attrs) && t.deviceCount == n.deviceCount {
				templateResources = swapDelete(templateResources, i)
				nodeResources = swapDelete(nodeResources, j)

				// Update address map with a new moved element and clear old entry
				delete(nodeResourceAddresses, identifier)
				if j < len(nodeResources) {
					movedIdentifier := resourceGroupIdentifier{
						driver:    nodeResources[j].driver,
						signature: nodeResources[j].signature,
					}
					nodeResourceAddresses[movedIdentifier] = j
				}

				break
			}
		}
	}

	// Phase 2 & 3: Fuzzy Match & Diff Reporting
	for i := len(templateResources) - 1; i >= 0; i-- {
		t := templateResources[i]
		bestMatchIndex := -1
		bestOverlap := -1
		bestMatchCountDelta := int64(math.MaxInt64)

		// Find the best unmatched candidate
		for j, n := range nodeResources {
			if t.driver != n.driver {
				continue
			}

			overlap := computeSetOverlap(t.attrs, n.attrs)
			deviceCountDelta := int64(t.deviceCount) - int64(n.deviceCount)

			if overlap < bestOverlap {
				continue
			}

			// Primary objective: Maximize attribute overlap
			// Secondary objective: Minimize difference in device counts
			if overlap > bestOverlap || absInt64(deviceCountDelta) < absInt64(bestMatchCountDelta) {
				bestMatchCountDelta = deviceCountDelta
				bestOverlap = overlap
				bestMatchIndex = j
			}
		}

		// Completely unmatched driver pool
		if bestMatchIndex == -1 {
			continue
		}

		bestMatch := nodeResources[bestMatchIndex]
		templateResources = swapDelete(templateResources, i)
		nodeResources = swapDelete(nodeResources, bestMatchIndex)
		hasMissing := mapsHaveKeyDifference(t.attrs, bestMatch.attrs)
		hasExtra := mapsHaveKeyDifference(bestMatch.attrs, t.attrs)

		// Only report if a real diff exists (e.g. different attributes or device count)
		if hasMissing || hasExtra || bestMatchCountDelta != 0 {
			deltasBuffer = append(deltasBuffer, resourceDelta{
				Driver:               t.driver,
				TemplateResourcePool: t.pool,
				NodeResourcePool:     bestMatch.pool,
				DeviceCountDelta:     bestMatchCountDelta,
				TemplateSignatureMap: t.attrs,
				NodeSignatureMap:     bestMatch.attrs,
			})
		}
	}

	// Report on remaining template resources
	for _, t := range templateResources {
		deltasBuffer = append(deltasBuffer, resourceDelta{
			Driver:               t.driver,
			TemplateResourcePool: t.pool,
			TemplateSignatureMap: t.attrs,
			DeviceCountDelta:     int64(t.deviceCount),
		})
	}

	// Report on remaining node resources
	for _, n := range nodeResources {
		deltasBuffer = append(deltasBuffer, resourceDelta{
			Driver:           n.driver,
			NodeResourcePool: n.pool,
			NodeSignatureMap: n.attrs,
			DeviceCountDelta: -int64(n.deviceCount),
		})
	}

	return deltasBuffer
}

// poolReference is used to identify a resource pool.
type poolReference struct {
	driver string
	pool   string
}

type poolState struct {
	generation    int64
	count         int64
	completeCount int64
}

// findComparedDrivers returns a list of drivers that should be compared using following
// algorithm:
//  1. Collect all drivers present in the template - those are drivers we care about.
//  2. Detect drivers in flux (drivers with at least a single resource pool with
//     multiple generation numbers).
//  3. Return the list of drivers that are not in flux.
func findComparedDrivers(templateSlices, nodeSlices []*v1.ResourceSlice) []string {
	templateDrivers := make([]string, 0, countOfDriversEstimate)
	for _, t := range templateSlices {
		if slices.Contains(templateDrivers, t.Spec.Driver) {
			continue
		}

		templateDrivers = append(templateDrivers, t.Spec.Driver)
	}

	driversInFlux := detectDriversInFlux(nodeSlices)

	comparedDrivers := make([]string, 0, len(templateDrivers))
	for _, driver := range templateDrivers {
		if slices.Contains(driversInFlux, driver) {
			continue
		}

		comparedDrivers = append(comparedDrivers, driver)
	}

	return comparedDrivers
}

// detectDriversInFlux detects drivers which have at least a single resource pool with multiple generation numbers.
func detectDriversInFlux(resourceSlices []*v1.ResourceSlice) []string {
	poolStates := make(map[poolReference]poolState, len(resourceSlices))
	var driversInFlux []string
	for _, rs := range resourceSlices {
		poolDriver := rs.Spec.Driver
		if slices.Contains(driversInFlux, poolDriver) {
			continue
		}

		poolGeneration := rs.Spec.Pool.Generation
		poolRef := poolReference{driver: poolDriver, pool: rs.Spec.Pool.Name}

		if state, ok := poolStates[poolRef]; ok {
			if state.generation != poolGeneration {
				driversInFlux = append(driversInFlux, poolDriver)
			}

			state.count++
			poolStates[poolRef] = state
			continue
		}

		poolStates[poolRef] = poolState{
			generation:    poolGeneration,
			completeCount: rs.Spec.Pool.ResourceSliceCount,
			count:         1,
		}
	}

	for ref, state := range poolStates {
		if state.count != state.completeCount && !slices.Contains(driversInFlux, ref.driver) {
			driversInFlux = append(driversInFlux, ref.driver)
		}
	}

	return driversInFlux
}

// poolShapeIdentifier is used to identify a group of devices with the same attributes, driver and resource pool.
type poolShapeIdentifier struct {
	driver    string
	pool      string
	signature uint64
}

// buildResourceGroups extracts resource groups from resource slices and splits them on a basis
// of driver, pool and attribute signature.
func buildResourceGroups(resourceSlices []*v1.ResourceSlice, allowedDrivers []string) []resourceGroup {
	groups := make([]resourceGroup, 0, len(resourceSlices))
	address := make(map[poolShapeIdentifier]int, len(resourceSlices))

	for _, rs := range resourceSlices {
		if allowedDrivers != nil && !slices.Contains(allowedDrivers, rs.Spec.Driver) {
			continue
		}

		driver := rs.Spec.Driver
		pool := rs.Spec.Pool.Name
		for _, device := range rs.Spec.Devices {
			signature := computeAttrMapHash(device.Attributes)
			identifier := poolShapeIdentifier{driver: driver, pool: pool, signature: signature}

			foundGroup := false
			if idx, ok := address[identifier]; ok {
				// Fast path: if the attributes match the attributes of the resource group at the found index
				if attributesMatch(groups[idx].attrs, device.Attributes) {
					groups[idx].deviceCount++
					foundGroup = true
				} else {
					// Slow path: in case of a hash collision, iterate over the resource groups to find the matching resource group
					for i := range groups {
						if groups[i].signature == signature && groups[i].driver == driver && groups[i].pool == pool {
							if attributesMatch(groups[i].attrs, device.Attributes) {
								groups[i].deviceCount++
								foundGroup = true
								break
							}
						}
					}
				}
			}

			if foundGroup {
				continue
			}

			groups = append(groups, resourceGroup{
				driver:      driver,
				pool:        pool,
				attrs:       device.Attributes,
				signature:   signature,
				deviceCount: 1,
			})
			address[identifier] = len(groups) - 1
		}
	}

	return groups
}

// attributesMatch checks if two maps of attributes exactly match.
func attributesMatch(attributes, deviceAttributes map[v1.QualifiedName]v1.DeviceAttribute) bool {
	if len(attributes) != len(deviceAttributes) {
		return false
	}

	for k := range deviceAttributes {
		if _, ok := attributes[k]; !ok {
			return false
		}
	}

	return true
}

// computeSetOverlap counts overlap without allocating a new set.
func computeSetOverlap(a, b map[v1.QualifiedName]v1.DeviceAttribute) int {
	count := 0
	// Always iterate over the smaller map
	if len(a) > len(b) {
		a, b = b, a
	}
	for k := range a {
		if _, exists := b[k]; exists {
			count++
		}
	}
	return count
}

// computeAttrMapHash computes a hash of the attribute map. It uses an XOR operation
// on the individual key hashes to guarantee order-independent determinism without
// requiring memory allocations for sorting strings.
func computeAttrMapHash(keys map[v1.QualifiedName]v1.DeviceAttribute) uint64 {
	var hash uint64

	for k := range keys {
		hash ^= maphash.String(mapHashSeed, string(k))
	}

	return hash
}

// swapDelete removes an element from a slice by swapping it with the last
// element and shrinking the slice.
func swapDelete[T any](s []T, i int) []T {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

// absInt64 returns the absolute value of an int64.
func absInt64(a int64) int64 {
	if a < 0 {
		return -a
	}
	return a
}

// mapKeyDifference returns the keys that are in the base map but not in the subtract map.
func mapKeyDifference(base, subtract map[v1.QualifiedName]v1.DeviceAttribute) []string {
	var diff []string
	for k := range base {
		if _, ok := subtract[k]; !ok {
			diff = append(diff, string(k))
		}
	}
	return diff
}

// mapsHaveKeyDifference checks if two maps have any difference.
func mapsHaveKeyDifference(base, subtract map[v1.QualifiedName]v1.DeviceAttribute) bool {
	if len(base) != len(subtract) {
		return true
	}

	for k := range base {
		if _, ok := subtract[k]; !ok {
			return true
		}
	}

	return false
}

// mapKeys extracts the keys from a map and returns them as a slice.
func mapKeys(m map[v1.QualifiedName]v1.DeviceAttribute) []string {
	res := make([]string, 0, len(m))
	for k := range m {
		res = append(res, string(k))
	}
	return res
}
