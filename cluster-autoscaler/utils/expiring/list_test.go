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
	"testing"
	"testing/quick"
	"time"

	klog "k8s.io/klog/v2"
	clock "k8s.io/utils/clock/testing"
)

func TestToSlice(t *testing.T) {
	if err := quick.Check(identityCheck, nil); err != nil {
		t.Error(err)
	}
}

func identityCheck(list []int) bool {
	l := NewList()
	l.registerElementsFrom(list)
	return l.equals(list)
}

func TestDropNotNewer(t *testing.T) {
	if err := quick.Check(dropChecks, nil); err != nil {
		t.Error(err)
	}
}

func dropChecks(l1, l2, l3 []int) bool {
	t0 := time.Now()
	c := clock.NewFakePassiveClock(t0)
	t1, t2 := t0.Add(1*time.Minute), t0.Add(2*time.Minute)
	l := newListWithClock(c)
	l.registerElementsFrom(l1)
	c.SetTime(t1)
	l.registerElementsFrom(l2)
	c.SetTime(t2)
	if !l.equals(append(l1, l2...)) {
		return false
	}
	l.DropNotNewerThan(t0)
	if !l.equals(l2) {
		return false
	}
	l.registerElementsFrom(l3)
	if !l.equals(append(l2, l3...)) {
		return false
	}
	l.DropNotNewerThan(t1)
	if !l.equals(l3) {
		return false
	}
	l.DropNotNewerThan(t2)
	return len(l.ToSlice()) == 0
}

func (e *List) registerElementsFrom(list []int) {
	for _, i := range list {
		e.RegisterElement(i)
	}
}

func (e *List) equals(want []int) bool {
	got := e.ToSlice()
	if len(got) != len(want) {
		klog.Errorf("len(%v) != len(%v)", got, want)
		return false
	}
	for i, g := range got {
		w := want[i]
		if g.(int) != w {
			klog.Errorf("%v != %v (difference at index %v)", got, want, i)
			return false
		}
	}
	return true
}
