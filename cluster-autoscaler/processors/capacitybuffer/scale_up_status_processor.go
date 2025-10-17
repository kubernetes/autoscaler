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
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	v1alpha1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
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

	scaleUpStatus.PodsRemainUnschedulable, fakePodsRemainUnschedulable = filterOutFakeUnschedulablePods(scaleUpStatus.PodsRemainUnschedulable)
	scaleUpStatus.PodsTriggeredScaleUp, fakePodsTriggeredScaleUp = filterOutFakePods(scaleUpStatus.PodsTriggeredScaleUp)
	scaleUpStatus.PodsAwaitEvaluation, _ = filterOutFakePods(scaleUpStatus.PodsAwaitEvaluation)

	p.createEventsForBuffersWithUnschedulablePods(context, scaleUpStatus, fakePodsRemainUnschedulable)
	p.createEventsForBuffersWithPodsTriggeredScaleUp(context, scaleUpStatus, fakePodsTriggeredScaleUp)
}

func (p *FakePodsScaleUpStatusProcessor) createEventsForBuffersWithUnschedulablePods(context *ca_context.AutoscalingContext, scaleUpStatus *status.ScaleUpStatus, fakePodsRemainUnschedulable []status.NoScaleUpInfo) {
	if scaleUpStatus.Result != status.ScaleUpSuccessful && scaleUpStatus.Result != status.ScaleUpError {
		consideredNodeGroupsMap := status.NodeGroupListToMapById(scaleUpStatus.ConsideredNodeGroups)
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
			context.Recorder.Event(bufferInfo.buffer, apiv1.EventTypeNormal, "NotTriggerScaleUp",
				fmt.Sprintf("capacity buffer %d fake pods didn't trigger scale-up: %s",
					bufferInfo.numberOfPods, strings.Join(bufferInfo.reasonMessages, ",")))
		}
	}
}

func (p *FakePodsScaleUpStatusProcessor) createEventsForBuffersWithPodsTriggeredScaleUp(context *ca_context.AutoscalingContext, scaleUpStatus *status.ScaleUpStatus, fakePodsTriggeredScaleUp []*apiv1.Pod) {
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

// filterOutFakeUnschedulablePods filters out NoScaleUpInfo for capacity buffers fake pods
func filterOutFakeUnschedulablePods(noScaleUpInfo []status.NoScaleUpInfo) ([]status.NoScaleUpInfo, []status.NoScaleUpInfo) {
	filteredInfo := make([]status.NoScaleUpInfo, 0)
	filteredOutInfo := make([]status.NoScaleUpInfo, 0)

	for _, info := range noScaleUpInfo {
		currentPod := info.Pod
		if !isFakeCapacityBuffersPod(currentPod) {
			filteredInfo = append(filteredInfo, info)
			continue
		}
		filteredOutInfo = append(filteredOutInfo, info)
	}
	return filteredInfo, filteredOutInfo
}

// filterFakePods filters out capacity buffer fake pods
func filterOutFakePods(pods []*apiv1.Pod) ([]*apiv1.Pod, []*apiv1.Pod) {
	filteredPods := make([]*apiv1.Pod, 0)
	filteredOutPods := make([]*apiv1.Pod, 0)

	for _, pod := range pods {
		if !isFakeCapacityBuffersPod(pod) {
			filteredPods = append(filteredPods, pod)
			continue
		}
		filteredOutPods = append(filteredOutPods, pod)
	}
	return filteredPods, filteredOutPods
}

// CleanUp is called at CA termination
func (p *FakePodsScaleUpStatusProcessor) CleanUp() {}
