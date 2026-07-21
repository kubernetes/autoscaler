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

import "time"

// Entry represents a single cache item.
type item[V any] struct {
	// value is the cached value.
	value V
	// expiry is the time at which the cache entry expires.
	expiry time.Time
}

// ExpiryString returns the expiry time of the cache entry as a formatted string.
// Used for logging
func (i *item[V]) ExpiryString() string {
	return i.expiry.Format(time.RFC3339)
}
