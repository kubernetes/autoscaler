/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
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

	tests := []struct {
		name             string
		setupNodes       map[string]unneededNodeState
		unneededList     []string
		unremovableList  []string
		updateThresholds map[string]time.Duration
		observeDeletions []string
		wantTrackedNodes []string
	}{
		{
			name:             "add new unneeded nodes",
			setupNodes:       map[string]unneededNodeState{},
			unneededList:     []string{"node1", "node2"},
			unremovableList:  []string{},
			updateThresholds: map[string]time.Duration{},
			wantTrackedNodes: []string{"node1", "node2"},
		},
		{
			name: "observe deletion with threshold",
			setupNodes: map[string]unneededNodeState{
				"node1": {unneededSince: baseTime, threshold: 2 * time.Second},
			},
			unneededList:     []string{},
			unremovableList:  []string{},
			observeDeletions: []string{"node1"},
			wantTrackedNodes: []string{},
		},
		{
			name: "node removed from unneeded but not unremovable",
			setupNodes: map[string]unneededNodeState{
				"node1": {unneededSince: baseTime, threshold: 1 * time.Second},
				"node2": {unneededSince: baseTime, threshold: 0},
			},
			unneededList:     []string{"node1"},
			unremovableList:  []string{"node2"},
			wantTrackedNodes: []string{"node1"},
		},
		{
			name: "update threshold",
			setupNodes: map[string]unneededNodeState{
				"node1": {unneededSince: baseTime, threshold: 1 * time.Second},
			},
			unneededList:     []string{"node1"},
			updateThresholds: map[string]time.Duration{"node1": 4 * time.Second},
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
			for _, name := range tt.observeDeletions {
				sd.ScaledDownNodes = append(sd.ScaledDownNodes, &status.ScaleDownNode{Node: &apiv1.Node{
					ObjectMeta: metav1.ObjectMeta{Name: name},
				}})
			}
			for _, name := range tt.unremovableList {
				sd.UnremovableNodes = append(sd.UnremovableNodes, &status.UnremovableNode{
					Node: &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: name}},
				})
			}

			// Process unremovable/deleted nodes
			tracker.Process(&ca_context.AutoscalingContext{}, sd)

			// Step 3: call UpdateScaleDownCandidates again with empty list to remove nodes
			tracker.UpdateScaleDownCandidates([]*apiv1.Node{}, currentTime.Add(5*time.Second))

			// Check tracked nodes
			gotTracked := tracker.getTrackedNodes()
			expectedMap := make(map[string]struct{})
			for _, n := range tt.wantTrackedNodes {
				expectedMap[n] = struct{}{}
			}
			for _, n := range gotTracked {
				if _, ok := expectedMap[n]; !ok {
					t.Errorf("[%s] unexpected tracked node %q", tt.name, n)
				}
				delete(expectedMap, n)
			}
			for n := range expectedMap {
				t.Errorf("[%s] expected node %q to be tracked, but was not", tt.name, n)
			}
		})
	}
}
