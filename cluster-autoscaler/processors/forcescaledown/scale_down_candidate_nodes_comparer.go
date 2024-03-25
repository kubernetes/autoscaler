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

package forcescaledown

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	"k8s.io/klog/v2"
)

// ScaleDownCandidateNodesCompare is sorting scale down candidates so that force-scale-down nodes appear first.
type ScaleDownCandidateNodesCompare struct{}

// NewScaleDownCandidateNodesCompare return ScaleDownCandidateNodesCompare struct.
func NewScaleDownCandidateNodesCompare() *ScaleDownCandidateNodesCompare {
	return &ScaleDownCandidateNodesCompare{}
}

// ScaleDownEarlierThan return true if node1 has force-scale-down taint and node2 doesn't.
func (p *ScaleDownCandidateNodesCompare) ScaleDownEarlierThan(node1, node2 *apiv1.Node) bool {
	if taints.HasForceScaleDownTaint(node1) && !taints.HasForceScaleDownTaint(node2) {
		klog.V(4).Infof("Prioritizing node %s over %s, because only %s has the force-scale-down taint", node1.Name, node2.Name, node1.Name)
		return true
	}
	return false
}
