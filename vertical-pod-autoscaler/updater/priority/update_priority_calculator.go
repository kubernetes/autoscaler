/*
Copyright 2017 The Kubernetes Authors.

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

package priority

import (
	"math"
	"sort"

	"github.com/golang/glog"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

const (
	// ignore change priority that is smaller than 10%
	defaultUpdateThreshod = 0.10
)

// UpdatePriorityCalculator is responsible for prioritizing updates on pods.
// It can returns a sorted list of pods in order of update priority.
// Update priority is proportional to fraction by which resources should be increased / decreased.
// i.e. pod with 10M current memory and recommendation 20M will have higher update priority
// than pod with 100M current memory and 150M recommendation (100% increase vs 50% increase)
type UpdatePriorityCalculator struct {
	resourcesPolicy *vpa_types.PodResourcePolicy
	pods            []podPriority
	config          *UpdateConfig
}

// UpdateConfig holds configuration for UpdatePriorityCalculator
type UpdateConfig struct {
	// MinChangePriority is the minimum change priority that will trigger a update.
	// TODO: should have separate for Mem and CPU?
	MinChangePriority float64
}

// NewUpdatePriorityCalculator creates new UpdatePriorityCalculator for the given resources policy and configuration.
// If the given policy is nil, there will be no policy restriction on update.
// If the given config is nil, default values are used.
func NewUpdatePriorityCalculator(policy *vpa_types.PodResourcePolicy, config *UpdateConfig) UpdatePriorityCalculator {
	if config == nil {
		config = &UpdateConfig{MinChangePriority: defaultUpdateThreshod}
	}
	return UpdatePriorityCalculator{resourcesPolicy: policy, config: config}
}

// AddPod adds pod to the UpdatePriorityCalculator.
func (calc *UpdatePriorityCalculator) AddPod(pod *apiv1.Pod, recommendation *vpa_types.RecommendedPodResources) {
	updatePriority := calc.getUpdatePriority(pod, recommendation)

	if updatePriority < calc.config.MinChangePriority {
		glog.V(2).Infof("pod not accepted for update %v, priority too low: %v", pod.Name, updatePriority)
		return
	}

	glog.V(2).Infof("pod accepted for update %v with priority %v", pod.Name, updatePriority)
	calc.pods = append(calc.pods, podPriority{pod, updatePriority})
}

// GetSortedPods returns a list of pods ordered by update priority (highest update priority first)
func (calc *UpdatePriorityCalculator) GetSortedPods() []*apiv1.Pod {
	sort.Sort(byPriority(calc.pods))

	result := make([]*apiv1.Pod, len(calc.pods))
	for i, podPrio := range calc.pods {
		result[i] = podPrio.pod
	}

	return result
}

func (calc *UpdatePriorityCalculator) getUpdatePriority(pod *apiv1.Pod, recommendation *vpa_types.RecommendedPodResources) float64 {
	var priority float64

	for _, podContainer := range pod.Spec.Containers {
		recommendedRequest, err := vpa_api_util.GetCappedRecommendationForContainer(podContainer, recommendation, calc.resourcesPolicy)
		if err != nil {
			glog.V(2).Infof("no recommendation for container %v in pod %v", podContainer.Name, pod.Name)
			continue
		}
		for resourceName, recommended := range recommendedRequest.Target {
			var requested *resource.Quantity

			if request, ok := podContainer.Resources.Requests[resourceName]; ok {
				requested = &request
			}
			resourceDiff := getPercentageDiff(requested, &recommended)
			priority += math.Abs(resourceDiff)
		}
	}
	return priority
}

func getPercentageDiff(request, recommendation *resource.Quantity) float64 {
	if request == nil {
		// resource requirement is not currently specified
		// any recommendation for this resource we will treat as 100% change
		return 1.0
	}
	if recommendation == nil || recommendation.IsZero() {
		return 0
	}
	return float64(recommendation.MilliValue()-request.MilliValue()) / float64(request.MilliValue())
}

type podPriority struct {
	pod      *apiv1.Pod
	priority float64
}
type byPriority []podPriority

func (list byPriority) Len() int {
	return len(list)
}
func (list byPriority) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}
func (list byPriority) Less(i, j int) bool {
	// reverse ordering, highest priority first
	return list[i].priority > list[j].priority
}
