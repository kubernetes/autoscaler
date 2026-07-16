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
	"context"

	"github.com/samber/lo"
	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// AllocationTracker is an opaque object used to track the allocation status of individual devices. It can be mutated
// via Commit and ReleaseInstanceTypes, which update allocation status based on NodeClaim constraints applied by the
// scheduler. Allocation status is queried via IsAllocated. Not that the allocation status of a device is not
// independent - it's dependent on the NodeClaim and InstanceType we're attempting to perform the allocation against.
// This is a unique property of Karpenter's NodeClaim model for DRA.
type AllocationTracker struct {
	// PreallocatedDevices represents the devices which are already allocated on the API server. This value is immutable
	// after set during Allocator construction.
	PreallocatedDevices sets.Set[DeviceID]

	// PreallocatedConsumedCapacity holds the aggregated consumed capacity for multi-allocatable devices
	// from existing cluster allocations. Multi-allocatable devices appear here instead of PreallocatedDevices.
	// Immutable after construction. Map: deviceID → dimensionName → consumed quantity.
	PreallocatedConsumedCapacity map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity

	// InflightClusterAllocations contains the allocation metadata by device for a given device ID. Note that entries in
	// this structure are not immutable - as instance types are released by NodeClaims, devices also have the potential to
	// be released and removed from the map.
	InflightClusterAllocations map[DeviceID]*InflightAllocationMetadata

	// InflightClusterAllocationsByNodeClaim is the inverse of inflightClusterAllocations, tracking the allocated devices
	// by NodeClaim and device ID. This is an acceleration data structure used to lookup the impacted devices when an
	// instance type is released for a NodeClaim.
	InflightClusterAllocationsByNodeClaim map[NodeClaimID]map[InstanceTypeID]sets.Set[DeviceID]

	InflightTemplateAllocations map[NodeClaimID]map[InstanceTypeID]sets.Set[DeviceID]

	// InflightConsumedCapacity tracks consumed capacity committed by earlier pods in this scheduling loop
	// for multi-allocatable devices. Updated via commitCapacity/releaseCapacity using the same pessimistic-max
	// pattern as SharedCounters but keyed per-device rather than per-pool.
	// Map: deviceID → dimensionName → consumed quantity.
	InflightConsumedCapacity map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity
	// consumedCapacityByNodeClaimIT stores per-NodeClaim per-IT capacity consumption for precise release.
	// Map: nodeClaimID → instanceTypeID → deviceID → dimensionName → consumed quantity.
	consumedCapacityByNodeClaimIT map[NodeClaimID]map[InstanceTypeID]map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity

	// templateConsumedCapacity tracks consumed capacity for multi-allocatable template devices
	// per (NodeClaim, IT). Each IT has its own independent device set (no pessimistic-max needed).
	// Sparse — only devices with actual cross-pod consumption get entries. Incremented on Commit.
	// Map: nodeClaimID → instanceTypeID → deviceID → dimensionName → consumed quantity.
	templateConsumedCapacity map[NodeClaimID]map[InstanceTypeID]map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity

	// RemainingCounters tracks the remaining counter budgets per pool. Initialized lazily per pool
	// (total - preallocated consumption), then decremented on Commit and incremented on Release.
	// Map: poolKey → counterSetName → counterName → remaining counter.
	RemainingCounters map[PoolKey]map[string]map[string]resourcev1.Counter
	// countersByNodeClaimIT stores per-NodeClaim per-IT counter consumption for precise release
	// when instance types are pruned.
	// Map: nodeClaimID → instanceTypeID → poolKey → counterSetName → counterName → consumed counter.
	countersByNodeClaimIT map[NodeClaimID]map[InstanceTypeID]map[PoolKey]map[string]map[string]resourcev1.Counter

	// templateRemainingCounters tracks the remaining template counter budgets per (NodeClaim, IT, Pool).
	// Initialized lazily with the full SharedCounters budget on first access, then decremented on Commit.
	// Separate from RemainingCounters because template and in-cluster pools can share the same PoolKey,
	// template counters are per-IT (no pessimistic-max), and they don't affect global RemainingCounters.
	// Map: nodeClaimID → instanceTypeID → poolKey → counterSetName → counterName → remaining counter.
	templateRemainingCounters map[NodeClaimID]map[InstanceTypeID]map[PoolKey]map[string]map[string]resourcev1.Counter
}

func NewAllocationTracker(allocatedState AllocatedDeviceState) *AllocationTracker {
	preallocatedDevices := make(sets.Set[DeviceID], len(allocatedState.ExclusiveDevices))
	for id := range allocatedState.ExclusiveDevices {
		preallocatedDevices.Insert(DeviceID{
			DeviceID: id,
		})
	}
	preallocatedCapacity := make(map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity, len(allocatedState.ConsumedCapacity))
	for id, capacity := range allocatedState.ConsumedCapacity {
		preallocatedCapacity[DeviceID{DeviceID: id}] = capacity
	}
	return &AllocationTracker{
		PreallocatedDevices:                   preallocatedDevices,
		PreallocatedConsumedCapacity:          preallocatedCapacity,
		InflightClusterAllocations:            make(map[DeviceID]*InflightAllocationMetadata),
		InflightClusterAllocationsByNodeClaim: make(map[NodeClaimID]map[InstanceTypeID]sets.Set[DeviceID]),
		InflightTemplateAllocations:           make(map[NodeClaimID]map[InstanceTypeID]sets.Set[DeviceID]),
		InflightConsumedCapacity:              make(map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity),
		consumedCapacityByNodeClaimIT:         make(map[NodeClaimID]map[InstanceTypeID]map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity),
		RemainingCounters:                     make(map[PoolKey]map[string]map[string]resourcev1.Counter),
		countersByNodeClaimIT:                 make(map[NodeClaimID]map[InstanceTypeID]map[PoolKey]map[string]map[string]resourcev1.Counter),
		templateRemainingCounters:             make(map[NodeClaimID]map[InstanceTypeID]map[PoolKey]map[string]map[string]resourcev1.Counter),
		templateConsumedCapacity:              make(map[NodeClaimID]map[InstanceTypeID]map[DeviceID]map[resourcev1.QualifiedName]resource.Quantity),
	}
}

// InflightAllocationMetadata constains the nodeClaim that a device was allocated for and the set of instance types for
// that nodeclaim. This set of instance types may be a subset of the nodeclaim's total instance types - if so the
// allocation is released when all instance types are released.
type InflightAllocationMetadata struct {
	// NodeClaimID represents the NodeClaim that the ResourceClaim is indirectly bound to (through the pod). This device
	// may have only been allocated to satisfy the claim for a subset of instance types on the NodeClaim.
	NodeClaimID NodeClaimID
	// InstanceTypes represents the instance types for the NodeClaim that allocated this device to satisfy a ResourceClaim
	InstanceTypes sets.Set[InstanceTypeID]
}

// nolint:gocyclo
func (at *AllocationTracker) Commit(alloc *allocation) {
	for it, deviceIDs := range alloc.deviceIDsByIT {
		for _, id := range deviceIDs {
			if id.Template {
				// Template multi-alloc devices are tracked via templateConsumedCapacity, not
				// binary allocation.
				if itCapacity, ok := alloc.templateCapacityConsumptionByIT[it]; ok {
					if _, hasCapacity := itCapacity[id]; hasCapacity {
						continue
					}
				}
				at.insertAllocation(at.InflightTemplateAllocations, id, alloc.nodeClaimID, it)
				continue
			}
			// Multi-allocatable devices are tracked via capacity, not binary allocation.
			if itCapacity, ok := alloc.capacityConsumptionByIT[it]; ok {
				if _, hasCapacity := itCapacity[id]; hasCapacity {
					continue
				}
			}
			at.insertAllocation(at.InflightClusterAllocationsByNodeClaim, id, alloc.nodeClaimID, it)

			meta, ok := at.InflightClusterAllocations[id]
			if ok {
				if meta.NodeClaimID != alloc.nodeClaimID {
					panic("device is already allocated for a different nodeclaim")
				}
				if meta.InstanceTypes.Has(it) {
					panic("device is already allocated for instance type")
				}
				meta.InstanceTypes.Insert(it)
			} else {
				meta = &InflightAllocationMetadata{
					NodeClaimID:   alloc.nodeClaimID,
					InstanceTypes: make(sets.Set[InstanceTypeID]),
				}
				meta.InstanceTypes.Insert(it)
				at.InflightClusterAllocations[id] = meta
			}
		}
	}
	at.commitCounters(alloc.nodeClaimID, alloc.counterConsumptionByIT)
	for itID, totals := range alloc.templateCounterTotalsByIT {
		at.InitTemplateRemainingCounters(alloc.nodeClaimID, itID, totals)
	}
	at.commitTemplateCounters(alloc.nodeClaimID, alloc.templateCounterConsumptionByIT)
	at.commitCapacity(alloc.nodeClaimID, alloc.capacityConsumptionByIT)
	at.commitTemplateCapacity(alloc.nodeClaimID, alloc.templateCapacityConsumptionByIT)
}

func (at *AllocationTracker) insertAllocation(
	allocationMap map[NodeClaimID]map[InstanceTypeID]sets.Set[DeviceID],
	deviceID DeviceID,
	nodeClaimID NodeClaimID,
	instanceTypeID InstanceTypeID,
) {
	nodeClaimAllocs, ok := allocationMap[nodeClaimID]
	if !ok {
		nodeClaimAllocs = make(map[InstanceTypeID]sets.Set[DeviceID])
		allocationMap[nodeClaimID] = nodeClaimAllocs
	}
	itAllocs, ok := nodeClaimAllocs[instanceTypeID]
	if !ok {
		itAllocs = make(sets.Set[DeviceID])
		nodeClaimAllocs[instanceTypeID] = itAllocs
	}
	if itAllocs.Has(deviceID) {
		panic("device is already allocated for instance type")
	}
	itAllocs.Insert(deviceID)
}

func (at *AllocationTracker) ReleaseInstanceTypes(ctx context.Context, nodeClaim NodeClaimID, instanceTypes ...InstanceTypeID) {
	released := make(map[InstanceTypeID]sets.Set[DeviceID])
	for _, instanceType := range instanceTypes {
		devices := at.InflightClusterAllocationsByNodeClaim[nodeClaim][instanceType]
		released[instanceType] = devices
		delete(at.InflightClusterAllocationsByNodeClaim[nodeClaim], instanceType)
		for id := range devices {
			meta, ok := at.InflightClusterAllocations[id]
			if !ok {
				panic("missing reference count for device ID")
			}
			if !meta.InstanceTypes.Has(instanceType) {
				panic("inflight allocation metadata for device is missing instance type reference")
			}
			meta.InstanceTypes.Delete(instanceType)
			if len(meta.InstanceTypes) == 0 {
				delete(at.InflightClusterAllocations, id)
			}
		}
		delete(at.InflightTemplateAllocations[nodeClaim], instanceType)
	}
	at.releaseCounters(nodeClaim, instanceTypes)
	at.releaseTemplateCounters(nodeClaim, instanceTypes)
	at.releaseCapacity(nodeClaim, instanceTypes)
	at.releaseTemplateCapacity(nodeClaim, instanceTypes)

	if len(released) != 0 && log.FromContext(ctx).V(1).Enabled() {
		log.FromContext(ctx).V(1).Info("releasing allocations", "nodeClaimID", nodeClaim.Value(), "devicesByInstanceType", lo.MapEntries(released, func(it InstanceTypeID, ids sets.Set[DeviceID]) (string, []string) {
			return it.Value(), lo.Map(ids.UnsortedList(), func(id DeviceID, _ int) string { return id.String() })
		}))
	}
}

func (at *AllocationTracker) IsAllocated(deviceID DeviceID, nodeClaim NodeClaim, instanceType InstanceTypeID) bool {
	if deviceID.Template {
		// Template devices are NodeClaim and instance type local. The device is only considered allocated if there's an entry
		// for the given NodeClaim-InstanceType combination.
		nodeClaimAllocs, ok := at.InflightTemplateAllocations[nodeClaim.ID()]
		if !ok {
			return false
		}
		instanceTypeAllocs, ok := nodeClaimAllocs[instanceType]
		if !ok {
			return false
		}
		if instanceTypeAllocs.Has(deviceID) {
			return true
		}
		return false
	}

	// The device is already marked as allocated on the cluster
	if at.PreallocatedDevices.Has(deviceID) {
		return true
	}
	if meta, ok := at.InflightClusterAllocations[deviceID]; ok {
		// If the device is marked as allocating for a different NodeClaim, we pessimistically assume it will be allocated
		if meta.NodeClaimID != nodeClaim.ID() {
			return true
		}
		// The device is marked as allocating for this NodeClaim and instance type combination so it can't be allocated again
		if meta.InstanceTypes.Has(instanceType) {
			return true
		}
		// The device is marked as allocating for this NodeClaim, but only for different instance types. Since the NodeClaim
		// will collapse to a single instance type, we can allocate it once per instance type.
		return false
	}
	// The device is neither marked as allocating in the cluster nor in the allocator's state, it's not allocated
	return false
}
