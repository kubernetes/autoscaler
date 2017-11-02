/*
Copyright 2015 The Kubernetes Authors.

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

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
	client "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"
)

// FastGetPodsToMove returns a list of pods that should be moved elsewhere if the node
// is drained. Raises error if there is an unreplicated pod.
// Based on kubectl drain code. It makes an assumption that RC, DS, Jobs and RS were deleted
// along with their pods (no abandoned pods with dangling created-by annotation). Useful for fast
// checks.
func FastGetPodsToMove(nodeInfo *schedulercache.NodeInfo, skipNodesWithSystemPods bool, skipNodesWithLocalStorage bool,
	pdbs []*policyv1.PodDisruptionBudget) ([]*apiv1.Pod, error) {
	pods, err := drain.GetPodsForDeletionOnNodeDrain(
		nodeInfo.Pods(),
		pdbs,
		false,
		skipNodesWithSystemPods,
		skipNodesWithLocalStorage,
		false,
		nil,
		0,
		time.Now())

	if err != nil {
		return pods, err
	}
	if err := checkPdbs(pods, pdbs); err != nil {
		return []*apiv1.Pod{}, err
	}

	return pods, nil
}

// DetailedGetPodsForMove returns a list of pods that should be moved elsewhere if the node
// is drained. Raises error if there is an unreplicated pod.
// Based on kubectl drain code. It checks whether RC, DS, Jobs and RS that created these pods
// still exist.
func DetailedGetPodsForMove(nodeInfo *schedulercache.NodeInfo, skipNodesWithSystemPods bool,
	skipNodesWithLocalStorage bool, client client.Interface, minReplicaCount int32,
	pdbs []*policyv1.PodDisruptionBudget) ([]*apiv1.Pod, error) {
	pods, err := drain.GetPodsForDeletionOnNodeDrain(
		nodeInfo.Pods(),
		pdbs,
		false,
		skipNodesWithSystemPods,
		skipNodesWithLocalStorage,
		true,
		client,
		minReplicaCount,
		time.Now())
	if err != nil {
		return pods, err
	}
	if err := checkPdbs(pods, pdbs); err != nil {
		return []*apiv1.Pod{}, err
	}

	return pods, nil
}

func checkPdbs(pods []*apiv1.Pod, pdbs []*policyv1.PodDisruptionBudget) error {
	// TODO: make it more efficient.
	for _, pdb := range pdbs {
		selector, err := metav1.LabelSelectorAsSelector(pdb.Spec.Selector)
		if err != nil {
			return err
		}
		for _, pod := range pods {
			if pod.Namespace == pdb.Namespace && selector.Matches(labels.Set(pod.Labels)) {
				if pdb.Status.PodDisruptionsAllowed < 1 {
					return fmt.Errorf("no enough pod disruption budget to move %s/%s", pod.Namespace, pod.Name)
				}
			}
		}
	}
	return nil
}
