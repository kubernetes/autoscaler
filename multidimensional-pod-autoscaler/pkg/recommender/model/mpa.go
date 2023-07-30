package model

import (
	"sort"
	"time"

	autoscaling "k8s.io/api/autoscaling/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	mpa_types "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1alpha1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_model "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	metrics_quality "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/quality"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

// Map from MPA annotation key to value.
type mpaAnnotationsMap map[string]string

// Map from MPA condition type to condition.
type mpaConditionsMap map[mpa_types.MultidimPodAutoscalerConditionType]mpa_types.MultidimPodAutoscalerCondition

func (conditionsMap *mpaConditionsMap) Set(
	conditionType mpa_types.MultidimPodAutoscalerConditionType,
	status bool, reason string, message string) *mpaConditionsMap {
	oldCondition, alreadyPresent := (*conditionsMap)[conditionType]
	condition := mpa_types.MultidimPodAutoscalerCondition{
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

func (conditionsMap *mpaConditionsMap) AsList() []mpa_types.MultidimPodAutoscalerCondition {
	conditions := make([]mpa_types.MultidimPodAutoscalerCondition, 0, len(*conditionsMap))
	for _, condition := range *conditionsMap {
		conditions = append(conditions, condition)
	}

	// Sort conditions by type to avoid elements floating on the list
	sort.Slice(conditions, func(i, j int) bool {
		return conditions[i].Type < conditions[j].Type
	})

	return conditions
}

func (conditionsMap *mpaConditionsMap) ConditionActive(conditionType mpa_types.MultidimPodAutoscalerConditionType) bool {
	condition, found := (*conditionsMap)[conditionType]
	return found && condition.Status == apiv1.ConditionTrue
}

// Mpa (Multidimensional Pod Autoscaler) object is responsible for horizontal and vertical scaling
// of Pods matching a given label selector.
type Mpa struct {
	ID MpaID
	// Labels selector that determines which Pods are controlled by this MPA
	// object. Can be nil, in which case no Pod is matched.
	PodSelector labels.Selector
	// Map of the object annotations (key-value pairs).
	Annotations mpaAnnotationsMap
	// Map of the status conditions (keys are condition types).
	Conditions mpaConditionsMap
	// Most recently computed recommendation. Can be nil.
	Recommendation *vpa_types.RecommendedPodResources
	// All container aggregations that contribute to this MPA.
	// TODO: Garbage collect old AggregateContainerStates.
	aggregateContainerStates aggregateContainerStatesMap
	// Pod Resource Policy provided in the MPA API object. Can be nil.
	ResourcePolicy *vpa_types.PodResourcePolicy
	// Initial checkpoints of AggregateContainerStates for containers.
	// The key is container name.
	ContainersInitialAggregateState vpa_model.ContainerNameToAggregateStateMap
	// UpdateMode describes how recommendations will be applied to pods
	UpdateMode *vpa_types.UpdateMode
	// Created denotes timestamp of the original MPA object creation
	Created time.Time
	// CheckpointWritten indicates when last checkpoint for the MPA object was stored.
	CheckpointWritten time.Time
	// IsV1Beta1API is set to true if MPA object has labelSelector defined as in v1beta1 api.
	IsV1Beta1API bool
	// ScaleTargetRef points to the controller managing the set of pods.
	ScaleTargetRef *autoscaling.CrossVersionObjectReference
	// PodCount contains number of live Pods matching a given MPA object.
	PodCount int

	// Added for HPA-related fields.
	// TODO: Currently HPA-related logic is directly manipulating the MPA object but not the MPA
	// model here.
	Metrics []autoscalingv2.MetricSpec
	MinReplicas int32
	MaxReplicas int32
	HorizontalScalingBehavior *autoscalingv2.HorizontalPodAutoscalerBehavior
	DesiredReplicas int32
	CurrentMetrics []autoscalingv2.MetricStatus
}

// NewMpa returns a new Mpa with a given ID and pod selector. Doesn't set the
// links to the matched aggregations.
func NewMpa(id MpaID, selector labels.Selector, created time.Time) *Mpa {
	mpa := &Mpa{
		ID:                              id,
		PodSelector:                     selector,
		aggregateContainerStates:        make(aggregateContainerStatesMap),
		ContainersInitialAggregateState: make(vpa_model.ContainerNameToAggregateStateMap),
		Created:                         created,
		Annotations:                     make(mpaAnnotationsMap),
		Conditions:                      make(mpaConditionsMap),
		IsV1Beta1API:                    false,
		PodCount:                        0,
	}
	return mpa
}

// UseAggregationIfMatching checks if the given aggregation matches (contributes to) this MPA
// and adds it to the set of MPA's aggregations if that is the case.
func (mpa *Mpa) UseAggregationIfMatching(aggregationKey vpa_model.AggregateStateKey, aggregation *vpa_model.AggregateContainerState) {
	if mpa.UsesAggregation(aggregationKey) {
		// Already linked, we can return quickly.
		return
	}
	if mpa.matchesAggregation(aggregationKey) {
		mpa.aggregateContainerStates[aggregationKey] = aggregation
		aggregation.IsUnderVPA = true
		aggregation.UpdateMode = mpa.UpdateMode
		aggregation.UpdateFromPolicy(vpa_api_util.GetContainerResourcePolicy(aggregationKey.ContainerName(), mpa.ResourcePolicy))
	}
}

// UsesAggregation returns true iff an aggregation with the given key contributes to the MPA.
func (mpa *Mpa) UsesAggregation(aggregationKey vpa_model.AggregateStateKey) bool {
	_, exists := mpa.aggregateContainerStates[aggregationKey]
	return exists
}

// DeleteAggregation deletes aggregation used by this container
func (mpa *Mpa) DeleteAggregation(aggregationKey vpa_model.AggregateStateKey) {
	state, ok := mpa.aggregateContainerStates[aggregationKey]
	if !ok {
		return
	}
	state.MarkNotAutoscaled()
	delete(mpa.aggregateContainerStates, aggregationKey)
}

// matchesAggregation returns true iff the MPA matches the given aggregation key.
func (mpa *Mpa) matchesAggregation(aggregationKey vpa_model.AggregateStateKey) bool {
	if mpa.ID.Namespace != aggregationKey.Namespace() {
		return false
	}
	return mpa.PodSelector != nil && mpa.PodSelector.Matches(aggregationKey.Labels())
}

// SetResourcePolicy updates the resource policy of the MPA and the scaling
// policies of aggregators under this MPA.
func (mpa *Mpa) SetResourcePolicy(resourcePolicy *vpa_types.PodResourcePolicy) {
	if resourcePolicy == mpa.ResourcePolicy {
		return
	}
	mpa.ResourcePolicy = resourcePolicy
	for container, state := range mpa.aggregateContainerStates {
		state.UpdateFromPolicy(vpa_api_util.GetContainerResourcePolicy(container.ContainerName(), mpa.ResourcePolicy))
	}
}

// SetUpdateMode updates the update mode of the MPA and aggregators under this MPA.
func (mpa *Mpa) SetUpdateMode(updatePolicy *mpa_types.PodUpdatePolicy) {
	if updatePolicy == nil {
		mpa.UpdateMode = nil
	} else {
		if updatePolicy.UpdateMode == mpa.UpdateMode {
			return
		}
		mpa.UpdateMode = updatePolicy.UpdateMode
	}
	for _, state := range mpa.aggregateContainerStates {
		state.UpdateMode = mpa.UpdateMode
	}
}

// Set HPA-related constraints.
func (mpa *Mpa) SetHPAConstraints(metrics []autoscalingv2.MetricSpec, minReplicas int32, maxReplicas int32, hpaBehavior *autoscalingv2.HorizontalPodAutoscalerBehavior) {
	mpa.Metrics = metrics
	mpa.MinReplicas = minReplicas
	mpa.MaxReplicas = maxReplicas
	mpa.HorizontalScalingBehavior = hpaBehavior
}

// Set the desired number of replicas.
func (mpa *Mpa) SetDesiredNumberOfReplicas(replicas int32) {
	mpa.DesiredReplicas = replicas
}

// Set the current metrics.
func (mpa *Mpa) SetCurrentMetrics(metrics []autoscalingv2.MetricStatus) {
	mpa.CurrentMetrics = metrics
}

// MergeCheckpointedState adds checkpointed MPA aggregations to the given aggregateStateMap.
func (mpa *Mpa) MergeCheckpointedState(aggregateContainerStateMap vpa_model.ContainerNameToAggregateStateMap) {
	for containerName, aggregation := range mpa.ContainersInitialAggregateState {
		aggregateContainerState, found := aggregateContainerStateMap[containerName]
		if !found {
			aggregateContainerState = vpa_model.NewAggregateContainerState()
			aggregateContainerStateMap[containerName] = aggregateContainerState
		}
		aggregateContainerState.MergeContainerState(aggregation)
	}
}

// AggregateStateByContainerName returns a map from container name to the aggregated state
// of all containers with that name, belonging to pods matched by the MPA.
func (mpa *Mpa) AggregateStateByContainerName() vpa_model.ContainerNameToAggregateStateMap {
	containerNameToAggregateStateMap := AggregateStateByContainerName(mpa.aggregateContainerStates)
	mpa.MergeCheckpointedState(containerNameToAggregateStateMap)
	return containerNameToAggregateStateMap
}

// HasRecommendation returns if the MPA object contains any recommendation
func (mpa *Mpa) HasRecommendation() bool {
	return (mpa.Recommendation != nil) && len(mpa.Recommendation.ContainerRecommendations) > 0
}

// UpdateRecommendation updates the recommended resources in the MPA and its
// aggregations with the given recommendation.
func (mpa *Mpa) UpdateRecommendation(recommendation *vpa_types.RecommendedPodResources) {
	for _, containerRecommendation := range recommendation.ContainerRecommendations {
		for container, state := range mpa.aggregateContainerStates {
			if container.ContainerName() == containerRecommendation.ContainerName {
				metrics_quality.ObserveRecommendationChange(state.LastRecommendation, containerRecommendation.UncappedTarget, mpa.UpdateMode, mpa.PodCount)
				state.LastRecommendation = containerRecommendation.UncappedTarget
			}
		}
	}
	mpa.Recommendation = recommendation
}

// UpdateConditions updates the conditions of MPA objects based on it's state.
// PodsMatched is passed to indicate if there are currently active pods in the
// cluster matching this MPA.
func (mpa *Mpa) UpdateConditions(podsMatched bool) {
	reason := ""
	msg := ""
	if podsMatched {
		delete(mpa.Conditions, mpa_types.NoPodsMatched)
	} else {
		reason = "NoPodsMatched"
		msg = "No pods match this MPA object"
		mpa.Conditions.Set(mpa_types.NoPodsMatched, true, reason, msg)
	}
	if mpa.HasRecommendation() {
		mpa.Conditions.Set(mpa_types.RecommendationProvided, true, "", "")
	} else {
		mpa.Conditions.Set(mpa_types.RecommendationProvided, false, reason, msg)
	}

}

// HasMatchedPods returns true if there are are currently active pods in the
// cluster matching this MPA, based on conditions. UpdateConditions should be
// called first.
func (mpa *Mpa) HasMatchedPods() bool {
	noPodsMatched, found := mpa.Conditions[mpa_types.NoPodsMatched]
	if found && noPodsMatched.Status == apiv1.ConditionTrue {
		return false
	}
	return true
}

// AsStatus returns this objects equivalent of MPA Status. UpdateConditions
// should be called first.
func (mpa *Mpa) AsStatus() *mpa_types.MultidimPodAutoscalerStatus {
	status := &mpa_types.MultidimPodAutoscalerStatus{
		Conditions: mpa.Conditions.AsList(),
	}
	if mpa.Recommendation != nil {
		status.Recommendation = mpa.Recommendation
	}
	return status
}
