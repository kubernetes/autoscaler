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

	"github.com/go-logr/logr"

	. "github.com/onsi/gomega"
)

func Test_Add_cachesEntry_withExpectedExpiry(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Create a new cache
	cache := New[string, string](logr.Discard())

	// Add an entry to the cache
	cache.Add("test-key", "test-value")

	// The cache should now contain the expected
	entry, found := cache.entries["test-key"]
	g.Expect(found).To(BeTrue())
	g.Expect(entry.value).To(Equal("test-value"))
}

func Test_AddAll_MergesEntriesAndOverwritesExistingEntries(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Create a new cache with a known initial entry
	cache := New[string, string](logr.Discard())
	cache.Add("existing", "old-value")

	// Add new entries to the cache
	entriesToAdd := map[string]string{
		"existing": "new-value",
		"new":      "new-entry",
	}
	cache.AddAll(entriesToAdd)

	// Verify the existing value was overwritten
	value, found := cache.Read("existing")
	g.Expect(found).To(BeTrue())
	g.Expect(value).To(Equal("new-value"))

	// Verify the new entry was added
	value, found = cache.Read("new")
	g.Expect(found).To(BeTrue())
	g.Expect(value).To(Equal("new-entry"))
}

func Test_ReplaceAll_ReplacesExistingEntries(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Create a new cache with some initial entries
	cache := New[string, string](logr.Discard())
	cache.Add("removed", "old-entry")
	cache.Add("retained", "old-value")

	// Replace cache entries with the new set of entries
	replacementEntries := map[string]string{
		"retained": "new-value",
		"new":      "new-entry",
	}
	cache.ReplaceAll(replacementEntries)

	// Verify the removed entry is no longer in the cache
	_, found := cache.Read("removed")
	g.Expect(found).To(BeFalse())

	// Verify the retained entry has the new value
	value, found := cache.Read("retained")
	g.Expect(found).To(BeTrue())
	g.Expect(value).To(Equal("new-value"))

	// Verify the new entry was added
	value, found = cache.Read("new")
	g.Expect(found).To(BeTrue())
	g.Expect(value).To(Equal("new-entry"))
}

func Test_AddAllAndReplaceAll_CanonicalizeKeys(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Create a new cache with a key canonicalizer that converts keys to lowercase
	cache := New[string, string](
		logr.Discard(),
		WithKeyCanonicalizer(strings.ToLower),
	)

	// Add entries using AddAll and verify they are stored with canonicalized keys
	entriesToAdd := map[string]string{
		"ADD-ALL": "Value",
	}
	cache.AddAll(entriesToAdd)
	entry, found := cache.entries["add-all"]
	g.Expect(found).To(BeTrue())
	g.Expect(entry.value).To(Equal("Value")) // Values should not be canonicalized

	// Replace entries using ReplaceAll and verify they are stored with canonicalized keys
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
}

func Test_Read_WhenCachePopulated_ReturnsCachedItem(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	// Create a new cache
	cache := New[string, string](logr.Discard())

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

func Test_Evict_RemovesCacheEntry(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	cache := New[string, string](logr.Discard())

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

	// Empty cache is empty
	cache := New[string, string](logr.Discard())
	g.Expect(cache.IsEmpty()).To(BeTrue())

	// Adding an entry should make the cache non-empty
	cache.Add("test-key", "test-value")
	g.Expect(cache.IsEmpty()).To(BeFalse())

	// Evicting the entry should make the cache empty again
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
