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

package framework

import (
	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1alpha3"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// NodeInfo contains all necessary information about a Node that Cluster Autoscaler needs to track.
type NodeInfo struct {
	// This type embeds the *schedulerframework.NodeInfo, and adds additional data on top - both on the
	// Node and Pod level.
	*schedulerframework.NodeInfo
	// Pods should always match the embedded NodeInfo.Pods. This way the schedulerframework.PodInfo can be
	// accessed via .NodeInfo.Pods[i], and the internal one via .Pods[i]. CA code normally only needs to
	// access the Pod, so iterating over the internal version should be enough.
	Pods []*PodInfo

	// LocalResourceSlices contains all node-local ResourceSlices exposed by this Node.
	LocalResourceSlices []*resourceapi.ResourceSlice
}

// PodInfo contains all necessary information about a Pod that Cluster Autoscaler needs to track.
type PodInfo struct {
	// This type embeds *apiv1.Pod to make the accesses easier - most of the code just needs to access the Pod.
	*apiv1.Pod

	// NeededResourceClaims contains ResourceClaim objects needed by the Pod.
	NeededResourceClaims []*resourceapi.ResourceClaim
}

// NewNodeInfo returns a new internal NodeInfo from the provided data.
func NewNodeInfo(node *apiv1.Node, slices []*resourceapi.ResourceSlice, pods ...*PodInfo) *NodeInfo {
	result := &NodeInfo{
		NodeInfo:            schedulerframework.NewNodeInfo(),
		LocalResourceSlices: slices,
	}
	if node != nil {
		result.NodeInfo.SetNode(node)
	}
	for _, pod := range pods {
		result.AddPod(pod)
	}
	return result
}

// WrapSchedulerNodeInfo wraps a *schedulerframework.NodeInfo into an internal *NodeInfo.
func WrapSchedulerNodeInfo(schedNodeInfo *schedulerframework.NodeInfo) *NodeInfo {
	var pods []*PodInfo
	for _, podInfo := range schedNodeInfo.Pods {
		pods = append(pods, &PodInfo{Pod: podInfo.Pod})
	}
	return &NodeInfo{
		NodeInfo: schedNodeInfo,
		Pods:     pods,
	}
}

// WrapSchedulerNodeInfos wraps a list of *schedulerframework.NodeInfos into internal *NodeInfos.
func WrapSchedulerNodeInfos(schedNodeInfos []*schedulerframework.NodeInfo) []*NodeInfo {
	var result []*NodeInfo
	for _, schedNodeInfo := range schedNodeInfos {
		result = append(result, WrapSchedulerNodeInfo(schedNodeInfo))
	}
	return result
}

// AddPod adds the given Pod and associated data to the NodeInfo. The Pod is also added to the
// embedded schedulerframework.NodeInfo, so that the two lists stay in sync.
func (n *NodeInfo) AddPod(pod *PodInfo) {
	n.NodeInfo.AddPod(pod.Pod)
	n.Pods = append(n.Pods, pod)
}

// ToScheduler returns the embedded *schedulerframework.NodeInfo portion of the tracked data.
func (n *NodeInfo) ToScheduler() *schedulerframework.NodeInfo {
	return n.NodeInfo
}
