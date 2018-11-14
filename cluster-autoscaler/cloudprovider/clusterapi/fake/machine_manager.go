/*
Copyright 2016 The Kubernetes Authors.

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

package fake

import (
	"github.com/stretchr/testify/mock"
	"k8s.io/api/core/v1"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

// MachineManagerMock mocks for MachineManager
type MachineManagerMock struct {
	mock.Mock
}

// AllDeployments returns all MachineDeployments of the cluster
func (m *MachineManagerMock) AllDeployments() []*v1alpha1.MachineDeployment {
	args := m.Called()
	return args.Get(0).([]*v1alpha1.MachineDeployment)
}

// DeploymentForNode returns the MachineDeployment that created a specific node
func (m *MachineManagerMock) DeploymentForNode(node *v1.Node) *v1alpha1.MachineDeployment {
	args := m.Called(node)
	return args.Get(0).(*v1alpha1.MachineDeployment)
}

// NodesForDeployment returns all nodes that were created by a specific MachineDeployment
func (m *MachineManagerMock) NodesForDeployment(md *v1alpha1.MachineDeployment) []*v1.Node {
	args := m.Called(md)
	return args.Get(0).([]*v1.Node)
}

// SetDeploymentSize sets a MachineDeployment's replica count
func (m *MachineManagerMock) SetDeploymentSize(md *v1alpha1.MachineDeployment, size int) error {
	args := m.Called(md, size)
	return args.Error(0)
}

// Refresh reloads the ClusterapiMachineManager's cached representation of the cluster state
func (m *MachineManagerMock) Refresh() error {
	args := m.Called()
	return args.Error(0)
}
