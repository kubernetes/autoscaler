/*
Copyright 2016 The Kubernetes Authors.

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

package simulator

import (
	"errors"
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/pdb"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability/rules"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/autoscaler/cluster-autoscaler/utils/tpu"

	"k8s.io/klog/v2"
)

// NodeToBeRemoved contain information about a node that can be removed.
type NodeToBeRemoved struct {
	// Node to be removed.
	Node *apiv1.Node
	// IsRisky indicates that node has high chance to fail during removal.
	IsRisky bool
	// PodsToReschedule contains pods on the node that should be rescheduled elsewhere.
	PodsToReschedule []*apiv1.Pod
	DaemonSetPods    []*apiv1.Pod
}

// UnremovableNode represents a node that can't be removed by CA.
type UnremovableNode struct {
	Node        *apiv1.Node
	Reason      UnremovableReason
	BlockingPod *drain.BlockingPod
}

// UnremovableReason represents a reason why a node can't be removed by CA.
type UnremovableReason int

const (
	// NoReason - sanity check, this should never be set explicitly. If this is found in the wild, it means that it was
	// implicitly initialized and might indicate a bug.
	NoReason UnremovableReason = iota
	// ScaleDownDisabledAnnotation - node can't be removed because it has a "scale down disabled" annotation.
	ScaleDownDisabledAnnotation
	// ScaleDownUnreadyDisabled - node can't be removed because it is unready and scale down is disabled for unready nodes.
	ScaleDownUnreadyDisabled
	// NotAutoscaled - node can't be removed because it doesn't belong to an autoscaled node group.
	NotAutoscaled
	// NotUnneededLongEnough - node can't be removed because it wasn't unneeded for long enough.
	NotUnneededLongEnough
	// NotUnreadyLongEnough - node can't be removed because it wasn't unready for long enough.
	NotUnreadyLongEnough
	// NodeGroupMinSizeReached - node can't be removed because its node group is at its minimal size already.
	NodeGroupMinSizeReached
	// NodeGroupMaxDeletionCountReached - node can't be removed because max node count to be removed value set in planner reached
	NodeGroupMaxDeletionCountReached
	// AtomicScaleDownFailed - node can't be removed as node group has ZeroOrMaxNodeScaling enabled and number of nodes to remove are not equal to target size
	AtomicScaleDownFailed
	// MinimalResourceLimitExceeded - node can't be removed because it would violate cluster-wide minimal resource limits.
	MinimalResourceLimitExceeded
	// CurrentlyBeingDeleted - node can't be removed because it's already in the process of being deleted.
	CurrentlyBeingDeleted
	// NotUnderutilized - node can't be removed because it's not underutilized.
	NotUnderutilized
	// NotUnneededOtherReason - node can't be removed because it's not marked as unneeded for other reasons (e.g. it wasn't inspected at all in a given autoscaler loop).
	NotUnneededOtherReason
	// RecentlyUnremovable - node can't be removed because it was recently found to be unremovable.
	RecentlyUnremovable
	// NoPlaceToMovePods - node can't be removed because there's no place to move its pods to.
	NoPlaceToMovePods
	// BlockedByPod - node can't be removed because a pod running on it can't be moved. The reason why should be in BlockingPod.
	BlockedByPod
	// UnexpectedError - node can't be removed because of an unexpected error.
	UnexpectedError
	// NoNodeInfo - node can't be removed because it doesn't have any node info in the cluster snapshot.
	NoNodeInfo
)

// RemovalSimulator is a helper object for simulating node removal scenarios.
type RemovalSimulator struct {
	listers             kube_util.ListerRegistry
	clusterSnapshot     clustersnapshot.ClusterSnapshot
	canPersist          bool
	deleteOptions       options.NodeDeleteOptions
	drainabilityRules   rules.Rules
	schedulingSimulator *scheduling.HintingSimulator
}

// NewRemovalSimulator returns a new RemovalSimulator.
func NewRemovalSimulator(listers kube_util.ListerRegistry, clusterSnapshot clustersnapshot.ClusterSnapshot, deleteOptions options.NodeDeleteOptions, drainabilityRules rules.Rules, persistSuccessfulSimulations bool) *RemovalSimulator {
	return &RemovalSimulator{
		listers:             listers,
		clusterSnapshot:     clusterSnapshot,
		canPersist:          persistSuccessfulSimulations,
		deleteOptions:       deleteOptions,
		drainabilityRules:   drainabilityRules,
		schedulingSimulator: scheduling.NewHintingSimulator(),
	}
}

// SimulateNodeRemoval simulates removing a node from the cluster to check
// whether it is possible to move its pods. Depending on
// the outcome, exactly one of (NodeToBeRemoved, UnremovableNode) will be
// populated in the return value, the other will be nil.
func (r *RemovalSimulator) SimulateNodeRemoval(
	nodeName string,
	destinationMap map[string]bool,
	timestamp time.Time,
	remainingPdbTracker pdb.RemainingPdbTracker,
) (*NodeToBeRemoved, *UnremovableNode) {
	nodeInfo, err := r.clusterSnapshot.GetNodeInfo(nodeName)
	if err != nil {
		klog.Errorf("Can't retrieve node %s from snapshot, err: %v", nodeName, err)
		unremovableReason := UnexpectedError
		if errors.Is(err, clustersnapshot.ErrNodeNotFound) {
			unremovableReason = NoNodeInfo
		}
		unremovableNode := &UnremovableNode{Node: &apiv1.Node{ObjectMeta: metav1.ObjectMeta{Name: nodeName}}, Reason: unremovableReason}
		return nil, unremovableNode
	}
	klog.V(2).Infof("Simulating node %s removal", nodeName)

	podsToRemove, daemonSetPods, blockingPod, err := GetPodsToMove(nodeInfo, r.deleteOptions, r.drainabilityRules, r.listers, remainingPdbTracker, timestamp)
	if err != nil {
		klog.V(2).Infof("Node %s cannot be removed: %v", nodeName, err)
		if blockingPod != nil {
			return nil, &UnremovableNode{Node: nodeInfo.Node(), Reason: BlockedByPod, BlockingPod: blockingPod}
		}
		return nil, &UnremovableNode{Node: nodeInfo.Node(), Reason: UnexpectedError}
	}

	err = r.withForkedSnapshot(func() error {
		return r.findPlaceFor(nodeName, podsToRemove, destinationMap, timestamp)
	})
	if err != nil {
		klog.V(2).Infof("Node %s is not suitable for removal: %v", nodeName, err)
		return nil, &UnremovableNode{Node: nodeInfo.Node(), Reason: NoPlaceToMovePods}
	}
	klog.V(2).Infof("Node %s may be removed", nodeName)
	return &NodeToBeRemoved{
		Node:             nodeInfo.Node(),
		PodsToReschedule: podsToRemove,
		DaemonSetPods:    daemonSetPods,
	}, nil
}

func (r *RemovalSimulator) withForkedSnapshot(f func() error) (err error) {
	r.clusterSnapshot.Fork()
	defer func() {
		if err == nil && r.canPersist {
			cleanupErr := r.clusterSnapshot.Commit()
			if cleanupErr != nil {
				klog.Fatalf("Got error when calling ClusterSnapshot.Commit(); %v", cleanupErr)
			}
		} else {
			r.clusterSnapshot.Revert()
		}
	}()
	err = f()
	return err
}

func (r *RemovalSimulator) findPlaceFor(removedNode string, pods []*apiv1.Pod, nodes map[string]bool, timestamp time.Time) error {
	isCandidateNode := func(nodeInfo *framework.NodeInfo) bool {
		return nodeInfo.Node().Name != removedNode && nodes[nodeInfo.Node().Name]
	}

	pods = tpu.ClearTPURequests(pods)

	// Unschedule the pods from the Node in the snapshot first, so that they can be scheduled elsewhere by TrySchedulePods().
	for _, pod := range pods {
		if err := r.clusterSnapshot.UnschedulePod(pod.Namespace, pod.Name, removedNode); err != nil {
			// just log error
			klog.Errorf("Simulating removal of %s/%s return error; %v", pod.Namespace, pod.Name, err)
		}
	}

	if err := r.replaceWithTaintedGhostNode(removedNode, timestamp); err != nil {
		return err
	}

	newpods := make([]*apiv1.Pod, 0, len(pods))
	for _, podptr := range pods {
		newpod := *podptr
		newpod.Spec.NodeName = ""
		newpods = append(newpods, &newpod)
	}

	statuses, _, err := r.schedulingSimulator.TrySchedulePods(r.clusterSnapshot, newpods, isCandidateNode, true)
	if err != nil {
		return err
	}
	if len(statuses) != len(newpods) {
		return fmt.Errorf("can reschedule only %d out of %d pods", len(statuses), len(newpods))
	}

	// After successful scheduling simulation, remove the tainted ghost node so that
	// persisted snapshot state (used by subsequent simulations when canPersist=true)
	// correctly reflects the node being gone.
	return r.clusterSnapshot.RemoveNodeInfo(removedNode)
}

// replaceWithTaintedGhostNode replaces the given node in the snapshot with a
// pod-less copy carrying the ToBeDeletedByClusterAutoscaler NoSchedule taint.
// This mirrors what happens in reality during drain: the node is tainted but
// stays in the cluster. Keeping it in the snapshot is critical for
// PodTopologySpread constraints with the default nodeTaintsPolicy=Ignore,
// where the scheduler still counts tainted nodes as topology domains even
// though pods can't schedule on them. Removing the node entirely would
// eliminate its domain from topology calculations, making the simulation
// overly optimistic and causing scale-down/scale-up oscillation.
func (r *RemovalSimulator) replaceWithTaintedGhostNode(nodeName string, timestamp time.Time) error {
	nodeInfo, err := r.clusterSnapshot.GetNodeInfo(nodeName)
	if err != nil {
		return fmt.Errorf("couldn't get NodeInfo for removed node %s: %v", nodeName, err)
	}
	if err = r.clusterSnapshot.RemoveNodeInfo(nodeName); err != nil {
		return fmt.Errorf("couldn't remove NodeInfo for %s: %v", nodeName, err)
	}
	taintedNode := nodeInfo.Node().DeepCopy()
	taintedNode.Spec.Taints = append(taintedNode.Spec.Taints, apiv1.Taint{
		Key:    taints.ToBeDeletedTaint,
		Value:  fmt.Sprint(timestamp.Unix()),
		Effect: apiv1.TaintEffectNoSchedule,
	})
	ghostNodeInfo := framework.NewNodeInfo(taintedNode, nodeInfo.LocalResourceSlices)
	if nodeInfo.CSINode != nil {
		ghostNodeInfo.SetCSINode(nodeInfo.CSINode)
	}
	if err = r.clusterSnapshot.AddNodeInfo(ghostNodeInfo); err != nil {
		return fmt.Errorf("couldn't add tainted ghost node for %s: %v", nodeName, err)
	}
	return nil
}

// DropOldHints drops old scheduling hints.
func (r *RemovalSimulator) DropOldHints() {
	r.schedulingSimulator.DropOldHints()
}
