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
	"time"

	apiv1 "k8s.io/api/core/v1"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/klog/v2"
)

type unneededNodeState struct {
	unneededSince time.Time
	threshold     time.Duration
}

// NodeLatencyTracker is a concrete implementation of LatencyTracker.
// It keeps track of nodes that are marked as unneeded, when they became unneeded,
// and thresholds to adjust node removal latency metrics.
type NodeLatencyTracker struct {
	unneededNodes map[string]unneededNodeState
}

// NewNodeLatencyTracker creates a new tracker.
func NewNodeLatencyTracker() *NodeLatencyTracker {
	return &NodeLatencyTracker{
		unneededNodes: make(map[string]unneededNodeState),
	}
}

// UpdateScaleDownCandidates updates tracked unneeded nodes and reports those that became needed again.
func (t *NodeLatencyTracker) UpdateScaleDownCandidates(list []*apiv1.Node, timestamp time.Time) {
	currentSet := make(map[string]struct{}, len(list))
	for _, node := range list {
		currentSet[node.Name] = struct{}{}
		if _, exists := t.unneededNodes[node.Name]; !exists {
			t.unneededNodes[node.Name] = unneededNodeState{
				unneededSince: timestamp,
				threshold:     0,
			}
			klog.V(6).Infof("Started tracking unneeded node %s at %v", node.Name, timestamp)
		}
	}
}

// Process updates unremovableNodes and reports node removal latency based on scale-down status.
func (t *NodeLatencyTracker) Process(autoscalingCtx *ca_context.AutoscalingContext, status *status.ScaleDownStatus) {
	for _, unremovableNode := range status.UnremovableNodes {
		nodeName := unremovableNode.Node.Name
		if info, exists := t.unneededNodes[nodeName]; exists {
			duration := time.Since(info.unneededSince)
			metrics.UpdateScaleDownNodeRemovalLatency(false, duration)
			klog.V(4).Infof("Node %q is unremovable, became needed again (unneeded for %s)", nodeName, duration)
			delete(t.unneededNodes, nodeName)
		}
	}
	for _, scaledDownNode := range status.ScaledDownNodes {
		nodeName := scaledDownNode.Node.Name
		if info, exists := t.unneededNodes[nodeName]; exists {
			duration := time.Since(info.unneededSince)
			metrics.UpdateScaleDownNodeRemovalLatency(true, duration-info.threshold)
			if duration > 3*time.Minute {
				klog.V(2).Infof(
					"Observing deletion for node %s, unneeded for %s (threshold was %s).",
					nodeName, duration, info.threshold,
				)
			} else {
				klog.V(6).Infof(
					"Observing deletion for node %s, unneeded for %s (threshold was %s).",
					nodeName, duration, info.threshold,
				)
			}
			delete(t.unneededNodes, nodeName)
		}
	}
	for nodeName := range t.unneededNodes {
		klog.V(6).Infof("Node %q was skipped during a current scale-down loop (not unremovable, scale-down not started), continuing to track latency.", nodeName)
	}
}

// UpdateThreshold updates the scale-down threshold for a tracked node.
func (t *NodeLatencyTracker) UpdateThreshold(nodeName string, threshold time.Duration) {
	if info, exists := t.unneededNodes[nodeName]; exists {
		info.threshold = threshold
		t.unneededNodes[nodeName] = info
		klog.V(6).Infof("Updated threshold for node %q to %s", nodeName, threshold)
	} else {
		klog.Warningf("Attempted to update threshold for unknown node %q", nodeName)
	}
}

// getTrackedNodes returns the names of all nodes currently tracked as unneeded.
func (t *NodeLatencyTracker) getTrackedNodes() []string {
	names := make([]string, 0, len(t.unneededNodes))
	for name := range t.unneededNodes {
		names = append(names, name)
	}
	return names
}

// CleanUp cleans up internal structures.
func (t *NodeLatencyTracker) CleanUp() {
	t.unneededNodes = make(map[string]unneededNodeState)
}
