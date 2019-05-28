/*
Copyright 2019 The Kubernetes Authors.

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
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta2"

	"k8s.io/klog"
)

// NewLimitProportionMaintainingRecommendationProcessor constructs new RecommendationsProcessor that adjusts recommendation
// for given pod to maintain proportion to limit and restrictions on limits
func NewLimitProportionMaintainingRecommendationProcessor() RecommendationProcessor {
	return &limitProportionMaintainingRecommendationProcessor{}
}

type limitProportionMaintainingRecommendationProcessor struct {
	limitsRangeCalculator LimitRangeCalculator
}

// Apply returns a recommendation for the given pod, adjusted to obey policy and limits.
func (c *limitProportionMaintainingRecommendationProcessor) Apply(
	podRecommendation *vpa_types.RecommendedPodResources,
	policy *vpa_types.PodResourcePolicy,
	conditions []vpa_types.VerticalPodAutoscalerCondition,
	pod *apiv1.Pod) (*vpa_types.RecommendedPodResources, ContainerToAnnotationsMap, error) {
}
