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

package utils

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/daemonset"
	"k8s.io/autoscaler/cluster-autoscaler/utils/deletetaint"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"k8s.io/klog"
)

const (
	// ReschedulerTaintKey is the name of the taint created by rescheduler.
	ReschedulerTaintKey = "CriticalAddonsOnly"
	// IgnoreTaintPrefix any taint starting with it will be filtered out from autoscaler nodes.
	IgnoreTaintPrefix = "ignore-taint.cluster-autoscaler.kubernetes.io/"
)

// GetNodeInfosForGroups finds NodeInfos for all node groups used to manage the given nodes. It also returns a node group to sample node mapping.
func GetNodeInfosForGroups(nodes []*apiv1.Node, nodeInfoCache map[string]*schedulernodeinfo.NodeInfo, cloudProvider cloudprovider.CloudProvider, listers kube_util.ListerRegistry,
	// TODO(mwielgus): This returns map keyed by url, while most code (including scheduler) uses node.Name for a key.
	// TODO(mwielgus): Review error policy - sometimes we may continue with partial errors.
	daemonsets []*appsv1.DaemonSet, predicateChecker simulator.PredicateChecker, ignoredTaints TaintKeySet) (map[string]*schedulernodeinfo.NodeInfo, errors.AutoscalerError) {
	result := make(map[string]*schedulernodeinfo.NodeInfo)
	seenGroups := make(map[string]bool)

	podsForNodes, err := getPodsForNodes(listers)
	if err != nil {
		return map[string]*schedulernodeinfo.NodeInfo{}, err
	}

	// processNode returns information whether the nodeTemplate was generated and if there was an error.
	processNode := func(node *apiv1.Node) (bool, string, errors.AutoscalerError) {
		nodeGroup, err := cloudProvider.NodeGroupForNode(node)
		if err != nil {
			return false, "", errors.ToAutoscalerError(errors.CloudProviderError, err)
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			return false, "", nil
		}
		id := nodeGroup.Id()
		if _, found := result[id]; !found {
			// Build nodeInfo.
			nodeInfo, err := simulator.BuildNodeInfoForNode(node, podsForNodes)
			if err != nil {
				return false, "", err
			}
			sanitizedNodeInfo, err := sanitizeNodeInfo(nodeInfo, id, ignoredTaints)
			if err != nil {
				return false, "", err
			}
			result[id] = sanitizedNodeInfo
			return true, id, nil
		}
		return false, "", nil
	}

	for _, node := range nodes {
		// Broken nodes might have some stuff missing. Skipping.
		if !kube_util.IsNodeReadyAndSchedulable(node) {
			continue
		}
		added, id, typedErr := processNode(node)
		if typedErr != nil {
			return map[string]*schedulernodeinfo.NodeInfo{}, typedErr
		}
		if added && nodeInfoCache != nil {
			if nodeInfoCopy, err := deepCopyNodeInfo(result[id]); err == nil {
				nodeInfoCache[id] = nodeInfoCopy
			}
		}
	}
	for _, nodeGroup := range cloudProvider.NodeGroups() {
		id := nodeGroup.Id()
		seenGroups[id] = true
		if _, found := result[id]; found {
			continue
		}

		// No good template, check cache of previously running nodes.
		if nodeInfoCache != nil {
			if nodeInfo, found := nodeInfoCache[id]; found {
				if nodeInfoCopy, err := deepCopyNodeInfo(nodeInfo); err == nil {
					result[id] = nodeInfoCopy
					continue
				}
			}
		}

		// No good template, trying to generate one. This is called only if there are no
		// working nodes in the node groups. By default CA tries to use a real-world example.
		nodeInfo, err := GetNodeInfoFromTemplate(nodeGroup, daemonsets, predicateChecker, ignoredTaints)
		if err != nil {
			if err == cloudprovider.ErrNotImplemented {
				continue
			} else {
				klog.Errorf("Unable to build proper template node for %s: %v", id, err)
				return map[string]*schedulernodeinfo.NodeInfo{}, errors.ToAutoscalerError(errors.CloudProviderError, err)
			}
		}
		result[id] = nodeInfo
	}

	// Remove invalid node groups from cache
	for id := range nodeInfoCache {
		if _, ok := seenGroups[id]; !ok {
			delete(nodeInfoCache, id)
		}
	}

	// Last resort - unready/unschedulable nodes.
	for _, node := range nodes {
		// Allowing broken nodes
		if !kube_util.IsNodeReadyAndSchedulable(node) {
			added, _, typedErr := processNode(node)
			if typedErr != nil {
				return map[string]*schedulernodeinfo.NodeInfo{}, typedErr
			}
			nodeGroup, err := cloudProvider.NodeGroupForNode(node)
			if err != nil {
				return map[string]*schedulernodeinfo.NodeInfo{}, errors.ToAutoscalerError(
					errors.CloudProviderError, err)
			}
			if added {
				klog.Warningf("Built template for %s based on unready/unschedulable node %s", nodeGroup.Id(), node.Name)
			}
		}
	}

	return result, nil
}

func getPodsForNodes(listers kube_util.ListerRegistry) (map[string][]*apiv1.Pod, errors.AutoscalerError) {
	pods, err := listers.ScheduledPodLister().List()
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.ApiCallError, err)
	}
	podsForNodes := map[string][]*apiv1.Pod{}
	for _, p := range pods {
		podsForNodes[p.Spec.NodeName] = append(podsForNodes[p.Spec.NodeName], p)
	}
	return podsForNodes, nil
}

// GetNodeInfoFromTemplate returns NodeInfo object built base on TemplateNodeInfo returned by NodeGroup.TemplateNodeInfo().
func GetNodeInfoFromTemplate(nodeGroup cloudprovider.NodeGroup, daemonsets []*appsv1.DaemonSet, predicateChecker simulator.PredicateChecker, ignoredTaints TaintKeySet) (*schedulernodeinfo.NodeInfo, errors.AutoscalerError) {
	id := nodeGroup.Id()
	baseNodeInfo, err := nodeGroup.TemplateNodeInfo()
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.CloudProviderError, err)
	}

	pods, err := daemonset.GetDaemonSetPodsForNode(baseNodeInfo, daemonsets, predicateChecker)
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}
	pods = append(pods, baseNodeInfo.Pods()...)
	fullNodeInfo := schedulernodeinfo.NewNodeInfo(pods...)
	fullNodeInfo.SetNode(baseNodeInfo.Node())
	sanitizedNodeInfo, typedErr := sanitizeNodeInfo(fullNodeInfo, id, ignoredTaints)
	if typedErr != nil {
		return nil, typedErr
	}
	return sanitizedNodeInfo, nil
}

// FilterOutNodesFromNotAutoscaledGroups return subset of input nodes for which cloud provider does not
// return autoscaled node group.
func FilterOutNodesFromNotAutoscaledGroups(nodes []*apiv1.Node, cloudProvider cloudprovider.CloudProvider) ([]*apiv1.Node, errors.AutoscalerError) {
	result := make([]*apiv1.Node, 0)

	for _, node := range nodes {
		nodeGroup, err := cloudProvider.NodeGroupForNode(node)
		if err != nil {
			return []*apiv1.Node{}, errors.ToAutoscalerError(errors.CloudProviderError, err)
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			result = append(result, node)
		}
	}
	return result, nil
}

func deepCopyNodeInfo(nodeInfo *schedulernodeinfo.NodeInfo) (*schedulernodeinfo.NodeInfo, errors.AutoscalerError) {
	newPods := make([]*apiv1.Pod, 0)
	for _, pod := range nodeInfo.Pods() {
		newPods = append(newPods, pod.DeepCopy())
	}

	// Build a new node info.
	newNodeInfo := schedulernodeinfo.NewNodeInfo(newPods...)
	if err := newNodeInfo.SetNode(nodeInfo.Node().DeepCopy()); err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}
	return newNodeInfo, nil
}

func sanitizeNodeInfo(nodeInfo *schedulernodeinfo.NodeInfo, nodeGroupName string, ignoredTaints TaintKeySet) (*schedulernodeinfo.NodeInfo, errors.AutoscalerError) {
	// Sanitize node name.
	sanitizedNode, err := sanitizeTemplateNode(nodeInfo.Node(), nodeGroupName, ignoredTaints)
	if err != nil {
		return nil, err
	}

	// Update nodename in pods.
	sanitizedPods := make([]*apiv1.Pod, 0)
	for _, pod := range nodeInfo.Pods() {
		sanitizedPod := pod.DeepCopy()
		sanitizedPod.Spec.NodeName = sanitizedNode.Name
		sanitizedPods = append(sanitizedPods, sanitizedPod)
	}

	// Build a new node info.
	sanitizedNodeInfo := schedulernodeinfo.NewNodeInfo(sanitizedPods...)
	if err := sanitizedNodeInfo.SetNode(sanitizedNode); err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}
	return sanitizedNodeInfo, nil
}

func sanitizeTemplateNode(node *apiv1.Node, nodeGroup string, ignoredTaints TaintKeySet) (*apiv1.Node, errors.AutoscalerError) {
	newNode := node.DeepCopy()
	nodeName := fmt.Sprintf("template-node-for-%s-%d", nodeGroup, rand.Int63())
	newNode.Labels = make(map[string]string, len(node.Labels))
	for k, v := range node.Labels {
		if k != apiv1.LabelHostname {
			newNode.Labels[k] = v
		} else {
			newNode.Labels[k] = nodeName
		}
	}
	newNode.Name = nodeName
	newTaints := make([]apiv1.Taint, 0)
	for _, taint := range node.Spec.Taints {
		// Rescheduler can put this taint on a node while evicting non-critical pods.
		// New nodes will not have this taint and so we should strip it when creating
		// template node.
		switch taint.Key {
		case ReschedulerTaintKey:
			klog.V(4).Infof("Removing rescheduler taint when creating template from node %s", node.Name)
			continue
		case deletetaint.ToBeDeletedTaint:
			klog.V(4).Infof("Removing autoscaler taint when creating template from node %s", node.Name)
			continue
		case deletetaint.DeletionCandidateTaint:
			klog.V(4).Infof("Removing autoscaler soft taint when creating template from node %s", node.Name)
			continue
		}

		// ignore conditional taints as they represent a transient node state.
		if exists := NodeConditionTaints[taint.Key]; exists {
			klog.V(4).Infof("Removing node condition taint %s, when creating template from node %s", taint.Key, node.Name)
			continue
		}

		if exists := ignoredTaints[taint.Key]; exists {
			klog.V(4).Infof("Removing ignored taint %s, when creating template from node %s", taint.Key, node.Name)
			continue
		}

		if strings.HasPrefix(taint.Key, IgnoreTaintPrefix) {
			klog.V(4).Infof("Removing taint %s based on prefix, when creation template from node %s", taint.Key, node.Name)
			continue
		}

		newTaints = append(newTaints, taint)
	}
	newNode.Spec.Taints = newTaints
	return newNode, nil
}

func hasHardInterPodAffinity(affinity *apiv1.Affinity) bool {
	if affinity == nil {
		return false
	}
	if affinity.PodAffinity != nil {
		if len(affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution) > 0 {
			return true
		}
	}
	if affinity.PodAntiAffinity != nil {
		if len(affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) > 0 {
			return true
		}
	}
	return false
}

// GetNodeCoresAndMemory extracts cpu and memory resources out of Node object
func GetNodeCoresAndMemory(node *apiv1.Node) (int64, int64) {
	cores := getNodeResource(node, apiv1.ResourceCPU)
	memory := getNodeResource(node, apiv1.ResourceMemory)
	return cores, memory
}

func getNodeResource(node *apiv1.Node, resource apiv1.ResourceName) int64 {
	nodeCapacity, found := node.Status.Capacity[resource]
	if !found {
		return 0
	}

	nodeCapacityValue := nodeCapacity.Value()
	if nodeCapacityValue < 0 {
		nodeCapacityValue = 0
	}

	return nodeCapacityValue
}

// UpdateClusterStateMetrics updates metrics related to cluster state
func UpdateClusterStateMetrics(csr *clusterstate.ClusterStateRegistry) {
	if csr == nil || reflect.ValueOf(csr).IsNil() {
		return
	}
	metrics.UpdateClusterSafeToAutoscale(csr.IsClusterHealthy())
	readiness := csr.GetClusterReadiness()
	metrics.UpdateNodesCount(readiness.Ready, readiness.Unready+readiness.LongNotStarted, readiness.NotStarted, readiness.LongUnregistered, readiness.Unregistered)
}

// GetOldestCreateTime returns oldest creation time out of the pods in the set
func GetOldestCreateTime(pods []*apiv1.Pod) time.Time {
	oldest := time.Now()
	for _, pod := range pods {
		if oldest.After(pod.CreationTimestamp.Time) {
			oldest = pod.CreationTimestamp.Time
		}
	}
	return oldest
}

// GetOldestCreateTimeWithGpu returns oldest creation time out of pods with GPU in the set
func GetOldestCreateTimeWithGpu(pods []*apiv1.Pod) (bool, time.Time) {
	oldest := time.Now()
	gpuFound := false
	for _, pod := range pods {
		if gpu.PodRequestsGpu(pod) {
			gpuFound = true
			if oldest.After(pod.CreationTimestamp.Time) {
				oldest = pod.CreationTimestamp.Time
			}
		}
	}
	return gpuFound, oldest
}
