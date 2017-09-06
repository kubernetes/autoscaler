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

	apiv1 "k8s.io/api/core/v1"
	extensionsv1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	kube_record "k8s.io/client-go/tools/record"
	"k8s.io/kubernetes/pkg/api/testapi"
	kubeletapis "k8s.io/kubernetes/pkg/kubelet/apis"

	"github.com/stretchr/testify/assert"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"
)

func TestPodSchedulableMap(t *testing.T) {
	rc1 := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc1",
			Namespace: "default",
			SelfLink:  testapi.Default.SelfLink("replicationcontrollers", "rc"),
			UID:       "12345678-1234-1234-1234-123456789012",
		},
	}

	rc2 := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc2",
			Namespace: "default",
			SelfLink:  testapi.Default.SelfLink("replicationcontrollers", "rc"),
			UID:       "12345678-1234-1234-1234-12345678901a",
		},
	}

	pMap := make(podSchedulableMap)

	podInRc1_1 := BuildTestPod("podInRc1_1", 500, 1000)
	podInRc1_1.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)

	podInRc2 := BuildTestPod("podInRc2", 500, 1000)
	podInRc2.OwnerReferences = GenerateOwnerReferences(rc2.Name, "ReplicationController", "extensions/v1beta1", rc2.UID)

	// Basic sanity checks
	_, found := pMap.get(podInRc1_1)
	assert.False(t, found)
	pMap.set(podInRc1_1, true)
	sched, found := pMap.get(podInRc1_1)
	assert.True(t, found)
	assert.True(t, sched)

	// Pod in different RC
	_, found = pMap.get(podInRc2)
	assert.False(t, found)
	pMap.set(podInRc2, false)
	sched, found = pMap.get(podInRc2)
	assert.True(t, found)
	assert.False(t, sched)

	// Another replica in rc1
	podInRc1_2 := BuildTestPod("podInRc1_1", 500, 1000)
	podInRc1_2.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)
	sched, found = pMap.get(podInRc1_2)
	assert.True(t, found)
	assert.True(t, sched)

	// A pod in rc1, but with different requests
	differentPodInRc1 := BuildTestPod("differentPodInRc1", 1000, 1000)
	differentPodInRc1.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)
	_, found = pMap.get(differentPodInRc1)
	assert.False(t, found)
	pMap.set(differentPodInRc1, false)
	sched, found = pMap.get(differentPodInRc1)
	assert.True(t, found)
	assert.False(t, sched)

	// A non-repliated pod
	nonReplicatedPod := BuildTestPod("nonReplicatedPod", 1000, 1000)
	_, found = pMap.get(nonReplicatedPod)
	assert.False(t, found)
	pMap.set(nonReplicatedPod, false)
	_, found = pMap.get(nonReplicatedPod)
	assert.False(t, found)

	// Verify information about first pod has not been overwritten by adding
	// other pods
	sched, found = pMap.get(podInRc1_1)
	assert.True(t, found)
	assert.True(t, sched)
}

func TestFilterOutSchedulable(t *testing.T) {
	rc1 := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc1",
			Namespace: "default",
			SelfLink:  testapi.Default.SelfLink("replicationcontrollers", "rc"),
			UID:       "12345678-1234-1234-1234-123456789012",
		},
	}

	rc2 := apiv1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rc2",
			Namespace: "default",
			SelfLink:  testapi.Default.SelfLink("replicationcontrollers", "rc"),
			UID:       "12345678-1234-1234-1234-12345678901a",
		},
	}

	p1 := BuildTestPod("p1", 1500, 200000)
	p2_1 := BuildTestPod("p2_2", 3000, 200000)
	p2_1.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)
	p2_2 := BuildTestPod("p2_2", 3000, 200000)
	p2_2.OwnerReferences = GenerateOwnerReferences(rc1.Name, "ReplicationController", "extensions/v1beta1", rc1.UID)
	p3_1 := BuildTestPod("p3", 100, 200000)
	p3_1.OwnerReferences = GenerateOwnerReferences(rc2.Name, "ReplicationController", "extensions/v1beta1", rc2.UID)
	p3_2 := BuildTestPod("p3", 100, 200000)
	p3_2.OwnerReferences = GenerateOwnerReferences(rc2.Name, "ReplicationController", "extensions/v1beta1", rc2.UID)
	unschedulablePods := []*apiv1.Pod{p1, p2_1, p2_2, p3_1, p3_2}

	scheduledPod1 := BuildTestPod("s1", 100, 200000)
	scheduledPod2 := BuildTestPod("s2", 1500, 200000)
	scheduledPod1.Spec.NodeName = "node1"
	scheduledPod2.Spec.NodeName = "node1"

	node := BuildTestNode("node1", 2000, 2000000)
	SetNodeReadyState(node, true, time.Time{})

	predicateChecker := simulator.NewTestPredicateChecker()

	res := FilterOutSchedulable(unschedulablePods, []*apiv1.Node{node}, []*apiv1.Pod{scheduledPod1, p1, p2_1, p2_2, p3_1, p3_2}, predicateChecker)
	assert.Equal(t, 2, len(res))
	assert.Equal(t, p2_1, res[0])
	assert.Equal(t, p2_2, res[1])

	res2 := FilterOutSchedulable(unschedulablePods, []*apiv1.Node{node}, []*apiv1.Pod{scheduledPod1, scheduledPod2, p1, p2_1, p2_2, p3_1, p3_2}, predicateChecker)
	assert.Equal(t, 3, len(res2))
	assert.Equal(t, p1, res2[0])
	assert.Equal(t, p2_1, res2[1])
	assert.Equal(t, p2_2, res2[2])
}

func TestGetNodeInfosForGroups(t *testing.T) {
	n1 := BuildTestNode("n1", 100, 1000)
	SetNodeReadyState(n1, true, time.Now())
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Now())
	n3 := BuildTestNode("n3", 1000, 1000)
	SetNodeReadyState(n3, false, time.Now())
	n4 := BuildTestNode("n4", 1000, 1000)
	SetNodeReadyState(n4, false, time.Now())

	p1 := BuildTestPod("p1", 80, 0)
	p2 := BuildTestPod("p2", 800, 0)
	p3 := BuildTestPod("p3", 800, 0)
	p1.Spec.NodeName = "n1"
	p2.Spec.NodeName = "n2"
	p3.Spec.NodeName = "n4"

	tn := BuildTestNode("T1-abc", 4000, 1000000)
	tni := schedulercache.NewNodeInfo()
	tni.SetNode(tn)

	// Cloud provider
	provider1 := testprovider.NewTestAutoprovisioningCloudProvider(nil, nil,
		nil, nil,
		nil, map[string]*schedulercache.NodeInfo{"n3": tni, "n4": tni})
	provider1.AddNodeGroup("n1", 1, 10, 1)   // Nodegroup with ready node.
	provider1.AddNodeGroup("n2", 1, 10, 1)   // Nodegroup with ready and unready node.
	provider1.AddNodeGroup("n3", 1, 10, 1)   // Nodegroup with unready node.
	provider1.AddNodeGroup("n4", 0, 1000, 0) // Nodegroup without nodes.
	provider1.AddNode("n1", n1)
	provider1.AddNode("n2", n2)
	provider1.AddNode("n2", n3)
	provider1.AddNode("n3", n4)

	provider2 := testprovider.NewTestAutoprovisioningCloudProvider(nil, nil,
		nil, nil,
		nil, nil)
	provider2.AddNodeGroup("n5", 1, 10, 1) // Nodegroup without nodes.

	fakeClient := &fake.Clientset{}
	fakeClient.Fake.AddReactor("list", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, &apiv1.PodList{Items: []apiv1.Pod{}}, nil
	})

	predicateChecker := simulator.NewTestPredicateChecker()

	res, err := GetNodeInfosForGroups([]*apiv1.Node{n1, n2, n3, n4}, provider1, fakeClient,
		[]*extensionsv1.DaemonSet{}, predicateChecker)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(res))
	_, found := res["n1"]
	assert.True(t, found)
	_, found = res["n2"]
	assert.True(t, found)
	_, found = res["n3"]
	assert.True(t, found)
	_, found = res["n4"]
	assert.True(t, found)

	// Test for a nodegroup without nodes and TempleteNodeInfo not implemented by cloud proivder
	res, err = GetNodeInfosForGroups([]*apiv1.Node{}, provider2, fakeClient,
		[]*extensionsv1.DaemonSet{}, predicateChecker)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(res))
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

func TestSanitizeNodeInfo(t *testing.T) {
	pod := BuildTestPod("p1", 80, 0)
	pod.Spec.NodeName = "n1"

	node := BuildTestNode("node", 1000, 1000)

	nodeInfo := schedulercache.NewNodeInfo(pod)
	nodeInfo.SetNode(node)

	res, err := sanitizeNodeInfo(nodeInfo, "test-group")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res.Pods()))
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

func TestGetPotentiallyUnneededNodes(t *testing.T) {
	ng1_1 := BuildTestNode("ng1-1", 1000, 1000)
	ng1_2 := BuildTestNode("ng1-2", 1000, 1000)
	ng2_1 := BuildTestNode("ng2-1", 1000, 1000)
	noNg := BuildTestNode("no-ng", 1000, 1000)
	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNodeGroup("ng2", 1, 10, 1)
	provider.AddNode("ng1", ng1_1)
	provider.AddNode("ng1", ng1_2)
	provider.AddNode("ng2", ng2_1)

	context := &AutoscalingContext{
		CloudProvider: provider,
	}

	result := getPotentiallyUnneededNodes(context, []*apiv1.Node{ng1_1, ng1_2, ng2_1, noNg})
	assert.Equal(t, 2, len(result))
	ok1 := result[0].Name == "ng1-1" && result[1].Name == "ng1-2"
	ok2 := result[1].Name == "ng1-1" && result[0].Name == "ng1-2"
	assert.True(t, ok1 || ok2)
}

func TestConfigurePredicateCheckerForLoop(t *testing.T) {
	p1 := BuildTestPod("p1", 500, 1000)
	p1.Spec.Affinity = &apiv1.Affinity{}
	p2 := BuildTestPod("p2", 500, 1000)
	p3 := BuildTestPod("p3", 500, 1000)

	predicateChecker := simulator.NewTestPredicateChecker()

	predicateChecker.SetAffinityPredicateEnabled(false)
	ConfigurePredicateCheckerForLoop([]*apiv1.Pod{p1}, []*apiv1.Pod{}, predicateChecker)
	assert.True(t, predicateChecker.IsAffinityPredicateEnabled())

	ConfigurePredicateCheckerForLoop([]*apiv1.Pod{}, []*apiv1.Pod{p1}, predicateChecker)
	assert.True(t, predicateChecker.IsAffinityPredicateEnabled())

	ConfigurePredicateCheckerForLoop([]*apiv1.Pod{p2}, []*apiv1.Pod{p3}, predicateChecker)
	assert.False(t, predicateChecker.IsAffinityPredicateEnabled())
}
