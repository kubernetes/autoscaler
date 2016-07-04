/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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
	kube_api "k8s.io/kubernetes/pkg/api"
)

// CloudProvider contains configuration info and functions for interacting with
// cloud provider (GCE, AWS, etc).
type CloudProvider interface {
	// Name returns name of the cloud provider.
	Name() string

	// NodeGroups returns all node groups configured for this cloud provider.
	NodeGroups() []*NodeGroup

	// NodeGroupForNode returns the node group for the given node, nil if the node
	// should not be processed by cluster autoscaler, or non-nil error if such
	// occurred.
	NodeGroupForNode(*kube_api.Node) (*NodeGroup, error)
}

// NodeGroup contains configuration info and functions to control a set
// of nodes that have the same capacity and set of labels.
type NodeGroup interface {
	// MaxSize returns maximum size of the node group.
	MaxSize() int

	// MinSize returns minimum size of the node group.
	MinSize() int

	// Size returns the current size of the node group.
	Size() (int, error)

	// GetSampleNode returns a sample node that belongs to this node group, or error
	// if occurred.
	SampleNode() (*kube_api.Node, error)

	// IncreaseSize increases the size of the node group. To delete a node you need
	// to explicitly name it and use DeleteNode. This function should wait until
	// the node appears in the Kubernetes cluster.
	IncreaseSize(delta int) error

	// DeleteNode deletes the node from this node group. Error is returned either on
	// failure or if the given node doesn't belong to this node group. This function
	// should wait until the node is removed from the Kubernetes cluster.
	DeleteNode(*kube_api.Node) error

	// Id returns an unique identifier of the node group.
	Id() string

	// Debug returns a string containing all information regarding this node group.
	Debug() string
}
