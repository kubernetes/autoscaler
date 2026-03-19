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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodBuilder helps building pods for tests.
type PodBuilder interface {
	WithName(name string) PodBuilder
	AddContainer(container corev1.Container) PodBuilder
	AddInitContainer(initContainer corev1.Container) PodBuilder
	AddContainerStatus(containerStatus corev1.ContainerStatus) PodBuilder
	AddInitContainerStatus(initContainerStatus corev1.ContainerStatus) PodBuilder
	WithCreator(creatorObjectMeta *metav1.ObjectMeta, creatorTypeMeta *metav1.TypeMeta) PodBuilder
	WithLabels(labels map[string]string) PodBuilder
	WithAnnotations(annotations map[string]string) PodBuilder
	WithPhase(phase corev1.PodPhase) PodBuilder
	WithQOSClass(class corev1.PodQOSClass) PodBuilder
	WithPodConditions(conditions []corev1.PodCondition) PodBuilder
	WithCPULimit(cpuLimit resource.Quantity) PodBuilder
	WithCPURequest(cpuRequest resource.Quantity) PodBuilder
	WithMemLimit(memLimit resource.Quantity) PodBuilder
	WithMemRequest(memRequest resource.Quantity) PodBuilder
	AddPodStatus(podStatus corev1.PodStatus) PodBuilder
	Get() *corev1.Pod
}

// Pod returns new PodBuilder.
func Pod() PodBuilder {
	return &podBuilderImpl{
		containers:        make([]corev1.Container, 0),
		containerStatuses: make([]corev1.ContainerStatus, 0),
	}
}

type podBuilderImpl struct {
	name                  string
	containers            []corev1.Container
	initContainers        []corev1.Container
	creatorObjectMeta     *metav1.ObjectMeta
	creatorTypeMeta       *metav1.TypeMeta
	labels                map[string]string
	annotations           map[string]string
	phase                 corev1.PodPhase
	containerStatuses     []corev1.ContainerStatus
	initContainerStatuses []corev1.ContainerStatus
	qosClass              corev1.PodQOSClass
	conditions            []corev1.PodCondition
	cpuRequest            *resource.Quantity
	memRequest            *resource.Quantity
	cpuLimit              *resource.Quantity
	memLimit              *resource.Quantity
	podStatus             *corev1.PodStatus
}

func (pb *podBuilderImpl) WithLabels(labels map[string]string) PodBuilder {
	r := *pb
	r.labels = labels
	return &r
}

func (pb *podBuilderImpl) WithAnnotations(annotations map[string]string) PodBuilder {
	r := *pb
	r.annotations = annotations
	return &r
}

func (pb *podBuilderImpl) WithName(name string) PodBuilder {
	r := *pb
	r.name = name
	return &r
}

func (pb *podBuilderImpl) AddContainer(container corev1.Container) PodBuilder {
	r := *pb
	r.containers = append(r.containers, container)
	return &r
}

func (pb *podBuilderImpl) AddInitContainer(initContainer corev1.Container) PodBuilder {
	r := *pb
	r.initContainers = append(r.initContainers, initContainer)
	return &r
}

func (pb *podBuilderImpl) WithCreator(creatorObjectMeta *metav1.ObjectMeta, creatorTypeMeta *metav1.TypeMeta) PodBuilder {
	r := *pb
	r.creatorObjectMeta = creatorObjectMeta
	r.creatorTypeMeta = creatorTypeMeta
	return &r
}

func (pb *podBuilderImpl) WithPhase(phase corev1.PodPhase) PodBuilder {
	r := *pb
	r.phase = phase
	return &r
}

func (pb *podBuilderImpl) AddContainerStatus(containerStatus corev1.ContainerStatus) PodBuilder {
	r := *pb
	r.containerStatuses = append(r.containerStatuses, containerStatus)
	return &r
}

func (pb *podBuilderImpl) AddInitContainerStatus(initContainerStatus corev1.ContainerStatus) PodBuilder {
	r := *pb
	r.initContainerStatuses = append(r.initContainerStatuses, initContainerStatus)
	return &r
}

func (pb *podBuilderImpl) WithQOSClass(class corev1.PodQOSClass) PodBuilder {
	r := *pb
	r.qosClass = class
	return &r
}

func (pb *podBuilderImpl) WithPodConditions(conditions []corev1.PodCondition) PodBuilder {
	r := *pb
	r.conditions = conditions
	return &r
}

func (pb *podBuilderImpl) WithCPURequest(cpuRequest resource.Quantity) PodBuilder {
	r := *pb
	r.cpuRequest = &cpuRequest
	return &r
}

func (pb *podBuilderImpl) WithMemRequest(memRequest resource.Quantity) PodBuilder {
	r := *pb
	r.memRequest = &memRequest
	return &r
}

func (pb *podBuilderImpl) WithCPULimit(cpuLimit resource.Quantity) PodBuilder {
	r := *pb
	r.cpuLimit = &cpuLimit
	return &r
}

func (pb *podBuilderImpl) WithMemLimit(memLimit resource.Quantity) PodBuilder {
	r := *pb
	r.memLimit = &memLimit
	return &r
}

func (pb *podBuilderImpl) AddPodStatus(podStatus corev1.PodStatus) PodBuilder {
	r := *pb
	r.podStatus = &podStatus
	return &r
}

func (pb *podBuilderImpl) Get() *corev1.Pod {
	startTime := metav1.Time{
		Time: testTimestamp,
	}

	status := corev1.PodStatus{
		StartTime:  &startTime,
		Conditions: pb.conditions,
	}
	if pb.podStatus != nil {
		status = *pb.podStatus
		status.StartTime = &startTime
		status.Conditions = pb.conditions
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      pb.name,
		},
		Spec: corev1.PodSpec{
			Containers:     pb.containers,
			InitContainers: pb.initContainers,
			Resources: &corev1.ResourceRequirements{
				Requests: corev1.ResourceList{},
				Limits:   corev1.ResourceList{},
			},
		},
		Status: status,
	}

	if pb.labels != nil {
		pod.Labels = pb.labels
	}

	if pb.annotations != nil {
		pod.Annotations = pb.annotations
	}

	if pb.creatorObjectMeta != nil && pb.creatorTypeMeta != nil {
		isController := true
		pod.OwnerReferences = []metav1.OwnerReference{
			{
				UID:        pb.creatorObjectMeta.UID,
				Name:       pb.creatorObjectMeta.Name,
				APIVersion: pb.creatorObjectMeta.ResourceVersion,
				Kind:       pb.creatorTypeMeta.Kind,
				Controller: &isController,
			},
		}
	}
	if pb.phase != "" {
		pod.Status.Phase = pb.phase
	}
	if pb.qosClass != "" {
		pod.Status.QOSClass = pb.qosClass
	}
	if pb.containerStatuses != nil {
		pod.Status.ContainerStatuses = pb.containerStatuses
	}
	if pb.initContainerStatuses != nil {
		pod.Status.InitContainerStatuses = pb.initContainerStatuses
	}
	if pb.cpuRequest != nil {
		pod.Spec.Resources.Requests[corev1.ResourceCPU] = *pb.cpuRequest
	}
	if pb.memRequest != nil {
		pod.Spec.Resources.Requests[corev1.ResourceMemory] = *pb.memRequest
	}
	if pb.cpuLimit != nil {
		pod.Spec.Resources.Limits[corev1.ResourceCPU] = *pb.cpuLimit
	}
	if pb.memLimit != nil {
		pod.Spec.Resources.Limits[corev1.ResourceMemory] = *pb.memLimit
	}

	return pod
}
