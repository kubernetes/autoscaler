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
	"maps"
	"slices"
	"time"

	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/klog/v2"

	processor "k8s.io/autoscaler/cluster-autoscaler/processors/status"
)

const (
	// scaleDownLatencyLogThreshold is the duration after which a scale-down
	// deletion is considered "slow". Deletions that take
	// longer than this threshold will be logged at a more visible level
	scaleDownLatencyLogThreshold = 3 * time.Minute
)

type unneededNodeState struct {
	unneededSince     time.Time
	removalThreshold  time.Duration
	latestDelayReason string
}

// NodeLatencyTracker keeps track of nodes that are marked as unneeded, when they became unneeded,
// and removalThresholds to emit node removal latency metrics.
type NodeLatencyTracker struct {
	unneededNodes map[string]unneededNodeState
	wrapped       processor.ScaleDownStatusProcessor
}

// NewNodeLatencyTracker creates a new tracker.
func NewNodeLatencyTracker(wrapped processor.ScaleDownStatusProcessor) *NodeLatencyTracker {
	return &NodeLatencyTracker{
		unneededNodes: make(map[string]unneededNodeState),
		wrapped:       wrapped,
	}
}

// UpdateScaleDownCandidates updates tracked unneeded nodes and reports those that became needed again.
func (t *NodeLatencyTracker) UpdateScaleDownCandidates(list []*scaledown.UnneededNode, timestamp time.Time) {
	currentSet := make(map[string]struct{}, len(list))
	for _, candidate := range list {
		nodeName := candidate.Node.Name
		currentSet[nodeName] = struct{}{}
		if info, exists := t.unneededNodes[nodeName]; !exists {
			t.unneededNodes[nodeName] = unneededNodeState{
				unneededSince:    timestamp,
				removalThreshold: candidate.RemovalThreshold,
			}
			klog.V(6).Infof("Started tracking unneeded node %s at %v with removal threshold %v.", nodeName, timestamp, candidate.RemovalThreshold)
		} else {
			if info.removalThreshold != candidate.RemovalThreshold {
				info.removalThreshold = candidate.RemovalThreshold
				t.unneededNodes[nodeName] = info
				klog.V(6).Infof("Updated removal threshold for tracked node %s to %v.", nodeName, candidate.RemovalThreshold)
			}
		}
	}
	for nodeName := range t.unneededNodes {
		if _, exists := currentSet[nodeName]; !exists {
			t.recordAndCleanup(nodeName, false)
		}
	}
}

// Process updates unremovableNodes and reports node removal latency based on scale-down status.
func (t *NodeLatencyTracker) Process(autoscalingCtx *ca_context.AutoscalingContext, status *status.ScaleDownStatus) {
	if t.wrapped != nil {
		t.wrapped.Process(autoscalingCtx, status)
	}

	t.updateLatestDelayReasons(status.UnremovableNodes)

	for _, node := range status.ScaledDownNodes {
		t.recordAndCleanup(node.Node.Name, true)
	}

	if klog.V(6).Enabled() {
		for nodeName := range t.unneededNodes {
			klog.Infof("Node %q remains in unneeded list (not scaled down). Continuing to track latency.", nodeName)
		}
	}
}

// recordAndCleanup calculates the time a node spent in the "unneeded" state, updates
// relevant Prometheus metrics, and removes the node from internal tracking.
func (t *NodeLatencyTracker) recordAndCleanup(nodeName string, isRemoved bool) {
	info, exists := t.unneededNodes[nodeName]
	if !exists {
		return
	}
	delete(t.unneededNodes, nodeName)

	duration := time.Since(info.unneededSince)
	latency := duration - info.removalThreshold

	delayReason := info.latestDelayReason
	if delayReason == "" {
		delayReason = "none"
	}

	if latency > 0 {
		metrics.UpdateScaleDownNodeRemovalLatency(isRemoved, delayReason, latency)
	} else {
		klog.V(6).Infof("Node %q was unneeded for %s (threshold %s). Latency %s is <= 0, skipping metric. isRemoved: %v, delayReason: %v",
			nodeName, duration, info.removalThreshold, latency, isRemoved, delayReason)
	}
	if isRemoved {
		t.logDeletion(nodeName, duration, info.removalThreshold, latency)
	} else {
		klog.V(4).Infof("Node %q became needed again (unneeded for %s). Blocker: %q. Latency: %s",
			nodeName, duration, delayReason, latency)
	}
}

// logDeletion handles the logging for scaled-down nodes,
// using a higher verbosity (V2) if the latency exceeds the configured threshold.
func (t *NodeLatencyTracker) logDeletion(nodeName string, duration, threshold, latency time.Duration) {
	level := klog.Level(6)
	if latency > scaleDownLatencyLogThreshold {
		level = klog.Level(2)
	}
	klog.V(level).Infof("Observing deletion for node %s, unneeded for %s (removal threshold was %s).",
		nodeName, duration, threshold)
}

// getTrackedNodes returns the names of all nodes currently tracked as unneeded.
func (t *NodeLatencyTracker) getTrackedNodes() []string {
	return slices.Collect(maps.Keys(t.unneededNodes))
}

// CleanUp cleans up internal structures.
func (t *NodeLatencyTracker) CleanUp() {
	t.unneededNodes = make(map[string]unneededNodeState)
	if t.wrapped != nil {
		t.wrapped.CleanUp()
	}
}

func (t *NodeLatencyTracker) updateLatestDelayReasons(unremovableNodes []*status.UnremovableNode) {
	for _, val := range unremovableNodes {
		if info, exists := t.unneededNodes[val.Node.Name]; exists {
			if isBlocker(val.Reason) {
				reasonStr := val.Reason.String()
				info.latestDelayReason = reasonStr
				t.unneededNodes[val.Node.Name] = info
				klog.V(6).Infof("Tracking latest delay reason %s for node %s.", reasonStr, val.Node.Name)
			}
		}
	}
}

func isBlocker(reason simulator.UnremovableReason) bool {
	return reason != simulator.NoReason && reason != simulator.NotUnneededLongEnough && reason != simulator.NotUnreadyLongEnough
}
