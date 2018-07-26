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
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodBuilder helps building pods for tests.
type PodBuilder interface {
	WithName(name string) PodBuilder
	AddContainer(container apiv1.Container) PodBuilder
	WithCreator(creatorObjectMeta *metav1.ObjectMeta, creatorTypeMeta *metav1.TypeMeta) PodBuilder
	Get() *apiv1.Pod
}

// Pod returns new PodBuilder.
func Pod() PodBuilder {
	return &podBuilderImpl{
		containers: make([]apiv1.Container, 0),
	}
}

type container struct {
	name string
	cpu  string
	mem  string
}

type podBuilderImpl struct {
	name              string
	containers        []apiv1.Container
	creatorObjectMeta *metav1.ObjectMeta
	creatorTypeMeta   *metav1.TypeMeta
}

func (pb *podBuilderImpl) WithName(name string) PodBuilder {
	r := *pb
	r.name = name
	return &r
}

func (pb *podBuilderImpl) AddContainer(container apiv1.Container) PodBuilder {
	r := *pb
	r.containers = append(r.containers, container)
	return &r
}

func (pb *podBuilderImpl) WithCreator(creatorObjectMeta *metav1.ObjectMeta, creatorTypeMeta *metav1.TypeMeta) PodBuilder {
	r := *pb
	r.creatorObjectMeta = creatorObjectMeta
	r.creatorTypeMeta = creatorTypeMeta
	return &r
}

func (pb *podBuilderImpl) Get() *apiv1.Pod {
	startTime := metav1.Time{testTimestamp}
	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      pb.name,
			SelfLink:  fmt.Sprintf("/api/v1/namespaces/default/pods/%s", pb.name),
		},
		Spec: apiv1.PodSpec{
			Containers: pb.containers,
		},
		Status: apiv1.PodStatus{
			StartTime: &startTime,
		},
	}

	if pb.creatorObjectMeta != nil && pb.creatorTypeMeta != nil {
		isController := true
		pod.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
			{
				UID:        pb.creatorObjectMeta.UID,
				Name:       pb.creatorObjectMeta.Name,
				APIVersion: pb.creatorObjectMeta.ResourceVersion,
				Kind:       pb.creatorTypeMeta.Kind,
				Controller: &isController,
			},
		}
	}

	return pod
}
