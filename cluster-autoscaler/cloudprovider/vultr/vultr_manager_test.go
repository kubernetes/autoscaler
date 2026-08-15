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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/vultr/govultr"
)

func TestManager_newManager(t *testing.T) {

	t.Run("basic token and cluster id", func(t *testing.T) {
		config := `{"token": "123-456", "cluster_id": "abc"}`

		manager, err := newManager(strings.NewReader(config))
		require.NoError(t, err)

		assert.Equal(t, manager.clusterID, "abc", "invalid cluster id")
	})

	t.Run("missing token", func(t *testing.T) {
		config := `{"token": "", "cluster_id": "abc"}`

		_, err := newManager(strings.NewReader(config))
		assert.EqualError(t, err, errors.New("empty token was supplied").Error())
	})

	t.Run("missing cluster id", func(t *testing.T) {
		config := `{"token": "123-345", "cluster_id": ""}`

		_, err := newManager(strings.NewReader(config))
		assert.EqualError(t, err, errors.New("empty cluster ID was supplied").Error())
	})
}

func TestManager_Refresh(t *testing.T) {
	config := `{"token": "123-456", "cluster_id": "abc"}`

	manager, err := newManager(strings.NewReader(config))
	require.NoError(t, err)

	client := &vultrClientMock{}
	ctx := context.Background()

	client.On("ListNodePools", ctx, manager.clusterID, nil).Return(
		[]govultr.NodePool{
			{
				ID:         "1234",
				AutoScaler: true,
				MinNodes:   1,
				MaxNodes:   2,
			},
			{
				ID:         "4567",
				AutoScaler: true,
				MinNodes:   5,
				MaxNodes:   8,
			},
			{
				ID:         "9876",
				AutoScaler: false,
				MinNodes:   5,
				MaxNodes:   8,
			},
		},
		&govultr.Meta{},
		nil,
	).Once()

	manager.client = client

	err = manager.Refresh()
	assert.NoError(t, err)
	assert.Equal(t, len(manager.nodeGroups), 2, "number of nodepools do not match")

	assert.Equal(t, manager.nodeGroups[0].minSize, 1, "minimum node for first group does not match")
	assert.Equal(t, manager.nodeGroups[0].MaxSize(), 2, "minimum node for first group does not match")
	//
	assert.Equal(t, manager.nodeGroups[1].minSize, 5, "minimum node for first group does not match")
	assert.Equal(t, manager.nodeGroups[1].maxSize, 8, "minimum node for first group does not match")

}
