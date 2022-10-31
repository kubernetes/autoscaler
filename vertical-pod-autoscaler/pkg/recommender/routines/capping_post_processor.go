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
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	vpa_utils "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
	"k8s.io/klog/v2"
)

// CappingPostProcessor ensure that the policy is applied to recommendation
// it applies policy for fields: MinAllowed and MaxAllowed
type CappingPostProcessor struct{}

var _ RecommendationPostProcessor = &CappingPostProcessor{}

// Process apply the capping post-processing to the recommendation. (use to be function getCappedRecommendation)
func (c CappingPostProcessor) Process(vpa *model.Vpa, recommendation *vpa_types.RecommendedPodResources, policy *vpa_types.PodResourcePolicy) *vpa_types.RecommendedPodResources {
	// TODO: maybe rename the vpa_utils.ApplyVPAPolicy to something that mention that it is doing capping only
	cappedRecommendation, err := vpa_utils.ApplyVPAPolicy(recommendation, policy)
	if err != nil {
		klog.Errorf("Failed to apply policy for VPA %v/%v: %v", vpa.ID.Namespace, vpa.ID.VpaName, err)
		return recommendation
	}
	return cappedRecommendation
}
