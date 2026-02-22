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
	"k8s.io/klog/v2"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/limits"
	vpa_utils "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

type cappingPostProcessor struct {
	globalMaxAllowed limits.GlobalMaxAllowed
}

var _ RecommendationPostProcessor = &cappingPostProcessor{}

// NewCappingRecommendationProcessor constructs new RecommendationPostProcessor that adjusts recommendation
// for given pod to obey VPA resources policy and a global max allowed configuration.
func NewCappingRecommendationProcessor(globalMaxAllowed limits.GlobalMaxAllowed) RecommendationPostProcessor {
	return &cappingPostProcessor{
		globalMaxAllowed: globalMaxAllowed,
	}
}

// Process apply the capping post-processing to the recommendation. (use to be function getCappedRecommendation)
func (c cappingPostProcessor) Process(vpa *vpa_types.VerticalPodAutoscaler, recommendation *vpa_types.RecommendedPodResources) *vpa_types.RecommendedPodResources {
	// TODO: maybe rename the vpa_utils.ApplyVPAPolicy to something that mention that it is doing capping only
	cappedRecommendation, err := vpa_utils.ApplyVPAPolicy(recommendation, vpa.Spec.ResourcePolicy, c.globalMaxAllowed)
	if err != nil {
		klog.ErrorS(err, "Failed to apply policy for VPA", "vpa", klog.KObj(vpa))
		return recommendation
	}
	return cappedRecommendation
}

// ProcessPodLevel applies capping to Pod-level recommendations.
func (c cappingPostProcessor) ProcessPodLevel(vpa *vpa_types.VerticalPodAutoscaler, recommendation *vpa_types.RecommendedPodResources) *vpa_types.RecommendedPodResources {
	cappedRecommendation, err := vpa_utils.ApplyRecommenderLevelPolicies(recommendation, vpa.Spec.ResourcePolicy, c.globalMaxAllowed)
	if err != nil {
		klog.ErrorS(err, "Failed to apply Pod-level policy for VPA", "vpa", klog.KObj(vpa))
		return recommendation
	}
	return cappedRecommendation
}
