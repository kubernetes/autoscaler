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
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/types"
	schedulerinterface "k8s.io/kube-scheduler/framework"
	schedulerimpl "k8s.io/kubernetes/pkg/scheduler/framework"
)

// PodInfo contains all necessary information about a Pod that Cluster Autoscaler needs to track.
// Use NewPodInfo to create new objects. The fields are exported for convenience, but should be treated as read-only.
// Manual initialization may result in errors.
type PodInfo struct {
	// This type embeds *apiv1.Pod to make the accesses easier - most of the code just needs to access the Pod.
	*apiv1.Pod
	// Embed *schedulerimpl.PodInfo to implement the interface.
	*schedulerimpl.PodInfo
	// NeededResourceClaims contains ResourceClaim objects needed by the Pod.
	//
	// TODO(DRA): Beware that when modifying the DRA Snapshot (e.g., scheduling a pod), it uses a
	// copy-on-write mechanism for ResourceClaims. Existing PodInfos/NodeInfos are not updated with
	// the new pointers. Consequently, the ResourceClaims in this field may contain stale data,
	// specifically in the `Status.ReservedFor` field, lacking recent modifications.
	// Currently, this is not an issue as CA logic outside the ClusterSnapshot does not depend on
	// the `ReservedFor` field, but it should be addressed if this dependency changes in the future.
	NeededResourceClaims []*resourceapi.ResourceClaim
}

// NewPodInfo returns a new internal PodInfo from the provided data.
func NewPodInfo(pod *apiv1.Pod, claims []*resourceapi.ResourceClaim) *PodInfo {
	pi, _ := schedulerimpl.NewPodInfo(pod)
	return &PodInfo{Pod: pod, PodInfo: pi, NeededResourceClaims: claims}
}

// NodeInfo contains all necessary information about a Node that Cluster Autoscaler needs to track.
// It's essentially a wrapper around schedulerimpl.NodeInfo, with extra data on top.
type NodeInfo struct {
	// Embed NodeInfo to implement the interface.
	*schedulerimpl.NodeInfo

	// Extra node-level data needed only by CA below.

	// LocalResourceSlices contains all node-local ResourceSlices exposed by this Node.
	LocalResourceSlices []*resourceapi.ResourceSlice

	// CSINode contains the CSI node exposed by this Node.
	CSINode *storagev1.CSINode
}

// Pods returns the Pods scheduled on this NodeInfo, along with all their associated data.
func (n *NodeInfo) Pods() []*PodInfo {
	if n == nil {
		return nil
	}
	var result []*PodInfo
	for _, pod := range n.GetPods() {
		switch v := pod.(type) {
		case *PodInfo:
			result = append(result, v)
		default:
			panic(fmt.Errorf("unexpected type %T in internal NodeInfo - this must not happen", pod))
		}
	}
	return result
}

// AddPod adds the given Pod and associated data to the NodeInfo.
func (n *NodeInfo) AddPod(pod *PodInfo) {
	n.NodeInfo.AddPodInfo(pod)
}

// AddPodInfo adds the given PodInfo to the NodeInfo.
// The underlying type must be *PodInfo.
func (n *NodeInfo) AddPodInfo(pod schedulerinterface.PodInfo) {
	// Allow only internal PodInfo.
	// It's better to panic rather than progress without DRA info and make wrong decision.
	switch pod.(type) {
	case *PodInfo:
	default:
		panic(fmt.Errorf("unexpected type %T in internal NodeInfo - this must not happen", pod))
	}

	n.NodeInfo.AddPodInfo(pod)
}

// DeepCopy clones the NodeInfo.
func (n *NodeInfo) DeepCopy() *NodeInfo {
	var newPods []*PodInfo
	for _, podInfo := range n.Pods() {
		var newClaims []*resourceapi.ResourceClaim
		for _, claim := range podInfo.NeededResourceClaims {
			newClaims = append(newClaims, claim.DeepCopy())
		}
		newPods = append(newPods, NewPodInfo(podInfo.Pod.DeepCopy(), newClaims))
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

// Snapshot returns a shallow copy of NodeInfo that is efficient to create.
func (n *NodeInfo) Snapshot() schedulerinterface.NodeInfo {
	if n == nil {
		return nil
	}
	return &NodeInfo{
		NodeInfo:            n.NodeInfo.Snapshot().(*schedulerimpl.NodeInfo),
		LocalResourceSlices: append([]*resourceapi.ResourceSlice(nil), n.LocalResourceSlices...),
		CSINode:             n.CSINode,
	}
}

// ResourceClaims returns all ResourceClaims contained in the PodInfos in this NodeInfo. Shared claims
// are taken into account, each claim should only be returned once.
//
// TODO(DRA): Beware that it may return stale ResourceClaim data.
// See PodInfo.NeededResourceClaims comment.
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
		NodeInfo:            schedulerimpl.NewNodeInfo(),
		LocalResourceSlices: slices,
	}
	if node != nil {
		result.SetNode(node)
	}
	for _, pod := range pods {
		result.AddPod(pod)
	}
	return result
}
