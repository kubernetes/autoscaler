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

package test

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type podStatusBuilder struct {
	cpuRequest *resource.Quantity
	memRequest *resource.Quantity
	cpuLimit   *resource.Quantity
	memLimit   *resource.Quantity
}

// PodStatus returns object that helps build PodStatus for tests.
func PodStatus() *podStatusBuilder {
	return &podStatusBuilder{}
}

func (ps *podStatusBuilder) WithCPURequest(cpuRequest resource.Quantity) *podStatusBuilder {
	r := *ps
	r.cpuRequest = &cpuRequest
	return &r
}

func (ps *podStatusBuilder) WithMemRequest(memRequest resource.Quantity) *podStatusBuilder {
	r := *ps
	r.memRequest = &memRequest
	return &r
}

func (ps *podStatusBuilder) WithCPULimit(cpuLimit resource.Quantity) *podStatusBuilder {
	r := *ps
	r.cpuLimit = &cpuLimit
	return &r
}

func (ps *podStatusBuilder) WithMemLimit(memLimit resource.Quantity) *podStatusBuilder {
	r := *ps
	r.memLimit = &memLimit
	return &r
}

func (ps *podStatusBuilder) Get() corev1.PodStatus {
	podStatus := corev1.PodStatus{
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{},
			Limits:   corev1.ResourceList{},
		},
	}
	if ps.cpuRequest != nil {
		podStatus.Resources.Requests[corev1.ResourceCPU] = *ps.cpuRequest
	}
	if ps.memRequest != nil {
		podStatus.Resources.Requests[corev1.ResourceMemory] = *ps.memRequest
	}
	if ps.cpuLimit != nil {
		podStatus.Resources.Limits[corev1.ResourceCPU] = *ps.cpuLimit
	}
	if ps.memLimit != nil {
		podStatus.Resources.Limits[corev1.ResourceMemory] = *ps.memLimit
	}
	return podStatus
}
