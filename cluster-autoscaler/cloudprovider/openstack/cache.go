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

package openstack

import (
	"fmt"
	"reflect"
	"sync"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"

	"k8s.io/klog"
)

// ASGInformation is a wrapper for ASG.
type ASGInformation struct {
	Config   ASG
}

// OpenStackCache is used for caching cluster resources state.
//
// It is needed to:
// - keep track of autoscaled ASGs in the cluster,
// - keep track of instances and which ASG they belong to,
// - limit repetitive OpenStack API calls.
//
// Cached resources:
// 1) ASG configuration,
// 2) instance->ASG mapping,
// 3) resource limits (self-imposed quotas),
//
// How it works:
// - asgs (1), resource limits (3) and machine types (4) are only stored in this cache,
// not updated by it.
// - instancesCache (2) is based on registered asgs (1). For each asg, its instances
// are fetched from OpenStack API using openstackService.
// - instancesCache (2) is NOT updated automatically when asgs field (1) is updated. Calling
// RegenerateInstancesCache is required to sync it with registered asgs.
type OpenStackCache struct {
	// Cache content.
	asgs            []*ASGInformation
	instancesCache  map[OpenStackRef]ASG
	resourceLimiter *cloudprovider.ResourceLimiter
	// Locks. Rules of locking:
	// - asgsMutex protects only asgs.
	// - cacheMutex protects instancesCache, resourceLimiter and machinesCache.
	// - if both locks are needed, cacheMutex must be obtained before asgsMutex.
	cacheMutex sync.Mutex
	asgsMutex  sync.Mutex
	// Service used to refresh cache.
	OpenStackService AutoscalingOpenStackClient
}

// NewOpenStackCache creates empty OpenStackCache.
func NewOpenStackCache(openstackService AutoscalingOpenStackClient) OpenStackCache {
	return OpenStackCache{
		asgs:           []*ASGInformation{},
		instancesCache: map[OpenStackRef]ASG{},
		OpenStackService:     openstackService,
	}
}

//  Methods locking on asgsMutex.

// RegisterASG returns true if the node group wasn't in cache before, or its config was updated.
func (osc *OpenStackCache) RegisterASG(asg ASG) bool {
	osc.asgsMutex.Lock()
	defer osc.asgsMutex.Unlock()

	for i := range osc.asgs {
		if oldASG := osc.asgs[i].Config; oldASG.OpenStackRef() == asg.OpenStackRef() {
			if !reflect.DeepEqual(oldASG, asg) {
				osc.asgs[i].Config = asg
				klog.V(4).Infof("Updated ASG %s", asg.OpenStackRef().String())
				return true
			}
			return false
		}
	}

	klog.V(1).Infof("Registering %s", asg.OpenStackRef().String())
	osc.asgs = append(osc.asgs, &ASGInformation{
		Config: asg,
	})
	return true
}

// UnregisterASG returns true if the node group has been removed, and false if it was already missing from cache.
func (osc *OpenStackCache) UnregisterASG(toBeRemoved ASG) bool {
	osc.asgsMutex.Lock()
	defer osc.asgsMutex.Unlock()

	newASGs := make([]*ASGInformation, 0, len(osc.asgs))
	found := false
	for _, asg := range osc.asgs {
		if asg.Config.OpenStackRef() == toBeRemoved.OpenStackRef() {
			klog.V(1).Infof("Unregistered ASG %s", toBeRemoved.OpenStackRef().String())
			found = true
		} else {
			newASGs = append(newASGs, asg)
		}
	}
	osc.asgs = newASGs
	return found
}

// GetASGs returns a copy of asgs list.
func (osc *OpenStackCache) GetASGs() []*ASGInformation {
	osc.asgsMutex.Lock()
	defer osc.asgsMutex.Unlock()

	asgs := make([]*ASGInformation, 0, len(osc.asgs))
	for _, asg := range osc.asgs {
		asgs = append(asgs, &ASGInformation{
			Config:   asg.Config,
		})
	}
	return asgs
}

// Methods locking on cacheMutex.

// GetASGForInstance returns ASG to which the given instance belongs.
// Attempts to regenerate cache if there is a ASG with matching prefix in asgs list.
// matching prefix, but instance doesn't belong to it.
func (osc *OpenStackCache) GetASGForInstance(instance *OpenStackRef) (ASG, error) {
	osc.cacheMutex.Lock()
	defer osc.cacheMutex.Unlock()

	if asg, found := osc.instancesCache[*instance]; found {
		return asg, nil
	}

	for _, asg := range osc.GetASGs() {
		if asg.Config.OpenStackRef().Project == instance.Project {
			if err := osc.regenerateCache(); err != nil {
				return nil, fmt.Errorf("Error while looking for ASG for instance %+v, error: %v", *instance, err)
			}
			if asg, found := osc.instancesCache[*instance]; found {
				return asg, nil
			}
			return nil, fmt.Errorf("Instance %+v does not belong to any configured ASG", *instance)
		}
	}
	// Instance doesn't belong to any configured asg.
	return nil, nil
}

// RegenerateInstancesCache triggers instances cache regeneration under lock.
func (osc *OpenStackCache) RegenerateInstancesCache() error {
	osc.cacheMutex.Lock()
	defer osc.cacheMutex.Unlock()

	return osc.regenerateCache()
}

// internal method - should only be called after locking on cacheMutex.
func (osc *OpenStackCache) regenerateCache() error {
	newInstancesCache := make(map[OpenStackRef]ASG)

	for _, asgInfo := range osc.GetASGs() {
		asg := asgInfo.Config
		klog.V(4).Infof("Regenerating ASG information for %s", asg.OpenStackRef().String())

		instances, err := osc.OpenStackService.FetchASGInstances(asg.OpenStackRef())
		if err != nil {
			klog.V(4).Infof("Failed ASG info request for %s: %v", asg.OpenStackRef().String(), err)
			return err
		}
		for _, ref := range instances {
			newInstancesCache[ref] = asg
		}
	}

	osc.instancesCache = newInstancesCache
	return nil
}

// SetResourceLimiter sets resource limiter.
func (osc *OpenStackCache) SetResourceLimiter(resourceLimiter *cloudprovider.ResourceLimiter) {
	osc.cacheMutex.Lock()
	defer osc.cacheMutex.Unlock()

	osc.resourceLimiter = resourceLimiter
}

// GetResourceLimiter returns resource limiter.
func (osc *OpenStackCache) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	osc.cacheMutex.Lock()
	defer osc.cacheMutex.Unlock()

	return osc.resourceLimiter, nil
}
