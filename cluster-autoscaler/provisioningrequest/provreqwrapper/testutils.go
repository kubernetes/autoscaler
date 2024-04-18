/*
Copyright 2024 The Kubernetes Authors.

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

package provreqwrapper

import (
	"fmt"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/apis/provisioningrequest/autoscaling.x-k8s.io/v1beta1"
)

// BuildTestProvisioningRequest builds ProvisioningRequest wrapper.
func BuildTestProvisioningRequest(namespace, name, cpu, memory, gpu string, podCount int32,
	antiAffinity bool, creationTimestamp time.Time, provisioningRequestClass string) *ProvisioningRequest {
	gpuResource := resource.Quantity{}
	tolerations := []apiv1.Toleration{}
	if len(gpu) > 0 {
		gpuResource = resource.MustParse(gpu)
		tolerations = append(tolerations, apiv1.Toleration{Key: "nvidia.com/gpu", Operator: apiv1.TolerationOpExists})
	}

	affinity := &apiv1.Affinity{}
	if antiAffinity {
		affinity.PodAntiAffinity = &apiv1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []apiv1.PodAffinityTerm{
				{
					LabelSelector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      "app",
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{"test-app"},
							},
						},
					},
					TopologyKey: "failure-domain.beta.kubernetes.io/zone",
				},
			},
		}
	}
	return NewProvisioningRequest(
		&v1beta1.ProvisioningRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:              name,
				Namespace:         namespace,
				CreationTimestamp: v1.NewTime(creationTimestamp),
			},
			Spec: v1beta1.ProvisioningRequestSpec{
				ProvisioningClassName: provisioningRequestClass,
				PodSets: []v1beta1.PodSet{
					{
						PodTemplateRef: v1beta1.Reference{Name: fmt.Sprintf("%s-template-name", name)},
						Count:          podCount,
					},
				},
			},
			Status: v1beta1.ProvisioningRequestStatus{
				Conditions: []metav1.Condition{},
			},
		},
		[]*apiv1.PodTemplate{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:              fmt.Sprintf("%s-template-name", name),
					Namespace:         namespace,
					CreationTimestamp: v1.NewTime(creationTimestamp),
				},
				Template: apiv1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "test-app",
						},
					},
					Spec: apiv1.PodSpec{
						Tolerations: tolerations,
						Affinity:    affinity,
						Containers: []apiv1.Container{
							{
								Name:    "pi",
								Image:   "perl",
								Command: []string{"/bin/sh"},
								Resources: apiv1.ResourceRequirements{
									Limits: apiv1.ResourceList{
										apiv1.ResourceCPU:    resource.MustParse(cpu),
										apiv1.ResourceMemory: resource.MustParse(memory),
										"nvidia.com/gpu":     gpuResource,
									},
									Requests: apiv1.ResourceList{
										apiv1.ResourceCPU:    resource.MustParse(cpu),
										apiv1.ResourceMemory: resource.MustParse(memory),
										"nvidia.com/gpu":     gpuResource,
									},
								},
							},
						},
					},
				},
			},
		})
}
