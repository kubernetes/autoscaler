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

package translator

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1alpha1"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	"k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

type resourceQuotasTranslator struct {
	client *cbclient.CapacityBufferClient
	usages map[types.UID]corev1.ResourceList
}

// NewResourceQuotasTranslator creates an instance of resourceQuotasTranslator.
func NewResourceQuotasTranslator(client *cbclient.CapacityBufferClient) *resourceQuotasTranslator {
	return &resourceQuotasTranslator{
		client: client,
		usages: make(map[types.UID]corev1.ResourceList),
	}
}

func (r *resourceQuotasTranslator) Translate(buffers []*v1.CapacityBuffer) []error {
	r.usages = make(map[types.UID]corev1.ResourceList)
	errors := []error{}

	for _, buffer := range buffers {
		// Skip buffers that are not ready for provisioning or have no replicas
		if buffer.Status.PodTemplateRef == nil || buffer.Status.Replicas == nil || *buffer.Status.Replicas <= 0 {
			klog.V(4).Infof("ResourceQuotasTranslator: Skipping buffer %s (not ready or no replicas)", buffer.Name)
			continue
		}

		podTemplate, err := r.client.GetPodTemplate(buffer.Namespace, buffer.Status.PodTemplateRef.Name)
		if err != nil {
			err = fmt.Errorf("ResourceQuotaTranslator: Couldn't get pod template, error: %v", err)
			errors = append(errors, err)
			continue
		}

		podReqs := calculatePodRequests(podTemplate)
		if len(podReqs) == 0 {
			continue
		}

		quotas, err := r.client.ListResourceQuotas(buffer.Namespace)
		if err != nil {
			err = fmt.Errorf("ResourceQuotaTranslator: Failed to list resource quotas, error: %v", err)
			errors = append(errors, err)
			continue
		}

		currentReplicas := *buffer.Status.Replicas
		allowedReplicas := currentReplicas
		blockingQuotas := []string{}

		for _, quota := range quotas {
			maxReplicasForQuota := getMaxReplicasForQuota(quota, podReqs, r.usages[quota.UID])
			if maxReplicasForQuota < int64(allowedReplicas) {
				allowedReplicas = int32(maxReplicasForQuota)
				blockingQuotas = append(blockingQuotas, quota.Name)
			}
		}

		if allowedReplicas < currentReplicas {
			klog.V(2).Infof("ResourceQuotasTranslator: Limiting buffer %s from %d to %d due to quotas: %v", buffer.Name, currentReplicas, allowedReplicas, blockingQuotas)
			buffer.Status.Replicas = ptr.To(allowedReplicas)
			msg := fmt.Sprintf("Buffer replicas limited from %d to %d due to quotas: %s", currentReplicas, allowedReplicas, strings.Join(blockingQuotas, ", "))
			common.UpdateBufferStatusLimitedByQuotas(buffer, true, msg)
		} else {
			// Not limited, but maybe previous run had it set, so clear it/set to False
			common.UpdateBufferStatusLimitedByQuotas(buffer, false, "")
		}

		if allowedReplicas > 0 {
			r.updateUsages(quotas, podReqs, allowedReplicas)
		}
	}
	return errors
}

func (r *resourceQuotasTranslator) updateUsages(quotas []*corev1.ResourceQuota, podReqs corev1.ResourceList, replicas int32) {
	for _, quota := range quotas {
		if r.usages[quota.UID] == nil {
			r.usages[quota.UID] = make(corev1.ResourceList)
		}

		// Check if quota actually limits any of the resources we use
		applies := false
		for resName := range podReqs {
			if _, ok := quota.Status.Hard[resName]; ok {
				applies = true
				break
			}
		}

		if applies {
			for resName, quantity := range podReqs {
				currentUsage := r.usages[quota.UID][resName]
				quantityCopy := quantity.DeepCopy()
				for i := int32(0); i < replicas; i++ {
					currentUsage.Add(quantityCopy)
				}
				r.usages[quota.UID][resName] = currentUsage
			}
		}
	}
}

func getMaxReplicasForQuota(quota *corev1.ResourceQuota, podReqs corev1.ResourceList, reserved corev1.ResourceList) int64 {
	// Assume unlimited
	maxReplicas := int64(1<<63 - 1)

	for resName, hardLimit := range quota.Status.Hard {
		reqQuantity, found := podReqs[resName]
		if !found {
			continue
		}

		usedQuantity := quota.Status.Used[resName]
		reservedQuantity := reserved[resName]

		// Available = Hard - Used - Reserved
		available := hardLimit.DeepCopy()
		available.Sub(usedQuantity)
		available.Sub(reservedQuantity)

		if available.Value() < 0 {
			return 0
		}

		reqValue := reqQuantity.MilliValue()
		if reqValue == 0 {
			// Request is 0, so infinite replicas allowed for this resource
			continue
		}

		availValue := available.MilliValue()
		fit := availValue / reqValue
		if fit < maxReplicas {
			maxReplicas = fit
		}
	}
	return maxReplicas
}

func calculatePodRequests(podTemplate *corev1.PodTemplate) corev1.ResourceList {
	reqs := make(corev1.ResourceList)
	for _, container := range podTemplate.Template.Spec.Containers {
		for name, quantity := range container.Resources.Requests {
			if val, ok := reqs[name]; ok {
				val.Add(quantity)
				reqs[name] = val
			} else {
				reqs[name] = quantity.DeepCopy()
			}
		}
	}
	// Implicitly every pod consumes 1 "pods" resource
	reqs[corev1.ResourcePods] = *resource.NewQuantity(1, resource.DecimalSI)
	return reqs
}

func (r *resourceQuotasTranslator) CleanUp() {
	r.usages = make(map[types.UID]corev1.ResourceList)
}
