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

package api

import (
	v1 "k8s.io/api/core/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

// NewSequentialProcessor constructs RecommendationProcessor that will use provided RecommendationProcessor objects
func NewSequentialProcessor(processors []RecommendationProcessor) RecommendationProcessor {
	return &sequentialRecommendationProcessor{processors: processors}
}

type sequentialRecommendationProcessor struct {
	processors []RecommendationProcessor
}

// Apply chains calls to underlying RecommendationProcessors in order provided on object construction
func (p *sequentialRecommendationProcessor) Apply(vpa *vpa_types.VerticalPodAutoscaler,
	pod *v1.Pod) (*vpa_types.RecommendedPodResources, ContainerToAnnotationsMap, error) {

	var recommendation *vpa_types.RecommendedPodResources

	accumulatedContainerToAnnotationsMap := ContainerToAnnotationsMap{}

	for _, processor := range p.processors {
		var (
			err                       error
			containerToAnnotationsMap ContainerToAnnotationsMap
		)
		recommendation, containerToAnnotationsMap, err = processor.Apply(vpa, pod)
		vpa.Status.Recommendation = recommendation

		for container, newAnnotations := range containerToAnnotationsMap {
			annotations, found := accumulatedContainerToAnnotationsMap[container]
			if found {
				accumulatedContainerToAnnotationsMap[container] = append(annotations, newAnnotations...)
			} else {
				accumulatedContainerToAnnotationsMap[container] = newAnnotations
			}
		}

		if err != nil {
			return nil, accumulatedContainerToAnnotationsMap, err
		}
	}
	return recommendation, accumulatedContainerToAnnotationsMap, nil
}
