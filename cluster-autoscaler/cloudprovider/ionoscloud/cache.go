/*
Copyright 2020 The Kubernetes Authors.

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

package ionoscloud

import (
	"sync"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/klog/v2"
)

const nodeGroupCacheEntryTimeout = 2 * time.Minute

type nodeGroupCacheEntry struct {
	data cloudprovider.NodeGroup
	ts   time.Time
}

var timeNow = time.Now

// IonosCache caches resources to reduce API calls.
// Cached state includes autoscaling limits, sizes and target sizes, a mapping of instances to node
// groups, and a simple lock mechanism to prevent invalid node group writes.
type IonosCache struct {
	mutex sync.Mutex

	nodeGroups           map[string]nodeGroupCacheEntry
	nodesToNodeGroups    map[string]string
	instances            map[string]cloudprovider.Instance
	nodeGroupSizes       map[string]int
	nodeGroupTargetSizes map[string]int
	nodeGroupLockTable   map[string]bool
}

// NewIonosCache initializes a new IonosCache.
func NewIonosCache() *IonosCache {
	return &IonosCache{
		nodeGroups:           map[string]nodeGroupCacheEntry{},
		nodesToNodeGroups:    map[string]string{},
		instances:            map[string]cloudprovider.Instance{},
		nodeGroupSizes:       map[string]int{},
		nodeGroupTargetSizes: map[string]int{},
		nodeGroupLockTable:   map[string]bool{},
	}
}

// GetInstancesForNodeGroup returns the list of cached instances a node group.
func (cache *IonosCache) GetInstancesForNodeGroup(id string) []cloudprovider.Instance {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	var nodeIds []string
	for nodeId, nodeGroupId := range cache.nodesToNodeGroups {
		if nodeGroupId == id {
			nodeIds = append(nodeIds, nodeId)
		}
	}
	instances := make([]cloudprovider.Instance, len(nodeIds))
	for i, id := range nodeIds {
		instances[i] = cache.instances[id]
	}
	return instances
}

// AddNodeGroup adds a node group to the cache.
func (cache *IonosCache) AddNodeGroup(newPool cloudprovider.NodeGroup) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.nodeGroups[newPool.Id()] = nodeGroupCacheEntry{data: newPool}
}

func (cache *IonosCache) removeNodesForNodeGroupNoLock(id string) {
	for nodeId, nodeGroupId := range cache.nodesToNodeGroups {
		if nodeGroupId == id {
			delete(cache.instances, nodeId)
			delete(cache.nodesToNodeGroups, nodeId)
		}
	}
}

// RemoveInstanceFromCache deletes an instance and its respective mapping to the node group from
// the cache.
func (cache *IonosCache) RemoveInstanceFromCache(id string) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	klog.V(5).Infof("Removed instance %s from cache", id)
	nodeGroupId := cache.nodesToNodeGroups[id]
	delete(cache.nodesToNodeGroups, id)
	delete(cache.instances, id)
	cache.updateNodeGroupTimestampNoLock(nodeGroupId)
}

// SetInstancesCache overwrites all cached instances and node group mappings.
func (cache *IonosCache) SetInstancesCache(nodeGroupInstances map[string][]cloudprovider.Instance) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.nodesToNodeGroups = map[string]string{}
	cache.instances = map[string]cloudprovider.Instance{}

	for id, instances := range nodeGroupInstances {
		cache.setInstancesCacheForNodeGroupNoLock(id, instances)
	}
}

// SetInstancesCacheForNodeGroup overwrites cached instances and mappings for a node group.
func (cache *IonosCache) SetInstancesCacheForNodeGroup(id string, instances []cloudprovider.Instance) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.removeNodesForNodeGroupNoLock(id)
	cache.setInstancesCacheForNodeGroupNoLock(id, instances)
}

func (cache *IonosCache) setInstancesCacheForNodeGroupNoLock(id string, instances []cloudprovider.Instance) {
	for _, instance := range instances {
		nodeId := convertToNodeId(instance.Id)
		cache.nodesToNodeGroups[nodeId] = id
		cache.instances[nodeId] = instance
	}
	cache.updateNodeGroupTimestampNoLock(id)
}

// GetNodeGroupIds returns an unsorted list of cached node group ids.
func (cache *IonosCache) GetNodeGroupIds() []string {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	return cache.getNodeGroupIds()
}

func (cache *IonosCache) getNodeGroupIds() []string {
	ids := make([]string, 0, len(cache.nodeGroups))
	for id := range cache.nodeGroups {
		ids = append(ids, id)
	}
	return ids
}

// GetNodeGroups returns an unsorted list of cached node groups.
func (cache *IonosCache) GetNodeGroups() []cloudprovider.NodeGroup {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	nodeGroups := make([]cloudprovider.NodeGroup, 0, len(cache.nodeGroups))
	for id := range cache.nodeGroups {
		nodeGroups = append(nodeGroups, cache.nodeGroups[id].data)
	}
	return nodeGroups
}

// GetInstances returns an unsorted list of all cached instances.
func (cache *IonosCache) GetInstances() []cloudprovider.Instance {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	instances := make([]cloudprovider.Instance, 0, len(cache.nodesToNodeGroups))
	for _, instance := range cache.instances {
		instances = append(instances, instance)
	}
	return instances
}

// GetNodeGroupForNode returns the node group for the given node.
// Returns nil if either the mapping or the node group is not cached.
func (cache *IonosCache) GetNodeGroupForNode(nodeId string) cloudprovider.NodeGroup {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	id, found := cache.nodesToNodeGroups[nodeId]
	if !found {
		return nil
	}
	entry, found := cache.nodeGroups[id]
	if !found {
		return nil
	}
	return entry.data
}

// TryLockNodeGroup tries to write a node group lock entry.
// Returns true if the write was successful.
func (cache *IonosCache) TryLockNodeGroup(nodeGroup cloudprovider.NodeGroup) bool {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if cache.nodeGroupLockTable[nodeGroup.Id()] {
		return false
	}

	klog.V(4).Infof("Acquired lock for node group %s", nodeGroup.Id())
	cache.nodeGroupLockTable[nodeGroup.Id()] = true
	return true
}

// UnlockNodeGroup deletes a node group lock entry if it exists.
func (cache *IonosCache) UnlockNodeGroup(nodeGroup cloudprovider.NodeGroup) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if _, found := cache.nodeGroupLockTable[nodeGroup.Id()]; found {
		klog.V(5).Infof("Released lock for node group %s", nodeGroup.Id())
		delete(cache.nodeGroupLockTable, nodeGroup.Id())
	}
}

// GetNodeGroupSize gets the node group's size. Return true if the size was in the cache.
func (cache *IonosCache) GetNodeGroupSize(id string) (int, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	size, found := cache.nodeGroupSizes[id]
	return size, found
}

// SetNodeGroupSize sets the node group's size.
func (cache *IonosCache) SetNodeGroupSize(id string, size int) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	previousSize := cache.nodeGroupSizes[id]
	klog.V(5).Infof("Updated node group size cache: nodeGroup=%s, previousSize=%d, currentSize=%d", id, previousSize, size)
	cache.nodeGroupSizes[id] = size
}

// GetNodeGroupTargetSize gets the node group's target size. Return true if the target size was in the
// cache.
func (cache *IonosCache) GetNodeGroupTargetSize(id string) (int, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	size, found := cache.nodeGroupTargetSizes[id]
	return size, found
}

// SetNodeGroupTargetSize sets the node group's target size.
func (cache *IonosCache) SetNodeGroupTargetSize(id string, size int) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	previousSize := cache.nodeGroupSizes[id]
	klog.V(5).Infof("Updated node group target size cache: nodeGroup=%s, previousTargetSize=%d, currentTargetSize=%d", id, previousSize, size)
	cache.nodeGroupTargetSizes[id] = size
}

// InvalidateNodeGroupTargetSize deletes a node group's target size entry.
func (cache *IonosCache) InvalidateNodeGroupTargetSize(id string) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	delete(cache.nodeGroupTargetSizes, id)
}

// NodeGroupNeedsRefresh returns true when the instances for the given node group have not been
// updated for more than 2 minutes.
func (cache *IonosCache) NodeGroupNeedsRefresh(id string) bool {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	return timeNow().Sub(cache.nodeGroups[id].ts) > nodeGroupCacheEntryTimeout
}

func (cache *IonosCache) updateNodeGroupTimestampNoLock(id string) {
	if entry, ok := cache.nodeGroups[id]; ok {
		entry.ts = timeNow()
		cache.nodeGroups[id] = entry
	}
}
