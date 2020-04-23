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

package simulator

import (
	"errors"

	apiv1 "k8s.io/api/core/v1"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
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

	// Fork creates a fork of snapshot state. All modifications can later be reverted to moment of forking via Revert()
	// Forking already forked snapshot is not allowed and will result with an error.
	Fork() error
	// Revert reverts snapshot state to moment of forking.
	Revert() error
	// Commit commits changes done after forking.
	Commit() error
	// Clear reset cluster snapshot to empty, unforked state
	Clear()
}

var errNodeNotFound = errors.New("node not found")
