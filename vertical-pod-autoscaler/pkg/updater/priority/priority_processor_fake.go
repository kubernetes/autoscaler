/*
Copyright 2020 The Kubernetes Authors.

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

package priority

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

type fakePriorityProcessor struct {
	priorities map[string]PodPriority
}

// NewFakeProcessor returns a fake processor for testing that can be initialized
// with a map from pod name to priority expected to be returned.
func NewFakeProcessor(priorities map[string]PodPriority) PriorityProcessor {
	return &fakePriorityProcessor{
		priorities: priorities,
	}
}

func (f *fakePriorityProcessor) GetUpdatePriority(pod *corev1.Pod, vpa *vpa_types.VerticalPodAutoscaler,
	recommendation *vpa_types.RecommendedPodResources) PodPriority {
	prio, ok := f.priorities[pod.Name]
	if !ok {
		panic(fmt.Sprintf("Unexpected pod name: %v", pod.Name))
	}
	return PodPriority{
		ScaleUp:                 prio.ScaleUp,
		ResourceDiff:            prio.ResourceDiff,
		OutsideRecommendedRange: prio.OutsideRecommendedRange,
	}
}
