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
	"time"

	"k8s.io/utils/clock"
)

// Option is a functional option for configuring the cache.
type Option func(*config)

// WithClock sets the clock implementation to use for time-based operations in the cache.
func WithClock(clk clock.PassiveClock) Option {
	return func(c *config) {
		c.clock = clk
	}
}

// WithTTL sets the time-to-live for cache entries.
func WithTTL(ttl time.Duration) Option {
	return func(c *config) {
		c.ttl = ttl
	}
}

// WithKeyCanonicalizer sets the function used to transform keys before cache operations.
func WithKeyCanonicalizer[K comparable](canonicalize func(K) K) Option {
	return func(c *config) {
		c.keyCanonicalizer = canonicalize
	}
}
