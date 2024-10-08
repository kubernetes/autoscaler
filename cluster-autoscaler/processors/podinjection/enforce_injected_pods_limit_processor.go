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
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
)

const (
	// InjectedMetricsLabel is the label for unschedulable pods metric for injected pods.
	InjectedMetricsLabel = "injected"
	// SkippedInjectionMetricsLabel is the label for unschedulable pods metric for pods that was not injected due to limit.
	SkippedInjectionMetricsLabel = "skipped_injection"
)

// EnforceInjectedPodsLimitProcessor is a PodListProcessor used to limit the number of injected fake pods.
type EnforceInjectedPodsLimitProcessor struct {
	podLimit int
}

// NewEnforceInjectedPodsLimitProcessor return an instance of EnforceInjectedPodsLimitProcessor
func NewEnforceInjectedPodsLimitProcessor(podLimit int) *EnforceInjectedPodsLimitProcessor {
	return &EnforceInjectedPodsLimitProcessor{
		podLimit: podLimit,
	}
}

// Process filters unschedulablePods and enforces the limit of the number of injected pods
func (p *EnforceInjectedPodsLimitProcessor) Process(ctx *context.AutoscalingContext, unschedulablePods []*apiv1.Pod) ([]*apiv1.Pod, error) {

	numberOfFakePodsToRemove := len(unschedulablePods) - p.podLimit
	removedFakePodsCount := 0
	injectedFakePodsCount := 0
	var unschedulablePodsAfterProcessing []*apiv1.Pod

	for _, pod := range unschedulablePods {
		if IsFake(pod) {
			if removedFakePodsCount < numberOfFakePodsToRemove {
				removedFakePodsCount += 1
				continue
			}
			injectedFakePodsCount += 1
		}

		unschedulablePodsAfterProcessing = append(unschedulablePodsAfterProcessing, pod)
	}

	metrics.UpdateUnschedulablePodsCountWithLabel(injectedFakePodsCount, InjectedMetricsLabel)
	metrics.UpdateUnschedulablePodsCountWithLabel(removedFakePodsCount, SkippedInjectionMetricsLabel)

	return unschedulablePodsAfterProcessing, nil
}

// CleanUp is called at CA termination
func (p *EnforceInjectedPodsLimitProcessor) CleanUp() {
}
