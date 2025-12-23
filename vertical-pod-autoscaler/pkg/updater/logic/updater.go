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
	"time"

	"golang.org/x/time/rate"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	corescheme "k8s.io/client-go/kubernetes/scheme"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/patch"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/scheme"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/priority"
	restriction "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/restriction"
	utils "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/utils"
	metrics_updater "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/updater"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/status"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

// logDeprecationWarnings logs deprecation warnings for VPAs using deprecated modes
func logDeprecationWarnings(vpa *vpa_types.VerticalPodAutoscaler) {
	if vpa.Spec.UpdatePolicy != nil && vpa.Spec.UpdatePolicy.UpdateMode != nil &&
		*vpa.Spec.UpdatePolicy.UpdateMode == vpa_types.UpdateModeAuto {

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
	podLister                    v1lister.PodLister
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
		ignoredNamespaces: ignoredNamespaces,
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

	inPlaceFeatureEnable := features.Enabled(features.InPlaceOrRecreate)

	seenPods := sets.New[*apiv1.Pod]()

	for _, vpa := range vpaList {
		if slices.Contains(u.ignoredNamespaces, vpa.Namespace) {
			klog.V(3).InfoS("Skipping VPA object in ignored namespace", "vpa", klog.KObj(vpa), "namespace", vpa.Namespace)
			continue
		}

		// Log deprecation warnings for VPAs using deprecated modes
		logDeprecationWarnings(vpa)

		if vpa_api_util.GetUpdateMode(vpa) != vpa_types.UpdateModeRecreate &&
			vpa_api_util.GetUpdateMode(vpa) != vpa_types.UpdateModeAuto && vpa_api_util.GetUpdateMode(vpa) != vpa_types.UpdateModeInPlaceOrRecreate {
			klog.V(3).InfoS("Skipping VPA object because its mode is not  \"InPlaceOrRecreate\", \"Recreate\" or \"Auto\"", "vpa", klog.KObj(vpa))
			continue
		}
		selector, err := u.selectorFetcher.Fetch(ctx, vpa)
		if err != nil {
			klog.V(3).InfoS("Skipping VPA object because we cannot fetch selector", "vpa", klog.KObj(vpa))
			continue
		}
		podsWithSelector, err := u.podLister.List(selector)
		if err != nil {
			klog.ErrorS(err, "Failed to get pods", "selector", selector)
			continue
		}

		// handle the case of overlapping VPA selectors
		for _, pod := range podsWithSelector {
			seenPods.Insert(pod)
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

	timer.ObserveStep("ListPods")
	allLivePods := filterDeletedPodsFromSet(seenPods)

	controlledPods := make(map[*vpa_types.VerticalPodAutoscaler][]*apiv1.Pod)
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

	// NOTE: this loop assumes that controlledPods are filtered
	// to contain only Pods controlled by a VPA in auto, recreate, or inPlaceOrRecreate mode
	for vpa, livePods := range controlledPods {
		vpaSize := len(livePods)
		updateMode := vpa_api_util.GetUpdateMode(vpa)
		controlledPodsCounter.Add(vpaSize, updateMode, vpaSize)
		creatorToSingleGroupStatsMap, podToReplicaCreatorMap, err := u.restrictionFactory.GetCreatorMaps(livePods, vpa)
		if err != nil {
			klog.ErrorS(err, "Failed to get creator maps")
			continue
		}

		evictionLimiter := u.restrictionFactory.NewPodsEvictionRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)
		inPlaceLimiter := u.restrictionFactory.NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap, podToReplicaCreatorMap)

		podsForInPlace := make([]*apiv1.Pod, 0)
		podsForEviction := make([]*apiv1.Pod, 0)

		if updateMode == vpa_types.UpdateModeInPlaceOrRecreate && inPlaceFeatureEnable {
			podsForInPlace = u.getPodsUpdateOrder(filterNonInPlaceUpdatablePods(livePods, inPlaceLimiter), vpa)
			inPlaceUpdatablePodsCounter.Add(vpaSize, len(podsForInPlace))
		} else {
			// If the feature gate is not enabled but update mode is InPlaceOrRecreate, updater will always fallback to eviction.
			if updateMode == vpa_types.UpdateModeInPlaceOrRecreate {
				klog.InfoS("Warning: feature gate is not enabled for this updateMode", "featuregate", features.InPlaceOrRecreate, "updateMode", vpa_types.UpdateModeInPlaceOrRecreate)
			}
			podsForEviction = u.getPodsUpdateOrder(filterNonEvictablePods(livePods, evictionLimiter), vpa)
			evictablePodsCounter.Add(vpaSize, updateMode, len(podsForEviction))
		}

		withInPlaceUpdatable := false
		withInPlaceUpdated := false
		withEvictable := false
		withEvicted := false

		for _, pod := range podsForInPlace {
			withInPlaceUpdatable = true
			decision := inPlaceLimiter.CanInPlaceUpdate(pod)

			if decision == utils.InPlaceDeferred {
				klog.V(0).InfoS("In-place update deferred", "pod", klog.KObj(pod))
				continue
			} else if decision == utils.InPlaceEvict {
				podsForEviction = append(podsForEviction, pod)
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
				klog.V(0).InfoS("In-place resize failed, falling back to eviction", "error", err, "pod", klog.KObj(pod))
				metrics_updater.RecordFailedInPlaceUpdate(vpaSize, vpa.Name, vpa.Namespace, "InPlaceUpdateError")
				podsForEviction = append(podsForEviction, pod)
				continue
			}
			withInPlaceUpdated = true
			metrics_updater.AddInPlaceUpdatedPod(vpaSize, vpa.Name, vpa.Namespace)
		}

		for _, pod := range podsForEviction {
			withEvictable = true
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
func (u *updater) getPodsUpdateOrder(pods []*apiv1.Pod, vpa *vpa_types.VerticalPodAutoscaler) []*apiv1.Pod {
	priorityCalculator := priority.NewUpdatePriorityCalculator(
		vpa,
		nil,
		u.recommendationProcessor,
		u.priorityProcessor)

	for _, pod := range pods {
		priorityCalculator.AddPod(pod, time.Now())
	}

	return priorityCalculator.GetSortedPods(u.evictionAdmission)
}

func filterPods(pods []*apiv1.Pod, predicate func(*apiv1.Pod) bool) []*apiv1.Pod {
	result := make([]*apiv1.Pod, 0)
	for _, pod := range pods {
		if predicate(pod) {
			result = append(result, pod)
		}
	}
	return result
}

func filterNonInPlaceUpdatablePods(pods []*apiv1.Pod, inplaceRestriction restriction.PodsInPlaceRestriction) []*apiv1.Pod {
	return filterPods(pods, func(pod *apiv1.Pod) bool {
		return inplaceRestriction.CanInPlaceUpdate(pod) != utils.InPlaceDeferred
	})
}

func filterNonEvictablePods(pods []*apiv1.Pod, evictionRestriction restriction.PodsEvictionRestriction) []*apiv1.Pod {
	return filterPods(pods, evictionRestriction.CanEvict)
}

func filterDeletedPodsFromSet(pods sets.Set[*apiv1.Pod]) []*apiv1.Pod {
	result := make([]*apiv1.Pod, 0)
	for p := range pods {
		if p.DeletionTimestamp == nil {
			result = append(result, p)
		}
	}
	return result
}

func newPodLister(kubeClient kube_client.Interface, namespace string) v1lister.PodLister {
	selector := fields.ParseSelectorOrDie("spec.nodeName!=" + "" + ",status.phase!=" +
		string(apiv1.PodSucceeded) + ",status.phase!=" + string(apiv1.PodFailed))
	podListWatch := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "pods", namespace, selector)
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podLister := v1lister.NewPodLister(store)
	podReflector := cache.NewReflector(podListWatch, &apiv1.Pod{}, store, time.Hour)
	stopCh := make(chan struct{})
	go podReflector.Run(stopCh)

	return podLister
}

func newEventRecorder(kubeClient kube_client.Interface) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(4)
	if _, isFake := kubeClient.(*fake.Clientset); !isFake {
		eventBroadcaster.StartRecordingToSink(&clientv1.EventSinkImpl{Interface: clientv1.New(kubeClient.CoreV1().RESTClient()).Events("")})
	} else {
		eventBroadcaster.StartRecordingToSink(&clientv1.EventSinkImpl{Interface: kubeClient.CoreV1().Events("")})
	}

	vpascheme := scheme.Scheme
	if err := corescheme.AddToScheme(vpascheme); err != nil {
		klog.ErrorS(err, "Error adding core scheme")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	return eventBroadcaster.NewRecorder(vpascheme, apiv1.EventSource{Component: "vpa-updater"})
}
