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
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	civocloud "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/civo/civo-cloud-sdk-go"
)

func TestNodeGroup_TargetSize(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		numberOfNodes := 3

		client := &civoClientMock{}
		ng := testNodeGroup(client, &civocloud.KubernetesPool{
			Count: numberOfNodes,
		}, 1, 10)

		size, err := ng.TargetSize()
		assert.NoError(t, err)
		assert.Equal(t, numberOfNodes, size, "target size is not correct")
	})
}

func TestNodeGroup_IncreaseSize(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		numberOfNodes := 3
		client := &civoClientMock{}
		ng := testNodeGroup(client, &civocloud.KubernetesPool{
			Count: numberOfNodes,
		}, 1, 10)

		delta := 2

		newCount := numberOfNodes + delta
		client.On("UpdateKubernetesClusterPool",
			ng.clusterID,
			ng.id,
			&civocloud.KubernetesClusterPoolUpdateConfig{
				Count:  newCount,
				Region: "test",
			},
		).Return(
			&civocloud.KubernetesPool{Count: newCount},
			nil,
		).Once()

		err := ng.IncreaseSize(delta)
		assert.NoError(t, err)
	})

	t.Run("successful increase to maximum", func(t *testing.T) {
		// Increase from 2 nodes to 3 (but 2 worker nodes which is the max)
		numberOfNodes := 2
		maxNodes := 3

		client := &civoClientMock{}
		ng := testNodeGroup(client, &civocloud.KubernetesPool{
			Count: numberOfNodes,
		}, 1, maxNodes)

		delta := 1

		newCount := numberOfNodes + delta
		client.On("UpdateKubernetesClusterPool",
			ng.clusterID,
			ng.id,
			&civocloud.KubernetesClusterPoolUpdateConfig{
				Count:  newCount,
				Region: "test",
			},
		).Return(
			&civocloud.KubernetesPool{Count: newCount},
			nil,
		).Once()

		err := ng.IncreaseSize(delta)
		assert.NoError(t, err)
	})

	t.Run("negative increase", func(t *testing.T) {
		numberOfNodes := 3
		client := &civoClientMock{}
		ng := testNodeGroup(client, &civocloud.KubernetesPool{
			Count: numberOfNodes,
		}, 1, 10)

		delta := -1
		err := ng.IncreaseSize(delta)
		exp := fmt.Errorf("delta must be positive, have: %d", delta)
		assert.EqualError(t, err, exp.Error(), "size increase must be positive")
	})

	t.Run("zero increase", func(t *testing.T) {
		numberOfNodes := 3
		client := &civoClientMock{}
		ng := testNodeGroup(client, &civocloud.KubernetesPool{
			Count: numberOfNodes,
		}, 1, 10)

		delta := 0
		err := ng.IncreaseSize(delta)
		exp := fmt.Errorf("delta must be positive, have: %d", delta)
		assert.EqualError(t, err, exp.Error(), "size increase must be positive")
	})

	t.Run("large increase above maximum", func(t *testing.T) {
		numberOfNodes := 95
		maxNodes := 100
		delta := 10
		client := &civoClientMock{}
		ng := testNodeGroup(client, &civocloud.KubernetesPool{
			Count: numberOfNodes,
		}, 1, maxNodes)

		exp := fmt.Errorf("size increase is too large. current: %d desired: %d max: %d",
			numberOfNodes, numberOfNodes+delta, ng.MaxSize())
		err := ng.IncreaseSize(delta)
		assert.EqualError(t, err, exp.Error(), "size increase is too large")
	})
}

func TestNodeGroup_DecreaseTargetSize(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		numberOfNodes := 5

		client := &civoClientMock{}
		ng := testNodeGroup(client, &civocloud.KubernetesPool{
			Count: numberOfNodes,
		}, 1, 10)

		delta := -2

		newCount := numberOfNodes + delta
		client.On("UpdateKubernetesClusterPool",
			ng.clusterID,
			ng.id,
			&civocloud.KubernetesClusterPoolUpdateConfig{
				Count:  newCount,
				Region: "test",
			},
		).Return(
			&civocloud.KubernetesPool{Count: newCount},
			nil,
		).Once()

		err := ng.DecreaseTargetSize(delta)
		assert.NoError(t, err)
	})

	t.Run("positive decrease", func(t *testing.T) {
		numberOfNodes := 5
		client := &civoClientMock{}
		ng := testNodeGroup(client, &civocloud.KubernetesPool{
			Count: numberOfNodes,
		}, 1, 10)

		delta := 1
		err := ng.DecreaseTargetSize(delta)

		exp := fmt.Errorf("delta must be negative, have: %d", delta)
		assert.EqualError(t, err, exp.Error(), "size decrease must be negative")
	})

	t.Run("zero decrease", func(t *testing.T) {
		numberOfNodes := 5
		client := &civoClientMock{}
		ng := testNodeGroup(client, &civocloud.KubernetesPool{
			Count: numberOfNodes,
		}, 1, 10)

		delta := 0
		exp := fmt.Errorf("delta must be negative, have: %d", delta)

		err := ng.DecreaseTargetSize(delta)
		assert.EqualError(t, err, exp.Error(), "size decrease must be negative")
	})

	t.Run("small decrease below minimum", func(t *testing.T) {
		delta := -2
		numberOfNodes := 2
		client := &civoClientMock{}
		ng := testNodeGroup(client, &civocloud.KubernetesPool{
			Count: numberOfNodes,
		}, 1, 5)

		exp := fmt.Errorf("size decrease is too small. current: %d desired: %d min: %d",
			numberOfNodes, numberOfNodes+delta, ng.MinSize())
		err := ng.DecreaseTargetSize(delta)
		assert.EqualError(t, err, exp.Error(), "size decrease is too small")
	})
}

func TestNodeGroup_DeleteNodes(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &civoClientMock{}
		ng := testNodeGroup(client, &civocloud.KubernetesPool{
			Count: 3,
		}, 1, 10)

		nodes := []*apiv1.Node{
			{Spec: apiv1.NodeSpec{ProviderID: "civo://1"}},
			{Spec: apiv1.NodeSpec{ProviderID: "civo://2"}},
			{Spec: apiv1.NodeSpec{ProviderID: "civo://3"}},
		}

		// this should be called three times (the number of nodes)
		client.On("DeleteKubernetesClusterPoolInstance",
			ng.clusterID,
			ng.id,
			"1",
		).Return(
			&civocloud.SimpleResponse{Result: "success"},
			nil,
		).Once()
		client.On("DeleteKubernetesClusterPoolInstance",
			ng.clusterID,
			ng.id,
			"2",
		).Return(
			&civocloud.SimpleResponse{Result: "success"},
			nil,
		).Once()
		client.On("DeleteKubernetesClusterPoolInstance",
			ng.clusterID,
			ng.id,
			"3",
		).Return(
			&civocloud.SimpleResponse{Result: "success"},
			nil,
		).Once()

		err := ng.DeleteNodes(nodes)
		assert.NoError(t, err)
	})

	t.Run("client deleting node fails", func(t *testing.T) {
		client := &civoClientMock{}
		ng := testNodeGroup(client, &civocloud.KubernetesPool{
			Count: 3,
		}, 1, 10)

		nodes := []*apiv1.Node{
			{Spec: apiv1.NodeSpec{ProviderID: "civo://1"}},
			{Spec: apiv1.NodeSpec{ProviderID: "civo://2"}},
			{Spec: apiv1.NodeSpec{ProviderID: "civo://3"}},
		}

		// client is called twice, first run is successfully but the second one
		// fails with a random error. In this case DeleteNodes() should return
		// immediately.
		client.On("DeleteKubernetesClusterPoolInstance",
			ng.clusterID,
			ng.id,
			"1",
		).Return(
			&civocloud.SimpleResponse{Result: "success"},
			nil,
		).Once()
		client.On("DeleteKubernetesClusterPoolInstance",
			ng.clusterID,
			ng.id,
			"2",
		).Return(
			&civocloud.SimpleResponse{},
			errors.New("random error"),
		).Once()

		err := ng.DeleteNodes(nodes)
		assert.Error(t, err)
	})
}

func TestNodeGroup_Nodes(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &civoClientMock{}
		ng := testNodeGroup(client, &civocloud.KubernetesPool{
			ID:    "1",
			Count: 5,
			Instances: []civocloud.KubernetesInstance{
				{
					ID:       "1",
					Hostname: "kube-node-1",
					Status:   "ACTIVE",
				},
				{
					ID:       "2",
					Hostname: "kube-node-2",
					Status:   "BUILDING",
				},
				{
					ID:       "3",
					Hostname: "kube-node-3",
					Status:   "DELETING",
				},
				{
					ID:       "4",
					Hostname: "kube-node-4",
					Status:   "BANANAS",
				},
				{
					ID:       "5",
					Hostname: "kube-node-5",
				},
			}}, 1, 10)

		exp := []cloudprovider.Instance{
			{
				Id: "civo://1",
				Status: &cloudprovider.InstanceStatus{
					State: cloudprovider.InstanceRunning,
				},
			},
			{
				Id: "civo://2",
				Status: &cloudprovider.InstanceStatus{
					State: cloudprovider.InstanceCreating,
				},
			},
			{
				Id: "civo://3",
				Status: &cloudprovider.InstanceStatus{
					State: cloudprovider.InstanceDeleting,
				},
			},
			{
				Id: "civo://4",
				Status: &cloudprovider.InstanceStatus{
					ErrorInfo: &cloudprovider.InstanceErrorInfo{
						ErrorClass:   cloudprovider.OtherErrorClass,
						ErrorCode:    "no-code-civo",
						ErrorMessage: "BANANAS",
					},
				},
			},
			{
				Id: "civo://5",
			},
		}

		nodes, err := ng.Nodes()
		assert.NoError(t, err)
		assert.Equal(t, exp, nodes, "nodes do not match")
	})

	t.Run("failure (nil node pool)", func(t *testing.T) {
		client := &civoClientMock{}
		ng := testNodeGroup(client, nil, 1, 10)

		_, err := ng.Nodes()
		assert.Error(t, err, "Nodes() should return an error")
	})
}

func TestNodeGroup_Debug(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &civoClientMock{}
		ng := testNodeGroup(client, &civocloud.KubernetesPool{Count: 2}, 1, 200)
		d := ng.Debug()
		exp := "cluster ID: 1 (min:1 max:200)"
		assert.Equal(t, exp, d, "debug string do not match")
	})
}

func TestNodeGroup_Exist(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &civoClientMock{}
		ng := testNodeGroup(client, &civocloud.KubernetesPool{Count: 3}, 1, 200)
		exist := ng.Exist()
		assert.Equal(t, true, exist, "node group should exist")
	})

	t.Run("failure", func(t *testing.T) {
		client := &civoClientMock{}
		ng := testNodeGroup(client, nil, 1, 200)
		exist := ng.Exist()
		assert.Equal(t, false, exist, "node group should not exist")
	})
}

func testNodeGroup(client nodeGroupClient, np *civocloud.KubernetesPool, min int, max int) *NodeGroup {
	Region = "test"
	return &NodeGroup{
		id:        "1",
		clusterID: "1",
		client:    client,
		nodePool:  np,
		minSize:   min,
		maxSize:   max,
	}
}
