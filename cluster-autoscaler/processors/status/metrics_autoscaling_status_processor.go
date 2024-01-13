package status

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/utils/backoff"
)

const (
	// unknownErrorCode means that the cloud provider has not provided an error code.
	unknownErrorCode = "unknown"
)

// BackoffReasonStatus contains information about backoff status and reason
type BackoffReasonStatus map[string]bool

// MetricsAutoscalingStatusProcessor is used to update metrics after each autoscaling iteration.
type MetricsAutoscalingStatusProcessor struct {
	backoffReasonStatus map[string]BackoffReasonStatus
}

// Process queries the health status and backoff situation of all node groups and updates metrics after each autoscaling iteration.
func (p *MetricsAutoscalingStatusProcessor) Process(context *context.AutoscalingContext, csr *clusterstate.ClusterStateRegistry, now time.Time) error {
	for _, nodeGroup := range context.CloudProvider.NodeGroups() {
		if !nodeGroup.Exist() {
			continue
		}
		metrics.UpdateNodeGroupHealthStatus(nodeGroup.Id(), csr.IsNodeGroupHealthy(nodeGroup.Id()))
		backoffStatus := csr.BackoffStatusForNodeGroup(nodeGroup, now)
		p.updateNodeGroupBackoffStatusMetrics(nodeGroup.Id(), backoffStatus)
	}
	return nil
}

// CleanUp cleans up the processor's internal structures.
func (p *MetricsAutoscalingStatusProcessor) CleanUp() {
}

// updateNodeGroupBackoffStatusMetrics updates metrics about backoff situation and reason of the node group
func (p *MetricsAutoscalingStatusProcessor) updateNodeGroupBackoffStatusMetrics(nodeGroup string, backoffStatus backoff.Status) {
	if _, ok := p.backoffReasonStatus[nodeGroup]; ok {
		for reason := range p.backoffReasonStatus[nodeGroup] {
			p.backoffReasonStatus[nodeGroup][reason] = false
		}
	} else {
		p.backoffReasonStatus[nodeGroup] = make(BackoffReasonStatus)
	}

	if backoffStatus.IsBackedOff {
		errorCode := backoffStatus.ErrorInfo.ErrorCode
		if errorCode == "" {
			// prevent error code from being empty.
			errorCode = unknownErrorCode
		}
		p.backoffReasonStatus[nodeGroup][errorCode] = true
	}
	metrics.UpdateNodeGroupBackOffStatus(nodeGroup, p.backoffReasonStatus[nodeGroup])
}
