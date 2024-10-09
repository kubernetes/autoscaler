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

package dynamicresources

import (
	"fmt"

	resourceapi "k8s.io/api/resource/v1alpha3"
	"k8s.io/apimachinery/pkg/types"
)

type snapshotClaimTracker Snapshot

func (s snapshotClaimTracker) List() ([]*resourceapi.ResourceClaim, error) {
	var result []*resourceapi.ResourceClaim
	for _, claim := range s.resourceClaimsByRef {
		result = append(result, claim)
	}
	return result, nil
}

func (s snapshotClaimTracker) Get(namespace, claimName string) (*resourceapi.ResourceClaim, error) {
	claim, found := s.resourceClaimsByRef[resourceClaimRef{Name: claimName, Namespace: namespace}]
	if !found {
		return nil, fmt.Errorf("claim %s/%s not found", namespace, claimName)
	}
	return claim, nil
}

func (s snapshotClaimTracker) ListAllAllocated() ([]*resourceapi.ResourceClaim, error) {
	var result []*resourceapi.ResourceClaim
	for _, claim := range s.resourceClaimsByRef {
		if ClaimAllocated(claim) {
			result = append(result, claim)
		}
	}
	return result, nil
}

func (s snapshotClaimTracker) SignalClaimPendingAllocation(claimUid types.UID, allocatedClaim *resourceapi.ResourceClaim) error {
	ref := resourceClaimRef{Name: allocatedClaim.Name, Namespace: allocatedClaim.Namespace}
	claim, found := s.resourceClaimsByRef[ref]
	if !found {
		return fmt.Errorf("claim %s/%s not found", allocatedClaim.Namespace, allocatedClaim.Name)
	}
	if claim.UID != claimUid {
		return fmt.Errorf("claim %s/%s: snapshot has UID %q, allocation came for UID %q - shouldn't happenn", allocatedClaim.Namespace, allocatedClaim.Name, claim.UID, claimUid)
	}
	s.resourceClaimsByRef[ref] = allocatedClaim
	return nil
}

func (s snapshotClaimTracker) ClaimHasPendingAllocation(claimUid types.UID) bool {
	return false
}

func (s snapshotClaimTracker) RemoveClaimPendingAllocation(claimUid types.UID) (deleted bool) {
	//TODO implement me
	panic("implement me")
}

func (s snapshotClaimTracker) AssumeClaimAfterAPICall(claim *resourceapi.ResourceClaim) error {
	//TODO implement me
	panic("implement me")
}

func (s snapshotClaimTracker) AssumedClaimRestore(namespace, claimName string) {
	//TODO implement me
	panic("implement me")
}

func (s snapshotClaimTracker) GetOriginal(namespace, claimName string) (*resourceapi.ResourceClaim, error) {
	//TODO implement me
	panic("implement me")
}
