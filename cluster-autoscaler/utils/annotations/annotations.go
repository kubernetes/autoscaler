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

package annotations

import apiv1 "k8s.io/api/core/v1"

const (
	// NodeUpcomingAnnotation is an annotation CA adds to nodes which are upcoming.
	NodeUpcomingAnnotation = "cluster-autoscaler.k8s.io/upcoming-node"

	// NodeSalvoAnnotation is an annotation CA adds to upcoming nodes injected during a Salvo scale-up loop.
	NodeSalvoAnnotation = "cluster-autoscaler.k8s.io/salvo-node"

	// PodScaleUpDelayAnnotationKey is an annotation how long pod can wait to be scaled up.
	PodScaleUpDelayAnnotationKey = "cluster-autoscaler.kubernetes.io/pod-scale-up-delay"
)

// IsSalvoNode returns true if the node was injected during an iterative Salvo scale-up loop.
func IsSalvoNode(node *apiv1.Node) bool {
	if node == nil || node.Annotations == nil {
		return false
	}
	return node.Annotations[NodeSalvoAnnotation] == "true"
}

// IsUpcomingNode returns true if the node is marked as upcoming in the cluster snapshot.
func IsUpcomingNode(node *apiv1.Node) bool {
	if node == nil || node.Annotations == nil {
		return false
	}
	return node.Annotations[NodeUpcomingAnnotation] == "true"
}
