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
	"k8s.io/apimachinery/pkg/util/sets"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	processor "k8s.io/autoscaler/cluster-autoscaler/processors/status"
)

func TestNodeLatencyTracker_Decorator(t *testing.T) {
	mock := &mockStatusProcessor{}
	tracker := NewNodeLatencyTracker(mock)

	tracker.Process(&ca_context.AutoscalingContext{}, &status.ScaleDownStatus{})

	if !mock.processCalled {
		t.Errorf("Process() did not call wrapped.Process()")
	}

	tracker.CleanUp()

	if !mock.cleanUpCalled {
		t.Errorf("CleanUp() did not call wrapped.CleanUp()")
	}
}

func TestNodeLatencyTracker(t *testing.T) {
	baseTime := time.Now()

	tests := []struct {
		name             string
		setupNodes       []string
		updateThresholds map[string]time.Duration
		unneededList     []string
		scaledDownList   []string
		unremovableList  []string
		wantTrackedNodes []string
		wantThresholds   map[string]time.Duration
	}{
		{
			name:             "add new unneeded nodes",
			setupNodes:       []string{"node1"},
			unneededList:     []string{"node1", "node2", "node3"},
			scaledDownList:   []string{},
			unremovableList:  []string{},
			wantTrackedNodes: []string{"node1", "node2", "node3"},
		},
		{
			name:             "node is scaled down",
			setupNodes:       []string{},
			unneededList:     []string{"node1"},
			scaledDownList:   []string{"node1"},
			unremovableList:  []string{},
			wantTrackedNodes: []string{},
		},
		{
			name:             "node becomes needed",
			setupNodes:       []string{},
			unneededList:     []string{"node1"},
			scaledDownList:   []string{},
			unremovableList:  []string{"node1"},
			wantTrackedNodes: []string{},
		},
		{
			name:             "node is still unneeded",
			setupNodes:       []string{"node1"},
			unneededList:     []string{"node1"},
			scaledDownList:   []string{},
			unremovableList:  []string{},
			wantTrackedNodes: []string{"node1"},
		},
		{
			name: "mix of different nodes",
			setupNodes: []string{
				"node1",
				"node2",
				"node3",
				"node4",
			},
			unneededList:     []string{"node1", "node4", "node5", "node6"},
			scaledDownList:   []string{"node4"},
			unremovableList:  []string{"node6"},
			wantTrackedNodes: []string{"node1", "node5"},
		},
		{
			name:             "process node that was never tracked",
			setupNodes:       []string{},
			unneededList:     []string{"node1"},
			scaledDownList:   []string{"node2"},
			unremovableList:  []string{"node3"},
			wantTrackedNodes: []string{"node1"},
		},
		{
			name:             "update threshold for known node",
			setupNodes:       []string{},
			updateThresholds: map[string]time.Duration{"node1": 4 * time.Second},
			unneededList:     []string{"node1"},
			scaledDownList:   []string{},
			unremovableList:  []string{},
			wantTrackedNodes: []string{"node1"},
			wantThresholds:   map[string]time.Duration{"node1": 4 * time.Second},
		},
		{
			name:             "update threshold for unknown node",
			updateThresholds: map[string]time.Duration{"node2": 5 * time.Second},
			unneededList:     []string{"node1"},
			scaledDownList:   []string{},
			unremovableList:  []string{},
			wantTrackedNodes: []string{"node1"},
			wantThresholds:   map[string]time.Duration{"node1": 0 * time.Second},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewNodeLatencyTracker(processor.NewDefaultScaleDownStatusProcessor())

			if len(tt.setupNodes) > 0 {
				tracker.UpdateScaleDownCandidates(nodesFromNames(tt.setupNodes), baseTime)
			}

			tracker.UpdateScaleDownCandidates(nodesFromNames(tt.unneededList), baseTime.Add(5*time.Second))

			for node, threshold := range tt.updateThresholds {
				tracker.UpdateThreshold(node, threshold)
			}

			sd := newScaleDownStatus(tt.scaledDownList, tt.unremovableList)
			tracker.Process(&ca_context.AutoscalingContext{}, sd)

			gotTracked := tracker.getTrackedNodes()
			gotSet := sets.New(gotTracked...)
			expectedSet := sets.New(tt.wantTrackedNodes...)

			if !gotSet.Equal(expectedSet) {
				t.Errorf("[%s] incorrect tracked nodes:\ngot:  %v\nwant: %v", tt.name, sets.List(gotSet), sets.List(expectedSet))
			}

			for _, nodeName := range tt.wantTrackedNodes {
				wantThreshold := tt.wantThresholds[nodeName]
				actualThreshold := tracker.unneededNodes[nodeName].threshold

				if actualThreshold != wantThreshold {
					t.Errorf("[%s] node %q: incorrect threshold: got %v, want %v", tt.name, nodeName, actualThreshold, wantThreshold)
				}
			}
		})
	}
}

// nodesFromNames is a test helper to convert a slice of node names into a slice of *apiv1.Node
func nodesFromNames(names []string) []*apiv1.Node {
	list := make([]*apiv1.Node, len(names))
	for i, name := range names {
		list[i] = &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: name}}
	}
	return list
}

// newScaleDownStatus is a test helper to create a ScaleDownStatus object
func newScaleDownStatus(scaledDown, unremovable []string) *status.ScaleDownStatus {
	sd := &status.ScaleDownStatus{}
	for _, name := range scaledDown {
		sd.ScaledDownNodes = append(sd.ScaledDownNodes, &status.ScaleDownNode{Node: &apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: name},
		}})
	}
	for _, name := range unremovable {
		sd.UnremovableNodes = append(sd.UnremovableNodes, &status.UnremovableNode{Node: &apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: name},
		}})
	}
	return sd
}

type mockStatusProcessor struct {
	processCalled bool
	cleanUpCalled bool
}

func (m *mockStatusProcessor) Process(autoscalingCtx *ca_context.AutoscalingContext, status *status.ScaleDownStatus) {
	m.processCalled = true
}

func (m *mockStatusProcessor) CleanUp() {
	m.cleanUpCalled = true
}
