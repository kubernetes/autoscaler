package snapshot

import (
	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1beta1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

func mergeNestedDeltaMaps[K1 comparable, K2 comparable, V any](maps []map[K1]map[K2]*V) map[K1]map[K2]*V {
	merged := make(map[K1]map[K2]*V)
	ignored := make(map[K1]bool)

	for _, mapping := range maps {
		for k1, nestedMap := range mapping {
			mergedNested, exists := merged[k1]

			if ignored[k1] {
				continue
			}

			if nestedMap == nil {
				ignored[k1] = true
				continue
			}

			if exists && mergedNested == nil {
				continue
			}

			if !exists {
				merged[k1] = nestedMap
			} else {
				toMerge := []map[K2]*V{mergedNested, nestedMap}
				merged[k1] = mergeDeltaMaps(toMerge)
			}
		}
	}

	for k1, nestedMap := range merged {
		if nestedMap == nil {
			delete(merged, k1)
		}

		for k2, value := range nestedMap {
			if value == nil {
				delete(nestedMap, k2)
			}
		}
	}

	return merged
}

func mergeDeltaMaps[K comparable, V any](maps []map[K]*V) map[K]*V {
	merged := make(map[K]*V)

	for _, mapping := range maps {
		for k, v := range mapping {
			if _, exists := merged[k]; exists {
				continue
			}

			merged[k] = v
		}
	}

	for k, v := range merged {
		if v == nil {
			delete(merged, k)
		}
	}

	return merged
}

func mergeSnapshots(deltas []Snapshot) Snapshot {
	claims := []map[ResourceClaimId]*resourceapi.ResourceClaim{}
	devices := []map[string]*resourceapi.DeviceClass{}
	nodeSlices := []map[string]map[string]*resourceapi.ResourceSlice{}
	globalSlices := []map[string]*resourceapi.ResourceSlice{}

	for i := len(deltas) - 1; i >= 0; i-- {
		delta := deltas[i]
		claims = append(claims, delta.resourceClaimsById)
		devices = append(devices, delta.deviceClasses)
		nodeSlices = append(nodeSlices, delta.resourceSlicesByNodeName)
		globalSlices = append(globalSlices, delta.nonNodeLocalResourceSlices)
	}

	return Snapshot{
		resourceClaimsById:         mergeDeltaMaps(claims),
		deviceClasses:              mergeDeltaMaps(devices),
		nonNodeLocalResourceSlices: mergeDeltaMaps(globalSlices),
		resourceSlicesByNodeName:   mergeNestedDeltaMaps(nodeSlices),
	}
}

type DeltaSnapshot struct {
	deltas         []Snapshot
	mergedSnapshot Snapshot
}

// ResourceClaims exposes the Snapshot as schedulerframework.ResourceClaimTracker, in order to interact with
// the scheduler framework.
func (s DeltaSnapshot) ResourceClaims() schedulerframework.ResourceClaimTracker {
	return snapshotClaimTracker(s.mergedSnapshot)
}

// ResourceSlices exposes the Snapshot as schedulerframework.ResourceSliceLister, in order to interact with
// the scheduler framework.
func (s DeltaSnapshot) ResourceSlices() schedulerframework.ResourceSliceLister {
	return snapshotSliceLister(s.mergedSnapshot)
}

// DeviceClasses exposes the Snapshot as schedulerframework.DeviceClassLister, in order to interact with
// the scheduler framework.
func (s DeltaSnapshot) DeviceClasses() schedulerframework.DeviceClassLister {
	return snapshotClassLister(s.mergedSnapshot)
}

// WrapSchedulerNodeInfo wraps the provided *schedulerframework.NodeInfo into an internal *framework.NodeInfo, adding
// dra information. Node-local ResourceSlices are added to the NodeInfo, and all ResourceClaims referenced by each Pod
// are added to each PodInfo. Returns an error if any of the Pods is missing a ResourceClaim.
func (s DeltaSnapshot) WrapSchedulerNodeInfo(schedNodeInfo *schedulerframework.NodeInfo) (*framework.NodeInfo, error) {
	return s.mergedSnapshot.WrapSchedulerNodeInfo(schedNodeInfo)
}

// Clone returns a copy of this Snapshot that can be independently modified without affecting this Snapshot.
// The only mutable objects in the Snapshot are ResourceClaims, so they are deep-copied. The rest is only a
// shallow copy.
// func (s DeltaSnapshot) Clone() DeltaSnapshot {
// 	merged := mergeSnapshots(s.deltas)
// 	return DeltaSnapshot{deltas: []Snapshot{merged}}
// }

// AddClaims adds additional ResourceClaims to the Snapshot. It can be used e.g. if we need to duplicate a Pod that
// owns ResourceClaims. Returns an error if any of the claims is already tracked in the snapshot.
func (s DeltaSnapshot) AddClaims(newClaims []*resourceapi.ResourceClaim) error {
	if err := s.mergedSnapshot.AddClaims(newClaims); err != nil {
		return err
	}

	if err := s.deltas[len(s.deltas)-1].AddClaims(newClaims); err != nil {
		return err
	}

	return nil
}

// PodClaims returns ResourceClaims objects for all claims referenced by the Pod. If any of the referenced claims
// isn't tracked in the Snapshot, an error is returned.
func (s DeltaSnapshot) PodClaims(pod *apiv1.Pod) ([]*resourceapi.ResourceClaim, error) {
	return s.mergedSnapshot.PodClaims(pod)
}

// RemovePodOwnedClaims iterates over all the claims referenced by the Pod, and removes the ones owned by the Pod from the Snapshot.
// Claims referenced by the Pod but not owned by it are not removed, but the Pod's reservation is removed from them.
// This method removes all relevant claims that are in the snapshot, and doesn't error out if any of the claims are missing.
func (s DeltaSnapshot) RemovePodOwnedClaims(pod *apiv1.Pod) {
	s.mergedSnapshot.RemovePodOwnedClaims(pod)
	s.deltas[len(s.deltas)-1].RemovePodOwnedClaims(pod)
}

// ReservePodClaims adds a reservation for the provided Pod to all the claims it references. If any of the referenced
// claims isn't tracked in the Snapshot, or if any of the claims are already at maximum reservation count, an error is
// returned.
func (s DeltaSnapshot) ReservePodClaims(pod *apiv1.Pod) error {
	if err := s.mergedSnapshot.ReservePodClaims(pod); err != nil {
		return err
	}

	if err := s.deltas[len(s.deltas)-1].ReservePodClaims(pod); err != nil {
		return err
	}

	return nil
}

// UnreservePodClaims removes reservations for the provided Pod from all the claims it references. If any of the referenced
// claims isn't tracked in the Snapshot, an error is returned. If a claim is owned by the pod, or if the claim has no more reservations,
// its allocation is cleared.
func (s DeltaSnapshot) UnreservePodClaims(pod *apiv1.Pod) error {
	if err := s.mergedSnapshot.UnreservePodClaims(pod); err != nil {
		return err
	}

	if err := s.deltas[len(s.deltas)-1].UnreservePodClaims(pod); err != nil {
		return err
	}

	return nil
}

// NodeResourceSlices returns all node-local ResourceSlices for the given Node.
func (s DeltaSnapshot) NodeResourceSlices(nodeName string) ([]*resourceapi.ResourceSlice, bool) {
	return s.mergedSnapshot.NodeResourceSlices(nodeName)
}

// AddNodeResourceSlices adds additional node-local ResourceSlices to the Snapshot. This should be used whenever a Node with
// node-local ResourceSlices is duplicated in the cluster snapshot.
func (s DeltaSnapshot) AddNodeResourceSlices(nodeName string, slices []*resourceapi.ResourceSlice) error {
	if err := s.mergedSnapshot.AddNodeResourceSlices(nodeName, slices); err != nil {
		return err
	}

	if err := s.deltas[len(s.deltas)-1].AddNodeResourceSlices(nodeName, slices); err != nil {
		return err
	}

	return nil
}

// RemoveNodeResourceSlices removes all node-local ResourceSlices for the Node with the given nodeName.
// It's a no-op if there aren't any slices to remove.
func (s DeltaSnapshot) RemoveNodeResourceSlices(nodeName string) {
	s.mergedSnapshot.RemoveNodeResourceSlices(nodeName)
	s.deltas[len(s.deltas)-1].RemoveNodeResourceSlices(nodeName)
}

func (s DeltaSnapshot) Fork() {
	s.deltas = append(s.deltas, Snapshot{})
}

func (s DeltaSnapshot) Revert() {
	s.deltas = s.deltas[:len(s.deltas)-1]
	s.mergedSnapshot = mergeSnapshots(s.deltas)
}

func (s DeltaSnapshot) Commit() error {
	if len(s.deltas) <= 1 {
		// do nothing
		return nil
	}

	forkedSnapshots := s.deltas[len(s.deltas)-2:]
	deltasWithoutForked := s.deltas[:len(s.deltas)-2]
	s.deltas = append(deltasWithoutForked, mergeSnapshots(forkedSnapshots))
	return nil
}
