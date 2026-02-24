/*
Copyright The Kubernetes Authors.

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

package fakepods

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	podutils "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
	resourcehelper "k8s.io/component-helpers/resource"
)

// DefaultingResolver performs a basic defaulting of a Pod.
type DefaultingResolver struct{}

// NewDefaultingResolver creates a new DefaultingResolver.
func NewDefaultingResolver() *DefaultingResolver {
	return &DefaultingResolver{}
}

// Resolve builds a Pod based on the provided template.
//
// Pod only goes through the basic defaulting logic in order to ensure that
// resource requests are defaulted from resource limits. The Pod will not go through
// the mutating and validating webhooks, so the resulting Pod might not be identical
// to a real Pod created with the same spec, and might also be invalid.
func (r *DefaultingResolver) Resolve(_ context.Context, namespace string, template *corev1.PodTemplateSpec) (*corev1.Pod, error) {
	pod := podutils.GetPodFromTemplate(template)
	pod.Namespace = namespace
	defaultPodResources(pod)
	return pod, nil
}

// defaultPodResources mimics defaulting behavior from https://github.com/kubernetes/kubernetes/blob/62277ef5d29d2aed692ae8013d5eb289bf75c0b5/pkg/apis/core/v1/defaults.go#L164
func defaultPodResources(pod *corev1.Pod) {
	for i := range pod.Spec.Containers {
		defaultContainerRequests(&pod.Spec.Containers[i])
	}
	for i := range pod.Spec.InitContainers {
		defaultContainerRequests(&pod.Spec.InitContainers[i])
	}
	defaultPodRequests(pod)
}

func defaultContainerRequests(container *corev1.Container) {
	if container.Resources.Requests == nil {
		container.Resources.Requests = make(corev1.ResourceList)
	}
	for name, quantity := range container.Resources.Limits {
		if _, ok := container.Resources.Requests[name]; !ok {
			container.Resources.Requests[name] = quantity
		}
	}
}

// Source: https://github.com/kubernetes/kubernetes/blob/6df7e09ad9bbe6ef1354ea889df07a28bbf8363d/pkg/apis/core/v1/defaults.go#L433
func defaultPodRequests(obj *corev1.Pod) {
	// We only populate defaults when the pod-level resources are partly specified already.
	if obj.Spec.Resources == nil {
		return
	}

	if len(obj.Spec.Resources.Limits) == 0 {
		return
	}

	var podReqs corev1.ResourceList
	podReqs = obj.Spec.Resources.Requests
	if podReqs == nil {
		podReqs = make(corev1.ResourceList)
	}

	aggrCtrReqs := resourcehelper.AggregateContainerRequests(obj, resourcehelper.PodResourcesOptions{})

	// When containers specify requests for a resource (supported by
	// PodLevelResources feature) and pod-level requests are not set, the pod-level requests
	// default to the effective requests of all the containers for that resource.
	for key, aggrCtrLim := range aggrCtrReqs {
		// Default pod level requests for overcommittable resources from aggregated container requests.
		if _, exists := podReqs[key]; !exists && resourcehelper.IsSupportedPodLevelResource(key) && strings.HasPrefix(string(key), corev1.ResourceHugePagesPrefix) {
			podReqs[key] = aggrCtrLim.DeepCopy()
		}
	}

	// When no containers specify requests for a resource, the pod-level requests
	// will default to match the pod-level limits, if pod-level
	// limits exist for that resource.
	// Defaulting for pod level hugepages requests is dependent on defaultHugePagePodLimits,
	// if defaultHugePagePodLimits defined the limit, the request will be set here.
	for key, podLim := range obj.Spec.Resources.Limits {
		if _, exists := podReqs[key]; !exists && resourcehelper.IsSupportedPodLevelResource(key) {
			podReqs[key] = podLim.DeepCopy()
		}
	}

	// Only set pod-level resource requests in the PodSpec if the requirements map
	// contains entries after collecting container-level requests and pod-level limits.
	if len(podReqs) > 0 {
		obj.Spec.Resources.Requests = podReqs
	}
}
