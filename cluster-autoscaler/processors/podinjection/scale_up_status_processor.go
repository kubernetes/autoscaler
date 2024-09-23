/*
Copyright 2024 The Kubernetes Authors.

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

package podinjection

import (
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	podinjectionbackoff "k8s.io/autoscaler/cluster-autoscaler/processors/podinjection/backoff"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/klog/v2"
)

// FakePodsScaleUpStatusProcessor is a ScaleUpStatusProcessor used for filtering out fake pods from scaleup status.
type FakePodsScaleUpStatusProcessor struct {
	fakePodControllerBackoffRegistry *podinjectionbackoff.ControllerRegistry
}

// NewFakePodsScaleUpStatusProcessor return an instance of FakePodsScaleUpStatusProcessor
func NewFakePodsScaleUpStatusProcessor(fakePodRegistry *podinjectionbackoff.ControllerRegistry) *FakePodsScaleUpStatusProcessor {
	return &FakePodsScaleUpStatusProcessor{
		fakePodControllerBackoffRegistry: fakePodRegistry,
	}
}

// Process updates scaleupStatus to remove all fake pods from
// PodsRemainUnschedulable, PodsAwaitEvaluation & PodsTriggeredScaleup
func (a *FakePodsScaleUpStatusProcessor) Process(_ *ca_context.AutoscalingContext, scaleUpStatus *status.ScaleUpStatus) {
	controllersToBackoff := extractFakePodsControllersUIDs(scaleUpStatus.PodsRemainUnschedulable)
	for uid := range controllersToBackoff {
		a.fakePodControllerBackoffRegistry.BackoffController(uid, time.Now())
	}

	scaleUpStatus.PodsRemainUnschedulable = filterFakePods(scaleUpStatus.PodsRemainUnschedulable, func(noScaleUpInfo status.NoScaleUpInfo) *apiv1.Pod { return noScaleUpInfo.Pod }, "PodsRemainUnschedulable")
	scaleUpStatus.PodsAwaitEvaluation = filterFakePods(scaleUpStatus.PodsAwaitEvaluation, func(pod *apiv1.Pod) *apiv1.Pod { return pod }, "PodsAwaitEvaluation")
	scaleUpStatus.PodsTriggeredScaleUp = filterFakePods(scaleUpStatus.PodsTriggeredScaleUp, func(pod *apiv1.Pod) *apiv1.Pod { return pod }, "PodsTriggeredScaleUp")
}

// filterFakePods removes fake pods from the input list of T using passed getPod(T)
// Uses `resourceName` to log which resource it has modified
// Returns a list containing only non-fake pods
func filterFakePods[T any](podsWrappers []T, getPod func(T) *apiv1.Pod, resourceName string) []T {
	filteredPodsSouces := make([]T, 0)
	removedPods := make([]*apiv1.Pod, 0)

	for _, podsWrapper := range podsWrappers {
		currentPod := getPod(podsWrapper)
		if !IsFake(currentPod) {
			filteredPodsSouces = append(filteredPodsSouces, podsWrapper)
			continue
		}

		controllerRef := v1.GetControllerOf(currentPod)
		if controllerRef == nil {
			klog.Infof("Failed to find controller for pod %s, ignoring.", currentPod.Name)
			continue
		}

		removedPods = append(removedPods, currentPod)
		klog.V(5).Infof("Filtering out pod %s from PodsRemainUnschedulable with controller reference %s", currentPod.Name, controllerRef.Name)
	}

	logRemovedPods(removedPods, resourceName)
	return filteredPodsSouces
}

// extractFakePodsControllersUIDs extracts the uids from NoScaleUpInfos with fake pods
func extractFakePodsControllersUIDs(NoScaleUpInfos []status.NoScaleUpInfo) map[types.UID]bool {
	uids := make(map[types.UID]bool)
	for _, NoScaleUpInfo := range NoScaleUpInfos {
		if IsFake(NoScaleUpInfo.Pod) {
			uids[NoScaleUpInfo.Pod.UID] = true
		}
	}
	return uids
}

// logRemovedPods logs the removed pods from resourceName
func logRemovedPods(removedPods []*apiv1.Pod, resourceName string) {
	if len(removedPods) == 0 {
		return
	}
	controllerRefNames := make([]string, len(removedPods))
	for idx, pod := range removedPods {
		controllerRefNames[idx] = v1.GetControllerOf(pod).Name
	}
	klog.Infof("Filtered out %d pods from %s for controllers %s", len(removedPods), resourceName, strings.Join(controllerRefNames, ", "))
}

// CleanUp is called at CA termination
func (a *FakePodsScaleUpStatusProcessor) CleanUp() {}
