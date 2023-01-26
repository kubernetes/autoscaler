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
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	"strings"
)

// IntegerCPUPostProcessor ensures that the recommendation delivers an integer value for CPU
// This is need for users who want to use CPU Management with static policy: https://kubernetes.io/docs/tasks/administer-cluster/cpu-management-policies/#static-policy
type IntegerCPUPostProcessor struct{}

const (
	// The user interface for that post processor is an annotation on the VPA object with the following format:
	// vpa-post-processor.kubernetes.io/{containerName}_integerCPU=true
	vpaPostProcessorPrefix           = "vpa-post-processor.kubernetes.io/"
	vpaPostProcessorIntegerCPUSuffix = "_integerCPU"
	vpaPostProcessorIntegerCPUValue  = "true"
)

var _ RecommendationPostProcessor = &IntegerCPUPostProcessor{}

// Process apply the capping post-processing to the recommendation.
// For this post processor the CPU value is rounded up to an integer
func (p *IntegerCPUPostProcessor) Process(vpa *model.Vpa, recommendation *vpa_types.RecommendedPodResources, policy *vpa_types.PodResourcePolicy) *vpa_types.RecommendedPodResources {

	amendedRecommendation := recommendation.DeepCopy()

	for key, value := range vpa.Annotations {
		containerName := extractContainerName(key, vpaPostProcessorPrefix, vpaPostProcessorIntegerCPUSuffix)
		if containerName == "" || value != vpaPostProcessorIntegerCPUValue {
			continue
		}

		for _, r := range amendedRecommendation.ContainerRecommendations {
			if r.ContainerName != containerName {
				continue
			}
			setIntegerCPURecommendation(r.Target)
			setIntegerCPURecommendation(r.LowerBound)
			setIntegerCPURecommendation(r.UpperBound)
			setIntegerCPURecommendation(r.UncappedTarget)
		}
	}
	return amendedRecommendation
}

func setIntegerCPURecommendation(recommendation apiv1.ResourceList) {
	for resourceName, recommended := range recommendation {
		if resourceName != apiv1.ResourceCPU {
			continue
		}
		recommended.RoundUp(resource.Scale(0))
		recommendation[resourceName] = recommended
	}
}

// extractContainerName return the container name for the feature based on annotation key
// if the returned value is empty that means that the key does not match
func extractContainerName(key, prefix, suffix string) string {
	if !strings.HasPrefix(key, prefix) {
		return ""
	}
	if !strings.HasSuffix(key, suffix) {
		return ""
	}

	return key[len(prefix) : len(key)-len(suffix)]
}
