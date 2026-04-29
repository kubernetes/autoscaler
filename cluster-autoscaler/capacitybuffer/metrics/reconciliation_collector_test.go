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

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	filters "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/filters"
	clocktesting "k8s.io/utils/clock/testing"
	"k8s.io/utils/ptr"
)

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

func TestReconciliationTimestampCollector_Collect(t *testing.T) {
	now := time.Unix(1710410000, 0) // Fixed "now" for deterministic tests
	time5mAgo := now.Add(-5 * time.Minute)
	time10mAgo := now.Add(-10 * time.Minute)
	time15mAgo := now.Add(-15 * time.Minute)
	time20mAgo := now.Add(-20 * time.Minute)

	testCases := []struct {
		name                string
		buffers             []*v1beta1.CapacityBuffer
		initialCache        map[types.UID]time.Time
		supportedStrategies []string
		wantNew             float64
		wantReconciled      float64
		wantFinalSnapshot   map[types.UID]time.Time
	}{
		{
			name:                "No buffers",
			buffers:             []*v1beta1.CapacityBuffer{},
			initialCache:        nil,
			supportedStrategies: []string{""},
			wantNew:             float64(now.Unix()),
			wantReconciled:      float64(now.Unix()),
			wantFinalSnapshot:   map[types.UID]time.Time{},
		},
		{
			name: "New and reconciled buffers",
			buffers: []*v1beta1.CapacityBuffer{
				newTestCapacityBuffer("new-uid", "", time10mAgo, nil),
				newTestCapacityBuffer("proc-uid", "", now, nil),
			},
			initialCache: map[types.UID]time.Time{
				"proc-uid": time5mAgo,
			},
			supportedStrategies: []string{""},
			wantNew:             float64(time10mAgo.Unix()),
			wantReconciled:      float64(time5mAgo.Unix()),
			wantFinalSnapshot: map[types.UID]time.Time{
				"proc-uid": time5mAgo,
			},
		},
		{
			name: "Complex scenario with multiple buffer states",
			buffers: []*v1beta1.CapacityBuffer{
				newTestCapacityBuffer("unsupported-uid", "unsupported", now, nil),
				newTestCapacityBuffer("skipped-uid", "", now, []metav1.Condition{{Type: "Ready", Status: "True"}}),
				newTestCapacityBuffer("new-uid-1", "", time10mAgo, nil),
				newTestCapacityBuffer("new-uid-2", "", time20mAgo, nil),
				newTestCapacityBuffer("proc-uid-1", "", now, nil),
				newTestCapacityBuffer("proc-uid-2", "", now, nil),
			},
			initialCache: map[types.UID]time.Time{
				"proc-uid-1":      time5mAgo,
				"proc-uid-2":      time15mAgo,
				"unsupported-uid": time5mAgo,
			},
			supportedStrategies: []string{""},
			wantNew:             float64(time20mAgo.Unix()),
			wantReconciled:      float64(time15mAgo.Unix()),
			wantFinalSnapshot: map[types.UID]time.Time{
				"proc-uid-1":      time5mAgo,
				"proc-uid-2":      time15mAgo,
				"unsupported-uid": time5mAgo,
			},
		},
		{
			name: "Cleanup of stale cache entries",
			buffers: []*v1beta1.CapacityBuffer{
				newTestCapacityBuffer("active-uid", "", now, nil),
				newTestCapacityBuffer("unsupported-uid", "unsupported", now, nil),
			},
			initialCache: map[types.UID]time.Time{
				"active-uid":      time5mAgo,
				"unsupported-uid": time5mAgo,
				"stale-uid":       time10mAgo,
			},
			supportedStrategies: []string{""},
			wantNew:             float64(now.Unix()),
			wantReconciled:      float64(time5mAgo.Unix()),
			wantFinalSnapshot: map[types.UID]time.Time{
				"active-uid":      time5mAgo,
				"unsupported-uid": time5mAgo,
			},
		},
		{
			name: "Oldest value is reported when cache has multiple entries",
			buffers: []*v1beta1.CapacityBuffer{
				newTestCapacityBuffer("proc-uid-old", "", now, nil),
				newTestCapacityBuffer("proc-uid-new", "", now, nil),
			},
			initialCache: map[types.UID]time.Time{
				"proc-uid-old": time20mAgo,
				"proc-uid-new": time5mAgo,
			},
			supportedStrategies: []string{""},
			wantNew:             float64(now.Unix()),
			wantReconciled:      float64(time20mAgo.Unix()),
			wantFinalSnapshot: map[types.UID]time.Time{
				"proc-uid-old": time20mAgo,
				"proc-uid-new": time5mAgo,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeClock := clocktesting.NewFakeClock(now)

			c := newFakeClient(tc.buffers)
			reconciledCache := NewReconciliationCache()
			if len(tc.initialCache) > 0 {
				// Inject initial cache directly for testing varied timestamps
				for uid, ts := range tc.initialCache {
					buf := []*v1beta1.CapacityBuffer{{ObjectMeta: metav1.ObjectMeta{UID: uid}}}
					reconciledCache.Update(buf, ts)
				}
			}

			filter := filters.NewStrategyFilter(tc.supportedStrategies)

			collector := &reconciliationTimestampCollector{
				client:                 c,
				reconciledBuffers:      reconciledCache,
				supportedBuffersFilter: filter,
				clock:                  fakeClock,
			}

			metricsChan := make(chan prometheus.Metric)
			go func() {
				collector.Collect(metricsChan)
				close(metricsChan)
			}()

			var gotNew, gotReconciled float64
			var foundNew, foundReconciled bool
			for m := range metricsChan {
				metricDto := &dto.Metric{}
				m.Write(metricDto)
				isNew := false
				for _, lp := range metricDto.Label {
					if *lp.Name == "new_buffer" && *lp.Value == "true" {
						isNew = true
						break
					}
				}
				if isNew {
					gotNew = *metricDto.Gauge.Value
					foundNew = true
				} else {
					gotReconciled = *metricDto.Gauge.Value
					foundReconciled = true
				}
			}

			assert.True(t, foundNew, "new_buffer=true metric not found")
			assert.True(t, foundReconciled, "new_buffer=false metric not found")
			assert.Equal(t, tc.wantNew, gotNew, "new_buffer=true timestamp mismatch")
			assert.Equal(t, tc.wantReconciled, gotReconciled, "new_buffer=false timestamp mismatch")

			finalSnapshot := reconciledCache.Snapshot()
			assert.Equal(t, tc.wantFinalSnapshot, finalSnapshot)
		})
	}
}
