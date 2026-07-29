/*
Copyright 2022 The Kubernetes Authors.

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

package actuation

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

type mockClock struct {
	startTime   time.Time
	currentTime time.Time
	mutex       sync.Mutex
}

func NewMockClock(startTime time.Time) *mockClock {
	return &mockClock{
		startTime:   startTime,
		currentTime: startTime,
	}
}

func (m *mockClock) Now() time.Time {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.currentTime
}

func (m *mockClock) SetTime(t time.Time) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.currentTime = t
}

type TestCustomNodeLister struct {
	nodes                  map[string]*apiv1.Node
	getCallCount           map[string]int
	nodeTaintAfterDuration map[string]time.Duration
	clock                  *mockClock
}

// List returns all nodes in test lister.
func (l *TestCustomNodeLister) List() ([]*apiv1.Node, error) {
	var nodes []*apiv1.Node
	for _, node := range l.nodes {
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (l *TestCustomNodeLister) Get(name string) (*apiv1.Node, error) {
	for _, node := range l.nodes {
		if node.Name == name {
			l.getCallCount[node.Name] += 1
			if expectedDuration, ok := l.nodeTaintAfterDuration[node.Name]; ok {
				if l.clock.Now().Sub(l.clock.startTime) >= expectedDuration {
					toBeDeletedTaint := apiv1.Taint{Key: taints.ToBeDeletedTaint, Effect: apiv1.TaintEffectNoSchedule}
					node.Spec.Taints = append(node.Spec.Taints, toBeDeletedTaint)
				}
			}
			return node, nil
		}
	}
	return nil, fmt.Errorf("Node %s not found", name)
}

// Return new TestCustomNodeLister object
func NewTestCustomNodeLister(nodes map[string]*apiv1.Node, nodeTaintAfterDuration map[string]time.Duration, clock *mockClock) *TestCustomNodeLister {
	getCallCounts := map[string]int{}
	for name := range nodes {
		getCallCounts[name] = 0
	}
	return &TestCustomNodeLister{
		nodes:                  nodes,
		getCallCount:           getCallCounts,
		nodeTaintAfterDuration: nodeTaintAfterDuration,
		clock:                  clock,
	}
}

func TestUpdateLatencyCalculation(t *testing.T) {
	oldTimeout := waitForTaintingTimeoutDuration
	waitForTaintingTimeoutDuration = 150 * time.Millisecond
	defer func() { waitForTaintingTimeoutDuration = oldTimeout }()

	testCases := []struct {
		description string
		startTime   time.Time
		nodes       []string
		// If an entry is not added for a node, that node will never get tainted
		nodeTaintAfterDuration    map[string]time.Duration
		wantLatency               time.Duration
		wantResultChanOpen        bool
		simulateZeroExpectedCount bool
	}{
		{
			description:            "latency when tainting a single node - node is tainted in the first call to the lister",
			startTime:              time.Now(),
			nodes:                  []string{"n1"},
			nodeTaintAfterDuration: map[string]time.Duration{"n1": 100 * time.Millisecond},
			wantLatency:            100 * time.Millisecond,
			wantResultChanOpen:     true,
		},
		{
			description:            "latency when tainting a single node - node is not tainted in the first call to the lister",
			startTime:              time.Now(),
			nodes:                  []string{"n1"},
			nodeTaintAfterDuration: map[string]time.Duration{"n1": 100 * time.Millisecond},
			wantLatency:            100 * time.Millisecond,
			wantResultChanOpen:     true,
		},
		{
			description:            "latency when tainting multiple nodes - nodes are tainted in the first calls to the lister",
			startTime:              time.Now(),
			nodes:                  []string{"n1", "n2"},
			nodeTaintAfterDuration: map[string]time.Duration{"n1": 100 * time.Millisecond, "n2": 150 * time.Millisecond},
			wantLatency:            150 * time.Millisecond,
			wantResultChanOpen:     true,
		},
		{
			description:            "latency when tainting multiple nodes - nodes are not tainted in the first calls to the lister",
			startTime:              time.Now(),
			nodes:                  []string{"n1", "n2"},
			nodeTaintAfterDuration: map[string]time.Duration{"n1": 100 * time.Millisecond, "n2": 150 * time.Millisecond},
			wantLatency:            150 * time.Millisecond,
			wantResultChanOpen:     true,
		},
		{
			description:            "Some nodes fails to taint before timeout",
			startTime:              time.Now(),
			nodes:                  []string{"n1", "n3"},
			nodeTaintAfterDuration: map[string]time.Duration{"n1": 100 * time.Millisecond},
			wantResultChanOpen:     false,
		},
		{
			description:               "Expected count is zero resulting in channel closure",
			startTime:                 time.Now(),
			nodes:                     []string{"n1"},
			nodeTaintAfterDuration:    map[string]time.Duration{},
			wantResultChanOpen:        false,
			simulateZeroExpectedCount: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			mc := NewMockClock(tc.startTime)
			nodes := map[string]*apiv1.Node{}
			for _, name := range tc.nodes {
				node := test.BuildTestNode(name, 100, 100)
				nodes[name] = node
			}
			nodeLister := NewTestCustomNodeLister(nodes, tc.nodeTaintAfterDuration, mc)
			updateLatencyTracker := NewUpdateLatencyTrackerForTesting(nodeLister, mc.Now)

			// Synthetically advance mock clock upon each sleep.
			updateLatencyTracker.sleep = func(d time.Duration) {
				mc.SetTime(mc.Now().Add(d))
				// We must explicitly yield because advancing a fake clock doesn't block.
				// Without this, the Start() loop could starve the main test thread on single-core setups (e.g. GOMAXPROCS=1).
				runtime.Gosched()
			}
			go updateLatencyTracker.Start()
			for _, node := range nodes {
				updateLatencyTracker.StartTimeChan <- nodeTaintStartTime{node.Name, tc.startTime}
			}
			if tc.simulateZeroExpectedCount {
				close(updateLatencyTracker.ExpectedNodeCountChan)
				// Wait slightly to ensure Start() loop processes the closed channel.
				time.Sleep(50 * time.Millisecond)
			} else {
				updateLatencyTracker.ExpectedNodeCountChan <- len(tc.nodes)

				latency, ok := <-updateLatencyTracker.ResultChan
				assert.Equal(t, tc.wantResultChanOpen, ok)
				if ok {
					assert.Equal(t, tc.wantLatency, latency)
				}
			}
		})
	}
}
