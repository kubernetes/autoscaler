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

import (
	"fmt"
	"testing"
)

func TestPersistentMapBasic(t *testing.T) {
	m := NewpersistentMap[string, int]()
	if m.Len() != 0 {
		t.Errorf("Expected empty map, got len %d", m.Len())
	}

	// Store
	m1 := m.Store("a", 1)
	if m1.Len() != 1 {
		t.Errorf("Expected len 1, got %d", m1.Len())
	}
	if m.Len() != 0 {
		t.Errorf("Original map mutated: expected len 0, got %d", m.Len())
	}

	// Load
	val, ok := m1.Load("a")
	if !ok || val != 1 {
		t.Errorf("Expected to load 1 for key 'a', got %v, %t", val, ok)
	}

	// Overwrite
	m2 := m1.Store("a", 2)
	if m2.Len() != 1 {
		t.Errorf("Expected len 1 after overwrite, got %d", m2.Len())
	}
	val, ok = m2.Load("a")
	if !ok || val != 2 {
		t.Errorf("Expected to load 2 for key 'a', got %v, %t", val, ok)
	}
	val, ok = m1.Load("a")
	if !ok || val != 1 {
		t.Errorf("Original map affected by overwrite: expected 1, got %v, %t", val, ok)
	}

	// Delete
	m3 := m2.Delete("a")
	if m3.Len() != 0 {
		t.Errorf("Expected len 0 after delete, got %d", m3.Len())
	}
	_, ok = m3.Load("a")
	if ok {
		t.Error("Expected key 'a' to be deleted")
	}
}

func TestPersistentMapMultipleKeys(t *testing.T) {
	m := NewpersistentMap[string, int]()
	keys := []string{"foo", "bar", "baz", "qux"}
	for i, k := range keys {
		m = m.Store(k, i+10)
	}

	if m.Len() != len(keys) {
		t.Errorf("Expected len %d, got %d", len(keys), m.Len())
	}

	for i, k := range keys {
		val, ok := m.Load(k)
		if !ok || val != i+10 {
			t.Errorf("Expected %d for key %q, got %v, %t", i+10, k, val, ok)
		}
	}
}

func TestTransientMapMutations(t *testing.T) {
	m := NewpersistentMap[string, int]()
	transient := m.AsTransient()

	if transient.Len() != 0 {
		t.Errorf("Expected empty transient, got len %d", transient.Len())
	}

	// Store in transient
	transient.Store("a", 1)
	transient.Store("b", 2)
	transient.Store("c", 3)

	if transient.Len() != 3 {
		t.Errorf("Expected transient len 3, got %d", transient.Len())
	}

	// Verify original is still empty
	if m.Len() != 0 {
		t.Errorf("Original map mutated during transient work: got len %d", m.Len())
	}

	// Convert to persistent
	p := transient.Persistent()
	if p.Len() != 3 {
		t.Errorf("Expected persistent len 3, got %d", p.Len())
	}

	for k, expected := range map[string]int{"a": 1, "b": 2, "c": 3} {
		val, ok := p.Load(k)
		if !ok || val != expected {
			t.Errorf("Expected %d for key %q, got %v, %t", expected, k, val, ok)
		}
	}

	// Verify transient methods panic after transition
	verifyPanic(t, "Store", func() {
		transient.Store("d", 4)
	})
	verifyPanic(t, "Delete", func() {
		transient.Delete("a")
	})
	verifyPanic(t, "Persistent", func() {
		transient.Persistent()
	})
}

func verifyPanic(t *testing.T, name string, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected %s to panic, but it did not", name)
		}
	}()
	f()
}

func TestPersistentMapRange(t *testing.T) {
	input := map[string]int{
		"k1": 100,
		"k2": 200,
		"k3": 300,
	}

	m := NewpersistentMapFromNativeMap(input)
	visited := make(map[string]int)

	m.Range(func(k string, v int) bool {
		visited[k] = v
		return true
	})

	if len(visited) != len(input) {
		t.Errorf("Range visited %d elements, expected %d", len(visited), len(input))
	}

	for k, v := range input {
		if visited[k] != v {
			t.Errorf("Range value mismatch for key %q: got %d, expected %d", k, visited[k], v)
		}
	}
}

func TestPersistentMapScale(t *testing.T) {
	const count = 5000
	m := NewpersistentMap[int, int]()

	// 1. Bulk insertion
	session := &editSession{}
	transient := &transientMap[int, int]{session: session}
	for i := 0; i < count; i++ {
		transient.Store(i, i*10)
	}

	m = transient.Persistent()
	if m.Len() != count {
		t.Errorf("Expected scale map len %d, got %d", count, m.Len())
	}

	// 2. Bulk validation
	for i := 0; i < count; i++ {
		val, ok := m.Load(i)
		if !ok || val != i*10 {
			t.Fatalf("Load failed for key %d: got %v, %t, expected %d", i, val, ok, i*10)
		}
	}

	// 3. Bulk deletions of even keys
	transient = m.AsTransient()
	for i := 0; i < count; i += 2 {
		transient.Delete(i)
	}
	m = transient.Persistent()

	if m.Len() != count/2 {
		t.Errorf("Expected len %d after deleting half, got %d", count/2, m.Len())
	}

	// 4. Validate remaining keys
	for i := 0; i < count; i++ {
		val, ok := m.Load(i)
		if i%2 == 0 {
			if ok {
				t.Fatalf("Key %d should have been deleted, but was found with value %d", i, val)
			}
		} else {
			if !ok || val != i*10 {
				t.Fatalf("Load failed for key %d: got %v, %t, expected %d", i, val, ok, i*10)
			}
		}
	}
}

func TestPersistentMapComplexStructKeys(t *testing.T) {
	type complexKey struct {
		ID   int
		Name string
	}

	m := NewpersistentMap[complexKey, string]()
	k1 := complexKey{ID: 1, Name: "one"}
	k2 := complexKey{ID: 2, Name: "two"}

	m = m.Store(k1, "value-one")
	m = m.Store(k2, "value-two")

	val, ok := m.Load(k1)
	if !ok || val != "value-one" {
		t.Errorf("Expected 'value-one', got %q, %t", val, ok)
	}

	val, ok = m.Load(k2)
	if !ok || val != "value-two" {
		t.Errorf("Expected 'value-two', got %q, %t", val, ok)
	}
}

func TestSafeEqual(t *testing.T) {
	// Simple comparable values
	if !safeEqual(10, 10) {
		t.Error("Expected 10 == 10")
	}
	if safeEqual(10, 20) {
		t.Error("Expected 10 != 20")
	}

	// Strings
	if !safeEqual("hello", "hello") {
		t.Error("Expected strings to be equal")
	}

	// Slices (normally non-comparable, should return false instead of panicking)
	slice1 := []int{1, 2}
	slice2 := []int{1, 2}
	if safeEqual(slice1, slice2) {
		t.Error("Expected slices to not be equal under safeEqual")
	}

	// Same slice variable (since boxing copies the slice header, interface data pointers are different, so it should safely return false without panicking)
	if safeEqual(slice1, slice1) {
		t.Error("Expected slice compared to itself to be false under safeEqual (due to boxing pointer difference)")
	}
}

func TestNewpersistentMapVarargs(t *testing.T) {
	m := NewpersistentMap[string, int](
		persistentMapItem[string, int]{Key: "a", Value: 1},
		persistentMapItem[string, int]{Key: "b", Value: 2},
	)

	if m.Len() != 2 {
		t.Errorf("Expected len 2, got %d", m.Len())
	}
	if val, ok := m.Load("a"); !ok || val != 1 {
		t.Errorf("Expected 'a' to be 1, got %d, %t", val, ok)
	}
	if val, ok := m.Load("b"); !ok || val != 2 {
		t.Errorf("Expected 'b' to be 2, got %d, %t", val, ok)
	}
}

func TestArrayNodeTransitions(t *testing.T) {
	// Array node has 32 slots. If a bitmap node grows past 16 occupied slots,
	// it transitions to an arrayNode. If it shrinks below 8, it transitions back.
	// Let's create keys that end up in the same node but different slots at a specific level.
	// Actually, just inserting 30 keys with slightly different values will naturally disperse them,
	// but to make sure they share a prefix, we can just insert many keys.
	m := NewpersistentMap[string, int]()
	const numKeys = 50
	for i := 0; i < numKeys; i++ {
		m = m.Store(fmt.Sprintf("key-%d", i), i)
	}

	// Verify all keys are present
	for i := 0; i < numKeys; i++ {
		val, ok := m.Load(fmt.Sprintf("key-%d", i))
		if !ok || val != i {
			t.Errorf("Expected %d for key-%d, got %v, %t", i, i, val, ok)
		}
	}

	// Delete them one by one
	for i := 0; i < numKeys; i++ {
		m = m.Delete(fmt.Sprintf("key-%d", i))
		val, ok := m.Load(fmt.Sprintf("key-%d", i))
		if ok {
			t.Errorf("key-%d should be deleted, but loaded value %v", i, val)
		}
	}

	if m.Len() != 0 {
		t.Errorf("Expected empty map after all deletes, got len %d", m.Len())
	}
}
