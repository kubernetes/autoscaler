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
	"time"

	"golang.org/x/time/rate"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/eviction"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/priority"
	metrics_updater "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/updater"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/status"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
)

// Updater performs updates on pods if recommended by Vertical Pod Autoscaler
type Updater interface {
	// RunOnce represents single iteration in the main-loop of Updater
	RunOnce(context.Context)
}

type updater struct {
	vpaLister                    vpa_lister.VerticalPodAutoscalerLister
	podLister                    v1lister.PodLister
	eventRecorder                record.EventRecorder
	evictionFactory              eviction.PodsEvictionRestrictionFactory
	recommendationProcessor      vpa_api_util.RecommendationProcessor
	evictionAdmission            priority.PodEvictionAdmission
	priorityProcessor            priority.PriorityProcessor
	evictionRateLimiter          *rate.Limiter
	selectorFetcher              target.VpaTargetSelectorFetcher
	useAdmissionControllerStatus bool
	statusValidator              status.Validator
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
	priorityProcessor priority.PriorityProcessor,
	namespace string,
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
		useAdmissionControllerStatus: useAdmissionControllerStatus,
		statusValidator: status.NewValidator(
			kubeClient,
			status.AdmissionControllerStatusName,
			statusNamespace,
		),
	}, nil
}

// RunOnce represents single iteration in the main-loop of Updater
func (u *updater) RunOnce(ctx context.Context) {
	timer := metrics_updater.NewExecutionTimer()
	defer timer.ObserveTotal()

	if u.useAdmissionControllerStatus {
		isValid, err := u.statusValidator.IsStatusValid(status.AdmissionControllerStatusTimeout)
		if err != nil {
			klog.Errorf("Error getting Admission Controller status: %v. Skipping eviction loop", err)
			return
		}
		if !isValid {
			klog.Warningf("Admission Controller status has been refreshed more than %v ago. Skipping eviction loop",
				status.AdmissionControllerStatusTimeout)
			return
		}
	}

	vpaList, err := u.vpaLister.List(labels.Everything())
	if err != nil {
		klog.Fatalf("failed get VPA list: %v", err)
	}
	timer.ObserveStep("ListVPAs")

	vpas := make([]*vpa_api_util.VpaWithSelector, 0)

	for _, vpa := range vpaList {
		if vpa_api_util.GetUpdateMode(vpa) != vpa_types.UpdateModeRecreate &&
			vpa_api_util.GetUpdateMode(vpa) != vpa_types.UpdateModeAuto {
			klog.V(3).Infof("skipping VPA object %v because its mode is not \"Recreate\" or \"Auto\"", vpa.Name)
			continue
		}
		selector, err := u.selectorFetcher.Fetch(vpa)
		if err != nil {
			klog.V(3).Infof("skipping VPA object %v because we cannot fetch selector", vpa.Name)
			continue
		}

		vpas = append(vpas, &vpa_api_util.VpaWithSelector{
			Vpa:      vpa,
			Selector: selector,
		})
	}

	if len(vpas) == 0 {
		klog.Warningf("no VPA objects to process")
		if u.evictionAdmission != nil {
			u.evictionAdmission.CleanUp()
		}
		return
	}

	podsList, err := u.podLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("failed to get pods list: %v", err)
		return
	}
	timer.ObserveStep("ListPods")
	allLivePods := filterDeletedPods(podsList)

	controlledPods := make(map[*vpa_types.VerticalPodAutoscaler][]*apiv1.Pod)
	for _, pod := range allLivePods {
		controllingVPA := vpa_api_util.GetControllingVPAForPod(pod, vpas)
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

	// using defer to protect against 'return' after evictionRateLimiter.Wait
	defer controlledPodsCounter.Observe()
	defer evictablePodsCounter.Observe()
	defer vpasWithEvictablePodsCounter.Observe()
	defer vpasWithEvictedPodsCounter.Observe()

	// NOTE: this loop assumes that controlledPods are filtered
	// to contain only Pods controlled by a VPA in auto or recreate mode
	for vpa, livePods := range controlledPods {
		vpaSize := len(livePods)
		controlledPodsCounter.Add(vpaSize, vpaSize)
		evictionLimiter := u.evictionFactory.NewPodsEvictionRestriction(livePods)
		podsForUpdate := u.getPodsUpdateOrder(filterNonEvictablePods(livePods, evictionLimiter), vpa)
		evictablePodsCounter.Add(vpaSize, len(podsForUpdate))

		withEvictable := false
		withEvicted := false
		for _, pod := range podsForUpdate {
			withEvictable = true
			if !evictionLimiter.CanEvict(pod) {
				continue
			}
			err := u.evictionRateLimiter.Wait(ctx)
			if err != nil {
				klog.Warningf("evicting pod %v failed: %v", pod.Name, err)
				return
			}
			klog.V(2).Infof("evicting pod %v", pod.Name)
			evictErr := evictionLimiter.Evict(pod, u.eventRecorder)
			if evictErr != nil {
				klog.Warningf("evicting pod %v failed: %v", pod.Name, evictErr)
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

func getRateLimiter(evictionRateLimit float64, evictionRateLimitBurst int) *rate.Limiter {
	var evictionRateLimiter *rate.Limiter
	if evictionRateLimit <= 0 {
		// As a special case if the rate is set to rate.Inf, the burst rate is ignored
		// see https://github.com/golang/time/blob/master/rate/rate.go#L37
		evictionRateLimiter = rate.NewLimiter(rate.Inf, 0)
		klog.V(1).Info("Rate limit disabled")
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

func filterNonEvictablePods(pods []*apiv1.Pod, evictionRestriciton eviction.PodsEvictionRestriction) []*apiv1.Pod {
	result := make([]*apiv1.Pod, 0)
	for _, pod := range pods {
		if evictionRestriciton.CanEvict(pod) {
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
	podListWatch := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "pods", apiv1.NamespaceAll, selector)
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podLister := v1lister.NewPodLister(store)
	podReflector := cache.NewReflector(podListWatch, &apiv1.Pod{}, store, time.Hour)
	stopCh := make(chan struct{})
	go podReflector.Run(stopCh)

	return podLister
}

func newEventRecorder(kubeClient kube_client.Interface) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.V(4).Infof)
	if _, isFake := kubeClient.(*fake.Clientset); !isFake {
		eventBroadcaster.StartRecordingToSink(&clientv1.EventSinkImpl{Interface: clientv1.New(kubeClient.CoreV1().RESTClient()).Events("")})
	}
	return eventBroadcaster.NewRecorder(scheme.Scheme, apiv1.EventSource{Component: "vpa-updater"})
}
