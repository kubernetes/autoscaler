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

package gce

import (
	"context"
	"fmt"
	"sync"

	gce "google.golang.org/api/compute/v1"
	"k8s.io/client-go/util/workqueue"
	klog "k8s.io/klog/v2"
)

// MigTargetSizesProvider allows obtaining target sizes of MIGs
type MigTargetSizesProvider interface {
	// GetMigTargetSize returns targetSize for MIG with given ref
	GetMigTargetSize(migRef GceRef) (int64, error)
}

type cachingMigTargetSizesProvider struct {
	mutex     sync.Mutex
	cache     *GceCache
	gceClient AutoscalingGceClient
	projectId string
}

// NewCachingMigTargetSizesProvider creates an instance of caching MigTargetSizesProvider
func NewCachingMigTargetSizesProvider(cache *GceCache, gceClient AutoscalingGceClient, projectId string) MigTargetSizesProvider {
	return &cachingMigTargetSizesProvider{
		cache:     cache,
		gceClient: gceClient,
		projectId: projectId,
	}
}

func (c *cachingMigTargetSizesProvider) GetMigTargetSize(migRef GceRef) (int64, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	targetSize, found := c.cache.GetMigTargetSize(migRef)

	if found {
		return targetSize, nil
	}

	newTargetSizes, err := c.fillInMigTargetSizeAndBaseNameCaches()

	size, found := newTargetSizes[migRef]
	if err != nil || !found {
		// fallback to querying for single mig
		targetSize, err = c.gceClient.FetchMigTargetSize(migRef)
		if err != nil {
			return 0, err
		}
		c.cache.SetMigTargetSize(migRef, targetSize)
		return targetSize, nil
	}

	// we are good
	return size, nil
}

func (c *cachingMigTargetSizesProvider) fillInMigTargetSizeAndBaseNameCaches() (map[GceRef]int64, error) {
	var zones []string
	for zone := range c.listAllZonesForMigs() {
		zones = append(zones, zone)
	}

	migs := make([][]*gce.InstanceGroupManager, len(zones))
	errors := make([]error, len(zones))
	workqueue.ParallelizeUntil(context.Background(), len(zones), len(zones), func(piece int) {
		migs[piece], errors[piece] = c.gceClient.FetchAllMigs(zones[piece])
	})

	for idx, err := range errors {
		if err != nil {
			klog.Errorf("Error listing migs from zone %v; err=%v", zones[idx], err)
			return nil, fmt.Errorf("%v", errors)
		}
	}

	newMigTargetSizeCache := map[GceRef]int64{}
	newMigBasenameCache := map[GceRef]string{}
	for idx, zone := range zones {
		registeredMigRefs := c.getMigRefs()
		for _, zoneMig := range migs[idx] {
			zoneMigRef := GceRef{
				c.projectId,
				zone,
				zoneMig.Name,
			}

			if registeredMigRefs[zoneMigRef] {
				newMigTargetSizeCache[zoneMigRef] = zoneMig.TargetSize
				newMigBasenameCache[zoneMigRef] = zoneMig.BaseInstanceName
			}
		}
	}

	for migRef, targetSize := range newMigTargetSizeCache {
		c.cache.SetMigTargetSize(migRef, targetSize)
	}

	for migRef, baseName := range newMigBasenameCache {
		c.cache.SetMigBasename(migRef, baseName)
	}

	return newMigTargetSizeCache, nil
}

func (c *cachingMigTargetSizesProvider) getMigRefs() map[GceRef]bool {
	migRefs := make(map[GceRef]bool)
	for _, mig := range c.cache.GetMigs() {
		migRefs[mig.GceRef()] = true
	}
	return migRefs
}

func (c *cachingMigTargetSizesProvider) listAllZonesForMigs() map[string]bool {
	zones := map[string]bool{}
	for _, mig := range c.cache.GetMigs() {
		zones[mig.GceRef().Zone] = true
	}
	return zones
}
