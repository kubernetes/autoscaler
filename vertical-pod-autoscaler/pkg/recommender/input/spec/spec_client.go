/*
Copyright 2017 The Kubernetes Authors.

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

package spec

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	listersv1 "k8s.io/client-go/listers/core/v1"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	resourcehelpers "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/resources"
)

// BasicPodSpec contains basic information defining a pod and its containers.
type BasicPodSpec struct {
	// ID identifies a pod within a cluster.
	ID model.PodID
	// Labels of the pod. It is used to match pods with certain VPA opjects.
	PodLabels map[string]string
	// List of containers within this pod.
	Containers []BasicContainerSpec
	// List of init containers within this pod.
	InitContainers []BasicContainerSpec
	// PodPhase describing current life cycle phase of the Pod.
	Phase corev1.PodPhase
}

// BasicContainerSpec contains basic information defining a container.
type BasicContainerSpec struct {
	// ID identifies the container within a cluster.
	ID model.ContainerID
	// Name of the image running within the container.
	Image string
	// Currently requested resources for this container.
	Request model.Resources
}

// SpecClient provides information about pods and containers Specification
type SpecClient interface {
	// Returns BasicPodSpec for each pod in the cluster
	GetPodSpecs() ([]*BasicPodSpec, error)
}

type specClient struct {
	podLister listersv1.PodLister
}

// NewSpecClient creates new client which can be used to get basic information about pods specification
// It requires PodLister which is a data source for this client.
func NewSpecClient(podLister listersv1.PodLister) SpecClient {
	return &specClient{
		podLister: podLister,
	}
}

func (client *specClient) GetPodSpecs() ([]*BasicPodSpec, error) {
	var podSpecs []*BasicPodSpec

	pods, err := client.podLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}
	for _, pod := range pods {
		basicPodSpec := newBasicPodSpec(pod)
		podSpecs = append(podSpecs, basicPodSpec)
	}
	return podSpecs, nil
}

func newBasicPodSpec(pod *corev1.Pod) *BasicPodSpec {
	containerSpecs := newContainerSpecs(pod, pod.Spec.Containers, false /* isInitContainer */)
	initContainerSpecs := newContainerSpecs(pod, pod.Spec.InitContainers, true /* isInitContainer */)

	basicPodSpec := &BasicPodSpec{
		ID:             podID(pod),
		PodLabels:      pod.Labels,
		Containers:     containerSpecs,
		InitContainers: initContainerSpecs,
		Phase:          pod.Status.Phase,
	}
	return basicPodSpec
}

func newContainerSpecs(pod *corev1.Pod, containers []corev1.Container, isInitContainer bool) []BasicContainerSpec {
	var containerSpecs []BasicContainerSpec
	for _, container := range containers {
		containerSpec := newContainerSpec(pod, container, isInitContainer)
		containerSpecs = append(containerSpecs, containerSpec)
	}
	return containerSpecs
}

func newContainerSpec(pod *corev1.Pod, container corev1.Container, isInitContainer bool) BasicContainerSpec {
	containerSpec := BasicContainerSpec{
		ID: model.ContainerID{
			PodID:         podID(pod),
			ContainerName: container.Name,
		},
		Image:   container.Image,
		Request: calculateRequestedResources(pod, container, isInitContainer),
	}
	return containerSpec
}

func calculateRequestedResources(pod *corev1.Pod, container corev1.Container, isInitContainer bool) model.Resources {
	requestsAndLimitsFn := resourcehelpers.ContainerRequestsAndLimits
	if isInitContainer {
		requestsAndLimitsFn = resourcehelpers.InitContainerRequestsAndLimits
	}
	requests, _ := requestsAndLimitsFn(container.Name, pod)

	cpuQuantity := requests[corev1.ResourceCPU]
	cpuMillicores := cpuQuantity.MilliValue()

	memoryQuantity := requests[corev1.ResourceMemory]
	memoryBytes := memoryQuantity.Value()

	return model.Resources{
		model.ResourceCPU:    model.ResourceAmount(cpuMillicores),
		model.ResourceMemory: model.ResourceAmount(memoryBytes),
	}
}

func podID(pod *corev1.Pod) model.PodID {
	return model.PodID{
		PodName:   pod.Name,
		Namespace: pod.Namespace,
	}
}
