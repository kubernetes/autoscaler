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

package legacy

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	autoscaler_errors "k8s.io/autoscaler/cluster-autoscaler/utils/errors"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	testprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/actuation"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/deletiontracker"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/eligibility"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/unremovable"
	. "k8s.io/autoscaler/cluster-autoscaler/core/test"
	"k8s.io/autoscaler/cluster-autoscaler/core/utils"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
	"k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/status"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
)

func TestFindUnneededNodes(t *testing.T) {
	var autoscalererr autoscaler_errors.AutoscalerError

	p1 := BuildTestPod("p1", 100, 0)
	p1.Spec.NodeName = "n1"

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")
	replicaSets := generateReplicaSets()

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
		eligibility.ScaleDownDisabledKey: "true",
	}
	// Node info not found.
	n6 := BuildTestNode("n6", 1000, 10)
	// Node without utilization.
	n7 := BuildTestNode("n7", 0, 10)
	// Node being deleted.
	n8 := BuildTestNode("n8", 1000, 10)
	n8.Spec.Taints = []apiv1.Taint{{Key: taints.ToBeDeletedTaint, Value: strconv.FormatInt(time.Now().Unix()-301, 10)}}
	// Nod being deleted recently.
	n9 := BuildTestNode("n9", 1000, 10)
	n9.Spec.Taints = []apiv1.Taint{{Key: taints.ToBeDeletedTaint, Value: strconv.FormatInt(time.Now().Unix()-60, 10)}}

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

	podLister := kube_util.NewTestPodLister([]*apiv1.Pod{p1, p2, p3, p4, p5, p6})
	pdbLister := kube_util.NewTestPodDisruptionBudgetLister([]*policyv1.PodDisruptionBudget{})

	rsLister, err := kube_util.NewTestReplicaSetLister(replicaSets)
	assert.NoError(t, err)
	registry := kube_util.NewListerRegistry(nil, nil, podLister, nil, pdbLister, nil, nil, nil, rsLister, nil)

	options := config.AutoscalingOptions{
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUtilizationThreshold: 0.35,
		},
		UnremovableNodeRecheckTimeout: 5 * time.Minute,
	}
	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, registry, provider, nil, nil)
	assert.NoError(t, err)

	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, NewBackoff())
	wrapper := newWrapperForTesting(&context, clusterStateRegistry, nil)
	sd := wrapper.sd
	allNodes := []*apiv1.Node{n1, n2, n3, n4, n5, n7, n8, n9}

	clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, allNodes, []*apiv1.Pod{p1, p2, p3, p4, p5, p6})
	autoscalererr = sd.UpdateUnneededNodes(allNodes, allNodes, time.Now())
	assert.NoError(t, autoscalererr)

	assert.Equal(t, 3, len(sd.unneededNodes.AsList()))
	assert.True(t, sd.unneededNodes.Contains("n2"))
	assert.True(t, sd.unneededNodes.Contains("n7"))
	assert.True(t, sd.unneededNodes.Contains("n8"))
	for _, n := range []string{"n1", "n2", "n3", "n4", "n7", "n8"} {
		_, found := sd.nodeUtilizationMap[n]
		assert.True(t, found, n)
	}
	for _, n := range []string{"n5", "n6", "n9"} {
		_, found := sd.nodeUtilizationMap[n]
		assert.False(t, found, n)
	}

	sd.unremovableNodes = unremovable.NewNodes()
	sd.unneededNodes.Update([]simulator.NodeToBeRemoved{{Node: n1}, {Node: n2}, {Node: n3}, {Node: n4}}, time.Now())
	allNodes = []*apiv1.Node{n1, n2, n3, n4}
	clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, allNodes, []*apiv1.Pod{p1, p2, p3, p4})
	autoscalererr = sd.UpdateUnneededNodes(allNodes, allNodes, time.Now())
	assert.NoError(t, autoscalererr)

	assert.Equal(t, 1, len(sd.unneededNodes.AsList()))
	assert.True(t, sd.unneededNodes.Contains("n2"))
	for _, n := range []string{"n1", "n2", "n3", "n4"} {
		_, found := sd.nodeUtilizationMap[n]
		assert.True(t, found, n)
	}
	for _, n := range []string{"n5", "n6"} {
		_, found := sd.nodeUtilizationMap[n]
		assert.False(t, found, n)
	}

	sd.unremovableNodes = unremovable.NewNodes()
	scaleDownCandidates := []*apiv1.Node{n1, n3, n4}
	clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, allNodes, []*apiv1.Pod{p1, p2, p3, p4})
	autoscalererr = sd.UpdateUnneededNodes(allNodes, scaleDownCandidates, time.Now())
	assert.NoError(t, autoscalererr)

	assert.Equal(t, 0, len(sd.unneededNodes.AsList()))

	// Node n1 is unneeded, but should be skipped because it has just recently been found to be unremovable
	allNodes = []*apiv1.Node{n1}
	clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, allNodes, []*apiv1.Pod{})
	autoscalererr = sd.UpdateUnneededNodes(allNodes, allNodes, time.Now())
	assert.NoError(t, autoscalererr)

	assert.Equal(t, 0, len(sd.unneededNodes.AsList()))
	// Verify that no other nodes are in unremovable map.
	assert.Equal(t, 1, len(sd.unremovableNodes.AsList()))

	// But it should be checked after timeout
	clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, allNodes, []*apiv1.Pod{})
	autoscalererr = sd.UpdateUnneededNodes(allNodes, allNodes, time.Now().Add(context.UnremovableNodeRecheckTimeout+time.Second))
	assert.NoError(t, autoscalererr)

	assert.Equal(t, 1, len(sd.unneededNodes.AsList()))
	// Verify that nodes that are no longer unremovable are removed.
	assert.Equal(t, 0, len(sd.unremovableNodes.AsList()))
}

func TestFindUnneededGPUNodes(t *testing.T) {
	var autoscalererr autoscaler_errors.AutoscalerError

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")
	replicaSets := generateReplicaSets()

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

	podLister := kube_util.NewTestPodLister([]*apiv1.Pod{p1, p2, p3})
	pdbLister := kube_util.NewTestPodDisruptionBudgetLister([]*policyv1.PodDisruptionBudget{})

	rsLister, err := kube_util.NewTestReplicaSetLister(replicaSets)
	assert.NoError(t, err)
	registry := kube_util.NewListerRegistry(nil, nil, podLister, nil, pdbLister, nil, nil, nil, rsLister, nil)

	options := config.AutoscalingOptions{
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUtilizationThreshold:    0.35,
			ScaleDownGpuUtilizationThreshold: 0.3,
		},
		UnremovableNodeRecheckTimeout: 5 * time.Minute,
	}
	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, registry, provider, nil, nil)
	assert.NoError(t, err)

	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, NewBackoff())
	wrapper := newWrapperForTesting(&context, clusterStateRegistry, nil)
	sd := wrapper.sd
	allNodes := []*apiv1.Node{n1, n2, n3}

	clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, allNodes, []*apiv1.Pod{p1, p2, p3})

	autoscalererr = sd.UpdateUnneededNodes(allNodes, allNodes, time.Now())
	assert.NoError(t, autoscalererr)
	assert.Equal(t, 1, len(sd.unneededNodes.AsList()))
	assert.True(t, sd.unneededNodes.Contains("n2"))
	for _, n := range []string{"n1", "n2", "n3"} {
		_, found := sd.nodeUtilizationMap[n]
		assert.True(t, found, n)
	}
}

func TestFindUnneededWithPerNodeGroupThresholds(t *testing.T) {
	var autoscalererr autoscaler_errors.AutoscalerError

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "apps/v1", "")
	replicaSets := generateReplicaSets()

	provider := testprovider.NewTestCloudProvider(nil, nil)

	// this test focuses on utilization checks
	// add a super large node, so every pod always has a place to drain
	sink := BuildTestNode("sink", 100000, 100000)
	AddGpusToNode(sink, 20)
	SetNodeReadyState(sink, true, time.Time{})
	provider.AddNodeGroup("sink_group", 1, 1, 1)
	provider.AddNode("sink_group", sink)

	allNodes := []*apiv1.Node{sink}
	scaleDownCandidates := []*apiv1.Node{}
	allPods := []*apiv1.Pod{}

	// set up 2 node groups with nodes with different utilizations
	cpuUtilizations := []int64{30, 40, 50, 60, 90}
	for i := 1; i < 3; i++ {
		ngName := fmt.Sprintf("n%d", i)
		provider.AddNodeGroup(ngName, 0, len(cpuUtilizations), len(cpuUtilizations))
		for _, u := range cpuUtilizations {
			nodeName := fmt.Sprintf("%s_%d", ngName, u)
			node := BuildTestNode(nodeName, 1000, 10)
			SetNodeReadyState(node, true, time.Time{})
			provider.AddNode(ngName, node)
			allNodes = append(allNodes, node)
			scaleDownCandidates = append(scaleDownCandidates, node)
			pod := BuildTestPod(fmt.Sprintf("p_%s", nodeName), u*10, 0)
			pod.Spec.NodeName = nodeName
			pod.OwnerReferences = ownerRef
			allPods = append(allPods, pod)
		}
	}

	globalOptions := config.AutoscalingOptions{
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUtilizationThreshold:    0.5,
			ScaleDownGpuUtilizationThreshold: 0.5,
		},
	}

	cases := map[string]struct {
		n1opts       *config.NodeGroupAutoscalingOptions
		n2opts       *config.NodeGroupAutoscalingOptions
		wantUnneeded []string
	}{
		"no per NodeGroup config": {
			wantUnneeded: []string{"n1_30", "n1_40", "n2_30", "n2_40"},
		},
		"one group has higher threshold": {
			n1opts: &config.NodeGroupAutoscalingOptions{
				ScaleDownUtilizationThreshold: 0.75,
			},
			wantUnneeded: []string{"n1_30", "n1_40", "n1_50", "n1_60", "n2_30", "n2_40"},
		},
		"one group has lower gpu threshold (which should be ignored)": {
			n1opts: &config.NodeGroupAutoscalingOptions{
				ScaleDownUtilizationThreshold:    0.75,
				ScaleDownGpuUtilizationThreshold: 0.1,
			},
			wantUnneeded: []string{"n1_30", "n1_40", "n1_50", "n1_60", "n2_30", "n2_40"},
		},
		"both group have different thresholds": {
			n1opts: &config.NodeGroupAutoscalingOptions{
				ScaleDownUtilizationThreshold: 0.75,
			},
			n2opts: &config.NodeGroupAutoscalingOptions{
				ScaleDownUtilizationThreshold: 0.55,
			},
			wantUnneeded: []string{"n1_30", "n1_40", "n1_50", "n1_60", "n2_30", "n2_40", "n2_50"},
		},
		"both group have the same custom threshold": {
			n1opts: &config.NodeGroupAutoscalingOptions{
				ScaleDownUtilizationThreshold: 0.35,
			},
			n2opts: &config.NodeGroupAutoscalingOptions{
				ScaleDownUtilizationThreshold: 0.35,
			},
			wantUnneeded: []string{"n1_30", "n2_30"},
		},
	}
	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			podLister := kube_util.NewTestPodLister(allPods)
			pdbLister := kube_util.NewTestPodDisruptionBudgetLister([]*policyv1.PodDisruptionBudget{})

			rsLister, err := kube_util.NewTestReplicaSetLister(replicaSets)
			assert.NoError(t, err)
			registry := kube_util.NewListerRegistry(nil, nil, podLister, nil, pdbLister, nil, nil, nil, rsLister, nil)

			context, err := NewScaleTestAutoscalingContext(globalOptions, &fake.Clientset{}, registry, provider, nil, nil)
			assert.NoError(t, err)
			clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, NewBackoff())
			wrapper := newWrapperForTesting(&context, clusterStateRegistry, nil)
			sd := wrapper.sd
			clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, allNodes, allPods)

			ng1 := provider.GetNodeGroup("n1").(*testprovider.TestNodeGroup)
			ng1.SetOptions(tc.n1opts)
			ng2 := provider.GetNodeGroup("n2").(*testprovider.TestNodeGroup)
			ng2.SetOptions(tc.n2opts)

			autoscalererr = sd.UpdateUnneededNodes(allNodes, scaleDownCandidates, time.Now())
			assert.NoError(t, autoscalererr)
			assert.Equal(t, len(tc.wantUnneeded), len(sd.unneededNodes.AsList()))
			for _, node := range tc.wantUnneeded {
				assert.True(t, sd.unneededNodes.Contains(node))
			}
		})
	}
}

func TestPodsWithPreemptionsFindUnneededNodes(t *testing.T) {
	var autoscalererr autoscaler_errors.AutoscalerError

	// shared owner reference
	ownerRef := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")
	replicaSets := generateReplicaSets()
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

	podLister := kube_util.NewTestPodLister([]*apiv1.Pod{})
	pdbLister := kube_util.NewTestPodDisruptionBudgetLister([]*policyv1.PodDisruptionBudget{})

	rsLister, err := kube_util.NewTestReplicaSetLister(replicaSets)
	assert.NoError(t, err)
	registry := kube_util.NewListerRegistry(nil, nil, podLister, nil, pdbLister, nil, nil, nil, rsLister, nil)

	options := config.AutoscalingOptions{
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUtilizationThreshold: 0.35,
		},
	}
	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, registry, provider, nil, nil)
	assert.NoError(t, err)

	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, NewBackoff())
	wrapper := newWrapperForTesting(&context, clusterStateRegistry, nil)
	sd := wrapper.sd

	allNodes := []*apiv1.Node{n1, n2, n3, n4}
	clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, allNodes, []*apiv1.Pod{p1, p2, p3, p4})
	autoscalererr = sd.UpdateUnneededNodes(allNodes, allNodes, time.Now())
	assert.NoError(t, autoscalererr)
	assert.Equal(t, 2, len(sd.unneededNodes.AsList()))
	assert.True(t, sd.unneededNodes.Contains("n2"))
	assert.True(t, sd.unneededNodes.Contains("n3"))
	for _, n := range []string{"n1", "n2", "n3", "n4"} {
		_, found := sd.nodeUtilizationMap[n]
		assert.True(t, found, n)
	}
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
	replicaSets := generateReplicaSets()

	pods := make([]*apiv1.Pod, 0, numNodes)
	for i := 0; i < numNodes; i++ {
		p := BuildTestPod(fmt.Sprintf("p%v", i), 100, 0)
		p.Spec.NodeName = fmt.Sprintf("n%v", i)
		p.OwnerReferences = ownerRef
		pods = append(pods, p)
	}

	numCandidates := 30

	podLister := kube_util.NewTestPodLister(pods)
	pdbLister := kube_util.NewTestPodDisruptionBudgetLister([]*policyv1.PodDisruptionBudget{})

	rsLister, err := kube_util.NewTestReplicaSetLister(replicaSets)
	assert.NoError(t, err)
	registry := kube_util.NewListerRegistry(nil, nil, podLister, nil, pdbLister, nil, nil, nil, rsLister, nil)

	options := config.AutoscalingOptions{
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUtilizationThreshold: 0.35,
		},
		ScaleDownNonEmptyCandidatesCount: numCandidates,
		ScaleDownCandidatesPoolRatio:     1,
		ScaleDownCandidatesPoolMinCount:  1000,
	}
	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, registry, provider, nil, nil)
	assert.NoError(t, err)

	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, NewBackoff())
	wrapper := newWrapperForTesting(&context, clusterStateRegistry, nil)
	sd := wrapper.sd

	clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, pods)
	autoscalererr = sd.UpdateUnneededNodes(nodes, nodes, time.Now())
	assert.NoError(t, autoscalererr)
	assert.Equal(t, numCandidates, len(sd.unneededNodes.AsList()))
	// Simulate one of the unneeded nodes got deleted
	deleted := sd.unneededNodes.AsList()[len(sd.unneededNodes.AsList())-1]
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

	clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, pods)
	autoscalererr = sd.UpdateUnneededNodes(nodes, nodes, time.Now())
	assert.NoError(t, autoscalererr)
	// Check that the deleted node was replaced
	assert.Equal(t, numCandidates, len(sd.unneededNodes.AsList()))
	assert.False(t, sd.unneededNodes.Contains(deleted.Name))
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
	replicaSets := generateReplicaSets()

	pods := make([]*apiv1.Pod, 0, numNodes)
	for i := 0; i < numNodes-numEmpty; i++ {
		p := BuildTestPod(fmt.Sprintf("p%v", i), 100, 0)
		p.Spec.NodeName = fmt.Sprintf("n%v", i)
		p.OwnerReferences = ownerRef
		pods = append(pods, p)
	}

	numCandidates := 30

	podLister := kube_util.NewTestPodLister(pods)
	pdbLister := kube_util.NewTestPodDisruptionBudgetLister([]*policyv1.PodDisruptionBudget{})

	rsLister, err := kube_util.NewTestReplicaSetLister(replicaSets)
	assert.NoError(t, err)
	registry := kube_util.NewListerRegistry(nil, nil, podLister, nil, pdbLister, nil, nil, nil, rsLister, nil)

	options := config.AutoscalingOptions{
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUtilizationThreshold: 0.35,
		},
		ScaleDownNonEmptyCandidatesCount: numCandidates,
		ScaleDownCandidatesPoolRatio:     1.0,
		ScaleDownCandidatesPoolMinCount:  1000,
	}
	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, registry, provider, nil, nil)
	assert.NoError(t, err)

	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, NewBackoff())
	wrapper := newWrapperForTesting(&context, clusterStateRegistry, nil)
	sd := wrapper.sd

	clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, pods)
	autoscalererr = sd.UpdateUnneededNodes(nodes, nodes, time.Now())
	assert.NoError(t, autoscalererr)
	assert.Equal(t, numEmpty+numCandidates, len(sd.unneededNodes.AsList()))
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
	replicaSets := generateReplicaSets()

	pods := make([]*apiv1.Pod, 0, numNodes)
	for i := 0; i < numNodes; i++ {
		p := BuildTestPod(fmt.Sprintf("p%v", i), 100, 0)
		p.Spec.NodeName = fmt.Sprintf("n%v", i)
		p.OwnerReferences = ownerRef
		pods = append(pods, p)
	}

	numCandidates := 30

	podLister := kube_util.NewTestPodLister(pods)
	pdbLister := kube_util.NewTestPodDisruptionBudgetLister([]*policyv1.PodDisruptionBudget{})

	rsLister, err := kube_util.NewTestReplicaSetLister(replicaSets)
	assert.NoError(t, err)
	registry := kube_util.NewListerRegistry(nil, nil, podLister, nil, pdbLister, nil, nil, nil, rsLister, nil)

	options := config.AutoscalingOptions{
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUtilizationThreshold: 0.35,
		},
		ScaleDownNonEmptyCandidatesCount: numCandidates,
		ScaleDownCandidatesPoolRatio:     0.1,
		ScaleDownCandidatesPoolMinCount:  10,
	}
	context, err := NewScaleTestAutoscalingContext(options, &fake.Clientset{}, registry, provider, nil, nil)
	assert.NoError(t, err)

	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, NewBackoff())
	wrapper := newWrapperForTesting(&context, clusterStateRegistry, nil)
	sd := wrapper.sd
	clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, pods)
	autoscalererr = sd.UpdateUnneededNodes(nodes, nodes, time.Now())
	assert.NoError(t, autoscalererr)
	assert.NotEmpty(t, sd.unneededNodes)
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
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUnneededTime:         time.Minute,
			ScaleDownUtilizationThreshold: 0.5,
		},
		MaxGracefulTerminationSec: 60,
	}
	jobLister, err := kube_util.NewTestJobLister([]*batchv1.Job{&job})
	assert.NoError(t, err)
	podLister := kube_util.NewTestPodLister([]*apiv1.Pod{p1, p2})
	pdbLister := kube_util.NewTestPodDisruptionBudgetLister([]*policyv1.PodDisruptionBudget{})

	registry := kube_util.NewListerRegistry(nil, nil, podLister, nil, pdbLister, nil, nil, jobLister, nil, nil)

	context, err := NewScaleTestAutoscalingContext(options, fakeClient, registry, provider, nil, nil)
	assert.NoError(t, err)
	nodes := []*apiv1.Node{n1, n2}

	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, NewBackoff())
	wrapper := newWrapperForTesting(&context, clusterStateRegistry, nil)
	clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, []*apiv1.Pod{p1, p2})
	autoscalererr = wrapper.UpdateClusterState(nodes, nodes, nil, time.Now().Add(-5*time.Minute))
	assert.NoError(t, autoscalererr)
	empty, drain := wrapper.NodesToDelete(time.Now())
	scaleDownStatus, err := wrapper.StartDeletion(empty, drain)
	waitForDeleteToFinish(t, wrapper)
	assert.NoError(t, err)
	assert.Equal(t, status.ScaleDownNodeDeleteStarted, scaleDownStatus.Result)
	assert.Equal(t, n1.Name, utils.GetStringFromChan(deletedNodes))
	assert.Equal(t, n1.Name, utils.GetStringFromChan(updatedNodes))
}

func waitForDeleteToFinish(t *testing.T, wrapper *ScaleDownWrapper) {
	for start := time.Now(); time.Since(start) < 20*time.Second; time.Sleep(100 * time.Millisecond) {
		_, drained := wrapper.CheckStatus().DeletionsInProgress()
		if len(drained) == 0 {
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
	NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
		ScaleDownUnneededTime:            time.Minute,
		ScaleDownUtilizationThreshold:    0.5,
		ScaleDownGpuUtilizationThreshold: 0.5,
	},
	MaxGracefulTerminationSec: 60,
	MaxEmptyBulkDelete:        10,
	MinCoresTotal:             0,
	MinMemoryTotal:            0,
	MaxCoresTotal:             config.DefaultMaxClusterCores,
	MaxMemoryTotal:            config.DefaultMaxClusterMemory * units.GiB,
}

func TestScaleDownEmptyMultipleNodeGroups(t *testing.T) {
	config := &ScaleTestConfig{
		Nodes: []NodeConfig{
			{"n1_1", 1000, 1000, 0, true, "ng1"},
			{"n1_2", 1000, 1000, 0, true, "ng1"},
			{"n2_1", 1000, 1000, 0, true, "ng2"},
			{"n2_2", 1000, 1000, 0, true, "ng2"},
		},
		Options:                defaultScaleDownOptions,
		ExpectedScaleDowns:     []string{"n1_1", "n1_2", "n2_1", "n2_2"},
		ExpectedScaleDownCount: 2,
	}
	simpleScaleDownEmpty(t, config)
}

func TestScaleDownEmptySingleNodeGroup(t *testing.T) {
	config := &ScaleTestConfig{
		Nodes: []NodeConfig{
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
		Options:                defaultScaleDownOptions,
		ExpectedScaleDowns:     []string{"n1", "n2", "n3", "n4", "n5", "n6", "n7", "n8", "n9", "n10", "n11", "n12"},
		ExpectedScaleDownCount: 10,
	}
	simpleScaleDownEmpty(t, config)
}

func TestScaleDownEmptyMinCoresLimitHit(t *testing.T) {
	options := defaultScaleDownOptions
	options.MinCoresTotal = 2
	config := &ScaleTestConfig{
		Nodes: []NodeConfig{
			{"n1", 2000, 1000, 0, true, "ng1"},
			{"n2", 1000, 1000, 0, true, "ng1"},
		},
		Options:            options,
		ExpectedScaleDowns: []string{"n2"},
	}
	simpleScaleDownEmpty(t, config)
}

func TestScaleDownEmptyMinMemoryLimitHit(t *testing.T) {
	options := defaultScaleDownOptions
	options.MinMemoryTotal = 4000 * utils.MiB
	config := &ScaleTestConfig{
		Nodes: []NodeConfig{
			{"n1", 2000, 1000 * utils.MiB, 0, true, "ng1"},
			{"n2", 1000, 1000 * utils.MiB, 0, true, "ng1"},
			{"n3", 1000, 1000 * utils.MiB, 0, true, "ng1"},
			{"n4", 1000, 3000 * utils.MiB, 0, true, "ng1"},
		},
		Options:                options,
		ExpectedScaleDowns:     []string{"n1", "n2", "n3"},
		ExpectedScaleDownCount: 2,
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
	config := &ScaleTestConfig{
		Nodes: []NodeConfig{
			{"n1", 1000, 1000 * utils.MiB, 1, true, "ng1"},
			{"n2", 1000, 1000 * utils.MiB, 1, true, "ng1"},
			{"n3", 1000, 1000 * utils.MiB, 1, true, "ng1"},
			{"n4", 1000, 1000 * utils.MiB, 1, true, "ng1"},
			{"n5", 1000, 1000 * utils.MiB, 1, true, "ng1"},
			{"n6", 1000, 1000 * utils.MiB, 1, true, "ng1"},
		},
		Options:                options,
		ExpectedScaleDowns:     []string{"n1", "n2", "n3", "n4", "n5", "n6"},
		ExpectedScaleDownCount: 2,
	}
	simpleScaleDownEmpty(t, config)
}

func TestScaleDownEmptyMinGroupSizeLimitHit(t *testing.T) {
	options := defaultScaleDownOptions
	config := &ScaleTestConfig{
		Nodes: []NodeConfig{
			{"n1", 2000, 1000, 0, true, "ng1"},
		},
		Options:            options,
		ExpectedScaleDowns: []string{},
	}
	simpleScaleDownEmpty(t, config)
}

func TestScaleDownEmptyMinGroupSizeLimitHitWhenOneNodeIsBeingDeleted(t *testing.T) {
	nodeDeletionTracker := deletiontracker.NewNodeDeletionTracker(0 * time.Second)
	nodeDeletionTracker.StartDeletion("ng1", "n1")
	nodeDeletionTracker.StartDeletion("ng1", "n2")
	options := defaultScaleDownOptions
	config := &ScaleTestConfig{
		Nodes: []NodeConfig{
			{"n1", 2000, 1000, 0, true, "ng1"},
			{"n2", 2000, 1000, 0, true, "ng1"},
			{"n3", 2000, 1000, 0, true, "ng1"},
		},
		Options:             options,
		ExpectedScaleDowns:  []string{},
		NodeDeletionTracker: nodeDeletionTracker,
	}
	simpleScaleDownEmpty(t, config)
}

func simpleScaleDownEmpty(t *testing.T, config *ScaleTestConfig) {
	var autoscalererr autoscaler_errors.AutoscalerError

	updatedNodes := make(chan string, 30)
	deletedNodes := make(chan string, 30)
	fakeClient := &fake.Clientset{}

	nodes := make([]*apiv1.Node, len(config.Nodes))
	nodesMap := make(map[string]*apiv1.Node)
	groups := make(map[string][]*apiv1.Node)

	provider := testprovider.NewTestCloudProvider(nil, func(nodeGroup string, node string) error {
		deletedNodes <- node
		return nil
	})

	for i, n := range config.Nodes {
		node := BuildTestNode(n.Name, n.Cpu, n.Memory)
		if n.Gpu > 0 {
			AddGpusToNode(node, n.Gpu)
			node.Labels[provider.GPULabel()] = gpu.DefaultGPUType
		}
		SetNodeReadyState(node, n.Ready, time.Time{})
		nodesMap[n.Name] = node
		nodes[i] = node
		groups[n.Group] = append(groups[n.Group], node)
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

	resourceLimiter := context.NewResourceLimiterFromAutoscalingOptions(config.Options)
	provider.SetResourceLimiter(resourceLimiter)

	assert.NotNil(t, provider)

	podLister := kube_util.NewTestPodLister([]*apiv1.Pod{})
	pdbLister := kube_util.NewTestPodDisruptionBudgetLister([]*policyv1.PodDisruptionBudget{})

	registry := kube_util.NewListerRegistry(nil, nil, podLister, nil, pdbLister, nil, nil, nil, nil, nil)
	context, err := NewScaleTestAutoscalingContext(config.Options, fakeClient, registry, provider, nil, nil)
	assert.NoError(t, err)

	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, NewBackoff())
	wrapper := newWrapperForTesting(&context, clusterStateRegistry, config.NodeDeletionTracker)
	clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, []*apiv1.Pod{})
	autoscalererr = wrapper.UpdateClusterState(nodes, nodes, nil, time.Now().Add(-5*time.Minute))
	assert.NoError(t, autoscalererr)
	empty, drain := wrapper.NodesToDelete(time.Now())
	scaleDownStatus, err := wrapper.StartDeletion(empty, drain)

	assert.NoError(t, err)
	var expectedScaleDownResult status.ScaleDownResult
	if len(config.ExpectedScaleDowns) > 0 {
		expectedScaleDownResult = status.ScaleDownNodeDeleteStarted
	} else {
		expectedScaleDownResult = status.ScaleDownNoUnneeded
	}
	assert.Equal(t, expectedScaleDownResult, scaleDownStatus.Result)

	expectedScaleDownCount := config.ExpectedScaleDownCount
	if config.ExpectedScaleDownCount == 0 {
		// For backwards compatibility.
		expectedScaleDownCount = len(config.ExpectedScaleDowns)
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
	assert.Subset(t, config.ExpectedScaleDowns, deleted)
	_, nonEmptyDeletions := wrapper.CheckStatus().DeletionsInProgress()
	assert.Equal(t, 0, len(nonEmptyDeletions))
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
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUnneededTime:         time.Minute,
			ScaleDownUtilizationThreshold: 0.5,
			ScaleDownUnreadyTime:          time.Hour,
		},
		MaxGracefulTerminationSec: 60,
		ScaleDownUnreadyEnabled:   true,
	}

	podLister := kube_util.NewTestPodLister([]*apiv1.Pod{p2})
	pdbLister := kube_util.NewTestPodDisruptionBudgetLister([]*policyv1.PodDisruptionBudget{})

	registry := kube_util.NewListerRegistry(nil, nil, podLister, nil, pdbLister, nil, nil, nil, nil, nil)
	context, err := NewScaleTestAutoscalingContext(options, fakeClient, registry, provider, nil, nil)
	assert.NoError(t, err)

	nodes := []*apiv1.Node{n1, n2}

	// N1 is unready so it requires a bigger unneeded time.
	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, NewBackoff())
	wrapper := newWrapperForTesting(&context, clusterStateRegistry, nil)
	clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, []*apiv1.Pod{p2})
	autoscalererr = wrapper.UpdateClusterState(nodes, nodes, nil, time.Now().Add(-5*time.Minute))
	assert.NoError(t, autoscalererr)
	empty, drain := wrapper.NodesToDelete(time.Now())
	scaleDownStatus, err := wrapper.StartDeletion(empty, drain)
	waitForDeleteToFinish(t, wrapper)

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
	wrapper = newWrapperForTesting(&context, clusterStateRegistry, nil)
	clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, []*apiv1.Pod{p2})
	autoscalererr = wrapper.UpdateClusterState(nodes, nodes, nil, time.Now().Add(-2*time.Hour))
	assert.NoError(t, autoscalererr)
	empty, drain = wrapper.NodesToDelete(time.Now())
	scaleDownStatus, err = wrapper.StartDeletion(empty, drain)
	waitForDeleteToFinish(t, wrapper)

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
		NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
			ScaleDownUnneededTime:         time.Minute,
			ScaleDownUnreadyTime:          time.Hour,
			ScaleDownUtilizationThreshold: 0.5,
		},
		MaxGracefulTerminationSec: 60,
	}
	jobLister, err := kube_util.NewTestJobLister([]*batchv1.Job{&job})
	assert.NoError(t, err)
	registry := kube_util.NewListerRegistry(nil, nil, nil, nil, nil, nil, nil, jobLister, nil, nil)

	context, err := NewScaleTestAutoscalingContext(options, fakeClient, registry, provider, nil, nil)
	assert.NoError(t, err)

	nodes := []*apiv1.Node{n1, n2}

	clusterStateRegistry := clusterstate.NewClusterStateRegistry(provider, clusterstate.ClusterStateRegistryConfig{}, context.LogRecorder, NewBackoff())
	wrapper := newWrapperForTesting(&context, clusterStateRegistry, nil)
	clustersnapshot.InitializeClusterSnapshotOrDie(t, context.ClusterSnapshot, nodes, []*apiv1.Pod{p1, p2})
	autoscalererr = wrapper.UpdateClusterState(nodes, nodes, nil, time.Now().Add(-5*time.Minute))
	assert.NoError(t, autoscalererr)
	empty, drain := wrapper.NodesToDelete(time.Now())
	scaleDownStatus, err := wrapper.StartDeletion(empty, drain)
	waitForDeleteToFinish(t, wrapper)

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

func generateReplicaSets() []*appsv1.ReplicaSet {
	replicas := int32(5)
	return []*appsv1.ReplicaSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rs",
				Namespace: "default",
				SelfLink:  "api/v1/namespaces/default/replicasets/rs",
			},
			Spec: appsv1.ReplicaSetSpec{
				Replicas: &replicas,
			},
		},
	}
}

func newWrapperForTesting(ctx *context.AutoscalingContext, clusterStateRegistry *clusterstate.ClusterStateRegistry, ndt *deletiontracker.NodeDeletionTracker) *ScaleDownWrapper {
	ctx.MaxDrainParallelism = 1
	ctx.MaxScaleDownParallelism = 10
	ctx.NodeDeletionBatcherInterval = 0 * time.Second
	ctx.NodeDeleteDelayAfterTaint = 0 * time.Second
	if ndt == nil {
		ndt = deletiontracker.NewNodeDeletionTracker(0 * time.Second)
	}
	deleteOptions := simulator.NodeDeleteOptions{
		SkipNodesWithSystemPods:   true,
		SkipNodesWithLocalStorage: true,
		MinReplicaCount:           0,
	}
	sd := NewScaleDown(ctx, NewTestProcessors(ctx), ndt, deleteOptions)
	actuator := actuation.NewActuator(ctx, clusterStateRegistry, ndt, deleteOptions)
	return NewScaleDownWrapper(sd, actuator)
}
