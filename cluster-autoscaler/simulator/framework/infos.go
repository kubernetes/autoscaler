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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// PodInfo contains all necessary information about a Pod that Cluster Autoscaler needs to track.
type PodInfo struct {
	// This type embeds *apiv1.Pod to make the accesses easier - most of the code just needs to access the Pod.
	*apiv1.Pod

	// NeededResourceClaims contains ResourceClaim objects needed by the Pod.
	NeededResourceClaims []*resourceapi.ResourceClaim
}

type podExtraInfo struct {
	neededResourceClaims []*resourceapi.ResourceClaim
}

// NodeInfo contains all necessary information about a Node that Cluster Autoscaler needs to track.
// It's essentially a wrapper around schedulerframework.NodeInfo, with extra data on top.
type NodeInfo struct {
	// schedNodeInfo is the part of information needed by the scheduler.
	schedNodeInfo *schedulerframework.NodeInfo
	// podsExtraInfo contains extra pod-level data needed only by CA.
	podsExtraInfo map[types.UID]podExtraInfo

	// Extra node-level data needed only by CA below.

	// LocalResourceSlices contains all node-local ResourceSlices exposed by this Node.
	LocalResourceSlices []*resourceapi.ResourceSlice
}

// SetNode sets the Node in this NodeInfo
func (n *NodeInfo) SetNode(node *apiv1.Node) {
	n.schedNodeInfo.SetNode(node)
}

// Node returns the Node set in this NodeInfo.
func (n *NodeInfo) Node() *apiv1.Node {
	return n.schedNodeInfo.Node()
}

// Pods returns the Pods scheduled on this NodeInfo, along with all their associated data.
func (n *NodeInfo) Pods() []*PodInfo {
	var result []*PodInfo
	for _, pod := range n.schedNodeInfo.Pods {
		extraInfo := n.podsExtraInfo[pod.Pod.UID]
		podInfo := &PodInfo{Pod: pod.Pod, NeededResourceClaims: extraInfo.neededResourceClaims}
		result = append(result, podInfo)
	}
	return result
}

// AddPod adds the given Pod and associated data to the NodeInfo.
func (n *NodeInfo) AddPod(pod *PodInfo) {
	n.schedNodeInfo.AddPod(pod.Pod)
	n.podsExtraInfo[pod.UID] = podExtraInfo{neededResourceClaims: pod.NeededResourceClaims}
}

// RemovePod removes the given pod and its associated data from the NodeInfo.
func (n *NodeInfo) RemovePod(pod *apiv1.Pod) error {
	err := n.schedNodeInfo.RemovePod(klog.Background(), pod)
	if err != nil {
		return err
	}
	delete(n.podsExtraInfo, pod.UID)
	return nil
}

// ToScheduler returns the embedded *schedulerframework.NodeInfo portion of the tracked data.
func (n *NodeInfo) ToScheduler() *schedulerframework.NodeInfo {
	return n.schedNodeInfo
}

// DeepCopy clones the NodeInfo.
func (n *NodeInfo) DeepCopy() *NodeInfo {
	var newPods []*PodInfo
	for _, podInfo := range n.Pods() {
		var newClaims []*resourceapi.ResourceClaim
		for _, claim := range podInfo.NeededResourceClaims {
			newClaims = append(newClaims, claim.DeepCopy())
		}
		newPods = append(newPods, &PodInfo{Pod: podInfo.Pod.DeepCopy(), NeededResourceClaims: newClaims})
	}
	var newSlices []*resourceapi.ResourceSlice
	for _, slice := range n.LocalResourceSlices {
		newSlices = append(newSlices, slice.DeepCopy())
	}
	// Node() can be nil, but DeepCopy() handles nil receivers gracefully.
	return NewNodeInfo(n.Node().DeepCopy(), newSlices, newPods...)
}

// NewNodeInfo returns a new internal NodeInfo from the provided data.
func NewNodeInfo(node *apiv1.Node, slices []*resourceapi.ResourceSlice, pods ...*PodInfo) *NodeInfo {
	result := &NodeInfo{
		schedNodeInfo:       schedulerframework.NewNodeInfo(),
		podsExtraInfo:       map[types.UID]podExtraInfo{},
		LocalResourceSlices: slices,
	}
	if node != nil {
		result.schedNodeInfo.SetNode(node)
	}
	for _, pod := range pods {
		result.AddPod(pod)
	}
	return result
}

// WrapSchedulerNodeInfo wraps a *schedulerframework.NodeInfo into an internal *NodeInfo.
func WrapSchedulerNodeInfo(schedNodeInfo *schedulerframework.NodeInfo) *NodeInfo {
	return &NodeInfo{
		schedNodeInfo: schedNodeInfo,
		podsExtraInfo: map[types.UID]podExtraInfo{},
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
