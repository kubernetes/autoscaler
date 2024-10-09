/*
Copyright 2024 The Kubernetes Authors.

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
	"math/rand"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/dynamicresources"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/daemonset"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/labels"
	pod_util "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
)

// TemplateNodeInfoFromNodeGroupTemplate returns a template NodeInfo object based on NodeGroup.TemplateNodeInfo(). The template is sanitized, and only
// contains the pods that should appear on a new Node from the same node group (e.g. DaemonSet pods).
func TemplateNodeInfoFromNodeGroupTemplate(nodeGroup cloudprovider.NodeGroup, daemonsets []*appsv1.DaemonSet, taintConfig taints.TaintConfig) (*framework.NodeInfo, errors.AutoscalerError) {
	id := nodeGroup.Id()
	baseNodeInfo, err := nodeGroup.TemplateNodeInfo()
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.CloudProviderError, err)
	}
	labels.UpdateDeprecatedLabels(baseNodeInfo.Node().ObjectMeta.Labels)

	return TemplateNodeInfoFromExampleNodeInfo(baseNodeInfo, id, daemonsets, true, taintConfig)
}

// TemplateNodeInfoFromExampleNodeInfo returns a template NodeInfo object based on a real example NodeInfo from the cluster. The template is sanitized, and only
// contains the pods that should appear on a new Node from the same node group (e.g. DaemonSet pods).
func TemplateNodeInfoFromExampleNodeInfo(example *framework.NodeInfo, nodeGroupId string, daemonsets []*appsv1.DaemonSet, forceDaemonSets bool, taintConfig taints.TaintConfig) (*framework.NodeInfo, errors.AutoscalerError) {
	randSuffix := fmt.Sprintf("%d", rand.Int63())
	newNodeNameBase := fmt.Sprintf("template-node-for-%s", nodeGroupId)

	expectedPods, err := podsExpectedOnFreshNode(example, daemonsets, forceDaemonSets)
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}
	exampleWithOnlyExpectedPods := framework.NewNodeInfo(example.Node(), example.LocalResourceSlices, expectedPods...)

	templateNodeInfo, err := sanitizeNodeInfo(exampleWithOnlyExpectedPods, newNodeNameBase, randSuffix, &taintConfig)
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}
	return templateNodeInfo, nil
}

// FreshNodeInfoFromTemplateNodeInfo duplicates the provided template NodeInfo, returning a fresh NodeInfo that can be injected into the cluster snapshot.
// The NodeInfo is sanitized (names, UIDs are changed, etc.), so that it can be injected along other copies created from the same template.
func FreshNodeInfoFromTemplateNodeInfo(template *framework.NodeInfo, suffix string) (*framework.NodeInfo, error) {
	// Template node infos should already have taints and pods filtered, so not setting these parameters.
	return sanitizeNodeInfo(template, template.Node().Name, suffix, nil)
}

// DeepCopyNodeInfo clones the provided NodeInfo
func DeepCopyNodeInfo(nodeInfo *framework.NodeInfo) *framework.NodeInfo {
	var podsCopy []*framework.PodInfo
	for _, podInfo := range nodeInfo.Pods {
		var claimsCopy []*resourceapi.ResourceClaim
		for _, claim := range podInfo.NeededResourceClaims {
			claimsCopy = append(claimsCopy, claim.DeepCopy())
		}
		podsCopy = append(podsCopy, &framework.PodInfo{Pod: podInfo.Pod.DeepCopy(), NeededResourceClaims: claimsCopy})
	}
	var slicesCopy []*resourceapi.ResourceSlice
	for _, slice := range nodeInfo.LocalResourceSlices {
		slicesCopy = append(slicesCopy, slice.DeepCopy())
	}
	newNodeInfo := framework.NewNodeInfo(nodeInfo.Node().DeepCopy(), slicesCopy, podsCopy...)
	return newNodeInfo
}

func sanitizeNodeInfo(nodeInfo *framework.NodeInfo, newNodeNameBase string, namesSuffix string, taintConfig *taints.TaintConfig) (*framework.NodeInfo, error) {
	freshNodeName := fmt.Sprintf("%s-%s", newNodeNameBase, namesSuffix)
	freshNode := sanitizeNode(nodeInfo.Node(), freshNodeName, taintConfig)
	freshResourceSlices, oldPoolNames := dynamicresources.SanitizeNodeResourceSlices(nodeInfo.LocalResourceSlices, freshNode.Name, namesSuffix)
	result := framework.NewNodeInfo(freshNode, freshResourceSlices)

	for _, podInfo := range nodeInfo.Pods {
		freshPod := sanitizePod(podInfo.Pod, freshNode.Name, namesSuffix)
		freshResourceClaims, err := dynamicresources.SanitizePodResourceClaims(freshPod, podInfo.Pod, podInfo.NeededResourceClaims, false, namesSuffix, nodeInfo.Node().Name, freshNodeName, oldPoolNames)
		if err != nil {
			return nil, err
		}
		result.AddPod(&framework.PodInfo{Pod: freshPod, NeededResourceClaims: freshResourceClaims})
	}
	return result, nil
}

func sanitizeNode(node *apiv1.Node, newName string, taintConfig *taints.TaintConfig) *apiv1.Node {
	newNode := node.DeepCopy()
	newNode.Labels = make(map[string]string, len(node.Labels))
	for k, v := range node.Labels {
		if k != apiv1.LabelHostname {
			newNode.Labels[k] = v
		} else {
			newNode.Labels[k] = newName
		}
	}
	newNode.Name = newName
	newNode.UID = uuid.NewUUID()
	if taintConfig != nil {
		newNode.Spec.Taints = taints.SanitizeTaints(newNode.Spec.Taints, *taintConfig)
	}
	return newNode
}

func sanitizePod(pod *apiv1.Pod, nodeName, nameSuffix string) *apiv1.Pod {
	sanitizedPod := dynamicresources.SanitizeResourceClaimRefs(pod, nameSuffix)
	sanitizedPod.UID = uuid.NewUUID()
	sanitizedPod.Name = fmt.Sprintf("%s-%s", pod.Name, nameSuffix)
	sanitizedPod.Spec.NodeName = nodeName
	return sanitizedPod
}

func podsExpectedOnFreshNode(exampleNodeInfo *framework.NodeInfo, daemonsets []*appsv1.DaemonSet, forceDaemonSets bool) ([]*framework.PodInfo, error) {
	var result []*framework.PodInfo
	runningDS := make(map[types.UID]bool)
	for _, pod := range exampleNodeInfo.Pods {
		// Ignore scheduled pods in deletion phase
		if pod.DeletionTimestamp != nil {
			continue
		}
		// Add scheduled mirror and DS pods
		if pod_util.IsMirrorPod(pod.Pod) || pod_util.IsDaemonSetPod(pod.Pod) {
			result = append(result, pod)
		}
		// Mark DS pods as running
		controllerRef := metav1.GetControllerOf(pod)
		if controllerRef != nil && controllerRef.Kind == "DaemonSet" {
			runningDS[controllerRef.UID] = true
		}
	}
	// Add all pending DS pods if force scheduling DS
	if forceDaemonSets {
		var pendingDS []*appsv1.DaemonSet
		for _, ds := range daemonsets {
			if !runningDS[ds.UID] {
				pendingDS = append(pendingDS, ds)
			}
		}
		daemonPods, err := daemonset.GetDaemonSetPodsForNode(exampleNodeInfo, pendingDS)
		if err != nil {
			return nil, err
		}
		for _, pod := range daemonPods {
			result = append(result, pod)
		}
	}
	return result, nil
}
