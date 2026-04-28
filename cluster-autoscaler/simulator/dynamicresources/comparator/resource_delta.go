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
	"hash/maphash"
	"math"
	"slices"

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
	resourceDeltaTypeUnknown  resourceDeltaType = 1
	resourceDeltaTypeMissing  resourceDeltaType = 2
	resourceDeltaTypeExtra    resourceDeltaType = 3
	resourceDeltaTypeMismatch resourceDeltaType = 4
)

func (t resourceDeltaType) String() string {
	switch t {
	case resourceDeltaTypeMissing:
		return "Missing"
	case resourceDeltaTypeExtra:
		return "Extra"
	case resourceDeltaTypeMismatch:
		return "Mismatch"
	default:
		return "Unknown"
	}
}

type attributesMap map[v1.QualifiedName]v1.DeviceAttribute

// resourceDelta represents a discrepancy between expected and actual DRA resource topologies.
//
// State Matrix:
//
// | TemplatePool | NodePool | Delta Type | DeviceCountDelta          |
// |--------------|----------|------------|---------------------------|
// | Present      | Present  | Mismatch   | TemplateCount - NodeCount |
// | Present      | Empty    | Missing    | TemplateCount (Positive)  |
// | Empty        | Present  | Extra      | -NodeCount (Negative)     |
// | Empty        | Empty    | Unknown    | 0                         |
type resourceDelta struct {
	// TemplateSignatureMap is the signature of the attributes of the resource pool.
	TemplateSignatureMap attributesMap
	// NodeSignatureMap is the signature of the attributes of the resource pool.
	NodeSignatureMap attributesMap
	// Driver is the name of the driver.
	Driver string
	// TemplateResourcePool is the name of the resource pool in the template,
	// is empty if there's no matching node resource pool
	TemplateResourcePool string
	// NodeResourcePool is the name of the resource pool in the node,
	// is empty if there's no matching template resource pool
	NodeResourcePool string
	// DeviceCountDelta is the difference in the number of devices between the template and the node.
	DeviceCountDelta int64
}

// Type returns the type of the resource delta.
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

// resourceGroup represents a group of devices with the same attributes, driver and resource pool.
type resourceGroup struct {
	attrs       attributesMap
	driver      string
	pool        string
	deviceCount uint64
	signature   uint64
}

// nodeResourceIdentifier is used to identify a group of devices with the same attributes and driver.
type nodeResourceIdentifier struct {
	driver    string
	signature uint64
}

// resourcePoolIdentifier is used to identify a group of devices with the same attributes, driver and resource pool.
type resourcePoolIdentifier struct {
	driver    string
	pool      string
	signature uint64
}

// poolReference is used to identify a resource pool.
type poolReference struct {
	driver string
	pool   string
}

// poolState is used to track the state of a resource pool.
type poolState struct {
	generation    int64
	count         int64
	completeCount int64
}

// resourcePoolComparator is used to compare the resource topologies of the template and the node.
// Component heavily relies on reusable buffers to minimize allocations and thus is not thread-safe
// and functions apart of CompareResourcePools are not recommended to be called directly as it may
// break integrity of the state.
type resourcePoolComparator struct {
	// poolStates keeps track of the state of each pool, it serves as a helper
	// for determining if a pool is in flux or not.
	poolStates map[poolReference]poolState
	// nodeResourceAddressMap is a map of node resource identifiers to indices
	// in the nodeResources slice, may contain collisions when driver has multiple
	// resource pools with the same signature.
	nodeResourceAddressMap map[nodeResourceIdentifier]int
	// nodePoolAddressMap is a map of node resource pool identifiers to indices
	// in the nodeResources slice
	nodePoolAddressMap map[resourcePoolIdentifier]int
	// templatePoolAddressMap is a map of template resource pool identifiers to indices
	// in the templateResources slice
	templatePoolAddressMap map[resourcePoolIdentifier]int
	// driversInFlux represents drivers which have at least a single resource
	// pool with multiple generation numbers or incomplete resource pools.
	driversInFlux []string
	// comparedDrivers represents drivers are compared by the comparator.
	comparedDrivers []string
	// templateResources represents a list of resource groups from the template.
	templateResources []resourceGroup
	// nodeResources represents a list of resource groups from the node.
	nodeResources []resourceGroup
}

// newResourcePoolComparator creates a new resource pool comparator.
func newResourcePoolComparator() resourcePoolComparator {
	return resourcePoolComparator{
		poolStates:             make(map[poolReference]poolState),
		nodeResourceAddressMap: make(map[nodeResourceIdentifier]int),
		nodePoolAddressMap:     make(map[resourcePoolIdentifier]int),
		templatePoolAddressMap: make(map[resourcePoolIdentifier]int),
	}
}

// resetBuffers resets the buffers used by the resource pool comparator.
func (c *resourcePoolComparator) resetBuffers() {
	clear(c.poolStates)
	clear(c.nodeResourceAddressMap)
	clear(c.nodePoolAddressMap)
	clear(c.templatePoolAddressMap)
	c.driversInFlux = c.driversInFlux[:0]
	c.comparedDrivers = c.comparedDrivers[:0]
	c.templateResources = c.templateResources[:0]
	c.nodeResources = c.nodeResources[:0]
}

// compareDraResources compares the resource topologies of the template and the node and
// returns a list of ResourceDelta that represent the differences between the two.
//
// Algorithm for deltas detection consists of the following steps:
// 1. Define a list of drivers used for the current comparison, in order to do that we search drivers
// which have at least a single resource pool being incomplete or having multiple generations as this
// point of time and remove them from the list of drivers exposed through template or nodes resource slices
// 2. Organize list of node and template slices into resource groups ignoring non-compared drivers, each
// resource group representing different resource pool or device attribute signature. We assume that resource
// pool consists of homogeneous set of devices, but if this assumption doesn't hold true - resource
// pool will be split into N resource groups where N is the amount of different device signatures available in it
// 3. Mutually eliminate node and template resource groups if they exactly match by the amount of devices and
// attribute signature, these resource groups were fulfilled and there's no deltas between them
// 4. Search and eliminate most similar resource groups among node and template slices prioritizing attribute
// signature overlap and device count (in order) when match found - difference between them is reported as a
// resource delta.
// 5. Left out resource groups from any source are reported as missing when they are coming from template but are
// missing on the node, and as extra when resource is only defined in the node, but missing from the template
//
// Warnings:
//   - Drivers which have at least a single resource pool with multiple generation numbers
//     or incomplete resource pools are not compared.
//   - Function is not thread-safe and manipulates internal buffers of the comparator
//     to minimize re-allocations.
//   - Device attribute values are not compared at any point of time, we only ensure
//     that the same set of attribute keys is present.
func (c *resourcePoolComparator) CompareResourcePools(
	templateSlices []*v1.ResourceSlice,
	nodeSlices []*v1.ResourceSlice,
	deltasBuffer []resourceDelta,
) []resourceDelta {
	c.resetBuffers()

	c.populateDriversInFlux(nodeSlices)
	c.populateComparedDrivers(templateSlices, nodeSlices)
	c.populateTemplateGroups(templateSlices)
	c.populateNodeGroups(nodeSlices)
	c.eliminateExactMatchResources()
	deltasBuffer = c.eliminateFuzzyMatchResources(deltasBuffer)

	for _, t := range c.templateResources {
		deltasBuffer = append(deltasBuffer, resourceDelta{
			Driver:               t.driver,
			TemplateResourcePool: t.pool,
			TemplateSignatureMap: t.attrs,
			DeviceCountDelta:     int64(t.deviceCount),
		})
	}

	// Report on remaining node resources
	for _, n := range c.nodeResources {
		deltasBuffer = append(deltasBuffer, resourceDelta{
			Driver:           n.driver,
			NodeResourcePool: n.pool,
			NodeSignatureMap: n.attrs,
			DeviceCountDelta: -int64(n.deviceCount),
		})
	}

	return deltasBuffer
}

// populateDriversInFlux detects drivers which have at least a single resource pool with multiple generation numbers.
func (c *resourcePoolComparator) populateDriversInFlux(resourceSlices []*v1.ResourceSlice) {
	for _, rs := range resourceSlices {
		poolDriver := rs.Spec.Driver
		if slices.Contains(c.driversInFlux, poolDriver) {
			continue
		}

		poolGeneration := rs.Spec.Pool.Generation
		poolRef := poolReference{driver: poolDriver, pool: rs.Spec.Pool.Name}

		if state, ok := c.poolStates[poolRef]; ok {
			if state.generation != poolGeneration {
				c.driversInFlux = append(c.driversInFlux, poolDriver)
			}
			state.count++
			c.poolStates[poolRef] = state
			continue
		}

		c.poolStates[poolRef] = poolState{
			generation:    poolGeneration,
			completeCount: rs.Spec.Pool.ResourceSliceCount,
			count:         1,
		}
	}

	for ref, state := range c.poolStates {
		if state.count != state.completeCount && !slices.Contains(c.driversInFlux, ref.driver) {
			c.driversInFlux = append(c.driversInFlux, ref.driver)
		}
	}
}

// populateComparedDrivers populates the list of drivers that should be compared, assumes that
// drivers in flux are already filtered out.
func (c *resourcePoolComparator) populateComparedDrivers(templateSlices []*v1.ResourceSlice, nodeSlices []*v1.ResourceSlice) {
	for _, t := range templateSlices {
		if slices.Contains(c.comparedDrivers, t.Spec.Driver) {
			continue
		}

		if slices.Contains(c.driversInFlux, t.Spec.Driver) {
			continue
		}

		c.comparedDrivers = append(c.comparedDrivers, t.Spec.Driver)
	}

	for _, n := range nodeSlices {
		if slices.Contains(c.comparedDrivers, n.Spec.Driver) {
			continue
		}

		if slices.Contains(c.driversInFlux, n.Spec.Driver) {
			continue
		}

		c.comparedDrivers = append(c.comparedDrivers, n.Spec.Driver)
	}
}

// populateTemplateGroups populates the template resource groups from the template slices.
func (c *resourcePoolComparator) populateTemplateGroups(templateSlices []*v1.ResourceSlice) {
	c.templateResources = c.buildResourceGroups(templateSlices, c.templateResources, c.templatePoolAddressMap)
}

// populateNodeGroups populates the node resource groups from the node slices.
func (c *resourcePoolComparator) populateNodeGroups(nodeSlices []*v1.ResourceSlice) {
	c.nodeResources = c.buildResourceGroups(nodeSlices, c.nodeResources, c.nodePoolAddressMap)
	for i, n := range c.nodeResources {
		identifier := nodeResourceIdentifier{driver: n.driver, signature: n.signature}
		c.nodeResourceAddressMap[identifier] = i
	}
}

// eliminateExactMatchResources eliminates resource groups that match exactly by attribute signatures and device counts.
func (c *resourcePoolComparator) eliminateExactMatchResources() {
	for i := len(c.templateResources) - 1; i >= 0; i-- {
		t := c.templateResources[i]
		identifier := nodeResourceIdentifier{driver: t.driver, signature: t.signature}
		index, ok := c.nodeResourceAddressMap[identifier]

		// Fast path: check if the resource group at the found index matches exactly
		if ok && attributesMatch(t.attrs, c.nodeResources[index].attrs) && t.deviceCount == c.nodeResources[index].deviceCount {
			c.templateResources = swapDelete(c.templateResources, i)
			c.nodeResources = swapDelete(c.nodeResources, index)
			c.deleteNodeResourceAddress(identifier, index)

			continue
		}

		// Slow path: in case of a hash collision, iterate over the node resources to find the matching resource group
		for j, n := range c.nodeResources {
			if n.signature != t.signature || n.driver != t.driver {
				continue
			}

			if attributesMatch(t.attrs, n.attrs) && t.deviceCount == n.deviceCount {
				c.templateResources = swapDelete(c.templateResources, i)
				c.nodeResources = swapDelete(c.nodeResources, j)
				c.deleteNodeResourceAddress(identifier, j)
				break
			}
		}
	}
}

// eliminateFuzzyMatchResources eliminates resource groups that match with the best overlap and reports discrepancies.
func (c *resourcePoolComparator) eliminateFuzzyMatchResources(deltasBuffer []resourceDelta) []resourceDelta {
	for i := len(c.templateResources) - 1; i >= 0; i-- {
		template := &c.templateResources[i]
		bestMatchIndex := c.findNodeFuzzyMatch(template)

		// Completely unmatched driver pool
		if bestMatchIndex == -1 {
			continue
		}

		bestMatch := c.nodeResources[bestMatchIndex]
		deviceCountDelta := int64(template.deviceCount) - int64(bestMatch.deviceCount)
		deltasBuffer = append(deltasBuffer, resourceDelta{
			Driver:               template.driver,
			TemplateResourcePool: template.pool,
			NodeResourcePool:     bestMatch.pool,
			DeviceCountDelta:     deviceCountDelta,
			TemplateSignatureMap: template.attrs,
			NodeSignatureMap:     bestMatch.attrs,
		})

		c.templateResources = swapDelete(c.templateResources, i)
		c.nodeResources = swapDelete(c.nodeResources, bestMatchIndex)
	}

	return deltasBuffer
}

// findNodeFuzzyMatch finds the best matching resource group available on the node
// and returns the index of the found group in the nodeResources list or -1 if none
// found
func (c *resourcePoolComparator) findNodeFuzzyMatch(target *resourceGroup) int {
	bestMatchIndex := -1
	bestOverlap := -1
	bestDeviceCountDelta := int64(math.MaxInt64)

	for j, node := range c.nodeResources {
		if target.driver != node.driver {
			continue
		}

		overlap := computeSetOverlap(target.attrs, node.attrs)
		deviceCountDelta := int64(target.deviceCount) - int64(node.deviceCount)

		if overlap < bestOverlap {
			continue
		}

		// Primary objective: Maximize attribute overlap
		// Secondary objective: Minimize difference in device counts
		if overlap > bestOverlap || absInt64(deviceCountDelta) < absInt64(bestDeviceCountDelta) {
			bestDeviceCountDelta = deviceCountDelta
			bestOverlap = overlap
			bestMatchIndex = j
		}
	}

	return bestMatchIndex
}

// deleteNodeResourceAddress deletes the node resource address from the map and updates the map with a new moved element.
func (c *resourcePoolComparator) deleteNodeResourceAddress(identifier nodeResourceIdentifier, index int) {
	delete(c.nodeResourceAddressMap, identifier)
	if index < len(c.nodeResources) {
		movedIdentifier := nodeResourceIdentifier{
			driver:    c.nodeResources[index].driver,
			signature: c.nodeResources[index].signature,
		}
		c.nodeResourceAddressMap[movedIdentifier] = index
	}
}

// ingestPoolDevice ingests device into a resource group by either placing it into one of the
// known groups or creating one if it's not found
func (c *resourcePoolComparator) ingestPoolDevice(
	driver string,
	pool string,
	attributes attributesMap,
	groups []resourceGroup,
	addresses map[resourcePoolIdentifier]int,
) []resourceGroup {
	hash := computeAttrMapHash(attributes)
	identifier := resourcePoolIdentifier{driver: driver, pool: pool, signature: hash}

	if idx, ok := addresses[identifier]; ok {
		// Fast path: if the attributes match the attributes of the resource group at the found index
		if attributesMatch(groups[idx].attrs, attributes) {
			groups[idx].deviceCount++
			return groups
		}

		// Slow path: in case of a hash collision, iterate over the resource groups to find the matching resource group
		for i := range groups {
			if groups[i].signature == hash && groups[i].driver == driver && groups[i].pool == pool {
				if attributesMatch(groups[i].attrs, attributes) {
					groups[i].deviceCount++
					return groups
				}
			}
		}
	}

	// Insert a new resource group with a single device so far
	groups = append(groups, resourceGroup{
		driver:      driver,
		pool:        pool,
		attrs:       attributes,
		signature:   hash,
		deviceCount: 1,
	})
	addresses[identifier] = len(groups) - 1
	return groups
}

// buildResourceGroups extracts resource groups from resource slices and splits them on a basis
// of driver, pool and attribute signature.
func (c *resourcePoolComparator) buildResourceGroups(
	resourceSlices []*v1.ResourceSlice,
	groups []resourceGroup,
	addresses map[resourcePoolIdentifier]int,
) []resourceGroup {
	if len(c.comparedDrivers) == 0 {
		return groups
	}

	for _, rs := range resourceSlices {
		if !slices.Contains(c.comparedDrivers, rs.Spec.Driver) {
			continue
		}

		for _, device := range rs.Spec.Devices {
			groups = c.ingestPoolDevice(rs.Spec.Driver, rs.Spec.Pool.Name, device.Attributes, groups, addresses)
		}
	}

	return groups
}

// attributesMatch checks if two maps of attributes exactly match.
func attributesMatch(attributes, deviceAttributes attributesMap) bool {
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
func computeSetOverlap(a, b attributesMap) int {
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
func computeAttrMapHash(keys attributesMap) uint64 {
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
