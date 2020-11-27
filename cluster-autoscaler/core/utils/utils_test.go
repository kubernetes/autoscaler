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

	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

func TestGetNodeInfosForGroups(t *testing.T) {
	ready1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(ready1, true, time.Now())
	ready2 := BuildTestNode("n2", 2000, 2000)
	SetNodeReadyState(ready2, true, time.Now())
	unready3 := BuildTestNode("n3", 3000, 3000)
	SetNodeReadyState(unready3, false, time.Now())
	unready4 := BuildTestNode("n4", 4000, 4000)
	SetNodeReadyState(unready4, false, time.Now())

	tn := BuildTestNode("tn", 5000, 5000)
	tni := schedulerframework.NewNodeInfo()
	tni.SetNode(tn)

	// Cloud provider with TemplateNodeInfo implemented.
	provider1 := testprovider.NewTestAutoprovisioningCloudProvider(
		nil, nil, nil, nil, nil,
		map[string]*schedulerframework.NodeInfo{"ng3": tni, "ng4": tni})
	provider1.AddNodeGroup("ng1", 1, 10, 1) // Nodegroup with ready node.
	provider1.AddNode("ng1", ready1)
	provider1.AddNodeGroup("ng2", 1, 10, 1) // Nodegroup with ready and unready node.
	provider1.AddNode("ng2", ready2)
	provider1.AddNode("ng2", unready3)
	provider1.AddNodeGroup("ng3", 1, 10, 1) // Nodegroup with unready node.
	provider1.AddNode("ng3", unready4)
	provider1.AddNodeGroup("ng4", 0, 1000, 0) // Nodegroup without nodes.

	// Cloud provider with TemplateNodeInfo not implemented.
	provider2 := testprovider.NewTestAutoprovisioningCloudProvider(nil, nil, nil, nil, nil, nil)
	provider2.AddNodeGroup("ng5", 1, 10, 1) // Nodegroup without nodes.

	podLister := kube_util.NewTestPodLister([]*apiv1.Pod{})
	registry := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil, nil)

	predicateChecker, err := simulator.NewTestPredicateChecker()
	assert.NoError(t, err)

	res, err := GetNodeInfosForGroups([]*apiv1.Node{unready4, unready3, ready2, ready1}, nil,
		provider1, registry, []*appsv1.DaemonSet{}, predicateChecker, nil)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(res))
	info, found := res["ng1"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, ready1, info.Node())
	info, found = res["ng2"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, ready2, info.Node())
	info, found = res["ng3"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, tn, info.Node())
	info, found = res["ng4"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, tn, info.Node())

	// Test for a nodegroup without nodes and TemplateNodeInfo not implemented by cloud proivder
	res, err = GetNodeInfosForGroups([]*apiv1.Node{}, nil, provider2, registry,
		[]*appsv1.DaemonSet{}, predicateChecker, nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(res))
}

func TestGetNodeInfosForGroupsCache(t *testing.T) {
	ready1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(ready1, true, time.Now())
	ready2 := BuildTestNode("n2", 2000, 2000)
	SetNodeReadyState(ready2, true, time.Now())
	unready3 := BuildTestNode("n3", 3000, 3000)
	SetNodeReadyState(unready3, false, time.Now())
	unready4 := BuildTestNode("n4", 4000, 4000)
	SetNodeReadyState(unready4, false, time.Now())
	ready5 := BuildTestNode("n5", 5000, 5000)
	SetNodeReadyState(ready5, true, time.Now())
	ready6 := BuildTestNode("n6", 6000, 6000)
	SetNodeReadyState(ready6, true, time.Now())

	tn := BuildTestNode("tn", 10000, 10000)
	tni := schedulerframework.NewNodeInfo()
	tni.SetNode(tn)

	lastDeletedGroup := ""
	onDeleteGroup := func(id string) error {
		lastDeletedGroup = id
		return nil
	}

	// Cloud provider with TemplateNodeInfo implemented.
	provider1 := testprovider.NewTestAutoprovisioningCloudProvider(
		nil, nil, nil, onDeleteGroup, nil,
		map[string]*schedulerframework.NodeInfo{"ng3": tni, "ng4": tni})
	provider1.AddNodeGroup("ng1", 1, 10, 1) // Nodegroup with ready node.
	provider1.AddNode("ng1", ready1)
	provider1.AddNodeGroup("ng2", 1, 10, 1) // Nodegroup with ready and unready node.
	provider1.AddNode("ng2", ready2)
	provider1.AddNode("ng2", unready3)
	provider1.AddNodeGroup("ng3", 1, 10, 1) // Nodegroup with unready node (and 1 previously ready node).
	provider1.AddNode("ng3", unready4)
	provider1.AddNode("ng3", ready5)
	provider1.AddNodeGroup("ng4", 0, 1000, 0) // Nodegroup without nodes (and 1 previously ready node).
	provider1.AddNode("ng4", ready6)

	podLister := kube_util.NewTestPodLister([]*apiv1.Pod{})
	registry := kube_util.NewListerRegistry(nil, nil, podLister, nil, nil, nil, nil, nil, nil, nil)

	predicateChecker, err := simulator.NewTestPredicateChecker()
	assert.NoError(t, err)

	nodeInfoCache := make(map[string]*schedulerframework.NodeInfo)

	// Fill cache
	res, err := GetNodeInfosForGroups([]*apiv1.Node{unready4, unready3, ready2, ready1}, nodeInfoCache,
		provider1, registry, []*appsv1.DaemonSet{}, predicateChecker, nil)
	assert.NoError(t, err)
	// Check results
	assert.Equal(t, 4, len(res))
	info, found := res["ng1"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, ready1, info.Node())
	info, found = res["ng2"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, ready2, info.Node())
	info, found = res["ng3"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, tn, info.Node())
	info, found = res["ng4"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, tn, info.Node())
	// Check cache
	cachedInfo, found := nodeInfoCache["ng1"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, ready1, cachedInfo.Node())
	cachedInfo, found = nodeInfoCache["ng2"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, ready2, cachedInfo.Node())
	cachedInfo, found = nodeInfoCache["ng3"]
	assert.False(t, found)
	cachedInfo, found = nodeInfoCache["ng4"]
	assert.False(t, found)

	// Invalidate part of cache in two different ways
	provider1.DeleteNodeGroup("ng1")
	provider1.GetNodeGroup("ng3").Delete()
	assert.Equal(t, "ng3", lastDeletedGroup)

	// Check cache with all nodes removed
	res, err = GetNodeInfosForGroups([]*apiv1.Node{}, nodeInfoCache,
		provider1, registry, []*appsv1.DaemonSet{}, predicateChecker, nil)
	assert.NoError(t, err)
	// Check results
	assert.Equal(t, 2, len(res))
	info, found = res["ng2"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, ready2, info.Node())
	// Check ng4 result and cache
	info, found = res["ng4"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, tn, info.Node())
	// Check cache
	cachedInfo, found = nodeInfoCache["ng2"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, ready2, cachedInfo.Node())
	cachedInfo, found = nodeInfoCache["ng4"]
	assert.False(t, found)

	// Fill cache manually
	infoNg4Node6 := schedulerframework.NewNodeInfo()
	err2 := infoNg4Node6.SetNode(ready6.DeepCopy())
	assert.NoError(t, err2)
	nodeInfoCache = map[string]*schedulerframework.NodeInfo{"ng4": infoNg4Node6}
	// Check if cache was used
	res, err = GetNodeInfosForGroups([]*apiv1.Node{ready1, ready2}, nodeInfoCache,
		provider1, registry, []*appsv1.DaemonSet{}, predicateChecker, nil)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(res))
	info, found = res["ng2"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, ready2, info.Node())
	info, found = res["ng4"]
	assert.True(t, found)
	assertEqualNodeCapacities(t, ready6, info.Node())
}

func assertEqualNodeCapacities(t *testing.T, expected, actual *apiv1.Node) {
	t.Helper()
	assert.Equal(t, getNodeResource(expected, apiv1.ResourceCPU), getNodeResource(actual, apiv1.ResourceCPU), "CPU should be the same")
	assert.Equal(t, getNodeResource(expected, apiv1.ResourceMemory), getNodeResource(actual, apiv1.ResourceMemory), "Memory should be the same")
}

func TestSanitizeNodeInfo(t *testing.T) {
	pod := BuildTestPod("p1", 80, 0)
	pod.Spec.NodeName = "n1"

	node := BuildTestNode("node", 1000, 1000)

	nodeInfo := schedulerframework.NewNodeInfo(pod)
	nodeInfo.SetNode(node)

	res, err := sanitizeNodeInfo(nodeInfo, "test-group", nil)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(res.Pods))
}

func TestSanitizeLabels(t *testing.T) {
	oldNode := BuildTestNode("ng1-1", 1000, 1000)
	oldNode.Labels = map[string]string{
		apiv1.LabelHostname: "abc",
		"x":                 "y",
	}
	node, err := sanitizeTemplateNode(oldNode, "bzium", nil)
	assert.NoError(t, err)
	assert.NotEqual(t, node.Labels[apiv1.LabelHostname], "abc", nil)
	assert.Equal(t, node.Labels["x"], "y")
	assert.NotEqual(t, node.Name, oldNode.Name)
	assert.Equal(t, node.Labels[apiv1.LabelHostname], node.Name)
}

func TestGetNodeResource(t *testing.T) {
	node := BuildTestNode("n1", 1000, 2*MiB)

	cores := getNodeResource(node, apiv1.ResourceCPU)
	assert.Equal(t, int64(1), cores)

	memory := getNodeResource(node, apiv1.ResourceMemory)
	assert.Equal(t, int64(2*MiB), memory)

	unknownResourceValue := getNodeResource(node, "unknown resource")
	assert.Equal(t, int64(0), unknownResourceValue)

	// if we have no resources in capacity we expect getNodeResource to return 0
	nodeWithMissingCapacity := BuildTestNode("n1", 1000, 2*MiB)
	nodeWithMissingCapacity.Status.Capacity = apiv1.ResourceList{}

	cores = getNodeResource(nodeWithMissingCapacity, apiv1.ResourceCPU)
	assert.Equal(t, int64(0), cores)

	memory = getNodeResource(nodeWithMissingCapacity, apiv1.ResourceMemory)
	assert.Equal(t, int64(0), memory)

	// if we have negative values in resources we expect getNodeResource to return 0
	nodeWithNegativeCapacity := BuildTestNode("n1", -1000, -2*MiB)
	nodeWithNegativeCapacity.Status.Capacity = apiv1.ResourceList{}

	cores = getNodeResource(nodeWithNegativeCapacity, apiv1.ResourceCPU)
	assert.Equal(t, int64(0), cores)

	memory = getNodeResource(nodeWithNegativeCapacity, apiv1.ResourceMemory)
	assert.Equal(t, int64(0), memory)

}

func TestGetNodeCoresAndMemory(t *testing.T) {
	node := BuildTestNode("n1", 2000, 2048*MiB)

	cores, memory := GetNodeCoresAndMemory(node)
	assert.Equal(t, int64(2), cores)
	assert.Equal(t, int64(2048*MiB), memory)

	// if we have no cpu/memory defined in capacity we expect getNodeCoresAndMemory to return 0s
	nodeWithMissingCapacity := BuildTestNode("n1", 1000, 2*MiB)
	nodeWithMissingCapacity.Status.Capacity = apiv1.ResourceList{}

	cores, memory = GetNodeCoresAndMemory(nodeWithMissingCapacity)
	assert.Equal(t, int64(0), cores)
	assert.Equal(t, int64(0), memory)
}

func TestGetOldestPod(t *testing.T) {
	p1 := BuildTestPod("p1", 500, 1000)
	p1.CreationTimestamp = metav1.NewTime(time.Now().Add(-1 * time.Minute))
	p2 := BuildTestPod("p2", 500, 1000)
	p2.CreationTimestamp = metav1.NewTime(time.Now().Add(+1 * time.Minute))
	p3 := BuildTestPod("p3", 500, 1000)
	p3.CreationTimestamp = metav1.NewTime(time.Now())

	assert.Equal(t, p1.CreationTimestamp.Time, GetOldestCreateTime([]*apiv1.Pod{p1, p2, p3}))
	assert.Equal(t, p1.CreationTimestamp.Time, GetOldestCreateTime([]*apiv1.Pod{p3, p2, p1}))
}
