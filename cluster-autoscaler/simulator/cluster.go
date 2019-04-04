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
	"math/rand"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/glogx"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	scheduler_util "k8s.io/autoscaler/cluster-autoscaler/utils/scheduler"
	"k8s.io/autoscaler/cluster-autoscaler/utils/tpu"
	"k8s.io/kubernetes/pkg/scheduler/algorithm/predicates"

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/resource"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"k8s.io/klog"
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
func FindNodesToRemove(candidates []*apiv1.Node, allNodes []*apiv1.Node, pods []*apiv1.Pod,
	listers kube_util.ListerRegistry, predicateChecker *PredicateChecker, maxCount int,
	fastCheck bool, oldHints map[string]string, usageTracker *UsageTracker,
	timestamp time.Time,
	podDisruptionBudgets []*policyv1.PodDisruptionBudget,
) (nodesToRemove []NodeToBeRemoved, unremovableNodes []*apiv1.Node, podReschedulingHints map[string]string, finalError errors.AutoscalerError) {

	nodeNameToNodeInfo := scheduler_util.CreateNodeNameToInfoMap(pods, allNodes)
	result := make([]NodeToBeRemoved, 0)
	unremovable := make([]*apiv1.Node, 0)

	evaluationType := "Detailed evaluation"
	if fastCheck {
		evaluationType = "Fast evaluation"
	}
	newHints := make(map[string]string, len(oldHints))

candidateloop:
	for _, node := range candidates {
		klog.V(2).Infof("%s: %s for removal", evaluationType, node.Name)

		var podsToRemove []*apiv1.Pod
		var err error

		if nodeInfo, found := nodeNameToNodeInfo[node.Name]; found {
			if fastCheck {
				podsToRemove, err = FastGetPodsToMove(nodeInfo, *skipNodesWithSystemPods, *skipNodesWithLocalStorage,
					podDisruptionBudgets)
			} else {
				podsToRemove, err = DetailedGetPodsForMove(nodeInfo, *skipNodesWithSystemPods, *skipNodesWithLocalStorage, listers, int32(*minReplicaCount),
					podDisruptionBudgets)
			}
			if err != nil {
				klog.V(2).Infof("%s: node %s cannot be removed: %v", evaluationType, node.Name, err)
				unremovable = append(unremovable, node)
				continue candidateloop
			}
		} else {
			klog.V(2).Infof("%s: nodeInfo for %s not found", evaluationType, node.Name)
			unremovable = append(unremovable, node)
			continue candidateloop
		}
		findProblems := findPlaceFor(node.Name, podsToRemove, allNodes, nodeNameToNodeInfo, predicateChecker, oldHints, newHints,
			usageTracker, timestamp)

		if findProblems == nil {
			result = append(result, NodeToBeRemoved{
				Node:             node,
				PodsToReschedule: podsToRemove,
			})
			klog.V(2).Infof("%s: node %s may be removed", evaluationType, node.Name)
			if len(result) >= maxCount {
				break candidateloop
			}
		} else {
			klog.V(2).Infof("%s: node %s is not suitable for removal: %v", evaluationType, node.Name, findProblems)
			unremovable = append(unremovable, node)
		}
	}
	return result, unremovable, newHints, nil
}

// FindEmptyNodesToRemove finds empty nodes that can be removed.
func FindEmptyNodesToRemove(candidates []*apiv1.Node, pods []*apiv1.Pod) []*apiv1.Node {
	nodeNameToNodeInfo := scheduler_util.CreateNodeNameToInfoMap(pods, candidates)
	result := make([]*apiv1.Node, 0)
	for _, node := range candidates {
		if nodeInfo, found := nodeNameToNodeInfo[node.Name]; found {
			// Should block on all pods.
			podsToRemove, err := FastGetPodsToMove(nodeInfo, true, true, nil)
			if err == nil && len(podsToRemove) == 0 {
				result = append(result, node)
			}
		} else {
			// Node without pods.
			result = append(result, node)
		}
	}
	return result
}

// CalculateUtilization calculates utilization of a node, defined as maximum of (cpu, memory) or gpu utilization
// based on if the node has GPU or not. Per resource utilization is the sum of requests for it divided by allocatable.
// It also returns the individual cpu, memory and gpu utilization.
func CalculateUtilization(node *apiv1.Node, nodeInfo *schedulernodeinfo.NodeInfo, skipDaemonSetPods, skipMirrorPods bool, gpuLabel string) (utilInfo UtilizationInfo, err error) {
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

func calculateUtilizationOfResource(node *apiv1.Node, nodeInfo *schedulernodeinfo.NodeInfo, resourceName apiv1.ResourceName, skipDaemonSetPods, skipMirrorPods bool) (float64, error) {
	nodeAllocatable, found := node.Status.Allocatable[resourceName]
	if !found {
		return 0, fmt.Errorf("failed to get %v from %s", resourceName, node.Name)
	}
	if nodeAllocatable.MilliValue() == 0 {
		return 0, fmt.Errorf("%v is 0 at %s", resourceName, node.Name)
	}
	podsRequest := resource.MustParse("0")
	for _, pod := range nodeInfo.Pods() {
		// factor daemonset pods out of the utilization calculations
		if skipDaemonSetPods && isDaemonSet(pod) {
			continue
		}
		// factor mirror pods out of the utilization calculations
		if skipMirrorPods && drain.IsMirrorPod(pod) {
			continue
		}
		for _, container := range pod.Spec.Containers {
			if resourceValue, found := container.Resources.Requests[resourceName]; found {
				podsRequest.Add(resourceValue)
			}
		}
	}
	return float64(podsRequest.MilliValue()) / float64(nodeAllocatable.MilliValue()), nil
}

// TODO: We don't need to pass list of nodes here as they are already available in nodeInfos.
func findPlaceFor(removedNode string, pods []*apiv1.Pod, nodes []*apiv1.Node, nodeInfos map[string]*schedulernodeinfo.NodeInfo,
	predicateChecker *PredicateChecker, oldHints map[string]string, newHints map[string]string, usageTracker *UsageTracker,
	timestamp time.Time) error {

	newNodeInfos := make(map[string]*schedulernodeinfo.NodeInfo)
	for k, v := range nodeInfos {
		newNodeInfos[k] = v
	}

	podKey := func(pod *apiv1.Pod) string {
		return fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
	}

	loggingQuota := glogx.PodsLoggingQuota()

	tryNodeForPod := func(nodename string, pod *apiv1.Pod, predicateMeta predicates.PredicateMetadata) bool {
		nodeInfo, found := newNodeInfos[nodename]
		if found {
			if nodeInfo.Node() == nil {
				// NodeInfo is generated based on pods. It is possible that node is removed from
				// an api server faster than the pod that were running on them. In such a case
				// we have to skip this nodeInfo. It should go away pretty soon.
				klog.Warningf("No node in nodeInfo %s -> %v", nodename, nodeInfo)
				return false
			}
			err := predicateChecker.CheckPredicates(pod, predicateMeta, nodeInfo)
			if err != nil {
				glogx.V(4).UpTo(loggingQuota).Infof("Evaluation %s for %s/%s -> %v", nodename, pod.Namespace, pod.Name, err.VerboseError())
			} else {
				// TODO(mwielgus): Optimize it.
				klog.V(4).Infof("Pod %s/%s can be moved to %s", pod.Namespace, pod.Name, nodename)
				podsOnNode := nodeInfo.Pods()
				podsOnNode = append(podsOnNode, pod)
				newNodeInfo := schedulernodeinfo.NewNodeInfo(podsOnNode...)
				newNodeInfo.SetNode(nodeInfo.Node())
				newNodeInfos[nodename] = newNodeInfo
				newHints[podKey(pod)] = nodename
				return true
			}
		}
		return false
	}

	// TODO: come up with a better semi-random semi-utilization sorted
	// layout.
	shuffledNodes := shuffleNodes(nodes)

	pods = tpu.ClearTPURequests(pods)
	for _, podptr := range pods {
		newpod := *podptr
		newpod.Spec.NodeName = ""
		pod := &newpod

		foundPlace := false
		targetNode := ""
		predicateMeta := predicateChecker.GetPredicateMetadata(pod, newNodeInfos)
		loggingQuota.Reset()

		klog.V(5).Infof("Looking for place for %s/%s", pod.Namespace, pod.Name)

		hintedNode, hasHint := oldHints[podKey(pod)]
		if hasHint {
			if hintedNode != removedNode && tryNodeForPod(hintedNode, pod, predicateMeta) {
				foundPlace = true
				targetNode = hintedNode
			}
		}
		if !foundPlace {
			for _, node := range shuffledNodes {
				if node.Name == removedNode {
					continue
				}
				if tryNodeForPod(node.Name, pod, predicateMeta) {
					foundPlace = true
					targetNode = node.Name
					break
				}
			}
			if !foundPlace {
				glogx.V(4).Over(loggingQuota).Infof("%v other nodes evaluated for %s/%s", -loggingQuota.Left(), pod.Namespace, pod.Name)
				return fmt.Errorf("failed to find place for %s", podKey(pod))
			}
		}

		usageTracker.RegisterUsage(removedNode, targetNode, timestamp)
	}
	return nil
}

func shuffleNodes(nodes []*apiv1.Node) []*apiv1.Node {
	result := make([]*apiv1.Node, len(nodes))
	for i := range nodes {
		result[i] = nodes[i]
	}
	for i := range result {
		j := rand.Intn(len(result))
		result[i], result[j] = result[j], result[i]
	}
	return result
}

func isDaemonSet(pod *apiv1.Pod) bool {
	for _, ownerReference := range pod.ObjectMeta.OwnerReferences {
		if ownerReference.Kind == "DaemonSet" {
			return true
		}
	}
	return false
}
