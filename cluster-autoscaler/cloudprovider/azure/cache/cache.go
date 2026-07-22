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
	"time"

	"github.com/go-logr/logr"
	"k8s.io/utils/clock"
)

type Cache[K comparable, V any] struct {
	// entries is a map of cache entries, keyed by the cache key.
	entries map[K]*item[V]

	// keyCanonicalizer transforms keys before they are used to access entries.
	keyCanonicalizer func(K) K

	// ttl is the time-to-live for cache entries.
	ttl time.Duration

	// Clock implementation to use for time-based operations.
	// Overridden in tests to use a fake clock.
	clock clock.PassiveClock

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
		clock:            clock.RealClock{},
		ttl:              5 * time.Minute,
		keyCanonicalizer: func(key K) K { return key },
	}

	for _, option := range options {
		option(cfg)
	}

	return &Cache[K, V]{
		entries:          make(map[K]*item[V]),
		keyCanonicalizer: cfg.keyCanonicalizer.(func(K) K),
		log:              log,
		clock:            cfg.clock,
		ttl:              cfg.ttl,
	}
}

// Read returns the cached value for the given key, or loads it if not present or expired.
func (c *Cache[K, V]) Read(key K) (V, bool) {
	now := c.clock.Now()
	canonicalKey := c.keyCanonicalizer(key)

	// Check if the entry is present and not expired.
	entry, entryFound := c.lookupEntry(canonicalKey)
	if entryFound {
		if now.Before(entry.expiry) {
			c.log.V(4).Info("Cache hit",
				"key", key,
				"expiry", entry.ExpiryString())
			return entry.value, true
		}

		// Entry is expired, evict it.
		c.log.V(4).Info("Cache entry expired, evicting",
			"key", canonicalKey,
			"expiry", entry.ExpiryString())
		c.evict(canonicalKey)
	}

	var zero V
	return zero, false
}

// Add adds a new entry to the cache with the given key and value, and sets its expiry time.
// If an entry with the same key already exists, it will be overwritten.
// key is the cache key.
// value is the value to be cached.
// now is the current time, used to calculate the expiry time.
func (c *Cache[K, V]) Add(key K, value V) {
	canonicalKey := c.keyCanonicalizer(key)
	toCache := item[V]{
		value:  value,
		expiry: c.clock.Now().Add(c.ttl),
	}

	c.log.V(4).Info(
		"Adding cache entry",
		"key", canonicalKey,
		"expiry", toCache.ExpiryString())

	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.entries[canonicalKey] = &toCache
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

// Evict removes the cache entry (if present)
func (c *Cache[K, V]) Evict(key K) {
	c.evict(c.keyCanonicalizer(key))
}

func (c *Cache[K, V]) evict(key K) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.entries, key)
}

func (c *Cache[K, V]) newEntries(entries map[K]V) map[K]*item[V] {
	expiry := c.clock.Now().Add(c.ttl)
	toCache := make(map[K]*item[V], len(entries))

	for key, value := range entries {
		toCache[c.keyCanonicalizer(key)] = &item[V]{
			value:  value,
			expiry: expiry,
		}
	}

	return toCache
}

// lookupEntry looks up the cache entry for the given key.
// key is the cache key.
// Returns the cached value (if present) and a boolean indicating whether the entry was found.
func (c *Cache[K, V]) lookupEntry(key K) (*item[V], bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	i, ok := c.entries[key]
	return i, ok
}
