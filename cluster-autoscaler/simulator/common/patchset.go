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

import "unsafe"

// PatchSet manages a stack of patches, allowing for fork/revert/commit operations.
// It provides a view of the data as if all patches were applied sequentially.
//
// Time Complexities:
//   - Fork(): O(1).
//   - Commit(): O(1).
//   - Revert(): O(1).
//   - FindValue(key): O(log M) where M is the number of keys in the map (effectively O(1) lookup).
//   - AsMap(): O(M) where M is the number of keys in the map.
//   - SetCurrent(key, value): O(log M) (in-place mutable speed for transient top layer!).
//   - DeleteCurrent(key): O(log M) (in-place mutable speed for transient top layer!).
//   - InCurrentPatch(key): O(log M).
type PatchSet[K comparable, V any] struct {
	stack []*persistentMap[K, V] // Frozen immutable lower layers
	top   *transientMap[K, V]    // Active mutable transient topmost layer
}

// NewPatchSetFromMap creates a new PatchSet initialized with the provided map.
// This is more performant than creating a Patch and passing it to NewPatchSet
// as it populates the persistent map in a single pass without allocation of Patch structures.
func NewPatchSetFromMap[K comparable, V any](m map[K]V) *PatchSet[K, V] {
	base := NewpersistentMapFromNativeMap(m)
	return &PatchSet[K, V]{
		stack: []*persistentMap[K, V]{base},
		top:   base.AsTransient(),
	}
}

// NewPatchSet creates a new PatchSet, initializing it with the provided base patches.
func NewPatchSet[K comparable, V any](patches ...*Patch[K, V]) *PatchSet[K, V] {
	if len(patches) == 0 {
		base := NewpersistentMap[K, V]()
		return &PatchSet[K, V]{
			stack: []*persistentMap[K, V]{base},
			top:   base.AsTransient(),
		}
	}

	stack := make([]*persistentMap[K, V], len(patches))
	state := NewpersistentMap[K, V]()
	for i, patch := range patches {
		state = patch.applyTo(state)
		stack[i] = state
	}

	return &PatchSet[K, V]{
		stack: stack,
		top:   stack[len(stack)-1].AsTransient(),
	}
}

func (p *Patch[K, V]) applyTo(m *persistentMap[K, V]) *persistentMap[K, V] {
	session := &editSession{}
	transient := &transientMap[K, V]{session: session, root: m.root, size: m.size}

	for k, v := range p.modified {
		transient.Store(k, v)
	}

	for k := range p.deleted {
		transient.Delete(k)
	}

	return transient.Persistent()
}

// Fork adds a new, empty patch layer to the top of the stack.
// Subsequent modifications will be recorded in this new layer.
func (p *PatchSet[K, V]) Fork() {
	if len(p.stack) == 0 {
		base := NewpersistentMap[K, V]()
		p.stack = []*persistentMap[K, V]{base}
		p.top = base.AsTransient()
		return
	}

	// Freeze the active transient layer and store it back to the stack
	frozen := p.top.Persistent()
	p.stack[len(p.stack)-1] = frozen

	// Append it as the new layer
	p.stack = append(p.stack, frozen)

	// Create the new active transient layer forked from it
	p.top = frozen.AsTransient()
}

// Commit merges the topmost patch layer into the one below it.
// If there's only one layer (or none), it's a no-op.
func (p *PatchSet[K, V]) Commit() {
	if len(p.stack) < 2 {
		return
	}

	// Freeze the active transient layer
	frozen := p.top.Persistent()

	topIndex := len(p.stack) - 1
	prevIndex := len(p.stack) - 2

	// Overwrite the parent layer with the frozen state
	p.stack[prevIndex] = frozen

	// Drop the top layer from the stack
	p.stack = p.stack[:topIndex]

	// Recreate the active transient from the new topmost layer
	p.top = p.stack[len(p.stack)-1].AsTransient()
}

// Revert removes the topmost patch layer.
// Any modifications or deletions recorded in that layer are discarded.
func (p *PatchSet[K, V]) Revert() {
	if len(p.stack) <= 1 {
		return
	}

	// Discard the active transient layer (Persistent is NOT called, so it is just abandoned)
	p.stack = p.stack[:len(p.stack)-1]

	// Recreate the active transient from the restored parent layer
	p.top = p.stack[len(p.stack)-1].AsTransient()
}

// FindValue searches for the effective value of a key by looking through the patches
// from top to bottom. It returns the value and true if found, or the zero value and false
// if the key is deleted or not found in any patch.
func (p *PatchSet[K, V]) FindValue(key K) (value V, found bool) {
	var zero V

	if len(p.stack) == 0 {
		return zero, false
	}

	return p.top.Load(key)
}

// AsMap merges all patches into a single map representing the current effective state.
// It iterates through all patches from bottom to top, applying modifications and deletions.
func (p *PatchSet[K, V]) AsMap() map[K]V {
	if len(p.stack) == 0 {
		return make(map[K]V)
	}

	return p.top.ToNativeMap()
}

// SetCurrent adds or updates a key-value pair in the topmost patch layer.
func (p *PatchSet[K, V]) SetCurrent(key K, value V) {
	if len(p.stack) == 0 {
		p.Fork()
	}

	p.top.Store(key, value)
}

// DeleteCurrent marks a key as deleted in the topmost patch layer.
func (p *PatchSet[K, V]) DeleteCurrent(key K) {
	if len(p.stack) == 0 {
		p.Fork()
	}

	p.top.Delete(key)
}

// InCurrentPatch checks if the key is available in the topmost patch layer.
func (p *PatchSet[K, V]) InCurrentPatch(key K) bool {
	if len(p.stack) == 0 {
		return false
	}
	topVal, active := p.FindValue(key)
	if !active {
		return false
	}

	if len(p.stack) == 1 {
		return true
	}

	prevMap := p.stack[len(p.stack)-2]
	prevVal, prevFound := prevMap.Load(key)
	if !prevFound {
		return true
	}

	return !safeEqual(topVal, prevVal)
}

// interfaceHeader represents the internal memory layout of a Go empty interface (any).
type interfaceHeader struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
}

// safeEqual checks if two values are equal or represent the same instance,
// safely handling non-comparable types like slices without panicking,
// using high-performance pointer comparisons.
func safeEqual(a, b any) bool {
	ha := *(*interfaceHeader)(unsafe.Pointer(&a))
	hb := *(*interfaceHeader)(unsafe.Pointer(&b))

	// If types are different, they are not equal.
	if ha.typ != hb.typ {
		return false
	}

	// Fast path: if the data pointers are identical, they are the same instance/value.
	if ha.data == hb.data {
		return true
	}

	// Slow path: different data pointers.
	return standardEqual(a, b)
}

// standardEqual performs standard Go comparison with panic recovery for non-comparable types.
func standardEqual(a, b any) (res bool) {
	defer func() {
		if r := recover(); r != nil {
			res = false
		}
	}()
	return a == b
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

	cloned := &PatchSet[K, V]{
		stack: make([]*persistentMap[K, V], len(ps.stack)),
	}

	cloneMap := func(m *persistentMap[K, V]) *persistentMap[K, V] {
		session := &editSession{}
		transient := &transientMap[K, V]{session: session}
		m.Range(func(k K, v V) bool {
			transient.Store(cloneKey(k), cloneValue(v))
			return true
		})
		return transient.Persistent()
	}

	for i, m := range ps.stack {
		if i == len(ps.stack)-1 {
			session := &editSession{}
			transient := &transientMap[K, V]{session: session}
			ps.top.Range(func(k K, v V) bool {
				transient.Store(cloneKey(k), cloneValue(v))
				return true
			})
			cloned.stack[i] = transient.Persistent()
		} else {
			cloned.stack[i] = cloneMap(m)
		}
	}

	cloned.top = cloned.stack[len(cloned.stack)-1].AsTransient()
	return cloned
}
