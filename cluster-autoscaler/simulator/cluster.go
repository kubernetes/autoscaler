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
	"flag"
	"fmt"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/autoscaler/cluster-autoscaler/utils/klogx"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	pod_util "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
	"k8s.io/autoscaler/cluster-autoscaler/utils/tpu"

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"

	klog "k8s.io/klog/v2"
)

var (
	skipNodesWithSystemPods = flag.Bool("skip-nodes-with-system-pods", true,
		"If true cluster autoscaler will never delete nodes with pods from kube-system (except for DaemonSet "+
			"or mirror pods)")
	skipNodesWithLocalStorage = flag.Bool("skip-nodes-with-local-storage", true,
		"If true cluster autoscaler will never delete nodes with pods with local storage, e.g. EmptyDir or HostPath")

	minReplicaCount = flag.Int("min-replica-count", 0,
		"Minimum number or replicas that a replica set or replication controller should have to allow their pods deletion in scale down")
)

// NodeToBeRemoved contain information about a node that can be removed.
type NodeToBeRemoved struct {
	// Node to be removed.
	Node *apiv1.Node
	// PodsToReschedule contains pods on the node that should be rescheduled elsewhere.
	PodsToReschedule []*apiv1.Pod
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

// UtilizationInfo contains utilization information for a node.
type UtilizationInfo struct {
	CpuUtil float64
	MemUtil float64
	GpuUtil float64
	// Resource name of highest utilization resource
	ResourceName apiv1.ResourceName
	// Max(CpuUtil, MemUtil) or GpuUtils
	Utilization float64
}

// FindNodesToRemove finds nodes that can be removed. Returns also an information about good
// rescheduling location for each of the pods.
func FindNodesToRemove(
	candidates []string,
	destinations []string,
	listers kube_util.ListerRegistry,
	clusterSnapshot ClusterSnapshot,
	predicateChecker PredicateChecker,
	maxCount int,
	fastCheck bool,
	oldHints map[string]string,
	usageTracker *UsageTracker,
	timestamp time.Time,
	podDisruptionBudgets []*policyv1.PodDisruptionBudget,
) (nodesToRemove []NodeToBeRemoved, unremovableNodes []*UnremovableNode, podReschedulingHints map[string]string, finalError errors.AutoscalerError) {

	result := make([]NodeToBeRemoved, 0)
	unremovable := make([]*UnremovableNode, 0)

	evaluationType := "Detailed evaluation"
	if fastCheck {
		evaluationType = "Fast evaluation"
	}
	newHints := make(map[string]string, len(oldHints))

	destinationMap := make(map[string]bool, len(destinations))
	for _, destination := range destinations {
		destinationMap[destination] = true
	}

candidateloop:
	for _, nodeName := range candidates {
		nodeInfo, err := clusterSnapshot.NodeInfos().Get(nodeName)
		if err != nil {
			klog.Errorf("Can't retrieve node %s from snapshot, err: %v", nodeName, err)
		}
		klog.V(2).Infof("%s: %s for removal", evaluationType, nodeName)

		var podsToRemove []*apiv1.Pod
		var blockingPod *drain.BlockingPod

		if _, found := destinationMap[nodeName]; !found {
			klog.V(2).Infof("%s: nodeInfo for %s not found", evaluationType, nodeName)
			unremovable = append(unremovable, &UnremovableNode{Node: nodeInfo.Node(), Reason: UnexpectedError})
			continue candidateloop
		}

		if fastCheck {
			podsToRemove, blockingPod, err = FastGetPodsToMove(nodeInfo, *skipNodesWithSystemPods, *skipNodesWithLocalStorage,
				podDisruptionBudgets)
		} else {
			podsToRemove, blockingPod, err = DetailedGetPodsForMove(nodeInfo, *skipNodesWithSystemPods, *skipNodesWithLocalStorage, listers, int32(*minReplicaCount),
				podDisruptionBudgets)
		}

		if err != nil {
			klog.V(2).Infof("%s: node %s cannot be removed: %v", evaluationType, nodeName, err)
			if blockingPod != nil {
				unremovable = append(unremovable, &UnremovableNode{Node: nodeInfo.Node(), Reason: BlockedByPod, BlockingPod: blockingPod})
			} else {
				unremovable = append(unremovable, &UnremovableNode{Node: nodeInfo.Node(), Reason: UnexpectedError})
			}
			continue candidateloop
		}

		findProblems := findPlaceFor(nodeName, podsToRemove, destinationMap, clusterSnapshot,
			predicateChecker, oldHints, newHints, usageTracker, timestamp)

		if findProblems == nil {
			result = append(result, NodeToBeRemoved{
				Node:             nodeInfo.Node(),
				PodsToReschedule: podsToRemove,
			})
			klog.V(2).Infof("%s: node %s may be removed", evaluationType, nodeName)
			if len(result) >= maxCount {
				break candidateloop
			}
		} else {
			klog.V(2).Infof("%s: node %s is not suitable for removal: %v", evaluationType, nodeName, findProblems)
			unremovable = append(unremovable, &UnremovableNode{Node: nodeInfo.Node(), Reason: NoPlaceToMovePods})
		}
	}
	return result, unremovable, newHints, nil
}

// FindEmptyNodesToRemove finds empty nodes that can be removed.
func FindEmptyNodesToRemove(snapshot ClusterSnapshot, candidates []string) []string {
	result := make([]string, 0)
	for _, node := range candidates {
		nodeInfo, err := snapshot.NodeInfos().Get(node)
		if err != nil {
			klog.Errorf("Can't retrieve node %s from snapshot, err: %v", node, err)
			continue
		}
		// Should block on all pods.
		podsToRemove, _, err := FastGetPodsToMove(nodeInfo, true, true, nil)
		if err == nil && len(podsToRemove) == 0 {
			result = append(result, node)
		}
	}
	return result
}

// CalculateUtilization calculates utilization of a node, defined as maximum of (cpu, memory) or gpu utilization
// based on if the node has GPU or not. Per resource utilization is the sum of requests for it divided by allocatable.
// It also returns the individual cpu, memory and gpu utilization.
func CalculateUtilization(node *apiv1.Node, nodeInfo *schedulerframework.NodeInfo, skipDaemonSetPods, skipMirrorPods bool, gpuLabel string) (utilInfo UtilizationInfo, err error) {
	if gpu.NodeHasGpu(gpuLabel, node) {
		gpuUtil, err := calculateUtilizationOfResource(node, nodeInfo, gpu.ResourceNvidiaGPU, skipDaemonSetPods, skipMirrorPods)
		if err != nil {
			klog.V(3).Infof("node %s has unready GPU", node.Name)
			// Return 0 if GPU is unready. This will guarantee we can still scale down a node with unready GPU.
			return UtilizationInfo{GpuUtil: 0, ResourceName: gpu.ResourceNvidiaGPU, Utilization: 0}, nil
		}

		// Skips cpu and memory utilization calculation for node with GPU.
		return UtilizationInfo{GpuUtil: gpuUtil, ResourceName: gpu.ResourceNvidiaGPU, Utilization: gpuUtil}, nil
	}

	cpu, err := calculateUtilizationOfResource(node, nodeInfo, apiv1.ResourceCPU, skipDaemonSetPods, skipMirrorPods)
	if err != nil {
		return UtilizationInfo{}, err
	}
	mem, err := calculateUtilizationOfResource(node, nodeInfo, apiv1.ResourceMemory, skipDaemonSetPods, skipMirrorPods)
	if err != nil {
		return UtilizationInfo{}, err
	}

	utilization := UtilizationInfo{CpuUtil: cpu, MemUtil: mem}

	if cpu > mem {
		utilization.ResourceName = apiv1.ResourceCPU
		utilization.Utilization = cpu
	} else {
		utilization.ResourceName = apiv1.ResourceMemory
		utilization.Utilization = mem
	}

	return utilization, nil
}

func calculateUtilizationOfResource(node *apiv1.Node, nodeInfo *schedulerframework.NodeInfo, resourceName apiv1.ResourceName, skipDaemonSetPods, skipMirrorPods bool) (float64, error) {
	nodeAllocatable, found := node.Status.Allocatable[resourceName]
	if !found {
		return 0, fmt.Errorf("failed to get %v from %s", resourceName, node.Name)
	}
	if nodeAllocatable.MilliValue() == 0 {
		return 0, fmt.Errorf("%v is 0 at %s", resourceName, node.Name)
	}
	podsRequest := resource.MustParse("0")
	for _, podInfo := range nodeInfo.Pods {
		// factor daemonset pods out of the utilization calculations
		if skipDaemonSetPods && pod_util.IsDaemonSetPod(podInfo.Pod) {
			continue
		}
		// factor mirror pods out of the utilization calculations
		if skipMirrorPods && pod_util.IsMirrorPod(podInfo.Pod) {
			continue
		}
		for _, container := range podInfo.Pod.Spec.Containers {
			if resourceValue, found := container.Resources.Requests[resourceName]; found {
				podsRequest.Add(resourceValue)
			}
		}
	}
	return float64(podsRequest.MilliValue()) / float64(nodeAllocatable.MilliValue()), nil
}

func findPlaceFor(removedNode string, pods []*apiv1.Pod, nodes map[string]bool,
	clusterSnapshot ClusterSnapshot, predicateChecker PredicateChecker, oldHints map[string]string, newHints map[string]string, usageTracker *UsageTracker,
	timestamp time.Time) error {

	if err := clusterSnapshot.Fork(); err != nil {
		return err
	}
	defer func() {
		err := clusterSnapshot.Revert()
		if err != nil {
			klog.Fatalf("Got error when calling ClusterSnapshot.Revert(); %v", err)
		}
	}()

	podKey := func(pod *apiv1.Pod) string {
		return fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
	}

	loggingQuota := klogx.PodsLoggingQuota()

	tryNodeForPod := func(nodename string, pod *apiv1.Pod) bool {
		if err := predicateChecker.CheckPredicates(clusterSnapshot, pod, nodename); err != nil {
			klogx.V(4).UpTo(loggingQuota).Infof("Evaluation %s for %s/%s -> %v", nodename, pod.Namespace, pod.Name, err.VerboseMessage())
			return false
		}

		klog.V(4).Infof("Pod %s/%s can be moved to %s", pod.Namespace, pod.Name, nodename)
		if err := clusterSnapshot.AddPod(pod, nodename); err != nil {
			klog.Errorf("Simulating scheduling of %s/%s to %s return error; %v", pod.Namespace, pod.Name, nodename, err)
			return false
		}
		newHints[podKey(pod)] = nodename
		return true
	}

	pods = tpu.ClearTPURequests(pods)

	// remove pods from clusterSnapshot first
	for _, pod := range pods {
		if err := clusterSnapshot.RemovePod(pod.Namespace, pod.Name, removedNode); err != nil {
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

		loggingQuota.Reset()

		klog.V(5).Infof("Looking for place for %s/%s", pod.Namespace, pod.Name)

		if hintedNode, hasHint := oldHints[podKey(pod)]; hasHint {
			if hintedNode != removedNode && tryNodeForPod(hintedNode, pod) {
				foundPlace = true
				targetNode = hintedNode
			}
		}

		if !foundPlace {
			for nodeName := range nodes {
				if nodeName == removedNode {
					continue
				}
				if tryNodeForPod(nodeName, pod) {
					foundPlace = true
					targetNode = nodeName
					break
				}
			}
			if !foundPlace {
				klogx.V(4).Over(loggingQuota).Infof("%v other nodes evaluated for %s/%s", -loggingQuota.Left(), pod.Namespace, pod.Name)
				return fmt.Errorf("failed to find place for %s", podKey(pod))
			}
		}

		usageTracker.RegisterUsage(removedNode, targetNode, timestamp)
	}
	return nil
}
