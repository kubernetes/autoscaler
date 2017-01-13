/*
Copyright 2016 The Kubernetes Authors.

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

package cloudprovider

import (
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
)

// CloudProvider contains configuration info and functions for interacting with
// cloud provider (GCE, AWS, etc).
type CloudProvider interface {
	// Name returns name of the cloud provider.
	Name() string

	// NodeGroups returns all node groups configured for this cloud provider.
	NodeGroups() []NodeGroup

	// NodeGroupForNode returns the node group for the given node, nil if the node
	// should not be processed by cluster autoscaler, or non-nil error if such
	// occurred.
	NodeGroupForNode(*apiv1.Node) (NodeGroup, error)
}

// NodeGroup contains configuration info and functions to control a set
// of nodes that have the same capacity and set of labels.
type NodeGroup interface {
	// MaxSize returns maximum size of the node group.
	MaxSize() int

	// MinSize returns minimum size of the node group.
	MinSize() int

	// TargetSize returns the current target size of the node group. It is possible that the
	// number of nodes in Kubernetes is different at the moment but should be equal
	// to Size() once everything stabilizes (new nodes finish startup and registration or
	// removed nodes are deleted completely)
	TargetSize() (int, error)

	// IncreaseSize increases the size of the node group. To delete a node you need
	// to explicitly name it and use DeleteNode. This function should wait until
	// node group size is updated.
	IncreaseSize(delta int) error

	// DeleteNodes deletes nodes from this node group. Error is returned either on
	// failure or if the given node doesn't belong to this node group. This function
	// should wait until node group size is updated.
	DeleteNodes([]*apiv1.Node) error

	// DecreaseTargetSize decreases the target size of the node group. This function
	// doesn't permit to delete any existing node and can be used only to reduce the
	// request for new nodes that have not been yet fulfilled. Delta should be negative.
	// It is assumed that cloud provider will not delete the existing nodes if the size
	// when there is an option to just decrease the target.
	DecreaseTargetSize(delta int) error

	// Id returns an unique identifier of the node group.
	Id() string

	// Debug returns a string containing all information regarding this node group.
	Debug() string

	// Nodes returns a list of all nodes that belong to this node group.
	Nodes() ([]string, error)
}
