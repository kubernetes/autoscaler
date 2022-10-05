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

package routines

import (
	"fmt"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"strings"
)

// KnownPostProcessors represent the names of the known post-processors
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
	names []KnownPostProcessors
}

// Build returns the list of post processors or an error
// implements interface RecommendationPostProcessorsBuilder
func (f *RecommendationPostProcessorFactory) Build() ([]RecommendationPostProcessor, error) {
	var processors []RecommendationPostProcessor
	for _, name := range f.names {
		switch name {
		case Capping:
			processors = append(processors, &cappingPostProcessor{})
		default:
			return nil, fmt.Errorf("unknown Post Processor: %s", name)
		}
	}
	return processors, nil
}

// ParsePostProcessors parses a comma separated list of post processor names, validate it,
// returns the list of post-processors or an error.
// if `endsWithCapping` is set to true, the parse will automatically append the `capping` post-processor
// at the end of the list if it is not already there.
func ParsePostProcessors(postProcessorNames string, endsWithCapping bool) ([]RecommendationPostProcessor, error) {
	var names []KnownPostProcessors
	cappingLast := false
	for _, name := range strings.Split(postProcessorNames, ",") {
		names = append(names, KnownPostProcessors(name))
		cappingLast = name == string(Capping)
	}
	if endsWithCapping && !cappingLast {
		names = append(names, Capping)
	}
	b := RecommendationPostProcessorFactory{names: names}
	return b.Build()
}
