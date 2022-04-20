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
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

func newCacheEntry(data cloudprovider.NodeGroup, ts time.Time) nodeGroupCacheEntry {
	return nodeGroupCacheEntry{data: data, ts: ts}
}

func TestCache_AddNodeGroup(t *testing.T) {
	cache := NewIonosCache()
	require.Empty(t, cache.GetNodeGroups())
	cache.AddNodeGroup(&nodePool{id: "123", min: 1, max: 3})
	require.Equal(t, []cloudprovider.NodeGroup{&nodePool{id: "123", min: 1, max: 3}}, cache.GetNodeGroups())
}

func TestCache_RemoveInstanceFromCache(t *testing.T) {
	firstTime := timeNow().Add(-2*time.Minute - 1*time.Second)
	cache := NewIonosCache()
	cache.nodeGroups["2"] = newCacheEntry(&nodePool{id: "2"}, firstTime)
	cache.nodesToNodeGroups["b2"] = "2"

	require.NotNil(t, cache.GetNodeGroupForNode("b2"))
	require.True(t, cache.NodeGroupNeedsRefresh("2"))

	cache.RemoveInstanceFromCache("b2")
	require.Nil(t, cache.GetNodeGroupForNode("b2"))
	require.False(t, cache.NodeGroupNeedsRefresh("2"))
}

func TestCache_SetInstancesCacheForNodeGroup(t *testing.T) {
	cache := NewIonosCache()
	cache.AddNodeGroup(&nodePool{id: "1"})
	cache.AddNodeGroup(&nodePool{id: "2"})
	cache.nodesToNodeGroups["a3"] = "1"
	cache.nodesToNodeGroups["b1"] = "2"
	instances := []cloudprovider.Instance{newInstance("a1"), newInstance("a2")}

	require.NotNil(t, cache.GetNodeGroupForNode("a3"))
	cache.SetInstancesCacheForNodeGroup("1", instances)

	require.Nil(t, cache.GetNodeGroupForNode("a3"))
}

func TestCache_GetNodeGroupIDs(t *testing.T) {
	cache := NewIonosCache()
	require.Empty(t, cache.GetNodeGroupIds())
	cache.AddNodeGroup(&nodePool{id: "1"})
	require.Equal(t, []string{"1"}, cache.GetNodeGroupIds())
	cache.AddNodeGroup(&nodePool{id: "2"})
	require.ElementsMatch(t, []string{"1", "2"}, cache.GetNodeGroupIds())
}

func TestCache_GetNodeGroups(t *testing.T) {
	cache := NewIonosCache()
	require.Empty(t, cache.GetNodeGroups())
	cache.AddNodeGroup(&nodePool{id: "1"})
	require.Equal(t, []cloudprovider.NodeGroup{&nodePool{id: "1"}}, cache.GetNodeGroups())
	cache.AddNodeGroup(&nodePool{id: "2"})
	require.ElementsMatch(t, []*nodePool{{id: "1"}, {id: "2"}}, cache.GetNodeGroups())
}

func TestCache_GetNodeGroupForNode(t *testing.T) {
	cache := NewIonosCache()
	require.Nil(t, cache.GetNodeGroupForNode("a1"))
	cache.AddNodeGroup(&nodePool{id: "1"})
	require.Nil(t, cache.GetNodeGroupForNode("a1"))
	cache.nodesToNodeGroups["a1"] = "1"
	require.EqualValues(t, &nodePool{id: "1"}, cache.GetNodeGroupForNode("a1"))
}

func TestCache_LockUnlockNodeGroup(t *testing.T) {
	cache := NewIonosCache()
	nodePool := &nodePool{id: "1"}
	require.True(t, cache.TryLockNodeGroup(nodePool))
	require.False(t, cache.TryLockNodeGroup(nodePool))
	cache.UnlockNodeGroup(nodePool)
	require.True(t, cache.TryLockNodeGroup(nodePool))
}

func TestCache_GetSetNodeGroupSize(t *testing.T) {
	cache := NewIonosCache()

	size, found := cache.GetNodeGroupSize("1")
	require.False(t, found)
	require.Zero(t, size)

	cache.SetNodeGroupSize("2", 1)
	size, found = cache.GetNodeGroupSize("1")
	require.False(t, found)
	require.Zero(t, size)

	cache.SetNodeGroupSize("1", 2)
	size, found = cache.GetNodeGroupSize("1")
	require.True(t, found)
	require.Equal(t, 2, size)
}

func TestCache_GetSetNodeGroupTargetSize(t *testing.T) {
	cache := NewIonosCache()

	size, found := cache.GetNodeGroupTargetSize("1")
	require.False(t, found)
	require.Zero(t, size)

	cache.SetNodeGroupTargetSize("2", 1)
	size, found = cache.GetNodeGroupTargetSize("1")
	require.False(t, found)
	require.Zero(t, size)

	cache.SetNodeGroupTargetSize("1", 2)
	size, found = cache.GetNodeGroupTargetSize("1")
	require.True(t, found)
	require.Equal(t, 2, size)

	cache.InvalidateNodeGroupTargetSize("1")
	size, found = cache.GetNodeGroupTargetSize("1")
	require.False(t, found)
	require.Zero(t, size)
}

func TestCache_NodeGroupNeedsRefresh(t *testing.T) {
	fixedTime := time.Now().Round(time.Second)
	timeNow = func() time.Time { return fixedTime }
	defer func() { timeNow = time.Now }()

	cache := NewIonosCache()
	require.True(t, cache.NodeGroupNeedsRefresh("test"))

	cache.AddNodeGroup(&nodePool{id: "test"})
	require.True(t, cache.NodeGroupNeedsRefresh("test"))
	cache.SetInstancesCacheForNodeGroup("test", nil)
	require.False(t, cache.NodeGroupNeedsRefresh("test"))

	timeNow = func() time.Time { return fixedTime.Add(nodeGroupCacheEntryTimeout) }
	require.False(t, cache.NodeGroupNeedsRefresh("test"))
	timeNow = func() time.Time { return fixedTime.Add(nodeGroupCacheEntryTimeout + 1*time.Second) }
	require.True(t, cache.NodeGroupNeedsRefresh("test"))
}
