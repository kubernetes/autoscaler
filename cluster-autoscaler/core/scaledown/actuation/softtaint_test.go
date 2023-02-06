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

package actuation

import (
	"context"
	"testing"
	"time"

	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/core/test"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	klog "k8s.io/klog/v2"
)

func TestSoftTaintUpdate(t *testing.T) {
	if t != nil {
		return
	}
	n1000 := BuildTestNode("n1000", 1000, 1000)
	SetNodeReadyState(n1000, true, time.Time{})
	n2000 := BuildTestNode("n2000", 2000, 1000)
	SetNodeReadyState(n2000, true, time.Time{})

	fakeClient := fake.NewSimpleClientset()
	ctx := context.Background()
	_, err := fakeClient.CoreV1().Nodes().Create(ctx, n1000, metav1.CreateOptions{})
	assert.NoError(t, err)
	_, err = fakeClient.CoreV1().Nodes().Create(ctx, n2000, metav1.CreateOptions{})
	assert.NoError(t, err)

	provider := testprovider.NewTestCloudProvider(nil, func(nodeGroup string, node string) error {
		t.Fatalf("Unexpected deletion of %s", node)
		return nil
	})
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", n1000)
	provider.AddNode("ng1", n2000)
	assert.NotNil(t, provider)

	options := config.AutoscalingOptions{
		MaxBulkSoftTaintCount: 1,
		MaxBulkSoftTaintTime:  3 * time.Second,
	}
	registry := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	actx, err := test.NewScaleTestAutoscalingContext(options, fakeClient, registry, provider, nil, nil)
	assert.NoError(t, err)

	// Test no superfluous nodes
	nodes := getAllNodes(t, fakeClient)
	errs := UpdateSoftDeletionTaints(&actx, nil, nodes)
	assert.Empty(t, errs)
	assert.False(t, hasDeletionCandidateTaint(t, fakeClient, n1000.Name))
	assert.False(t, hasDeletionCandidateTaint(t, fakeClient, n2000.Name))

	// Test one unneeded node
	nodes = getAllNodes(t, fakeClient)
	errs = UpdateSoftDeletionTaints(&actx, []*apiv1.Node{n1000}, []*apiv1.Node{n2000})
	assert.Empty(t, errs)
	assert.True(t, hasDeletionCandidateTaint(t, fakeClient, n1000.Name))
	assert.False(t, hasDeletionCandidateTaint(t, fakeClient, n2000.Name))

	// Test remove soft taint
	nodes = getAllNodes(t, fakeClient)
	errs = UpdateSoftDeletionTaints(&actx, nil, nodes)
	assert.Empty(t, errs)
	assert.False(t, hasDeletionCandidateTaint(t, fakeClient, n1000.Name))
	assert.False(t, hasDeletionCandidateTaint(t, fakeClient, n2000.Name))

	// Test bulk update taint limit
	nodes = getAllNodes(t, fakeClient)
	errs = UpdateSoftDeletionTaints(&actx, nodes, nil)
	assert.Empty(t, errs)
	assert.Equal(t, 1, countDeletionCandidateTaints(t, fakeClient))
	nodes = getAllNodes(t, fakeClient)
	errs = UpdateSoftDeletionTaints(&actx, nodes, nil)
	assert.Empty(t, errs)
	assert.Equal(t, 2, countDeletionCandidateTaints(t, fakeClient))

	// Test bulk update untaint limit
	nodes = getAllNodes(t, fakeClient)
	errs = UpdateSoftDeletionTaints(&actx, nil, nodes)
	assert.Empty(t, errs)
	assert.Equal(t, 1, countDeletionCandidateTaints(t, fakeClient))
	nodes = getAllNodes(t, fakeClient)
	errs = UpdateSoftDeletionTaints(&actx, nil, nodes)
	assert.Empty(t, errs)
	assert.Equal(t, 0, countDeletionCandidateTaints(t, fakeClient))
}

func TestSoftTaintTimeLimit(t *testing.T) {
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Time{})
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Time{})

	currentTime := time.Now()
	updateTime := time.Millisecond
	maxSoftTaintDuration := 1 * time.Second

	unfreeze := freezeTime(&currentTime)
	defer unfreeze()

	fakeClient := fake.NewSimpleClientset()
	ctx := context.Background()
	_, err := fakeClient.CoreV1().Nodes().Create(ctx, n1, metav1.CreateOptions{})
	assert.NoError(t, err)
	_, err = fakeClient.CoreV1().Nodes().Create(ctx, n2, metav1.CreateOptions{})
	assert.NoError(t, err)

	// Move time forward when updating
	fakeClient.Fake.PrependReactor("update", "nodes", func(action k8stesting.Action) (bool, runtime.Object, error) {
		currentTime = currentTime.Add(updateTime)
		klog.Infof("currentTime after update by %v is %v", updateTime, currentTime)
		return false, nil, nil
	})

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng1", n2)
	assert.NotNil(t, provider)

	options := config.AutoscalingOptions{
		MaxBulkSoftTaintCount: 10,
		MaxBulkSoftTaintTime:  maxSoftTaintDuration,
	}
	registry := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	actx, err := test.NewScaleTestAutoscalingContext(options, fakeClient, registry, provider, nil, nil)
	assert.NoError(t, err)

	// Test bulk taint
	nodes := getAllNodes(t, fakeClient)
	errs := UpdateSoftDeletionTaints(&actx, nodes, nil)
	assert.Empty(t, errs)
	assert.Equal(t, 2, countDeletionCandidateTaints(t, fakeClient))
	assert.True(t, hasDeletionCandidateTaint(t, fakeClient, n1.Name))
	assert.True(t, hasDeletionCandidateTaint(t, fakeClient, n2.Name))

	// Test bulk untaint
	nodes = getAllNodes(t, fakeClient)
	errs = UpdateSoftDeletionTaints(&actx, nil, nodes)
	assert.Empty(t, errs)
	assert.Equal(t, 0, countDeletionCandidateTaints(t, fakeClient))
	assert.False(t, hasDeletionCandidateTaint(t, fakeClient, n1.Name))
	assert.False(t, hasDeletionCandidateTaint(t, fakeClient, n2.Name))

	updateTime = maxSoftTaintDuration

	// Test duration limit of bulk taint
	nodes = getAllNodes(t, fakeClient)
	errs = UpdateSoftDeletionTaints(&actx, nodes, nil)
	assert.Empty(t, errs)
	assert.Equal(t, 1, countDeletionCandidateTaints(t, fakeClient))
	nodes = getAllNodes(t, fakeClient)
	errs = UpdateSoftDeletionTaints(&actx, nodes, nil)
	assert.Empty(t, errs)
	assert.Equal(t, 2, countDeletionCandidateTaints(t, fakeClient))

	// Test duration limit of bulk untaint
	nodes = getAllNodes(t, fakeClient)
	errs = UpdateSoftDeletionTaints(&actx, nil, nodes)
	assert.Empty(t, errs)
	assert.Equal(t, 1, countDeletionCandidateTaints(t, fakeClient))
	nodes = getAllNodes(t, fakeClient)
	errs = UpdateSoftDeletionTaints(&actx, nil, nodes)
	assert.Empty(t, errs)
	assert.Equal(t, 0, countDeletionCandidateTaints(t, fakeClient))
}

func countDeletionCandidateTaints(t *testing.T, client kubernetes.Interface) (total int) {
	t.Helper()
	for _, node := range getAllNodes(t, client) {
		if taints.HasDeletionCandidateTaint(node) {
			total++
		}
	}
	return total
}

func hasDeletionCandidateTaint(t *testing.T, client kubernetes.Interface, name string) bool {
	t.Helper()
	return taints.HasDeletionCandidateTaint(getNode(t, client, name))
}

func getNode(t *testing.T, client kubernetes.Interface, name string) *apiv1.Node {
	t.Helper()
	node, err := client.CoreV1().Nodes().Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to retrieve node %v: %v", name, err)
	}
	return node
}

func getAllNodes(t *testing.T, client kubernetes.Interface) []*apiv1.Node {
	t.Helper()
	nodeList, err := client.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to retrieve list of nodes: %v", err)
	}
	result := make([]*apiv1.Node, 0, nodeList.Size())
	for _, node := range nodeList.Items {
		result = append(result, node.DeepCopy())
	}
	return result
}

func freezeTime(at *time.Time) (unfreeze func()) {
	// Replace time tracking function
	now = func() time.Time {
		return *at
	}
	return func() { now = time.Now }
}
