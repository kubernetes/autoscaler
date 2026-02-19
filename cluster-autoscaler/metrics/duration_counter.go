/*
Copyright 2026 The Kubernetes Authors.

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

package metrics

import (
	"sync/atomic"
	"time"

	"k8s.io/klog/v2"
)

const DurationCountersCapacity = 128

// DurationCounter is a thread-safe counter for durations. It is used to
// count the total duration of operations with the same key under assumption
// that the number of keys is small is for the most part unchanging.
type DurationCounter struct {
	counters [DurationCountersCapacity]atomic.Int64
	counterLabels [DurationCountersCapacity]string
	entriesRegistered atomic.Int64
}

// NewDurationCounter creates a new DurationCounter.
func NewDurationCounter() *DurationCounter {
	return &DurationCounter{
		counters: [DurationCountersCapacity]atomic.Int64{},
	}
}

// RegisterLabel registers a label for a given key.
func (dc *DurationCounter) RegisterLabel(key int, label string) {
	dc.counterLabels[key] = label
	dc.entriesRegistered.Add(1)
}

// Increment increments the counter for the given key by the given duration.
func (dc *DurationCounter) Increment(key int, duration time.Duration) {
	dc.counters[key].Add(int64(duration))
}

// Snapshot returns a snapshot of the current counters, may be inaccurate
// if the counter is getting incremented concurrently.
func (dc *DurationCounter) Snapshot() map[string]time.Duration {
	result := make(map[string]time.Duration, dc.entriesRegistered.Load())
	for i := range dc.counters {
		label := dc.counterLabels[i]
		counter := dc.counters[i].Load()

		// Do not store entries with zero values as it indicates that
		// the key was never incremented during this iteration of the counter
		// or wasn't registered at all
		if counter == 0 {
			continue
		}

		if label == "" {
			klog.Warningf("DurationCounter: counter %d is incremented, while label is not registered", i)
			continue
		}

		result[label] = time.Duration(counter)
	}

	return result
}

// Reset resets the counters without removing the keys.
func (dc *DurationCounter) Reset() {
	for i := range dc.counters {
		dc.counters[i].Store(0)
	}
}
