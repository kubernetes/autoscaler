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
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/metrics"
)

const (
	EnforceFakePodsLimitDuration = "total-time-to-enforce-injected-pods-limit"
)

// EnforceFakePodsLimitProcessor is a PodListProcessor used to limit the number of injected fake pods.
type EnforceInjectedPodsLimitProcessor struct {
	podLimit int
}

// NewEnforceFakePodsLimitProcessor return an instance of EnforceFakePodsLimitProcessor
func NewEnforceInjectedPodsLimitProcessor(podLimit int) *EnforceInjectedPodsLimitProcessor {
	return &EnforceInjectedPodsLimitProcessor{
		podLimit: podLimit,
	}
}

func (p *EnforceInjectedPodsLimitProcessor) Process(ctx *context.AutoscalingContext, unschedulablePods []*apiv1.Pod) ([]*apiv1.Pod, error) {

	defer metrics.UpdateDurationFromStart(EnforceFakePodsLimitDuration, time.Now())

	numberOfFakePodsToRemove := len(unschedulablePods) - p.podLimit
	var unschedulablePodsAfterProcessing []*apiv1.Pod

	for _, pod := range unschedulablePods {
		if IsFake(pod) && numberOfFakePodsToRemove > 0 {
			numberOfFakePodsToRemove -= 1
			continue
		}

		unschedulablePodsAfterProcessing = append(unschedulablePodsAfterProcessing, pod)
	}

	return unschedulablePodsAfterProcessing, nil
}

func (p *EnforceInjectedPodsLimitProcessor) CleanUp() {
}
