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

package logic

import (
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/model"
)

// PodResourceRecommender computes resource recommendation for a Vpa object.
type PodResourceRecommender interface {
	GetRecommendedPodResources(vpa *model.Vpa) RecommendedPodResources
}

// RecommendedPodResources is a Map from container name to recommended resources.
type RecommendedPodResources map[string]RecommendedContainerResources

// RecommendedContainerResources is the recommendation of resources for a
// container.
type RecommendedContainerResources struct {
	// Recommended optimal amount of resources.
	Target model.Resources
	// Recommended minimum amount of resources.
	MinRecommended model.Resources
	// Recommended maximum amount of resources.
	MaxRecommended model.Resources
}

type podResourceRecommender struct {
	targetEstimator     ResourceEstimator
	lowerBoundEstimator ResourceEstimator
	upperBoundEstimator ResourceEstimator
}

// NewPodResourceRecommender returns a new PodResourceRecommender which is an
// aggregation of three ResourceEstimators, one for each of the target, min
// and max recommended resources.
func NewPodResourceRecommender(
	targetEstimator ResourceEstimator,
	lowerBoundEstimator ResourceEstimator,
	upperBoundEstimator ResourceEstimator) PodResourceRecommender {
	return &podResourceRecommender{targetEstimator, lowerBoundEstimator, upperBoundEstimator}
}

// Returns recommended resources for a given Vpa object.
func (r *podResourceRecommender) GetRecommendedPodResources(vpa *model.Vpa) RecommendedPodResources {
	aggregateContainerStateMap := model.BuildAggregateContainerStateMap(vpa, model.MergeForRecommendation, time.Unix(0, 0))
	var recommendation = make(RecommendedPodResources)
	for containerName, aggregatedContainerState := range aggregateContainerStateMap {
		if aggregatedContainerState.TotalSamplesCount > 0 {
			recommendation[containerName] = r.getRecommendedContainerResources(aggregatedContainerState)
		}
	}
	return recommendation
}

// Takes AggregateContainerState and returns a container recommendation.
func (r *podResourceRecommender) getRecommendedContainerResources(s *model.AggregateContainerState) RecommendedContainerResources {
	return RecommendedContainerResources{
		r.targetEstimator.GetResourceEstimation(s),
		r.lowerBoundEstimator.GetResourceEstimation(s),
		r.upperBoundEstimator.GetResourceEstimation(s),
	}
}
