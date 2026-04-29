/*
Copyright The Kubernetes Authors.

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
	"errors"

	v1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	schedulingapi "k8s.io/api/scheduling/v1alpha2"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/dynamic-resource-allocation/structured"
	fwk "k8s.io/kube-scheduler/framework"
)

var (
	errorNotFound = errors.New("object not found: dra manager is disabled")
)

// NoOpDRAManager is a No-Op implementation of the SharedDRAManager
// interface, it's intended to be used only when DRA is disabled
type NoOpDRAManager struct{}

type noOpResourceClaimTracker NoOpDRAManager
type noOpResourceSliceLister NoOpDRAManager
type noOpDeviceClassLister NoOpDRAManager
type noOpDeviceClassResolver NoOpDRAManager
type noOpPodGroupLister NoOpDRAManager

// NewNoOpDRAManager returns a no-op SharedDRAManager.
func NewNoOpDRAManager() fwk.SharedDRAManager {
	return NoOpDRAManager{}
}

// ResourceClaims returns a no-op ResourceClaimTracker.
func (dm NoOpDRAManager) ResourceClaims() fwk.ResourceClaimTracker {
	return (noOpResourceClaimTracker)(dm)
}

// ResourceSlices returns a no-op ResourceSliceLister.
func (dm NoOpDRAManager) ResourceSlices() fwk.ResourceSliceLister {
	return (noOpResourceSliceLister)(dm)
}

// DeviceClasses returns a no-op DeviceClassLister.
func (dm NoOpDRAManager) DeviceClasses() fwk.DeviceClassLister {
	return (noOpDeviceClassLister)(dm)
}

// DeviceClassResolver returns a no-op DeviceClassResolver.
func (dm NoOpDRAManager) DeviceClassResolver() fwk.DeviceClassResolver {
	return (noOpDeviceClassResolver)(dm)
}

// PodGroups returns a no-op PodGroupLister.
func (dm NoOpDRAManager) PodGroups() fwk.PodGroupLister {
	return (noOpPodGroupLister)(dm)
}

// -------------------------------------
// Resource Claim Tracker Implementation
// -------------------------------------

// List is a no-op implementation of the List method on fwk.ResourceClaimTracker.
func (ct noOpResourceClaimTracker) List() ([]*resourceapi.ResourceClaim, error) {
	return nil, nil
}

// Get is a no-op implementation of the Get method on fwk.ResourceClaimTracker.
func (ct noOpResourceClaimTracker) Get(namespace, claimName string) (*resourceapi.ResourceClaim, error) {
	return nil, errorNotFound
}

// ListAllAllocatedDevices is a no-op implementation of the ListAllAllocatedDevices method on fwk.ResourceClaimTracker.
func (ct noOpResourceClaimTracker) ListAllAllocatedDevices() (sets.Set[structured.DeviceID], error) {
	return nil, nil
}

// GatherAllocatedState is a no-op implementation of the GatherAllocatedState method on fwk.ResourceClaimTracker.
func (ct noOpResourceClaimTracker) GatherAllocatedState() (*structured.AllocatedState, error) {
	return nil, nil
}

// SignalClaimPendingAllocation is a no-op implementation of the SignalClaimPendingAllocation method on fwk.ResourceClaimTracker.
func (ct noOpResourceClaimTracker) SignalClaimPendingAllocation(claimUID types.UID, allocatedClaim *resourceapi.ResourceClaim) error {
	return nil
}

// GetPendingAllocation is a no-op implementation of the GetPendingAllocation method on fwk.ResourceClaimTracker.
func (ct noOpResourceClaimTracker) GetPendingAllocation(claimUID types.UID) *resourceapi.AllocationResult {
	return nil
}

// MaybeRemoveClaimPendingAllocation is a no-op implementation of the MaybeRemoveClaimPendingAllocation method on fwk.ResourceClaimTracker.
func (ct noOpResourceClaimTracker) MaybeRemoveClaimPendingAllocation(claimUID types.UID, forceRemove bool) (deleted bool) {
	return false
}

// AssumeClaimAfterAPICall is a no-op implementation of the AssumeClaimAfterAPICall method on fwk.ResourceClaimTracker.
func (ct noOpResourceClaimTracker) AssumeClaimAfterAPICall(claim *resourceapi.ResourceClaim) error {
	return nil
}

// AssumedClaimRestore is a no-op implementation of the AssumedClaimRestore method on fwk.ResourceClaimTracker.
func (ct noOpResourceClaimTracker) AssumedClaimRestore(namespace, claimName string) {

}

// ListWithDeviceTaintRules is a no-op implementation of the ListWithDeviceTaintRules method on fwk.ResourceSliceLister.
func (sl noOpResourceSliceLister) ListWithDeviceTaintRules() ([]*resourceapi.ResourceSlice, error) {
	return nil, nil
}

// ----------------------------------
// Device Class Lister Implementation
// ----------------------------------

// List is a no-op implementation of the List method on fwk.DeviceClassLister.
func (dcl noOpDeviceClassLister) List() ([]*resourceapi.DeviceClass, error) {
	return nil, nil
}

// Get is a no-op implementation of the Get method on fwk.DeviceClassLister.
func (dcl noOpDeviceClassLister) Get(className string) (*resourceapi.DeviceClass, error) {
	return nil, errorNotFound
}

// ------------------------------------
// Device Class Resolver Implementation
// ------------------------------------

// GetDeviceClass is a no-op implementation of the GetDeviceClass method on fwk.DeviceClassResolver.
func (dcr noOpDeviceClassResolver) GetDeviceClass(resourceName v1.ResourceName) *resourceapi.DeviceClass {
	return nil
}

// -------------------------------
// Pod Group Lister Implementation
// -------------------------------

// Get is a no-op implementation of the Get method on fwk.PodGroupLister.
func (pgl noOpPodGroupLister) Get(namespace, podGroupName string) (*schedulingapi.PodGroup, error) {
	return nil, errorNotFound
}
