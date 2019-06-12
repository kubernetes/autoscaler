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
	"sort"
	"time"

	autoscaling "k8s.io/api/autoscaling/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

// Map from VPA condition type to condition.
type vpaConditionsMap map[vpa_types.VerticalPodAutoscalerConditionType]vpa_types.VerticalPodAutoscalerCondition

func (conditionsMap *vpaConditionsMap) Set(
	conditionType vpa_types.VerticalPodAutoscalerConditionType,
	status bool, reason string, message string) *vpaConditionsMap {
	oldCondition, alreadyPresent := (*conditionsMap)[conditionType]
	condition := vpa_types.VerticalPodAutoscalerCondition{
		Type:    conditionType,
		Reason:  reason,
		Message: message,
	}
	if status {
		condition.Status = apiv1.ConditionTrue
	} else {
		condition.Status = apiv1.ConditionFalse
	}
	if alreadyPresent && oldCondition.Status == condition.Status {
		condition.LastTransitionTime = oldCondition.LastTransitionTime
	} else {
		condition.LastTransitionTime = metav1.Now()
	}
	(*conditionsMap)[conditionType] = condition
	return conditionsMap
}

func (conditionsMap *vpaConditionsMap) AsList() []vpa_types.VerticalPodAutoscalerCondition {
	conditions := make([]vpa_types.VerticalPodAutoscalerCondition, 0, len(*conditionsMap))
	for _, condition := range *conditionsMap {
		conditions = append(conditions, condition)
	}

	// Sort conditions by type to avoid elements floating on the list
	sort.Slice(conditions, func(i, j int) bool {
		return conditions[i].Type < conditions[j].Type
	})

	return conditions
}

func (conditionsMap *vpaConditionsMap) ConditionActive(conditionType vpa_types.VerticalPodAutoscalerConditionType) bool {
	condition, found := (*conditionsMap)[conditionType]
	return found && condition.Status == apiv1.ConditionTrue
}

// Vpa (Vertical Pod Autoscaler) object is responsible for vertical scaling of
// Pods matching a given label selector.
type Vpa struct {
	ID VpaID
	// Labels selector that determines which Pods are controlled by this VPA
	// object. Can be nil, in which case no Pod is matched.
	PodSelector labels.Selector
	// Map of the status conditions (keys are condition types).
	Conditions vpaConditionsMap
	// Most recently computed recommendation. Can be nil.
	Recommendation *vpa_types.RecommendedPodResources
	// All container aggregations that contribute to this VPA.
	// TODO: Garbage collect old AggregateContainerStates.
	aggregateContainerStates aggregateContainerStatesMap
	// Pod Resource Policy provided in the VPA API object. Can be nil.
	ResourcePolicy *vpa_types.PodResourcePolicy
	// Initial checkpoints of AggregateContainerStates for containers.
	// The key is container name.
	ContainersInitialAggregateState ContainerNameToAggregateStateMap
	// UpdateMode describes how recommendations will be applied to pods
	UpdateMode *vpa_types.UpdateMode
	// Created denotes timestamp of the original VPA object creation
	Created time.Time
	// CheckpointWritten indicates when last checkpoint for the VPA object was stored.
	CheckpointWritten time.Time
	// IsV1Beta1API is set to true if VPA object has labelSelector defined as in v1beta1 api.
	IsV1Beta1API bool
	// TargetRef points to the controller managing the set of pods.
	TargetRef *autoscaling.CrossVersionObjectReference
}

// NewVpa returns a new Vpa with a given ID and pod selector. Doesn't set the
// links to the matched aggregations.
func NewVpa(id VpaID, selector labels.Selector, created time.Time) *Vpa {
	vpa := &Vpa{
		ID:                              id,
		PodSelector:                     selector,
		aggregateContainerStates:        make(aggregateContainerStatesMap),
		ContainersInitialAggregateState: make(ContainerNameToAggregateStateMap),
		Created:                         created,
		Conditions:                      make(vpaConditionsMap),
		IsV1Beta1API:                    false,
	}
	return vpa
}

// UseAggregationIfMatching checks if the given aggregation matches (contributes to) this VPA
// and adds it to the set of VPA's aggregations if that is the case. Returns true
// if the aggregation matches VPA.
func (vpa *Vpa) UseAggregationIfMatching(aggregationKey AggregateStateKey, aggregation *AggregateContainerState) bool {
	if vpa.UsesAggregation(aggregationKey) {
		return true
	}
	if vpa.matchesAggregation(aggregationKey) {
		vpa.aggregateContainerStates[aggregationKey] = aggregation
		aggregation.IsUnderVPA = true
		aggregation.UpdateMode = vpa.UpdateMode
		return true
	}
	return false
}

// UpdateRecommendation updates the recommended resources in the VPA and its
// aggregations with the given recommendation.
func (vpa *Vpa) UpdateRecommendation(recommendation *vpa_types.RecommendedPodResources) {
	vpa.Recommendation = recommendation
	for _, containerRecommendation := range recommendation.ContainerRecommendations {
		for container, state := range vpa.aggregateContainerStates {
			if container.ContainerName() == containerRecommendation.ContainerName {
				state.LastRecommendation = containerRecommendation.UncappedTarget
			}
		}
	}
}

// UsesAggregation returns true iff an aggregation with the given key contributes to the VPA.
func (vpa *Vpa) UsesAggregation(aggregationKey AggregateStateKey) bool {
	_, exists := vpa.aggregateContainerStates[aggregationKey]
	return exists
}

// DeleteAggregation deletes aggregation used by this container
func (vpa *Vpa) DeleteAggregation(aggregationKey AggregateStateKey) {
	state := vpa.aggregateContainerStates[aggregationKey]
	state.MarkNotAutoscaled()
	delete(vpa.aggregateContainerStates, aggregationKey)
}

// MergeCheckpointedState adds checkpointed VPA aggregations to the given aggregateStateMap.
func (vpa *Vpa) MergeCheckpointedState(aggregateContainerStateMap ContainerNameToAggregateStateMap) {
	for containerName, aggregation := range vpa.ContainersInitialAggregateState {
		aggregateContainerState, found := aggregateContainerStateMap[containerName]
		if !found {
			aggregateContainerState = NewAggregateContainerState()
			aggregateContainerStateMap[containerName] = aggregateContainerState
		}
		aggregateContainerState.MergeContainerState(aggregation)
	}
}

// AggregateStateByContainerName returns a map from container name to the aggregated state
// of all containers with that name, belonging to pods matched by the VPA.
func (vpa *Vpa) AggregateStateByContainerName() ContainerNameToAggregateStateMap {
	containerNameToAggregateStateMap := AggregateStateByContainerName(vpa.aggregateContainerStates)
	vpa.MergeCheckpointedState(containerNameToAggregateStateMap)
	return containerNameToAggregateStateMap
}

// HasRecommendation returns if the VPA object contains any recommendation
func (vpa *Vpa) HasRecommendation() bool {
	return (vpa.Recommendation != nil) && len(vpa.Recommendation.ContainerRecommendations) > 0
}

// matchesAggregation returns true iff the VPA matches the given aggregation key.
func (vpa *Vpa) matchesAggregation(aggregationKey AggregateStateKey) bool {
	if vpa.ID.Namespace != aggregationKey.Namespace() {
		return false
	}
	return vpa.PodSelector != nil && vpa.PodSelector.Matches(aggregationKey.Labels())
}

// UpdateConditions updates the conditions of VPA objects based on it's state.
// PodsMatched is passed to indicate if there are currently active pods in the
// cluster matching this VPA.
func (vpa *Vpa) UpdateConditions(podsMatched bool) {
	reason := ""
	msg := ""
	if podsMatched {
		delete(vpa.Conditions, vpa_types.NoPodsMatched)
	} else {
		reason = "NoPodsMatched"
		msg = "No pods match this VPA object"
		vpa.Conditions.Set(vpa_types.NoPodsMatched, true, reason, msg)
	}
	if vpa.HasRecommendation() {
		vpa.Conditions.Set(vpa_types.RecommendationProvided, true, "", "")
	} else {
		vpa.Conditions.Set(vpa_types.RecommendationProvided, false, reason, msg)
	}

}

// AsStatus returns this objects equivalent of VPA Status. UpdateConditions
// should be called first.
func (vpa *Vpa) AsStatus() *vpa_types.VerticalPodAutoscalerStatus {
	status := &vpa_types.VerticalPodAutoscalerStatus{
		Conditions: vpa.Conditions.AsList(),
	}
	if vpa.Recommendation != nil {
		status.Recommendation = vpa.Recommendation
	}
	return status
}

// HasMatchedPods returns true if there are are currently active pods in the
// cluster matching this VPA, based on conditions. UpdateConditions should be
// called first.
func (vpa *Vpa) HasMatchedPods() bool {
	noPodsMatched, found := vpa.Conditions[vpa_types.NoPodsMatched]
	if found && noPodsMatched.Status == apiv1.ConditionTrue {
		return false
	}
	return true
}
