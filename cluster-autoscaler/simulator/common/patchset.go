/*
Copyright 2025 The Kubernetes Authors.

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

package common

// PatchSet manages a stack of patches, allowing for fork/revert/commit operations.
// It provides a view of the data as if all patches were applied sequentially.
//
// Time Complexities:
//   - Fork(): O(1).
//   - Commit(): O(P), where P is the number of modified/deleted entries
//     in the current patch or no-op for PatchSet with a single patch.
//   - Revert(): O(P), where P is the number of modified/deleted entries
//     in the topmost patch or no-op for PatchSet with a single patch.
//   - FindValue(key): O(1) for cached keys, O(N * P) - otherwise.
//   - AsMap(): O(N * P) as it needs to iterate through all the layers
//     modifications and deletions to get the latest state. Best case -
//     if the cache is currently in sync with the PatchSet data complexity becomes
//     O(C), where C is the actual number key/value pairs in flattened PatchSet.
//   - SetCurrent(key, value): O(1).
//   - DeleteCurrent(key): O(1).
//   - InCurrentPatch(key): O(1).
//
// Variables used in complexity analysis:
//   - N: The number of patch layers in the PatchSet.
//   - P: The number of modified/deleted entries in a single patch layer
//
// Caching:
//
// The PatchSet employs a lazy caching mechanism to speed up access to data. When a specific item is requested,
// the cache is checked first. If the item isn't cached, its effective value is computed by traversing the layers
// of patches from the most recent to the oldest, and the result is then stored in the cache. For operations that
// require the entire dataset like AsMap, if a cache is fully up-to-date, the data is served directly from the cache.
// Otherwise, the entire dataset is rebuilt by applying all patches, and this process fully populates the cache.
// Direct modifications to the current patch update the specific item in the cache immediately. Reverting the latest
// patch will clear affected items from the cache and mark the cache as potentially out-of-sync, as underlying values may
// no longer represent the PatchSet as a whole. Committing and forking does not invalidate the cache, as the effective
// values remain consistent from the perspective of read operations.
type PatchSet[K comparable, V any] struct {
	// patches is a stack of individual modification layers. The base data is
	// at index 0, and subsequent modifications are layered on top.
	// PatchSet should always contain at least a single patch.
	patches []*Patch[K, V]

	// cache stores the computed effective value for keys that have been accessed.
	// A nil pointer indicates the key is effectively deleted or not present.
	cache map[K]*V

	// cacheInSync indicates whether the cache map accurately reflects the
	// current state derived from applying all patches in the 'patches' slice.
	// When false, the cache may be stale and needs to be rebuilt or validated
	// before being fully trusted for all keys.
	cacheInSync bool
}

// NewPatchSet creates a new PatchSet, initializing it with the provided base patches.
func NewPatchSet[K comparable, V any](patches ...*Patch[K, V]) *PatchSet[K, V] {
	return &PatchSet[K, V]{
		patches:     patches,
		cache:       make(map[K]*V),
		cacheInSync: false,
	}
}

// Fork adds a new, empty patch layer to the top of the stack.
// Subsequent modifications will be recorded in this new layer.
func (p *PatchSet[K, V]) Fork() {
	p.patches = append(p.patches, NewPatch[K, V]())
}

// Commit merges the topmost patch layer into the one below it.
// If there's only one layer (or none), it's a no-op.
func (p *PatchSet[K, V]) Commit() {
	if len(p.patches) < 2 {
		return
	}

	currentPatch := p.patches[len(p.patches)-1]
	previousPatch := p.patches[len(p.patches)-2]
	mergePatchesInPlace(previousPatch, currentPatch)
	p.patches = p.patches[:len(p.patches)-1]
}

// Revert removes the topmost patch layer.
// Any modifications or deletions recorded in that layer are discarded.
func (p *PatchSet[K, V]) Revert() {
	if len(p.patches) <= 1 {
		return
	}

	currentPatch := p.patches[len(p.patches)-1]
	p.patches = p.patches[:len(p.patches)-1]

	for key := range currentPatch.modified {
		delete(p.cache, key)
	}

	for key := range currentPatch.deleted {
		delete(p.cache, key)
	}

	p.cacheInSync = false
}

// FindValue searches for the effective value of a key by looking through the patches
// from top to bottom. It returns the value and true if found, or the zero value and false
// if the key is deleted or not found in any patch.
func (p *PatchSet[K, V]) FindValue(key K) (value V, found bool) {
	var zero V

	if cachedValue, cacheHit := p.cache[key]; cacheHit {
		if cachedValue == nil {
			return zero, false
		}

		return *cachedValue, true
	}

	value = zero
	found = false
	for i := len(p.patches) - 1; i >= 0; i-- {
		patch := p.patches[i]
		if patch.IsDeleted(key) {
			break
		}

		foundValue, ok := patch.Get(key)
		if ok {
			value = foundValue
			found = true
			break
		}
	}

	if found {
		p.cache[key] = &value
	} else {
		p.cache[key] = nil
	}

	return value, found
}

// AsMap merges all patches into a single map representing the current effective state.
// It iterates through all patches from bottom to top, applying modifications and deletions.
// The cache is populated with the results during this process.
func (p *PatchSet[K, V]) AsMap() map[K]V {
	if p.cacheInSync {
		patchSetMap := make(map[K]V, len(p.cache))
		for key, value := range p.cache {
			if value != nil {
				patchSetMap[key] = *value
			}
		}
		return patchSetMap
	}

	keysCount := p.totalKeyCount()
	patchSetMap := make(map[K]V, keysCount)

	for _, patch := range p.patches {
		for key, value := range patch.modified {
			patchSetMap[key] = value
		}

		for key := range patch.deleted {
			delete(patchSetMap, key)
		}
	}

	for key, value := range patchSetMap {
		p.cache[key] = &value
	}
	p.cacheInSync = true

	return patchSetMap
}

// SetCurrent adds or updates a key-value pair in the topmost patch layer.
func (p *PatchSet[K, V]) SetCurrent(key K, value V) {
	if len(p.patches) == 0 {
		p.Fork()
	}

	currentPatch := p.patches[len(p.patches)-1]
	currentPatch.Set(key, value)
	p.cache[key] = &value
}

// DeleteCurrent marks a key as deleted in the topmost patch layer.
func (p *PatchSet[K, V]) DeleteCurrent(key K) {
	if len(p.patches) == 0 {
		p.Fork()
	}

	currentPatch := p.patches[len(p.patches)-1]
	currentPatch.Delete(key)
	p.cache[key] = nil
}

// InCurrentPatch checks if the key is available in the topmost patch layer.
func (p *PatchSet[K, V]) InCurrentPatch(key K) bool {
	if len(p.patches) == 0 {
		return false
	}

	currentPatch := p.patches[len(p.patches)-1]
	_, found := currentPatch.Get(key)
	return found
}

// totalKeyCount calculates an approximate total number of key-value
// pairs across all patches taking the highest number of records possible
// this calculation does not consider deleted records - thus it is likely
// to be not accurate
func (p *PatchSet[K, V]) totalKeyCount() int {
	count := 0
	for _, patch := range p.patches {
		count += len(patch.modified)
	}

	return count
}

// ClonePatchSet creates a deep copy of a PatchSet object with the same patch layers
// structure, while copying keys and values using cloneKey and cloneValue functions
// provided.
//
// This function is intended for testing purposes only.
func ClonePatchSet[K comparable, V any](ps *PatchSet[K, V], cloneKey func(K) K, cloneValue func(V) V) *PatchSet[K, V] {
	if ps == nil {
		return nil
	}

	cloned := NewPatchSet[K, V]()
	for _, patch := range ps.patches {
		clonedPatch := NewPatch[K, V]()
		for key, value := range patch.modified {
			clonedKey, clonedValue := cloneKey(key), cloneValue(value)
			clonedPatch.Set(clonedKey, clonedValue)
		}

		for key := range patch.deleted {
			clonedKey := cloneKey(key)
			clonedPatch.Delete(clonedKey)
		}

		cloned.patches = append(cloned.patches, clonedPatch)
	}

	return cloned
}
