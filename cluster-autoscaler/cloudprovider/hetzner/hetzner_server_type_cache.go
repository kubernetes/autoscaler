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
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/hetzner/hcloud-go/hcloud"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
)

const (
	serverTypeCacheKey    = "hetzner-server-type-cache"
	serverTypeCachedTTL   = time.Minute * 10
	serverTypeCacheMinTTL = 5
	serverTypeCacheMaxTTL = 60
)

type serverTypeCache struct {
	cache.Store
	mngJitterClock      clock.Clock
	hcloudClient        *hcloud.Client
	hcloudClientContext context.Context
}

type serverTypeClock struct {
	clock.Clock

	jitter bool
	sync.RWMutex
}

func (c *serverTypeClock) Since(ts time.Time) time.Duration {
	since := time.Since(ts)
	c.RLock()
	defer c.RUnlock()
	if c.jitter {
		return since + (time.Second * time.Duration(rand.IntnRange(serverTypeCacheMinTTL, serverTypeCacheMaxTTL)))
	}
	return since
}

type serverTypeCachedObject struct {
	name        string
	serverTypes []*hcloud.ServerType
}

func newServerTypeCache(ctx context.Context, hcloudClient *hcloud.Client) *serverTypeCache {
	jc := &serverTypeClock{}
	return newServerTypeCacheWithClock(
		ctx,
		hcloudClient,
		jc,
		cache.NewExpirationStore(func(obj interface{}) (s string, e error) {
			return obj.(serverTypeCachedObject).name, nil
		}, &cache.TTLPolicy{
			TTL:   serverTypeCachedTTL,
			Clock: jc,
		}),
	)
}

func newServerTypeCacheWithClock(ctx context.Context, hcloudClient *hcloud.Client, jc clock.Clock, store cache.Store) *serverTypeCache {
	return &serverTypeCache{
		store,
		jc,
		hcloudClient,
		ctx,
	}
}

func (m *serverTypeCache) serverTypes() ([]*hcloud.ServerType, error) {
	klog.Warning("Fetching server types from Hetzner API")

	serverTypes, err := m.hcloudClient.ServerType.All(m.hcloudClientContext)
	if err != nil {
		return nil, err
	}

	cacheObject := serverTypeCachedObject{
		name:        serverTypeCacheKey,
		serverTypes: serverTypes,
	}

	if err := m.Add(cacheObject); err != nil {
		return nil, err
	}

	return serverTypes, nil
}

func (m *serverTypeCache) getAllServerTypes() ([]*hcloud.ServerType, error) {
	// List expires old entries
	cacheList := m.List()
	klog.V(5).Infof("Current serverTypeCache len: %d\n", len(cacheList))

	if obj, found, err := m.GetByKey(serverTypeCacheKey); err == nil && found {
		foundServerTypes := obj.(serverTypeCachedObject)

		return foundServerTypes.serverTypes, nil
	}

	return m.serverTypes()
}

func (m *serverTypeCache) getServerType(name string) (*hcloud.ServerType, error) {
	serverTypes, err := m.getAllServerTypes()
	if err != nil {
		return nil, err
	}

	for _, serverType := range serverTypes {
		if serverType.Name == name {
			return serverType, nil
		}
	}

	return nil, errors.New("server type not found")
}
