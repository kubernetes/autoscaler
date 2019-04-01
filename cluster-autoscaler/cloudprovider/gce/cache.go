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

	gce "google.golang.org/api/compute/v1"
	"k8s.io/klog"
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
// - instanceRefToMigRef (2) is based on registered migs (1). For each mig, its instances
// are fetched from GCE API using gceService.
// - instanceRefToMigRef (2) is NOT updated automatically when migs field (1) is updated. Calling
// RegenerateInstancesCache is required to sync it with registered migs.
type GceCache struct {
	// Cache content.
	migs                map[GceRef]*MigInformation
	instanceRefToMigRef map[GceRef]GceRef
	resourceLimiter     *cloudprovider.ResourceLimiter
	machinesCache       map[MachineTypeKey]*gce.MachineType
	migTargetSizeCache  map[GceRef]int64
	// Locks. Rules of locking:
	// - migsMutex protects only migs.
	// - cacheMutex protects instanceRefToMigRef, resourceLimiter, machinesCache and migTargetSizeCache.
	// - if both locks are needed, cacheMutex must be obtained before migsMutex.
	cacheMutex sync.Mutex
	migsMutex  sync.Mutex
	// Service used to refresh cache.
	GceService AutoscalingGceClient
}

// NewGceCache creates empty GceCache.
func NewGceCache(gceService AutoscalingGceClient) GceCache {
	return GceCache{
		migs:                map[GceRef]*MigInformation{},
		instanceRefToMigRef: map[GceRef]GceRef{},
		machinesCache:       map[MachineTypeKey]*gce.MachineType{},
		GceService:          gceService,
		migTargetSizeCache:  map[GceRef]int64{},
	}
}

//  Methods locking on migsMutex.

// RegisterMig returns true if the node group wasn't in cache before, or its config was updated.
func (gc *GceCache) RegisterMig(newMig Mig) bool {
	gc.migsMutex.Lock()
	defer gc.migsMutex.Unlock()

	oldMigInformation, found := gc.migs[newMig.GceRef()]
	if found {
		if !reflect.DeepEqual(oldMigInformation.Config, newMig) {
			gc.migs[newMig.GceRef()].Config = newMig
			klog.V(4).Infof("Updated Mig %s", newMig.GceRef().String())
			return true
		}
		return false
	}

	klog.V(1).Infof("Registering %s", newMig.GceRef().String())
	// TODO(aleksandra-malinowska): fetch and set MIG basename here.
	newMigInformation := &MigInformation{
		Config: newMig,
	}
	gc.migs[newMig.GceRef()] = newMigInformation
	return true
}

// UnregisterMig returns true if the node group has been removed, and false if it was already missing from cache.
func (gc *GceCache) UnregisterMig(toBeRemoved Mig) bool {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()
	gc.migsMutex.Lock()
	defer gc.migsMutex.Unlock()

	_, found := gc.migs[toBeRemoved.GceRef()]
	if found {
		klog.V(1).Infof("Unregistered Mig %s", toBeRemoved.GceRef().String())
		delete(gc.migs, toBeRemoved.GceRef())
		gc.removeInstancesForMig(toBeRemoved.GceRef())
		return true
	}
	return false
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

func (gc *GceCache) updateMigBasename(migRef GceRef, basename string) {
	gc.migsMutex.Lock()
	defer gc.migsMutex.Unlock()

	mig, found := gc.migs[migRef]
	if found {
		mig.Basename = basename
	}
	// TODO: is found == false a possiblity?
}

// Methods locking on cacheMutex.

// GetMigForInstance returns Mig to which the given instance belongs.
// Attempts to regenerate cache if there is a Mig with matching prefix in migs list.
// TODO(aleksandra-malinowska): reconsider failing when there's a Mig with
// matching prefix, but instance doesn't belong to it.
func (gc *GceCache) GetMigForInstance(instanceRef GceRef) (Mig, error) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	if migRef, found := gc.instanceRefToMigRef[instanceRef]; found {
		mig, found := gc.getMig(migRef)
		if !found {
			return nil, fmt.Errorf("instance %+v belongs to unregistered mig %+v", instanceRef, migRef)
		}
		return mig.Config, nil
	}

	for _, mig := range gc.GetMigs() {
		if mig.Config.GceRef().Project == instanceRef.Project &&
			mig.Config.GceRef().Zone == instanceRef.Zone &&
			strings.HasPrefix(instanceRef.Name, mig.Basename) {
			if err := gc.regenerateCache(); err != nil {
				return nil, fmt.Errorf("error while looking for MIG for instance %+v, error: %v", instanceRef, err)
			}

			migRef, found := gc.instanceRefToMigRef[instanceRef]
			if !found {
				return nil, fmt.Errorf("instance %+v belongs to unknown mig", instanceRef)
			}
			mig, found := gc.getMig(migRef)
			if !found {
				return nil, fmt.Errorf("instance %+v belongs to unregistered mig %+v", instanceRef, migRef)
			}
			return mig.Config, nil
		}
	}
	// Instance doesn't belong to any configured mig.
	return nil, nil
}

func (gc *GceCache) removeInstancesForMig(migRef GceRef) {
	for instanceRef, instanceMigRef := range gc.instanceRefToMigRef {
		if migRef == instanceMigRef {
			delete(gc.instanceRefToMigRef, instanceRef)
		}
	}
}

func (gc *GceCache) getMig(migRef GceRef) (*MigInformation, bool) {
	gc.migsMutex.Lock()
	defer gc.migsMutex.Unlock()
	mig, found := gc.migs[migRef]
	return mig, found
}

// RegenerateInstancesCache triggers instances cache regeneration under lock.
func (gc *GceCache) RegenerateInstancesCache() error {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	return gc.regenerateCache()
}

// internal method - should only be called after locking on cacheMutex.
func (gc *GceCache) regenerateCache() error {
	newInstancesCache := make(map[GceRef]GceRef)
	for _, migInfo := range gc.GetMigs() {
		mig := migInfo.Config
		klog.V(4).Infof("Regenerating MIG information for %s", mig.GceRef().String())

		basename, err := gc.GceService.FetchMigBasename(mig.GceRef())
		if err != nil {
			return err
		}
		gc.updateMigBasename(mig.GceRef(), basename)

		instances, err := gc.GceService.FetchMigInstances(mig.GceRef())
		if err != nil {
			klog.V(4).Infof("Failed MIG info request for %s: %v", mig.GceRef().String(), err)
			return err
		}
		for _, instance := range instances {
			gceRef, err := GceRefFromProviderId(instance.Id)
			if err != nil {
				return err
			}
			newInstancesCache[gceRef] = mig.GceRef()
		}
	}

	gc.instanceRefToMigRef = newInstancesCache
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

// GetMigTargetSize returns the cached targetSize for a GceRef
func (gc *GceCache) GetMigTargetSize(ref GceRef) (int64, bool) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	size, found := gc.migTargetSizeCache[ref]
	if found {
		klog.V(5).Infof("target size cache hit for %s", ref)
	}
	return size, found
}

// SetMigTargetSize sets targetSize for a GceRef
func (gc *GceCache) SetMigTargetSize(ref GceRef, size int64) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	gc.migTargetSizeCache[ref] = size
}

// InvalidateTargetSizeCache clears the target size cache
func (gc *GceCache) InvalidateTargetSizeCache() {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	klog.V(5).Infof("target size cache invalidated")
	gc.migTargetSizeCache = map[GceRef]int64{}
}

// InvalidateTargetSizeCacheForMig clears the target size cache
func (gc *GceCache) InvalidateTargetSizeCacheForMig(ref GceRef) {
	gc.cacheMutex.Lock()
	defer gc.cacheMutex.Unlock()

	if _, found := gc.migTargetSizeCache[ref]; found {
		klog.V(5).Infof("target size cache invalidated for %s", ref)
		delete(gc.migTargetSizeCache, ref)
	}
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
