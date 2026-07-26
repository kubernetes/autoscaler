/*
Copyright 2018 The Kubernetes Authors.

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

package routines

import (
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	scopeutil "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/scope"
	api_utils "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

func isNoopContainerPolicy(policy *vpa_types.ContainerResourcePolicy) bool {
	// Fast path for the common "no effective overrides" policy shape.
	// We use it to skip policy copy/apply work for every scope group.
	if policy == nil {
		return true
	}
	if policy.Mode != nil && *policy.Mode != vpa_types.ContainerScalingModeAuto {
		return false
	}
	return len(policy.MinAllowed) == 0 &&
		len(policy.MaxAllowed) == 0 &&
		policy.ControlledResources == nil &&
		policy.ControlledValues == nil &&
		policy.OOMBumpUpRatio == nil &&
		policy.OOMMinBumpUp == nil
}

func isNoopPodPolicy(policy *vpa_types.PodResourcePolicy) bool {
	if policy == nil {
		return true
	}
	for i := range policy.ContainerPolicies {
		if !isNoopContainerPolicy(&policy.ContainerPolicies[i]) {
			return false
		}
	}
	return true
}

// GetContainerNameToAggregateStateMap returns ContainerNameToAggregateStateMap for pods.
func GetContainerNameToAggregateStateMap(vpa *model.Vpa) model.ContainerNameToAggregateStateMap {
	containerNameToAggregateStateMap := vpa.AggregateStateByContainerName()
	if isNoopPodPolicy(vpa.ResourcePolicy) {
		// No filtering/transformation is needed; avoid allocating and copying maps.
		return containerNameToAggregateStateMap
	}
	filteredContainerNameToAggregateStateMap := make(model.ContainerNameToAggregateStateMap)

	for containerName, aggregatedContainerState := range containerNameToAggregateStateMap {
		containerResourcePolicy := api_utils.GetContainerResourcePolicy(containerName, vpa.ResourcePolicy)
		autoscalingDisabled := containerResourcePolicy != nil && containerResourcePolicy.Mode != nil &&
			*containerResourcePolicy.Mode == vpa_types.ContainerScalingModeOff
		if !autoscalingDisabled {
			if isNoopContainerPolicy(containerResourcePolicy) {
				// Reuse already aggregated state to avoid per-container recomputation.
				filteredContainerNameToAggregateStateMap[containerName] = aggregatedContainerState
				continue
			}
			aggregatedContainerState.UpdateFromPolicy(containerResourcePolicy)
			filteredContainerNameToAggregateStateMap[containerName] = aggregatedContainerState
		}
	}
	return filteredContainerNameToAggregateStateMap
}

// GetContainerNameToAggregateStateMapByScopeValue returns per-scope aggregated states for pods.
func GetContainerNameToAggregateStateMapByScopeValue(vpa *model.Vpa) map[string]model.ContainerNameToAggregateStateMap {
	scopeGroups := vpa.AggregateStateByScopeValueAndContainerName(scopeutil.AggregationLabelKey(vpa.Scope))
	if isNoopPodPolicy(vpa.ResourcePolicy) {
		// Common DaemonSet case: all container policies are effectively defaults.
		// Returning grouped state directly avoids a full second map allocation pass.
		return scopeGroups
	}
	filteredGroups := make(map[string]model.ContainerNameToAggregateStateMap, len(scopeGroups))

	for scopeValue, containerNameToAggregateStateMap := range scopeGroups {
		filteredGroups[scopeValue] = make(model.ContainerNameToAggregateStateMap, len(containerNameToAggregateStateMap))
		for containerName, aggregatedContainerState := range containerNameToAggregateStateMap {
			containerResourcePolicy := api_utils.GetContainerResourcePolicy(containerName, vpa.ResourcePolicy)
			autoscalingDisabled := containerResourcePolicy != nil && containerResourcePolicy.Mode != nil &&
				*containerResourcePolicy.Mode == vpa_types.ContainerScalingModeOff
			if !autoscalingDisabled {
				if isNoopContainerPolicy(containerResourcePolicy) {
					// No policy override: reuse the state instead of allocating/copying per scope group.
					filteredGroups[scopeValue][containerName] = aggregatedContainerState
					continue
				}
				// Policy overrides should not mutate shared aggregate state.
				policyAwareState := model.NewAggregateContainerState()
				policyAwareState.MergeContainerState(aggregatedContainerState)
				policyAwareState.UpdateFromPolicy(containerResourcePolicy)
				filteredGroups[scopeValue][containerName] = policyAwareState
			}
		}
	}

	return filteredGroups
}
