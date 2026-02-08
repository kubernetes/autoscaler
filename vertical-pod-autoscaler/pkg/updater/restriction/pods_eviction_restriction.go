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

	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

// PodsEvictionRestriction controls pods evictions. It ensures that we will not evict too
// many pods from one replica set. For replica set will allow to evict one pod or more if
// evictionToleranceFraction is configured.
type PodsEvictionRestriction interface {
	// Evict sends eviction instruction to the api client.
	// Returns error if pod cannot be evicted or if client returned error.
	Evict(pod *apiv1.Pod, vpa *vpa_types.VerticalPodAutoscaler, eventRecorder record.EventRecorder) error
	// CanEvict checks if pod can be safely evicted
	CanEvict(pod *apiv1.Pod) bool
}

// PodsEvictionRestrictionImpl is the implementation of the PodsEvictionRestriction interface.
type PodsEvictionRestrictionImpl struct {
	client                       kube_client.Interface
	podToReplicaCreatorMap       map[string]podReplicaCreator
	creatorToSingleGroupStatsMap map[podReplicaCreator]singleGroupStats
	clock                        clock.Clock
	lastInPlaceAttemptTimeMap    map[string]time.Time
}

// CanEvict checks if pod can be safely evicted
func (e *PodsEvictionRestrictionImpl) CanEvict(pod *apiv1.Pod) bool {
	cr, present := e.podToReplicaCreatorMap[getPodID(pod)]
	if present {
		singleGroupStats, present := e.creatorToSingleGroupStatsMap[cr]
		if pod.Status.Phase == apiv1.PodPending {
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
func (e *PodsEvictionRestrictionImpl) Evict(podToEvict *apiv1.Pod, vpa *vpa_types.VerticalPodAutoscaler, eventRecorder record.EventRecorder) error {
	cr, present := e.podToReplicaCreatorMap[getPodID(podToEvict)]
	if !present {
		return fmt.Errorf("pod not suitable for eviction %s/%s: not in replicated pods map", podToEvict.Namespace, podToEvict.Name)
	}

	if !e.CanEvict(podToEvict) {
		return fmt.Errorf("cannot evict pod %s/%s: eviction budget exceeded", podToEvict.Namespace, podToEvict.Name)
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
	eventRecorder.Event(podToEvict, apiv1.EventTypeNormal, "EvictedByVPA",
		"Pod was evicted by VPA Updater to apply resource recommendation.")

	eventRecorder.Event(vpa, apiv1.EventTypeNormal, "EvictedPod",
		"VPA Updater evicted Pod "+podToEvict.Name+" to apply resource recommendation.")

	if podToEvict.Status.Phase != apiv1.PodPending {
		singleGroupStats, present := e.creatorToSingleGroupStatsMap[cr]
		if !present {
			return fmt.Errorf("internal error - cannot find stats for replication group %v", cr)
		}
		singleGroupStats.evicted = singleGroupStats.evicted + 1
		e.creatorToSingleGroupStatsMap[cr] = singleGroupStats
	}

	return nil
}
