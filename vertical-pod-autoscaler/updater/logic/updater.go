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
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/updater/eviction"
	"k8s.io/autoscaler/vertical-pod-autoscaler/updater/priority"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	vpa_clientset "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/poc.autoscaling.k8s.io/v1alpha1"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
	kube_client "k8s.io/client-go/kubernetes"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/golang/glog"
)

// Updater performs updates on pods if recommended by Vertical Pod Autoscaler
type Updater interface {
	// RunOnce represents single iteration in the main-loop of Updater
	RunOnce()
}

type updater struct {
	vpaLister               vpa_lister.VerticalPodAutoscalerLister
	podLister               v1lister.PodLister
	evictionFactory         eviction.PodsEvictionRestrictionFactory
	recommendationProcessor vpa_api_util.RecommendationProcessor
}

// NewUpdater creates Updater with given configuration
func NewUpdater(kubeClient kube_client.Interface, vpaClient *vpa_clientset.Clientset, minReplicasForEvicition int, evictionToleranceFraction float64, recommendationProcessor vpa_api_util.RecommendationProcessor) Updater {
	return &updater{
		vpaLister:               vpa_api_util.NewAllVpasLister(vpaClient, make(chan struct{})),
		podLister:               newPodLister(kubeClient),
		evictionFactory:         eviction.NewPodsEvictionRestrictionFactory(kubeClient, minReplicasForEvicition, evictionToleranceFraction),
		recommendationProcessor: recommendationProcessor,
	}
}

// RunOnce represents single iteration in the main-loop of Updater
func (u *updater) RunOnce() {
	vpaList, err := u.vpaLister.List(labels.Everything())
	if err != nil {
		glog.Fatalf("failed get VPA list: %v", err)
	}

	vpas := make([]*vpa_types.VerticalPodAutoscaler, 0)

	for _, vpa := range vpaList {
		if vpa.Spec.UpdatePolicy.UpdateMode != vpa_types.UpdateModeAuto {
			glog.V(3).Infof("skipping VPA object %v because its mode is not \"Auto\"", vpa.Name)
			continue
		}
		vpas = append(vpas, vpa)
	}

	if len(vpas) == 0 {
		glog.Warningf("no VPA objects to process")
		return
	}

	podsList, err := u.podLister.List(labels.Everything())
	if err != nil {
		glog.Errorf("failed to get pods list: %v", err)
		return
	}
	livePods := filterDeletedPods(podsList)

	controlledPods := make(map[*vpa_types.VerticalPodAutoscaler][]*apiv1.Pod)
	for _, pod := range livePods {
		controllingVPA := vpa_api_util.GetControllingVPAForPod(pod, vpas)
		if controllingVPA != nil {
			controlledPods[controllingVPA] = append(controlledPods[controllingVPA], pod)
		}
	}

	for vpa, livePods := range controlledPods {
		evictionLimiter := u.evictionFactory.NewPodsEvictionRestriction(livePods)
		podsForUpdate := u.getPodsForUpdate(filterNonEvictablePods(livePods, evictionLimiter), vpa)

		for _, pod := range podsForUpdate {
			if !evictionLimiter.CanEvict(pod) {
				continue
			}
			glog.V(2).Infof("evicting pod %v", pod.Name)
			evictErr := evictionLimiter.Evict(pod)
			if evictErr != nil {
				glog.Warningf("evicting pod %v failed: %v", pod.Name, evictErr)
			}
		}
	}
}

// getPodsForUpdate returns list of pods that should be updated ordered by update priority
func (u *updater) getPodsForUpdate(pods []*apiv1.Pod, vpa *vpa_types.VerticalPodAutoscaler) []*apiv1.Pod {
	priorityCalculator := priority.NewUpdatePriorityCalculator(&vpa.Spec.ResourcePolicy, nil, u.recommendationProcessor)
	recommendation := vpa.Status.Recommendation

	for _, pod := range pods {
		priorityCalculator.AddPod(pod, &recommendation, time.Now())
	}

	return priorityCalculator.GetSortedPods()
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
