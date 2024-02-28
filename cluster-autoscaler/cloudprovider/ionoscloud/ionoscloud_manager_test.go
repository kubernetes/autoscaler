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
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	ionos "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/ionoscloud/ionos-cloud-sdk-go"
	"k8s.io/utils/ptr"
)

var (
	pollInterval = 10 * time.Millisecond
	pollTimeout  = 50 * time.Millisecond
)

func TestLoadConfigFromEnv(t *testing.T) {
	cases := []struct {
		name      string
		env       map[string]string
		expectErr bool
		expectCfg *Config
	}{
		{
			name:      "missing cluster ID",
			expectErr: true,
		},
		{
			name:      "missing both token and tokens path",
			env:       map[string]string{envKeyClusterID: "1"},
			expectErr: true,
		},
		{
			name:      "invalid value for insecure",
			env:       map[string]string{envKeyClusterID: "1", envKeyToken: "token", envKeyInsecure: "fasle"},
			expectErr: true,
		},
		{
			name:      "invalid value for interval",
			env:       map[string]string{envKeyClusterID: "1", envKeyToken: "token", envKeyPollInterval: "10Ghz"},
			expectErr: true,
		},
		{
			name:      "invalid value for timeout",
			env:       map[string]string{envKeyClusterID: "1", envKeyToken: "token", envKeyPollTimeout: "1ly"},
			expectErr: true,
		},
		{
			name:      "invalid header format",
			env:       map[string]string{envKeyClusterID: "1", envKeyToken: "token", envKeyAdditionalHeaders: "foo=bar,baz=qux"},
			expectErr: true,
		},
		{
			name: "use defaults",
			env: map[string]string{
				envKeyClusterID: "test",
				envKeyToken:     "token",
			},
			expectCfg: &Config{
				ClusterID:    "test",
				PollInterval: defaultInterval,
				PollTimeout:  defaultTimeout,
				Token:        "token",
			},
		},
		{
			name: "all fields set",
			env: map[string]string{
				envKeyClusterID:         "test",
				envKeyEndpoint:          "/dev/null",
				envKeyInsecure:          "1",
				envKeyPollInterval:      "42ms",
				envKeyPollTimeout:       "1337s",
				envKeyToken:             "token",
				envKeyTokensPath:        "/etc/passwd",
				envKeyAdditionalHeaders: "foo:bar;; baz:qux; ",
			},
			expectCfg: &Config{
				ClusterID:         "test",
				Endpoint:          "/dev/null",
				Insecure:          true,
				PollInterval:      42 * time.Millisecond,
				PollTimeout:       1337 * time.Second,
				Token:             "token",
				TokensPath:        "/etc/passwd",
				AdditionalHeaders: map[string]string{"foo": "bar", "baz": "qux"},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			for key, value := range c.env {
				t.Setenv(key, value)
			}

			cfg, err := LoadConfigFromEnv()
			require.Equalf(t, c.expectErr, err != nil, "expected error: %t, got: %v", c.expectErr, err)
			require.Equal(t, c.expectCfg, cfg)
		})
	}
}

func TestCreateIonosCloudManager(t *testing.T) {
	t.Setenv(envKeyClusterID, "test")
	t.Setenv(envKeyToken, "token")

	manager, err := CreateIonosCloudManager(nil, "ua")
	require.Nil(t, manager)
	require.Error(t, err)
}

func newKubernetesNodePool(state string, size int32) *ionos.KubernetesNodePool {
	return &ionos.KubernetesNodePool{
		Id:         ptr.To("test"),
		Metadata:   &ionos.DatacenterElementMetadata{State: ptr.To(state)},
		Properties: &ionos.KubernetesNodePoolProperties{NodeCount: ptr.To(size)},
	}
}

func newKubernetesNode(id, state string) ionos.KubernetesNode {
	return ionos.KubernetesNode{
		Id:       ptr.To(id),
		Metadata: &ionos.KubernetesNodeMetadata{State: ptr.To(state)},
	}
}

func newInstance(id string) cloudprovider.Instance {
	return cloudprovider.Instance{Id: convertToInstanceID(id)}
}

func newInstanceWithState(id string, state cloudprovider.InstanceState) cloudprovider.Instance {
	instance := newInstance(id)
	instance.Status = &cloudprovider.InstanceStatus{
		State: state,
	}
	return instance
}

func newAPINode(id string) *apiv1.Node {
	return &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: convertToInstanceID(id),
		},
	}
}

func newAPIResponse(statusCode int) *ionos.APIResponse {
	return &ionos.APIResponse{
		Response: &http.Response{
			StatusCode: statusCode,
		},
	}
}

type ManagerTestSuite struct {
	suite.Suite
	*require.Assertions
	mockAPIClient *MockAPIClient
	manager       *ionosCloudManagerImpl
	nodePool      *nodePool
}

func (s *ManagerTestSuite) SetupTest() {
	s.Assertions = s.Require()
	s.mockAPIClient = NewMockAPIClient(s.T())
	client := NewAutoscalingClient(&Config{
		ClusterID:    "cluster",
		Token:        "token",
		PollInterval: pollInterval,
		PollTimeout:  pollTimeout,
	}, "ua")
	client.client = s.mockAPIClient

	s.manager = newManager(client)
	s.nodePool = &nodePool{
		id:      "test",
		min:     1,
		max:     3,
		manager: s.manager,
	}
}

func (s *ManagerTestSuite) OnGetKubernetesNodePool(retval *ionos.KubernetesNodePool, reterr error) *mock.Call {
	req := ionos.ApiK8sNodepoolsFindByIdRequest{}
	nodepool := ionos.KubernetesNodePool{}
	if retval != nil {
		nodepool = *retval
	}
	statusCode := 200
	if reterr != nil {
		statusCode = 500
	}
	return s.mockAPIClient.
		On("K8sNodepoolsFindById", mock.Anything, s.manager.client.cfg.ClusterID, s.nodePool.id).Return(req).
		On("K8sNodepoolsFindByIdExecute", req).Return(nodepool, newAPIResponse(statusCode), reterr)
}

func (s *ManagerTestSuite) OnUpdateKubernetesNodePool(size int32, reterr error) *mock.Call {
	// use actual client here to fill in private fields
	cl := ionos.APIClient{}.KubernetesApi
	origReq := cl.K8sNodepoolsPut(context.Background(), s.manager.client.cfg.ClusterID, s.nodePool.id)
	expect := origReq.KubernetesNodePool(resizeRequestBody(size))
	statusCode := 202
	if reterr != nil {
		statusCode = 500
	}
	return s.mockAPIClient.
		On("K8sNodepoolsPut", mock.Anything, s.manager.client.cfg.ClusterID, s.nodePool.id).Return(origReq).
		On("K8sNodepoolsPutExecute", expect).Return(ionos.KubernetesNodePool{}, newAPIResponse(statusCode), reterr)
}

func (s *ManagerTestSuite) OnListKubernetesNodes(retval *ionos.KubernetesNodes, reterr error) *mock.Call {
	origReq := ionos.ApiK8sNodepoolsNodesGetRequest{}
	req := ionos.ApiK8sNodepoolsNodesGetRequest{}.Depth(1)
	nodes := ionos.KubernetesNodes{}
	if retval != nil {
		nodes = *retval
	}
	statusCode := 200
	if reterr != nil {
		statusCode = 500
	}
	return s.mockAPIClient.
		On("K8sNodepoolsNodesGet", mock.Anything, s.manager.client.cfg.ClusterID, s.nodePool.id).Return(origReq).
		On("K8sNodepoolsNodesGetExecute", req).Return(nodes, newAPIResponse(statusCode), reterr)
}

func (s *ManagerTestSuite) OnDeleteKubernetesNode(id string, reterr error) *mock.Call {
	req := ionos.ApiK8sNodepoolsNodesDeleteRequest{}
	statusCode := 202
	if reterr != nil {
		statusCode = 500
	}
	return s.mockAPIClient.
		On("K8sNodepoolsNodesDelete", mock.Anything, s.manager.client.cfg.ClusterID, s.nodePool.id, id).Return(req).
		On("K8sNodepoolsNodesDeleteExecute", req).Return(newAPIResponse(statusCode), reterr)
}

func TestIonosCloudManager(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}

func (s *ManagerTestSuite) TestGetNodeGroupSize_Error() {
	s.OnListKubernetesNodes(nil, errors.New("error")).Once()

	size, err := s.manager.GetNodeGroupSize(s.nodePool)
	s.Error(err)
	s.Zero(size)
}

func (s *ManagerTestSuite) TestGetNodeGroupSize_OK() {
	s.OnListKubernetesNodes(&ionos.KubernetesNodes{
		Items: &[]ionos.KubernetesNode{
			newKubernetesNode("node-1", K8sNodeStateReady),
			newKubernetesNode("node-1", K8sNodeStateReady),
		},
	}, nil).Once()

	size, err := s.manager.GetNodeGroupSize(s.nodePool)
	s.NoError(err)
	s.Equal(2, size)
}

func (s *ManagerTestSuite) TestGetNodeGroupTargetSize_Error() {
	s.OnGetKubernetesNodePool(nil, errors.New("error")).Once()

	size, err := s.manager.GetNodeGroupTargetSize(s.nodePool)
	s.Error(err)
	s.Zero(size)
}

func (s *ManagerTestSuite) TestGetNodeGroupTargetSize_OK() {
	s.OnGetKubernetesNodePool(newKubernetesNodePool(ionos.Active, 2), nil).Once()

	size, err := s.manager.GetNodeGroupTargetSize(s.nodePool)
	s.NoError(err)
	s.Equal(2, size)
}

func (s *ManagerTestSuite) TestSetNodeGroupSize_ResizeError() {
	s.manager.cache.SetNodeGroupSize(s.nodePool.Id(), 1)
	s.OnUpdateKubernetesNodePool(2, errors.New("error")).Once()

	s.Error(s.manager.SetNodeGroupSize(s.nodePool, 2))
	s.Empty(s.manager.cache.GetNodeGroups())
	size, found := s.manager.cache.GetNodeGroupSize(s.nodePool.Id())
	s.True(found)
	s.Equal(1, size)
}

func (s *ManagerTestSuite) TestSetNodeGroupSize_WaitGetError() {
	s.OnUpdateKubernetesNodePool(2, nil).Once()
	s.OnGetKubernetesNodePool(newKubernetesNodePool(ionos.Updating, 1), nil).Times(3)
	s.OnGetKubernetesNodePool(nil, errors.New("error")).Once()

	err := s.manager.SetNodeGroupSize(s.nodePool, 2)
	s.Error(err)
	s.False(wait.Interrupted(err))
	s.Empty(s.manager.cache.GetNodeGroups())
}

func (s *ManagerTestSuite) TestSetNodeGroupSize_WaitTimeout() {
	s.OnUpdateKubernetesNodePool(2, nil).Once()
	pollCount := 0
	s.OnGetKubernetesNodePool(newKubernetesNodePool(ionos.Updating, 1), nil).
		Run(func(_ mock.Arguments) {
			pollCount++
		})

	err := s.manager.SetNodeGroupSize(s.nodePool, 2)
	s.Error(err)
	s.True(wait.Interrupted(err))
	// The poll count may vary, so just do this to prevent flakes.
	s.GreaterOrEqual(pollCount, int(pollTimeout/pollInterval))
	s.Empty(s.manager.cache.GetNodeGroups())
}

func (s *ManagerTestSuite) TestSetNodeGroupSize_RefreshNodesError() {
	s.OnUpdateKubernetesNodePool(2, nil).Once()
	s.OnGetKubernetesNodePool(newKubernetesNodePool(ionos.Updating, 1), nil).Times(3)
	s.OnGetKubernetesNodePool(newKubernetesNodePool(ionos.Active, 2), nil).Once()
	s.OnListKubernetesNodes(nil, errors.New("error")).Once()

	s.Error(s.manager.SetNodeGroupSize(s.nodePool, 2))
	s.Empty(s.manager.cache.GetNodeGroups())
}

func (s *ManagerTestSuite) TestSetNodeGroupSize_OK() {
	s.OnUpdateKubernetesNodePool(2, nil).Once()
	s.OnGetKubernetesNodePool(newKubernetesNodePool(ionos.Updating, 1), nil).Times(3)
	s.OnGetKubernetesNodePool(newKubernetesNodePool(ionos.Active, 2), nil).Once()
	s.OnListKubernetesNodes(&ionos.KubernetesNodes{
		Items: &[]ionos.KubernetesNode{
			newKubernetesNode("node-1", K8sNodeStateReady),
			newKubernetesNode("node-2", K8sNodeStateReady),
		},
	}, nil).Once()

	_, found := s.manager.cache.GetNodeGroupSize(s.nodePool.Id())
	s.False(found)
	s.NoError(s.manager.SetNodeGroupSize(s.nodePool, 2))
	size, found := s.manager.cache.GetNodeGroupSize(s.nodePool.Id())
	s.True(found)
	s.Equal(2, size)
}

func (s *ManagerTestSuite) TestGetInstancesForNodeGroup_Error() {
	s.OnListKubernetesNodes(nil, errors.New("error")).Once()

	instances, err := s.manager.GetInstancesForNodeGroup(s.nodePool)
	s.Error(err)
	s.Empty(instances)
}

func (s *ManagerTestSuite) TestGetInstancesForNodeGroup_RefreshOK() {
	s.manager.cache.AddNodeGroup(s.nodePool)
	s.OnListKubernetesNodes(&ionos.KubernetesNodes{
		Items: &[]ionos.KubernetesNode{
			newKubernetesNode("node-1", K8sNodeStateReady),
			newKubernetesNode("node-2", K8sNodeStateReady),
			newKubernetesNode("node-3", K8sNodeStateProvisioning),
		},
	}, nil).Once()

	expectInstances := []cloudprovider.Instance{
		newInstanceWithState("node-1", cloudprovider.InstanceRunning),
		newInstanceWithState("node-2", cloudprovider.InstanceRunning),
		newInstanceWithState("node-3", cloudprovider.InstanceCreating),
	}
	instances, err := s.manager.GetInstancesForNodeGroup(s.nodePool)
	s.NoError(err)
	s.ElementsMatch(expectInstances, instances)
}

func (s *ManagerTestSuite) TestGetNodeGroupForNode_NoMatchingNodes() {
	s.Nil(s.manager.GetNodeGroupForNode(newAPINode("node-1")))
}

func (s *ManagerTestSuite) TestGetNodeGroupForNode_NoMatchingNodePools() {
	s.manager.cache.nodesToNodeGroups["node-1"] = "foo"

	s.Nil(s.manager.GetNodeGroupForNode(newAPINode("node-1")))
}

func (s *ManagerTestSuite) TestGetNodeGroupForNode_OK() {
	s.manager.cache.nodesToNodeGroups["node-1"] = s.nodePool.Id()
	s.manager.cache.nodeGroups[s.nodePool.Id()] = s.nodePool

	nodePool := s.manager.GetNodeGroupForNode(newAPINode("node-1"))
	s.Equal(s.nodePool, nodePool)
}

func (s *ManagerTestSuite) TestGetNodeGroups_OK() {
	nodePools := []*nodePool{
		{id: "1", min: 1, max: 3, manager: s.manager},
		{id: "2", min: 1, max: 5, manager: s.manager},
		{id: "3", min: 1, max: 10, manager: s.manager},
	}
	for _, nodePool := range nodePools {
		s.manager.cache.AddNodeGroup(nodePool)
	}

	s.ElementsMatch(nodePools, s.manager.GetNodeGroups())
}

func (s *ManagerTestSuite) TestTryLockNodeGroup_LockUnlock() {
	s.True(s.manager.TryLockNodeGroup(s.nodePool))
	s.False(s.manager.TryLockNodeGroup(s.nodePool))
	s.manager.UnlockNodeGroup(&nodePool{id: "some other node pool"})
	s.False(s.manager.TryLockNodeGroup(s.nodePool))
	s.manager.UnlockNodeGroup(s.nodePool)
	s.True(s.manager.TryLockNodeGroup(s.nodePool))
}

func (s *ManagerTestSuite) TestDeleteNode_GetSizeError() {
	s.OnListKubernetesNodes(nil, errors.New("error")).Once()
	s.Error(s.manager.DeleteNode(s.nodePool, "deleteme"))
}

func (s *ManagerTestSuite) TestDeleteNode_DeleteError() {
	s.manager.cache.SetNodeGroupSize(s.nodePool.Id(), 2)
	s.manager.cache.SetNodeGroupTargetSize(s.nodePool.Id(), 2)
	s.OnDeleteKubernetesNode("notfound", errors.New("error")).Once()

	s.Error(s.manager.DeleteNode(s.nodePool, "notfound"))
	size, found := s.manager.cache.GetNodeGroupSize(s.nodePool.Id())
	s.True(found)
	s.Equal(2, size)
}

func (s *ManagerTestSuite) TestDeleteNode_WaitError() {
	s.manager.cache.SetNodeGroupSize(s.nodePool.Id(), 2)
	s.manager.cache.SetNodeGroupTargetSize(s.nodePool.Id(), 2)
	s.OnDeleteKubernetesNode("testnode", nil).Once()
	s.OnGetKubernetesNodePool(newKubernetesNodePool(ionos.Updating, 1), nil).Twice()
	s.OnGetKubernetesNodePool(nil, errors.New("error")).Once()

	s.Error(s.manager.DeleteNode(s.nodePool, "testnode"))
	size, found := s.manager.cache.GetNodeGroupSize(s.nodePool.Id())
	s.True(found)
	s.Equal(2, size)
	_, found = s.manager.cache.GetNodeGroupTargetSize(s.nodePool.Id())
	s.False(found)
}

func (s *ManagerTestSuite) TestDeleteNode_OK() {
	s.manager.cache.SetNodeGroupSize(s.nodePool.Id(), 2)
	s.manager.cache.SetNodeGroupTargetSize(s.nodePool.Id(), 2)
	s.manager.cache.SetInstancesCacheForNodeGroup(s.nodePool.Id(), []cloudprovider.Instance{
		newInstance("testnode"), newInstance("othernode"),
	})
	s.OnDeleteKubernetesNode("testnode", nil).Once()
	s.OnGetKubernetesNodePool(newKubernetesNodePool(ionos.Updating, 1), nil).Times(3)
	s.OnGetKubernetesNodePool(newKubernetesNodePool(ionos.Active, 1), nil).Once()

	s.NoError(s.manager.DeleteNode(s.nodePool, "testnode"))
	size, found := s.manager.cache.GetNodeGroupSize(s.nodePool.Id())
	s.True(found)
	s.Equal(1, size)
	_, found = s.manager.cache.GetNodeGroupTargetSize(s.nodePool.Id())
	s.False(found)
	s.Nil(s.manager.cache.GetNodeGroupForNode("testnode"))
}

func (s *ManagerTestSuite) TestInitExplicitNodeGroups_InvalidFormat() {
	s.Error(s.manager.initExplicitNodeGroups([]string{"in:valid"}))
}

func (s *ManagerTestSuite) TestInitExplicitNodeGroups_InvalidMinValue() {
	s.Error(s.manager.initExplicitNodeGroups([]string{"invalid:3:test"}))
}

func (s *ManagerTestSuite) TestInitExplicitNodeGroups_InvalidMaxValue() {
	s.Error(s.manager.initExplicitNodeGroups([]string{"1:invalid:test"}))
}

func (s *ManagerTestSuite) TestInitExplicitNodeGroups_InvalidIDValue() {
	s.Error(s.manager.initExplicitNodeGroups([]string{"1:3:invalid"}))
}

func (s *ManagerTestSuite) TestInitExplicitNodeGroups_GetNodePoolError() {
	id := uuid.NewString()
	s.nodePool.id = id
	s.OnGetKubernetesNodePool(nil, errors.New("error")).Once()

	s.Error(s.manager.initExplicitNodeGroups([]string{"1:3:" + id}))
}

func (s *ManagerTestSuite) TestInitExplicitNodeGroups_ListNodesError() {
	id := uuid.NewString()
	s.nodePool.id = id
	kNodePool := newKubernetesNodePool(ionos.Active, 2)
	s.OnGetKubernetesNodePool(kNodePool, nil).Once()
	s.OnListKubernetesNodes(nil, errors.New("error")).Once()

	s.Error(s.manager.initExplicitNodeGroups([]string{"1:3:" + id}))
}

func (s *ManagerTestSuite) TestInitExplicitNodeGroups_OK() {
	id := uuid.NewString()
	s.nodePool.id = id
	kNodePool := newKubernetesNodePool(ionos.Active, 2)
	s.OnGetKubernetesNodePool(kNodePool, nil).Once()
	s.OnListKubernetesNodes(&ionos.KubernetesNodes{
		Items: &[]ionos.KubernetesNode{
			newKubernetesNode("node-1", K8sNodeStateReady),
			newKubernetesNode("node-2", K8sNodeStateReady),
		},
	}, nil).Once()

	s.NoError(s.manager.initExplicitNodeGroups([]string{"1:3:" + id}))
	s.Equal([]cloudprovider.NodeGroup{&nodePool{
		id:      id,
		min:     1,
		max:     3,
		manager: s.manager,
	}}, s.manager.cache.GetNodeGroups())
	size, found := s.manager.cache.GetNodeGroupTargetSize(id)
	s.True(found)
	s.Equal(2, size)
	size, found = s.manager.cache.GetNodeGroupSize(id)
	s.True(found)
	s.Equal(2, size)
}
