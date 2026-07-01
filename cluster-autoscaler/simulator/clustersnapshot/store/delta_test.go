/*
Copyright 2024 The Kubernetes Authors.

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

package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

func buildCaches(store *DeltaSnapshotStore) {
	lister := (*deltaSnapshotStoreNodeLister)(store)
	lister.HavePodsWithAffinityList()
	lister.HavePodsWithRequiredAntiAffinityList()
}

func TestPodCacheSurvivesAddWithoutAffinity(t *testing.T) {
	nodes := clustersnapshot.CreateTestNodes(2)
	store := NewDeltaSnapshotStore(1)
	for _, node := range nodes {
		require.NoError(t, store.StoreNodeInfo(framework.NewNodeInfo(node, nil)))
	}

	buildCaches(store)
	assert.NotNil(t, store.data.havePodsWithAffinity)
	assert.NotNil(t, store.data.havePodsWithRequiredAntiAffinity)

	pod := test.BuildTestPod("plain-pod", 100, 100)
	pod.Spec.NodeName = nodes[0].Name
	require.NoError(t, store.StorePodInfo(framework.NewPodInfo(pod, nil), nodes[0].Name))

	assert.NotNil(t, store.data.havePodsWithAffinity, "cache should not be cleared when adding pod without affinity")
	assert.NotNil(t, store.data.havePodsWithRequiredAntiAffinity, "cache should not be cleared when adding pod without affinity")
	assert.Empty(t, store.data.havePodsWithAffinity)
	assert.Empty(t, store.data.havePodsWithRequiredAntiAffinity)
}

func TestPodCacheUpdatedOnAffinityPodAdd(t *testing.T) {
	nodes := clustersnapshot.CreateTestNodes(2)
	store := NewDeltaSnapshotStore(1)
	for _, node := range nodes {
		require.NoError(t, store.StoreNodeInfo(framework.NewNodeInfo(node, nil)))
	}

	buildCaches(store)
	assert.Empty(t, store.data.havePodsWithAffinity)

	pod := test.BuildTestPod("affinity-pod", 100, 100)
	pod.Spec.Affinity = &apiv1.Affinity{
		PodAffinity: &apiv1.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []apiv1.PodAffinityTerm{
				{TopologyKey: "kubernetes.io/hostname"},
			},
		},
	}
	require.NoError(t, store.StorePodInfo(framework.NewPodInfo(pod, nil), nodes[0].Name))

	assert.NotNil(t, store.data.havePodsWithAffinity)
	assert.Len(t, store.data.havePodsWithAffinity, 1)
	assert.Equal(t, nodes[0].Name, store.data.havePodsWithAffinity[0].Node().Name)
}

func TestPodCacheUpdatedOnRequiredAntiAffinityPodAdd(t *testing.T) {
	nodes := clustersnapshot.CreateTestNodes(2)
	store := NewDeltaSnapshotStore(1)
	for _, node := range nodes {
		require.NoError(t, store.StoreNodeInfo(framework.NewNodeInfo(node, nil)))
	}

	buildCaches(store)
	assert.Empty(t, store.data.havePodsWithRequiredAntiAffinity)

	pod := test.BuildTestPod("anti-affinity-pod", 100, 100, test.WithPodHostnameAntiAffinity(map[string]string{"app": "test"}))
	require.NoError(t, store.StorePodInfo(framework.NewPodInfo(pod, nil), nodes[0].Name))

	assert.NotNil(t, store.data.havePodsWithRequiredAntiAffinity)
	assert.Len(t, store.data.havePodsWithRequiredAntiAffinity, 1)
	assert.Equal(t, nodes[0].Name, store.data.havePodsWithRequiredAntiAffinity[0].Node().Name)

	// Should also be in havePodsWithAffinity (PodAntiAffinity counts)
	assert.Len(t, store.data.havePodsWithAffinity, 1)
}

func TestPodCacheUpdatedOnPodRemove(t *testing.T) {
	nodes := clustersnapshot.CreateTestNodes(2)
	store := NewDeltaSnapshotStore(1)
	for _, node := range nodes {
		require.NoError(t, store.StoreNodeInfo(framework.NewNodeInfo(node, nil)))
	}

	pod := test.BuildTestPod("anti-affinity-pod", 100, 100, test.WithPodHostnameAntiAffinity(map[string]string{"app": "test"}))
	require.NoError(t, store.StorePodInfo(framework.NewPodInfo(pod, nil), nodes[0].Name))

	buildCaches(store)
	assert.Len(t, store.data.havePodsWithAffinity, 1)
	assert.Len(t, store.data.havePodsWithRequiredAntiAffinity, 1)

	require.NoError(t, store.RemovePodInfo(pod.Namespace, pod.Name, nodes[0].Name))

	assert.NotNil(t, store.data.havePodsWithAffinity, "cache should not be nil after remove")
	assert.NotNil(t, store.data.havePodsWithRequiredAntiAffinity, "cache should not be nil after remove")
	assert.Empty(t, store.data.havePodsWithAffinity)
	assert.Empty(t, store.data.havePodsWithRequiredAntiAffinity)
}

func TestPodCacheNotBuiltPrematurely(t *testing.T) {
	nodes := clustersnapshot.CreateTestNodes(2)
	store := NewDeltaSnapshotStore(1)
	for _, node := range nodes {
		require.NoError(t, store.StoreNodeInfo(framework.NewNodeInfo(node, nil)))
	}

	// Don't build caches - they should be nil
	assert.Nil(t, store.data.havePodsWithAffinity)
	assert.Nil(t, store.data.havePodsWithRequiredAntiAffinity)

	pod := test.BuildTestPod("affinity-pod", 100, 100, test.WithPodHostnameAntiAffinity(map[string]string{"app": "test"}))
	require.NoError(t, store.StorePodInfo(framework.NewPodInfo(pod, nil), nodes[0].Name))

	// Cache should still be nil - not built prematurely
	assert.Nil(t, store.data.havePodsWithAffinity, "cache should stay nil when not previously built")
	assert.Nil(t, store.data.havePodsWithRequiredAntiAffinity, "cache should stay nil when not previously built")

	// But when we build it now, it should reflect the added pod
	buildCaches(store)
	assert.Len(t, store.data.havePodsWithAffinity, 1)
	assert.Len(t, store.data.havePodsWithRequiredAntiAffinity, 1)
}

func TestPodCacheCorrectAfterForkAndAddPod(t *testing.T) {
	nodes := clustersnapshot.CreateTestNodes(3)
	store := NewDeltaSnapshotStore(1)
	for _, node := range nodes {
		require.NoError(t, store.StoreNodeInfo(framework.NewNodeInfo(node, nil)))
	}

	// Add a pod with affinity to base
	pod1 := test.BuildTestPod("base-affinity-pod", 100, 100, test.WithPodHostnameAntiAffinity(map[string]string{"app": "test"}))
	require.NoError(t, store.StorePodInfo(framework.NewPodInfo(pod1, nil), nodes[0].Name))
	buildCaches(store)
	assert.Len(t, store.data.havePodsWithAffinity, 1)

	// Fork and add another affinity pod on a different node
	store.Fork()
	pod2 := test.BuildTestPod("fork-affinity-pod", 100, 100, test.WithPodHostnameAntiAffinity(map[string]string{"app": "test2"}))
	require.NoError(t, store.StorePodInfo(framework.NewPodInfo(pod2, nil), nodes[1].Name))

	// The forked layer's cache should reflect both nodes
	lister := (*deltaSnapshotStoreNodeLister)(store)
	affinityList, err := lister.HavePodsWithAffinityList()
	require.NoError(t, err)
	assert.Len(t, affinityList, 2)

	// Revert and check base is unchanged
	store.Revert()
	assert.Len(t, store.data.havePodsWithAffinity, 1)
	assert.Equal(t, nodes[0].Name, store.data.havePodsWithAffinity[0].Node().Name)
}

func TestPodCacheNodeNotDuplicatedOnMultipleAdds(t *testing.T) {
	nodes := clustersnapshot.CreateTestNodes(1)
	store := NewDeltaSnapshotStore(1)
	require.NoError(t, store.StoreNodeInfo(framework.NewNodeInfo(nodes[0], nil)))

	buildCaches(store)

	// Add two affinity pods to the same node
	pod1 := test.BuildTestPod("affinity-pod-1", 100, 100, test.WithPodHostnameAntiAffinity(map[string]string{"app": "a"}))
	require.NoError(t, store.StorePodInfo(framework.NewPodInfo(pod1, nil), nodes[0].Name))
	pod2 := test.BuildTestPod("affinity-pod-2", 100, 100, test.WithPodHostnameAntiAffinity(map[string]string{"app": "b"}))
	require.NoError(t, store.StorePodInfo(framework.NewPodInfo(pod2, nil), nodes[0].Name))

	assert.Len(t, store.data.havePodsWithAffinity, 1, "node should appear only once even with multiple affinity pods")
	assert.Len(t, store.data.havePodsWithRequiredAntiAffinity, 1)
}
