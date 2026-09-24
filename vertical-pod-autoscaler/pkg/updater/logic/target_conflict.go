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

package logic

import (
	"fmt"
	"slices"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

const targetConflictReason = "MultipleActiveVPATargetsFound"
const noTargetConflictReason = "NoConflictingActiveVPATargets"

// reconcileTargetConflicts groups the given VPAs by their targetRef and, within
// each group of active VPAs (update mode other than "Off"), sets the
// TargetConflict status condition to True on every VPA that overlaps with
// another one, meaning both control the same resource of the same container.
// VPAs that control disjoint containers or resources do not conflict. VPAs
// that no longer conflict but previously had the condition set to True have
// it cleared.
//
// This surfaces the conflict to the user; it does not change which VPA
// actually controls a pod (see vpa_api_util.Stronger for that logic).
func (u *updater) reconcileTargetConflicts(vpaList []*vpa_types.VerticalPodAutoscaler) {
	vpaGroups := make(map[string][]*vpa_types.VerticalPodAutoscaler, len(vpaList))
	var ineligible []*vpa_types.VerticalPodAutoscaler
	for _, vpa := range vpaList {
		if slices.Contains(u.ignoredNamespaces, vpa.Namespace) {
			continue
		}
		if vpa.Spec.TargetRef == nil {
			ineligible = append(ineligible, vpa)
			continue
		}
		// "Off" VPAs don't update pods at all, so they can't conflict.
		if vpa_api_util.GetUpdateMode(vpa) == vpa_types.UpdateModeOff {
			ineligible = append(ineligible, vpa)
			continue
		}
		key := vpa_api_util.TargetRefIndexKey(vpa.Namespace, vpa.Spec.TargetRef.Kind, vpa.Spec.TargetRef.Name)
		vpaGroups[key] = append(vpaGroups[key], vpa)
	}

	for _, group := range vpaGroups {
		for _, vpa := range group {
			names := conflictingVpaNames(vpa, group)
			u.updateTargetConflictCondition(vpa, len(names) > 1, names)
		}
	}

	// VPAs that became ineligible (e.g. switched to Off, or dropped their
	// targetRef) may still be carrying a stale TargetConflict=True condition
	// from a previous run. Clear it since they can no longer be part of a
	// conflict. updateTargetConflictCondition is a no-op if there was never
	// a condition to begin with.
	for _, vpa := range ineligible {
		u.updateTargetConflictCondition(vpa, false, nil)
	}
}

// updateTargetConflictCondition patches the TargetConflict condition on a
// single VPA if it needs to change, and emits a Warning/Normal event only on
// a state transition (conflict appearing or clearing).
func (u *updater) updateTargetConflictCondition(vpa *vpa_types.VerticalPodAutoscaler, conflicting bool, conflictingNames []string) {
	oldStatus := vpa.Status.DeepCopy()

	wasConflicting := false
	if c := getVpaCondition(oldStatus, vpa_types.TargetConflict); c != nil {
		wasConflicting = c.Status == corev1.ConditionTrue
	}

	// Nothing to do: no conflict now and none previously recorded.
	if !conflicting && !wasConflicting {
		return
	}

	// transitioned is true only when the conflict appears or clears. While a
	// conflict persists we still rebuild the condition, because the set of
	// conflicting VPAs (and so the message) can change without a transition;
	// UpdateVpaStatusIfNeeded skips the patch when nothing differs.
	transitioned := conflicting != wasConflicting

	condStatus := corev1.ConditionFalse
	reason := noTargetConflictReason
	message := "No conflicting active VPAs found for this target."
	if conflicting {
		condStatus = corev1.ConditionTrue
		reason = targetConflictReason
		message = fmt.Sprintf("Conflict: multiple active VPAs target the same resource: %s", strings.Join(conflictingNames, ", "))
	}

	newStatus := oldStatus.DeepCopy()
	setVpaCondition(newStatus, vpa_types.TargetConflict, condStatus, reason, message, vpa.Generation)

	_, err := vpa_api_util.UpdateVpaStatusIfNeeded(
		u.vpaClient.AutoscalingV1().VerticalPodAutoscalers(vpa.Namespace),
		vpa.Name,
		newStatus,
		oldStatus,
	)
	if err != nil {
		klog.ErrorS(err, "Failed to update VPA TargetConflict condition", "vpa", klog.KObj(vpa))
		return
	}

	// Only emit an event on a state transition, not on message-only updates.
	if !transitioned {
		return
	}

	eventType := corev1.EventTypeWarning
	if !conflicting {
		eventType = corev1.EventTypeNormal
	}
	u.eventRecorder.Event(vpa, eventType, reason, message)
}

// getVpaCondition returns a pointer to the condition of the given type, or
// nil if it isn't present.
func getVpaCondition(status *vpa_types.VerticalPodAutoscalerStatus, condType vpa_types.VerticalPodAutoscalerConditionType) *vpa_types.VerticalPodAutoscalerCondition {
	for i := range status.Conditions {
		if status.Conditions[i].Type == condType {
			return &status.Conditions[i]
		}
	}
	return nil
}

// setVpaCondition sets or updates a condition on the given status, bumping
// LastTransitionTime only when the status value actually changes.
//
// observedGeneration is the VPA's .metadata.generation at the time this
// condition was computed. It is recorded on the condition itself rather than
// on .status.observedGeneration, which is owned by the recommender.
func setVpaCondition(status *vpa_types.VerticalPodAutoscalerStatus, condType vpa_types.VerticalPodAutoscalerConditionType, condStatus corev1.ConditionStatus, reason, message string, observedGeneration int64) {
	now := metav1.Now()
	for i := range status.Conditions {
		if status.Conditions[i].Type == condType {
			if status.Conditions[i].Status != condStatus {
				status.Conditions[i].LastTransitionTime = now
			}
			status.Conditions[i].Status = condStatus
			status.Conditions[i].Reason = reason
			status.Conditions[i].Message = message
			status.Conditions[i].ObservedGeneration = observedGeneration
			return
		}
	}
	status.Conditions = append(status.Conditions, vpa_types.VerticalPodAutoscalerCondition{
		Type:               condType,
		Status:             condStatus,
		LastTransitionTime: now,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: observedGeneration,
	})
}

// conflictingVpaNames returns the sorted names of vpa plus every other VPA in
// group whose controlled (container, resource) set overlaps with vpa's. If
// nothing overlaps, only vpa's own name is returned.
func conflictingVpaNames(vpa *vpa_types.VerticalPodAutoscaler, group []*vpa_types.VerticalPodAutoscaler) []string {
	names := []string{vpa.Name}
	for _, other := range group {
		if other != vpa && vpasOverlap(vpa, other) {
			names = append(names, other.Name)
		}
	}
	// Sorted so the condition message and event text are stable across runs.
	slices.Sort(names)
	return names
}

// vpasOverlap reports whether two VPAs would both manage at least one resource
// of the same container. The wildcard policy stands in for every container
// that has no explicit policy in either VPA.
func vpasOverlap(a, b *vpa_types.VerticalPodAutoscaler) bool {
	containers := map[string]struct{}{vpa_types.DefaultContainerResourcePolicy: {}}
	for _, v := range []*vpa_types.VerticalPodAutoscaler{a, b} {
		if v.Spec.ResourcePolicy == nil {
			continue
		}
		for _, p := range v.Spec.ResourcePolicy.ContainerPolicies {
			containers[p.ContainerName] = struct{}{}
		}
	}
	for name := range containers {
		resourcesB := controlledResourcesFor(b, name)
		for r := range controlledResourcesFor(a, name) {
			if _, ok := resourcesB[r]; ok {
				return true
			}
		}
	}
	return false
}

// controlledResourcesFor returns the resources the VPA manages for the given
// container: none if the container policy mode is Off, the policy's
// controlledResources if set, otherwise CPU and memory.
func controlledResourcesFor(vpa *vpa_types.VerticalPodAutoscaler, containerName string) map[corev1.ResourceName]struct{} {
	policy := vpa_api_util.GetContainerResourcePolicy(containerName, vpa.Spec.ResourcePolicy)
	if policy != nil && policy.Mode != nil && *policy.Mode == vpa_types.ContainerScalingModeOff {
		return nil
	}
	resources := []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory}
	if policy != nil && policy.ControlledResources != nil {
		resources = *policy.ControlledResources
	}
	set := make(map[corev1.ResourceName]struct{}, len(resources))
	for _, r := range resources {
		set[r] = struct{}{}
	}
	return set
}
