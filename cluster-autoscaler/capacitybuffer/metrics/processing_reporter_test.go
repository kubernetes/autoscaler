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
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	filters "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/filters"
	clocktesting "k8s.io/utils/clock/testing"
	"k8s.io/utils/ptr"
)

type mockMetrics struct {
	mock.Mock
}

func (m *mockMetrics) ObserveCapacityBuffersProcessingIntervalSeconds(isNewBuffer bool, duration time.Duration) {
	m.Called(isNewBuffer, duration)
}

type fakeClient struct {
	buffers []*v1beta1.CapacityBuffer
}

func (m *fakeClient) ListCapacityBuffers(namespace string) ([]*v1beta1.CapacityBuffer, error) {
	return m.buffers, nil
}

func newFakeClient(buffers []*v1beta1.CapacityBuffer) *fakeClient {
	return &fakeClient{buffers: buffers}
}

func newTestCapacityBuffer(uid, strategy string, creationTime time.Time, conditions []metav1.Condition) *v1beta1.CapacityBuffer {
	return &v1beta1.CapacityBuffer{
		ObjectMeta: metav1.ObjectMeta{
			UID:               types.UID(uid),
			CreationTimestamp: metav1.Time{Time: creationTime},
		},
		Spec: v1beta1.CapacityBufferSpec{
			ProvisioningStrategy: ptr.To(strategy),
		},
		Status: v1beta1.CapacityBufferStatus{
			Conditions: conditions,
		},
	}
}

func TestProcessingIntervalMetricReporter_Loop(t *testing.T) {
	now := time.Now()
	creationTime := now.Add(-10 * time.Minute)
	cachedTime := now.Add(-5 * time.Minute)

	testCases := []struct {
		name                  string
		buffers               []*v1beta1.CapacityBuffer
		snapshot              map[string]time.Time
		supportedStrategies   []string
		expectedNew           time.Duration
		expectedProcessed     time.Duration
		expectedFinalSnapshot map[string]time.Time
	}{
		{
			name:                  "No buffers",
			buffers:               []*v1beta1.CapacityBuffer{},
			snapshot:              map[string]time.Time{},
			supportedStrategies:   []string{""},
			expectedNew:           0,
			expectedProcessed:     0,
			expectedFinalSnapshot: map[string]time.Time{},
		},
		{
			name: "New and processed buffers",
			buffers: []*v1beta1.CapacityBuffer{
				newTestCapacityBuffer("new-uid", "", creationTime, nil),
				newTestCapacityBuffer("proc-uid", "", now, nil),
			},
			snapshot: map[string]time.Time{
				"proc-uid": cachedTime,
			},
			supportedStrategies: []string{""},
			expectedNew:         10 * time.Minute,
			expectedProcessed:   5 * time.Minute,
			expectedFinalSnapshot: map[string]time.Time{
				"proc-uid": cachedTime,
			},
		},
		{
			name: "Complex scenario with multiple buffer states",
			buffers: []*v1beta1.CapacityBuffer{
				newTestCapacityBuffer("unsupported-uid", "unsupported", now, nil),
				newTestCapacityBuffer("skipped-uid", "", now, []metav1.Condition{{Type: "Ready", Status: "True"}}),
				newTestCapacityBuffer("new-uid-1", "", now.Add(-10*time.Minute), nil),
				newTestCapacityBuffer("new-uid-2", "", now.Add(-20*time.Minute), nil),
				newTestCapacityBuffer("proc-uid-1", "", now, nil),
				newTestCapacityBuffer("proc-uid-2", "", now, nil),
			},
			snapshot: map[string]time.Time{
				"proc-uid-1": now.Add(-5 * time.Minute),
				"proc-uid-2": now.Add(-15 * time.Minute),
			},
			supportedStrategies: []string{""},
			expectedNew:         20 * time.Minute,
			expectedProcessed:   15 * time.Minute,
			expectedFinalSnapshot: map[string]time.Time{
				"proc-uid-1": now.Add(-5 * time.Minute),
				"proc-uid-2": now.Add(-15 * time.Minute),
			},
		},
		{
			name: "Cleanup of stale cache entries",
			buffers: []*v1beta1.CapacityBuffer{
				newTestCapacityBuffer("active-uid", "", now, nil),
			},
			snapshot: map[string]time.Time{
				"active-uid": now.Add(-5 * time.Minute),
				"stale-uid":  now.Add(-10 * time.Minute),
			},
			supportedStrategies: []string{""},
			expectedNew:         0,
			expectedProcessed:   5 * time.Minute,
			expectedFinalSnapshot: map[string]time.Time{
				"active-uid": now.Add(-5 * time.Minute),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeClock := clocktesting.NewFakeClock(now)

			m := &mockMetrics{}
			m.On("ObserveCapacityBuffersProcessingIntervalSeconds", true, tc.expectedNew).Once()
			m.On("ObserveCapacityBuffersProcessingIntervalSeconds", false, tc.expectedProcessed).Once()

			c := newFakeClient(tc.buffers)
			processedCache := NewProcessingCache()
			processedCache.Update(tc.snapshot)

			filter := filters.NewStrategyFilter(tc.supportedStrategies)

			reporter := &processingIntervalMetricReporter{
				client:                 c,
				processedBuffers:       processedCache,
				supportedBuffersFilter: filter,
				metrics:                m,
				clock:                  fakeClock,
			}

			reporter.loop()

			m.AssertExpectations(t)

			finalSnapshot := processedCache.Snapshot()
			assert.Equal(t, tc.expectedFinalSnapshot, finalSnapshot)
		})
	}
}
