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

package predicate

import (
	"context"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	csisnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/csi/snapshot"
	drasnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	drautils "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/utils"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/dynamic-resource-allocation/resourceclaim"
	schedulerinterface "k8s.io/kube-scheduler/framework"
	schedulerimpl "k8s.io/kubernetes/pkg/scheduler/framework"
)

// PredicateSnapshot implements ClusterSnapshot on top of a ClusterSnapshotStore by using
// SchedulerBasedPredicateChecker to check scheduler predicates.
type PredicateSnapshot struct {
	clustersnapshot.ClusterSnapshotStore
	pluginRunner                 *SchedulerPluginRunner
	draEnabled                   bool
	enableCSINodeAwareScheduling bool
	parallelism                  int
	draSnapshot                  *drasnapshot.Snapshot
	csiSnapshot                  *csisnapshot.Snapshot
}

// NewPredicateSnapshot builds a PredicateSnapshot.
func NewPredicateSnapshot(snapshotStore clustersnapshot.ClusterSnapshotStore, fwHandle *framework.Handle, draEnabled bool, parallelism int, enableCSINodeAwareScheduling bool) *PredicateSnapshot {
	snapshot := &PredicateSnapshot{
		ClusterSnapshotStore:         snapshotStore,
		draEnabled:                   draEnabled,
		enableCSINodeAwareScheduling: enableCSINodeAwareScheduling,
		parallelism:                  parallelism,
		draSnapshot:                  drasnapshot.NewEmptySnapshot(),
		csiSnapshot:                  csisnapshot.NewEmptySnapshot(),
	}
	// Plugin runner really only needs a framework.SharedLister for running the plugins, but it also needs to run the provided Node-matching functions
	// which operate on *framework.NodeInfo. The only object that allows obtaining *framework.NodeInfos is PredicateSnapshot, so we have an ugly circular
	// dependency between PluginRunner and PredicateSnapshot.
	// TODO: Refactor PluginRunner so that it doesn't depend on PredicateSnapshot (e.g. move retrieving NodeInfos out of PluginRunner, to PredicateSnapshot).
	snapshot.pluginRunner = NewSchedulerPluginRunner(fwHandle, snapshot, parallelism)
	return snapshot
}

// SetClusterState resets the snapshot to an unforked state and replaces the contents of the snapshot
// with the provided data. scheduledPods are correlated to their Nodes based on spec.NodeName.
// The provided draSnapshot and csiSnapshot are treated as the source of truth and are eagerly
// loaded into the created NodeInfo/PodInfo objects.
func (s *PredicateSnapshot) SetClusterState(nodes []*apiv1.Node, scheduledPods []*apiv1.Pod, draSnapshot *drasnapshot.Snapshot, csiSnapshot *csisnapshot.Snapshot) error {
	s.ClusterSnapshotStore.Clear()

	if draSnapshot == nil {
		draSnapshot = drasnapshot.NewEmptySnapshot()
	}
	s.draSnapshot = draSnapshot

	if csiSnapshot == nil {
		csiSnapshot = csisnapshot.NewEmptySnapshot()
	}
	s.csiSnapshot = csiSnapshot

	nodeInfos := make([]*framework.NodeInfo, len(nodes))
	nodeNameToIdx := make(map[string]int, len(nodes))
	for i, node := range nodes {
		var slices []*resourceapi.ResourceSlice
		if s.draEnabled && draSnapshot != nil {
			slices, _ = draSnapshot.NodeResourceSlices(node.Name)
		}
		ni := framework.NewNodeInfo(node, slices)

		if s.enableCSINodeAwareScheduling && csiSnapshot != nil {
			csiNode, err := csiSnapshot.Get(node.Name)
			if err != nil {
				return fmt.Errorf("couldn't obtain csi node: %v", err)
			}
			ni.SetCSINode(csiNode)
		}

		nodeInfos[i] = ni
		nodeNameToIdx[node.Name] = i
	}

	if s.parallelism > 1 {
		if err := s.setClusterStatePodsParallelized(nodeInfos, nodeNameToIdx, scheduledPods); err != nil {
			return err
		}
	} else {
		// TODO(macsko): Migrate to setClusterStatePodsParallelized for parallelism == 1
		// after making sure the implementation is always correct in CA 1.33.
		if err := s.setClusterStatePodsSequential(nodeInfos, nodeNameToIdx, scheduledPods); err != nil {
			return err
		}
	}

	// We build NodeInfo objects with all their pods before adding them to the store.
	// This allows parallelizing PodInfo construction and enables the store to
	// perform bulk internal state / cache updates per-node rather than per-pod.
	for _, ni := range nodeInfos {
		if err := s.ClusterSnapshotStore.StoreNodeInfo(ni); err != nil {
			return err
		}
	}

	return nil
}

func (s *PredicateSnapshot) setClusterStatePodsSequential(nodeInfos []*framework.NodeInfo, nodeNameToIdx map[string]int, scheduledPods []*apiv1.Pod) error {
	for _, pod := range scheduledPods {
		if nodeIdx, ok := nodeNameToIdx[pod.Spec.NodeName]; ok {
			var claims []*resourceapi.ResourceClaim
			if s.draEnabled && s.draSnapshot != nil {
				var err error
				claims, err = s.draSnapshot.PodClaims(pod)
				if err != nil {
					return fmt.Errorf("couldn't obtain pod %s/%s claims: %v", pod.Namespace, pod.Name, err)
				}
			}
			podInfo := framework.NewPodInfo(pod, claims)
			nodeInfos[nodeIdx].AddPodInfo(podInfo)
		}
	}
	return nil
}

func (s *PredicateSnapshot) setClusterStatePodsParallelized(nodeInfos []*framework.NodeInfo, nodeNameToIdx map[string]int, scheduledPods []*apiv1.Pod) error {
	podInfosForNode := make([][]*framework.PodInfo, len(nodeInfos))
	for _, pod := range scheduledPods {
		nodeIdx, ok := nodeNameToIdx[pod.Spec.NodeName]
		if !ok {
			continue
		}

		var claims []*resourceapi.ResourceClaim
		if s.draEnabled && s.draSnapshot != nil {
			var err error
			claims, err = s.draSnapshot.PodClaims(pod)
			if err != nil {
				return fmt.Errorf("couldn't obtain pod %s/%s claims: %v", pod.Namespace, pod.Name, err)
			}
		}

		podInfo := framework.NewPodInfo(pod, claims)
		podInfosForNode[nodeIdx] = append(podInfosForNode[nodeIdx], podInfo)
	}

	ctx := context.Background()
	workqueue.ParallelizeUntil(ctx, s.parallelism, len(nodeInfos), func(nodeIdx int) {
		nodeInfo := nodeInfos[nodeIdx]
		for _, pi := range podInfosForNode[nodeIdx] {
			nodeInfo.AddPodInfo(pi)
		}
	})

	return nil
}

// GetNodeInfo returns an internal NodeInfo wrapping the relevant schedulerimpl.NodeInfo.
//
// TODO(DRA): Beware that it may return stale ResourceClaim data.
// See framework.PodInfo.NeededResourceClaims comment.
func (s *PredicateSnapshot) GetNodeInfo(nodeName string) (*framework.NodeInfo, error) {
	schedNodeInfo, err := s.ClusterSnapshotStore.NodeInfos().Get(nodeName)
	if err != nil {
		return nil, err
	}
	nodeInfo, ok := schedNodeInfo.(*framework.NodeInfo)
	if !ok {
		return nil, fmt.Errorf("expected: %T, got: %T in the underlying store", &framework.NodeInfo{}, schedNodeInfo)
	}

	return nodeInfo, nil
}

// ListNodeInfos returns internal NodeInfos wrapping all schedulerimpl.NodeInfos in the snapshot.
//
// TODO(DRA): Beware that it may return stale ResourceClaim data.
// See PodInfo.NeededResourceClaims comment.
func (s *PredicateSnapshot) ListNodeInfos() ([]*framework.NodeInfo, error) {
	schedNodeInfos, err := s.ClusterSnapshotStore.NodeInfos().List()
	if err != nil {
		return nil, err
	}
	var result []*framework.NodeInfo
	for _, schedNodeInfo := range schedNodeInfos {
		nodeInfo, ok := schedNodeInfo.(*framework.NodeInfo)
		if !ok {
			return nil, fmt.Errorf("expected: %T, got: %T in the underlying store", &framework.NodeInfo{}, schedNodeInfo)
		}

		result = append(result, nodeInfo)
	}
	return result, nil
}

// AddNodeInfo adds the provided internal NodeInfo to the snapshot.
// The DRA slices and CSI data are explicitly pushed into the underlying
// draSnapshot and csiSnapshot before being stored.
func (s *PredicateSnapshot) AddNodeInfo(nodeInfo *framework.NodeInfo) error {
	if s.draEnabled {
		// TODO(DRA): Add transaction-like clean-up in case of errors here - don't modify the state on any errors.
		if len(nodeInfo.LocalResourceSlices) > 0 {
			if err := s.draSnapshot.AddNodeResourceSlices(nodeInfo.Node().Name, nodeInfo.LocalResourceSlices); err != nil {
				return fmt.Errorf("couldn't add ResourceSlices to DRA snapshot: %v", err)
			}
		}

		for _, podInfo := range nodeInfo.Pods() {
			if len(podInfo.NeededResourceClaims) > 0 {
				if err := s.modifyResourceClaimsForNewPod(podInfo); err != nil {
					return err
				}
				if err := s.verifyScheduledPodResourceClaims(podInfo.Pod, nodeInfo.Node()); err != nil {
					return err
				}
			}
		}
	}

	if s.enableCSINodeAwareScheduling {
		if nodeInfo.CSINode != nil {
			if err := s.csiSnapshot.AddCSINode(nodeInfo.CSINode); err != nil {
				return err
			}
		}
	}

	return s.ClusterSnapshotStore.StoreNodeInfo(nodeInfo)
}

// RemoveNodeInfo removes a NodeInfo matching the provided nodeName from the snapshot.
// The DRA slices and CSI data are removed from the underlying draSnapshot and csiSnapshot.
func (s *PredicateSnapshot) RemoveNodeInfo(nodeName string) error {
	nodeInfo, err := s.GetNodeInfo(nodeName)
	if err != nil {
		return err
	}

	if err := s.ClusterSnapshotStore.RemoveNodeInfo(nodeName); err != nil {
		return err
	}

	if s.draEnabled {
		s.draSnapshot.RemoveNodeResourceSlices(nodeName)

		for _, pod := range nodeInfo.Pods() {
			s.draSnapshot.RemovePodOwnedClaims(pod.Pod)
		}
	}
	if s.enableCSINodeAwareScheduling {
		// generally a node name is same as csi node name and hence we should be safe
		if nodeInfo.CSINode != nil {
			s.csiSnapshot.RemoveCSINode(nodeName)
		}
	}
	return nil
}

// SchedulePod adds pod to the snapshot and schedules it to given node.
// The scheduler's Reserve phase computes DRA claim allocations, which are pushed to the draSnapshot.
// The updated claims are then pulled to construct the final PodInfo.
func (s *PredicateSnapshot) SchedulePod(pod *apiv1.Pod, nodeName string) clustersnapshot.SchedulingError {
	node, cycleState, schedErr := s.pluginRunner.RunFiltersOnNode(pod, nodeName)
	if schedErr != nil {
		return schedErr
	}

	if s.draEnabled && len(pod.Spec.ResourceClaims) > 0 {
		// TODO(DRA): Add transaction-like clean-up in case of errors here - don't modify the state on any errors.
		if err := s.modifyResourceClaimsForScheduledPod(pod, node, cycleState); err != nil {
			return clustersnapshot.NewSchedulingInternalError(pod, err.Error())
		}
		if err := s.verifyScheduledPodResourceClaims(pod, node); err != nil {
			return clustersnapshot.NewSchedulingInternalError(pod, err.Error())
		}
	}

	podInfo, err := s.createPodInfo(pod)
	if err != nil {
		return clustersnapshot.NewSchedulingInternalError(pod, err.Error())
	}
	if err := s.ClusterSnapshotStore.StorePodInfo(podInfo, nodeName); err != nil {
		return clustersnapshot.NewSchedulingInternalError(pod, err.Error())
	}
	return nil
}

// SchedulePodOnAnyNodeMatching adds pod to the snapshot and schedules it to any node matching the provided function.
// Data flow is the same as SchedulePod.
func (s *PredicateSnapshot) SchedulePodOnAnyNodeMatching(pod *apiv1.Pod, opts clustersnapshot.SchedulingOptions) (string, clustersnapshot.SchedulingError) {
	node, cycleState, schedErr := s.pluginRunner.RunFiltersUntilPassingNode(pod, opts)
	if schedErr != nil {
		return "", schedErr
	}

	if s.draEnabled && len(pod.Spec.ResourceClaims) > 0 {
		// TODO(DRA): Add transaction-like clean-up in case of errors here - don't modify the state on any errors.
		if err := s.modifyResourceClaimsForScheduledPod(pod, node, cycleState); err != nil {
			return "", clustersnapshot.NewSchedulingInternalError(pod, err.Error())
		}
		if err := s.verifyScheduledPodResourceClaims(pod, node); err != nil {
			return "", clustersnapshot.NewSchedulingInternalError(pod, err.Error())
		}
	}

	podInfo, err := s.createPodInfo(pod)
	if err != nil {
		return "", clustersnapshot.NewSchedulingInternalError(pod, err.Error())
	}
	if err := s.ClusterSnapshotStore.StorePodInfo(podInfo, node.Name); err != nil {
		return "", clustersnapshot.NewSchedulingInternalError(pod, err.Error())
	}
	return node.Name, nil
}

// UnschedulePod removes the given Pod from the given Node inside the snapshot.
func (s *PredicateSnapshot) UnschedulePod(namespace string, podName string, nodeName string) error {
	if s.draEnabled {
		nodeInfo, err := s.GetNodeInfo(nodeName)
		if err != nil {
			return err
		}

		var foundPod *apiv1.Pod
		for _, pod := range nodeInfo.Pods() {
			if pod.Namespace == namespace && pod.Name == podName {
				foundPod = pod.Pod
				break
			}
		}
		if foundPod == nil {
			return fmt.Errorf("pod %s/%s not found on node %s", namespace, podName, nodeName)
		}

		if len(foundPod.Spec.ResourceClaims) > 0 {
			if err := s.draSnapshot.UnreservePodClaims(foundPod); err != nil {
				return err
			}
		}
	}

	return s.ClusterSnapshotStore.RemovePodInfo(namespace, podName, nodeName)
}

// ForceAddPod adds the given Pod to the Node with the given nodeName inside the snapshot without checking scheduler predicates.
// This method will allocate internal PodInfo and include the DRA-related information (taken from the DRA snapshot).
// It assumes the draSnapshot is already populated by the caller and just pulls from it to build the PodInfo.
// It does NOT compute or reserve new claims.
func (s *PredicateSnapshot) ForceAddPod(pod *apiv1.Pod, nodeName string) error {
	podInfo, err := s.createPodInfo(pod)
	if err != nil {
		return err
	}

	return s.ClusterSnapshotStore.StorePodInfo(podInfo, nodeName)
}

// ForceRemovePod removes the given Pod from the snapshot.
func (s *PredicateSnapshot) ForceRemovePod(namespace string, podName string, nodeName string) error {
	return s.RemovePodInfo(namespace, podName, nodeName)
}

// CheckPredicates checks whether scheduler predicates pass for the given pod on the given node.
func (s *PredicateSnapshot) CheckPredicates(pod *apiv1.Pod, nodeName string) clustersnapshot.SchedulingError {
	_, _, err := s.pluginRunner.RunFiltersOnNode(pod, nodeName)
	return err
}

// DraSnapshot returns an interface that allows accessing and modifying the DRA objects in the snapshot.
func (s *PredicateSnapshot) DraSnapshot() *drasnapshot.Snapshot {
	return s.draSnapshot
}

// CsiSnapshot returns an interface that allows accessing and modifying the CSINode objects in the snapshot.
func (s *PredicateSnapshot) CsiSnapshot() *csisnapshot.Snapshot {
	return s.csiSnapshot
}

// Clear resets the snapshot to an empty, unforked state.
func (s *PredicateSnapshot) Clear() {
	s.ClusterSnapshotStore.Clear()
	s.draSnapshot = drasnapshot.NewEmptySnapshot()
	s.csiSnapshot = csisnapshot.NewEmptySnapshot()
}

// Fork creates a fork of snapshot state. All modifications can later be reverted to moment of forking via Revert().
func (s *PredicateSnapshot) Fork() {
	s.ClusterSnapshotStore.Fork()
	s.draSnapshot.Fork()
	s.csiSnapshot.Fork()
}

// Revert reverts snapshot state to moment of forking.
func (s *PredicateSnapshot) Revert() {
	s.ClusterSnapshotStore.Revert()
	s.draSnapshot.Revert()
	s.csiSnapshot.Revert()
}

// Commit commits changes done after forking.
func (s *PredicateSnapshot) Commit() error {
	if err := s.ClusterSnapshotStore.Commit(); err != nil {
		return err
	}
	s.draSnapshot.Commit()
	s.csiSnapshot.Commit()
	return nil
}

// ResourceClaims exposes snapshot as ResourceClaimTracker
func (s *PredicateSnapshot) ResourceClaims() schedulerinterface.ResourceClaimTracker {
	return s.draSnapshot.ResourceClaims()
}

// ResourceSlices exposes snapshot as ResourceSliceLister.
func (s *PredicateSnapshot) ResourceSlices() schedulerinterface.ResourceSliceLister {
	return s.draSnapshot.ResourceSlices()
}

// DeviceClasses exposes the snapshot as DeviceClassLister.
func (s *PredicateSnapshot) DeviceClasses() schedulerinterface.DeviceClassLister {
	return s.draSnapshot.DeviceClasses()
}

// DeviceClassResolver exposes the snapshot as DeviceClassResolver.
func (s *PredicateSnapshot) DeviceClassResolver() schedulerinterface.DeviceClassResolver {
	return s.draSnapshot.DeviceClassResolver()
}

// CSINodes returns the CSI nodes snapshot.
func (s *PredicateSnapshot) CSINodes() schedulerinterface.CSINodeLister {
	return s.csiSnapshot.CSINodes()
}

// verifyScheduledPodResourceClaims verifies that all needed claims are tracked in the DRA snapshot, allocated, and available on the Node.
func (s *PredicateSnapshot) verifyScheduledPodResourceClaims(pod *apiv1.Pod, node *apiv1.Node) error {
	claims, err := s.draSnapshot.PodClaims(pod)
	if err != nil {
		return fmt.Errorf("couldn't obtain pod %s/%s claims: %v", pod.Namespace, pod.Name, err)
	}
	for _, claim := range claims {
		if available, err := drautils.ClaimAvailableOnNode(claim, node); err != nil || !available {
			return fmt.Errorf("pod %s/%s needs claim %s to schedule, but it isn't available on node %s (allocated: %v, available: %v, err: %v)", pod.Namespace, pod.Name, claim.Name, node.Name, drautils.ClaimAllocated(claim), available, err)
		}
	}
	return nil
}

func (s *PredicateSnapshot) createPodInfo(pod *apiv1.Pod) (*framework.PodInfo, error) {
	var claims []*resourceapi.ResourceClaim
	if s.draEnabled {
		var err error
		claims, err = s.draSnapshot.PodClaims(pod)
		if err != nil {
			return nil, fmt.Errorf("couldn't obtain pod %s/%s claims: %v", pod.Namespace, pod.Name, err)
		}
	}
	return framework.NewPodInfo(pod, claims), nil
}

func (s *PredicateSnapshot) modifyResourceClaimsForScheduledPod(pod *apiv1.Pod, node *apiv1.Node, postFilterState *schedulerimpl.CycleState) error {
	// We need to run the scheduler Reserve phase to allocate the appropriate ResourceClaims in the DRA snapshot. The allocations are
	// actually computed and cached in the Filter phase, and Reserve only grabs them from the cycle state. So this should be quick, but
	// it needs the cycle state from after running the Filter phase.
	if err := s.pluginRunner.RunReserveOnNode(pod, node.Name, postFilterState); err != nil {
		return fmt.Errorf("error while trying to run Reserve node %s for pod %s/%s: %v", node.Name, pod.Namespace, pod.Name, err)
	}

	// The pod isn't added to the ReservedFor field of the claim during the Reserve phase (it happens later, in PreBind). We can just do it
	// manually here. It shouldn't fail, it only fails if ReservedFor is at max length already, but that is checked during the Filter phase.
	if err := s.draSnapshot.ReservePodClaims(pod); err != nil {
		return fmt.Errorf("couldn't add pod %s/%s reservations to claims, this shouldn't happen: %v", pod.Namespace, pod.Name, err)
	}
	return nil
}

func (s *PredicateSnapshot) modifyResourceClaimsForNewPod(podInfo *framework.PodInfo) error {
	// PodInfo.NeededResourceClaims contains both Pod-owned and shared ResourceClaims that the Pod needs to schedule. When a new Pod is added to
	// ClusterSnapshot, we need to add the Pod-owned claims to the DRA snapshot. The shared ones should already be in the DRA snapshot,
	// so we don't add them. The claims should already be allocated in the provided PodInfo.
	var podOwnedClaims []*resourceapi.ResourceClaim
	for _, claim := range podInfo.NeededResourceClaims {
		if err := resourceclaim.IsForPod(podInfo.Pod, claim); err == nil {
			podOwnedClaims = append(podOwnedClaims, claim)
		}
	}
	if err := s.draSnapshot.AddClaims(podOwnedClaims); err != nil {
		return fmt.Errorf("couldn't add ResourceSlices for pod %s/%s to DRA snapshot: %v", podInfo.Namespace, podInfo.Name, err)
	}

	// The Pod-owned claims should already be reserved for the Pod after sanitization, but we need to add the reservation for the new Pod
	// to the shared claims. This can fail if doing so would exceed the max reservation limit for a claim.
	if err := s.draSnapshot.ReservePodClaims(podInfo.Pod); err != nil {
		return fmt.Errorf("couldn't add pod %s/%s reservations to claims, this shouldn't happen: %v", podInfo.Namespace, podInfo.Name, err)
	}
	return nil
}
