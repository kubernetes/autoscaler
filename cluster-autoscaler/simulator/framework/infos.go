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
	resourceapi "k8s.io/api/resource/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	schedulerinterface "k8s.io/kube-scheduler/framework"
	schedulerimpl "k8s.io/kubernetes/pkg/scheduler/framework"
)

// PodInfo contains all necessary information about a Pod that Cluster Autoscaler needs to track.
// TODO: Rewrite PodInfo to be an interface extending schedulerinterface.PodInfo
type PodInfo struct {
	// This type embeds *apiv1.Pod to make the accesses easier - most of the code just needs to access the Pod.
	*apiv1.Pod
	// PodExtraInfo is an embedded struct containing all additional information that CA needs to track about a Pod.
	PodExtraInfo
}

// PodExtraInfo contains all necessary information about a Pod that Cluster Autoscaler needs to track, apart from the Pod itself.
// This is extracted from PodInfo so that it can be stored separately from the Pod.
type PodExtraInfo struct {
	// NeededResourceClaims contains ResourceClaim objects needed by the Pod.
	NeededResourceClaims []*resourceapi.ResourceClaim
}

// NodeInfo contains all necessary information about a Node that Cluster Autoscaler needs to track.
// It's essentially a wrapper around schedulerimpl.NodeInfo, with extra data on top.
// TODO: Rewrite NodeInfo to be an interface extending schedulerinterface.NodeInfo
type NodeInfo struct {
	// schedNodeInfo is the part of information needed by the scheduler.
	schedNodeInfo schedulerinterface.NodeInfo
	// podsExtraInfo contains extra pod-level data needed only by CA.
	podsExtraInfo map[types.UID]PodExtraInfo

	// Extra node-level data needed only by CA below.

	// LocalResourceSlices contains all node-local ResourceSlices exposed by this Node.
	LocalResourceSlices []*resourceapi.ResourceSlice

	// CSINode contains the CSI node exposed by this Node.
	CSINode *storagev1.CSINode
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
	for _, pod := range n.schedNodeInfo.GetPods() {
		extraInfo := n.podsExtraInfo[pod.GetPod().UID]
		podInfo := &PodInfo{Pod: pod.GetPod(), PodExtraInfo: extraInfo}
		result = append(result, podInfo)
	}
	return result
}

// AddPod adds the given Pod and associated data to the NodeInfo.
func (n *NodeInfo) AddPod(pod *PodInfo) {
	podInfo, _ := schedulerimpl.NewPodInfo(pod.Pod)
	n.schedNodeInfo.AddPodInfo(podInfo)
	if len(pod.PodExtraInfo.NeededResourceClaims) > 0 {
		n.podsExtraInfo[pod.UID] = pod.PodExtraInfo
	}
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

// ToScheduler returns the embedded schedulerinterface.NodeInfo portion of the tracked data.
func (n *NodeInfo) ToScheduler() schedulerinterface.NodeInfo {
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
		newPods = append(newPods, &PodInfo{Pod: podInfo.Pod.DeepCopy(), PodExtraInfo: PodExtraInfo{NeededResourceClaims: newClaims}})
	}
	var newSlices []*resourceapi.ResourceSlice
	for _, slice := range n.LocalResourceSlices {
		newSlices = append(newSlices, slice.DeepCopy())
	}
	// Node() can be nil, but DeepCopy() handles nil receivers gracefully.
	ni := NewNodeInfo(n.Node().DeepCopy(), newSlices, newPods...)
	if n.CSINode != nil {
		ni.SetCSINode(n.CSINode.DeepCopy())
	}
	return ni
}

// ResourceClaims returns all ResourceClaims contained in the PodInfos in this NodeInfo. Shared claims
// are taken into account, each claim should only be returned once.
func (n *NodeInfo) ResourceClaims() []*resourceapi.ResourceClaim {
	processedClaims := map[types.UID]bool{}
	var result []*resourceapi.ResourceClaim
	for _, pod := range n.Pods() {
		for _, claim := range pod.NeededResourceClaims {
			if processedClaims[claim.UID] {
				// Shared claim, already grouped.
				continue
			}
			result = append(result, claim)
			processedClaims[claim.UID] = true
		}
	}
	return result
}

// SetCSINode adds a CSINode to the NodeInfo.
func (n *NodeInfo) SetCSINode(csiNode *storagev1.CSINode) *NodeInfo {
	n.CSINode = csiNode
	return n
}

// NewNodeInfo returns a new internal NodeInfo from the provided data.
func NewNodeInfo(node *apiv1.Node, slices []*resourceapi.ResourceSlice, pods ...*PodInfo) *NodeInfo {
	result := &NodeInfo{
		schedNodeInfo:       schedulerimpl.NewNodeInfo(),
		podsExtraInfo:       map[types.UID]PodExtraInfo{},
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

// WrapSchedulerNodeInfo wraps a schedulerinterface.NodeInfo into an internal *NodeInfo.
func WrapSchedulerNodeInfo(schedNodeInfo schedulerinterface.NodeInfo, slices []*resourceapi.ResourceSlice, podExtraInfos map[types.UID]PodExtraInfo) *NodeInfo {
	if podExtraInfos == nil {
		podExtraInfos = map[types.UID]PodExtraInfo{}
	}
	return &NodeInfo{
		schedNodeInfo:       schedNodeInfo,
		podsExtraInfo:       podExtraInfos,
		LocalResourceSlices: slices,
	}
}

// NewPodInfo is a convenience function for creating new PodInfos without typing out the "PodExtraInfo" part.
func NewPodInfo(pod *apiv1.Pod, claims []*resourceapi.ResourceClaim) *PodInfo {
	return &PodInfo{Pod: pod, PodExtraInfo: PodExtraInfo{NeededResourceClaims: claims}}
}
