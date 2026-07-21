/*
Copyright 2018 The Kubernetes Authors.

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

package cache

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/go-logr/logr"
)

func Test_WithClock_whenPassedToNew_ConfiguresClock(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	clk := newFakePassiveClock()

	cache := New[string, string](
		logr.Discard(),
		WithClock(clk))

	g.Expect(cache.clock).To(Equal(clk))
}

func Test_WithEntryTTL_whenPassedToNew_ConfiguresEntryTTL(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	ttl := 42 * time.Second

	cache := New[string, string](
		logr.Discard(),
		WithTTL(ttl))

	g.Expect(cache.ttl).To(Equal(ttl))
}
