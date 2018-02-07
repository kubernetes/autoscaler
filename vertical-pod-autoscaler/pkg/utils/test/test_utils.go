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

package test

import (
	"fmt"
	"log"

	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/poc.autoscaling.k8s.io/v1alpha1"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/poc.autoscaling.k8s.io/v1alpha1"
	v1 "k8s.io/client-go/listers/core/v1"
)

// BuildTestPod creates a pod with specified resources.
func BuildTestPod(name, containerName, cpu, mem string, creatorObjectMeta *metav1.ObjectMeta, creatorTypeMeta *metav1.TypeMeta) *apiv1.Pod {
	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      name,
			SelfLink:  fmt.Sprintf("/api/v1/namespaces/default/pods/%s", name),
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{BuildTestContainer(containerName, cpu, mem)},
		},
	}

	if creatorObjectMeta != nil && creatorTypeMeta != nil {
		isController := true
		pod.ObjectMeta.OwnerReferences = []metav1.OwnerReference{
			{
				UID:        creatorObjectMeta.UID,
				Name:       creatorObjectMeta.Name,
				APIVersion: creatorObjectMeta.ResourceVersion,
				Kind:       creatorTypeMeta.Kind,
				Controller: &isController,
			},
		}
	}

	if len(cpu) > 0 {
		cpuVal, _ := resource.ParseQuantity(cpu)
		pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceCPU] = cpuVal
	}
	if len(mem) > 0 {
		memVal, _ := resource.ParseQuantity(mem)
		pod.Spec.Containers[0].Resources.Requests[apiv1.ResourceMemory] = memVal
	}

	return pod
}

// BuildTestContainer creates container with specified resources
func BuildTestContainer(containerName, cpu, mem string) apiv1.Container {
	container := apiv1.Container{
		Name: containerName,
		Resources: apiv1.ResourceRequirements{
			Requests: apiv1.ResourceList{},
		},
	}

	if len(cpu) > 0 {
		cpuVal, _ := resource.ParseQuantity(cpu)
		container.Resources.Requests[apiv1.ResourceCPU] = cpuVal
	}
	if len(mem) > 0 {
		memVal, _ := resource.ParseQuantity(mem)
		container.Resources.Requests[apiv1.ResourceMemory] = memVal
	}

	return container
}

// BuildTestPolicy creates ResourcesPolicy with specified constraints
func BuildTestPolicy(containerName, minCpu, maxCpu, minMemory, maxMemory string) *vpa_types.PodResourcePolicy {
	minCpuVal, _ := resource.ParseQuantity(minCpu)
	maxCpuVal, _ := resource.ParseQuantity(maxCpu)
	minMemVal, _ := resource.ParseQuantity(minMemory)
	maxMemVal, _ := resource.ParseQuantity(maxMemory)
	return &vpa_types.PodResourcePolicy{ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
		Name: containerName,
		MinAllowed: apiv1.ResourceList{
			apiv1.ResourceMemory: minMemVal,
			apiv1.ResourceCPU:    minCpuVal,
		},
		MaxAllowed: apiv1.ResourceList{
			apiv1.ResourceMemory: maxMemVal,
			apiv1.ResourceCPU:    maxCpuVal,
		},
	},
	}}
}

// BuildTestVerticalPodAutoscaler creates VerticalPodAutoscaler withs specified policy constraints
func BuildTestVerticalPodAutoscaler(containerName, targetCpu, minCpu, maxCpu, targetMemory, minMemory, maxMemory string, selector string) *vpa_types.VerticalPodAutoscaler {
	resourcesPolicy := BuildTestPolicy(containerName, minCpu, maxCpu, minMemory, maxMemory)

	labelSelector, err := metav1.ParseToLabelSelector(selector)
	if err != nil {
		log.Fatal(err)
	}
	targetCpuVal, _ := resource.ParseQuantity(targetCpu)
	targetMemoryVal, _ := resource.ParseQuantity(targetMemory)

	return &vpa_types.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
		},
		Spec: vpa_types.VerticalPodAutoscalerSpec{
			Selector:       labelSelector,
			UpdatePolicy:   vpa_types.PodUpdatePolicy{},
			ResourcePolicy: *resourcesPolicy,
		},
		Status: vpa_types.VerticalPodAutoscalerStatus{
			Recommendation: vpa_types.RecommendedPodResources{
				ContainerRecommendations: []vpa_types.RecommendedContainerResources{
					{
						Name: containerName,
						Target: apiv1.ResourceList{
							apiv1.ResourceMemory: targetMemoryVal,
							apiv1.ResourceCPU:    targetCpuVal,
						},
					},
				},
			},
		},
	}
}

// Recommendation creates Recommendation with specified container name and resources
func Recommendation(containerName, cpu, mem string) *vpa_types.RecommendedPodResources {
	result := &vpa_types.RecommendedPodResources{ContainerRecommendations: []vpa_types.RecommendedContainerResources{
		{Name: containerName,
			Target: make(map[apiv1.ResourceName]resource.Quantity, 0)}},
	}
	if len(cpu) > 0 {
		cpuVal, _ := resource.ParseQuantity(cpu)
		result.ContainerRecommendations[0].Target[apiv1.ResourceCPU] = cpuVal
	}

	if len(mem) > 0 {
		memVal, _ := resource.ParseQuantity(mem)
		result.ContainerRecommendations[0].Target[apiv1.ResourceMemory] = memVal
	}

	return result
}

// RecommenderAPIMock is a mock of RecommenderAPI
type RecommenderAPIMock struct {
	mock.Mock
}

// GetRecommendation is mock implementation of RecommenderAPI.GetRecommendation
func (m *RecommenderAPIMock) GetRecommendation(spec *apiv1.PodSpec) (*vpa_types.RecommendedPodResources, error) {
	args := m.Called(spec)
	var returnArg *vpa_types.RecommendedPodResources
	if args.Get(0) != nil {
		returnArg = args.Get(0).(*vpa_types.RecommendedPodResources)
	}
	return returnArg, args.Error(1)
}

// RecommenderMock is a mock of Recommender
type RecommenderMock struct {
	mock.Mock
}

// Get is a mock implementation of Recommender.Get
func (m *RecommenderMock) Get(spec *apiv1.PodSpec) (*vpa_types.RecommendedPodResources, error) {
	args := m.Called(spec)
	var returnArg *vpa_types.RecommendedPodResources
	if args.Get(0) != nil {
		returnArg = args.Get(0).(*vpa_types.RecommendedPodResources)
	}
	return returnArg, args.Error(1)
}

// PodsEvictionRestrictionMock is a mock of PodsEvictionRestriction
type PodsEvictionRestrictionMock struct {
	mock.Mock
}

// Evict is a mock implementation of PodsEvictionRestriction.Evict
func (m *PodsEvictionRestrictionMock) Evict(pod *apiv1.Pod) error {
	args := m.Called(pod)
	return args.Error(0)
}

// CanEvict is a mock implementation of PodsEvictionRestriction.CanEvict
func (m *PodsEvictionRestrictionMock) CanEvict(pod *apiv1.Pod) bool {
	args := m.Called(pod)
	return args.Bool(0)
}

// PodListerMock is a mock of PodLister
type PodListerMock struct {
	mock.Mock
}

// Pods is a mock implementation of PodLister.Pods
func (m *PodListerMock) Pods(namespace string) v1.PodNamespaceLister {
	args := m.Called(namespace)
	var returnArg v1.PodNamespaceLister
	if args.Get(0) != nil {
		returnArg = args.Get(0).(v1.PodNamespaceLister)
	}
	return returnArg
}

// List is a mock implementation of PodLister.List
func (m *PodListerMock) List(selector labels.Selector) (ret []*apiv1.Pod, err error) {
	args := m.Called()
	var returnArg []*apiv1.Pod
	if args.Get(0) != nil {
		returnArg = args.Get(0).([]*apiv1.Pod)
	}
	return returnArg, args.Error(1)
}

// Get is not implemented for this mock
func (m *PodListerMock) Get(name string) (*apiv1.Pod, error) {
	return nil, fmt.Errorf("unimplemented")
}

// VerticalPodAutoscalerListerMock is a mock of VerticalPodAutoscalerLister or
// VerticalPodAutoscalerNamespaceLister - the crucial List method is the same.
type VerticalPodAutoscalerListerMock struct {
	mock.Mock
}

// List is a mock implementation of VerticalPodAutoscalerLister.List
func (m *VerticalPodAutoscalerListerMock) List(selector labels.Selector) (ret []*vpa_types.VerticalPodAutoscaler, err error) {
	args := m.Called()
	var returnArg []*vpa_types.VerticalPodAutoscaler
	if args.Get(0) != nil {
		returnArg = args.Get(0).([]*vpa_types.VerticalPodAutoscaler)
	}
	return returnArg, args.Error(1)
}

// VerticalPodAutoscalers is a mock implementation of returning a lister for namespace.
func (m *VerticalPodAutoscalerListerMock) VerticalPodAutoscalers(namespace string) vpa_lister.VerticalPodAutoscalerNamespaceLister {
	args := m.Called(namespace)
	var returnArg vpa_lister.VerticalPodAutoscalerNamespaceLister
	if args.Get(0) != nil {
		returnArg = args.Get(0).(vpa_lister.VerticalPodAutoscalerNamespaceLister)
	}
	return returnArg
}

// Get is not implemented for this mock
func (m *VerticalPodAutoscalerListerMock) Get(name string) (*vpa_types.VerticalPodAutoscaler, error) {
	return nil, fmt.Errorf("unimplemented")
}
