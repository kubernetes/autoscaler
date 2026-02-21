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
	"sort"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/annotations"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

// UpdatePriorityCalculator is responsible for prioritizing updates on pods.
// It can returns a sorted list of pods in order of update priority.
// Update priority is proportional to fraction by which resources should be increased / decreased.
// i.e. pod with 10M current memory and recommendation 20M will have higher update priority
// than pod with 100M current memory and 150M recommendation (100% increase vs 50% increase)
type UpdatePriorityCalculator struct {
	vpa                     *vpa_types.VerticalPodAutoscaler
	pods                    []prioritizedPod
	config                  UpdateConfig
	recommendationProcessor vpa_api_util.RecommendationProcessor
	priorityProcessor       PriorityProcessor
}

// UpdateConfig holds configuration for UpdatePriorityCalculator
type UpdateConfig struct {
	// MinChangePriority is the minimum change priority that will trigger a update.
	// TODO: should have separate for Mem and CPU?
	MinChangePriority          float64
	PodLifetimeUpdateThreshold time.Duration
	EvictAfterOOMThreshold     time.Duration
}

// NewUpdatePriorityCalculator creates new UpdatePriorityCalculator for the given VPA object
// an update config.
// If the vpa resource policy is nil, there will be no policy restriction on update.
// If the given update config is nil, default values are used.
func NewUpdatePriorityCalculator(vpa *vpa_types.VerticalPodAutoscaler,
	config UpdateConfig,
	recommendationProcessor vpa_api_util.RecommendationProcessor,
	priorityProcessor PriorityProcessor) UpdatePriorityCalculator {
	return UpdatePriorityCalculator{
		vpa:                     vpa,
		config:                  config,
		recommendationProcessor: recommendationProcessor,
		priorityProcessor:       priorityProcessor}
}

// AddPod adds pod to the UpdatePriorityCalculator.
// The caller must hold the lock protecting the calculator.
func (calc *UpdatePriorityCalculator) AddPod(pod *corev1.Pod, now time.Time, infeasibleAttempts map[types.UID]*vpa_types.RecommendedPodResources) {
	processedRecommendation, _, err := calc.recommendationProcessor.Apply(calc.vpa, pod)
	if err != nil {
		klog.V(2).ErrorS(err, "Cannot process recommendation for pod", "pod", klog.KObj(pod))
		return
	}

	// Check if this recommendation was already tried and failed as infeasible
	// only for InPlace update mode
	if vpa_api_util.GetUpdateMode(calc.vpa) == vpa_types.UpdateModeInPlace {
		if lastAttempt, exists := infeasibleAttempts[pod.UID]; exists {
			// Only retry if the new recommendation has at least one resource lower than the last attempt
			if !hasLowerResourceRecommendation(lastAttempt, processedRecommendation) {
				klog.V(4).InfoS("Skipping pod, recommendation not lower than last infeasible attempt",
					"pod", klog.KObj(pod))
				return
			}
			// Recommendation has lower resource, will retry
			klog.V(4).InfoS("Recommendation changed since last infeasible attempt, will retry",
				"pod", klog.KObj(pod))
		}
	}

	hasObservedContainers, vpaContainerSet := parseVpaObservedContainers(pod)

	updatePriority := calc.priorityProcessor.GetUpdatePriority(pod, calc.vpa, processedRecommendation)

	quickOOM := false
	for i := range pod.Status.ContainerStatuses {
		cs := &pod.Status.ContainerStatuses[i]
		if hasObservedContainers && !vpaContainerSet.Has(cs.Name) {
			// Containers not observed by Admission Controller are not supported
			// by the quick OOM logic.
			klog.V(4).InfoS("Not listed in VPA observed containers label. Skipping container quick OOM calculations", "label", annotations.VpaObservedContainersLabel, "observedContainers", pod.GetAnnotations()[annotations.VpaObservedContainersLabel], "containerName", cs.Name, "vpa", klog.KObj(calc.vpa))
			continue
		}
		crp := vpa_api_util.GetContainerResourcePolicy(cs.Name, calc.vpa.Spec.ResourcePolicy)
		if crp != nil && crp.Mode != nil && *crp.Mode == vpa_types.ContainerScalingModeOff {
			// Containers with ContainerScalingModeOff are not considered
			// during the quick OOM calculation.
			klog.V(4).InfoS("Container with ContainerScalingModeOff. Skipping container quick OOM calculations", "containerName", cs.Name)
			continue
		}
		evictOOMThreshold := calc.getEvictOOMThreshold()
		terminationState := &cs.LastTerminationState
		if terminationState.Terminated != nil &&
			terminationState.Terminated.Reason == "OOMKilled" &&
			terminationState.Terminated.FinishedAt.Sub(terminationState.Terminated.StartedAt.Time) < evictOOMThreshold {
			quickOOM = true
			klog.V(2).InfoS("Quick OOM detected in pod", "pod", klog.KObj(pod), "containerName", cs.Name)
		}
	}

	// The update is allowed in following cases:
	// - the request is outside the recommended range for some container.
	// - the pod lives for at least 24h and the resource diff is >= MinChangePriority.
	// - a vpa scaled container OOMed in less than evictAfterOOMThreshold.
	if !updatePriority.OutsideRecommendedRange && !quickOOM {
		if pod.Status.StartTime == nil {
			// TODO: Set proper condition on the VPA.
			klog.V(4).InfoS("Not updating pod, missing field pod.Status.StartTime", "pod", klog.KObj(pod))
			return
		}
		if now.Before(pod.Status.StartTime.Add(calc.config.PodLifetimeUpdateThreshold)) {
			klog.V(4).InfoS("Not updating a short-lived pod, request within recommended range", "pod", klog.KObj(pod))
			return
		}
		if updatePriority.ResourceDiff < calc.config.MinChangePriority {
			klog.V(4).InfoS("Not updating pod, resource diff too low", "pod", klog.KObj(pod), "updatePriority", updatePriority)
			return
		}
	}

	// If the pod has quick OOMed then evict only if the resources will change
	if quickOOM && updatePriority.ResourceDiff == 0 {
		klog.V(4).InfoS("Not updating pod because resource would not change", "pod", klog.KObj(pod))
		return
	}
	klog.V(2).InfoS("Pod accepted for update", "pod", klog.KObj(pod), "updatePriority", updatePriority.ResourceDiff, "processedRecommendations", calc.GetProcessedRecommendationTargets(processedRecommendation))
	calc.pods = append(calc.pods, prioritizedPod{
		pod:            pod,
		priority:       updatePriority,
		recommendation: processedRecommendation})
}

// GetSortedPods returns a list of pods ordered by update priority (highest update priority first)
func (calc *UpdatePriorityCalculator) GetSortedPods(admission PodEvictionAdmission) []*corev1.Pod {
	sort.Sort(byPriorityDesc(calc.pods))

	result := []*corev1.Pod{}
	for _, podPrio := range calc.pods {
		if admission.Admit(podPrio.pod, podPrio.recommendation) {
			result = append(result, podPrio.pod)
		} else {
			klog.V(2).InfoS("Pod removed from update queue by PodEvictionAdmission", "pod", klog.KObj(podPrio.pod))
		}
	}

	return result
}

// GetProcessedRecommendationTargets takes a RecommendedPodResources object and returns a formatted string
// with the recommended pod resources. Specifically, it formats the target and uncapped target CPU and memory.
func (calc *UpdatePriorityCalculator) GetProcessedRecommendationTargets(r *vpa_types.RecommendedPodResources) string {
	sb := &strings.Builder{}
	for _, cr := range r.ContainerRecommendations {
		sb.WriteString(cr.ContainerName)
		sb.WriteString(": ")
		if cr.Target != nil {
			sb.WriteString("target: ")
			if !cr.Target.Memory().IsZero() {
				sb.WriteString(strconv.FormatInt(cr.Target.Memory().ScaledValue(resource.Kilo), 10))
				sb.WriteString("k ")
			}
			if !cr.Target.Cpu().IsZero() {
				sb.WriteString(strconv.FormatInt(cr.Target.Cpu().MilliValue(), 10))
				sb.WriteString("m; ")
			}
		}
		if cr.UncappedTarget != nil {
			sb.WriteString("uncappedTarget: ")
			if !cr.UncappedTarget.Memory().IsZero() {
				sb.WriteString(strconv.FormatInt(cr.Target.Memory().ScaledValue(resource.Kilo), 10))
				sb.WriteString("k ")
			}
			if !cr.UncappedTarget.Cpu().IsZero() {
				sb.WriteString(strconv.FormatInt(cr.Target.Cpu().MilliValue(), 10))
				sb.WriteString("m;")
			}
		}
	}
	return sb.String()
}

// getEvictOOMThreshold returns the duration to wait after an OOM event before
// considering the pod for eviction. It uses the VPA-specific EvictAfterOOMSeconds
// if the PerVPAConfig feature flag is enabled and the value is set, otherwise
// falls back to the global evictAfterOOMThreshold flag.
func (calc *UpdatePriorityCalculator) getEvictOOMThreshold() time.Duration {
	evictOOMThreshold := calc.config.EvictAfterOOMThreshold

	if calc.vpa.Spec.UpdatePolicy == nil || calc.vpa.Spec.UpdatePolicy.EvictAfterOOMSeconds == nil {
		return evictOOMThreshold
	}

	if !features.Enabled(features.PerVPAConfig) {
		klog.V(4).InfoS("feature flag is off, falling back to default EvictAfterOOMThreshold", "flagName", features.PerVPAConfig)
		return evictOOMThreshold
	}
	seconds := calc.vpa.Spec.UpdatePolicy.EvictAfterOOMSeconds
	duration := time.Duration(*seconds) * time.Second
	return duration
}

func parseVpaObservedContainers(pod *corev1.Pod) (bool, sets.Set[string]) {
	observedContainers, hasObservedContainers := pod.GetAnnotations()[annotations.VpaObservedContainersLabel]
	vpaContainerSet := sets.New[string]()
	if hasObservedContainers {
		if containers, err := annotations.ParseVpaObservedContainersValue(observedContainers); err != nil {
			klog.ErrorS(err, "VPA annotation failed to parse", "pod", klog.KObj(pod), "annotation", observedContainers)
			hasObservedContainers = false
		} else {
			vpaContainerSet.Insert(containers...)
		}
	}
	return hasObservedContainers, vpaContainerSet
}

// hasLowerResourceRecommendation checks if the new recommendation has any resource
// (CPU or Memory) that is lower than the last recommendation. This is used to determine
// if a previously infeasible update attempt should be retried.
func hasLowerResourceRecommendation(lastAttempt, newRecommendation *vpa_types.RecommendedPodResources) bool {
	if lastAttempt == nil || newRecommendation == nil {
		return false
	}

	// Create a map of container names to their recommendations for easier lookup
	lastAttemptMap := make(map[string]*vpa_types.RecommendedContainerResources)
	for i := range lastAttempt.ContainerRecommendations {
		cr := &lastAttempt.ContainerRecommendations[i]
		lastAttemptMap[cr.ContainerName] = cr
	}

	// Check if any resource in the new recommendation is lower than the last attempt
	for i := range newRecommendation.ContainerRecommendations {
		newCR := &newRecommendation.ContainerRecommendations[i]
		lastCR, exists := lastAttemptMap[newCR.ContainerName]
		if !exists {
			continue
		}

		// Compare target resources
		for resourceName, newTarget := range newCR.Target {
			if lastTarget, exists := lastCR.Target[resourceName]; exists {
				if newTarget.Cmp(lastTarget) < 0 {
					return true
				}
			}
		}
	}

	return false
}

type prioritizedPod struct {
	pod            *corev1.Pod
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
