/*
Copyright 2018 The Kubernetes Authors.

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

package status

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	klog "k8s.io/klog/v2"
)

// ScaleDownStatus represents the state of scale down.
type ScaleDownStatus struct {
	Result            ScaleDownResult
	ScaledDownNodes   []*ScaleDownNode
	UnremovableNodes  []*UnremovableNode
	RemovedNodeGroups []cloudprovider.NodeGroup
	NodeDeleteResults map[string]NodeDeleteResult
}

// SetUnremovableNodesInfo sets the status of nodes that were found to be unremovable.
func (s *ScaleDownStatus) SetUnremovableNodesInfo(unremovableNodesMap map[string]*simulator.UnremovableNode, nodeUtilizationMap map[string]simulator.UtilizationInfo, cp cloudprovider.CloudProvider) {
	s.UnremovableNodes = make([]*UnremovableNode, 0, len(unremovableNodesMap))

	for _, unremovableNode := range unremovableNodesMap {
		nodeGroup, err := cp.NodeGroupForNode(unremovableNode.Node)
		if err != nil {
			klog.Errorf("Couldn't find node group for unremovable node in cloud provider %s", unremovableNode.Node.Name)
			continue
		}

		var utilInfoPtr *simulator.UtilizationInfo
		if utilInfo, found := nodeUtilizationMap[unremovableNode.Node.Name]; found {
			utilInfoPtr = &utilInfo
			// It's okay if we don't find the util info, it's not computed for some unremovable nodes that are skipped early in the loop.
		}

		s.UnremovableNodes = append(s.UnremovableNodes, &UnremovableNode{
			Node:        unremovableNode.Node,
			NodeGroup:   nodeGroup,
			UtilInfo:    utilInfoPtr,
			Reason:      unremovableNode.Reason,
			BlockingPod: unremovableNode.BlockingPod,
		})
	}
}

// UnremovableNode represents the state of a node that couldn't be removed.
type UnremovableNode struct {
	Node        *apiv1.Node
	NodeGroup   cloudprovider.NodeGroup
	UtilInfo    *simulator.UtilizationInfo
	Reason      simulator.UnremovableReason
	BlockingPod *drain.BlockingPod
}

// ScaleDownNode represents the state of a node that's being scaled down.
type ScaleDownNode struct {
	Node        *apiv1.Node
	NodeGroup   cloudprovider.NodeGroup
	EvictedPods []*apiv1.Pod
	UtilInfo    simulator.UtilizationInfo
}

// ScaleDownResult represents the result of scale down.
type ScaleDownResult int

const (
	// ScaleDownError - scale down finished with error.
	ScaleDownError ScaleDownResult = iota
	// ScaleDownNoUnneeded - no unneeded nodes and no errors.
	ScaleDownNoUnneeded
	// ScaleDownNoNodeDeleted - unneeded nodes present but not available for deletion.
	ScaleDownNoNodeDeleted
	// ScaleDownNodeDeleteStarted - a node deletion process was started.
	ScaleDownNodeDeleteStarted
	// ScaleDownNotTried - the scale down wasn't even attempted, e.g. an autoscaling iteration was skipped, or
	// an error occurred before the scale up logic.
	ScaleDownNotTried
	// ScaleDownInCooldown - the scale down wasn't even attempted, because it's in a cooldown state (it's suspended for a scheduled period of time).
	ScaleDownInCooldown
	// ScaleDownInProgress - the scale down wasn't attempted, because a previous scale-down was still in progress.
	ScaleDownInProgress
)

// NodeDeleteResultType denotes the type of the result of node deletion. It provides deeper
// insight into why the node failed to be deleted.
type NodeDeleteResultType int

const (
	// NodeDeleteOk - the node was deleted successfully.
	NodeDeleteOk NodeDeleteResultType = iota

	// NodeDeleteErrorFailedToMarkToBeDeleted - node deletion failed because the node couldn't be marked to be deleted.
	NodeDeleteErrorFailedToMarkToBeDeleted
	// NodeDeleteErrorFailedToEvictPods - node deletion failed because some of the pods couldn't be evicted from the node.
	NodeDeleteErrorFailedToEvictPods
	// NodeDeleteErrorFailedToDelete - failed to delete the node from the cloud provider.
	NodeDeleteErrorFailedToDelete
)

// NodeDeleteResult contains information about the result of a node deletion.
type NodeDeleteResult struct {
	// Err contains nil if the delete was successful and an error otherwise.
	Err error
	// ResultType contains the type of the result of a node deletion.
	ResultType NodeDeleteResultType
	// PodEvictionResults maps pod names to the result of their eviction.
	PodEvictionResults map[string]PodEvictionResult
}

// ScaleDownStatusProcessor processes the status of the cluster after a scale-down.
type ScaleDownStatusProcessor interface {
	Process(context *context.AutoscalingContext, status *ScaleDownStatus)
	CleanUp()
}

// NewDefaultScaleDownStatusProcessor creates a default instance of ScaleUpStatusProcessor.
func NewDefaultScaleDownStatusProcessor() ScaleDownStatusProcessor {
	return &NoOpScaleDownStatusProcessor{}
}

// PodEvictionResult contains the result of an eviction of a pod.
type PodEvictionResult struct {
	Pod      *apiv1.Pod
	TimedOut bool
	Err      error
}

// WasEvictionSuccessful tells if the pod was successfully evicted.
func (per PodEvictionResult) WasEvictionSuccessful() bool {
	return per.Err == nil && !per.TimedOut
}

// NoOpScaleDownStatusProcessor is a ScaleDownStatusProcessor implementations useful for testing.
type NoOpScaleDownStatusProcessor struct{}

// Process processes the status of the cluster after a scale-down.
func (p *NoOpScaleDownStatusProcessor) Process(context *context.AutoscalingContext, status *ScaleDownStatus) {
}

// CleanUp cleans up the processor's internal structures.
func (p *NoOpScaleDownStatusProcessor) CleanUp() {
}
