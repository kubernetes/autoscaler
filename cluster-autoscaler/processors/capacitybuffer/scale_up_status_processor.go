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
	"strings"

	apiv1 "k8s.io/api/core/v1"
	v1alpha1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
)

// FakePodsScaleUpStatusProcessor is a ScaleUpStatusProcessor used for filtering out fake pods from scaleup status.
type FakePodsScaleUpStatusProcessor struct {
	buffersRegistry *capacityBuffersFakePodsRegistry
}

type bufferInfo struct {
	buffer         *v1alpha1.CapacityBuffer
	numberOfPods   int
	reasonMessages []string
}

// NewFakePodsScaleUpStatusProcessor return an instance of FakePodsScaleUpStatusProcessor
func NewFakePodsScaleUpStatusProcessor(buffersRegistry *capacityBuffersFakePodsRegistry) *FakePodsScaleUpStatusProcessor {
	return &FakePodsScaleUpStatusProcessor{buffersRegistry: buffersRegistry}
}

// Process updates scaleupStatus to remove all capacity buffer fake pods from
// PodsRemainUnschedulable, PodsAwaitEvaluation & PodsTriggeredScaleup
func (p *FakePodsScaleUpStatusProcessor) Process(context *ca_context.AutoscalingContext, scaleUpStatus *status.ScaleUpStatus) {
	var fakePodsRemainUnschedulable []status.NoScaleUpInfo
	var fakePodsTriggeredScaleUp []*apiv1.Pod

	scaleUpStatus.PodsRemainUnschedulable, fakePodsRemainUnschedulable = filterOutCapacityBuffersPod(scaleUpStatus.PodsRemainUnschedulable, func(noScaleUpInfo status.NoScaleUpInfo) *apiv1.Pod { return noScaleUpInfo.Pod })
	scaleUpStatus.PodsTriggeredScaleUp, fakePodsTriggeredScaleUp = filterOutCapacityBuffersPod(scaleUpStatus.PodsTriggeredScaleUp, func(pod *apiv1.Pod) *apiv1.Pod { return pod })
	scaleUpStatus.PodsAwaitEvaluation, _ = filterOutCapacityBuffersPod(scaleUpStatus.PodsAwaitEvaluation, func(pod *apiv1.Pod) *apiv1.Pod { return pod })

	p.createBuffersNoScaleUpEvents(context, scaleUpStatus, fakePodsRemainUnschedulable)
	p.createBuffersScaleUpEvents(context, scaleUpStatus, fakePodsTriggeredScaleUp)
}

func (p *FakePodsScaleUpStatusProcessor) createBuffersNoScaleUpEvents(context *ca_context.AutoscalingContext, scaleUpStatus *status.ScaleUpStatus, fakePodsRemainUnschedulable []status.NoScaleUpInfo) {
	if scaleUpStatus.Result != status.ScaleUpSuccessful && scaleUpStatus.Result != status.ScaleUpError {
		consideredNodeGroupsMap := cloudprovider.NodeGroupListToMapById(scaleUpStatus.ConsideredNodeGroups)
		buffersInfo := map[string]*bufferInfo{}
		for _, noScaleUpInfo := range fakePodsRemainUnschedulable {
			parentCapacityBuffer, found := p.buffersRegistry.fakePodsUIDToBuffer[string(noScaleUpInfo.Pod.UID)]
			if found {
				bufferUID := string(parentCapacityBuffer.UID)
				if _, found := buffersInfo[bufferUID]; !found {
					buffersInfo[bufferUID] = &bufferInfo{
						buffer: parentCapacityBuffer,
					}
				}
				buffersInfo[bufferUID].numberOfPods += 1
				buffersInfo[bufferUID].reasonMessages = append(buffersInfo[bufferUID].reasonMessages,
					status.ReasonsMessage(scaleUpStatus.Result, noScaleUpInfo, consideredNodeGroupsMap))
			}
		}

		for _, bufferInfo := range buffersInfo {
			context.Recorder.Eventf(bufferInfo.buffer, apiv1.EventTypeNormal, "NotTriggerScaleUp",
				"capacity buffer %d fake pods didn't trigger scale-up: %s",
				bufferInfo.numberOfPods, strings.Join(bufferInfo.reasonMessages, ","))
		}
	}
}

func (p *FakePodsScaleUpStatusProcessor) createBuffersScaleUpEvents(context *ca_context.AutoscalingContext, scaleUpStatus *status.ScaleUpStatus, fakePodsTriggeredScaleUp []*apiv1.Pod) {
	if len(scaleUpStatus.ScaleUpInfos) > 0 && len(fakePodsTriggeredScaleUp) > 0 {
		buffersInfo := map[string]*bufferInfo{}
		for _, pod := range fakePodsTriggeredScaleUp {
			parentCapacityBuffer, found := p.buffersRegistry.fakePodsUIDToBuffer[string(pod.UID)]
			if found {
				bufferUID := string(parentCapacityBuffer.UID)
				if _, found := buffersInfo[bufferUID]; !found {
					buffersInfo[bufferUID] = &bufferInfo{
						buffer: parentCapacityBuffer,
					}
				}
				buffersInfo[bufferUID].numberOfPods += 1
			}
		}
		for _, bufferInfo := range buffersInfo {
			context.Recorder.Eventf(bufferInfo.buffer, apiv1.EventTypeNormal, "TriggeredScaleUp",
				"capacity buffer %d fake pods triggered scale-up: %v", bufferInfo.numberOfPods, scaleUpStatus.ScaleUpInfos)
		}
	}
}

// filterOutCapacityBuffersPod filters out the fake pods created for capcity biffers
// from the input list of T using passed getPod(T) and returns a the filtered and the filtered out lists
func filterOutCapacityBuffersPod[T any](podsWrappers []T, getPod func(T) *apiv1.Pod) ([]T, []T) {
	filteredPodsSources := make([]T, 0)
	filteredOutPodsSources := make([]T, 0)
	for _, podsWrapper := range podsWrappers {
		currentPod := getPod(podsWrapper)
		if IsFakeCapacityBuffersPod(currentPod) {
			filteredOutPodsSources = append(filteredOutPodsSources, podsWrapper)
		} else {
			filteredPodsSources = append(filteredPodsSources, podsWrapper)
		}
	}
	return filteredPodsSources, filteredOutPodsSources
}

// CleanUp is called at CA termination
func (p *FakePodsScaleUpStatusProcessor) CleanUp() {}
