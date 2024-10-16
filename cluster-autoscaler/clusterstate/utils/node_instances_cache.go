/*
Copyright 2016 The Kubernetes Authors.

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

package utils

import (
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"

	klog "k8s.io/klog/v2"
)

const (
	// CloudProviderNodeInstancesCacheRefreshInterval is the interval between nodegroup instances cache refreshes.
	CloudProviderNodeInstancesCacheRefreshInterval = 2 * time.Minute
	// CloudProviderNodeInstancesCacheEntryFreshnessThreshold is the duration when cache entry is fresh.
	CloudProviderNodeInstancesCacheEntryFreshnessThreshold = 5 * time.Minute
)

type cloudProviderNodeInstancesCacheEntry struct {
	instances   []cloudprovider.Instance
	refreshTime time.Time
}

func (cacheEntry *cloudProviderNodeInstancesCacheEntry) isStale() bool {
	return cacheEntry.refreshTime.Add(CloudProviderNodeInstancesCacheEntryFreshnessThreshold).Before(time.Now())
}

// CloudProviderNodeInstancesCache caches cloud provider node instances.
type CloudProviderNodeInstancesCache struct {
	sync.Mutex
	cloudProviderNodeInstances map[string]*cloudProviderNodeInstancesCacheEntry
	cloudProvider              cloudprovider.CloudProvider
}

// NewCloudProviderNodeInstancesCache creates new cache instance.
func NewCloudProviderNodeInstancesCache(cloudProvider cloudprovider.CloudProvider) *CloudProviderNodeInstancesCache {
	return &CloudProviderNodeInstancesCache{
		cloudProviderNodeInstances: map[string]*cloudProviderNodeInstancesCacheEntry{},
		cloudProvider:              cloudProvider,
	}
}

func (cache *CloudProviderNodeInstancesCache) updateCacheEntryLocked(nodeGroup cloudprovider.NodeGroup, cacheEntry *cloudProviderNodeInstancesCacheEntry) {
	cache.Lock()
	defer cache.Unlock()
	cache.cloudProviderNodeInstances[nodeGroup.Id()] = cacheEntry
}

func (cache *CloudProviderNodeInstancesCache) removeCacheEntryLocked(nodeGroup cloudprovider.NodeGroup) {
	cache.Lock()
	defer cache.Unlock()
	delete(cache.cloudProviderNodeInstances, nodeGroup.Id())
}

func (cache *CloudProviderNodeInstancesCache) getCacheEntryLocked(nodeGroup cloudprovider.NodeGroup) (*cloudProviderNodeInstancesCacheEntry, bool) {
	cache.Lock()
	defer cache.Unlock()
	cacheEntry, found := cache.cloudProviderNodeInstances[nodeGroup.Id()]
	return cacheEntry, found
}

func (cache *CloudProviderNodeInstancesCache) removeEntriesForNonExistingNodeGroupsLocked(nodeGroups []cloudprovider.NodeGroup) {
	cache.Lock()
	defer cache.Unlock()
	nodeGroupExists := map[string]bool{}
	for _, nodeGroup := range nodeGroups {
		nodeGroupExists[nodeGroup.Id()] = true
	}
	for nodeGroupId := range cache.cloudProviderNodeInstances {
		if !nodeGroupExists[nodeGroupId] {
			delete(cache.cloudProviderNodeInstances, nodeGroupId)
		}
	}
}

// GetCloudProviderNodeInstances returns cloud provider node instances for all node groups returned by cloud provider.
func (cache *CloudProviderNodeInstancesCache) GetCloudProviderNodeInstances() (map[string][]cloudprovider.Instance, error) {
	nodeGroups := cache.cloudProvider.NodeGroups()

	// Fetch missing node instances.
	var wg sync.WaitGroup
	for _, nodeGroup := range nodeGroups {
		nodeGroup := nodeGroup
		if _, found := cache.getCacheEntryLocked(nodeGroup); !found {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if _, err := cache.fetchCloudProviderNodeInstancesForNodeGroup(nodeGroup); err != nil {
					klog.Errorf("Failed to fetch cloud provider node instances for %v, error %v", nodeGroup.Id(), err)
				}
			}()
		}
	}
	wg.Wait()

	// Get data from cache.
	results := map[string][]cloudprovider.Instance{}
	for _, nodeGroup := range nodeGroups {
		nodeGroupInstances, err := cache.GetCloudProviderNodeInstancesForNodeGroup(nodeGroup)
		if err != nil {
			return nil, err
		}
		results[nodeGroup.Id()] = nodeGroupInstances
	}
	return results, nil
}

// GetCloudProviderNodeInstancesForNodeGroup returns cloud provider node instances for the given node group.
func (cache *CloudProviderNodeInstancesCache) GetCloudProviderNodeInstancesForNodeGroup(nodeGroup cloudprovider.NodeGroup) ([]cloudprovider.Instance, error) {
	cacheEntry, found := cache.getCacheEntryLocked(nodeGroup)
	if found {
		if cacheEntry.isStale() {
			klog.Warningf("Entry cloudProviderNodeInstances is stale, refresh time is %v", cacheEntry.refreshTime)
		}
		klog.V(5).Infof("Get cached cloud provider node instances for %v", nodeGroup.Id())
		return cacheEntry.instances, nil
	}
	klog.V(5).Infof("Cloud provider node instances for %v hasn't been found in cache, fetch from cloud provider", nodeGroup.Id())
	return cache.fetchCloudProviderNodeInstancesForNodeGroup(nodeGroup)
}

func (cache *CloudProviderNodeInstancesCache) fetchCloudProviderNodeInstancesForNodeGroup(nodeGroup cloudprovider.NodeGroup) ([]cloudprovider.Instance, error) {
	nodeGroupInstances, err := nodeGroup.Nodes()
	if err != nil {
		return nil, err
	}
	cache.updateCacheEntryLocked(nodeGroup, &cloudProviderNodeInstancesCacheEntry{nodeGroupInstances, time.Now()})
	return nodeGroupInstances, nil
}

// InvalidateCacheEntry removes entry for the given node group from cache.
func (cache *CloudProviderNodeInstancesCache) InvalidateCacheEntry(nodeGroup cloudprovider.NodeGroup) {
	klog.V(5).Infof("Invalidate entry in cloud provider node instances cache %v", nodeGroup.Id())
	cache.removeCacheEntryLocked(nodeGroup)
}

// Refresh refreshes cache.
func (cache *CloudProviderNodeInstancesCache) Refresh() {
	klog.V(4).Infof("Start refreshing cloud provider node instances cache")
	refreshStart := time.Now()

	nodeGroups := cache.cloudProvider.NodeGroups()
	cache.removeEntriesForNonExistingNodeGroupsLocked(nodeGroups)
	for _, nodeGroup := range nodeGroups {
		nodeGroupInstances, err := nodeGroup.Nodes()
		if err != nil {
			klog.Errorf("Failed to get cloud provider node instance for node group %v, error %v", nodeGroup.Id(), err)
		}
		cache.updateCacheEntryLocked(nodeGroup, &cloudProviderNodeInstancesCacheEntry{nodeGroupInstances, time.Now()})
	}
	klog.Infof("Refresh cloud provider node instances cache finished, refresh took %v", time.Now().Sub(refreshStart))
}

// Start starts components running in background.
func (cache *CloudProviderNodeInstancesCache) Start(interrupt chan struct{}) {
	go wait.Until(func() {
		cache.Refresh()
	}, CloudProviderNodeInstancesCacheRefreshInterval, interrupt)
}
