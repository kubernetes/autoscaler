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
	"time"

	"github.com/stretchr/testify/mock"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/record"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_types_v1beta1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1beta1"
	vpa_lister "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1"
	vpa_lister_v1beta1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/listers/autoscaling.k8s.io/v1beta1"
)

var (
	timeLayout       = "2006-01-02 15:04:05"
	testTimestamp, _ = time.Parse(timeLayout, "2017-04-18 17:35:05")
)

// BuildTestPolicy creates ResourcesPolicy with specified constraints
func BuildTestPolicy(containerName, minCPU, maxCPU, minMemory, maxMemory string) *vpa_types.PodResourcePolicy {
	minCPUVal, _ := resource.ParseQuantity(minCPU)
	maxCPUVal, _ := resource.ParseQuantity(maxCPU)
	minMemVal, _ := resource.ParseQuantity(minMemory)
	maxMemVal, _ := resource.ParseQuantity(maxMemory)
	return &vpa_types.PodResourcePolicy{ContainerPolicies: []vpa_types.ContainerResourcePolicy{{
		ContainerName: containerName,
		MinAllowed: apiv1.ResourceList{
			apiv1.ResourceMemory: minMemVal,
			apiv1.ResourceCPU:    minCPUVal,
		},
		MaxAllowed: apiv1.ResourceList{
			apiv1.ResourceMemory: maxMemVal,
			apiv1.ResourceCPU:    maxCPUVal,
		},
	},
	}}
}

// Resources creates a ResourceList with given amount of cpu and memory.
func Resources(cpu, mem string) apiv1.ResourceList {
	result := make(apiv1.ResourceList)
	if len(cpu) > 0 {
		cpuVal, _ := resource.ParseQuantity(cpu)
		result[apiv1.ResourceCPU] = cpuVal
	}
	if len(mem) > 0 {
		memVal, _ := resource.ParseQuantity(mem)
		result[apiv1.ResourceMemory] = memVal
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
func (m *PodsEvictionRestrictionMock) Evict(pod *apiv1.Pod, vpa *vpa_types.VerticalPodAutoscaler, eventRecorder record.EventRecorder) error {
	args := m.Called(pod, eventRecorder)
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

// VerticalPodAutoscalerV1Beta1ListerMock is a mock of VerticalPodAutoscalerLister or
// VerticalPodAutoscalerNamespaceLister - the crucial List method is the same.
type VerticalPodAutoscalerV1Beta1ListerMock struct {
	mock.Mock
}

// List is a mock implementation of VerticalPodAutoscalerLister.List
func (m *VerticalPodAutoscalerV1Beta1ListerMock) List(selector labels.Selector) (ret []*vpa_types_v1beta1.VerticalPodAutoscaler, err error) {
	args := m.Called()
	var returnArg []*vpa_types_v1beta1.VerticalPodAutoscaler
	if args.Get(0) != nil {
		returnArg = args.Get(0).([]*vpa_types_v1beta1.VerticalPodAutoscaler)
	}
	return returnArg, args.Error(1)
}

// VerticalPodAutoscalers is a mock implementation of returning a lister for namespace.
func (m *VerticalPodAutoscalerV1Beta1ListerMock) VerticalPodAutoscalers(namespace string) vpa_lister_v1beta1.VerticalPodAutoscalerNamespaceLister {
	args := m.Called(namespace)
	var returnArg vpa_lister_v1beta1.VerticalPodAutoscalerNamespaceLister
	if args.Get(0) != nil {
		returnArg = args.Get(0).(vpa_lister_v1beta1.VerticalPodAutoscalerNamespaceLister)
	}
	return returnArg
}

// Get is not implemented for this mock
func (m *VerticalPodAutoscalerV1Beta1ListerMock) Get(name string) (*vpa_types_v1beta1.VerticalPodAutoscaler, error) {
	return nil, fmt.Errorf("unimplemented")
}

// RecommendationProcessorMock is mock implementation of RecommendationProcessor
type RecommendationProcessorMock struct {
	mock.Mock
}

// Apply is a mock implementation of RecommendationProcessor.Apply
func (m *RecommendationProcessorMock) Apply(vpa *vpa_types.VerticalPodAutoscaler,
	pod *apiv1.Pod) (*vpa_types.RecommendedPodResources, map[string][]string, error) {
	args := m.Called()
	var returnArg *vpa_types.RecommendedPodResources
	if args.Get(0) != nil {
		returnArg = args.Get(0).(*vpa_types.RecommendedPodResources)
	}
	var annotations map[string][]string
	if args.Get(1) != nil {
		annotations = args.Get(1).(map[string][]string)
	}
	return returnArg, annotations, args.Error(1)
}

// FakeRecommendationProcessor is a dummy implementation of RecommendationProcessor
type FakeRecommendationProcessor struct{}

// Apply is a dummy implementation of RecommendationProcessor.Apply which returns provided podRecommendation
func (f *FakeRecommendationProcessor) Apply(vpa *vpa_types.VerticalPodAutoscaler,
	pod *apiv1.Pod) (*vpa_types.RecommendedPodResources, map[string][]string, error) {
	return vpa.Status.Recommendation, nil, nil
}

// fakeEventRecorder is a dummy implementation of record.EventRecorder.
type fakeEventRecorder struct{}

// Event is a dummy implementation of record.EventRecorder interface.
func (f *fakeEventRecorder) Event(object runtime.Object, eventtype, reason, message string) {}

// Eventf is a dummy implementation of record.EventRecorder interface.
func (f *fakeEventRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
}

// PastEventf is a dummy implementation of record.EventRecorder interface.
func (f *fakeEventRecorder) PastEventf(object runtime.Object, timestamp metav1.Time, eventtype, reason, messageFmt string, args ...interface{}) {
}

// AnnotatedEventf is a dummy implementation of record.EventRecorder interface.
func (f *fakeEventRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
}

// FakeEventRecorder returns a dummy implementation of record.EventRecorder.
func FakeEventRecorder() record.EventRecorder {
	return &fakeEventRecorder{}
}

// mockedEventRecorder is a dummy implementation of record.EventRecorder.
type mockedEventRecorder struct {
	mock.Mock
}

// Event is a mocked implementation of record.EventRecorder interface.
func (m *mockedEventRecorder) Event(object runtime.Object, eventtype, reason, message string) {
	m.Called(object, eventtype, reason, message)
}

// Eventf is a dummy implementation of record.EventRecorder interface.
func (m *mockedEventRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
}

// PastEventf is a dummy implementation of record.EventRecorder interface.
func (m *mockedEventRecorder) PastEventf(object runtime.Object, timestamp metav1.Time, eventtype, reason, messageFmt string, args ...interface{}) {
}

// AnnotatedEventf is a dummy implementation of record.EventRecorder interface.
func (m *mockedEventRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
}

// MockEventRecorder returns a dummy implementation of record.EventRecorder.
func MockEventRecorder() *mockedEventRecorder {
	return &mockedEventRecorder{}
}
