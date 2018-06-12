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
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	v1lister "k8s.io/client-go/listers/core/v1"
)

// BasicPodSpec contains basic information defining a pod and its containers.
type BasicPodSpec struct {
	// ID identifies a pod within a cluster.
	ID model.PodID
	// Labels of the pod. It is used to match pods with certain VPA opjects.
	PodLabels map[string]string
	// List of containers within this pod.
	Containers []BasicContainerSpec
	// PodPhase describing current life cycle phase of the Pod.
	Phase v1.PodPhase
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

//SpecClient provides information about pods and containers Specification
type SpecClient interface {
	// Returns BasicPodSpec for each pod in the cluster
	GetPodSpecs() ([]*BasicPodSpec, error)
}

type specClient struct {
	podLister v1lister.PodLister
}

// NewSpecClient creates new client which can be used to get basic information about pods specification
// It requires PodLister which is a data source for this client.
func NewSpecClient(podLister v1lister.PodLister) SpecClient {
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
func newBasicPodSpec(pod *v1.Pod) *BasicPodSpec {
	podId := model.PodID{
		PodName:   pod.Name,
		Namespace: pod.Namespace,
	}
	containerSpecs := newContainerSpecs(podId, pod)

	basicPodSpec := &BasicPodSpec{
		ID:         podId,
		PodLabels:  pod.Labels,
		Containers: containerSpecs,
		Phase:      pod.Status.Phase,
	}
	return basicPodSpec
}

func newContainerSpecs(podID model.PodID, pod *v1.Pod) []BasicContainerSpec {
	var containerSpecs []BasicContainerSpec

	for _, container := range pod.Spec.Containers {
		containerSpec := newContainerSpec(podID, container)
		containerSpecs = append(containerSpecs, containerSpec)
	}

	return containerSpecs
}

func newContainerSpec(podID model.PodID, container v1.Container) BasicContainerSpec {
	containerSpec := BasicContainerSpec{
		ID: model.ContainerID{
			PodID:         podID,
			ContainerName: container.Name,
		},
		Image:   container.Image,
		Request: calculateRequestedResources(container),
	}
	return containerSpec
}

func calculateRequestedResources(container v1.Container) model.Resources {
	cpuQuantity := container.Resources.Requests[v1.ResourceCPU]
	cpuMillicores := cpuQuantity.MilliValue()

	memoryQuantity := container.Resources.Requests[v1.ResourceMemory]
	memoryBytes := memoryQuantity.Value()

	return model.Resources{
		model.ResourceCPU:    model.ResourceAmount(cpuMillicores),
		model.ResourceMemory: model.ResourceAmount(memoryBytes),
	}

}
