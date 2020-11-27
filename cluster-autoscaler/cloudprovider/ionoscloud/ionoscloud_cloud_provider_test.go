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

package ionoscloud

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

type NodeGroupTestSuite struct {
	suite.Suite
	manager    *MockIonosCloudManager
	nodePool   *nodePool
	deleteNode []*apiv1.Node
}

func (s *NodeGroupTestSuite) SetupTest() {
	s.manager = &MockIonosCloudManager{}
	s.nodePool = &nodePool{
		id:      "test",
		min:     1,
		max:     3,
		manager: s.manager,
	}
	s.deleteNode = []*apiv1.Node{
		{Spec: apiv1.NodeSpec{ProviderID: convertToInstanceId("testnode")}},
	}
}

func (s *NodeGroupTestSuite) TearDownTest() {
	s.manager.AssertExpectations(s.T())
}

func TestNodeGroup(t *testing.T) {
	suite.Run(t, new(NodeGroupTestSuite))
}

func (s *NodeGroupTestSuite) TestId() {
	s.Equal("test", s.nodePool.Id())
}

func (s *NodeGroupTestSuite) TestDebug() {
	s.Equal("Id=test, Min=1, Max=3", s.nodePool.Debug())
}

func (s *NodeGroupTestSuite) TestMaxSize() {
	s.Equal(3, s.nodePool.MaxSize())
}

func (s *NodeGroupTestSuite) TestMinSize() {
	s.Equal(1, s.nodePool.MinSize())
}

func (s *NodeGroupTestSuite) TestTemplateNodeInfo() {
	nodeInfo, err := s.nodePool.TemplateNodeInfo()
	s.Nil(nodeInfo)
	s.Equal(cloudprovider.ErrNotImplemented, err)
}

func (s *NodeGroupTestSuite) TestExist() {
	var nodePool *nodePool
	s.True((nodePool).Exist())
}

func (s *NodeGroupTestSuite) TestCreate() {
	nodeGroup, err := s.nodePool.Create()
	s.Nil(nodeGroup)
	s.Equal(cloudprovider.ErrNotImplemented, err)
}

func (s *NodeGroupTestSuite) TestDelete() {
	err := s.nodePool.Delete()
	s.Equal(cloudprovider.ErrNotImplemented, err)
}

func (s *NodeGroupTestSuite) TestAutoprovisioned() {
	var nodePool *nodePool
	s.False(nodePool.Autoprovisioned())
}

func (s *NodeGroupTestSuite) TestTargetSize_Error() {
	s.manager.On("GetNodeGroupTargetSize", s.nodePool).Return(0, fmt.Errorf("error")).Once()
	_, err := s.nodePool.TargetSize()
	s.Error(err)
}

func (s *NodeGroupTestSuite) TestTargetSize_OK() {
	s.manager.On("GetNodeGroupTargetSize", s.nodePool).Return(3, nil).Once()
	size, err := s.nodePool.TargetSize()
	s.NoError(err)
	s.Equal(3, size)
}

func (s *NodeGroupTestSuite) TestIncreaseSize_InvalidDelta() {
	s.Error(s.nodePool.IncreaseSize(0))
}

func (s *NodeGroupTestSuite) TestIncreaseSize_GetSizeError() {
	s.manager.On("GetNodeGroupSize", s.nodePool).Return(0, fmt.Errorf("error")).Once()
	s.Error(s.nodePool.IncreaseSize(2))
}

func (s *NodeGroupTestSuite) TestIncreaseSize_ExceedMax() {
	s.manager.On("GetNodeGroupSize", s.nodePool).Return(2, nil).Once()
	s.Error(s.nodePool.IncreaseSize(2))
}

func (s *NodeGroupTestSuite) TestIncreaseSize_SetSizeError() {
	s.manager.On("GetNodeGroupSize", s.nodePool).Return(2, nil).Once()
	s.manager.On("SetNodeGroupSize", s.nodePool, 3).Return(fmt.Errorf("error")).Once()
	s.Error(s.nodePool.IncreaseSize(1))
}

func (s *NodeGroupTestSuite) TestIncreaseSize_OK() {
	s.manager.On("GetNodeGroupSize", s.nodePool).Return(2, nil).Once()
	s.manager.On("SetNodeGroupSize", s.nodePool, 3).Return(nil).Once()
	s.NoError(s.nodePool.IncreaseSize(1))
}

func (s *NodeGroupTestSuite) TestDeleteNodes_Locked() {
	s.manager.On("TryLockNodeGroup", s.nodePool).Return(false).Once()
	s.Error(s.nodePool.DeleteNodes(s.deleteNode))
}

func (s *NodeGroupTestSuite) TestDeleteNodes_DeleteError() {
	s.manager.On("TryLockNodeGroup", s.nodePool).Return(true).Once()
	s.manager.On("UnlockNodeGroup", s.nodePool).Return().Once()
	s.manager.On("DeleteNode", s.nodePool, "testnode").Return(fmt.Errorf("error")).Once()
	s.Error(s.nodePool.DeleteNodes(s.deleteNode))
}

func (s *NodeGroupTestSuite) TestDeleteNodes_OK() {
	s.manager.On("TryLockNodeGroup", s.nodePool).Return(true).Once()
	s.manager.On("UnlockNodeGroup", s.nodePool).Return().Once()
	s.manager.On("DeleteNode", s.nodePool, "testnode").Return(nil).Once()
	s.NoError(s.nodePool.DeleteNodes(s.deleteNode))
}

func (s *NodeGroupTestSuite) TestDecreaseTargetSize_InvalidDelta() {
	s.Error(s.nodePool.DecreaseTargetSize(0))
}

func (s *NodeGroupTestSuite) TestDecreaseTargetSize_GetTargetSizeError() {
	s.manager.On("GetNodeGroupTargetSize", s.nodePool).Return(0, fmt.Errorf("error")).Once()
	s.Error(s.nodePool.DecreaseTargetSize(-1))
}

func (s *NodeGroupTestSuite) TestDecreaseTargetSize_ExceedMin() {
	s.manager.On("GetNodeGroupTargetSize", s.nodePool).Return(1, nil).Once()
	s.Error(s.nodePool.DecreaseTargetSize(-2))
}

func (s *NodeGroupTestSuite) TestDecreaseTargetSize_NotImplemented() {
	s.manager.On("GetNodeGroupTargetSize", s.nodePool).Return(1, nil).Once()
	s.Error(s.nodePool.DecreaseTargetSize(-1))
}

type CloudProviderTestSuite struct {
	suite.Suite
	manager  *MockIonosCloudManager
	provider *IonosCloudCloudProvider
}

func (s *CloudProviderTestSuite) SetupTest() {
	s.manager = &MockIonosCloudManager{}
	s.provider = BuildIonosCloudCloudProvider(s.manager, &cloudprovider.ResourceLimiter{})
}

func (s *CloudProviderTestSuite) TearDownTest() {
	s.manager.AssertExpectations(s.T())
}

func TestIonosCloudCloudProvider(t *testing.T) {
	suite.Run(t, new(CloudProviderTestSuite))
}

func (s *CloudProviderTestSuite) TestName() {
	var ionoscloud *IonosCloudCloudProvider
	s.Equal(cloudprovider.IonoscloudProviderName, ionoscloud.Name())
}

func (s *CloudProviderTestSuite) TestNodeGroups() {
	s.manager.On("GetNodeGroups").Return([]cloudprovider.NodeGroup{&nodePool{id: "test"}}).Once()
	s.Equal([]cloudprovider.NodeGroup{&nodePool{id: "test"}}, s.provider.NodeGroups())
}

func (s *CloudProviderTestSuite) TestNodeGroupForNode_CacheHit() {
	node := newAPINode("test")
	s.manager.On("GetNodeGroupForNode", node).Return(&nodePool{id: "test"}).Once()

	nodeGroup, err := s.provider.NodeGroupForNode(node)
	s.NoError(err)
	s.Equal(&nodePool{id: "test"}, nodeGroup)
}

func (s *CloudProviderTestSuite) TestNodeGroupForNode_Error() {
	node := newAPINode("test")
	nodePool := &nodePool{id: "test", manager: s.manager}
	s.manager.On("GetNodeGroupForNode", node).Return(nil).Once()
	s.manager.On("GetNodeGroups").Return([]cloudprovider.NodeGroup{nodePool}).Once()
	s.manager.On("GetInstancesForNodeGroup", nodePool).Return(nil, fmt.Errorf("error")).Once()

	nodeGroup, err := s.provider.NodeGroupForNode(node)
	s.Nil(nodeGroup)
	s.Error(err)
}

func (s *CloudProviderTestSuite) TestNodeGroupForNode_Found() {
	node := newAPINode("test")
	nodePool := &nodePool{id: "test", manager: s.manager}
	s.manager.On("GetNodeGroupForNode", node).Return(nil).Once()
	s.manager.On("GetNodeGroups").Return([]cloudprovider.NodeGroup{nodePool}).Once()
	s.manager.On("GetInstancesForNodeGroup", nodePool).Return([]cloudprovider.Instance{
		newInstance("foo"), newInstance("test"),
	}, nil).Once()

	nodeGroup, err := s.provider.NodeGroupForNode(node)
	s.Equal(nodePool, nodeGroup)
	s.NoError(err)
}

func (s *CloudProviderTestSuite) TestNodeGroupForNode_NotFound() {
	node := newAPINode("test")
	nodePool := &nodePool{id: "test", manager: s.manager}
	s.manager.On("GetNodeGroupForNode", node).Return(nil).Once()
	s.manager.On("GetNodeGroups").Return([]cloudprovider.NodeGroup{nodePool}).Once()
	s.manager.On("GetInstancesForNodeGroup", nodePool).Return([]cloudprovider.Instance{
		newInstance("foo"), newInstance("bar"),
	}, nil).Once()

	nodeGroup, err := s.provider.NodeGroupForNode(node)
	s.Nil(nodeGroup)
	s.NoError(err)
}

func (s *CloudProviderTestSuite) TestPricing_NotImplemented() {
	var ionoscloud *IonosCloudCloudProvider
	pricingModel, err := ionoscloud.Pricing()
	s.Nil(pricingModel)
	s.Equal(cloudprovider.ErrNotImplemented, err)
}

func (s *CloudProviderTestSuite) TestGetAvailableMachineTypes_NotImplemented() {
	var ionoscloud *IonosCloudCloudProvider
	machineTypes, err := ionoscloud.GetAvailableMachineTypes()
	s.Nil(machineTypes)
	s.Equal(cloudprovider.ErrNotImplemented, err)
}

func (s *CloudProviderTestSuite) TestNewNodeGroup_NotImplemented() {
	var ionoscloud *IonosCloudCloudProvider
	nodeGroup, err := ionoscloud.NewNodeGroup("test", nil, nil, nil, nil)
	s.Nil(nodeGroup)
	s.Equal(cloudprovider.ErrNotImplemented, err)
}

func (s *CloudProviderTestSuite) TestGetResourceLimiter() {
	rl, err := s.provider.GetResourceLimiter()
	s.NotNil(rl)
	s.NoError(err)
}

func (s *CloudProviderTestSuite) TestGPULabel() {
	var ionoscloud *IonosCloudCloudProvider
	s.Equal(GPULabel, ionoscloud.GPULabel())
}

func (s *CloudProviderTestSuite) TestGetAvailableGPUTypes() {
	var ionoscloud *IonosCloudCloudProvider
	s.Nil(ionoscloud.GetAvailableGPUTypes())
}

func (s *CloudProviderTestSuite) TestCleanup() {
	s.NoError(s.provider.Cleanup())
}

func (s *CloudProviderTestSuite) TestRefresh() {
	s.NoError(s.provider.Refresh())
}
