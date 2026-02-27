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

package core

import (
	"reflect"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/azure/deallocate"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	klog "k8s.io/klog/v2"
)

// cleanUpTaintsFromDeallocatedNodes cleans up the ToBeDeleted taint from already deallocated nodes
// so in the future they can be started and workloads can be scheduled on them
// This is a best effort check for "Deallocation" because it can take sometime for cloudprovider
// to apply the shutdown taint. In those cases, we resort to the unreacahble taint.
// TODO: When removing deallocate mode, remove this func.
func (a *StaticAutoscaler) cleanUpTaintsFromDeallocatedNodes(allNodes []*apiv1.Node) {
	for _, node := range allNodes {
		if !(taints.HasShutdownTaint(node) || taints.HasUnreachableTaint(node)) || !taints.HasToBeDeletedTaint(node) {
			continue
		}
		nodeGroup, err := a.AutoscalingContext.CloudProvider.NodeGroupForNode(node)
		if err != nil {
			klog.V(5).Infof("Failed to get node group for %s: %v", node.Name, err)
			continue
		}
		if nodeGroup == nil || reflect.ValueOf(nodeGroup).IsNil() {
			klog.V(5).Infof("No node group for node %s, skipping", node)
			continue
		}
		// TODO: remove this check when deallocate mode is removed
		ng, ok := nodeGroup.(deallocate.PolicyNodeGroup)
		if ok && ng.ScaleDownPolicy() == deallocate.Deallocate {
			klog.V(3).Infof("Node %s is deallocated - attempting to remove taint from the node.", node.Name)
			if _, err := taints.CleanToBeDeleted(node, a.ClientSet, a.AutoscalingContext.CordonNodeBeforeTerminate); err != nil {
				klog.Errorf("error while removing taint from node %s: %s", node.Name, err.Error())
			}
		}
	}
}
