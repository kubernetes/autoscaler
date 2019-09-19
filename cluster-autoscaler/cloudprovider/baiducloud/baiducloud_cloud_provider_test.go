/*
Copyright 2018 The Kubernetes Authors.

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

package baiducloud

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

var testBaiducloudManager = &BaiducloudManager{
	cloudConfig: &CloudConfig{},
	cceClient:   nil,
	asgs:        newAutoScalingGroups(&CloudConfig{}, nil),
	interrupt:   nil,
}

func testProvider(t *testing.T, m *BaiducloudManager) *baiducloudCloudProvider {
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	discoveryOpts := cloudprovider.NodeGroupDiscoveryOptions{
		NodeGroupSpecs: []string{"1:10:k8s-worker-asg-1"},
	}
	provider, err := BuildBaiducloudCloudProvider(m, discoveryOpts, resourceLimiter)
	assert.NoError(t, err)
	return provider.(*baiducloudCloudProvider)
}

func TestBuildBaiduCloudProvider(t *testing.T) {
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	discoveryOpts := cloudprovider.NodeGroupDiscoveryOptions{
		NodeGroupSpecs: []string{"1:10:k8s-worker-asg-1"},
	}
	_, err := BuildBaiducloudCloudProvider(testBaiducloudManager, discoveryOpts, resourceLimiter)
	assert.NoError(t, err)
}

func TestName(t *testing.T) {
	provider := testProvider(t, testBaiducloudManager)
	assert.Equal(t, provider.Name(), cloudprovider.BaiducloudProviderName)
}

func TestNodeGroups(t *testing.T) {
	provider := testProvider(t, testBaiducloudManager)

	nodeGroups := provider.NodeGroups()
	assert.Equal(t, len(nodeGroups), 1)
	assert.Equal(t, nodeGroups[0].Id(), "k8s-worker-asg-1")
	assert.Equal(t, nodeGroups[0].MinSize(), 1)
	assert.Equal(t, nodeGroups[0].MaxSize(), 10)
}

func TestGPULabel(t *testing.T) {
	provider := testProvider(t, testBaiducloudManager)
	GPULabel := provider.GPULabel()
	assert.Equal(t, GPULabel, "baidu/nvidia_name")
}

func TestCleanup(t *testing.T) {
	provider := testProvider(t, testBaiducloudManager)
	err := provider.Cleanup()
	assert.NoError(t, err)
}

func TestRefresh(t *testing.T) {
	provider := testProvider(t, testBaiducloudManager)
	err := provider.Refresh()
	assert.NoError(t, err)
}
