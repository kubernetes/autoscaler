/*
Copyright 2019 The Kubernetes Authors.

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

package mocktarget

import (
	gomock "github.com/golang/mock/gomock"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	labels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	mpa_types "k8s.io/autoscaler/multidimensional-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1alpha1"
	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/scale"
)

// MockVpaTargetSelectorFetcher is a mock of VpaTargetSelectorFetcher interface
type MockVpaTargetSelectorFetcher struct {
	ctrl     *gomock.Controller
	recorder *_MockVpaTargetSelectorFetcherRecorder
}

// Recorder for MockVpaTargetSelectorFetcher (not exported)
type _MockVpaTargetSelectorFetcherRecorder struct {
	mock *MockVpaTargetSelectorFetcher
}

// NewMockVpaTargetSelectorFetcher returns mock instance of a mock of VpaTargetSelectorFetcher
func NewMockVpaTargetSelectorFetcher(ctrl *gomock.Controller) *MockVpaTargetSelectorFetcher {
	mock := &MockVpaTargetSelectorFetcher{ctrl: ctrl}
	mock.recorder = &_MockVpaTargetSelectorFetcherRecorder{mock}
	return mock
}

// MockMpaTargetSelectorFetcher is a mock of MpaTargetSelectorFetcher interface
type MockMpaTargetSelectorFetcher struct {
	ctrl     *gomock.Controller
	recorder *_MockMpaTargetSelectorFetcherRecorder
	mapper   restmapper.DeferredDiscoveryRESTMapper
}

// Recorder for MockMpaTargetSelectorFetcher (not exported)
type _MockMpaTargetSelectorFetcherRecorder struct {
	mock *MockMpaTargetSelectorFetcher
}

// NewMockMpaTargetSelectorFetcher returns mock instance of a mock of MpaTargetSelectorFetcher
func NewMockMpaTargetSelectorFetcher(ctrl *gomock.Controller) *MockMpaTargetSelectorFetcher {
	mock := &MockMpaTargetSelectorFetcher{ctrl: ctrl}
	mock.recorder = &_MockMpaTargetSelectorFetcherRecorder{mock}
	return mock
}

// EXPECT enables configuring expectaions
func (_m *MockVpaTargetSelectorFetcher) EXPECT() *_MockVpaTargetSelectorFetcherRecorder {
	return _m.recorder
}

// Fetch enables configuring expectations on Fetch method
func (_m *MockVpaTargetSelectorFetcher) Fetch(vpa *vpa_types.VerticalPodAutoscaler) (labels.Selector, error) {
	ret := _m.ctrl.Call(_m, "Fetch", vpa)
	ret0, _ := ret[0].(labels.Selector)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockVpaTargetSelectorFetcherRecorder) Fetch(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Fetch", arg0)
}

// EXPECT enables configuring expectaions
func (_m *MockMpaTargetSelectorFetcher) EXPECT() *_MockMpaTargetSelectorFetcherRecorder {
	return _m.recorder
}

// Fetch enables configuring expectations on Fetch method
func (_m *MockMpaTargetSelectorFetcher) Fetch(mpa *mpa_types.MultidimPodAutoscaler) (labels.Selector, error) {
	ret := _m.ctrl.Call(_m, "Fetch", mpa)
	ret0, _ := ret[0].(labels.Selector)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockMpaTargetSelectorFetcherRecorder) Fetch(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Fetch", arg0)
}

func (_m *MockMpaTargetSelectorFetcher) GetRESTMappings(groupKind schema.GroupKind) ([]*apimeta.RESTMapping, error) {
	return _m.mapper.RESTMappings(groupKind)
}

func (_m *MockMpaTargetSelectorFetcher) Scales(namespace string) (scale.ScaleInterface) {
	return nil
}
