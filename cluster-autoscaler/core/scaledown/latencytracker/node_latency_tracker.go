/*
Copyright 2019 The Kubernetes Authors.

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
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/klog/v2"
)

// LatencyTracker defines the interface for tracking node removal latency.
// Implementations record when nodes become unneeded, observe deletion events,
// and expose thresholds for measuring node removal duration.
type LatencyTracker interface {
	ObserveDeletion(nodeName string, timestamp time.Time)
	UpdateStateWithUnneededList(list []*apiv1.Node, currentlyInDeletion map[string]bool, timestamp time.Time)
	UpdateThreshold(nodeName string, threshold time.Duration)
	GetTrackedNodes() []string
}
type nodeInfo struct {
	unneededSince time.Time
	threshold     time.Duration
}

// NodeLatencyTracker is a concrete implementation of LatencyTracker.
// It keeps track of nodes that are marked as unneeded, when they became unneeded,
// and thresholds to adjust node removal latency metrics.
type NodeLatencyTracker struct {
	nodes map[string]nodeInfo
}

// NewNodeLatencyTracker creates a new tracker.
func NewNodeLatencyTracker() *NodeLatencyTracker {
	return &NodeLatencyTracker{
		nodes: make(map[string]nodeInfo),
	}
}

// UpdateStateWithUnneededList records unneeded nodes and handles missing ones.
func (t *NodeLatencyTracker) UpdateStateWithUnneededList(
	list []*apiv1.Node,
	currentlyInDeletion map[string]bool,
	timestamp time.Time,
) {
	currentSet := make(map[string]struct{}, len(list))
	for _, node := range list {
		currentSet[node.Name] = struct{}{}

		if _, exists := t.nodes[node.Name]; !exists {
			t.nodes[node.Name] = nodeInfo{
				unneededSince: timestamp,
				threshold:     0,
			}
			klog.V(4).Infof("Started tracking unneeded node %s at %v", node.Name, timestamp)
		}
	}

	for name, info := range t.nodes {
		if _, stillUnneeded := currentSet[name]; !stillUnneeded {
			if _, inDeletion := currentlyInDeletion[name]; !inDeletion {
				duration := timestamp.Sub(info.unneededSince)
				metrics.UpdateScaleDownNodeRemovalLatency(false, duration-info.threshold)
				delete(t.nodes, name)
				klog.V(4).Infof("Node %q reported as deleted/missing (unneeded for %s, threshold %s)",
					name, duration, info.threshold)
			}
		}
	}
}

// ObserveDeletion is called by the actuator just before node deletion.
func (t *NodeLatencyTracker) ObserveDeletion(nodeName string, timestamp time.Time) {
	if info, exists := t.nodes[nodeName]; exists {
		duration := timestamp.Sub(info.unneededSince)

		klog.V(4).Infof(
			"Observing deletion for node %s, unneeded for %s (threshold was %s).",
			nodeName, duration, info.threshold,
		)

		metrics.UpdateScaleDownNodeRemovalLatency(true, duration-info.threshold)
		delete(t.nodes, nodeName)
	}
}

// UpdateThreshold updates the scale-down threshold for a tracked node.
func (t *NodeLatencyTracker) UpdateThreshold(nodeName string, threshold time.Duration) {
	if info, exists := t.nodes[nodeName]; exists {
		info.threshold = threshold
		t.nodes[nodeName] = info
		klog.V(4).Infof("Updated threshold for node %q to %s", nodeName, threshold)
	} else {
		klog.Warningf("Attempted to update threshold for unknown node %q", nodeName)
	}
}

// GetTrackedNodes returns the names of all nodes currently tracked as unneeded.
func (t *NodeLatencyTracker) GetTrackedNodes() []string {
	names := make([]string, 0, len(t.nodes))
	for name := range t.nodes {
		names = append(names, name)
	}
	return names
}
