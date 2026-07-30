/*
Copyright The Kubernetes Authors.

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

package atomic

import (
	"strconv"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/utils/annotations"
	klog "k8s.io/klog/v2"
)

// IsAtomicNodeGroup returns the node group and true if the given node belongs to an atomic node group (ZeroOrMaxNodeScaling).
func IsAtomicNodeGroup(autoscalingCtx *ca_context.AutoscalingContext, node *apiv1.Node) (cloudprovider.NodeGroup, bool) {
	nodeGroup, err := autoscalingCtx.CloudProvider.NodeGroupForNode(node)
	if err != nil || nodeGroup == nil {
		return nil, false
	}
	autoscalingOptions, err := nodeGroup.GetOptions(autoscalingCtx.NodeGroupDefaults)
	if err != nil && err != cloudprovider.ErrNotImplemented {
		klog.Errorf("Failed to get autoscaling options for node group %s: %v", nodeGroup.Id(), err)
		return nil, false
	}
	if autoscalingOptions != nil && autoscalingOptions.ZeroOrMaxNodeScaling {
		return nodeGroup, true
	}
	return nil, false
}

// CountRegisteredNodesForGroup returns the number of registered (non-upcoming) nodes in allNodes belonging to nodeGroup.
func CountRegisteredNodesForGroup(ng cloudprovider.NodeGroup, allNodes []*apiv1.Node) (int, error) {
	nodesInGroup, err := ng.Nodes()
	if err != nil {
		return 0, err
	}
	nodeByProviderID := make(map[string]bool, len(nodesInGroup))
	for _, node := range nodesInGroup {
		nodeByProviderID[node.Id] = true
	}
	count := 0
	for _, node := range allNodes {
		if val, ok := node.Annotations[annotations.NodeUpcomingAnnotation]; ok {
			if res, ok := strconv.ParseBool(val); ok == nil && res {
				continue
			}
		}
		if nodeByProviderID[node.Spec.ProviderID] {
			count++
		}
	}
	return count, nil
}
