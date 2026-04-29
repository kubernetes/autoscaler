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

package controller

import (
	"fmt"
	"math"

	"gopkg.in/inf.v0"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"
	podutils "k8s.io/autoscaler/cluster-autoscaler/utils/pod"
	"k8s.io/klog/v2"
)

type resourceQuotaAllocator struct {
	client *cbclient.CapacityBufferClient
}

// newResourceQuotaAllocator creates an instance of resourceQuotaAllocator.
func newResourceQuotaAllocator(client *cbclient.CapacityBufferClient) *resourceQuotaAllocator {
	return &resourceQuotaAllocator{
		client: client,
	}
}

// Allocate limits the buffers' replicas if they exceed ResourceQuotas.
//
// The algorithm for allocating the quotas to the buffers is as follows:
// - get resource quotas and buffers in the namespace
// - for each buffer:
//   - build a fake pod from the buffer's pod template
//   - get resource requests from the pod (in case of buffers we're not interested in the limits)
//   - for each quota:
//   - check if the pod matches the quota scope. If it doesn't, go to next quota.
//   - calculate how many replicas would fit into the quota's limits. It's done by comparing
//     the pod resource requests with RQ .Status.Used and already calculated usages from the buffers.
//   - if number of allowed replicas < the current buffer's .Status.Replicas, add LimitedByQuota condition
//     to the buffer and limit the buffer's .Status.Replicas
//   - update usages by the buffers for matching quotas (which will be used when handling next buffers)
func (r *resourceQuotaAllocator) Allocate(namespace string, buffers []*v1.CapacityBuffer) []error {
	quotas, err := r.client.ListResourceQuotas(namespace)
	if err != nil {
		return []error{fmt.Errorf("resourceQuotaAllocator: Failed to list resource quotas, error: %v", err)}
	}

	usages := make(map[types.UID]corev1.ResourceList)
	var errors []error

	for _, buffer := range buffers {
		if buffer.Namespace != namespace {
			// should not happen, as buffers are filtered by the namespace in the controller
			klog.Warningf("resourceQuotaAllocator: buffer %q namespace mismatch: %q, expected: %q", buffer.Name, buffer.Namespace, namespace)
			continue
		}
		// Skip buffers that are not ready for provisioning or have no replicas
		if buffer.Status.PodTemplateRef == nil || buffer.Status.PodTemplateGeneration == nil || buffer.Status.Replicas == nil {
			klog.V(4).Infof("resourceQuotaAllocator: Skipping buffer %s (not ready or no replicas)", buffer.Name)
			continue
		}

		podTemplate, err := r.client.GetPodTemplate(buffer.Namespace, buffer.Status.PodTemplateRef.Name)
		if err != nil {
			klog.V(4).Infof("resourceQuotaAllocator: Skipping buffer %s (pod template not found)", buffer.Name)
			continue
		}
		pod := podutils.GetPodFromTemplate(&podTemplate.Template)
		pod.Namespace = buffer.Namespace

		podReqs := calculatePodRequests(pod)
		if len(podReqs) == 0 {
			continue
		}

		currentReplicas := *buffer.Status.Replicas
		allowedReplicas := currentReplicas
		var blockingQuotas []string
		var matchingQuotas []*corev1.ResourceQuota

		for _, quota := range quotas {
			matches, err := podMatchesQuotaScope(pod, quota)
			if err != nil {
				err = fmt.Errorf("resourceQuotaAllocator: failed to check if pod matches quota %q, error: %w", quota.Name, err)
				errors = append(errors, err)
				continue
			}
			if !matches {
				continue
			}
			matchingQuotas = append(matchingQuotas, quota)
			maxReplicasForQuota := getMaxReplicasForQuota(quota, podReqs, usages[quota.UID])
			if maxReplicasForQuota < currentReplicas {
				blockingQuotas = append(blockingQuotas, quota.Name)
			}
			if maxReplicasForQuota < allowedReplicas {
				allowedReplicas = maxReplicasForQuota
			}
		}

		if allowedReplicas < currentReplicas {
			klog.V(4).Infof("resourceQuotaAllocator: Limiting buffer %s from %d to %d due to quotas: %v", buffer.Name, currentReplicas, allowedReplicas, blockingQuotas)
			common.MarkBufferAsLimitedByQuota(buffer, currentReplicas, allowedReplicas, blockingQuotas)
		} else {
			common.UpdateBufferStatusLimitedByQuotas(buffer, false, "")
		}

		if allowedReplicas > 0 {
			r.updateUsages(matchingQuotas, usages, podReqs, allowedReplicas)
		}
	}
	return errors
}

func calculatePodRequests(pod *corev1.Pod) corev1.ResourceList {
	requests := podutils.PodRequests(pod)

	result := corev1.ResourceList{}

	// to be considered in the future - it's arguable whether it makes sense
	// to scale up a buffer if no pod can be created in the namespace.
	// requests[corev1.ResourcePods] = *resource.NewQuantity(1, resource.DecimalSI)

	for resource, request := range requests {
		result[resource] = request
		result[maskResourceWithPrefix(resource, corev1.DefaultResourceRequestsPrefix)] = request
	}
	return result
}

func podMatchesQuotaScope(pod *corev1.Pod, quota *corev1.ResourceQuota) (bool, error) {
	selectors := getScopeSelectorsFromQuota(quota)
	matches := true
	for _, selector := range selectors {
		innerMatch, err := podMatchesSelector(pod, selector)
		if err != nil {
			return false, err
		}
		matches = matches && innerMatch
	}
	return matches, nil
}

func getMaxReplicasForQuota(quota *corev1.ResourceQuota, podReqs corev1.ResourceList, reserved corev1.ResourceList) int32 {
	maxReplicas := int64(math.MaxInt64)

	for resName, hardLimit := range quota.Status.Hard {
		reqQuantity, found := podReqs[resName]
		if !found || reqQuantity.IsZero() {
			continue
		}

		usedQuantity := quota.Status.Used[resName]
		reservedQuantity := reserved[resName]

		available := hardLimit.DeepCopy()
		available.Sub(usedQuantity)
		available.Sub(reservedQuantity)

		if available.Value() < 0 {
			return 0
		}

		reqValue := reqQuantity.AsDec()
		availValue := available.AsDec()
		// QuoRound with s=0 and inf.RoundDown divides two numbers and scales the result down to the nearest integer.
		// This works correctly both for milliCPUs and for large integers in case of memory
		divResult := new(inf.Dec).QuoRound(availValue, reqValue, 0, inf.RoundDown)
		fit := divResult.UnscaledBig().Int64()
		if fit < maxReplicas {
			maxReplicas = fit
		}
	}
	if maxReplicas > math.MaxInt32 {
		return int32(math.MaxInt32)
	}
	return int32(maxReplicas)
}

func (r *resourceQuotaAllocator) updateUsages(quotas []*corev1.ResourceQuota, usages map[types.UID]corev1.ResourceList, podReqs corev1.ResourceList, replicas int32) {
	for _, quota := range quotas {
		if usages[quota.UID] == nil {
			usages[quota.UID] = make(corev1.ResourceList)
		}

		for resName, quantity := range podReqs {
			if _, ok := quota.Status.Hard[resName]; !ok {
				continue
			}
			currentUsage := usages[quota.UID][resName]
			quantity.Mul(int64(replicas))
			currentUsage.Add(quantity)
			usages[quota.UID][resName] = currentUsage
		}
	}
}

func maskResourceWithPrefix(resource corev1.ResourceName, prefix string) corev1.ResourceName {
	return corev1.ResourceName(fmt.Sprintf("%s%s", prefix, string(resource)))
}

// Most of the code below is copied from https://github.com/kubernetes/kubernetes/blob/cc55e3447816e49c0bd7128668da49b856294536/pkg/quota/v1/evaluator/core/pods.go
// this is internal k8s code, so it should not be used as a package.

func getScopeSelectorsFromQuota(quota *corev1.ResourceQuota) []corev1.ScopedResourceSelectorRequirement {
	var selectors []corev1.ScopedResourceSelectorRequirement
	for _, scope := range quota.Spec.Scopes {
		selectors = append(selectors, corev1.ScopedResourceSelectorRequirement{
			ScopeName: scope,
			Operator:  corev1.ScopeSelectorOpExists,
		})
	}
	if quota.Spec.ScopeSelector != nil {
		selectors = append(selectors, quota.Spec.ScopeSelector.MatchExpressions...)
	}
	return selectors
}

func podMatchesSelector(pod *corev1.Pod, selector corev1.ScopedResourceSelectorRequirement) (bool, error) {
	switch selector.ScopeName {
	case corev1.ResourceQuotaScopeNotTerminating:
		// we treat all buffer pods as not terminating
		return true, nil
	case corev1.ResourceQuotaScopeTerminating:
		return false, nil
	case corev1.ResourceQuotaScopeBestEffort:
		return isBestEffort(pod), nil
	case corev1.ResourceQuotaScopeNotBestEffort:
		return !isBestEffort(pod), nil
	case corev1.ResourceQuotaScopePriorityClass:
		return podMatchesPriorityClass(pod, selector)
	case corev1.ResourceQuotaScopeCrossNamespacePodAffinity:
		return usesCrossNamespacePodAffinity(pod), nil
	}
	return false, nil
}

// isBestEffort checks if Pod has BestEffort QoS
//
// pod is BestEffort if it doesn't have CPU and Memory limits and requests.
// It's assumed that we ran pod defaulting, so we can just look up for memory and CPU in the requests.
func isBestEffort(pod *corev1.Pod) bool {
	podReqs := podutils.PodRequests(pod)
	if _, ok := podReqs[corev1.ResourceCPU]; ok {
		return false
	}
	if _, ok := podReqs[corev1.ResourceMemory]; ok {
		return false
	}
	return true
}

func podMatchesPriorityClass(pod *corev1.Pod, selector corev1.ScopedResourceSelectorRequirement) (bool, error) {
	if selector.Operator == corev1.ScopeSelectorOpExists {
		return len(pod.Spec.PriorityClassName) != 0, nil
	}
	labelSelector, err := scopedResourceSelectorRequirementsAsSelector(selector)
	if err != nil {
		return false, fmt.Errorf("failed to parse and convert selector: %v", err)
	}
	var m map[string]string
	if len(pod.Spec.PriorityClassName) != 0 {
		m = map[string]string{string(corev1.ResourceQuotaScopePriorityClass): pod.Spec.PriorityClassName}
	}
	if labelSelector.Matches(labels.Set(m)) {
		return true, nil
	}
	return false, nil
}

func scopedResourceSelectorRequirementsAsSelector(ssr corev1.ScopedResourceSelectorRequirement) (labels.Selector, error) {
	selector := labels.NewSelector()
	var op selection.Operator
	switch ssr.Operator {
	case corev1.ScopeSelectorOpIn:
		op = selection.In
	case corev1.ScopeSelectorOpNotIn:
		op = selection.NotIn
	case corev1.ScopeSelectorOpExists:
		op = selection.Exists
	case corev1.ScopeSelectorOpDoesNotExist:
		op = selection.DoesNotExist
	default:
		return nil, fmt.Errorf("%q is not a valid scope selector operator", ssr.Operator)
	}
	r, err := labels.NewRequirement(string(ssr.ScopeName), op, ssr.Values)
	if err != nil {
		return nil, err
	}
	selector = selector.Add(*r)
	return selector, nil
}

func usesCrossNamespacePodAffinity(pod *corev1.Pod) bool {
	if pod == nil || pod.Spec.Affinity == nil {
		return false
	}

	affinity := pod.Spec.Affinity.PodAffinity
	if affinity != nil {
		if isCrossNamespacePodAffinityTerms(affinity.RequiredDuringSchedulingIgnoredDuringExecution) {
			return true
		}
		if isCrossNamespaceWeightedPodAffinityTerms(affinity.PreferredDuringSchedulingIgnoredDuringExecution) {
			return true
		}
	}

	antiAffinity := pod.Spec.Affinity.PodAntiAffinity
	if antiAffinity != nil {
		if isCrossNamespacePodAffinityTerms(antiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) {
			return true
		}
		if isCrossNamespaceWeightedPodAffinityTerms(antiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) {
			return true
		}
	}

	return false
}

func isCrossNamespacePodAffinityTerms(terms []corev1.PodAffinityTerm) bool {
	for _, t := range terms {
		if isCrossNamespacePodAffinityTerm(&t) {
			return true
		}
	}
	return false
}

func isCrossNamespaceWeightedPodAffinityTerms(terms []corev1.WeightedPodAffinityTerm) bool {
	for _, t := range terms {
		if isCrossNamespacePodAffinityTerm(&t.PodAffinityTerm) {
			return true
		}
	}
	return false
}

func isCrossNamespacePodAffinityTerm(term *corev1.PodAffinityTerm) bool {
	return len(term.Namespaces) != 0 || term.NamespaceSelector != nil
}
