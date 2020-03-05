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
	"flag"
	"sort"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/annotations"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
	"k8s.io/klog"
)

const (
	// Ignore change priority that is smaller than 10%.
	defaultUpdateThreshold = 0.10
	// Pods that live for at least that long can be evicted even if their
	// request is within the [MinRecommended...MaxRecommended] range.
	podLifetimeUpdateThreshold = time.Hour * 12
)

var (
	evictAfterOOMThreshold = flag.Duration("evict-after-oom-threshold", 10*time.Minute,
		`Evict pod that has only one container and it OOMed in less than
		evict-after-oom-threshold since start.`)
)

// UpdatePriorityCalculator is responsible for prioritizing updates on pods.
// It can returns a sorted list of pods in order of update priority.
// Update priority is proportional to fraction by which resources should be increased / decreased.
// i.e. pod with 10M current memory and recommendation 20M will have higher update priority
// than pod with 100M current memory and 150M recommendation (100% increase vs 50% increase)
type UpdatePriorityCalculator struct {
	vpa                     *vpa_types.VerticalPodAutoscaler
	pods                    []prioritizedPod
	config                  *UpdateConfig
	recommendationProcessor vpa_api_util.RecommendationProcessor
	priorityProcessor       PriorityProcessor
}

// UpdateConfig holds configuration for UpdatePriorityCalculator
type UpdateConfig struct {
	// MinChangePriority is the minimum change priority that will trigger a update.
	// TODO: should have separate for Mem and CPU?
	MinChangePriority float64
}

// NewUpdatePriorityCalculator creates new UpdatePriorityCalculator for the given VPA object
// an update config.
// If the vpa resource policy is nil, there will be no policy restriction on update.
// If the given update config is nil, default values are used.
func NewUpdatePriorityCalculator(vpa *vpa_types.VerticalPodAutoscaler,
	config *UpdateConfig,
	recommendationProcessor vpa_api_util.RecommendationProcessor,
	priorityProcessor PriorityProcessor) UpdatePriorityCalculator {
	if config == nil {
		config = &UpdateConfig{MinChangePriority: defaultUpdateThreshold}
	}
	return UpdatePriorityCalculator{
		vpa:                     vpa,
		config:                  config,
		recommendationProcessor: recommendationProcessor,
		priorityProcessor:       priorityProcessor}
}

// AddPod adds pod to the UpdatePriorityCalculator.
func (calc *UpdatePriorityCalculator) AddPod(pod *apiv1.Pod, now time.Time) {
	processedRecommendation, _, err := calc.recommendationProcessor.Apply(calc.vpa.Status.Recommendation, calc.vpa.Spec.ResourcePolicy, calc.vpa.Status.Conditions, pod)
	if err != nil {
		klog.V(2).Infof("cannot process recommendation for pod %s/%s: %v", pod.Namespace, pod.Name, err)
		return
	}

	hasObservedContainers, vpaContainerSet := parseVpaObservedContainers(pod)

	updatePriority := calc.priorityProcessor.GetUpdatePriority(pod, calc.vpa, processedRecommendation)

	quickOOM := false
	for i := range pod.Status.ContainerStatuses {
		cs := &pod.Status.ContainerStatuses[i]
		if hasObservedContainers && !vpaContainerSet.Has(cs.Name) {
			// Containers not observed by Admission Controller are not supported
			// by the quick OOM logic.
			klog.V(4).Infof("Not listed in %s:%s. Skipping container %s quick OOM calculations",
				annotations.VpaObservedContainersLabel, pod.GetAnnotations()[annotations.VpaObservedContainersLabel], cs.Name)
			continue
		}
		crp := vpa_api_util.GetContainerResourcePolicy(cs.Name, calc.vpa.Spec.ResourcePolicy)
		if crp != nil && crp.Mode != nil && *crp.Mode == vpa_types.ContainerScalingModeOff {
			// Containers with ContainerScalingModeOff are not considered
			// during the quick OOM calculation.
			klog.V(4).Infof("Container with ContainerScalingModeOff. Skipping container %s quick OOM calculations", cs.Name)
			continue
		}
		terminationState := &cs.LastTerminationState
		if terminationState.Terminated != nil &&
			terminationState.Terminated.Reason == "OOMKilled" &&
			terminationState.Terminated.FinishedAt.Time.Sub(terminationState.Terminated.StartedAt.Time) < *evictAfterOOMThreshold {
			quickOOM = true
			klog.V(2).Infof("quick OOM detected in pod %v/%v, container %v", pod.Namespace, pod.Name, cs.Name)
		}
	}

	// The update is allowed in following cases:
	// - the request is outside the recommended range for some container.
	// - the pod lives for at least 24h and the resource diff is >= MinChangePriority.
	// - a vpa scaled container OOMed in less than evictAfterOOMThreshold.
	if !updatePriority.OutsideRecommendedRange && !quickOOM {
		if pod.Status.StartTime == nil {
			// TODO: Set proper condition on the VPA.
			klog.V(2).Infof("not updating pod %v/%v, missing field pod.Status.StartTime", pod.Namespace, pod.Name)
			return
		}
		if now.Before(pod.Status.StartTime.Add(podLifetimeUpdateThreshold)) {
			klog.V(2).Infof("not updating a short-lived pod %v/%v, request within recommended range", pod.Namespace, pod.Name)
			return
		}
		if updatePriority.ResourceDiff < calc.config.MinChangePriority {
			klog.V(2).Infof("not updating pod %v/%v, resource diff too low: %v", pod.Namespace, pod.Name, updatePriority)
			return
		}
	}

	// If the pod has quick OOMed then evict only if the resources will change
	if quickOOM && updatePriority.ResourceDiff == 0 {
		klog.V(2).Infof("not updating pod %v/%v because resource would not change", pod.Namespace, pod.Name)
		return
	}
	klog.V(2).Infof("pod accepted for update %v/%v with priority %v", pod.Namespace, pod.Name, updatePriority.ResourceDiff)
	calc.pods = append(calc.pods, prioritizedPod{
		pod:            pod,
		priority:       updatePriority,
		recommendation: processedRecommendation})
}

// GetSortedPods returns a list of pods ordered by update priority (highest update priority first)
func (calc *UpdatePriorityCalculator) GetSortedPods(admission PodEvictionAdmission) []*apiv1.Pod {
	sort.Sort(byPriorityDesc(calc.pods))

	result := []*apiv1.Pod{}
	for _, podPrio := range calc.pods {
		if admission == nil || admission.Admit(podPrio.pod, podPrio.recommendation) {
			result = append(result, podPrio.pod)
		} else {
			klog.V(2).Infof("pod removed from update queue by PodEvictionAdmission: %v", podPrio.pod.Name)
		}
	}

	return result
}

func parseVpaObservedContainers(pod *apiv1.Pod) (bool, sets.String) {
	observedContainers, hasObservedContainers := pod.GetAnnotations()[annotations.VpaObservedContainersLabel]
	vpaContainerSet := sets.NewString()
	if hasObservedContainers {
		if containers, err := annotations.ParseVpaObservedContainersValue(observedContainers); err != nil {
			klog.Errorf("Vpa annotation %s failed to parse: %v", observedContainers, err)
			hasObservedContainers = false
		} else {
			vpaContainerSet.Insert(containers...)
		}
	}
	return hasObservedContainers, vpaContainerSet
}

type prioritizedPod struct {
	pod            *apiv1.Pod
	priority       PodPriority
	recommendation *vpa_types.RecommendedPodResources
}

// PodPriority contains data for a pod update that can be used to prioritize between updates.
type PodPriority struct {
	// Is any container outside of the recommended range.
	OutsideRecommendedRange bool
	// Does any container want to grow.
	ScaleUp bool
	// Relative difference between the total requested and total recommended resources.
	ResourceDiff float64
}

type byPriorityDesc []prioritizedPod

func (list byPriorityDesc) Len() int {
	return len(list)
}
func (list byPriorityDesc) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

// Less implements reverse ordering by priority (highest priority first).
// This means we return true if priority at index j is lower than at index i.
func (list byPriorityDesc) Less(i, j int) bool {
	return list[j].priority.Less(list[i].priority)
}

// Less returns true if p is lower than other.
func (p PodPriority) Less(other PodPriority) bool {
	// 1. If any container wants to grow, the pod takes precedence.
	// TODO: A better policy would be to prioritize scaling down when
	// (a) the pod is pending
	// (b) there is general resource shortage
	// and prioritize scaling up otherwise.
	if p.ScaleUp != other.ScaleUp {
		return other.ScaleUp
	}
	// 2. A pod with larger value of resourceDiff takes precedence.
	return p.ResourceDiff < other.ResourceDiff
}
