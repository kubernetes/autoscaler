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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
)

func TestReconciliationCache(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name        string
		initialData []*v1beta1.CapacityBuffer
		updateWith  []*v1beta1.CapacityBuffer
		wantResult  map[types.UID]time.Time
	}{
		{
			name:        "Empty initial, update with data",
			initialData: []*v1beta1.CapacityBuffer{},
			updateWith:  []*v1beta1.CapacityBuffer{{ObjectMeta: metav1.ObjectMeta{UID: "uid1"}}},
			wantResult:  map[types.UID]time.Time{"uid1": now},
		},
		{
			name:        "Existing data, update with new data (merge)",
			initialData: []*v1beta1.CapacityBuffer{{ObjectMeta: metav1.ObjectMeta{UID: "uid1"}}},
			updateWith:  []*v1beta1.CapacityBuffer{{ObjectMeta: metav1.ObjectMeta{UID: "uid2"}}},
			wantResult:  map[types.UID]time.Time{"uid1": now, "uid2": now},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := NewReconciliationCache()
			if len(tc.initialData) > 0 {
				r.Update(tc.initialData, now)
			}
			r.Update(tc.updateWith, now)
			snapshot := r.Snapshot()
			assert.Equal(t, tc.wantResult, snapshot)

			// Verify snapshot is a deep copy by modifying it and checking the cache remains unchanged
			snapshot["deep-copy-check"] = time.Now()
			assert.NotContains(t, r.Snapshot(), "deep-copy-check")
		})
	}
}

func TestReconciliationCache_Prune(t *testing.T) {
	now := time.Now()
	initBuffers := []*v1beta1.CapacityBuffer{
		{ObjectMeta: metav1.ObjectMeta{UID: "uid1"}},
		{ObjectMeta: metav1.ObjectMeta{UID: "uid2"}},
		{ObjectMeta: metav1.ObjectMeta{UID: "uid3"}},
	}

	testCases := []struct {
		name             string
		supportedBuffers []*v1beta1.CapacityBuffer
		wantSnapshot     map[types.UID]time.Time
	}{
		{
			name:             "Prune all",
			supportedBuffers: []*v1beta1.CapacityBuffer{},
			wantSnapshot:     map[types.UID]time.Time{},
		},
		{
			name: "Prune none",
			supportedBuffers: []*v1beta1.CapacityBuffer{
				{ObjectMeta: metav1.ObjectMeta{UID: "uid1"}},
				{ObjectMeta: metav1.ObjectMeta{UID: "uid2"}},
				{ObjectMeta: metav1.ObjectMeta{UID: "uid3"}},
				{ObjectMeta: metav1.ObjectMeta{UID: "extra"}},
			},
			wantSnapshot: map[types.UID]time.Time{
				"uid1": now,
				"uid2": now,
				"uid3": now,
			},
		},
		{
			name: "Prune some",
			supportedBuffers: []*v1beta1.CapacityBuffer{
				{ObjectMeta: metav1.ObjectMeta{UID: "uid2"}},
			},
			wantSnapshot: map[types.UID]time.Time{
				"uid2": now,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := NewReconciliationCache()
			r.Update(initBuffers, now)
			r.Prune(tc.supportedBuffers)
			assert.Equal(t, tc.wantSnapshot, r.Snapshot())
		})
	}
}
