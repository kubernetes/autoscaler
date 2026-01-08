/*
Copyright 2020 The Kubernetes Authors.

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

package clustersnapshot

import (
	"errors"

	apiv1 "k8s.io/api/core/v1"
	csisnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/csi/snapshot"
	drasnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"
)

// ClusterSnapshot is abstraction of cluster state used for predicate simulations.
// It exposes mutation methods and can be viewed as scheduler's SharedLister.
type ClusterSnapshot interface {
	ClusterSnapshotStore

	// AddNodeInfo adds the given NodeInfo to the snapshot without checking scheduler predicates. The Node and the Pods are added,
	// as well as any DRA objects passed along them.
	AddNodeInfo(nodeInfo *framework.NodeInfo) error
	// RemoveNodeInfo removes the given NodeInfo from the snapshot The Node and the Pods are removed, as well as
	// any DRA objects owned by them.
	RemoveNodeInfo(nodeName string) error
	// GetNodeInfo returns an internal NodeInfo for a given Node - all information about the Node tracked in the snapshot.
	// This means the Node itself, its scheduled Pods, as well as all relevant DRA objects. The internal NodeInfos
	// obtained via this method should always be used in CA code instead of directly using *schedulerframework.NodeInfo.
	GetNodeInfo(nodeName string) (*framework.NodeInfo, error)
	// ListNodeInfos returns internal NodeInfos for all Nodes tracked in the snapshot. See the comment on GetNodeInfo.
	ListNodeInfos() ([]*framework.NodeInfo, error)

	// ForceAddPod adds the given Pod to the Node with the given nodeName inside the snapshot without checking scheduler predicates.
	// This method will allocate internal PodInfo and include all the DRA-related information (taken from the DRA snapshot).
	// It will store it to NodeInfo via StorePodInfo.
	ForceAddPod(pod *apiv1.Pod, nodeName string) error
	// ForceRemovePod removes the given Pod (and all DRA objects it owns) from the snapshot.
	ForceRemovePod(namespace string, podName string, nodeName string) error

	// SchedulePod tries to schedule the given Pod on the Node with the given name inside the snapshot,
	// checking scheduling predicates. The pod is only scheduled if the predicates pass. If the pod is scheduled,
	// all relevant DRA objects are modified to reflect that. Returns nil if the pod got scheduled, and a non-nil
	// error explaining why not otherwise. The error Type() can be checked against SchedulingInternalError to distinguish
	// failing predicates from unexpected errors.
	SchedulePod(pod *apiv1.Pod, nodeName string) SchedulingError
	// SchedulePodOnAnyNodeMatching tries to schedule the given Pod on any Node for which nodeMatches returns
	// true. Scheduling predicates are checked, and the pod is scheduled only if there is a matching Node with passing
	// predicates. If the pod is scheduled, all relevant DRA objects are modified to reflect that, and the name of the
	// Node its scheduled on and nil are returned. If the pod can't be scheduled on any Node, an empty string and a non-nil
	// error explaining why are returned. The error Type() can be checked against SchedulingInternalError to distinguish
	// failing predicates from unexpected errors.
	SchedulePodOnAnyNodeMatching(pod *apiv1.Pod, nodeMatches func(*framework.NodeInfo) bool) (matchingNode string, err SchedulingError)
	// UnschedulePod removes the given Pod from the given Node inside the snapshot, and modifies all relevant DRA objects
	// to reflect the removal. The pod can then be scheduled on another Node in the snapshot using the Schedule methods.
	UnschedulePod(namespace string, podName string, nodeName string) error

	// CheckPredicates runs scheduler predicates to check if the given Pod would be able to schedule on the Node with the given
	// name. Returns nil if predicates pass, or a non-nil error specifying why they didn't otherwise. The error Type() can be
	// checked against SchedulingInternalError to distinguish failing predicates from unexpected errors. Doesn't mutate the snapshot.
	CheckPredicates(pod *apiv1.Pod, nodeName string) SchedulingError

	// TODO(DRA): Move unschedulable Pods inside ClusterSnapshot (since their DRA objects are already here), refactor PodListProcessor.
}

// ClusterSnapshotStore is the "low-level" part of ClusterSnapshot, responsible for storing the snapshot state and mutating it directly,
// without going through scheduler predicates. ClusterSnapshotStore shouldn't be directly used outside the clustersnapshot pkg, its methods
// should be accessed via ClusterSnapshot.
type ClusterSnapshotStore interface {
	framework.SharedLister

	// SetClusterState resets the snapshot to an unforked state and replaces the contents of the snapshot
	// with the provided data. scheduledPods are correlated to their Nodes based on spec.NodeName.
	SetClusterState(nodes []*apiv1.Node, scheduledPods []*apiv1.Pod, draSnapshot *drasnapshot.Snapshot, csiSnapshot *csisnapshot.Snapshot) error

	// StorePodInfo adds the given PodInfo to the Node with the given nodeName inside the snapshot.
	StorePodInfo(podInfo *framework.PodInfo, nodeName string) error
	// RemovePodInfo removes the given Pod from the snapshot.
	RemovePodInfo(namespace string, podName string, nodeName string) error

	// StoreNodeInfo adds the given NodeInfo to the snapshot without checking scheduler predicates, and
	// without taking DRA objects into account. It expects the DRA objects are already configured for the given NodeInfo.
	// This shouldn't be used outside the clustersnapshot pkg, use ClusterSnapshot.AddNodeInfo() instead.
	StoreNodeInfo(nodeInfo *framework.NodeInfo) error
	// RemoveNodeInfo removes the given NodeInfo from the snapshot without deallocating DRA objects from the DraSnapshot.
	// This shouldn't be used outside the clustersnapshot pkg, use ClusterSnapshot.RemoveNodeInfo() instead.
	RemoveNodeInfo(nodeName string) error

	// DraSnapshot returns an interface that allows accessing and modifying the DRA objects in the snapshot.
	DraSnapshot() *drasnapshot.Snapshot

	// CsiSnapshot returns an interface that allows accessing and modifying the CSINode objects in the snapshot.
	CsiSnapshot() *csisnapshot.Snapshot

	// Fork creates a fork of snapshot state. All modifications can later be reverted to moment of forking via Revert().
	// Use WithForkedSnapshot() helper function instead if possible.
	Fork()
	// Revert reverts snapshot state to moment of forking.
	Revert()
	// Commit commits changes done after forking.
	Commit() error
}

// ErrNodeNotFound means that a node wasn't found in the snapshot.
var ErrNodeNotFound = errors.New("node not found")

// WithForkedSnapshot is a helper function for snapshot that makes sure all Fork() calls are closed with Commit() or Revert() calls.
// The function return (error, error) pair. The first error comes from the passed function, the second error indicate the success of the function itself.
func WithForkedSnapshot(snapshot ClusterSnapshot, f func() (bool, error)) (error, error) {
	var commit bool
	var err, cleanupErr error
	snapshot.Fork()
	defer func() {
		if commit {
			cleanupErr = snapshot.Commit()
			if cleanupErr != nil {
				klog.Errorf("Got error when calling ClusterSnapshot.Commit(), will try to revert; %v", cleanupErr)
			}
		}
		if !commit || cleanupErr != nil {
			snapshot.Revert()
		}
	}()
	commit, err = f()
	return err, cleanupErr
}
