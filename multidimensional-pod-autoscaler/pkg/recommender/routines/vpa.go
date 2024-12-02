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
	"k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/recommender/model"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_model "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	api_utils "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

// GetContainerNameToAggregateStateMap returns ContainerNameToAggregateStateMap for pods.
// Updated to integrate with the MPA API.
func GetContainerNameToAggregateStateMap(mpa *model.Mpa) vpa_model.ContainerNameToAggregateStateMap {
	containerNameToAggregateStateMap := mpa.AggregateStateByContainerName()
	filteredContainerNameToAggregateStateMap := make(vpa_model.ContainerNameToAggregateStateMap)

	for containerName, aggregatedContainerState := range containerNameToAggregateStateMap {
		containerResourcePolicy := api_utils.GetContainerResourcePolicy(containerName, mpa.ResourcePolicy)
		autoscalingDisabled := containerResourcePolicy != nil && containerResourcePolicy.Mode != nil &&
			*containerResourcePolicy.Mode == vpa_types.ContainerScalingModeOff
		if !autoscalingDisabled && aggregatedContainerState.TotalSamplesCount > 0 {
			aggregatedContainerState.UpdateFromPolicy(containerResourcePolicy)
			vpaAggregatedContainerState := vpa_model.AggregateContainerState{
                AggregateCPUUsage: aggregatedContainerState.AggregateCPUUsage,
                AggregateMemoryPeaks: aggregatedContainerState.AggregateMemoryPeaks,
                FirstSampleStart: aggregatedContainerState.FirstSampleStart,
                LastSampleStart: aggregatedContainerState.LastSampleStart,
                TotalSamplesCount: aggregatedContainerState.TotalSamplesCount,
                CreationTime: aggregatedContainerState.CreationTime,
                LastRecommendation: aggregatedContainerState.LastRecommendation,
                IsUnderVPA: aggregatedContainerState.IsUnderVPA,
                UpdateMode: aggregatedContainerState.UpdateMode,
                ScalingMode: aggregatedContainerState.ScalingMode,
                ControlledResources: aggregatedContainerState.ControlledResources,
        }
			filteredContainerNameToAggregateStateMap[containerName] = &vpaAggregatedContainerState
		}
	}
	return filteredContainerNameToAggregateStateMap
}
