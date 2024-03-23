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
	"strconv"
	"strings"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/annotations"
	vpa_api_util "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

var (
	defaultUpdateThreshold = flag.Float64("pod-update-threshold", 0.1, "Ignore updates that have priority lower than the value of this flag")

	podLifetimeUpdateThreshold = flag.Duration("in-recommendation-bounds-eviction-lifetime-threshold", time.Hour*12, "Pods that live for at least that long can be evicted even if their request is within the [MinRecommended...MaxRecommended] range")

	evictAfterOOMThreshold = flag.Duration("evict-after-oom-threshold", 10*time.Minute,
		`Evict pod that has OOMed in less than evict-after-oom-threshold since start.`)
)

// UpdatePriorityCalculator is responsible for prioritizing updates on pods.
// It can returns a sorted list of pods in order of update priority.
// Update priority is proportional to fraction by which resources should be increased / decreased.
// i.e. pod with 10M current memory and recommendation 20M will have higher update priority
// than pod with 100M current memory and 150M recommendation (100% increase vs 50% increase)
type UpdatePriorityCalculator struct {
	vpa                     *vpa_types.VerticalPodAutoscaler
	pods                    []PrioritizedPod
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
		config = &UpdateConfig{MinChangePriority: *defaultUpdateThreshold}
	}
	return UpdatePriorityCalculator{
		vpa:                     vpa,
		config:                  config,
		recommendationProcessor: recommendationProcessor,
		priorityProcessor:       priorityProcessor}
}

// AddPod adds pod to the UpdatePriorityCalculator.
func (calc *UpdatePriorityCalculator) AddPod(pod *apiv1.Pod, now time.Time) {
	processedRecommendation, _, err := calc.recommendationProcessor.Apply(calc.vpa, pod)
	if err != nil {
		klog.V(2).ErrorS(err, "Cannot process recommendation for pod", "pod", klog.KObj(pod))
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
		terminationState := &cs.LastTerminationState
		if terminationState.Terminated != nil &&
			terminationState.Terminated.Reason == "OOMKilled" &&
			terminationState.Terminated.FinishedAt.Time.Sub(terminationState.Terminated.StartedAt.Time) < *evictAfterOOMThreshold {
			quickOOM = true
			klog.V(2).InfoS("Quick OOM detected in pod", "pod", klog.KObj(pod), "containerName", cs.Name)
		}
	}

	disruptionlessRecommendation := calc.CalcualteDisruptionFreeActions(pod, processedRecommendation)

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
		if now.Before(pod.Status.StartTime.Add(*podLifetimeUpdateThreshold)) {
			// TODO(jkyros): do we need an in-place update threshold arg ?
			// If our recommendations are disruptionless, we can bypass the threshold limit
			if len(disruptionlessRecommendation.ContainerRecommendations) > 0 {
				klog.V(2).Infof("pod accepted for DISRUPTIONLESS (%d/%d) update %v/%v with priority %v", len(disruptionlessRecommendation.ContainerRecommendations), len(processedRecommendation.ContainerRecommendations), pod.Namespace, pod.Name, updatePriority.ResourceDiff)
				updatePriority.Disruptionless = true
				calc.pods = append(calc.pods, PrioritizedPod{
					pod:            pod,
					priority:       updatePriority,
					recommendation: disruptionlessRecommendation})
			}
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
	calc.pods = append(calc.pods, PrioritizedPod{
		pod:            pod,
		priority:       updatePriority,
		recommendation: processedRecommendation})
}

// GetSortedPrioritizedPods returns a list of prioritized pods ordered by update priority (highest update priority first). Used instead
// of GetSortedPods when we need access to the priority information
func (calc *UpdatePriorityCalculator) GetSortedPrioritizedPods(admission PodEvictionAdmission) []*PrioritizedPod {
	sort.Sort(byPriorityDesc(calc.pods))

	//result := []*apiv1.Pod{}
	result := []*PrioritizedPod{}
	for num, podPrio := range calc.pods {
		if admission.Admit(podPrio.pod, podPrio.recommendation) {
			result = append(result, &calc.pods[num])
		} else {
			klog.V(2).Infof("pod removed from update queue by PodEvictionAdmission: %v", podPrio.pod.Name)
		}
	}

	return result
}

// GetSortedPods returns a list of pods ordered by update priority (highest update priority first)
func (calc *UpdatePriorityCalculator) GetSortedPods(admission PodEvictionAdmission) []*apiv1.Pod {
	sort.Sort(byPriorityDesc(calc.pods))

	result := []*apiv1.Pod{}
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

func parseVpaObservedContainers(pod *apiv1.Pod) (bool, sets.Set[string]) {
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

// PrioritizedPod contains the priority and recommendation details for a pod.
// TODO(jkyros): I made this public, but there may be a cleaner way
type PrioritizedPod struct {
	pod            *apiv1.Pod
	priority       PodPriority
	recommendation *vpa_types.RecommendedPodResources
}

// IsDisruptionless returns the disruptionless status of the underlying pod priority
// TODO(jkyros): scope issues, maybe not the best place to put Disruptionless
func (p PrioritizedPod) IsDisruptionless() bool {
	return p.priority.Disruptionless
}

// Pod returns the underlying private pod
func (p PrioritizedPod) Pod() *apiv1.Pod {
	return p.pod
}

// PodPriority contains data for a pod update that can be used to prioritize between updates.
type PodPriority struct {
	// Is any container outside of the recommended range.
	OutsideRecommendedRange bool
	// Does any container want to grow.
	ScaleUp bool
	// Relative difference between the total requested and total recommended resources.
	ResourceDiff float64
	// Is this update disruptionless
	Disruptionless bool
}

type byPriorityDesc []PrioritizedPod

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

// CalcualteDisruptionFreeActions calculates the set of actions we think we can perform without disruption based on the pod/container resize/restart
// policies and returns that set of actions.
func (calc *UpdatePriorityCalculator) CalcualteDisruptionFreeActions(pod *apiv1.Pod, recommendation *vpa_types.RecommendedPodResources) *vpa_types.RecommendedPodResources {

	var disruptionlessRecommendation = &vpa_types.RecommendedPodResources{}

	for _, container := range pod.Spec.Containers {
		// If we don't have a resize policy, we can't check it
		if len(container.ResizePolicy) == 0 {
			continue
		}

		// So we get whatever the recommendation was for this container
		resourceRec := getRecommendationForContainerName(container.Name, recommendation)
		// If we didn't find a recommendation for this container, we don't have anything to do
		if resourceRec == nil {
			continue
		}
		// Then we go through all the resource recommendations it has
		for resource := range resourceRec.Target {
			// And we look up what the restart policy is for those resources
			resourceRestartPolicy := getRestartPolicyForResource(resource, container.ResizePolicy)
			// If we don't have one, that's probably bad
			if resourceRestartPolicy == nil {
				continue
			}
			// If we do have one, and it's disruptive, then we know this won't work
			if *resourceRestartPolicy != apiv1.NotRequired {
				continue
			}

		}

		// And if we made it here, we should theoretically be able to do this without disruption
		disruptionlessRecommendation.ContainerRecommendations = append(disruptionlessRecommendation.ContainerRecommendations, *resourceRec)

	}

	return disruptionlessRecommendation
}

// getRecommendationForContainerName searches through the list of ContainerRecommendations until it finds one matching the named container. Used
// to match up containers with their recommendations (we have container, we want resource recommendation)
func getRecommendationForContainerName(name string, recommendation *vpa_types.RecommendedPodResources) *vpa_types.RecommendedContainerResources {
	for _, recommendationContainer := range recommendation.ContainerRecommendations {
		if recommendationContainer.ContainerName == name {
			return &recommendationContainer
		}
	}
	return nil
}

// getRestartPolicyForResource searches through the list of resources in the resize policy until it finds the one matching the named resource. Used
// to match up restart policies with our resource recommendations (we have resource, we want policy).
func getRestartPolicyForResource(resourceName apiv1.ResourceName, policy []apiv1.ContainerResizePolicy) *apiv1.ResourceResizeRestartPolicy {
	// TODO(jkyros): can there be duplicate policies for resources? we just take the first one now
	for _, resizePolicy := range policy {
		if resizePolicy.ResourceName == resourceName {
			return &resizePolicy.RestartPolicy
		}
	}
	return nil
}
