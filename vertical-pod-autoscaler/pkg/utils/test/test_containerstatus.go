/*
Copyright 2019 The Kubernetes Authors.

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

type containerStatusBuilder struct {
	name       string
	cpuRequest *resource.Quantity
	memRequest *resource.Quantity
	cpuLimit   *resource.Quantity
	memLimit   *resource.Quantity
}

// ContainerStatus returns object that helps build containerStatus for tests.
func ContainerStatus() *containerStatusBuilder {
	return &containerStatusBuilder{}
}

func (cs *containerStatusBuilder) WithName(name string) *containerStatusBuilder {
	r := *cs
	r.name = name
	return &r
}

func (cs *containerStatusBuilder) WithCPURequest(cpuRequest resource.Quantity) *containerStatusBuilder {
	r := *cs
	r.cpuRequest = &cpuRequest
	return &r
}

func (cs *containerStatusBuilder) WithMemRequest(memRequest resource.Quantity) *containerStatusBuilder {
	r := *cs
	r.memRequest = &memRequest
	return &r
}

func (cs *containerStatusBuilder) WithCPULimit(cpuLimit resource.Quantity) *containerStatusBuilder {
	r := *cs
	r.cpuLimit = &cpuLimit
	return &r
}

func (cs *containerStatusBuilder) WithMemLimit(memLimit resource.Quantity) *containerStatusBuilder {
	r := *cs
	r.memLimit = &memLimit
	return &r
}

func (cs *containerStatusBuilder) Get() corev1.ContainerStatus {
	containerStatus := corev1.ContainerStatus{
		Name: cs.name,
		Resources: &corev1.ResourceRequirements{
			Requests: corev1.ResourceList{},
			Limits:   corev1.ResourceList{},
		},
	}
	if cs.cpuRequest != nil {
		containerStatus.Resources.Requests[corev1.ResourceCPU] = *cs.cpuRequest
	}
	if cs.memRequest != nil {
		containerStatus.Resources.Requests[corev1.ResourceMemory] = *cs.memRequest
	}
	if cs.cpuLimit != nil {
		containerStatus.Resources.Limits[corev1.ResourceCPU] = *cs.cpuLimit
	}
	if cs.memLimit != nil {
		containerStatus.Resources.Limits[corev1.ResourceMemory] = *cs.memLimit
	}
	return containerStatus
}
