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
	resourceapi "k8s.io/api/resource/v1alpha3"
	"k8s.io/autoscaler/cluster-autoscaler/dynamicresources"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// ClusterSnapshot is abstraction of cluster state used for predicate simulations.
// It exposes mutation methods and can be viewed as scheduler's SharedLister.
type ClusterSnapshot interface {
	framework.SharedLister

	// Initialize clears the snapshot and initializes it with real objects from the cluster - Nodes,
	// scheduled pods, and all DRA objects (including e.g. ResourceClaims referenced by unschedulable Pods,
	// or non-Node-local ResourceSlices).
	Initialize(nodes []*apiv1.Node, scheduledPods []*apiv1.Pod, draSnapshot dynamicresources.Snapshot) error

	// SchedulePod schedules the given Pod onto the Node with the given nodeName inside the snapshot. If reserveState is passed,
	// and the Pod references ResourceClaims, the Reserve phase of the scheduler framework is run in order to allocate the
	// claims inside the snapshot. Returns an error if the pod references a ResourceClaim that isn't tracked in the snapshot, and
	// allocated to the given Node.
	SchedulePod(pod *apiv1.Pod, nodeName string, reserveState *schedulerframework.CycleState) error
	// UnschedulePod removes the given Pod from the given Node inside the snapshot. The ResourceClaims referenced by the Pod are
	// deallocated and left in the snapshot, so that the Pod can be scheduled on another Node using SchedulePod.
	UnschedulePod(namespace string, podName string, nodeName string) error

	// AddNodeInfo adds the given NodeInfo to the snapshot. The Node and the Pods are added, as well as
	// any DRA objects passed along them.
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

	// AddResourceClaims adds additional ResourceClaims to the snapshot. It can be used e.g. if we need to duplicate a Pod that
	// owns ResourceClaims. Returns an error if any of the claims is already tracked in the snapshot.
	AddResourceClaims(extraClaims []*resourceapi.ResourceClaim) error
	// GetPodResourceClaims returns all ResourceClaims referenced by the given pod. It can be used to retrieve ResourceClaims
	// for pods that aren't scheduled. Returns an error if any of the claims referenced by the pod aren't tracked in the snapshot.
	GetPodResourceClaims(pod *apiv1.Pod) ([]*resourceapi.ResourceClaim, error)

	// Fork creates a fork of snapshot state. All modifications can later be reverted to moment of forking via Revert().
	// Use WithForkedSnapshot() helper function instead if possible.
	Fork()
	// Revert reverts snapshot state to moment of forking.
	Revert()
	// Commit commits changes done after forking.
	Commit() error
	// Clear reset cluster snapshot to empty, unforked state.
	Clear()
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
