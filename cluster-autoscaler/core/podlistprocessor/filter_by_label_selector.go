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

package podlistprocessor

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/klog/v2"
)

// FilterByLabelSelectorPodListProcessor filters pods based on a label selector.
// If the selector matches a pod's labels, the pod is included (for include selectors)
// or excluded (for exclude selectors using NotIn operator).
type FilterByLabelSelectorPodListProcessor struct{}

// NewFilterByLabelSelectorPodListProcessor creates a new FilterByLabelSelectorPodListProcessor.
func NewFilterByLabelSelectorPodListProcessor() *FilterByLabelSelectorPodListProcessor {
	return &FilterByLabelSelectorPodListProcessor{}
}

// Process filters the list of unschedulable pods based on the PodLabelSelector
// from the AutoscalingContext options.
func (p *FilterByLabelSelectorPodListProcessor) Process(
	ctx *context.AutoscalingContext,
	unschedulablePods []*apiv1.Pod) ([]*apiv1.Pod, error) {

	selector := ctx.PodLabelSelector
	if selector == nil || selector.Empty() {
		// No selector configured, return all pods
		return unschedulablePods, nil
	}

	result := make([]*apiv1.Pod, 0, len(unschedulablePods))
	for _, pod := range unschedulablePods {
		podLabels := labels.Set(pod.Labels)
		if selector.Matches(podLabels) {
			result = append(result, pod)
		} else {
			klog.V(4).Infof("Braze: Filtering out pod %s/%s because it doesn't match selector %s (pod labels: %v)",
				pod.Namespace, pod.Name, selector.String(), pod.Labels)
		}
	}

	if len(result) < len(unschedulablePods) {
		klog.V(2).Infof("Braze: Filtered %d pods using selector %s, %d pods remaining",
			len(unschedulablePods)-len(result), selector.String(), len(result))
	}

	return result, nil
}

// CleanUp cleans up the processor's internal structures.
func (p *FilterByLabelSelectorPodListProcessor) CleanUp() {
}
