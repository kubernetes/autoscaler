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

package main

import (
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/updater/apimock"
	"k8s.io/autoscaler/vertical-pod-autoscaler/updater/eviction"
	"k8s.io/autoscaler/vertical-pod-autoscaler/updater/priority"
	"k8s.io/autoscaler/vertical-pod-autoscaler/updater/recommender"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	kube_client "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
	v1lister "k8s.io/kubernetes/pkg/client/listers/core/v1"
)

// Updater performs updates on pods if recommended by Vertical Pod Autoscaler
type Updater interface {
	// RunOnce represents single iteration in the main-loop of Updater
	RunOnce()
}

type updater struct {
	vpaLister        apimock.VerticalPodAutoscalerLister // wait for VPA api
	podLister        v1lister.PodLister
	recommender      recommender.CachingRecommender
	evictionFactrory eviction.PodsEvictionRestrictionFactory
}

// NewUpdater creates Updater with given configuration
func NewUpdater(kubeClient kube_client.Interface, cacheTTl time.Duration, minReplicasForEvicition int, evictionToleranceFraction float64) Updater {
	return &updater{
		vpaLister:        newVpaLister(kubeClient),
		podLister:        newPodLister(kubeClient),
		recommender:      recommender.NewCachingRecommender(cacheTTl, apimock.NewRecommenderAPI()),
		evictionFactrory: eviction.NewPodsEvictionRestrictionFactory(kubeClient, minReplicasForEvicition, evictionToleranceFraction),
	}
}

// RunOnce represents single iteration in the main-loop of Updater
func (u *updater) RunOnce() {
	vpaList, err := u.vpaLister.List()
	if err != nil {
		glog.Fatalf("failed get VPA list: %v", err)
	}

	if len(vpaList) == 0 {
		glog.Warningf("no VPA objects to process")
		return
	}

	for _, vpa := range vpaList {
		glog.V(2).Infof("processing VPA object targetting %v", vpa.Spec.Target.Selector)
		selector, err := labels.Parse(vpa.Spec.Target.Selector)
		if err != nil {
			glog.Errorf("error processing VPA object: failed to create pod selector: %v", err)
			continue
		}

		podsList, err := u.podLister.List(selector)
		if err != nil {
			glog.Errorf("failed get pods list for selector %v: %v", selector, err)
			continue
		}

		livePods := filterDeletedPods(podsList)
		if len(livePods) == 0 {
			glog.Warningf("no live pods matching selector %v", selector)
			continue
		}

		evictionLimiter := u.evictionFactrory.NewPodsEvictionRestriction(livePods)
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
func (u *updater) getPodsForUpdate(pods []*apiv1.Pod, vpa *apimock.VerticalPodAutoscaler) []*apiv1.Pod {
	priorityCalculator := priority.NewUpdatePriorityCalculator(&vpa.Spec.ResourcesPolicy, nil)

	for _, pod := range pods {
		recommendation, err := u.recommender.Get(&pod.Spec)
		if err != nil {
			glog.Errorf("error while getting recommendation for pod %v: %v", pod.Name, err)
			continue
		}

		if recommendation == nil {
			if len(vpa.Status.Recommendation.Containers) == 0 {
				glog.Warningf("no recommendation for pod: %v", pod.Name)
				continue
			}

			glog.Warningf("fallback to default VPA recommendation for pod: %v", pod.Name)
			recommendation = vpa.Status.Recommendation
		}

		priorityCalculator.AddPod(pod, recommendation)
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

func newVpaLister(kubeClient kube_client.Interface) apimock.VerticalPodAutoscalerLister {
	return apimock.NewVpaLister(kubeClient)
}

func newPodLister(kubeClient kube_client.Interface) v1lister.PodLister {
	selector := fields.ParseSelectorOrDie("spec.nodeName!=" + "" + ",status.phase!=" +
		string(apiv1.PodSucceeded) + ",status.phase!=" + string(apiv1.PodFailed))
	podListWatch := cache.NewListWatchFromClient(kubeClient.CoreV1().RESTClient(), "pods", apiv1.NamespaceAll, selector)
	store := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	podLister := v1lister.NewPodLister(store)
	podReflector := cache.NewReflector(podListWatch, &apiv1.Pod{}, store, time.Hour)
	podReflector.Run()

	return podLister
}
