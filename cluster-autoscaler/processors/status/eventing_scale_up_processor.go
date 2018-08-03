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
	for _, pod := range status.PodsRemainUnschedulable {
		context.Recorder.Event(pod, apiv1.EventTypeNormal, "NotTriggerScaleUp",
			"pod didn't trigger scale-up (it wouldn't fit if a new node is added)")
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
