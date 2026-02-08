/*
Copyright 2026 The Kubernetes Authors.

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

package skipreschedule

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

const (
	// SkipRescheduleAnnotationKey is an annotation that indicates the pod should be
	// skipped during rescheduling simulation. Pods with this annotation:
	// - Are evictable (can be terminated during scale-down)
	// - Receive SkipDrain status (not simulated for rescheduling to other nodes)
	// - Are NOT included in template nodes (unlike daemonset-pod annotation)
	//
	// Use case: Ephemeral pods that don't need to survive node deletion and don't
	// need destination simulation. The pod will simply be terminated when the node
	// is removed.
	SkipRescheduleAnnotationKey = "cluster-autoscaler.kubernetes.io/skip-reschedule"
)

// Rule is a drainability rule on how to handle pods that should skip rescheduling simulation.
type Rule struct{}

// New creates a new Rule.
func New() *Rule {
	return &Rule{}
}

// Name returns the name of the rule.
func (r *Rule) Name() string {
	return "SkipReschedule"
}

// Drainable decides what to do with skip-reschedule pods on node drain.
// Pods with the skip-reschedule annotation will receive SkipDrain status,
// meaning they won't block drain and won't be simulated for rescheduling.
func (Rule) Drainable(drainCtx *drainability.DrainContext, pod *apiv1.Pod, _ *framework.NodeInfo) drainability.Status {
	if HasSkipRescheduleAnnotation(pod) {
		return drainability.NewSkipStatus()
	}
	return drainability.NewUndefinedStatus()
}

// HasSkipRescheduleAnnotation returns true if the pod has the skip-reschedule annotation set to "true".
func HasSkipRescheduleAnnotation(pod *apiv1.Pod) bool {
	return pod.Annotations[SkipRescheduleAnnotationKey] == "true"
}
