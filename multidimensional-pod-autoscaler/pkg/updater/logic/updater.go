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

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	mpa_types "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1alpha1"
	mpa_clientset "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/client/clientset/versioned"
	mpa_lister "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1alpha1"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/target"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/updater/eviction"
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/updater/priority"
	mpa_api_util "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/utils/mpa"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	metrics_updater "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/updater"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/status"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/kubernetes/scheme"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	v1lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
)

// Updater performs updates on pods if recommended by Multidimensional Pod Autoscaler
type Updater interface {
	// RunOnce represents single iteration in the main-loop of Updater.
	RunOnce(context.Context)

	// RunOnceUpdatingDeployment represents single iteration in the main-loop of Updater
	// which does pod eviction and deployment updates.
	RunOnceUpdatingDeployment(context.Context)
}

type updater struct {
	mpaLister                    mpa_lister.MultidimPodAutoscalerLister
	podLister                    v1lister.PodLister
	eventRecorder                record.EventRecorder
	evictionFactory              eviction.PodsEvictionRestrictionFactory
	recommendationProcessor      mpa_api_util.RecommendationProcessor
	evictionAdmission            priority.PodEvictionAdmission
	priorityProcessor            priority.PriorityProcessor
	evictionRateLimiter          *rate.Limiter
	selectorFetcher              target.MpaTargetSelectorFetcher
	useAdmissionControllerStatus bool
	statusValidator              status.Validator
}

// NewUpdater creates Updater with given configuration
func NewUpdater(
	kubeClient kube_client.Interface,
	mpaClient *mpa_clientset.Clientset,
	minReplicasForEvicition int,
	evictionRateLimit float64,
	evictionRateBurst int,
	evictionToleranceFraction float64,
	useAdmissionControllerStatus bool,
	statusNamespace string,
	recommendationProcessor mpa_api_util.RecommendationProcessor,
	evictionAdmission priority.PodEvictionAdmission,
	selectorFetcher target.MpaTargetSelectorFetcher,
	priorityProcessor priority.PriorityProcessor,
	namespace string,
) (Updater, error) {
	evictionRateLimiter := getRateLimiter(evictionRateLimit, evictionRateBurst)
	factory, err := eviction.NewPodsEvictionRestrictionFactory(kubeClient, minReplicasForEvicition, evictionToleranceFraction)
	if err != nil {
		return nil, fmt.Errorf("Failed to create eviction restriction factory: %v", err)
	}
	return &updater{
		mpaLister:                    mpa_api_util.NewMpasLister(mpaClient, make(chan struct{}), namespace),
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

	mpaList, err := u.mpaLister.List(labels.Everything())
	if err != nil {
		klog.Fatalf("failed get MPA list: %v", err)
	}
	timer.ObserveStep("ListMPAs")
	klog.V(4).Infof("Retrieved all MPA objects.")

	mpas := make([]*mpa_api_util.MpaWithSelector, 0)

	for _, mpa := range mpaList {
		if mpa_api_util.GetUpdateMode(mpa) != vpa_types.UpdateModeRecreate &&
			mpa_api_util.GetUpdateMode(mpa) != vpa_types.UpdateModeAuto {
			klog.V(3).Infof("skipping MPA object %v because its mode is not \"Recreate\" or \"Auto\"", mpa.Name)
			continue
		}
		selector, err := u.selectorFetcher.Fetch(mpa)
		if err != nil {
			klog.V(3).Infof("skipping MPA object %v because we cannot fetch selector", mpa.Name)
			continue
		}

		mpas = append(mpas, &mpa_api_util.MpaWithSelector{
			Mpa:      mpa,
			Selector: selector,
		})
	}

	if len(mpas) == 0 {
		klog.Warningf("no MPA objects to process")
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
	klog.V(4).Infof("Retrieved all live pods.")

	controlledPods := make(map[*mpa_types.MultidimPodAutoscaler][]*apiv1.Pod)
	for _, pod := range allLivePods {
		controllingMPA := mpa_api_util.GetControllingMPAForPod(pod, mpas)
		if controllingMPA != nil {
			controlledPods[controllingMPA.Mpa] = append(controlledPods[controllingMPA.Mpa], pod)
		}
	}
	timer.ObserveStep("FilterPods")
	klog.V(4).Infof("Matched the MPA object for all pods.")

	if u.evictionAdmission != nil {
		u.evictionAdmission.LoopInit(allLivePods, controlledPods)
	}
	timer.ObserveStep("AdmissionInit")

	// wrappers for metrics which are computed every loop run
	controlledPodsCounter := metrics_updater.NewControlledPodsCounter()
	evictablePodsCounter := metrics_updater.NewEvictablePodsCounter()
	mpasWithEvictablePodsCounter := metrics_updater.NewVpasWithEvictablePodsCounter()
	mpasWithEvictedPodsCounter := metrics_updater.NewVpasWithEvictedPodsCounter()

	// using defer to protect against 'return' after evictionRateLimiter.Wait
	defer controlledPodsCounter.Observe()
	defer evictablePodsCounter.Observe()
	defer mpasWithEvictablePodsCounter.Observe()
	defer mpasWithEvictedPodsCounter.Observe()

	// NOTE: this loop assumes that controlledPods are filtered
	// to contain only Pods controlled by a MPA in auto or recreate mode
	for mpa, livePods := range controlledPods {
		mpaSize := len(livePods)
		controlledPodsCounter.Add(mpaSize, mpaSize)
		evictionLimiter := u.evictionFactory.NewPodsEvictionRestriction(livePods, mpa)
		podsForUpdate := u.getPodsUpdateOrder(filterNonEvictablePods(livePods, evictionLimiter), mpa)
		evictablePodsCounter.Add(mpaSize, len(podsForUpdate))

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
				metrics_updater.AddEvictedPod(mpaSize)
			}
		}

		if withEvictable {
			mpasWithEvictablePodsCounter.Add(mpaSize, 1)
		}
		if withEvicted {
			mpasWithEvictedPodsCounter.Add(mpaSize, 1)
		}
	}
	timer.ObserveStep("EvictPods")
	klog.V(4).Infof("Evicted all eligible pods.")
}

// RunOnceUpdatingDeployment represents single iteration in the main-loop of Updater which evicts
// pods for VPA and updates the Deployment for HPA.
func (u *updater) RunOnceUpdatingDeployment(ctx context.Context) {
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

	mpaList, err := u.mpaLister.List(labels.Everything())
	if err != nil {
		klog.Fatalf("failed get MPA list: %v", err)
	}
	timer.ObserveStep("ListMPAs")
	klog.V(4).Infof("Retrieved all MPA objects.")

	mpas := make([]*mpa_api_util.MpaWithSelector, 0)

	for _, mpa := range mpaList {
		if mpa_api_util.GetUpdateMode(mpa) != vpa_types.UpdateModeRecreate &&
			mpa_api_util.GetUpdateMode(mpa) != vpa_types.UpdateModeAuto {
			klog.V(3).Infof("skipping MPA object %v because its mode is not \"Recreate\" or \"Auto\"", mpa.Name)
			continue
		}
		selector, err := u.selectorFetcher.Fetch(mpa)
		if err != nil {
			klog.V(3).Infof("skipping MPA object %v because we cannot fetch selector", mpa.Name)
			continue
		}

		mpas = append(mpas, &mpa_api_util.MpaWithSelector{
			Mpa:      mpa,
			Selector: selector,
		})

		// Update the number of replicas.
		targetGV, err := schema.ParseGroupVersion(mpa.Spec.ScaleTargetRef.APIVersion)
		if err != nil {
			klog.Errorf("%s: FailedGetScale - error: %v", v1.EventTypeWarning, err.Error())
			u.eventRecorder.Event(mpa, v1.EventTypeWarning, "FailedGetScale", err.Error())
			return
		}
		fmt.Printf("targetGV %v targetGV.Group: %v\n", targetGV, targetGV.Group)
		targetGK := schema.GroupKind{
			Group: targetGV.Group,
			Kind:  mpa.Spec.ScaleTargetRef.Kind,
		}
		fmt.Printf("targetGK: %v\n", targetGK)
		mappings, err := u.selectorFetcher.GetRESTMappings(targetGK)
		fmt.Printf("mapping: %v\n", mappings)
		if err != nil {
			klog.Errorf("%s: FailedGetScale - error: %v", v1.EventTypeWarning, err.Error())
			u.eventRecorder.Event(mpa, v1.EventTypeWarning, "FailedGetScale", err.Error())
			return
		}
		scale, targetGR, err := u.scaleForResourceMappings(ctx, mpa.Namespace, mpa.Spec.ScaleTargetRef.Name, mappings)
		if err != nil {
			klog.Errorf("%s: FailedGetScale - error: %v", v1.EventTypeWarning, err.Error())
			u.eventRecorder.Event(mpa, v1.EventTypeWarning, "FailedGetScale", err.Error())
			return
		}
		desiredReplicas := mpa.Status.DesiredReplicas
		if (desiredReplicas == scale.Spec.Replicas) {
			// No need to update the number of replicas.
			klog.V(4).Infof("No need to change the number of replicas for MPA %v", mpa.Name)
			continue
		} else if (desiredReplicas > *mpa.Spec.Constraints.MaxReplicas || desiredReplicas < *mpa.Spec.Constraints.MinReplicas) {
			// Constraints not satisfied. Should not be out of bound because it should have been
			// checked in the recommender.
			continue
		} else {
			klog.V(4).Infof("Updating the number of replicas from %d to %d for MPA %v", scale.Spec.Replicas, desiredReplicas, mpa.Name)
			scale.Spec.Replicas = desiredReplicas
			_, err = u.selectorFetcher.Scales(mpa.Namespace).Update(ctx, targetGR, scale, metav1.UpdateOptions{})
			if err != nil {
				u.eventRecorder.Eventf(mpa, v1.EventTypeWarning, "FailedRescale", "New size: %d; error: %v", desiredReplicas, err.Error())
				klog.Errorf("%s: FailedRescale - New size: %d; error: %v", v1.EventTypeWarning, desiredReplicas, err.Error())
				return
			}
			klog.V(4).Infof("%s: Successfully rescaled the number of replicas to %d based on MPA %v", v1.EventTypeNormal, desiredReplicas, mpa.Name)
		}
	}

	if len(mpas) == 0 {
		klog.Warningf("no MPA objects to process")
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
	klog.V(4).Infof("Retrieved all live pods.")

	controlledPods := make(map[*mpa_types.MultidimPodAutoscaler][]*apiv1.Pod)
	for _, pod := range allLivePods {
		controllingMPA := mpa_api_util.GetControllingMPAForPod(pod, mpas)
		if controllingMPA != nil {
			controlledPods[controllingMPA.Mpa] = append(controlledPods[controllingMPA.Mpa], pod)
		}
	}
	timer.ObserveStep("FilterPods")
	klog.V(4).Infof("Matched the MPA object for all pods.")

	if u.evictionAdmission != nil {
		u.evictionAdmission.LoopInit(allLivePods, controlledPods)
	}
	timer.ObserveStep("AdmissionInit")

	// wrappers for metrics which are computed every loop run
	controlledPodsCounter := metrics_updater.NewControlledPodsCounter()
	evictablePodsCounter := metrics_updater.NewEvictablePodsCounter()
	mpasWithEvictablePodsCounter := metrics_updater.NewVpasWithEvictablePodsCounter()
	mpasWithEvictedPodsCounter := metrics_updater.NewVpasWithEvictedPodsCounter()

	// using defer to protect against 'return' after evictionRateLimiter.Wait
	defer controlledPodsCounter.Observe()
	defer evictablePodsCounter.Observe()
	defer mpasWithEvictablePodsCounter.Observe()
	defer mpasWithEvictedPodsCounter.Observe()

	// NOTE: this loop assumes that controlledPods are filtered
	// to contain only Pods controlled by a MPA in auto or recreate mode
	for mpa, livePods := range controlledPods {
		mpaSize := len(livePods)
		controlledPodsCounter.Add(mpaSize, mpaSize)
		evictionLimiter := u.evictionFactory.NewPodsEvictionRestriction(livePods, mpa)
		podsForUpdate := u.getPodsUpdateOrder(filterNonEvictablePods(livePods, evictionLimiter), mpa)
		evictablePodsCounter.Add(mpaSize, len(podsForUpdate))

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
				metrics_updater.AddEvictedPod(mpaSize)
			}
		}

		if withEvictable {
			mpasWithEvictablePodsCounter.Add(mpaSize, 1)
		}
		if withEvicted {
			mpasWithEvictedPodsCounter.Add(mpaSize, 1)
		}
	}
	timer.ObserveStep("EvictPods")
	klog.V(4).Infof("Evicted all eligible pods.")
}

// scaleForResourceMappings attempts to fetch the scale for the resource with the given name and
// namespace, trying each RESTMapping in turn until a working one is found. If none work, the first
// error is returned. It returns both the scale, as well as the group-resource from the working
// mapping.
func (u *updater) scaleForResourceMappings(ctx context.Context, namespace, name string, mappings []*apimeta.RESTMapping) (*autoscalingv1.Scale, schema.GroupResource, error) {
	var firstErr error
	for i, mapping := range mappings {
		targetGR := mapping.Resource.GroupResource()
		scale, err := u.selectorFetcher.Scales(namespace).Get(ctx, targetGR, name, metav1.GetOptions{})
		if err == nil {
			return scale, targetGR, nil
		}

		// if this is the first error, remember it,
		// then go on and try other mappings until we find a good one
		if i == 0 {
			firstErr = err
		}
	}

	// make sure we handle an empty set of mappings
	if firstErr == nil {
		firstErr = fmt.Errorf("unrecognized resource")
	}

	return nil, schema.GroupResource{}, firstErr
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
func (u *updater) getPodsUpdateOrder(pods []*apiv1.Pod, mpa *mpa_types.MultidimPodAutoscaler) []*apiv1.Pod {
	priorityCalculator := priority.NewUpdatePriorityCalculator(
		mpa,
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
	eventBroadcaster.StartLogging(klog.V(4).Infof)
	if _, isFake := kubeClient.(*fake.Clientset); !isFake {
		eventBroadcaster.StartRecordingToSink(&clientv1.EventSinkImpl{Interface: clientv1.New(kubeClient.CoreV1().RESTClient()).Events("")})
	}
	return eventBroadcaster.NewRecorder(scheme.Scheme, apiv1.EventSource{Component: "mpa-updater"})
}
