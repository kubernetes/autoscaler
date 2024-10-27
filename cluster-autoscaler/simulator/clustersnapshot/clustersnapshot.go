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
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// ClusterSnapshot is abstraction of cluster state used for predicate simulations.
// It exposes mutation methods and can be viewed as scheduler's SharedLister.
type ClusterSnapshot interface {
	schedulerframework.SharedLister
	// AddNode adds node to the snapshot.
	AddNode(node *apiv1.Node) error
	// AddNodes adds nodes to the snapshot.
	AddNodes(nodes []*apiv1.Node) error
	// RemoveNode removes nodes (and pods scheduled to it) from the snapshot.
	RemoveNode(nodeName string) error
	// AddPod adds pod to the snapshot and schedules it to given node.
	AddPod(pod *apiv1.Pod, nodeName string) error
	// RemovePod removes pod from the snapshot.
	RemovePod(namespace string, podName string, nodeName string) error
	// AddNodeWithPods adds a node and set of pods to be scheduled to this node to the snapshot.
	AddNodeWithPods(node *apiv1.Node, pods []*apiv1.Pod) error
	// IsPVCUsedByPods returns if the pvc is used by any pod, key = <namespace>/<pvc_name>
	IsPVCUsedByPods(key string) bool

	// AddNodeInfo adds the given NodeInfo to the snapshot. The Node and the Pods are added, as well as
	// any DRA objects passed along them.
	AddNodeInfo(nodeInfo *framework.NodeInfo) error
	// GetNodeInfo returns an internal NodeInfo for a given Node - all information about the Node tracked in the snapshot.
	// This means the Node itself, its scheduled Pods, as well as all relevant DRA objects. The internal NodeInfos
	// obtained via this method should always be used in CA code instead of directly using *schedulerframework.NodeInfo.
	GetNodeInfo(nodeName string) (*framework.NodeInfo, error)
	// ListNodeInfos returns internal NodeInfos for all Nodes tracked in the snapshot. See the comment on GetNodeInfo.
	ListNodeInfos() ([]*framework.NodeInfo, error)

	// Fork creates a fork of snapshot state. All modifications can later be reverted to moment of forking via Revert().
	// Use WithForkedSnapshot() helper function instead if possible.
	Fork()
	// Revert reverts snapshot state to moment of forking.
	Revert()
	// Commit commits changes done after forking.
	Commit() error
	// Clear reset cluster snapshot to empty, unforked state.
	Clear()
	// Export returns a shallow copy of the snapshot.
	Export() ClusterSnapshot
	// Rebase rebases the snapshot to a new base snapshot.
	Rebase(base ClusterSnapshot) error
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
