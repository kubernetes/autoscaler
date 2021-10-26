/*
Copyright 2021 The Kubernetes Authors.

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
	"net/url"
	"path"
	"sync"

	gce "google.golang.org/api/compute/v1"
	"k8s.io/client-go/util/workqueue"
	klog "k8s.io/klog/v2"
)

// MigInfoProvider allows obtaining information about MIGs
type MigInfoProvider interface {
	// GetMigTargetSize returns target size for given MIG ref
	GetMigTargetSize(migRef GceRef) (int64, error)
	// GetMigBasename returns basename for given MIG ref
	GetMigBasename(migRef GceRef) (string, error)
	// GetMigInstanceTemplateName returns instance template name for given MIG ref
	GetMigInstanceTemplateName(migRef GceRef) (string, error)
	// GetMigInstanceTemplate returns instance template for given MIG ref
	GetMigInstanceTemplate(migRef GceRef) (*gce.InstanceTemplate, error)
}

type cachingMigInfoProvider struct {
	mutex     sync.Mutex
	cache     *GceCache
	gceClient AutoscalingGceClient
	projectId string
}

// NewCachingMigInfoProvider creates an instance of caching MigInfoProvider
func NewCachingMigInfoProvider(cache *GceCache, gceClient AutoscalingGceClient, projectId string) MigInfoProvider {
	return &cachingMigInfoProvider{
		cache:     cache,
		gceClient: gceClient,
		projectId: projectId,
	}
}

func (c *cachingMigInfoProvider) GetMigTargetSize(migRef GceRef) (int64, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	targetSize, found := c.cache.GetMigTargetSize(migRef)
	if found {
		return targetSize, nil
	}

	err := c.fillMigInfoCache()
	targetSize, found = c.cache.GetMigTargetSize(migRef)
	if err == nil && found {
		return targetSize, nil
	}

	// fallback to querying for single mig
	targetSize, err = c.gceClient.FetchMigTargetSize(migRef)
	if err != nil {
		return 0, err
	}
	c.cache.SetMigTargetSize(migRef, targetSize)
	return targetSize, nil
}

func (c *cachingMigInfoProvider) GetMigBasename(migRef GceRef) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	basename, found := c.cache.GetMigBasename(migRef)
	if found {
		return basename, nil
	}

	err := c.fillMigInfoCache()
	basename, found = c.cache.GetMigBasename(migRef)
	if err == nil && found {
		return basename, nil
	}

	// fallback to querying for single mig
	basename, err = c.gceClient.FetchMigBasename(migRef)
	if err != nil {
		return "", err
	}
	c.cache.SetMigBasename(migRef, basename)
	return basename, nil
}

func (c *cachingMigInfoProvider) GetMigInstanceTemplateName(migRef GceRef) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	templateName, found := c.cache.GetMigInstanceTemplateName(migRef)
	if found {
		return templateName, nil
	}

	err := c.fillMigInfoCache()
	templateName, found = c.cache.GetMigInstanceTemplateName(migRef)
	if err == nil && found {
		return templateName, nil
	}

	// fallback to querying for single mig
	templateName, err = c.gceClient.FetchMigTemplateName(migRef)
	if err != nil {
		return "", err
	}
	c.cache.SetMigInstanceTemplateName(migRef, templateName)
	return templateName, nil
}

func (c *cachingMigInfoProvider) GetMigInstanceTemplate(migRef GceRef) (*gce.InstanceTemplate, error) {
	templateName, err := c.GetMigInstanceTemplateName(migRef)
	if err != nil {
		return nil, err
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	template, found := c.cache.GetMigInstanceTemplate(migRef)
	if found && template.Name == templateName {
		return template, nil
	}

	klog.V(2).Infof("Instance template of mig %v changed to %v", migRef.Name, templateName)
	template, err = c.gceClient.FetchMigTemplate(migRef, templateName)
	if err != nil {
		return nil, err
	}
	c.cache.SetMigInstanceTemplate(migRef, template)
	return template, nil
}

// filMigInfoCache needs to be called with mutex locked
func (c *cachingMigInfoProvider) fillMigInfoCache() error {
	var zones []string
	for zone := range c.listAllZonesWithMigs() {
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
			return fmt.Errorf("%v", errors)
		}
	}

	registeredMigRefs := c.getRegisteredMigRefs()
	for idx, zone := range zones {
		for _, zoneMig := range migs[idx] {
			zoneMigRef := GceRef{
				c.projectId,
				zone,
				zoneMig.Name,
			}

			if registeredMigRefs[zoneMigRef] {
				c.cache.SetMigTargetSize(zoneMigRef, zoneMig.TargetSize)
				c.cache.SetMigBasename(zoneMigRef, zoneMig.BaseInstanceName)

				templateUrl, err := url.Parse(zoneMig.InstanceTemplate)
				if err == nil {
					_, templateName := path.Split(templateUrl.EscapedPath())
					c.cache.SetMigInstanceTemplateName(zoneMigRef, templateName)
				}
			}
		}
	}

	return nil
}

func (c *cachingMigInfoProvider) getRegisteredMigRefs() map[GceRef]bool {
	migRefs := make(map[GceRef]bool)
	for _, mig := range c.cache.GetMigs() {
		migRefs[mig.GceRef()] = true
	}
	return migRefs
}

func (c *cachingMigInfoProvider) listAllZonesWithMigs() map[string]bool {
	zones := map[string]bool{}
	for _, mig := range c.cache.GetMigs() {
		zones[mig.GceRef().Zone] = true
	}
	return zones
}
