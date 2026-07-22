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
	"strings"
	"testing"
	"time"

	"github.com/go-logr/logr"

	. "github.com/onsi/gomega"

	clock "k8s.io/utils/clock/testing"
)

func Test_Add_cachesEntry_withExpectedExpiry(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	clock := newFakePassiveClock()

	cache := New[string, string](
		logr.Discard(),
		WithClock(clock),
		WithTTL(5*time.Minute),
	)

	// Directly add an entry to the cache
	cache.Add("test-key", "test-value")

	// The cache should now contain the entry with the expected expiry
	entry, found := cache.entries["test-key"]
	g.Expect(found).To(BeTrue())
	g.Expect(entry.value).To(Equal("test-value"))
	g.Expect(entry.expiry).To(Equal(clock.Now().Add(5 * time.Minute)))
}

func Test_AddAll_MergesEntriesAndOverwritesExistingEntries(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cache := New[string, string](logr.Discard())
	cache.Add("existing", "old-value")

	entriesToAdd := map[string]string{
		"existing": "new-value",
		"new":      "new-entry",
	}
	cache.AddAll(entriesToAdd)

	value, found := cache.Read("existing")
	g.Expect(found).To(BeTrue())
	g.Expect(value).To(Equal("new-value"))

	value, found = cache.Read("new")
	g.Expect(found).To(BeTrue())
	g.Expect(value).To(Equal("new-entry"))
}

func Test_ReplaceAll_ReplacesExistingEntries(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cache := New[string, string](logr.Discard())
	cache.Add("removed", "old-entry")
	cache.Add("retained", "old-value")

	replacementEntries := map[string]string{
		"retained": "new-value",
		"new":      "new-entry",
	}
	cache.ReplaceAll(replacementEntries)

	_, found := cache.Read("removed")
	g.Expect(found).To(BeFalse())

	value, found := cache.Read("retained")
	g.Expect(found).To(BeTrue())
	g.Expect(value).To(Equal("new-value"))

	value, found = cache.Read("new")
	g.Expect(found).To(BeTrue())
	g.Expect(value).To(Equal("new-entry"))
}

func Test_AddAllAndReplaceAll_CanonicalizeKeysAndSetExpectedExpiry(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	clock := newFakePassiveClock()
	cache := New[string, string](
		logr.Discard(),
		WithClock(clock),
		WithTTL(5*time.Minute),
		WithKeyCanonicalizer(strings.ToLower),
	)

	entriesToAdd := map[string]string{
		"ADD-ALL": "value",
	}
	cache.AddAll(entriesToAdd)
	entry, found := cache.entries["add-all"]
	g.Expect(found).To(BeTrue())
	g.Expect(entry.expiry).To(Equal(clock.Now().Add(5 * time.Minute)))

	replacementEntries := map[string]string{
		"REPLACE-ALL": "value",
	}
	cache.ReplaceAll(replacementEntries)

	_, found = cache.Read("add-all")
	g.Expect(found).To(BeFalse())

	value, found := cache.Read("replace-all")
	g.Expect(found).To(BeTrue())
	g.Expect(value).To(Equal("value"))
	entry, found = cache.entries["replace-all"]
	g.Expect(found).To(BeTrue())
	g.Expect(entry.expiry).To(Equal(clock.Now().Add(5 * time.Minute)))
}

func Test_Read_WhenCachePopulated_ReturnsCachedItem(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	clock := newFakePassiveClock()

	cache := New[string, string](
		logr.Discard(),
		WithClock(clock),
		WithTTL(5*time.Minute),
	)

	// First populate the cache
	cache.Add("test-key", "test-value")

	// First read should return the cached value
	value, ok := cache.Read("test-key")
	g.Expect(ok).To(BeTrue())
	g.Expect(value).To(Equal("test-value"))

	// Second read should also return cached value
	value, ok = cache.Read("test-key")
	g.Expect(ok).To(BeTrue())
	g.Expect(value).To(Equal("test-value"))
}

func Test_Read_WhenCacheEntryExpired_ReturnsNoValue(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	clock := newFakePassiveClock()

	cache := New[string, string](
		logr.Discard(),
		WithClock(clock),
		WithTTL(5*time.Minute),
	)

	// Populate the cache
	cache.Add("test-key", "loaded-value-1")

	// Verify the cache entry is present before advancing time
	value, ok := cache.Read("test-key")
	g.Expect(ok).To(BeTrue())
	g.Expect(value).To(Equal("loaded-value-1"))

	// Advance time to expire the cache entry
	clock.SetTime(clock.Now().Add(6 * time.Minute))

	// Second read should return no-value
	_, ok = cache.Read("test-key")
	g.Expect(ok).To(BeFalse())
}

func Test_ReadAll_ReturnsOnlyCurrentEntries(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	clock := newFakePassiveClock()
	cache := New[string, string](
		logr.Discard(),
		WithClock(clock),
		WithTTL(5*time.Minute),
	)

	cache.Add("stale", "stale-value")
	clock.SetTime(clock.Now().Add(4 * time.Minute))
	cache.Add("current", "current-value")
	clock.SetTime(clock.Now().Add(2 * time.Minute))

	g.Expect(cache.ReadAll()).To(ConsistOf("current-value"))
}

func Test_Evict_RemovesCacheEntry(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	clock := newFakePassiveClock()

	cache := New[string, string](
		logr.Discard(),
		WithClock(clock),
		WithTTL(5*time.Minute),
	)

	// Populate the cache
	cache.Add("test-key", "test-value")

	// Verify the cache entry is present before eviction
	value, ok := cache.Read("test-key")
	g.Expect(ok).To(BeTrue())
	g.Expect(value).To(Equal("test-value"))

	// Evict the cache entry
	cache.Evict("test-key")

	// Verify the cache entry is no longer present after eviction
	_, ok = cache.Read("test-key")
	g.Expect(ok).To(BeFalse())
}

func Test_IsEmpty_ReflectsStoredEntries(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cache := New[string, string](logr.Discard())
	g.Expect(cache.IsEmpty()).To(BeTrue())

	cache.Add("test-key", "test-value")
	g.Expect(cache.IsEmpty()).To(BeFalse())

	cache.Evict("test-key")
	g.Expect(cache.IsEmpty()).To(BeTrue())
}

func Test_Cache_WithKeyCanonicalizer_UsesCanonicalKeyForAllOperations(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cache := New[string, string](
		logr.Discard(),
		WithKeyCanonicalizer(strings.ToLower),
	)

	cache.Add("TEST-KEY", "test-value")

	value, ok := cache.Read("test-key")
	g.Expect(ok).To(BeTrue())
	g.Expect(value).To(Equal("test-value"))
	g.Expect(cache.entries).To(HaveKey("test-key"))

	cache.Evict("TEST-KEY")

	_, ok = cache.Read("test-key")
	g.Expect(ok).To(BeFalse())
}

func Test_Cache_WithoutKeyCanonicalizer_PreservesKeyComparison(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cache := New[string, string](logr.Discard())
	cache.Add("TEST-KEY", "test-value")

	_, ok := cache.Read("test-key")
	g.Expect(ok).To(BeFalse())
}

func newFakePassiveClock() *clock.FakePassiveClock {
	now, _ := time.Parse(time.RFC3339, "2026-08-01T15:00:00Z")
	return clock.NewFakePassiveClock(now)
}
