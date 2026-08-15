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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/linode/linodego"
)

func TestManager_newManager(t *testing.T) {
	cfg := strings.NewReader(`
[globalxxx]
linode-token=123123123
lke-cluster-id=456456
`)
	_, err := newManager(cfg)
	assert.Error(t, err)

	cfg = strings.NewReader(`
[global]
linode-token=123123123
lke-cluster-id=456456
`)
	_, err = newManager(cfg)
	assert.NoError(t, err)
}

func TestManager_refresh(t *testing.T) {

	cfg := strings.NewReader(`
[global]
linode-token=123123123
lke-cluster-id=456456
defaut-min-size-per-linode-type=2
defaut-max-size-per-linode-type=10
do-not-import-pool-id=888
do-not-import-pool-id=999

[nodegroup "g6-standard-1"]
min-size=1
max-size=2

[nodegroup "g6-standard-2"]
min-size=4
max-size=5
`)
	m, err := newManager(cfg)
	assert.NoError(t, err)

	client := linodeClientMock{}
	m.client = &client
	ctx := context.Background()

	// test multiple pools with same type
	client.On(
		"ListLKEClusterPools", ctx, 456456, nil,
	).Return(
		[]linodego.LKEClusterPool{
			{ID: 1, Count: 1, Type: "g6-standard-1", Linodes: []linodego.LKEClusterPoolLinode{{ID: "aaa", InstanceID: 123}}},
			{ID: 2, Count: 1, Type: "g6-standard-1", Linodes: []linodego.LKEClusterPoolLinode{{ID: "bbb", InstanceID: 345}}},
			{ID: 3, Count: 1, Type: "g6-standard-1", Linodes: []linodego.LKEClusterPoolLinode{{ID: "ccc", InstanceID: 678}}},
		},
		nil,
	).Once()
	err = m.refresh()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(m.nodeGroups))
	assert.Equal(t, 3, len(m.nodeGroups["g6-standard-1"].lkePools))

	// test skip pools with count > 1
	client.On(
		"ListLKEClusterPools", ctx, 456456, nil,
	).Return(
		[]linodego.LKEClusterPool{
			{ID: 1, Count: 1, Type: "g6-standard-1"},
			{ID: 2, Count: 1, Type: "g6-standard-1"},
			{ID: 3, Count: 2, Type: "g6-standard-1"},
		},
		nil,
	).Once()
	err = m.refresh()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(m.nodeGroups))
	assert.Equal(t, 2, len(m.nodeGroups["g6-standard-1"].lkePools))

	// test multiple pools with different types
	client.On(
		"ListLKEClusterPools", ctx, 456456, nil,
	).Return(
		[]linodego.LKEClusterPool{
			{ID: 1, Count: 1, Type: "g6-standard-1"},
			{ID: 2, Count: 1, Type: "g6-standard-1"},
			{ID: 3, Count: 1, Type: "g6-standard-2"},
		},
		nil,
	).Once()
	err = m.refresh()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(m.nodeGroups))
	assert.Equal(t, 2, len(m.nodeGroups["g6-standard-1"].lkePools))
	assert.Equal(t, 1, len(m.nodeGroups["g6-standard-2"].lkePools))

	// test avoid import of specific pools
	client.On(
		"ListLKEClusterPools", ctx, 456456, nil,
	).Return(
		[]linodego.LKEClusterPool{
			{ID: 1, Count: 1, Type: "g6-standard-1"},
			{ID: 888, Count: 1, Type: "g6-standard-1"},
			{ID: 999, Count: 1, Type: "g6-standard-1"},
		},
		nil,
	).Once()
	err = m.refresh()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(m.nodeGroups))
	assert.Equal(t, 1, len(m.nodeGroups["g6-standard-1"].lkePools))

	// test api error
	client.On(
		"ListLKEClusterPools", ctx, 456456, nil,
	).Return(
		[]linodego.LKEClusterPool{},
		fmt.Errorf("error on API call"),
	).Once()
	err = m.refresh()
	assert.Error(t, err)

}
