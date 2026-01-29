/*
Copyright 2025 The Kubernetes Authors.

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
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	gomock "go.uber.org/mock/gomock"
)

// MockStorageAccountClient is a mock of storage account client interface.
type MockStorageAccountClient struct {
	ctrl     *gomock.Controller
	recorder *MockStorageAccountClientMockRecorder
}

// MockStorageAccountClientMockRecorder is the mock recorder for MockStorageAccountClient.
type MockStorageAccountClientMockRecorder struct {
	mock *MockStorageAccountClient
}

// NewMockStorageAccountClient creates a new mock instance.
func NewMockStorageAccountClient(ctrl *gomock.Controller) *MockStorageAccountClient {
	mock := &MockStorageAccountClient{ctrl: ctrl}
	mock.recorder = &MockStorageAccountClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockStorageAccountClient) EXPECT() *MockStorageAccountClientMockRecorder {
	return m.recorder
}

// ListKeys mocks base method.
func (m *MockStorageAccountClient) ListKeys(ctx context.Context, resourceGroupName, accountName string) ([]*armstorage.AccountKey, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListKeys", ctx, resourceGroupName, accountName)
	ret0, _ := ret[0].([]*armstorage.AccountKey)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListKeys indicates an expected call of ListKeys.
func (mr *MockStorageAccountClientMockRecorder) ListKeys(ctx, resourceGroupName, accountName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListKeys", nil, ctx, resourceGroupName, accountName)
}

// Create mocks base method.
func (m *MockStorageAccountClient) Create(ctx context.Context, resourceGroupName, accountName string, resource *armstorage.AccountCreateParameters) (*armstorage.Account, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, resourceGroupName, accountName, resource)
	ret0, _ := ret[0].(*armstorage.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockStorageAccountClientMockRecorder) Create(ctx, resourceGroupName, accountName, resource any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", nil, ctx, resourceGroupName, accountName, resource)
}

// GetProperties mocks base method.
func (m *MockStorageAccountClient) GetProperties(ctx context.Context, resourceGroupName, accountName string, options *armstorage.AccountsClientGetPropertiesOptions) (*armstorage.Account, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProperties", ctx, resourceGroupName, accountName, options)
	ret0, _ := ret[0].(*armstorage.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProperties indicates an expected call of GetProperties.
func (mr *MockStorageAccountClientMockRecorder) GetProperties(ctx, resourceGroupName, accountName, options any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProperties", nil, ctx, resourceGroupName, accountName, options)
}

// Delete mocks base method.
func (m *MockStorageAccountClient) Delete(ctx context.Context, resourceGroupName, accountName string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, resourceGroupName, accountName)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockStorageAccountClientMockRecorder) Delete(ctx, resourceGroupName, accountName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", nil, ctx, resourceGroupName, accountName)
}

// List mocks base method.
func (m *MockStorageAccountClient) List(ctx context.Context, resourceGroupName string) ([]*armstorage.Account, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx, resourceGroupName)
	ret0, _ := ret[0].([]*armstorage.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockStorageAccountClientMockRecorder) List(ctx, resourceGroupName any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", nil, ctx, resourceGroupName)
}

// Update mocks base method.
func (m *MockStorageAccountClient) Update(ctx context.Context, resourceGroupName, accountName string, parameters *armstorage.AccountUpdateParameters) (*armstorage.Account, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, resourceGroupName, accountName, parameters)
	ret0, _ := ret[0].(*armstorage.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Update indicates an expected call of Update.
func (mr *MockStorageAccountClientMockRecorder) Update(ctx, resourceGroupName, accountName, parameters any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", nil, ctx, resourceGroupName, accountName, parameters)
}
