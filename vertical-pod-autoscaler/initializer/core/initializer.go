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

package core

import (
	"log"
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/apimock"
	recommender "k8s.io/autoscaler/vertical-pod-autoscaler/recommender_mock"

	"k8s.io/api/admissionregistration/v1alpha1"
	"k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kubernetes/pkg/api"
	kubeclient "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
	admissionregistrationv1alpha1 "k8s.io/kubernetes/pkg/client/clientset_generated/clientset/typed/admissionregistration/v1alpha1"

	"github.com/golang/glog"
)

const (
	// VPAInitializerConfigName is a unique name of the VPA initializer config,
	// used to register the initializer with apiserver.
	VPAInitializerConfigName = "vpa-initializer-config.k8s.io"
	// VPAInitializerName is a unique name of VPA initializer, used to mark
	// pods that need to be initialized by VPA initializer.
	VPAInitializerName = "vpa-initializer.k8s.io"

	retries     = 5
	retryPeriod = 1000 * time.Millisecond
)

// Initializer initializes all uninitialized pods matched by an existing VPA object.
// Takes resource recommendations from recommender or recommendations cached
// in VPA object if recommender is unavailable and applies them to containers
// in the pod.
// On startup initializers registers itself with API Server by creating a VPA
// initializer configuration.
type Initializer interface {
	// Run runs the initializer by syncing caches and registering with the API server
	Run(stopCh <-chan struct{})
}

type initializer struct {
	client      kubeclient.Interface
	vpaLister   apimock.VerticalPodAutoscalerLister
	podSynced   cache.InformerSynced
	informer    cache.SharedInformer
	recommender recommender.CachingRecommender
	registerer  admissionregistrationv1alpha1.InitializerConfigurationInterface
}

// Run starts and syncs the initializer's caches and registers initializer with
// the API sever.
func (initializer *initializer) Run(stopCh <-chan struct{}) {
	glog.Infof("starting VPA initializer")
	defer glog.Infof("shutting down VPA initializer")

	err := initializer.register()
	if err != nil {
		log.Fatalf("failed to register VPA initializer: %v", err)
	}
	defer initializer.unregister()
	go initializer.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, initializer.podSynced) {
		glog.Fatalf("Failed to sync caches for VPA initializer")
	}

	<-stopCh
}

func (initializer *initializer) register() error {
	configuration := newConfiguration()
	config, err := initializer.registerer.Create(configuration)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return nil
		}
		return err
	}
	glog.V(1).Infof("registered VPA initializer with config name %v", config.Name)
	return nil
}

func (initializer *initializer) unregister() {
	err := initializer.registerer.Delete(VPAInitializerConfigName, &metav1.DeleteOptions{})
	if err != nil {
		glog.Error("failed to unregister VPA initializer: %v", err)
	}
}

func (initializer *initializer) updateResourceRequests(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		glog.Errorf("can't update resource requests, object is not v1.Pod")
		return
	}

	if !shouldInitialize(pod) {
		glog.V(3).Infof("not updating pod: %v, VPA initializer not in the beginning of pending initializers list", pod.Name)
		return
	}
	initializedPod, err := initializer.initializePod(pod)

	if err != nil {
		glog.Errorf("unable to initialize pod: %v", pod.Name)
		// Mark failed initialization for pod.
		failedPod, err := markAsFailed(pod)
		if err != nil {
			glog.Errorf("unable to mark failed initialization for pod %v: %v", pod.Name, err)
			return
		}
		err = initializer.doUpdatePod(failedPod)
		if err != nil {
			glog.Errorf("unable to update pod %v after failed initialization: %v", pod.Name, err)
		}
		return
	}
	err = initializer.doUpdatePod(initializedPod)
	if err != nil {
		glog.Errorf("error updating pod %v: %v", initializedPod.Name, err)
	}
}

func (initializer *initializer) doUpdatePod(pod *v1.Pod) error {
	var err error
	for i := 0; i < retries; i++ {
		if _, err = initializer.client.CoreV1().Pods(pod.Namespace).Update(pod); err == nil {
			return nil
		}
		time.Sleep(retryPeriod)
	}
	return err
}

func markAsFailed(pod *v1.Pod) (*v1.Pod, error) {
	failedPodCopy, err := api.Scheme.Copy(pod)
	if err != nil {
		return nil, err
	}
	failedPod := failedPodCopy.(*v1.Pod)
	markInitializationFailure(failedPod)
	return failedPod, nil
}

// initializePod returns a pod in an initialized state, updating resource
// requests if applicable.
func (initializer *initializer) initializePod(pod *v1.Pod) (*v1.Pod, error) {
	glog.V(2).Infof("updating requirements for pod %v.", pod.Name)

	updatedPodCopy, err := api.Scheme.Copy(pod)
	if err != nil {
		return nil, err
	}
	updatedPod := updatedPodCopy.(*v1.Pod)
	markInitializationSuccess(updatedPod)

	vpaConfig := initializer.getMatchingVPA(pod)
	if vpaConfig == nil {
		glog.V(2).Infof("no matching VPA found for pod %v", pod.Name)
		return updatedPod, nil
	}

	recommendation, err := initializer.recommender.Get(&pod.Spec)
	if err != nil || recommendation == nil {
		if vpaConfig.Status.Recommendation != nil {
			// fallback to recommendation cached in VPA config
			recommendation = vpaConfig.Status.Recommendation
		} else {
			// no recommendation to apply
			glog.V(3).Infof("no recommendation to apply for pod %v", pod.Name)
			return updatedPod, nil
		}
	}

	glog.V(2).Infof("applying recommended resources for pod %v: %+v", pod.Name, recommendation)
	initializer.applyRecomendedResources(updatedPod, recommendation, vpaConfig.Spec.ResourcesPolicy)
	return updatedPod, nil
}

func shouldInitialize(pod *v1.Pod) bool {
	return pod.ObjectMeta.Initializers != nil && len(pod.ObjectMeta.Initializers.Pending) > 0 && pod.ObjectMeta.Initializers.Pending[0].Name == VPAInitializerName
}

// markInitializationSuccess denotes successful initialization for pod
func markInitializationSuccess(pod *v1.Pod) {
	if len(pod.ObjectMeta.Initializers.Pending) == 1 {
		pod.ObjectMeta.Initializers = nil
	} else {
		pod.ObjectMeta.Initializers.Pending = pod.ObjectMeta.Initializers.Pending[1:]
	}
}

// markInitializationFailure denotes failed initialization for pod
func markInitializationFailure(pod *v1.Pod) {
	pod.ObjectMeta.Initializers.Result = &metav1.Status{Status: metav1.StatusFailure}
}

// applyRecomendedResources overwrites pod resources Request field with recommended values.
func (initializer *initializer) applyRecomendedResources(pod *v1.Pod, recommendation *apimock.Recommendation, policy apimock.ResourcesPolicy) {
	for _, container := range pod.Spec.Containers {
		containerRecommendation := getRecommendationForContainer(recommendation, container)
		if containerRecommendation == nil {
			continue
		}
		containerPolicy := getContainerPolicy(container.Name, &policy)
		applyVPAPolicy(containerRecommendation, containerPolicy)
		for resource, recommended := range containerRecommendation.Resources {
			requested, exists := container.Resources.Requests[resource]
			if exists {
				// overwriting existing resource spec
				glog.V(2).Infof("updating resources request for pod %v container %v resource %v old value: %v new value: %v",
					pod.Name, container.Name, resource, requested, recommended)
			} else {
				// adding new resource spec
				glog.V(2).Infof("updating resources request for pod %v container %v resource %v old value: none new value: %v",
					pod.Name, container.Name, resource, recommended)
			}

			container.Resources.Requests[resource] = recommended
		}
	}

}

// applyVPAPolicy updates recommendation if recommended resources exceed limits defined in VPA resources policy
func applyVPAPolicy(recommendation *apimock.ContainerRecommendation, policy *apimock.ContainerPolicy) {
	for resourceName, recommended := range recommendation.Resources {
		if policy == nil {
			continue
		}
		resourcePolicy, found := policy.ResourcePolicy[resourceName]
		if !found {
			continue
		}
		if !resourcePolicy.Min.IsZero() && recommended.Value() < resourcePolicy.Min.Value() {
			glog.Warningf("recommendation outside of policy bounds : min value : %v recommended : %v",
				resourcePolicy.Min.Value(), recommended)
			recommendation.Resources[resourceName] = resourcePolicy.Min
		}
		if !resourcePolicy.Max.IsZero() && recommended.Value() > resourcePolicy.Max.Value() {
			glog.Warningf("recommendation outside of policy bounds : max value : %v recommended : %v",
				resourcePolicy.Max.Value(), recommended)
			recommendation.Resources[resourceName] = resourcePolicy.Max
		}
	}
}

func getRecommendationForContainer(recommendation *apimock.Recommendation, container v1.Container) *apimock.ContainerRecommendation {
	for i, containerRec := range recommendation.Containers {
		if containerRec.Name == container.Name {
			return &recommendation.Containers[i]
		}
	}
	return nil
}

func getContainerPolicy(containerName string, policy *apimock.ResourcesPolicy) *apimock.ContainerPolicy {
	if policy != nil {
		for i, container := range policy.Containers {
			if containerName == container.Name {
				return &policy.Containers[i]
			}
		}
	}
	return nil
}

// This will be cached as part of VerticalPodAutoscalerLister.
func (initializer *initializer) getMatchingVPA(pod *v1.Pod) *apimock.VerticalPodAutoscaler {
	configs, err := initializer.vpaLister.List()
	if err != nil {
		glog.Error("failed to get vpa configs: %v", err)
		return nil
	}
	for _, vpaConfig := range configs {
		selector, err := labels.Parse(vpaConfig.Spec.Target.Selector)
		if err != nil {
			continue
		}
		if selector.Matches(labels.Set(pod.GetLabels())) {
			return vpaConfig
		}
		glog.V(4).Infof("pod %v didn't match selector. Selector: %+v, labels: %+v", pod.Name, selector, pod.GetLabels())
	}
	return nil
}

// NewInitializer returns a VPA initializer.
func NewInitializer(kubeClient kubeclient.Interface, cacheTtl time.Duration) Initializer {
	i := &initializer{
		client:      kubeClient,
		vpaLister:   newVPALister(kubeClient),
		registerer:  newRegisterer(kubeClient),
		recommender: recommender.NewCachingRecommender(cacheTtl, apimock.NewRecommenderAPI()),
	}

	i.informer = newInformer(kubeClient, i.updateResourceRequests)
	i.podSynced = i.informer.HasSynced

	return i
}

func newVPALister(kubeClient kubeclient.Interface) apimock.VerticalPodAutoscalerLister {
	return apimock.NewVpaLister(kubeClient)
}

func newInformer(kubeClient kubeclient.Interface, updateResourceRequestFunc func(interface{})) cache.SharedInformer {
	selector := fields.ParseSelectorOrDie("status.phase!=" +
		string(v1.PodSucceeded) + ",status.phase!=" + string(v1.PodFailed))
	podListWatch := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.FieldSelector = selector.String()
			options.IncludeUninitialized = true
			return kubeClient.CoreV1().Pods(metav1.NamespaceAll).List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = selector.String()
			options.IncludeUninitialized = true
			return kubeClient.CoreV1().Pods(metav1.NamespaceAll).Watch(options)
		},
	}
	informer := cache.NewSharedInformer(
		podListWatch,
		&v1.Pod{},
		time.Second*0)

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: updateResourceRequestFunc,
	})
	return informer
}

func newConfiguration() *v1alpha1.InitializerConfiguration {
	// Initialize all pods.
	allPodsRule := v1alpha1.Rule{
		APIGroups:   []string{"*"},
		APIVersions: []string{"*"},
		Resources:   []string{"pods"},
	}
	// If initializer fails, allow for pod creation.
	failPolicy := v1alpha1.Ignore
	vpaInitializer := v1alpha1.Initializer{
		Name:          VPAInitializerName,
		Rules:         []v1alpha1.Rule{allPodsRule},
		FailurePolicy: &failPolicy,
	}
	configuration := &v1alpha1.InitializerConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: v1.NamespaceDefault,
			Name:      VPAInitializerConfigName,
		},
		Initializers: []v1alpha1.Initializer{vpaInitializer},
	}
	return configuration
}

func newRegisterer(kubeClient kubeclient.Interface) admissionregistrationv1alpha1.InitializerConfigurationInterface {
	return kubeClient.AdmissionregistrationV1alpha1().InitializerConfigurations()
}
