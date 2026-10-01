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

package pressure

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	testclock "k8s.io/utils/clock/testing"
	"k8s.io/utils/ptr"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

const mi = 1024 * 1024

var (
	testBase   = time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	testTarget = Target{
		ContainerID: model.ContainerID{PodID: model.PodID{Namespace: "ns", PodName: "p"}, ContainerName: "c"},
		PodUID:      "uid",
		Node:        "n",
		LimitBytes:  512 * mi,
	}
	testKey = targetKey{podUID: "uid", container: "c"}
)

// testSummary returns a one-container Summary. total is in the configured PSI unit.
func testSummary(at, started time.Time, total, workingSet uint64) summary {
	memory := &memoryStats{Time: at, WorkingSetBytes: &workingSet, PSI: &psiStats{}}
	memory.PSI.Some.Total = total
	pod := podStats{Containers: []containerStats{{Name: "c", StartTime: started, Memory: memory}}}
	pod.PodRef.UID = "uid"
	return summary{Pods: []podStats{pod}}
}

func drainEvents(o *KubeletObserver) int {
	n := 0
	for {
		select {
		case <-o.ch:
			n++
		default:
			return n
		}
	}
}

func TestProcess(t *testing.T) {
	type step struct {
		at time.Duration
		// started overrides the case's container start offset, to simulate a restart.
		started    time.Duration
		total      uint64
		workingSet uint64
		// limit, when set, resizes the target's memory limit before this step.
		limit int64
		// clockSkew is the observer clock minus the sample time; 1s when zero.
		clockSkew  time.Duration
		wantEvents int
	}
	testCases := []struct {
		name           string
		unit           time.Duration
		started        time.Duration
		resizeState    string
		steps          []step
		wantQualifying *int
	}{
		{
			// 41.9s -> 55.7s of stall over 60s is a rate of 0.23; 486/512 of the limit is 0.95.
			name:    "two qualifying windows emit one event",
			started: -time.Hour,
			steps: []step{
				{at: 0, total: 41_900_000, workingSet: 470 * mi},
				{at: time.Minute, total: 55_700_000, workingSet: 486 * mi},
				{at: 2 * time.Minute, total: 69_500_000, workingSet: 486 * mi, wantEvents: 1},
			},
		},
		{
			name:    "working set below the fraction resets the count",
			started: -time.Hour,
			steps: []step{
				{at: 0, total: 0, workingSet: 486 * mi},
				{at: time.Minute, total: 14_000_000, workingSet: 486 * mi},
				{at: 2 * time.Minute, total: 28_000_000, workingSet: 400 * mi},
				{at: 3 * time.Minute, total: 42_000_000, workingSet: 486 * mi},
				{at: 4 * time.Minute, total: 56_000_000, workingSet: 486 * mi, wantEvents: 1},
			},
		},
		{
			name:    "stall below the ratio never qualifies",
			started: -time.Hour,
			steps: []step{
				{at: 0, total: 0, workingSet: 500 * mi},
				{at: time.Minute, total: 3_000_000, workingSet: 500 * mi},
				{at: 2 * time.Minute, total: 6_000_000, workingSet: 500 * mi},
				{at: 3 * time.Minute, total: 9_000_000, workingSet: 500 * mi},
			},
		},
		{
			name:    "startup grace suppresses events",
			started: 0,
			steps: []step{
				{at: 10 * time.Second, total: 0, workingSet: 486 * mi},
				{at: 20 * time.Second, total: 6_000_000, workingSet: 486 * mi},
				{at: 30 * time.Second, total: 12_000_000, workingSet: 486 * mi},
				{at: 70 * time.Second, total: 36_000_000, workingSet: 486 * mi, wantEvents: 1},
			},
		},
		{
			name:    "stale samples are discarded",
			started: -time.Hour,
			steps: []step{
				{at: 0, total: 0, workingSet: 486 * mi, clockSkew: 3 * time.Minute},
				{at: time.Minute, total: 14_000_000, workingSet: 486 * mi, clockSkew: 3 * time.Minute},
				{at: 2 * time.Minute, total: 28_000_000, workingSet: 486 * mi, clockSkew: 3 * time.Minute},
			},
		},
		{
			// A first sample dated ten minutes ahead must not become the baseline, or every later valid
			// sample would look older and be ignored.
			name:    "future-dated first sample does not block later samples",
			started: -time.Hour,
			steps: []step{
				{at: 10 * time.Minute, total: 0, workingSet: 486 * mi, clockSkew: -10 * time.Minute},
				{at: time.Minute, total: 0, workingSet: 486 * mi},
				{at: 2 * time.Minute, total: 14_000_000, workingSet: 486 * mi},
				{at: 3 * time.Minute, total: 28_000_000, workingSet: 486 * mi, wantEvents: 1},
			},
		},
		{
			name:    "future samples are discarded",
			started: -time.Hour,
			steps: []step{
				{at: 0, total: 0, workingSet: 486 * mi, clockSkew: -10 * time.Second},
				{at: time.Minute, total: 14_000_000, workingSet: 486 * mi, clockSkew: -10 * time.Second},
				{at: 2 * time.Minute, total: 28_000_000, workingSet: 486 * mi, clockSkew: -10 * time.Second},
			},
		},
		{
			// The counter after the restart is larger by an in-range amount, so only the identity check
			// stops the second window from qualifying.
			name:    "container restart resets the baseline",
			started: -time.Hour,
			steps: []step{
				{at: 0, total: 0, workingSet: 486 * mi},
				{at: time.Minute, total: 14_000_000, workingSet: 486 * mi},
				{at: 2 * time.Minute, started: -30 * time.Minute, total: 28_000_000, workingSet: 486 * mi},
			},
			wantQualifying: ptr.To(0),
		},
		{
			// 486/520 still passes the working-set condition, so only the limit check stops the second
			// window from qualifying.
			name:    "limit change resets the baseline",
			started: -time.Hour,
			steps: []step{
				{at: 0, total: 0, workingSet: 486 * mi},
				{at: time.Minute, total: 14_000_000, workingSet: 486 * mi},
				{at: 2 * time.Minute, total: 28_000_000, workingSet: 486 * mi, limit: 520 * mi},
			},
			wantQualifying: ptr.To(0),
		},
		{
			name:    "gap longer than two intervals resets the baseline",
			started: -time.Hour,
			steps: []step{
				{at: 0, total: 0, workingSet: 486 * mi},
				{at: time.Minute, total: 14_000_000, workingSet: 486 * mi},
				{at: 4 * time.Minute, total: 56_000_000, workingSet: 486 * mi},
			},
			wantQualifying: ptr.To(0),
		},
		{
			name:    "nanosecond counters with a nanosecond unit",
			unit:    time.Nanosecond,
			started: -time.Hour,
			steps: []step{
				{at: 0, total: 0, workingSet: 486 * mi},
				{at: time.Minute, total: 14_000_000_000, workingSet: 486 * mi},
				{at: 2 * time.Minute, total: 28_000_000_000, workingSet: 486 * mi, wantEvents: 1},
			},
		},
		{
			// Read as microseconds, a nanosecond counter gives a rate of ~233.
			name:    "nanosecond counters with the microsecond default are discarded",
			started: -time.Hour,
			steps: []step{
				{at: 0, total: 0, workingSet: 486 * mi},
				{at: time.Minute, total: 14_000_000_000, workingSet: 486 * mi},
				{at: 2 * time.Minute, total: 28_000_000_000, workingSet: 486 * mi},
			},
		},
		{
			// Timestamps have one-second resolution, so a 60s window can hold 61s of stall.
			name:    "full stall with whole-second timestamps qualifies",
			started: -time.Hour,
			steps: []step{
				{at: 0, total: 0, workingSet: 500 * mi},
				{at: time.Minute, total: 61_000_000, workingSet: 500 * mi},
				{at: 2 * time.Minute, total: 122_000_000, workingSet: 500 * mi, wantEvents: 1},
			},
		},
		{
			name:        "outstanding resize suppresses events",
			started:     -time.Hour,
			resizeState: "resize_deferred",
			steps: []step{
				{at: 0, total: 0, workingSet: 486 * mi},
				{at: time.Minute, total: 14_000_000, workingSet: 486 * mi},
				{at: 2 * time.Minute, total: 28_000_000, workingSet: 486 * mi},
			},
		},
		{
			name:    "cooldown suppresses a second event",
			started: -time.Hour,
			steps: []step{
				{at: 0, total: 0, workingSet: 486 * mi},
				{at: time.Minute, total: 14_000_000, workingSet: 486 * mi},
				{at: 2 * time.Minute, total: 28_000_000, workingSet: 486 * mi, wantEvents: 1},
				{at: 3 * time.Minute, total: 42_000_000, workingSet: 486 * mi},
				{at: 4 * time.Minute, total: 56_000_000, workingSet: 486 * mi},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := DefaultConfig()
			if tc.unit != 0 {
				cfg.PSITotalUnit = tc.unit
			}
			o := NewObserver(nil, cfg)
			target := testTarget
			target.ResizeState = tc.resizeState
			o.SetTargets([]Target{target})
			for i, s := range tc.steps {
				if s.limit != 0 {
					target.LimitBytes = s.limit
					o.SetTargets([]Target{target})
				}
				started := tc.started
				if s.started != 0 {
					started = s.started
				}
				skew := s.clockSkew
				if skew == 0 {
					skew = time.Second
				}
				at := testBase.Add(s.at)
				o.clock = testclock.NewFakePassiveClock(at.Add(skew))
				o.process(testSummary(at, testBase.Add(started), s.total, s.workingSet))
				assert.Equal(t, s.wantEvents, drainEvents(o), "events after step %d", i)
			}
			if tc.wantQualifying != nil {
				st := o.state[testKey]
				if assert.NotNil(t, st) {
					assert.Equal(t, *tc.wantQualifying, st.qualifying)
					assert.Equal(t, target.LimitBytes, st.limit)
				}
			}
		})
	}
}

func TestDeriveConcurrency(t *testing.T) {
	testCases := []struct {
		name    string
		latency time.Duration
		nodes   int
		want    int
	}{
		{name: "fits the sweep timeout", latency: 450 * time.Millisecond, nodes: 5000, want: 50},
		{name: "at least one worker", latency: 450 * time.Millisecond, nodes: 10, want: 1},
		{name: "capped at MaxConcurrent", latency: 2 * time.Second, nodes: 5000, want: 50},
		{name: "below the cap", latency: 2 * time.Second, nodes: 1000, want: 45},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			o := NewObserver(nil, DefaultConfig())
			for range 100 {
				o.latencies = append(o.latencies, tc.latency)
			}
			assert.Equal(t, tc.want, o.deriveConcurrencyLocked(tc.nodes))
		})
	}
}

func TestQueueFullDropsWithoutCooldown(t *testing.T) {
	o := NewObserver(nil, DefaultConfig())
	o.SetTargets([]Target{testTarget})
	started := testBase.Add(-time.Hour)
	feed := func(at time.Duration, total uint64) {
		o.clock = testclock.NewFakePassiveClock(testBase.Add(at + time.Second))
		o.process(testSummary(testBase.Add(at), started, total, 486*mi))
	}
	o.ch = make(chan Info) // nobody receives, so the send is dropped
	feed(0, 0)
	feed(time.Minute, 14_000_000)
	feed(2*time.Minute, 28_000_000)
	st := o.state[testKey]
	assert.True(t, st.cooldownUntil.IsZero(), "a dropped event does not start a cooldown")
	assert.Equal(t, 0, st.qualifying, "a dropped event needs two fresh windows again")

	o.ch = make(chan Info, 1)
	feed(3*time.Minute, 42_000_000)
	assert.Equal(t, 0, drainEvents(o))
	feed(4*time.Minute, 56_000_000)
	assert.Equal(t, 1, drainEvents(o), "the container is delivered once the queue has room")
}
