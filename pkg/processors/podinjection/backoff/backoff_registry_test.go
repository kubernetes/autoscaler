/*
Copyright 2024 The Kubernetes Authors.

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

package podinjectionbackoff

import (
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
)

func TestBackoffControllerOfPod(t *testing.T) {
	c1 := types.UID("c1")
	c2 := types.UID("c2")
	clock := &clock{}

	testCases := map[string]struct {
		backoffCounts                map[types.UID]int
		spendTime                    time.Duration
		expectedBackedoffControllers map[types.UID]controllerEntry
	}{
		"backing-off a controller adds its controller UID in backoff correctly": {
			backoffCounts: map[types.UID]int{
				c1: 1,
			},
			expectedBackedoffControllers: map[types.UID]controllerEntry{
				c1: {
					until: clock.now.Add(baseBackoff),
				},
			},
		},
		"backing-off an already backed-off controller exponentially increases backoff duration": {
			backoffCounts: map[types.UID]int{
				c1: 2,
			},
			expectedBackedoffControllers: map[types.UID]controllerEntry{
				c1: {
					until: clock.now.Add(time.Duration(float64(baseBackoff) * backoff.DefaultMultiplier)),
				},
			},
		},
		"backing-off a controller doesn't affect other controllers": {
			backoffCounts: map[types.UID]int{
				c1: 1,
				c2: 2,
			},
			expectedBackedoffControllers: map[types.UID]controllerEntry{
				c1: {
					until: clock.now.Add(baseBackoff),
				},
				c2: {
					until: clock.now.Add(time.Duration(float64(baseBackoff) * backoff.DefaultMultiplier)),
				},
			},
		},
		"backing-off a past backed-off controller resets backoff": {
			backoffCounts: map[types.UID]int{
				c1: 1,
			},
			spendTime: baseBackoff * 2,
			expectedBackedoffControllers: map[types.UID]controllerEntry{
				c1: {
					until: clock.now.Add(baseBackoff * 2).Add(baseBackoff),
				},
			},
		},
		"back-off duration doesn't exceed backoffThreshold": {
			backoffCounts: map[types.UID]int{
				c1: 15,
			},
			expectedBackedoffControllers: map[types.UID]controllerEntry{
				c1: {
					until: clock.now.Add(backoffThreshold),
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Reset time between test cases
			clock.now = time.Time{}
			clock.now = clock.now.Add(tc.spendTime)

			registry := NewFakePodControllerRegistry()

			for uid, backoffCount := range tc.backoffCounts {
				for i := 0; i < backoffCount; i++ {
					registry.BackoffController(uid, clock.now)
				}
			}

			assert.Equal(t, len(registry.backedOffControllers), len(tc.expectedBackedoffControllers))
			for uid, backoffController := range tc.expectedBackedoffControllers {
				assert.NotNil(t, registry.backedOffControllers[uid])
				assert.Equal(t, backoffController.until, registry.backedOffControllers[uid].until)
			}
		})
	}
}

type clock struct {
	now time.Time
}

func (c *clock) Now() time.Time {
	return c.now
}
