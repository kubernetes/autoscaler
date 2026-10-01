/*
Copyright The Kubernetes Authors.

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

package pressure

import (
	"context"
	"slices"

	corev1 "k8s.io/api/core/v1"
	listersv1 "k8s.io/client-go/listers/core/v1"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	resourcehelpers "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/resources"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

// Observer reports sustained memory stall for eligible containers.
type Observer interface {
	// Run polls until ctx is done.
	Run(ctx context.Context)
	// SetTargets replaces the set of containers to observe.
	SetTargets(targets []Target)
	// Channel returns accepted pressure observations.
	Channel() <-chan Info
	// Eligible reports whether an observation's container is still in the current target set, so one
	// queued before an opt-out or a Pod replacement is discarded.
	Eligible(info Info) bool
}

// BuildTargets returns the running containers whose VPA opts in to pressure detection. It reads full
// Pods from the lister because BasicPodSpec carries no UID, node name, statuses or conditions. VPAs
// without an opted-in container policy are skipped before their Pods are listed.
func BuildTargets(clusterState model.ClusterState, pods listersv1.PodLister) []Target {
	var targets []Target
	for _, vpa := range clusterState.VPAs() {
		if !anyPressureOptIn(vpa) {
			continue
		}
		for _, podID := range clusterState.GetMatchingPods(vpa) {
			pod, err := pods.Pods(podID.Namespace).Get(podID.PodName)
			if err != nil || pod.Status.Phase != corev1.PodRunning || pod.Spec.NodeName == "" {
				continue
			}
			resizeState := resizeStateOf(pod)
			for _, c := range pod.Spec.Containers {
				if !vpa_api_util.IsPressureDetectionEnabled(c.Name, vpa.ResourcePolicy) ||
					!managesMemory(vpa_api_util.GetContainerResourcePolicy(c.Name, vpa.ResourcePolicy)) {
					continue
				}
				_, limits := resourcehelpers.ContainerRequestsAndLimits(c.Name, pod)
				targets = append(targets, Target{
					ContainerID: model.ContainerID{PodID: podID, ContainerName: c.Name},
					PodUID:      string(pod.UID),
					Node:        pod.Spec.NodeName,
					LimitBytes:  limits.Memory().Value(),
					ResizeState: resizeState,
				})
			}
		}
	}
	return targets
}

func anyPressureOptIn(vpa *model.Vpa) bool {
	if vpa.ResourcePolicy == nil {
		return false
	}
	for _, p := range vpa.ResourcePolicy.ContainerPolicies {
		if p.PressureDetection != nil && *p.PressureDetection == vpa_types.PressureDetectionEnabled {
			return true
		}
	}
	return false
}

// managesMemory reports whether a container policy lets VPA recommend memory, with the defaults
// AggregateContainerState.UpdateFromPolicy applies: scaling mode Auto and model.DefaultControlledResources.
func managesMemory(policy *vpa_types.ContainerResourcePolicy) bool {
	if policy != nil && policy.Mode != nil && *policy.Mode == vpa_types.ContainerScalingModeOff {
		return false
	}
	controlled := model.DefaultControlledResources
	if policy != nil && policy.ControlledResources != nil {
		controlled = *model.ResourceNamesApiToModel(*policy.ControlledResources)
	}
	return slices.Contains(controlled, model.ResourceMemory)
}

// resizeStateOf returns the Pod's outstanding resize state, or "" when none. Conditions count only when
// True, matching the updater's in-place restriction.
func resizeStateOf(pod *corev1.Pod) string {
	for _, c := range pod.Status.Conditions {
		if c.Status != corev1.ConditionTrue {
			continue
		}
		switch {
		case c.Type == corev1.PodResizeInProgress:
			return "resize_in_progress"
		case c.Type == corev1.PodResizePending && c.Reason == corev1.PodReasonInfeasible:
			return "resize_infeasible"
		case c.Type == corev1.PodResizePending:
			return "resize_deferred"
		}
	}
	return ""
}
