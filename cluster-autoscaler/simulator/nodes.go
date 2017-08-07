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
	"time"

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_client "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"
)

// GetRequiredPodsForNode returns a list od pods that would appear on the node if the
// node was just created (like daemonset and manifest-run pods). It reuses kubectl
// drain command to get the list.
func GetRequiredPodsForNode(nodename string, client kube_client.Interface) ([]*apiv1.Pod, errors.AutoscalerError) {

	// TODO: we should change this to use informer
	podListResult, err := client.Core().Pods(apiv1.NamespaceAll).List(
		metav1.ListOptions{FieldSelector: fields.SelectorFromSet(fields.Set{"spec.nodeName": nodename}).String()})
	if err != nil {
		return []*apiv1.Pod{}, errors.ToAutoscalerError(errors.ApiCallError, err)
	}
	allPods := make([]*apiv1.Pod, 0)
	for i := range podListResult.Items {
		allPods = append(allPods, &podListResult.Items[i])
	}

	podsToRemoveList, err := drain.GetPodsForDeletionOnNodeDrain(
		allPods,
		[]*policyv1.PodDisruptionBudget{}, // PDBs are irrelevant when considering new node.
		true, // Force all removals.
		false,
		false,
		false, // Setting this to true requires client to be not-null.
		nil,
		0,
		time.Now())
	if err != nil {
		return []*apiv1.Pod{}, errors.ToAutoscalerError(errors.InternalError, err)
	}

	podsToRemoveMap := make(map[string]struct{})
	for _, pod := range podsToRemoveList {
		podsToRemoveMap[pod.SelfLink] = struct{}{}
	}

	podsOnNewNode := make([]*apiv1.Pod, 0)
	for _, pod := range allPods {
		if pod.DeletionTimestamp != nil {
			continue
		}

		if _, found := podsToRemoveMap[pod.SelfLink]; !found {
			podsOnNewNode = append(podsOnNewNode, pod)
		}
	}
	return podsOnNewNode, nil
}

// BuildNodeInfoForNode build a NodeInfo structure for the given node as if the node was just created.
func BuildNodeInfoForNode(node *apiv1.Node, client kube_client.Interface) (*schedulercache.NodeInfo, errors.AutoscalerError) {
	requiredPods, err := GetRequiredPodsForNode(node.Name, client)
	if err != nil {
		return nil, err
	}
	result := schedulercache.NewNodeInfo(requiredPods...)
	if err := result.SetNode(node); err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}
	return result, nil
}
