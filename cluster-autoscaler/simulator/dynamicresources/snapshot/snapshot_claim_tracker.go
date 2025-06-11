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

	resourceapi "k8s.io/api/resource/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/dynamic-resource-allocation/structured"
)

type snapshotClaimTracker struct {
	snapshot *Snapshot
}

func (ct snapshotClaimTracker) List() ([]*resourceapi.ResourceClaim, error) {
	return ct.snapshot.listResourceClaims(), nil
}

func (ct snapshotClaimTracker) Get(namespace, claimName string) (*resourceapi.ResourceClaim, error) {
	claimId := ResourceClaimId{Name: claimName, Namespace: namespace}
	claim, found := ct.snapshot.getResourceClaim(claimId)
	if !found {
		return nil, fmt.Errorf("claim %s/%s not found", namespace, claimName)
	}
	return claim, nil
}

func (ct snapshotClaimTracker) ListAllAllocatedDevices() (sets.Set[structured.DeviceID], error) {
	result := sets.New[structured.DeviceID]()
	for _, claim := range ct.snapshot.listResourceClaims() {
		result = result.Union(claimAllocatedDevices(claim))
	}
	return result, nil
}

func (ct snapshotClaimTracker) SignalClaimPendingAllocation(claimUid types.UID, allocatedClaim *resourceapi.ResourceClaim) error {
	// The DRA scheduler plugin calls this at the end of the scheduling phase, in Reserve. Then, the allocation is persisted via an API
	// call during the binding phase.
	//
	// In Cluster Autoscaler only the scheduling phase is run, so SignalClaimPendingAllocation() is used to obtain the allocation
	// and persist it in-memory in the snapshot.
	claimId := ResourceClaimId{Name: allocatedClaim.Name, Namespace: allocatedClaim.Namespace}
	claim, found := ct.snapshot.getResourceClaim(claimId)
	if !found {
		return fmt.Errorf("claim %s/%s not found", allocatedClaim.Namespace, allocatedClaim.Name)
	}
	if claim.UID != claimUid {
		return fmt.Errorf("claim %s/%s: snapshot has UID %q, allocation came for UID %q - shouldn't happenn", allocatedClaim.Namespace, allocatedClaim.Name, claim.UID, claimUid)
	}

	ct.snapshot.configureResourceClaim(allocatedClaim)
	return nil
}

func (ct snapshotClaimTracker) ClaimHasPendingAllocation(claimUid types.UID) bool {
	// The DRA scheduler plugin calls this at the beginning of Filter, and fails the filter if true is returned to handle race conditions.
	//
	// In the scheduler implementation, ClaimHasPendingAllocation() starts answering true after SignalClaimPendingAllocation()
	// is called at the end of the scheduling phase, until RemoveClaimPendingAllocation() is called after the allocation API call
	// is made in the asynchronous bind phase.
	//
	// In Cluster Autoscaler only the scheduling phase is run, and SignalClaimPendingAllocation() synchronously persists the allocation
	// in-memory. So the race conditions don't apply, and this should always return false not to block the filter.
	return false
}

func (ct snapshotClaimTracker) RemoveClaimPendingAllocation(claimUid types.UID) (deleted bool) {
	// This method is only called during the Bind phase of scheduler framework, which is never run by CA. We need to implement
	// it to satisfy the interface, but it should never be called.
	panic("snapshotClaimTracker.RemoveClaimPendingAllocation() was called - this should never happen")
}

func (ct snapshotClaimTracker) AssumeClaimAfterAPICall(claim *resourceapi.ResourceClaim) error {
	// This method is only called during the Bind phase of scheduler framework, which is never run by CA. We need to implement
	// it to satisfy the interface, but it should never be called.
	panic("snapshotClaimTracker.AssumeClaimAfterAPICall() was called - this should never happen")
}

func (ct snapshotClaimTracker) AssumedClaimRestore(namespace, claimName string) {
	// This method is only called during the Bind phase of scheduler framework, which is never run by CA. We need to implement
	// it to satisfy the interface, but it should never be called.
	panic("snapshotClaimTracker.AssumedClaimRestore() was called - this should never happen")
}

// claimAllocatedDevices returns ids of all devices allocated in the provided claim.
func claimAllocatedDevices(claim *resourceapi.ResourceClaim) sets.Set[structured.DeviceID] {
	if claim.Status.Allocation == nil {
		return nil
	}
	result := sets.New[structured.DeviceID]()
	for _, device := range claim.Status.Allocation.Devices.Results {
		result.Insert(structured.MakeDeviceID(device.Driver, device.Pool, device.Device))
	}
	return result
}
