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

package magnum

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/magnum/gophercloud/openstack/containerinfra/v1/nodegroups"
)

const testNodeGroupUUID = "013701a6-4fcb-457d-91a4-44113d0f9b8d"

type magnumManagerMock struct {
	mock.Mock
}

func (m *magnumManagerMock) nodeGroupSize(nodegroup string) (int, error) {
	args := m.Called(nodegroup)
	return args.Int(0), args.Error(1)
}

func (m *magnumManagerMock) updateNodeCount(nodegroup string, nodes int) error {
	args := m.Called(nodegroup, nodes)
	return args.Error(0)
}

func (m *magnumManagerMock) getNodes(nodegroup string) ([]cloudprovider.Instance, error) {
	args := m.Called(nodegroup)
	return args.Get(0).([]cloudprovider.Instance), args.Error(1)
}

func (m *magnumManagerMock) deleteNodes(nodegroup string, nodes []NodeRef, updatedNodeCount int) error {
	args := m.Called(nodegroup, nodes, updatedNodeCount)
	return args.Error(0)
}

func (m *magnumManagerMock) isNodeInNodeGroup(node *apiv1.Node, nodegroup string) (bool, error) {
	args := m.Called(node, nodegroup)
	return args.Bool(0), args.Error(1)
}

func (m *magnumManagerMock) autoDiscoverNodeGroups(cfgs []magnumAutoDiscoveryConfig) ([]*nodegroups.NodeGroup, error) {
	args := m.Called(cfgs)
	return args.Get(0).([]*nodegroups.NodeGroup), args.Error(1)
}

func (m *magnumManagerMock) fetchNodeGroupStackIDs(nodegroup string) (nodeGroupStacks, error) {
	args := m.Called(nodegroup)
	return args.Get(0).(nodeGroupStacks), args.Error(1)
}

func (m *magnumManagerMock) uniqueNameAndIDForNodeGroup(nodegroup string) (string, string, error) {
	args := m.Called(nodegroup)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *magnumManagerMock) nodeGroupForNode(node *apiv1.Node) (string, error) {
	args := m.Called(node)
	return args.String(0), args.Error(1)
}

func createTestNodeGroup(manager magnumManager) *magnumNodeGroup {
	ng := magnumNodeGroup{
		magnumManager:     manager,
		id:                "TestNodeGroup",
		UUID:              testNodeGroupUUID,
		clusterUpdateLock: &sync.Mutex{},
		minSize:           1,
		maxSize:           10,
		targetSize:        1,
		deletedNodes:      make(map[string]time.Time),
		nodeTemplate:      &MagnumNodeTemplate{},
	}
	return &ng
}

func TestIncreaseSize(t *testing.T) {
	manager := &magnumManagerMock{}
	ng := createTestNodeGroup(manager)

	// Test all working normally
	t.Run("success", func(t *testing.T) {
		manager.On("updateNodeCount", testNodeGroupUUID, 2).Return(nil).Once()
		err := ng.IncreaseSize(1)
		assert.NoError(t, err)
		assert.Equal(t, 2, ng.targetSize, "target size not updated")
	})

	// Test negative increase
	t.Run("negative increase", func(t *testing.T) {
		err := ng.IncreaseSize(-1)
		assert.Error(t, err)
		assert.Equal(t, "size increase must be positive", err.Error())
	})

	// Test zero increase
	t.Run("zero increase", func(t *testing.T) {
		err := ng.IncreaseSize(0)
		assert.Error(t, err)
		assert.Equal(t, "size increase must be positive", err.Error())
	})

	// Test increase too large
	t.Run("increase too large", func(t *testing.T) {
		ng.targetSize = 1
		err := ng.IncreaseSize(10)
		assert.Error(t, err)
		assert.Equal(t, "size increase too large, desired:11 max:10", err.Error())
	})

	// Test update node count fails
	t.Run("update node count fails", func(t *testing.T) {
		ng.targetSize = 1
		manager.On("updateNodeCount", testNodeGroupUUID, 2).Return(errors.New("manager error")).Once()
		err := ng.IncreaseSize(1)
		assert.Error(t, err)
		assert.Equal(t, "could not increase cluster size: manager error", err.Error())
	})
}

func TestDecreaseSize(t *testing.T) {
	manager := &magnumManagerMock{}
	ng := createTestNodeGroup(manager)
	ng.targetSize = 3

	// Test all working normally
	t.Run("success", func(t *testing.T) {
		manager.On("updateNodeCount", testNodeGroupUUID, 2).Return(nil).Once()
		err := ng.DecreaseTargetSize(-1)
		assert.NoError(t, err)
		assert.Equal(t, 2, ng.targetSize, "target size not updated")
	})

	// Test positive decrease
	t.Run("positive decrease", func(t *testing.T) {
		err := ng.DecreaseTargetSize(1)
		assert.Error(t, err)
		assert.Equal(t, "size decrease must be negative", err.Error())
	})

	// Test zero decrease
	t.Run("zero decrease", func(t *testing.T) {
		err := ng.DecreaseTargetSize(0)
		assert.Error(t, err)
		assert.Equal(t, "size decrease must be negative", err.Error())
	})

	// Test decrease too large
	t.Run("decrease too large", func(t *testing.T) {
		ng.targetSize = 3
		err := ng.DecreaseTargetSize(-4)
		assert.Error(t, err)
		assert.Equal(t, "size decrease too large, desired:-1 min:1", err.Error())
	})

	// Test update node count fails
	t.Run("update node count fails", func(t *testing.T) {
		ng.targetSize = 3
		manager.On("updateNodeCount", testNodeGroupUUID, 2).Return(errors.New("manager error")).Once()
		err := ng.DecreaseTargetSize(-1)
		assert.Error(t, err)
		assert.Equal(t, "could not decrease target size: manager error", err.Error())
	})
}

var systemUUIDs []string

var nodesToDelete []*apiv1.Node
var nodeRefs []NodeRef

func init() {
	postDeleteSleepDuration = 100 * time.Millisecond

	systemUUIDs = []string{
		"4d030dc5-f294-4154-aaba-43a004f9fccc",
		"ce120bf5-12f1-4497-be3c-1b07ac9a433a",
		"7cce7826-93dd-4805-bc8e-2735d3600433",
		"8bf92ce1-93f7-401c-9a24-d1dc2e75dc5d",
		"1ef6ab64-07c3-4821-85fb-ff428271a6a0",
	}

	for i, UUID := range systemUUIDs {
		nodeRefs = append(nodeRefs,
			NodeRef{Name: fmt.Sprintf("cluster-abc-minion-%d", i+1),
				SystemUUID: UUID,
				ProviderID: fmt.Sprintf("openstack:///%s", UUID),
				IsFake:     false,
			},
		)
	}

	for i, UUID := range systemUUIDs {
		node := apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("cluster-abc-minion-%d", i+1),
				UID:  "ac558cdb-c7da-11e9-b4fd-fa163ecf721a", // need any UID or it is assumed to be a fake node
			},
			Status: apiv1.NodeStatus{
				NodeInfo: apiv1.NodeSystemInfo{
					SystemUUID: UUID,
				},
			},
			Spec: apiv1.NodeSpec{
				ProviderID: fmt.Sprintf("openstack:///%s", UUID),
			},
		}

		nodesToDelete = append(nodesToDelete, &node)
	}
}

func TestDeleteNodes(t *testing.T) {
	manager := &magnumManagerMock{}
	ng := createTestNodeGroup(manager)

	// Test all working normally
	t.Run("success", func(t *testing.T) {
		ng.targetSize = 10
		manager.On("deleteNodes", testNodeGroupUUID, nodeRefs, 5).Return(nil).Once()
		err := ng.DeleteNodes(nodesToDelete)
		assert.NoError(t, err)
		assert.Equal(t, 5, ng.targetSize)
	})

	// Test call to deleteNodes on manager failing
	t.Run("error", func(t *testing.T) {
		ng.targetSize = 10
		manager.On("deleteNodes", testNodeGroupUUID, nodeRefs, 5).Return(errors.New("manager error")).Once()
		err := ng.DeleteNodes(nodesToDelete)
		assert.Error(t, err)
		assert.Equal(t, "manager error deleting nodes: manager error", err.Error())
	})
}

// TestNodes checks that the NodeGroup takes into account
// nodes that have recently been requested for deletion
// when retrieving the list of Instances, reporting them
// as InstanceDeleting to prevent the autoscaler from trying
// to delete them multiple times.
func TestNodes(t *testing.T) {
	manager := &magnumManagerMock{}
	ng := createTestNodeGroup(manager)
	ng.targetSize = 6

	// Nodes which are running, should be reported as is.
	runningNodes := []cloudprovider.Instance{
		{Id: "openstack:///1", Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
		{Id: "openstack:///2", Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
		{Id: "openstack:///3", Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
	}

	// Nodes which have been requested for deleting but the state may not be updated yet on the OpenStack side.
	deletingNodes := []cloudprovider.Instance{
		{Id: "openstack:///4", Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
		{Id: "openstack:///5", Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}},
		{Id: "openstack:///6", Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceDeleting}},
	}

	// How the deleting nodes should be reported.
	expectedDeletingNodes := []cloudprovider.Instance{
		{Id: "openstack:///4", Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceDeleting}},
		{Id: "openstack:///5", Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceDeleting}},
		{Id: "openstack:///6", Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceDeleting}},
	}

	nodesToDelete := []*apiv1.Node{
		{Spec: apiv1.NodeSpec{ProviderID: "openstack:///4"}},
		{Spec: apiv1.NodeSpec{ProviderID: "openstack:///5"}},
		{Spec: apiv1.NodeSpec{ProviderID: "openstack:///6"}},
	}

	manager.On("deleteNodes", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	err := ng.DeleteNodes(nodesToDelete)
	require.NoError(t, err)

	allNodes := append(runningNodes, deletingNodes...)
	expectedNodes := append(runningNodes, expectedDeletingNodes...)

	manager.On("getNodes", testNodeGroupUUID).Return(allNodes, nil).Once()

	nodes, err := ng.Nodes()
	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedNodes, nodes)

	// After some time the nodes have actually been deleted in OpenStack.
	// The deleted nodes map should be cleaned after 10 minutes.
	for node, deletedTime := range ng.deletedNodes {
		// Fast forward time by 20 minutes.
		ng.deletedNodes[node] = deletedTime.Add(-20 * time.Minute)
	}

	manager.On("getNodes", testNodeGroupUUID).Return(runningNodes, nil).Once()

	nodes, err = ng.Nodes()
	assert.NoError(t, err)
	assert.ElementsMatch(t, runningNodes, nodes)
	assert.Equalf(t, 0, len(ng.deletedNodes), "node group deletedNodes map was not cleaned")
}
