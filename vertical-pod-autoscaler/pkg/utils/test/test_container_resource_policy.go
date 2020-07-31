/*
Copyright 2018 The Kubernetes Authors.

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
	core "k8s.io/api/core/v1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

// ContainerResourcePolicyBuilder helps building test instances of ContainerResourcePolicy.
type ContainerResourcePolicyBuilder interface {
	WithContainer(containerName string) ContainerResourcePolicyBuilder
	WithMinAllowed(cpu, memory string) ContainerResourcePolicyBuilder
	WithMaxAllowed(cpu, memory string) ContainerResourcePolicyBuilder
	WithControlledValues(mode vpa_types.ContainerControlledValues) ContainerResourcePolicyBuilder
	WithPreventScaleDown(PreventScaleDown bool) ContainerResourcePolicyBuilder

	Get() vpa_types.ContainerResourcePolicy
}

// ContainerResourcePolicy returns a new ContainerResourcePolicyBuilder.
func ContainerResourcePolicy() ContainerResourcePolicyBuilder {
	return &containerResourcePolicyBuilder{}
}

type containerResourcePolicyBuilder struct {
	containerName    string
	minAllowed       core.ResourceList
	maxAllowed       core.ResourceList
	ControlledValues *vpa_types.ContainerControlledValues
	PreventScaleDown bool
}

func (b *containerResourcePolicyBuilder) WithContainer(containerName string) ContainerResourcePolicyBuilder {
	c := *b
	c.containerName = containerName
	return &c
}

func (b *containerResourcePolicyBuilder) WithMinAllowed(cpu, memory string) ContainerResourcePolicyBuilder {
	c := *b
	c.minAllowed = Resources(cpu, memory)
	return &c
}

func (b *containerResourcePolicyBuilder) WithMaxAllowed(cpu, memory string) ContainerResourcePolicyBuilder {
	c := *b
	c.maxAllowed = Resources(cpu, memory)
	return &c
}

func (b *containerResourcePolicyBuilder) WithControlledValues(mode vpa_types.ContainerControlledValues) ContainerResourcePolicyBuilder {
	c := *b
	c.ControlledValues = &mode
	return &c
}

func (b *containerResourcePolicyBuilder) WithPreventScaleDown(PreventScaleDown bool) ContainerResourcePolicyBuilder {
	c := *b
	c.PreventScaleDown = PreventScaleDown
	return &c
}

func (b *containerResourcePolicyBuilder) Get() vpa_types.ContainerResourcePolicy {
	if b.containerName == "" {
		panic("Must call WithContainer() before Get()")
	}
	return vpa_types.ContainerResourcePolicy{
		ContainerName:    b.containerName,
		MinAllowed:       b.minAllowed,
		MaxAllowed:       b.maxAllowed,
		ControlledValues: b.ControlledValues,
		PreventScaleDown: b.PreventScaleDown,
	}
}
