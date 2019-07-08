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
	"fmt"
	"time"

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
	RunOnce()
}

type updater struct {
	vpaLister               vpa_lister.VerticalPodAutoscalerLister
	podLister               v1lister.PodLister
	eventRecorder           record.EventRecorder
	evictionFactory         eviction.PodsEvictionRestrictionFactory
	recommendationProcessor vpa_api_util.RecommendationProcessor
	evictionAdmission       priority.PodEvictionAdmission
	selectorFetcher         target.VpaTargetSelectorFetcher
}

// NewUpdater creates Updater with given configuration
func NewUpdater(kubeClient kube_client.Interface, vpaClient *vpa_clientset.Clientset, minReplicasForEvicition int, evictionToleranceFraction float64, recommendationProcessor vpa_api_util.RecommendationProcessor, evictionAdmission priority.PodEvictionAdmission, selectorFetcher target.VpaTargetSelectorFetcher) (Updater, error) {
	factory, err := eviction.NewPodsEvictionRestrictionFactory(kubeClient, minReplicasForEvicition, evictionToleranceFraction)
	if err != nil {
		return nil, fmt.Errorf("Failed to create eviction restriction factory: %v", err)
	}
	return &updater{
		vpaLister:               vpa_api_util.NewAllVpasLister(vpaClient, make(chan struct{})),
		podLister:               newPodLister(kubeClient),
		eventRecorder:           newEventRecorder(kubeClient),
		evictionFactory:         factory,
		recommendationProcessor: recommendationProcessor,
		evictionAdmission:       evictionAdmission,
		selectorFetcher:         selectorFetcher,
	}, nil
}

// RunOnce represents single iteration in the main-loop of Updater
func (u *updater) RunOnce() {
	timer := metrics_updater.NewExecutionTimer()

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
		timer.ObserveTotal()
		return
	}

	podsList, err := u.podLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("failed to get pods list: %v", err)
		timer.ObserveTotal()
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

	for vpa, livePods := range controlledPods {
		evictionLimiter := u.evictionFactory.NewPodsEvictionRestriction(livePods)
		podsForUpdate := u.getPodsUpdateOrder(filterNonEvictablePods(livePods, evictionLimiter), vpa)

		for _, pod := range podsForUpdate {
			if !evictionLimiter.CanEvict(pod) {
				continue
			}
			klog.V(2).Infof("evicting pod %v", pod.Name)
			evictErr := evictionLimiter.Evict(pod, u.eventRecorder)
			if evictErr != nil {
				klog.Warningf("evicting pod %v failed: %v", pod.Name, evictErr)
			}
		}
	}
	timer.ObserveStep("EvictPods")
	timer.ObserveTotal()
}

// getPodsUpdateOrder returns list of pods that should be updated ordered by update priority
func (u *updater) getPodsUpdateOrder(pods []*apiv1.Pod, vpa *vpa_types.VerticalPodAutoscaler) []*apiv1.Pod {
	priorityCalculator := priority.NewUpdatePriorityCalculator(vpa.Spec.ResourcePolicy, vpa.Status.Conditions, nil, u.recommendationProcessor)
	recommendation := vpa.Status.Recommendation

	for _, pod := range pods {
		priorityCalculator.AddPod(pod, recommendation, time.Now())
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

func newPodLister(kubeClient kube_client.Interface) v1lister.PodLister {
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
