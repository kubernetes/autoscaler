/*
Copyright 2025 The Kubernetes Authors.

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
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsinformer "k8s.io/client-go/informers/apps/v1"
	coreinformer "k8s.io/client-go/informers/core/v1"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/patch"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

const (
	resyncPeriod time.Duration = 1 * time.Minute
)

// ControllerKind is the type of controller that can manage a pod.
type controllerKind string

const (
	replicationController controllerKind = "ReplicationController"
	statefulSet           controllerKind = "StatefulSet"
	replicaSet            controllerKind = "ReplicaSet"
	daemonSet             controllerKind = "DaemonSet"
	job                   controllerKind = "Job"
)

type podReplicaCreator struct {
	Namespace string
	Name      string
	Kind      controllerKind
}

// PodsRestrictionFactory is a factory for creating PodsEvictionRestriction and PodsInPlaceRestriction.
type PodsRestrictionFactory interface {
	GetCreatorMaps(pods []*apiv1.Pod, vpa *vpa_types.VerticalPodAutoscaler) (map[podReplicaCreator]singleGroupStats, map[string]podReplicaCreator, error)
	NewPodsEvictionRestriction(creatorToSingleGroupStatsMap map[podReplicaCreator]singleGroupStats, podToReplicaCreatorMap map[string]podReplicaCreator) PodsEvictionRestriction
	NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap map[podReplicaCreator]singleGroupStats, podToReplicaCreatorMap map[string]podReplicaCreator) PodsInPlaceRestriction
}

// PodsRestrictionFactoryImpl is the implementation of the PodsRestrictionFactory interface.
type PodsRestrictionFactoryImpl struct {
	client                    kube_client.Interface
	rcInformer                cache.SharedIndexInformer // informer for Replication Controllers
	ssInformer                cache.SharedIndexInformer // informer for Stateful Sets
	rsInformer                cache.SharedIndexInformer // informer for Replica Sets
	dsInformer                cache.SharedIndexInformer // informer for Daemon Sets
	minReplicas               int
	evictionToleranceFraction float64
	clock                     clock.Clock
	lastInPlaceAttemptTimeMap map[string]time.Time
	patchCalculators          []patch.Calculator
}

// NewPodsRestrictionFactory creates a new PodsRestrictionFactory.
func NewPodsRestrictionFactory(client kube_client.Interface, minReplicas int, evictionToleranceFraction float64, patchCalculators []patch.Calculator) (PodsRestrictionFactory, error) {
	rcInformer, err := setupInformer(client, replicationController)
	if err != nil {
		return nil, fmt.Errorf("Failed to create rcInformer: %v", err)
	}
	ssInformer, err := setupInformer(client, statefulSet)
	if err != nil {
		return nil, fmt.Errorf("Failed to create ssInformer: %v", err)
	}
	rsInformer, err := setupInformer(client, replicaSet)
	if err != nil {
		return nil, fmt.Errorf("Failed to create rsInformer: %v", err)
	}
	dsInformer, err := setupInformer(client, daemonSet)
	if err != nil {
		return nil, fmt.Errorf("Failed to create dsInformer: %v", err)
	}
	return &PodsRestrictionFactoryImpl{
		client:                    client,
		rcInformer:                rcInformer, // informer for Replication Controllers
		ssInformer:                ssInformer, // informer for Stateful Sets
		rsInformer:                rsInformer, // informer for Replica Sets
		dsInformer:                dsInformer, // informer for Daemon Sets
		minReplicas:               minReplicas,
		evictionToleranceFraction: evictionToleranceFraction,
		clock:                     &clock.RealClock{},
		lastInPlaceAttemptTimeMap: make(map[string]time.Time),
		patchCalculators:          patchCalculators,
	}, nil
}

func (f *PodsRestrictionFactoryImpl) getReplicaCount(creator podReplicaCreator) (int, error) {
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
	case daemonSet:
		dsObj, exists, err := f.dsInformer.GetStore().GetByKey(creator.Namespace + "/" + creator.Name)
		if err != nil {
			return 0, fmt.Errorf("daemon set %s/%s is not available, err: %v", creator.Namespace, creator.Name, err)
		}
		if !exists {
			return 0, fmt.Errorf("daemon set %s/%s does not exist", creator.Namespace, creator.Name)
		}
		ds, ok := dsObj.(*appsv1.DaemonSet)
		if !ok {
			return 0, fmt.Errorf("Failed to parse DaemonSet")
		}
		if ds.Status.NumberReady == 0 {
			return 0, fmt.Errorf("daemon set %s/%s has no number ready pods", creator.Namespace, creator.Name)
		}
		return int(ds.Status.NumberReady), nil
	}
	return 0, nil
}

// GetCreatorMaps is a helper function that returns a map of pod replica creators to their single group stats
// and a map of pod ids to pod replica creator from a list of pods and it's corresponding VPA.
func (f *PodsRestrictionFactoryImpl) GetCreatorMaps(pods []*apiv1.Pod, vpa *vpa_types.VerticalPodAutoscaler) (map[podReplicaCreator]singleGroupStats, map[string]podReplicaCreator, error) {
	livePods := make(map[podReplicaCreator][]*apiv1.Pod)

	for _, pod := range pods {
		creator, err := getPodReplicaCreator(pod)
		if err != nil {
			klog.ErrorS(err, "Failed to obtain replication info for pod", "pod", klog.KObj(pod))
			continue
		}
		if creator == nil {
			klog.V(0).InfoS("Pod is not managed by any controller", "pod", klog.KObj(pod))
			continue
		}
		livePods[*creator] = append(livePods[*creator], pod)
	}

	podToReplicaCreatorMap := make(map[string]podReplicaCreator)
	creatorToSingleGroupStatsMap := make(map[podReplicaCreator]singleGroupStats)

	// Use per-VPA minReplicas if present, fall back to the global setting.
	required := f.minReplicas
	if vpa.Spec.UpdatePolicy != nil && vpa.Spec.UpdatePolicy.MinReplicas != nil {
		required = int(*vpa.Spec.UpdatePolicy.MinReplicas)
		klog.V(3).InfoS("Overriding minReplicas from global to per-VPA value", "globalMinReplicas", f.minReplicas, "vpaMinReplicas", required, "vpa", klog.KObj(vpa))
	}

	for creator, replicas := range livePods {
		actual := len(replicas)
		if actual < required {
			klog.V(2).InfoS("Too few replicas", "kind", creator.Kind, "object", klog.KRef(creator.Namespace, creator.Name), "livePods", actual, "requiredPods", required, "globalMinReplicas", f.minReplicas)
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
				klog.ErrorS(err, "Failed to obtain replication info", "kind", creator.Kind, "object", klog.KRef(creator.Namespace, creator.Name))
				continue
			}
		}

		singleGroup := singleGroupStats{}
		singleGroup.configured = configured
		singleGroup.evictionTolerance = int(float64(configured) * f.evictionToleranceFraction) // truncated
		for _, pod := range replicas {
			podToReplicaCreatorMap[getPodID(pod)] = creator
			if pod.Status.Phase == apiv1.PodPending {
				singleGroup.pending = singleGroup.pending + 1
			}
			if isInPlaceUpdating(pod) {
				singleGroup.inPlaceUpdateOngoing = singleGroup.inPlaceUpdateOngoing + 1
			}
		}
		singleGroup.running = len(replicas) - singleGroup.pending
		creatorToSingleGroupStatsMap[creator] = singleGroup

	}
	return creatorToSingleGroupStatsMap, podToReplicaCreatorMap, nil
}

// NewPodsEvictionRestriction creates a new PodsEvictionRestriction.
func (f *PodsRestrictionFactoryImpl) NewPodsEvictionRestriction(creatorToSingleGroupStatsMap map[podReplicaCreator]singleGroupStats, podToReplicaCreatorMap map[string]podReplicaCreator) PodsEvictionRestriction {
	return &PodsEvictionRestrictionImpl{
		client:                       f.client,
		podToReplicaCreatorMap:       podToReplicaCreatorMap,
		creatorToSingleGroupStatsMap: creatorToSingleGroupStatsMap,
		clock:                        f.clock,
		lastInPlaceAttemptTimeMap:    f.lastInPlaceAttemptTimeMap,
	}
}

// NewPodsInPlaceRestriction creates a new PodsInPlaceRestriction.
func (f *PodsRestrictionFactoryImpl) NewPodsInPlaceRestriction(creatorToSingleGroupStatsMap map[podReplicaCreator]singleGroupStats, podToReplicaCreatorMap map[string]podReplicaCreator) PodsInPlaceRestriction {
	return &PodsInPlaceRestrictionImpl{
		client:                       f.client,
		podToReplicaCreatorMap:       podToReplicaCreatorMap,
		creatorToSingleGroupStatsMap: creatorToSingleGroupStatsMap,
		clock:                        f.clock,
		lastInPlaceAttemptTimeMap:    f.lastInPlaceAttemptTimeMap,
		patchCalculators:             f.patchCalculators,
	}
}

func getPodID(pod *apiv1.Pod) string {
	if pod == nil {
		return ""
	}
	return pod.Namespace + "/" + pod.Name
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

func setupInformer(kubeClient kube_client.Interface, kind controllerKind) (cache.SharedIndexInformer, error) {
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
	case daemonSet:
		informer = appsinformer.NewDaemonSetInformer(kubeClient, apiv1.NamespaceAll,
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

type singleGroupStats struct {
	configured             int
	pending                int
	running                int
	evictionTolerance      int
	evicted                int
	inPlaceUpdateOngoing   int // number of pods from last loop that are still in-place updating
	inPlaceUpdateInitiated int // number of pods from the current loop that have newly requested in-place resize
}

// isPodDisruptable checks if all pods are running and eviction tolerance is small, we can
// disrupt the current pod.
func (s *singleGroupStats) isPodDisruptable() bool {
	shouldBeAlive := s.configured - s.evictionTolerance
	actuallyAlive := s.running - (s.evicted + s.inPlaceUpdateInitiated)
	return actuallyAlive > shouldBeAlive ||
		(s.configured == s.running && s.evictionTolerance == 0 && s.evicted == 0 && s.inPlaceUpdateInitiated == 0)
	// we don't want to block pods from being considered for eviction if tolerance is small and some pods are potentially stuck resizing
}

// isInPlaceUpdating checks whether or not the given pod is currently in the middle of an in-place update
func isInPlaceUpdating(podToCheck *apiv1.Pod) bool {
	return podToCheck.Status.Resize != ""
}
