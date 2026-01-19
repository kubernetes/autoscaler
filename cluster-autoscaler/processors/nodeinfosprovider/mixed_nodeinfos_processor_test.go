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

package nodeinfosprovider

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	caerror "k8s.io/autoscaler/cluster-autoscaler/utils/errors"

	apiv1 "k8s.io/api/core/v1"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	ca_context "k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/testsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

var (
	cacheTtl = 1 * time.Second
)

func TestMixedNodeInfosProvider(t *testing.T) {
	now := time.Now()
	cacheTtl = time.Minute

	type testNode struct {
		name           string
		milliCpu       int64
		mem            int64
		ready          bool
		lastTransition time.Time
		toBeDeleted    bool
	}

	type testNodeGroup struct {
		id       string
		nodes    []testNode
		template *testNode
	}

	testCases := []struct {
		name          string
		groups        []testNodeGroup
		errorNodes    []testNode
		withCache     map[string]cacheItem
		wantNodeInfos map[string]testNode
		wantCache     map[string]cacheItem
		wantError     caerror.AutoscalerError
	}{
		{
			name:   "Nodegroup without nodes and templates",
			groups: []testNodeGroup{},
		},
		{
			name: "Nodegroup with ready node",
			groups: []testNodeGroup{
				{
					id: "ng",
					nodes: []testNode{
						{name: "n", milliCpu: 1000, mem: 20, ready: true, lastTransition: now.Add(-2 * time.Minute)},
					},
				},
			},
			wantNodeInfos: map[string]testNode{
				"ng": {milliCpu: 1000, mem: 20},
			},
			wantCache: map[string]cacheItem{
				"ng": {NodeInfo: framework.NewTestNodeInfo(BuildTestNode("n", 1000, 20)), added: now},
			},
		},
		{
			name: "Nodegroup with template",
			groups: []testNodeGroup{
				{
					id:       "ng",
					template: &testNode{milliCpu: 1000, mem: 20},
				},
			},
			wantNodeInfos: map[string]testNode{
				"ng": {milliCpu: 1000, mem: 20},
			},
		},
		{
			name: "Nodegroup with unready node",
			groups: []testNodeGroup{
				{
					id: "ng",
					nodes: []testNode{
						{name: "n", milliCpu: 1000, mem: 20, ready: false},
					},
				},
			},
			wantNodeInfos: map[string]testNode{
				"ng": {milliCpu: 1000, mem: 20},
			},
		},
		{
			name: "Ready node is picked over unready node",
			groups: []testNodeGroup{
				{
					id: "ng",
					nodes: []testNode{
						{name: "unready", milliCpu: 1000, mem: 20, ready: false},
						{name: "ready", milliCpu: 2000, mem: 40, ready: true, lastTransition: now.Add(-2 * time.Minute)},
					},
				},
			},
			wantNodeInfos: map[string]testNode{
				"ng": {milliCpu: 2000, mem: 40},
			},
			wantCache: map[string]cacheItem{
				"ng": {NodeInfo: framework.NewTestNodeInfo(BuildTestNode("n", 2000, 40)), added: now},
			},
		},
		{
			name: "Ready node is picked over template",
			groups: []testNodeGroup{
				{
					id: "ng",
					nodes: []testNode{
						{name: "ready", milliCpu: 2000, mem: 40, ready: true, lastTransition: now.Add(-2 * time.Minute)},
					},
					template: &testNode{milliCpu: 1000, mem: 20},
				},
			},
			wantNodeInfos: map[string]testNode{
				"ng": {milliCpu: 2000, mem: 40},
			},
			wantCache: map[string]cacheItem{
				"ng": {NodeInfo: framework.NewTestNodeInfo(BuildTestNode("n", 2000, 40)), added: now},
			},
		},
		{
			name: "Template is picked over unready node",
			groups: []testNodeGroup{
				{
					id: "ng",
					nodes: []testNode{
						{name: "unready", milliCpu: 1000, mem: 20, ready: false},
					},
					template: &testNode{milliCpu: 2000, mem: 40},
				},
			},
			wantNodeInfos: map[string]testNode{
				"ng": {milliCpu: 2000, mem: 40},
			},
		},
		{
			name: "Template is picked over to be deleted node",
			groups: []testNodeGroup{
				{
					id: "ng",
					nodes: []testNode{
						{name: "to-be-deleted", milliCpu: 1000, mem: 20, ready: true, lastTransition: now.Add(-2 * time.Minute), toBeDeleted: true},
					},
					template: &testNode{milliCpu: 2000, mem: 40},
				},
			},
			wantNodeInfos: map[string]testNode{
				"ng": {milliCpu: 2000, mem: 40},
			},
		},
		{
			name: "Template is picked over just created node",
			groups: []testNodeGroup{
				{
					id: "ng",
					nodes: []testNode{
						{name: "just-ready", milliCpu: 1000, mem: 20, ready: true, lastTransition: now},
					},
					template: &testNode{milliCpu: 2000, mem: 40},
				},
			},
			wantNodeInfos: map[string]testNode{
				"ng": {milliCpu: 2000, mem: 40},
			},
		},
		{
			name: "Fresh cache used",
			groups: []testNodeGroup{
				{
					id: "ng",
				},
			},
			withCache: map[string]cacheItem{
				"ng": {NodeInfo: framework.NewTestNodeInfo(BuildTestNode("n", 2000, 40)), added: now.Add(-30 * time.Second)},
			},
			wantNodeInfos: map[string]testNode{
				"ng": {milliCpu: 2000, mem: 40},
			},
			wantCache: map[string]cacheItem{
				"ng": {NodeInfo: framework.NewTestNodeInfo(BuildTestNode("n", 2000, 40)), added: now.Add(-30 * time.Second)},
			},
		},
		{
			name: "Old cache not used and cleared",
			groups: []testNodeGroup{
				{
					id: "ng",
				},
			},
			withCache: map[string]cacheItem{
				"ng": {NodeInfo: framework.NewTestNodeInfo(BuildTestNode("n", 2000, 40)), added: now.Add(-2 * time.Minute)},
			},
			wantNodeInfos: map[string]testNode{},
			wantCache:     map[string]cacheItem{},
		},
		{
			name: "Ready node picked over cache, updates cache",
			groups: []testNodeGroup{
				{
					id: "ng",
					nodes: []testNode{
						{name: "ready", milliCpu: 4000, mem: 80, ready: true, lastTransition: now.Add(-2 * time.Minute)},
					},
				},
			},
			withCache: map[string]cacheItem{
				"ng": {NodeInfo: framework.NewTestNodeInfo(BuildTestNode("n", 2000, 40)), added: now.Add(-30 * time.Second)},
			},
			wantNodeInfos: map[string]testNode{
				"ng": {milliCpu: 4000, mem: 80},
			},
			wantCache: map[string]cacheItem{
				"ng": {NodeInfo: framework.NewTestNodeInfo(BuildTestNode("n", 4000, 80)), added: now},
			},
		},
		{
			name: "Cache picked over templates and unready nodes",
			groups: []testNodeGroup{
				{
					id: "ng",
					nodes: []testNode{
						{name: "unready", milliCpu: 4000, mem: 80, ready: false},
					},
					template: &testNode{milliCpu: 3000, mem: 60},
				},
			},
			withCache: map[string]cacheItem{
				"ng": {NodeInfo: framework.NewTestNodeInfo(BuildTestNode("n", 2000, 40)), added: now.Add(-30 * time.Second)},
			},
			wantNodeInfos: map[string]testNode{
				"ng": {milliCpu: 2000, mem: 40},
			},
			wantCache: map[string]cacheItem{
				"ng": {NodeInfo: framework.NewTestNodeInfo(BuildTestNode("n", 2000, 40)), added: now.Add(-30 * time.Second)},
			},
		},
		{
			name: "Nodes triggering error in CloudProvider are skipped",
			groups: []testNodeGroup{
				{
					id: "ng",
					nodes: []testNode{
						{name: "ready", milliCpu: 2000, mem: 40, ready: true, lastTransition: now.Add(-2 * time.Minute)},
					},
				},
			},
			errorNodes: []testNode{
				{name: "error-node-ready", milliCpu: 4000, mem: 80, ready: true, lastTransition: now.Add(-2 * time.Minute)},
				{name: "error-node-unready", milliCpu: 4000, mem: 80, ready: false},
			},
			wantNodeInfos: map[string]testNode{
				"ng": {milliCpu: 2000, mem: 40},
			},
			wantCache: map[string]cacheItem{
				"ng": {NodeInfo: framework.NewTestNodeInfo(BuildTestNode("n", 2000, 40)), added: now},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			machineTemplates := make(map[string]*framework.NodeInfo)
			for _, tg := range tc.groups {
				if tg.template == nil {
					continue
				}

				mt := *tg.template
				nodeInfo := framework.NewTestNodeInfo(BuildTestNode(mt.name, mt.milliCpu, mt.mem))
				machineTemplates[tg.id] = nodeInfo
			}

			var allNodes []*apiv1.Node
			var errNodes []string
			for _, tn := range tc.errorNodes {
				node := BuildTestNode(tn.name, tn.milliCpu, tn.mem)
				errNodes = append(errNodes, node.Name)
				allNodes = append(allNodes, node)
			}

			provider := testprovider.NewTestCloudProviderBuilder().
				WithMachineTemplates(machineTemplates).
				WithNodeProcessingError(errNodes).Build()

			for _, tg := range tc.groups {
				provider.AddNodeGroup(tg.id, 0, 0, 0)
				for _, tn := range tg.nodes {
					node := BuildTestNode(tn.name, tn.milliCpu, tn.mem)
					SetNodeReadyState(node, tn.ready, tn.lastTransition)
					if tn.toBeDeleted {
						setToBeDeletedTaint(node)
					}
					provider.AddNode(tg.id, node)
					allNodes = append(allNodes, node)
				}
			}

			snapshot := testsnapshot.NewTestSnapshotOrDie(t)
			err := snapshot.SetClusterState(allNodes, nil, nil, nil)
			assert.NoError(t, err)

			podLister := kube_util.NewTestPodLister([]*apiv1.Pod{})
			registry := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil)

			ctx := &ca_context.AutoscalingContext{
				CloudProvider:   provider,
				ClusterSnapshot: snapshot,
				AutoscalingKubeClients: ca_context.AutoscalingKubeClients{
					ListerRegistry: registry,
				},
			}

			processor := NewMixedTemplateNodeInfoProvider(&cacheTtl, false)
			if tc.withCache != nil {
				processor.nodeInfoCache = tc.withCache
			}
			nodeInfos, err := processor.Process(ctx, allNodes, nil, taints.TaintConfig{}, now)

			assert.Equal(t, tc.wantError, err)
			if tc.wantError == nil {
				assert.Equal(t, len(tc.wantNodeInfos), len(nodeInfos))
				for id, tn := range tc.wantNodeInfos {
					want := BuildTestNode(tn.name, tn.milliCpu, tn.mem)

					info, found := nodeInfos[id]
					assert.True(t, found)
					assertEqualNodeCapacities(t, want, info.Node())
				}

				assert.Equal(t, len(tc.wantCache), len(processor.nodeInfoCache))
				for id, want := range tc.wantCache {
					cached, found := processor.nodeInfoCache[id]
					assert.True(t, found)
					assert.Equal(t, want.added, cached.added)
					assertEqualNodeCapacities(t, want.Node(), cached.Node())
				}
			}
		})
	}
}

func assertEqualNodeCapacities(t *testing.T, expected, actual *apiv1.Node) {
	t.Helper()
	assert.NotEqual(t, actual.Status, nil, "")
	assert.Equal(t, getNodeResource(expected, apiv1.ResourceCPU), getNodeResource(actual, apiv1.ResourceCPU), "CPU should be the same")
	assert.Equal(t, getNodeResource(expected, apiv1.ResourceMemory), getNodeResource(actual, apiv1.ResourceMemory), "Memory should be the same")
}

func getNodeResource(node *apiv1.Node, resource apiv1.ResourceName) int64 {
	nodeCapacity, found := node.Status.Capacity[resource]
	if !found {
		return 0
	}

	nodeCapacityValue := nodeCapacity.Value()
	if nodeCapacityValue < 0 {
		nodeCapacityValue = 0
	}

	return nodeCapacityValue
}

func setToBeDeletedTaint(node *apiv1.Node) {
	node.Spec.Taints = append(node.Spec.Taints, apiv1.Taint{
		Key:    taints.ToBeDeletedTaint,
		Value:  fmt.Sprint(time.Now().Unix()),
		Effect: apiv1.TaintEffectNoSchedule,
	})
}
