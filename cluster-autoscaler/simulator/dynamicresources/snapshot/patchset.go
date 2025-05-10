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
// It also implements lazy caching while also attempting to maintain cache up to
// date while performing fetch operations.
type patchSet[K comparable, V any] struct {
	patches []*patch[K, V]
	cache   map[K]*V
}

// newPatchSet creates a new patchSet, initializing it with the provided base patches.
func newPatchSet[K comparable, V any](patches ...*patch[K, V]) *patchSet[K, V] {
	return &patchSet[K, V]{
		patches: patches,
		cache:   make(map[K]*V),
	}
}

// Fork adds a new, empty patch layer to the top of the stack.
// Subsequent modifications will be recorded in this new layer.
func (ps *patchSet[K, V]) Fork() {
	ps.patches = append(ps.patches, newPatch[K, V]())
}

// Commit merges the topmost patch layer into the one below it.
// If there's only one layer (or none), it's a no-op.
func (ps *patchSet[K, V]) Commit() {
	if len(ps.patches) < 2 {
		return
	}

	currentPatch := ps.patches[len(ps.patches)-1]
	previousPatch := ps.patches[len(ps.patches)-2]
	mergedPatch := mergePatchesInPlace(previousPatch, currentPatch)
	ps.patches = ps.patches[:len(ps.patches)-1]
	ps.patches[len(ps.patches)-1] = mergedPatch
}

// Revert removes the topmost patch layer.
// Any modifications or deletions recorded in that layer are discarded.
func (ps *patchSet[K, V]) Revert() {
	if len(ps.patches) <= 1 {
		return
	}

	currentPatch := ps.patches[len(ps.patches)-1]
	ps.patches = ps.patches[:len(ps.patches)-1]

	for key := range currentPatch.Modified {
		delete(ps.cache, key)
	}

	for key := range currentPatch.Deleted {
		delete(ps.cache, key)
	}
}

// FindValue searches for the effective value of a key by looking through the patches
// from top to bottom. It returns the value and true if found, or the zero value and false
// if the key is deleted or not found in any patch.
func (ps *patchSet[K, V]) FindValue(key K) (value V, found bool) {
	var zero V

	if cachedValue, cacheHit := ps.cache[key]; cacheHit {
		if cachedValue == nil {
			return zero, false
		}

		return *cachedValue, true
	}

	value = zero
	found = false
	for i := len(ps.patches) - 1; i >= 0; i-- {
		patch := ps.patches[i]
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
		ps.cache[key] = &value
	} else {
		ps.cache[key] = nil
	}

	return value, found
}

// AsMap merges all patches into a single map representing the current effective state.
// It iterates through all patches from bottom to top, applying modifications and deletions.
// The cache is populated with the results during this process.
func (ps *patchSet[K, V]) AsMap() map[K]V {
	keysCount := ps.approximateKeysCount()
	merged := make(map[K]V, keysCount)

	for _, patch := range ps.patches {
		for key, value := range patch.Modified {
			merged[key] = value
		}

		for key := range patch.Deleted {
			delete(merged, key)
		}
	}

	for key, value := range merged {
		ps.cache[key] = &value
	}

	return merged
}

// SetCurrent adds or updates a key-value pair in the topmost patch layer.
func (ps *patchSet[K, V]) SetCurrent(key K, value V) {
	currentPatch := ps.patches[len(ps.patches)-1]
	currentPatch.Set(key, value)
	ps.cache[key] = &value
}

// DeleteCurrent marks a key as deleted in the topmost patch layer.
func (ps *patchSet[K, V]) DeleteCurrent(key K) {
	currentPatch := ps.patches[len(ps.patches)-1]
	currentPatch.Delete(key)
	ps.cache[key] = nil
}

// InCurrentPatch checks if the key is available in the topmost patch layer.
func (ps *patchSet[K, V]) InCurrentPatch(key K) bool {
	currentPatch := ps.patches[len(ps.patches)-1]
	_, found := currentPatch.Get(key)
	return found
}

// approximateKeysCount calculates an approximate total number of key-value
// pairs across all patches taking the highest number of records possible
// this calculation does not consider deleted records - thus it is likely
// to be not accurate
func (ps *patchSet[K, V]) approximateKeysCount() int {
	count := 0
	for _, patch := range ps.patches {
		count += len(patch.Modified)
	}

	return count
}
