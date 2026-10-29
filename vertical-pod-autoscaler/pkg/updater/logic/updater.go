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
	"errors"
	"fmt"
	"slices"
	"time"

	"golang.org/x/time/rate"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/informers"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	corescheme "k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	listersv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
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
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/status"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

// TODO(omerap12): Should we make this configurable?
const maxVPALookupRetries = 5

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
	// RunBoostWorker processes the startup boost unboost queue until ctx is cancelled.
	RunBoostWorker(context.Context)
	// PodAddHandler handles pod add events for the startup boost unboost queue.
	PodAddHandler(obj any)
	// PodUpdateHandler handles pod update events for the startup boost unboost queue.
	PodUpdateHandler(oldObj, curObj any)
	// ShutDown shuts down the boost work queue.
	ShutDown()
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
	statusTimeout                time.Duration
	controllerFetcher            controllerfetcher.ControllerFetcher
	ignoredNamespaces            []string
	infeasibleAttempts           map[types.UID]*vpa_types.RecommendedPodResources
	defaultUpdateThreshold       float64
	podLifetimeUpdateThreshold   time.Duration
	evictAfterOOMThreshold       time.Duration
	requireObservedGeneration    bool
	podInformer                  cache.SharedIndexInformer
	cpuStartupBoostQueue         workqueue.TypedRateLimitingInterface[string]
}

// NewUpdater creates Updater with given configuration
func NewUpdater(
	kubeClient kube_client.Interface,
	vpaClient *vpa_clientset.Clientset,
	kubeInformerFactory informers.SharedInformerFactory,
	podInformerFactory informers.SharedInformerFactory,
	minReplicasForEviction int,
	evictionRateLimit float64,
	evictionRateBurst int,
	evictionToleranceFraction float64,
	useAdmissionControllerStatus bool,
	inPlaceSkipDisruptionBudget bool,
	defaultUpdateThreshold float64,
	podLifetimeUpdateThreshold time.Duration,
	evictAfterOOMThreshold time.Duration,
	statusLeaseName string,
	statusNamespace string,
	statusTimeout time.Duration,
	recommendationProcessor vpa_api_util.RecommendationProcessor,
	evictionAdmission priority.PodEvictionAdmission,
	selectorFetcher target.VpaTargetSelectorFetcher,
	controllerFetcher controllerfetcher.ControllerFetcher,
	priorityProcessor priority.PriorityProcessor,
	namespace string,
	ignoredNamespaces []string,
	patchCalculators []patch.Calculator,
	requireObservedGeneration bool,
) (Updater, error) {
	evictionRateLimiter := getRateLimiter(evictionRateLimit, evictionRateBurst)
	// TODO: Create in-place rate limits for the in-place rate limiter
	inPlaceRateLimiter := getRateLimiter(evictionRateLimit, evictionRateBurst)
	factory := restriction.NewPodsRestrictionFactory(
		kubeClient,
		kubeInformerFactory,
		minReplicasForEviction,
		evictionToleranceFraction,
		patchCalculators,
		inPlaceSkipDisruptionBudget,
	)

	u := &updater{
		vpaLister:                    vpa_api_util.NewVpasLister(vpaClient, make(chan struct{}), namespace),
		podLister:                    podInformerFactory.Core().V1().Pods().Lister(),
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
			statusLeaseName,
			statusNamespace,
		),
		statusTimeout:              statusTimeout,
		infeasibleAttempts:         make(map[types.UID]*vpa_types.RecommendedPodResources),
		ignoredNamespaces:          ignoredNamespaces,
		defaultUpdateThreshold:     defaultUpdateThreshold,
		podLifetimeUpdateThreshold: podLifetimeUpdateThreshold,
		evictAfterOOMThreshold:     evictAfterOOMThreshold,
		requireObservedGeneration:  requireObservedGeneration,
	}
	if features.Enabled(features.CPUStartupBoost) {
		u.podInformer = podInformerFactory.Core().V1().Pods().Informer()
		u.cpuStartupBoostQueue = workqueue.NewTypedRateLimitingQueueWithConfig(
			workqueue.NewTypedItemExponentialFailureRateLimiter[string](100*time.Millisecond, 1000*time.Second),
			workqueue.TypedRateLimitingQueueConfig[string]{
				Name: "cpu-startup-boost",
			},
		)
		if _, err := u.podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    u.PodAddHandler,
			UpdateFunc: u.PodUpdateHandler,
		}); err != nil {
			return nil, fmt.Errorf("adding Pod event handler: %w", err)
		}
	}

	return u, nil
}

// RunOnce represents single iteration in the main-loop of Updater
func (u *updater) RunOnce(ctx context.Context) {
	timer := metrics_updater.NewExecutionTimer()
	defer timer.ObserveTotal()

	if u.useAdmissionControllerStatus {
		isValid, err := u.statusValidator.IsStatusValid(ctx, u.statusTimeout)
		if err != nil {
			klog.ErrorS(err, "Error getting Admission Controller status. Skipping update loop")
			metrics_updater.RecordAdmissionControllerStatusInvalid("error")
			return
		}
		if !isValid {
			klog.V(0).InfoS("Admission Controller status is not valid. Skipping update loop", "timeout", u.statusTimeout)
			metrics_updater.RecordAdmissionControllerStatusInvalid("invalid")
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

	inPlaceFeatureEnabled := features.Enabled(features.InPlace)
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
			!vpa_api_util.HasStartupBoost(vpa) {
			klog.V(3).InfoS("Skipping VPA object because its mode is not  \"InPlaceOrRecreate\", \"InPlace\", \"Recreate\" or \"Auto\" and it doesn't have startupBoost configured", "vpa", klog.KObj(vpa))
			continue
		}

		// Only act on the VPA status once the recommender has observed the
		// current spec. This avoids applying recommendations computed from a
		// completely different spec (for example, a recommendation computed
		// when the update mode was Off before the update mode was changed to
		// Recreate with minAllowed/maxAllowed capping)
		if u.requireObservedGeneration && (vpa.Status.ObservedGeneration == nil || *vpa.Status.ObservedGeneration != vpa.Generation) {
			klog.V(3).InfoS("Skipping VPA object because the recommender has not observed its current generation yet", "vpa", klog.KObj(vpa), "generation", vpa.Generation, "observedGeneration", ptr.Deref(vpa.Status.ObservedGeneration, 0))
			continue
		}

		selector, err := u.selectorFetcher.Fetch(ctx, vpa)
		if err != nil {
			klog.V(3).ErrorS(err, "Skipping VPA object because we cannot fetch selector", "vpa", klog.KObj(vpa))
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

	controlledPods := make(map[*vpa_types.VerticalPodAutoscaler][]*corev1.Pod)
	livePodUIDs := set.New[types.UID]()
	for _, pod := range allLivePods {
		livePodUIDs.Insert(pod.UID)
		controllingVPA := vpa_api_util.GetControllingVPAForPod(ctx, pod, vpas, u.controllerFetcher)
		if controllingVPA != nil {
			controlledPods[controllingVPA.Vpa] = append(controlledPods[controllingVPA.Vpa], pod)
		}
	}

	// Clean up stale infeasible attempts for pods that no longer exist
	if len(u.infeasibleAttempts) > 0 {
		u.cleanupStaleInfeasibleAttempts(livePodUIDs)
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
		metrics_updater.InitCounters(vpaSize, vpa.Name, vpa.Namespace, updateMode)

		inPlaceLimiter := u.restrictionFactory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)
		withInPlaceUpdated := false

		// Exclude pods with an active CPU startup boost from standard eviction/in-place processing.
		// These pods are handled separately by the RunBoostWorker queue, which unbooosts them
		// via in-place update once the boost duration expires — avoiding unnecessary evictions.
		podsAvailableForUpdate := livePods
		if features.Enabled(features.CPUStartupBoost) && vpa_api_util.HasStartupBoost(vpa) {
			podsAvailableForUpdate = filterPods(livePods, func(pod *corev1.Pod) bool {
				return !vpa_api_util.PodHasCPUBoostInProgressAnnotation(pod)
			})
		}

		if updateMode == vpa_types.UpdateModeOff || updateMode == vpa_types.UpdateModeInitial {
			continue
		}

		evictionLimiter := u.restrictionFactory.NewPodsEvictionRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)
		podsForEviction := make([]*corev1.Pod, 0)
		podsForInPlace := make([]*corev1.Pod, 0)
		withInPlaceUpdatable := false
		withEvictable := false

		if (updateMode == vpa_types.UpdateModeInPlaceOrRecreate) || (updateMode == vpa_types.UpdateModeInPlace && inPlaceFeatureEnabled) {
			podsForInPlace = u.getPodsUpdateOrder(filterNonInPlaceUpdatablePods(podsAvailableForUpdate, inPlaceLimiter, vpa, len(livePods), u.infeasibleAttempts, u.eventRecorder), vpa)
			inPlaceUpdatablePodsCounter.Add(vpaSize, len(podsForInPlace))
			if len(podsForInPlace) > 0 {
				withInPlaceUpdatable = true
			}
		} else {
			// If the feature gate is not enabled but update mode is InPlace, updater will do nothing.
			if updateMode == vpa_types.UpdateModeInPlace {
				klog.InfoS("Warning: feature gate is not enabled for this updateMode", "featuregate", features.InPlace, "updateMode", updateMode)
				continue
			}
			// We evict the pod when the mode is set to Recreate or Auto. The latter mode is deprecated.
			podsForEviction = u.getPodsUpdateOrder(filterNonEvictablePods(podsAvailableForUpdate, evictionLimiter), vpa)
			evictablePodsCounter.Add(vpaSize, updateMode, len(podsForEviction))
			if len(podsForEviction) > 0 {
				withEvictable = true
			}
		}

		withEvicted := false

		for _, pod := range podsForInPlace {
			decision := inPlaceLimiter.CanInPlaceUpdate(pod, vpa, u.infeasibleAttempts)

			switch decision {
			case utils.InPlaceDeferred:
				// this can be happening for both InPlace and InPlaceOrRecreate
				if updateMode == vpa_types.UpdateModeInPlaceOrRecreate {
					klog.V(0).InfoS("In-place update deferred", "pod", klog.KObj(pod))
					continue
				}
				// Pod passed priority calculator, meaning recommendations differ from spec.
				// Retry the in-place update implicitly by ignoring until the next loop.
				klog.V(2).InfoS("In-place update deferred", "pod", klog.KObj(pod))
			case utils.InPlaceEvict:
				// This should only happen for InPlaceOrRecreate mode
				podsForEviction = append(podsForEviction, pod)
				continue

			case utils.InPlaceInfeasibleCached:
				// This should be unreachable — filterNonInPlaceUpdatablePods already handles InPlaceInfeasibleCached and filters those pods out.
				klog.V(2).InfoS("In-place update infeasible, no resource is lower in new recommendation, skipping pod", "pod", klog.KObj(pod))
				continue
			case utils.InPlaceInfeasible:
				// This should only happen for InPlace mode
				// Status is Infeasible, but recommendation has changed
				// (or first infeasible attempt)
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
					if isInfeasibleError(err) {
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

func (u *updater) cleanupStaleInfeasibleAttempts(livePodUIDs set.Set[types.UID]) {
	for podID := range u.infeasibleAttempts {
		if !livePodUIDs.Has(podID) {
			delete(u.infeasibleAttempts, podID)
		}
	}
}

// recordInfeasibleAttempt stores the recommendation that failed as infeasible
func (u *updater) recordInfeasibleAttempt(pod *corev1.Pod, vpa *vpa_types.VerticalPodAutoscaler) {
	u.infeasibleAttempts[pod.UID] = vpa.Status.Recommendation
	klog.V(2).InfoS("Recorded infeasible attempt, will retry when recommendation changes", "pod", klog.KObj(pod))
}

func (u *updater) PodAddHandler(obj any) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		klog.InfoS("Expected Pod", "got", obj)
		return
	}
	if vpa_api_util.IsPodReady(pod) && vpa_api_util.PodHasCPUBoostInProgressAnnotation(pod) && !slices.Contains(u.ignoredNamespaces, pod.Namespace) {
		u.enqueuePod(pod)
	}
}

func (u *updater) PodUpdateHandler(_, curObj any) {
	curPod, ok := curObj.(*corev1.Pod)
	if !ok {
		klog.InfoS("Expected Pod", "got", curObj)
		return
	}

	if vpa_api_util.IsPodReady(curPod) && vpa_api_util.PodHasCPUBoostInProgressAnnotation(curPod) && !slices.Contains(u.ignoredNamespaces, curPod.Namespace) {
		u.enqueuePod(curPod)
	}
}

func (u *updater) enqueuePod(obj any) {
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		return
	}
	u.cpuStartupBoostQueue.Add(key)
}

func (u *updater) RunBoostWorker(ctx context.Context) {
	logger := klog.FromContext(ctx).WithName("boost-worker")
	ctx = klog.NewContext(ctx, logger)
	if !cache.WaitForCacheSync(ctx.Done(), u.podInformer.HasSynced) {
		klog.ErrorS(nil, "Failed to sync pod informer cache")
		return
	}
	for u.processNextBoostItem(ctx) {
	}
	logger.Info("VPA updater is shutting down")
}

func (u *updater) ShutDown() {
	u.cpuStartupBoostQueue.ShutDown()
}

func (u *updater) processNextBoostItem(ctx context.Context) bool {
	key, quit := u.cpuStartupBoostQueue.Get()
	if quit {
		return false
	}
	defer u.cpuStartupBoostQueue.Done(key)

	logger := klog.FromContext(ctx)

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		logger.Error(err, "Failed to split key", "key", key)
		u.cpuStartupBoostQueue.Forget(key)
		return true
	}
	pod, err := u.podLister.Pods(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.V(4).Info("Pod no longer exists, skipping", "key", key)
			u.cpuStartupBoostQueue.Forget(key)
			return true
		}
		// other transient error we should retry
		logger.Error(err, "Failed to get pod", "pod", name, "namespace", namespace)
		u.cpuStartupBoostQueue.AddRateLimited(key)
		return true
	}

	logger = logger.WithValues("pod", klog.KObj(pod))

	if !vpa_api_util.PodHasCPUBoostInProgressAnnotation(pod) {
		logger.V(4).Info("Pod is not in boosted state, skipping")
		u.cpuStartupBoostQueue.Forget(key)
		return true
	}

	vpaWithSelector, err := u.getControllingVPAForPod(ctx, pod)
	if err != nil {
		logger.Error(err, "Failed to get controlling VPA")
		u.cpuStartupBoostQueue.AddRateLimited(key)
		return true
	}
	if vpaWithSelector == nil {
		if u.cpuStartupBoostQueue.NumRequeues(key) > maxVPALookupRetries {
			logger.V(2).Info("No controlling VPA found for pod after retries, deferring until next resync", "maxRetries", maxVPALookupRetries)
			u.cpuStartupBoostQueue.Forget(key)
			return true
		}
		logger.V(4).Info("No controlling VPA found for pod, re-enqueueing")
		u.cpuStartupBoostQueue.AddRateLimited(key)
		return true
	}
	vpa := vpaWithSelector.Vpa

	expiredAnnotations := vpa_api_util.GetExpiredStartupCPUBoostAnnotations(pod, vpa)
	if len(expiredAnnotations) > 0 {
		logger.V(4).Info("Found expired boost annotations", "count", len(expiredAnnotations))
		allPodsPerVPA, err := u.podLister.Pods(namespace).List(vpaWithSelector.Selector)
		if err != nil {
			logger.Error(err, "Failed to list pods for VPA", "vpa", klog.KObj(vpa))
			u.cpuStartupBoostQueue.AddRateLimited(key)
			return true
		}
		livePods := filterDeletedPods(allPodsPerVPA)
		vpaSize := len(livePods)

		creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := u.restrictionFactory.GetCreatorMaps(livePods, vpa)
		if err != nil {
			logger.Error(err, "Failed to get creator maps for unboosting")
			u.cpuStartupBoostQueue.AddRateLimited(key)
			return true
		}
		inPlaceLimiter := u.restrictionFactory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

		if !inPlaceLimiter.CanUnboost(pod, vpa) {
			logger.V(4).Info("Cannot unboost pod yet, re-enqueueing")
			u.cpuStartupBoostQueue.AddRateLimited(key)
			return true
		}

		logger.V(2).Info("Unboosting pod")
		if err := u.inPlaceRateLimiter.Wait(ctx); err != nil {
			logger.Error(err, "In-place rate limiter wait failed for unboosting")
			u.cpuStartupBoostQueue.AddRateLimited(key)
			return true
		}
		if err := inPlaceLimiter.InPlaceUpdate(pod, vpa, u.eventRecorder); err != nil {
			logger.Error(err, "Unboosting failed")
			metrics_updater.RecordFailedInPlaceUpdate(vpaSize, vpa.Name, vpa.Namespace, "UnboostError")
			u.cpuStartupBoostQueue.AddRateLimited(key)
			return true
		}
		logger.V(2).Info("Successfully unboosted pod")
		metrics_updater.AddInPlaceUpdatedPod(vpaSize, vpa.Name, vpa.Namespace)
		u.cpuStartupBoostQueue.Forget(key)
		return true
	}

	remaining := vpa_api_util.GetBoostRemainingDuration(pod, vpa)
	if remaining > 0 {
		logger.V(4).Info("Boost still active, re-enqueueing", "remainingDuration", remaining)
		u.cpuStartupBoostQueue.Forget(key)
		u.cpuStartupBoostQueue.AddAfter(key, remaining)
	} else if len(expiredAnnotations) == 0 && vpa_api_util.PodHasCPUBoostInProgressAnnotation(pod) {
		// Pod has boost annotations but neither expired nor remaining — likely the pod
		// is not yet Ready in the lister cache. Re-enqueue to retry shortly.
		logger.V(4).Info("Pod has boost annotations but not yet Ready in cache, re-enqueueing")
		u.cpuStartupBoostQueue.AddRateLimited(key)
	} else {
		logger.V(4).Info("No remaining boost duration")
		u.cpuStartupBoostQueue.Forget(key)
	}
	return true
}

// getControllingVPAForPod gets the controlling VPA for a pod.
//
// TODO(omerap12): We should revisit this.
// Every time the boost worker processes a pod, it lists ALL VPAs in the pod's namespace and fetches selectors for each.
// In clusters with many VPAs this is expensive (I know there is a lister so the list operation is in memory but maybe we can do something better).
//
// TODO(omerap12): Move this function to the utils package since it shares code with the admission controller.
func (u *updater) getControllingVPAForPod(ctx context.Context, pod *corev1.Pod) (*vpa_api_util.VpaWithSelector, error) {
	vpaList, err := u.vpaLister.VerticalPodAutoscalers(pod.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}
	vpas := make([]*vpa_api_util.VpaWithSelector, 0, len(vpaList))
	for _, vpa := range vpaList {
		selector, err := u.selectorFetcher.Fetch(ctx, vpa)
		if err != nil {
			klog.V(3).ErrorS(err, "Skipping VPA object because we cannot fetch selector", "vpa", klog.KObj(vpa))
			continue
		}
		vpas = append(vpas, &vpa_api_util.VpaWithSelector{
			Vpa:      vpa,
			Selector: selector,
		})
	}
	return vpa_api_util.GetControllingVPAForPod(ctx, pod, vpas, u.controllerFetcher), nil
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

	for _, pod := range pods {
		priorityCalculator.AddPod(pod, time.Now(), u.infeasibleAttempts)
	}

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

func filterNonInPlaceUpdatablePods(pods []*corev1.Pod, inplaceRestriction restriction.PodsInPlaceRestriction, vpa *vpa_types.VerticalPodAutoscaler, vpaCount int, infeasibleAttempts map[types.UID]*vpa_types.RecommendedPodResources, eventRecorder record.EventRecorder) []*corev1.Pod {
	return filterPods(pods, func(pod *corev1.Pod) bool {
		updateMode := *vpa.Spec.UpdatePolicy.UpdateMode
		decision := inplaceRestriction.CanInPlaceUpdate(pod, vpa, infeasibleAttempts)
		switch decision {
		case utils.InPlaceApproved:
			return true
		case utils.InPlaceInfeasible, utils.InPlaceDeferred:
			// For InPlace mode, include infeasible/deferred pods to retry (no backoff for alpha)
			// or to check if recommendation changed and apply a new patch while a previous update is in progress
			return updateMode == vpa_types.UpdateModeInPlace
		case utils.InPlaceEvict:
			// For InPlaceOrRecreate, include so they can be redirected to eviction in the loop
			return updateMode == vpa_types.UpdateModeInPlaceOrRecreate
		case utils.InPlaceInfeasibleCached:
			// Cached infeasibility means we've already determined this pod cannot be
			// updated in-place and cached the result to avoid redundant checks
			eventRecorder.Event(pod, corev1.EventTypeNormal, "InPlaceResizeSkipped", "Previously recorded in-place update was infeasible, current recommendation is not lower than previous, skipping resize")
			metrics_updater.RecordInPlaceInfeasibleCached(vpaCount, vpa.Name, vpa.Namespace)
			return false
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

// isInfeasibleError checks if an error indicates the resize is infeasible
// due to insufficient node capacity, as reported by the PodResizeValidator
// admission plugin (available from Kubernetes 1.36).
// See https://issues.k8s.io/136043
func isInfeasibleError(err error) bool {
	var statusErr *apierrors.StatusError
	if !errors.As(err, &statusErr) {
		return false
	}
	if statusErr.ErrStatus.Reason != metav1.StatusReasonForbidden {
		return false
	}
	if statusErr.ErrStatus.Details == nil {
		return false
	}
	for _, cause := range statusErr.ErrStatus.Details.Causes {
		if cause.Type == metav1.CauseType(utils.InfeasibleCauseNodeCapacity) {
			return true
		}
	}
	return false
}
