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

package backoff

import (
	"time"
)

type backoffInfo struct {
	duration            time.Duration
	backoffUntil        time.Time
	lastFailedExecution time.Time
}

// Backoff handles backing off executions.
type Backoff struct {
	maxBackoffDuration     time.Duration
	initialBackoffDuration time.Duration
	backoffResetTimeout    time.Duration
	backoffInfo            map[string]backoffInfo
}

// NewBackoff creates an instance of  Backoff.
func NewBackoff(initialBackoffDuration time.Duration, maxBackoffDuration time.Duration, backoffResetTimeout time.Duration) *Backoff {
	return &Backoff{maxBackoffDuration, initialBackoffDuration, backoffResetTimeout, make(map[string]backoffInfo)}
}

// RemoveStaleBackoffData removes stale backoff data.
func (b *Backoff) RemoveStaleBackoffData(currentTime time.Time) {
	for key, backoffInfo := range b.backoffInfo {
		if backoffInfo.lastFailedExecution.Add(b.backoffResetTimeout).Before(currentTime) {
			delete(b.backoffInfo, key)
		}
	}
}

// Backoff execution for the given key. Returns time till execution is backed off.
func (b *Backoff) Backoff(key string, currentTime time.Time) time.Time {
	duration := b.initialBackoffDuration
	if backoffInfo, found := b.backoffInfo[key]; found {
		// Multiple concurrent scale-ups failing shouldn't cause backoff
		// duration to increase, so we only increase it if we're not in
		// backoff right now.
		if backoffInfo.backoffUntil.Before(currentTime) {
			duration = 2 * backoffInfo.duration
			if duration > b.maxBackoffDuration {
				duration = b.maxBackoffDuration
			}
		}
	}
	backoffUntil := currentTime.Add(duration)
	b.backoffInfo[key] = backoffInfo{
		duration:            duration,
		backoffUntil:        backoffUntil,
		lastFailedExecution: currentTime,
	}
	return backoffUntil
}

// RemoveBackoff removes backoff data for the given key.
func (b *Backoff) RemoveBackoff(key string) {
	delete(b.backoffInfo, key)
}

// IsBackedOff returns true if execution is backed off for the given key.
func (b *Backoff) IsBackedOff(key string, currentTime time.Time) bool {
	backoffInfo, found := b.backoffInfo[key]
	return found && backoffInfo.backoffUntil.After(currentTime)
}
