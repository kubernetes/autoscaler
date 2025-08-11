/*
Copyright 2019 The Kubernetes Authors.

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

package datacrunch

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	datacrunchclient "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/datacrunch/datacrunch-go"
)

func TestServersCache(t *testing.T) {
	c := newServersCache(context.Background(), nil)

	// add initial cache entry, to test that it will be replaced
	serversOld := []*datacrunchclient.Instance{
		{
			ID:          "old-id",
			Hostname:    "test-old",
			Description: "group1",
		},
	}

	err := c.Add(serversCachedObject{
		name:    serversCacheKey,
		servers: serversOld,
	})
	require.NoError(t, err)

	servers := []*datacrunchclient.Instance{
		{
			ID:          "id1",
			Hostname:    "test1",
			Description: "group1",
		},
		{
			ID:          "id2",
			Hostname:    "test2",
			Description: "group1",
		},
	}

	cacheObject := serversCachedObject{
		name:    serversCacheKey,
		servers: servers,
	}

	err = c.Add(cacheObject)

	require.NoError(t, err)
	obj, ok, err := c.GetByKey(serversCacheKey)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, serversCacheKey, obj.(serversCachedObject).name)
	foundserverss := obj.(serversCachedObject).servers
	assert.Equal(t, 2, len(foundserverss))
	assert.Equal(t, "test1", foundserverss[0].Hostname)

	foundservers, err := c.getServer("test2")
	require.NoError(t, err)
	assert.Equal(t, "test2", foundservers.Hostname)

	server, err := c.getServer("test3")
	require.Nil(t, server)
	require.NoError(t, err)
}
