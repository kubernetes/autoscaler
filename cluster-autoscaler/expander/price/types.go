/*
Copyright 2017 The Kubernetes Authors.

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

package price

import (
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
)

// PreferredNodeProvider provides a node that would be, in-longer run, the most suited to the cluster
// needs.
type PreferredNodeProvider interface {
	// Node returns the preferred node for the cluster
	Node() (*apiv1.Node, error)
}

// NodeUnfitness is a function that returns a value that how much the evaluated node is "different"
// from the preferred node for the cluster. The value should be greater or equal 1 where 1 means that
// the node is perfect. 10 means that the node is very unfitted, although there is no upper bound on the
// value of this function.
type NodeUnfitness func(preferredNode, evaluatedNode *apiv1.Node) float64
