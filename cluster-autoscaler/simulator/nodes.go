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
	kube_api "k8s.io/kubernetes/pkg/api"
	kube_client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/fields"
	cmd "k8s.io/kubernetes/pkg/kubectl/cmd"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"
)

// GetRequiredPodsForNode returns a list od pods that would appear on the node if the
// node was just created (like deamonset and manifest-run pods). It reuses kubectl
// drain command to get the list.
func GetRequiredPodsForNode(nodename string, client *kube_client.Client) ([]*kube_api.Pod, error) {
	podsToRemoveList, _, _, err := cmd.GetPodsForDeletionOnNodeDrain(client, nodename,
		kube_api.Codecs.UniversalDecoder(), true, true)
	if err != nil {
		return []*kube_api.Pod{}, err
	}

	podsToRemoveMap := make(map[string]struct{})
	for _, pod := range podsToRemoveList {
		podsToRemoveMap[pod.SelfLink] = struct{}{}
	}

	allPodList, err := client.Pods(kube_api.NamespaceAll).List(
		kube_api.ListOptions{FieldSelector: fields.SelectorFromSet(fields.Set{"spec.nodeName": nodename})})
	if err != nil {
		return []*kube_api.Pod{}, err
	}

	podsOnNewNode := make([]*kube_api.Pod, 0)
	for i, pod := range allPodList.Items {
		if _, found := podsToRemoveMap[pod.SelfLink]; !found {
			podsOnNewNode = append(podsOnNewNode, &allPodList.Items[i])
		}
	}
	return podsOnNewNode, nil
}

// BuildNodeInfoForNode build a NodeInfo structure for the given node as if the node was just created.
func BuildNodeInfoForNode(node *kube_api.Node, client *kube_client.Client) (*schedulercache.NodeInfo, error) {
	requiredPods, err := GetRequiredPodsForNode(node.Name, client)
	if err != nil {
		return nil, err
	}
	result := schedulercache.NewNodeInfo(requiredPods...)
	if err := result.SetNode(node); err != nil {
		return nil, err
	}
	return result, nil
}
