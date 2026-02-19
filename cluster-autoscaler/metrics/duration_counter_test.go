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
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDurationCounter(t *testing.T) {
	dc := NewDurationCounter()
	assert.NotNil(t, dc)
}

func TestDurationCounter_Increment(t *testing.T) {
	dc := NewDurationCounter()

	dc.RegisterLabel(1, "key1")
	dc.RegisterLabel(2, "key2")

	// Test single increment
	dc.Increment(1, 5*time.Second)
	snapshot := dc.Snapshot()
	assert.Equal(t, 5*time.Second, snapshot["key1"])
	assert.Equal(t, 1, len(snapshot))

	// Test multiple increments same key
	dc.Increment(1, 3*time.Second)
	snapshot = dc.Snapshot()
	assert.Equal(t, 8*time.Second, snapshot["key1"])
	assert.Equal(t, 1, len(snapshot))

	// Test new key
	dc.Increment(2, 10*time.Second)
	snapshot = dc.Snapshot()
	assert.Equal(t, 8*time.Second, snapshot["key1"])
	assert.Equal(t, 10*time.Second, snapshot["key2"])
	assert.Equal(t, 2, len(snapshot))

	// Test unregistered key
	dc.Increment(3, 15*time.Second)
	snapshot = dc.Snapshot()
	assert.Equal(t, 8*time.Second, snapshot["key1"])
	assert.Equal(t, 10*time.Second, snapshot["key2"])
	assert.Equal(t, 2, len(snapshot))
}

func TestDurationCounter_Snapshot(t *testing.T) {
	dc := NewDurationCounter()
	dc.RegisterLabel(1, "key1")
	dc.RegisterLabel(2, "key2")
	dc.Increment(1, 1*time.Second)
	dc.Increment(2, 2*time.Second)

	snapshot := dc.Snapshot()
	assert.Equal(t, 1*time.Second, snapshot["key1"])
	assert.Equal(t, 2*time.Second, snapshot["key2"])

	dc.Reset()
	snapshot = dc.Snapshot()
	assert.Empty(t, snapshot)

	dc.Increment(2, 1*time.Second)
	snapshot = dc.Snapshot()
	assert.Equal(t, 1*time.Second, snapshot["key2"])
}

func TestDurationCounter_Reset(t *testing.T) {
	dc := NewDurationCounter()
	dc.RegisterLabel(1, "key1")
	dc.RegisterLabel(2, "key2")
	dc.Increment(1, 10*time.Second)
	dc.Increment(2, 20*time.Second)

	dc.Reset()
	snapshot := dc.Snapshot()
	assert.Empty(t, snapshot)

	// Ensure keys can be reused after reset
	dc.Increment(1, 5*time.Second)
	snapshot = dc.Snapshot()
	assert.Equal(t, 5*time.Second, snapshot["key1"])
}

func TestDurationCounter_RaceCondition(t *testing.T) {
	dc := NewDurationCounter()
	var wg sync.WaitGroup
	routines := 100
	iterations := 1000

	// Register all the keys
	dc.RegisterLabel(0, "key_concurrent")
	for i := 1; i < DurationCountersCapacity; i++ {
		dc.RegisterLabel(i, fmt.Sprintf("key_%d", i))
	}

	// Concurrent increments on the same key
	wg.Add(routines)
	for i := 0; i < routines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				dc.Increment(0, 1*time.Nanosecond)
			}
		}()
	}

	// Concurrent increments on different keys
	wg.Add(routines)
	for i := 0; i < routines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < DurationCountersCapacity - 1; j++ {
				key := (j + 1) % DurationCountersCapacity
				dc.Increment(key, 1*time.Nanosecond)
			}
		}(i)
	}

	wg.Wait()

	snapshot := dc.Snapshot()

	// Verify same key total
	expectedDuration := time.Duration(routines*iterations) * time.Nanosecond
	assert.Equal(t, expectedDuration, snapshot["key_concurrent"])

	// Verify different keys
	for i := 1; i < DurationCountersCapacity; i++ {
		key := fmt.Sprintf("key_%d", i)
		expected := time.Duration(routines) * time.Nanosecond
		assert.Equal(t, expected, snapshot[key])
	}
}

func BenchmarkDurationCounter_SingleCounter(b *testing.B) {
	dc := NewDurationCounter()
	dc.RegisterLabel(0, "key")
	for b.Loop() {
		dc.Increment(0, 1*time.Nanosecond)
	}
}

func BenchmarkDurationCounter_StaticSetOfKeys(b *testing.B) {
	dc := NewDurationCounter()
	for i := 0; i < DurationCountersCapacity; i++ {
		dc.RegisterLabel(i, fmt.Sprintf("key_%d", i))
	}
	for i := 0; i < b.N; i++ {
		dc.Increment(i%DurationCountersCapacity, 1*time.Nanosecond)
	}
}
