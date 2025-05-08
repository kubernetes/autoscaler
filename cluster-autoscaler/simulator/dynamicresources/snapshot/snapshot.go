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
	"k8s.io/apimachinery/pkg/util/sets"
	drautils "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/utils"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/dynamic-resource-allocation/structured"
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// ResourceClaimId is a unique identifier for a ResourceClaim.
type ResourceClaimId struct {
	Name      string
	Namespace string
}

func (id ResourceClaimId) String() string {
	return fmt.Sprintf("%s/%s", id.Namespace, id.Name)
}

// GetClaimId returns the unique identifier for a ResourceClaim.
func GetClaimId(claim *resourceapi.ResourceClaim) ResourceClaimId {
	return ResourceClaimId{Name: claim.Name, Namespace: claim.Namespace}
}

/*

LAYER 1
A -> [1,2,3]
LAYER 2
A -> [4]

*/

// BasicSnapshot contains a snapshot of all DRA objects taken at a ~single point in time. The BasicSnapshot should be
// obtained via the Provider. Then, it can be modified using the exposed methods, to simulate scheduling actions
// in the cluster.
type BasicSnapshot struct {
	resourceClaimsById         map[ResourceClaimId]*resourceapi.ResourceClaim
	resourceSlicesByNodeName   map[string][]*resourceapi.ResourceSlice
	nonNodeLocalResourceSlices []*resourceapi.ResourceSlice
	deviceClasses              map[string]*resourceapi.DeviceClass
}

// NewBasicSnapshot returns a Snapshot created from the provided data.
func NewBasicSnapshot(claims map[ResourceClaimId]*resourceapi.ResourceClaim, nodeLocalSlices map[string][]*resourceapi.ResourceSlice, globalSlices []*resourceapi.ResourceSlice, deviceClasses map[string]*resourceapi.DeviceClass) BasicSnapshot {
	if claims == nil {
		claims = map[ResourceClaimId]*resourceapi.ResourceClaim{}
	}
	if nodeLocalSlices == nil {
		nodeLocalSlices = map[string][]*resourceapi.ResourceSlice{}
	}
	if deviceClasses == nil {
		deviceClasses = map[string]*resourceapi.DeviceClass{}
	}
	return BasicSnapshot{
		resourceClaimsById:         claims,
		resourceSlicesByNodeName:   nodeLocalSlices,
		nonNodeLocalResourceSlices: globalSlices,
		deviceClasses:              deviceClasses,
	}
}

// NewBasicSnapshot returns a Snapshot created from the provided data.
func NewSnapshot(claims map[ResourceClaimId]*resourceapi.ResourceClaim, nodeLocalSlices map[string][]*resourceapi.ResourceSlice, globalSlices []*resourceapi.ResourceSlice, deviceClasses map[string]*resourceapi.DeviceClass) *Snapshot {
	basicSnapshot := NewBasicSnapshot(claims, nodeLocalSlices, globalSlices, deviceClasses)
	return &Snapshot{
		snapshot:   basicSnapshot,
		patches:    make([]*snapshotPatch, 0),
		forkPoints: make([]int, 0),
	}
}

// NewBasicSnapshot returns a Snapshot created from the provided data.
func NewEmptySnapshot() *Snapshot {
	basicSnapshot := NewBasicSnapshot(nil, nil, nil, nil)
	return &Snapshot{
		snapshot:   basicSnapshot,
		patches:    make([]*snapshotPatch, 0),
		forkPoints: make([]int, 0),
	}
}

// WrapSchedulerNodeInfo wraps the provided *schedulerframework.NodeInfo into an internal *framework.NodeInfo, adding
// dra information. Node-local ResourceSlices are added to the NodeInfo, and all ResourceClaims referenced by each Pod
// are added to each PodInfo. Returns an error if any of the Pods is missing a ResourceClaim.
func (s BasicSnapshot) WrapSchedulerNodeInfo(schedNodeInfo *schedulerframework.NodeInfo) (*framework.NodeInfo, error) {
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
func (s BasicSnapshot) Clone() BasicSnapshot {
	result := BasicSnapshot{
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
		result.resourceSlicesByNodeName[nodeName] = append(result.resourceSlicesByNodeName[nodeName], slices...)
	}
	result.nonNodeLocalResourceSlices = append(result.nonNodeLocalResourceSlices, s.nonNodeLocalResourceSlices...)
	for className, class := range s.deviceClasses {
		result.deviceClasses[className] = class
	}

	return result
}

// AddClaims adds additional ResourceClaims to the Snapshot. It can be used e.g. if we need to duplicate a Pod that
// owns ResourceClaims. Returns an error if any of the claims is already tracked in the snapshot.
func (s BasicSnapshot) AddClaims(newClaims []*resourceapi.ResourceClaim) error {
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

func (s BasicSnapshot) RemoveClaims(claims []*resourceapi.ResourceClaim) {
	for _, claim := range claims {
		claimId := GetClaimId(claim)
		delete(s.resourceClaimsById, claimId)
	}
}

func (s BasicSnapshot) ConfigureClaim(claim *resourceapi.ResourceClaim, newState *resourceapi.ResourceClaim) error {
	ref := GetClaimId(claim)
	if newStateRef := GetClaimId(newState); ref != newStateRef {
		return fmt.Errorf("unable to configure claim %s with %s", ref, newStateRef)
	}

	if claim.UID != newState.UID {
		return fmt.Errorf("claim %s is tracked with UUID: %s, unable to change it to %s", ref, claim.UID, newState.UID)
	}

	_, tracked := s.resourceClaimsById[ref]
	if !tracked {
		return fmt.Errorf("claim %s is not tracked in a snapshot", ref)
	}

	s.resourceClaimsById[ref] = newState
	return nil
}

// PodClaims returns ResourceClaims objects for all claims referenced by the Pod. If any of the referenced claims
// isn't tracked in the Snapshot, an error is returned.
func (s BasicSnapshot) PodClaims(pod *apiv1.Pod) ([]*resourceapi.ResourceClaim, error) {
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

func (s BasicSnapshot) podClaimsNoError(pod *apiv1.Pod) map[ResourceClaimId]*resourceapi.ResourceClaim {
	result := make(map[ResourceClaimId]*resourceapi.ResourceClaim)
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

		result[claimId] = claim
	}

	return result
}

// RemovePodOwnedClaims iterates over all the claims referenced by the Pod, and removes the ones owned by the Pod from the Snapshot.
// Claims referenced by the Pod but not owned by it are not removed, but the Pod's reservation is removed from them.
// This method removes all relevant claims that are in the snapshot, and doesn't error out if any of the claims are missing.
func (s BasicSnapshot) RemovePodOwnedClaims(pod *apiv1.Pod) {
	podClaims := s.podClaimsNoError(pod)
	for claimId, claim := range podClaims {
		if drautils.PodOwnsClaim(pod, claim) {
			delete(s.resourceClaimsById, claimId)
		} else {
			drautils.ClearPodReservationInPlace(claim, pod)
		}
	}
}

// ReservePodClaims adds a reservation for the provided Pod to all the claims it references. If any of the referenced
// claims isn't tracked in the Snapshot, or if any of the claims are already at maximum reservation count, an error is
// returned.
func (s BasicSnapshot) ReservePodClaims(pod *apiv1.Pod) error {
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
func (s BasicSnapshot) UnreservePodClaims(pod *apiv1.Pod) error {
	claims, err := s.PodClaims(pod)
	if err != nil {
		return err
	}
	for _, claim := range claims {
		podOwnedClaim := drautils.PodOwnsClaim(pod, claim)

		drautils.ClearPodReservationInPlace(claim, pod)
		if podOwnedClaim || !drautils.ClaimInUse(claim) {
			drautils.DeallocateClaimInPlace(claim)
		}
	}
	return nil
}

// NodeResourceSlices returns all node-local ResourceSlices for the given Node.
func (s BasicSnapshot) NodeResourceSlices(nodeName string) ([]*resourceapi.ResourceSlice, bool) {
	slices, found := s.resourceSlicesByNodeName[nodeName]
	return slices, found
}

// AddNodeResourceSlices adds additional node-local ResourceSlices to the Snapshot. This should be used whenever a Node with
// node-local ResourceSlices is duplicated in the cluster snapshot.
func (s BasicSnapshot) AddNodeResourceSlices(nodeName string, slices []*resourceapi.ResourceSlice) error {
	if _, alreadyInSnapshot := s.resourceSlicesByNodeName[nodeName]; alreadyInSnapshot {
		return fmt.Errorf("node %q ResourceSlices already present", nodeName)
	}
	s.resourceSlicesByNodeName[nodeName] = slices
	return nil
}

// RemoveNodeResourceSlices removes all node-local ResourceSlices for the Node with the given nodeName.
// It's a no-op if there aren't any slices to remove.
func (s BasicSnapshot) RemoveNodeResourceSlices(nodeName string) {
	delete(s.resourceSlicesByNodeName, nodeName)
}

func (s BasicSnapshot) ListDeviceClasses() ([]*resourceapi.DeviceClass, error) {
	var result []*resourceapi.DeviceClass
	for _, class := range s.deviceClasses {
		result = append(result, class)
	}
	return result, nil
}

func (s BasicSnapshot) GetDeviceClass(className string) (*resourceapi.DeviceClass, error) {
	class, found := s.deviceClasses[className]
	if !found {
		return nil, fmt.Errorf("DeviceClass %q not found", className)
	}
	return class, nil
}

func (s BasicSnapshot) ListResourceSlices() ([]*resourceapi.ResourceSlice, error) {
	var result []*resourceapi.ResourceSlice
	for _, slices := range s.resourceSlicesByNodeName {
		result = append(result, slices...)
	}
	result = append(result, s.nonNodeLocalResourceSlices...)
	return result, nil
}

func (s BasicSnapshot) ListResourceClaims() ([]*resourceapi.ResourceClaim, error) {
	var result []*resourceapi.ResourceClaim
	for _, claim := range s.resourceClaimsById {
		result = append(result, claim)
	}
	return result, nil
}

func (s BasicSnapshot) GetResourceClaim(namespace, claimName string) (*resourceapi.ResourceClaim, error) {
	claim, found := s.resourceClaimsById[ResourceClaimId{Name: claimName, Namespace: namespace}]
	if !found {
		return nil, fmt.Errorf("claim %s/%s not found", namespace, claimName)
	}
	return claim, nil
}

func (s BasicSnapshot) ListAllAllocatedDevices() (sets.Set[structured.DeviceID], error) {
	result := sets.New[structured.DeviceID]()
	for _, claim := range s.resourceClaimsById {
		result = result.Union(claimAllocatedDevices(claim))
	}
	return result, nil
}

type snapshotPatch struct {
	Name   string
	Apply  func(*BasicSnapshot) error
	Revert func(*BasicSnapshot) error
}

func (p *snapshotPatch) String() string {
	return p.Name
}

type Snapshot struct {
	patches    []*snapshotPatch
	forkPoints []int
	snapshot   BasicSnapshot
}

// ResourceClaims exposes the Snapshot as schedulerframework.ResourceClaimTracker, in order to interact with
// the scheduler framework.
func (ds *Snapshot) ResourceClaimTracker() schedulerframework.ResourceClaimTracker {
	return snapshotClaimTracker{ds}
}

// ResourceSlices exposes the Snapshot as schedulerframework.ResourceSliceLister, in order to interact with
// the scheduler framework.
func (ds *Snapshot) ResourceSliceLister() schedulerframework.ResourceSliceLister {
	return snapshotSliceLister{ds}
}

// DeviceClasses exposes the Snapshot as schedulerframework.DeviceClassLister, in order to interact with
// the scheduler framework.
func (ds *Snapshot) DeviceClassLister() schedulerframework.DeviceClassLister {
	return snapshotClassLister{ds}
}

// WrapSchedulerNodeInfo wraps the provided *schedulerframework.NodeInfo into an internal *framework.NodeInfo, adding
// dra information. Node-local ResourceSlices are added to the NodeInfo, and all ResourceClaims referenced by each Pod
// are added to each PodInfo. Returns an error if any of the Pods is missing a ResourceClaim.
func (ds *Snapshot) WrapSchedulerNodeInfo(schedNodeInfo *schedulerframework.NodeInfo) (*framework.NodeInfo, error) {
	return ds.snapshot.WrapSchedulerNodeInfo(schedNodeInfo)
}

// Clone returns a copy of this Snapshot that can be independently moxdified without affecting this Snapshot.
// The only mutable objects in the Snapshot are ResourceClaims, so they are deep-copied. The rest is only a
// shallow copy.
func (ds *Snapshot) Clone() *Snapshot {
	patches := make([]*snapshotPatch, len(ds.patches))
	copy(patches, ds.patches)

	return &Snapshot{
		snapshot: ds.snapshot.Clone(),
		patches:  patches, // shallow copy as patches are immutable
	}
}

// AddClaims adds additional ResourceClaims to the Snapshot. It can be used e.g. if we need to duplicate a Pod that
// owns ResourceClaims. Returns an error if any of the claims is already tracked in the snapshot.
func (ds *Snapshot) AddClaims(claims []*resourceapi.ResourceClaim) error {
	return ds.applyPatch(
		"Snapshot.AddClaims",
		func(s *BasicSnapshot) error { return s.AddClaims(claims) },
		func(s *BasicSnapshot) error {
			s.RemoveClaims(claims)
			return nil
		},
	)
}

// RemoveClaims: TODO(mfuhol)
func (ds *Snapshot) RemoveClaims(claims []*resourceapi.ResourceClaim) {
	err := ds.applyPatch(
		"Snapshot.RemoveClaims",
		func(s *BasicSnapshot) error {
			s.RemoveClaims(claims)
			return nil
		},
		func(s *BasicSnapshot) error { return s.AddClaims(claims) },
	)

	if err != nil {
		klog.Errorf("Snapshot.RemoveClaims ignored error: %v", err)
	}
}

// ConfigureClaim: TODO(mfuhol)
func (ds *Snapshot) ConfigureClaim(claim *resourceapi.ResourceClaim, newState *resourceapi.ResourceClaim) error {
	return ds.applyPatch(
		"Snapshot.ConfigureClaim",
		func(s *BasicSnapshot) error { return s.ConfigureClaim(claim, newState) },
		func(s *BasicSnapshot) error { return s.ConfigureClaim(newState, claim) },
	)
}

// RemovePodOwnedClaims iterates over all the claims referenced by the Pod, and removes the ones owned by the Pod from the Snapshot.
// Claims referenced by the Pod but not owned by it are not removed, but the Pod's reservation is removed from them.
// This method removes all relevant claims that are in the snapshot, and doesn't error out if any of the claims are missing.
func (ds *Snapshot) RemovePodOwnedClaims(pod *apiv1.Pod) {
	podClaims := ds.snapshot.podClaimsNoError(pod)
	podClaimsSlice := make([]*resourceapi.ResourceClaim, len(podClaims))
	for _, claim := range podClaims {
		podClaimsSlice = append(podClaimsSlice, claim)
	}

	err := ds.applyPatch(
		"Snapshot.RemovePodOwnedClaims",
		func(s *BasicSnapshot) error {
			s.RemovePodOwnedClaims(pod)
			return nil
		},
		func(s *BasicSnapshot) error { return s.AddClaims(podClaimsSlice) },
	)

	if err != nil {
		klog.Errorf("Snapshot.RemovePodOwnedClaims ignored error: %v", err)
	}
}

// ReservePodClaims adds a reservation for the provided Pod to all the claims it references. If any of the referenced
// claims isn't tracked in the Snapshot, or if any of the claims are already at maximum reservation count, an error is
// returned.
func (ds *Snapshot) ReservePodClaims(pod *apiv1.Pod) error {
	return ds.applyPatch(
		"Snapshot.ReservePodClaims",
		func(s *BasicSnapshot) error { return s.ReservePodClaims(pod) },
		func(s *BasicSnapshot) error { return s.UnreservePodClaims(pod) },
	)
}

// UnreservePodClaims removes reservations for the provided Pod from all the claims it references. If any of the referenced
// claims isn't tracked in the Snapshot, an error is returned. If a claim is owned by the pod, or if the claim has no more reservations,
// its allocation is cleared.
func (ds *Snapshot) UnreservePodClaims(pod *apiv1.Pod) error {
	return ds.applyPatch(
		"Snapshot.UnreservePodClaims",
		func(s *BasicSnapshot) error { return s.UnreservePodClaims(pod) },
		func(s *BasicSnapshot) error { return s.ReservePodClaims(pod) },
	)
}

// AddNodeResourceSlices adds additional node-local ResourceSlices to the Snapshot. This should be used whenever a Node with
// node-local ResourceSlices is duplicated in the cluster snapshot.
func (ds *Snapshot) AddNodeResourceSlices(nodeName string, slices []*resourceapi.ResourceSlice) error {
	return ds.applyPatch(
		"AddNodeResourceSlices",
		func(s *BasicSnapshot) error { return s.AddNodeResourceSlices(nodeName, slices) },
		func(s *BasicSnapshot) error {
			s.RemoveNodeResourceSlices(nodeName)
			return nil
		},
	)
}

// RemoveNodeResourceSlices removes all node-local ResourceSlices for the Node with the given nodeName.
// It's a no-op if there aren't any slices to remove.
func (ds *Snapshot) RemoveNodeResourceSlices(nodeName string) {
	resourceSlices, found := ds.snapshot.NodeResourceSlices(nodeName)
	if !found {
		return
	}

	err := ds.applyPatch(
		"Snapshot.RemoveNodeResourceSlices",
		func(s *BasicSnapshot) error {
			s.RemoveNodeResourceSlices(nodeName)
			return nil
		},
		func(s *BasicSnapshot) error { return s.AddNodeResourceSlices(nodeName, resourceSlices) },
	)

	if err != nil {
		klog.Errorf("Snapshot.RemoveNodeResourceSlices ignored error: %v", err)
	}
}

// PodClaims returns ResourceClaims objects for all claims referenced by the Pod. If any of the referenced claims
// isn't tracked in the Snapshot, an error is returned.
func (ds *Snapshot) PodClaims(pod *apiv1.Pod) ([]*resourceapi.ResourceClaim, error) {
	return ds.snapshot.PodClaims(pod)
}

// NodeResourceSlices returns all node-local ResourceSlices for the given Node.
func (ds *Snapshot) NodeResourceSlices(nodeName string) ([]*resourceapi.ResourceSlice, bool) {
	return ds.snapshot.NodeResourceSlices(nodeName)
}

func (ds *Snapshot) ListDeviceClasses() ([]*resourceapi.DeviceClass, error) {
	return ds.snapshot.ListDeviceClasses()
}

func (ds *Snapshot) GetDeviceClass(className string) (*resourceapi.DeviceClass, error) {
	return ds.snapshot.GetDeviceClass(className)
}

func (ds *Snapshot) ListResourceSlices() ([]*resourceapi.ResourceSlice, error) {
	return ds.snapshot.ListResourceSlices()
}

func (ds *Snapshot) ListResourceClaims() ([]*resourceapi.ResourceClaim, error) {
	return ds.snapshot.ListResourceClaims()
}

func (ds *Snapshot) GetResourceClaim(namespace, claimName string) (*resourceapi.ResourceClaim, error) {
	return ds.snapshot.GetResourceClaim(namespace, claimName)
}

func (ds *Snapshot) ListAllAllocatedDevices() (sets.Set[structured.DeviceID], error) {
	return ds.snapshot.ListAllAllocatedDevices()
}

func (ds *Snapshot) Commit() error {
	ds.forkPoints = ds.forkPoints[:len(ds.forkPoints)-1]
	return nil
}

func (ds *Snapshot) Fork() {
	ds.forkPoints = append(ds.forkPoints, ds.patchNumber())
}

func (ds *Snapshot) Revert() error {
	err := ds.revertToPatch(ds.forkPoints[len(ds.forkPoints)-1])
	if err != nil {
		return err
	}

	ds.forkPoints = ds.forkPoints[:len(ds.forkPoints)-1]
	return nil
}

func (ds *Snapshot) patchNumber() int {
	return len(ds.patches)
}

func (ds *Snapshot) revertToPatch(patchNumber int) error {
	currentPatchNumber := ds.patchNumber()
	if patchNumber > currentPatchNumber {
		return fmt.Errorf("unable to rollback to patch number above current: %d > %d (current)", patchNumber, currentPatchNumber)
	}

	return ds.revertNPatches(currentPatchNumber - patchNumber)
}

func (ds *Snapshot) revertNPatches(n int) error {
	for i := 0; i < n; i++ {
		err := ds.revertPatch()
		if err != nil {
			return err
		}
	}

	return nil
}

func (ds *Snapshot) lastPatch() *snapshotPatch {
	return ds.patches[ds.patchNumber()-1]
}

func (ds *Snapshot) revertPatch() error {
	patch := ds.lastPatch()
	err := patch.Revert(&ds.snapshot)
	if err != nil {
		return fmt.Errorf("failed to revert patch to the snapshot: %w", err)
	}

	ds.patches = ds.patches[:len(ds.patches)-1]
	return nil
}

func (ds *Snapshot) applyPatch(name string, do, undo func(*BasicSnapshot) error) error {
	patch := &snapshotPatch{Name: name, Apply: do, Revert: undo}
	err := patch.Apply(&ds.snapshot)
	if err != nil {
		return fmt.Errorf("failed to update the snapshot: %w", err)
	}

	ds.patches = append(ds.patches, patch)

	return nil
}

func (s BasicSnapshot) claimForPod(pod *apiv1.Pod, claimRef apiv1.PodResourceClaim) (*resourceapi.ResourceClaim, error) {
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
