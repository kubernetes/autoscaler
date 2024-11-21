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
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/daemonset"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/labels"
	pod_util "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
)

type nodeGroupTemplateNodeInfoGetter interface {
	Id() string
	TemplateNodeInfo() (*framework.NodeInfo, error)
}

// SanitizedTemplateNodeInfoFromNodeGroup returns a template NodeInfo object based on NodeGroup.TemplateNodeInfo(). The template is sanitized, and only
// contains the pods that should appear on a new Node from the same node group (e.g. DaemonSet pods).
func SanitizedTemplateNodeInfoFromNodeGroup(nodeGroup nodeGroupTemplateNodeInfoGetter, daemonsets []*appsv1.DaemonSet, taintConfig taints.TaintConfig) (*framework.NodeInfo, errors.AutoscalerError) {
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
	sanitizedExample := sanitizeNodeInfo(example, newNodeNameBase, randSuffix, &taintConfig)
	expectedPods, err := podsExpectedOnFreshNode(sanitizedExample, daemonsets, forceDaemonSets, randSuffix)
	if err != nil {
		return nil, err
	}
	// No need to sanitize the expected pods again - they either come from sanitizedExample and were sanitized above,
	// or were added by podsExpectedOnFreshNode and sanitized there.
	return framework.NewNodeInfo(sanitizedExample.Node(), nil, expectedPods...), nil
}

// NodeInfoSanitizedDeepCopy duplicates the provided template NodeInfo, returning a fresh NodeInfo that can be injected into the cluster snapshot.
// The NodeInfo is sanitized (names, UIDs are changed, etc.), so that it can be injected along other copies created from the same template.
func NodeInfoSanitizedDeepCopy(template *framework.NodeInfo, suffix string) *framework.NodeInfo {
	// Template node infos should already have taints and pods filtered, so not setting these parameters.
	return sanitizeNodeInfo(template, template.Node().Name, suffix, nil)
}

func sanitizeNodeInfo(nodeInfo *framework.NodeInfo, newNodeNameBase string, namesSuffix string, taintConfig *taints.TaintConfig) *framework.NodeInfo {
	freshNodeName := fmt.Sprintf("%s-%s", newNodeNameBase, namesSuffix)
	freshNode := sanitizeNode(nodeInfo.Node(), freshNodeName, taintConfig)
	result := framework.NewNodeInfo(freshNode, nil)

	for _, podInfo := range nodeInfo.Pods() {
		freshPod := sanitizePod(podInfo.Pod, freshNode.Name, namesSuffix)
		result.AddPod(framework.NewPodInfo(freshPod, nil))
	}
	return result
}

func sanitizeNode(node *apiv1.Node, newName string, taintConfig *taints.TaintConfig) *apiv1.Node {
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

func sanitizePod(pod *apiv1.Pod, nodeName, nameSuffix string) *apiv1.Pod {
	sanitizedPod := pod.DeepCopy()
	sanitizedPod.UID = uuid.NewUUID()
	sanitizedPod.Name = fmt.Sprintf("%s-%s", pod.Name, nameSuffix)
	sanitizedPod.Spec.NodeName = nodeName
	return sanitizedPod
}

func podsExpectedOnFreshNode(sanitizedExampleNodeInfo *framework.NodeInfo, daemonsets []*appsv1.DaemonSet, forceDaemonSets bool, nameSuffix string) ([]*framework.PodInfo, errors.AutoscalerError) {
	var result []*framework.PodInfo
	runningDS := make(map[types.UID]bool)
	for _, pod := range sanitizedExampleNodeInfo.Pods() {
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
		// The provided nodeInfo has to have taints properly sanitized, or this won't work correctly.
		daemonPods, err := daemonset.GetDaemonSetPodsForNode(sanitizedExampleNodeInfo, pendingDS)
		if err != nil {
			return nil, errors.ToAutoscalerError(errors.InternalError, err)
		}
		for _, pod := range daemonPods {
			// There's technically no need to sanitize these pods since they're created from scratch, but
			// it's nice to have the same suffix for all names in one sanitized NodeInfo when debugging.
			result = append(result, &framework.PodInfo{Pod: sanitizePod(pod.Pod, sanitizedExampleNodeInfo.Node().Name, nameSuffix)})
		}
	}
	return result, nil
}
