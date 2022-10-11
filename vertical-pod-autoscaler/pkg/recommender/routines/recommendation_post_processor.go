/*
Copyright 2022 The Kubernetes Authors.

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
	"fmt"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"k8s.io/klog/v2"
)

// KnownPostProcessors represent the PostProcessorsNames of the known post-processors
type KnownPostProcessors string

const (
	// Capping is post-processor name to ensure that recommendation stays within [MinAllowed-MaxAllowed] range
	Capping KnownPostProcessors = "capping"
)

// RecommendationPostProcessor can amend the recommendation according to the defined policies
type RecommendationPostProcessor interface {
	Process(vpa *model.Vpa, recommendation *vpa_types.RecommendedPodResources,
		policy *vpa_types.PodResourcePolicy) *vpa_types.RecommendedPodResources
}

// RecommendationPostProcessorsBuilder helps for the creation of the pre-processors list
type RecommendationPostProcessorsBuilder interface {
	// Build returns the list of post processors or an error
	Build() ([]RecommendationPostProcessor, error)
}

// RecommendationPostProcessorFactory helps to build processors by their name
// using a struct and not a simple function to hold parameters that might be necessary to create some post processors in the future
type RecommendationPostProcessorFactory struct {
	PostProcessorsNames []KnownPostProcessors
}

// Build returns the list of post processors or an error
// implements interface RecommendationPostProcessorsBuilder
func (f *RecommendationPostProcessorFactory) Build() ([]RecommendationPostProcessor, error) {
	var processors []RecommendationPostProcessor
	for _, name := range f.PostProcessorsNames {
		switch name {
		case Capping:
			processors = append(processors, &cappingPostProcessor{})
		default:
			return nil, fmt.Errorf("unknown Post Processor: %s", name)
		}
	}
	klog.Infof("List of recommendation post-processors: %v", processors)
	return processors, nil
}
