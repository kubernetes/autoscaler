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

package eligibility

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/azure/deallocate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	klog "k8s.io/klog/v2"
)

func processNodeGroupDeallocate(context *context.AutoscalingContext, node *apiv1.Node) bool {
	nodeGroup, err := context.CloudProvider.NodeGroupForNode(node)
	if err != nil {
		klog.Errorf("Error while checking node group for %s: %v", node.Name, err)
		return true
	}
	// Skip calculating unreadiness for already deallocated nodes. This is because CA attempts to delete unready unneeded nodes
	if shouldSkipDeletionWhenDeallocated(nodeGroup, node) {
		klog.V(3).Infof("Skipping %s from scale-down considerations as the nodegroup has Deallocate policy and node is currently deallocated", node.Name)
		return true
	}
	return false
}

// shouldSkipDeletionWhenDeallocated returns true if we should skip the unneeded node calculation for this node
// This is only done for Deallocation mode to avoid unnecessary candidate consideration/Deallocating those since
// they remain in NotReady state
func shouldSkipDeletionWhenDeallocated(nodeGroup cloudprovider.NodeGroup, node *apiv1.Node) bool {
	policyNg, ok := nodeGroup.(deallocate.PolicyNodeGroup)
	if ok && policyNg.ScaleDownPolicy() != deallocate.Deallocate {
		return false
	}
	ready, _, _ := kube_util.GetReadinessState(node)
	return taints.HasShutdownTaint(node) && !ready
}
