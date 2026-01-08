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

package snapshot

import (
	"fmt"
	"maps"

	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/common"
	drautils "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/utils"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	resourceclaim "k8s.io/dynamic-resource-allocation/resourceclaim"
	"k8s.io/klog/v2"
	schedulerinterface "k8s.io/kube-scheduler/framework"
)

// ResourceClaimId is a unique identifier for a ResourceClaim.
type ResourceClaimId struct {
	Name      string
	Namespace string
}

// GetClaimId returns the unique identifier for a ResourceClaim.
func GetClaimId(claim *resourceapi.ResourceClaim) ResourceClaimId {
	return ResourceClaimId{Name: claim.Name, Namespace: claim.Namespace}
}

// Snapshot contains a snapshot of all DRA objects taken at a ~single point in time. The Snapshot should be
// obtained via the Provider. Then, it can be modified using the exposed methods, to simulate scheduling actions
// in the cluster.
type Snapshot struct {
	resourceClaims *common.PatchSet[ResourceClaimId, *resourceapi.ResourceClaim]
	resourceSlices *common.PatchSet[string, []*resourceapi.ResourceSlice]
	deviceClasses  *common.PatchSet[string, *resourceapi.DeviceClass]
}

// nonNodeLocalResourceSlicesIdentifier is a special key used in the resourceSlices patchSet
// to store ResourceSlices that apply do not apply to specific nodes. The string itself is
// using a value which kubernetes node names cannot possibly have to avoid having collisions.
const nonNodeLocalResourceSlicesIdentifier = "--NON-LOCAL--"

// NewSnapshot returns a Snapshot created from the provided data.
func NewSnapshot(claims map[ResourceClaimId]*resourceapi.ResourceClaim, nodeLocalSlices map[string][]*resourceapi.ResourceSlice, nonNodeLocalSlices []*resourceapi.ResourceSlice, deviceClasses map[string]*resourceapi.DeviceClass) *Snapshot {
	slices := make(map[string][]*resourceapi.ResourceSlice, len(nodeLocalSlices)+1)
	maps.Copy(slices, nodeLocalSlices)
	slices[nonNodeLocalResourceSlicesIdentifier] = nonNodeLocalSlices

	claimsPatch := common.NewPatchFromMap(claims)
	slicesPatch := common.NewPatchFromMap(slices)
	devicesPatch := common.NewPatchFromMap(deviceClasses)
	return &Snapshot{
		resourceClaims: common.NewPatchSet(claimsPatch),
		resourceSlices: common.NewPatchSet(slicesPatch),
		deviceClasses:  common.NewPatchSet(devicesPatch),
	}
}

// NewEmptySnapshot returns a zero initialized Snapshot.
func NewEmptySnapshot() *Snapshot {
	claimsPatch := common.NewPatch[ResourceClaimId, *resourceapi.ResourceClaim]()
	slicesPatch := common.NewPatch[string, []*resourceapi.ResourceSlice]()
	devicesPatch := common.NewPatch[string, *resourceapi.DeviceClass]()
	return &Snapshot{
		resourceClaims: common.NewPatchSet(claimsPatch),
		resourceSlices: common.NewPatchSet(slicesPatch),
		deviceClasses:  common.NewPatchSet(devicesPatch),
	}
}

// ResourceClaims exposes the Snapshot as schedulerinterface.ResourceClaimTracker, in order to interact with
// the scheduler framework.
func (s *Snapshot) ResourceClaims() schedulerinterface.ResourceClaimTracker {
	return snapshotClaimTracker{snapshot: s}
}

// ResourceSlices exposes the Snapshot as schedulerinterface.ResourceSliceLister, in order to interact with
// the scheduler framework.
func (s *Snapshot) ResourceSlices() schedulerinterface.ResourceSliceLister {
	return snapshotSliceLister{snapshot: s}
}

// DeviceClasses exposes the Snapshot as schedulerinterface.DeviceClassLister, in order to interact with
// the scheduler framework.
func (s *Snapshot) DeviceClasses() schedulerinterface.DeviceClassLister {
	return snapshotClassLister{snapshot: s}
}

// DeviceClassResolver exposes the Snapshot as schedulerinterface.DeviceClassResolver, in order to interact with
// the scheduler framework.
func (s *Snapshot) DeviceClassResolver() schedulerinterface.DeviceClassResolver {
	return newSnapshotDeviceClassResolver(s)
}

// WrapSchedulerNodeInfo wraps the provided schedulerinterface.NodeInfo into an internal *framework.NodeInfo, adding
// dra information. Node-local ResourceSlices are added to the NodeInfo, and all ResourceClaims referenced by each Pod
// are added to each PodInfo. Returns an error if any of the Pods is missing a ResourceClaim.
func (s *Snapshot) WrapSchedulerNodeInfo(schedNodeInfo schedulerinterface.NodeInfo) (*framework.NodeInfo, error) {
	podExtraInfos := make(map[types.UID]framework.PodExtraInfo, len(schedNodeInfo.GetPods()))
	for _, pod := range schedNodeInfo.GetPods() {
		podClaims, err := s.PodClaims(pod.GetPod())
		if err != nil {
			return nil, err
		}
		if len(podClaims) > 0 {
			podExtraInfos[pod.GetPod().UID] = framework.PodExtraInfo{NeededResourceClaims: podClaims}
		}
	}
	nodeSlices, _ := s.NodeResourceSlices(schedNodeInfo.Node().Name)
	return framework.WrapSchedulerNodeInfo(schedNodeInfo, nodeSlices, podExtraInfos), nil
}

// AddClaims adds additional ResourceClaims to the Snapshot. It can be used e.g. if we need to duplicate a Pod that
// owns ResourceClaims. Returns an error if any of the claims is already tracked in the snapshot.
func (s *Snapshot) AddClaims(newClaims []*resourceapi.ResourceClaim) error {
	for _, claim := range newClaims {
		if _, found := s.resourceClaims.FindValue(GetClaimId(claim)); found {
			return fmt.Errorf("claim %s/%s already tracked in the snapshot", claim.Namespace, claim.Name)
		}
	}

	for _, claim := range newClaims {
		s.resourceClaims.SetCurrent(GetClaimId(claim), claim)
	}

	return nil
}

// PodClaims returns ResourceClaims objects for all claims referenced by the Pod. If any of the referenced claims
// isn't tracked in the Snapshot, an error is returned.
func (s *Snapshot) PodClaims(pod *apiv1.Pod) ([]*resourceapi.ResourceClaim, error) {
	return s.findPodClaims(pod, false)
}

// RemovePodOwnedClaims iterates over all the claims referenced by the Pod, and removes the ones owned by the Pod from the Snapshot.
// Claims referenced by the Pod but not owned by it are not removed, but the Pod's reservation is removed from them.
// This method removes all relevant claims that are in the snapshot, and doesn't error out if any of the claims are missing.
func (s *Snapshot) RemovePodOwnedClaims(pod *apiv1.Pod) {
	claims, err := s.findPodClaims(pod, true)
	if err != nil {
		klog.Errorf("Snapshot.RemovePodOwnedClaims ignored an error: %s", err)
	}

	for _, claim := range claims {
		claimId := GetClaimId(claim)
		if err := resourceclaim.IsForPod(pod, claim); err == nil {
			s.resourceClaims.DeleteCurrent(claimId)
			continue
		}

		claim := s.ensureClaimWritable(claim)
		drautils.ClearPodReservationInPlace(claim, pod)
		s.resourceClaims.SetCurrent(claimId, claim)
	}
}

// ReservePodClaims adds a reservation for the provided Pod to all the claims it references. If any of the referenced
// claims isn't tracked in the Snapshot, or if any of the claims are already at maximum reservation count, an error is
// returned.
func (s *Snapshot) ReservePodClaims(pod *apiv1.Pod) error {
	claims, err := s.findPodClaims(pod, false)
	if err != nil {
		return err
	}

	for _, claim := range claims {
		if drautils.ClaimFullyReserved(claim) && !resourceclaim.IsReservedForPod(pod, claim) {
			return fmt.Errorf("claim %s/%s already has max number of reservations set, can't add more", claim.Namespace, claim.Name)
		}
	}

	for _, claim := range claims {
		claimId := GetClaimId(claim)
		claim := s.ensureClaimWritable(claim)
		drautils.AddPodReservationInPlace(claim, pod)
		s.resourceClaims.SetCurrent(claimId, claim)
	}

	return nil
}

// UnreservePodClaims removes reservations for the provided Pod from all the claims it references. If any of the referenced
// claims isn't tracked in the Snapshot, an error is returned. If a claim is owned by the pod, or if the claim has no more reservations,
// its allocation is cleared.
func (s *Snapshot) UnreservePodClaims(pod *apiv1.Pod) error {
	claims, err := s.findPodClaims(pod, false)
	if err != nil {
		return err
	}

	for _, claim := range claims {
		claimId := GetClaimId(claim)
		claim := s.ensureClaimWritable(claim)
		drautils.ClearPodReservationInPlace(claim, pod)
		if err := resourceclaim.IsForPod(pod, claim); err == nil || !drautils.ClaimInUse(claim) {
			drautils.DeallocateClaimInPlace(claim)
		}

		s.resourceClaims.SetCurrent(claimId, claim)
	}
	return nil
}

// NodeResourceSlices returns all node-local ResourceSlices for the given Node.
func (s *Snapshot) NodeResourceSlices(nodeName string) ([]*resourceapi.ResourceSlice, bool) {
	slices, found := s.resourceSlices.FindValue(nodeName)
	return slices, found
}

// AddNodeResourceSlices adds additional node-local ResourceSlices to the Snapshot. This should be used whenever a Node with
// node-local ResourceSlices is duplicated in the cluster snapshot.
func (s *Snapshot) AddNodeResourceSlices(nodeName string, slices []*resourceapi.ResourceSlice) error {
	if _, alreadyInSnapshot := s.NodeResourceSlices(nodeName); alreadyInSnapshot {
		return fmt.Errorf("node %q ResourceSlices already present", nodeName)
	}

	s.resourceSlices.SetCurrent(nodeName, slices)
	return nil
}

// RemoveNodeResourceSlices removes all node-local ResourceSlices for the Node with the given node name.
// It's a no-op if there aren't any slices to remove.
func (s *Snapshot) RemoveNodeResourceSlices(nodeName string) {
	s.resourceSlices.DeleteCurrent(nodeName)
}

// Commit persists changes done in the topmost layer merging them into the one below it.
func (s *Snapshot) Commit() {
	s.deviceClasses.Commit()
	s.resourceClaims.Commit()
	s.resourceSlices.Commit()
}

// Revert removes the topmost patch layer for all resource types, discarding
// any modifications or deletions recorded there.
func (s *Snapshot) Revert() {
	s.deviceClasses.Revert()
	s.resourceClaims.Revert()
	s.resourceSlices.Revert()
}

// Fork adds a new, empty patch layer to the top of the stack for all
// resource types. Subsequent modifications will be recorded in this
// new layer until Commit() or Revert() are invoked.
func (s *Snapshot) Fork() {
	s.deviceClasses.Fork()
	s.resourceClaims.Fork()
	s.resourceSlices.Fork()
}

// listDeviceClasses retrieves all effective DeviceClasses from the snapshot.
func (s *Snapshot) listDeviceClasses() []*resourceapi.DeviceClass {
	deviceClasses := s.deviceClasses.AsMap()
	deviceClassesList := make([]*resourceapi.DeviceClass, 0, len(deviceClasses))
	for _, class := range deviceClasses {
		deviceClassesList = append(deviceClassesList, class)
	}

	return deviceClassesList
}

// listResourceClaims retrieves all effective ResourceClaims from the snapshot.
func (s *Snapshot) listResourceClaims() []*resourceapi.ResourceClaim {
	claims := s.resourceClaims.AsMap()
	claimsList := make([]*resourceapi.ResourceClaim, 0, len(claims))
	for _, claim := range claims {
		claimsList = append(claimsList, claim)
	}

	return claimsList
}

// configureResourceClaim updates or adds a ResourceClaim in the current patch layer.
// This is typically used internally when a claim's state (like allocation) changes.
func (s *Snapshot) configureResourceClaim(claim *resourceapi.ResourceClaim) {
	claimId := ResourceClaimId{Name: claim.Name, Namespace: claim.Namespace}
	s.resourceClaims.SetCurrent(claimId, claim)
}

// getResourceClaim retrieves a specific ResourceClaim by its ID from the snapshot.
func (s *Snapshot) getResourceClaim(claimId ResourceClaimId) (*resourceapi.ResourceClaim, bool) {
	return s.resourceClaims.FindValue(claimId)
}

// listResourceSlices retrieves all effective ResourceSlices from the snapshot.
func (s *Snapshot) listResourceSlices() []*resourceapi.ResourceSlice {
	resourceSlices := s.resourceSlices.AsMap()
	resourceSlicesList := make([]*resourceapi.ResourceSlice, 0, len(resourceSlices))
	for _, nodeSlices := range resourceSlices {
		resourceSlicesList = append(resourceSlicesList, nodeSlices...)
	}

	return resourceSlicesList
}

// getDeviceClass retrieves a specific DeviceClass by its name from the snapshot.
func (s *Snapshot) getDeviceClass(className string) (*resourceapi.DeviceClass, bool) {
	return s.deviceClasses.FindValue(className)
}

// findPodClaims retrieves all ResourceClaim objects referenced by a given pod.
// If ignoreNotTracked is true, it skips claims not found in the snapshot; otherwise, it returns an error.
func (s *Snapshot) findPodClaims(pod *apiv1.Pod, ignoreNotTracked bool) ([]*resourceapi.ResourceClaim, error) {
	if len(pod.Spec.ResourceClaims) == 0 {
		return nil, nil
	}

	result := make([]*resourceapi.ResourceClaim, len(pod.Spec.ResourceClaims))
	for claimIndex, claimRef := range pod.Spec.ResourceClaims {
		claimName := claimRefToName(pod, claimRef)
		if claimName == "" {
			if !ignoreNotTracked {
				return nil, fmt.Errorf(
					"error while obtaining ResourceClaim %s for pod %s/%s: couldn't determine ResourceClaim name",
					claimRef.Name,
					pod.Namespace,
					pod.Name,
				)
			}

			continue
		}

		claimId := ResourceClaimId{Name: claimName, Namespace: pod.Namespace}
		claim, found := s.resourceClaims.FindValue(claimId)
		if !found {
			if !ignoreNotTracked {
				return nil, fmt.Errorf(
					"error while obtaining ResourceClaim %s for pod %s/%s: couldn't find ResourceClaim in the snapshot",
					claimRef.Name,
					pod.Namespace,
					pod.Name,
				)
			}

			continue
		}

		result[claimIndex] = claim
	}

	return result, nil
}

// ensureClaimWritable returns a resource claim suitable for inplace modifications,
// in case if requested claim is stored in the current patch - the same object
// is returned, otherwise a deep-copy is created. This is required for resource claim
// state changing operations to implement copy-on-write policy for inplace modifications
// when there's no claim tracked on the current layer of the patchset.
func (s *Snapshot) ensureClaimWritable(claim *resourceapi.ResourceClaim) *resourceapi.ResourceClaim {
	claimId := GetClaimId(claim)
	if s.resourceClaims.InCurrentPatch(claimId) {
		return claim
	}

	return claim.DeepCopy()
}

// claimRefToName determines the actual name of a ResourceClaim based on a PodResourceClaim reference.
// It first checks if the name is directly specified in the reference. If not, it looks up the name
// in the pod's status. Returns an empty string if the name cannot be determined.
func claimRefToName(pod *apiv1.Pod, claimRef apiv1.PodResourceClaim) string {
	if claimRef.ResourceClaimName != nil {
		return *claimRef.ResourceClaimName
	}
	for _, claimStatus := range pod.Status.ResourceClaimStatuses {
		if claimStatus.Name == claimRef.Name && claimStatus.ResourceClaimName != nil {
			return *claimStatus.ResourceClaimName
		}
	}
	return ""
}
