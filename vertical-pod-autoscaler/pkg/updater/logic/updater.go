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
	DeferredResizeUpdateTimeout = 5 * time.Minute

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
		podsForUpdate := u.getPodsUpdateOrder(filterNonEvictablePods(livePods, evictionLimiter), vpa)
		evictablePodsCounter.Add(vpaSize, len(podsForUpdate))

		withInPlaceUpdatable := false
		withInPlaceUpdated := false
		withEvictable := false
		withEvicted := false

		for _, pod := range podsForUpdate {
			if vpa_api_util.GetUpdateMode(vpa) == vpa_types.UpdateModeInPlaceOrRecreate {
				withInPlaceUpdatable = true
				fallBackToEviction, err := u.AttemptInPlaceUpdate(ctx, vpa, pod, evictionLimiter)
				if err != nil {
					klog.V(0).InfoS("In-place update failed", "error", err, "pod", klog.KObj(pod))
					return
				}
				if fallBackToEviction {
					klog.V(4).InfoS("Falling back to eviction for pod", "pod", klog.KObj(pod))
				} else {
					withInPlaceUpdated = true
					metrics_updater.AddInPlaceUpdatedPod(vpaSize)
					continue
				}
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

		if withInPlaceUpdatable {
			vpasWithInPlaceUpdatablePodsCounter.Add(vpaSize, 1)
		}
		if withInPlaceUpdated {
			vpasWithInPlaceUpdatedPodsCounter.Add(vpaSize, 1)
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

func (u *updater) AttemptInPlaceUpdate(ctx context.Context, vpa *vpa_types.VerticalPodAutoscaler, pod *apiv1.Pod, evictionLimiter eviction.PodsEvictionRestriction) (fallBackToEviction bool, err error) {
	klog.V(4).InfoS("Checking preconditions for attemping in-place update", "pod", klog.KObj(pod))
	if !evictionLimiter.CanInPlaceUpdate(pod) {
		if pod.Status.QOSClass == apiv1.PodQOSGuaranteed {
			klog.V(4).InfoS("Can't resize pod in-place due to QOSClass change, falling back to eviction", "pod", klog.KObj(pod), "qosClass", pod.Status.QOSClass)
			return true, nil
		}
		if eviction.IsInPlaceUpdating(pod) {
			lastInPlaceUpdateTime, exists := u.lastInPlaceUpdateAttemptTimeMap[eviction.GetPodID(pod)]
			if !exists {
				klog.V(4).InfoS("In-place update in progress for pod but no lastInPlaceUpdateTime found, setting it to now", "pod", klog.KObj(pod))
				lastInPlaceUpdateTime = time.Now()
				u.lastInPlaceUpdateAttemptTimeMap[eviction.GetPodID(pod)] = lastInPlaceUpdateTime
			}
			// TODO(maxcao13): fix this after 1.33 KEP changes
			// if currently inPlaceUpdating, we should only fallback to eviction if the update has failed. i.e: one of the following conditions:
			// 1. .status.resize: Infeasible
			// 2. .status.resize: Deferred + more than 1 minute has elapsed since the lastInPlaceUpdateTime
			// 3. .status.resize: InProgress + more than 1 hour has elapsed since the lastInPlaceUpdateTime
			switch pod.Status.Resize {
			case apiv1.PodResizeStatusDeferred:
				if time.Since(lastInPlaceUpdateTime) > DeferredResizeUpdateTimeout {
					klog.V(4).InfoS(fmt.Sprintf("In-place update deferred for more than %v, falling back to eviction", DeferredResizeUpdateTimeout), "pod", klog.KObj(pod))
					fallBackToEviction = true
				} else {
					klog.V(4).InfoS("In-place update deferred, NOT falling back to eviction yet", "pod", klog.KObj(pod))
				}
			case apiv1.PodResizeStatusInProgress:
				if time.Since(lastInPlaceUpdateTime) > InProgressResizeUpdateTimeout {
					klog.V(4).InfoS(fmt.Sprintf("In-place update in progress for more than %v, falling back to eviction", InProgressResizeUpdateTimeout), "pod", klog.KObj(pod))
					fallBackToEviction = true
				} else {
					klog.V(4).InfoS("In-place update in progress, NOT falling back to eviction yet", "pod", klog.KObj(pod))
				}
			case apiv1.PodResizeStatusInfeasible:
				klog.V(4).InfoS("In-place update infeasible, falling back to eviction", "pod", klog.KObj(pod))
				fallBackToEviction = true
			default:
				klog.V(4).InfoS("In-place update status unknown, falling back to eviction", "pod", klog.KObj(pod))
				fallBackToEviction = true
			}
			return fallBackToEviction, nil
		}
		klog.V(4).InfoS("Can't in-place update pod, but not falling back to eviction. Waiting for next loop", "pod", klog.KObj(pod))
		return false, nil
	}

	// TODO(jkyros): need our own rate limiter or can we freeload off the eviction one?
	err = u.evictionRateLimiter.Wait(ctx)
	if err != nil {
		klog.ErrorS(err, "Eviction rate limiter wait failed for in-place resize", "pod", klog.KObj(pod))
		return false, err
	}

	klog.V(2).InfoS("Actuating in-place update", "pod", klog.KObj(pod))
	u.lastInPlaceUpdateAttemptTimeMap[eviction.GetPodID(pod)] = time.Now()
	err = evictionLimiter.InPlaceUpdate(pod, vpa, u.eventRecorder)
	return false, err
}
