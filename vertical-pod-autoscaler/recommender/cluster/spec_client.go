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

package cluster

import (
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/model"
	v1lister "k8s.io/kubernetes/pkg/client/listers/core/v1"
)

//SpecClient provides information about pods and containers Specification
type SpecClient interface {
	// Returns BasicPodSpec for each pod in the cluster
	GetPodSpecs() ([]*model.BasicPodSpec, error)
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

func (client *specClient) GetPodSpecs() ([]*model.BasicPodSpec, error) {
	var podSpecs []*model.BasicPodSpec

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
func newBasicPodSpec(pod *v1.Pod) *model.BasicPodSpec {
	podId := model.PodID{
		PodName:   pod.Name,
		Namespace: pod.Namespace,
	}
	containerSpecs := newContainerSpecs(podId, pod)

	basicPodSpec := &model.BasicPodSpec{
		ID:         podId,
		PodLabels:  pod.Labels,
		Containers: containerSpecs,
	}
	return basicPodSpec
}

func newContainerSpecs(podID model.PodID, pod *v1.Pod) []model.BasicContainerSpec {
	var containerSpecs []model.BasicContainerSpec

	for _, container := range pod.Spec.Containers {
		containerSpec := newContainerSpec(podID, container)
		containerSpecs = append(containerSpecs, containerSpec)
	}

	return containerSpecs
}

func newContainerSpec(podID model.PodID, container v1.Container) model.BasicContainerSpec {
	containerSpec := model.BasicContainerSpec{
		ID: model.ContainerID{
			PodID:         podID,
			ContainerName: container.Name,
		},
		Image:   container.Image,
		Request: calculateRequestedResources(container),
	}
	return containerSpec
}

func calculateRequestedResources(container v1.Container) map[model.MetricName]model.ResourceAmount {
	cpuQuantity := container.Resources.Requests[v1.ResourceCPU]
	cpuMillicores := cpuQuantity.MilliValue()

	memoryQuantity := container.Resources.Requests[v1.ResourceMemory]
	memoryBytes := memoryQuantity.Value()

	return map[model.MetricName]model.ResourceAmount{
		model.ResourceCPU:    model.ResourceAmount(cpuMillicores),
		model.ResourceMemory: model.ResourceAmount(memoryBytes),
	}

}
