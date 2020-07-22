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

package eviction

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsinformer "k8s.io/client-go/informers/apps/v1"
	coreinformer "k8s.io/client-go/informers/core/v1"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
)

const (
	resyncPeriod time.Duration = 1 * time.Minute
)

// PodsEvictionRestriction controls pods evictions. It ensures that we will not evict too
// many pods from one replica set. For replica set will allow to evict one pod or more if
// evictionToleranceFraction is configured.
type PodsEvictionRestriction interface {
	// Evict sends eviction instruction to the api client.
	// Returns error if pod cannot be evicted or if client returned error.
	Evict(pod *apiv1.Pod, eventRecorder record.EventRecorder) error
	// CanEvict checks if pod can be safely evicted
	CanEvict(pod *apiv1.Pod) bool
}

type podsEvictionRestrictionImpl struct {
	client                       kube_client.Interface
	podToReplicaCreatorMap       map[string]podReplicaCreator
	creatorToSingleGroupStatsMap map[podReplicaCreator]singleGroupStats
}

type singleGroupStats struct {
	configured        int
	pending           int
	running           int
	evictionTolerance int
	evicted           int
}

// PodsEvictionRestrictionFactory creates PodsEvictionRestriction
type PodsEvictionRestrictionFactory interface {
	// NewPodsEvictionRestriction creates PodsEvictionRestriction for given set of pods.
	NewPodsEvictionRestriction(pods []*apiv1.Pod) PodsEvictionRestriction
}

type podsEvictionRestrictionFactoryImpl struct {
	client                    kube_client.Interface
	rcInformer                cache.SharedIndexInformer // informer for Replication Controllers
	ssInformer                cache.SharedIndexInformer // informer for Stateful Sets
	rsInformer                cache.SharedIndexInformer // informer for Replica Sets
	minReplicas               int
	evictionToleranceFraction float64
}

type controllerKind string

const (
	replicationController controllerKind = "ReplicationController"
	statefulSet           controllerKind = "StatefulSet"
	replicaSet            controllerKind = "ReplicaSet"
	job                   controllerKind = "Job"
)

type podReplicaCreator struct {
	Namespace string
	Name      string
	Kind      controllerKind
}

// CanEvict checks if pod can be safely evicted
func (e *podsEvictionRestrictionImpl) CanEvict(pod *apiv1.Pod) bool {
	cr, present := e.podToReplicaCreatorMap[getPodID(pod)]
	if present {
		singleGroupStats, present := e.creatorToSingleGroupStatsMap[cr]
		if pod.Status.Phase == apiv1.PodPending {
			return true
		}
		if present {
			shouldBeAlive := singleGroupStats.configured - singleGroupStats.evictionTolerance
			if singleGroupStats.running-singleGroupStats.evicted > shouldBeAlive {
				return true
			}
			// If all pods are running and eviction tollerance is small evict 1 pod.
			if singleGroupStats.running == singleGroupStats.configured &&
				singleGroupStats.evictionTolerance == 0 &&
				singleGroupStats.evicted == 0 {
				return true
			}
		}
	}
	return false
}

// Evict sends eviction instruction to api client. Returns error if pod cannot be evicted or if client returned error
// Does not check if pod was actually evicted after eviction grace period.
func (e *podsEvictionRestrictionImpl) Evict(podToEvict *apiv1.Pod, eventRecorder record.EventRecorder) error {
	cr, present := e.podToReplicaCreatorMap[getPodID(podToEvict)]
	if !present {
		return fmt.Errorf("pod not suitable for eviction %v : not in replicated pods map", podToEvict.Name)
	}

	if !e.CanEvict(podToEvict) {
		return fmt.Errorf("cannot evict pod %v : eviction budget exceeded", podToEvict.Name)
	}

	eviction := &policyv1.Eviction{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: podToEvict.Namespace,
			Name:      podToEvict.Name,
		},
	}
	err := e.client.CoreV1().Pods(podToEvict.Namespace).Evict(context.TODO(), eviction)
	if err != nil {
		klog.Errorf("failed to evict pod %s/%s, error: %v", podToEvict.Namespace, podToEvict.Name, err)
		return err
	}
	eventRecorder.Event(podToEvict, apiv1.EventTypeNormal, "EvictedByVPA",
		"Pod was evicted by VPA Updater to apply resource recommendation.")

	if podToEvict.Status.Phase != apiv1.PodPending {
		singleGroupStats, present := e.creatorToSingleGroupStatsMap[cr]
		if !present {
			return fmt.Errorf("Internal error - cannot find stats for replication group %v", cr)
		}
		singleGroupStats.evicted = singleGroupStats.evicted + 1
		e.creatorToSingleGroupStatsMap[cr] = singleGroupStats
	}

	return nil
}

// NewPodsEvictionRestrictionFactory creates PodsEvictionRestrictionFactory
func NewPodsEvictionRestrictionFactory(client kube_client.Interface, minReplicas int,
	evictionToleranceFraction float64) (PodsEvictionRestrictionFactory, error) {
	rcInformer, err := setUpInformer(client, replicationController)
	if err != nil {
		return nil, fmt.Errorf("Failed to create rcInformer: %v", err)
	}
	ssInformer, err := setUpInformer(client, statefulSet)
	if err != nil {
		return nil, fmt.Errorf("Failed to create ssInformer: %v", err)
	}
	rsInformer, err := setUpInformer(client, replicaSet)
	if err != nil {
		return nil, fmt.Errorf("Failed to create rsInformer: %v", err)
	}
	return &podsEvictionRestrictionFactoryImpl{
		client:                    client,
		rcInformer:                rcInformer, // informer for Replication Controllers
		ssInformer:                ssInformer, // informer for Replica Sets
		rsInformer:                rsInformer, // informer for Stateful Sets
		minReplicas:               minReplicas,
		evictionToleranceFraction: evictionToleranceFraction}, nil
}

// NewPodsEvictionRestriction creates PodsEvictionRestriction for a given set of pods.
func (f *podsEvictionRestrictionFactoryImpl) NewPodsEvictionRestriction(pods []*apiv1.Pod) PodsEvictionRestriction {
	// We can evict pod only if it is a part of replica set
	// For each replica set we can evict only a fraction of pods.
	// Evictions may be later limited by pod disruption budget if configured.

	livePods := make(map[podReplicaCreator][]*apiv1.Pod)

	for _, pod := range pods {
		creator, err := getPodReplicaCreator(pod)
		if err != nil {
			klog.Errorf("failed to obtain replication info for pod %s: %v", pod.Name, err)
			continue
		}
		if creator == nil {
			klog.Warningf("pod %s not replicated", pod.Name)
			continue
		}
		livePods[*creator] = append(livePods[*creator], pod)
	}

	podToReplicaCreatorMap := make(map[string]podReplicaCreator)
	creatorToSingleGroupStatsMap := make(map[podReplicaCreator]singleGroupStats)

	for creator, replicas := range livePods {
		actual := len(replicas)
		if actual < f.minReplicas {
			klog.V(2).Infof("too few replicas for %v %v/%v. Found %v live pods",
				creator.Kind, creator.Namespace, creator.Name, actual)
			continue
		}

		var configured int
		if creator.Kind == job {
			// Job has no replicas configuration, so we will use actual number of live pods as replicas count.
			configured = actual
		} else {
			var err error
			configured, err = f.getReplicaCount(creator)
			if err != nil {
				klog.Errorf("failed to obtain replication info for %v %v/%v. %v",
					creator.Kind, creator.Namespace, creator.Name, err)
				continue
			}
		}

		singleGroup := singleGroupStats{}
		singleGroup.configured = configured
		singleGroup.evictionTolerance = int(float64(configured) * f.evictionToleranceFraction)
		for _, pod := range replicas {
			podToReplicaCreatorMap[getPodID(pod)] = creator
			if pod.Status.Phase == apiv1.PodPending {
				singleGroup.pending = singleGroup.pending + 1
			}
		}
		singleGroup.running = len(replicas) - singleGroup.pending
		creatorToSingleGroupStatsMap[creator] = singleGroup
	}
	return &podsEvictionRestrictionImpl{
		client:                       f.client,
		podToReplicaCreatorMap:       podToReplicaCreatorMap,
		creatorToSingleGroupStatsMap: creatorToSingleGroupStatsMap}
}

func getPodReplicaCreator(pod *apiv1.Pod) (*podReplicaCreator, error) {
	creator := managingControllerRef(pod)
	if creator == nil {
		return nil, nil
	}
	podReplicaCreator := &podReplicaCreator{
		Namespace: pod.Namespace,
		Name:      creator.Name,
		Kind:      controllerKind(creator.Kind),
	}
	return podReplicaCreator, nil
}

func getPodID(pod *apiv1.Pod) string {
	if pod == nil {
		return ""
	}
	return pod.Namespace + "/" + pod.Name
}

func (f *podsEvictionRestrictionFactoryImpl) getReplicaCount(creator podReplicaCreator) (int, error) {
	switch creator.Kind {
	case replicationController:
		rcObj, exists, err := f.rcInformer.GetStore().GetByKey(creator.Namespace + "/" + creator.Name)
		if err != nil {
			return 0, fmt.Errorf("replication controller %s/%s is not available, err: %v", creator.Namespace, creator.Name, err)
		}
		if !exists {
			return 0, fmt.Errorf("replication controller %s/%s does not exist", creator.Namespace, creator.Name)
		}
		rc, ok := rcObj.(*apiv1.ReplicationController)
		if !ok {
			return 0, fmt.Errorf("Failed to parse Replication Controller")
		}
		if rc.Spec.Replicas == nil || *rc.Spec.Replicas == 0 {
			return 0, fmt.Errorf("replication controller %s/%s has no replicas config", creator.Namespace, creator.Name)
		}
		return int(*rc.Spec.Replicas), nil

	case replicaSet:
		rsObj, exists, err := f.rsInformer.GetStore().GetByKey(creator.Namespace + "/" + creator.Name)
		if err != nil {
			return 0, fmt.Errorf("replica set %s/%s is not available, err: %v", creator.Namespace, creator.Name, err)
		}
		if !exists {
			return 0, fmt.Errorf("replica set %s/%s does not exist", creator.Namespace, creator.Name)
		}
		rs, ok := rsObj.(*appsv1.ReplicaSet)
		if !ok {
			return 0, fmt.Errorf("Failed to parse Replicaset")
		}
		if rs.Spec.Replicas == nil || *rs.Spec.Replicas == 0 {
			return 0, fmt.Errorf("replica set %s/%s has no replicas config", creator.Namespace, creator.Name)
		}
		return int(*rs.Spec.Replicas), nil

	case statefulSet:
		ssObj, exists, err := f.ssInformer.GetStore().GetByKey(creator.Namespace + "/" + creator.Name)
		if err != nil {
			return 0, fmt.Errorf("stateful set %s/%s is not available, err: %v", creator.Namespace, creator.Name, err)
		}
		if !exists {
			return 0, fmt.Errorf("stateful set %s/%s does not exist", creator.Namespace, creator.Name)
		}
		ss, ok := ssObj.(*appsv1.StatefulSet)
		if !ok {
			return 0, fmt.Errorf("Failed to parse StatefulSet")
		}
		if ss.Spec.Replicas == nil || *ss.Spec.Replicas == 0 {
			return 0, fmt.Errorf("stateful set %s/%s has no replicas config", creator.Namespace, creator.Name)
		}
		return int(*ss.Spec.Replicas), nil
	}

	return 0, nil
}

func managingControllerRef(pod *apiv1.Pod) *metav1.OwnerReference {
	var managingController metav1.OwnerReference
	for _, ownerReference := range pod.ObjectMeta.GetOwnerReferences() {
		if *ownerReference.Controller {
			managingController = ownerReference
			break
		}
	}
	return &managingController
}

func setUpInformer(kubeClient kube_client.Interface, kind controllerKind) (cache.SharedIndexInformer, error) {
	var informer cache.SharedIndexInformer
	switch kind {
	case replicationController:
		informer = coreinformer.NewReplicationControllerInformer(kubeClient, apiv1.NamespaceAll,
			resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	case replicaSet:
		informer = appsinformer.NewReplicaSetInformer(kubeClient, apiv1.NamespaceAll,
			resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	case statefulSet:
		informer = appsinformer.NewStatefulSetInformer(kubeClient, apiv1.NamespaceAll,
			resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc})
	default:
		return nil, fmt.Errorf("Unknown controller kind: %v", kind)
	}
	stopCh := make(chan struct{})
	go informer.Run(stopCh)
	synced := cache.WaitForCacheSync(stopCh, informer.HasSynced)
	if !synced {
		return nil, fmt.Errorf("Failed to sync %v cache.", kind)
	}
	return informer, nil
}
