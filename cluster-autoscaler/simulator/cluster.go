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

	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/tpu"

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	klog "k8s.io/klog/v2"
)

// NodeToBeRemoved contain information about a node that can be removed.
type NodeToBeRemoved struct {
	// Node to be removed.
	Node *apiv1.Node
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
	listers          kube_util.ListerRegistry
	clusterSnapshot  ClusterSnapshot
	predicateChecker PredicateChecker
	usageTracker     *UsageTracker
	canPersist       bool
	deleteOptions    NodeDeleteOptions
}

// NewRemovalSimulator returns a new RemovalSimulator.
func NewRemovalSimulator(listers kube_util.ListerRegistry, clusterSnapshot ClusterSnapshot, predicateChecker PredicateChecker,
	usageTracker *UsageTracker, deleteOptions NodeDeleteOptions, persistSuccessfulSimulations bool) *RemovalSimulator {
	return &RemovalSimulator{
		listers:          listers,
		clusterSnapshot:  clusterSnapshot,
		predicateChecker: predicateChecker,
		usageTracker:     usageTracker,
		canPersist:       persistSuccessfulSimulations,
		deleteOptions:    deleteOptions,
	}
}

// FindNodesToRemove finds nodes that can be removed. Returns also an
// information about good rescheduling location for each of the pods.
func (r *RemovalSimulator) FindNodesToRemove(
	candidates []string,
	destinations []string,
	oldHints map[string]string,
	timestamp time.Time,
	pdbs []*policyv1.PodDisruptionBudget,
) (nodesToRemove []NodeToBeRemoved, unremovableNodes []*UnremovableNode, podReschedulingHints map[string]string, finalError errors.AutoscalerError) {
	result := make([]NodeToBeRemoved, 0)
	unremovable := make([]*UnremovableNode, 0)
	newHints := make(map[string]string, len(oldHints))

	destinationMap := make(map[string]bool, len(destinations))
	for _, destination := range destinations {
		destinationMap[destination] = true
	}

	for _, nodeName := range candidates {
		rn, urn := r.CheckNodeRemoval(nodeName, destinationMap, oldHints, newHints, timestamp, pdbs)
		if rn != nil {
			result = append(result, *rn)
		} else if urn != nil {
			unremovable = append(unremovable, urn)
		}
	}
	return result, unremovable, newHints, nil
}

// CheckNodeRemoval checks whether a specific node can be removed. Depending on
// the outcome, exactly one of (NodeToBeRemoved, UnremovableNode) will be
// populated in the return value, the other will be nil.
func (r *RemovalSimulator) CheckNodeRemoval(
	nodeName string,
	destinationMap map[string]bool,
	oldHints map[string]string,
	newHints map[string]string,
	timestamp time.Time,
	pdbs []*policyv1.PodDisruptionBudget,
) (*NodeToBeRemoved, *UnremovableNode) {
	nodeInfo, err := r.clusterSnapshot.NodeInfos().Get(nodeName)
	if err != nil {
		klog.Errorf("Can't retrieve node %s from snapshot, err: %v", nodeName, err)
	}
	klog.V(2).Infof("%s for removal", nodeName)

	if _, found := destinationMap[nodeName]; !found {
		klog.V(2).Infof("nodeInfo for %s not found", nodeName)
		return nil, &UnremovableNode{Node: nodeInfo.Node(), Reason: UnexpectedError}
	}

	podsToRemove, daemonSetPods, blockingPod, err := GetPodsToMove(nodeInfo, r.deleteOptions, r.listers, pdbs, timestamp)
	if err != nil {
		klog.V(2).Infof("node %s cannot be removed: %v", nodeName, err)
		if blockingPod != nil {
			return nil, &UnremovableNode{Node: nodeInfo.Node(), Reason: BlockedByPod, BlockingPod: blockingPod}
		}
		return nil, &UnremovableNode{Node: nodeInfo.Node(), Reason: UnexpectedError}
	}

	err = r.withForkedSnapshot(func() error {
		return r.findPlaceFor(nodeName, podsToRemove, destinationMap, oldHints, newHints, timestamp)
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
		nodeInfo, err := r.clusterSnapshot.NodeInfos().Get(node)
		if err != nil {
			klog.Errorf("Can't retrieve node %s from snapshot, err: %v", node, err)
			continue
		}
		// Should block on all pods
		podsToRemove, _, _, err := GetPodsToMove(nodeInfo, r.deleteOptions, nil, nil, timestamp)
		if err == nil && len(podsToRemove) == 0 {
			result = append(result, node)
		}
	}
	return result
}

func (r *RemovalSimulator) withForkedSnapshot(f func() error) (err error) {
	if err = r.clusterSnapshot.Fork(); err != nil {
		return err
	}
	defer func() {
		if err == nil && r.canPersist {
			cleanupErr := r.clusterSnapshot.Commit()
			if cleanupErr != nil {
				klog.Fatalf("Got error when calling ClusterSnapshot.Commit(); %v", cleanupErr)
			}
		} else {
			cleanupErr := r.clusterSnapshot.Revert()
			if cleanupErr != nil {
				klog.Fatalf("Got error when calling ClusterSnapshot.Revert(); %v", cleanupErr)
			}
		}
	}()
	err = f()
	return err
}

func (r *RemovalSimulator) findPlaceFor(removedNode string, pods []*apiv1.Pod, nodes map[string]bool,
	oldHints map[string]string, newHints map[string]string, timestamp time.Time) error {
	podKey := func(pod *apiv1.Pod) string {
		return fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
	}

	isCandidateNode := func(nodeName string) bool {
		return nodeName != removedNode && nodes[nodeName]
	}

	pods = tpu.ClearTPURequests(pods)

	// remove pods from clusterSnapshot first
	for _, pod := range pods {
		if err := r.clusterSnapshot.RemovePod(pod.Namespace, pod.Name, removedNode); err != nil {
			// just log error
			klog.Errorf("Simulating removal of %s/%s return error; %v", pod.Namespace, pod.Name, err)
		}
	}

	for _, podptr := range pods {
		newpod := *podptr
		newpod.Spec.NodeName = ""
		pod := &newpod

		foundPlace := false
		targetNode := ""

		klog.V(5).Infof("Looking for place for %s/%s", pod.Namespace, pod.Name)

		if hintedNode, hasHint := oldHints[podKey(pod)]; hasHint && isCandidateNode(hintedNode) {
			if err := r.predicateChecker.CheckPredicates(r.clusterSnapshot, pod, hintedNode); err == nil {
				klog.V(4).Infof("Pod %s/%s can be moved to %s", pod.Namespace, pod.Name, hintedNode)
				if err := r.clusterSnapshot.AddPod(pod, hintedNode); err != nil {
					return fmt.Errorf("Simulating scheduling of %s/%s to %s return error; %v", pod.Namespace, pod.Name, hintedNode, err)
				}
				newHints[podKey(pod)] = hintedNode
				foundPlace = true
				targetNode = hintedNode
			}
		}

		if !foundPlace {
			newNodeName, err := r.predicateChecker.FitsAnyNodeMatching(r.clusterSnapshot, pod, func(nodeInfo *schedulerframework.NodeInfo) bool {
				return isCandidateNode(nodeInfo.Node().Name)
			})
			if err == nil {
				klog.V(4).Infof("Pod %s/%s can be moved to %s", pod.Namespace, pod.Name, newNodeName)
				if err := r.clusterSnapshot.AddPod(pod, newNodeName); err != nil {
					return fmt.Errorf("Simulating scheduling of %s/%s to %s return error; %v", pod.Namespace, pod.Name, newNodeName, err)
				}
				newHints[podKey(pod)] = newNodeName
				targetNode = newNodeName
			} else {
				return fmt.Errorf("failed to find place for %s", podKey(pod))
			}
		}

		r.usageTracker.RegisterUsage(removedNode, targetNode, timestamp)
	}
	return nil
}
