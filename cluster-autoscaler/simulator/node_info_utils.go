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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	drautils "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/utils"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/daemonset"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/labels"
	podutils "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
)

type nodeGroupTemplateNodeInfoGetter interface {
	Id() string
	TemplateNodeInfo() (*framework.NodeInfo, error)
}

// SanitizedTemplateNodeInfoFromNodeGroup returns a template NodeInfo object based on NodeGroup.TemplateNodeInfo(). The template is sanitized, and only
// contains the pods that should appear on a new Node from the same node group (e.g. DaemonSet pods).
func SanitizedTemplateNodeInfoFromNodeGroup(nodeGroup nodeGroupTemplateNodeInfoGetter, daemonsets []*appsv1.DaemonSet, taintConfig taints.TaintConfig) (*framework.NodeInfo, errors.AutoscalerError) {
	// TODO(DRA): Figure out how to handle TemplateNodeInfo() returning DaemonSet Pods using DRA. Currently, things only work correctly if such pods are
	// already allocated by TemplateNodeInfo(). It might be better for TemplateNodeInfo() to return unallocated claims, and to run scheduler predicates and
	// compute the allocations here.
	baseNodeInfo, err := nodeGroup.TemplateNodeInfo()
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.CloudProviderError, err).AddPrefix("failed to obtain template NodeInfo from node group %q: ", nodeGroup.Id())
	}
	labels.UpdateDeprecatedLabels(baseNodeInfo.Node().ObjectMeta.Labels)

	return SanitizedTemplateNodeInfoFromNodeInfo(baseNodeInfo, nodeGroup.Id(), daemonsets, true, taintConfig)
}

// SanitizedTemplateNodeInfoFromNodeInfo returns a template NodeInfo object based on a real example NodeInfo from the cluster. The template is sanitized, and only
// contains the pods that should appear on a new Node from the same node group (e.g. DaemonSet pods).
func SanitizedTemplateNodeInfoFromNodeInfo(example *framework.NodeInfo, nodeGroupId string, daemonsets []*appsv1.DaemonSet, forceDaemonSets bool, taintConfig taints.TaintConfig) (*framework.NodeInfo, errors.AutoscalerError) {
	randSuffix := fmt.Sprintf("%d", rand.Int63())
	newNodeNameBase := fmt.Sprintf("template-node-for-%s", nodeGroupId)

	// We need to sanitize the example before determining the DS pods, since taints are checked there, and
	// we might need to filter some out during sanitization.
	sanitizedExample, err := createSanitizedNodeInfo(example, newNodeNameBase, randSuffix, &taintConfig)
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}
	expectedPods, err := podsExpectedOnFreshNode(sanitizedExample, daemonsets, forceDaemonSets, randSuffix)
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}
	// No need to sanitize the expected pods again - they either come from sanitizedExample and were sanitized above,
	// or were added by podsExpectedOnFreshNode and sanitized there.
	return framework.NewNodeInfo(sanitizedExample.Node(), sanitizedExample.LocalResourceSlices, expectedPods...), nil
}

// SanitizedNodeInfo duplicates the provided template NodeInfo, returning a fresh NodeInfo that can be injected into the cluster snapshot.
// The NodeInfo is sanitized (names, UIDs are changed, etc.), so that it can be injected along other copies created from the same template.
func SanitizedNodeInfo(template *framework.NodeInfo, suffix string) (*framework.NodeInfo, error) {
	// Template node infos should already have taints and pods filtered, so not setting these parameters.
	return createSanitizedNodeInfo(template, template.Node().Name, suffix, nil)
}

func createSanitizedNodeInfo(nodeInfo *framework.NodeInfo, newNodeNameBase string, namesSuffix string, taintConfig *taints.TaintConfig) (*framework.NodeInfo, error) {
	freshNodeName := fmt.Sprintf("%s-%s", newNodeNameBase, namesSuffix)
	freshNode := createSanitizedNode(nodeInfo.Node(), freshNodeName, taintConfig)
	freshResourceSlices, oldPoolNames, err := drautils.SanitizedNodeResourceSlices(nodeInfo.LocalResourceSlices, freshNode.Name, namesSuffix)
	if err != nil {
		return nil, err
	}
	result := framework.NewNodeInfo(freshNode, freshResourceSlices)

	for _, podInfo := range nodeInfo.Pods() {
		freshPod := createSanitizedPod(podInfo.Pod, freshNode.Name, namesSuffix)
		freshResourceClaims, err := drautils.SanitizedPodResourceClaims(freshPod, podInfo.Pod, podInfo.NeededResourceClaims, namesSuffix, freshNodeName, nodeInfo.Node().Name, oldPoolNames)
		if err != nil {
			return nil, err
		}
		result.AddPod(framework.NewPodInfo(freshPod, freshResourceClaims))
	}
	return result, nil
}

func createSanitizedNode(node *apiv1.Node, newName string, taintConfig *taints.TaintConfig) *apiv1.Node {
	newNode := node.DeepCopy()
	newNode.UID = uuid.NewUUID()

	newNode.Name = newName
	if newNode.Labels == nil {
		newNode.Labels = make(map[string]string)
	}
	newNode.Labels[apiv1.LabelHostname] = newName

	if taintConfig != nil {
		newNode.Spec.Taints = taints.SanitizeTaints(newNode.Spec.Taints, *taintConfig)
	}
	return newNode
}

func createSanitizedPod(pod *apiv1.Pod, nodeName, nameSuffix string) *apiv1.Pod {
	sanitizedPod := drautils.SanitizedResourceClaimRefs(pod, nameSuffix)
	sanitizedPod.UID = uuid.NewUUID()
	sanitizedPod.Name = fmt.Sprintf("%s-%s", pod.Name, nameSuffix)
	sanitizedPod.Spec.NodeName = nodeName
	return sanitizedPod
}

func podsExpectedOnFreshNode(sanitizedExampleNodeInfo *framework.NodeInfo, daemonsets []*appsv1.DaemonSet, forceDaemonSets bool, nameSuffix string) ([]*framework.PodInfo, error) {
	var result []*framework.PodInfo
	runningDS := make(map[types.UID]bool)
	for _, pod := range sanitizedExampleNodeInfo.Pods() {
		// Ignore scheduled pods in deletion phase
		if pod.DeletionTimestamp != nil {
			continue
		}
		// Add scheduled mirror and DS pods
		if podutils.IsMirrorPod(pod.Pod) || podutils.IsDaemonSetPod(pod.Pod) {
			result = append(result, pod)
		}
		// Mark DS pods as running
		controllerRef := metav1.GetControllerOf(pod)
		if controllerRef != nil && controllerRef.Kind == "DaemonSet" {
			runningDS[controllerRef.UID] = true
		}
	}
	// Add all pending DS pods if force scheduling DS
	// TODO(DRA): Figure out how to make this work for DS pods using DRA. Currently such pods would get force-added to the
	// ClusterSnapshot, but the ResourceClaims reflecting their DRA usage on the Node wouldn't. So CA would be overestimating
	// available DRA resources on the Node.
	if forceDaemonSets {
		var pendingDS []*appsv1.DaemonSet
		for _, ds := range daemonsets {
			if !runningDS[ds.UID] {
				pendingDS = append(pendingDS, ds)
			}
		}
		// The provided nodeInfo has to have taints properly sanitized, or this won't work correctly.
		daemonPods, err := daemonset.GetDaemonSetPodsForNode(sanitizedExampleNodeInfo, pendingDS)
		if err != nil {
			return nil, err
		}
		for _, pod := range daemonPods {
			// There's technically no need to sanitize these pods since they're created from scratch, but
			// it's nice to have the same suffix for all names in one sanitized NodeInfo when debugging.
			result = append(result, &framework.PodInfo{Pod: createSanitizedPod(pod.Pod, sanitizedExampleNodeInfo.Node().Name, nameSuffix)})
		}
	}
	return result, nil
}
