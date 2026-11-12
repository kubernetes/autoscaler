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
	"sync"

	"github.com/go-logr/logr"
)

// Cache is a generic cache implementation that stores key-value pairs
// with optional key canonicalization and logging support.
// Cache expiry is handled outside of this implementation.
type Cache[K comparable, V any] struct {
	// entries is a map of cache entries, keyed by the cache key.
	entries map[K]V

	// keyCanonicalizer transforms keys before they are used to access entries.
	keyCanonicalizer func(K) K

	// log is a logger for logging cache operations.
	log logr.Logger

	// mutex is a read-write mutex to protect concurrent access to the cache.
	mutex sync.RWMutex
}

// New creates a new Cache instance.
// K is the type of the cache key.
// V is the type of the cache value.
// log is a logger for logging cache operations.
// options are functional options for configuring the cache.
func New[K comparable, V any](
	log logr.Logger,
	options ...Option,
) *Cache[K, V] {
	cfg := &config{
		keyCanonicalizer: func(key K) K { return key },
	}

	for _, option := range options {
		option(cfg)
	}

	return &Cache[K, V]{
		entries:          make(map[K]V),
		keyCanonicalizer: cfg.keyCanonicalizer.(func(K) K),
		log:              log,
	}
}

// Read returns the cached value for the given key.
func (c *Cache[K, V]) Read(key K) (V, bool) {
	canonicalKey := c.keyCanonicalizer(key)

	// Check if the entry is present
	entry, entryFound := c.lookupEntry(canonicalKey)
	if entryFound {
		return entry, true
	}

	var zero V
	return zero, false
}

// ReadAll returns all cached values. The returned values have no guaranteed order.
func (c *Cache[K, V]) ReadAll() []V {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	values := make([]V, 0, len(c.entries))
	for _, entry := range c.entries {
		values = append(values, entry)
	}

	return values
}

// Add adds a new entry to the cache with the given key and value.
// If an entry with the same key already exists, it will be overwritten.
// key is the cache key.
// value is the value to be cached.
func (c *Cache[K, V]) Add(key K, value V) {
	canonicalKey := c.keyCanonicalizer(key)

	c.log.V(4).Info(
		"Adding cache entry",
		"key", canonicalKey)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries[canonicalKey] = value
}

// AddAll adds all entries to the cache. Entries with existing keys are overwritten.
func (c *Cache[K, V]) AddAll(entries map[K]V) {
	toCache := c.newEntries(entries)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	for key, entry := range toCache {
		c.entries[key] = entry
	}
}

// ReplaceAll replaces all cache entries with the given entries.
func (c *Cache[K, V]) ReplaceAll(entries map[K]V) {
	toCache := c.newEntries(entries)

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries = toCache
}

// IsEmpty reports whether the cache has no stored entries.
func (c *Cache[K, V]) IsEmpty() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.entries) == 0
}

// Evict removes the cache entry (if present)
func (c *Cache[K, V]) Evict(key K) {
	c.evict(c.keyCanonicalizer(key))
}

func (c *Cache[K, V]) evict(key K) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.entries, key)
}

func (c *Cache[K, V]) newEntries(entries map[K]V) map[K]V {
	toCache := make(map[K]V, len(entries))

	for key, value := range entries {
		toCache[c.keyCanonicalizer(key)] = value
	}

	return toCache
}

// lookupEntry looks up the cache entry for the given key.
// key is the cache key.
// Returns the cached value (if present) and a boolean indicating whether the entry was found.
func (c *Cache[K, V]) lookupEntry(key K) (V, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	i, ok := c.entries[key]
	return i, ok
}
