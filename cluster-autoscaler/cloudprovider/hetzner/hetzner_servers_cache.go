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
	"errors"
	"strconv"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/hetzner/hcloud-go/hcloud"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
)

const (
	serversCacheKey    = "hetzner-servers-cache"
	serversCachedTTL   = time.Minute * 1
	serversCacheMinTTL = 5
	serversCacheMaxTTL = 60
)

type serversCache struct {
	cache.Store
	mngJitterClock      clock.Clock
	hcloudClient        *hcloud.Client
	hcloudClientContext context.Context
}

type serversClock struct {
	clock.Clock

	jitter bool
	sync.RWMutex
}

func (c *serversClock) Since(ts time.Time) time.Duration {
	since := time.Since(ts)
	c.RLock()
	defer c.RUnlock()
	if c.jitter {
		return since + (time.Second * time.Duration(rand.IntnRange(serversCacheMinTTL, serversCacheMaxTTL)))
	}
	return since
}

type serversCachedObject struct {
	name    string
	servers []*hcloud.Server
}

func newServersCache(ctx context.Context, hcloudClient *hcloud.Client) *serversCache {
	jc := &serversClock{}
	return newServersCacheWithClock(
		ctx,
		hcloudClient,
		jc,
		cache.NewExpirationStore(func(obj interface{}) (s string, e error) {
			return obj.(serversCachedObject).name, nil
		}, &cache.TTLPolicy{
			TTL:   serversCachedTTL,
			Clock: jc,
		}),
	)
}

func newServersCacheWithClock(ctx context.Context, hcloudClient *hcloud.Client, jc clock.Clock, store cache.Store) *serversCache {
	return &serversCache{
		store,
		jc,
		hcloudClient,
		ctx,
	}
}

func (m *serversCache) servers() ([]*hcloud.Server, error) {
	klog.Warning("Fetching servers from Hetzner API")

	servers, err := m.hcloudClient.Server.All(m.hcloudClientContext)
	if err != nil {
		return nil, err
	}

	cacheObject := serversCachedObject{
		name:    serversCacheKey,
		servers: servers,
	}

	if err := m.Add(cacheObject); err != nil {
		return nil, err
	}

	return servers, nil
}

func (m *serversCache) getAllServers() ([]*hcloud.Server, error) {
	// List expires old entries
	cacheList := m.List()
	klog.V(5).Infof("Current serversCache len: %d\n", len(cacheList))

	if obj, found, err := m.GetByKey(serversCacheKey); err == nil && found {
		foundServers := obj.(serversCachedObject)

		return foundServers.servers, nil
	}

	return m.servers()
}

func (m *serversCache) getServer(nodeIdOrName string) (*hcloud.Server, error) {
	servers, err := m.getAllServers()
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		if server.Name == nodeIdOrName || strconv.FormatInt(server.ID, 10) == nodeIdOrName {
			return server, nil
		}
	}

	return nil, errors.New("server not found")
}

func (m *serversCache) getServersByNodeGroupName(nodeGroup string) ([]*hcloud.Server, error) {
	servers, err := m.getAllServers()
	if err != nil {
		return nil, err
	}

	foundServers := make([]*hcloud.Server, 0)

	for _, server := range servers {
		if server.Labels[nodeGroupLabel] == nodeGroup {
			foundServers = append(foundServers, server)
		}
	}

	return foundServers, nil
}
