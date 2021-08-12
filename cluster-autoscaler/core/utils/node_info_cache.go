package utils

import (
	"time"

	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// NodeInfoCache - Caches node-group IDs -> NodeInfo with a given expiration.
// Note: this cache is not thread-safe.
type NodeInfoCache struct {
	nodeInfo   map[string]cacheEntry
	expiration time.Duration
}

type cacheEntry struct {
	*schedulerframework.NodeInfo
	lastUpdated time.Time
}

const NoExpiration = -1

func NewNodeInfoCache(expiration time.Duration) *NodeInfoCache {
	return &NodeInfoCache{
		nodeInfo:   make(map[string]cacheEntry),
		expiration: expiration,
	}
}

func (c *NodeInfoCache) Set(nodeGroupID string, nodeInfo *schedulerframework.NodeInfo) {
	c.nodeInfo[nodeGroupID] = cacheEntry{NodeInfo: nodeInfo, lastUpdated: time.Now()}
}

var now = time.Now

func (c *NodeInfoCache) Get(nodeGroupID string) (*schedulerframework.NodeInfo, bool) {
	info, found := c.nodeInfo[nodeGroupID]
	if !found {
		return nil, false
	}

	if c.expiration == NoExpiration {
		return info.NodeInfo, true
	}

	expirationTime := info.lastUpdated.Add(c.expiration)
	if expired := now().After(expirationTime); expired {
		delete(c.nodeInfo, nodeGroupID)
		return nil, false
	}

	return info.NodeInfo, true
}

func (c *NodeInfoCache) Delete(nodeGroupID string) {
	delete(c.nodeInfo, nodeGroupID)
}

func (c *NodeInfoCache) Range(handler func(nodeGroupID string, nodeInfo *schedulerframework.NodeInfo)) {
	for k, v := range c.nodeInfo {
		handler(k, v.NodeInfo)
	}
}
