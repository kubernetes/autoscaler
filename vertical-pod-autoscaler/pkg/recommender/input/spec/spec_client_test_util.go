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
	"fmt"

	"github.com/stretchr/testify/mock"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	v1lister "k8s.io/client-go/listers/core/v1"
)

var scheme = runtime.NewScheme()
var codecs = serializer.NewCodecFactory(scheme)

func init() {
	v1.AddToScheme(scheme)
}

const pod1Yaml = `
apiVersion: v1
kind: Pod
metadata:
  name: Pod1
  labels:
    Pod1LabelKey: Pod1LabelValue
spec:
  containers:
  - name: Name11
    image: Name11Image
    resources:
      requests:
        memory: "512Mi"
        cpu: "500m"
  - name: Name12
    image: Name12Image
    resources:
      requests:
        memory: "1024Mi"
        cpu: "1000m"
`

const pod2Yaml = `
apiVersion: v1
kind: Pod
metadata:
  name: Pod2
  labels:
    Pod2LabelKey: Pod2LabelValue
spec:
  containers:
  - name: Name21
    image: Name21Image
    resources:
      requests:
        memory: "2048Mi"
        cpu: "2000m"
  - name: Name22
    image: Name22Image
    resources:
      requests:
        memory: "4096Mi"
        cpu: "4000m"
`

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
	podSpecs []*BasicPodSpec
	podYamls []string
}

func newEmptySpecClientTestCase() *specClientTestCase {
	return &specClientTestCase{}
}

func newSpecClientTestCase() *specClientTestCase {
	podID1 := model.PodID{Namespace: "", PodName: "Pod1"}
	podID2 := model.PodID{Namespace: "", PodName: "Pod2"}

	containerSpec11 := newTestContainerSpec(podID1, "Name11", 500, 512*1024*1024)
	containerSpec12 := newTestContainerSpec(podID1, "Name12", 1000, 1024*1024*1024)
	containerSpec21 := newTestContainerSpec(podID2, "Name21", 2000, 2048*1024*1024)
	containerSpec22 := newTestContainerSpec(podID2, "Name22", 4000, 4096*1024*1024)

	podSpec1 := newTestPodSpec(podID1, containerSpec11, containerSpec12)
	podSpec2 := newTestPodSpec(podID2, containerSpec21, containerSpec22)

	return &specClientTestCase{
		podSpecs: []*BasicPodSpec{podSpec1, podSpec2},
		podYamls: []string{pod1Yaml, pod2Yaml},
	}
}

func newTestContainerSpec(podID model.PodID, containerName string, milicores int, memory int64) BasicContainerSpec {
	containerID := model.ContainerID{
		PodID:         podID,
		ContainerName: containerName,
	}
	requestedResources := model.Resources{
		model.ResourceCPU:    model.ResourceAmount(milicores),
		model.ResourceMemory: model.ResourceAmount(memory),
	}
	return BasicContainerSpec{
		ID:      containerID,
		Image:   containerName + "Image",
		Request: requestedResources,
	}
}

func newTestPodSpec(podId model.PodID, containerSpecs ...BasicContainerSpec) *BasicPodSpec {
	return &BasicPodSpec{
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
	for _, yaml := range tc.podYamls {
		pods = append(pods, newPod(yaml))
	}
	return pods
}

func newPod(yaml string) *v1.Pod {
	decode := codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(yaml), nil, nil)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	return obj.(*v1.Pod)
}
