/*
Copyright The Kubernetes Authors.

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

package pretty

import (
	"time"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/patrickmn/go-cache"
)

// ChangeMonitor is used to reduce logging when discovering information that may change. The values recorded expire after
// 24 hours by default to prevent a value from being logged at startup only which could impede debugging if full sets
// of logs aren't available.
type ChangeMonitor struct {
	lastSeen *cache.Cache
}

func NewChangeMonitor() *ChangeMonitor {
	return &ChangeMonitor{
		lastSeen: cache.New(24*time.Hour, 12*time.Hour),
	}
}

// HasChanged takes a key and value and returns true if the hash of the value has changed since the last tine the
// change monitor was called.
func (c *ChangeMonitor) HasChanged(key string, value any) bool {
	hv, _ := hashstructure.Hash(value, hashstructure.FormatV2, &hashstructure.HashOptions{SlicesAsSets: true})
	existing, ok := c.lastSeen.Get(key)
	var existingHash uint64
	if ok {
		existingHash = existing.(uint64)
	}
	if !ok || existingHash != hv {
		c.lastSeen.SetDefault(key, hv)
		return true
	}
	return false
}
