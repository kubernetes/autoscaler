/*
Copyright 2016 The Kubernetes Authors.

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

package expiring

import (
	"container/list"
	"time"

	"k8s.io/utils/clock"
)

type elementWithTimestamp struct {
	value interface{}
	added time.Time
}

// List tracks elements along with their creation times.
// This is essentially a linked list with timestamp on each entry, allowing
// dropping old entries. This struct is not thread safe.
// TODO(x13n): Migrate to generics once supported by Go stdlib (container/list
// in particular).
type List struct {
	lst   list.List
	clock clock.PassiveClock
}

// NewList creates a new expiring list.
func NewList() *List {
	return newListWithClock(clock.RealClock{})
}

// Warning: This object doesn't support time travel. Subsequent calls to
// clock.Now are expected to return non-decreasing time values.
func newListWithClock(clock clock.PassiveClock) *List {
	return &List{
		clock: clock,
	}
}

// ToSlice converts the underlying list of elements into a slice.
func (e *List) ToSlice() []interface{} {
	p := e.lst.Front()
	ps := make([]interface{}, 0, e.lst.Len())
	for p != nil {
		ps = append(ps, p.Value.(*elementWithTimestamp).value)
		p = p.Next()
	}
	return ps
}

// ToSliceWithTimestamp is identical to ToSlice, but additionally returns the
// timestamp of newest entry (or current time if there are no entries).
func (e *List) ToSliceWithTimestamp() ([]interface{}, time.Time) {
	if e.lst.Len() == 0 {
		return nil, e.clock.Now()
	}
	return e.ToSlice(), e.lst.Back().Value.(*elementWithTimestamp).added
}

// RegisterElement adds new element to the list.
func (e *List) RegisterElement(elem interface{}) {
	e.lst.PushBack(&elementWithTimestamp{elem, e.clock.Now()})
}

// DropNotNewerThan removes all elements of the list that are older or exactly
// as old as the provided time.
func (e *List) DropNotNewerThan(expiry time.Time) {
	p := e.lst.Front()
	for p != nil {
		if p.Value.(*elementWithTimestamp).added.After(expiry) {
			// First not-expired element on the list, skip checking
			// the rest.
			return
		}
		d := p
		p = p.Next()
		e.lst.Remove(d)
	}
}
