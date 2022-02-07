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

package vultr

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/vultr/govultr"
)

func TestNodeGroup_Debug(t *testing.T) {
	client := &vultrClientMock{}
	ng := testData(client, &govultr.NodePool{
		NodeQuantity: 3,
		MinNodes:     1,
		MaxNodes:     10,
	})

	d := ng.Debug()
	exp := "node group ID: a (min:1 max:10)"
	assert.Equal(t, exp, d, "debug string do not match")
}

func TestNodeGroup_TargetSize(t *testing.T) {
	nodes := 5

	client := &vultrClientMock{}
	ng := testData(client, &govultr.NodePool{
		NodeQuantity: nodes,
	})

	size, err := ng.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, nodes, size, "target size is not correct")
}

func TestNodeGroup_IncreaseSize(t *testing.T) {
	client := &vultrClientMock{}

	t.Run("basic increase", func(t *testing.T) {
		nodeQuant := 2
		delta := 1
		ng := testData(client, &govultr.NodePool{NodeQuantity: nodeQuant, MinNodes: 2, MaxNodes: 3})

		newQaunt := nodeQuant + delta
		client.On("UpdateNodePool", context.Background(), ng.clusterID, ng.id,
			&govultr.NodePoolReqUpdate{NodeQuantity: newQaunt}).Return(&govultr.NodePool{NodeQuantity: newQaunt}, nil).Once()

		err := ng.IncreaseSize(delta)
		assert.NoError(t, err)
	})

	t.Run("negative increase", func(t *testing.T) {
		numberOfNodes := 3
		delta := -1
		client := &vultrClientMock{}
		ng := testData(client, &govultr.NodePool{
			NodeQuantity: numberOfNodes,
		})
		err := ng.IncreaseSize(delta)

		exp := fmt.Errorf("delta must be positive, have: %d", delta)
		assert.EqualError(t, err, exp.Error(), "size increase must be positive")
	})

	t.Run("zero increase", func(t *testing.T) {
		numberOfNodes := 3
		delta := 0
		client := &vultrClientMock{}
		ng := testData(client, &govultr.NodePool{
			NodeQuantity: numberOfNodes,
		})

		exp := fmt.Errorf("delta must be positive, have: %d", delta)

		err := ng.IncreaseSize(delta)
		assert.EqualError(t, err, exp.Error(), "size increase must be positive")
	})

	t.Run("large increase above maximum", func(t *testing.T) {
		nodes := 15
		delta := 10

		client := &vultrClientMock{}
		ng := testData(client, &govultr.NodePool{
			NodeQuantity: nodes,
		})

		exp := fmt.Errorf("size increase is too large. current: %d desired: %d max: %d",
			nodes, nodes+delta, ng.MaxSize())

		err := ng.IncreaseSize(delta)
		assert.EqualError(t, err, exp.Error(), "size increase is too large")
	})
}

func TestNodeGroup_DecreaseTargetSize(t *testing.T) {
	t.Run("basic decrease", func(t *testing.T) {
		client := &vultrClientMock{}

		nodeQuant := 3
		delta := -1
		ng := testData(client, &govultr.NodePool{NodeQuantity: nodeQuant, MinNodes: 2, MaxNodes: 3})

		newQaunt := nodeQuant + delta
		client.On("UpdateNodePool", context.Background(), ng.clusterID, ng.id,
			&govultr.NodePoolReqUpdate{NodeQuantity: newQaunt}).Return(&govultr.NodePool{NodeQuantity: newQaunt}, nil).Once()

		err := ng.DecreaseTargetSize(delta)
		assert.NoError(t, err)
	})

	t.Run("positive decrease", func(t *testing.T) {
		nodes := 5
		client := &vultrClientMock{}
		ng := testData(client, &govultr.NodePool{
			NodeQuantity: nodes,
		})

		delta := 1
		err := ng.DecreaseTargetSize(delta)

		exp := fmt.Errorf("delta must be negative, have: %d", delta)
		assert.EqualError(t, err, exp.Error(), "size decrease must be negative")
	})

	t.Run("small decrease below minimum", func(t *testing.T) {
		delta := -2
		numberOfNodes := 3
		client := &vultrClientMock{}
		ng := testData(client, &govultr.NodePool{
			NodeQuantity: numberOfNodes,
			MinNodes:     2,
			MaxNodes:     5,
		})

		exp := fmt.Errorf("size decrease is too small. current: %d desired: %d min: %d",
			numberOfNodes, numberOfNodes+delta, ng.MinSize())
		err := ng.DecreaseTargetSize(delta)
		assert.EqualError(t, err, exp.Error(), "size decrease is too small")
	})
}

func TestNodeGroup_Nodes(t *testing.T) {
	client := &vultrClientMock{}
	ng := testData(client, &govultr.NodePool{
		Nodes: []govultr.Node{
			{
				ID: "a-1",
			},
			{
				ID: "a-2",
			},
		},
	})

	instances := []cloudprovider.Instance{
		{
			Id: "vultr://a-1",
		},
		{
			Id: "vultr://a-2",
		},
	}

	nodes, err := ng.Nodes()
	assert.NoError(t, err)
	assert.Equal(t, instances, nodes, "nodes do not match")
}

func TestNodeGroup_DeleteNodes(t *testing.T) {

	t.Run("deleting node", func(t *testing.T) {

		ctx := context.Background()
		client := &vultrClientMock{}
		ng := testData(client, &govultr.NodePool{NodeQuantity: 2, MinNodes: 2, MaxNodes: 3, Nodes: []govultr.Node{{ID: "a"}}})

		nodes := []*apiv1.Node{
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{nodeIDLabel: "a"}}},
		}

		client.On("DeleteNodePoolInstance", ctx, ng.clusterID, ng.id, "a").Return(nil).Once()

		err := ng.DeleteNodes(nodes)
		assert.NoError(t, err)
	})

	t.Run("delete failure", func(t *testing.T) {
		ctx := context.Background()
		client := &vultrClientMock{}
		ng := testData(client, &govultr.NodePool{NodeQuantity: 2, MinNodes: 2, MaxNodes: 3, Nodes: []govultr.Node{{ID: "a"}}})

		nodes := []*apiv1.Node{
			{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{nodeIDLabel: "a"}}},
		}

		client.On("DeleteNodePoolInstance", ctx, ng.clusterID, ng.id, "a").Return(errors.New("error")).Once()

		err := ng.DeleteNodes(nodes)
		assert.Error(t, err)
	})
}

func TestNodeGroup_Exist(t *testing.T) {
	client := &vultrClientMock{}
	nodeGroup := testData(client, &govultr.NodePool{MinNodes: 1, MaxNodes: 2})

	assert.Equal(t, true, nodeGroup.Exist(), "nodegroup should exist")

}

func testData(client vultrClient, np *govultr.NodePool) *NodeGroup {

	return &NodeGroup{
		id:        "a",
		clusterID: "a",
		client:    client,
		nodePool:  np,
		minSize:   np.MinNodes,
		maxSize:   np.MaxNodes,
	}
}
