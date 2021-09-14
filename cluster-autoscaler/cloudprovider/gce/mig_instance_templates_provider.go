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
	"sync"

	gce "google.golang.org/api/compute/v1"
	"k8s.io/client-go/util/workqueue"
	klog "k8s.io/klog/v2"
)

// MigInstanceTemplatesProvider allows obtaining instance templates for MIGs
type MigInstanceTemplatesProvider interface {
	// GetMigInstanceTemplate returns instance template for MIG with given ref
	GetMigInstanceTemplate(migRef GceRef) (*gce.InstanceTemplate, error)
	// Refresh InstanceTemplate cache
	Refresh(migRefs []GceRef, concurrency int)
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
	instanceTemplate, found := p.getMigInstanceTemplateFromCache(migRef)
	if found {
		return instanceTemplate, nil
	}

	instanceTemplate, err := p.gceClient.FetchMigTemplate(migRef)
	if err != nil {
		return nil, err
	}
	p.setMigInstanceTemplateToCache(migRef, instanceTemplate)

	return instanceTemplate, nil
}

// Refresh InstanceTemplate cache
func (p *CachingMigInstanceTemplatesProvider) Refresh(gceRefs []GceRef, concurrency int) {
	gceRefsToInstanceTemplates := map[GceRef]*gce.InstanceTemplate{}
	gceRefsToInstanceTemplatesLock := sync.Mutex{}
	workqueue.ParallelizeUntil(context.Background(), concurrency, len(gceRefs), func(piece int) {
		instanceTemplate, err := p.gceClient.FetchMigTemplate(gceRefs[piece])
		if err != nil {
			klog.Errorf("Error while regenerating instance template cache for %s: %v",
				gceRefs[piece], err)
			return
		}
		gceRefsToInstanceTemplatesLock.Lock()
		gceRefsToInstanceTemplates[gceRefs[piece]] = instanceTemplate
		gceRefsToInstanceTemplatesLock.Unlock()
	})

	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.cache.SetAllMigInstanceTemplates(gceRefsToInstanceTemplates)
}

func (p *CachingMigInstanceTemplatesProvider) getMigInstanceTemplateFromCache(migRef GceRef) (*gce.InstanceTemplate, bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.cache.GetMigInstanceTemplate(migRef)
}

func (p *CachingMigInstanceTemplatesProvider) setMigInstanceTemplateToCache(migRef GceRef, instanceTemplate *gce.InstanceTemplate) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.cache.SetMigInstanceTemplate(migRef, instanceTemplate)
}
