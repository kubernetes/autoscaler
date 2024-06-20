/*
Copyright 2020 The Kubernetes Authors.

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
	"fmt"
	"sync"
	"time"

	"k8s.io/client-go/tools/cache"

	"sigs.k8s.io/cloud-provider-azure/pkg/util/deepcopy"
)

// AzureCacheReadType defines the read type for cache data
type AzureCacheReadType int

const (
	// CacheReadTypeDefault returns data from cache if cache entry not expired
	// if cache entry expired, then it will refetch the data using getter
	// save the entry in cache and then return
	CacheReadTypeDefault AzureCacheReadType = iota
	// CacheReadTypeUnsafe returns data from cache even if the cache entry is
	// active/expired. If entry doesn't exist in cache, then data is fetched
	// using getter, saved in cache and returned
	CacheReadTypeUnsafe
	// CacheReadTypeForceRefresh force refreshes the cache even if the cache entry
	// is not expired
	CacheReadTypeForceRefresh
)

// GetFunc defines a getter function for timedCache.
type GetFunc func(key string) (interface{}, error)

// AzureCacheEntry is the internal structure stores inside TTLStore.
type AzureCacheEntry struct {
	Key  string
	Data interface{}

	// The lock to ensure not updating same entry simultaneously.
	Lock sync.Mutex
	// time when entry was fetched and created
	CreatedOn time.Time
}

// cacheKeyFunc defines the key function required in TTLStore.
func cacheKeyFunc(obj interface{}) (string, error) {
	return obj.(*AzureCacheEntry).Key, nil
}

// Resource operations
type Resource interface {
	Get(key string, crt AzureCacheReadType) (interface{}, error)
	GetWithDeepCopy(key string, crt AzureCacheReadType) (interface{}, error)
	Delete(key string) error
	Set(key string, data interface{})
	Update(key string, data interface{})

	GetStore() cache.Store
	Lock()
	Unlock()
}

// TimedCache is a cache with TTL.
type TimedCache struct {
	Store     cache.Store
	MutexLock sync.RWMutex
	TTL       time.Duration

	resourceProvider Resource
}

type ResourceProvider struct {
	Getter GetFunc
}

// NewTimedCache creates a new azcache.Resource.
func NewTimedCache(ttl time.Duration, getter GetFunc, disabled bool) (Resource, error) {
	if getter == nil {
		return nil, fmt.Errorf("getter is not provided")
	}

	provider := &ResourceProvider{
		Getter: getter,
	}

	if disabled {
		return provider, nil
	}

	timedCache := &TimedCache{
		// switch to using NewStore instead of NewTTLStore so that we can
		// reuse entries for calls that are fine with reading expired/stalled data.
		// with NewTTLStore, entries are not returned if they have already expired.
		Store:            cache.NewStore(cacheKeyFunc),
		MutexLock:        sync.RWMutex{},
		TTL:              ttl,
		resourceProvider: provider,
	}
	return timedCache, nil
}

// getInternal returns AzureCacheEntry by key. If the key is not cached yet,
// it returns a AzureCacheEntry with nil data.
func (t *TimedCache) getInternal(key string) (*AzureCacheEntry, error) {
	entry, exists, err := t.Store.GetByKey(key)
	if err != nil {
		return nil, err
	}
	// if entry exists, return the entry
	if exists {
		return entry.(*AzureCacheEntry), nil
	}

	// lock here to ensure if entry doesn't exist, we add a new entry
	// avoiding overwrites
	t.Lock()
	defer t.Unlock()

	// Another goroutine might have written the same key.
	entry, exists, err = t.Store.GetByKey(key)
	if err != nil {
		return nil, err
	}
	if exists {
		return entry.(*AzureCacheEntry), nil
	}

	// Still not found, add new entry with nil data.
	// Note the data will be filled later by getter.
	newEntry := &AzureCacheEntry{
		Key:  key,
		Data: nil,
	}
	_ = t.Store.Add(newEntry)
	return newEntry, nil
}

// Get returns the requested item by key.
func (t *TimedCache) Get(key string, crt AzureCacheReadType) (interface{}, error) {
	return t.get(key, crt)
}

func (c *ResourceProvider) Get(key string, _ AzureCacheReadType) (interface{}, error) {
	return c.Getter(key)
}

// Get returns the requested item by key with deep copy.
func (t *TimedCache) GetWithDeepCopy(key string, crt AzureCacheReadType) (interface{}, error) {
	data, err := t.get(key, crt)
	copied := deepcopy.Copy(data)
	return copied, err
}

func (c *ResourceProvider) GetWithDeepCopy(key string, _ AzureCacheReadType) (interface{}, error) {
	return c.Getter(key)
}

func (t *TimedCache) get(key string, crt AzureCacheReadType) (interface{}, error) {
	entry, err := t.getInternal(key)
	if err != nil {
		return nil, err
	}

	entry.Lock.Lock()
	defer entry.Lock.Unlock()

	// entry exists and if cache is not force refreshed
	if entry.Data != nil && crt != CacheReadTypeForceRefresh {
		// allow unsafe read, so return data even if expired
		if crt == CacheReadTypeUnsafe {
			return entry.Data, nil
		}
		// if cached data is not expired, return cached data
		if crt == CacheReadTypeDefault && time.Since(entry.CreatedOn) < t.TTL {
			return entry.Data, nil
		}
	}
	// Data is not cached yet, cache data is expired or requested force refresh
	// cache it by getter. entry is locked before getting to ensure concurrent
	// gets don't result in multiple ARM calls.
	data, err := t.resourceProvider.Get(key, CacheReadTypeDefault /* not matter */)
	if err != nil {
		return nil, err
	}

	// set the data in cache and also set the last update time
	// to now as the data was recently fetched
	entry.Data = data
	entry.CreatedOn = time.Now().UTC()

	return entry.Data, nil
}

// Delete removes an item from the cache.
func (t *TimedCache) Delete(key string) error {
	return t.Store.Delete(&AzureCacheEntry{
		Key: key,
	})
}

func (c *ResourceProvider) Delete(_ string) error {
	return nil
}

// Set sets the data cache for the key.
// It is only used for testing.
func (t *TimedCache) Set(key string, data interface{}) {
	_ = t.Store.Add(&AzureCacheEntry{
		Key:       key,
		Data:      data,
		CreatedOn: time.Now().UTC(),
	})
}

func (c *ResourceProvider) Set(_ string, _ interface{}) {}

// Update updates the data cache for the key.
func (t *TimedCache) Update(key string, data interface{}) {
	if entry, err := t.getInternal(key); err == nil {
		entry.Lock.Lock()
		defer entry.Lock.Unlock()
		entry.Data = data
		entry.CreatedOn = time.Now().UTC()
	} else {
		_ = t.Store.Update(&AzureCacheEntry{
			Key:       key,
			Data:      data,
			CreatedOn: time.Now().UTC(),
		})
	}
}

func (c *ResourceProvider) Update(_ string, _ interface{}) {}

func (t *TimedCache) GetStore() cache.Store {
	return t.Store
}

func (c *ResourceProvider) GetStore() cache.Store {
	return nil
}

func (t *TimedCache) Lock() {
	t.MutexLock.Lock()
}

func (t *TimedCache) Unlock() {
	t.MutexLock.Unlock()
}

func (c *ResourceProvider) Lock() {}

func (c *ResourceProvider) Unlock() {}
