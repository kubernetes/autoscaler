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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	autoscalingapi "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func getKey(key string) scaleCacheKey {
	return scaleCacheKey{
		namespace: "ns",
		groupResource: schema.GroupResource{
			Group:    "group",
			Resource: "resource",
		},
		name: key,
	}
}

func getScale() *autoscalingapi.Scale {
	return &autoscalingapi.Scale{}
}

func TestControllerCache_InitiallyNotPresent(t *testing.T) {
	c := newControllerCacheStorage(time.Second, 10*time.Second, 1)
	key := getKey("foo")
	present, _, _ := c.Get(key.namespace, key.groupResource, key.name)
	assert.False(t, present)
}

func TestControllerCache_Refresh_NotExisting(t *testing.T) {
	key := getKey("foo")
	c := newControllerCacheStorage(time.Second, 10*time.Second, 1)
	present, _, _ := c.Get(key.namespace, key.groupResource, key.name)
	assert.False(t, present)

	// Refreshing key that isn't in the cache doesn't insert it
	c.Refresh(key.namespace, key.groupResource, key.name, getScale(), nil)
	present, _, _ = c.Get(key.namespace, key.groupResource, key.name)
	assert.False(t, present)
}

func TestControllerCache_Insert(t *testing.T) {
	key := getKey("foo")
	c := newControllerCacheStorage(time.Second, 10*time.Second, 1)
	present, _, _ := c.Get(key.namespace, key.groupResource, key.name)
	assert.False(t, present)

	c.Insert(key.namespace, key.groupResource, key.name, getScale(), nil)
	present, val, err := c.Get(key.namespace, key.groupResource, key.name)
	assert.True(t, present)
	assert.Equal(t, getScale(), val)
	assert.Nil(t, err)
}

func TestControllerCache_InsertAndRefresh(t *testing.T) {
	key := getKey("foo")
	c := newControllerCacheStorage(time.Second, 10*time.Second, 1)
	present, _, _ := c.Get(key.namespace, key.groupResource, key.name)
	assert.False(t, present)

	c.Insert(key.namespace, key.groupResource, key.name, getScale(), nil)
	present, val, err := c.Get(key.namespace, key.groupResource, key.name)
	assert.True(t, present)
	assert.Equal(t, getScale(), val)
	assert.Nil(t, err)

	c.Refresh(key.namespace, key.groupResource, key.name, nil, fmt.Errorf("err"))
	present, val, err = c.Get(key.namespace, key.groupResource, key.name)
	assert.True(t, present)
	assert.Nil(t, val)
	assert.Errorf(t, err, "err")
}

func TestControllerCache_InsertExistingKey(t *testing.T) {
	key := getKey("foo")
	c := newControllerCacheStorage(time.Second, 10*time.Second, 1)
	present, _, _ := c.Get(key.namespace, key.groupResource, key.name)
	assert.False(t, present)

	c.Insert(key.namespace, key.groupResource, key.name, getScale(), nil)
	present, val, err := c.Get(key.namespace, key.groupResource, key.name)
	assert.True(t, present)
	assert.Equal(t, getScale(), val)
	assert.Nil(t, err)

	// We might overwrite old values or keep them, either way should be fine.
	c.Insert(key.namespace, key.groupResource, key.name, nil, fmt.Errorf("err"))
	present, _, _ = c.Get(key.namespace, key.groupResource, key.name)
	assert.True(t, present)
}

func TestControllerCache_GetRefreshesDeleteAfter(t *testing.T) {
	oldNow := now
	defer func() { now = oldNow }()
	startTime := oldNow()
	timeNow := startTime
	now = func() time.Time {
		return timeNow
	}

	key := getKey("foo")
	c := newControllerCacheStorage(time.Second, 10*time.Second, 1)
	c.Insert(key.namespace, key.groupResource, key.name, nil, nil)
	assert.Equal(t, startTime.Add(10*time.Second), c.cache[key].deleteAfter)

	timeNow = startTime.Add(5 * time.Second)
	c.Get(key.namespace, key.groupResource, key.name)
	assert.Equal(t, startTime.Add(15*time.Second), c.cache[key].deleteAfter)
}

func assertTimeBetween(t *testing.T, got, expectAfter, expectBefore time.Time) {
	assert.True(t, got.After(expectAfter), "expected %v to be after %v", got, expectAfter)
	assert.False(t, got.After(expectBefore), "expected %v to not be after %v", got, expectBefore)
}

func TestControllerCache_GetChangesLifeTimeNotFreshness(t *testing.T) {
	oldNow := now
	defer func() { now = oldNow }()
	startTime := oldNow()
	timeNow := startTime
	now = func() time.Time {
		return timeNow
	}

	key := getKey("foo")
	c := newControllerCacheStorage(time.Second, 10*time.Second, 1)
	c.Insert(key.namespace, key.groupResource, key.name, nil, nil)
	cacheEntry := c.cache[key]
	// scheduled to delete 10s after insert
	assert.Equal(t, startTime.Add(10*time.Second), cacheEntry.deleteAfter)
	// scheduled to refresh (1-2)s after insert (with jitter)
	firstRefreshAfter := cacheEntry.refreshAfter
	assertTimeBetween(t, firstRefreshAfter, startTime.Add(time.Second), startTime.Add(2*time.Second))

	timeNow = startTime.Add(5 * time.Second)
	c.Get(key.namespace, key.groupResource, key.name)
	cacheEntry = c.cache[key]
	// scheduled to delete 10s after get (15s after insert)
	assert.Equal(t, startTime.Add(15*time.Second), cacheEntry.deleteAfter)
	// refresh the same as before calling Get
	assert.Equal(t, firstRefreshAfter, cacheEntry.refreshAfter)
}

func TestControllerCache_GetKeysToRefresh(t *testing.T) {
	oldNow := now
	defer func() { now = oldNow }()
	startTime := oldNow()
	timeNow := startTime
	now = func() time.Time {
		return timeNow
	}

	key1 := getKey("foo")
	c := newControllerCacheStorage(time.Second, 10*time.Second, 1)
	c.Insert(key1.namespace, key1.groupResource, key1.name, nil, nil)
	cacheEntry := c.cache[key1]
	// scheduled to refresh (1-2)s after insert (with jitter)
	refreshAfter := cacheEntry.refreshAfter
	assertTimeBetween(t, refreshAfter, startTime.Add(time.Second), startTime.Add(2*time.Second))

	timeNow = startTime.Add(5 * time.Second)
	key2 := getKey("bar")
	c.Insert(key2.namespace, key2.groupResource, key2.name, nil, nil)
	cacheEntry = c.cache[key2]
	// scheduled to refresh (1-2)s after insert (with jitter)
	refreshAfter = cacheEntry.refreshAfter
	assertTimeBetween(t, refreshAfter, startTime.Add(6*time.Second), startTime.Add(7*time.Second))

	assert.ElementsMatch(t, []scaleCacheKey{key1}, c.GetKeysToRefresh())
}

func TestControllerCache_Clear(t *testing.T) {
	oldNow := now
	defer func() { now = oldNow }()
	startTime := oldNow()
	timeNow := startTime
	now = func() time.Time {
		return timeNow
	}

	key1 := getKey("foo")
	c := newControllerCacheStorage(time.Second, 10*time.Second, 1)
	c.Insert(key1.namespace, key1.groupResource, key1.name, nil, nil)
	assert.Equal(t, startTime.Add(10*time.Second), c.cache[key1].deleteAfter)

	timeNow = startTime.Add(15 * time.Second)
	key2 := getKey("bar")
	c.Insert(key2.namespace, key2.groupResource, key2.name, nil, nil)
	assert.Equal(t, startTime.Add(25*time.Second), c.cache[key2].deleteAfter)

	c.RemoveExpired()
	assert.Equal(t, 1, len(c.cache))
	assert.Contains(t, c.cache, key2)
}
