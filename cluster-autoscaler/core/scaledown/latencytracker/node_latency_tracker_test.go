/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package latencytracker

import (
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNodeLatencyTracker(t *testing.T) {
	baseTime := time.Now()

	tests := []struct {
		name                  string
		setupNodes            map[string]NodeInfo
		unneededList          []string
		currentlyInDeletion   map[string]bool
		updateThresholds      map[string]time.Duration
		observeDeletion       []string
		expectedTrackedNodes  []string
		expectedDeletionTimes map[string]time.Duration
	}{
		{
			name:                 "add new unneeded nodes",
			setupNodes:           map[string]NodeInfo{},
			unneededList:         []string{"node1", "node2"},
			currentlyInDeletion:  map[string]bool{},
			updateThresholds:     map[string]time.Duration{},
			observeDeletion:      []string{},
			expectedTrackedNodes: []string{"node1", "node2"},
		},
		{
			name: "observe deletion with threshold",
			setupNodes: map[string]NodeInfo{
				"node1": {UnneededSince: baseTime, Threshold: 2 * time.Second},
			},
			unneededList:         []string{},
			currentlyInDeletion:  map[string]bool{},
			updateThresholds:     map[string]time.Duration{},
			observeDeletion:      []string{"node1"},
			expectedTrackedNodes: []string{},
			expectedDeletionTimes: map[string]time.Duration{
				"node1": 3 * time.Second, // simulate observation 5s after UnneededSince, threshold 2s
			},
		},
		{
			name: "remove unneeded node not in deletion",
			setupNodes: map[string]NodeInfo{
				"node1": {UnneededSince: baseTime, Threshold: 1 * time.Second},
				"node2": {UnneededSince: baseTime, Threshold: 0},
			},
			unneededList:         []string{"node2"}, // node1 is removed from unneeded
			currentlyInDeletion:  map[string]bool{},
			updateThresholds:     map[string]time.Duration{},
			observeDeletion:      []string{},
			expectedTrackedNodes: []string{"node2"},
			expectedDeletionTimes: map[string]time.Duration{
				"node1": 5*time.Second - 1*time.Second, // assume current timestamp baseTime+5s
			},
		},
		{
			name: "update threshold",
			setupNodes: map[string]NodeInfo{
				"node1": {UnneededSince: baseTime, Threshold: 1 * time.Second},
			},
			unneededList:        []string{"node1"},
			currentlyInDeletion: map[string]bool{},
			updateThresholds: map[string]time.Duration{
				"node1": 4 * time.Second,
			},
			observeDeletion:      []string{},
			expectedTrackedNodes: []string{"node1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewNodeLatencyTracker()
			for name, info := range tt.setupNodes {
				tracker.nodes[name] = info
			}

			for node, threshold := range tt.updateThresholds {
				tracker.UpdateThreshold(node, threshold)
			}
			unneededNodes := make([]*apiv1.Node, len(tt.unneededList))
			for i, name := range tt.unneededList {
				unneededNodes[i] = &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: name}}
			}
			// simulate current timestamp as baseTime + 5s
			currentTime := baseTime.Add(5 * time.Second)
			tracker.UpdateStateWithUnneededList(unneededNodes, tt.currentlyInDeletion, currentTime)

			// Observe deletions
			for _, node := range tt.observeDeletion {
				tracker.ObserveDeletion(node, currentTime)
			}

			// Check tracked nodes
			gotTracked := tracker.GetTrackedNodes()
			expectedMap := make(map[string]struct{})
			for _, n := range tt.expectedTrackedNodes {
				expectedMap[n] = struct{}{}
			}
			for _, n := range gotTracked {
				if _, ok := expectedMap[n]; !ok {
					t.Errorf("unexpected tracked node %q", n)
				}
				delete(expectedMap, n)
			}
			for n := range expectedMap {
				t.Errorf("expected node %q to be tracked, but was not", n)
			}

			for node, expectedDuration := range tt.expectedDeletionTimes {
				info, ok := tt.setupNodes[node]
				if !ok {
					continue
				}
				duration := currentTime.Sub(info.UnneededSince) - info.Threshold
				if duration != expectedDuration {
					t.Errorf("node %q expected deletion duration %v, got %v", node, expectedDuration, duration)
				}
			}
		})
	}
}
