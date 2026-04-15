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

var _ fwk.SharedDRAManager = NoOpDRAManager{}

func NewNoOpDRAManager() fwk.SharedDRAManager {
	return NoOpDRAManager{}
}

func (dm NoOpDRAManager) ResourceClaims() fwk.ResourceClaimTracker {
	return (noOpResourceClaimTracker)(dm)
}

func (dm NoOpDRAManager) ResourceSlices() fwk.ResourceSliceLister {
	return (noOpResourceSliceLister)(dm)
}

func (dm NoOpDRAManager) DeviceClasses() fwk.DeviceClassLister {
	return (noOpDeviceClassLister)(dm)
}

func (dm NoOpDRAManager) DeviceClassResolver() fwk.DeviceClassResolver {
	return (noOpDeviceClassResolver)(dm)
}

func (dm NoOpDRAManager) PodGroups() fwk.PodGroupLister {
	return (noOpPodGroupLister)(dm)
}

// -------------------------------------
// Resource Claim Tracker Implementation
// -------------------------------------

func (ct noOpResourceClaimTracker) List() ([]*resourceapi.ResourceClaim, error) {
	return nil, nil
}

func (ct noOpResourceClaimTracker) Get(namespace, claimName string) (*resourceapi.ResourceClaim, error) {
	return nil, errorNotFound
}

func (ct noOpResourceClaimTracker) ListAllAllocatedDevices() (sets.Set[structured.DeviceID], error) {
	return nil, nil
}

func (ct noOpResourceClaimTracker) GatherAllocatedState() (*structured.AllocatedState, error) {
	return nil, nil
}

func (ct noOpResourceClaimTracker) SignalClaimPendingAllocation(claimUID types.UID, allocatedClaim *resourceapi.ResourceClaim) error {
	return nil
}

func (ct noOpResourceClaimTracker) GetPendingAllocation(claimUID types.UID) *resourceapi.AllocationResult {
	return nil
}

func (ct noOpResourceClaimTracker) MaybeRemoveClaimPendingAllocation(claimUID types.UID, forceRemove bool) (deleted bool) {
	return false
}

func (ct noOpResourceClaimTracker) AssumeClaimAfterAPICall(claim *resourceapi.ResourceClaim) error {
	return nil
}

func (ct noOpResourceClaimTracker) AssumedClaimRestore(namespace, claimName string) {

}

func (sl noOpResourceSliceLister) ListWithDeviceTaintRules() ([]*resourceapi.ResourceSlice, error) {
	return nil, nil
}

// ----------------------------------
// Device Class Lister Implementation
// ----------------------------------

func (dcl noOpDeviceClassLister) List() ([]*resourceapi.DeviceClass, error) {
	return nil, nil
}

func (dcl noOpDeviceClassLister) Get(className string) (*resourceapi.DeviceClass, error) {
	return nil, errorNotFound
}

// ------------------------------------
// Device Class Resolver Implementation
// ------------------------------------

func (dcr noOpDeviceClassResolver) GetDeviceClass(resourceName v1.ResourceName) *resourceapi.DeviceClass {
	return nil
}

// -------------------------------
// Pod Group Lister Implementation
// -------------------------------

func (pgl noOpPodGroupLister) Get(namespace, podGroupName string) (*schedulingapi.PodGroup, error) {
	return nil, errorNotFound
}
