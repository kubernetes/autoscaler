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

package aksk

import "sync"

// MemoryCache presents a thread safe memory cache
type MemoryCache struct {
	sync.Mutex                    // handling r/w for cache
	cacheHolder map[string]string // cache holder
	cacheKeys   []string          // cache keys
	MaxCount    int               // max cache entry count
}

// NewCache inits an new MemoryCache
func NewCache(maxCount int) *MemoryCache {
	return &MemoryCache{
		cacheHolder: make(map[string]string, maxCount),
		MaxCount:    maxCount,
	}
}

// Add an new cache item
func (cache *MemoryCache) Add(cacheKey string, cacheData string) {
	cache.Lock()
	defer cache.Unlock()

	if len(cache.cacheKeys) >= cache.MaxCount && len(cache.cacheKeys) > 1 {
		delete(cache.cacheHolder, cache.cacheKeys[0]) // delete first item
		cache.cacheKeys = append(cache.cacheKeys[1:]) // pop first one
	}

	cache.cacheHolder[cacheKey] = cacheData
	cache.cacheKeys = append(cache.cacheKeys, cacheKey)
}

// Get a cache item by its key
func (cache *MemoryCache) Get(cacheKey string) string {
	cache.Lock()
	defer cache.Unlock()

	return cache.cacheHolder[cacheKey]
}
