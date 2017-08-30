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
	"math/big"
	"time"

	"github.com/stretchr/testify/mock"

	"k8s.io/api/core/v1"
	k8sapiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/autoscaler/vertical-pod-autoscaler/recommender/model"
	v1lister "k8s.io/kubernetes/pkg/client/listers/core/v1"
)

type podListerMock struct {
	mock.Mock
}

func (m *podListerMock) List(selector labels.Selector) (ret []*v1.Pod, err error) {
	args := m.Called()
	return args.Get(0).([]*v1.Pod), args.Error(1)
}

func (m *podListerMock) Pods(namespace string) v1lister.PodNamespaceLister {
	args := m.Called()
	return args.Get(0).(v1lister.PodNamespaceLister)
}

type specClientTestCase struct {
	namespace *v1.Namespace
	podSpecs  []*model.BasicPodSpec
}

func newEmptySpecClientTestCase() *specClientTestCase {
	return &specClientTestCase{}
}

func newSpecClientTestCase() *specClientTestCase {
	namespaceName := "test-namespace"

	testCase := &specClientTestCase{
		namespace: &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespaceName}},
	}

	podID1 := model.PodID{Namespace: namespaceName, PodName: "Pod1"}
	podID2 := model.PodID{Namespace: namespaceName, PodName: "Pod2"}

	containerSpec11 := newTestContainerSpec(podID1, "Name11", 500, 512)
	containerSpec12 := newTestContainerSpec(podID1, "Name12", 1000, 1024)
	containerSpec21 := newTestContainerSpec(podID2, "Name21", 2000, 2048)
	containerSpec22 := newTestContainerSpec(podID2, "Name22", 4000, 4096)

	testCase.podSpecs = append(testCase.podSpecs, newTestPodSpec(podID1, containerSpec11, containerSpec12))
	testCase.podSpecs = append(testCase.podSpecs, newTestPodSpec(podID2, containerSpec21, containerSpec22))

	return testCase
}
func newTestContainerSpec(podID model.PodID, containerName string, milicores int, memory int) model.BasicContainerSpec {
	containerID := model.ContainerID{
		PodID:         podID,
		ContainerName: containerName,
	}
	requestedResources := map[model.MetricName]model.ResourceAmount{
		model.ResourceCPU:    model.ResourceAmount(milicores),
		model.ResourceMemory: model.ResourceAmount(memory),
	}
	return model.BasicContainerSpec{
		ID:      containerID,
		Image:   containerName + "Image",
		Request: requestedResources,
	}
}

func newTestPodSpec(podId model.PodID, containerSpecs ...model.BasicContainerSpec) *model.BasicPodSpec {
	return &model.BasicPodSpec{
		ID:         podId,
		PodLabels:  map[string]string{podId.PodName + "LabelKey": podId.PodName + "LabelValue"},
		Containers: containerSpecs,
	}
}

func (tc *specClientTestCase) createFakeSpecClient() SpecClient {
	podListerMock := new(podListerMock)
	podListerMock.On("List").Return(tc.getFakePods(), nil)

	return NewSpecClient(podListerMock)
}

func (tc *specClientTestCase) getFakePods() []*v1.Pod {
	pods := []*v1.Pod{}
	for _, podSpec := range tc.podSpecs {
		pods = append(pods, newPod(podSpec))
	}
	return pods
}

func newPod(podSpec *model.BasicPodSpec) *v1.Pod {

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         podSpec.ID.Namespace,
			Name:              podSpec.ID.PodName,
			CreationTimestamp: metav1.Time{time.Now()},
			Labels:            podSpec.PodLabels,
		},
		Spec: v1.PodSpec{
			Containers: make([]v1.Container, len(podSpec.Containers)),
		},
	}

	for i, containerSpec := range podSpec.Containers {
		pod.Spec.Containers[i] = v1.Container{
			Name:  containerSpec.ID.ContainerName,
			Image: containerSpec.Image,
			Resources: v1.ResourceRequirements{
				Requests: calculateResourceList(containerSpec.Request),
			},
		}
	}
	return pod
}

func calculateResourceList(usage map[model.MetricName]model.ResourceAmount) k8sapiv1.ResourceList {
	cpuCores := big.NewRat(int64(usage[model.ResourceCPU]), 1000)
	cpuQuantityString := cpuCores.FloatString(3)

	memoryBytes := big.NewInt(int64(usage[model.ResourceMemory]))
	memoryQuantityString := memoryBytes.String()

	resourceMap := map[k8sapiv1.ResourceName]resource.Quantity{
		k8sapiv1.ResourceCPU:    resource.MustParse(cpuQuantityString),
		k8sapiv1.ResourceMemory: resource.MustParse(memoryQuantityString),
	}
	return k8sapiv1.ResourceList(resourceMap)
}
