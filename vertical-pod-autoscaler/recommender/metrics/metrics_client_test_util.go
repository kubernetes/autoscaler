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

package metrics

import (
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	clientapiv1 "k8s.io/client-go/pkg/api/v1"
	core "k8s.io/client-go/testing"
	"k8s.io/kubernetes/pkg/api/v1"
	v1lister "k8s.io/kubernetes/pkg/client/listers/core/v1"
	metricsapi "k8s.io/metrics/pkg/apis/metrics/v1alpha1"
	"k8s.io/metrics/pkg/client/clientset_generated/clientset/fake"
	"time"
)

type PodListerMock struct {
	mock.Mock
}

func (m *PodListerMock) List(selector labels.Selector) (ret []*v1.Pod, err error) {
	args := m.Called()
	return args.Get(0).([]*v1.Pod), args.Error(1)
}

func (m *PodListerMock) Pods(namespace string) v1lister.PodNamespaceLister {
	args := m.Called()
	return args.Get(0).(v1lister.PodNamespaceLister)
}

type NamespaceListerMock struct {
	mock.Mock
}

func (m *NamespaceListerMock) List(selector labels.Selector) (ret []*v1.Namespace, err error) {
	args := m.Called()
	return args.Get(0).([]*v1.Namespace), args.Error(1)
}

func (m *NamespaceListerMock) Get(name string) (*v1.Namespace, error) {
	args := m.Called()
	return args.Get(0).(*v1.Namespace), args.Error(1)
}

type metricsClientTestCase struct {
	podCreationTimestamp time.Time
	snapshotTimestamp    time.Time
	snapshotWindow       time.Duration
	namespace            *v1.Namespace
	pod1Snaps, pod2Snaps []*ContainerUtilizationSnapshot
}

func newMetricsClientTestCase() *metricsClientTestCase {
	namespaceName := "test-namespace"

	testCase := &metricsClientTestCase{
		podCreationTimestamp: time.Now().AddDate(0, 0, -1),
		snapshotTimestamp:    time.Now(),
		snapshotWindow:       time.Duration(1234),
		namespace:            &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespaceName}},
	}

	id1 := containerId{Namespace: namespaceName, PodName: "Pod1", ContainerName: "Name1"}
	id2 := containerId{Namespace: namespaceName, PodName: "Pod1", ContainerName: "Name2"}
	id3 := containerId{Namespace: namespaceName, PodName: "Pod2", ContainerName: "Name1"}
	id4 := containerId{Namespace: namespaceName, PodName: "Pod2", ContainerName: "Name2"}

	testCase.pod1Snaps = append(testCase.pod1Snaps, testCase.newContainerUtilizationSnapshot(id1, 500, 512, 400, 333))
	testCase.pod1Snaps = append(testCase.pod1Snaps, testCase.newContainerUtilizationSnapshot(id2, 1000, 1024, 800, 666))
	testCase.pod2Snaps = append(testCase.pod2Snaps, testCase.newContainerUtilizationSnapshot(id3, 500, 512, 401, 334))
	testCase.pod2Snaps = append(testCase.pod2Snaps, testCase.newContainerUtilizationSnapshot(id4, 1000, 1024, 801, 667))

	return testCase
}

func newEmptyMetricsClientTestCase() *metricsClientTestCase {
	return &metricsClientTestCase{}
}

func (tc *metricsClientTestCase) newContainerUtilizationSnapshot(id containerId, cpuReq int64, memReq int64, cpuUsage int64, memUsage int64) *ContainerUtilizationSnapshot {
	return &ContainerUtilizationSnapshot{
		Id:             id,
		CreationTime:   tc.podCreationTimestamp,
		SnapshotTime:   tc.snapshotTimestamp,
		SnapshotWindow: tc.snapshotWindow,
		Image:          id.ContainerName + "Image",
		Request: clientapiv1.ResourceList{
			clientapiv1.ResourceCPU:    *resource.NewQuantity(cpuReq, resource.DecimalSI),
			clientapiv1.ResourceMemory: *resource.NewQuantity(memReq, resource.DecimalSI),
		},
		Usage: clientapiv1.ResourceList{
			clientapiv1.ResourceCPU:    *resource.NewQuantity(cpuUsage, resource.DecimalSI),
			clientapiv1.ResourceMemory: *resource.NewQuantity(memUsage, resource.DecimalSI),
		},
	}
}

func (tc *metricsClientTestCase) createFakeMetricsClient() MetricsClient {
	fakeMetricsGetter := &fake.Clientset{}
	fakeMetricsGetter.AddReactor("list", "podmetricses", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		return true, tc.getFakePodMetricsList(), nil
	})

	podListerMock := new(PodListerMock)
	podListerMock.On("List").Return(tc.getFakePods(), nil)

	namespaceListerMock := new(NamespaceListerMock)
	namespaceListerMock.On("List").Return(tc.getFakeNamespaces(), nil)

	return NewMetricsClient(fakeMetricsGetter.MetricsV1alpha1(), podListerMock, namespaceListerMock)
}

func (tc *metricsClientTestCase) getFakeNamespaces() []*v1.Namespace {
	namespaces := []*v1.Namespace{}

	if tc.namespace != nil {
		namespaces = append(namespaces, tc.namespace)
	}

	return namespaces
}

func (tc *metricsClientTestCase) getFakePodMetricsList() *metricsapi.PodMetricsList {
	metrics := &metricsapi.PodMetricsList{}

	if tc.pod1Snaps != nil {
		metrics.Items = append(metrics.Items, makePodMetrics(tc.pod1Snaps))
	}
	if tc.pod2Snaps != nil {
		metrics.Items = append(metrics.Items, makePodMetrics(tc.pod2Snaps))
	}

	return metrics
}

func (tc *metricsClientTestCase) getFakePods() []*v1.Pod {
	pods := []*v1.Pod{}

	if tc.pod1Snaps != nil {
		pods = append(pods, newPod(tc.pod1Snaps))
	}
	if tc.pod2Snaps != nil {
		pods = append(pods, newPod(tc.pod2Snaps))
	}

	return pods
}

func newPod(snaps []*ContainerUtilizationSnapshot) *v1.Pod {
	if len(snaps) == 0 {
		return nil
	}
	firstSnap := snaps[0]
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         firstSnap.Id.Namespace,
			Name:              firstSnap.Id.PodName,
			CreationTimestamp: metav1.Time{firstSnap.CreationTime},
		},
		Spec: v1.PodSpec{
			Containers: make([]v1.Container, len(snaps)),
		},
	}

	for i, snap := range snaps {
		pod.Spec.Containers[i] = v1.Container{
			Name:  snap.Id.ContainerName,
			Image: snap.Image,
			Resources: v1.ResourceRequirements{
				Requests: convertToKubernetesApi(snap.Request),
			},
		}
	}
	return pod
}

func makePodMetrics(snaps []*ContainerUtilizationSnapshot) metricsapi.PodMetrics {
	firstSnap := snaps[0]
	podMetrics := metricsapi.PodMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: firstSnap.Id.Namespace,
			Name:      firstSnap.Id.PodName,
		},
		Timestamp:  metav1.Time{Time: firstSnap.SnapshotTime},
		Window:     metav1.Duration{Duration: firstSnap.SnapshotWindow},
		Containers: make([]metricsapi.ContainerMetrics, len(snaps)),
	}

	for i, snap := range snaps {
		podMetrics.Containers[i] = metricsapi.ContainerMetrics{
			Name:  snap.Id.ContainerName,
			Usage: snap.Usage,
		}
	}
	return podMetrics
}

func (tc *metricsClientTestCase) getAllSnaps() []*ContainerUtilizationSnapshot {
	return append(tc.pod1Snaps, tc.pod2Snaps...)
}
