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

package core

import (
	"fmt"
	"testing"
	"time"

	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	kube_record "k8s.io/client-go/tools/record"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset/fake"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"

	"github.com/stretchr/testify/assert"
)

func TestFilterOutSchedulable(t *testing.T) {
	p1 := BuildTestPod("p1", 1500, 200000)
	p2 := BuildTestPod("p2", 3000, 200000)
	p3 := BuildTestPod("p3", 100, 200000)
	unschedulablePods := []*apiv1.Pod{p1, p2, p3}

	scheduledPod1 := BuildTestPod("s1", 100, 200000)
	scheduledPod2 := BuildTestPod("s2", 1500, 200000)
	scheduledPod1.Spec.NodeName = "node1"
	scheduledPod2.Spec.NodeName = "node1"

	node := BuildTestNode("node1", 2000, 2000000)
	SetNodeReadyState(node, true, time.Time{})

	predicateChecker := simulator.NewTestPredicateChecker()

	res := FilterOutSchedulable(unschedulablePods, []*apiv1.Node{node}, []*apiv1.Pod{scheduledPod1}, predicateChecker)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, p2, res[0])

	res2 := FilterOutSchedulable(unschedulablePods, []*apiv1.Node{node}, []*apiv1.Pod{scheduledPod1, scheduledPod2}, predicateChecker)
	assert.Equal(t, 2, len(res2))
	assert.Equal(t, p1, res2[0])
	assert.Equal(t, p2, res2[1])
}

func TestRemoveOldUnregisteredNodes(t *testing.T) {
	deletedNodes := make(chan string, 10)

	now := time.Now()

	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	ng1_1.Spec.ProviderID = "ng1-1"
	ng1_2 := BuildTestNode("ng1-2", 1000, 1000)
	ng1_2.Spec.ProviderID = "ng1-2"
	provider := testprovider.NewTestCloudProvider(nil, func(nodegroup string, node string) error {
		deletedNodes <- fmt.Sprintf("%s/%s", nodegroup, node)
		return nil
	})
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng1", ng1_2)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder)
	err := clusterState.UpdateNodes([]*apiv1.Node{ng1_1}, now.Add(-time.Hour))
	assert.NoError(t, err)

	context := &AutoscalingContext{
		AutoscalingOptions: AutoscalingOptions{
			UnregisteredNodeRemovalTime: 45 * time.Minute,
		},
		CloudProvider:        provider,
		ClusterStateRegistry: clusterState,
	}
	unregisteredNodes := clusterState.GetUnregisteredNodes()
	assert.Equal(t, 1, len(unregisteredNodes))

	// Nothing should be removed. The unregistered node is not old enough.
	removed, err := removeOldUnregisteredNodes(unregisteredNodes, context, now.Add(-50*time.Minute))
	assert.NoError(t, err)
	assert.False(t, removed)

	// ng1_2 should be removed.
	removed, err = removeOldUnregisteredNodes(unregisteredNodes, context, now)
	assert.NoError(t, err)
	assert.True(t, removed)
	deletedNode := getStringFromChan(deletedNodes)
	assert.Equal(t, "ng1/ng1-2", deletedNode)
}

func TestSanitizeLabels(t *testing.T) {
	oldNode := BuildTestNode("ng1-1", 1000, 1000)
	oldNode.Labels = map[string]string{
		kubeletapis.LabelHostname: "abc",
		"x": "y",
	}
	node, err := sanitizeTemplateNode(oldNode, "bzium")
	assert.NoError(t, err)
	assert.NotEqual(t, node.Labels[kubeletapis.LabelHostname], "abc")
	assert.Equal(t, node.Labels["x"], "y")
	assert.NotEqual(t, node.Name, oldNode.Name)
	assert.Equal(t, node.Labels[kubeletapis.LabelHostname], node.Name)
}

func TestSanitizeTaints(t *testing.T) {
	oldNode := BuildTestNode("ng1-1", 1000, 1000)
	taints := make([]apiv1.Taint, 0)
	taints = append(taints, apiv1.Taint{
		Key:    ReschedulerTaintKey,
		Value:  "test1",
		Effect: apiv1.TaintEffectNoSchedule,
	})
	taints = append(taints, apiv1.Taint{
		Key:    "test-taint",
		Value:  "test2",
		Effect: apiv1.TaintEffectNoSchedule,
	})
	oldNode.Spec.Taints = taints
	node, err := sanitizeTemplateNode(oldNode, "bzium")
	assert.NoError(t, err)
	assert.Equal(t, len(node.Spec.Taints), 1)
	assert.Equal(t, node.Spec.Taints[0].Key, "test-taint")
}

func TestRemoveFixNodeTargetSize(t *testing.T) {
	sizeChanges := make(chan string, 10)
	now := time.Now()

	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	ng1_1.Spec.ProviderID = "ng1-1"
	provider := testprovider.NewTestCloudProvider(func(nodegroup string, delta int) error {
		sizeChanges <- fmt.Sprintf("%s/%d", nodegroup, delta)
		return nil
	}, nil)
	provider.AddNodeGroup("ng1", 1, 10, 3)
	provider.AddNode("ng1", ng1_1)

	fakeClient := &fake.Clientset{}
	fakeLogRecorder, _ := utils.NewStatusMapRecorder(fakeClient, "kube-system", kube_record.NewFakeRecorder(5), false)
	clusterState := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{
		MaxTotalUnreadyPercentage: 10,
		OkTotalUnreadyCount:       1,
	}, fakeLogRecorder)
	err := clusterState.UpdateNodes([]*apiv1.Node{ng1_1}, now.Add(-time.Hour))
	assert.NoError(t, err)

	context := &AutoscalingContext{
		AutoscalingOptions: AutoscalingOptions{
			UnregisteredNodeRemovalTime: 45 * time.Minute,
		},
		CloudProvider:        provider,
		ClusterStateRegistry: clusterState,
	}

	// Nothing should be fixed. The incorrect size state is not old enough.
	removed, err := fixNodeGroupSize(context, now.Add(-50*time.Minute))
	assert.NoError(t, err)
	assert.False(t, removed)

	// Node group should be decreased.
	removed, err = fixNodeGroupSize(context, now)
	assert.NoError(t, err)
	assert.True(t, removed)
	change := getStringFromChan(sizeChanges)
	assert.Equal(t, "ng1/-2", change)
}
