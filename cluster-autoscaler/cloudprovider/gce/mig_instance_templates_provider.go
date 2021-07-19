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
	"math/rand"
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
	mutex     sync.Mutex
	cache     *GceCache
	gceClient AutoscalingGceClient
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
	instanceTemplate, lastRefresh := p.getMigInstanceTemplateFromCache(migRef)
	if instanceTemplate != nil {
		return instanceTemplate, nil
	}

	instanceTemplate, err := p.gceClient.FetchMigTemplate(migRef)
	if err != nil {
		return nil, err
	}
	p.setMigInstanceTemplateToCache(migRef, instanceTemplate, lastRefresh)

	return instanceTemplate, nil
}

func (p *CachingMigInstanceTemplatesProvider) getMigInstanceTemplateFromCache(migRef GceRef) (*gce.InstanceTemplate, time.Time) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	instanceTemplateEntry, found := p.cache.GetMigInstanceTemplate(migRef)
	if !found {
		// first time seen: spread refreshes by randomizing the first deadline (between now and the max ttl)
		splay := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(int(migInstanceCacheRefreshInterval.Seconds() + 1))
		return nil, time.Now().Add(-time.Second * time.Duration(splay))
	}

	if instanceTemplateEntry.lastRefresh.Add(migInstanceCacheRefreshInterval).Before(time.Now()) {
		p.cache.InvalidateMigInstanceTemplate(migRef)
		return nil, time.Now()
	}

	return instanceTemplateEntry.template, instanceTemplateEntry.lastRefresh
}

func (p *CachingMigInstanceTemplatesProvider) setMigInstanceTemplateToCache(migRef GceRef, instanceTemplate *gce.InstanceTemplate, lastRefresh time.Time) {
	p.cache.InvalidateLongObsoleteMigInstanceTemplates(migInstanceCacheRefreshInterval * 2)
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.cache.SetMigInstanceTemplate(migRef, &instanceTemplateCacheEntry{
		template:    instanceTemplate,
		lastRefresh: lastRefresh,
	})
}
