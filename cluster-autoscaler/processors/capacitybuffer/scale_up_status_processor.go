/*
Copyright 2025 The Kubernetes Authors.

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

package capacitybufferpodlister

import (
	apiv1 "k8s.io/api/core/v1"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/klog/v2"
)

// FakePodsScaleUpStatusProcessor is a ScaleUpStatusProcessor used for filtering out fake pods from scaleup status.
type FakePodsScaleUpStatusProcessor struct{}

// NewFakePodsScaleUpStatusProcessor return an instance of FakePodsScaleUpStatusProcessor
func NewFakePodsScaleUpStatusProcessor() *FakePodsScaleUpStatusProcessor {
	return &FakePodsScaleUpStatusProcessor{}
}

// Process updates scaleupStatus to remove all capacity buffer fake pods from
// PodsRemainUnschedulable, PodsAwaitEvaluation & PodsTriggeredScaleup
func (a *FakePodsScaleUpStatusProcessor) Process(_ *ca_context.AutoscalingContext, scaleUpStatus *status.ScaleUpStatus) {
	scaleUpStatus.PodsRemainUnschedulable = filterFakePods(scaleUpStatus.PodsRemainUnschedulable, func(noScaleUpInfo status.NoScaleUpInfo) *apiv1.Pod { return noScaleUpInfo.Pod }, "PodsRemainUnschedulable")
	scaleUpStatus.PodsAwaitEvaluation = filterFakePods(scaleUpStatus.PodsAwaitEvaluation, func(pod *apiv1.Pod) *apiv1.Pod { return pod }, "PodsAwaitEvaluation")
	scaleUpStatus.PodsTriggeredScaleUp = filterFakePods(scaleUpStatus.PodsTriggeredScaleUp, func(pod *apiv1.Pod) *apiv1.Pod { return pod }, "PodsTriggeredScaleUp")
}

// filterFakePods removes capacity buffer fake pods from the input list of T using passed getPod(T)
// Returns a list containing only non-fake pods
func filterFakePods[T any](podsWrappers []T, getPod func(T) *apiv1.Pod, resourceName string) []T {
	filteredPodsSources := make([]T, 0)
	removedPods := make([]*apiv1.Pod, 0)

	for _, podsWrapper := range podsWrappers {
		currentPod := getPod(podsWrapper)
		if !isFakeCapacityBuffersPod(currentPod) {
			filteredPodsSources = append(filteredPodsSources, podsWrapper)
			continue
		}
		removedPods = append(removedPods, currentPod)
	}

	klog.Infof("Filtered out %d pods from %s", len(removedPods), resourceName)
	return filteredPodsSources
}

// CleanUp is called at CA termination
func (a *FakePodsScaleUpStatusProcessor) CleanUp() {}
