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

package logic

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	"golang.org/x/time/rate"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	corescheme "k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/set"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/patch"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/scheme"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/priority"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/restriction"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/utils"
	metrics_updater "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/updater"
	resourcehelpers "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/resources"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/status"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

// logDeprecationWarnings logs deprecation warnings for VPAs using deprecated modes
func logDeprecationWarnings(vpa *vpa_types.VerticalPodAutoscaler) {
	if vpa.Spec.UpdatePolicy != nil &&
		vpa.Spec.UpdatePolicy.UpdateMode != nil &&
		*vpa.Spec.UpdatePolicy.UpdateMode == vpa_types.UpdateModeAuto { //nolint:staticcheck
		klog.InfoS("VPA uses deprecated UpdateMode 'Auto'. This mode is deprecated and will be removed in a future API version. Please use explicit update modes like 'Recreate', 'Initial', or 'InPlaceOrRecreate'",
			"vpa", klog.KObj(vpa), "issue", "https://github.com/kubernetes/autoscaler/issues/8424")
	}
}

// Updater performs updates on pods if recommended by Vertical Pod Autoscaler
type Updater interface {
	// RunOnce represents single iteration in the main-loop of Updater
	RunOnce(context.Context)
}

type updater struct {
	vpaLister                    vpa_lister.VerticalPodAutoscalerLister
	podLister                    listersv1.PodLister
	eventRecorder                record.EventRecorder
	restrictionFactory           restriction.PodsRestrictionFactory
	recommendationProcessor      vpa_api_util.RecommendationProcessor
	evictionAdmission            priority.PodEvictionAdmission
	priorityProcessor            priority.PriorityProcessor
	evictionRateLimiter          *rate.Limiter
	inPlaceRateLimiter           *rate.Limiter
	selectorFetcher              target.VpaTargetSelectorFetcher
	useAdmissionControllerStatus bool
	statusValidator              status.Validator
	controllerFetcher            controllerfetcher.ControllerFetcher
	ignoredNamespaces            []string
	// infeasibleMu guards infeasibleAttempts.
	infeasibleMu               sync.RWMutex
	infeasibleAttempts         map[types.UID]*vpa_types.RecommendedPodResources
	defaultUpdateThreshold     float64
	podLifetimeUpdateThreshold time.Duration
	evictAfterOOMThreshold     time.Duration
}

// NewUpdater creates Updater with given configuration
func NewUpdater(
	kubeClient kube_client.Interface,
	vpaClient *vpa_clientset.Clientset,
	minReplicasForEviction int,
	evictionRateLimit float64,
	evictionRateBurst int,
	evictionToleranceFraction float64,
	useAdmissionControllerStatus bool,
	inPlaceSkipDisruptionBudget bool,
	defaultUpdateThreshold float64,
	podLifetimeUpdateThreshold time.Duration,
	evictAfterOOMThreshold time.Duration,
	statusNamespace string,
	recommendationProcessor vpa_api_util.RecommendationProcessor,
	evictionAdmission priority.PodEvictionAdmission,
	selectorFetcher target.VpaTargetSelectorFetcher,
	controllerFetcher controllerfetcher.ControllerFetcher,
	priorityProcessor priority.PriorityProcessor,
	namespace string,
	ignoredNamespaces []string,
	patchCalculators []patch.Calculator,
) (Updater, error) {
	evictionRateLimiter := getRateLimiter(evictionRateLimit, evictionRateBurst)
	// TODO: Create in-place rate limits for the in-place rate limiter
	inPlaceRateLimiter := getRateLimiter(evictionRateLimit, evictionRateBurst)
	factory, err := restriction.NewPodsRestrictionFactory(
		kubeClient,
		minReplicasForEviction,
		evictionToleranceFraction,
		patchCalculators,
		inPlaceSkipDisruptionBudget,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create restriction factory: %v", err)
	}

	return &updater{
		vpaLister:                    vpa_api_util.NewVpasLister(vpaClient, make(chan struct{}), namespace),
		podLister:                    newPodLister(kubeClient, namespace),
		eventRecorder:                newEventRecorder(kubeClient),
		restrictionFactory:           factory,
		recommendationProcessor:      recommendationProcessor,
		evictionRateLimiter:          evictionRateLimiter,
		inPlaceRateLimiter:           inPlaceRateLimiter,
		evictionAdmission:            evictionAdmission,
		priorityProcessor:            priorityProcessor,
		selectorFetcher:              selectorFetcher,
		controllerFetcher:            controllerFetcher,
		useAdmissionControllerStatus: useAdmissionControllerStatus,
		statusValidator: status.NewValidator(
			kubeClient,
			status.AdmissionControllerStatusName,
			statusNamespace,
		),
		infeasibleAttempts:         make(map[types.UID]*vpa_types.RecommendedPodResources),
		ignoredNamespaces:          ignoredNamespaces,
		defaultUpdateThreshold:     defaultUpdateThreshold,
		podLifetimeUpdateThreshold: podLifetimeUpdateThreshold,
		evictAfterOOMThreshold:     evictAfterOOMThreshold,
	}, nil
}

// RunOnce represents single iteration in the main-loop of Updater
func (u *updater) RunOnce(ctx context.Context) {
	timer := metrics_updater.NewExecutionTimer()
	defer timer.ObserveTotal()

	if u.useAdmissionControllerStatus {
		isValid, err := u.statusValidator.IsStatusValid(ctx, status.AdmissionControllerStatusTimeout)
		if err != nil {
			klog.ErrorS(err, "Error getting Admission Controller status. Skipping eviction loop")
			return
		}
		if !isValid {
			klog.V(0).InfoS("Admission Controller status is not valid. Skipping eviction loop", "timeout", status.AdmissionControllerStatusTimeout)
			return
		}
	}

	vpaList, err := u.vpaLister.List(labels.Everything())
	if err != nil {
		klog.ErrorS(err, "Failed to get VPA list")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	timer.ObserveStep("ListVPAs")

	vpas := make([]*vpa_api_util.VpaWithSelector, 0)

	inPlaceOrRecreateFeatureEnable := features.Enabled(features.InPlaceOrRecreate)
	inPlaceFeatureEnable := features.Enabled(features.InPlace)
	cpuStartupBoostEnabled := features.Enabled(features.CPUStartupBoost)

	for _, vpa := range vpaList {
		if slices.Contains(u.ignoredNamespaces, vpa.Namespace) {
			klog.V(3).InfoS("Skipping VPA object in ignored namespace", "vpa", klog.KObj(vpa), "namespace", vpa.Namespace)
			continue
		}
		// Log deprecation warnings for VPAs using deprecated modes
		logDeprecationWarnings(vpa)

		updateMode := vpa_api_util.GetUpdateMode(vpa)
		if updateMode != vpa_types.UpdateModeRecreate &&
			updateMode != vpa_types.UpdateModeAuto && //nolint:staticcheck
			updateMode != vpa_types.UpdateModeInPlaceOrRecreate &&
			updateMode != vpa_types.UpdateModeInPlace &&
			vpa.Spec.StartupBoost == nil {
			klog.V(3).InfoS("Skipping VPA object because its mode is not  \"InPlaceOrRecreate\", \"InPlace\", \"Recreate\" or \"Auto\" and it doesn't have startupBoost configured", "vpa", klog.KObj(vpa))
			continue
		}
		selector, err := u.selectorFetcher.Fetch(ctx, vpa)
		if err != nil {
			klog.V(3).InfoS("Skipping VPA object because we cannot fetch selector", "vpa", klog.KObj(vpa))
			continue
		}

		vpas = append(vpas, &vpa_api_util.VpaWithSelector{
			Vpa:      vpa,
			Selector: selector,
		})
	}

	if len(vpas) == 0 {
		klog.V(0).InfoS("No VPA objects to process")
		if u.evictionAdmission != nil {
			u.evictionAdmission.CleanUp()
		}
		return
	}

	podsList, err := u.podLister.List(labels.Everything())
	if err != nil {
		klog.ErrorS(err, "Failed to get pods list")
		return
	}
	timer.ObserveStep("ListPods")
	allLivePods := filterDeletedPods(podsList)

	// Clean up stale infeasible attempts for pods that no longer exist
	if len(u.infeasibleAttempts) > 0 {
		u.cleanupStaleInfeasibleAttempts(allLivePods)
	}

	controlledPods := make(map[*vpa_types.VerticalPodAutoscaler][]*corev1.Pod)
	for _, pod := range allLivePods {
		controllingVPA := vpa_api_util.GetControllingVPAForPod(ctx, pod, vpas, u.controllerFetcher)
		if controllingVPA != nil {
			controlledPods[controllingVPA.Vpa] = append(controlledPods[controllingVPA.Vpa], pod)
		}
	}
	timer.ObserveStep("FilterPods")

	if u.evictionAdmission != nil {
		u.evictionAdmission.LoopInit(allLivePods, controlledPods)
	}
	timer.ObserveStep("AdmissionInit")

	// wrappers for metrics which are computed every loop run
	controlledPodsCounter := metrics_updater.NewControlledPodsCounter()
	evictablePodsCounter := metrics_updater.NewEvictablePodsCounter()
	inPlaceUpdatablePodsCounter := metrics_updater.NewInPlaceUpdatablePodsCounter()
	vpasWithEvictablePodsCounter := metrics_updater.NewVpasWithEvictablePodsCounter()
	vpasWithEvictedPodsCounter := metrics_updater.NewVpasWithEvictedPodsCounter()

	vpasWithInPlaceUpdatablePodsCounter := metrics_updater.NewVpasWithInPlaceUpdatablePodsCounter()
	vpasWithInPlaceUpdatedPodsCounter := metrics_updater.NewVpasWithInPlaceUpdatedPodsCounter()

	// using defer to protect against 'return' after evictionRateLimiter.Wait
	defer controlledPodsCounter.Observe()
	defer evictablePodsCounter.Observe()
	defer vpasWithEvictablePodsCounter.Observe()
	defer vpasWithEvictedPodsCounter.Observe()
	// separate counters for in-place
	defer inPlaceUpdatablePodsCounter.Observe()
	defer vpasWithInPlaceUpdatablePodsCounter.Observe()
	defer vpasWithInPlaceUpdatedPodsCounter.Observe()

	for vpa, livePods := range controlledPods {
		vpaSize := len(livePods)
		updateMode := vpa_api_util.GetUpdateMode(vpa)
		controlledPodsCounter.Add(vpaSize, updateMode, vpaSize)
		creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := u.restrictionFactory.GetCreatorMaps(livePods, vpa)
		if err != nil {
			klog.ErrorS(err, "Failed to get creator maps")
			continue
		}

		inPlaceLimiter := u.restrictionFactory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)
		podsAvailableForUpdate := make([]*corev1.Pod, 0)
		podsToUnboost := make([]*corev1.Pod, 0)
		withInPlaceUpdated := false

		if cpuStartupBoostEnabled && vpa.Spec.StartupBoost != nil {
			// First, handle unboosting for pods that have finished their startup period.
			for _, pod := range livePods {
				if vpa_api_util.PodHasCPUBoostInProgressAnnotation(pod) {
					if vpa_api_util.IsPodReadyAndStartupBoostDurationPassed(pod, vpa) {
						podsToUnboost = append(podsToUnboost, pod)
					}
				} else {
					podsAvailableForUpdate = append(podsAvailableForUpdate, pod)
				}
			}

			// Perform unboosting
			for _, pod := range podsToUnboost {
				if inPlaceLimiter.CanUnboost(pod, vpa) {
					klog.V(2).InfoS("Unboosting pod", "pod", klog.KObj(pod))
					err = u.inPlaceRateLimiter.Wait(ctx)
					if err != nil {
						klog.V(0).InfoS("In-place rate limiter wait failed for unboosting", "error", err)
						return
					}
					err := inPlaceLimiter.InPlaceUpdate(pod, vpa, u.eventRecorder)
					if err != nil {
						klog.V(0).InfoS("Unboosting failed", "error", err, "pod", klog.KObj(pod))
						metrics_updater.RecordFailedInPlaceUpdate(vpaSize, vpa.Name, vpa.Namespace, "UnboostError")
					} else {
						klog.V(2).InfoS("Successfully unboosted pod", "pod", klog.KObj(pod))
						withInPlaceUpdated = true
						metrics_updater.AddInPlaceUpdatedPod(vpaSize, vpa.Name, vpa.Namespace)
					}
				}
			}
		} else {
			// CPU Startup Boost is not enabled or configured for this VPA,
			// so all live pods are available for potential standard VPA updates.
			podsAvailableForUpdate = livePods
		}

		if updateMode == vpa_types.UpdateModeOff || updateMode == vpa_types.UpdateModeInitial {
			continue
		}

		evictionLimiter := u.restrictionFactory.NewPodsEvictionRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)
		podsForEviction := make([]*corev1.Pod, 0)
		podsForInPlace := make([]*corev1.Pod, 0)
		withInPlaceUpdatable := false
		withEvictable := false

		if (updateMode == vpa_types.UpdateModeInPlaceOrRecreate && inPlaceOrRecreateFeatureEnable) || (updateMode == vpa_types.UpdateModeInPlace && inPlaceFeatureEnable) {
			podsForInPlace = u.getPodsUpdateOrder(filterNonInPlaceUpdatablePods(podsAvailableForUpdate, inPlaceLimiter, updateMode), vpa)
			inPlaceUpdatablePodsCounter.Add(vpaSize, len(podsForInPlace))
			if len(podsForInPlace) > 0 {
				withInPlaceUpdatable = true
			}
		} else {
			// If the feature gate is not enabled but update mode is InPlaceOrRecreate, updater will always fallback to eviction.
			if updateMode == vpa_types.UpdateModeInPlaceOrRecreate {
				klog.InfoS("Warning: feature gate is not enabled for this updateMode", "featuregate", features.InPlaceOrRecreate, "updateMode", vpa_types.UpdateModeInPlaceOrRecreate)
				// If the feature gate is not enabled but update mode is InPlace, updater will do nothing.
			} else if updateMode == vpa_types.UpdateModeInPlace {
				klog.InfoS("Warning: feature gate is not enabled for this updateMode", "featuregate", features.InPlace, "updateMode", vpa_types.UpdateModeInPlace)
				continue
			}
			podsForEviction = u.getPodsUpdateOrder(filterNonEvictablePods(podsAvailableForUpdate, evictionLimiter), vpa)
			evictablePodsCounter.Add(vpaSize, updateMode, len(podsForEviction))
			if len(podsForEviction) > 0 {
				withEvictable = true
			}
		}

		withEvicted := false

		for _, pod := range podsForInPlace {
			decision := inPlaceLimiter.CanInPlaceUpdate(pod, updateMode)

			switch decision {
			case utils.InPlaceDeferred:
				// Pod passed priority calculator, meaning recommendations differ from spec.
				// Retry the in-place update.
				klog.V(0).InfoS("In-place update deferred", "pod", klog.KObj(pod))
				// Fall through to attempt in-place update
			case utils.InPlaceEvict:
				// This should only happen for InPlaceOrRecreate mode
				podsForEviction = append(podsForEviction, pod)
				continue
			case utils.InPlaceInfeasible:
				// if the recommendation hasn't changed we skip the pod
				if resourcehelpers.RecommendationsEqual(u.infeasibleAttempts[pod.UID], vpa.Status.Recommendation) {
					klog.V(2).InfoS("In-place update infeasible, recommendation unchanged, skipping pod", "pod", klog.KObj(pod))
					continue
				}

				// Status is Infeasible, but recommendation has changed
				// Retry in-place update (no backoff for alpha)
				// this status should only be returned with InPlace update mode (InPlaceOrRecreate will return InPlaceEvict in case of infeasible state)
				// Fall through to attempt in-place update
				klog.V(2).InfoS("In-place update infeasible, retrying with new recommendation", "pod", klog.KObj(pod))
				u.recordInfeasibleAttempt(pod, vpa)
			case utils.InPlaceApproved:
				klog.V(2).InfoS("In-place update approved", "pod", klog.KObj(pod))
				// Proceed with in-place update
			default:
				klog.ErrorS(nil, "Unexpected in-place update decision, skipping pod", "decision", decision, "pod", klog.KObj(pod))
				continue
			}

			err = u.inPlaceRateLimiter.Wait(ctx)
			if err != nil {
				klog.V(0).InfoS("In-place rate limiter wait failed for in-place resize", "error", err)
				metrics_updater.RecordFailedInPlaceUpdate(vpaSize, vpa.Name, vpa.Namespace, "InPlaceUpdateRateLimiterWaitFailed")
				return
			}
			err := inPlaceLimiter.InPlaceUpdate(pod, vpa, u.eventRecorder)
			if err != nil {
				reason := "InPlaceUpdateError"
				// For InPlace mode, don't evict pods even if we get an error
				if updateMode == vpa_types.UpdateModeInPlace {
					// TODO: check for an admission plugin error because of OS and node capacity checks
					// Check if it's an infeasibility error
					// infeasible patches are rejected at API server level (soon),
					// so spec.resources remains unchanged. We must track the attempted
					// recommendation to prevent infinite retry loops.
					// This work is still in progress (https://github.com/kubernetes/kubernetes/pull/136043)
					// Currently isInfeasibleError return false
					if isInfeasibleError(err) {
						// TODO: this will be changed when we know how errors shoule be look like
						// depends on https://github.com/kubernetes/kubernetes/pull/136043
						reason = "InPlaceUpdateInfeasible"
						u.recordInfeasibleAttempt(pod, vpa)
					}
					klog.V(0).InfoS("In-place resize failed", "error", err, "pod", klog.KObj(pod), "reason", reason)
					metrics_updater.RecordFailedInPlaceUpdate(vpaSize, vpa.Name, vpa.Namespace, reason)
					continue
				}
				klog.V(0).InfoS("In-place resize failed, falling back to eviction", "error", err, "pod", klog.KObj(pod))
				metrics_updater.RecordFailedInPlaceUpdate(vpaSize, vpa.Name, vpa.Namespace, reason)
				podsForEviction = append(podsForEviction, pod)
				continue
			}
			withInPlaceUpdated = true
			metrics_updater.AddInPlaceUpdatedPod(vpaSize, vpa.Name, vpa.Namespace)
		}

		for _, pod := range podsForEviction {
			if !evictionLimiter.CanEvict(pod) {
				continue
			}
			err = u.evictionRateLimiter.Wait(ctx)
			if err != nil {
				klog.V(0).InfoS("Eviction rate limiter wait failed", "error", err)
				metrics_updater.RecordFailedEviction(vpaSize, vpa.Name, vpa.Namespace, updateMode, "EvictionRateLimiterWaitFailed")
				return
			}
			klog.V(2).InfoS("Evicting pod", "pod", klog.KObj(pod))
			evictErr := evictionLimiter.Evict(pod, vpa, u.eventRecorder)
			if evictErr != nil {
				klog.V(0).InfoS("Eviction failed", "error", evictErr, "pod", klog.KObj(pod))
				metrics_updater.RecordFailedEviction(vpaSize, vpa.Name, vpa.Namespace, updateMode, "EvictionError")
			} else {
				withEvicted = true
				metrics_updater.AddEvictedPod(vpaSize, vpa.Name, vpa.Namespace, updateMode)
			}
		}

		if withInPlaceUpdatable {
			vpasWithInPlaceUpdatablePodsCounter.Add(vpaSize, 1)
		}
		if withInPlaceUpdated {
			vpasWithInPlaceUpdatedPodsCounter.Add(vpaSize, 1)
		}
		if withEvictable {
			vpasWithEvictablePodsCounter.Add(vpaSize, updateMode, 1)
		}
		if withEvicted {
			vpasWithEvictedPodsCounter.Add(vpaSize, updateMode, 1)
		}
	}
	timer.ObserveStep("EvictPods")
}

func (u *updater) cleanupStaleInfeasibleAttempts(livePods []*corev1.Pod) {
	livePodKeys := set.New[types.UID]()
	for _, pod := range livePods {
		livePodKeys.Insert(pod.UID)
	}

	u.infeasibleMu.Lock()
	defer u.infeasibleMu.Unlock()

	for podID := range u.infeasibleAttempts {
		if !livePodKeys.Has(podID) {
			delete(u.infeasibleAttempts, podID)
		}
	}
}

// recordInfeasibleAttempt stores the recommendation that failed as infeasible
func (u *updater) recordInfeasibleAttempt(pod *corev1.Pod, vpa *vpa_types.VerticalPodAutoscaler) {
	processedRecommendation, _, err := u.recommendationProcessor.Apply(vpa, pod)
	if err != nil {
		klog.V(2).ErrorS(err, "Failed to get recommendation for infeasible attempt recording", "pod", klog.KObj(pod))
		return
	}

	u.infeasibleMu.Lock()
	u.infeasibleAttempts[pod.UID] = processedRecommendation
	u.infeasibleMu.Unlock()

	klog.V(2).InfoS("Recorded infeasible attempt, will retry when recommendation changes", "pod", klog.KObj(pod))
}

func getRateLimiter(rateLimit float64, rateLimitBurst int) *rate.Limiter {
	var rateLimiter *rate.Limiter
	if rateLimit <= 0 {
		// As a special case if the rate is set to rate.Inf, the burst rate is ignored
		// see https://github.com/golang/time/blob/master/rate/rate.go#L37
		rateLimiter = rate.NewLimiter(rate.Inf, 0)
		klog.V(1).InfoS("Rate limit disabled")
	} else {
		rateLimiter = rate.NewLimiter(rate.Limit(rateLimit), rateLimitBurst)
	}
	return rateLimiter
}

// getPodsUpdateOrder returns list of pods that should be updated ordered by update priority
func (u *updater) getPodsUpdateOrder(pods []*corev1.Pod, vpa *vpa_types.VerticalPodAutoscaler) []*corev1.Pod {
	updateconfig := priority.UpdateConfig{
		MinChangePriority:          u.defaultUpdateThreshold,
		PodLifetimeUpdateThreshold: u.podLifetimeUpdateThreshold,
		EvictAfterOOMThreshold:     u.evictAfterOOMThreshold,
	}
	priorityCalculator := priority.NewUpdatePriorityCalculator(
		vpa,
		updateconfig,
		u.recommendationProcessor,
		u.priorityProcessor)

	u.infeasibleMu.RLock()
	for _, pod := range pods {
		priorityCalculator.AddPod(pod, time.Now(), u.infeasibleAttempts)
	}
	u.infeasibleMu.RUnlock()

	return priorityCalculator.GetSortedPods(u.evictionAdmission)
}

func filterPods(pods []*corev1.Pod, predicate func(*corev1.Pod) bool) []*corev1.Pod {
	result := make([]*corev1.Pod, 0)
	for _, pod := range pods {
		if predicate(pod) {
			result = append(result, pod)
		}
	}
	return result
}

func filterNonInPlaceUpdatablePods(pods []*corev1.Pod, inplaceRestriction restriction.PodsInPlaceRestriction, updateMode vpa_types.UpdateMode) []*corev1.Pod {
	return filterPods(pods, func(pod *corev1.Pod) bool {
		decision := inplaceRestriction.CanInPlaceUpdate(pod, updateMode)
		switch decision {
		case utils.InPlaceApproved:
			return true
		case utils.InPlaceInfeasible:
			// For InPlace mode, include infeasible pods to retry (no backoff for alpha)
			return updateMode == vpa_types.UpdateModeInPlace
		case utils.InPlaceEvict:
			// For InPlaceOrRecreate, include so they can be redirected to eviction in the loop
			return updateMode == vpa_types.UpdateModeInPlaceOrRecreate
		case utils.InPlaceDeferred:
			// For InPlace mode, include deferred pods so we can check if recommendation
			// changed and apply a new patch while a previous update is in progress
			return updateMode == vpa_types.UpdateModeInPlace
		default:
			return false
		}
	})
}

func filterNonEvictablePods(pods []*corev1.Pod, evictionRestriction restriction.PodsEvictionRestriction) []*corev1.Pod {
	return filterPods(pods, evictionRestriction.CanEvict)
}

func filterDeletedPods(pods []*corev1.Pod) []*corev1.Pod {
	return filterPods(pods, func(pod *corev1.Pod) bool {
		return pod.DeletionTimestamp == nil
	})
}

func newPodLister(kubeClient kube_client.Interface, namespace string) listersv1.PodLister {
	selector := fields.ParseSelectorOrDie("spec.nodeName!=" + "" + ",status.phase!=" +
		string(corev1.PodSucceeded) + ",status.phase!=" + string(corev1.PodFailed))
	podListWatch := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "pods", namespace, selector)
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podLister := listersv1.NewPodLister(store)
	podReflector := cache.NewReflector(podListWatch, &corev1.Pod{}, store, time.Hour)
	stopCh := make(chan struct{})
	go podReflector.Run(stopCh)

	return podLister
}

func newEventRecorder(kubeClient kube_client.Interface) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(4)
	if _, isFake := kubeClient.(*fake.Clientset); !isFake {
		eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: typedcorev1.New(kubeClient.CoreV1().RESTClient()).Events("")})
	} else {
		eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	}

	vpascheme := scheme.Scheme
	if err := corescheme.AddToScheme(vpascheme); err != nil {
		klog.ErrorS(err, "Error adding core scheme")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	return eventBroadcaster.NewRecorder(vpascheme, corev1.EventSource{Component: "vpa-updater"})
}

// isInfeasibleError checks if an error indicates the resize is infeasible.
// Infeasible error detection at the admission controller level is still in progress
// (https://github.com/kubernetes/kubernetes/pull/136043).
// This is just a placeholder until that work lands and we know the exact error format.
func isInfeasibleError(err error) bool {
	return false
}

func (u *updater) CleanupInfeasibleAttempts(livePods []*corev1.Pod) {
	u.infeasibleMu.Lock()
	defer u.infeasibleMu.Unlock()

	// Build a set of existing pod UIDs
	seenPods := sets.New[types.UID]()
	for _, pod := range livePods {
		seenPods.Insert(pod.UID)
	}

	// Remove entries for pods that no longer exist
	for podUID := range u.infeasibleAttempts {
		if !seenPods.Has(podUID) {
			delete(u.infeasibleAttempts, podUID)
			klog.V(4).InfoS("Cleaned up infeasible attempt for non-existent pod", "podUID", podUID)
		}
	}
}
