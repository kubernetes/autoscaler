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
	"context"
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	controllerfetcher "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/target/controller_fetcher"
	vpa_utils "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
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
	Pods map[PodID]*PodState
	// VPA objects in the cluster.
	Vpas map[VpaID]*Vpa
	// VPA objects in the cluster that have no recommendation mapped to the first
	// time we've noticed the recommendation missing or last time we logged
	// a warning about it.
	EmptyVPAs map[VpaID]time.Time
	// Observed VPAs. Used to check if there are updates needed.
	ObservedVpas []*vpa_types.VerticalPodAutoscaler

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

// AggregateStateKey determines the set of containers for which the usage samples
// are kept aggregated in the model.
type AggregateStateKey interface {
	Namespace() string
	ContainerName() string
	Labels() labels.Labels
}

// String representation of the labels.LabelSet. This is the value returned by
// labelSet.String(). As opposed to the LabelSet object, it can be used as a map key.
type labelSetKey string

// Map of label sets keyed by their string representation.
type labelSetMap map[labelSetKey]labels.Set

// AggregateContainerStatesMap is a map from AggregateStateKey to AggregateContainerState.
type aggregateContainerStatesMap map[AggregateStateKey]*AggregateContainerState

// PodState holds runtime information about a single Pod.
type PodState struct {
	// Unique id of the Pod.
	ID PodID
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
		Pods:                          make(map[PodID]*PodState),
		Vpas:                          make(map[VpaID]*Vpa),
		EmptyVPAs:                     make(map[VpaID]time.Time),
		aggregateStateMap:             make(aggregateContainerStatesMap),
		labelSetMap:                   make(labelSetMap),
		lastAggregateContainerStateGC: time.Unix(0, 0),
		gcInterval:                    gcInterval,
	}
}

// ContainerUsageSampleWithKey holds a ContainerUsageSample together with the
// ID of the container it belongs to.
type ContainerUsageSampleWithKey struct {
	ContainerUsageSample
	Container ContainerID
}

// AddOrUpdatePod updates the state of the pod with a given PodID, if it is
// present in the cluster object. Otherwise a new pod is created and added to
// the Cluster object.
// If the labels of the pod have changed, it updates the links between the containers
// and the aggregations.
func (cluster *ClusterState) AddOrUpdatePod(podID PodID, newLabels labels.Set, phase apiv1.PodPhase) {
	pod, podExists := cluster.Pods[podID]
	if !podExists {
		pod = newPod(podID)
		cluster.Pods[podID] = pod
	}

	newlabelSetKey := cluster.getLabelSetKey(newLabels)
	if podExists && pod.labelSetKey != newlabelSetKey {
		// This Pod is already counted in the old VPA, remove the link.
		cluster.removePodFromItsVpa(pod)
	}
	if !podExists || pod.labelSetKey != newlabelSetKey {
		pod.labelSetKey = newlabelSetKey
		// Set the links between the containers and aggregations based on the current pod labels.
		for containerName, container := range pod.Containers {
			containerID := ContainerID{PodID: podID, ContainerName: containerName}
			container.aggregator = cluster.findOrCreateAggregateContainerState(containerID)
		}

		cluster.addPodToItsVpa(pod)
	}
	// Tally the number of containers for later when we're averaging the recommendations
	cluster.setVPAContainersPerPod(pod)
	pod.Phase = phase
}

// setVPAContainersPerPod sets the number of containers per pod seen for pods connected to this VPA
// so that later when were spreading the minimum recommendations we're spreading them over the correct
// number and not just the number of all aggregates that have *ever* been in the pod. (We don't want minimum resources
// to erroneously shrink)
func (cluster *ClusterState) setVPAContainersPerPod(pod *PodState) {
	for _, vpa := range cluster.Vpas {
		if vpa_utils.PodLabelsMatchVPA(pod.ID.Namespace, cluster.labelSetMap[pod.labelSetKey], vpa.ID.Namespace, vpa.PodSelector) {
			// We want the "high water mark" of the most containers in the pod in the event
			// that we're rolling out a pod that has an additional container
			if len(pod.Containers) > vpa.ContainersPerPod {
				vpa.ContainersPerPod = len(pod.Containers)
			}
		}
	}

}

// addPodToItsVpa increases the count of Pods associated with a VPA object.
// Does a scan similar to findOrCreateAggregateContainerState so could be optimized if needed.
func (cluster *ClusterState) addPodToItsVpa(pod *PodState) {
	for _, vpa := range cluster.Vpas {
		if vpa_utils.PodLabelsMatchVPA(pod.ID.Namespace, cluster.labelSetMap[pod.labelSetKey], vpa.ID.Namespace, vpa.PodSelector) {
			vpa.PodCount++
		}
	}
}

// removePodFromItsVpa decreases the count of Pods associated with a VPA object.
func (cluster *ClusterState) removePodFromItsVpa(pod *PodState) {
	for _, vpa := range cluster.Vpas {
		if vpa_utils.PodLabelsMatchVPA(pod.ID.Namespace, cluster.labelSetMap[pod.labelSetKey], vpa.ID.Namespace, vpa.PodSelector) {
			vpa.PodCount--
		}
	}
}

// GetContainer returns the ContainerState object for a given ContainerID or
// null if it's not present in the model.
func (cluster *ClusterState) GetContainer(containerID ContainerID) *ContainerState {
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
func (cluster *ClusterState) DeletePod(podID PodID) {
	pod, found := cluster.Pods[podID]
	if found {
		cluster.removePodFromItsVpa(pod)
	}
	delete(cluster.Pods, podID)
}

// AddOrUpdateContainer creates a new container with the given ContainerID and
// adds it to the parent pod in the ClusterState object, if not yet present.
// Requires the pod to be added to the ClusterState first. Otherwise an error is
// returned.
func (cluster *ClusterState) AddOrUpdateContainer(containerID ContainerID, request Resources) error {
	pod, podExists := cluster.Pods[containerID.PodID]
	if !podExists {
		return NewKeyError(containerID.PodID)
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
		return NewKeyError(sample.Container.PodID)
	}
	containerState, containerExists := pod.Containers[sample.Container.ContainerName]
	if !containerExists {
		return NewKeyError(sample.Container)
	}
	if !containerState.AddSample(&sample.ContainerUsageSample) {
		return fmt.Errorf("sample discarded (invalid or out of order)")
	}
	return nil
}

// RecordOOM adds info regarding OOM event in the model as an artificial memory sample.
func (cluster *ClusterState) RecordOOM(containerID ContainerID, timestamp time.Time, requestedMemory ResourceAmount) error {
	pod, podExists := cluster.Pods[containerID.PodID]
	if !podExists {
		return NewKeyError(containerID.PodID)
	}
	containerState, containerExists := pod.Containers[containerID.ContainerName]
	if !containerExists {
		return NewKeyError(containerID.ContainerName)
	}
	err := containerState.RecordOOM(timestamp, requestedMemory)
	if err != nil {
		return fmt.Errorf("error while recording OOM for %v, Reason: %v", containerID, err)
	}
	return nil
}

// AddOrUpdateVpa adds a new VPA with a given ID to the ClusterState if it
// didn't yet exist. If the VPA already existed but had a different pod
// selector, the pod selector is updated. Updates the links between the VPA and
// all aggregations it matches.
func (cluster *ClusterState) AddOrUpdateVpa(apiObject *vpa_types.VerticalPodAutoscaler, selector labels.Selector) error {
	vpaID := VpaID{Namespace: apiObject.Namespace, VpaName: apiObject.Name}
	annotationsMap := apiObject.Annotations
	conditionsMap := make(vpaConditionsMap)
	for _, condition := range apiObject.Status.Conditions {
		conditionsMap[condition.Type] = condition
	}
	var currentRecommendation *vpa_types.RecommendedPodResources
	if conditionsMap[vpa_types.RecommendationProvided].Status == apiv1.ConditionTrue {
		currentRecommendation = apiObject.Status.Recommendation
	}

	vpa, vpaExists := cluster.Vpas[vpaID]
	if vpaExists && (vpa.PodSelector.String() != selector.String()) {
		// Pod selector was changed. Delete the VPA object and recreate
		// it with the new selector.
		if err := cluster.DeleteVpa(vpaID); err != nil {
			return err
		}
		vpaExists = false
	}
	if !vpaExists {
		vpa = NewVpa(vpaID, selector, apiObject.CreationTimestamp.Time)
		cluster.Vpas[vpaID] = vpa
		for aggregationKey, aggregation := range cluster.aggregateStateMap {
			vpa.UseAggregationIfMatching(aggregationKey, aggregation)
		}
		vpa.PodCount = len(cluster.GetMatchingPods(vpa))
	}

	// Default this to the minimum, we will tally the true number when we load the pods later
	// TODO(jkyros): This is gross, it depends on the order I know it currently loads things in, but
	// that might not be the case someday
	vpa.ContainersPerPod = 1
	vpa.TargetRef = apiObject.Spec.TargetRef
	vpa.Annotations = annotationsMap
	vpa.Conditions = conditionsMap
	vpa.Recommendation = currentRecommendation
	vpa.SetUpdateMode(apiObject.Spec.UpdatePolicy)
	vpa.SetResourcePolicy(apiObject.Spec.ResourcePolicy)
	vpa.SetAPIVersion(apiObject.GetObjectKind().GroupVersionKind().Version)
	return nil
}

// DeleteVpa removes a VPA with the given ID from the ClusterState.
func (cluster *ClusterState) DeleteVpa(vpaID VpaID) error {
	vpa, vpaExists := cluster.Vpas[vpaID]
	if !vpaExists {
		return NewKeyError(vpaID)
	}
	for _, state := range vpa.aggregateContainerStates {
		state.MarkNotAutoscaled()
	}
	delete(cluster.Vpas, vpaID)
	delete(cluster.EmptyVPAs, vpaID)
	return nil
}

func newPod(id PodID) *PodState {
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
func (cluster *ClusterState) MakeAggregateStateKey(pod *PodState, containerName string) AggregateStateKey {
	return aggregateStateKey{
		namespace:     pod.ID.Namespace,
		containerName: containerName,
		labelSetKey:   pod.labelSetKey,
		labelSetMap:   &cluster.labelSetMap,
	}
}

// aggregateStateKeyForContainerID returns the AggregateStateKey for the ContainerID.
// The pod with the corresponding PodID must already be present in the ClusterState.
func (cluster *ClusterState) aggregateStateKeyForContainerID(containerID ContainerID) AggregateStateKey {
	pod, podExists := cluster.Pods[containerID.PodID]
	if !podExists {
		panic(fmt.Sprintf("Pod not present in the ClusterState: %s/%s", containerID.PodID.Namespace, containerID.PodID.PodName))
	}
	return cluster.MakeAggregateStateKey(pod, containerID.ContainerName)
}

// findOrCreateAggregateContainerState returns (possibly newly created) AggregateContainerState
// that should be used to aggregate usage samples from container with a given ID.
// The pod with the corresponding PodID must already be present in the ClusterState.
func (cluster *ClusterState) findOrCreateAggregateContainerState(containerID ContainerID) *AggregateContainerState {
	aggregateStateKey := cluster.aggregateStateKeyForContainerID(containerID)
	aggregateContainerState, aggregateStateExists := cluster.aggregateStateMap[aggregateStateKey]
	if !aggregateStateExists {
		aggregateContainerState = NewAggregateContainerState()
		cluster.aggregateStateMap[aggregateStateKey] = aggregateContainerState
		// Link the new aggregation to the existing VPAs.
		for _, vpa := range cluster.Vpas {
			vpa.UseAggregationIfMatching(aggregateStateKey, aggregateContainerState)
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
func (cluster *ClusterState) garbageCollectAggregateCollectionStates(ctx context.Context, now time.Time, controllerFetcher controllerfetcher.ControllerFetcher) {
	klog.V(1).InfoS("Garbage collection of AggregateCollectionStates triggered")
	keysToDelete := make([]AggregateStateKey, 0)
	contributiveKeys := cluster.getContributiveAggregateStateKeys(ctx, controllerFetcher)
	for key, aggregateContainerState := range cluster.aggregateStateMap {
		isKeyContributive := contributiveKeys[key]
		if !isKeyContributive && aggregateContainerState.isEmpty() {
			keysToDelete = append(keysToDelete, key)
			klog.V(1).InfoS("Removing empty and not contributive AggregateCollectionState", "key", key)
			continue
		}
		if aggregateContainerState.isExpired(now) {
			keysToDelete = append(keysToDelete, key)
			klog.V(1).InfoS("Removing expired AggregateCollectionState", "key", key)
		}
	}
	for _, key := range keysToDelete {
		delete(cluster.aggregateStateMap, key)
		for _, vpa := range cluster.Vpas {
			vpa.DeleteAggregation(key)
		}
	}
}

// RateLimitedGarbageCollectAggregateCollectionStates removes obsolete AggregateCollectionStates from the ClusterState.
// It performs clean up only if more than `gcInterval` passed since the last time it performed a cleanup.
// AggregateCollectionState is obsolete in following situations:
// 1) It has no samples and there are no more contributive pods - a pod is contributive in any of following situations:
//
//	a) It is in an active state - i.e. not PodSucceeded nor PodFailed.
//	b) Its associated controller (e.g. Deployment) still exists.
//
// 2) The last sample is too old to give meaningful recommendation (>8 days),
// 3) There are no samples and the aggregate state was created >8 days ago.
func (cluster *ClusterState) RateLimitedGarbageCollectAggregateCollectionStates(ctx context.Context, now time.Time, controllerFetcher controllerfetcher.ControllerFetcher) {
	if now.Sub(cluster.lastAggregateContainerStateGC) < cluster.gcInterval {
		return
	}
	cluster.garbageCollectAggregateCollectionStates(ctx, now, controllerFetcher)
	cluster.lastAggregateContainerStateGC = now
}

func (cluster *ClusterState) getContributiveAggregateStateKeys(ctx context.Context, controllerFetcher controllerfetcher.ControllerFetcher) map[AggregateStateKey]bool {
	contributiveKeys := map[AggregateStateKey]bool{}
	for _, pod := range cluster.Pods {
		// Pod is considered contributive in any of following situations:
		// 1) It is in active state - i.e. not PodSucceeded nor PodFailed.
		// 2) Its associated controller (e.g. Deployment) still exists.
		podControllerExists := cluster.GetControllerForPodUnderVPA(ctx, pod, controllerFetcher) != nil
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
func (cluster *ClusterState) RecordRecommendation(vpa *Vpa, now time.Time) error {
	if vpa.Recommendation != nil && len(vpa.Recommendation.ContainerRecommendations) > 0 {
		delete(cluster.EmptyVPAs, vpa.ID)
		return nil
	}
	lastLogged, ok := cluster.EmptyVPAs[vpa.ID]
	if !ok {
		cluster.EmptyVPAs[vpa.ID] = now
	} else {
		if lastLogged.Add(RecommendationMissingMaxDuration).Before(now) {
			cluster.EmptyVPAs[vpa.ID] = now
			return fmt.Errorf("VPA %s/%s is missing recommendation for more than %v", vpa.ID.Namespace, vpa.ID.VpaName, RecommendationMissingMaxDuration)
		}
	}
	return nil
}

// GetMatchingPods returns a list of currently active pods that match the
// given VPA. Traverses through all pods in the cluster - use sparingly.
func (cluster *ClusterState) GetMatchingPods(vpa *Vpa) []PodID {
	matchingPods := []PodID{}
	for podID, pod := range cluster.Pods {
		if vpa_utils.PodLabelsMatchVPA(podID.Namespace, cluster.labelSetMap[pod.labelSetKey],
			vpa.ID.Namespace, vpa.PodSelector) {
			matchingPods = append(matchingPods, podID)
		}
	}
	return matchingPods
}

// GetControllerForPodUnderVPA returns controller associated with given Pod. Returns nil if Pod is not controlled by a VPA object.
func (cluster *ClusterState) GetControllerForPodUnderVPA(ctx context.Context, pod *PodState, controllerFetcher controllerfetcher.ControllerFetcher) *controllerfetcher.ControllerKeyWithAPIVersion {
	controllingVPA := cluster.GetControllingVPA(pod)
	if controllingVPA != nil {
		controller := &controllerfetcher.ControllerKeyWithAPIVersion{
			ControllerKey: controllerfetcher.ControllerKey{
				Namespace: controllingVPA.ID.Namespace,
				Kind:      controllingVPA.TargetRef.Kind,
				Name:      controllingVPA.TargetRef.Name,
			},
			ApiVersion: controllingVPA.TargetRef.APIVersion,
		}
		topLevelController, _ := controllerFetcher.FindTopMostWellKnownOrScalable(ctx, controller)
		return topLevelController
	}
	return nil
}

// GetControllingVPA returns a VPA object controlling given Pod.
func (cluster *ClusterState) GetControllingVPA(pod *PodState) *Vpa {
	for _, vpa := range cluster.Vpas {
		if vpa_utils.PodLabelsMatchVPA(pod.ID.Namespace, cluster.labelSetMap[pod.labelSetKey],
			vpa.ID.Namespace, vpa.PodSelector) {
			return vpa
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

// Namespace returns the namespace for the aggregateStateKey.
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
