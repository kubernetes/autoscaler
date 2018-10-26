/*
Copyright 2018 The Kubernetes Authors.

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

package nodegroupset

import (
	schedulercache "k8s.io/kubernetes/pkg/scheduler/cache"
)

// GkeNodepoolLabel is a label specifying GKE node pool particular node belongs to.
const GkeNodepoolLabel = "cloud.google.com/gke-nodepool"

func nodesFromSameGkeNodePool(n1, n2 *schedulercache.NodeInfo) bool {
	n1GkeNodePool := n1.Node().Labels[GkeNodepoolLabel]
	n2GkeNodePool := n2.Node().Labels[GkeNodepoolLabel]
	return n1GkeNodePool != "" && n1GkeNodePool == n2GkeNodePool
}

// IsGkeNodeInfoSimilar compares if two nodes should be considered part of the
// same NodeGroupSet. This is true if they either belong to the same GKE nodepool
// or match usual conditions checked by IsNodeInfoSimilar.
func IsGkeNodeInfoSimilar(n1, n2 *schedulercache.NodeInfo) bool {
	if nodesFromSameGkeNodePool(n1, n2) {
		return true
	}
	return IsNodeInfoSimilar(n1, n2)
}
