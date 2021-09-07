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
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/linode/linodego"
	"github.com/stretchr/testify/assert"
)

func TestManager_newManager(t *testing.T) {
	t.Run("fails on empty buffer", func(t *testing.T) {
		_, err := newManager(bytes.NewBuffer(nil))
		assert.Error(t, err)
	})

	t.Run("fails on read error", func(t *testing.T) {
		_, err := newManager(&readerErrMock{})
		assert.Error(t, err)
	})

	t.Run("fails without clusterID", func(t *testing.T) {
		cfg := strings.NewReader(`{"token": "bogus"}`)
		_, err := newManager(cfg)
		assert.Error(t, err)
	})

	t.Run("fails without token", func(t *testing.T) {
		cfg := strings.NewReader(`{"clusterID": 123}`)
		_, err := newManager(cfg)
		assert.Error(t, err)
	})

	t.Run("successfully creates manager", func(t *testing.T) {
		cfg := strings.NewReader(`
			{"clusterID": 123, "token": "bogus", "apiVersion": "v4beta", "url": "api2.linode.com"}
		`)
		_, err := newManager(cfg)
		assert.NoError(t, err)
	})

	t.Run("gets token from env if not in config", func(t *testing.T) {
		token := "bogus"
		restore := testEnvVar(linodeTokenEnvVar, token)
		defer restore()

		cfg := strings.NewReader(`{"clusterID": 123}`)
		m, err := newManager(cfg)
		assert.NoError(t, err)
		assert.Equal(t, m.config.Token, token)
	})

	t.Run("gets clusterID from env if not in config", func(t *testing.T) {
		clusterID := 123
		restore := testEnvVar(lkeClusterIDEnvVar, strconv.Itoa(clusterID))
		defer restore()

		cfg := strings.NewReader(`{"token": "bogus"}`)
		m, err := newManager(cfg)
		assert.NoError(t, err)
		assert.Equal(t, m.config.ClusterID, clusterID)
	})

	t.Run("sucessfully creates manager without file", func(t *testing.T) {
		tokenRestore := testEnvVar(linodeTokenEnvVar, "bogus")
		defer tokenRestore()
		clusterIDRestore := testEnvVar(lkeClusterIDEnvVar, "123")
		defer clusterIDRestore()

		m, err := newManager(nil)
		assert.NoError(t, err)
		assert.NotNil(t, m)
	})
}

func TestManager_refreshAfterInterval(t *testing.T) {
	cfg := strings.NewReader(`{
    "clusterID": 456456,
    "token": "123123123"
}`)
	m, err := newManager(cfg)
	assert.NoError(t, err)

	client := &linodeClientMock{}
	m.client = client
	ctx := context.Background()
	pools := []linodego.LKEClusterPool{
		makeMockNodePool(123, makeTestNodePoolNodes(1001, 1003), linodego.LKEClusterPoolAutoscaler{
			Min:     5,
			Max:     10,
			Enabled: true,
		}),
		makeMockNodePool(124, makeTestNodePoolNodes(1004, 1025), linodego.LKEClusterPoolAutoscaler{
			Min:     1,
			Max:     50,
			Enabled: true,
		}),
	}

	client.On(
		"ListLKEClusterPools", ctx, m.config.ClusterID, nil,
	).Return(pools, nil).Once()

	t.Run("not ran when lastRefresh was less than 5 minutes ago", func(t *testing.T) {
		now := time.Now()
		m.lastRefresh = now.Add(-1 * time.Minute)

		err := m.refreshAfterInterval()
		assert.NoError(t, err)
		client.AssertNotCalled(t, "ListLKEClusterPools")
	})

	t.Run("successfully refreshes nodepools", func(t *testing.T) {
		now := time.Now()
		m.lastRefresh = now.Add(-5 * time.Minute)

		err := m.refreshAfterInterval()
		assert.NoError(t, err)
		assert.Len(t, m.nodeGroups, 2)
		assert.Equal(t, m.nodeGroups[123], nodeGroupFromPool(client, m.config.ClusterID, &pools[0]))
		assert.Equal(t, m.nodeGroups[124], nodeGroupFromPool(client, m.config.ClusterID, &pools[1]))
		assert.True(t, m.lastRefresh.After(now))
	})

	client.On(
		"ListLKEClusterPools", ctx, 456456, nil,
	).Return(
		[]linodego.LKEClusterPool{},
		fmt.Errorf("error on API call"),
	).Once()

	t.Run("fails on generic ListLKEClusterPools error", func(t *testing.T) {
		now := time.Now()
		initialLastRefresh := now.Add(-5 * time.Minute)
		m.lastRefresh = initialLastRefresh

		err := m.refreshAfterInterval()
		assert.Error(t, err)
		assert.Equal(t, m.lastRefresh, initialLastRefresh)
	})
}
