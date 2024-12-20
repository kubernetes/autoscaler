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
	"os"
	"slices"
	"time"

	"golang.org/x/time/rate"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
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
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/eviction"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/priority"
	metrics_updater "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/updater"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/status"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

const (
	// DeferredResizeUpdateTimeout defines the duration during which an in-place resize request
	// is considered deferred. If the resize is not completed within this time, it falls back to eviction.
	DeferredResizeUpdateTimeout = 1 * time.Minute

	// InProgressResizeUpdateTimeout defines the duration during which an in-place resize request
	// is considered in progress. If the resize is not completed within this time, it falls back to eviction.
	InProgressResizeUpdateTimeout = 1 * time.Hour
)

// Updater performs updates on pods if recommended by Vertical Pod Autoscaler
type Updater interface {
	// RunOnce represents single iteration in the main-loop of Updater
	RunOnce(context.Context)
}

type updater struct {
	vpaLister                       vpa_lister.VerticalPodAutoscalerLister
	podLister                       v1lister.PodLister
	eventRecorder                   record.EventRecorder
	evictionFactory                 eviction.PodsEvictionRestrictionFactory
	recommendationProcessor         vpa_api_util.RecommendationProcessor
	evictionAdmission               priority.PodEvictionAdmission
	priorityProcessor               priority.PriorityProcessor
	evictionRateLimiter             *rate.Limiter
	selectorFetcher                 target.VpaTargetSelectorFetcher
	useAdmissionControllerStatus    bool
	statusValidator                 status.Validator
	controllerFetcher               controllerfetcher.ControllerFetcher
	ignoredNamespaces               []string
	patchCalculators                []patch.Calculator
	lastInPlaceUpdateAttemptTimeMap map[string]time.Time
}

// NewUpdater creates Updater with given configuration
func NewUpdater(
	kubeClient kube_client.Interface,
	vpaClient *vpa_clientset.Clientset,
	minReplicasForEvicition int,
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
	factory, err := eviction.NewPodsEvictionRestrictionFactory(kubeClient, minReplicasForEvicition, evictionToleranceFraction)
	if err != nil {
		return nil, fmt.Errorf("Failed to create eviction restriction factory: %v", err)
	}
	return &updater{
		vpaLister:                    vpa_api_util.NewVpasLister(vpaClient, make(chan struct{}), namespace),
		podLister:                    newPodLister(kubeClient, namespace),
		eventRecorder:                newEventRecorder(kubeClient),
		evictionFactory:              factory,
		recommendationProcessor:      recommendationProcessor,
		evictionRateLimiter:          evictionRateLimiter,
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
		ignoredNamespaces:               ignoredNamespaces,
		patchCalculators:                patchCalculators,
		lastInPlaceUpdateAttemptTimeMap: make(map[string]time.Time),
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
		os.Exit(255)
	}
	timer.ObserveStep("ListVPAs")

	vpas := make([]*vpa_api_util.VpaWithSelector, 0)

	for _, vpa := range vpaList {
		if slices.Contains(u.ignoredNamespaces, vpa.Namespace) {
			klog.V(3).InfoS("Skipping VPA object in ignored namespace", "vpa", klog.KObj(vpa), "namespace", vpa.Namespace)
			continue
		}
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
	vpasWithEvictablePodsCounter := metrics_updater.NewVpasWithEvictablePodsCounter()
	vpasWithEvictedPodsCounter := metrics_updater.NewVpasWithEvictedPodsCounter()

	vpasWithInPlaceUpdatablePodsCounter := metrics_updater.NewVpasWithInPlaceUpdtateablePodsCounter()
	vpasWithInPlaceUpdatedPodsCounter := metrics_updater.NewVpasWithInPlaceUpdtatedPodsCounter()

	// using defer to protect against 'return' after evictionRateLimiter.Wait
	defer controlledPodsCounter.Observe()
	defer evictablePodsCounter.Observe()
	defer vpasWithEvictablePodsCounter.Observe()
	defer vpasWithEvictedPodsCounter.Observe()
	// separate counters for in-place
	defer vpasWithInPlaceUpdatablePodsCounter.Observe()
	defer vpasWithInPlaceUpdatedPodsCounter.Observe()

	// NOTE: this loop assumes that controlledPods are filtered
	// to contain only Pods controlled by a VPA in auto, recreate, or inPlaceOrRecreate mode
	for vpa, livePods := range controlledPods {
		vpaSize := len(livePods)
		controlledPodsCounter.Add(vpaSize, vpaSize)
		evictionLimiter := u.evictionFactory.NewPodsEvictionRestriction(livePods, vpa, u.patchCalculators)
		// TODO(jkyros): I need to know the priority details here so I can use them to determine what we want to do to the pod
		// previously it was just "evict" but now we have to make decisions, so we need to know
		// TODO(macao): Does the below filter mean that we may not be able to inplace scale pods that could be inplaced scaled, but non-evictable?
		// Seems so, and if so, we should rework this filter to allow inplace scaling for non-evictable pods (is that what John was commenting above me? :p)
		podsForUpdate := u.getPodsUpdateOrder(filterNonEvictablePods(livePods, evictionLimiter), vpa)
		evictablePodsCounter.Add(vpaSize, len(podsForUpdate))

		withEvictable := false
		withEvicted := false

		for _, prioritizedPod := range podsForUpdate {

			pod := prioritizedPod.Pod()

			fallBackToEviction, err := u.AttemptInPlaceScalingIfPossible(ctx, vpaSize, vpa, pod, evictionLimiter, vpasWithInPlaceUpdatablePodsCounter, vpasWithInPlaceUpdatedPodsCounter)
			if err != nil {
				klog.Warningf("error attemptng to scale pod %v in-place: %v", pod.Name, err)
				return
			}
			// If in-place scaling was possible, and it isn't stuck, then skip eviction
			if fallBackToEviction {
				// TODO(jkyros): this needs to be cleaner, but we absolutely need to make sure a disruptionless update doesn't "sneak through"
				if prioritizedPod.IsDisruptionless() {
					klog.Infof("Not falling back to eviction, %v was supposed to be disruptionless", pod.Name)
					continue
				} else {
					klog.V(4).Infof("Falling back to eviction for %s", pod.Name)
				}
			} else {
				klog.Infof("Not falling back to eviction, because we don't have a recommendation yet? %v", pod.Name)
				continue
			}

			withEvictable = true
			if !evictionLimiter.CanEvict(pod) {
				continue
			}
			err = u.evictionRateLimiter.Wait(ctx)
			if err != nil {
				klog.V(0).InfoS("Eviction rate limiter wait failed", "error", err)
				return
			}
			klog.V(2).InfoS("Evicting pod", "pod", klog.KObj(pod))
			evictErr := evictionLimiter.Evict(pod, vpa, u.eventRecorder)
			if evictErr != nil {
				klog.V(0).InfoS("Eviction failed", "error", evictErr, "pod", klog.KObj(pod))
			} else {
				withEvicted = true
				metrics_updater.AddEvictedPod(vpaSize)
			}
		}

		if withEvictable {
			vpasWithEvictablePodsCounter.Add(vpaSize, 1)
		}
		if withEvicted {
			vpasWithEvictedPodsCounter.Add(vpaSize, 1)
		}
	}
	timer.ObserveStep("EvictPods")
}

// VpaRecommendationProvided checks the VPA status to see if it has provided a recommendation yet. Used
// to make sure we don't get bogus values for in-place scaling
// TODO(jkyros):  take this out when you find the proper place to gate this
func VpaRecommendationProvided(vpa *vpa_types.VerticalPodAutoscaler) bool {
	// for _, condition := range vpa.Status.Conditions {
	// 	if condition.Type == vpa_types.RecommendationProvided && condition.Status == apiv1.ConditionTrue {
	// 		return true
	// 	}
	// }
	// TODO(maxcao13): The above condition doesn't work in tests because sometimes there is no recommender to set this status
	// so we should check the recommendation field directly. Or we can set the above condition manually in tests.
	return vpa.Status.Recommendation != nil
}

func getRateLimiter(evictionRateLimit float64, evictionRateLimitBurst int) *rate.Limiter {
	var evictionRateLimiter *rate.Limiter
	if evictionRateLimit <= 0 {
		// As a special case if the rate is set to rate.Inf, the burst rate is ignored
		// see https://github.com/golang/time/blob/master/rate/rate.go#L37
		evictionRateLimiter = rate.NewLimiter(rate.Inf, 0)
		klog.V(1).InfoS("Rate limit disabled")
	} else {
		evictionRateLimiter = rate.NewLimiter(rate.Limit(evictionRateLimit), evictionRateLimitBurst)
	}
	return evictionRateLimiter
}

// getPodsUpdateOrder returns list of pods that should be updated ordered by update priority
func (u *updater) getPodsUpdateOrder(pods []*apiv1.Pod, vpa *vpa_types.VerticalPodAutoscaler) []*priority.PrioritizedPod {
	priorityCalculator := priority.NewUpdatePriorityCalculator(
		vpa,
		nil,
		u.recommendationProcessor,
		u.priorityProcessor)

	for _, pod := range pods {
		priorityCalculator.AddPod(pod, time.Now())
	}

	return priorityCalculator.GetSortedPrioritizedPods(u.evictionAdmission)
}

func filterNonEvictablePods(pods []*apiv1.Pod, evictionRestriction eviction.PodsEvictionRestriction) []*apiv1.Pod {
	result := make([]*apiv1.Pod, 0)
	for _, pod := range pods {
		if evictionRestriction.CanEvict(pod) {
			result = append(result, pod)
		}
	}
	return result
}

func filterDeletedPods(pods []*apiv1.Pod) []*apiv1.Pod {
	result := make([]*apiv1.Pod, 0)
	for _, pod := range pods {
		if pod.DeletionTimestamp == nil {
			result = append(result, pod)
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
		os.Exit(255)
	}

	return eventBroadcaster.NewRecorder(vpascheme, apiv1.EventSource{Component: "vpa-updater"})
}

func (u *updater) AttemptInPlaceScalingIfPossible(ctx context.Context, vpaSize int, vpa *vpa_types.VerticalPodAutoscaler, pod *apiv1.Pod, evictionLimiter eviction.PodsEvictionRestriction, vpasWithInPlaceUpdatablePodsCounter *metrics_updater.SizeBasedGauge, vpasWithInPlaceUpdatedPodsCounter *metrics_updater.SizeBasedGauge) (fallBackToEviction bool, err error) {
	// TODO(jkyros): We're somehow jumping the gun here, I'm not sure if it's a race condition or what but evictions
	// don't hit it (maybe they take too long?). We end up with 0's for resource recommendations because we
	// queue this for in-place before the VPA has made a recommendation.

	if !VpaRecommendationProvided(vpa) {
		klog.V(4).Infof("VPA hasn't made a recommendation yet, we're early on %s for some reason", pod.Name)
		// TODO(jkyros): so we must have had some erroneous evictions before, but we were passing the test suite? But for now if I want to test
		// in-place I need it to not evict immediately if I can't in-place (because then it will never in-place)
		fallBackToEviction = false
		return
	}

	if vpa_api_util.GetUpdateMode(vpa) == vpa_types.UpdateModeInPlaceOrRecreate {

		// separate counters/stats for in-place updates
		withInPlaceUpdatable := false
		withInPlaceUpdated := false

		klog.V(4).Infof("Looks like we might be able to in-place update %s..", pod.Name)
		withInPlaceUpdatable = true
		// If I can't update
		if !evictionLimiter.CanInPlaceUpdate(pod) {
			// If we are already updating, wait for the next loop to progress
			if !eviction.IsInPlaceUpdating(pod) {
				// But it's not in-place updating, something went wrong (e.g. the operation would change pods QoS)
				// fall back to eviction
				klog.V(4).Infof("Can't in-place update pod %s, falling back to eviction, it might say no", pod.Name)
				fallBackToEviction = true
				return
			}

			lastUpdateTime, exists := u.lastInPlaceUpdateAttemptTimeMap[eviction.GetPodID(pod)]
			if !exists {
				klog.V(4).Infof("In-place update in progress for %s/%s, but no lastInPlaceUpdateTime found, setting it to now", pod.Namespace, pod.Name)
				lastUpdateTime = time.Now()
				u.lastInPlaceUpdateAttemptTimeMap[eviction.GetPodID(pod)] = lastUpdateTime
			}

			// if currently inPlaceUpdating, we should only fallback to eviction if the update has failed. i.e: one of the following conditions:
			// 1. .status.resize: Infeasible
			// 2. .status.resize: Deferred + more than 1 minute has elapsed since the lastInPlaceUpdateTime
			// 3. .status.resize: InProgress + more than 1 hour has elapsed since the lastInPlaceUpdateTime
			switch pod.Status.Resize {
			case apiv1.PodResizeStatusDeferred:
				if time.Since(lastUpdateTime) > DeferredResizeUpdateTimeout {
					klog.V(4).InfoS(fmt.Sprintf("In-place update deferred for more than %v, falling back to eviction", DeferredResizeUpdateTimeout), "pod", pod.Name)
					fallBackToEviction = true
				} else {
					klog.V(4).Infof("In-place update deferred for %s, NOT falling back to eviction yet", pod.Name)
				}
			case apiv1.PodResizeStatusInProgress:
				if time.Since(lastUpdateTime) > InProgressResizeUpdateTimeout {
					klog.V(4).InfoS(fmt.Sprintf("In-place update in progress for more than %v, falling back to eviction", InProgressResizeUpdateTimeout), "pod", pod.Name)
					fallBackToEviction = true
				} else {
					klog.V(4).InfoS("In-place update in progress, NOT falling back to eviction yet", "pod", pod.Name)
				}
			case apiv1.PodResizeStatusInfeasible:
				klog.V(4).InfoS("In-place update infeasible, falling back to eviction", "pod", pod.Name)
				fallBackToEviction = true
			default:
				klog.V(4).InfoS("In-place update status unknown, falling back to eviction", "pod", pod.Name)
				fallBackToEviction = true
			}
			return
		}

		// TODO(jkyros): need our own rate limiter or can we freeload off the eviction one?
		err = u.evictionRateLimiter.Wait(ctx)
		if err != nil {
			// TODO(jkyros): whether or not we fall back to eviction here probably depends on *why* we failed
			klog.Warningf("updating pod %v failed: %v", pod.Name, err)
			return
		}

		klog.V(2).Infof("attempting to in-place update pod %v", pod.Name)
		u.lastInPlaceUpdateAttemptTimeMap[eviction.GetPodID(pod)] = time.Now()
		evictErr := evictionLimiter.InPlaceUpdate(pod, vpa, u.eventRecorder)
		if evictErr != nil {
			klog.Warningf("updating pod %v failed: %v", pod.Name, evictErr)
		} else {
			// TODO(jkyros): come back later for stats
			withInPlaceUpdated = false
			metrics_updater.AddInPlaceUpdatedPod(vpaSize)
		}

		if withInPlaceUpdatable {
			vpasWithInPlaceUpdatablePodsCounter.Add(vpaSize, 1)
		}
		if withInPlaceUpdated {
			vpasWithInPlaceUpdatedPodsCounter.Add(vpaSize, 1)
		}

	} else {
		// If our update mode doesn't support in-place, then evict
		fallBackToEviction = true
		return
	}

	// counters for in-place update

	return
}
