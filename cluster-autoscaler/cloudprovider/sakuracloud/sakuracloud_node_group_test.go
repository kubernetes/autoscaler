/*
Copyright The Kubernetes Authors.

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

package sakuracloud

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
)

func testManagerWithServers(servers []sakuraServer) *sakuracloudManager {
	return &sakuracloudManager{
		zone:       "is1a",
		servers:    servers,
		nodeGroups: map[string]*sakuracloudNodeGroup{},
	}
}

func TestIncreaseSizeValidation(t *testing.T) {
	m := testManagerWithServers(nil)
	ng := &sakuracloudNodeGroup{id: "pool", manager: m, config: testGroupConfig(), targetSize: 2}

	assert.Error(t, ng.IncreaseSize(0))
	assert.Error(t, ng.IncreaseSize(-1))
	// 2 + 1 > MaxSize(2)
	assert.Error(t, ng.IncreaseSize(1))
}

func TestDecreaseTargetSize(t *testing.T) {
	m := testManagerWithServers([]sakuraServer{
		{ID: "1", Name: "pool-a", Tags: []string{groupTagPrefix + "pool"}},
	})
	ng := &sakuracloudNodeGroup{id: "pool", manager: m, config: testGroupConfig(), targetSize: 2}

	assert.Error(t, ng.DecreaseTargetSize(1))
	// target 2 -> 0 would drop below the 1 live server.
	assert.Error(t, ng.DecreaseTargetSize(-2))
	assert.NoError(t, ng.DecreaseTargetSize(-1))
	size, err := ng.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 1, size)
}

func TestNodes(t *testing.T) {
	up := "up"
	m := testManagerWithServers([]sakuraServer{
		{ID: "1", Name: "pool-a", Tags: []string{groupTagPrefix + "pool"}, Instance: &struct {
			Status string `json:"Status"`
		}{Status: up}},
		{ID: "2", Name: "pool-b", Tags: []string{groupTagPrefix + "pool"}},
		{ID: "3", Name: "other-a", Tags: []string{groupTagPrefix + "other"}},
	})
	ng := &sakuracloudNodeGroup{id: "pool", manager: m, config: testGroupConfig()}

	instances, err := ng.Nodes()
	assert.NoError(t, err)
	assert.Len(t, instances, 2)
	assert.Equal(t, "sakuracloud://is1a/pool-a", instances[0].Id)
}

func TestTemplateNodeInfo(t *testing.T) {
	cfg := testGroupConfig()
	cfg.Labels = map[string]string{"cloud": "sakura", "role": "sakura-spot"}
	cfg.Taints = []apiv1.Taint{{Key: "dedicated", Value: "sakura-ops", Effect: apiv1.TaintEffectNoSchedule}}
	m := testManagerWithServers(nil)
	ng := &sakuracloudNodeGroup{id: "pool", manager: m, config: cfg}

	info, err := ng.TemplateNodeInfo()
	assert.NoError(t, err)
	node := info.Node()
	assert.Equal(t, "sakura", node.Labels["cloud"])
	assert.Equal(t, "sakura-spot", node.Labels["role"])
	assert.Len(t, node.Spec.Taints, 1)
	assert.Equal(t, "dedicated", node.Spec.Taints[0].Key)
	cpu := node.Status.Capacity[apiv1.ResourceCPU]
	assert.Equal(t, int64(2), cpu.Value())
	mem := node.Status.Capacity[apiv1.ResourceMemory]
	assert.Equal(t, int64(4*1024*1024*1024), mem.Value())
}

func TestSyncTargetSize(t *testing.T) {
	m := testManagerWithServers([]sakuraServer{
		{ID: "1", Name: "pool-a", Tags: []string{groupTagPrefix + "pool"}},
	})
	ng := &sakuracloudNodeGroup{id: "pool", manager: m, config: testGroupConfig(), targetSize: 0}

	// With nothing provisioning, the live count is the truth.
	ng.syncTargetSize()
	size, _ := ng.TargetSize()
	assert.Equal(t, 1, size)

	// While provisioning, do not reconcile downwards.
	ng.provisioning = 1
	ng.targetSize = 2
	ng.syncTargetSize()
	size, _ = ng.TargetSize()
	assert.Equal(t, 2, size)
}
