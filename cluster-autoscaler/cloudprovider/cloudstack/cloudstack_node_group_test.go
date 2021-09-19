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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

func createASG() *asg {
	s := createMockService()
	s.On("ScaleCluster",
		testConfig.clusterID, workerCount).After(100 * time.Millisecond).
		Return(createScaleUpClusterDetails())
	s.On("RemoveNodesFromCluster",
		testConfig.clusterID, nodeIDs).After(100 * time.Millisecond).
		Return(createScaleDownClusterDetails())
	manager, _ := newManager(testConfig,
		withCKSService(s))
	return manager.asg
}

func TestMaxSize(t *testing.T) {
	asg := createASG()
	assert.Equal(t, testConfig.maxSize, asg.MaxSize())
}

func TestMinSize(t *testing.T) {
	asg := createASG()
	assert.Equal(t, testConfig.minSize, asg.MinSize())
}

func TestTargetSize(t *testing.T) {
	clusterDetails := createClusterDetails()
	asg := createASG()
	targetSize, err := asg.TargetSize()
	assert.Equal(t, clusterDetails.WorkerCount, targetSize)
	assert.Equal(t, nil, err)
}

func TestId(t *testing.T) {
	asg := createASG()
	assert.Equal(t, testConfig.clusterID, asg.Id())
}

func TestNodes(t *testing.T) {
	clusterDetails := createClusterDetails()
	asg := createASG()
	nodes, err := asg.Nodes()
	assert.Equal(t, len(clusterDetails.VirtualMachines), len(nodes))
	for i, node := range nodes {
		assert.Equal(t, clusterDetails.VirtualMachines[i].ID, node.Id)
	}
	assert.Equal(t, nil, err)
}

func TestExist(t *testing.T) {
	asg := createASG()
	assert.Equal(t, true, asg.Exist())
}

func TestAutoprovisioned(t *testing.T) {
	asg := createASG()
	assert.Equal(t, false, asg.Autoprovisioned())
}

func TestCreate(t *testing.T) {
	asg := createASG()
	_, err := asg.Create()
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

func TestDelete(t *testing.T) {
	asg := createASG()
	assert.Equal(t, cloudprovider.ErrNotImplemented, asg.Delete())
}

func TestTemplateNodeInfo(t *testing.T) {
	asg := createASG()
	_, err := asg.TemplateNodeInfo()
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

func testIncreaseSizeError(t *testing.T) {
	clusterDetails := createClusterDetails()

	// Above the max size
	asg := createASG()
	err := asg.IncreaseSize(100)
	assert.Equal(t, clusterDetails, asg.cluster)
	assert.NotEqual(t, nil, err)

	// Delta must be positive
	err = asg.IncreaseSize(-1)
	assert.Equal(t, clusterDetails, asg.cluster)
	assert.NotEqual(t, nil, err)
}

func testIncreaseSizeSuccess(t *testing.T) {
	clusterDetails := createClusterDetails()
	scaleUpClusterDetails := createScaleUpClusterDetails()
	scaleDownClusterDetails := createScaleDownClusterDetails()
	delta := clusterDetails.WorkerCount - scaleDownClusterDetails.WorkerCount
	asg := createASG()
	fmt.Println(asg.cluster.WorkerCount)
	err := asg.IncreaseSize(delta)
	assert.Equal(t, scaleUpClusterDetails, asg.cluster)
	assert.Equal(t, nil, err)
}

func TestIncreaseSize(t *testing.T) {
	t.Run("testIncreaseSizeError", testIncreaseSizeError)
	t.Run("testIncreaseSizeSuccess", testIncreaseSizeSuccess)
}

func TestDecreaseTargetSize(t *testing.T) {
	asg := createASG()
	err := asg.DecreaseTargetSize(1)
	assert.NotEqual(t, nil, err)
}

func testDeleteNodesError(t *testing.T) {
	nodes := []*apiv1.Node{
		{
			Status: v1.NodeStatus{
				NodeInfo: v1.NodeSystemInfo{
					SystemUUID: "vm1",
				},
			},
		},
		{
			Status: v1.NodeStatus{
				NodeInfo: v1.NodeSystemInfo{
					SystemUUID: "vm2",
				},
			},
		},
		{
			Status: v1.NodeStatus{
				NodeInfo: v1.NodeSystemInfo{
					SystemUUID: "vm3",
				},
			},
		},
	}

	// Can't go below minSize
	asg := createASG()
	err := asg.DeleteNodes(nodes)
	clusterDetails := createClusterDetails()
	assert.Equal(t, clusterDetails, asg.cluster)
	assert.NotEqual(t, nil, err)
}

func testDeleteNodesSuccess(t *testing.T) {
	asg := createASG()
	err := asg.DeleteNodes([]*apiv1.Node{
		{
			Status: v1.NodeStatus{
				NodeInfo: v1.NodeSystemInfo{
					SystemUUID: "vm2",
				},
			},
		},
		{
			Status: v1.NodeStatus{
				NodeInfo: v1.NodeSystemInfo{
					SystemUUID: "vm3",
				},
			},
		},
	})
	scaleDownClusterDetails := createScaleDownClusterDetails()
	assert.Equal(t, scaleDownClusterDetails, asg.cluster)
	assert.Equal(t, nil, err)
}

func TestDeleteNodes(t *testing.T) {
	t.Run("testDeleteNodesError", testDeleteNodesError)
	t.Run("testDeleteNodesSuccess", testDeleteNodesSuccess)
}
