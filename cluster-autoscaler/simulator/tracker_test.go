/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package simulator

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUsageTracker(t *testing.T) {
	tracker := NewUsageTracker()
	now := time.Now()
	tracker.RegisterUsage("A", "B", now.Add(-5*time.Minute))
	tracker.RegisterUsage("A", "C", now.Add(-10*time.Minute))
	tracker.RegisterUsage("D", "C", now.Add(-35*time.Minute))
	tracker.RegisterUsage("D", "C", now.Add(-25*time.Minute))
	tracker.RegisterUsage("D", "C", now.Add(-20*time.Minute))
	tracker.RegisterUsage("C", "E", now.Add(-20*time.Minute))

	for i := 0; i < maxUsageRecorded+5; i++ {
		tracker.RegisterUsage(fmt.Sprintf("X%d", i), "X", now)
		tracker.RegisterUsage("Y", fmt.Sprintf("X%d", i), now)
	}

	C, _ := tracker.Get("C")
	X, _ := tracker.Get("X")
	Y, _ := tracker.Get("Y")

	// Checking regular nodes.
	assert.Equal(t, 1, len(C.using))
	assert.Contains(t, C.using, "E")
	assert.Contains(t, C.usedBy, "A")
	assert.Contains(t, C.usedBy, "D")

	assert.Equal(t, 2, len(C.usedBy))
	assert.False(t, C.usedByTooMany)
	assert.False(t, C.usingTooMany)

	// Checking overflow.
	assert.True(t, X.usedByTooMany)
	assert.False(t, X.usingTooMany)
	assert.True(t, Y.usingTooMany)
	assert.False(t, Y.usedByTooMany)

	// Checking cleanup
	tracker.CleanUp(now.Add(-17 * time.Minute))

	C, foundC := tracker.Get("C")
	_, foundD := tracker.Get("D")
	_, foundE := tracker.Get("E")

	assert.True(t, foundC)
	assert.Contains(t, C.usedBy, "A")
	assert.NotContains(t, C.usedBy, "D")

	assert.False(t, foundD)
	assert.False(t, foundE)
}

func TestRemove(t *testing.T) {
	tracker := NewUsageTracker()
	now := time.Now()
	tracker.RegisterUsage("A", "B", now)
	tracker.RegisterUsage("A", "C", now)
	tracker.RegisterUsage("X", "C", now)
	tracker.RegisterUsage("C", "Z", now)
	tracker.RegisterUsage("M", "N", now)

	utilization := map[string]time.Time{
		"A": now,
		"C": now,
		"X": now,
		"M": now,
	}

	RemoveNodeFromTracker(tracker, "A", utilization)

	_, foundA := tracker.Get("A")
	C, foundC := tracker.Get("C")
	_, foundX := tracker.Get("X")

	assert.False(t, foundA)
	assert.True(t, foundC)
	assert.True(t, foundX)
	assert.NotContains(t, C.usedBy, "A")
	assert.Contains(t, C.usedBy, "X")

	_, foundA = utilization["A"]
	_, foundC = utilization["C"]
	_, foundX = utilization["X"]

	assert.False(t, foundA)
	assert.True(t, foundC)
	assert.False(t, foundX)
}
