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

package dynamicresources

import (
	"fmt"

	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"sigs.k8s.io/karpenter/pkg/cloudprovider"
)

// checkCapacity verifies that a multi-allocatable device has sufficient remaining capacity
// for the given request. Returns the computed consumed capacity map and a pass/fail bool.
// For non-multi-allocatable devices, returns (nil, true) immediately.
func (a *allocator) checkCapacity(device cloudprovider.Device, deviceID DeviceID, rd *RequestData) (map[resourcev1.QualifiedName]resource.Quantity, bool) {
	if !device.AllowMultipleAllocations {
		return nil, true
	}
	if requestsContainNonExistCapacity(rd.CapacityRequests, device.Capacity) {
		return nil, false
	}
	consumed, err := computeConsumedCapacity(rd.CapacityRequests, device.Capacity)
	if err != nil {
		return nil, false
	}
	if consumed == nil {
		return nil, true
	}
	var sources []map[resourcev1.QualifiedName]resource.Quantity
	if deviceID.Template {
		if tc := a.allocationTracker.TemplateConsumedCapacityForIT(a.nodeClaim.ID(), a.itID); tc != nil {
			sources = append(sources, tc[deviceID])
		}
		sources = append(sources, a.templateAllocatingCapacity[deviceID])
	} else {
		sources = append(sources,
			a.allocationTracker.PreallocatedConsumedCapacity[deviceID],
			a.allocationTracker.InflightConsumedCapacity[deviceID],
			a.allocatingCapacity[deviceID],
		)
	}
	for name, qty := range consumed {
		total := device.Capacity[name].Value
		var used resource.Quantity
		for _, src := range sources {
			if q, ok := src[name]; ok {
				used.Add(q)
			}
		}
		used.Add(qty)
		if used.Cmp(total) > 0 {
			return nil, false
		}
	}
	return consumed, true
}

// deductAllocatingCapacity adds consumed capacity to the DFS-local allocating state.
func (a *allocator) deductAllocatingCapacity(consumed map[resourcev1.QualifiedName]resource.Quantity, deviceID DeviceID, template bool) {
	if len(consumed) == 0 {
		return
	}
	if template {
		a.templateAllocatingCapacity[deviceID] = addCapacity(a.templateAllocatingCapacity[deviceID], consumed)
	} else {
		a.allocatingCapacity[deviceID] = addCapacity(a.allocatingCapacity[deviceID], consumed)
	}
}

// restoreAllocatingCapacity reverses consumed capacity from the DFS-local allocating state.
func (a *allocator) restoreAllocatingCapacity(consumed map[resourcev1.QualifiedName]resource.Quantity, deviceID DeviceID, template bool) {
	if len(consumed) == 0 {
		return
	}
	if template {
		a.templateAllocatingCapacity[deviceID] = subCapacity(a.templateAllocatingCapacity[deviceID], consumed)
	} else {
		a.allocatingCapacity[deviceID] = subCapacity(a.allocatingCapacity[deviceID], consumed)
	}
}

// commitCapacity stores per-IT capacity consumption and increments InflightConsumedCapacity by
// the delta between the new pessimistic max and the old one.
//
//nolint:gocyclo
func (at *AllocationTracker) commitCapacity(nodeClaimID NodeClaimID, newCapacityByIT map[InstanceTypeID]map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity) {
	if len(newCapacityByIT) == 0 {
		return
	}
	storedConsumedCapacityByIT, ok := at.consumedCapacityByNodeClaimIT[nodeClaimID]
	if !ok {
		storedConsumedCapacityByIT = make(map[InstanceTypeID]map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity)
		at.consumedCapacityByNodeClaimIT[nodeClaimID] = storedConsumedCapacityByIT
	}

	oldMax := pessimisticCapacityMax(storedConsumedCapacityByIT)

	// Merge new consumption into stored state.
	for it, deviceCapacity := range newCapacityByIT {
		storedDevices, ok := storedConsumedCapacityByIT[it]
		if !ok {
			storedConsumedCapacityByIT[it] = deviceCapacity
			continue
		}
		for deviceID, consumedByDevice := range deviceCapacity {
			storedConsumedByDevice, ok := storedDevices[deviceID]
			if !ok {
				storedDevices[deviceID] = consumedByDevice
				continue
			}
			for name, qty := range consumedByDevice {
				existing := storedConsumedByDevice[name]
				existing.Add(qty)
				storedConsumedByDevice[name] = existing
			}
		}
	}

	newMax := pessimisticCapacityMax(storedConsumedCapacityByIT)

	// Add the delta (newMax - oldMax) to InflightConsumedCapacity.
	for deviceID, newConsumedByDevice := range newMax {
		for name, newQty := range newConsumedByDevice {
			delta := newQty.DeepCopy()
			if oldConsumedByDevice, ok := oldMax[deviceID]; ok {
				if oldQty, ok := oldConsumedByDevice[name]; ok {
					delta.Sub(oldQty)
				}
			}
			if delta.Sign() > 0 {
				if at.InflightConsumedCapacity[deviceID] == nil {
					at.InflightConsumedCapacity[deviceID] = make(map[resourcev1.QualifiedName]resource.Quantity)
				}
				existing := at.InflightConsumedCapacity[deviceID][name]
				existing.Add(delta)
				at.InflightConsumedCapacity[deviceID][name] = existing
			}
		}
	}
}

// commitTemplateCapacity adds per-IT template capacity consumption to the tracker's consumed state.
// Template capacity doesn't need pessimistic-max treatment — each IT has its own independent device set.
func (at *AllocationTracker) commitTemplateCapacity(nodeClaimID NodeClaimID, consumptionByIT map[InstanceTypeID]map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity) {
	if len(consumptionByIT) == 0 {
		return
	}
	consumedByIT, ok := at.templateConsumedCapacity[nodeClaimID]
	if !ok {
		consumedByIT = make(map[InstanceTypeID]map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity)
		at.templateConsumedCapacity[nodeClaimID] = consumedByIT
	}
	for itID, devices := range consumptionByIT {
		consumedDevices, ok := consumedByIT[itID]
		if !ok {
			consumedByIT[itID] = devices
			continue
		}
		for deviceID, consumedByDevice := range devices {
			storedConsumedByDevice, ok := consumedDevices[deviceID]
			if !ok {
				consumedDevices[deviceID] = consumedByDevice
				continue
			}
			for name, qty := range consumedByDevice {
				existing := storedConsumedByDevice[name]
				existing.Add(qty)
				storedConsumedByDevice[name] = existing
			}
		}
	}
}

// releaseCapacity adjusts InflightConsumedCapacity when instance types are pruned. Recomputes
// the pessimistic max from remaining ITs and subtracts the delta.
//
//nolint:gocyclo
func (at *AllocationTracker) releaseCapacity(nodeClaimID NodeClaimID, releasedITs []InstanceTypeID) {
	storedConsumedCapacityByIT, ok := at.consumedCapacityByNodeClaimIT[nodeClaimID]
	if !ok {
		return
	}

	oldMax := pessimisticCapacityMax(storedConsumedCapacityByIT)

	for _, itID := range releasedITs {
		delete(storedConsumedCapacityByIT, itID)
	}

	newMax := pessimisticCapacityMax(storedConsumedCapacityByIT)

	// Subtract the delta (oldMax - newMax) from InflightConsumedCapacity.
	for deviceID, oldConsumedByDevice := range oldMax {
		for name, oldQty := range oldConsumedByDevice {
			delta := oldQty.DeepCopy()
			if newConsumedByDevice, ok := newMax[deviceID]; ok {
				if newQty, ok := newConsumedByDevice[name]; ok {
					delta.Sub(newQty)
				}
			}
			if delta.Sign() > 0 {
				inflight := at.InflightConsumedCapacity[deviceID]
				if inflight == nil {
					continue
				}
				existing := inflight[name]
				existing.Sub(delta)
				inflight[name] = existing
				if existing.Sign() <= 0 {
					delete(inflight, name)
				}
				if len(inflight) == 0 {
					delete(at.InflightConsumedCapacity, deviceID)
				}
			}
		}
	}

	if len(storedConsumedCapacityByIT) == 0 {
		delete(at.consumedCapacityByNodeClaimIT, nodeClaimID)
	}
}

// releaseTemplateCapacity removes template capacity state for pruned instance types.
func (at *AllocationTracker) releaseTemplateCapacity(nodeClaimID NodeClaimID, releasedITs []InstanceTypeID) {
	consumedByIT, ok := at.templateConsumedCapacity[nodeClaimID]
	if !ok {
		return
	}
	for _, itID := range releasedITs {
		delete(consumedByIT, itID)
	}
	if len(consumedByIT) == 0 {
		delete(at.templateConsumedCapacity, nodeClaimID)
	}
}

// TemplateConsumedCapacityForIT returns the template consumed capacity for the given
// (NodeClaim, IT) pair. Returns nil if not yet initialized.
func (at *AllocationTracker) TemplateConsumedCapacityForIT(nodeClaimID NodeClaimID, itID InstanceTypeID) map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity {
	consumedByIT, ok := at.templateConsumedCapacity[nodeClaimID]
	if !ok {
		return nil
	}
	return consumedByIT[itID]
}

// pessimisticCapacityMax computes the maximum consumed capacity per device per dimension across all ITs.
func pessimisticCapacityMax(capacityByIT map[InstanceTypeID]map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity) map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity {
	if len(capacityByIT) == 0 {
		return nil
	}
	maxByDevice := make(map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity)
	for _, devices := range capacityByIT {
		for deviceID, consumedByDevice := range devices {
			maxConsumedByDevice, ok := maxByDevice[deviceID]
			if !ok {
				maxConsumedByDevice = make(map[resourcev1.QualifiedName]resource.Quantity, len(consumedByDevice))
				maxByDevice[deviceID] = maxConsumedByDevice
			}
			for name, qty := range consumedByDevice {
				if existing, ok := maxConsumedByDevice[name]; !ok || qty.Cmp(existing) > 0 {
					maxConsumedByDevice[name] = qty.DeepCopy()
				}
			}
		}
	}
	return maxByDevice
}

// computeConsumedCapacity computes the consumed capacity for all dimensions on a device given
// the request's capacity requirements. Returns nil if the device has no capacity dimensions.
// Returns an error if a requested dimension doesn't exist on the device or violates policy.
func computeConsumedCapacity(
	capacityRequests map[resourcev1.QualifiedName]resource.Quantity,
	deviceCapacity map[resourcev1.QualifiedName]resourcev1.DeviceCapacity,
) (map[resourcev1.QualifiedName]resource.Quantity, error) {
	if len(deviceCapacity) == 0 {
		return nil, nil
	}
	consumed := make(map[resourcev1.QualifiedName]resource.Quantity, len(deviceCapacity))
	for name, cap := range deviceCapacity {
		var requestedVal *resource.Quantity
		if capacityRequests != nil {
			if rv, ok := capacityRequests[name]; ok {
				requestedVal = &rv
			}
		}
		c := calculateConsumedCapacity(requestedVal, cap)
		if violatesPolicy(c, cap.RequestPolicy) {
			return nil, fmt.Errorf("capacity request violates policy for dimension %s", name)
		}
		consumed[name] = c
	}
	return consumed, nil
}

// addCapacity adds the quantities from src into dst, initializing dst if nil.
func addCapacity(dst, src map[resourcev1.QualifiedName]resource.Quantity) map[resourcev1.QualifiedName]resource.Quantity {
	if len(src) == 0 {
		return dst
	}
	if dst == nil {
		dst = make(map[resourcev1.QualifiedName]resource.Quantity, len(src))
	}
	for name, qty := range src {
		existing := dst[name]
		existing.Add(qty)
		dst[name] = existing
	}
	return dst
}

// subCapacity subtracts the quantities in src from dst.
func subCapacity(dst, src map[resourcev1.QualifiedName]resource.Quantity) map[resourcev1.QualifiedName]resource.Quantity {
	if len(src) == 0 {
		return dst
	}
	for name, qty := range src {
		existing := dst[name]
		existing.Sub(qty)
		dst[name] = existing
	}
	return dst
}

// requestsContainNonExistCapacity returns true if the request references capacity dimensions
// that don't exist on the device.
// Note: equivalent to upstream requestsContainNonExistCapacity in k8s.io/dynamic-resource-allocation
func requestsContainNonExistCapacity(
	capacityRequests map[resourcev1.QualifiedName]resource.Quantity,
	deviceCapacity map[resourcev1.QualifiedName]resourcev1.DeviceCapacity,
) bool {
	for name := range capacityRequests {
		if _, ok := deviceCapacity[name]; !ok {
			return true
		}
	}
	return false
}

// calculateConsumedCapacity returns valid capacity to be consumed regarding the requested capacity and device capacity policy.
// If no requestPolicy, return capacity.Value. If no requestVal, fill the quantity by fillEmptyRequest function
// Otherwise, use requestPolicy to calculate the consumed capacity from request if applicable.
// Note: equivalent to upstream calculateConsumedCapacity in k8s.io/dynamic-resource-allocation
func calculateConsumedCapacity(requestedVal *resource.Quantity, capacity resourcev1.DeviceCapacity) resource.Quantity {
	if requestedVal == nil {
		return fillEmptyRequest(capacity)
	}
	if capacity.RequestPolicy == nil {
		return requestedVal.DeepCopy()
	}
	switch {
	case capacity.RequestPolicy.ValidRange != nil && capacity.RequestPolicy.ValidRange.Min != nil:
		return roundUpRange(requestedVal, capacity.RequestPolicy.ValidRange)
	case capacity.RequestPolicy.ValidValues != nil:
		return roundUpValidValues(requestedVal, capacity.RequestPolicy.ValidValues)
	}
	return requestedVal.DeepCopy()
}

// fillEmptyRequest returns RequestPolicy.Default if defined, otherwise the full device capacity.
// Note: equivalent to upstream fillEmptyRequest in k8s.io/dynamic-resource-allocation
func fillEmptyRequest(capacity resourcev1.DeviceCapacity) resource.Quantity {
	if capacity.RequestPolicy != nil && capacity.RequestPolicy.Default != nil {
		return capacity.RequestPolicy.Default.DeepCopy()
	}
	return capacity.Value.DeepCopy()
}

// roundUpRange rounds requestedVal up to fit within the specified validRange.
// If requestedVal < Min, returns Min.
// If Step is specified, rounds up to the nearest Min + N*Step.
// If no Step is specified and requestedVal >= Min, it returns requestedVal as is.
// Note: equivalent to upstream roundUpRange in k8s.io/dynamic-resource-allocation
func roundUpRange(requestedVal *resource.Quantity, validRange *resourcev1.CapacityRequestPolicyRange) resource.Quantity {
	if requestedVal.Cmp(*validRange.Min) < 0 {
		return validRange.Min.DeepCopy()
	}
	if validRange.Step == nil {
		return requestedVal.DeepCopy()
	}
	requestedInt64 := requestedVal.Value()
	step := validRange.Step.Value()
	min := validRange.Min.Value()
	added := requestedInt64 - min
	n := added / step
	if added%step != 0 {
		n++
	}
	return *resource.NewQuantity(min+step*n, resource.BinarySI)
}

// roundUpValidValues returns the first valid value >= requestedVal. If none exists, returns requestedVal.
// Note: equivalent to upstream roundUpValidValues in k8s.io/dynamic-resource-allocation
func roundUpValidValues(requestedVal *resource.Quantity, validValues []resource.Quantity) resource.Quantity {
	// validValues must be sorted ascending (enforced by API validation).
	for _, validValue := range validValues {
		if requestedVal.Cmp(validValue) <= 0 {
			return validValue.DeepCopy()
		}
	}
	return requestedVal.DeepCopy()
}

// violatesPolicy checks whether a consumed value violates the device's request policy.
// Note: equivalent to upstream violatesPolicy in k8s.io/dynamic-resource-allocation
func violatesPolicy(consumedVal resource.Quantity, policy *resourcev1.CapacityRequestPolicy) bool {
	if policy == nil {
		return false
	}
	if policy.Default != nil && consumedVal.Cmp(*policy.Default) == 0 {
		return false
	}
	switch {
	case policy.ValidRange != nil:
		return violateValidRange(consumedVal, *policy.ValidRange)
	case len(policy.ValidValues) > 0:
		return violateValidValues(consumedVal, policy.ValidValues)
	}
	return false
}

// Note: equivalent to upstream violateValidRange in k8s.io/dynamic-resource-allocation
func violateValidRange(val resource.Quantity, validRange resourcev1.CapacityRequestPolicyRange) bool {
	if validRange.Max != nil && val.Cmp(*validRange.Max) > 0 {
		return true
	}
	if validRange.Step != nil {
		requestedInt64 := val.Value()
		step := validRange.Step.Value()
		min := validRange.Min.Value()
		if (requestedInt64-min)%step != 0 {
			return true
		}
	}
	return false
}

// Note: equivalent to upstream violateValidValues in k8s.io/dynamic-resource-allocation
func violateValidValues(val resource.Quantity, validValues []resource.Quantity) bool {
	for i := range validValues {
		if val.Cmp(validValues[i]) == 0 {
			return false
		}
	}
	return true
}
