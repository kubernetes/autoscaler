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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// PodBuilder helps building pods for tests.
type PodBuilder interface {
	WithName(name string) PodBuilder
	WithUID(types.UID) PodBuilder
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
	uid                   types.UID
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

func (pb *podBuilderImpl) WithUID(uid types.UID) PodBuilder {
	r := *pb
	r.uid = uid
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

func (pb *podBuilderImpl) Get() *corev1.Pod {
	startTime := metav1.Time{
		Time: testTimestamp,
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      pb.name,
		},
		Spec: corev1.PodSpec{
			Containers:     pb.containers,
			InitContainers: pb.initContainers,
		},
		Status: corev1.PodStatus{
			StartTime:  &startTime,
			Conditions: pb.conditions,
		},
	}

	if pb.labels != nil {
		pod.Labels = pb.labels
	}

	if pb.annotations != nil {
		pod.Annotations = pb.annotations
	}

	if pb.uid != "" {
		pod.UID = pb.uid
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

	return pod
}
