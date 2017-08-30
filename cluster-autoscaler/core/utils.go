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
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"

	apiv1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	kube_client "k8s.io/client-go/kubernetes"
	api "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/helper"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
)

const (
	// ReschedulerTaintKey is the name of the taint created by rescheduler.
	ReschedulerTaintKey = "CriticalAddonsOnly"
)

// Following data structure is used to avoid running predicates #pending_pods * #nodes
// times (which turned out to be very expensive if there are thousands of pending pods).
// This optimization is based on the assumption that if there are that many pods they're
// likely created by controllers (deployment, replication controller, ...).
// So instead of running all predicates for every pod we first check whether we've
// already seen identical pod (in this step we're not binpacking, just checking if
// the pod would fit anywhere right now) and if so we use the result we already
// calculated.
// To decide if two pods are similar enough we check if they have identical label
// and spec and are owned by the same controller. The problem is the whole
// podSchedulableInfo struct is not hashable and keeping a list and runnig deep
// equality checks would likely also be expensive. So instead we use controller
// UID as a key in initial lookup and only run full comparison on a set of
// podSchedulableInfos created for pods owned by this controller.
type podSchedulableInfo struct {
	spec        apiv1.PodSpec
	labels      map[string]string
	schedulable bool
}

type podSchedulableMap map[string][]podSchedulableInfo

func (psi *podSchedulableInfo) match(pod *apiv1.Pod) bool {
	return reflect.DeepEqual(pod.Labels, psi.labels) && helper.Semantic.DeepEqual(pod.Spec, psi.spec)
}

func (podMap podSchedulableMap) get(pod *apiv1.Pod) (bool, bool) {
	ref := drain.ControllerRef(pod)
	if ref == nil {
		return false, false
	}
	uid := string(ref.UID)
	if infos, found := podMap[uid]; found {
		for _, info := range infos {
			if info.match(pod) {
				return info.schedulable, true
			}
		}
	}
	return false, false
}

func (podMap podSchedulableMap) set(pod *apiv1.Pod, schedulable bool) {
	ref := drain.ControllerRef(pod)
	if ref == nil {
		return
	}
	uid := string(ref.UID)
	podMap[uid] = append(podMap[uid], podSchedulableInfo{
		spec:        pod.Spec,
		labels:      pod.Labels,
		schedulable: schedulable,
	})
}

// FilterOutSchedulable checks whether pods from <unschedulableCandidates> marked as unschedulable
// by Scheduler actually can't be scheduled on any node and filter out the ones that can.
func FilterOutSchedulable(unschedulableCandidates []*apiv1.Pod, nodes []*apiv1.Node, allPods []*apiv1.Pod, predicateChecker *simulator.PredicateChecker) []*apiv1.Pod {
	unschedulablePods := []*apiv1.Pod{}
	nodeNameToNodeInfo := createNodeNameToInfoMap(allPods, nodes)
	podSchedulable := make(podSchedulableMap)

	for _, pod := range unschedulableCandidates {
		if schedulable, found := podSchedulable.get(pod); found {
			if !schedulable {
				unschedulablePods = append(unschedulablePods, pod)
			} else {
				glog.V(4).Infof("Pod %s marked as unschedulable can be scheduled (based on simulation run for other pod owned by the same controller). Ignoring in scale up.", pod.Name)
			}
			continue
		}
		if nodeName, err := predicateChecker.FitsAny(pod, nodeNameToNodeInfo); err == nil {
			glog.V(4).Infof("Pod %s marked as unschedulable can be scheduled on %s. Ignoring in scale up.", pod.Name, nodeName)
			podSchedulable.set(pod, true)
		} else {
			unschedulablePods = append(unschedulablePods, pod)
			podSchedulable.set(pod, false)
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

// getPotentiallyUnneededNodes returns nodes that are:
// - managed by the cluster autoscaler
// - in groups with size > min size
func getPotentiallyUnneededNodes(context *AutoscalingContext, nodes []*apiv1.Node) []*apiv1.Node {
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
		size, err := nodeGroup.TargetSize()
		if err != nil {
			glog.Errorf("Error while checking node group size %s: %v", nodeGroup.Id(), err)
			continue
		}
		if size <= nodeGroup.MinSize() {
			glog.V(1).Infof("Skipping %s - node group min size reached", node.Name)
			continue
		}
		result = append(result, node)
	}
	return result
}

// ConfigurePredicateCheckerForLoop can be run to update predicateChecker configuration
// based on current state of the cluster.
func ConfigurePredicateCheckerForLoop(unschedulablePods []*apiv1.Pod, schedulablePods []*apiv1.Pod, predicateChecker *simulator.PredicateChecker) {
	podsWithAffinityFound := false
	for _, pod := range unschedulablePods {
		if pod.Spec.Affinity != nil {
			podsWithAffinityFound = true
			break
		}
	}
	if !podsWithAffinityFound {
		for _, pod := range schedulablePods {
			if pod.Spec.Affinity != nil {
				podsWithAffinityFound = true
				break
			}
		}
	}
	predicateChecker.SetAffinityPredicateEnabled(podsWithAffinityFound)
	if !podsWithAffinityFound {
		glog.V(1).Info("No pod using affinity / antiaffinity found in cluster, disabling affinity predicate for this loop")
	}
}
