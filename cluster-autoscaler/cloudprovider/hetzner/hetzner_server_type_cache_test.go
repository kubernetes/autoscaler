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

package hetzner

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/hetzner/hcloud-go/hcloud"
)

func TestServerTypeCache(t *testing.T) {
	c := newServerTypeCache(context.Background(), nil)

	serverTypes := []*hcloud.ServerType{
		{
			Name: "test1",
		},
		{
			Name: "test2",
		},
	}

	cacheObject := serverTypeCachedObject{
		name:        serverTypeCacheKey,
		serverTypes: serverTypes,
	}

	err := c.Add(cacheObject)

	require.NoError(t, err)
	obj, ok, err := c.GetByKey(serverTypeCacheKey)
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, serverTypeCacheKey, obj.(serverTypeCachedObject).name)
	foundServerTypes := obj.(serverTypeCachedObject).serverTypes
	assert.Equal(t, 2, len(foundServerTypes))
	assert.Equal(t, "test1", foundServerTypes[0].Name)

	foundServerType, err := c.getServerType("test2")
	require.NoError(t, err)
	assert.Equal(t, "test2", foundServerType.Name)

	_, err = c.getServerType("test3")
	require.Error(t, err)
}
