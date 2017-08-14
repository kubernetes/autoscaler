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

package core

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/daemonset"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/kubernetes/pkg/api"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	podv1 "k8s.io/kubernetes/pkg/api/v1/pod"
	extensionsv1 "k8s.io/kubernetes/pkg/apis/extensions/v1beta1"
	kube_client "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
)

const (
	// ReschedulerTaintKey is the name of the taint created by rescheduler.
	ReschedulerTaintKey = "CriticalAddonsOnly"
)

// GetAllNodesAvailableTime returns time when the newest node became available for scheduler.
// TODO: This function should use LastTransitionTime from NodeReady condition.
func GetAllNodesAvailableTime(nodes []*apiv1.Node) time.Time {
	var result time.Time
	for _, node := range nodes {
		if node.CreationTimestamp.After(result) {
			result = node.CreationTimestamp.Time
		}
	}
	return result.Add(1 * time.Minute)
}

// SlicePodsByPodScheduledTime slices given pod array into those where PodScheduled condition
// have been updated after the thresold and others.
// Each pod must be in condition "Scheduled: False; Reason: Unschedulable"
func SlicePodsByPodScheduledTime(pods []*apiv1.Pod, threshold time.Time) (oldPods []*apiv1.Pod, newPods []*apiv1.Pod) {
	for _, pod := range pods {
		_, condition := podv1.GetPodCondition(&pod.Status, apiv1.PodScheduled)
		if condition != nil {
			if condition.LastTransitionTime.After(threshold) {
				newPods = append(newPods, pod)
			} else {
				oldPods = append(oldPods, pod)
			}
		}
	}
	return
}

// ResetPodScheduledCondition resets pod condition PodScheduled to "unknown" for all the pods with LastTransitionTime
// not after the threshold time.
func ResetPodScheduledCondition(kubeClient kube_client.Interface, pods []*apiv1.Pod) {
	for _, pod := range pods {
		if err := resetPodScheduledConditionForPod(kubeClient, pod); err != nil {
			glog.Errorf("Error during reseting pod condition for %s/%s: %v", pod.Namespace, pod.Name, err)
		}
	}
}

func resetPodScheduledConditionForPod(kubeClient kube_client.Interface, pod *apiv1.Pod) error {
	_, condition := podv1.GetPodCondition(&pod.Status, apiv1.PodScheduled)
	if condition != nil {
		glog.V(4).Infof("Reseting pod condition for %s/%s, last transition: %s",
			pod.Namespace, pod.Name, condition.LastTransitionTime.Time.String())
		condition.Status = apiv1.ConditionUnknown
		condition.LastTransitionTime = metav1.Time{
			Time: time.Now(),
		}
		_, err := kubeClient.Core().Pods(pod.Namespace).UpdateStatus(pod)
		return err
	}
	return fmt.Errorf("Expected condition PodScheduled")
}

// FilterOutSchedulable checks whether pods from <unschedulableCandidates> marked as unschedulable
// by Scheduler actually can't be scheduled on any node and filter out the ones that can.
func FilterOutSchedulable(unschedulableCandidates []*apiv1.Pod, nodes []*apiv1.Node, allPods []*apiv1.Pod, predicateChecker *simulator.PredicateChecker) []*apiv1.Pod {
	unschedulablePods := []*apiv1.Pod{}
	nodeNameToNodeInfo := createNodeNameToInfoMap(allPods, nodes)

	for _, pod := range unschedulableCandidates {
		if nodeName, err := predicateChecker.FitsAny(pod, nodeNameToNodeInfo); err == nil {
			glog.Warningf("Pod %s marked as unschedulable can be scheduled on %s. Ignoring in scale up.", pod.Name, nodeName)
		} else {
			unschedulablePods = append(unschedulablePods, pod)
		}
	}

	return unschedulablePods
}

// TODO: move this function to scheduler utils.
func createNodeNameToInfoMap(pods []*apiv1.Pod, nodes []*apiv1.Node) map[string]*schedulercache.NodeInfo {
	nodeNameToNodeInfo := schedulercache.CreateNodeNameToInfoMap(pods, nodes)
	for _, node := range nodes {
		if nodeInfo, found := nodeNameToNodeInfo[node.Name]; found {
			nodeInfo.SetNode(node)
		}
	}

	// Some pods may be out of sync with node lists. Removing incomplete node infos.
	keysToRemove := make([]string, 0)
	for key, nodeInfo := range nodeNameToNodeInfo {
		if nodeInfo.Node() == nil {
			keysToRemove = append(keysToRemove, key)
		}
	}
	for _, key := range keysToRemove {
		delete(nodeNameToNodeInfo, key)
	}

	return nodeNameToNodeInfo
}

// GetNodeInfosForGroups finds NodeInfos for all node groups used to manage the given nodes. It also returns a node group to sample node mapping.
// TODO(mwielgus): This returns map keyed by url, while most code (including scheduler) uses node.Name for a key.
//
// TODO(mwielgus): Review error policy - sometimes we may continue with partial errors.
func GetNodeInfosForGroups(nodes []*apiv1.Node, cloudProvider cloudprovider.CloudProvider, kubeClient kube_client.Interface,
	daemonsets []*extensionsv1.DaemonSet, predicateChecker *simulator.PredicateChecker) (map[string]*schedulercache.NodeInfo, errors.AutoscalerError) {
	result := make(map[string]*schedulercache.NodeInfo)

	// processNode returns information whether the nodeTemplate was generated and if there was an error.
	processNode := func(node *apiv1.Node) (bool, errors.AutoscalerError) {
		nodeGroup, err := cloudProvider.NodeGroupForNode(node)
		if err != nil {
			return false, errors.ToAutoscalerError(errors.CloudProviderError, err)
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			return false, nil
		}
		id := nodeGroup.Id()
		if _, found := result[id]; !found {
			// Build nodeInfo.
			nodeInfo, err := simulator.BuildNodeInfoForNode(node, kubeClient)
			if err != nil {
				return false, err
			}
			sanitizedNodeInfo, err := sanitizeNodeInfo(nodeInfo, id)
			if err != nil {
				return false, err
			}
			result[id] = sanitizedNodeInfo
			return true, nil
		}
		return false, nil
	}

	for _, node := range nodes {
		// Broken nodes might have some stuff missing. Skipping.
		if !kube_util.IsNodeReadyAndSchedulable(node) {
			continue
		}
		_, typedErr := processNode(node)
		if typedErr != nil {
			return map[string]*schedulercache.NodeInfo{}, typedErr
		}
	}
	for _, nodeGroup := range cloudProvider.NodeGroups() {
		id := nodeGroup.Id()
		if _, found := result[id]; found {
			continue
		}

		// No good template, trying to generate one. This is called only if there are no
		// working nodes in the node groups. By default CA tries to usa a real-world example.
		baseNodeInfo, err := nodeGroup.TemplateNodeInfo()
		if err != nil {
			if err == cloudprovider.ErrNotImplemented {
				continue
			} else {
				glog.Errorf("Unable to build proper template node for %s: %v", id, err)
				return map[string]*schedulercache.NodeInfo{}, errors.ToAutoscalerError(
					errors.CloudProviderError, err)
			}
		}
		pods := daemonset.GetDaemonSetPodsForNode(baseNodeInfo, daemonsets, predicateChecker)
		pods = append(pods, baseNodeInfo.Pods()...)
		fullNodeInfo := schedulercache.NewNodeInfo(pods...)
		fullNodeInfo.SetNode(baseNodeInfo.Node())
		sanitizedNodeInfo, typedErr := sanitizeNodeInfo(fullNodeInfo, id)
		if typedErr != nil {
			return map[string]*schedulercache.NodeInfo{}, typedErr
		}
		result[id] = sanitizedNodeInfo
	}

	// Last resort - unready/unschedulable nodes.
	for _, node := range nodes {
		// Allowing broken nodes
		if !kube_util.IsNodeReadyAndSchedulable(node) {
			added, typedErr := processNode(node)
			if typedErr != nil {
				return map[string]*schedulercache.NodeInfo{}, typedErr
			}
			nodeGroup, err := cloudProvider.NodeGroupForNode(node)
			if err != nil {
				return map[string]*schedulercache.NodeInfo{}, errors.ToAutoscalerError(
					errors.CloudProviderError, err)
			}
			if added {
				glog.Warningf("Built template for %s based on unready/unschedulable node %s", nodeGroup.Id(), node.Name)
			}
		}
	}

	return result, nil
}

func sanitizeNodeInfo(nodeInfo *schedulercache.NodeInfo, nodeGroupName string) (*schedulercache.NodeInfo, errors.AutoscalerError) {
	// Sanitize node name.
	sanitizedNode, err := sanitizeTemplateNode(nodeInfo.Node(), nodeGroupName)
	if err != nil {
		return nil, err
	}

	// Update nodename in pods.
	sanitizedPods := make([]*apiv1.Pod, 0)
	for _, pod := range nodeInfo.Pods() {
		obj, err := api.Scheme.DeepCopy(pod)
		if err != nil {
			return nil, errors.ToAutoscalerError(errors.InternalError, err)
		}
		sanitizedPod := obj.(*apiv1.Pod)
		sanitizedPod.Spec.NodeName = sanitizedNode.Name
		sanitizedPods = append(sanitizedPods, sanitizedPod)
	}

	// Build a new node info.
	sanitizedNodeInfo := schedulercache.NewNodeInfo(sanitizedPods...)
	if err := sanitizedNodeInfo.SetNode(sanitizedNode); err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}
	return sanitizedNodeInfo, nil
}

func sanitizeTemplateNode(node *apiv1.Node, nodeGroup string) (*apiv1.Node, errors.AutoscalerError) {
	obj, err := api.Scheme.DeepCopy(node)
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}
	nodeName := fmt.Sprintf("template-node-for-%s-%d", nodeGroup, rand.Int63())
	newNode := obj.(*apiv1.Node)
	newNode.Labels = make(map[string]string, len(node.Labels))
	for k, v := range node.Labels {
		if k != kubeletapis.LabelHostname {
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
		if taint.Key == ReschedulerTaintKey {
			glog.V(4).Infof("Removing rescheduler taint when creating template from node %s", node.Name)
		} else {
			newTaints = append(newTaints, taint)
		}
	}
	newNode.Spec.Taints = newTaints
	return newNode, nil
}

// Removes unregistered nodes if needed. Returns true if anything was removed and error if such occurred.
func removeOldUnregisteredNodes(unregisteredNodes []clusterstate.UnregisteredNode, context *AutoscalingContext,
	currentTime time.Time) (bool, error) {
	removedAny := false
	for _, unregisteredNode := range unregisteredNodes {
		if unregisteredNode.UnregisteredSince.Add(context.UnregisteredNodeRemovalTime).Before(currentTime) {
			glog.V(0).Infof("Removing unregistered node %v", unregisteredNode.Node.Name)
			nodeGroup, err := context.CloudProvider.NodeGroupForNode(unregisteredNode.Node)
			if err != nil {
				glog.Warningf("Failed to get node group for %s: %v", unregisteredNode.Node.Name, err)
				return removedAny, err
			}
			if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
				glog.Warningf("No node group for node %s, skipping", unregisteredNode.Node.Name)
				continue
			}
			err = nodeGroup.DeleteNodes([]*apiv1.Node{unregisteredNode.Node})
			if err != nil {
				glog.Warningf("Failed to remove node %s: %v", unregisteredNode.Node.Name, err)
				return removedAny, err
			}
			removedAny = true
		}
	}
	return removedAny, nil
}

// Sets the target size of node groups to the current number of nodes in them
// if the difference was constant for a prolonged time. Returns true if managed
// to fix something.
func fixNodeGroupSize(context *AutoscalingContext, currentTime time.Time) (bool, error) {
	fixed := false
	for _, nodeGroup := range context.CloudProvider.NodeGroups() {
		incorrectSize := context.ClusterStateRegistry.GetIncorrectNodeGroupSize(nodeGroup.Id())
		if incorrectSize == nil {
			continue
		}
		if incorrectSize.FirstObserved.Add(context.UnregisteredNodeRemovalTime).Before(currentTime) {
			delta := incorrectSize.CurrentSize - incorrectSize.ExpectedSize
			if delta < 0 {
				glog.V(0).Infof("Decreasing size of %s, expected=%d current=%d delta=%d", nodeGroup.Id(),
					incorrectSize.ExpectedSize,
					incorrectSize.CurrentSize,
					delta)
				if err := nodeGroup.DecreaseTargetSize(delta); err != nil {
					return fixed, fmt.Errorf("Failed to decrease %s: %v", nodeGroup.Id(), err)
				}
				fixed = true
			}
		}
	}
	return fixed, nil
}

// getManagedNodes returns the nodes managed by the cluster autoscaler.
func getManagedNodes(context *AutoscalingContext, nodes []*apiv1.Node) []*apiv1.Node {
	result := make([]*apiv1.Node, 0, len(nodes))
	for _, node := range nodes {
		nodeGroup, err := context.CloudProvider.NodeGroupForNode(node)
		if err != nil {
			glog.Warningf("Error while checking node group for %s: %v", node.Name, err)
			continue
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			glog.V(4).Infof("Skipping %s - no node group config", node.Name)
			continue
		}
		result = append(result, node)
	}
	return result
}
