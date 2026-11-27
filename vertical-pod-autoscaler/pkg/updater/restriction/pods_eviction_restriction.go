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

package restriction

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	lifecyclev1alpha1 "k8s.io/api/lifecycle/v1alpha1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
	"k8s.io/utils/set"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

// PodsEvictionRestriction controls pods evictions. It ensures that we will not evict too
// many pods from one replica set. For replica set will allow to evict one pod or more if
// evictionToleranceFraction is configured.
type PodsEvictionRestriction interface {
	// Evict sends eviction instruction to the api client.
	// Returns error if pod cannot be evicted or if client returned error.
	Evict(pod *corev1.Pod, vpa *vpa_types.VerticalPodAutoscaler, eventRecorder record.EventRecorder) error
	// CanEvict checks if pod can be safely evicted
	CanEvict(pod *corev1.Pod) bool
}

// PodsEvictionRestrictionImpl is the implementation of the PodsEvictionRestriction interface.
type PodsEvictionRestrictionImpl struct {
	client                       kube_client.Interface
	podToReplicaCreatorMap       map[string]podReplicaCreator
	creatorToSingleGroupStatsMap map[podReplicaCreator]singleGroupStats
	clock                        clock.Clock
	lastInPlaceAttemptTimeMap    map[string]time.Time
	activeEvictionRequests       set.Set[types.UID]
	useEvictionRequest           bool
	evictionRequestRequester     string
}

// CanEvict checks if pod can be safely evicted
func (e *PodsEvictionRestrictionImpl) CanEvict(pod *corev1.Pod) bool {
	if e.activeEvictionRequests != nil && e.activeEvictionRequests.Has(pod.UID) {
		klog.V(3).InfoS("Cannot evict pod, active EvictionRequest already exists", "pod", klog.KObj(pod))
		return false
	}
	cr, present := e.podToReplicaCreatorMap[getPodID(pod)]
	if present {
		singleGroupStats, present := e.creatorToSingleGroupStatsMap[cr]
		if pod.Status.Phase == corev1.PodPending {
			return true
		}
		if present {
			if singleGroupStats.belowMinReplicas {
				klog.V(2).InfoS("Cannot evict pod, group is below minReplicas", "pod", klog.KObj(pod))
				return false
			}
			if isInPlaceUpdating(pod) {
				return CanEvictInPlacingPod(pod, singleGroupStats, e.lastInPlaceAttemptTimeMap, e.clock)
			}
			return singleGroupStats.isPodDisruptable()
		}
	}
	return false
}

// Evict sends eviction instruction to api client. Returns error if pod cannot be evicted or if client returned error
// Does not check if pod was actually evicted after eviction grace period.
func (e *PodsEvictionRestrictionImpl) Evict(podToEvict *corev1.Pod, vpa *vpa_types.VerticalPodAutoscaler, eventRecorder record.EventRecorder) error {
	cr, present := e.podToReplicaCreatorMap[getPodID(podToEvict)]
	if !present {
		return fmt.Errorf("pod not suitable for eviction %s/%s: not in replicated pods map", podToEvict.Namespace, podToEvict.Name)
	}

	if !e.CanEvict(podToEvict) {
		return fmt.Errorf("cannot evict pod %s/%s: eviction budget exceeded", podToEvict.Namespace, podToEvict.Name)
	}

	if e.useEvictionRequest {
		reqErr := e.evictViaEvictionRequest(podToEvict, vpa, eventRecorder)
		if reqErr == nil {
			if podToEvict.Status.Phase != corev1.PodPending {
				singleGroupStats, present := e.creatorToSingleGroupStatsMap[cr]
				if !present {
					return fmt.Errorf("internal error - cannot find stats for replication group %v", cr)
				}
				singleGroupStats.evicted = singleGroupStats.evicted + 1
				e.creatorToSingleGroupStatsMap[cr] = singleGroupStats
			}
			return nil
		}
		klog.V(2).InfoS("Failed to create EvictionRequest, falling back to imperative eviction", "pod", klog.KObj(podToEvict), "error", reqErr)
	}

	eviction := &policyv1.Eviction{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: podToEvict.Namespace,
			Name:      podToEvict.Name,
		},
	}
	err := e.client.CoreV1().Pods(podToEvict.Namespace).EvictV1(context.TODO(), eviction)
	if err != nil {
		klog.ErrorS(err, "Failed to evict pod", "pod", klog.KObj(podToEvict))
		return err
	}
	eventRecorder.Event(podToEvict, corev1.EventTypeNormal, "EvictedByVPA",
		"Pod was evicted by VPA Updater to apply resource recommendation.")

	eventRecorder.Event(vpa, corev1.EventTypeNormal, "EvictedPod",
		"VPA Updater evicted Pod "+podToEvict.Name+" to apply resource recommendation.")

	if podToEvict.Status.Phase != corev1.PodPending {
		singleGroupStats, present := e.creatorToSingleGroupStatsMap[cr]
		if !present {
			return fmt.Errorf("internal error - cannot find stats for replication group %v", cr)
		}
		singleGroupStats.evicted = singleGroupStats.evicted + 1
		e.creatorToSingleGroupStatsMap[cr] = singleGroupStats
	}

	return nil
}

func (e *PodsEvictionRestrictionImpl) evictViaEvictionRequest(podToEvict *corev1.Pod, vpa *vpa_types.VerticalPodAutoscaler, eventRecorder record.EventRecorder) error {
	requester := e.evictionRequestRequester
	if requester == "" {
		requester = "vpa.autoscaling.k8s.io/updater"
	}
	req := &lifecyclev1alpha1.EvictionRequest{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-vpa-", podToEvict.Name),
			Namespace:    podToEvict.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "vpa-updater",
				"vpa.autoscaling.k8s.io/name":  vpa.Name,
			},
		},
		Spec: lifecyclev1alpha1.EvictionRequestSpec{
			Target: lifecyclev1alpha1.EvictionRequestTarget{
				Pod: &lifecyclev1alpha1.EvictionRequestPodReference{
					Name: podToEvict.Name,
					UID:  podToEvict.UID,
				},
			},
			Requester: requester,
			Intent:    lifecyclev1alpha1.EvictionRequestIntentEviction,
		},
	}
	created, err := e.client.LifecycleV1alpha1().EvictionRequests(podToEvict.Namespace).Create(context.TODO(), req, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	eventRecorder.Event(podToEvict, corev1.EventTypeNormal, "EvictionRequestedByVPA",
		"EvictionRequest "+created.Name+" was created by VPA Updater to apply resource recommendation.")

	eventRecorder.Event(vpa, corev1.EventTypeNormal, "RequestedPodEviction",
		"VPA Updater requested eviction of Pod "+podToEvict.Name+" via EvictionRequest "+created.Name+".")

	if e.activeEvictionRequests == nil {
		e.activeEvictionRequests = set.New[types.UID]()
	}
	e.activeEvictionRequests.Insert(podToEvict.UID)
	return nil
}
