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
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type containerBuilder struct {
	name       string
	cpuRequest *resource.Quantity
	memRequest *resource.Quantity
	cpuLimit   *resource.Quantity
	memLimit   *resource.Quantity
}

// Container returns object that helps build containers for tests.
func Container() *containerBuilder {
	return &containerBuilder{}
}

func (cb *containerBuilder) WithName(name string) *containerBuilder {
	r := *cb
	r.name = name
	return &r
}

func (cb *containerBuilder) WithCPURequest(cpuRequest resource.Quantity) *containerBuilder {
	r := *cb
	r.cpuRequest = &cpuRequest
	return &r
}

func (cb *containerBuilder) WithMemRequest(memRequest resource.Quantity) *containerBuilder {
	r := *cb
	r.memRequest = &memRequest
	return &r
}

func (cb *containerBuilder) WithCPULimit(cpuLimit resource.Quantity) *containerBuilder {
	r := *cb
	r.cpuLimit = &cpuLimit
	return &r
}

func (cb *containerBuilder) WithMemLimit(memLimit resource.Quantity) *containerBuilder {
	r := *cb
	r.memLimit = &memLimit
	return &r
}

func (cb *containerBuilder) Get() apiv1.Container {
	container := apiv1.Container{
		Name: cb.name,
		Resources: apiv1.ResourceRequirements{
			Requests: apiv1.ResourceList{},
			Limits:   apiv1.ResourceList{},
		},
	}
	if cb.cpuRequest != nil {
		container.Resources.Requests[apiv1.ResourceCPU] = *cb.cpuRequest
	}
	if cb.memRequest != nil {
		container.Resources.Requests[apiv1.ResourceMemory] = *cb.memRequest
	}
	if cb.cpuLimit != nil {
		container.Resources.Limits[apiv1.ResourceCPU] = *cb.cpuLimit
	}
	if cb.memLimit != nil {
		container.Resources.Limits[apiv1.ResourceMemory] = *cb.memLimit
	}
	return container
}
