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

package clusterapi

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/clusterapi/fake"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"testing"
)

func newTestMachineManager(t *testing.T) *fake.MachineManagerMock {
	manager := new(fake.MachineManagerMock)

	return manager
}

func newTestProvider(t *testing.T) *ClusterapiCloudProvider {
	manager := newTestMachineManager(t)
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})

	return &ClusterapiCloudProvider{
		resourceLimiter: resourceLimiter,
		machineManager:  manager,
	}
}

func TestBuildClusterapiCloudProvider(t *testing.T) {
	machineManager := newTestMachineManager(t)
	machineManager.On("Refresh").Return(nil)

	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	_, err := BuildClusterapiCloudProvider(machineManager, resourceLimiter)

	assert.NoError(t, err)
	machineManager.AssertExpectations(t)
}

func TestName(t *testing.T) {
	provider := newTestProvider(t)
	assert.Equal(t, provider.Name(), "clusterapi")
}

func TestNodeGroups(t *testing.T) {
	md1 := buildTestMachineDeployment("md1", 1, 0, 10)
	md2 := buildTestMachineDeployment("md2", 2, 0, 10)

	machineManager := newTestMachineManager(t)
	machineManager.On("Refresh").Return(nil)
	machineManager.On("AllDeployments").Return([]*v1alpha1.MachineDeployment{md1, md2})

	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	cp, err := BuildClusterapiCloudProvider(machineManager, resourceLimiter)
	assert.NoError(t, err)

	nodeGroups := cp.NodeGroups()
	assert.Len(t, nodeGroups, 2)
	assert.Equal(t, "md1", nodeGroups[0].Id())
	assert.Equal(t, "md2", nodeGroups[1].Id())

	machineManager.AssertExpectations(t)
}

func TestNodeGroupForNodeAndNodesOfNodeGroup(t *testing.T) {
	md1 := buildTestMachineDeployment("md1", 1, 0, 10)
	ms1 := buildTestMachineSet(md1, "ms1", 2)

	n11 := buildTestNode("n11")
	buildTestMachine(ms1, "m11", n11)

	n12 := buildTestNode("n12")
	buildTestMachine(ms1, "m12", n12)

	md2 := buildTestMachineDeployment("md2", 2, 0, 10)
	buildTestMachineSet(md2, "ms2", 2)

	n21 := buildTestNode("n21")
	buildTestMachine(ms1, "m21", n21)

	machineManager := newTestMachineManager(t)
	machineManager.On("Refresh").Return(nil)

	machineManager.On("DeploymentForNode", n11).Return(md1)
	machineManager.On("DeploymentForNode", n12).Return(md1)
	machineManager.On("NodesForDeployment", md1).Return([]*v1.Node{n11, n12})

	machineManager.On("DeploymentForNode", n21).Return(md2)
	machineManager.On("NodesForDeployment", md2).Return([]*v1.Node{n21})

	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	cp, err := BuildClusterapiCloudProvider(machineManager, resourceLimiter)
	assert.NoError(t, err)

	nodeGroup, err := cp.NodeGroupForNode(n11)
	assert.NoError(t, err)
	assert.Equal(t, "md1", nodeGroup.Id())

	nodeGroup, err = cp.NodeGroupForNode(n12)
	assert.NoError(t, err)
	assert.Equal(t, "md1", nodeGroup.Id())

	nodes, err := nodeGroup.Nodes()
	assert.NoError(t, err)
	assert.Equal(t, []cloudprovider.Instance{{Id: "n11"}, {Id: "n12"}}, nodes)

	nodeGroup, err = cp.NodeGroupForNode(n21)
	assert.NoError(t, err)
	assert.Equal(t, "md2", nodeGroup.Id())

	nodes, err = nodeGroup.Nodes()
	assert.NoError(t, err)
	assert.Equal(t, []cloudprovider.Instance{{Id: "n21"}}, nodes)

	machineManager.AssertExpectations(t)
}

func TestRefresh(t *testing.T) {
	machineManager := newTestMachineManager(t)
	machineManager.On("Refresh").Return(nil)

	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	cp, err := BuildClusterapiCloudProvider(machineManager, resourceLimiter)
	assert.NoError(t, err)

	assert.NoError(t, cp.Refresh())

	machineManager.AssertExpectations(t)
	machineManager.AssertNumberOfCalls(t, "Refresh", 2)
}
