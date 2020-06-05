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
	ctx "context"
	"fmt"
	"sort"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	autoscaler_errors "k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"

	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	"k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
	kube_client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	klog "k8s.io/klog/v2"

	"strconv"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/utils/deletetaint"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
)

func TestFindUnneededNodes(t *testing.T) {
	var autoscalererr autoscaler_errors.AutoscalerError

	p1 := BuildTestPod("p1", 100, 0)
	p1.Spec.NodeName = "n1"

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")

	p2 := BuildTestPod("p2", 300, 0)
	p2.Spec.NodeName = "n2"
	p2.OwnerReferences = ownerRef

	p3 := BuildTestPod("p3", 400, 0)
	p3.OwnerReferences = ownerRef
	p3.Spec.NodeName = "n3"

	p4 := BuildTestPod("p4", 2000, 0)
	p4.OwnerReferences = ownerRef
	p4.Spec.NodeName = "n4"

	p5 := BuildTestPod("p5", 100, 0)
	p5.OwnerReferences = ownerRef
	p5.Spec.NodeName = "n5"

	p6 := BuildTestPod("p6", 500, 0)
	p6.OwnerReferences = ownerRef
	p6.Spec.NodeName = "n7"

	// Node with not replicated pod.
	n1 := BuildTestNode("n1", 1000, 10)
	// Node can be deleted.
	n2 := BuildTestNode("n2", 1000, 10)
	// Node with high utilization.
	n3 := BuildTestNode("n3", 1000, 10)
	// Node with big pod.
	n4 := BuildTestNode("n4", 10000, 10)
	// No scale down node.
	n5 := BuildTestNode("n5", 1000, 10)
	n5.Annotations = map[string]string{
		ScaleDownDisabledKey: "true",
	}
	// Node info not found.
	n6 := BuildTestNode("n6", 1000, 10)
	// Node without utilization.
	n7 := BuildTestNode("n7", 0, 10)
	// Node being deleted.
	n8 := BuildTestNode("n8", 1000, 10)
	n8.Spec.Taints = []apiv1.Taint{{Key: deletetaint.ToBeDeletedTaint, Value: strconv.FormatInt(time.Now().Unix()-301, 10)}}
	// Nod being deleted recently.
	n9 := BuildTestNode("n9", 1000, 10)
	n9.Spec.Taints = []apiv1.Taint{{Key: deletetaint.ToBeDeletedTaint, Value: strconv.FormatInt(time.Now().Unix()-60, 10)}}

	SetNodeReadyState(n1, true, time.Time{})
	SetNodeReadyState(n2, true, time.Time{})
	SetNodeReadyState(n3, true, time.Time{})
	SetNodeReadyState(n4, true, time.Time{})
	SetNodeReadyState(n5, true, time.Time{})
	SetNodeReadyState(n6, true, time.Time{})
	SetNodeReadyState(n7, true, time.Time{})
	SetNodeReadyState(n8, true, time.Time{})
	SetNodeReadyState(n9, true, time.Time{})

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng1", n2)
	provider.AddNode("ng1", n3)
	provider.AddNode("ng1", n4)
	provider.AddNode("ng1", n5)
	provider.AddNode("ng1", n7)
	provider.AddNode("ng1", n8)
	provider.AddNode("ng1", n9)

	options := config.AutoscalingOptions{
		ScaleDownUtilizationThreshold: 0.35,
		UnremovableNodeRecheckTimeout: 5 * time.Minute,
	}
	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, nil)
	assert.NoError(t, err)

	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, newBackoff())
	sd := NewScaleDown(&context, clusterStateRegistry)
	allNodes := []*apiv1.Node{n1, n2, n3, n4, n5, n7, n8, n9}

	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, allNodes, []*apiv1.Pod{p1, p2, p3, p4, p5, p6})
	autoscalererr = sd.UpdateUnneededNodes(allNodes, allNodes, time.Now(), nil)
	assert.NoError(t, autoscalererr)

	assert.Equal(t, 3, len(sd.unneededNodes))
	_, found := sd.unneededNodes["n2"]
	assert.True(t, found)
	_, found = sd.unneededNodes["n7"]
	assert.True(t, found)
	addTime, found := sd.unneededNodes["n8"]
	assert.True(t, found)
	assert.Contains(t, sd.podLocationHints, p2.Namespace+"/"+p2.Name)
	assert.Equal(t, 6, len(sd.nodeUtilizationMap))

	sd.unremovableNodes = make(map[string]time.Time)
	sd.unneededNodes["n1"] = time.Now()
	allNodes = []*apiv1.Node{n1, n2, n3, n4}
	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, allNodes, []*apiv1.Pod{p1, p2, p3, p4})
	autoscalererr = sd.UpdateUnneededNodes(allNodes, allNodes, time.Now(), nil)
	assert.NoError(t, autoscalererr)

	sd.unremovableNodes = make(map[string]time.Time)

	assert.Equal(t, 1, len(sd.unneededNodes))
	addTime2, found := sd.unneededNodes["n2"]
	assert.True(t, found)
	assert.Equal(t, addTime, addTime2)
	assert.Equal(t, 4, len(sd.nodeUtilizationMap))

	sd.unremovableNodes = make(map[string]time.Time)
	scaleDownCandidates := []*apiv1.Node{n1, n3, n4}
	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, allNodes, []*apiv1.Pod{p1, p2, p3, p4})
	autoscalererr = sd.UpdateUnneededNodes(allNodes, scaleDownCandidates, time.Now(), nil)
	assert.NoError(t, autoscalererr)

	assert.Equal(t, 0, len(sd.unneededNodes))

	// Node n1 is unneeded, but should be skipped because it has just recently been found to be unremovable
	allNodes = []*apiv1.Node{n1}
	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, allNodes, []*apiv1.Pod{})
	autoscalererr = sd.UpdateUnneededNodes(allNodes, allNodes, time.Now(), nil)
	assert.NoError(t, autoscalererr)

	assert.Equal(t, 0, len(sd.unneededNodes))
	// Verify that no other nodes are in unremovable map.
	assert.Equal(t, 1, len(sd.unremovableNodes))

	// But it should be checked after timeout
	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, allNodes, []*apiv1.Pod{})
	autoscalererr = sd.UpdateUnneededNodes(allNodes, allNodes, time.Now().Add(context.UnremovableNodeRecheckTimeout+time.Second), nil)
	assert.NoError(t, autoscalererr)

	assert.Equal(t, 1, len(sd.unneededNodes))
	// Verify that nodes that are no longer unremovable are removed.
	assert.Equal(t, 0, len(sd.unremovableNodes))
}

func TestFindUnneededGPUNodes(t *testing.T) {
	var autoscalererr autoscaler_errors.AutoscalerError

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")

	p1 := BuildTestPod("p1", 100, 0)
	p1.Spec.NodeName = "n1"
	p1.OwnerReferences = ownerRef
	RequestGpuForPod(p1, 1)
	TolerateGpuForPod(p1)

	p2 := BuildTestPod("p2", 400, 0)
	p2.Spec.NodeName = "n2"
	p2.OwnerReferences = ownerRef
	RequestGpuForPod(p2, 1)
	TolerateGpuForPod(p2)

	p3 := BuildTestPod("p3", 300, 0)
	p3.Spec.NodeName = "n3"
	p3.OwnerReferences = ownerRef
	p3.ObjectMeta.Annotations["cluster-autoscaler.kubernetes.io/safe-to-evict"] = "false"
	RequestGpuForPod(p3, 1)
	TolerateGpuForPod(p3)

	// Node with low cpu utilization and high gpu utilization
	n1 := BuildTestNode("n1", 1000, 10)
	AddGpusToNode(n1, 2)
	// Node with high cpu utilization and low gpu utilization
	n2 := BuildTestNode("n2", 1000, 10)
	AddGpusToNode(n2, 4)
	// Node with low gpu utilization and pods on node can not be interrupted
	n3 := BuildTestNode("n3", 1000, 10)
	AddGpusToNode(n3, 8)

	SetNodeReadyState(n1, true, time.Time{})
	SetNodeReadyState(n2, true, time.Time{})
	SetNodeReadyState(n3, true, time.Time{})

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng1", n2)
	provider.AddNode("ng1", n3)

	options := config.AutoscalingOptions{
		ScaleDownUtilizationThreshold:    0.35,
		ScaleDownGpuUtilizationThreshold: 0.3,
		UnremovableNodeRecheckTimeout:    5 * time.Minute,
	}
	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, nil)
	assert.NoError(t, err)

	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, newBackoff())
	sd := NewScaleDown(&context, clusterStateRegistry)
	allNodes := []*apiv1.Node{n1, n2, n3}

	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, allNodes, []*apiv1.Pod{p1, p2, p3})

	autoscalererr = sd.UpdateUnneededNodes(allNodes, allNodes, time.Now(), nil)
	assert.NoError(t, autoscalererr)
	assert.Equal(t, 1, len(sd.unneededNodes))
	_, found := sd.unneededNodes["n2"]
	assert.True(t, found)

	assert.Contains(t, sd.podLocationHints, p2.Namespace+"/"+p2.Name)
	assert.Equal(t, 3, len(sd.nodeUtilizationMap))
}

func TestPodsWithPreemptionsFindUnneededNodes(t *testing.T) {
	var autoscalererr autoscaler_errors.AutoscalerError

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")
	var priority100 int32 = 100

	p1 := BuildTestPod("p1", 600, 0)
	p1.OwnerReferences = ownerRef
	p1.Spec.Priority = &priority100
	p1.Status.NominatedNodeName = "n1"

	p2 := BuildTestPod("p2", 100, 0)
	p2.OwnerReferences = ownerRef
	p2.Spec.NodeName = "n2"

	p3 := BuildTestPod("p3", 100, 0)
	p3.OwnerReferences = ownerRef
	p3.Spec.Priority = &priority100
	p3.Status.NominatedNodeName = "n2"

	p4 := BuildTestPod("p4", 1200, 0)
	p4.OwnerReferences = ownerRef
	p4.Spec.Priority = &priority100
	p4.Status.NominatedNodeName = "n4"

	// Node with pod waiting for lower priority pod preemption, highly utilized. Can't be deleted.
	n1 := BuildTestNode("n1", 1000, 10)
	// Node with two small pods that can be moved.
	n2 := BuildTestNode("n2", 1000, 10)
	// Node without pods.
	n3 := BuildTestNode("n3", 1000, 10)
	// Node with big pod waiting for lower priority pod preemption. Can't be deleted.
	n4 := BuildTestNode("n4", 10000, 10)

	SetNodeReadyState(n1, true, time.Time{})
	SetNodeReadyState(n2, true, time.Time{})
	SetNodeReadyState(n3, true, time.Time{})
	SetNodeReadyState(n4, true, time.Time{})

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng1", n2)
	provider.AddNode("ng1", n3)
	provider.AddNode("ng1", n4)

	options := config.AutoscalingOptions{
		ScaleDownUtilizationThreshold: 0.35,
	}
	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, nil)
	assert.NoError(t, err)

	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, newBackoff())
	sd := NewScaleDown(&context, clusterStateRegistry)

	allNodes := []*apiv1.Node{n1, n2, n3, n4}
	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, allNodes, []*apiv1.Pod{p1, p2, p3, p4})
	autoscalererr = sd.UpdateUnneededNodes(allNodes, allNodes, time.Now(), nil)
	assert.NoError(t, autoscalererr)
	assert.Equal(t, 2, len(sd.unneededNodes))
	klog.Warningf("Unneeded nodes %v", sd.unneededNodes)
	_, found := sd.unneededNodes["n2"]
	assert.True(t, found)
	_, found = sd.unneededNodes["n3"]
	assert.True(t, found)
	assert.Contains(t, sd.podLocationHints, p2.Namespace+"/"+p2.Name)
	assert.Contains(t, sd.podLocationHints, p3.Namespace+"/"+p3.Name)
	assert.Equal(t, 4, len(sd.nodeUtilizationMap))
}

func TestFindUnneededMaxCandidates(t *testing.T) {
	var autoscalererr autoscaler_errors.AutoscalerError

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 100, 2)

	numNodes := 100
	nodes := make([]*apiv1.Node, 0, numNodes)
	for i := 0; i < numNodes; i++ {
		n := BuildTestNode(fmt.Sprintf("n%v", i), 1000, 10)
		SetNodeReadyState(n, true, time.Time{})
		provider.AddNode("ng1", n)
		nodes = append(nodes, n)
	}

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")

	pods := make([]*apiv1.Pod, 0, numNodes)
	for i := 0; i < numNodes; i++ {
		p := BuildTestPod(fmt.Sprintf("p%v", i), 100, 0)
		p.Spec.NodeName = fmt.Sprintf("n%v", i)
		p.OwnerReferences = ownerRef
		pods = append(pods, p)
	}

	numCandidates := 30

	options := config.AutoscalingOptions{
		ScaleDownUtilizationThreshold:    0.35,
		ScaleDownNonEmptyCandidatesCount: numCandidates,
		ScaleDownCandidatesPoolRatio:     1,
		ScaleDownCandidatesPoolMinCount:  1000,
	}
	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, nil)
	assert.NoError(t, err)

	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, newBackoff())
	sd := NewScaleDown(&context, clusterStateRegistry)

	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, pods)
	autoscalererr = sd.UpdateUnneededNodes(nodes, nodes, time.Now(), nil)
	assert.NoError(t, autoscalererr)
	assert.Equal(t, numCandidates, len(sd.unneededNodes))
	// Simulate one of the unneeded nodes got deleted
	deleted := sd.unneededNodesList[len(sd.unneededNodesList)-1]
	for i, node := range nodes {
		if node.Name == deleted.Name {
			// Move pod away from the node
			var newNode int
			if i >= 1 {
				newNode = i - 1
			} else {
				newNode = i + 1
			}
			pods[i].Spec.NodeName = nodes[newNode].Name
			nodes[i] = nodes[len(nodes)-1]
			nodes[len(nodes)-1] = nil
			nodes = nodes[:len(nodes)-1]
			break
		}
	}

	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, pods)
	autoscalererr = sd.UpdateUnneededNodes(nodes, nodes, time.Now(), nil)
	assert.NoError(t, autoscalererr)
	// Check that the deleted node was replaced
	assert.Equal(t, numCandidates, len(sd.unneededNodes))
	assert.NotContains(t, sd.unneededNodes, deleted)
}

func TestFindUnneededEmptyNodes(t *testing.T) {
	var autoscalererr autoscaler_errors.AutoscalerError

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 100, 100)

	// 30 empty nodes and 70 heavily underutilized.
	numNodes := 100
	numEmpty := 30
	nodes := make([]*apiv1.Node, 0, numNodes)
	for i := 0; i < numNodes; i++ {
		n := BuildTestNode(fmt.Sprintf("n%v", i), 1000, 10)
		SetNodeReadyState(n, true, time.Time{})
		provider.AddNode("ng1", n)
		nodes = append(nodes, n)
	}

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")

	pods := make([]*apiv1.Pod, 0, numNodes)
	for i := 0; i < numNodes-numEmpty; i++ {
		p := BuildTestPod(fmt.Sprintf("p%v", i), 100, 0)
		p.Spec.NodeName = fmt.Sprintf("n%v", i)
		p.OwnerReferences = ownerRef
		pods = append(pods, p)
	}

	numCandidates := 30

	options := config.AutoscalingOptions{
		ScaleDownUtilizationThreshold:    0.35,
		ScaleDownNonEmptyCandidatesCount: numCandidates,
		ScaleDownCandidatesPoolRatio:     1.0,
		ScaleDownCandidatesPoolMinCount:  1000,
	}
	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, nil)
	assert.NoError(t, err)

	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, newBackoff())
	sd := NewScaleDown(&context, clusterStateRegistry)
	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, pods)
	autoscalererr = sd.UpdateUnneededNodes(nodes, nodes, time.Now(), nil)
	assert.NoError(t, autoscalererr)
	for _, node := range sd.unneededNodesList {
		t.Log(node.Name)
	}
	assert.Equal(t, numEmpty+numCandidates, len(sd.unneededNodes))
}

func TestFindUnneededNodePool(t *testing.T) {
	var autoscalererr autoscaler_errors.AutoscalerError

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 100, 100)

	numNodes := 100
	nodes := make([]*apiv1.Node, 0, numNodes)
	for i := 0; i < numNodes; i++ {
		n := BuildTestNode(fmt.Sprintf("n%v", i), 1000, 10)
		SetNodeReadyState(n, true, time.Time{})
		provider.AddNode("ng1", n)
		nodes = append(nodes, n)
	}

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")

	pods := make([]*apiv1.Pod, 0, numNodes)
	for i := 0; i < numNodes; i++ {
		p := BuildTestPod(fmt.Sprintf("p%v", i), 100, 0)
		p.Spec.NodeName = fmt.Sprintf("n%v", i)
		p.OwnerReferences = ownerRef
		pods = append(pods, p)
	}

	numCandidates := 30

	options := config.AutoscalingOptions{
		ScaleDownUtilizationThreshold:    0.35,
		ScaleDownNonEmptyCandidatesCount: numCandidates,
		ScaleDownCandidatesPoolRatio:     0.1,
		ScaleDownCandidatesPoolMinCount:  10,
	}
	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, nil, provider, nil)
	assert.NoError(t, err)

	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, newBackoff())
	sd := NewScaleDown(&context, clusterStateRegistry)
	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, pods)
	autoscalererr = sd.UpdateUnneededNodes(nodes, nodes, time.Now(), nil)
	assert.NoError(t, autoscalererr)
	assert.NotEmpty(t, sd.unneededNodes)
}

func TestDeleteNode(t *testing.T) {
	// common parameters
	nodeDeleteFailedFunc :=
		func(string, string) error {
			return fmt.Errorf("won't remove node")
		}
	podNotFoundFunc :=
		func(action core.Action) (bool, runtime.Object, error) {
			return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
		}

	// scenarios
	testScenarios := []struct {
		name               string
		pods               []string
		drainSuccess       bool
		nodeDeleteSuccess  bool
		expectedDeletion   bool
		expectedResultType status.NodeDeleteResultType
	}{
		{
			name:               "successful attempt to delete node with pods",
			pods:               []string{"p1", "p2"},
			drainSuccess:       true,
			nodeDeleteSuccess:  true,
			expectedDeletion:   true,
			expectedResultType: status.NodeDeleteOk,
		},
		/* Temporarily disabled as it takes several minutes due to hardcoded timeout.
		* TODO(aleksandra-malinowska): move MaxPodEvictionTime to AutoscalingContext.
		{
			name:              "failed on drain",
			pods:              []string{"p1", "p2"},
			drainSuccess:      false,
			nodeDeleteSuccess: true,
			expectedDeletion:  false,
			expectedResultType: status.NodeDeleteErrorFailedToEvictPods,
		},
		*/
		{
			name:               "failed on node delete",
			pods:               []string{"p1", "p2"},
			drainSuccess:       true,
			nodeDeleteSuccess:  false,
			expectedDeletion:   false,
			expectedResultType: status.NodeDeleteErrorFailedToDelete,
		},
		{
			name:               "successful attempt to delete empty node",
			pods:               []string{},
			drainSuccess:       true,
			nodeDeleteSuccess:  true,
			expectedDeletion:   true,
			expectedResultType: status.NodeDeleteOk,
		},
		{
			name:               "failed attempt to delete empty node",
			pods:               []string{},
			drainSuccess:       true,
			nodeDeleteSuccess:  false,
			expectedDeletion:   false,
			expectedResultType: status.NodeDeleteErrorFailedToDelete,
		},
	}

	for _, scenario := range testScenarios {
		// run each scenario as an independent test
		t.Run(scenario.name, func(t *testing.T) {
			// set up test channels
			updatedNodes := make(chan string, 10)
			deletedNodes := make(chan string, 10)
			deletedPods := make(chan string, 10)

			// set up test data
			n1 := BuildTestNode("n1", 1000, 1000)
			SetNodeReadyState(n1, true, time.Time{})
			pods := make([]*apiv1.Pod, len(scenario.pods))
			for i, podName := range scenario.pods {
				pod := BuildTestPod(podName, 100, 0)
				pods[i] = pod
			}

			// set up fake provider
			deleteNodeHandler := nodeDeleteFailedFunc
			if scenario.nodeDeleteSuccess {
				deleteNodeHandler =
					func(nodeGroup string, node string) error {
						deletedNodes <- node
						return nil
					}
			}
			provider := testprovider.NewTestCloudProvider(nil, deleteNodeHandler)
			provider.AddNodeGroup("ng1", 1, 100, 100)
			provider.AddNode("ng1", n1)

			// set up fake client
			fakeClient := &fake.Clientset{}
			fakeNode := n1.DeepCopy()
			fakeClient.Fake.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
				return true, fakeNode.DeepCopy(), nil
			})
			fakeClient.Fake.AddReactor("update", "nodes",
				func(action core.Action) (bool, runtime.Object, error) {
					update := action.(core.UpdateAction)
					obj := update.GetObject().(*apiv1.Node)
					taints := make([]string, 0, len(obj.Spec.Taints))
					for _, taint := range obj.Spec.Taints {
						taints = append(taints, taint.Key)
					}
					updatedNodes <- fmt.Sprintf("%s-%s", obj.Name, taints)
					fakeNode = obj.DeepCopy()
					return true, obj, nil
				})
			fakeClient.Fake.AddReactor("create", "pods",
				func(action core.Action) (bool, runtime.Object, error) {
					if !scenario.drainSuccess {
						return true, nil, fmt.Errorf("won't evict")
					}
					createAction := action.(core.CreateAction)
					if createAction == nil {
						return false, nil, nil
					}
					eviction := createAction.GetObject().(*policyv1.Eviction)
					if eviction == nil {
						return false, nil, nil
					}
					deletedPods <- eviction.Name
					return true, nil, nil
				})
			fakeClient.Fake.AddReactor("get", "pods", podNotFoundFunc)

			// build context
			registry := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
			context, err := NewScaleTestAutoscalingContext(config.AutoscalingOptions{}, fakeClient, registry, provider, nil)
			assert.NoError(t, err)

			clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, newBackoff())
			sd := NewScaleDown(&context, clusterStateRegistry)

			// attempt delete
			result := sd.deleteNode(n1, pods, provider.GetNodeGroup("ng1"))

			// verify
			if scenario.expectedDeletion {
				assert.NoError(t, result.Err)
				assert.Equal(t, n1.Name, utils.GetStringFromChanImmediately(deletedNodes))
			} else {
				assert.NotNil(t, result.Err)
			}
			assert.Equal(t, utils.NothingReturned, utils.GetStringFromChanImmediately(deletedNodes))
			assert.Equal(t, scenario.expectedResultType, result.ResultType)

			taintedUpdate := fmt.Sprintf("%s-%s", n1.Name, []string{deletetaint.ToBeDeletedTaint})
			assert.Equal(t, taintedUpdate, utils.GetStringFromChan(updatedNodes))
			if !scenario.expectedDeletion {
				untaintedUpdate := fmt.Sprintf("%s-%s", n1.Name, []string{})
				assert.Equal(t, untaintedUpdate, utils.GetStringFromChanImmediately(updatedNodes))
			}
			assert.Equal(t, utils.NothingReturned, utils.GetStringFromChanImmediately(updatedNodes))
		})
	}
}

func TestDrainNode(t *testing.T) {
	deletedPods := make(chan string, 10)
	fakeClient := &fake.Clientset{}

	p1 := BuildTestPod("p1", 100, 0)
	p2 := BuildTestPod("p2", 300, 0)
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Time{})

	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
		createAction := action.(core.CreateAction)
		if createAction == nil {
			return false, nil, nil
		}
		eviction := createAction.GetObject().(*policyv1.Eviction)
		if eviction == nil {
			return false, nil, nil
		}
		deletedPods <- eviction.Name
		return true, nil, nil
	})
	_, err := drainNode(n1, []*apiv1.Pod{p1, p2}, fakeClient, kube_util.CreateEventRecorder(fakeClient), 20, 5*time.Second, 0*time.Second, PodEvictionHeadroom)
	assert.NoError(t, err)
	deleted := make([]string, 0)
	deleted = append(deleted, utils.GetStringFromChan(deletedPods))
	deleted = append(deleted, utils.GetStringFromChan(deletedPods))
	sort.Strings(deleted)
	assert.Equal(t, p1.Name, deleted[0])
	assert.Equal(t, p2.Name, deleted[1])
}

func TestDrainNodeWithRescheduled(t *testing.T) {
	deletedPods := make(chan string, 10)
	fakeClient := &fake.Clientset{}

	p1 := BuildTestPod("p1", 100, 0)
	p2 := BuildTestPod("p2", 300, 0)
	p2Rescheduled := BuildTestPod("p2", 300, 0)
	p2Rescheduled.Spec.NodeName = "n2"
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Time{})

	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)
		if getAction == nil {
			return false, nil, nil
		}
		if getAction.GetName() == "p2" {
			return true, p2Rescheduled, nil
		}
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
		createAction := action.(core.CreateAction)
		if createAction == nil {
			return false, nil, nil
		}
		eviction := createAction.GetObject().(*policyv1.Eviction)
		if eviction == nil {
			return false, nil, nil
		}
		deletedPods <- eviction.Name
		return true, nil, nil
	})
	_, err := drainNode(n1, []*apiv1.Pod{p1, p2}, fakeClient, kube_util.CreateEventRecorder(fakeClient), 20, 5*time.Second, 0*time.Second, PodEvictionHeadroom)
	assert.NoError(t, err)
	deleted := make([]string, 0)
	deleted = append(deleted, utils.GetStringFromChan(deletedPods))
	deleted = append(deleted, utils.GetStringFromChan(deletedPods))
	sort.Strings(deleted)
	assert.Equal(t, p1.Name, deleted[0])
	assert.Equal(t, p2.Name, deleted[1])
}

func TestDrainNodeWithRetries(t *testing.T) {
	deletedPods := make(chan string, 10)
	// Simulate pdb of size 1 by making the 'eviction' goroutine:
	// - read from (at first empty) channel
	// - if it's empty, fail and write to it, then retry
	// - succeed on successful read.
	ticket := make(chan bool, 1)
	fakeClient := &fake.Clientset{}

	p1 := BuildTestPod("p1", 100, 0)
	p2 := BuildTestPod("p2", 300, 0)
	p3 := BuildTestPod("p3", 300, 0)
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Time{})

	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
		createAction := action.(core.CreateAction)
		if createAction == nil {
			return false, nil, nil
		}
		eviction := createAction.GetObject().(*policyv1.Eviction)
		if eviction == nil {
			return false, nil, nil
		}
		select {
		case <-ticket:
			deletedPods <- eviction.Name
			return true, nil, nil
		default:
			select {
			case ticket <- true:
			default:
			}
			return true, nil, fmt.Errorf("too many concurrent evictions")
		}
	})
	_, err := drainNode(n1, []*apiv1.Pod{p1, p2, p3}, fakeClient, kube_util.CreateEventRecorder(fakeClient), 20, 5*time.Second, 0*time.Second, PodEvictionHeadroom)
	assert.NoError(t, err)
	deleted := make([]string, 0)
	deleted = append(deleted, utils.GetStringFromChan(deletedPods))
	deleted = append(deleted, utils.GetStringFromChan(deletedPods))
	deleted = append(deleted, utils.GetStringFromChan(deletedPods))
	sort.Strings(deleted)
	assert.Equal(t, p1.Name, deleted[0])
	assert.Equal(t, p2.Name, deleted[1])
	assert.Equal(t, p3.Name, deleted[2])
}

func TestDrainNodeEvictionFailure(t *testing.T) {
	fakeClient := &fake.Clientset{}

	p1 := BuildTestPod("p1", 100, 0)
	p2 := BuildTestPod("p2", 100, 0)
	p3 := BuildTestPod("p3", 100, 0)
	p4 := BuildTestPod("p4", 100, 0)
	n1 := BuildTestNode("n1", 1000, 1000)
	e2 := fmt.Errorf("eviction_error: p2")
	e4 := fmt.Errorf("eviction_error: p4")
	SetNodeReadyState(n1, true, time.Time{})

	fakeClient.Fake.AddReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
		createAction := action.(core.CreateAction)
		if createAction == nil {
			return false, nil, nil
		}
		eviction := createAction.GetObject().(*policyv1.Eviction)
		if eviction == nil {
			return false, nil, nil
		}

		if eviction.Name == "p2" {
			return true, nil, e2
		}
		if eviction.Name == "p4" {
			return true, nil, e4
		}
		return true, nil, nil
	})

	evictionResults, err := drainNode(n1, []*apiv1.Pod{p1, p2, p3, p4}, fakeClient, kube_util.CreateEventRecorder(fakeClient), 20, 0*time.Second, 0*time.Second, PodEvictionHeadroom)
	assert.Error(t, err)
	assert.Equal(t, 4, len(evictionResults))
	assert.Equal(t, *p1, *evictionResults["p1"].Pod)
	assert.Equal(t, *p2, *evictionResults["p2"].Pod)
	assert.Equal(t, *p3, *evictionResults["p3"].Pod)
	assert.Equal(t, *p4, *evictionResults["p4"].Pod)
	assert.NoError(t, evictionResults["p1"].Err)
	assert.Contains(t, evictionResults["p2"].Err.Error(), e2.Error())
	assert.NoError(t, evictionResults["p3"].Err)
	assert.Contains(t, evictionResults["p4"].Err.Error(), e4.Error())
	assert.False(t, evictionResults["p1"].TimedOut)
	assert.True(t, evictionResults["p2"].TimedOut)
	assert.False(t, evictionResults["p3"].TimedOut)
	assert.True(t, evictionResults["p4"].TimedOut)
	assert.True(t, evictionResults["p1"].WasEvictionSuccessful())
	assert.False(t, evictionResults["p2"].WasEvictionSuccessful())
	assert.True(t, evictionResults["p3"].WasEvictionSuccessful())
	assert.False(t, evictionResults["p4"].WasEvictionSuccessful())
}

func TestDrainNodeDisappearanceFailure(t *testing.T) {
	fakeClient := &fake.Clientset{}

	p1 := BuildTestPod("p1", 100, 0)
	p2 := BuildTestPod("p2", 100, 0)
	p3 := BuildTestPod("p3", 100, 0)
	p4 := BuildTestPod("p4", 100, 0)
	e2 := fmt.Errorf("disappearance_error: p2")
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Time{})

	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)
		if getAction == nil {
			return false, nil, nil
		}
		if getAction.GetName() == "p2" {
			return true, nil, e2
		}
		if getAction.GetName() == "p4" {
			return true, nil, nil
		}
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("create", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, nil
	})

	evictionResults, err := drainNode(n1, []*apiv1.Pod{p1, p2, p3, p4}, fakeClient, kube_util.CreateEventRecorder(fakeClient), 0, 0*time.Second, 0*time.Second, 0*time.Second)
	assert.Error(t, err)
	assert.Equal(t, 4, len(evictionResults))
	assert.Equal(t, *p1, *evictionResults["p1"].Pod)
	assert.Equal(t, *p2, *evictionResults["p2"].Pod)
	assert.Equal(t, *p3, *evictionResults["p3"].Pod)
	assert.Equal(t, *p4, *evictionResults["p4"].Pod)
	assert.NoError(t, evictionResults["p1"].Err)
	assert.Contains(t, evictionResults["p2"].Err.Error(), e2.Error())
	assert.NoError(t, evictionResults["p3"].Err)
	assert.NoError(t, evictionResults["p4"].Err)
	assert.False(t, evictionResults["p1"].TimedOut)
	assert.True(t, evictionResults["p2"].TimedOut)
	assert.False(t, evictionResults["p3"].TimedOut)
	assert.True(t, evictionResults["p4"].TimedOut)
	assert.True(t, evictionResults["p1"].WasEvictionSuccessful())
	assert.False(t, evictionResults["p2"].WasEvictionSuccessful())
	assert.True(t, evictionResults["p3"].WasEvictionSuccessful())
	assert.False(t, evictionResults["p4"].WasEvictionSuccessful())
}

func TestScaleDown(t *testing.T) {
	var autoscalererr autoscaler_errors.AutoscalerError

	deletedPods := make(chan string, 10)
	updatedNodes := make(chan string, 10)
	deletedNodes := make(chan string, 10)
	fakeClient := &fake.Clientset{}

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "job",
			Namespace: "default",
			SelfLink:  "/apivs/batch/v1/namespaces/default/jobs/job",
		},
	}
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Time{})
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Time{})
	p1 := BuildTestPod("p1", 100, 0)
	p1.OwnerReferences = GenerateOwnerReferences(job.Name, "Job", "batch/v1", "")

	p2 := BuildTestPod("p2", 800, 0)

	p1.Spec.NodeName = "n1"
	p2.Spec.NodeName = "n2"

	fakeClient.Fake.AddReactor("list", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, &apiv1.PodList{Items: []apiv1.Pod{*p1, *p2}}, nil
	})
	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)
		switch getAction.GetName() {
		case n1.Name:
			return true, n1, nil
		case n2.Name:
			return true, n2, nil
		}
		return true, nil, fmt.Errorf("wrong node: %v", getAction.GetName())
	})
	fakeClient.Fake.AddReactor("delete", "pods", func(action core.Action) (bool, runtime.Object, error) {
		deleteAction := action.(core.DeleteAction)
		deletedPods <- deleteAction.GetName()
		return true, nil, nil
	})
	fakeClient.Fake.AddReactor("update", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		update := action.(core.UpdateAction)
		obj := update.GetObject().(*apiv1.Node)
		updatedNodes <- obj.Name
		return true, obj, nil
	})

	provider := testprovider.NewTestCloudProvider(nil, func(nodeGroup string, node string) error {
		deletedNodes <- node
		return nil
	})
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng1", n2)
	assert.NotNil(t, provider)

	options := config.AutoscalingOptions{
		ScaleDownUtilizationThreshold: 0.5,
		ScaleDownUnneededTime:         time.Minute,
		MaxGracefulTerminationSec:     60,
	}
	jobLister, err := kube_util.NewTestJobLister([]*batchv1.Job{&job})
	assert.NoError(t, err)
	registry := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, nil, nil, jobLister, nil, nil)

	context, err := NewScaleTestAutoscalingContext(options, fakeClient, registry, provider, nil)
	assert.NoError(t, err)
	nodes := []*apiv1.Node{n1, n2}

	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, newBackoff())
	scaleDown := NewScaleDown(&context, clusterStateRegistry)
	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, []*apiv1.Pod{p1, p2})
	autoscalererr = scaleDown.UpdateUnneededNodes(nodes, nodes, time.Now().Add(-5*time.Minute), nil)
	assert.NoError(t, autoscalererr)
	scaleDownStatus, err := scaleDown.TryToScaleDown(time.Now(), nil)
	waitForDeleteToFinish(t, scaleDown)
	assert.NoError(t, err)
	assert.Equal(t, status.ScaleDownNodeDeleteStarted, scaleDownStatus.Result)
	assert.Equal(t, n1.Name, utils.GetStringFromChan(deletedNodes))
	assert.Equal(t, n1.Name, utils.GetStringFromChan(updatedNodes))
}

func waitForDeleteToFinish(t *testing.T, sd *ScaleDown) {
	for start := time.Now(); time.Since(start) < 20*time.Second; time.Sleep(100 * time.Millisecond) {
		if !sd.nodeDeletionTracker.IsNonEmptyNodeDeleteInProgress() {
			return
		}
	}
	t.Fatalf("Node delete not finished")
}

// this IGNORES duplicates
func assertEqualSet(t *testing.T, a []string, b []string) {
	assertSubset(t, a, b)
	assertSubset(t, b, a)
}

// this IGNORES duplicates
func assertSubset(t *testing.T, a []string, b []string) {
	for _, x := range a {
		found := false
		for _, y := range b {
			if x == y {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Failed to find %s (from %s) in %v", x, a, b)
		}
	}
}

var defaultScaleDownOptions = config.AutoscalingOptions{
	ScaleDownUtilizationThreshold:    0.5,
	ScaleDownGpuUtilizationThreshold: 0.5,
	ScaleDownUnneededTime:            time.Minute,
	MaxGracefulTerminationSec:        60,
	MaxEmptyBulkDelete:               10,
	MinCoresTotal:                    0,
	MinMemoryTotal:                   0,
	MaxCoresTotal:                    config.DefaultMaxClusterCores,
	MaxMemoryTotal:                   config.DefaultMaxClusterMemory * units.GiB,
}

func TestScaleDownEmptyMultipleNodeGroups(t *testing.T) {
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1_1", 1000, 1000, 0, true, "ng1"},
			{"n1_2", 1000, 1000, 0, true, "ng1"},
			{"n2_1", 1000, 1000, 0, true, "ng2"},
			{"n2_2", 1000, 1000, 0, true, "ng2"},
		},
		options:                defaultScaleDownOptions,
		expectedScaleDowns:     []string{"n1_1", "n1_2", "n2_1", "n2_2"},
		expectedScaleDownCount: 2,
	}
	simpleScaleDownEmpty(t, config)
}

func TestScaleDownEmptySingleNodeGroup(t *testing.T) {
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 1000, 1000, 0, true, "ng1"},
			{"n2", 1000, 1000, 0, true, "ng1"},
			{"n3", 1000, 1000, 0, true, "ng1"},
			{"n4", 1000, 1000, 0, true, "ng1"},
			{"n5", 1000, 1000, 0, true, "ng1"},
			{"n6", 1000, 1000, 0, true, "ng1"},
			{"n7", 1000, 1000, 0, true, "ng1"},
			{"n8", 1000, 1000, 0, true, "ng1"},
			{"n9", 1000, 1000, 0, true, "ng1"},
			{"n10", 1000, 1000, 0, true, "ng1"},
			{"n11", 1000, 1000, 0, true, "ng1"},
			{"n12", 1000, 1000, 0, true, "ng1"},
		},
		options:                defaultScaleDownOptions,
		expectedScaleDowns:     []string{"n1", "n2", "n3", "n4", "n5", "n6", "n7", "n8", "n9", "n10", "n11", "n12"},
		expectedScaleDownCount: 10,
	}
	simpleScaleDownEmpty(t, config)
}

func TestScaleDownEmptyMinCoresLimitHit(t *testing.T) {
	options := defaultScaleDownOptions
	options.MinCoresTotal = 2
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 2000, 1000, 0, true, "ng1"},
			{"n2", 1000, 1000, 0, true, "ng1"},
		},
		options:            options,
		expectedScaleDowns: []string{"n2"},
	}
	simpleScaleDownEmpty(t, config)
}

func TestScaleDownEmptyMinMemoryLimitHit(t *testing.T) {
	options := defaultScaleDownOptions
	options.MinMemoryTotal = 4000 * utils.MiB
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 2000, 1000 * utils.MiB, 0, true, "ng1"},
			{"n2", 1000, 1000 * utils.MiB, 0, true, "ng1"},
			{"n3", 1000, 1000 * utils.MiB, 0, true, "ng1"},
			{"n4", 1000, 3000 * utils.MiB, 0, true, "ng1"},
		},
		options:                options,
		expectedScaleDowns:     []string{"n1", "n2", "n3"},
		expectedScaleDownCount: 2,
	}
	simpleScaleDownEmpty(t, config)
}

func TestScaleDownEmptyMinGpuLimitHit(t *testing.T) {
	options := defaultScaleDownOptions
	options.GpuTotal = []config.GpuLimits{
		{
			GpuType: gpu.DefaultGPUType,
			Min:     4,
			Max:     50,
		},
		{
			GpuType: "nvidia-tesla-p100", // this one should not trigger
			Min:     5,
			Max:     50,
		},
	}
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 1000, 1000 * utils.MiB, 1, true, "ng1"},
			{"n2", 1000, 1000 * utils.MiB, 1, true, "ng1"},
			{"n3", 1000, 1000 * utils.MiB, 1, true, "ng1"},
			{"n4", 1000, 1000 * utils.MiB, 1, true, "ng1"},
			{"n5", 1000, 1000 * utils.MiB, 1, true, "ng1"},
			{"n6", 1000, 1000 * utils.MiB, 1, true, "ng1"},
		},
		options:                options,
		expectedScaleDowns:     []string{"n1", "n2", "n3", "n4", "n5", "n6"},
		expectedScaleDownCount: 2,
	}
	simpleScaleDownEmpty(t, config)
}

func TestScaleDownEmptyMinGroupSizeLimitHit(t *testing.T) {
	options := defaultScaleDownOptions
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 2000, 1000, 0, true, "ng1"},
		},
		options:            options,
		expectedScaleDowns: []string{},
	}
	simpleScaleDownEmpty(t, config)
}

func TestScaleDownEmptyMinGroupSizeLimitHitWhenOneNodeIsBeingDeleted(t *testing.T) {
	nodeDeletionTracker := NewNodeDeletionTracker()
	nodeDeletionTracker.StartDeletion("ng1")
	nodeDeletionTracker.StartDeletion("ng1")
	options := defaultScaleDownOptions
	config := &scaleTestConfig{
		nodes: []nodeConfig{
			{"n1", 2000, 1000, 0, true, "ng1"},
			{"n2", 2000, 1000, 0, true, "ng1"},
			{"n3", 2000, 1000, 0, true, "ng1"},
		},
		options:             options,
		expectedScaleDowns:  []string{},
		nodeDeletionTracker: nodeDeletionTracker,
	}
	simpleScaleDownEmpty(t, config)
}

func simpleScaleDownEmpty(t *testing.T, config *scaleTestConfig) {
	var autoscalererr autoscaler_errors.AutoscalerError

	updatedNodes := make(chan string, 30)
	deletedNodes := make(chan string, 30)
	fakeClient := &fake.Clientset{}

	nodes := make([]*apiv1.Node, len(config.nodes))
	nodesMap := make(map[string]*apiv1.Node)
	groups := make(map[string][]*apiv1.Node)

	provider := testprovider.NewTestCloudProvider(nil, func(nodeGroup string, node string) error {
		deletedNodes <- node
		return nil
	})

	for i, n := range config.nodes {
		node := BuildTestNode(n.name, n.cpu, n.memory)
		if n.gpu > 0 {
			AddGpusToNode(node, n.gpu)
			node.Labels[provider.GPULabel()] = gpu.DefaultGPUType
		}
		SetNodeReadyState(node, n.ready, time.Time{})
		nodesMap[n.name] = node
		nodes[i] = node
		groups[n.group] = append(groups[n.group], node)
	}

	fakeClient.Fake.AddReactor("list", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, &apiv1.PodList{Items: []apiv1.Pod{}}, nil
	})
	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)
		if node, found := nodesMap[getAction.GetName()]; found {
			return true, node, nil
		}
		return true, nil, fmt.Errorf("wrong node: %v", getAction.GetName())

	})
	fakeClient.Fake.AddReactor("update", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		update := action.(core.UpdateAction)
		obj := update.GetObject().(*apiv1.Node)
		updatedNodes <- obj.Name
		return true, obj, nil
	})

	for name, nodesInGroup := range groups {
		provider.AddNodeGroup(name, 1, 10, len(nodesInGroup))
		for _, n := range nodesInGroup {
			provider.AddNode(name, n)
		}
	}

	resourceLimiter := context.NewResourceLimiterFromAutoscalingOptions(config.options)
	provider.SetResourceLimiter(resourceLimiter)

	assert.NotNil(t, provider)

	registry := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	context, err := NewScaleTestAutoscalingContext(config.options, fakeClient, registry, provider, nil)
	assert.NoError(t, err)

	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, newBackoff())
	scaleDown := NewScaleDown(&context, clusterStateRegistry)
	if config.nodeDeletionTracker != nil {
		scaleDown.nodeDeletionTracker = config.nodeDeletionTracker
	}

	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, []*apiv1.Pod{})
	autoscalererr = scaleDown.UpdateUnneededNodes(nodes, nodes, time.Now().Add(-5*time.Minute), nil)
	assert.NoError(t, autoscalererr)
	scaleDownStatus, err := scaleDown.TryToScaleDown(time.Now(), nil)
	assert.False(t, scaleDown.nodeDeletionTracker.IsNonEmptyNodeDeleteInProgress())

	assert.NoError(t, err)
	var expectedScaleDownResult status.ScaleDownResult
	if len(config.expectedScaleDowns) > 0 {
		expectedScaleDownResult = status.ScaleDownNodeDeleteStarted
	} else {
		expectedScaleDownResult = status.ScaleDownNoUnneeded
	}
	assert.Equal(t, expectedScaleDownResult, scaleDownStatus.Result)

	expectedScaleDownCount := config.expectedScaleDownCount
	if config.expectedScaleDownCount == 0 {
		// For backwards compatibility.
		expectedScaleDownCount = len(config.expectedScaleDowns)
	}

	// Check the channel (and make sure there isn't more than there should be).
	// Report only up to 10 extra nodes found.
	deleted := make([]string, 0, expectedScaleDownCount+10)
	for i := 0; i < expectedScaleDownCount+10; i++ {
		d := utils.GetStringFromChan(deletedNodes)
		if d == utils.NothingReturned { // a closed channel yields empty value
			break
		}
		deleted = append(deleted, d)
	}

	assert.Equal(t, expectedScaleDownCount, len(deleted))
	assert.Subset(t, config.expectedScaleDowns, deleted)
}

func TestNoScaleDownUnready(t *testing.T) {
	var autoscalererr autoscaler_errors.AutoscalerError
	fakeClient := &fake.Clientset{}
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, false, time.Now().Add(-3*time.Minute))
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Time{})
	p2 := BuildTestPod("p2", 800, 0)
	p2.Spec.NodeName = "n2"

	fakeClient.Fake.AddReactor("list", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, &apiv1.PodList{Items: []apiv1.Pod{*p2}}, nil
	})
	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)
		switch getAction.GetName() {
		case n1.Name:
			return true, n1, nil
		case n2.Name:
			return true, n2, nil
		}
		return true, nil, fmt.Errorf("wrong node: %v", getAction.GetName())
	})

	provider := testprovider.NewTestCloudProvider(nil, func(nodeGroup string, node string) error {
		t.Fatalf("Unexpected deletion of %s", node)
		return nil
	})
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng1", n2)

	options := config.AutoscalingOptions{
		ScaleDownUtilizationThreshold: 0.5,
		ScaleDownUnneededTime:         time.Minute,
		ScaleDownUnreadyTime:          time.Hour,
		MaxGracefulTerminationSec:     60,
	}
	registry := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	context, err := NewScaleTestAutoscalingContext(options, fakeClient, registry, provider, nil)
	assert.NoError(t, err)

	nodes := []*apiv1.Node{n1, n2}

	// N1 is unready so it requires a bigger unneeded time.
	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, newBackoff())
	scaleDown := NewScaleDown(&context, clusterStateRegistry)
	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, []*apiv1.Pod{p2})
	autoscalererr = scaleDown.UpdateUnneededNodes(nodes, nodes, time.Now().Add(-5*time.Minute), nil)
	assert.NoError(t, autoscalererr)
	scaleDownStatus, err := scaleDown.TryToScaleDown(time.Now(), nil)
	waitForDeleteToFinish(t, scaleDown)

	assert.NoError(t, err)
	assert.Equal(t, status.ScaleDownNoUnneeded, scaleDownStatus.Result)

	deletedNodes := make(chan string, 10)

	provider = testprovider.NewTestCloudProvider(nil, func(nodeGroup string, node string) error {
		deletedNodes <- node
		return nil
	})
	SetNodeReadyState(n1, false, time.Now().Add(-3*time.Hour))
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng1", n2)

	// N1 has been unready for 2 hours, ok to delete.
	context.CloudProvider = provider
	scaleDown = NewScaleDown(&context, clusterStateRegistry)
	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, []*apiv1.Pod{p2})
	autoscalererr = scaleDown.UpdateUnneededNodes(nodes, nodes, time.Now().Add(-2*time.Hour), nil)
	assert.NoError(t, autoscalererr)
	scaleDownStatus, err = scaleDown.TryToScaleDown(time.Now(), nil)
	waitForDeleteToFinish(t, scaleDown)

	assert.NoError(t, err)
	assert.Equal(t, status.ScaleDownNodeDeleteStarted, scaleDownStatus.Result)
	assert.Equal(t, n1.Name, utils.GetStringFromChan(deletedNodes))
}

func TestScaleDownNoMove(t *testing.T) {
	var autoscalererr autoscaler_errors.AutoscalerError

	fakeClient := &fake.Clientset{}

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "job",
			Namespace: "default",
			SelfLink:  "/apivs/extensions/v1beta1/namespaces/default/jobs/job",
		},
	}
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Time{})

	// N2 is unready so no pods can be moved there.
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, false, time.Time{})

	p1 := BuildTestPod("p1", 100, 0)
	p1.OwnerReferences = GenerateOwnerReferences(job.Name, "Job", "extensions/v1beta1", "")

	p2 := BuildTestPod("p2", 800, 0)
	p1.Spec.NodeName = "n1"
	p2.Spec.NodeName = "n2"

	fakeClient.Fake.AddReactor("list", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, &apiv1.PodList{Items: []apiv1.Pod{*p1, *p2}}, nil
	})
	fakeClient.Fake.AddReactor("get", "pods", func(action core.Action) (bool, runtime.Object, error) {
		return true, nil, errors.NewNotFound(apiv1.Resource("pod"), "whatever")
	})
	fakeClient.Fake.AddReactor("get", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		getAction := action.(core.GetAction)
		switch getAction.GetName() {
		case n1.Name:
			return true, n1, nil
		case n2.Name:
			return true, n2, nil
		}
		return true, nil, fmt.Errorf("wrong node: %v", getAction.GetName())
	})
	fakeClient.Fake.AddReactor("delete", "pods", func(action core.Action) (bool, runtime.Object, error) {
		t.FailNow()
		return false, nil, nil
	})
	fakeClient.Fake.AddReactor("update", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		t.FailNow()
		return false, nil, nil
	})
	provider := testprovider.NewTestCloudProvider(nil, func(nodeGroup string, node string) error {
		t.FailNow()
		return nil
	})
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng1", n2)
	assert.NotNil(t, provider)

	options := config.AutoscalingOptions{
		ScaleDownUtilizationThreshold: 0.5,
		ScaleDownUnneededTime:         time.Minute,
		ScaleDownUnreadyTime:          time.Hour,
		MaxGracefulTerminationSec:     60,
	}
	jobLister, err := kube_util.NewTestJobLister([]*batchv1.Job{&job})
	assert.NoError(t, err)
	registry := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, nil, nil, jobLister, nil, nil)

	context, err := NewScaleTestAutoscalingContext(options, fakeClient, registry, provider, nil)
	assert.NoError(t, err)

	nodes := []*apiv1.Node{n1, n2}

	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, newBackoff())
	scaleDown := NewScaleDown(&context, clusterStateRegistry)
	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, []*apiv1.Pod{p1, p2})
	autoscalererr = scaleDown.UpdateUnneededNodes(nodes, nodes, time.Now().Add(-5*time.Minute), nil)
	assert.NoError(t, autoscalererr)
	scaleDownStatus, err := scaleDown.TryToScaleDown(time.Now(), nil)
	waitForDeleteToFinish(t, scaleDown)

	assert.NoError(t, err)
	assert.Equal(t, status.ScaleDownNoUnneeded, scaleDownStatus.Result)
}

func getCountOfChan(c chan string) int {
	count := 0
	for {
		select {
		case <-c:
			count++
		default:
			return count
		}
	}
}

func TestCalculateCoresAndMemoryTotal(t *testing.T) {
	nodeConfigs := []nodeConfig{
		{"n1", 2000, 7500 * utils.MiB, 0, true, "ng1"},
		{"n2", 2000, 7500 * utils.MiB, 0, true, "ng1"},
		{"n3", 2000, 7500 * utils.MiB, 0, true, "ng1"},
		{"n4", 12000, 8000 * utils.MiB, 0, true, "ng1"},
		{"n5", 16000, 7500 * utils.MiB, 0, true, "ng1"},
		{"n6", 8000, 6000 * utils.MiB, 0, true, "ng1"},
		{"n7", 6000, 16000 * utils.MiB, 0, true, "ng1"},
	}
	nodes := make([]*apiv1.Node, len(nodeConfigs))
	for i, n := range nodeConfigs {
		node := BuildTestNode(n.name, n.cpu, n.memory)
		SetNodeReadyState(node, n.ready, time.Now())
		nodes[i] = node
	}

	nodes[6].Spec.Taints = []apiv1.Taint{
		{
			Key:    deletetaint.ToBeDeletedTaint,
			Value:  fmt.Sprint(time.Now().Unix()),
			Effect: apiv1.TaintEffectNoSchedule,
		},
	}

	coresTotal, memoryTotal := calculateScaleDownCoresMemoryTotal(nodes, time.Now())

	assert.Equal(t, int64(42), coresTotal)
	assert.Equal(t, int64(44000*utils.MiB), memoryTotal)
}

func TestFilterOutMasters(t *testing.T) {
	nodeConfigs := []nodeConfig{
		{"n1", 2000, 4000, 0, false, "ng1"},
		{"n2", 2000, 4000, 0, true, "ng2"},
		{"n3", 2000, 8000, 0, true, ""}, // real master
		{"n4", 1000, 2000, 0, true, "ng3"},
		{"n5", 1000, 2000, 0, true, "ng3"},
		{"n6", 2000, 8000, 0, true, ""}, // same machine type, no node group, no api server
		{"n7", 2000, 8000, 0, true, ""}, // real master
	}
	nodes := make([]*schedulerframework.NodeInfo, len(nodeConfigs))
	nodeMap := make(map[string]*schedulerframework.NodeInfo, len(nodeConfigs))
	for i, n := range nodeConfigs {
		node := BuildTestNode(n.name, n.cpu, n.memory)
		SetNodeReadyState(node, n.ready, time.Now())
		nodeInfo := schedulerframework.NewNodeInfo()
		err := nodeInfo.SetNode(node)
		assert.NoError(t, err)
		nodes[i] = nodeInfo
		nodeMap[n.name] = nodeInfo
	}

	BuildTestPodWithExtra := func(name, namespace, node string, labels map[string]string) *apiv1.Pod {
		pod := BuildTestPod(name, 100, 200)
		pod.Spec.NodeName = node
		pod.Namespace = namespace
		pod.Labels = labels
		return pod
	}

	pods := []*apiv1.Pod{
		BuildTestPodWithExtra("kube-apiserver-kubernetes-master", "kube-system", "n2", map[string]string{}),                                          // without label
		BuildTestPodWithExtra("kube-apiserver-kubernetes-master", "fake-kube-system", "n6", map[string]string{"component": "kube-apiserver"}),        // wrong namespace
		BuildTestPodWithExtra("kube-apiserver-kubernetes-master", "kube-system", "n3", map[string]string{"component": "kube-apiserver"}),             // real api server
		BuildTestPodWithExtra("hidden-name", "kube-system", "n7", map[string]string{"component": "kube-apiserver"}),                                  // also a real api server
		BuildTestPodWithExtra("kube-apiserver-kubernetes-master", "kube-system", "n1", map[string]string{"component": "kube-apiserver-dev"}),         // wrong label
		BuildTestPodWithExtra("custom-deployment", "custom", "n5", map[string]string{"component": "custom-component", "custom-key": "custom-value"}), // unrelated pod
	}

	for _, pod := range pods {
		if node, found := nodeMap[pod.Spec.NodeName]; found {
			node.AddPod(pod)
		}
	}

	withoutMasters := filterOutMasters(nodes)

	withoutMastersNames := make([]string, len(withoutMasters))
	for i, n := range withoutMasters {
		withoutMastersNames[i] = n.Name
	}
	assertEqualSet(t, []string{"n1", "n2", "n4", "n5", "n6"}, withoutMastersNames)
}

func TestCheckScaleDownDeltaWithinLimits(t *testing.T) {
	type testcase struct {
		limits            scaleDownResourcesLimits
		delta             scaleDownResourcesDelta
		exceededResources []string
	}
	tests := []testcase{
		{
			limits:            scaleDownResourcesLimits{"a": 10},
			delta:             scaleDownResourcesDelta{"a": 10},
			exceededResources: []string{},
		},
		{
			limits:            scaleDownResourcesLimits{"a": 10},
			delta:             scaleDownResourcesDelta{"a": 11},
			exceededResources: []string{"a"},
		},
		{
			limits:            scaleDownResourcesLimits{"a": 10},
			delta:             scaleDownResourcesDelta{"b": 10},
			exceededResources: []string{},
		},
		{
			limits:            scaleDownResourcesLimits{"a": scaleDownLimitUnknown},
			delta:             scaleDownResourcesDelta{"a": 0},
			exceededResources: []string{},
		},
		{
			limits:            scaleDownResourcesLimits{"a": scaleDownLimitUnknown},
			delta:             scaleDownResourcesDelta{"a": 1},
			exceededResources: []string{"a"},
		},
		{
			limits:            scaleDownResourcesLimits{"a": 10, "b": 20, "c": 30},
			delta:             scaleDownResourcesDelta{"a": 11, "b": 20, "c": 31},
			exceededResources: []string{"a", "c"},
		},
	}

	for _, test := range tests {
		checkResult := test.limits.checkScaleDownDeltaWithinLimits(test.delta)
		if len(test.exceededResources) == 0 {
			assert.Equal(t, scaleDownLimitsNotExceeded(), checkResult)
		} else {
			assert.Equal(t, scaleDownLimitsCheckResult{true, test.exceededResources}, checkResult)
		}
	}
}

func getNode(t *testing.T, client kube_client.Interface, name string) *apiv1.Node {
	t.Helper()
	node, err := client.CoreV1().Nodes().Get(ctx.TODO(), name, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to retrieve node %v: %v", name, err)
	}
	return node
}

func hasDeletionCandidateTaint(t *testing.T, client kube_client.Interface, name string) bool {
	t.Helper()
	return deletetaint.HasDeletionCandidateTaint(getNode(t, client, name))
}

func getAllNodes(t *testing.T, client kube_client.Interface) []*apiv1.Node {
	t.Helper()
	nodeList, err := client.CoreV1().Nodes().List(ctx.TODO(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("Failed to retrieve list of nodes: %v", err)
	}
	result := make([]*apiv1.Node, 0, nodeList.Size())
	for _, node := range nodeList.Items {
		result = append(result, node.DeepCopy())
	}
	return result
}

func countDeletionCandidateTaints(t *testing.T, client kube_client.Interface) (total int) {
	t.Helper()
	for _, node := range getAllNodes(t, client) {
		if deletetaint.HasDeletionCandidateTaint(node) {
			total++
		}
	}
	return total
}

func TestSoftTaint(t *testing.T) {
	var err error
	var autoscalererr autoscaler_errors.AutoscalerError

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "job",
			Namespace: "default",
			SelfLink:  "/apivs/batch/v1/namespaces/default/jobs/job",
		},
	}
	n1000 := BuildTestNode("n1000", 1000, 1000)
	SetNodeReadyState(n1000, true, time.Time{})
	n2000 := BuildTestNode("n2000", 2000, 1000)
	SetNodeReadyState(n2000, true, time.Time{})

	p500 := BuildTestPod("p500", 500, 0)
	p700 := BuildTestPod("p700", 700, 0)
	p1200 := BuildTestPod("p1200", 1200, 0)
	p500.Spec.NodeName = "n2000"
	p700.Spec.NodeName = "n1000"
	p1200.Spec.NodeName = "n2000"

	fakeClient := fake.NewSimpleClientset()
	_, err = fakeClient.CoreV1().Nodes().Create(ctx.TODO(), n1000, metav1.CreateOptions{})
	assert.NoError(t, err)
	_, err = fakeClient.CoreV1().Nodes().Create(ctx.TODO(), n2000, metav1.CreateOptions{})
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
		ScaleDownUtilizationThreshold: 0.5,
		ScaleDownUnneededTime:         10 * time.Minute,
		MaxGracefulTerminationSec:     60,
		MaxBulkSoftTaintCount:         1,
		MaxBulkSoftTaintTime:          3 * time.Second,
	}
	jobLister, err := kube_util.NewTestJobLister([]*batchv1.Job{&job})
	assert.NoError(t, err)
	registry := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, nil, nil, jobLister, nil, nil)

	context, err := NewScaleTestAutoscalingContext(options, fakeClient, registry, provider, nil)
	assert.NoError(t, err)

	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, newBackoff())
	scaleDown := NewScaleDown(&context, clusterStateRegistry)

	// Test no superfluous nodes
	nodes := []*apiv1.Node{n1000, n2000}
	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, []*apiv1.Pod{p500, p700, p1200})
	autoscalererr = scaleDown.UpdateUnneededNodes(nodes, nodes, time.Now().Add(-5*time.Minute), nil)
	assert.NoError(t, autoscalererr)
	errs := scaleDown.SoftTaintUnneededNodes(getAllNodes(t, fakeClient))
	assert.Empty(t, errs)
	assert.False(t, hasDeletionCandidateTaint(t, fakeClient, n1000.Name))
	assert.False(t, hasDeletionCandidateTaint(t, fakeClient, n2000.Name))

	// Test one unneeded node
	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, []*apiv1.Pod{p500, p1200})
	autoscalererr = scaleDown.UpdateUnneededNodes(nodes, nodes, time.Now().Add(-5*time.Minute), nil)
	assert.NoError(t, autoscalererr)
	errs = scaleDown.SoftTaintUnneededNodes(getAllNodes(t, fakeClient))
	assert.Empty(t, errs)
	assert.True(t, hasDeletionCandidateTaint(t, fakeClient, n1000.Name))
	assert.False(t, hasDeletionCandidateTaint(t, fakeClient, n2000.Name))

	// Test remove soft taint
	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, []*apiv1.Pod{p500, p700, p1200})
	autoscalererr = scaleDown.UpdateUnneededNodes(nodes, nodes, time.Now().Add(-5*time.Minute), nil)
	assert.NoError(t, autoscalererr)
	errs = scaleDown.SoftTaintUnneededNodes(getAllNodes(t, fakeClient))
	assert.Empty(t, errs)
	assert.False(t, hasDeletionCandidateTaint(t, fakeClient, n1000.Name))
	assert.False(t, hasDeletionCandidateTaint(t, fakeClient, n2000.Name))

	// Test bulk update taint limit
	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, []*apiv1.Pod{})
	autoscalererr = scaleDown.UpdateUnneededNodes(nodes, nodes, time.Now().Add(-5*time.Minute), nil)
	assert.NoError(t, autoscalererr)
	errs = scaleDown.SoftTaintUnneededNodes(getAllNodes(t, fakeClient))
	assert.Empty(t, errs)
	assert.Equal(t, 1, countDeletionCandidateTaints(t, fakeClient))
	errs = scaleDown.SoftTaintUnneededNodes(getAllNodes(t, fakeClient))
	assert.Empty(t, errs)
	assert.Equal(t, 2, countDeletionCandidateTaints(t, fakeClient))

	// Test bulk update untaint limit
	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, []*apiv1.Pod{p500, p700, p1200})
	autoscalererr = scaleDown.UpdateUnneededNodes(nodes, nodes, time.Now().Add(-5*time.Minute), nil)
	assert.NoError(t, autoscalererr)
	errs = scaleDown.SoftTaintUnneededNodes(getAllNodes(t, fakeClient))
	assert.Empty(t, errs)
	assert.Equal(t, 1, countDeletionCandidateTaints(t, fakeClient))
	errs = scaleDown.SoftTaintUnneededNodes(getAllNodes(t, fakeClient))
	assert.Empty(t, errs)
	assert.Equal(t, 0, countDeletionCandidateTaints(t, fakeClient))
}

func TestSoftTaintTimeLimit(t *testing.T) {
	var autoscalererr autoscaler_errors.AutoscalerError

	job := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "job",
			Namespace: "default",
			SelfLink:  "/apivs/batch/v1/namespaces/default/jobs/job",
		},
	}
	n1 := BuildTestNode("n1", 1000, 1000)
	SetNodeReadyState(n1, true, time.Time{})
	n2 := BuildTestNode("n2", 1000, 1000)
	SetNodeReadyState(n2, true, time.Time{})

	p1 := BuildTestPod("p1", 1000, 0)
	p2 := BuildTestPod("p2", 1000, 0)
	p1.Spec.NodeName = "n1"
	p2.Spec.NodeName = "n2"

	currentTime := time.Now()
	updateTime := time.Millisecond
	maxSoftTaintDuration := 1 * time.Second

	// Replace time tracking function
	now = func() time.Time {
		return currentTime
	}
	defer func() {
		now = time.Now
		return
	}()

	fakeClient := fake.NewSimpleClientset()
	_, err := fakeClient.CoreV1().Nodes().Create(ctx.TODO(), n1, metav1.CreateOptions{})
	assert.NoError(t, err)
	_, err = fakeClient.CoreV1().Nodes().Create(ctx.TODO(), n2, metav1.CreateOptions{})
	assert.NoError(t, err)

	// Move time forward when updating
	fakeClient.Fake.PrependReactor("update", "nodes", func(action core.Action) (bool, runtime.Object, error) {
		currentTime = currentTime.Add(updateTime)
		return false, nil, nil
	})

	provider := testprovider.NewTestCloudProvider(nil, nil)
	provider.AddNodeGroup("ng1", 1, 10, 2)
	provider.AddNode("ng1", n1)
	provider.AddNode("ng1", n2)
	assert.NotNil(t, provider)

	options := config.AutoscalingOptions{
		ScaleDownUtilizationThreshold: 0.5,
		ScaleDownUnneededTime:         10 * time.Minute,
		MaxGracefulTerminationSec:     60,
		MaxBulkSoftTaintCount:         10,
		MaxBulkSoftTaintTime:          maxSoftTaintDuration,
	}
	jobLister, err := kube_util.NewTestJobLister([]*batchv1.Job{&job})
	assert.NoError(t, err)
	registry := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, nil, nil, jobLister, nil, nil)

	context, err := NewScaleTestAutoscalingContext(options, fakeClient, registry, provider, nil)
	assert.NoError(t, err)

	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, newBackoff())
	scaleDown := NewScaleDown(&context, clusterStateRegistry)

	// Test bulk taint
	nodes := []*apiv1.Node{n1, n2}
	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, []*apiv1.Pod{})
	autoscalererr = scaleDown.UpdateUnneededNodes(nodes, nodes, time.Now().Add(-5*time.Minute), nil)
	assert.NoError(t, autoscalererr)
	errs := scaleDown.SoftTaintUnneededNodes(getAllNodes(t, fakeClient))
	assert.Empty(t, errs)
	assert.Equal(t, 2, countDeletionCandidateTaints(t, fakeClient))
	assert.True(t, hasDeletionCandidateTaint(t, fakeClient, n1.Name))
	assert.True(t, hasDeletionCandidateTaint(t, fakeClient, n2.Name))

	// Test bulk untaint
	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, []*apiv1.Pod{p1, p2})
	autoscalererr = scaleDown.UpdateUnneededNodes(nodes, nodes, time.Now().Add(-5*time.Minute), nil)
	assert.NoError(t, autoscalererr)
	errs = scaleDown.SoftTaintUnneededNodes(getAllNodes(t, fakeClient))
	assert.Empty(t, errs)
	assert.Equal(t, 0, countDeletionCandidateTaints(t, fakeClient))
	assert.False(t, hasDeletionCandidateTaint(t, fakeClient, n1.Name))
	assert.False(t, hasDeletionCandidateTaint(t, fakeClient, n2.Name))

	updateTime = maxSoftTaintDuration

	// Test duration limit of bulk taint
	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, []*apiv1.Pod{})
	autoscalererr = scaleDown.UpdateUnneededNodes(nodes, nodes, time.Now().Add(-5*time.Minute), nil)
	assert.NoError(t, autoscalererr)
	errs = scaleDown.SoftTaintUnneededNodes(getAllNodes(t, fakeClient))
	assert.Empty(t, errs)
	assert.Equal(t, 1, countDeletionCandidateTaints(t, fakeClient))
	errs = scaleDown.SoftTaintUnneededNodes(getAllNodes(t, fakeClient))
	assert.Empty(t, errs)
	assert.Equal(t, 2, countDeletionCandidateTaints(t, fakeClient))

	// Test duration limit of bulk untaint
	simulator.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, []*apiv1.Pod{p1, p2})
	autoscalererr = scaleDown.UpdateUnneededNodes(nodes, nodes, time.Now().Add(-5*time.Minute), nil)
	assert.NoError(t, autoscalererr)
	errs = scaleDown.SoftTaintUnneededNodes(getAllNodes(t, fakeClient))
	assert.Empty(t, errs)
	assert.Equal(t, 1, countDeletionCandidateTaints(t, fakeClient))
	errs = scaleDown.SoftTaintUnneededNodes(getAllNodes(t, fakeClient))
	assert.Empty(t, errs)
	assert.Equal(t, 0, countDeletionCandidateTaints(t, fakeClient))
}

func TestWaitForDelayDeletion(t *testing.T) {
	type testcase struct {
		name                 string
		timeout              time.Duration
		addAnnotation        bool
		removeAnnotation     bool
		expectCallingGetNode bool
	}
	tests := []testcase{
		{
			name:             "annotation not set",
			timeout:          6 * time.Second,
			addAnnotation:    false,
			removeAnnotation: false,
		},
		{
			name:             "annotation set and removed",
			timeout:          6 * time.Second,
			addAnnotation:    true,
			removeAnnotation: true,
		},
		{
			name:             "annotation set but not removed",
			timeout:          6 * time.Second,
			addAnnotation:    true,
			removeAnnotation: false,
		},
		{
			name:             "timeout is 0 - mechanism disable",
			timeout:          0 * time.Second,
			addAnnotation:    true,
			removeAnnotation: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			node := BuildTestNode("n1", 1000, 10)
			nodeWithAnnotation := BuildTestNode("n1", 1000, 10)
			nodeWithAnnotation.Annotations = map[string]string{DelayDeletionAnnotationPrefix + "ingress": "true"}
			allNodeLister := kubernetes.NewTestNodeLister(nil)
			if test.addAnnotation {
				if test.removeAnnotation {
					allNodeLister.SetNodes([]*apiv1.Node{node})
				} else {
					allNodeLister.SetNodes([]*apiv1.Node{nodeWithAnnotation})
				}
			}
			var err error
			if test.addAnnotation {
				err = waitForDelayDeletion(nodeWithAnnotation, allNodeLister, test.timeout)
			} else {
				err = waitForDelayDeletion(node, allNodeLister, test.timeout)
			}
			assert.NoError(t, err)
		})
	}
}
