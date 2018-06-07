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
	"time"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
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
	return conditions
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
	// Value of the Status.LastUpdateTime fetched from the VPA API object.
	LastUpdateTime time.Time
	// Initial checkpoints of AggregateContainerStates for containers.
	// The key is container name.
	ContainersInitialAggregateState ContainerNameToAggregateStateMap
}

// NewVpa returns a new Vpa with a given ID and pod selector. Doesn't set the
// links to the matched aggregations.
func NewVpa(id VpaID, selector labels.Selector) *Vpa {
	vpa := &Vpa{
		ID:                              id,
		PodSelector:                     selector,
		aggregateContainerStates:        make(aggregateContainerStatesMap),
		ContainersInitialAggregateState: make(ContainerNameToAggregateStateMap),
	}
	return vpa
}

// UseAggregationIfMatching checks if the given aggregation matches (contributes to) this VPA
// and adds it to the set of VPA's aggregations if that is the case.
func (vpa *Vpa) UseAggregationIfMatching(aggregationKey AggregateStateKey, aggregation *AggregateContainerState) {
	if !vpa.UsesAggregation(aggregationKey) && vpa.matchesAggregation(aggregationKey) {
		vpa.aggregateContainerStates[aggregationKey] = aggregation
	}
}

// UsesAggregation returns true iff an aggregation with the given key contributes to the VPA.
func (vpa *Vpa) UsesAggregation(aggregationKey AggregateStateKey) bool {
	_, exists := vpa.aggregateContainerStates[aggregationKey]
	return exists
}

// MergeCheckpointedState adds checkpointed VPA aggregations to the given aggregateStateMap.
func (vpa *Vpa) MergeCheckpointedState(aggregateContainerStateMap ContainerNameToAggregateStateMap) {
	for containerName, aggregation := range vpa.ContainersInitialAggregateState {
		aggregateContainerStateMap[containerName].MergeContainerState(aggregation)
	}
}

// AggregateStateByContainerName returns a map from container name to the aggregated state
// of all containers with that name, belonging to pods matched by the VPA.
func (vpa *Vpa) AggregateStateByContainerName() ContainerNameToAggregateStateMap {
	containerNameToAggregateStateMap := AggregateStateByContainerName(vpa.aggregateContainerStates)
	vpa.MergeCheckpointedState(containerNameToAggregateStateMap)
	return containerNameToAggregateStateMap
}

// matchesAggregation returns true iff the VPA matches the given aggregation key.
func (vpa *Vpa) matchesAggregation(aggregationKey AggregateStateKey) bool {
	if vpa.ID.Namespace != aggregationKey.Namespace() {
		return false
	}
	return vpa.PodSelector != nil && vpa.PodSelector.Matches(aggregationKey.Labels())
}
