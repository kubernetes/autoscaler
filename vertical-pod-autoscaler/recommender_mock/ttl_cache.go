/*
Copyright 2017 The Kubernetes Authors.

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

package recommender

import (
	"sync"
	"time"
)

// TTLCache represents cache with ttl
type TTLCache struct {
	ttl    time.Duration
	mutex  sync.RWMutex
	items  map[string]*cachedValue
	stopGC chan struct{}
}

type cachedValue struct {
	value          interface{}
	expirationTime time.Time
}

// NewTTLCache reates TTLCache for given TTL
func NewTTLCache(ttl time.Duration) *TTLCache {
	return &TTLCache{
		ttl:    ttl,
		items:  make(map[string]*cachedValue),
		stopGC: make(chan struct{})}
}

func (r *cachedValue) isFresh() bool {
	return r.expirationTime.After(time.Now())
}

// Get Returns value present in cache for given cache key, or nil if key is not found or value TTL has expired.
func (c *TTLCache) Get(cacheKey *string) interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	value, found := c.items[*cacheKey]
	if found && value.isFresh() {
		return value.value
	}
	return nil
}

// Set adds given value for given key
func (c *TTLCache) Set(cacheKey *string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.items[*cacheKey] = &cachedValue{value: value, expirationTime: time.Now().Add(c.ttl)}
}

func (c *TTLCache) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for key, item := range c.items {
		if !item.isFresh() {
			delete(c.items, key)
		}
	}
}

// StartCacheGC starts background garbage collector worker which on every time interval removes expired cache entries
// If StartCacheGC was called, in order to properly remove cache object, call StopCacheGC.
// Otherwise TTLCache will never be garbage collected since background worker holds reference to it.
func (c *TTLCache) StartCacheGC(interval time.Duration) {
	ticker := time.Tick(interval)
	go (func() {
		for {
			select {
			case <-ticker:
				c.cleanup()
			case <-c.stopGC:
				return
			}
		}
	})()
}

// StopCacheGC stops background cache garbage collector
func (c *TTLCache) StopCacheGC() {
	c.stopGC <- struct{}{}
}
