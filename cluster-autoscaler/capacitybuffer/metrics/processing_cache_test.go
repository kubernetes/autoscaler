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

package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProcessingCache(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name           string
		initialData    map[string]time.Time
		updateWith     map[string]time.Time
		expectedResult map[string]time.Time
	}{
		{
			name:           "Empty initial, update with data",
			initialData:    map[string]time.Time{},
			updateWith:     map[string]time.Time{"uid1": now},
			expectedResult: map[string]time.Time{"uid1": now},
		},
		{
			name:           "Existing data, update with empty",
			initialData:    map[string]time.Time{"uid1": now},
			updateWith:     map[string]time.Time{},
			expectedResult: map[string]time.Time{"uid1": now},
		},
		{
			name:           "Existing data, update with new data (merge)",
			initialData:    map[string]time.Time{"uid1": now},
			updateWith:     map[string]time.Time{"uid2": now.Add(time.Minute)},
			expectedResult: map[string]time.Time{"uid1": now, "uid2": now.Add(time.Minute)},
		},
		{
			name:           "Existing data, update existing key",
			initialData:    map[string]time.Time{"uid1": now},
			updateWith:     map[string]time.Time{"uid1": now.Add(time.Minute)},
			expectedResult: map[string]time.Time{"uid1": now.Add(time.Minute)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewProcessingCache()
			if len(tc.initialData) > 0 {
				p.Update(tc.initialData)
			}
			p.Update(tc.updateWith)
			snapshot := p.Snapshot()
			assert.Equal(t, tc.expectedResult, snapshot)

			// Verify snapshot is a deep copy by modifying it and checking the cache remains unchanged
			snapshot["deep-copy-check"] = time.Now()
			assert.NotContains(t, p.Snapshot(), "deep-copy-check")
		})
	}
}

func TestProcessingCache_Prune(t *testing.T) {
	now := time.Now()
	initMap := map[string]time.Time{
		"uid1": now,
		"uid2": now,
		"uid3": now,
	}

	testCases := []struct {
		name          string
		supportedUIDs []string
		expected      map[string]time.Time
	}{
		{
			name:          "Prune all",
			supportedUIDs: []string{},
			expected:      map[string]time.Time{},
		},
		{
			name:          "Prune none",
			supportedUIDs: []string{"uid1", "uid2", "uid3", "extra"},
			expected: map[string]time.Time{
				"uid1": now,
				"uid2": now,
				"uid3": now,
			},
		},
		{
			name:          "Prune some",
			supportedUIDs: []string{"uid2"},
			expected: map[string]time.Time{
				"uid2": now,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewProcessingCache()
			p.Update(initMap)
			p.Prune(tc.supportedUIDs)
			assert.Equal(t, tc.expected, p.Snapshot())
		})
	}
}
