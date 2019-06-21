/*
Copyright 2018 The Kubernetes Authors.

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

package status

import (
	"fmt"
	"k8s.io/klog"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
)

// EventingScaleUpStatusProcessor processes the state of the cluster after
// a scale-up by emitting relevant events for pods depending on their post
// scale-up status.
type EventingScaleUpStatusProcessor struct{}

// Process processes the state of the cluster after a scale-up by emitting
// relevant events for pods depending on their post scale-up status.
func (p *EventingScaleUpStatusProcessor) Process(context *context.AutoscalingContext, status *ScaleUpStatus) {
	if status.Result != ScaleUpSuccessful && status.Result != ScaleUpError {
		for _, noScaleUpInfo := range status.PodsRemainUnschedulable {
			context.Recorder.Event(noScaleUpInfo.Pod, apiv1.EventTypeNormal, "NotTriggerScaleUp",
				fmt.Sprintf("pod didn't trigger scale-up (it wouldn't fit if a new node is"+
					" added): %s", ReasonsMessage(noScaleUpInfo)))
		}
	} else {
		klog.V(4).Infof("Skipping event processing for unschedulable pods since there is a" +
			" ScaleUp attempt this loop")
	}
	if len(status.ScaleUpInfos) > 0 {
		for _, pod := range status.PodsTriggeredScaleUp {
			context.Recorder.Eventf(pod, apiv1.EventTypeNormal, "TriggeredScaleUp",
				"pod triggered scale-up: %v", status.ScaleUpInfos)
		}
	}
}

// CleanUp cleans up the processor's internal structures.
func (p *EventingScaleUpStatusProcessor) CleanUp() {
}

// ReasonsMessage aggregates reasons from NoScaleUpInfos.
func ReasonsMessage(noScaleUpInfo NoScaleUpInfo) string {
	messages := []string{}
	aggregated := map[string]int{}
	for _, reasons := range noScaleUpInfo.RejectedNodeGroups {
		for _, reason := range reasons.Reasons() {
			aggregated[reason]++
		}
	}
	for _, reasons := range noScaleUpInfo.SkippedNodeGroups {
		for _, reason := range reasons.Reasons() {
			aggregated[reason]++
		}
	}
	for msg, count := range aggregated {
		messages = append(messages, fmt.Sprintf("%d %s", count, msg))
	}
	return strings.Join(messages, ", ")
}
