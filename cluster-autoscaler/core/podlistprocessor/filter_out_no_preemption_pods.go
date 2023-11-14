/*
Copyright 2023 The Kubernetes Authors.

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

package podlistprocessor

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/pods"
)

type filterOutNoPreemptionPodsListProcessor struct {
	scheduledPods []*apiv1.Pod
}

type ScheduledPodsNotifier interface {
	Update(scheduledPods []*apiv1.Pod)
	Register(pods.PodListProcessor)
}

type scheduledPodsNotifier struct {
	// CONTINUE HERE
	processors []pods.PodListProcessor
}

func NewScheduledPodsNotifier() *scheduledPodsNotifier {
	return &scheduledPodsNotifier{}
}

func (s *scheduledPodsNotifier) Update(scheduledPods []*apiv1.Pod) {
	s.Update(scheduledPods)
}

func NewFilterOutNoPreemptionPodsListProcessor() *filterOutNoPreemptionPodsListProcessor {
	return &filterOutNoPreemptionPodsListProcessor{}
}

func (p *filterOutNoPreemptionPodsListProcessor) Process(
	context *context.AutoscalingContext,
	unschedulablePods []*apiv1.Pod) ([]*apiv1.Pod, error) {
	return []*apiv1.Pod{}, nil
}

func (p *filterOutNoPreemptionPodsListProcessor) Update(
	scheduledPods []*apiv1.Pod) {
	p.scheduledPods = scheduledPods
}

func (p *filterOutNoPreemptionPodsListProcessor) CleanUp() {

}
