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

package magnum

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	apiv1 "k8s.io/api/core/v1"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/compute/v2/flavors"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/containerinfra/v1/nodegroups"
)

func (m *magnumManagerMock) getFlavorById(flavorId string) (*flavors.Flavor, error) {
	flavor := flavors.Flavor{}
	flavor.VCPUs = 2
	flavor.RAM = 2048
	flavor.Disk = 25
	flavor.Name = "s1.small"
	flavor.ID = "s1.small"
	flavor.IsPublic = true

	return &flavor, nil
}

// magnumManagerDiscoveryMock overrides magnumManagerMock's autoDiscoverNodeGroups
// to return a random set of node groups each time, so that refreshNodeGroups
// has to add/remove from the cloud providers list of node groups.
type magnumManagerDiscoveryMock struct {
	magnumManagerMock
}

// autoDiscoverNodeGroups disregards the mock return values and generates a random list of node groups.
func (m *magnumManagerDiscoveryMock) autoDiscoverNodeGroups(cfgs []magnumAutoDiscoveryConfig) ([]*nodegroups.NodeGroup, error) {
	m.Called(cfgs)
	ngs := []*nodegroups.NodeGroup{}
	two := 2
	for i := 0; i < rand.Intn(20); i++ {
		newUUID, err := uuid.NewV4()
		if err != nil {
			return nil, fmt.Errorf("failed to produce a random UUID: %v", err)
		}
		newUUIDStr := newUUID.String()
		ngs = append(ngs, &nodegroups.NodeGroup{Name: newUUIDStr, NodeCount: 1, MinNodeCount: 1, MaxNodeCount: &two})
	}
	return ngs, nil
}

// TestRefreshNodeGroupsRace checks for data races that could be caused
// by adding/removing node groups if the autoscaler uses the NodeGroups
// method to retrieve the list at the same time.
func TestRefreshNodeGroupsRace(t *testing.T) {
	manager := &magnumManagerDiscoveryMock{}
	provider := magnumCloudProvider{
		magnumManager:        manager,
		usingAutoDiscovery:   true,
		autoDiscoveryConfigs: nil,
		nodeGroupsLock:       &sync.Mutex{},
		clusterUpdateLock:    &sync.Mutex{},
	}

	manager.On("autoDiscoverNodeGroups", mock.AnythingOfType("[]magnum.magnumAutoDiscoveryConfig")).Return(nil, nil)
	manager.On("fetchNodeGroupStackIDs", mock.AnythingOfType("string")).Return(nodeGroupStacks{}, nil)

	var wg sync.WaitGroup

	startTime := time.Now()
	wg.Add(2)

	// Continuously refresh the node groups list with random node groups.
	go func() {
		for time.Since(startTime) < 2*time.Second {
			err := provider.refreshNodeGroups()
			assert.NoError(t, err)
		}
		wg.Done()
	}()

	// Continuously read from the node groups list.
	go func() {
		for time.Since(startTime) < 2*time.Second {
			ngs := provider.NodeGroups()
			assert.NotNil(t, ngs)
		}
		wg.Done()
	}()

	wg.Wait()
}

// TestNodeGroups checks that the groups returned by NodeGroups
// can be properly modified by the autoscaler (that it returns
// references, not copies).
func TestNodeGroups(t *testing.T) {
	manager := &magnumManagerMock{}

	clusterLock := &sync.Mutex{}
	provider := magnumCloudProvider{
		magnumManager:     manager,
		nodeGroupsLock:    &sync.Mutex{},
		clusterUpdateLock: clusterLock,
	}

	manager.On("updateNodeCount", mock.AnythingOfType("string"), mock.AnythingOfType("int")).Return(nil)

	ng1 := &magnumNodeGroup{magnumManager: manager, id: "ng1", targetSize: 2, maxSize: 4, clusterUpdateLock: clusterLock}
	ng2 := &magnumNodeGroup{magnumManager: manager, id: "ng2", targetSize: 2, maxSize: 4, clusterUpdateLock: clusterLock}

	provider.nodeGroups = []*magnumNodeGroup{ng1, ng2}

	for _, ng := range provider.NodeGroups() {
		err := ng.IncreaseSize(1)
		require.NoError(t, err)
	}

	assert.Equal(t, 3, ng1.targetSize, "targetSize was not updated on node group 1")
	assert.Equal(t, 3, ng2.targetSize, "targetSize was not updated on node group 2")
}

// TestRefreshNodeGroupsAdd checks that refreshNodeGroups correctly
// registers a newly added node group.
func TestRefreshNodeGroupsAdd(t *testing.T) {
	manager := &magnumManagerMock{}
	provider := magnumCloudProvider{
		magnumManager:        manager,
		usingAutoDiscovery:   true,
		autoDiscoveryConfigs: nil,
		nodeGroupsLock:       &sync.Mutex{},
		clusterUpdateLock:    &sync.Mutex{},
	}

	autoDiscoverySpec, err := parseMagnumAutoDiscoverySpec("magnum:role=autoscaling")
	require.NoError(t, err, "error parsing auto discovery spec")
	provider.autoDiscoveryConfigs = []magnumAutoDiscoveryConfig{autoDiscoverySpec}

	three := 3

	initialNodeGroups := []*nodegroups.NodeGroup{
		{
			UUID:         "ece653dd-2544-4f2e-b553-3c136af0ffa6",
			Name:         "test-ng-1",
			Role:         "autoscaling",
			NodeCount:    1,
			MinNodeCount: 1,
			MaxNodeCount: &three,
		},
	}

	manager.On("autoDiscoverNodeGroups", []magnumAutoDiscoveryConfig{autoDiscoverySpec}).Return(initialNodeGroups, nil).Once()
	manager.On("fetchNodeGroupStackIDs", mock.AnythingOfType("string")).Return(nodeGroupStacks{}, nil)

	secondNodeGroups := []*nodegroups.NodeGroup{
		{
			UUID:         "ece653dd-2544-4f2e-b553-3c136af0ffa6",
			Name:         "test-ng-1",
			Role:         "autoscaling",
			NodeCount:    1,
			MinNodeCount: 1,
			MaxNodeCount: &three,
		},
		{
			UUID:         "946ffca0-719c-4a3f-a157-3a091e4c97f5",
			Name:         "test-ng-2",
			Role:         "autoscaling",
			NodeCount:    1,
			MinNodeCount: 1,
			MaxNodeCount: &three,
		},
	}

	manager.On("autoDiscoverNodeGroups", []magnumAutoDiscoveryConfig{autoDiscoverySpec}).Return(secondNodeGroups, nil).Once()

	// Find the initial node group
	err = provider.refreshNodeGroups()
	require.NoError(t, err)
	require.Equal(t, 1, len(provider.nodeGroups), "wrong number of initial node groups")

	// Find the second node group
	err = provider.refreshNodeGroups()
	require.NoError(t, err)
	assert.Equal(t, 2, len(provider.nodeGroups), "wrong number of node groups after refresh")
	assert.ElementsMatch(t, []string{"test-ng-1-ece653dd", "test-ng-2-946ffca0"}, []string{provider.nodeGroups[0].Id(), provider.nodeGroups[1].Id()}, "didn't find both node groups")
}
func TestNodeInf(t *testing.T) {
	manager := &magnumManagerMock{}
	provider := magnumCloudProvider{
		magnumManager:        manager,
		usingAutoDiscovery:   true,
		autoDiscoveryConfigs: nil,
		nodeGroupsLock:       &sync.Mutex{},
		clusterUpdateLock:    &sync.Mutex{},
	}

	autoDiscoverySpec, err := parseMagnumAutoDiscoverySpec("magnum:role=autoscaling")
	require.NoError(t, err, "error parsing auto discovery spec")
	provider.autoDiscoveryConfigs = []magnumAutoDiscoveryConfig{autoDiscoverySpec}

	three := 3
	labels := make(map[string]string)
	labels["Label1"] = "Label1"
	labels["Label2"] = "Label2"

	nodeGroupWithLabels := []*nodegroups.NodeGroup{
		{
			UUID:         "1ce653dd-2544-4f2e-b553-3c136af0ffa6",
			Name:         "ng-with-labels",
			Role:         "autoscaling",
			NodeCount:    0,
			MinNodeCount: 1,
			MaxNodeCount: &three,
			FlavorID:     "s1.small",
			Labels:       labels,
		},
	}

	nodeGroupWithoutLabels := []*nodegroups.NodeGroup{
		{
			UUID:         "2ce653dd-2544-4f2e-b553-3c136af0ffa6",
			Name:         "ng-without-labels",
			Role:         "autoscaling",
			NodeCount:    0,
			MinNodeCount: 1,
			MaxNodeCount: &three,
			FlavorID:     "s1.small",
		},
	}

	manager.On("autoDiscoverNodeGroups", []magnumAutoDiscoveryConfig{autoDiscoverySpec}).Return(nodeGroupWithLabels, nil).Once()
	manager.On("autoDiscoverNodeGroups", []magnumAutoDiscoveryConfig{autoDiscoverySpec}).Return(nodeGroupWithoutLabels, nil).Once()

	manager.On("", mock.AnythingOfType("string")).Return(manager.getFlavorById("s1.small"))
	manager.On("getFlavorById", mock.AnythingOfType("string")).Return(manager.getFlavorById("s1.small"))
	manager.On("fetchNodeGroupStackIDs", mock.AnythingOfType("string")).Return(nodeGroupStacks{}, nil)

	err = provider.refreshNodeGroups()
	require.NoError(t, err)
	require.Equal(t, 1, len(provider.nodeGroups), "wrong number of initial node groups")

	nodeInfo, err := provider.nodeGroups[0].TemplateNodeInfo()
	assert.NoError(t, err)
	assert.Equal(t, len(nodeInfo.Pods()), 1, "should have one template pod")
	assert.Equal(t, nodeInfo.Node().Status.Capacity.Cpu().ToDec().Value(), int64(2000), "should match cpu capacity ")
	assert.Equal(t, nodeInfo.Node().Status.Capacity.Memory().ToDec().Value(), int64(2048*1024*1024), "should match memory capacity")
	assert.Equal(t, 3, len(nodeInfo.Node().Labels), "We should have 3 labels")
	assert.Equal(t, nodeInfo.Node().Labels["magnum.openstack.org/nodegroup"], "ng-with-labels", "Magnum nodegroup label should be")
	assert.Equal(t, nodeInfo.Node().Labels["Label1"], "Label1")
	assert.Equal(t, nodeInfo.Node().Labels["Label2"], "Label2")

	err = provider.refreshNodeGroups()
	require.NoError(t, err)
	require.Equal(t, 1, len(provider.nodeGroups), "wrong number of initial node groups")

	nodeInfo, err = provider.nodeGroups[0].TemplateNodeInfo()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(nodeInfo.Pods()), "should have one template pod")
	assert.Equal(t, int64(2000), nodeInfo.Node().Status.Capacity.Cpu().ToDec().Value(), "should match cpu capacity ")
	assert.Equal(t, int64(2048*1024*1024), nodeInfo.Node().Status.Capacity.Memory().ToDec().Value(), "should match memory capacity")
	assert.Equal(t, 1, len(nodeInfo.Node().Labels), "We should have 1 labels")
	assert.Equal(t, "ng-without-labels", nodeInfo.Node().Labels["magnum.openstack.org/nodegroup"], "Magnum nodegroup label should be")
}

// TestRefreshNodeGroupsRemove checks that refreshNodeGroups correctly
// removes a node group that no longer exists.
func TestRefreshNodeGroupsRemove(t *testing.T) {
	manager := &magnumManagerMock{}
	provider := magnumCloudProvider{
		magnumManager:        manager,
		usingAutoDiscovery:   true,
		autoDiscoveryConfigs: nil,
		nodeGroupsLock:       &sync.Mutex{},
		clusterUpdateLock:    &sync.Mutex{},
	}

	autoDiscoverySpec, err := parseMagnumAutoDiscoverySpec("magnum:role=autoscaling")
	assert.NoError(t, err, "error parsing auto discovery spec")
	provider.autoDiscoveryConfigs = []magnumAutoDiscoveryConfig{autoDiscoverySpec}

	three := 3

	initialNodeGroups := []*nodegroups.NodeGroup{
		{
			UUID:         "ece653dd-2544-4f2e-b553-3c136af0ffa6",
			Name:         "test-ng-1",
			Role:         "autoscaling",
			NodeCount:    1,
			MinNodeCount: 1,
			MaxNodeCount: &three,
		},
		{
			UUID:         "946ffca0-719c-4a3f-a157-3a091e4c97f5",
			Name:         "test-ng-2",
			Role:         "autoscaling",
			NodeCount:    1,
			MinNodeCount: 1,
			MaxNodeCount: &three,
		},
	}

	manager.On("autoDiscoverNodeGroups", []magnumAutoDiscoveryConfig{autoDiscoverySpec}).Return(initialNodeGroups, nil).Once()
	manager.On("fetchNodeGroupStackIDs", mock.AnythingOfType("string")).Return(nodeGroupStacks{}, nil)

	secondNodeGroups := []*nodegroups.NodeGroup{
		{
			UUID:         "ece653dd-2544-4f2e-b553-3c136af0ffa6",
			Name:         "test-ng-1",
			Role:         "autoscaling",
			NodeCount:    1,
			MinNodeCount: 1,
			MaxNodeCount: &three,
		},
	}

	manager.On("autoDiscoverNodeGroups", []magnumAutoDiscoveryConfig{autoDiscoverySpec}).Return(secondNodeGroups, nil).Once()

	// Find the initial node groups
	err = provider.refreshNodeGroups()
	require.NoError(t, err)
	require.Equal(t, 2, len(provider.nodeGroups), "wrong number of initial node groups")

	// Remove one node group
	err = provider.refreshNodeGroups()
	require.NoError(t, err)
	assert.Equal(t, 1, len(provider.nodeGroups), "wrong number of node groups after refresh")
	assert.Equal(t, "test-ng-1-ece653dd", provider.nodeGroups[0].Id(), "remaining node group has the wrong ID")
}

// TestRefreshNodeGroupsReplace checks that refreshNodeGroups correctly
// handles a new node group being added at the same time as one is removed.
func TestRefreshNodeGroupsReplace(t *testing.T) {
	manager := &magnumManagerMock{}
	provider := magnumCloudProvider{
		magnumManager:        manager,
		usingAutoDiscovery:   true,
		autoDiscoveryConfigs: nil,
		nodeGroupsLock:       &sync.Mutex{},
		clusterUpdateLock:    &sync.Mutex{},
	}

	autoDiscoverySpec, err := parseMagnumAutoDiscoverySpec("magnum:role=autoscaling")
	require.NoError(t, err, "error parsing auto discovery spec")
	provider.autoDiscoveryConfigs = []magnumAutoDiscoveryConfig{autoDiscoverySpec}

	three := 3

	initialNodeGroups := []*nodegroups.NodeGroup{
		{
			UUID:         "ece653dd-2544-4f2e-b553-3c136af0ffa6",
			Name:         "test-ng-1",
			Role:         "autoscaling",
			NodeCount:    1,
			MinNodeCount: 1,
			MaxNodeCount: &three,
		},
		{
			UUID:         "946ffca0-719c-4a3f-a157-3a091e4c97f5",
			Name:         "test-ng-2",
			Role:         "autoscaling",
			NodeCount:    1,
			MinNodeCount: 1,
			MaxNodeCount: &three,
		},
	}

	manager.On("autoDiscoverNodeGroups", []magnumAutoDiscoveryConfig{autoDiscoverySpec}).Return(initialNodeGroups, nil).Once()
	manager.On("fetchNodeGroupStackIDs", mock.AnythingOfType("string")).Return(nodeGroupStacks{}, nil)

	secondNodeGroups := []*nodegroups.NodeGroup{
		{
			UUID:         "ece653dd-2544-4f2e-b553-3c136af0ffa6",
			Name:         "test-ng-1",
			Role:         "autoscaling",
			NodeCount:    1,
			MinNodeCount: 1,
			MaxNodeCount: &three,
		},
		{
			UUID:         "1a61c05a-95e3-4aea-a0d4-088285e30e2c",
			Name:         "test-ng-3",
			Role:         "autoscaling",
			NodeCount:    1,
			MinNodeCount: 1,
			MaxNodeCount: &three,
		},
	}

	manager.On("autoDiscoverNodeGroups", []magnumAutoDiscoveryConfig{autoDiscoverySpec}).Return(secondNodeGroups, nil).Once()

	// Find the initial node group
	err = provider.refreshNodeGroups()
	require.NoError(t, err)
	require.Equal(t, 2, len(provider.nodeGroups), "wrong number of initial node groups")
	require.ElementsMatch(t, []string{"test-ng-1-ece653dd", "test-ng-2-946ffca0"}, []string{provider.nodeGroups[0].Id(), provider.nodeGroups[1].Id()}, "didn't find both initial node groups")

	// Find remove one node group and add another
	err = provider.refreshNodeGroups()
	require.NoError(t, err)
	assert.Equal(t, 2, len(provider.nodeGroups), "wrong number of node groups after refresh")
	assert.ElementsMatch(t, []string{"test-ng-1-ece653dd", "test-ng-3-1a61c05a"}, []string{provider.nodeGroups[0].Id(), provider.nodeGroups[1].Id()}, "didn't find both final node groups")
}

// TestRefreshNodeGroupsUpdate checks that refreshNodeGroups correctly
// updates a node group that has had its min and max node counts updated.
func TestRefreshNodeGroupsUpdate(t *testing.T) {
	manager := &magnumManagerMock{}
	provider := magnumCloudProvider{
		magnumManager:        manager,
		usingAutoDiscovery:   true,
		autoDiscoveryConfigs: nil,
		nodeGroupsLock:       &sync.Mutex{},
		clusterUpdateLock:    &sync.Mutex{},
	}

	autoDiscoverySpec, err := parseMagnumAutoDiscoverySpec("magnum:role=autoscaling")
	require.NoError(t, err, "error parsing auto discovery spec")
	provider.autoDiscoveryConfigs = []magnumAutoDiscoveryConfig{autoDiscoverySpec}

	three := 3
	four := 4

	initialNodeGroups := []*nodegroups.NodeGroup{
		{
			UUID:         "ece653dd-2544-4f2e-b553-3c136af0ffa6",
			Name:         "test-ng-1",
			Role:         "autoscaling",
			NodeCount:    2,
			MinNodeCount: 1,
			MaxNodeCount: &three,
		},
	}

	manager.On("autoDiscoverNodeGroups", []magnumAutoDiscoveryConfig{autoDiscoverySpec}).Return(initialNodeGroups, nil).Once()
	manager.On("fetchNodeGroupStackIDs", mock.AnythingOfType("string")).Return(nodeGroupStacks{}, nil)

	secondNodeGroups := []*nodegroups.NodeGroup{
		{
			UUID:         "ece653dd-2544-4f2e-b553-3c136af0ffa6",
			Name:         "test-ng-1",
			Role:         "autoscaling",
			NodeCount:    2,
			MinNodeCount: 2,
			MaxNodeCount: &four,
		},
	}

	manager.On("autoDiscoverNodeGroups", []magnumAutoDiscoveryConfig{autoDiscoverySpec}).Return(secondNodeGroups, nil).Once()

	// Find the initial node group
	err = provider.refreshNodeGroups()
	require.NoError(t, err)
	require.Equal(t, 1, len(provider.nodeGroups), "wrong number of initial node groups")
	require.Equal(t, 1, provider.nodeGroups[0].MinSize(), "wrong initial min node count")
	require.Equal(t, 3, provider.nodeGroups[0].MaxSize(), "wrong initial max node count")

	// Update the node group
	err = provider.refreshNodeGroups()
	require.NoError(t, err)
	assert.Equal(t, 1, len(provider.nodeGroups), "wrong number of node groups after refresh")
	assert.Equal(t, 2, provider.nodeGroups[0].MinSize(), "wrong updated min node count")
	assert.Equal(t, 4, provider.nodeGroups[0].MaxSize(), "wrong updated max node count")
}

// TestRefreshNodeGroupsEmpty checks that refreshNodeGroups correctly
// works when autodiscovery does not find any node groups to autoscale.
func TestRefreshNodeGroupsEmpty(t *testing.T) {
	manager := &magnumManagerMock{}
	provider := magnumCloudProvider{
		magnumManager:        manager,
		usingAutoDiscovery:   true,
		autoDiscoveryConfigs: nil,
		nodeGroupsLock:       &sync.Mutex{},
		clusterUpdateLock:    &sync.Mutex{},
	}

	autoDiscoverySpec, err := parseMagnumAutoDiscoverySpec("magnum:role=autoscaling")
	require.NoError(t, err, "error parsing auto discovery spec")
	provider.autoDiscoveryConfigs = []magnumAutoDiscoveryConfig{autoDiscoverySpec}

	initialNodeGroups := []*nodegroups.NodeGroup{}

	manager.On("autoDiscoverNodeGroups", []magnumAutoDiscoveryConfig{autoDiscoverySpec}).Return(initialNodeGroups, nil).Once()
	//manager.On("fetchNodeGroupStackIDs", mock.AnythingOfType("string")).Return(nodeGroupStacks{}, nil)

	secondNodeGroups := []*nodegroups.NodeGroup{}

	manager.On("autoDiscoverNodeGroups", []magnumAutoDiscoveryConfig{autoDiscoverySpec}).Return(secondNodeGroups, nil).Once()

	// Find the initial node group
	err = provider.refreshNodeGroups()
	require.NoError(t, err)
	require.Equal(t, 0, len(provider.nodeGroups), "wrong number of initial node groups")

	// Find the second node group
	err = provider.refreshNodeGroups()
	require.NoError(t, err)
	assert.Equal(t, 0, len(provider.nodeGroups), "wrong number of node groups after refresh")
}

func TestProviderNodeGroupForNode(t *testing.T) {
	manager := &magnumManagerMock{}
	provider := magnumCloudProvider{
		magnumManager:        manager,
		usingAutoDiscovery:   true,
		autoDiscoveryConfigs: nil,
		nodeGroupsLock:       &sync.Mutex{},
		clusterUpdateLock:    &sync.Mutex{},
	}

	// Node group UUIDs
	uuid1 := "1ee71848-0d97-4f71-b442-6eb26e6a3d21"
	uuid2 := "22e2ced8-e565-4e91-99d3-5236cf6a6fe1"
	uuid3 := "345eba74-a0a2-4350-a301-30eaade334e3"

	// Two of the node groups are registered with the autoscaler
	nodegroup1 := &magnumNodeGroup{id: "ng1", UUID: uuid1}
	nodegroup2 := &magnumNodeGroup{id: "ng2", UUID: uuid2}

	provider.AddNodeGroup(nodegroup1)
	provider.AddNodeGroup(nodegroup2)

	node1 := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "openstack:///node1",
		},
	}

	node2 := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "openstack:///node2",
		},
	}

	node3 := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "openstack:///node3",
		},
	}

	node4 := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "openstack:///node4",
		},
	}

	manager.On("nodeGroupForNode", node1).Return(uuid1, nil)
	manager.On("nodeGroupForNode", node2).Return(uuid2, nil)
	manager.On("nodeGroupForNode", node3).Return(uuid3, nil)
	manager.On("nodeGroupForNode", node4).Return("", fmt.Errorf("manager error"))

	t.Run("ng1", func(t *testing.T) {
		ng, err := provider.NodeGroupForNode(node1)
		assert.NoError(t, err)
		assert.Equal(t, nodegroup1, ng)
	})

	t.Run("ng2", func(t *testing.T) {
		ng, err := provider.NodeGroupForNode(node2)
		assert.NoError(t, err)
		assert.Equal(t, nodegroup2, ng)
	})

	t.Run("ng3", func(t *testing.T) {
		ng, err := provider.NodeGroupForNode(node3)
		assert.NoError(t, err)
		assert.Equal(t, nil, ng)
	})

	t.Run("error", func(t *testing.T) {
		_, err := provider.NodeGroupForNode(node4)
		assert.Error(t, err)
	})
}
