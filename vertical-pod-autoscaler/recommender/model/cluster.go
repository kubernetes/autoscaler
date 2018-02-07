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
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
)

// ClusterState holds all runtime information about the cluster required for the
// VPA operations, i.e. configuration of resources (pods, containers,
// VPA objects), aggregated utilization of compute resources (CPU, memory) and
// events (container OOMs).
// All input to the VPA Recommender algorithm lives in this structure.
// TODO(kgrygiel): Limit the ClusterState object to a single namespace.
type ClusterState struct {
	// Pods in the cluster.
	Pods map[PodID]*PodState
	// VPA objects in the cluster.
	Vpas map[VpaID]*Vpa
}

// PodState holds runtime information about a single Pod.
type PodState struct {
	// Unique id of the Pod.
	ID PodID
	// Set of labels attached to the Pod.
	Labels labels.Set
	// Containers that belong to the Pod, keyed by the container name.
	Containers map[string]*ContainerState
	// All VPA objects that match this Pod. While it is incorrect to let
	// multiple VPA objects match the same pod, the model has no means to
	// prevent such situation. In such case the pod is controlled by one of the
	// matching VPAs.
	MatchingVpas map[VpaID]*Vpa
}

// NewClusterState returns a new ClusterState with no pods.
func NewClusterState() *ClusterState {
	return &ClusterState{
		make(map[PodID]*PodState), // empty pods map.
		make(map[VpaID]*Vpa),      // empty vpas map.
	}
}

// ContainerUsageSampleWithKey holds a ContainerUsageSample together with the
// ID of the container it belongs to.
type ContainerUsageSampleWithKey struct {
	ContainerUsageSample
	Container ContainerID
}

// AddOrUpdatePod udpates the state of the pod with a given PodID, if it is
// present in the cluster object. Otherwise a new pod is created and added to
// the Cluster object.
// If the labels of the pod have changed, it updates the links between the pod
// and the matching Vpa.
func (cluster *ClusterState) AddOrUpdatePod(podID PodID, newLabels labels.Set) {
	pod, podExists := cluster.Pods[podID]
	if !podExists {
		pod = newPod(podID)
		cluster.Pods[podID] = pod
	}
	if !podExists || !labels.Equals(pod.Labels, newLabels) {
		// Update the labels and the links between the pod and Vpas.
		pod.Labels = newLabels
		for _, vpa := range cluster.Vpas {
			vpa.UpdatePodLink(pod)
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
func (cluster *ClusterState) DeletePod(podID PodID) error {
	pod, podExists := cluster.Pods[podID]
	if !podExists {
		return NewKeyError(podID)
	}
	// Set labels to nil so that no VPA matches the pod.
	pod.Labels = nil
	for _, vpa := range pod.MatchingVpas {
		vpa.UpdatePodLink(pod)
	}
	delete(cluster.Pods, podID)
	return nil
}

// AddOrUpdateContainer creates a new container with the given ContainerID and
// adds it to the parent pod in the ClusterState object, if not yet present.
// Requires the pod to be added to the ClusterState first. Otherwise an error is
// returned.
func (cluster *ClusterState) AddOrUpdateContainer(containerID ContainerID) error {
	pod, podExists := cluster.Pods[containerID.PodID]
	if !podExists {
		return NewKeyError(containerID.PodID)
	}
	if _, containerExists := pod.Containers[containerID.ContainerName]; !containerExists {
		pod.Containers[containerID.ContainerName] = NewContainerState()
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
	containerState.AddSample(&sample.ContainerUsageSample)
	return nil
}

// AddOrUpdateVpa adds a new VPA with a given ID to the ClusterState if it
// didn't yet exist. If the VPA already existed but had a different pod
// selector, the pod selector is updated. Updates the links between the VPA and
// all pods it matches.
func (cluster *ClusterState) AddOrUpdateVpa(apiObject *vpa_types.VerticalPodAutoscaler) error {
	vpaID := VpaID{Namespace: apiObject.Namespace, VpaName: apiObject.Name}
	conditionsMap := make(vpaConditionsMap)
	for _, condition := range apiObject.Status.Conditions {
		conditionsMap[condition.Type] = condition
	}
	var currentRecommendation *vpa_types.RecommendedPodResources
	if conditionsMap[vpa_types.RecommendationProvided].Status == apiv1.ConditionTrue {
		currentRecommendation = &apiObject.Status.Recommendation
	}
	selector, err := metav1.LabelSelectorAsSelector(apiObject.Spec.Selector)
	if err != nil {
		errMsg := fmt.Sprintf("couldn't convert selector into a corresponding internal selector object: %v", err)
		conditionsMap.Set(vpa_types.Configured, false, "InvalidSelector", errMsg)
	}

	vpa, vpaExists := cluster.Vpas[vpaID]
	if vpaExists && (err != nil || vpa.PodSelector.String() != selector.String()) {
		// Pod selector was changed. Delete the VPA object and recreate
		// it with the new selector.
		if err := cluster.DeleteVpa(vpaID); err != nil {
			return err
		}
		vpaExists = false
	}
	if !vpaExists {
		vpa = NewVpa(vpaID, selector)
		cluster.Vpas[vpaID] = vpa
		for _, pod := range cluster.Pods {
			vpa.UpdatePodLink(pod)
		}
	}
	vpa.Conditions = conditionsMap
	vpa.Recommendation = currentRecommendation
	vpa.LastUpdateTime = apiObject.Status.LastUpdateTime
	return nil
}

// DeleteVpa removes a VPA with the given ID from the ClusterState.
func (cluster *ClusterState) DeleteVpa(vpaID VpaID) error {
	vpa, vpaExists := cluster.Vpas[vpaID]
	if !vpaExists {
		return NewKeyError(vpaID)
	}
	// Change the selector to not match any pod and detach all pods.
	vpa.PodSelector = nil
	for _, pod := range vpa.Pods {
		vpa.UpdatePodLink(pod)
	}
	delete(cluster.Vpas, vpaID)
	return nil
}

func newPod(id PodID) *PodState {
	return &PodState{
		ID:           id,
		Labels:       make(map[string]string),
		Containers:   make(map[string]*ContainerState),
		MatchingVpas: make(map[VpaID]*Vpa),
	}
}
