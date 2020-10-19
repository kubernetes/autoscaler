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

package cloudstack

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
)

var (
	provider   *cloudStackCloudProvider
	testConfig = &clusterConfig{
		clusterID: "abc",
		minSize:   1,
		maxSize:   6,
	}
)

func init() {
	s := createMockService()
	s.On("Close").Return()
	manager, _ := newManager(testConfig,
		withCKSService(s))
	provider = &cloudStackCloudProvider{
		manager:         manager,
		resourceLimiter: &cloudprovider.ResourceLimiter{},
	}
}

func TestCreateClusterConfig(t *testing.T) {
	opts := config.AutoscalingOptions{
		NodeGroups: []string{"2:3:abcd"},
	}
	cfg, err := createClusterConfig(opts)

	assert.Equal(t, nil, err)
	assert.Equal(t, 2, cfg.minSize)
	assert.Equal(t, 3, cfg.maxSize)
	assert.Equal(t, "abcd", cfg.clusterID)

	opts.NodeGroups = []string{}
	cfg, err = createClusterConfig(opts)
	assert.NotEqual(t, nil, err)

	opts.NodeGroups = []string{"2:abcd"}
	cfg, err = createClusterConfig(opts)
	assert.NotEqual(t, nil, err)

	opts.NodeGroups = []string{"2:three:abcd"}
	cfg, err = createClusterConfig(opts)
	assert.NotEqual(t, nil, err)

	opts.NodeGroups = []string{"two:three:abcd"}
	cfg, err = createClusterConfig(opts)
	assert.NotEqual(t, nil, err)
}

func TestName(t *testing.T) {
	assert.Equal(t, cloudprovider.CloudStackProviderName, provider.Name())
}

func TestNodeGroups(t *testing.T) {
	asgs := provider.NodeGroups()
	assert.Equal(t, 1, len(asgs))
	assert.Equal(t, testConfig.clusterID, asgs[0].Id())
	assert.Equal(t, testConfig.maxSize, asgs[0].MaxSize())
	assert.Equal(t, testConfig.minSize, asgs[0].MinSize())
}

func testNodeExists(t *testing.T) {
	node := &v1.Node{
		Status: v1.NodeStatus{
			NodeInfo: v1.NodeSystemInfo{
				SystemUUID: "vm1",
			},
		},
	}
	asg, err := provider.NodeGroupForNode(node)
	assert.Equal(t, nil, err)
	assert.Equal(t, testConfig.clusterID, asg.Id())
	assert.Equal(t, testConfig.maxSize, asg.MaxSize())
	assert.Equal(t, testConfig.minSize, asg.MinSize())
}

func testNodeNotExist(t *testing.T) {
	node := &v1.Node{
		Status: v1.NodeStatus{
			NodeInfo: v1.NodeSystemInfo{
				SystemUUID: "vm5",
			},
		},
	}
	_, err := provider.NodeGroupForNode(node)
	assert.NotEqual(t, nil, err)
}

func TestNodeGroupForNode(t *testing.T) {
	t.Run("testNodeExists", testNodeExists)
	t.Run("testNodeNotExist", testNodeNotExist)
}

func TestGetAvailableMachineTypes(t *testing.T) {
	types, err := provider.GetAvailableMachineTypes()
	assert.Equal(t, availableMachineTypes, types)
	assert.Equal(t, nil, err)
}

func TestGPULabel(t *testing.T) {
	assert.Equal(t, GPULabel, provider.GPULabel())
}

func TestGetAvailableGPUTypes(t *testing.T) {
	assert.Equal(t, availableGPUTypes, provider.GetAvailableGPUTypes())
}

func TestPricing(t *testing.T) {
	_, err := provider.Pricing()
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

func TestNewNodeGroup(t *testing.T) {
	_, err := provider.NewNodeGroup("machineType", map[string]string{}, map[string]string{}, []v1.Taint{}, map[string]resource.Quantity{})
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

func TestCleanup(t *testing.T) {
	assert.Equal(t, nil, provider.Cleanup())
}

func TestGetResourceLimiter(t *testing.T) {
	rl, err := provider.GetResourceLimiter()
	assert.Equal(t, &cloudprovider.ResourceLimiter{}, rl)
	assert.Equal(t, nil, err)
}

func TestRefresh(t *testing.T) {
	asg := provider.manager.asg
	asg.cluster = createScaleUpClusterDetails()
	assert.Equal(t, nil, provider.Refresh())
	assert.Equal(t, createClusterDetails(), asg.cluster)
}
