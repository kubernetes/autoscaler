package latencytracker

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/metrics"

	"k8s.io/klog/v2"
)

type NodeInfo struct {
	Name          string
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

func (t *NodeLatencyTracker) UpdateStateWithUnneededList(list []NodeInfo, timestamp time.Time) {
	currentSet := make(map[string]struct{}, len(list))
	for _, info := range list {
		currentSet[info.Name] = struct{}{}
		_, exists := t.nodes[info.Name]
		if !exists {
			t.nodes[info.Name] = NodeInfo{
				Name:          info.Name,
				UnneededSince: info.UnneededSince,
				Threshold:     info.Threshold,
			}
			klog.V(2).Infof("Started tracking unneeded node %s at %v with ScaleDownUnneededTime=%v",
				info.Name, info.UnneededSince, info.Threshold)
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
