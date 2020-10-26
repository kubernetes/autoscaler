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

package controllerfetcher

import (
	"sync"
	"time"

	autoscalingapi "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
)

// Allows tests to inject their time.
var now = time.Now

type scaleCacheKey struct {
	namespace     string
	groupResource schema.GroupResource
	name          string
}

type scaleCacheEntry struct {
	refreshAfter time.Time
	deleteAfter  time.Time
	resource     *autoscalingapi.Scale
	err          error
}

// Cache for responses to get queries on controllers. Thread safe.
// Usage:
// - `Get` cached response. If there is one use it, otherwise make query and
// - `Insert` the response you got into the cache.
// When you create a `controllerCacheStorage` you should start two go routines:
// - One for refreshing cache entries, which calls `GetKeysToRefresh` then for
//   each key makes query to the API server and calls `Refresh` to update
//  content of the cache.
// - Second for removing stale entries which periodically calls `RemoveExpired`
// Each entry is refreshed after duration
// `validityTime` * (1 + `jitterFactor`)
// passes and is removed if there are no reads for it for more than `lifeTime`.
//
// Sometimes refreshing might take longer than refreshAfter (for example when
// VPA is starting in a big cluster and tries to fetch all controllers). To
// handle such situation lifeTime should be longer than refreshAfter so the main
// VPA loop can do its work quickly, using the cached information (instead of
// getting stuck on refreshing the cache).
// TODO(jbartosik): Add a way to detect when we don't refresh cache frequently
// enough.
// TODO(jbartosik): Add a way to learn how long we keep entries around so we can
// decide if / how we want to optimize entry refreshes.
type controllerCacheStorage struct {
	cache        map[scaleCacheKey]scaleCacheEntry
	mux          sync.Mutex
	validityTime time.Duration
	jitterFactor float64
	lifeTime     time.Duration
}

// Returns bool indicating whether the entry was present in the cache and the cached response.
// Updates deleteAfter for the element.
func (cc *controllerCacheStorage) Get(namespace string, groupResource schema.GroupResource, name string) (ok bool, controller *autoscalingapi.Scale, err error) {
	key := scaleCacheKey{namespace: namespace, groupResource: groupResource, name: name}
	cc.mux.Lock()
	defer cc.mux.Unlock()
	r, ok := cc.cache[key]
	if ok {
		r.deleteAfter = now().Add(cc.lifeTime)
		cc.cache[key] = r
	}
	return ok, r.resource, r.err
}

// If key is in the cache, refresh updates the cached value, error and refresh
// time (but not time to remove).
// If the key is missing from the cache does nothing (relevant when we're
// concurrently updating cache and removing stale entries from it, to avoid
// adding back an entry which we just removed).
func (cc *controllerCacheStorage) Refresh(namespace string, groupResource schema.GroupResource, name string, controller *autoscalingapi.Scale, err error) {
	key := scaleCacheKey{namespace: namespace, groupResource: groupResource, name: name}
	cc.mux.Lock()
	defer cc.mux.Unlock()
	old, ok := cc.cache[key]
	if !ok {
		return
	}
	// We refresh entries that are waiting to be removed. So when we refresh an
	// entry we mustn't change entries deleteAfter time (otherwise we risk never
	// removing entries that are not being read).
	cc.cache[key] = scaleCacheEntry{
		refreshAfter: now().Add(wait.Jitter(cc.validityTime, cc.jitterFactor)),
		deleteAfter:  old.deleteAfter,
		resource:     controller,
		err:          err,
	}
}

// If the key is missing from the cache, updates the cached value, error and refresh time (but not deleteAfter time).
// If key is in the cache, does nothing (to make sure updating element doesn't change its deleteAfter time).
func (cc *controllerCacheStorage) Insert(namespace string, groupResource schema.GroupResource, name string, controller *autoscalingapi.Scale, err error) {
	key := scaleCacheKey{namespace: namespace, groupResource: groupResource, name: name}
	cc.mux.Lock()
	defer cc.mux.Unlock()
	if _, ok := cc.cache[key]; ok {
		return
	}
	now := now()
	cc.cache[key] = scaleCacheEntry{
		refreshAfter: now.Add(wait.Jitter(cc.validityTime, cc.jitterFactor)),
		deleteAfter:  now.Add(cc.lifeTime),
		resource:     controller,
		err:          err,
	}
}

// Removes entries which we didn't read in a while from the cache.
func (cc *controllerCacheStorage) RemoveExpired() {
	klog.V(1).Info("Removing entries from controllerCacheStorage")
	cc.mux.Lock()
	defer cc.mux.Unlock()
	now := now()
	removed := 0
	for k, v := range cc.cache {
		if now.After(v.deleteAfter) {
			removed += 1
			delete(cc.cache, k)
		}
	}
	klog.V(1).Infof("Removed %d entries from controllerCacheStorage", removed)
}

// Returns a list of keys for which values need to be refreshed
func (cc *controllerCacheStorage) GetKeysToRefresh() []scaleCacheKey {
	result := make([]scaleCacheKey, 0)
	cc.mux.Lock()
	defer cc.mux.Unlock()
	now := now()
	for k, v := range cc.cache {
		if now.After(v.refreshAfter) {
			result = append(result, k)
		}
	}
	return result
}

func newControllerCacheStorage(validityTime, lifeTime time.Duration, jitterFactor float64) controllerCacheStorage {
	return controllerCacheStorage{
		cache:        make(map[scaleCacheKey]scaleCacheEntry),
		validityTime: validityTime,
		jitterFactor: jitterFactor,
		lifeTime:     lifeTime,
	}
}
