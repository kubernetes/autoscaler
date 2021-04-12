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

package bizflycloud

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/bizflycloud/gobizfly"
)

// Due to newManager require authenticate with our server, we will use newManagerTest for simple test case
func newManagerTest(configReader io.Reader) (*Manager, error) {
	cfg := &Config{}
	if configReader != nil {
		body, err := ioutil.ReadAll(configReader)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(body, cfg)
		if err != nil {
			return nil, err
		}
	}

	if cfg.Token == "" {
		return nil, errors.New("access token is not provided")
	}
	if cfg.ClusterID == "" {
		return nil, errors.New("cluster ID is not provided")
	}

	bizflyClient, err := gobizfly.NewClient()
	if err != nil {
		return nil, fmt.Errorf("couldn't initialize Bizflycloud client: %s", err)
	}

	m := &Manager{
		client:     bizflyClient.KubernetesEngine,
		clusterID:  cfg.ClusterID,
		nodeGroups: make([]*NodeGroup, 0),
	}

	return m, nil
}

func TestNewManager(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "token": "123123123", "url": "https://manage.bizflycloud.vn", "version": "test"}`
		manager, err := newManagerTest(bytes.NewBufferString(cfg))
		assert.NoError(t, err)
		assert.Equal(t, manager.clusterID, "123456", "cluster ID does not match")
	})
}

func TestBizflyCloudManager_Refresh(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &bizflyClientMock{}
		cfg := `{"cluster_id": "123456", "token": "123123123", "url": "https://manage.bizflycloud.vn", "version": "test"}`
		manager, err := newManagerTest(bytes.NewBufferString(cfg))
		assert.NoError(t, err)
		ctx := context.Background()

		client.On("Get", ctx, manager.clusterID, nil).Return(&gobizfly.FullCluster{
			ExtendedCluster: gobizfly.ExtendedCluster{
				Cluster: gobizfly.Cluster{
					UID:              "1",
					Name:             "test-1",
					VPCNetworkID:     "vpc-1",
					WorkerPoolsCount: 1,
				},
				WorkerPools: []gobizfly.ExtendedWorkerPool{},
			},
			Stat: gobizfly.ClusterStat{},
		}, nil).Once()

		manager.client = client
		assert.Equal(t, len(manager.nodeGroups), 0, "number of nodes do not match")
	})

}

func TestBizflyCloudManager_RefreshWithNodeSpec(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cfg := `{"cluster_id": "123456", "token": "123123123", "url": "https://manage.bizflycloud.vn", "version": "test"}`

		manager, err := newManagerTest(bytes.NewBufferString(cfg))
		assert.NoError(t, err)

		client := &bizflyClientMock{}
		ctx := context.Background()

		client.On("Get", ctx, manager.clusterID, nil).Return(&gobizfly.FullCluster{},
			nil,
		).Once()
		manager.client = client
		err = manager.Refresh()
		assert.NoError(t, err)
		assert.Equal(t, len(manager.nodeGroups), 0, "number of node groups do not match")
	})
}
