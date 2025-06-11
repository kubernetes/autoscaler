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

// Patch represents a single layer of modifications (additions/updates)
// and deletions for a set of key-value pairs. It's used as a building
// block for the patchSet.
type Patch[K comparable, V any] struct {
	modified map[K]V
	deleted  map[K]bool
}

// Set marks a key-value pair as modified in the current patch.
// If the key was previously marked as deleted, the deletion mark is removed.
func (p *Patch[K, V]) Set(key K, value V) {
	p.modified[key] = value
	delete(p.deleted, key)
}

// Delete marks a key as deleted in the current patch.
// If the key was previously marked as modified, the modification is removed.
func (p *Patch[K, V]) Delete(key K) {
	delete(p.modified, key)
	p.deleted[key] = true
}

// Get retrieves the modified value for a key within this specific patch.
// It returns the value and true if the key was found in this patch
// or zero value and false otherwise.
func (p *Patch[K, V]) Get(key K) (value V, found bool) {
	value, found = p.modified[key]
	return value, found
}

// IsDeleted checks if the key is marked as deleted within this specific patch.
func (p *Patch[K, V]) IsDeleted(key K) bool {
	return p.deleted[key]
}

// NewPatch creates a new, empty patch with no modifications or deletions.
func NewPatch[K comparable, V any]() *Patch[K, V] {
	return &Patch[K, V]{
		modified: make(map[K]V),
		deleted:  make(map[K]bool),
	}
}

// NewPatchFromMap creates a new patch initialized with the data from the provided map
// the data supplied is recorded as modified in the patch.
func NewPatchFromMap[M ~map[K]V, K comparable, V any](source M) *Patch[K, V] {
	if source == nil {
		source = make(M)
	}

	return &Patch[K, V]{
		modified: source,
		deleted:  make(map[K]bool),
	}
}

// mergePatchesInPlace merges two patches into one, while modifying patch a
// inplace taking records in b as a priority when overwrites are required
func mergePatchesInPlace[K comparable, V any](a *Patch[K, V], b *Patch[K, V]) *Patch[K, V] {
	for key, value := range b.modified {
		a.Set(key, value)
	}

	for key := range b.deleted {
		a.Delete(key)
	}

	return a
}
