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
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	resourceapi "k8s.io/api/resource/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	drautils "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/utils"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/dynamic-resource-allocation/resourceclaim"
	schedulerimpl "k8s.io/kubernetes/pkg/scheduler/framework"
)

// PredicateSnapshot implements ClusterSnapshot on top of a ClusterSnapshotStore by using
// SchedulerBasedPredicateChecker to check scheduler predicates.
type PredicateSnapshot struct {
	clustersnapshot.ClusterSnapshotStore
	pluginRunner                 *SchedulerPluginRunner
	draEnabled                   bool
	enableCSINodeAwareScheduling bool
}

// NewPredicateSnapshot builds a PredicateSnapshot.
func NewPredicateSnapshot(snapshotStore clustersnapshot.ClusterSnapshotStore, fwHandle *framework.Handle, draEnabled bool, parallelism int, enableCSINodeAwareScheduling bool) *PredicateSnapshot {
	snapshot := &PredicateSnapshot{
		ClusterSnapshotStore:         snapshotStore,
		draEnabled:                   draEnabled,
		enableCSINodeAwareScheduling: enableCSINodeAwareScheduling,
	}
	// Plugin runner really only needs a framework.SharedLister for running the plugins, but it also needs to run the provided Node-matching functions
	// which operate on *framework.NodeInfo. The only object that allows obtaining *framework.NodeInfos is PredicateSnapshot, so we have an ugly circular
	// dependency between PluginRunner and PredicateSnapshot.
	// TODO: Refactor PluginRunner so that it doesn't depend on PredicateSnapshot (e.g. move retrieving NodeInfos out of PluginRunner, to PredicateSnapshot).
	snapshot.pluginRunner = NewSchedulerPluginRunner(fwHandle, snapshot, parallelism)
	return snapshot
}

// GetNodeInfo returns an internal NodeInfo wrapping the relevant schedulerimpl.NodeInfo.
func (s *PredicateSnapshot) GetNodeInfo(nodeName string) (*framework.NodeInfo, error) {
	schedNodeInfo, err := s.ClusterSnapshotStore.NodeInfos().Get(nodeName)
	if err != nil {
		return nil, err
	}

	wrappedNodeInfo := framework.WrapSchedulerNodeInfo(schedNodeInfo, nil, nil)
	if s.draEnabled {
		wrappedNodeInfo, err = s.ClusterSnapshotStore.DraSnapshot().WrapSchedulerNodeInfo(schedNodeInfo)
		if err != nil {
			return nil, err
		}
	}

	if s.enableCSINodeAwareScheduling {
		wrappedNodeInfo, err = s.ClusterSnapshotStore.CsiSnapshot().AddCSINodeInfoToNodeInfo(wrappedNodeInfo)
		if err != nil {
			return nil, err
		}
	}
	return wrappedNodeInfo, nil
}

// ListNodeInfos returns internal NodeInfos wrapping all schedulerimpl.NodeInfos in the snapshot.
func (s *PredicateSnapshot) ListNodeInfos() ([]*framework.NodeInfo, error) {
	schedNodeInfos, err := s.ClusterSnapshotStore.NodeInfos().List()
	if err != nil {
		return nil, err
	}
	var result []*framework.NodeInfo
	for _, schedNodeInfo := range schedNodeInfos {
		wrappedNodeInfo := framework.WrapSchedulerNodeInfo(schedNodeInfo, nil, nil)

		var err error
		if s.draEnabled {
			wrappedNodeInfo, err = s.ClusterSnapshotStore.DraSnapshot().WrapSchedulerNodeInfo(schedNodeInfo)
			if err != nil {
				return nil, err
			}
		}

		if s.enableCSINodeAwareScheduling {
			wrappedNodeInfo, err = s.ClusterSnapshotStore.CsiSnapshot().AddCSINodeInfoToNodeInfo(wrappedNodeInfo)
			if err != nil {
				return nil, err
			}
		}
		result = append(result, wrappedNodeInfo)
	}
	return result, nil
}

// AddNodeInfo adds the provided internal NodeInfo to the snapshot.
func (s *PredicateSnapshot) AddNodeInfo(nodeInfo *framework.NodeInfo) error {
	if s.draEnabled {
		// TODO(DRA): Add transaction-like clean-up in case of errors here - don't modify the state on any errors.
		if len(nodeInfo.LocalResourceSlices) > 0 {
			if err := s.ClusterSnapshotStore.DraSnapshot().AddNodeResourceSlices(nodeInfo.Node().Name, nodeInfo.LocalResourceSlices); err != nil {
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
			if err := s.ClusterSnapshotStore.CsiSnapshot().AddCSINode(nodeInfo.CSINode); err != nil {
				return err
			}
		}
	}

	return s.ClusterSnapshotStore.AddSchedulerNodeInfo(nodeInfo.ToScheduler())
}

// RemoveNodeInfo removes a NodeInfo matching the provided nodeName from the snapshot.
func (s *PredicateSnapshot) RemoveNodeInfo(nodeName string) error {
	nodeInfo, err := s.GetNodeInfo(nodeName)
	if err != nil {
		return err
	}

	if err := s.ClusterSnapshotStore.RemoveSchedulerNodeInfo(nodeName); err != nil {
		return err
	}

	if s.draEnabled {
		s.ClusterSnapshotStore.DraSnapshot().RemoveNodeResourceSlices(nodeName)

		for _, pod := range nodeInfo.Pods() {
			s.ClusterSnapshotStore.DraSnapshot().RemovePodOwnedClaims(pod.Pod)
		}
	}
	if s.enableCSINodeAwareScheduling {
		// generally a node name is same as csi node name and hence we should be safe
		if nodeInfo.CSINode != nil {
			s.ClusterSnapshotStore.CsiSnapshot().RemoveCSINode(nodeName)
		}
	}
	return nil
}

// SchedulePod adds pod to the snapshot and schedules it to given node.
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

	if err := s.ClusterSnapshotStore.ForceAddPod(pod, nodeName); err != nil {
		return clustersnapshot.NewSchedulingInternalError(pod, err.Error())
	}
	return nil
}

// SchedulePodOnAnyNodeMatching adds pod to the snapshot and schedules it to any node matching the provided function.
func (s *PredicateSnapshot) SchedulePodOnAnyNodeMatching(pod *apiv1.Pod, anyNodeMatching func(*framework.NodeInfo) bool) (string, clustersnapshot.SchedulingError) {
	node, cycleState, schedErr := s.pluginRunner.RunFiltersUntilPassingNode(pod, anyNodeMatching)
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

	if err := s.ClusterSnapshotStore.ForceAddPod(pod, node.Name); err != nil {
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
			if err := s.ClusterSnapshotStore.DraSnapshot().UnreservePodClaims(foundPod); err != nil {
				return err
			}
		}
	}

	return s.ClusterSnapshotStore.ForceRemovePod(namespace, podName, nodeName)
}

// CheckPredicates checks whether scheduler predicates pass for the given pod on the given node.
func (s *PredicateSnapshot) CheckPredicates(pod *apiv1.Pod, nodeName string) clustersnapshot.SchedulingError {
	_, _, err := s.pluginRunner.RunFiltersOnNode(pod, nodeName)
	return err
}

// verifyScheduledPodResourceClaims verifies that all needed claims are tracked in the DRA snapshot, allocated, and available on the Node.
func (s *PredicateSnapshot) verifyScheduledPodResourceClaims(pod *apiv1.Pod, node *apiv1.Node) error {
	claims, err := s.ClusterSnapshotStore.DraSnapshot().PodClaims(pod)
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

func (s *PredicateSnapshot) modifyResourceClaimsForScheduledPod(pod *apiv1.Pod, node *apiv1.Node, postFilterState *schedulerimpl.CycleState) error {
	// We need to run the scheduler Reserve phase to allocate the appropriate ResourceClaims in the DRA snapshot. The allocations are
	// actually computed and cached in the Filter phase, and Reserve only grabs them from the cycle state. So this should be quick, but
	// it needs the cycle state from after running the Filter phase.
	if err := s.pluginRunner.RunReserveOnNode(pod, node.Name, postFilterState); err != nil {
		return fmt.Errorf("error while trying to run Reserve node %s for pod %s/%s: %v", node.Name, pod.Namespace, pod.Name, err)
	}

	// The pod isn't added to the ReservedFor field of the claim during the Reserve phase (it happens later, in PreBind). We can just do it
	// manually here. It shouldn't fail, it only fails if ReservedFor is at max length already, but that is checked during the Filter phase.
	if err := s.ClusterSnapshotStore.DraSnapshot().ReservePodClaims(pod); err != nil {
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
	if err := s.ClusterSnapshotStore.DraSnapshot().AddClaims(podOwnedClaims); err != nil {
		return fmt.Errorf("couldn't add ResourceSlices for pod %s/%s to DRA snapshot: %v", podInfo.Namespace, podInfo.Name, err)
	}

	// The Pod-owned claims should already be reserved for the Pod after sanitization, but we need to add the reservation for the new Pod
	// to the shared claims. This can fail if doing so would exceed the max reservation limit for a claim.
	if err := s.ClusterSnapshotStore.DraSnapshot().ReservePodClaims(podInfo.Pod); err != nil {
		return fmt.Errorf("couldn't add pod %s/%s reservations to claims, this shouldn't happen: %v", podInfo.Namespace, podInfo.Name, err)
	}
	return nil
}
