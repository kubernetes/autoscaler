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
	"encoding/json"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	appsinformer "k8s.io/client-go/informers/apps/v1"
	coreinformer "k8s.io/client-go/informers/core/v1"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"

	resource_updates "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/admission-controller/resource/pod/patch"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
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
	Evict(pod *apiv1.Pod, vpa *vpa_types.VerticalPodAutoscaler, eventRecorder record.EventRecorder) error
	// CanEvict checks if pod can be safely evicted
	CanEvict(pod *apiv1.Pod) bool

	// InPlaceUpdate updates the pod resources in-place
	InPlaceUpdate(pod *apiv1.Pod, vpa *vpa_types.VerticalPodAutoscaler, eventRecorder record.EventRecorder) error
	// CanEvict checks if pod can be safely evicted
	CanInPlaceUpdate(pod *apiv1.Pod) bool
}

type podsEvictionRestrictionImpl struct {
	client                       kube_client.Interface
	podToReplicaCreatorMap       map[string]podReplicaCreator
	creatorToSingleGroupStatsMap map[podReplicaCreator]singleGroupStats
	patchCalculators             []patch.Calculator
}

type singleGroupStats struct {
	configured        int
	pending           int
	running           int
	evictionTolerance int
	evicted           int
	inPlaceUpdating   int
}

// PodsEvictionRestrictionFactory creates PodsEvictionRestriction
type PodsEvictionRestrictionFactory interface {
	// NewPodsEvictionRestriction creates PodsEvictionRestriction for given set of pods,
	// controlled by a single VPA object.
	NewPodsEvictionRestriction(pods []*apiv1.Pod, vpa *vpa_types.VerticalPodAutoscaler, patchCalculators []patch.Calculator) PodsEvictionRestriction
}

type podsEvictionRestrictionFactoryImpl struct {
	client                    kube_client.Interface
	rcInformer                cache.SharedIndexInformer // informer for Replication Controllers
	ssInformer                cache.SharedIndexInformer // informer for Stateful Sets
	rsInformer                cache.SharedIndexInformer // informer for Replica Sets
	dsInformer                cache.SharedIndexInformer // informer for Daemon Sets
	minReplicas               int
	evictionToleranceFraction float64
}

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

// CanEvict checks if pod can be safely evicted
func (e *podsEvictionRestrictionImpl) CanEvict(pod *apiv1.Pod) bool {
	cr, present := e.podToReplicaCreatorMap[GetPodID(pod)]
	if present {
		singleGroupStats, present := e.creatorToSingleGroupStatsMap[cr]
		if pod.Status.Phase == apiv1.PodPending {
			return true
		}
		if present {

			shouldBeAlive := singleGroupStats.configured - singleGroupStats.evictionTolerance
			// TODO(jkyros): Come back and think through this better, but for now, take in-place updates into account because
			// they might cause disruption. We assume pods will not be both in-place updated and evicted in the same pass, but
			// we need eviction to take the numbers into account so we don't violate our disruption dolerances.
			// If we're already resizing this pod, don't do anything to it, unless we failed to resize it, then we want to evict it.
			if IsInPlaceUpdating(pod) {
				klog.V(4).InfoS("Pod disruption tolerance",
					"pod", pod.Name,
					"running", singleGroupStats.running,
					"configured", singleGroupStats.configured,
					"tolerance", singleGroupStats.evictionTolerance,
					"evicted", singleGroupStats.evicted,
					"updating", singleGroupStats.inPlaceUpdating)

				if singleGroupStats.running-(singleGroupStats.evicted+(singleGroupStats.inPlaceUpdating-1)) > shouldBeAlive {
					klog.V(4).Infof("Would be able to evict, but already resizing %s", pod.Name)

					if pod.Status.Resize == apiv1.PodResizeStatusInfeasible || pod.Status.Resize == apiv1.PodResizeStatusDeferred {
						klog.Warningf("Attempted in-place resize of %s impossible, should now evict", pod.Name)
						return true
					}
				}
				return false
			}

			if singleGroupStats.running-(singleGroupStats.evicted+singleGroupStats.inPlaceUpdating) > shouldBeAlive {
				return true
			}
			// If all pods are running and eviction tolerance is small evict 1 pod.
			if singleGroupStats.running == singleGroupStats.configured &&
				singleGroupStats.evictionTolerance == 0 &&
				singleGroupStats.evicted == 0 &&
				singleGroupStats.inPlaceUpdating == 0 {
				return true
			}

		}
	}
	return false
}

// Evict sends eviction instruction to api client. Returns error if pod cannot be evicted or if client returned error
// Does not check if pod was actually evicted after eviction grace period.
func (e *podsEvictionRestrictionImpl) Evict(podToEvict *apiv1.Pod, vpa *vpa_types.VerticalPodAutoscaler, eventRecorder record.EventRecorder) error {
	cr, present := e.podToReplicaCreatorMap[GetPodID(podToEvict)]
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
	dsInformer, err := setUpInformer(client, daemonSet)
	if err != nil {
		return nil, fmt.Errorf("Failed to create dsInformer: %v", err)
	}
	return &podsEvictionRestrictionFactoryImpl{
		client:                    client,
		rcInformer:                rcInformer, // informer for Replication Controllers
		ssInformer:                ssInformer, // informer for Replica Sets
		rsInformer:                rsInformer, // informer for Stateful Sets
		dsInformer:                dsInformer, // informer for Daemon Sets
		minReplicas:               minReplicas,
		evictionToleranceFraction: evictionToleranceFraction}, nil
}

// NewPodsEvictionRestriction creates PodsEvictionRestriction for a given set of pods,
// controlled by a single VPA object.
func (f *podsEvictionRestrictionFactoryImpl) NewPodsEvictionRestriction(pods []*apiv1.Pod, vpa *vpa_types.VerticalPodAutoscaler, patchCalculators []patch.Calculator) PodsEvictionRestriction {
	// We can evict pod only if it is a part of replica set
	// For each replica set we can evict only a fraction of pods.
	// Evictions may be later limited by pod disruption budget if configured.

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
		// TODO(maxcao13): Does minReplicas matter if the update is in-place? There should be no downtime if the update is disruptionless
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
		singleGroup.evictionTolerance = int(float64(configured) * f.evictionToleranceFraction)
		for _, pod := range replicas {
			podToReplicaCreatorMap[GetPodID(pod)] = creator
			if pod.Status.Phase == apiv1.PodPending {
				singleGroup.pending = singleGroup.pending + 1
			}
			if IsInPlaceUpdating(pod) {
				singleGroup.inPlaceUpdating = singleGroup.inPlaceUpdating + 1
			}
		}
		singleGroup.running = len(replicas) - singleGroup.pending
		creatorToSingleGroupStatsMap[creator] = singleGroup

	}
	return &podsEvictionRestrictionImpl{
		client:                       f.client,
		podToReplicaCreatorMap:       podToReplicaCreatorMap,
		creatorToSingleGroupStatsMap: creatorToSingleGroupStatsMap,
		patchCalculators:             patchCalculators,
	}
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

// GetPodID returns a string that uniquely identifies a pod by namespace and name
func GetPodID(pod *apiv1.Pod) string {
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

// CanInPlaceUpdate performs the same checks
func (e *podsEvictionRestrictionImpl) CanInPlaceUpdate(pod *apiv1.Pod) bool {

	cr, present := e.podToReplicaCreatorMap[GetPodID(pod)]
	if present {

		// If our QoS class is guaranteed, we can't change the resources without a restart
		// TODO(maxcao13): kubelet already prevents a resize of a guaranteed pod, so should we still check this early?
		if pod.Status.QOSClass == apiv1.PodQOSGuaranteed {
			klog.Warningf("Can't resize %s in-place, pod QoS is %s", pod.Name, pod.Status.QOSClass)
			return false
		}

		// If we're already resizing this pod, don't do it again
		if IsInPlaceUpdating(pod) {
			klog.Warningf("Not resizing %s, already resizing it", pod.Name)
			return false
		}

		noRestartPoliciesPopulated := true

		for _, container := range pod.Spec.Containers {
			// If some of these are populated, we know it at least understands resizing
			if len(container.ResizePolicy) > 0 {
				noRestartPoliciesPopulated = false
			}

			// TODO(maxcao13): Do we have to check the policy resource too? i.e. if only memory is getting scaled, then only check the memory resize policy?
			for _, policy := range container.ResizePolicy {
				if policy.RestartPolicy != apiv1.NotRequired {
					klog.Warningf("in-place resize of %s will cause container disruption, container %s restart policy is %v", pod.Name, container.Name, policy.RestartPolicy)
					// TODO(jkyros): is there something that prevents this from happening elsewhere in the API?
					if pod.Spec.RestartPolicy == apiv1.RestartPolicyNever {
						klog.Warningf("in-place resize of %s not possible, container %s resize policy is %v but pod restartPolicy is %v", pod.Name, container.Name, policy.RestartPolicy, pod.Spec.RestartPolicy)
						return false
					}

				}
			}
		}

		// If none of the policies are populated, our feature is probably not enabled, so we can't in-place regardless
		if noRestartPoliciesPopulated {
			klog.Warningf("impossible to resize %s in-place, container resize policies are not populated", pod.Name)
		}

		//TODO(jkyros): Come back and handle sidecar containers at some point since they're weird?
		singleGroupStats, present := e.creatorToSingleGroupStatsMap[cr]
		// If we're pending, we can't in-place resize
		// TODO(jkyros): are we sure we can't? Should I just set this to "if running"?
		if pod.Status.Phase == apiv1.PodPending {
			klog.V(4).Infof("Can't resize pending pod %s", pod.Name)
			return false
		}
		// This second "present" check is against the creator-to-group-stats map, not the pod-to-replica map
		// TODO(maxcao13): Not sure, but do we need disruption tolerance for in-place updates? I guess we do since they are not guaranteed to be disruptionless...
		// TODO(maxcao13): If this is okay, I may have to rename evictionTolerance to disruption tolerance
		if present {
			klog.V(4).InfoS("Checking pod disruption tolerance",
				"podName", pod.Name,
				"configuredPods", singleGroupStats.configured,
				"runningPods", singleGroupStats.running,
				"evictedPods", singleGroupStats.evicted,
				"inPlaceUpdatingPods", singleGroupStats.inPlaceUpdating,
				"evictionTolerance", singleGroupStats.evictionTolerance,
			)
			// minimum number of pods that should be running to tolerate disruptions
			shouldBeAlive := singleGroupStats.configured - singleGroupStats.evictionTolerance
			// number of pods that are actually running
			actuallyAlive := singleGroupStats.running - (singleGroupStats.evicted + singleGroupStats.inPlaceUpdating)
			if actuallyAlive > shouldBeAlive {
				klog.V(4).InfoS("Pod can be resized in-place; more pods are running than required", "podName", pod.Name, "shouldBeAlive", shouldBeAlive, "actuallyAlive", actuallyAlive)
				return true
			}

			// If all pods are running, no pods are being evicted or updated, and eviction tolerance is small, we can resize in-place
			if singleGroupStats.running == singleGroupStats.configured &&
				singleGroupStats.evictionTolerance == 0 &&
				singleGroupStats.evicted == 0 && singleGroupStats.inPlaceUpdating == 0 {
				klog.V(4).InfoS("Pod can be resized in-place; all pods are running and eviction tolerance is 0", "podName", pod.Name)
				return true
			}
		}
	}
	return false
}

// TODO(maxcao13): Maybe we want to move all this updating logic outisde of this method receiver or into a different package/file?
// InPlaceUpdate sends calculates patches and sends resize request to api client. Returns error if pod cannot be in-place updated or if client returned error.
// Does not check if pod was actually in-place updated after grace period.
func (e *podsEvictionRestrictionImpl) InPlaceUpdate(podToUpdate *apiv1.Pod, vpa *vpa_types.VerticalPodAutoscaler, eventRecorder record.EventRecorder) error {
	cr, present := e.podToReplicaCreatorMap[GetPodID(podToUpdate)]
	if !present {
		return fmt.Errorf("pod not suitable for eviction %v : not in replicated pods map", podToUpdate.Name)
	}

	// TODO(maxcao13): Not sure if we need to check again here, but commenting it out for now in case we do
	// if !e.CanInPlaceUpdate(podToUpdate) {
	// 	return fmt.Errorf("cannot update pod %v in place : number of in-flight updates exceeded", podToUpdate.Name)
	// }

	// TODO(maxcao13): There's maybe a more efficient way to do this, but this is what we have for now

	// separate patches since we have to patch resize and spec separately
	resourcePatches := []resource_updates.PatchRecord{}
	annotationPatches := []resource_updates.PatchRecord{}
	if podToUpdate.Annotations == nil {
		annotationPatches = append(annotationPatches, patch.GetAddEmptyAnnotationsPatch())
	}
	for i, calculator := range e.patchCalculators {
		p, err := calculator.CalculatePatches(podToUpdate, vpa)
		if err != nil {
			return fmt.Errorf("failed to calculate resource patch for pod %s/%s: %v", podToUpdate.Namespace, podToUpdate.Name, err)
		}
		klog.V(4).InfoS("Calculated patches for pod", "pod", klog.KObj(podToUpdate), "patches", p)
		// TODO(maxcao13): change how this works later, this is gross and depends on the resource calculator being first in the slice
		// we may not even want the updater to patch pod annotations at all
		if i == 0 {
			resourcePatches = append(resourcePatches, p...)
		} else {
			annotationPatches = append(annotationPatches, p...)
		}
	}
	if len(resourcePatches) > 0 {
		patch, err := json.Marshal(resourcePatches)
		if err != nil {
			klog.Errorf("Cannot marshal the patch %v: %v", resourcePatches, err)
			return err
		}

		res, err := e.client.CoreV1().Pods(podToUpdate.Namespace).Patch(context.TODO(), podToUpdate.Name, k8stypes.JSONPatchType, patch, metav1.PatchOptions{}, "resize")
		if err != nil {
			klog.ErrorS(err, "Failed to patch pod", "pod", klog.KObj(podToUpdate))
			return err
		}
		klog.V(4).InfoS("In-place patched pod /resize subresource using patches ", "pod", klog.KObj(res), "patches", string(patch))

		// TODO(maxcao13): whether or not we apply annotation patches should depend on resource patches?
		if len(annotationPatches) > 0 {
			patch, err := json.Marshal(annotationPatches)
			if err != nil {
				klog.Errorf("Cannot marshal the patch %v: %v", annotationPatches, err)
				return err
			}
			res, err = e.client.CoreV1().Pods(podToUpdate.Namespace).Patch(context.TODO(), podToUpdate.Name, k8stypes.JSONPatchType, patch, metav1.PatchOptions{})
			if err != nil {
				klog.ErrorS(err, "Failed to patch pod", "pod", klog.KObj(podToUpdate))
				return err
			}
			klog.V(4).InfoS("Patched pod annotations", "pod", klog.KObj(res), "patches", string(patch))
		}
	} else {
		err := fmt.Errorf("no patches to apply to %s", podToUpdate.Name)
		klog.ErrorS(err, "Failed to patch pod", "pod", klog.KObj(podToUpdate))
		return err
	}

	// TODO(maxcao13): If this keeps getting called on the same object with the same reason, it is considered a patch request.
	// And we fail to have the corresponding rbac for it. So figure out if we need this later.
	// TODO(maxcao13): Do we even need to emit an event? The api server might reject the resize request. Should we rename this to InPlaceResizeAttempted?
	// eventRecorder.Event(podToUpdate, apiv1.EventTypeNormal, "InPlaceResizedByVPA", "Pod was resized in place by VPA Updater.")

	// TODO(jkyros): You need to do this regardless once you update the pod, if it changes phases here as a result, you still
	// need to catalog what you did
	if podToUpdate.Status.Phase == apiv1.PodRunning {
		singleGroupStats, present := e.creatorToSingleGroupStatsMap[cr]
		if !present {
			klog.Errorf("Internal error - cannot find stats for replication group %v", cr)
		} else {
			singleGroupStats.inPlaceUpdating = singleGroupStats.inPlaceUpdating + 1
			e.creatorToSingleGroupStatsMap[cr] = singleGroupStats
		}
	} else {
		klog.Warningf("I updated, but my pod phase was %s", podToUpdate.Status.Phase)
	}

	return nil
}

// IsInPlaceUpdating checks whether or not the given pod is currently in the middle of an in-place update
func IsInPlaceUpdating(podToCheck *apiv1.Pod) (isUpdating bool) {
	// If the pod is currently updating we need to tally that
	if podToCheck.Status.Resize != "" {
		klog.V(4).InfoS("Pod is currently resizing", "pod", klog.KObj(podToCheck), "status", podToCheck.Status.Resize)
		// Proposed -> Deferred -> InProgress, but what about Infeasible?
		if podToCheck.Status.Resize == apiv1.PodResizeStatusInfeasible {
			klog.V(4).InfoS("Resource proposal for pod is Infeasible, we're probably stuck like this until we evict", "pod", klog.KObj(podToCheck))
		} else if podToCheck.Status.Resize == apiv1.PodResizeStatusDeferred {
			klog.V(4).InfoS("Resource proposal for pod is Deferred, we're probably stuck like this until we evict", "pod", klog.KObj(podToCheck))
		}
		return true
	}
	return false
}
