/*
Copyright 2018 The Kubernetes Authors.

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
	"fmt"
	"reflect"
	"strings"
	"sync"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"

	"github.com/golang/glog"
	gce "google.golang.org/api/compute/v1"
)

// MigInformation is a wrapper for Mig.
type MigInformation struct {
	Config   Mig
	Basename string
}

// MachineTypeKey is used to identify MachineType.
type MachineTypeKey struct {
	Zone        string
	MachineType string
}

// GceCache is used for caching cluster resources state.
//
// It is needed to:
// - keep track of autoscaled MIGs in the cluster,
// - keep track of instances and which MIG they belong to,
// - limit repetitive GCE API calls.
//
// Cached resources:
// 1) MIG configuration,
// 2) instance->MIG mapping,
// 3) resource limits (self-imposed quotas),
// 4) machine types.
//
// How it works:
// - migs (1), resource limits (3) and machine types (4) are only stored in this cache,
// not updated by it.
// - instancesCache (2) is based on registered migs (1). For each mig, its instances
// are fetched from GCE API using gceService.
// - instancesCache (2) is NOT updated automatically when migs field (1) is updated. Calling
// RegenerateInstancesCache is required to sync it with registered migs.
type GceCache struct {
	// Cache content.
	migs            []*MigInformation
	instancesCache  map[GceRef]Mig
	resourceLimiter *cloudprovider.ResourceLimiter
	machinesCache   map[MachineTypeKey]*gce.MachineType
	// Locks. Rules of locking:
	// - migsMutex protects only migs.
	// - cacheMutex protects instancesCache, resourceLimiter and machinesCache.
	// - if both locks are needed, cacheMutex must be obtained before migsMutex.
	cacheMutex sync.Mutex
	migsMutex  sync.Mutex
	// Service used to refresh cache.
	GceService AutoscalingGceClient
}

// NewGceCache creates empty GceCache.
func NewGceCache(gceService AutoscalingGceClient) GceCache {
	return GceCache{
		migs:           []*MigInformation{},
		instancesCache: map[GceRef]Mig{},
		machinesCache:  map[MachineTypeKey]*gce.MachineType{},
		GceService:     gceService,
	}
}

//  Methods locking on migsMutex.

// RegisterMig returns true if the node group wasn't in cache before, or its config was updated.
func (gc *GceCache) RegisterMig(mig Mig) bool {
	gc.migsMutex.Lock()
	defer gc.migsMutex.Unlock()

	for i := range gc.migs {
		if oldMig := gc.migs[i].Config; oldMig.GceRef() == mig.GceRef() {
			if !reflect.DeepEqual(oldMig, mig) {
				gc.migs[i].Config = mig
				glog.V(4).Infof("Updated Mig %s", mig.GceRef().String())
				return true
			}
			return false
		}
	}

	glog.V(1).Infof("Registering %s", mig.GceRef().String())
	// TODO(aleksandra-malinowska): fetch and set MIG basename here.
	gc.migs = append(gc.migs, &MigInformation{
		Config: mig,
	})
	return true
}

// UnregisterMig returns true if the node group has been removed, and false if it was alredy missing from cache.
func (gc *GceCache) UnregisterMig(toBeRemoved Mig) bool {
	gc.migsMutex.Lock()
	defer gc.migsMutex.Unlock()

	newMigs := make([]*MigInformation, 0, len(gc.migs))
	found := false
	for _, mig := range gc.migs {
		if mig.Config.GceRef() == toBeRemoved.GceRef() {
			glog.V(1).Infof("Unregistered Mig %s", toBeRemoved.GceRef().String())
			found = true
		} else {
			newMigs = append(newMigs, mig)
		}
	}
	gc.migs = newMigs
	return found
}

// GetMigs returns a copy of migs list.
func (gc *GceCache) GetMigs() []*MigInformation {
	gc.migsMutex.Lock()
	defer gc.migsMutex.Unlock()

	migs := make([]*MigInformation, 0, len(gc.migs))
	for _, mig := range gc.migs {
		migs = append(migs, &MigInformation{
			Basename: mig.Basename,
			Config:   mig.Config,
		})
	}
	return migs
}

func (gc *GceCache) updateMigBasename(ref GceRef, basename string) {
	gc.migsMutex.Lock()
	defer gc.migsMutex.Unlock()

	for _, mig := range gc.migs {
		if mig.Config.GceRef() == ref {
			mig.Basename = basename
		}
	}
}

// Methods locking on cacheMutex.

// GetMigForInstance returns Mig to which the given instance belongs.
// Attempts to regenerate cache if there is a Mig with matching prefix in migs list.
// TODO(aleksandra-malinowska): reconsider failing when there's a Mig with
// matching prefix, but instance doesn't belong to it.
func (gc *GceCache) GetMigForInstance(instance *GceRef) (Mig, error) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	if mig, found := gc.instancesCache[*instance]; found {
		return mig, nil
	}

	for _, mig := range gc.GetMigs() {
		if mig.Config.GceRef().Project == instance.Project &&
			mig.Config.GceRef().Zone == instance.Zone &&
			strings.HasPrefix(instance.Name, mig.Basename) {
			if err := gc.regenerateCache(); err != nil {
				return nil, fmt.Errorf("Error while looking for MIG for instance %+v, error: %v", *instance, err)
			}
			if mig, found := gc.instancesCache[*instance]; found {
				return mig, nil
			}
			return nil, fmt.Errorf("Instance %+v does not belong to any configured MIG", *instance)
		}
	}
	// Instance doesn't belong to any configured mig.
	return nil, nil
}

// RegenerateInstancesCache triggers instances cache regeneration under lock.
func (gc *GceCache) RegenerateInstancesCache() error {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	return gc.regenerateCache()
}

// internal method - should only be called after locking on cacheMutex.
func (gc *GceCache) regenerateCache() error {
	newInstancesCache := make(map[GceRef]Mig)

	for _, migInfo := range gc.GetMigs() {
		mig := migInfo.Config
		glog.V(4).Infof("Regenerating MIG information for %s", mig.GceRef().String())

		basename, err := gc.GceService.FetchMigBasename(mig.GceRef())
		if err != nil {
			return err
		}
		gc.updateMigBasename(mig.GceRef(), basename)

		instances, err := gc.GceService.FetchMigInstances(mig.GceRef())
		if err != nil {
			glog.V(4).Infof("Failed MIG info request for %s: %v", mig.GceRef().String(), err)
			return err
		}
		for _, ref := range instances {
			newInstancesCache[ref] = mig
		}
	}

	gc.instancesCache = newInstancesCache
	return nil
}

// SetResourceLimiter sets resource limiter.
func (gc *GceCache) SetResourceLimiter(resourceLimiter *cloudprovider.ResourceLimiter) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	gc.resourceLimiter = resourceLimiter
}

// GetResourceLimiter returns resource limiter.
func (gc *GceCache) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	return gc.resourceLimiter, nil
}

// GetMachineFromCache retrieves machine type from cache under lock.
func (gc *GceCache) GetMachineFromCache(machineType string, zone string) *gce.MachineType {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	return gc.machinesCache[MachineTypeKey{zone, machineType}]
}

// AddMachineToCache adds machine to cache under lock.
func (gc *GceCache) AddMachineToCache(machineType string, zone string, machine *gce.MachineType) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	gc.machinesCache[MachineTypeKey{zone, machineType}] = machine
}

// SetMachinesCache sets the machines cache under lock.
func (gc *GceCache) SetMachinesCache(machinesCache map[MachineTypeKey]*gce.MachineType) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	gc.machinesCache = machinesCache
}
