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
	"sync"
	"time"

	gce "google.golang.org/api/compute/v1"
)

const (
	migInstanceCacheRefreshInterval = 30 * time.Minute
)

// MigInstanceTemplatesProvider allows obtaining instance templates for MIGs
type MigInstanceTemplatesProvider interface {
	// GetMigInstanceTemplate returns instance template for MIG with given ref
	GetMigInstanceTemplate(migRef GceRef) (*gce.InstanceTemplate, error)
}

// CachingMigInstanceTemplatesProvider is caching implementation of MigInstanceTemplatesProvider
type CachingMigInstanceTemplatesProvider struct {
	mutex       sync.Mutex
	cache       *GceCache
	lastRefresh time.Time
	gceClient   AutoscalingGceClient
}

// NewCachingMigInstanceTemplatesProvider creates an instance of caching MigInstanceTemplatesProvider
func NewCachingMigInstanceTemplatesProvider(cache *GceCache, gceClient AutoscalingGceClient) *CachingMigInstanceTemplatesProvider {
	return &CachingMigInstanceTemplatesProvider{
		cache:     cache,
		gceClient: gceClient,
	}
}

// GetMigInstanceTemplate returns instance template for MIG with given ref
func (p *CachingMigInstanceTemplatesProvider) GetMigInstanceTemplate(migRef GceRef) (*gce.InstanceTemplate, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.lastRefresh.Add(migInstanceCacheRefreshInterval).After(time.Now()) {
		p.cache.InvalidateAllMigInstanceTemplates()
		p.lastRefresh = time.Now()
	}

	instanceTemplate, found := p.cache.GetMigInstanceTemplate(migRef)

	if found {
		return instanceTemplate, nil
	}

	instanceTemplate, err := p.gceClient.FetchMigTemplate(migRef)
	if err != nil {
		return nil, err
	}
	p.cache.SetMigInstanceTemplate(migRef, instanceTemplate)
	return instanceTemplate, nil
}
