/*
Copyright 2020 The Kubernetes Authors.

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

package azure

import (
	context "context"
	reflect "reflect"

	runtime "github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	armcontainerservice "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
	gomock "go.uber.org/mock/gomock"
)

// MockAgentPoolsClient is a mock of AgentPoolsClient interface.
type MockAgentPoolsClient struct {
	ctrl     *gomock.Controller
	recorder *MockAgentPoolsClientMockRecorder
}

// MockAgentPoolsClientMockRecorder is the mock recorder for MockAgentPoolsClient.
type MockAgentPoolsClientMockRecorder struct {
	mock *MockAgentPoolsClient
}

// NewMockAgentPoolsClient creates a new mock instance.
func NewMockAgentPoolsClient(ctrl *gomock.Controller) *MockAgentPoolsClient {
	mock := &MockAgentPoolsClient{ctrl: ctrl}
	mock.recorder = &MockAgentPoolsClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAgentPoolsClient) EXPECT() *MockAgentPoolsClientMockRecorder {
	return m.recorder
}

// BeginCreateOrUpdate mocks base method.
func (m *MockAgentPoolsClient) BeginCreateOrUpdate(arg0 context.Context, arg1, arg2, arg3 string, arg4 armcontainerservice.AgentPool, arg5 *armcontainerservice.AgentPoolsClientBeginCreateOrUpdateOptions) (*runtime.Poller[armcontainerservice.AgentPoolsClientCreateOrUpdateResponse], error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BeginCreateOrUpdate", arg0, arg1, arg2, arg3, arg4, arg5)
	ret0, _ := ret[0].(*runtime.Poller[armcontainerservice.AgentPoolsClientCreateOrUpdateResponse])
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BeginCreateOrUpdate indicates an expected call of BeginCreateOrUpdate.
func (mr *MockAgentPoolsClientMockRecorder) BeginCreateOrUpdate(arg0, arg1, arg2, arg3, arg4, arg5 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeginCreateOrUpdate", reflect.TypeOf((*MockAgentPoolsClient)(nil).BeginCreateOrUpdate), arg0, arg1, arg2, arg3, arg4, arg5)
}

// BeginDeleteMachines mocks base method.
func (m *MockAgentPoolsClient) BeginDeleteMachines(arg0 context.Context, arg1, arg2, arg3 string, arg4 armcontainerservice.AgentPoolDeleteMachinesParameter, arg5 *armcontainerservice.AgentPoolsClientBeginDeleteMachinesOptions) (*runtime.Poller[armcontainerservice.AgentPoolsClientDeleteMachinesResponse], error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BeginDeleteMachines", arg0, arg1, arg2, arg3, arg4, arg5)
	ret0, _ := ret[0].(*runtime.Poller[armcontainerservice.AgentPoolsClientDeleteMachinesResponse])
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BeginDeleteMachines indicates an expected call of BeginDeleteMachines.
func (mr *MockAgentPoolsClientMockRecorder) BeginDeleteMachines(arg0, arg1, arg2, arg3, arg4, arg5 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BeginDeleteMachines", reflect.TypeOf((*MockAgentPoolsClient)(nil).BeginDeleteMachines), arg0, arg1, arg2, arg3, arg4, arg5)
}

// Get mocks base method.
func (m *MockAgentPoolsClient) Get(arg0 context.Context, arg1, arg2, arg3 string, arg4 *armcontainerservice.AgentPoolsClientGetOptions) (armcontainerservice.AgentPoolsClientGetResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(armcontainerservice.AgentPoolsClientGetResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockAgentPoolsClientMockRecorder) Get(arg0, arg1, arg2, arg3, arg4 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockAgentPoolsClient)(nil).Get), arg0, arg1, arg2, arg3, arg4)
}
