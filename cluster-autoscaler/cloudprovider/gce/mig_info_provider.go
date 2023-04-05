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
	"strings"
	"sync"
	"time"

	gce "google.golang.org/api/compute/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/client-go/util/workqueue"
	klog "k8s.io/klog/v2"
)

// MigInfoProvider allows obtaining information about MIGs
type MigInfoProvider interface {
	// GetMigInstances returns instances for a given MIG ref
	GetMigInstances(migRef GceRef) ([]cloudprovider.Instance, error)
	// GetMigForInstance returns MIG ref for a given instance
	GetMigForInstance(instanceRef GceRef) (Mig, error)
	// RegenerateMigInstancesCache regenerates MIGs to instances mapping cache
	RegenerateMigInstancesCache() error
	// GetMigTargetSize returns target size for given MIG ref
	GetMigTargetSize(migRef GceRef) (int64, error)
	// GetMigBasename returns basename for given MIG ref
	GetMigBasename(migRef GceRef) (string, error)
	// GetMigInstanceTemplateName returns instance template name for given MIG ref
	GetMigInstanceTemplateName(migRef GceRef) (string, error)
	// GetMigInstanceTemplate returns instance template for given MIG ref
	GetMigInstanceTemplate(migRef GceRef) (*gce.InstanceTemplate, error)
	// GetMigMachineType returns machine type used by a MIG.
	// For custom machines cpu and memory information is based on parsing
	// machine name. For standard types it's retrieved from GCE API.
	GetMigMachineType(migRef GceRef) (MachineType, error)
}

type timeProvider interface {
	Now() time.Time
}

type cachingMigInfoProvider struct {
	migInfoMutex                   sync.Mutex
	cache                          *GceCache
	migLister                      MigLister
	gceClient                      AutoscalingGceClient
	projectId                      string
	concurrentGceRefreshes         int
	migInstanceMutex               sync.Mutex
	migInstancesMinRefreshWaitTime time.Duration
	migInstancesLastRefreshedInfo  map[string]time.Time
	timeProvider                   timeProvider
}

type realTime struct{}

func (r *realTime) Now() time.Time {
	return time.Now()
}

// NewCachingMigInfoProvider creates an instance of caching MigInfoProvider
func NewCachingMigInfoProvider(cache *GceCache, migLister MigLister, gceClient AutoscalingGceClient, projectId string, concurrentGceRefreshes int, migInstancesMinRefreshWaitTime time.Duration) MigInfoProvider {
	return &cachingMigInfoProvider{
		cache:                          cache,
		migLister:                      migLister,
		gceClient:                      gceClient,
		projectId:                      projectId,
		concurrentGceRefreshes:         concurrentGceRefreshes,
		migInstancesMinRefreshWaitTime: migInstancesMinRefreshWaitTime,
		migInstancesLastRefreshedInfo:  make(map[string]time.Time),
		timeProvider:                   &realTime{},
	}
}

// GetMigInstances returns instances for a given MIG ref
func (c *cachingMigInfoProvider) GetMigInstances(migRef GceRef) ([]cloudprovider.Instance, error) {
	instances, found := c.cache.GetMigInstances(migRef)
	if found {
		return instances, nil
	}

	err := c.fillMigInstances(migRef)
	if err != nil {
		return nil, err
	}
	instances, _ = c.cache.GetMigInstances(migRef)
	return instances, nil
}

// GetMigForInstance returns MIG ref for a given instance
func (c *cachingMigInfoProvider) GetMigForInstance(instanceRef GceRef) (Mig, error) {
	c.migInstanceMutex.Lock()
	defer c.migInstanceMutex.Unlock()

	mig, found, err := c.getCachedMigForInstance(instanceRef)
	if found {
		return mig, err
	}

	mig = c.findMigWithMatchingBasename(instanceRef)
	if mig == nil {
		return nil, nil
	}

	err = c.fillMigInstances(mig.GceRef())
	if err != nil {
		return nil, err
	}

	mig, found, err = c.getCachedMigForInstance(instanceRef)
	if !found {
		c.cache.MarkInstanceMigUnknown(instanceRef)
	}
	return mig, err
}

func (c *cachingMigInfoProvider) getCachedMigForInstance(instanceRef GceRef) (Mig, bool, error) {
	if migRef, found := c.cache.GetMigForInstance(instanceRef); found {
		mig, found := c.cache.GetMig(migRef)
		if !found {
			return nil, true, fmt.Errorf("instance %v belongs to unregistered mig %v", instanceRef, migRef)
		}
		return mig, true, nil
	} else if c.cache.IsMigUnknownForInstance(instanceRef) {
		return nil, true, nil
	}
	return nil, false, nil
}

// RegenerateMigInstancesCache regenerates MIGs to instances mapping cache
func (c *cachingMigInfoProvider) RegenerateMigInstancesCache() error {
	c.cache.InvalidateAllMigInstances()
	c.cache.InvalidateAllInstancesToMig()
	migs := c.migLister.GetMigs()
	errors := make([]error, len(migs))
	workqueue.ParallelizeUntil(context.Background(), c.concurrentGceRefreshes, len(migs), func(piece int) {
		errors[piece] = c.fillMigInstances(migs[piece].GceRef())
	}, workqueue.WithChunkSize(c.concurrentGceRefreshes))

	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *cachingMigInfoProvider) findMigWithMatchingBasename(instanceRef GceRef) Mig {
	for _, mig := range c.migLister.GetMigs() {
		migRef := mig.GceRef()
		basename, err := c.GetMigBasename(migRef)
		if err == nil && migRef.Project == instanceRef.Project && migRef.Zone == instanceRef.Zone && strings.HasPrefix(instanceRef.Name, basename) {
			return mig
		}
	}
	return nil
}

func (c *cachingMigInfoProvider) fillMigInstances(migRef GceRef) error {
	if val, ok := c.migInstancesLastRefreshedInfo[migRef.String()]; ok {
		// do not regenerate MIG instances cache if last refresh happened recently.
		if c.timeProvider.Now().Sub(val) < c.migInstancesMinRefreshWaitTime {
			klog.V(4).Infof("Not regenerating MIG instances cache for %s, as it was refreshed in last MinRefreshWaitTime (%s).", migRef.String(), c.migInstancesMinRefreshWaitTime)
			return nil
		}
	}
	klog.V(4).Infof("Regenerating MIG instances cache for %s", migRef.String())
	instances, err := c.gceClient.FetchMigInstances(migRef)
	if err != nil {
		c.migLister.HandleMigIssue(migRef, err)
		return err
	}
	// only save information for successful calls, given the errors above may be transient.
	c.migInstancesLastRefreshedInfo[migRef.String()] = c.timeProvider.Now()
	return c.cache.SetMigInstances(migRef, instances)
}

func (c *cachingMigInfoProvider) GetMigTargetSize(migRef GceRef) (int64, error) {
	c.migInfoMutex.Lock()
	defer c.migInfoMutex.Unlock()

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
		c.migLister.HandleMigIssue(migRef, err)
		return 0, err
	}
	c.cache.SetMigTargetSize(migRef, targetSize)
	return targetSize, nil
}

func (c *cachingMigInfoProvider) GetMigBasename(migRef GceRef) (string, error) {
	c.migInfoMutex.Lock()
	defer c.migInfoMutex.Unlock()

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
		c.migLister.HandleMigIssue(migRef, err)
		return "", err
	}
	c.cache.SetMigBasename(migRef, basename)
	return basename, nil
}

func (c *cachingMigInfoProvider) GetMigInstanceTemplateName(migRef GceRef) (string, error) {
	c.migInfoMutex.Lock()
	defer c.migInfoMutex.Unlock()

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
		c.migLister.HandleMigIssue(migRef, err)
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

// filMigInfoCache needs to be called with migInfoMutex locked
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
	for _, mig := range c.migLister.GetMigs() {
		migRefs[mig.GceRef()] = true
	}
	return migRefs
}

func (c *cachingMigInfoProvider) listAllZonesWithMigs() map[string]bool {
	zones := map[string]bool{}
	for _, mig := range c.migLister.GetMigs() {
		zones[mig.GceRef().Zone] = true
	}
	return zones
}

func (c *cachingMigInfoProvider) GetMigMachineType(migRef GceRef) (MachineType, error) {
	template, err := c.GetMigInstanceTemplate(migRef)
	if err != nil {
		return MachineType{}, err
	}
	machineName := template.Properties.MachineType
	if IsCustomMachine(machineName) {
		return NewCustomMachineType(machineName)
	}
	zone := migRef.Zone
	machine, found := c.cache.GetMachine(machineName, zone)
	if !found {
		rawMachine, err := c.gceClient.FetchMachineType(zone, machineName)
		if err != nil {
			c.migLister.HandleMigIssue(migRef, err)
			return MachineType{}, err
		}
		machine, err = NewMachineTypeFromAPI(machineName, rawMachine)
		if err != nil {
			c.migLister.HandleMigIssue(migRef, err)
			return MachineType{}, err
		}
		c.cache.AddMachine(machine, zone)
	}
	return machine, nil
}
