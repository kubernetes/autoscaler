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
	"errors"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/rand"
	datacrunchclient "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/datacrunch/datacrunch-go"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
)

const (
	serverTypeCacheKey    = "datacrunch-server-type-cache"
	serverTypeCachedTTL   = time.Minute * 10
	serverTypeCacheMinTTL = 5
	serverTypeCacheMaxTTL = 60
	availabilityCacheTTL  = time.Minute * 1 // Refresh availability cache every minute
)

// Add availability cache to serverTypeCache

type availabilityKey struct {
	InstanceType string
	Region       string
	IsSpot       bool
}

type availabilityCacheEntry struct {
	available bool
	timestamp time.Time
}

// serverTypeCache now also caches instance type availability per (instanceType, region)
type serverTypeCache struct {
	cache.Store
	mngJitterClock          clock.Clock
	datacrunchClient        *datacrunchclient.Client
	datacrunchClientContext context.Context

	availabilityCache map[availabilityKey]availabilityCacheEntry
	availabilityMu    sync.RWMutex
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
	serverTypes []*datacrunchclient.InstanceType
}

func newServerTypeCache(ctx context.Context, datacrunchClient *datacrunchclient.Client) *serverTypeCache {
	jc := &serverTypeClock{
		Clock: clock.RealClock{},
	}
	return newServerTypeCacheWithClock(
		ctx,
		datacrunchClient,
		jc,
		cache.NewExpirationStore(func(obj interface{}) (s string, e error) {
			return obj.(serverTypeCachedObject).name, nil
		}, &cache.TTLPolicy{
			TTL:   serverTypeCachedTTL,
			Clock: jc,
		}),
	)
}

func newServerTypeCacheWithClock(ctx context.Context, datacrunchClient *datacrunchclient.Client, jc clock.Clock, store cache.Store) *serverTypeCache {
	return &serverTypeCache{
		Store:                   store,
		mngJitterClock:          jc,
		datacrunchClient:        datacrunchClient,
		datacrunchClientContext: ctx,
		availabilityCache:       make(map[availabilityKey]availabilityCacheEntry),
	}
}

func (m *serverTypeCache) serverTypes() ([]*datacrunchclient.InstanceType, error) {
	klog.Warning("Fetching server types from DataCrunch API")

	instanceTypes, err := m.datacrunchClient.ListInstanceTypes()
	if err != nil {
		return nil, err
	}

	// Convert InstanceTypeList (which is []InstanceType) to []*InstanceType
	types := make([]*datacrunchclient.InstanceType, len(instanceTypes))
	for i := range instanceTypes {
		types[i] = &instanceTypes[i]
	}

	cacheObject := serverTypeCachedObject{
		name:        serverTypeCacheKey,
		serverTypes: types,
	}

	if err := m.Add(cacheObject); err != nil {
		return nil, err
	}

	return types, nil
}

func (m *serverTypeCache) getAllServerTypes() ([]*datacrunchclient.InstanceType, error) {
	// List expires old entries
	cacheList := m.List()
	klog.V(5).Infof("Current serverTypeCache len: %d\n", len(cacheList))

	if obj, found, err := m.GetByKey(serverTypeCacheKey); err == nil && found {
		foundServerTypes := obj.(serverTypeCachedObject)

		return foundServerTypes.serverTypes, nil
	}

	return m.serverTypes()
}

func (m *serverTypeCache) getServerType(instanceType string) (*datacrunchclient.InstanceType, error) {
	serverTypes, err := m.getAllServerTypes()
	if err != nil {
		return nil, err
	}

	for _, serverType := range serverTypes {
		if serverType.InstanceType == instanceType {
			return serverType, nil
		}
	}

	return nil, errors.New("server type not found")
}

// GetInstanceTypeAvailabilityCached checks the cache for (instanceType, region) availability, calls the API if not present, and caches the result.
func (m *serverTypeCache) GetInstanceTypeAvailabilityCached(instanceType, region string, isSpot bool) (bool, error) {
	key := availabilityKey{InstanceType: instanceType, Region: region, IsSpot: isSpot}
	m.availabilityMu.RLock()
	entry, ok := m.availabilityCache[key]
	m.availabilityMu.RUnlock()

	// Check if entry exists and is not expired
	if ok && m.mngJitterClock.Since(entry.timestamp) < availabilityCacheTTL {
		return entry.available, nil
	}

	// Not cached or expired, call API
	available, err := m.datacrunchClient.GetInstanceTypeAvailability(instanceType, isSpot, region)
	if err != nil {
		return false, err
	}

	// Update cache with new entry
	m.availabilityMu.Lock()
	m.availabilityCache[key] = availabilityCacheEntry{
		available: available,
		timestamp: m.mngJitterClock.Now(),
	}
	m.availabilityMu.Unlock()
	return available, nil
}
