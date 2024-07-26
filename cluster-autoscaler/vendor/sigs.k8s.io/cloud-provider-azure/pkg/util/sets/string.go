/*
Copyright 2024 The Kubernetes Authors.

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

package sets

import (
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

// IgnoreCaseSet is a set of strings that is case-insensitive.
type IgnoreCaseSet struct {
	set sets.Set[string]
}

// NewString creates a new IgnoreCaseSet with the given items.
func NewString(items ...string) *IgnoreCaseSet {
	var lowerItems []string
	for _, item := range items {
		lowerItems = append(lowerItems, strings.ToLower(item))
	}
	set := sets.New[string](lowerItems...)
	return &IgnoreCaseSet{set: set}
}

// Insert adds the given items to the set. It only works if the set is initialized.
func (s *IgnoreCaseSet) Insert(items ...string) {
	var lowerItems []string
	for _, item := range items {
		lowerItems = append(lowerItems, strings.ToLower(item))
	}
	for _, item := range lowerItems {
		s.set.Insert(item)
	}
}

// SafeInsert creates a new IgnoreCaseSet with the given items if the set is not initialized.
// This is the recommended way to insert elements into the set.
func SafeInsert(s *IgnoreCaseSet, items ...string) *IgnoreCaseSet {
	if s.Initialized() {
		s.Insert(items...)
		return s
	}
	return NewString(items...)
}

// Delete removes the given item from the set.
// It will be a no-op if the set is not initialized or the item is not in the set.
func (s *IgnoreCaseSet) Delete(item string) bool {
	var has bool
	item = strings.ToLower(item)
	if s.Initialized() && s.Has(item) {
		s.set.Delete(item)
		has = true
	}
	return has
}

// Has returns true if the given item is in the set, and the set is initialized.
func (s *IgnoreCaseSet) Has(item string) bool {
	if !s.Initialized() {
		return false
	}
	return s.set.Has(strings.ToLower(item))
}

// Initialized returns true if the set is initialized.
func (s *IgnoreCaseSet) Initialized() bool {
	return s != nil && s.set != nil
}

// UnsortedList returns the items in the set in an arbitrary order.
func (s *IgnoreCaseSet) UnsortedList() []string {
	if !s.Initialized() {
		return []string{}
	}
	return s.set.UnsortedList()
}

// Len returns the number of items in the set.
func (s *IgnoreCaseSet) Len() int {
	if !s.Initialized() {
		return 0
	}
	return s.set.Len()
}
