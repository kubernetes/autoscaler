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
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/pdb"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability/rules"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/scheduling"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/tpu"

	klog "k8s.io/klog/v2"
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
func NewRemovalSimulator(listers kube_util.ListerRegistry, clusterSnapshot clustersnapshot.ClusterSnapshot, predicateChecker predicatechecker.PredicateChecker,
	deleteOptions options.NodeDeleteOptions, drainabilityRules rules.Rules, persistSuccessfulSimulations bool) *RemovalSimulator {
	return &RemovalSimulator{
		listers:             listers,
		clusterSnapshot:     clusterSnapshot,
		canPersist:          persistSuccessfulSimulations,
		deleteOptions:       deleteOptions,
		drainabilityRules:   drainabilityRules,
		schedulingSimulator: scheduling.NewHintingSimulator(predicateChecker),
	}
}

// FindNodesToRemove finds nodes that can be removed.
func (r *RemovalSimulator) FindNodesToRemove(
	candidates []string,
	destinations []string,
	timestamp time.Time,
	remainingPdbTracker pdb.RemainingPdbTracker,
) (nodesToRemove []NodeToBeRemoved, unremovableNodes []*UnremovableNode) {
	destinationMap := make(map[string]bool, len(destinations))
	for _, destination := range destinations {
		destinationMap[destination] = true
	}

	for _, nodeName := range candidates {
		rn, urn := r.SimulateNodeRemoval(nodeName, destinationMap, timestamp, remainingPdbTracker)
		if rn != nil {
			nodesToRemove = append(nodesToRemove, *rn)
		} else if urn != nil {
			unremovableNodes = append(unremovableNodes, urn)
		}
	}
	return nodesToRemove, unremovableNodes
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
	}
	klog.V(2).Infof("Simulating node %s removal", nodeName)

	podsToRemove, daemonSetPods, blockingPod, err := GetPodsToMove(nodeInfo, r.deleteOptions, r.drainabilityRules, r.listers, remainingPdbTracker, timestamp)
	if err != nil {
		klog.V(2).Infof("node %s cannot be removed: %v", nodeName, err)
		if blockingPod != nil {
			return nil, &UnremovableNode{Node: nodeInfo.Node(), Reason: BlockedByPod, BlockingPod: blockingPod}
		}
		return nil, &UnremovableNode{Node: nodeInfo.Node(), Reason: UnexpectedError}
	}

	err = r.withForkedSnapshot(func() error {
		return r.findPlaceFor(nodeName, podsToRemove, destinationMap, timestamp)
	})
	if err != nil {
		klog.V(2).Infof("node %s is not suitable for removal: %v", nodeName, err)
		return nil, &UnremovableNode{Node: nodeInfo.Node(), Reason: NoPlaceToMovePods}
	}
	klog.V(2).Infof("node %s may be removed", nodeName)
	return &NodeToBeRemoved{
		Node:             nodeInfo.Node(),
		PodsToReschedule: podsToRemove,
		DaemonSetPods:    daemonSetPods,
	}, nil
}

// FindEmptyNodesToRemove finds empty nodes that can be removed.
func (r *RemovalSimulator) FindEmptyNodesToRemove(candidates []string, timestamp time.Time) []string {
	result := make([]string, 0)
	for _, node := range candidates {
		nodeInfo, err := r.clusterSnapshot.GetNodeInfo(node)
		if err != nil {
			klog.Errorf("Can't retrieve node %s from snapshot, err: %v", node, err)
			continue
		}
		// Should block on all pods
		podsToRemove, _, _, err := GetPodsToMove(nodeInfo, r.deleteOptions, r.drainabilityRules, nil, nil, timestamp)
		if err == nil && len(podsToRemove) == 0 {
			result = append(result, node)
		}
	}
	return result
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

	// remove pods from clusterSnapshot first
	for _, pod := range pods {
		if err := r.clusterSnapshot.UnschedulePod(pod.Namespace, pod.Name, removedNode); err != nil {
			// just log error
			klog.Errorf("Simulating removal of %s/%s return error; %v", pod.Namespace, pod.Name, err)
		}
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
	return nil
}

// DropOldHints drops old scheduling hints.
func (r *RemovalSimulator) DropOldHints() {
	r.schedulingSimulator.DropOldHints()
}
