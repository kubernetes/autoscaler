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
	"github.com/samber/lo"
	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"sigs.k8s.io/karpenter/pkg/cloudprovider"
)

// commitCounters stores per-IT counter consumption and decrements remaining counters by the
// delta between the new accumulated pessimistic max and the old one.
func (at *AllocationTracker) commitCounters(nodeClaimID NodeClaimID, newCounterConsumptionByIT map[InstanceTypeID]map[PoolKey]map[string]map[string]resourcev1.Counter) {
	if len(newCounterConsumptionByIT) == 0 {
		return
	}
	storedCounterSetsByIT, ok := at.countersByNodeClaimIT[nodeClaimID]
	if !ok {
		storedCounterSetsByIT = make(map[InstanceTypeID]map[PoolKey]map[string]map[string]resourcev1.Counter)
		at.countersByNodeClaimIT[nodeClaimID] = storedCounterSetsByIT
	}

	// Compute old pessimistic max before merging new consumption.
	var oldCounterMax map[PoolKey]map[string]map[string]resourcev1.Counter
	if len(storedCounterSetsByIT) > 0 {
		oldCounterMax = pessimisticCounterMax(storedCounterSetsByIT)
	}

	// Merge new consumption into stored state.
	for it, counterSetsByPool := range newCounterConsumptionByIT {
		storedCounterSetsByPool, ok := storedCounterSetsByIT[it]
		if !ok {
			storedCounterSetsByIT[it] = counterSetsByPool
			continue
		}

		for poolKey, counterSets := range counterSetsByPool {
			storedCounterSets, ok := storedCounterSetsByPool[poolKey]
			if !ok {
				storedCounterSetsByPool[poolKey] = counterSets
				continue
			}
			for counterSetName, counters := range counterSets {
				storedCounterSet, ok := storedCounterSets[counterSetName]
				if !ok {
					storedCounterSets[counterSetName] = counters
					continue
				}
				for counterName, counter := range counters {
					storedCounter := storedCounterSet[counterName]
					storedCounter.Value.Add(counter.Value)
					storedCounterSet[counterName] = storedCounter
				}
			}
		}
	}

	// Compute new pessimistic max after merging.
	newCounterMax := pessimisticCounterMax(storedCounterSetsByIT)

	// Subtract only the delta (newMax - oldMax) from remaining counters.
	subtractDeltaFromRemaining(at.RemainingCounters, oldCounterMax, newCounterMax)
}

// commitTemplateCounters subtracts per-IT template counter consumption directly from the
// pre-initialized remaining budgets. Template counters don't need pessimistic-max treatment —
// each IT has its own independent budget.
func (at *AllocationTracker) commitTemplateCounters(nodeClaimID NodeClaimID, consumptionByIT map[InstanceTypeID]map[PoolKey]map[string]map[string]resourcev1.Counter) {
	if len(consumptionByIT) == 0 {
		return
	}
	remainingCounterSetsByIT := at.templateRemainingCounters[nodeClaimID]
	if remainingCounterSetsByIT == nil {
		return
	}
	for itID, counterSetsByPool := range consumptionByIT {
		remainingCounterSetsByPool := remainingCounterSetsByIT[itID]
		if remainingCounterSetsByPool == nil {
			continue
		}
		for poolKey, counterSets := range counterSetsByPool {
			remainingCounterSets := remainingCounterSetsByPool[poolKey]
			for counterSetName, counters := range counterSets {
				remainingCounterSet := remainingCounterSets[counterSetName]
				for counterName, counter := range counters {
					remainingCounter := remainingCounterSet[counterName]
					remainingCounter.Value.Sub(counter.Value)
					remainingCounterSet[counterName] = remainingCounter
				}
			}
		}
	}
}

// subtractDeltaFromRemaining subtracts (newCounterMax - oldCounterMax) from remaining counters.
func subtractDeltaFromRemaining(remaining map[PoolKey]map[string]map[string]resourcev1.Counter, oldCounterMax, newCounterMax map[PoolKey]map[string]map[string]resourcev1.Counter) {
	for poolKey, newCounterSets := range newCounterMax {
		poolRemaining, ok := remaining[poolKey]
		if !ok {
			continue
		}
		for counterSetName, newCounters := range newCounterSets {
			counterSetRemaining, ok := poolRemaining[counterSetName]
			if !ok {
				continue
			}
			for counterName, newCounter := range newCounters {
				delta := newCounter.Value.DeepCopy()
				if old, ok := getCounter(oldCounterMax, poolKey, counterSetName, counterName); ok {
					delta.Sub(old.Value)
				}
				if delta.Sign() > 0 {
					remainingCounter, ok := counterSetRemaining[counterName]
					if !ok {
						continue
					}
					remainingCounter.Value.Sub(delta)
					counterSetRemaining[counterName] = remainingCounter
				}
			}
		}
	}
}

// releaseCounters adjusts remaining counters when instance types are pruned. Recomputes the
// pessimistic max from remaining ITs and adds back the delta.
func (at *AllocationTracker) releaseCounters(nodeClaimID NodeClaimID, releasedITs []InstanceTypeID) {
	storedCounterSetsByIT, ok := at.countersByNodeClaimIT[nodeClaimID]
	if !ok {
		return
	}

	oldCounterMax := pessimisticCounterMax(storedCounterSetsByIT)

	for _, itID := range releasedITs {
		delete(storedCounterSetsByIT, itID)
	}

	var newCounterMax map[PoolKey]map[string]map[string]resourcev1.Counter
	if len(storedCounterSetsByIT) > 0 {
		newCounterMax = pessimisticCounterMax(storedCounterSetsByIT)
	}
	addDeltaToRemaining(at.RemainingCounters, oldCounterMax, newCounterMax)

	if len(storedCounterSetsByIT) == 0 {
		delete(at.countersByNodeClaimIT, nodeClaimID)
	}
}

// releaseTemplateCounters removes template counter state for pruned instance types.
func (at *AllocationTracker) releaseTemplateCounters(nodeClaimID NodeClaimID, releasedITs []InstanceTypeID) {
	remainingCounterSetsByIT, ok := at.templateRemainingCounters[nodeClaimID]
	if !ok {
		return
	}
	for _, itID := range releasedITs {
		delete(remainingCounterSetsByIT, itID)
	}
	if len(remainingCounterSetsByIT) == 0 {
		delete(at.templateRemainingCounters, nodeClaimID)
	}
}

// addDeltaToRemaining adds (oldCounterMax - newCounterMax) back to remaining counters.
func addDeltaToRemaining(remaining map[PoolKey]map[string]map[string]resourcev1.Counter, oldCounterMax, newCounterMax map[PoolKey]map[string]map[string]resourcev1.Counter) {
	for poolKey, oldCounterSets := range oldCounterMax {
		poolRemaining, ok := remaining[poolKey]
		if !ok {
			continue
		}
		for counterSetName, oldCounters := range oldCounterSets {
			counterSetRemaining, ok := poolRemaining[counterSetName]
			if !ok {
				continue
			}
			for counterName, oldCounter := range oldCounters {
				delta := oldCounter.Value.DeepCopy()
				if new, ok := getCounter(newCounterMax, poolKey, counterSetName, counterName); ok {
					delta.Sub(new.Value)
				}
				if delta.Sign() > 0 {
					remainingCounter := counterSetRemaining[counterName]
					remainingCounter.Value.Add(delta)
					counterSetRemaining[counterName] = remainingCounter
				}
			}
		}
	}
}

// InitRemainingCounters initializes the remaining counter budget for a pool. Called lazily on first
// access for each pool during allocation. The initial value is the pool's total counter budget
// minus consumption from preallocated devices.
func (at *AllocationTracker) InitRemainingCounters(pool *Pool) {
	if _, ok := at.RemainingCounters[pool.Key]; ok {
		return
	}
	if len(pool.CounterSets) == 0 {
		return
	}
	remainingCounterSets := make(map[string]map[string]resourcev1.Counter, len(pool.CounterSets))
	for counterSetName, counters := range pool.CounterSets {
		remainingCounterSets[counterSetName] = make(map[string]resourcev1.Counter, len(counters))
		for counterName, counter := range counters {
			remainingCounterSets[counterSetName][counterName] = resourcev1.Counter{Value: counter.Value.DeepCopy()}
		}
	}
	// Deduct consumption from preallocated devices.
	for i := range pool.Devices {
		if !at.PreallocatedDevices.Has(pool.Devices[i].ID) &&
			!lo.HasKey(at.PreallocatedConsumedCapacity, pool.Devices[i].ID) {
			continue
		}
		deductFromCounters(remainingCounterSets, pool.Devices[i].Device)
	}
	for i := range pool.NonTargetingDevices {
		if !at.PreallocatedDevices.Has(pool.NonTargetingDevices[i].ID) &&
			!lo.HasKey(at.PreallocatedConsumedCapacity, pool.NonTargetingDevices[i].ID) {
			continue
		}
		deductFromCounters(remainingCounterSets, pool.NonTargetingDevices[i].Device)
	}
	at.RemainingCounters[pool.Key] = remainingCounterSets
}

// deductFromCounters subtracts a device's counter consumption from counter budgets.
func deductFromCounters(remainingCounterSets map[string]map[string]resourcev1.Counter, device cloudprovider.Device) {
	for _, consumption := range device.ConsumesCounters {
		counterSetRemaining, ok := remainingCounterSets[consumption.CounterSet]
		if !ok {
			continue
		}
		for counterName, counter := range consumption.Counters {
			remainingCounter, ok := counterSetRemaining[counterName]
			if !ok {
				continue
			}
			remainingCounter.Value.Sub(counter.Value)
			counterSetRemaining[counterName] = remainingCounter
		}
	}
}

// InitTemplateRemainingCounters lazily initializes the remaining counter budget for a
// (NodeClaim, IT) pair. The caller provides the total budget (computed from SharedCounters on
// the template slices). Subsequent calls for the same (NC, IT) are no-ops.
func (at *AllocationTracker) InitTemplateRemainingCounters(
	nodeClaimID NodeClaimID,
	itID InstanceTypeID,
	totals map[PoolKey]map[string]map[string]resourcev1.Counter,
) {
	remainingCounterSetsByIT, ok := at.templateRemainingCounters[nodeClaimID]
	if !ok {
		remainingCounterSetsByIT = make(map[InstanceTypeID]map[PoolKey]map[string]map[string]resourcev1.Counter)
		at.templateRemainingCounters[nodeClaimID] = remainingCounterSetsByIT
	}
	if _, ok := remainingCounterSetsByIT[itID]; ok {
		return
	}
	remainingCounterSetsByIT[itID] = totals
}

// TemplateRemainingForIT returns the remaining template counter budget for the given
// (NodeClaim, IT) pair. Returns nil if not yet initialized.
func (at *AllocationTracker) TemplateRemainingForIT(nodeClaimID NodeClaimID, itID InstanceTypeID) map[PoolKey]map[string]map[string]resourcev1.Counter {
	remainingCounterSetsByIT, ok := at.templateRemainingCounters[nodeClaimID]
	if !ok {
		return nil
	}
	return remainingCounterSetsByIT[itID]
}

// pessimisticCounterMax computes the maximum counter value per pool/counterSet/counter across all ITs.
// returns map: poolKey → counterSetName → counterName → remaining counter.
func pessimisticCounterMax(counterConsumptionByIT map[InstanceTypeID]map[PoolKey]map[string]map[string]resourcev1.Counter) map[PoolKey]map[string]map[string]resourcev1.Counter {
	counterMaxByPool := make(map[PoolKey]map[string]map[string]resourcev1.Counter)
	for _, counterSetsByPool := range counterConsumptionByIT {
		for poolKey, counterSets := range counterSetsByPool {
			maxCounterSets, ok := counterMaxByPool[poolKey]
			if !ok {
				maxCounterSets = make(map[string]map[string]resourcev1.Counter)
				counterMaxByPool[poolKey] = maxCounterSets
			}
			for counterSetName, counters := range counterSets {
				maxCounters, ok := maxCounterSets[counterSetName]
				if !ok {
					maxCounters = make(map[string]resourcev1.Counter)
					maxCounterSets[counterSetName] = maxCounters
				}
				for counterName, counter := range counters {
					maxCounter, ok := maxCounters[counterName]
					if !ok || counter.Value.Cmp(maxCounter.Value) > 0 {
						maxCounters[counterName] = resourcev1.Counter{Value: counter.Value.DeepCopy()}
					}
				}
			}
		}
	}
	return counterMaxByPool
}

func getCounter(m map[PoolKey]map[string]map[string]resourcev1.Counter, pool PoolKey, set, name string) (resourcev1.Counter, bool) {
	if m == nil {
		return resourcev1.Counter{}, false
	}
	sets, ok := m[pool]
	if !ok {
		return resourcev1.Counter{}, false
	}
	counters, ok := sets[set]
	if !ok {
		return resourcev1.Counter{}, false
	}
	c, ok := counters[name]
	return c, ok
}

// poolCountersExhausted returns true if any counter in the pool has been fully consumed
// by DFS-local tentative allocations, meaning no additional counter-consuming device from
// this pool can succeed.
func (a *allocator) poolCountersExhausted(pool *Pool) bool {
	if len(pool.CounterSets) == 0 {
		return false
	}
	remaining := a.allocationTracker.RemainingCounters[pool.Key]
	if remaining == nil {
		return false
	}
	allocating := a.allocatingCounters[pool.Key]
	if allocating == nil {
		return false
	}
	for counterSetName, counterSet := range allocating {
		counterSetRemaining, ok := remaining[counterSetName]
		if !ok {
			continue
		}
		for counterName, allocCounter := range counterSet {
			remCounter, ok := counterSetRemaining[counterName]
			if !ok {
				continue
			}
			if remCounter.Value.Value()-allocCounter.Value.Value() <= 0 {
				return true
			}
		}
	}
	return false
}

// checkCounters verifies that shared counters have sufficient remaining budget for the device.
// remainingCounterSets is the base budget (from AllocationTracker for in-cluster pools, or
// templateRemainingCounters for template pools). The DFS-local allocatingCounters are subtracted
// to account for tentative allocations in the current search.
func (a *allocator) checkCounters(device cloudprovider.Device, poolKey PoolKey, remainingCounterSets map[string]map[string]resourcev1.Counter, template bool) bool {
	if len(device.ConsumesCounters) == 0 {
		return true
	}
	if remainingCounterSets == nil {
		return false
	}
	allocatingCounterSets := lo.Ternary(template, a.templateAllocatingCounters[poolKey], a.allocatingCounters[poolKey])
	for _, consumption := range device.ConsumesCounters {
		counterSetRemaining, ok := remainingCounterSets[consumption.CounterSet]
		if !ok {
			return false
		}
		var allocatingCounters map[string]resourcev1.Counter
		if allocatingCounterSets != nil {
			allocatingCounters = allocatingCounterSets[consumption.CounterSet]
		}
		for counterName, counter := range consumption.Counters {
			remainingCounter, ok := counterSetRemaining[counterName]
			if !ok {
				return false
			}
			allocatingVal := int64(0)
			if allocatingCounters != nil {
				if ac, ok := allocatingCounters[counterName]; ok {
					allocatingVal = ac.Value.Value()
				}
			}
			if remainingCounter.Value.Value()-allocatingVal < counter.Value.Value() {
				return false
			}
		}
	}
	return true
}

// deductAllocatingCounters adds a device's counter consumption to the DFS-local allocating state.
func (a *allocator) deductAllocatingCounters(device cloudprovider.Device, poolKey PoolKey, template bool) {
	if len(device.ConsumesCounters) == 0 {
		return
	}
	counterMap := lo.Ternary(template, a.templateAllocatingCounters, a.allocatingCounters)
	allocatingCounterSets, ok := counterMap[poolKey]
	if !ok {
		allocatingCounterSets = make(map[string]map[string]resourcev1.Counter)
		counterMap[poolKey] = allocatingCounterSets
	}
	for _, consumption := range device.ConsumesCounters {
		allocatingCounters, ok := allocatingCounterSets[consumption.CounterSet]
		if !ok {
			allocatingCounters = make(map[string]resourcev1.Counter)
			allocatingCounterSets[consumption.CounterSet] = allocatingCounters
		}
		for counterName, counter := range consumption.Counters {
			allocatingCounter := allocatingCounters[counterName]
			allocatingCounter.Value.Add(counter.Value)
			allocatingCounters[counterName] = allocatingCounter
		}
	}
}

// restoreAllocatingCounters reverses a device's counter consumption from the DFS-local allocating state.
func (a *allocator) restoreAllocatingCounters(device cloudprovider.Device, poolKey PoolKey, template bool) {
	if len(device.ConsumesCounters) == 0 {
		return
	}
	counterMap := lo.Ternary(template, a.templateAllocatingCounters, a.allocatingCounters)
	allocatingCounterSets, ok := counterMap[poolKey]
	if !ok {
		return
	}
	for _, consumption := range device.ConsumesCounters {
		allocatingCounters, ok := allocatingCounterSets[consumption.CounterSet]
		if !ok {
			continue
		}
		for counterName, counter := range consumption.Counters {
			allocatingCounter, ok := allocatingCounters[counterName]
			if !ok {
				continue
			}
			allocatingCounter.Value.Sub(counter.Value)
			allocatingCounters[counterName] = allocatingCounter
		}
	}
}

// countersFeasible checks whether the remaining counter budgets can possibly
// satisfy the aggregate demand from all requests. This is a conservative
// lower-bound check: if even the minimum total consumption exceeds available
// budget, no DFS path can succeed. This is only done for AllMode requests
// as their eligible devices (both in-cluster and template) are pre-computed.
func (a *allocator) countersFeasible() bool {
	for _, cd := range a.claimData {
		for _, rd := range cd.Requests {
			if rd.AllocationMode == resourcev1.DeviceAllocationModeAll {
				if !a.allModeCountersFeasible(&rd) {
					return false
				}
			}
		}
	}
	return true
}

//nolint:gocyclo
func (a *allocator) allModeCountersFeasible(rd *RequestData) bool {
	// For All mode, we must allocate all predetermined devices.
	// Decrement from shadow copies of remaining counters as we iterate.
	// Map: poolKey -> counterSetName -> counterName -> Counter
	inClusterShadow := make(map[PoolKey]map[string]map[string]resourcev1.Counter)
	templateShadow := make(map[PoolKey]map[string]map[string]resourcev1.Counter)

	devices := rd.AllDevices
	if templateDevices, ok := rd.AllTemplateDevicesByIT[a.itID]; ok {
		devices = append(devices, templateDevices...)
	}

	for _, d := range devices {
		if len(d.ConsumesCounters) == 0 {
			continue
		}
		poolKey := PoolKey{Driver: d.ID.Driver, Pool: d.ID.Pool}

		var shadow map[PoolKey]map[string]map[string]resourcev1.Counter
		if d.ID.Template {
			shadow = templateShadow
		} else {
			shadow = inClusterShadow
		}

		if _, ok := shadow[poolKey]; !ok {
			var remaining map[string]map[string]resourcev1.Counter
			if d.ID.Template {
				if a.templateRemainingCounters != nil {
					remaining = a.templateRemainingCounters[poolKey]
				}
			} else {
				remaining = a.allocationTracker.RemainingCounters[poolKey]
			}
			// This is uninitialized before first DFS; defer to checkCounters which
			// initializes the remaining counters.
			if remaining == nil {
				return true
			}
			shadow[poolKey] = copyCounterSets(remaining)
		}
		poolShadow := shadow[poolKey]
		for _, consumption := range d.ConsumesCounters {
			counterSetsShadow, ok := poolShadow[consumption.CounterSet]
			if !ok {
				return false
			}
			for counterName, counter := range consumption.Counters {
				availCounter, ok := counterSetsShadow[counterName]
				if !ok {
					return false
				}
				availCounter.Value.Sub(counter.Value)
				if availCounter.Value.Cmp(resource.Quantity{}) < 0 {
					return false
				}
				counterSetsShadow[counterName] = availCounter
			}
		}
	}
	return true
}

func copyCounterSets(src map[string]map[string]resourcev1.Counter) map[string]map[string]resourcev1.Counter {
	cp := make(map[string]map[string]resourcev1.Counter, len(src))
	for counterSetName, counters := range src {
		cpCounters := make(map[string]resourcev1.Counter, len(counters))
		for counterName, counter := range counters {
			cpCounters[counterName] = resourcev1.Counter{Value: counter.Value.DeepCopy()}
		}
		cp[counterSetName] = cpCounters
	}
	return cp
}
