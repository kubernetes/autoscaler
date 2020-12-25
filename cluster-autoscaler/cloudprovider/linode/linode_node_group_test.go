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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/linode/linodego"
)

func TestNodeGroup_IncreaseSize(t *testing.T) {
	client := linodeClientMock{}
	ctx := context.Background()
	poolOpts := linodego.LKEClusterPoolCreateOptions{
		Count: 1,
		Type:  "g6-standard-1",
	}
	ng := NodeGroup{
		lkePools: map[int]*linodego.LKEClusterPool{
			1: {ID: 1, Count: 1, Type: "g6-standard-1"},
			2: {ID: 2, Count: 1, Type: "g6-standard-1"},
			3: {ID: 3, Count: 1, Type: "g6-standard-1"},
		},
		poolOpts:     poolOpts,
		client:       &client,
		lkeClusterID: 111,
		minSize:      1,
		maxSize:      7,
		id:           "g6-standard-1",
	}
	client.On(
		"CreateLKEClusterPool", ctx, ng.lkeClusterID, poolOpts,
	).Return(
		&linodego.LKEClusterPool{ID: 4, Count: 1, Type: "g6-standard-1"}, nil,
	).Once().On(
		"CreateLKEClusterPool", ctx, ng.lkeClusterID, poolOpts,
	).Return(
		&linodego.LKEClusterPool{ID: 5, Count: 1, Type: "g6-standard-1"}, nil,
	).Once().On(
		"CreateLKEClusterPool", ctx, ng.lkeClusterID, poolOpts,
	).Return(
		&linodego.LKEClusterPool{ID: 6, Count: 1, Type: "g6-standard-1"}, nil,
	).Once().On(
		"CreateLKEClusterPool", ctx, ng.lkeClusterID, poolOpts,
	).Return(
		&linodego.LKEClusterPool{ID: 6, Count: 1, Type: "g6-standard-1"}, fmt.Errorf("error on API call"),
	).Once()

	// test error on bad delta value
	err := ng.IncreaseSize(0)
	assert.Error(t, err)

	// test error on bad delta value
	err = ng.IncreaseSize(-1)
	assert.Error(t, err)

	// test error on a too large increase of nodes
	err = ng.IncreaseSize(5)
	assert.Error(t, err)

	// test ok to add a node
	err = ng.IncreaseSize(1)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(ng.lkePools))

	// test ok to add multiple nodes
	err = ng.IncreaseSize(2)
	assert.NoError(t, err)
	assert.Equal(t, 6, len(ng.lkePools))

	// test error on linode API call error
	err = ng.IncreaseSize(1)
	assert.Error(t, err, "no error on injected API call error")
}

func TestNodeGroup_DecreaseTargetSize(t *testing.T) {
	ng := &NodeGroup{}
	err := ng.DecreaseTargetSize(-1)
	assert.Error(t, err)
}

func TestNodeGroup_DeleteNodes(t *testing.T) {
	client := linodeClientMock{}
	ctx := context.Background()
	poolOpts := linodego.LKEClusterPoolCreateOptions{
		Count: 1,
		Type:  "g6-standard-1",
	}
	ng := NodeGroup{
		lkePools: map[int]*linodego.LKEClusterPool{
			1: {ID: 1, Count: 1, Type: "g6-standard-1", Linodes: []linodego.LKEClusterPoolLinode{{InstanceID: 123}}},
			2: {ID: 2, Count: 1, Type: "g6-standard-1", Linodes: []linodego.LKEClusterPoolLinode{{InstanceID: 223}}},
			3: {ID: 3, Count: 1, Type: "g6-standard-1", Linodes: []linodego.LKEClusterPoolLinode{{InstanceID: 323}}},
			4: {ID: 4, Count: 1, Type: "g6-standard-1", Linodes: []linodego.LKEClusterPoolLinode{{InstanceID: 423}}},
			5: {ID: 5, Count: 1, Type: "g6-standard-1", Linodes: []linodego.LKEClusterPoolLinode{{InstanceID: 523}}},
		},
		poolOpts:     poolOpts,
		client:       &client,
		lkeClusterID: 111,
		minSize:      1,
		maxSize:      6,
		id:           "g6-standard-1",
	}
	client.On(
		"DeleteLKEClusterPool", ctx, ng.lkeClusterID, 1,
	).Return(nil).On(
		"DeleteLKEClusterPool", ctx, ng.lkeClusterID, 2,
	).Return(nil).On(
		"DeleteLKEClusterPool", ctx, ng.lkeClusterID, 3,
	).Return(nil).On(
		"DeleteLKEClusterPool", ctx, ng.lkeClusterID, 4,
	).Return(fmt.Errorf("error on API call")).On(
		"DeleteLKEClusterPool", ctx, ng.lkeClusterID, 5,
	).Return(nil)

	nodes := []*apiv1.Node{
		{Spec: apiv1.NodeSpec{ProviderID: "linode://123"}},
		{Spec: apiv1.NodeSpec{ProviderID: "linode://223"}},
		{Spec: apiv1.NodeSpec{ProviderID: "linode://523"}},
	}

	// test of on deleting nodes
	err := ng.DeleteNodes(nodes)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(ng.lkePools))
	assert.NotNil(t, ng.lkePools[3])
	assert.NotNil(t, ng.lkePools[4])

	// test error on deleting a node with a malformed providerID
	nodes = []*apiv1.Node{
		{Spec: apiv1.NodeSpec{ProviderID: "linode://aaa"}},
	}
	err = ng.DeleteNodes(nodes)
	assert.Error(t, err)

	// test error on deleting a node we are not managing
	nodes = []*apiv1.Node{
		{Spec: apiv1.NodeSpec{ProviderID: "linode://555"}},
	}
	err = ng.DeleteNodes(nodes)
	assert.Error(t, err)

	// test error on deleting a node when the linode API call fails
	nodes = []*apiv1.Node{
		{Spec: apiv1.NodeSpec{ProviderID: "linode://423"}},
	}
	err = ng.DeleteNodes(nodes)
	assert.Error(t, err)
}

func TestNodeGroup_deleteLKEPool(t *testing.T) {
	client := linodeClientMock{}
	ctx := context.Background()
	poolOpts := linodego.LKEClusterPoolCreateOptions{
		Count: 1,
		Type:  "g6-standard-1",
	}
	ng := NodeGroup{
		lkePools: map[int]*linodego.LKEClusterPool{
			1: {ID: 1, Count: 1, Type: "g6-standard-1", Linodes: []linodego.LKEClusterPoolLinode{{InstanceID: 123}}},
			2: {ID: 2, Count: 1, Type: "g6-standard-1", Linodes: []linodego.LKEClusterPoolLinode{{InstanceID: 223}}},
			3: {ID: 3, Count: 1, Type: "g6-standard-1", Linodes: []linodego.LKEClusterPoolLinode{{InstanceID: 323}}},
			4: {ID: 4, Count: 1, Type: "g6-standard-1", Linodes: []linodego.LKEClusterPoolLinode{{InstanceID: 423}}},
			5: {ID: 5, Count: 1, Type: "g6-standard-1", Linodes: []linodego.LKEClusterPoolLinode{{InstanceID: 523}}},
		},
		poolOpts:     poolOpts,
		client:       &client,
		lkeClusterID: 111,
		minSize:      1,
		maxSize:      6,
		id:           "g6-standard-1",
	}
	client.On(
		"DeleteLKEClusterPool", ctx, ng.lkeClusterID, 3,
	).Return(nil)

	// test ok on deleting a pool from a node group
	err := ng.deleteLKEPool(3)
	assert.NoError(t, err)
	assert.Nil(t, ng.lkePools[3])

	// test error on deleting a pool from a node group that does not contain it
	err = ng.deleteLKEPool(6)
	assert.Error(t, err)
}

func TestNosdeGroup_Nodes(t *testing.T) {

	client := linodeClientMock{}
	poolOpts := linodego.LKEClusterPoolCreateOptions{
		Count: 1,
		Type:  "g6-standard-1",
	}
	ng := NodeGroup{
		lkePools: map[int]*linodego.LKEClusterPool{
			4: {ID: 4,
				Count: 1,
				Type:  "g6-standard-1",
				Linodes: []linodego.LKEClusterPoolLinode{
					{InstanceID: 423},
				},
			},
			5: {ID: 5,
				Count: 2,
				Type:  "g6-standard-1",
				Linodes: []linodego.LKEClusterPoolLinode{
					{InstanceID: 523}, {InstanceID: 623},
				},
			},
		},
		poolOpts:     poolOpts,
		client:       &client,
		lkeClusterID: 111,
		minSize:      1,
		maxSize:      6,
		id:           "g6-standard-1",
	}

	// test nodes returned from Nodes() are only the ones we are expecting
	instancesList, err := ng.Nodes()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(instancesList))
	assert.Contains(t, instancesList, cloudprovider.Instance{Id: "linode://423"})
	assert.Contains(t, instancesList, cloudprovider.Instance{Id: "linode://523"})
	assert.Contains(t, instancesList, cloudprovider.Instance{Id: "linode://623"})
	assert.NotContains(t, instancesList, cloudprovider.Instance{Id: "423"})
}

func TestNodeGroup_Others(t *testing.T) {
	client := linodeClientMock{}
	poolOpts := linodego.LKEClusterPoolCreateOptions{
		Count: 1,
		Type:  "g6-standard-1",
	}
	ng := NodeGroup{
		lkePools: map[int]*linodego.LKEClusterPool{
			1: {ID: 1, Count: 1, Type: "g6-standard-1"},
			2: {ID: 2, Count: 1, Type: "g6-standard-1"},
			3: {ID: 3, Count: 1, Type: "g6-standard-1"},
		},
		poolOpts:     poolOpts,
		client:       &client,
		lkeClusterID: 111,
		minSize:      1,
		maxSize:      7,
		id:           "g6-standard-1",
	}
	assert.Equal(t, 1, ng.MinSize())
	assert.Equal(t, 7, ng.MaxSize())
	ts, err := ng.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 3, ts)
	assert.Equal(t, "g6-standard-1", ng.Id())
	assert.Equal(t, "node group ID: g6-standard-1 (min:1 max:7)", ng.Debug())
	assert.Equal(t, true, ng.Exist())
	assert.Equal(t, false, ng.Autoprovisioned())
	_, err = ng.TemplateNodeInfo()
	assert.Error(t, err)
	_, err = ng.Create()
	assert.Error(t, err)
	err = ng.Delete()
	assert.Error(t, err)
}
