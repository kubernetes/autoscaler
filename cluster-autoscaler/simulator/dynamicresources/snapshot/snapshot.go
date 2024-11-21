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

	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	drautils "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/utils"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
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
	resourceClaimsById         map[ResourceClaimId]*resourceapi.ResourceClaim
	resourceSlicesByNodeName   map[string][]*resourceapi.ResourceSlice
	nonNodeLocalResourceSlices []*resourceapi.ResourceSlice
	deviceClasses              map[string]*resourceapi.DeviceClass
}

// NewSnapshot returns a Snapshot created from the provided data.
func NewSnapshot(claims map[ResourceClaimId]*resourceapi.ResourceClaim, nodeLocalSlices map[string][]*resourceapi.ResourceSlice, globalSlices []*resourceapi.ResourceSlice, deviceClasses map[string]*resourceapi.DeviceClass) Snapshot {
	if claims == nil {
		claims = map[ResourceClaimId]*resourceapi.ResourceClaim{}
	}
	if nodeLocalSlices == nil {
		nodeLocalSlices = map[string][]*resourceapi.ResourceSlice{}
	}
	if deviceClasses == nil {
		deviceClasses = map[string]*resourceapi.DeviceClass{}
	}
	return Snapshot{
		resourceClaimsById:         claims,
		resourceSlicesByNodeName:   nodeLocalSlices,
		nonNodeLocalResourceSlices: globalSlices,
		deviceClasses:              deviceClasses,
	}
}

// ResourceClaims exposes the Snapshot as schedulerframework.ResourceClaimTracker, in order to interact with
// the scheduler framework.
func (s Snapshot) ResourceClaims() schedulerframework.ResourceClaimTracker {
	return snapshotClaimTracker(s)
}

// ResourceSlices exposes the Snapshot as schedulerframework.ResourceSliceLister, in order to interact with
// the scheduler framework.
func (s Snapshot) ResourceSlices() schedulerframework.ResourceSliceLister {
	return snapshotSliceLister(s)
}

// DeviceClasses exposes the Snapshot as schedulerframework.DeviceClassLister, in order to interact with
// the scheduler framework.
func (s Snapshot) DeviceClasses() schedulerframework.DeviceClassLister {
	return snapshotClassLister(s)
}

// WrapSchedulerNodeInfo wraps the provided *schedulerframework.NodeInfo into an internal *framework.NodeInfo, adding
// dra information. Node-local ResourceSlices are added to the NodeInfo, and all ResourceClaims referenced by each Pod
// are added to each PodInfo. Returns an error if any of the Pods is missing a ResourceClaim.
func (s Snapshot) WrapSchedulerNodeInfo(schedNodeInfo *schedulerframework.NodeInfo) (*framework.NodeInfo, error) {
	podExtraInfos := map[types.UID]framework.PodExtraInfo{}
	for _, pod := range schedNodeInfo.Pods {
		podClaims, err := s.PodClaims(pod.Pod)
		if err != nil {
			return nil, err
		}
		if len(podClaims) > 0 {
			podExtraInfos[pod.Pod.UID] = framework.PodExtraInfo{NeededResourceClaims: podClaims}
		}
	}
	nodeSlices, _ := s.NodeResourceSlices(schedNodeInfo.Node().Name)
	return framework.WrapSchedulerNodeInfo(schedNodeInfo, nodeSlices, podExtraInfos), nil
}

// Clone returns a copy of this Snapshot that can be independently modified without affecting this Snapshot.
// The only mutable objects in the Snapshot are ResourceClaims, so they are deep-copied. The rest is only a
// shallow copy.
func (s Snapshot) Clone() Snapshot {
	result := Snapshot{
		resourceClaimsById:       map[ResourceClaimId]*resourceapi.ResourceClaim{},
		resourceSlicesByNodeName: map[string][]*resourceapi.ResourceSlice{},
		deviceClasses:            map[string]*resourceapi.DeviceClass{},
	}
	// The claims are mutable, they have to be deep-copied.
	for id, claim := range s.resourceClaimsById {
		result.resourceClaimsById[id] = claim.DeepCopy()
	}
	// The rest of the objects aren't mutable, so a shallow copy should be enough.
	for nodeName, slices := range s.resourceSlicesByNodeName {
		for _, slice := range slices {
			result.resourceSlicesByNodeName[nodeName] = append(result.resourceSlicesByNodeName[nodeName], slice)
		}
	}
	for _, slice := range s.nonNodeLocalResourceSlices {
		result.nonNodeLocalResourceSlices = append(result.nonNodeLocalResourceSlices, slice)
	}
	for className, class := range s.deviceClasses {
		result.deviceClasses[className] = class
	}
	return result
}

// AddClaims adds additional ResourceClaims to the Snapshot. It can be used e.g. if we need to duplicate a Pod that
// owns ResourceClaims. Returns an error if any of the claims is already tracked in the snapshot.
func (s Snapshot) AddClaims(newClaims []*resourceapi.ResourceClaim) error {
	for _, claim := range newClaims {
		if _, found := s.resourceClaimsById[GetClaimId(claim)]; found {
			return fmt.Errorf("claim %s/%s already tracked in the snapshot", claim.Namespace, claim.Name)
		}
	}
	for _, claim := range newClaims {
		s.resourceClaimsById[GetClaimId(claim)] = claim
	}
	return nil
}

// PodClaims returns ResourceClaims objects for all claims referenced by the Pod. If any of the referenced claims
// isn't tracked in the Snapshot, an error is returned.
func (s Snapshot) PodClaims(pod *apiv1.Pod) ([]*resourceapi.ResourceClaim, error) {
	var result []*resourceapi.ResourceClaim

	for _, claimRef := range pod.Spec.ResourceClaims {
		claim, err := s.claimForPod(pod, claimRef)
		if err != nil {
			return nil, fmt.Errorf("error while obtaining ResourceClaim %s for pod %s/%s: %v", claimRef.Name, pod.Namespace, pod.Name, err)
		}
		result = append(result, claim)
	}

	return result, nil
}

// RemovePodOwnedClaims iterates over all the claims referenced by the Pod, and removes the ones owned by the Pod from the Snapshot.
// Claims referenced by the Pod but not owned by it are not removed, but the Pod's reservation is removed from them.
// This method removes all relevant claims that are in the snapshot, and doesn't error out if any of the claims are missing.
func (s Snapshot) RemovePodOwnedClaims(pod *apiv1.Pod) {
	for _, podClaimRef := range pod.Spec.ResourceClaims {
		claimName := claimRefToName(pod, podClaimRef)
		if claimName == "" {
			// This most likely means that the Claim hasn't actually been created. Nothing to remove/modify, so continue to the next claim.
			continue
		}
		claimId := ResourceClaimId{Name: claimName, Namespace: pod.Namespace}
		claim, found := s.resourceClaimsById[claimId]
		if !found {
			// The claim isn't tracked in the snapshot for some reason. Nothing to remove/modify, so continue to the next claim.
			continue
		}
		if ownerName, ownerUid := drautils.ClaimOwningPod(claim); ownerName == pod.Name && ownerUid == pod.UID {
			delete(s.resourceClaimsById, claimId)
		} else {
			drautils.ClearPodReservationInPlace(claim, pod)
		}
	}
}

// ReservePodClaims adds a reservation for the provided Pod to all the claims it references. If any of the referenced
// claims isn't tracked in the Snapshot, or if any of the claims are already at maximum reservation count, an error is
// returned.
func (s Snapshot) ReservePodClaims(pod *apiv1.Pod) error {
	claims, err := s.PodClaims(pod)
	if err != nil {
		return err
	}
	for _, claim := range claims {
		if drautils.ClaimFullyReserved(claim) && !drautils.ClaimReservedForPod(claim, pod) {
			return fmt.Errorf("claim %s/%s already has max number of reservations set, can't add more", claim.Namespace, claim.Name)
		}
	}
	for _, claim := range claims {
		drautils.AddPodReservationInPlace(claim, pod)
	}
	return nil
}

// UnreservePodClaims removes reservations for the provided Pod from all the claims it references. If any of the referenced
// claims isn't tracked in the Snapshot, an error is returned. If a claim is owned by the pod, or if the claim has no more reservations,
// its allocation is cleared.
func (s Snapshot) UnreservePodClaims(pod *apiv1.Pod) error {
	claims, err := s.PodClaims(pod)
	if err != nil {
		return err
	}
	for _, claim := range claims {
		ownerPodName, ownerPodUid := drautils.ClaimOwningPod(claim)
		podOwnedClaim := ownerPodName == pod.Name && ownerPodUid == ownerPodUid

		drautils.ClearPodReservationInPlace(claim, pod)
		if podOwnedClaim || !drautils.ClaimInUse(claim) {
			drautils.DeallocateClaimInPlace(claim)
		}
	}
	return nil
}

// NodeResourceSlices returns all node-local ResourceSlices for the given Node.
func (s Snapshot) NodeResourceSlices(nodeName string) ([]*resourceapi.ResourceSlice, bool) {
	slices, found := s.resourceSlicesByNodeName[nodeName]
	return slices, found
}

// AddNodeResourceSlices adds additional node-local ResourceSlices to the Snapshot. This should be used whenever a Node with
// node-local ResourceSlices is duplicated in the cluster snapshot.
func (s Snapshot) AddNodeResourceSlices(nodeName string, slices []*resourceapi.ResourceSlice) error {
	if _, alreadyInSnapshot := s.resourceSlicesByNodeName[nodeName]; alreadyInSnapshot {
		return fmt.Errorf("node %q ResourceSlices already present", nodeName)
	}
	s.resourceSlicesByNodeName[nodeName] = slices
	return nil
}

// RemoveNodeResourceSlices removes all node-local ResourceSlices for the Node with the given nodeName.
// It's a no-op if there aren't any slices to remove.
func (s Snapshot) RemoveNodeResourceSlices(nodeName string) {
	delete(s.resourceSlicesByNodeName, nodeName)
}

func (s Snapshot) claimForPod(pod *apiv1.Pod, claimRef apiv1.PodResourceClaim) (*resourceapi.ResourceClaim, error) {
	claimName := claimRefToName(pod, claimRef)
	if claimName == "" {
		return nil, fmt.Errorf("couldn't determine ResourceClaim name")
	}

	claim, found := s.resourceClaimsById[ResourceClaimId{Name: claimName, Namespace: pod.Namespace}]
	if !found {
		return nil, fmt.Errorf("couldn't find ResourceClaim %q", claimName)
	}

	return claim, nil
}

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
