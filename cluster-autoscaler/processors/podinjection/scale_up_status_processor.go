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

func NewFakePodsScaleUpStatusProcessor(fakePodRegistry *podinjectionbackoff.ControllerRegistry) *FakePodsScaleUpStatusProcessor {
	return &FakePodsScaleUpStatusProcessor{
		fakePodControllerBackoffRegistry: fakePodRegistry,
	}
}

// Process updates scaleupStatus to remove all fake pods from
// PodsRemainUnschedulable, PodsAwaitEvaluation & PodsTriggeredScaleup
func (a *FakePodsScaleUpStatusProcessor) Process(_ *ca_context.AutoscalingContext, scaleUpStatus *status.ScaleUpStatus) {
	var controllersToBackoff []types.UID
	scaleUpStatus.PodsRemainUnschedulable, controllersToBackoff = filterNoScaleUpInfos(scaleUpStatus.PodsRemainUnschedulable)
	scaleUpStatus.PodsAwaitEvaluation = filterFakePods(scaleUpStatus.PodsAwaitEvaluation, "PodsAwaitEvaluation")
	scaleUpStatus.PodsTriggeredScaleUp = filterFakePods(scaleUpStatus.PodsTriggeredScaleUp, "PodsTriggeredScaleUp")

	for _, uid := range controllersToBackoff {
		a.fakePodControllerBackoffRegistry.BackoffController(uid, time.Now())
	}

}

// Returns pod's controller name if found
// Otherwise, returns an empty string
func extractUIDsAndControllerRefNames(controllerRefs map[types.UID]*v1.OwnerReference) ([]types.UID, []string) {
	uids := make([]types.UID, len(controllerRefs))
	controllerRefNames := make([]string, len(controllerRefs))
	idx := 0
	for uid, controllerRef := range controllerRefs {
		uids[idx] = uid
		controllerRefNames[idx] = controllerRef.Name
		idx++
	}
	return uids, controllerRefNames
}

// Removes NoScaleUpInfo entries which contain a fake pod from the input list
// Returns a list of NoScaleUpInfo which doesn't contain any fake pods
func filterNoScaleUpInfos(noScaleUpInfos []status.NoScaleUpInfo) ([]status.NoScaleUpInfo, []types.UID) {
	filteredNoScaleUpInfos := make([]status.NoScaleUpInfo, 0)
	filteredPodsCount := 0
	controllerRefs := make(map[types.UID]*v1.OwnerReference, 0)
	for _, noScaleUpInfo := range noScaleUpInfos {
		if !IsFake(noScaleUpInfo.Pod) {
			filteredNoScaleUpInfos = append(filteredNoScaleUpInfos, noScaleUpInfo)
			continue
		}
		controllerRef := v1.GetControllerOf(noScaleUpInfo.Pod)
		if controllerRef == nil {
			klog.Infof("Failed to find controller for pod %s, ignoring.", noScaleUpInfo.Pod.Name)
			continue
		}
		controllerRefs[controllerRef.UID] = controllerRef
		klog.V(5).Infof("Filtering out pod %s from PodsRemainUnschedulable with controller reference %s", noScaleUpInfo.Pod.Name, controllerRef.Name)
		filteredPodsCount++
	}
	controllerUIDs, controllerRefNames := extractUIDsAndControllerRefNames(controllerRefs)
	if filteredPodsCount > 0 {
		klog.Infof("Filtered out %d pods from PodsRemainUnschedulable for controllers %s", filteredPodsCount, strings.Join(controllerRefNames, ", "))
	}
	return filteredNoScaleUpInfos, controllerUIDs
}

// Removes fake pods from the input list
// Uses `resourceName` to log which resource it has modified
// Returns a list containing only non-fake pods
func filterFakePods(pods []*apiv1.Pod, resourceName string) []*apiv1.Pod {
	filteredPods := make([]*apiv1.Pod, 0)
	removedPods := make([]*apiv1.Pod, 0)
	controllerRefs := make(map[types.UID]*v1.OwnerReference, 0)
	for _, pod := range pods {
		if IsFake(pod) {
			controllerRef := v1.GetControllerOf(pod)
			if controllerRef == nil {
				klog.Infof("Failed to find controller for pod %s, ignoring.", pod.Name)
				continue
			}
			controllerRefs[controllerRef.UID] = controllerRef
			klog.V(5).Infof("Filtering out pod %s with controller reference %s", pod.Name, controllerRef)
			removedPods = append(removedPods, pod)
			continue
		}
		filteredPods = append(filteredPods, pod)
	}
	if len(removedPods) > 0 {
		_, controllerRefNames := extractUIDsAndControllerRefNames(controllerRefs)
		klog.Infof("Filtered out %d pods from %s for controllers %s", len(removedPods), resourceName, strings.Join(controllerRefNames, ", "))
	}
	return filteredPods
}

func (_ *FakePodsScaleUpStatusProcessor) CleanUp() {}
