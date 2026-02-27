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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	listersv1 "k8s.io/client-go/listers/core/v1"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

var scheme = runtime.NewScheme()
var codecs = serializer.NewCodecFactory(scheme)

func init() {
	utilruntime.Must(corev1.AddToScheme(scheme))
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
status:
  containerStatuses:
  - name: Name23
    resources:
      requests:
        memory: "250Mi"
        cpu: "30m"
  initContainerStatuses:
  - name: Name22-init
    resources:
      requests:
        memory: "350Mi"
        cpu: "40m"
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
  - name: Name23
    image: Name23Image
    resources:
      # Requests below will be ignored because
      # requests are also defined in containerStatus.
      requests:
        memory: "1Mi"
        cpu: "1m"
  initContainers:
  - name: Name21-init
    image: Name21-initImage
    resources:
      requests:
        memory: "128Mi"
        cpu: "40m"
  - name: Name22-init
    image: Name22-initImage
    resources:
      requests:
        # Requests below will be ignored because
        # requests are also defined in initContainerStatus.
        memory: "1Mi"
        cpu: "1m"
`

type podListerMock struct {
	mock.Mock
}

func (m *podListerMock) List(selector labels.Selector) (ret []*corev1.Pod, err error) {
	args := m.Called()
	return args.Get(0).([]*corev1.Pod), args.Error(1)
}

func (m *podListerMock) Pods(namespace string) listersv1.PodNamespaceLister {
	args := m.Called()
	return args.Get(0).(listersv1.PodNamespaceLister)
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
	containerSpec23 := newTestContainerSpec(podID2, "Name23", 30, 250*1024*1024)

	initContainerSpec21 := newTestContainerSpec(podID2, "Name21-init", 40, 128*1024*1024)
	initContainerSpec22 := newTestContainerSpec(podID2, "Name22-init", 40, 350*1024*1024)

	podSpec1 := newTestPodSpec(podID1, []BasicContainerSpec{containerSpec11, containerSpec12}, nil)
	podSpec2 := newTestPodSpec(podID2, []BasicContainerSpec{containerSpec21, containerSpec22, containerSpec23}, []BasicContainerSpec{initContainerSpec21, initContainerSpec22})

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

func newTestPodSpec(podId model.PodID, containerSpecs []BasicContainerSpec, initContainerSpecs []BasicContainerSpec) *BasicPodSpec {
	return &BasicPodSpec{
		ID:             podId,
		PodLabels:      map[string]string{podId.PodName + "LabelKey": podId.PodName + "LabelValue"},
		Containers:     containerSpecs,
		InitContainers: initContainerSpecs,
	}
}

const nativeSidecarPodYaml = `
apiVersion: v1
kind: Pod
metadata:
  name: PodWithSidecar
  labels:
    SidecarLabelKey: SidecarLabelValue
spec:
  containers:
  - name: main
    image: mainImage
    resources:
      requests:
        memory: "512Mi"
        cpu: "500m"
  initContainers:
  - name: sidecar
    image: sidecarImage
    restartPolicy: Always
    resources:
      requests:
        memory: "128Mi"
        cpu: "100m"
  - name: regular-init
    image: regularInitImage
    resources:
      requests:
        memory: "64Mi"
        cpu: "50m"
`

func newNativeSidecarSpecClientTestCase() *specClientTestCase {
	podID := model.PodID{Namespace: "", PodName: "PodWithSidecar"}

	containerSpec := newTestContainerSpec(podID, "main", 500, 512*1024*1024)
	sidecarSpec := BasicContainerSpec{
		ID: model.ContainerID{
			PodID:         podID,
			ContainerName: "sidecar",
		},
		Image:           "sidecarImage",
		Request:         model.Resources{model.ResourceCPU: model.ResourceAmount(100), model.ResourceMemory: model.ResourceAmount(128 * 1024 * 1024)},
		IsNativeSidecar: true,
	}
	regularInitSpec := newTestContainerSpec(podID, "regular-init", 50, 64*1024*1024)

	podSpec := &BasicPodSpec{
		ID:             podID,
		PodLabels:      map[string]string{"SidecarLabelKey": "SidecarLabelValue"},
		Containers:     []BasicContainerSpec{containerSpec},
		InitContainers: []BasicContainerSpec{sidecarSpec, regularInitSpec},
	}

	return &specClientTestCase{
		podSpecs: []*BasicPodSpec{podSpec},
		podYamls: []string{nativeSidecarPodYaml},
	}
}

func (tc *specClientTestCase) createFakeSpecClient() SpecClient {
	podListerMock := new(podListerMock)
	podListerMock.On("List").Return(tc.getFakePods(), nil)

	return NewSpecClient(podListerMock)
}

func (tc *specClientTestCase) getFakePods() []*corev1.Pod {
	pods := []*corev1.Pod{}
	for _, yaml := range tc.podYamls {
		pods = append(pods, newPod(yaml))
	}
	return pods
}

func newPod(yaml string) *corev1.Pod {
	decode := codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(yaml), nil, nil)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	return obj.(*corev1.Pod)
}
