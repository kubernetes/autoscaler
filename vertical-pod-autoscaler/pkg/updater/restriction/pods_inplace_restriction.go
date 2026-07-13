/*
Copyright 2015 The Kubernetes Authors.

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

package restriction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"

	resource_updates "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/patch"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/utils"
	resourcehelpers "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/resources"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

// TODO: Make these configurable by flags
const (
	// DeferredResizeUpdateTimeout defines the duration during which an in-place resize request
	// is considered deferred. If the resize is not completed within this time, it falls back to eviction.
	DeferredResizeUpdateTimeout = 5 * time.Minute

	// InProgressResizeUpdateTimeout defines the duration during which an in-place resize request
	// is considered in progress. If the resize is not completed within this time, it falls back to eviction.
	InProgressResizeUpdateTimeout = 1 * time.Hour
)

// PodsInPlaceRestriction controls pods in-place updates. It ensures that we will not update too
// many pods from one replica set. For a replica set, it will allow updating one pod or more if
// inPlaceToleranceFraction is configured.
type PodsInPlaceRestriction interface {
	// InPlaceUpdate attempts to actuate the in-place resize for the given pod.
	// Returns error if the client operation fails.
	InPlaceUpdate(pod *corev1.Pod, vpa *vpa_types.VerticalPodAutoscaler, eventRecorder record.EventRecorder) error

	// CanInPlaceUpdate checks if a pod can be safely updated in-place.
	// Returns:
	//   - InPlaceApproved: The pod can be updated in-place immediately.
	//   - InPlaceDeferred: The update should be deferred (e.g., pod is pending, already updating, or group is not disruptable).
	//   - InPlaceEvict: The pod should be evicted instead of updated in-place.
	//   - InPlaceInfeasible: The in-place update is infeasible (node can't accommodate it or error occurred). Will only retry when recommendation changes.
	//
	// The updateMode parameter affects the decision:
	//   - UpdateModeInPlace: Waits indefinitely for in-progress updates.
	//   - UpdateModeInPlaceOrRecreate: May return InPlaceEvict if the pod update exceeds the timeout threshold.
	CanInPlaceUpdate(pod *corev1.Pod, vpa *vpa_types.VerticalPodAutoscaler, infeasibleAttempts map[k8stypes.UID]*vpa_types.RecommendedPodResources) utils.InPlaceDecision
	// CanUnboost checks if a pod can be safely unboosted.
	CanUnboost(pod *corev1.Pod, vpa *vpa_types.VerticalPodAutoscaler) bool
}

// PodsInPlaceRestrictionImpl is the implementation of the PodsInPlaceRestriction interface.
type PodsInPlaceRestrictionImpl struct {
	client                       kube_client.Interface
	podToReplicaCreatorMap       map[string]podReplicaCreator
	creatorToSingleGroupStatsMap map[podReplicaCreator]singleGroupStats
	patchCalculators             []patch.Calculator
	clock                        clock.Clock
	lastInPlaceAttemptTimeMap    map[string]time.Time
	inPlaceSkipDisruptionBudget  bool
}

// CanInPlaceUpdate checks if pod can be safely updated
func (ip *PodsInPlaceRestrictionImpl) CanInPlaceUpdate(pod *corev1.Pod, vpa *vpa_types.VerticalPodAutoscaler, infeasibleAttempts map[k8stypes.UID]*vpa_types.RecommendedPodResources) utils.InPlaceDecision {
	// Use default mode when update mode is nil
	updateMode := vpa_types.UpdateModeRecreate
	if (vpa.Spec.UpdatePolicy != nil) && (vpa.Spec.UpdatePolicy.UpdateMode != nil) {
		updateMode = *vpa.Spec.UpdatePolicy.UpdateMode
	}

	if updateMode == vpa_types.UpdateModeInPlace && !features.Enabled(features.InPlace) {
		klog.V(4).InfoS("Can't in-place update pod, VPA updateMode is InPlace but InPlace feature gate is not enabled", "pod", klog.KObj(pod), "vpa", klog.KObj(vpa))
		return utils.InPlaceDeferred
	}

	cr, present := ip.podToReplicaCreatorMap[getPodID(pod)]
	if !present {
		klog.V(4).InfoS("Deferring in-place update because VPA cannot determine whether disrupting this Pod is safe; the Pod may be unmanaged, its controller information may be unavailable, or its replica group may be below the configured minimum size", "pod", klog.KObj(pod))
		klog.V(5).InfoS("Pod lacks recognized owner references, its controller is untracked, or min replicas not met", "pod", klog.KObj(pod), "ownerReferences", pod.OwnerReferences)
		return utils.InPlaceDeferred
	}

	if pod.Status.Phase == corev1.PodPending {
		klog.V(4).InfoS("Can't in-place update pod, pod is in Pending phase", "pod", klog.KObj(pod))
		return utils.InPlaceDeferred
	}

	singleGroupStats, present := ip.creatorToSingleGroupStatsMap[cr]
	if !present {
		klog.V(4).InfoS("Deferring in-place update because VPA has no replica availability information for the Pod's controller and cannot verify that the update is safe", "pod", klog.KObj(pod), "replicaCreator", cr)
		return utils.InPlaceDeferred
	}

	if vpa.Status.Recommendation == nil {
		klog.V(4).InfoS("Can't in-place update pod, no recommendation available yet. Waiting for next loop", "pod", klog.KObj(pod))
		return utils.InPlaceDeferred
	}

	recommendation := vpa.Status.Recommendation

	// if the update mode is inPlace we want to check the map for previous Infeasible attempts
	if updateMode == vpa_types.UpdateModeInPlace && infeasibleAttempts != nil {
		lastAttempt, exist := infeasibleAttempts[pod.UID]
		if exist && !resourcehelpers.RecommendationHasLowerResource(lastAttempt, recommendation) {
			klog.V(2).InfoS("Skipping in-place resize because the previous resize could not be accommodated and the new recommendation does not request fewer resources; VPA will retry after a lower recommendation", "pod", klog.KObj(pod))
			return utils.InPlaceInfeasibleCached
		}
	}

	resizeStatus := getResizeStatus(pod)

	// pod is in place updaing
	// utils.ResizeStatusNone means that  there is no container that is inPodResizePending or PodResizeInProgress state
	if resizeStatus != utils.ResizeStatusNone {
		// For InPlace mode: wait for all non-terminal statuses, never evict.
		// Infeasible attempts are tracked and only retried when recommendation changes.
		if updateMode == vpa_types.UpdateModeInPlace {
			switch resizeStatus {
			case utils.ResizeStatusInfeasible:
				// Infeasible means node can't accommodate the resize.
				// Store spec.resources and wait for recommendation to change before retrying.
				klog.V(4).InfoS("Not retrying in-place resize because the kubelet reported that the node cannot accommodate the requested resources; VPA will retry only after a recommendation changes", "pod", klog.KObj(pod))
				return utils.InPlaceInfeasible
			case utils.ResizeStatusDeferred:
				// Deferred means kubelet is waiting to apply the resize.
				// Do nothing, wait for kubelet to proceed.
				klog.V(4).InfoS("Deferring in-place update because the kubelet has postponed the requested resize; VPA will wait instead of sending another resize request", "pod", klog.KObj(pod))
				return utils.InPlaceDeferred
			case utils.ResizeStatusInProgress:
				// Resize is actively being applied, wait for completion.
				klog.V(4).InfoS("Deferring in-place update because the kubelet is applying the previous resize request; VPA will wait for it to complete", "pod", klog.KObj(pod))
				return utils.InPlaceDeferred
			case utils.ResizeStatusError:
				// Error during resize, will retry if recommendation changes.
				klog.V(4).ErrorS(nil, "Not retrying in-place resize because the kubelet reported an error; VPA will retry only after a recommendation changes", "pod", klog.KObj(pod))
				return utils.InPlaceInfeasible
			default:
				klog.V(4).InfoS("Deferring in-place update because the Pod reports an unrecognized resize status; VPA will wait rather than send a conflicting resize request", "pod", klog.KObj(pod), "resizeStatus", resizeStatus)
				return utils.InPlaceDeferred
			}
		}

		// For InPlaceOrRecreate mode, check timeout
		canEvict := CanEvictInPlacingPod(pod, singleGroupStats, ip.lastInPlaceAttemptTimeMap, ip.clock)
		if canEvict {
			klog.V(4).InfoS("Evicting Pod because its in-place resize has failed or stalled; InPlaceOrRecreate mode falls back to recreation when it is safe to disrupt the Pod", "pod", klog.KObj(pod), "resizeStatus", resizeStatus)
			return utils.InPlaceEvict
		}
		klog.V(4).InfoS("Deferring update because the Pod already has an in-place resize pending or in progress; VPA will wait for it to complete or for replica availability to permit recreation", "pod", klog.KObj(pod), "resizeStatus", resizeStatus)
		return utils.InPlaceDeferred
	}

	if ip.inPlaceSkipDisruptionBudget {
		if utils.IsNonDisruptiveResize(pod) {
			klog.V(4).InfoS("Approving in-place update because the requested resize does not require restarting containers; the configured policy permits this without consuming the replica disruption budget", "pod", klog.KObj(pod))
			return utils.InPlaceApproved
		}
		klog.V(4).InfoS("Not bypassing replica availability checks because the requested resize requires restarting containers", "pod", klog.KObj(pod))
	}

	if singleGroupStats.isPodDisruptable() {
		klog.V(4).InfoS("Approving in-place update because the replica group has enough available Pods to maintain its configured availability during the update", "pod", klog.KObj(pod))
		return utils.InPlaceApproved
	}

	klog.V(4).InfoS("Deferring in-place update because updating this Pod could reduce the replica group below its configured availability; VPA will retry when more replicas are available", "pod", klog.KObj(pod))
	return utils.InPlaceDeferred
}

// CanUnboost checks if a pod can be safely unboosted.
func (ip *PodsInPlaceRestrictionImpl) CanUnboost(pod *corev1.Pod, vpa *vpa_types.VerticalPodAutoscaler) bool {
	if !features.Enabled(features.CPUStartupBoost) {
		return false
	}
	if pod.Status.Phase == corev1.PodPending {
		return false
	}
	expiredAnnotations := vpa_api_util.GetExpiredStartupCPUBoostAnnotations(pod, vpa)
	if len(expiredAnnotations) == 0 {
		return false
	}
	klog.V(2).InfoS("Pod ready, checking if containers can be unboosted", "pod", klog.KObj(pod))
	cr, present := ip.podToReplicaCreatorMap[getPodID(pod)]
	if present {
		singleGroupStats, present := ip.creatorToSingleGroupStatsMap[cr]
		if present {
			if isInPlaceUpdating(pod) {
				return false
			}
			return singleGroupStats.isPodDisruptable()
		}
	}
	return false
}

// InPlaceUpdate sends calculates patches and sends resize request to api client. Returns error if client returned error.
// Does not check if pod was actually in-place updated after grace period.
// CanInPlaceUpdate / CanUnboost should be called first.
func (ip *PodsInPlaceRestrictionImpl) InPlaceUpdate(podToUpdate *corev1.Pod, vpa *vpa_types.VerticalPodAutoscaler, eventRecorder record.EventRecorder) error {
	cr, present := ip.podToReplicaCreatorMap[getPodID(podToUpdate)]
	if !present {
		return fmt.Errorf("pod not suitable for in-place update %v: not in replicated pods map", podToUpdate.Name)
	}

	// separate patches since we have to patch resize and spec separately
	resizePatches := []resource_updates.PatchRecord{}
	annotationPatches := []resource_updates.PatchRecord{}
	if podToUpdate.Annotations == nil {
		annotationPatches = append(annotationPatches, patch.GetAddEmptyAnnotationsPatch())
	}
	for _, calculator := range ip.patchCalculators {
		p, err := calculator.CalculatePatches(podToUpdate, vpa)
		if err != nil {
			return err
		}
		klog.V(4).InfoS("Calculated patches for pod", "pod", klog.KObj(podToUpdate), "patches", p)
		if calculator.PatchResourceTarget() == patch.Resize {
			resizePatches = append(resizePatches, p...)
		} else {
			annotationPatches = append(annotationPatches, p...)
		}
	}

	if len(resizePatches) == 0 {
		return errors.New("no resource patches were calculated to apply")
	}

	patch, err := json.Marshal(resizePatches)
	if err != nil {
		return err
	}

	res, err := ip.client.CoreV1().Pods(podToUpdate.Namespace).Patch(context.TODO(), podToUpdate.Name, k8stypes.JSONPatchType, patch, metav1.PatchOptions{}, "resize")
	if err != nil {
		return err
	}
	klog.V(4).InfoS("In-place patched pod /resize subresource using patches", "pod", klog.KObj(res), "patches", string(patch))

	if len(annotationPatches) > 0 {
		patch, err := json.Marshal(annotationPatches)
		if err != nil {
			return err
		}
		res, err = ip.client.CoreV1().Pods(podToUpdate.Namespace).Patch(context.TODO(), podToUpdate.Name, k8stypes.JSONPatchType, patch, metav1.PatchOptions{})
		if err != nil {
			klog.V(4).ErrorS(err, "Failed to patch pod annotations", "pod", klog.KObj(res), "patches", string(patch))
		} else {
			klog.V(4).InfoS("Patched pod annotations", "pod", klog.KObj(res), "patches", string(patch))
		}
	}

	eventRecorder.Event(podToUpdate, corev1.EventTypeNormal, "InPlaceResizedByVPA", "Pod was resized in place by VPA Updater.")

	singleGroupStats, present := ip.creatorToSingleGroupStatsMap[cr]
	if !present {
		klog.InfoS("Internal error - cannot find stats for replication group", "pod", klog.KObj(podToUpdate), "podReplicaCreator", cr)
	} else {
		singleGroupStats.inPlaceUpdateInitiated = singleGroupStats.inPlaceUpdateInitiated + 1
		ip.creatorToSingleGroupStatsMap[cr] = singleGroupStats
	}

	return nil
}

// CanEvictInPlacingPod checks if the pod can be evicted while it is currently in the middle of an in-place update.
func CanEvictInPlacingPod(pod *corev1.Pod, singleGroupStats singleGroupStats, lastInPlaceAttemptTimeMap map[string]time.Time, clock clock.Clock) bool {
	if !isInPlaceUpdating(pod) {
		return false
	}
	lastUpdate, exists := lastInPlaceAttemptTimeMap[getPodID(pod)]
	if !exists {
		klog.V(4).InfoS("Pod already reports an in-place resize, but VPA has no recorded start time; starting the fallback timeout now", "pod", klog.KObj(pod))
		lastUpdate = clock.Now()
		lastInPlaceAttemptTimeMap[getPodID(pod)] = lastUpdate
	}

	if singleGroupStats.isPodDisruptable() {
		// if currently inPlaceUpdating, we should only fallback to eviction if the update has failed. i.e: one of the following conditions:
		// - Infeasible
		// - Deferred + more than 5 minutes has elapsed since the lastInPlaceUpdateTime
		// - InProgress + more than 1 hour has elapsed since the lastInPlaceUpdateTime
		resizePendingCondition, ok := utils.GetPodCondition(pod, corev1.PodResizePending)
		if ok {
			switch resizePendingCondition.Reason {
			case corev1.PodReasonDeferred:
				elapsed := clock.Since(lastUpdate)
				if elapsed > DeferredResizeUpdateTimeout {
					klog.V(4).InfoS("Evicting Pod because the kubelet has deferred its in-place resize longer than the allowed timeout; recreation is now used as a fallback", "pod", klog.KObj(pod), "elapsed", elapsed, "timeout", DeferredResizeUpdateTimeout)
					return true
				}
			case corev1.PodReasonInfeasible:
				klog.V(4).InfoS("Evicting Pod because the kubelet reported that the node cannot accommodate its in-place resize; recreation is used as a fallback", "pod", klog.KObj(pod))
				return true
			default:
				klog.V(4).InfoS("Evicting Pod because the kubelet reported an unexpected pending-resize condition; recreation is used as a safe fallback", "pod", klog.KObj(pod), "resizeConditionReason", resizePendingCondition.Reason, "resizeConditionMessage", resizePendingCondition.Message)
				return true
			}
		} else {
			resizeInProgressCondition, ok := utils.GetPodCondition(pod, corev1.PodResizeInProgress)
			if ok {
				if resizeInProgressCondition.Reason == "" && resizeInProgressCondition.Message == "" {
					elapsed := clock.Since(lastUpdate)
					if elapsed > InProgressResizeUpdateTimeout {
						klog.V(4).InfoS("Evicting Pod because the kubelet has been applying its in-place resize longer than the allowed timeout; recreation is now used as a fallback", "pod", klog.KObj(pod), "elapsed", elapsed, "timeout", InProgressResizeUpdateTimeout)
						return true
					}
				} else if resizeInProgressCondition.Reason == corev1.PodReasonError {
					klog.V(4).InfoS("Evicting Pod because the kubelet reported an error while applying its in-place resize; recreation is used as a fallback", "pod", klog.KObj(pod), "resizeConditionMessage", resizeInProgressCondition.Message)
					return true
				} else {
					klog.V(4).InfoS("Evicting Pod because the kubelet reported an unexpected in-progress-resize condition; recreation is used as a safe fallback", "pod", klog.KObj(pod), "resizeConditionReason", resizeInProgressCondition.Reason, "resizeConditionMessage", resizeInProgressCondition.Message)
					return true
				}
			}
		}
		return false
	}
	klog.V(4).InfoS("Not evicting a Pod that is already resizing because the replica group cannot currently tolerate another disruption", "pod", klog.KObj(pod))
	return false
}
