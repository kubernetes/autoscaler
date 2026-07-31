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
	"time"

	"github.com/cenkalti/backoff/v4"
	"k8s.io/apimachinery/pkg/types"
)

const (
	baseBackoff      = 5 * time.Minute
	backoffThreshold = 30 * time.Minute
)

// controllerEntry describes a backed off controller
type controllerEntry struct {
	until   time.Time
	backoff backoff.ExponentialBackOff
}

// ControllerRegistry contains backed off controllers to be used in time-based backing off of controllers considered in fake pod injection
type ControllerRegistry struct {
	backedOffControllers map[types.UID]controllerEntry
}

// NewFakePodControllerRegistry Creates & returns an instance of fakePodControllerBackoffRegistry
func NewFakePodControllerRegistry() *ControllerRegistry {
	return &ControllerRegistry{
		backedOffControllers: make(map[types.UID]controllerEntry),
	}
}

// newExponentialBackOff creates an instance of ExponentialBackOff using non-default values.
func newExponentialBackOff(clock backoff.Clock) backoff.ExponentialBackOff {
	b := backoff.ExponentialBackOff{
		InitialInterval: baseBackoff,
		// Disables randomization for easier testing and better predictability
		RandomizationFactor: 0,
		Multiplier:          backoff.DefaultMultiplier,
		MaxInterval:         backoffThreshold,
		// Disable stopping if it reaches threshold
		MaxElapsedTime: 0,
		Stop:           backoff.Stop,
		Clock:          clock,
	}
	b.Reset()
	return b
}

// BackoffController Backs off a controller
// If the controller is already in backoff it's backoff time is exponentially increased
// If the controller was in backoff, it resets its entry and makes it in backoff
// If the controller is not in backoff and not stored, a new entry is created
func (r *ControllerRegistry) BackoffController(ownerUID types.UID, now time.Time) {
	if ownerUID == "" {
		return
	}

	controller, found := r.backedOffControllers[ownerUID]

	if !found || now.After(controller.until) {
		controller = controllerEntry{
			backoff: newExponentialBackOff(backoff.SystemClock),
		}
	}
	// NextBackOff() needs to be called to increase the next interval
	controller.until = now.Add(controller.backoff.NextBackOff())

	r.backedOffControllers[ownerUID] = controller
}

// BackOffUntil Returns the back off status a controller with id `uid`
func (r *ControllerRegistry) BackOffUntil(uid types.UID, now time.Time) time.Time {
	controller, found := r.backedOffControllers[uid]

	if !found {
		return time.Time{}
	}

	return controller.until
}
