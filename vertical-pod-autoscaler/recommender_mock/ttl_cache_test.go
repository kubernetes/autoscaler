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
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	cache := NewTTLCache(5 * time.Second)
	cacheKey := "hello"
	result := cache.Get(&cacheKey)
	assert.Nil(t, result, "Result should be nil")

	cache.Set(&cacheKey, "world")
	result = cache.Get(&cacheKey)
	assert.Equal(t, "world", result)
}

func TestExpiration(t *testing.T) {
	cache := NewTTLCache(5 * time.Second)
	key1 := "A"
	key2 := "B"
	key3 := "C"

	cache.Set(&key1, "1")
	cache.Set(&key2, "2")
	cache.Set(&key3, "3")
	cache.StartCacheGC(time.Second)

	<-time.After(2 * time.Second)
	assert.Equal(t, "1", cache.Get(&key1))
	cache.Set(&key1, "1")

	<-time.After(3 * time.Second)
	assert.Equal(t, "1", cache.Get(&key1))
	assert.Nil(t, cache.Get(&key2))
	assert.Nil(t, cache.Get(&key3))

	<-time.After(5 * time.Second)
	assert.Nil(t, cache.Get(&key1))

	cache.StopCacheGC()
	cache.Set(&key1, "1")
	<-time.After(6 * time.Second)
	assert.Nil(t, cache.Get(&key1))
}
