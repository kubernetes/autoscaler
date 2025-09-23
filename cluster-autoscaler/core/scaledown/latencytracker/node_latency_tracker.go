package latencytracker

import (
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/klog/v2"
)

type LatencyTracker interface {
	ObserveDeletion(nodeName string, timestamp time.Time)
	UpdateStateWithUnneededList(list []*apiv1.Node, currentlyInDeletion map[string]bool, timestamp time.Time)
}
type NodeInfo struct {
	UnneededSince time.Time
	Threshold     time.Duration
}

type NodeLatencyTracker struct {
	nodes map[string]NodeInfo
}

// NewNodeLatencyTracker creates a new tracker.
func NewNodeLatencyTracker() *NodeLatencyTracker {
	return &NodeLatencyTracker{
		nodes: make(map[string]NodeInfo),
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
			t.nodes[node.Name] = NodeInfo{
				UnneededSince: timestamp,
				Threshold:     0,
			}
			klog.V(2).Infof("Started tracking unneeded node %s at %v", node.Name, timestamp)
		}
	}

	for name, info := range t.nodes {
		if _, stillUnneeded := currentSet[name]; !stillUnneeded {
			if _, inDeletion := currentlyInDeletion[name]; !inDeletion {
				duration := timestamp.Sub(info.UnneededSince)
				metrics.UpdateScaleDownNodeDeletionDuration("false", duration-info.Threshold)
				delete(t.nodes, name)
				klog.V(2).Infof("Node %q reported as deleted/missing (unneeded for %s, threshold %s)",
					name, duration, info.Threshold)
			}
		}
	}
}

// ObserveDeletion is called by the actuator just before node deletion.
func (t *NodeLatencyTracker) ObserveDeletion(nodeName string, timestamp time.Time) {
	if info, exists := t.nodes[nodeName]; exists {
		duration := timestamp.Sub(info.UnneededSince)

		klog.V(2).Infof(
			"Observing deletion for node %s, unneeded for %s (threshold was %s).",
			nodeName, duration, info.Threshold,
		)

		metrics.UpdateScaleDownNodeDeletionDuration("true", duration-info.Threshold)
		delete(t.nodes, nodeName)
	}
}
