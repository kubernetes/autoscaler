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

package pod

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/kubernetes/pkg/features"
	"k8s.io/kubernetes/pkg/kubelet/types"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	resourcehelper "k8s.io/component-helpers/resource"
)

const (
	// DaemonSetPodAnnotationKey - annotation use to informs the cluster-autoscaler controller when a pod needs to be considered as a Daemonset's Pod.
	DaemonSetPodAnnotationKey = "cluster-autoscaler.kubernetes.io/daemonset-pod"
)

// IsDaemonSetPod returns true if the Pod should be considered as Pod managed by a DaemonSet
func IsDaemonSetPod(pod *apiv1.Pod) bool {
	controllerRef := metav1.GetControllerOf(pod)
	if controllerRef != nil && controllerRef.Kind == "DaemonSet" {
		return true
	}

	return pod.Annotations[DaemonSetPodAnnotationKey] == "true"
}

// IsMirrorPod checks whether the pod is a mirror pod.
func IsMirrorPod(pod *apiv1.Pod) bool {
	if pod.ObjectMeta.Annotations == nil {
		return false
	}
	_, found := pod.ObjectMeta.Annotations[types.ConfigMirrorAnnotationKey]
	return found
}

// IsStaticPod returns true if the pod is a static pod.
func IsStaticPod(pod *apiv1.Pod) bool {
	if pod.Annotations != nil {
		if source, ok := pod.Annotations[types.ConfigSourceAnnotationKey]; ok == true {
			return source != types.ApiserverSource
		}
	}
	return false
}

// FilterRecreatablePods filters pods that will be recreated by their controllers
func FilterRecreatablePods(pods []*apiv1.Pod) []*apiv1.Pod {
	filtered := make([]*apiv1.Pod, 0, len(pods))
	for _, p := range pods {
		if IsStaticPod(p) || IsMirrorPod(p) || IsDaemonSetPod(p) {
			continue
		}
		filtered = append(filtered, p)
	}
	return filtered
}

// ClearPodNodeNames removes node name from pods
func ClearPodNodeNames(pods []*apiv1.Pod) []*apiv1.Pod {
	newPods := make([]*apiv1.Pod, 0, len(pods))
	for _, podPtr := range pods {
		newPod := *podPtr
		newPod.Spec.NodeName = ""
		newPods = append(newPods, &newPod)
	}
	return newPods
}

// PodRequests calculates Pod requests using a common resource helper shared with the scheduler
func PodRequests(pod *apiv1.Pod) apiv1.ResourceList {
	inPlacePodVerticalScalingEnabled := utilfeature.DefaultFeatureGate.Enabled(features.InPlacePodVerticalScaling)
	podLevelResourcesEnabled := utilfeature.DefaultFeatureGate.Enabled(features.PodLevelResources)

	return resourcehelper.PodRequests(pod, resourcehelper.PodResourcesOptions{
		UseStatusResources:    inPlacePodVerticalScalingEnabled,
		SkipPodLevelResources: !podLevelResourcesEnabled,
	})
}

// GetPodFromTemplate generates a Pod from a PodTemplateSpec.
//
// Source: https://github.com/kubernetes/kubernetes/blob/f366ba158ab7f0370e4e988dca8b0330a5952f43/pkg/controller/controller_utils.go#L562
func GetPodFromTemplate(template *apiv1.PodTemplateSpec) *apiv1.Pod {
	desiredLabels := getPodsLabelSet(template)
	desiredFinalizers := getPodsFinalizers(template)
	desiredAnnotations := getPodsAnnotationSet(template)

	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      desiredLabels,
			Annotations: desiredAnnotations,
			Finalizers:  desiredFinalizers,
		},
	}

	pod.Spec = template.Spec
	return pod
}

func getPodsLabelSet(template *apiv1.PodTemplateSpec) labels.Set {
	desiredLabels := make(labels.Set)
	for k, v := range template.Labels {
		desiredLabels[k] = v
	}
	return desiredLabels
}

func getPodsFinalizers(template *apiv1.PodTemplateSpec) []string {
	desiredFinalizers := make([]string, len(template.Finalizers))
	copy(desiredFinalizers, template.Finalizers)
	return desiredFinalizers
}

func getPodsAnnotationSet(template *apiv1.PodTemplateSpec) labels.Set {
	desiredAnnotations := make(labels.Set)
	for k, v := range template.Annotations {
		desiredAnnotations[k] = v
	}
	return desiredAnnotations
}
