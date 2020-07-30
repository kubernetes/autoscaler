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

package huaweicloud

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"testing"
)

func buildTestHuaweicloudCloudProvider() *huaweicloudCloudProvider {
	mgr := &huaweicloudCloudManager{}
	rl := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: int64(0), cloudprovider.ResourceNameMemory: int64(0)},
		map[string]int64{cloudprovider.ResourceNameCores: config.DefaultMaxClusterCores, cloudprovider.ResourceNameMemory: config.DefaultMaxClusterMemory})
	ng := createTestNodeGroup(mgr)

	hcp := &huaweicloudCloudProvider{
		huaweiCloudManager: mgr,
		resourceLimiter:    rl,
		nodeGroups:         []NodeGroup{*ng},
	}
	return hcp
}

func TestHuaweicloudCloudProvider_Name(t *testing.T) {
	hcp := buildTestHuaweicloudCloudProvider()
	name := hcp.Name()
	assert.Equal(t, cloudprovider.HuaweicloudProviderName, name)
}

func TestHuaweicloudCloudProvider_NodeGroups(t *testing.T) {
	hcp := buildTestHuaweicloudCloudProvider()
	ngs := hcp.NodeGroups()
	assert.Equal(t, len(hcp.nodeGroups), len(ngs))

	for _, ng := range ngs {
		_, ok := ng.(*NodeGroup)
		if !ok {
			assert.Error(t, fmt.Errorf("nodeGroups method returns NodeGroup slice which are not of type *huaweicloud.NodeGroup\n"))
		}
	}
}

func TestHuaweicloudCloudProvider_NodeGroupForNode(t *testing.T) {
	hcp := buildTestHuaweicloudCloudProvider()
	nodes := []*apiv1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"node-role.kubernetes.io/master": ""},
			},
		},
		{},
	}
	for _, node := range nodes {
		_, err := hcp.NodeGroupForNode(node)
		assert.NoError(t, err)
	}
}

func TestHuaweicloudCloudProvider_GetResourceLimiter(t *testing.T) {
	hcp := buildTestHuaweicloudCloudProvider()
	rl, err := hcp.GetResourceLimiter()
	assert.NoError(t, err)
	assert.Equal(t, int64(config.DefaultMaxClusterCores), rl.GetMax(cloudprovider.ResourceNameCores))
	assert.Equal(t, int64(config.DefaultMaxClusterMemory), rl.GetMax(cloudprovider.ResourceNameMemory))
	assert.Equal(t, int64(0), rl.GetMin(cloudprovider.ResourceNameCores))
	assert.Equal(t, int64(0), rl.GetMin(cloudprovider.ResourceNameMemory))
}

func TestHuaweicloudCloudProvider_GPULabel(t *testing.T) {
	hcp := buildTestHuaweicloudCloudProvider()
	labels := hcp.GPULabel()
	assert.Equal(t, GPULabel, labels)
}

func TestHuaweicloudCloudProvider_GetAvailableGPUTypes(t *testing.T) {
	hcp := buildTestHuaweicloudCloudProvider()
	availableGPU := hcp.GetAvailableGPUTypes()
	assert.Equal(t, availableGPUTypes, availableGPU)
}

func TestHuaweicloudCloudProvider_Append(t *testing.T) {
	hcp := buildTestHuaweicloudCloudProvider()
	originalSize := len(hcp.nodeGroups)
	newNgs := []NodeGroup{*createTestNodeGroup(&huaweicloudCloudManager{})}
	hcp.Append(newNgs)
	assert.Equal(t, originalSize+len(newNgs), len(hcp.nodeGroups))
}

func TestHuaweicloudCloudProvider_GetInstanceID(t *testing.T) {
	hcp := buildTestHuaweicloudCloudProvider()
	node := apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: nodeIdToBeDeleted,
		},
	}
	id := hcp.GetInstanceID(&node)
	assert.Equal(t, nodeIdToBeDeleted, id)
}
