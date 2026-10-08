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
	"context"
	"encoding/json"
	"fmt"
	"slices"

	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

const targetConflictReason = "MultipleActiveVPATargetsFound"
const noTargetConflictReason = "NoConflictingActiveVPATargets"

// reconcileTargetConflicts groups the given VPAs by their targetRef. Within
// each group of active VPAs (update mode other than "Off", or "Off" with a
// startupBoost) it picks the controlling VPA with vpa_api_util.Stronger, the
// same logic the admission controller and updater use to choose a VPA for a
// pod. Every other VPA in the group gets the TargetConflict condition set to
// True, naming the controlling VPA. The controlling VPA is left untouched.
// VPAs that are not in conflict, or are no longer eligible, have a previously
// set TargetConflict condition cleared.
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
		// "Off" VPAs without a startupBoost don't act on pods, so they can't conflict.
		if vpa_api_util.GetUpdateMode(vpa) == vpa_types.UpdateModeOff && !vpa_api_util.HasStartupBoost(vpa) {
			ineligible = append(ineligible, vpa)
			continue
		}
		key := vpa_api_util.TargetRefIndexKey(vpa.Namespace, vpa.Spec.TargetRef.Kind, vpa.Spec.TargetRef.Name)
		vpaGroups[key] = append(vpaGroups[key], vpa)
	}

	for _, group := range vpaGroups {
		controlling := group[0]
		for _, vpa := range group[1:] {
			if vpa_api_util.Stronger(vpa, controlling) {
				controlling = vpa
			}
		}
		for _, vpa := range group {
			if vpa == controlling {
				// The controlling VPA (or the only VPA) is never in conflict.
				u.updateTargetConflictCondition(vpa, "")
				continue
			}
			u.updateTargetConflictCondition(vpa, controlling.Name)
		}
	}

	// VPAs that became ineligible (e.g. switched to Off, or dropped their
	// targetRef) may still carry a stale TargetConflict=True condition.
	// updateTargetConflictCondition is a no-op if there never was one.
	for _, vpa := range ineligible {
		u.updateTargetConflictCondition(vpa, "")
	}
}

// updateTargetConflictCondition patches the TargetConflict condition on a
// single VPA if it needs to change, and emits a Warning/Normal event only on
// a state transition (conflict appearing or clearing). controllingVpaName is
// the name of the VPA that controls the target instead of this one; an empty
// string means this VPA is not in conflict.
func (u *updater) updateTargetConflictCondition(vpa *vpa_types.VerticalPodAutoscaler, controllingVpaName string) {
	conflicting := controllingVpaName != ""
	wasConflicting := false
	if c := getVpaCondition(&vpa.Status, vpa_types.TargetConflict); c != nil {
		wasConflicting = c.Status == corev1.ConditionTrue
	}
	// Skip the copy below when there's nothing to do.
	if !conflicting && !wasConflicting {
		return
	}
	oldStatus := vpa.Status.DeepCopy()

	// transitioned is true only when the conflict appears or clears. While a
	// conflict persists we still rebuild the condition, because the controlling
	// VPA (and so the message) can change without a transition; the patch is
	// skipped when nothing differs.
	transitioned := conflicting != wasConflicting

	condStatus := corev1.ConditionFalse
	reason := noTargetConflictReason
	message := "No conflicting active VPAs found for this target."
	if conflicting {
		condStatus = corev1.ConditionTrue
		reason = targetConflictReason
		message = fmt.Sprintf("Conflict: multiple active VPAs target the same resource; VPA %q controls it and this VPA is not applied", controllingVpaName)
	}

	newStatus := oldStatus.DeepCopy()
	setVpaCondition(newStatus, vpa_types.TargetConflict, condStatus, reason, message, vpa.Generation)
	// Nothing changed (same status, reason, message and generation): don't
	// send a status patch every updater loop for an unchanged conflict.
	if apiequality.Semantic.DeepEqual(oldStatus.Conditions, newStatus.Conditions) {
		return
	}

	// Patch conditions only so a concurrent recommender status write isn't clobbered.
	if err := u.patchTargetConflictConditions(vpa, newStatus.Conditions); err != nil {
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

// patchTargetConflictConditions patches status.conditions only, leaving the
// rest of status (owned by the recommender) untouched.
func (u *updater) patchTargetConflictConditions(vpa *vpa_types.VerticalPodAutoscaler, conditions []vpa_types.VerticalPodAutoscalerCondition) error {
	patch := []struct {
		Op    string                                     `json:"op"`
		Path  string                                     `json:"path"`
		Value []vpa_types.VerticalPodAutoscalerCondition `json:"value"`
	}{{Op: "add", Path: "/status/conditions", Value: conditions}}
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("marshal TargetConflict patch: %v", err)
	}
	_, err = u.vpaClient.AutoscalingV1().VerticalPodAutoscalers(vpa.Namespace).
		Patch(context.TODO(), vpa.Name, types.JSONPatchType, patchBytes, metav1.PatchOptions{}, "status")
	return err
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
