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
	"maps"
	"sync"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/klog/v2"
)

// IonosCache caches resources to reduce API calls.
// Cached state includes autoscaling limits, sizes and target sizes, a mapping of instances to node
// groups, and a simple lock mechanism to prevent invalid node group writes.
type IonosCache struct {
	mutex sync.Mutex

	nodeGroups           map[string]cloudprovider.NodeGroup
	nodesToNodeGroups    map[string]string
	nodeGroupSizes       map[string]int
	nodeGroupTargetSizes map[string]int
	nodeGroupLockTable   map[string]bool
}

// NewIonosCache initializes a new IonosCache.
func NewIonosCache() *IonosCache {
	return &IonosCache{
		nodeGroups:           make(map[string]cloudprovider.NodeGroup),
		nodesToNodeGroups:    make(map[string]string),
		nodeGroupSizes:       make(map[string]int),
		nodeGroupTargetSizes: make(map[string]int),
		nodeGroupLockTable:   make(map[string]bool),
	}
}

// AddNodeGroup adds a node group to the cache.
func (cache *IonosCache) AddNodeGroup(newPool cloudprovider.NodeGroup) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.nodeGroups[newPool.Id()] = newPool
}

// RemoveInstanceFromCache deletes an instance and its respective mapping to the node group from
// the cache.
func (cache *IonosCache) RemoveInstanceFromCache(id string) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if _, ok := cache.nodesToNodeGroups[id]; ok {
		delete(cache.nodesToNodeGroups, id)
		klog.V(5).Infof("Removed instance %s from cache", id)
	}
}

// SetInstancesCacheForNodeGroup overwrites cached instances and mappings for a node group.
func (cache *IonosCache) SetInstancesCacheForNodeGroup(id string, instances []cloudprovider.Instance) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	maps.DeleteFunc(cache.nodesToNodeGroups, func(_, nodeGroupID string) bool {
		return nodeGroupID == id
	})
	cache.setInstancesCacheForNodeGroupNoLock(id, instances)
}

func (cache *IonosCache) setInstancesCacheForNodeGroupNoLock(id string, instances []cloudprovider.Instance) {
	for _, instance := range instances {
		nodeID := convertToNodeID(instance.Id)
		cache.nodesToNodeGroups[nodeID] = id
	}
}

// GetNodeGroupIDs returns an unsorted list of cached node group ids.
func (cache *IonosCache) GetNodeGroupIDs() []string {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	return cache.getNodeGroupIDs()
}

func (cache *IonosCache) getNodeGroupIDs() []string {
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
		nodeGroups = append(nodeGroups, cache.nodeGroups[id])
	}
	return nodeGroups
}

// GetNodeGroupForNode returns the node group for the given node.
// Returns nil if either the mapping or the node group is not cached.
func (cache *IonosCache) GetNodeGroupForNode(nodeID string) cloudprovider.NodeGroup {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	nodeGroupID, found := cache.nodesToNodeGroups[nodeID]
	if !found {
		return nil
	}
	nodeGroup, found := cache.nodeGroups[nodeGroupID]
	if !found {
		return nil
	}
	return nodeGroup
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
