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

package provider

import (
	reflect "reflect"

	network "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"
	gomock "go.uber.org/mock/gomock"
	v1 "k8s.io/api/core/v1"
)

// MockBackendPool is a mock of BackendPool interface.
type MockBackendPool struct {
	ctrl     *gomock.Controller
	recorder *MockBackendPoolMockRecorder
}

// MockBackendPoolMockRecorder is the mock recorder for MockBackendPool.
type MockBackendPoolMockRecorder struct {
	mock *MockBackendPool
}

// NewMockBackendPool creates a new mock instance.
func NewMockBackendPool(ctrl *gomock.Controller) *MockBackendPool {
	mock := &MockBackendPool{ctrl: ctrl}
	mock.recorder = &MockBackendPoolMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBackendPool) EXPECT() *MockBackendPoolMockRecorder {
	return m.recorder
}

// CleanupVMSetFromBackendPoolByCondition mocks base method.
func (m *MockBackendPool) CleanupVMSetFromBackendPoolByCondition(slb *network.LoadBalancer, service *v1.Service, nodes []*v1.Node, clusterName string, shouldRemoveVMSetFromSLB func(string) bool) (*network.LoadBalancer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CleanupVMSetFromBackendPoolByCondition", slb, service, nodes, clusterName, shouldRemoveVMSetFromSLB)
	ret0, _ := ret[0].(*network.LoadBalancer)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CleanupVMSetFromBackendPoolByCondition indicates an expected call of CleanupVMSetFromBackendPoolByCondition.
func (mr *MockBackendPoolMockRecorder) CleanupVMSetFromBackendPoolByCondition(slb, service, nodes, clusterName, shouldRemoveVMSetFromSLB interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CleanupVMSetFromBackendPoolByCondition", reflect.TypeOf((*MockBackendPool)(nil).CleanupVMSetFromBackendPoolByCondition), slb, service, nodes, clusterName, shouldRemoveVMSetFromSLB)
}

// EnsureHostsInPool mocks base method.
func (m *MockBackendPool) EnsureHostsInPool(service *v1.Service, nodes []*v1.Node, backendPoolID, vmSetName, clusterName, lbName string, backendPool network.BackendAddressPool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EnsureHostsInPool", service, nodes, backendPoolID, vmSetName, clusterName, lbName, backendPool)
	ret0, _ := ret[0].(error)
	return ret0
}

// EnsureHostsInPool indicates an expected call of EnsureHostsInPool.
func (mr *MockBackendPoolMockRecorder) EnsureHostsInPool(service, nodes, backendPoolID, vmSetName, clusterName, lbName, backendPool interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EnsureHostsInPool", reflect.TypeOf((*MockBackendPool)(nil).EnsureHostsInPool), service, nodes, backendPoolID, vmSetName, clusterName, lbName, backendPool)
}

// GetBackendPrivateIPs mocks base method.
func (m *MockBackendPool) GetBackendPrivateIPs(clusterName string, service *v1.Service, lb *network.LoadBalancer) ([]string, []string) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBackendPrivateIPs", clusterName, service, lb)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].([]string)
	return ret0, ret1
}

// GetBackendPrivateIPs indicates an expected call of GetBackendPrivateIPs.
func (mr *MockBackendPoolMockRecorder) GetBackendPrivateIPs(clusterName, service, lb interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBackendPrivateIPs", reflect.TypeOf((*MockBackendPool)(nil).GetBackendPrivateIPs), clusterName, service, lb)
}

// ReconcileBackendPools mocks base method.
func (m *MockBackendPool) ReconcileBackendPools(clusterName string, service *v1.Service, lb *network.LoadBalancer) (bool, bool, *network.LoadBalancer, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReconcileBackendPools", clusterName, service, lb)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(*network.LoadBalancer)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// ReconcileBackendPools indicates an expected call of ReconcileBackendPools.
func (mr *MockBackendPoolMockRecorder) ReconcileBackendPools(clusterName, service, lb interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReconcileBackendPools", reflect.TypeOf((*MockBackendPool)(nil).ReconcileBackendPools), clusterName, service, lb)
}
