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

package clusterstate

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/azure/deallocate"
	klog "k8s.io/klog/v2"
)

func (csr *ClusterStateRegistry) calcDeallocationNodes(delta int, totalUnready []string) int {
	if isAnyNodeGroupInDeallocationMode(csr.cloudProvider.NodeGroups()) {
		totalDeallocated := csr.totalReadiness.Deallocated
		delta = len(totalUnready) - len(totalDeallocated)
		klog.V(5).Infof("Cluster Health Check - totalUnready: %d, totalDeallocated: %d", len(totalUnready), len(totalDeallocated))
	}
	return delta
}

// IsNodeGroupHealthyDeallocate returns true if the node group is in deallocate mode.
func (csr *ClusterStateRegistry) IsNodeGroupHealthyDeallocate(nodeGroup cloudprovider.NodeGroup) bool {
	// NodeGroups with deallocated nodes can have a lot of Unready nodes - disable the health check
	// for those so they can be scaled up
	policyNg, ok := nodeGroup.(deallocate.PolicyNodeGroup)
	if ok && policyNg.ScaleDownPolicy() == deallocate.Deallocate {
		return true
	}
	return false
}

// isAnyNodeGroupInDeallocationMode returns true iff any nodegroup in the list is of Deallocate policy
func isAnyNodeGroupInDeallocationMode(ngs []cloudprovider.NodeGroup) bool {
	for _, ng := range ngs {
		policyNg, ok := ng.(deallocate.PolicyNodeGroup)
		if !ok {
			return false
		}
		if policyNg.ScaleDownPolicy() == deallocate.Deallocate {
			return true
		}
	}
	return false
}
