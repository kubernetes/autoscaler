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
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	processor "k8s.io/autoscaler/cluster-autoscaler/processors/status"
)

const testStepDuration = 1 * time.Minute

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

func TestNodeLatencyTracker_SimulationLoop(t *testing.T) {
	// A step represents one loop of the autoscaler:
	// 1. Planner calculates unneeded nodes (UpdateScaleDownCandidates)
	// 2. Actuator attempts deletion and reports status (Process)
	type step struct {
		unneededList     []string                 // Nodes found unneeded this loop
		thresholds       map[string]time.Duration // Specific thresholds for this loop
		scaledDownList   []string                 // Nodes successfully deleted this loop
		unremovableList  []string                 // Nodes failed to delete this loop
		wantTrackedNodes []string                 // Expected internal state after this loop
		wantThresholds   map[string]time.Duration // Expected threshold values in internal state
	}

	tests := []struct {
		name  string
		steps []step
	}{
		{
			name: "Standard lifecycle: Unneeded -> Tracked -> Deleted",
			steps: []step{
				{
					unneededList:     []string{"node1", "node2"},
					wantTrackedNodes: []string{"node1", "node2"},
				},
				{
					unneededList:     []string{"node1", "node2"},
					scaledDownList:   []string{"node1"},
					wantTrackedNodes: []string{"node2"},
				},
			},
		},
		{
			name: "Node becomes needed again (disappears from list)",
			steps: []step{
				{
					unneededList:     []string{"node1", "node2"},
					wantTrackedNodes: []string{"node1", "node2"},
				},
				{
					unneededList:     []string{"node2"},
					wantTrackedNodes: []string{"node2"},
				},
			},
		},
		{
			name: "Node becomes needed again (reported unremovable)",
			steps: []step{
				{
					unneededList:     []string{"node1"},
					wantTrackedNodes: []string{"node1"},
				},
				{
					unneededList:     []string{"node1"},
					unremovableList:  []string{"node1"},
					wantTrackedNodes: []string{},
				},
			},
		},
		{
			name: "Threshold updates dynamically",
			steps: []step{
				{
					unneededList:     []string{"node1"},
					thresholds:       map[string]time.Duration{"node1": 5 * time.Minute},
					wantTrackedNodes: []string{"node1"},
					wantThresholds:   map[string]time.Duration{"node1": 5 * time.Minute},
				},
				{
					unneededList:     []string{"node1"},
					thresholds:       map[string]time.Duration{"node1": 10 * time.Minute},
					wantTrackedNodes: []string{"node1"},
					wantThresholds:   map[string]time.Duration{"node1": 10 * time.Minute},
				},
			},
		},
		{
			name: "Mix of additions, deletions, and updates",
			steps: []step{
				{
					// Start tracking node1, node2
					unneededList:     []string{"node1", "node2"},
					wantTrackedNodes: []string{"node1", "node2"},
				},
				{
					// node1 gets deleted, node3 appears, node2 stays
					unneededList:     []string{"node1", "node2", "node3"},
					scaledDownList:   []string{"node1"},
					wantTrackedNodes: []string{"node2", "node3"},
				},
				{
					unneededList:     []string{"node3"},
					scaledDownList:   []string{"node2"},
					wantTrackedNodes: []string{"node3"},
				},
			},
		},
	}

	baseTime := time.Now()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tracker := NewNodeLatencyTracker(processor.NewDefaultScaleDownStatusProcessor())

			for i, step := range tc.steps {
				stepTime := baseTime.Add(testStepDuration)

				candidates := candidatesFromNames(step.unneededList, step.thresholds)
				tracker.UpdateScaleDownCandidates(candidates, stepTime)

				sd := newScaleDownStatus(step.scaledDownList, step.unremovableList)
				tracker.Process(&ca_context.AutoscalingContext{}, sd)

				gotTracked := tracker.getTrackedNodes()
				gotSet := sets.New(gotTracked...)
				expectedSet := sets.New(step.wantTrackedNodes...)

				if !gotSet.Equal(expectedSet) {
					t.Errorf("Step %d: incorrect tracked nodes:\ngot:  %v\nwant: %v", i, sets.List(gotSet), sets.List(expectedSet))
				}

				for _, nodeName := range step.wantTrackedNodes {
					wantThreshold := time.Duration(0)
					if val, ok := step.wantThresholds[nodeName]; ok {
						wantThreshold = val
					}

					if actualInfo, ok := tracker.unneededNodes[nodeName]; ok {
						if actualInfo.removalThreshold != wantThreshold {
							t.Errorf("Step %d: node %q incorrect threshold: got %v, want %v", i, nodeName, actualInfo.removalThreshold, wantThreshold)
						}
					} else {
						t.Errorf("Step %d: node %q expected to be tracked but was not", i, nodeName)
					}
				}
			}
		})
	}
}

// nodesFromNames is a test helper to convert a slice of node names and threshold  into a slice of *scaledown.UnneededNode
func candidatesFromNames(names []string, thresholds map[string]time.Duration) []*scaledown.UnneededNode {
	list := make([]*scaledown.UnneededNode, len(names))
	for i, name := range names {
		t := time.Duration(0)
		if val, ok := thresholds[name]; ok {
			t = val
		}
		list[i] = &scaledown.UnneededNode{
			Node:             &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: name}},
			RemovalThreshold: t,
		}
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
