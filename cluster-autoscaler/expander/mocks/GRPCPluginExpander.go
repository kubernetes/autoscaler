/*
Copyright 2021 The Kubernetes Authors.

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

package mocks

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	grpc "google.golang.org/grpc"
	"k8s.io/autoscaler/cluster-autoscaler/expander/grpcplugin/protos"
)

// MockExpanderClient is a mock of ExpanderClient interface.
type MockExpanderClient struct {
	ctrl     *gomock.Controller
	recorder *MockExpanderClientMockRecorder
}

// MockExpanderClientMockRecorder is the mock recorder for MockExpanderClient.
type MockExpanderClientMockRecorder struct {
	mock *MockExpanderClient
}

// NewMockExpanderClient creates a new mock instance.
func NewMockExpanderClient(ctrl *gomock.Controller) *MockExpanderClient {
	mock := &MockExpanderClient{ctrl: ctrl}
	mock.recorder = &MockExpanderClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExpanderClient) EXPECT() *MockExpanderClientMockRecorder {
	return m.recorder
}

// BestOptions mocks base method.
func (m *MockExpanderClient) BestOptions(ctx context.Context, in *protos.BestOptionsRequest, opts ...grpc.CallOption) (*protos.BestOptionsResponse, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, in}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "BestOptions", varargs...)
	ret0, _ := ret[0].(*protos.BestOptionsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BestOptions indicates an expected call of BestOptions.
func (mr *MockExpanderClientMockRecorder) BestOptions(ctx, in interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, in}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BestOptions", reflect.TypeOf((*MockExpanderClient)(nil).BestOptions), varargs...)
}

// MockExpanderServer is a mock of ExpanderServer interface.
type MockExpanderServer struct {
	ctrl     *gomock.Controller
	recorder *MockExpanderServerMockRecorder
}

// MockExpanderServerMockRecorder is the mock recorder for MockExpanderServer.
type MockExpanderServerMockRecorder struct {
	mock *MockExpanderServer
}

// NewMockExpanderServer creates a new mock instance.
func NewMockExpanderServer(ctrl *gomock.Controller) *MockExpanderServer {
	mock := &MockExpanderServer{ctrl: ctrl}
	mock.recorder = &MockExpanderServerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExpanderServer) EXPECT() *MockExpanderServerMockRecorder {
	return m.recorder
}

// BestOptions mocks base method.
func (m *MockExpanderServer) BestOptions(arg0 context.Context, arg1 *protos.BestOptionsRequest) (*protos.BestOptionsResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BestOptions", arg0, arg1)
	ret0, _ := ret[0].(*protos.BestOptionsResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BestOptions indicates an expected call of BestOptions.
func (mr *MockExpanderServerMockRecorder) BestOptions(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BestOptions", reflect.TypeOf((*MockExpanderServer)(nil).BestOptions), arg0, arg1)
}
