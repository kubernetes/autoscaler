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

package latencytracker

import (
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
)

func TestNodeLatencyTracker(t *testing.T) {
	baseTime := time.Now()
	defaultNodeState := func() unneededNodeState {
		return unneededNodeState{unneededSince: baseTime, threshold: 0}
	}

	tests := []struct {
		name             string
		setupNodes       map[string]unneededNodeState
		updateThresholds map[string]time.Duration
		unneededList     []string
		scaledDownList   []string
		unremovableList  []string
		wantTrackedNodes []string
	}{
		{
			name:             "add new unneeded nodes",
			setupNodes:       map[string]unneededNodeState{},
			unneededList:     []string{"node1", "node2"},
			scaledDownList:   []string{},
			unremovableList:  []string{},
			wantTrackedNodes: []string{"node1", "node2"},
		},
		{
			name:             "node is scaled down",
			setupNodes:       map[string]unneededNodeState{"node1": defaultNodeState()},
			unneededList:     []string{},
			scaledDownList:   []string{"node1"},
			unremovableList:  []string{},
			wantTrackedNodes: []string{},
		},
		{
			name:             "node becomes needed",
			setupNodes:       map[string]unneededNodeState{"node1": defaultNodeState()},
			unneededList:     []string{},
			scaledDownList:   []string{},
			unremovableList:  []string{"node1"},
			wantTrackedNodes: []string{},
		},
		{
			name:             "node is skipped",
			setupNodes:       map[string]unneededNodeState{"node1": defaultNodeState()},
			unneededList:     []string{},
			scaledDownList:   []string{},
			unremovableList:  []string{},
			wantTrackedNodes: []string{"node1"},
		},
		{
			name:             "node is still unneeded",
			setupNodes:       map[string]unneededNodeState{"node1": defaultNodeState()},
			unneededList:     []string{"node1"},
			scaledDownList:   []string{},
			unremovableList:  []string{},
			wantTrackedNodes: []string{"node1"},
		},
		{
			name: "mix of different nodes",
			setupNodes: map[string]unneededNodeState{
				"node1": defaultNodeState(),
				"node2": defaultNodeState(),
				"node3": defaultNodeState(),
				"node4": defaultNodeState(),
			},
			unneededList:     []string{"node1", "node5"},
			scaledDownList:   []string{"node3"},
			unremovableList:  []string{"node2"},
			wantTrackedNodes: []string{"node1", "node4", "node5"},
		},
		{
			name:             "update threshold for known node",
			setupNodes:       map[string]unneededNodeState{"node1": {unneededSince: baseTime, threshold: 1 * time.Second}},
			updateThresholds: map[string]time.Duration{"node1": 4 * time.Second},
			unneededList:     []string{"node1"},
			scaledDownList:   []string{},
			unremovableList:  []string{},
			wantTrackedNodes: []string{"node1"},
		},
		{
			name:             "process node that was never tracked",
			setupNodes:       map[string]unneededNodeState{},
			unneededList:     []string{"node1"},
			scaledDownList:   []string{"node2"},
			unremovableList:  []string{"node3"},
			wantTrackedNodes: []string{"node1"},
		},
		{
			name:             "update threshold for unknown node",
			setupNodes:       map[string]unneededNodeState{},
			updateThresholds: map[string]time.Duration{"node2": 5 * time.Second},
			unneededList:     []string{"node1"},
			scaledDownList:   []string{},
			unremovableList:  []string{},
			wantTrackedNodes: []string{"node1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewNodeLatencyTracker()
			for name, info := range tt.setupNodes {
				tracker.unneededNodes[name] = info
			}

			for node, threshold := range tt.updateThresholds {
				tracker.UpdateThreshold(node, threshold)
			}

			// Step 1: add unneeded nodes
			unneededNodes := make([]*apiv1.Node, len(tt.unneededList))
			for i, name := range tt.unneededList {
				unneededNodes[i] = &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: name}}
			}
			currentTime := baseTime.Add(5 * time.Second)
			tracker.UpdateScaleDownCandidates(unneededNodes, currentTime)

			// Step 2: simulate ScaleDownStatus
			sd := &status.ScaleDownStatus{}
			for _, name := range tt.scaledDownList {
				sd.ScaledDownNodes = append(sd.ScaledDownNodes, &status.ScaleDownNode{Node: &apiv1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: name},
				}})
			}
			for _, name := range tt.unremovableList {
				sd.UnremovableNodes = append(sd.UnremovableNodes, &status.UnremovableNode{Node: &apiv1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: name},
				}})
			}
			tracker.Process(&ca_context.AutoscalingContext{}, sd)

			// Step 3: Check final state of tracked nodes
			gotTracked := tracker.getTrackedNodes()
			expectedMap := make(map[string]struct{})
			for _, n := range tt.wantTrackedNodes {
				expectedMap[n] = struct{}{}
			}

			gotMap := make(map[string]struct{})
			for _, n := range gotTracked {
				gotMap[n] = struct{}{}
			}

			if len(gotMap) != len(expectedMap) {
				t.Errorf("[%s] incorrect number of tracked nodes. got %v, want %v", tt.name, gotTracked, tt.wantTrackedNodes)
			}

			for n := range expectedMap {
				if _, ok := gotMap[n]; !ok {
					t.Errorf("[%s] expected node %q to be tracked, but was not", tt.name, n)
				}
			}
			for n := range gotMap {
				if _, ok := expectedMap[n]; !ok {
					t.Errorf("[%s] unexpected tracked node %q", tt.name, n)
				}
			}

			// Specific check for threshold update test
			if tt.name == "update threshold for known node" {
				if info, ok := tracker.unneededNodes["node1"]; !ok {
					t.Errorf("[%s] node1 was not tracked", tt.name)
				} else if info.threshold != 4*time.Second {
					t.Errorf("[%s] expected threshold 4s, got %s", tt.name, info.threshold)
				}
			}
		})
	}
}
