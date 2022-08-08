/*
Copyright 2022 The Kubernetes Authors.

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

package civo

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	civocloud "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/civo/civo-cloud-sdk-go"
)

type civoClientMock struct {
	mock.Mock
}

func (m *civoClientMock) ListKubernetesClusterPools(clusterID string) ([]civocloud.KubernetesPool, error) {
	args := m.Called(clusterID)
	return args.Get(0).([]civocloud.KubernetesPool), args.Error(1)
}

func (m *civoClientMock) UpdateKubernetesClusterPool(cid, pid string, config *civocloud.KubernetesClusterPoolUpdateConfig) (*civocloud.KubernetesPool, error) {
	args := m.Called(cid, pid, config)
	return args.Get(0).(*civocloud.KubernetesPool), args.Error(1)
}

func (m *civoClientMock) DeleteKubernetesClusterPoolInstance(clusterID, poolID, instanceID string) (*civocloud.SimpleResponse, error) {
	args := m.Called(clusterID, poolID, instanceID)
	return args.Get(0).(*civocloud.SimpleResponse), args.Error(1)
}

func testCloudProvider(t *testing.T, client *civoClientMock) *civoCloudProvider {
	cfg := `{"cluster_id": "123456", "api_key": "123-123-123", "api_url": "https://api.civo.com", "region": "test"}`

	nodeGroupSpecs := []string{"1:10:workers"}
	nodeGroupDiscoveryOptions := cloudprovider.NodeGroupDiscoveryOptions{NodeGroupSpecs: nodeGroupSpecs}
	manager, err := newManager(bytes.NewBufferString(cfg), nodeGroupDiscoveryOptions)
	assert.NoError(t, err)
	rl := &cloudprovider.ResourceLimiter{}

	// fill the test provider with some example
	if client == nil {
		client = &civoClientMock{}

		client.On("ListKubernetesClusterPools", manager.clusterID).Return(
			[]civocloud.KubernetesPool{
				{
					ID:            "1",
					Count:         2,
					Size:          "small",
					InstanceNames: []string{"test-1", "test-2"},
					Instances: []civocloud.KubernetesInstance{
						{
							ID:       "1",
							Hostname: "test-1",
							Status:   "ACTIVE",
						},
						{
							ID:       "2",
							Hostname: "test-1",
							Status:   "ACTIVE",
						},
					},
				},
				{
					ID:            "2",
					Count:         2,
					Size:          "small",
					InstanceNames: []string{"test-1", "test-2"},
					Instances: []civocloud.KubernetesInstance{
						{
							ID:       "3",
							Hostname: "test-3",
							Status:   "ACTIVE",
						},
						{
							ID:       "4",
							Hostname: "test-4",
							Status:   "BUILDING",
						},
					},
				},
			},
			nil,
		).Once()
	}

	manager.client = client

	provider, err := newCivoCloudProvider(manager, rl)
	assert.NoError(t, err)
	return provider
}

func TestNewCivoCloudProvider(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_ = testCloudProvider(t, nil)
	})
}

func TestCivoCloudProvider_Name(t *testing.T) {
	provider := testCloudProvider(t, nil)

	t.Run("success", func(t *testing.T) {
		name := provider.Name()
		assert.Equal(t, cloudprovider.CivoProviderName, name, "provider name doesn't match")
	})
}

func TestCivoCloudProvider_NodeGroups(t *testing.T) {
	provider := testCloudProvider(t, nil)

	t.Run("success", func(t *testing.T) {
		nodegroups := provider.NodeGroups()
		assert.Equal(t, len(nodegroups), 2, "number of node groups does not match")
		nodes, _ := nodegroups[0].Nodes()
		assert.Equal(t, len(nodes), 2, "number of nodes in workers node group does not match")

	})

	t.Run("zero groups", func(t *testing.T) {
		provider.manager.nodeGroups = []*NodeGroup{}
		nodes := provider.NodeGroups()
		assert.Equal(t, len(nodes), 0, "number of nodes do not match")
	})
}

func TestCivoCloudProvider_NodeGroupForNode(t *testing.T) {
	cfg := `{"cluster_id": "123456", "api_key": "123-123-123", "api_url": "https://api.civo.com", "region": "test"}`

	nodeGroupSpecs := []string{"1:10:workers"}
	nodeGroupDiscoveryOptions := cloudprovider.NodeGroupDiscoveryOptions{NodeGroupSpecs: nodeGroupSpecs}
	manager, err := newManager(bytes.NewBufferString(cfg), nodeGroupDiscoveryOptions)
	assert.NoError(t, err)

	t.Run("success", func(t *testing.T) {
		client := &civoClientMock{}
		client.On("ListKubernetesClusterPools", manager.clusterID).Return(
			[]civocloud.KubernetesPool{
				{
					ID:    "1",
					Count: 2,
					Instances: []civocloud.KubernetesInstance{
						{
							ID:       "11",
							Hostname: "kube-node-11",
							Status:   "ACTIVE",
						},
						{
							ID:       "22",
							Hostname: "kube-node-22",
							Status:   "ACTIVE",
						},
					},
				},
				{
					ID:    "2",
					Count: 2,
					Instances: []civocloud.KubernetesInstance{
						{
							ID:       "111",
							Hostname: "kube-node-111",
							Status:   "ACTIVE",
						},
						{
							ID:       "222",
							Hostname: "kube-node-222",
							Status:   "ACTIVE",
						},
					},
				},
			},
			nil,
		).Once()

		provider := testCloudProvider(t, client)

		// let's get the nodeGroup for the node with ID 11
		node := &apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: "civo://11",
			},
		}

		nodeGroup, err := provider.NodeGroupForNode(node)
		require.NoError(t, err)
		require.NotNil(t, nodeGroup)
		require.Equal(t, nodeGroup.Id(), "1", "node group ID does not match")
	})

	t.Run("node does not exist", func(t *testing.T) {
		client := &civoClientMock{}
		client.On("ListKubernetesClusterPools", manager.clusterID).Return(
			[]civocloud.KubernetesPool{
				{
					ID:    "1",
					Count: 2,
					Instances: []civocloud.KubernetesInstance{
						{
							ID:       "11",
							Hostname: "kube-node-11",
							Status:   "ACTIVE",
						},
						{
							ID:       "22",
							Hostname: "kube-node-22",
							Status:   "ACTIVE",
						},
					},
				},
				{
					ID:    "2",
					Count: 2,
					Instances: []civocloud.KubernetesInstance{
						{
							ID:       "111",
							Hostname: "kube-node-111",
							Status:   "ACTIVE",
						},
						{
							ID:       "222",
							Hostname: "kube-node-222",
							Status:   "ACTIVE",
						},
					},
				},
			},
			nil,
		).Once()

		provider := testCloudProvider(t, client)

		node := &apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: "civo://non-existing-node",
			},
		}

		nodeGroup, err := provider.NodeGroupForNode(node)
		require.NoError(t, err)
		require.Nil(t, nodeGroup)
	})
}
