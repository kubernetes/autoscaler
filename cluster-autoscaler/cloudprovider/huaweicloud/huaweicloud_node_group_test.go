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
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	th "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/huaweicloud/huawei-cloud-sdk-go/testhelper"
	"sync"
	"testing"
	"time"
)

const (
	waitTimeStep        = 100 * time.Millisecond
	nodePoolName        = "crc-nodepool"
	nodePoolUID         = "deb25e14-6558-11ea-97c2-0255ac101d4a"
	minNodeCount        = 1
	maxNodeCount        = 10
	increaseSize        = 1
	decreaseSize        = -1
	nodePoolNodeCount   = 5
	nodeIdToBeDeleted   = "c1b6ff0c-6ee1-11ea-befa-0255ac101d4c"
	nodeNameToBeDeleted = "nodeToBeDeleted"
)

func createTestNodeGroup(manager *huaweicloudCloudManager) *NodeGroup {
	size := nodePoolNodeCount
	return &NodeGroup{
		huaweiCloudManager: manager,
		clusterUpdateMutex: &sync.Mutex{},
		nodePoolName:       nodePoolName,
		nodePoolId:         nodePoolUID,
		clusterName:        clusterUUID,
		autoscalingEnabled: true,
		minNodeCount:       minNodeCount,
		maxNodeCount:       maxNodeCount,
		targetSize:         &size,
		timeIncrement:      waitTimeStep,
	}
}

func TestNodeGroup_MaxSize(t *testing.T) {
	th.CreateServer()
	defer th.ShutDownServer()

	manager := createTestHuaweicloudManager()
	ng := createTestNodeGroup(manager)

	size := ng.MaxSize()
	assert.Equal(t, maxNodeCount, size)
}

func TestNodeGroup_MinSize(t *testing.T) {
	th.CreateServer()
	defer th.ShutDownServer()

	manager := createTestHuaweicloudManager()
	ng := createTestNodeGroup(manager)

	size := ng.MinSize()
	assert.Equal(t, minNodeCount, size)
}

func TestNodeGroup_TargetSize(t *testing.T) {
	th.CreateServer()
	defer th.ShutDownServer()

	manager := createTestHuaweicloudManager()
	ng := createTestNodeGroup(manager)

	size, err := ng.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, nodePoolNodeCount, size)
}

func TestNodeGroup_waitForClusterStatus(t *testing.T) {
	th.CreateServer()
	defer th.ShutDownServer()
	register()

	manager := createTestHuaweicloudManager()
	ng := createTestNodeGroup(manager)

	err := ng.waitForClusterStatus(clusterStatusAvailable, waitForCompleteStatusTimout)
	assert.NoError(t, err)
}

func TestNodeGroup_IncreaseSize(t *testing.T) {
	th.CreateServer()
	defer th.ShutDownServer()
	register()

	manager := createTestHuaweicloudManager()
	ng := createTestNodeGroup(manager)

	err := ng.IncreaseSize(increaseSize)
	assert.NoError(t, err)
	curSize, _ := ng.TargetSize()
	assert.Equal(t, nodePoolNodeCount+increaseSize, curSize)
}

func TestNodeGroup_DeleteNodes(t *testing.T) {
	th.CreateServer()
	defer th.ShutDownServer()
	register()

	manager := createTestHuaweicloudManager()
	ng := createTestNodeGroup(manager)
	nodes := []*apiv1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: nodeNameToBeDeleted,
			},
			Spec: apiv1.NodeSpec{
				ProviderID: nodeIdToBeDeleted,
			},
		},
	}
	isTestingDeleteNodes = true
	getRequestCount = -1

	err := ng.DeleteNodes(nodes)
	assert.NoError(t, err)
	curSize, _ := ng.TargetSize()
	assert.Equal(t, nodePoolNodeCount+decreaseSize, curSize)

	isTestingDeleteNodes = false
	getRequestCount = 0
}

func TestNodeGroup_DecreaseTargetSize(t *testing.T) {
	th.CreateServer()
	defer th.ShutDownServer()
	register()

	manager := createTestHuaweicloudManager()
	ng := createTestNodeGroup(manager)

	err := ng.DecreaseTargetSize(decreaseSize)
	assert.NoError(t, err)
	curSize, _ := ng.TargetSize()
	assert.Equal(t, nodePoolNodeCount+decreaseSize, curSize)
}

func TestNodeGroup_Id(t *testing.T) {
	th.CreateServer()
	defer th.ShutDownServer()

	manager := createTestHuaweicloudManager()
	ng := createTestNodeGroup(manager)

	name := ng.Id()
	assert.Equal(t, nodePoolName, name)
}
