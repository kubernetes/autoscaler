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

package snapshot

// patch represents a single layer of modifications (additions/updates)
// and deletions for a set of key-value pairs. It's used as a building
// block for the patchSet.
type patch[K comparable, V any] struct {
	Modified map[K]V
	Deleted  map[K]bool
}

// Set marks a key-value pair as modified in the current patch.
// If the key was previously marked as deleted, the deletion mark is removed.
func (p *patch[K, V]) Set(key K, value V) {
	p.Modified[key] = value
	delete(p.Deleted, key)
}

// Delete marks a key as deleted in the current patch.
// If the key was previously marked as modified, the modification is removed.
func (p *patch[K, V]) Delete(key K) {
	delete(p.Modified, key)
	p.Deleted[key] = true
}

// Get retrieves the modified value for a key within this specific patch.
// It returns the value and true if the key was found in this patch
// or zero value and false otherwise.
func (p *patch[K, V]) Get(key K) (value V, found bool) {
	value, found = p.Modified[key]
	return value, found
}

// IsDeleted checks if the key is marked as deleted within this specific patch.
func (p *patch[K, V]) IsDeleted(key K) bool {
	return p.Deleted[key]
}

// newPatch creates a new, empty patch with no modifications or deletions.
func newPatch[K comparable, V any]() *patch[K, V] {
	return &patch[K, V]{
		Modified: make(map[K]V),
		Deleted:  make(map[K]bool),
	}
}

// newPatchFromMap creates a new patch initialized with the data from the provided map
// the data supplied is recorded as modified in the patch.
func newPatchFromMap[M ~map[K]V, K comparable, V any](source M) *patch[K, V] {
	if source == nil {
		source = make(M)
	}

	return &patch[K, V]{
		Modified: source,
		Deleted:  make(map[K]bool),
	}
}

// mergePatchesInPlace merges two patches into one, while modifying patch a
// inplace taking records in b as a priority when overwrites are required
func mergePatchesInPlace[K comparable, V any](a *patch[K, V], b *patch[K, V]) *patch[K, V] {
	for key, value := range b.Modified {
		a.Set(key, value)
	}

	for key := range b.Deleted {
		a.Delete(key)
	}

	return a
}

// patchSet manages a stack of patches, allowing for fork/revert/commit operations.
// It provides a view of the data as if all patches were applied sequentially.
//
// Time Complexities:
//   - Fork(): O(1).
//   - Commit(): O(P), where P is the number of modified/deleted entries
//     in the current patch or no-op for patchSet with a single patch.
//   - Revert(): O(P), where P is the number of modified/deleted entries
//     in the topmost patch or no-op for patchSet with a single patch.
//   - FindValue(key): O(1) for cached keys, O(N * P) - otherwise.
//   - AsMap(): O(N * P) as it needs to iterate through all the layers
//     modifications and deletions to get the latest state. Best case -
//     if the cache is currently in sync with the patchSet data complexity becomes
//     O(C), where C is the actual number key/value pairs in flattened patchSet.
//   - SetCurrent(key, value): O(1).
//   - DeleteCurrent(key): O(1).
//   - InCurrentPatch(key): O(1).
//
// Variables used in complexity analysis:
//   - N: The number of patch layers in the patchSet.
//   - P: The number of modified/deleted entries in a single patch layer
//
// Caching:
//
// The patchSet employs a lazy caching mechanism to speed up access to data. When a specific item is requested,
// the cache is checked first. If the item isn't cached, its effective value is computed by traversing the layers
// of patches from the most recent to the oldest, and the result is then stored in the cache. For operations that
// require the entire dataset like AsMap, if a cache is fully up-to-date, the data is served directly from the cache.
// Otherwise, the entire dataset is rebuilt by applying all patches, and this process fully populates the cache.
// Direct modifications to the current patch update the specific item in the cache immediately. Reverting the latest
// patch will clear affected items from the cache and mark the cache as potentially out-of-sync, as underlying values may
// no longer represent the patchSet as a whole. Committing and forking does not invalidate the cache, as the effective
// values remain consistent from the perspective of read operations.
type patchSet[K comparable, V any] struct {
	// patches is a stack of individual modification layers. The base data is
	// at index 0, and subsequent modifications are layered on top.
	// PatchSet should always contain at least a single patch.
	patches []*patch[K, V]

	// cache stores the computed effective value for keys that have been accessed.
	// A nil pointer indicates the key is effectively deleted or not present.
	cache map[K]*V

	// cacheInSync indicates whether the cache map accurately reflects the
	// current state derived from applying all patches in the 'patches' slice.
	// When false, the cache may be stale and needs to be rebuilt or validated
	// before being fully trusted for all keys.
	cacheInSync bool
}

// newPatchSet creates a new patchSet, initializing it with the provided base patches.
func newPatchSet[K comparable, V any](patches ...*patch[K, V]) *patchSet[K, V] {
	return &patchSet[K, V]{
		patches:     patches,
		cache:       make(map[K]*V),
		cacheInSync: false,
	}
}

// Fork adds a new, empty patch layer to the top of the stack.
// Subsequent modifications will be recorded in this new layer.
func (p *patchSet[K, V]) Fork() {
	p.patches = append(p.patches, newPatch[K, V]())
}

// Commit merges the topmost patch layer into the one below it.
// If there's only one layer (or none), it's a no-op.
func (p *patchSet[K, V]) Commit() {
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
func (p *patchSet[K, V]) Revert() {
	if len(p.patches) <= 1 {
		return
	}

	currentPatch := p.patches[len(p.patches)-1]
	p.patches = p.patches[:len(p.patches)-1]

	for key := range currentPatch.Modified {
		delete(p.cache, key)
	}

	for key := range currentPatch.Deleted {
		delete(p.cache, key)
	}

	p.cacheInSync = false
}

// FindValue searches for the effective value of a key by looking through the patches
// from top to bottom. It returns the value and true if found, or the zero value and false
// if the key is deleted or not found in any patch.
func (p *patchSet[K, V]) FindValue(key K) (value V, found bool) {
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
func (p *patchSet[K, V]) AsMap() map[K]V {
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
		for key, value := range patch.Modified {
			patchSetMap[key] = value
		}

		for key := range patch.Deleted {
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
func (p *patchSet[K, V]) SetCurrent(key K, value V) {
	if len(p.patches) == 0 {
		p.Fork()
	}

	currentPatch := p.patches[len(p.patches)-1]
	currentPatch.Set(key, value)
	p.cache[key] = &value
}

// DeleteCurrent marks a key as deleted in the topmost patch layer.
func (p *patchSet[K, V]) DeleteCurrent(key K) {
	if len(p.patches) == 0 {
		p.Fork()
	}

	currentPatch := p.patches[len(p.patches)-1]
	currentPatch.Delete(key)
	p.cache[key] = nil
}

// InCurrentPatch checks if the key is available in the topmost patch layer.
func (p *patchSet[K, V]) InCurrentPatch(key K) bool {
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
func (p *patchSet[K, V]) totalKeyCount() int {
	count := 0
	for _, patch := range p.patches {
		count += len(patch.Modified)
	}

	return count
}
