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

package priority

import (
	"slices"
	"time"

	corev1 "k8s.io/api/core/v1"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/features"
	updater_config "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/updater/config"
	resourcehelpers "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/resources"
	vpa_utils "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
)

// NewScalingDirectionPodEvictionAdmission creates a PodEvictionAdmission object.
// It admits Pods for eviction based on the scaling direction, i.e. if a resource is scaled up (recommendation > requests) or scaled down (recommendation < requests).
// It also supports >= and <= relations.
func NewScalingDirectionPodEvictionAdmission(opts ...ScalingDirectionPodEvictionAdmissionOption) PodEvictionAdmission {
	s := &scalingDirectionPodEvictionAdmission{
		recentOOMThreshold: updater_config.DefaultUpdaterConfig().EvictAfterOOMThreshold,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// ScalingDirectionPodEvictionAdmissionOption is a function that configures a scalingDirectionPodEvictionAdmission.
type ScalingDirectionPodEvictionAdmissionOption func(*scalingDirectionPodEvictionAdmission)

// WithEvictAfterOOMThreshold sets the shared threshold for considering a container as recently OOM killed.
func WithEvictAfterOOMThreshold(threshold time.Duration) ScalingDirectionPodEvictionAdmissionOption {
	return func(s *scalingDirectionPodEvictionAdmission) {
		s.recentOOMThreshold = threshold
	}
}

type scalingDirectionPodEvictionAdmission struct {
	// evictionConfigs is the per-Pod eviction admission related configuration from VPA.
	evictionConfigs map[*corev1.Pod]vpaEvictionConfig
	// recentOOMThreshold is the shared threshold for considering a container as
	// recently OOM killed. It can be overridden by the per-VPA evictAfterOOMSeconds field.
	// The shared value comes from the --evict-after-oom-threshold flag (default 10 minutes).
	recentOOMThreshold time.Duration
}

type vpaEvictionConfig struct {
	requirements []*vpa_types.EvictionRequirement
	// recentOOMThreshold is this VPA's threshold for considering a container as recently OOM killed.
	recentOOMThreshold time.Duration
}

type container struct {
	requests corev1.ResourceList
	target   corev1.ResourceList // recommendation
	// hasRecentOOM is true if the container has been recently OOM killed.
	hasRecentOOM bool
}

// Admit admits a Pod for eviction in one of three cases
// * no EvictionRequirement exists for this Pod
// * no Resource requests are set for at least one Container in this Pod
// * all EvictionRequirements are evaluated to 'true' for at least one Container in this Pod
func (s *scalingDirectionPodEvictionAdmission) Admit(pod *corev1.Pod, resources *vpa_types.RecommendedPodResources) bool {
	config, ok := s.evictionConfigs[pod]
	if !ok {
		return true
	}
	for _, spec := range pod.Spec.Containers {
		recommendedResources := vpa_utils.GetRecommendationForContainer(spec.Name, resources)
		// if a container doesn't have a recommendation, the VPA has set `.containerPolicy.mode: off` for this container,
		// so we can skip this container
		if recommendedResources == nil {
			continue
		}
		requests, _ := resourcehelpers.ContainerRequestsAndLimits(spec.Name, pod)
		c := container{
			requests:     requests,
			target:       recommendedResources.Target,
			hasRecentOOM: config.hasRecentOOM(pod, spec.Name),
		}
		if c.admit(config.requirements) {
			return true
		}
	}
	return false
}

func (c container) admit(reqs []*vpa_types.EvictionRequirement) bool {
	if _, ok := c.requests[corev1.ResourceCPU]; !ok {
		return true
	}
	if _, ok := c.requests[corev1.ResourceMemory]; !ok {
		return true
	}
	return c.checkEvictionRequirements(reqs)
}

func (c container) checkEvictionRequirements(reqs []*vpa_types.EvictionRequirement) bool {
	for _, requirement := range reqs {
		var resultsForResources = []bool{}
		for _, resource := range requirement.Resources {
			ok := c.checkChangeRequirement(resource, requirement.ChangeRequirement)
			resultsForResources = append(resultsForResources, ok)
		}
		// *All* EvictionRequirements need to be evaluated to `true`. Each requirement passes if *at least one* of its resources satisfies changeRequirement. So if there's a single EvictionRequirement which evaluates to `false` because none of its resources satisfies changeRequirement, we can exit here and don't admit the Container
		if !slices.Contains(resultsForResources, true) {
			return false
		}
	}
	return true
}

func (c container) checkChangeRequirement(resource corev1.ResourceName, req vpa_types.EvictionChangeRequirement) bool {
	request, target := c.requests[resource], c.target[resource]
	switch req {
	case vpa_types.TargetHigherThanRequests:
		return target.MilliValue() > request.MilliValue()
	case vpa_types.TargetLowerThanRequests:
		return target.MilliValue() < request.MilliValue()
	case vpa_types.TargetHigherThanRequestsAndOOM:
		return c.hasRecentOOM && target.MilliValue() > request.MilliValue()
	}
	return false
}

// LoopInit initializes the object by creating a map holding all applicable EvictionRequirements for each Pod.
// The map is re-created on every call, to ensure that any changes to a VPA's EvictionRequirements are picked up and not leak any EvictionRequirements for no longer existing Pods.
func (s *scalingDirectionPodEvictionAdmission) LoopInit(_ []*corev1.Pod, vpaControlledPods map[*vpa_types.VerticalPodAutoscaler][]*corev1.Pod) {
	s.evictionConfigs = make(map[*corev1.Pod]vpaEvictionConfig)
	for vpa, pods := range vpaControlledPods {
		// When UpdatePolicy is not specified, the default policy will be followed, and the EvictionRequirements field will be nil
		if vpa.Spec.UpdatePolicy == nil {
			continue
		}
		config := newVPAEvictionConfig(vpa.Spec.UpdatePolicy, s.recentOOMThreshold)
		for _, pod := range pods {
			s.evictionConfigs[pod] = config
		}
	}
}

func (s *scalingDirectionPodEvictionAdmission) CleanUp() {
	s.evictionConfigs = make(map[*corev1.Pod]vpaEvictionConfig)
}

func newVPAEvictionConfig(policy *vpa_types.PodUpdatePolicy, recentOOMThreshold time.Duration) vpaEvictionConfig {
	c := vpaEvictionConfig{
		requirements:       policy.EvictionRequirements,
		recentOOMThreshold: recentOOMThreshold,
	}
	if policy.EvictAfterOOMSeconds == nil {
		return c
	}
	if !features.Enabled(features.PerVPAConfig) {
		return c
	}
	c.recentOOMThreshold = time.Duration(*policy.EvictAfterOOMSeconds) * time.Second
	return c
}

// hasRecentOOM returns true if the given container has recently been OOM killed or is currently in terminated state due to OOM.
func (c vpaEvictionConfig) hasRecentOOM(pod *corev1.Pod, containerName string) bool {
	for _, status := range pod.Status.ContainerStatuses {
		if status.Name != containerName {
			continue
		}
		return c.isRecentOOM(status.State.Terminated) || c.isRecentOOM(status.LastTerminationState.Terminated)
	}
	return false
}

func (c vpaEvictionConfig) isRecentOOM(t *corev1.ContainerStateTerminated) bool {
	return t != nil && t.Reason == "OOMKilled" && time.Since(t.FinishedAt.Time) < c.recentOOMThreshold
}
