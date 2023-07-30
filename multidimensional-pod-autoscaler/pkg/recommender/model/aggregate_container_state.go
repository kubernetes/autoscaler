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

// VPA collects CPU and memory usage measurements from all containers running in
// the cluster and aggregates them in memory in structures called
// AggregateContainerState.
// During aggregation the usage samples are grouped together by the key called
// AggregateStateKey and stored in structures such as histograms of CPU and
// memory usage, that are parts of the AggregateContainerState.
//
// The AggregateStateKey consists of the container name, the namespace and the
// set of labels on the pod the container belongs to. In other words, whenever
// two samples come from containers with the same name, in the same namespace
// and with the same pod labels, they end up in the same histogram.
//
// Recall that VPA produces one recommendation for all containers with a given
// name and namespace, having pod labels that match a given selector. Therefore
// for each VPA object and container name the recommender has to take all
// matching AggregateContainerStates and further aggregate them together, in
// order to obtain the final aggregation that is the input to the recommender
// function.

package model

import (
	"time"

	corev1 "k8s.io/api/core/v1"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_model "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

// The isStateExpired and isStateEmpty functions are from VPA lib (because isExpired and isEmpty
// are private functions so not callable).
func isStateExpired(a *vpa_model.AggregateContainerState, now time.Time) bool {
	if isStateEmpty(a) {
		return now.Sub(a.CreationTime) >= vpa_model.GetAggregationsConfig().GetMemoryAggregationWindowLength()
	}
	return now.Sub(a.LastSampleStart) >= vpa_model.GetAggregationsConfig().GetMemoryAggregationWindowLength()
}

func isStateEmpty(a *vpa_model.AggregateContainerState) bool {
	return a.TotalSamplesCount == 0
}

// AggregateStateByContainerName takes a set of AggregateContainerStates and merge them
// grouping by the container name. The result is a map from the container name to the aggregation
// from all input containers with the given name.
func AggregateStateByContainerName(aggregateContainerStateMap aggregateContainerStatesMap) vpa_model.ContainerNameToAggregateStateMap {
	containerNameToAggregateStateMap := make(vpa_model.ContainerNameToAggregateStateMap)
	for aggregationKey, aggregation := range aggregateContainerStateMap {
		containerName := aggregationKey.ContainerName()
		aggregateContainerState, isInitialized := containerNameToAggregateStateMap[containerName]
		if !isInitialized {
			aggregateContainerState = vpa_model.NewAggregateContainerState()
			containerNameToAggregateStateMap[containerName] = aggregateContainerState
		}
		aggregateContainerState.MergeContainerState(aggregation)
	}
	return containerNameToAggregateStateMap
}

// ContainerStateAggregatorProxy is a wrapper for ContainerStateAggregator
// that creates ContainerStateAgregator for container if it is no longer
// present in the cluster state.
type ContainerStateAggregatorProxy struct {
	containerID vpa_model.ContainerID
	cluster     *ClusterState
}

// NewContainerStateAggregatorProxy creates a ContainerStateAggregatorProxy
// pointing to the cluster state.
func NewContainerStateAggregatorProxy(cluster *ClusterState, containerID vpa_model.ContainerID) vpa_model.ContainerStateAggregator {
	return &ContainerStateAggregatorProxy{containerID, cluster}
}

// AddSample adds a container sample to the aggregator.
func (p *ContainerStateAggregatorProxy) AddSample(sample *vpa_model.ContainerUsageSample) {
	aggregator := p.cluster.findOrCreateAggregateContainerState(p.containerID)
	aggregator.AddSample(sample)
}

// SubtractSample subtracts a container sample from the aggregator.
func (p *ContainerStateAggregatorProxy) SubtractSample(sample *vpa_model.ContainerUsageSample) {
	aggregator := p.cluster.findOrCreateAggregateContainerState(p.containerID)
	aggregator.SubtractSample(sample)
}

// GetLastRecommendation returns last recorded recommendation.
func (p *ContainerStateAggregatorProxy) GetLastRecommendation() corev1.ResourceList {
	aggregator := p.cluster.findOrCreateAggregateContainerState(p.containerID)
	return aggregator.GetLastRecommendation()
}

// NeedsRecommendation returns true if the aggregator should have recommendation calculated.
func (p *ContainerStateAggregatorProxy) NeedsRecommendation() bool {
	aggregator := p.cluster.findOrCreateAggregateContainerState(p.containerID)
	return aggregator.NeedsRecommendation()
}

// GetUpdateMode returns update mode of VPA controlling the aggregator.
func (p *ContainerStateAggregatorProxy) GetUpdateMode() *vpa_types.UpdateMode {
	aggregator := p.cluster.findOrCreateAggregateContainerState(p.containerID)
	return aggregator.GetUpdateMode()
}

// GetScalingMode returns scaling mode of container represented by the aggregator.
func (p *ContainerStateAggregatorProxy) GetScalingMode() *vpa_types.ContainerScalingMode {
	aggregator := p.cluster.findOrCreateAggregateContainerState(p.containerID)
	return aggregator.GetScalingMode()
}
