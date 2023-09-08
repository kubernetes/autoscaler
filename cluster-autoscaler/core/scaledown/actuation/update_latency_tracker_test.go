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
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

// mockClock is used to mock time.Now() when testing UpdateLatencyTracker
// For the n th call to Now() it will return a timestamp after duration[n] to
// the startTime if n < the length of durations. Otherwise, it will return current time.
type mockClock struct {
	startTime time.Time
	durations []time.Duration
	index     int
	mutex     sync.Mutex
}

// Returns a new NewMockClock object
func NewMockClock(startTime time.Time, durations []time.Duration) mockClock {
	return mockClock{
		startTime: startTime,
		durations: durations,
		index:     0,
	}
}

// Returns a time after Nth duration from the start time if N < length of durations.
// Otherwise, returns the current time
func (m *mockClock) Now() time.Time {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	var timeToSend time.Time
	if m.index < len(m.durations) {
		timeToSend = m.startTime.Add(m.durations[m.index])
	} else {
		timeToSend = time.Now()
	}
	m.index += 1
	return timeToSend
}

// Returns the number of times that the Now function was called
func (m *mockClock) getIndex() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.index
}

// TestCustomNodeLister can be used to mock nodeLister Get call when testing delayed tainting
type TestCustomNodeLister struct {
	nodes                    map[string]*apiv1.Node
	getCallCount             map[string]int
	nodeTaintAfterNthGetCall map[string]int
}

// List returns all nodes in test lister.
func (l *TestCustomNodeLister) List() ([]*apiv1.Node, error) {
	var nodes []*apiv1.Node
	for _, node := range l.nodes {
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// Get returns node from test lister. Add ToBeDeletedTaint to the node
// during the N th call specified in the nodeTaintAfterNthGetCall
func (l *TestCustomNodeLister) Get(name string) (*apiv1.Node, error) {
	for _, node := range l.nodes {
		if node.Name == name {
			l.getCallCount[node.Name] += 1
			if _, ok := l.nodeTaintAfterNthGetCall[node.Name]; ok && l.getCallCount[node.Name] == l.nodeTaintAfterNthGetCall[node.Name] {
				toBeDeletedTaint := apiv1.Taint{Key: taints.ToBeDeletedTaint, Effect: apiv1.TaintEffectNoSchedule}
				node.Spec.Taints = append(node.Spec.Taints, toBeDeletedTaint)
			}
			return node, nil
		}
	}
	return nil, fmt.Errorf("Node %s not found", name)
}

// Return new TestCustomNodeLister object
func NewTestCustomNodeLister(nodes map[string]*apiv1.Node, nodeTaintAfterNthGetCall map[string]int) *TestCustomNodeLister {
	getCallCounts := map[string]int{}
	for name := range nodes {
		getCallCounts[name] = 0
	}
	return &TestCustomNodeLister{
		nodes:                    nodes,
		getCallCount:             getCallCounts,
		nodeTaintAfterNthGetCall: nodeTaintAfterNthGetCall,
	}
}

func TestUpdateLatencyCalculation(t *testing.T) {

	testCases := []struct {
		description string
		startTime   time.Time
		nodes       []string
		// If an entry is not added for a node, that node will never get tainted
		nodeTaintAfterNthGetCall map[string]int
		durations                []time.Duration
		wantLatency              time.Duration
		wantResultChanOpen       bool
	}{
		{
			description:              "latency when tainting a single node - node is tainted in the first call to the lister",
			startTime:                time.Now(),
			nodes:                    []string{"n1"},
			nodeTaintAfterNthGetCall: map[string]int{"n1": 1},
			durations:                []time.Duration{100 * time.Millisecond},
			wantLatency:              100 * time.Millisecond,
			wantResultChanOpen:       true,
		},
		{
			description:              "latency when tainting a single node - node is not tainted in the first call to the lister",
			startTime:                time.Now(),
			nodes:                    []string{"n1"},
			nodeTaintAfterNthGetCall: map[string]int{"n1": 3},
			durations:                []time.Duration{100 * time.Millisecond},
			wantLatency:              100 * time.Millisecond,
			wantResultChanOpen:       true,
		},
		{
			description:              "latency when tainting multiple nodes - nodes are tainted in the first calls to the lister",
			startTime:                time.Now(),
			nodes:                    []string{"n1", "n2"},
			nodeTaintAfterNthGetCall: map[string]int{"n1": 1, "n2": 1},
			durations:                []time.Duration{100 * time.Millisecond, 150 * time.Millisecond},
			wantLatency:              150 * time.Millisecond,
			wantResultChanOpen:       true,
		},
		{
			description:              "latency when tainting multiple nodes - nodes are not tainted in the first calls to the lister",
			startTime:                time.Now(),
			nodes:                    []string{"n1", "n2"},
			nodeTaintAfterNthGetCall: map[string]int{"n1": 3, "n2": 5},
			durations:                []time.Duration{100 * time.Millisecond, 150 * time.Millisecond},
			wantLatency:              150 * time.Millisecond,
			wantResultChanOpen:       true,
		},
		{
			description:              "Some nodes fails to taint before timeout",
			startTime:                time.Now(),
			nodes:                    []string{"n1", "n3"},
			nodeTaintAfterNthGetCall: map[string]int{"n1": 1},
			durations:                []time.Duration{100 * time.Millisecond, 150 * time.Millisecond},
			wantResultChanOpen:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			mc := NewMockClock(tc.startTime, tc.durations)
			nodes := map[string]*apiv1.Node{}
			for _, name := range tc.nodes {
				node := test.BuildTestNode(name, 100, 100)
				nodes[name] = node
			}
			nodeLister := NewTestCustomNodeLister(nodes, tc.nodeTaintAfterNthGetCall)
			updateLatencyTracker := NewUpdateLatencyTrackerForTesting(nodeLister, mc.Now)
			go updateLatencyTracker.Start()
			for _, node := range nodes {
				updateLatencyTracker.StartTimeChan <- nodeTaintStartTime{node.Name, tc.startTime}
			}
			updateLatencyTracker.AwaitOrStopChan <- true
			latency, ok := <-updateLatencyTracker.ResultChan
			assert.Equal(t, tc.wantResultChanOpen, ok)
			if ok {
				assert.Equal(t, tc.wantLatency, latency)
				assert.Equal(t, len(tc.durations), mc.getIndex())
			}
		})
	}
}
