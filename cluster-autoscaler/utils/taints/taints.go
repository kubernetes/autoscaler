/*
Copyright 2020 The Kubernetes Authors.

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

package taints

import (
	"strings"

	apiv1 "k8s.io/api/core/v1"
	coreUtils "k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"

	"k8s.io/klog"
)

// FilterOutNodesWithIgnoredTaints override the condition status of the given nodes to mark them as NotReady when they have
// filtered taints.
func FilterOutNodesWithIgnoredTaints(ignoredTaints coreUtils.TaintKeySet, allNodes, readyNodes []*apiv1.Node) ([]*apiv1.Node, []*apiv1.Node) {
	newAllNodes := make([]*apiv1.Node, 0)
	newReadyNodes := make([]*apiv1.Node, 0)
	nodesWithIgnoredTaints := make(map[string]*apiv1.Node)
	for _, node := range readyNodes {
		if len(node.Spec.Taints) == 0 {
			newReadyNodes = append(newReadyNodes, node)
		}
		for _, t := range node.Spec.Taints {
			_, hasIgnoredTaint := ignoredTaints[t.Key]
			if !hasIgnoredTaint && !strings.HasPrefix(t.Key, coreUtils.IgnoreTaintPrefix) {
				newReadyNodes = append(newReadyNodes, node)
				continue
			}
			nodesWithIgnoredTaints[node.Name] = kubernetes.GetUnreadyNodeCopy(node)
			klog.V(3).Infof("Overriding status of node %v, which seems to have ignored taint %q", node.Name, t.Key)
		}
	}
	// Override any node with ignored taint with its "unready" copy
	for _, node := range allNodes {
		if newNode, found := nodesWithIgnoredTaints[node.Name]; found {
			newAllNodes = append(newAllNodes, newNode)
		} else {
			newAllNodes = append(newAllNodes, node)
		}
	}
	return newAllNodes, newReadyNodes
}
