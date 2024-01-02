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
	"errors"
	"testing"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	civocloud "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/civo/civo-cloud-sdk-go"

	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "api_key": "123-123-123", "api_url": "https://api.civo.com", "region": "test"}`
		nodeGroupSpecs := []string{"1:10:workers"}
		nodeGroupDiscoveryOptions := cloudprovider.NodeGroupDiscoveryOptions{NodeGroupSpecs: nodeGroupSpecs}
		manager, err := newManager(bytes.NewBufferString(cfg), nodeGroupDiscoveryOptions)
		assert.NoError(t, err)
		assert.Equal(t, "123456", manager.clusterID, "cluster ID does not match")
		assert.Equal(t, nodeGroupDiscoveryOptions, manager.discoveryOpts, "node group discovery options do not match")
	})

	t.Run("empty api_key", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "api_key": "", "api_url": "https://api.civo.com", "region": "test"}`
		nodeGroupSpecs := []string{"1:10:workers"}
		nodeGroupDiscoveryOptions := cloudprovider.NodeGroupDiscoveryOptions{NodeGroupSpecs: nodeGroupSpecs}
		_, err := newManager(bytes.NewBufferString(cfg), nodeGroupDiscoveryOptions)
		assert.EqualError(t, err, errors.New("civo API Key was not provided").Error())
	})

	t.Run("empty cluster ID", func(t *testing.T) {
		cfg := `{"cluster_id": "", "api_key": "123-123-123", "api_url": "https://api.civo.com", "region": "test"}`
		nodeGroupSpecs := []string{"1:10:workers"}
		nodeGroupDiscoveryOptions := cloudprovider.NodeGroupDiscoveryOptions{NodeGroupSpecs: nodeGroupSpecs}
		_, err := newManager(bytes.NewBufferString(cfg), nodeGroupDiscoveryOptions)
		assert.EqualError(t, err, errors.New("cluster ID was not provided").Error())
	})
}

func TestCivoManager_Refresh(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "api_key": "123-123-123", "api_url": "https://api.civo.com", "region": "test"}`
		nodeGroupSpecs := []string{"1:10:workers"}
		nodeGroupDiscoveryOptions := cloudprovider.NodeGroupDiscoveryOptions{NodeGroupSpecs: nodeGroupSpecs}
		manager, err := newManager(bytes.NewBufferString(cfg), nodeGroupDiscoveryOptions)
		assert.NoError(t, err)

		client := &civoClientMock{}

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

		client.On("FindInstanceSizes", "small").Return(
			&civocloud.InstanceSize{
				Name:     "small",
				CPUCores: 1,
			}, nil,
		).Times(10)

		manager.client = client
		err = manager.Refresh()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(manager.nodeGroups), "number of node groups do not match")
	})
}

func TestCivoManager_RefreshWithNodeSpec(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "api_key": "123-123-123", "api_url": "https://api.civo.com", "region": "test"}`
		nodeGroupSpecs := []string{"1:10:workers"}
		nodeGroupDiscoveryOptions := cloudprovider.NodeGroupDiscoveryOptions{NodeGroupSpecs: nodeGroupSpecs}
		manager, err := newManager(bytes.NewBufferString(cfg), nodeGroupDiscoveryOptions)
		assert.NoError(t, err)

		client := &civoClientMock{}

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

		client.On("FindInstanceSizes", "small").Return(
			&civocloud.InstanceSize{
				Name:     "small",
				CPUCores: 1,
			}, nil,
		).Times(10)

		manager.client = client
		err = manager.Refresh()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(manager.nodeGroups), "number of node groups do not match")
		assert.Equal(t, 1, manager.nodeGroups[0].minSize, "minimum node for node group does not match")
		assert.Equal(t, 10, manager.nodeGroups[0].maxSize, "maximum node for node group does not match")
	})
}

func TestCivoManager_RefreshWithNodeSpecPool(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "api_key": "123-123-123", "api_url": "https://api.civo.com", "region": "test"}`
		nodeGroupSpecs := []string{"1:5:pool-1", "5:10:pool-2"}
		nodeGroupDiscoveryOptions := cloudprovider.NodeGroupDiscoveryOptions{NodeGroupSpecs: nodeGroupSpecs}
		manager, err := newManager(bytes.NewBufferString(cfg), nodeGroupDiscoveryOptions)
		assert.NoError(t, err)

		client := &civoClientMock{}

		client.On("ListKubernetesClusterPools", manager.clusterID).Return(
			[]civocloud.KubernetesPool{
				{
					ID:            "pool-1",
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
					ID:            "pool-2",
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

		client.On("FindInstanceSizes", "small").Return(
			&civocloud.InstanceSize{
				Name:     "small",
				CPUCores: 1,
			}, nil,
		).Times(10)

		manager.client = client
		err = manager.Refresh()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(manager.nodeGroups), "number of node groups do not match")
		assert.Equal(t, 1, manager.nodeGroups[0].minSize, "minimum node for node group does not match")
		assert.Equal(t, 5, manager.nodeGroups[0].maxSize, "maximum node for node group does not match")
		assert.Equal(t, 5, manager.nodeGroups[1].minSize, "minimum node for node group does not match")
		assert.Equal(t, 10, manager.nodeGroups[1].maxSize, "maximum node for node group does not match")
	})
}
