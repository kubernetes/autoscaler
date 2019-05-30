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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func TestCloudProviderNodeInstancesCache(t *testing.T) {
	// Fresh entry for node group in cache.
	nodeNg1_1 := BuildTestNode("ng1-1", 1000, 1000)
	instanceNg1_1 := cloudprovider.Instance{Id: nodeNg1_1.Name}
	// Fresh entry for node group in cache - checks Invalidate function.
	nodeNg2_1 := BuildTestNode("ng2-1", 1000, 1000)
	instanceNg2_1 := cloudprovider.Instance{Id: nodeNg2_1.Name}
	nodeNg2_2 := BuildTestNode("ng2-2", 1000, 1000)
	instanceNg2_2 := cloudprovider.Instance{Id: nodeNg2_2.Name}
	// Stale entry for node group in cache - check Refresh function.
	nodeNg3_1 := BuildTestNode("ng3-1", 1000, 1000)
	instanceNg3_1 := cloudprovider.Instance{Id: nodeNg3_1.Name}
	nodeNg3_2 := BuildTestNode("ng3-2", 1000, 1000)
	instanceNg3_2 := cloudprovider.Instance{Id: nodeNg3_2.Name}
	// Removed node group.
	nodeNg4_1 := BuildTestNode("ng4-1", 1000, 1000)
	instanceNg4_1 := cloudprovider.Instance{Id: nodeNg4_1.Name}

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 1)
	provider.AddNodeGroup("ng2", 1, 10, 1)
	provider.AddNodeGroup("ng3", 1, 10, 1)
	provider.AddNode("ng1", nodeNg1_1)
	provider.AddNode("ng2", nodeNg2_2)
	provider.AddNode("ng3", nodeNg3_2)

	cache := NewCloudProviderNodeInstancesCache(provider)
	cache.cloudProviderNodeInstances["ng1"] = &cloudProviderNodeInstancesCacheEntry{
		instances:   []cloudprovider.Instance{instanceNg1_1},
		refreshTime: time.Now(),
	}
	cache.cloudProviderNodeInstances["ng2"] = &cloudProviderNodeInstancesCacheEntry{
		instances:   []cloudprovider.Instance{instanceNg2_1},
		refreshTime: time.Now(),
	}
	cache.cloudProviderNodeInstances["ng3"] = &cloudProviderNodeInstancesCacheEntry{
		instances:   []cloudprovider.Instance{instanceNg3_1},
		refreshTime: time.Now().Add(-time.Hour),
	}
	cache.cloudProviderNodeInstances["ng4"] = &cloudProviderNodeInstancesCacheEntry{
		instances:   []cloudprovider.Instance{instanceNg4_1},
		refreshTime: time.Now(),
	}

	// Fetch stale instances.
	results, err := cache.GetCloudProviderNodeInstances()
	assert.NoError(t, err)
	assert.Equal(t, map[string][]cloudprovider.Instance{"ng1": {instanceNg1_1}, "ng2": {instanceNg2_1}, "ng3": {instanceNg3_1}}, results)
	assert.Equal(t, 4, len(cache.cloudProviderNodeInstances))

	// Invalidate entry in cache.
	cache.InvalidateCacheEntry(provider.GetNodeGroup("ng2"))
	results, err = cache.GetCloudProviderNodeInstances()
	assert.NoError(t, err)
	assert.Equal(t, map[string][]cloudprovider.Instance{"ng1": {instanceNg1_1}, "ng2": {instanceNg2_2}, "ng3": {instanceNg3_1}}, results)
	assert.Equal(t, 4, len(cache.cloudProviderNodeInstances))

	// Refresh cache.
	cache.Refresh()
	results, err = cache.GetCloudProviderNodeInstances()
	assert.NoError(t, err)
	assert.Equal(t, map[string][]cloudprovider.Instance{"ng1": {instanceNg1_1}, "ng2": {instanceNg2_2}, "ng3": {instanceNg3_2}}, results)
	assert.Equal(t, 3, len(cache.cloudProviderNodeInstances))
}
