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
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

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

func (m *magnumManagerMock) getNodes(nodegroup string) ([]string, error) {
	args := m.Called(nodegroup)
	return args.Get(0).([]string), args.Error(1)
}

func (m *magnumManagerMock) deleteNodes(nodegroup string, nodes []NodeRef, updatedNodeCount int) error {
	args := m.Called(nodegroup, nodes, updatedNodeCount)
	return args.Error(0)
}

func (m *magnumManagerMock) getClusterStatus() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *magnumManagerMock) canUpdate() (bool, string, error) {
	args := m.Called()
	return args.Bool(0), args.String(1), args.Error(2)
}

func (m *magnumManagerMock) templateNodeInfo(nodegroup string) (*schedulerframework.NodeInfo, error) {
	return &schedulerframework.NodeInfo{}, nil
}

func createTestNodeGroup(manager magnumManager) *magnumNodeGroup {
	current := 1
	ng := magnumNodeGroup{
		magnumManager:       manager,
		id:                  "TestNodeGroup",
		clusterUpdateMutex:  &sync.Mutex{},
		minSize:             1,
		maxSize:             10,
		targetSize:          &current,
		waitTimeStep:        100 * time.Millisecond,
		deleteBatchingDelay: 250 * time.Millisecond,
	}
	return &ng
}

func TestWaitForClusterStatus(t *testing.T) {
	manager := &magnumManagerMock{}
	ng := createTestNodeGroup(manager)

	// Test all working normally
	t.Run("success", func(t *testing.T) {
		manager.On("getClusterStatus").Return(clusterStatusUpdateComplete, nil).Once()
		err := ng.waitForClusterStatus(clusterStatusUpdateComplete, 200*time.Millisecond)
		assert.NoError(t, err)
	})

	// Test timeout
	t.Run("timeout", func(t *testing.T) {
		manager.On("getClusterStatus").Return(clusterStatusUpdateInProgress, nil).Times(2)
		err := ng.waitForClusterStatus(clusterStatusUpdateComplete, 200*time.Millisecond)
		assert.Error(t, err)
		assert.Equal(t, "timeout (200ms) waiting for UPDATE_COMPLETE status", err.Error())
	})

	// Test error returned from manager
	t.Run("manager error", func(t *testing.T) {
		manager.On("getClusterStatus").Return("", errors.New("manager error")).Once()
		err := ng.waitForClusterStatus(clusterStatusUpdateComplete, 200*time.Millisecond)
		assert.Error(t, err)
		assert.Equal(t, "error waiting for UPDATE_COMPLETE status: manager error", err.Error())
	})
}

func TestIncreaseSize(t *testing.T) {
	manager := &magnumManagerMock{}
	ng := createTestNodeGroup(manager)

	// Test all working normally
	t.Run("success", func(t *testing.T) {
		manager.On("nodeGroupSize", "TestNodeGroup").Return(1, nil).Once()
		manager.On("canUpdate").Return(true, "", nil).Once()
		manager.On("updateNodeCount", "TestNodeGroup", 2).Return(nil).Once()
		manager.On("getClusterStatus").Return(clusterStatusUpdateInProgress, nil).Once()
		manager.On("getClusterStatus").Return(clusterStatusUpdateComplete, nil).Once()
		err := ng.IncreaseSize(1)
		assert.NoError(t, err)
		assert.Equal(t, 2, *ng.targetSize, "target size not updated")
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

	// Test current total nodes fails
	t.Run("node group size fails", func(t *testing.T) {
		manager.On("nodeGroupSize", "TestNodeGroup").Return(0, errors.New("manager error")).Once()
		err := ng.IncreaseSize(1)
		assert.Error(t, err)
		assert.Equal(t, "could not check current nodegroup size: manager error", err.Error())
	})

	// Test increase too large
	t.Run("increase too large", func(t *testing.T) {
		manager.On("nodeGroupSize", "TestNodeGroup").Return(1, nil).Once()
		err := ng.IncreaseSize(10)
		assert.Error(t, err)
		assert.Equal(t, "size increase too large, desired:11 max:10", err.Error())
	})

	// Test cluster status prevents update
	t.Run("status prevents update", func(t *testing.T) {
		manager.On("nodeGroupSize", "TestNodeGroup").Return(1, nil).Once()
		manager.On("canUpdate").Return(false, clusterStatusUpdateInProgress, nil).Once()
		err := ng.IncreaseSize(1)
		assert.Error(t, err)
		assert.Equal(t, "can not add nodes, cluster is in UPDATE_IN_PROGRESS status", err.Error())
	})

	// Test cluster status check fails
	t.Run("status check fails", func(t *testing.T) {
		manager.On("nodeGroupSize", "TestNodeGroup").Return(1, nil).Once()
		manager.On("canUpdate").Return(false, "", errors.New("manager error")).Once()
		err := ng.IncreaseSize(1)
		assert.Error(t, err)
		assert.Equal(t, "can not increase node count: manager error", err.Error())
	})

	// Test update node count fails
	t.Run("update node count fails", func(t *testing.T) {
		*ng.targetSize = 1
		manager.On("nodeGroupSize", "TestNodeGroup").Return(1, nil).Once()
		manager.On("canUpdate").Return(true, "", nil).Once()
		manager.On("updateNodeCount", "TestNodeGroup", 2).Return(errors.New("manager error")).Once()
		err := ng.IncreaseSize(1)
		assert.Error(t, err)
		assert.Equal(t, "could not increase cluster size: manager error", err.Error())
	})
}

var machineIDs []string

var nodesToDelete []*apiv1.Node
var nodeRefs []NodeRef

func init() {
	machineIDs = []string{
		"4d030dc5f2944154aaba43a004f9fccc",
		"ce120bf512f14497be3c1b07ac9a433a",
		"7cce782693dd4805bc8e2735d3600433",
		"8bf92ce193f7401c9a24d1dc2e75dc5d",
		"1ef6ab6407c3482185fbff428271a6a0",
	}
	for i, machineID := range machineIDs {
		nodeRefs = append(nodeRefs,
			NodeRef{Name: fmt.Sprintf("cluster-abc-minion-%d", i+1),
				MachineID:  machineID,
				ProviderID: fmt.Sprintf("openstack:///%s", machineID),
				IPs:        []string{fmt.Sprintf("10.0.0.%d", 100+i)},
			},
		)
	}

	for i := 0; i < 5; i++ {
		node := apiv1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("cluster-abc-minion-%d", i+1),
			},
			Status: apiv1.NodeStatus{
				NodeInfo: apiv1.NodeSystemInfo{
					MachineID: machineIDs[i],
				},
				Addresses: []apiv1.NodeAddress{
					{Type: apiv1.NodeInternalIP, Address: fmt.Sprintf("10.0.0.%d", 100+i)},
				},
			},
			Spec: apiv1.NodeSpec{
				ProviderID: fmt.Sprintf("openstack:///%s", machineIDs[i]),
			},
		}

		nodesToDelete = append(nodesToDelete, &node)
	}
}

func TestDeleteNodes(t *testing.T) {
	manager := &magnumManagerMock{}
	ng := createTestNodeGroup(manager)
	ng.deleteBatchingDelay = 25 * time.Millisecond

	// Test all working normally
	t.Run("success", func(t *testing.T) {
		*ng.targetSize = 10
		ng.deleteNodesCachedSizeAt = time.Time{}
		manager.On("nodeGroupSize", "TestNodeGroup").Return(10, nil).Once()
		manager.On("canUpdate").Return(true, clusterStatusUpdateComplete, nil).Once()
		manager.On("deleteNodes", "TestNodeGroup", nodeRefs, 5).Return(nil).Once()
		manager.On("getClusterStatus").Return(clusterStatusUpdateInProgress, nil).Once()
		manager.On("getClusterStatus").Return(clusterStatusUpdateComplete, nil).Once()
		manager.On("nodeGroupSize", "TestNodeGroup").Return(5, nil).Once()
		err := ng.DeleteNodes(nodesToDelete)
		assert.NoError(t, err)
		assert.Equal(t, 5, *ng.targetSize)
	})

	// Test cluster status check failing
	t.Run("cluster status check fail", func(t *testing.T) {
		*ng.targetSize = 10
		ng.deleteNodesCachedSizeAt = time.Time{}
		manager.On("nodeGroupSize", "TestNodeGroup").Return(10, nil).Once()
		manager.On("canUpdate").Return(false, "", errors.New("manager error")).Once()
		err := ng.DeleteNodes(nodesToDelete)
		assert.Error(t, err)
		assert.Equal(t, "could not check if cluster is ready to delete nodes: manager error", err.Error())
	})

	// Test cluster status prevents update
	t.Run("cluster status prevents update", func(t *testing.T) {
		*ng.targetSize = 10
		ng.deleteNodesCachedSizeAt = time.Time{}
		manager.On("nodeGroupSize", "TestNodeGroup").Return(10, nil).Once()
		manager.On("canUpdate").Return(false, clusterStatusUpdateInProgress, nil).Once()
		err := ng.DeleteNodes(nodesToDelete)
		assert.Error(t, err)
		assert.Equal(t, fmt.Sprintf("can not delete nodes, cluster is in %s status", clusterStatusUpdateInProgress), err.Error())
	})

	// Test error returned when checking current cluster size before deleting
	t.Run("current size check fail", func(t *testing.T) {
		*ng.targetSize = 10
		ng.deleteNodesCachedSizeAt = time.Time{}
		manager.On("nodeGroupSize", "TestNodeGroup").Return(0, errors.New("manager error")).Once()
		err := ng.DeleteNodes(nodesToDelete)
		assert.Error(t, err)
		assert.Equal(t, "could not get current node count: manager error", err.Error())
	})

	// Test trying to delete more nodes than minimum would allow, when the initial target size is out of sync with the cluster
	t.Run("delete below min", func(t *testing.T) {
		*ng.targetSize = 10
		ng.minSize = 8
		ng.deleteNodesCachedSizeAt = time.Time{}
		manager.On("nodeGroupSize", "TestNodeGroup").Return(10, nil).Once()
		err := ng.DeleteNodes(nodesToDelete)
		assert.Error(t, err)
		assert.Equal(t, "deleting nodes would take nodegroup below minimum size", err.Error())
	})
	ng.minSize = 1

	// Test call to deleteNodes on manager failing
	t.Run("deleteNodes fails", func(t *testing.T) {
		*ng.targetSize = 10
		ng.deleteNodesCachedSizeAt = time.Time{}
		manager.On("nodeGroupSize", "TestNodeGroup").Return(10, nil).Once()
		manager.On("canUpdate").Return(true, clusterStatusUpdateComplete, nil).Once()
		manager.On("deleteNodes", "TestNodeGroup", nodeRefs, 5).Return(errors.New("manager error")).Once()
		err := ng.DeleteNodes(nodesToDelete)
		assert.Error(t, err)
		assert.Equal(t, "manager error deleting nodes: manager error", err.Error())
	})

	// Test final call to get new cluster size fails
	t.Run("get size after delete fails", func(t *testing.T) {
		*ng.targetSize = 10
		ng.deleteNodesCachedSizeAt = time.Time{}
		manager.On("nodeGroupSize", "TestNodeGroup").Return(10, nil).Once()
		manager.On("canUpdate").Return(true, clusterStatusUpdateComplete, nil).Once()
		manager.On("deleteNodes", "TestNodeGroup", nodeRefs, 5).Return(nil).Once()
		manager.On("getClusterStatus").Return(clusterStatusUpdateInProgress, nil).Once()
		manager.On("getClusterStatus").Return(clusterStatusUpdateComplete, nil).Once()
		manager.On("nodeGroupSize", "TestNodeGroup").Return(0, errors.New("manager error")).Once()
		err := ng.DeleteNodes(nodesToDelete)
		assert.Error(t, err)
		assert.Equal(t, "could not check new cluster size after scale down: manager error", err.Error())
		assert.Equal(t, 5, *ng.targetSize, "targetSize should have been set to expected size 5, it is %d", *ng.targetSize)
	})
}

func TestDeleteNodesBatching(t *testing.T) {
	manager := &magnumManagerMock{}
	ng := createTestNodeGroup(manager)
	*ng.targetSize = 10

	// Test all working normally
	manager.On("nodeGroupSize", "TestNodeGroup").Return(10, nil).Once()
	manager.On("canUpdate").Return(true, "", nil).Once()
	manager.On("deleteNodes", "TestNodeGroup", nodeRefs, 5).Return(nil).Once()
	manager.On("nodeGroupSize", "TestNodeGroup").Return(5, nil).Once()
	manager.On("getClusterStatus").Return(clusterStatusUpdateInProgress, nil).Once()
	manager.On("getClusterStatus").Return(clusterStatusUpdateComplete, nil).Once()

	go func() {
		time.Sleep(time.Millisecond * 100)
		err := ng.DeleteNodes(nodesToDelete[3:5])
		assert.NoError(t, err, "Delete call that should have been batched did not return nil")
	}()

	err := ng.DeleteNodes(nodesToDelete[0:3])
	assert.NoError(t, err)
	assert.Equal(t, 5, *ng.targetSize)
	manager.AssertExpectations(t)
}

func TestDeleteNodesBatchBelowMin(t *testing.T) {
	manager := &magnumManagerMock{}
	ng := createTestNodeGroup(manager)
	ng.minSize = 8
	*ng.targetSize = 10

	// Try to batch 5 nodes for deletion when minSize only allows for two to be deleted

	manager.On("nodeGroupSize", "TestNodeGroup").Return(10, nil).Once()
	manager.On("canUpdate").Return(true, "", nil).Once()
	manager.On("deleteNodes", "TestNodeGroup", nodeRefs[0:2], 8).Return(nil).Once()
	manager.On("nodeGroupSize", "TestNodeGroup").Return(8, nil).Once()
	manager.On("getClusterStatus").Return(clusterStatusUpdateInProgress, nil).Once()
	manager.On("getClusterStatus").Return(clusterStatusUpdateComplete, nil).Once()

	for i := 1; i < 5; i++ {
		go func(i int) {
			// Wait and preserve order
			time.Sleep(time.Millisecond*100 + 25*time.Millisecond*time.Duration(i))
			err := ng.DeleteNodes(nodesToDelete[i : i+1])
			if i == 1 {
				// One node should be added to the batch
				assert.NoError(t, err, "Delete call that should have been batched did not return nil")
			} else {
				// The rest should fail
				assert.Error(t, err)
				assert.Equal(t, "deleting nodes would take nodegroup below minimum size", err.Error())
			}
		}(i)
	}

	err := ng.DeleteNodes(nodesToDelete[0:1])
	assert.NoError(t, err)
	manager.AssertExpectations(t)
}
