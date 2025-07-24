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
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/rand"
	datacrunchclient "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/datacrunch/datacrunch-go"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
)

const (
	serversCacheKey    = "datacrunch-servers-cache"
	serversCachedTTL   = time.Minute * 1
	serversCacheMinTTL = 5
	serversCacheMaxTTL = 60
)

type serversCache struct {
	cache.Store
	mngJitterClock          clock.Clock
	datacrunchClient        *datacrunchclient.Client
	datacrunchClientContext context.Context
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
	servers []*datacrunchclient.Instance
}

func newServersCache(ctx context.Context, datacrunchClient *datacrunchclient.Client) *serversCache {
	jc := &serversClock{}
	return newServersCacheWithClock(
		ctx,
		datacrunchClient,
		jc,
		cache.NewExpirationStore(func(obj interface{}) (s string, e error) {
			return obj.(serversCachedObject).name, nil
		}, &cache.TTLPolicy{
			TTL:   serversCachedTTL,
			Clock: jc,
		}),
	)
}

func newServersCacheWithClock(ctx context.Context, datacrunchClient *datacrunchclient.Client, jc clock.Clock, store cache.Store) *serversCache {
	return &serversCache{
		store,
		jc,
		datacrunchClient,
		ctx,
	}
}

func (m *serversCache) servers() ([]*datacrunchclient.Instance, error) {
	klog.Warning("Fetching servers from DataCrunch API")

	instances, err := m.datacrunchClient.ListInstances("")
	if err != nil {
		return nil, err
	}

	// Convert InstanceList (which is []Instance) to []*Instance
	servers := make([]*datacrunchclient.Instance, len(instances))
	for i := range instances {
		servers[i] = &instances[i]
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

func (m *serversCache) getAllServers() ([]*datacrunchclient.Instance, error) {
	// List expires old entries
	cacheList := m.List()
	klog.V(5).Infof("Current serversCache len: %d\n", len(cacheList))

	if obj, found, err := m.GetByKey(serversCacheKey); err == nil && found {
		foundServers := obj.(serversCachedObject)
		return foundServers.servers, nil
	}

	return m.servers()
}

func (m *serversCache) getServer(nodeIdOrName string) (*datacrunchclient.Instance, error) {
	servers, err := m.getAllServers()
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		if server.Hostname == nodeIdOrName || server.ID == nodeIdOrName {
			return server, nil
		}
	}

	// return nil if server not found
	return nil, nil
}

func (m *serversCache) getServersByNodeGroupName(nodeGroup string) ([]*datacrunchclient.Instance, error) {
	servers, err := m.getAllServers()
	if err != nil {
		return nil, err
	}

	// DataCrunch does not have labels, so this is a stub. You may need to adapt this logic to your grouping method.
	foundServers := make([]*datacrunchclient.Instance, 0)
	for _, server := range servers {
		// Example: match by Description or another field if you use it for grouping
		if server.Description == nodeGroup {
			foundServers = append(foundServers, server)
		}
	}

	return foundServers, nil
}
