/*
Copyright 2025 The Kubernetes Authors.

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

// reconcileTargetConflicts groups the given VPAs by their targetRef and, for
// any target with more than one VPA using an active update mode (anything
// other than "Off"), sets the TargetConflict status condition to True on
// each conflicting VPA. VPAs that are no longer part of a conflicting group
// but previously had the condition set to True have it cleared.
//
// This surfaces the conflict to the user; it does not change which VPA
// actually controls a pod (see vpa_api_util.Stronger for that logic).
func (u *updater) reconcileTargetConflicts(vpaList []*vpa_types.VerticalPodAutoscaler) {
	groups := make(map[string][]*vpa_types.VerticalPodAutoscaler)
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
		groups[key] = append(groups[key], vpa)
	}

	for _, group := range groups {
		conflicting := len(group) > 1
		var message string
		if conflicting {
			names := make([]string, 0, len(group))
			for _, vpa := range group {
				names = append(names, vpa.Name)
			}
			slices.Sort(names)
			message = fmt.Sprintf("Conflict: multiple active VPAs target the same object: %s", strings.Join(names, ", "))
		}
		for _, vpa := range group {
			u.updateTargetConflictCondition(vpa, conflicting, message)
		}
	}

	// VPAs that became ineligible (e.g. switched to Off, or dropped their
	// targetRef) may still be carrying a stale TargetConflict=True condition
	// from a previous run. Clear it since they can no longer be part of a
	// conflict. updateTargetConflictCondition is a no-op if there was never
	// a condition to begin with.
	for _, vpa := range ineligible {
		u.updateTargetConflictCondition(vpa, false, "")
	}
}

// updateTargetConflictCondition patches the TargetConflict condition on a
// single VPA if it needs to change, and emits a Warning/Normal event only on
// a state transition (conflict appearing or clearing).
func (u *updater) updateTargetConflictCondition(vpa *vpa_types.VerticalPodAutoscaler, conflicting bool, message string) {
	oldStatus := vpa.Status.DeepCopy()

	wasConflicting := false
	if c := getVpaCondition(oldStatus, vpa_types.TargetConflict); c != nil {
		wasConflicting = c.Status == corev1.ConditionTrue
	}

	// Nothing to do: no conflict now and none previously recorded.
	if !conflicting && !wasConflicting {
		return
	}

	newStatus := oldStatus.DeepCopy()
	condStatus := corev1.ConditionFalse
	reason := noTargetConflictReason
	if conflicting {
		condStatus = corev1.ConditionTrue
		reason = targetConflictReason
	} else {
		message = "No conflicting active VPAs found for this target."
	}
	setVpaCondition(newStatus, vpa_types.TargetConflict, condStatus, reason, message)

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

	if conflicting != wasConflicting {
		eventType := corev1.EventTypeWarning
		if !conflicting {
			eventType = corev1.EventTypeNormal
		}
		u.eventRecorder.Event(vpa, eventType, reason, message)
	}
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
func setVpaCondition(status *vpa_types.VerticalPodAutoscalerStatus, condType vpa_types.VerticalPodAutoscalerConditionType, condStatus corev1.ConditionStatus, reason, message string) {
	now := metav1.Now()
	for i := range status.Conditions {
		if status.Conditions[i].Type == condType {
			if status.Conditions[i].Status != condStatus {
				status.Conditions[i].LastTransitionTime = now
			}
			status.Conditions[i].Status = condStatus
			status.Conditions[i].Reason = reason
			status.Conditions[i].Message = message
			return
		}
	}
	status.Conditions = append(status.Conditions, vpa_types.VerticalPodAutoscalerCondition{
		Type:               condType,
		Status:             condStatus,
		LastTransitionTime: now,
		Reason:             reason,
		Message:            message,
	})
}
