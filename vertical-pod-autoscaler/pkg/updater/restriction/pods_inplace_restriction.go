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
	utils "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/utils"
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
// many pods from one replica set. For replica set will allow to update one pod or more if
// inPlaceToleranceFraction is configured.
type PodsInPlaceRestriction interface {
	// InPlaceUpdate attempts to actuate the in-place resize.
	// Returns error if client returned error.
	InPlaceUpdate(pod *corev1.Pod, vpa *vpa_types.VerticalPodAutoscaler, eventRecorder record.EventRecorder) error
	// CanInPlaceUpdate checks if pod can be safely updated in-place. If not, it will return a decision to potentially evict the pod.
	CanInPlaceUpdate(pod *corev1.Pod) utils.InPlaceDecision
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
func (ip *PodsInPlaceRestrictionImpl) CanInPlaceUpdate(pod *corev1.Pod) utils.InPlaceDecision {
	if !features.Enabled(features.InPlaceOrRecreate) {
		return utils.InPlaceEvict
	}

	cr, present := ip.podToReplicaCreatorMap[getPodID(pod)]
	if present {
		singleGroupStats, present := ip.creatorToSingleGroupStatsMap[cr]
		if pod.Status.Phase == corev1.PodPending {
			return utils.InPlaceDeferred
		}
		if present {
			if isInPlaceUpdating(pod) {
				canEvict := CanEvictInPlacingPod(pod, singleGroupStats, ip.lastInPlaceAttemptTimeMap, ip.clock)
				if canEvict {
					return utils.InPlaceEvict
				}
				return utils.InPlaceDeferred
			}
			if ip.inPlaceSkipDisruptionBudget {
				if utils.IsNonDisruptiveResize(pod) {
					klog.V(4).InfoS("in-place-skip-disruption-budget enabled, skipping disruption budget check for in-place update")
					return utils.InPlaceApproved
				}
				klog.V(4).InfoS("in-place-skip-disruption-budget enabled, but pod has RestartContainer resize policy", "pod", klog.KObj(pod))
			}
			if singleGroupStats.isPodDisruptable() {
				return utils.InPlaceApproved
			}
		}
	}
	klog.V(4).InfoS("Can't in-place update pod, but not falling back to eviction. Waiting for next loop", "pod", klog.KObj(pod))
	return utils.InPlaceDeferred
}

// CanUnboost checks if a pod can be safely unboosted.
func (ip *PodsInPlaceRestrictionImpl) CanUnboost(pod *corev1.Pod, vpa *vpa_types.VerticalPodAutoscaler) bool {
	if !features.Enabled(features.CPUStartupBoost) {
		return false
	}
	durationPassed := vpa_api_util.IsPodReadyAndStartupBoostDurationPassed(pod, vpa)

	klog.V(2).InfoS("Checking if pod can be unboosted", "pod", klog.KObj(pod), "durationPassed", durationPassed)

	if !durationPassed {
		return false
	}
	cr, present := ip.podToReplicaCreatorMap[getPodID(pod)]
	if present {
		singleGroupStats, present := ip.creatorToSingleGroupStatsMap[cr]
		if present {
			return singleGroupStats.isPodDisruptable()
		}
	}
	return false
}

// InPlaceUpdate sends calculates patches and sends resize request to api client. Returns error if pod cannot be in-place updated or if client returned error.
// Does not check if pod was actually in-place updated after grace period.
func (ip *PodsInPlaceRestrictionImpl) InPlaceUpdate(podToUpdate *corev1.Pod, vpa *vpa_types.VerticalPodAutoscaler, eventRecorder record.EventRecorder) error {
	cr, present := ip.podToReplicaCreatorMap[getPodID(podToUpdate)]
	if !present {
		return fmt.Errorf("pod not suitable for in-place update %v: not in replicated pods map", podToUpdate.Name)
	}

	if ip.CanInPlaceUpdate(podToUpdate) != utils.InPlaceApproved {
		return fmt.Errorf("cannot in-place update pod %s", klog.KObj(podToUpdate))
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
		klog.V(4).InfoS("In-place update in progress for pod but no lastUpdateTime found, setting it to now", "pod", klog.KObj(pod))
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
				if clock.Since(lastUpdate) > DeferredResizeUpdateTimeout {
					klog.V(4).InfoS(fmt.Sprintf("In-place update deferred for more than %v, falling back to eviction", DeferredResizeUpdateTimeout), "pod", klog.KObj(pod))
					return true
				}
			case corev1.PodReasonInfeasible:
				klog.V(4).InfoS("In-place update infeasible, falling back to eviction", "pod", klog.KObj(pod))
				return true
			default:
				klog.V(4).InfoS("In-place update condition unknown, falling back to eviction", "pod", klog.KObj(pod), "condition", resizePendingCondition)
				return true
			}
		} else {
			resizeInProgressCondition, ok := utils.GetPodCondition(pod, corev1.PodResizeInProgress)
			if ok {
				if resizeInProgressCondition.Reason == "" && resizeInProgressCondition.Message == "" {
					if clock.Since(lastUpdate) > InProgressResizeUpdateTimeout {
						klog.V(4).InfoS(fmt.Sprintf("In-place update in progress for more than %v, falling back to eviction", InProgressResizeUpdateTimeout), "pod", klog.KObj(pod))
						return true
					}
				} else if resizeInProgressCondition.Reason == corev1.PodReasonError {
					klog.V(4).InfoS("In-place update error, falling back to eviction", "pod", klog.KObj(pod), "message", resizeInProgressCondition.Message)
					return true
				} else {
					klog.V(4).InfoS("In-place update condition unknown, falling back to eviction", "pod", klog.KObj(pod), "condition", resizeInProgressCondition)
					return true
				}
			}
		}
		return false
	}
	klog.V(4).InfoS("Would be able to evict, but already resizing", "pod", klog.KObj(pod))
	return false
}
