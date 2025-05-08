package snapshot

import (
	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1beta1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/dynamic-resource-allocation/structured"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

type Interface interface {
	// ResourceClaims exposes the Snapshot as schedulerframework.ResourceClaimTracker, in order to interact with
	// the scheduler framework.
	ResourceClaimTracker() schedulerframework.ResourceClaimTracker

	// ResourceSlices exposes the Snapshot as schedulerframework.ResourceSliceLister, in order to interact with
	// the scheduler framework.
	ResourceSliceLister() schedulerframework.ResourceSliceLister

	// DeviceClasses exposes the Snapshot as schedulerframework.DeviceClassLister, in order to interact with
	// the scheduler framework.
	DeviceClassLister() schedulerframework.DeviceClassLister

	// WrapSchedulerNodeInfo wraps the provided *schedulerframework.NodeInfo into an internal *framework.NodeInfo, adding
	// dra information. Node-local ResourceSlices are added to the NodeInfo, and all ResourceClaims referenced by each Pod
	// are added to each PodInfo. Returns an error if any of the Pods is missing a ResourceClaim.
	WrapSchedulerNodeInfo(schedNodeInfo *schedulerframework.NodeInfo) (*framework.NodeInfo, error)

	// AddClaims adds additional ResourceClaims to the Snapshot. It can be used e.g. if we need to duplicate a Pod that
	// owns ResourceClaims. Returns an error if any of the claims is already tracked in the snapshot.
	AddClaims(newClaims []*resourceapi.ResourceClaim) error

	// PodClaims returns ResourceClaims objects for all claims referenced by the Pod. If any of the referenced claims
	// isn't tracked in the Snapshot, an error is returned.
	PodClaims(pod *apiv1.Pod) ([]*resourceapi.ResourceClaim, error)

	// RemovePodOwnedClaims iterates over all the claims referenced by the Pod, and removes the ones owned by the Pod from the Snapshot.
	// Claims referenced by the Pod but not owned by it are not removed, but the Pod's reservation is removed from them.
	// This method removes all relevant claims that are in the snapshot, and doesn't error out if any of the claims are missing.
	RemovePodOwnedClaims(pod *apiv1.Pod)

	// ReservePodClaims adds a reservation for the provided Pod to all the claims it references. If any of the referenced
	// claims isn't tracked in the Snapshot, or if any of the claims are already at maximum reservation count, an error is
	// returned.
	ReservePodClaims(pod *apiv1.Pod) error

	// UnreservePodClaims removes reservations for the provided Pod from all the claims it references. If any of the referenced
	// claims isn't tracked in the Snapshot, an error is returned. If a claim is owned by the pod, or if the claim has no more reservations,
	// its allocation is cleared.
	UnreservePodClaims(pod *apiv1.Pod) error

	// NodeResourceSlices returns all node-local ResourceSlices for the given Node.
	NodeResourceSlices(nodeName string) ([]*resourceapi.ResourceSlice, bool)

	// AddNodeResourceSlices adds additional node-local ResourceSlices to the Snapshot. This should be used whenever a Node with
	// node-local ResourceSlices is duplicated in the cluster snapshot.
	AddNodeResourceSlices(nodeName string, slices []*resourceapi.ResourceSlice) error

	// RemoveNodeResourceSlices removes all node-local ResourceSlices for the Node with the given nodeName.
	// It's a no-op if there aren't any slices to remove.
	RemoveNodeResourceSlices(nodeName string)

	// TODO(mfuhol): doc
	RemoveClaims(claims []*resourceapi.ResourceClaim)
	ConfigureClaim(claim *resourceapi.ResourceClaim, newState *resourceapi.ResourceClaim) error
	ListDeviceClasses() ([]*resourceapi.DeviceClass, error)
	GetDeviceClass(className string) (*resourceapi.DeviceClass, error)
	ListResourceSlices() ([]*resourceapi.ResourceSlice, error)
	ListResourceClaims() ([]*resourceapi.ResourceClaim, error)
	GetResourceClaim(namespace, claimName string) (*resourceapi.ResourceClaim, error)
	ListAllAllocatedDevices() (sets.Set[structured.DeviceID], error)
	Commit() error
	Revert() error
	Fork()
}
