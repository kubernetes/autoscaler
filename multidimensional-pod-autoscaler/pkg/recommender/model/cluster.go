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

package model

import (
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	mpa_types "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1alpha1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/input/controller_fetcher"
	vpa_model "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	vpa_utils "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
	"k8s.io/klog/v2"
)

const (
	// RecommendationMissingMaxDuration is maximum time that we accept the recommendation can be missing.
	RecommendationMissingMaxDuration = 30 * time.Minute
)

// ClusterState holds all runtime information about the cluster required for the
// VPA operations, i.e. configuration of resources (pods, containers,
// VPA objects), aggregated utilization of compute resources (CPU, memory) and
// events (container OOMs).
// All input to the VPA Recommender algorithm lives in this structure.
type ClusterState struct {
	// Pods in the cluster.
	Pods map[vpa_model.PodID]*PodState

	// MPA objects in the cluster.
	Mpas map[MpaID]*Mpa
	// MPA objects in the cluster that have no recommendation mapped to the first
	// time we've noticed the recommendation missing or last time we logged
	// a warning about it.
	EmptyMPAs map[MpaID]time.Time
	// Observed MPAs. Used to check if there are updates needed.
	ObservedMpas []*mpa_types.MultidimPodAutoscaler

	// All container aggregations where the usage samples are stored.
	aggregateStateMap aggregateContainerStatesMap
	// Map with all label sets used by the aggregations. It serves as a cache
	// that allows to quickly access labels.Set corresponding to a labelSetKey.
	labelSetMap labelSetMap

	lastAggregateContainerStateGC time.Time
	gcInterval                    time.Duration
}

// StateMapSize is the number of pods being tracked by the VPA
func (cluster *ClusterState) StateMapSize() int {
	return len(cluster.aggregateStateMap)
}

// String representation of the labels.LabelSet. This is the value returned by
// labelSet.String(). As opposed to the LabelSet object, it can be used as a map key.
type labelSetKey string

// Map of label sets keyed by their string representation.
type labelSetMap map[labelSetKey]labels.Set

// AggregateContainerStatesMap is a map from AggregateStateKey to AggregateContainerState.
type aggregateContainerStatesMap map[vpa_model.AggregateStateKey]*vpa_model.AggregateContainerState

// PodState holds runtime information about a single Pod.
type PodState struct {
	// Unique id of the Pod.
	ID vpa_model.PodID
	// Set of labels attached to the Pod.
	labelSetKey labelSetKey
	// Containers that belong to the Pod, keyed by the container name.
	Containers map[string]*ContainerState
	// PodPhase describing current life cycle phase of the Pod.
	Phase apiv1.PodPhase
}

// NewClusterState returns a new ClusterState with no pods.
func NewClusterState(gcInterval time.Duration) *ClusterState {
	return &ClusterState{
		Pods:                          make(map[vpa_model.PodID]*PodState),
		Mpas:                          make(map[MpaID]*Mpa),
		EmptyMPAs:                     make(map[MpaID]time.Time),
		aggregateStateMap:             make(aggregateContainerStatesMap),
		labelSetMap:                   make(labelSetMap),
		lastAggregateContainerStateGC: time.Unix(0, 0),
		gcInterval:                    gcInterval,
	}
}

// ContainerUsageSampleWithKey holds a ContainerUsageSample together with the
// ID of the container it belongs to.
type ContainerUsageSampleWithKey struct {
	vpa_model.ContainerUsageSample
	Container vpa_model.ContainerID
}

// AddOrUpdatePod updates the state of the pod with a given PodID, if it is
// present in the cluster object. Otherwise a new pod is created and added to
// the Cluster object.
// If the labels of the pod have changed, it updates the links between the containers
// and the aggregations.
func (cluster *ClusterState) AddOrUpdatePod(podID vpa_model.PodID, newLabels labels.Set, phase apiv1.PodPhase) {
	pod, podExists := cluster.Pods[podID]
	if !podExists {
		pod = newPod(podID)
		cluster.Pods[podID] = pod
	}

	newlabelSetKey := cluster.getLabelSetKey(newLabels)
	if podExists && pod.labelSetKey != newlabelSetKey {
		// This Pod is already counted in the old MPA, remove the link.
		cluster.removePodFromItsMpa(pod)
	}
	if !podExists || pod.labelSetKey != newlabelSetKey {
		pod.labelSetKey = newlabelSetKey
		// Set the links between the containers and aggregations based on the current pod labels.
		for containerName, container := range pod.Containers {
			containerID := vpa_model.ContainerID{PodID: podID, ContainerName: containerName}
			container.aggregator = cluster.findOrCreateAggregateContainerState(containerID)
		}

		cluster.addPodToItsMpa(pod)
	}
	pod.Phase = phase
}

// addPodToItsVpa increases the count of Pods associated with a MPA object.
// Does a scan similar to findOrCreateAggregateContainerState so could be optimized if needed.
func (cluster *ClusterState) addPodToItsMpa(pod *PodState) {
	for _, mpa := range cluster.Mpas {
		if vpa_utils.PodLabelsMatchVPA(pod.ID.Namespace, cluster.labelSetMap[pod.labelSetKey], mpa.ID.Namespace, mpa.PodSelector) {
			mpa.PodCount++
		}
	}
}

// removePodFromItsVpa decreases the count of Pods associated with a VPA object.
func (cluster *ClusterState) removePodFromItsMpa(pod *PodState) {
	for _, mpa := range cluster.Mpas {
		if vpa_utils.PodLabelsMatchVPA(pod.ID.Namespace, cluster.labelSetMap[pod.labelSetKey], mpa.ID.Namespace, mpa.PodSelector) {
			mpa.PodCount--
		}
	}
}

// GetContainer returns the ContainerState object for a given ContainerID or
// null if it's not present in the model.
func (cluster *ClusterState) GetContainer(containerID vpa_model.ContainerID) *ContainerState {
	pod, podExists := cluster.Pods[containerID.PodID]
	if podExists {
		container, containerExists := pod.Containers[containerID.ContainerName]
		if containerExists {
			return container
		}
	}
	return nil
}

// DeletePod removes an existing pod from the cluster.
func (cluster *ClusterState) DeletePod(podID vpa_model.PodID) {
	pod, found := cluster.Pods[podID]
	if found {
		cluster.removePodFromItsMpa(pod)
	}
	delete(cluster.Pods, podID)
}

// AddOrUpdateContainer creates a new container with the given ContainerID and
// adds it to the parent pod in the ClusterState object, if not yet present.
// Requires the pod to be added to the ClusterState first. Otherwise an error is
// returned.
func (cluster *ClusterState) AddOrUpdateContainer(containerID vpa_model.ContainerID, request vpa_model.Resources) error {
	pod, podExists := cluster.Pods[containerID.PodID]
	if !podExists {
		return vpa_model.NewKeyError(containerID.PodID)
	}
	if container, containerExists := pod.Containers[containerID.ContainerName]; !containerExists {
		cluster.findOrCreateAggregateContainerState(containerID)
		pod.Containers[containerID.ContainerName] = NewContainerState(request, NewContainerStateAggregatorProxy(cluster, containerID))
	} else {
		// Container aleady exists. Possibly update the request.
		container.Request = request
	}
	return nil
}

// AddSample adds a new usage sample to the proper container in the ClusterState
// object. Requires the container as well as the parent pod to be added to the
// ClusterState first. Otherwise an error is returned.
func (cluster *ClusterState) AddSample(sample *ContainerUsageSampleWithKey) error {
	pod, podExists := cluster.Pods[sample.Container.PodID]
	if !podExists {
		return vpa_model.NewKeyError(sample.Container.PodID)
	}
	containerState, containerExists := pod.Containers[sample.Container.ContainerName]
	if !containerExists {
		return vpa_model.NewKeyError(sample.Container)
	}
	if !containerState.AddSample(&sample.ContainerUsageSample) {
		return fmt.Errorf("sample discarded (invalid or out of order)")
	}
	return nil
}

// RecordOOM adds info regarding OOM event in the model as an artificial memory sample.
func (cluster *ClusterState) RecordOOM(containerID vpa_model.ContainerID, timestamp time.Time, requestedMemory vpa_model.ResourceAmount) error {
	pod, podExists := cluster.Pods[containerID.PodID]
	if !podExists {
		return vpa_model.NewKeyError(containerID.PodID)
	}
	containerState, containerExists := pod.Containers[containerID.ContainerName]
	if !containerExists {
		return vpa_model.NewKeyError(containerID.ContainerName)
	}
	err := containerState.RecordOOM(timestamp, requestedMemory)
	if err != nil {
		return fmt.Errorf("error while recording OOM for %v, Reason: %v", containerID, err)
	}
	return nil
}

// AddOrUpdateMpa adds a new MPA with a given ID to the ClusterState if it
// didn't yet exist. If the MPA already existed but had a different pod
// selector, the pod selector is updated. Updates the links between the MPA and
// all aggregations it matches.
func (cluster *ClusterState) AddOrUpdateMpa(apiObject *mpa_types.MultidimPodAutoscaler, selector labels.Selector) error {
	mpaID := MpaID{Namespace: apiObject.Namespace, MpaName: apiObject.Name}
	annotationsMap := apiObject.Annotations
	conditionsMap := make(mpaConditionsMap)
	for _, condition := range apiObject.Status.Conditions {
		conditionsMap[condition.Type] = condition
	}
	var currentRecommendation *vpa_types.RecommendedPodResources
	if conditionsMap[mpa_types.RecommendationProvided].Status == apiv1.ConditionTrue {
		currentRecommendation = apiObject.Status.Recommendation
	}

	mpa, mpaExists := cluster.Mpas[mpaID]
	if mpaExists && (mpa.PodSelector.String() != selector.String()) {
		// Pod selector was changed. Delete the MPA object and recreate
		// it with the new selector.
		if err := cluster.DeleteMpa(mpaID); err != nil {
			return err
		}
		mpaExists = false
	}
	if !mpaExists {
		mpa = NewMpa(mpaID, selector, apiObject.CreationTimestamp.Time)
		cluster.Mpas[mpaID] = mpa
		for aggregationKey, aggregation := range cluster.aggregateStateMap {
			mpa.UseAggregationIfMatching(aggregationKey, aggregation)
		}
		mpa.PodCount = len(cluster.GetMatchingPods(mpa))
	}
	mpa.ScaleTargetRef = apiObject.Spec.ScaleTargetRef
	mpa.Annotations = annotationsMap
	mpa.Conditions = conditionsMap
	mpa.Recommendation = currentRecommendation
	mpa.SetUpdateMode(apiObject.Spec.UpdatePolicy)
	mpa.SetResourcePolicy(apiObject.Spec.ResourcePolicy)

	// For HPA-related fields.
	// mpa.SetHPAConstraints(apiObject.Spec.Metrics, *apiObject.Spec.Constraints.MinReplicas, *apiObject.Spec.Constraints.MaxReplicas, apiObject.Spec.Constraints.Behavior)
	return nil
}

// DeleteMpa removes a MPA with the given ID from the ClusterState.
func (cluster *ClusterState) DeleteMpa(mpaID MpaID) error {
	mpa, mpaExists := cluster.Mpas[mpaID]
	if !mpaExists {
		return vpa_model.NewKeyError(mpaID)
	}
	for _, state := range mpa.aggregateContainerStates {
		state.MarkNotAutoscaled()
	}
	delete(cluster.Mpas, mpaID)
	delete(cluster.EmptyMPAs, mpaID)
	return nil
}

func newPod(id vpa_model.PodID) *PodState {
	return &PodState{
		ID:         id,
		Containers: make(map[string]*ContainerState),
	}
}

// getLabelSetKey puts the given labelSet in the global labelSet map and returns a
// corresponding labelSetKey.
func (cluster *ClusterState) getLabelSetKey(labelSet labels.Set) labelSetKey {
	labelSetKey := labelSetKey(labelSet.String())
	cluster.labelSetMap[labelSetKey] = labelSet
	return labelSetKey
}

// MakeAggregateStateKey returns the AggregateStateKey that should be used
// to aggregate usage samples from a container with the given name in a given pod.
func (cluster *ClusterState) MakeAggregateStateKey(pod *PodState, containerName string) vpa_model.AggregateStateKey {
	return aggregateStateKey{
		namespace:     pod.ID.Namespace,
		containerName: containerName,
		labelSetKey:   pod.labelSetKey,
		labelSetMap:   &cluster.labelSetMap,
	}
}

// aggregateStateKeyForContainerID returns the AggregateStateKey for the ContainerID.
// The pod with the corresponding PodID must already be present in the ClusterState.
func (cluster *ClusterState) aggregateStateKeyForContainerID(containerID vpa_model.ContainerID) vpa_model.AggregateStateKey {
	pod, podExists := cluster.Pods[containerID.PodID]
	if !podExists {
		panic(fmt.Sprintf("Pod not present in the ClusterState: %v", containerID.PodID))
	}
	return cluster.MakeAggregateStateKey(pod, containerID.ContainerName)
}

// findOrCreateAggregateContainerState returns (possibly newly created) AggregateContainerState
// that should be used to aggregate usage samples from container with a given ID.
// The pod with the corresponding PodID must already be present in the ClusterState.
func (cluster *ClusterState) findOrCreateAggregateContainerState(containerID vpa_model.ContainerID) *vpa_model.AggregateContainerState {
	aggregateStateKey := cluster.aggregateStateKeyForContainerID(containerID)
	aggregateContainerState, aggregateStateExists := cluster.aggregateStateMap[aggregateStateKey]
	if !aggregateStateExists {
		aggregateContainerState = vpa_model.NewAggregateContainerState()
		cluster.aggregateStateMap[aggregateStateKey] = aggregateContainerState
		// Link the new aggregation to the existing VPAs.
		for _, mpa := range cluster.Mpas {
			mpa.UseAggregationIfMatching(aggregateStateKey, aggregateContainerState)
		}
	}
	return aggregateContainerState
}

// garbageCollectAggregateCollectionStates removes obsolete AggregateCollectionStates from the ClusterState.
// AggregateCollectionState is obsolete in following situations:
// 1) It has no samples and there are no more contributive pods - a pod is contributive in any of following situations:
//
//	a) It is in an active state - i.e. not PodSucceeded nor PodFailed.
//	b) Its associated controller (e.g. Deployment) still exists.
//
// 2) The last sample is too old to give meaningful recommendation (>8 days),
// 3) There are no samples and the aggregate state was created >8 days ago.
func (cluster *ClusterState) garbageCollectAggregateCollectionStates(now time.Time, controllerFetcher controllerfetcher.ControllerFetcher) {
	klog.V(1).Info("Garbage collection of AggregateCollectionStates triggered")
	keysToDelete := make([]vpa_model.AggregateStateKey, 0)
	contributiveKeys := cluster.getContributiveAggregateStateKeys(controllerFetcher)
	for key, aggregateContainerState := range cluster.aggregateStateMap {
		isKeyContributive := contributiveKeys[key]
		if !isKeyContributive && isStateEmpty(aggregateContainerState) {
			keysToDelete = append(keysToDelete, key)
			klog.V(1).Infof("Removing empty and not contributive AggregateCollectionState for %+v", key)
			continue
		}
		if isStateExpired(aggregateContainerState, now) {
			keysToDelete = append(keysToDelete, key)
			klog.V(1).Infof("Removing expired AggregateCollectionState for %+v", key)
		}
	}
	for _, key := range keysToDelete {
		delete(cluster.aggregateStateMap, key)
		for _, mpa := range cluster.Mpas {
			mpa.DeleteAggregation(key)
		}
	}
}

// RateLimitedGarbageCollectAggregateCollectionStates removes obsolete AggregateCollectionStates from the ClusterState.
// It performs clean up only if more than `gcInterval` passed since the last time it performed a clean up.
// AggregateCollectionState is obsolete in following situations:
// 1) It has no samples and there are no more contributive pods - a pod is contributive in any of following situations:
//
//	a) It is in an active state - i.e. not PodSucceeded nor PodFailed.
//	b) Its associated controller (e.g. Deployment) still exists.
//
// 2) The last sample is too old to give meaningful recommendation (>8 days),
// 3) There are no samples and the aggregate state was created >8 days ago.
func (cluster *ClusterState) RateLimitedGarbageCollectAggregateCollectionStates(now time.Time, controllerFetcher controllerfetcher.ControllerFetcher) {
	if now.Sub(cluster.lastAggregateContainerStateGC) < cluster.gcInterval {
		return
	}
	cluster.garbageCollectAggregateCollectionStates(now, controllerFetcher)
	cluster.lastAggregateContainerStateGC = now
}

func (cluster *ClusterState) getContributiveAggregateStateKeys(controllerFetcher controllerfetcher.ControllerFetcher) map[vpa_model.AggregateStateKey]bool {
	contributiveKeys := map[vpa_model.AggregateStateKey]bool{}
	for _, pod := range cluster.Pods {
		// Pod is considered contributive in any of following situations:
		// 1) It is in active state - i.e. not PodSucceeded nor PodFailed.
		// 2) Its associated controller (e.g. Deployment) still exists.
		podControllerExists := cluster.GetControllerForPodUnderVPA(pod, controllerFetcher) != nil
		podActive := pod.Phase != apiv1.PodSucceeded && pod.Phase != apiv1.PodFailed
		if podActive || podControllerExists {
			for container := range pod.Containers {
				contributiveKeys[cluster.MakeAggregateStateKey(pod, container)] = true
			}
		}
	}
	return contributiveKeys
}

// RecordRecommendation marks the state of recommendation in the cluster. We
// keep track of empty recommendations and log information about them
// periodically.
func (cluster *ClusterState) RecordRecommendation(mpa *Mpa, now time.Time) error {
	if mpa.Recommendation != nil && len(mpa.Recommendation.ContainerRecommendations) > 0 {
		delete(cluster.EmptyMPAs, mpa.ID)
		return nil
	}
	lastLogged, ok := cluster.EmptyMPAs[mpa.ID]
	if !ok {
		cluster.EmptyMPAs[mpa.ID] = now
	} else {
		if lastLogged.Add(RecommendationMissingMaxDuration).Before(now) {
			cluster.EmptyMPAs[mpa.ID] = now
			return fmt.Errorf("MPA %v/%v is missing recommendation for more than %v", mpa.ID.Namespace, mpa.ID.MpaName, RecommendationMissingMaxDuration)
		}
	}
	return nil
}

// GetMatchingPods returns a list of currently active pods that match the
// given MPA. Traverses through all pods in the cluster - use sparingly.
func (cluster *ClusterState) GetMatchingPods(mpa *Mpa) []vpa_model.PodID {
	matchingPods := []vpa_model.PodID{}
	for podID, pod := range cluster.Pods {
		if vpa_utils.PodLabelsMatchVPA(podID.Namespace, cluster.labelSetMap[pod.labelSetKey],
			mpa.ID.Namespace, mpa.PodSelector) {
			matchingPods = append(matchingPods, podID)
		}
	}
	return matchingPods
}

// GetControllerForPodUnderVPA returns controller associated with given Pod. Returns nil if Pod is not controlled by a VPA object.
func (cluster *ClusterState) GetControllerForPodUnderVPA(pod *PodState, controllerFetcher controllerfetcher.ControllerFetcher) *controllerfetcher.ControllerKeyWithAPIVersion {
	controllingMPA := cluster.GetControllingMPA(pod)
	if controllingMPA != nil {
		controller := &controllerfetcher.ControllerKeyWithAPIVersion{
			ControllerKey: controllerfetcher.ControllerKey{
				Namespace: controllingMPA.ID.Namespace,
				Kind:      controllingMPA.ScaleTargetRef.Kind,
				Name:      controllingMPA.ScaleTargetRef.Name,
			},
			ApiVersion: controllingMPA.ScaleTargetRef.APIVersion,
		}
		topLevelController, _ := controllerFetcher.FindTopMostWellKnownOrScalable(controller)
		return topLevelController
	}
	return nil
}

// GetControllingVPA returns a VPA object controlling given Pod.
func (cluster *ClusterState) GetControllingMPA(pod *PodState) *Mpa {
	for _, mpa := range cluster.Mpas {
		if vpa_utils.PodLabelsMatchVPA(pod.ID.Namespace, cluster.labelSetMap[pod.labelSetKey],
			mpa.ID.Namespace, mpa.PodSelector) {
			return mpa
		}
	}
	return nil
}

// Implementation of the AggregateStateKey interface. It can be used as a map key.
type aggregateStateKey struct {
	namespace     string
	containerName string
	labelSetKey   labelSetKey
	// Pointer to the global map from labelSetKey to labels.Set.
	// Note: a pointer is used so that two copies of the same key are equal.
	labelSetMap *labelSetMap
}

// Labels returns the namespace for the aggregateStateKey.
func (k aggregateStateKey) Namespace() string {
	return k.namespace
}

// ContainerName returns the name of the container for the aggregateStateKey.
func (k aggregateStateKey) ContainerName() string {
	return k.containerName
}

// Labels returns the set of labels for the aggregateStateKey.
func (k aggregateStateKey) Labels() labels.Labels {
	if k.labelSetMap == nil {
		return labels.Set{}
	}
	return (*k.labelSetMap)[k.labelSetKey]
}
