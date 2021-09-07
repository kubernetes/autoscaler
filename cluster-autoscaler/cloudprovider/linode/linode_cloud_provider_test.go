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

package linode

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/linode/linodego"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
)

func TestBuildLinode(t *testing.T) {
	token := "bogus"
	restoreToken := testEnvVar(linodeTokenEnvVar, token)
	defer restoreToken()

	clusterID := 123
	restoreClusterID := testEnvVar(lkeClusterIDEnvVar, strconv.Itoa(clusterID))
	defer restoreClusterID()

	lcp := BuildLinode(config.AutoscalingOptions{},
		cloudprovider.NodeGroupDiscoveryOptions{},
		&cloudprovider.ResourceLimiter{})

	assert.NotNil(t, lcp)
}

func TestCloudProvider_newLinodeCloudProvider(t *testing.T) {
	// test ok on correctly creating a linode provider
	rl := &cloudprovider.ResourceLimiter{}
	cfg := strings.NewReader(`{
    "clusterID": 456456,
    "token": "123123123"
}`)
	_, err := newLinodeCloudProvider(cfg, rl)
	assert.NoError(t, err)

	// test error on creating a linode provider when config is bad
	cfg = strings.NewReader("{bad config}")
	_, err = newLinodeCloudProvider(cfg, rl)
	assert.Error(t, err)
}

func TestCloudProvider_NodeGroups(t *testing.T) {
	ctx := context.TODO()
	client := &linodeClientMock{}
	cfg := strings.NewReader(`{
    "clusterID": 456456,
    "token": "123123123"
}`)
	m, _ := newManager(cfg)
	m.client = client
	lcp := linodeCloudProvider{manager: m}

	pools := []linodego.LKEClusterPool{
		// will be ignored, as autoscaler is disabled
		makeMockNodePool(123000, makeTestNodePoolNodes(581, 583), linodego.LKEClusterPoolAutoscaler{
			Min:     3,
			Max:     3,
			Enabled: false,
		}),
		makeMockNodePool(124000, makeTestNodePoolNodes(132, 133), linodego.LKEClusterPoolAutoscaler{
			Min:     1,
			Max:     2,
			Enabled: true,
		}),
		makeMockNodePool(125000, makeTestNodePoolNodes(127, 130), linodego.LKEClusterPoolAutoscaler{
			Min:     3,
			Max:     6,
			Enabled: true,
		}),
	}
	client.On("ListLKEClusterPools", ctx, m.config.ClusterID, nil).Return(pools, nil).Once()

	err := lcp.Refresh()
	assert.NoError(t, err)

	ngs := lcp.NodeGroups()
	assert.Equal(t, 2, len(ngs))
	assert.NotContains(t, ngs, nodeGroupFromPool(client, m.config.ClusterID, &pools[0]))
	assert.Contains(t, ngs, nodeGroupFromPool(client, m.config.ClusterID, &pools[1]))
	assert.Contains(t, ngs, nodeGroupFromPool(client, m.config.ClusterID, &pools[2]))
}

func TestCloudProvider_NodeGroupForNode(t *testing.T) {
	ctx := context.TODO()
	client := &linodeClientMock{}
	cfg := strings.NewReader(`{
		"clusterID": 12345,
		"token": "123123123"
	}`)
	m, _ := newManager(cfg)
	m.client = client
	lcp := linodeCloudProvider{manager: m}

	pools := []linodego.LKEClusterPool{
		makeMockNodePool(123000, makeTestNodePoolNodes(123, 125), linodego.LKEClusterPoolAutoscaler{
			Min:     1,
			Max:     5,
			Enabled: true,
		}),
		makeMockNodePool(124000, makeTestNodePoolNodes(127, 130), linodego.LKEClusterPoolAutoscaler{
			Min:     3,
			Max:     6,
			Enabled: true,
		}),
	}
	client.On("ListLKEClusterPools", ctx, m.config.ClusterID, nil).Return(pools, nil).Once()

	err := lcp.Refresh()
	assert.NoError(t, err)

	t.Run("fails on malformed ID", func(t *testing.T) {
		ng, err := lcp.NodeGroupForNode(&apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: "linode://invalid",
			},
		})
		assert.Error(t, err)
		assert.Nil(t, ng)
	})

	t.Run("fails when node does not exist", func(t *testing.T) {
		ng, err := lcp.NodeGroupForNode(&apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: "linode://131",
			},
		})
		assert.NoError(t, err)
		assert.Nil(t, ng)
	})

	t.Run("successfully retrieves the nodegroup", func(t *testing.T) {
		ng, err := lcp.NodeGroupForNode(&apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: "linode://125",
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, nodeGroupFromPool(client, 12345, &pools[0]), ng)
	})
}

func TestCloudProvider_others(t *testing.T) {
	resourceLimiter := &cloudprovider.ResourceLimiter{}
	lcp := &linodeCloudProvider{
		resourceLimiter: resourceLimiter,
	}
	assert.Equal(t, cloudprovider.LinodeProviderName, lcp.Name())
	_, err := lcp.GetAvailableMachineTypes()
	assert.Error(t, err)
	_, err = lcp.NewNodeGroup("", nil, nil, nil, nil)
	assert.Error(t, err)
	rl, err := lcp.GetResourceLimiter()
	assert.NoError(t, err)
	assert.Equal(t, resourceLimiter, rl)
	assert.Equal(t, "", lcp.GPULabel())
	assert.Nil(t, lcp.GetAvailableGPUTypes())
	assert.Nil(t, lcp.Cleanup())
	_, err2 := lcp.Pricing()
	assert.Error(t, err2)
}
