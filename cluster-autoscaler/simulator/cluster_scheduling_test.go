/*
Copyright 2026 The Kubernetes Authors.

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

package simulator

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/testsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

// TestTopologySpreadTaintScheduling verifies scheduler behavior in
// the post-taint state that motivates the ghost-node fix in SimulateNodeRemoval.
//
// The CAS simulation tests (SimulateNodeRemoval with default vs Honor policy)
// live in cluster_test.go. This test validates the scheduler-side premise: that
// after CAS taints a node with ToBeDeletedByClusterAutoscaler:NoSchedule, the
// replacement pod cannot be placed (default policy) or can (Honor).
func TestTopologySpreadTaintScheduling(t *testing.T) {
	// --- Setup: 3 nodes with hostname topology, 1 TSC pod per node ---
	node1 := BuildTestNode("node1", 1000, 2000000)
	node2 := BuildTestNode("node2", 1000, 2000000)
	node3 := BuildTestNode("node3", 1000, 2000000)
	node1.Labels["kubernetes.io/hostname"] = "node1"
	node2.Labels["kubernetes.io/hostname"] = "node2"
	node3.Labels["kubernetes.io/hostname"] = "node3"
	SetNodeReadyState(node1, true, time.Time{})
	SetNodeReadyState(node2, true, time.Time{})
	SetNodeReadyState(node3, true, time.Time{})

	ownerRefs := GenerateOwnerReferences("rs", "ReplicaSet", "extensions/v1beta1", "")

	// TSC with default nodeTaintsPolicy (nil → Ignore)
	tscDefault := apiv1.TopologySpreadConstraint{
		MaxSkew:           1,
		TopologyKey:       "kubernetes.io/hostname",
		WhenUnsatisfiable: apiv1.DoNotSchedule,
		LabelSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"app": "topo-app"},
		},
	}

	// TSC with nodeTaintsPolicy=Honor (the fix)
	honor := apiv1.NodeInclusionPolicyHonor
	tscHonor := apiv1.TopologySpreadConstraint{
		MaxSkew:           1,
		TopologyKey:       "kubernetes.io/hostname",
		WhenUnsatisfiable: apiv1.DoNotSchedule,
		LabelSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"app": "topo-app"},
		},
		NodeTaintsPolicy: &honor,
	}

	makePods := func(tsc apiv1.TopologySpreadConstraint) []*apiv1.Pod {
		nodeNames := []string{"node1", "node2", "node3"}
		pods := make([]*apiv1.Pod, len(nodeNames))
		for i, nodeName := range nodeNames {
			pod := BuildTestPod(fmt.Sprintf("pod%d", i+1), 100, 100000)
			pod.Labels = map[string]string{"app": "topo-app"}
			pod.OwnerReferences = ownerRefs
			pod.Spec.NodeName = nodeName
			pod.Spec.TopologySpreadConstraints = []apiv1.TopologySpreadConstraint{tsc}
			pods[i] = pod
		}
		return pods
	}

	// -----------------------------------------------------------------------
	// Post-taint state: replacement pod is UNSCHEDULABLE (default policy).
	//
	// In reality CAS taints node1 with NoSchedule, then drains it. After
	// eviction, the cluster state is:
	//   node1 (tainted, 0 matching pods)
	//   node2 (1 matching pod)
	//   node3 (1 matching pod)
	//
	// With default nodeTaintsPolicy=Ignore, PodTopologySpread still counts
	// node1 as a domain with matchNum=0, so minMatchNum=0.
	// Scheduling on node2: skew = (1+1) - 0 = 2 > maxSkew(1) → REJECT
	// Scheduling on node3: skew = (1+1) - 0 = 2 > maxSkew(1) → REJECT
	// Scheduling on node1: rejected by TaintToleration (NoSchedule)
	//
	// Result: pod is unschedulable on ALL nodes → CAS will trigger scale-up.
	// -----------------------------------------------------------------------
	t.Run("post_taint_scheduling_fails_with_default_nodeTaintsPolicy", func(t *testing.T) {
		// Build tainted node1 (after CAS marks it ToBeDeleted)
		taintedNode1 := node1.DeepCopy()
		taintedNode1.Spec.Taints = []apiv1.Taint{{
			Key:    "ToBeDeletedByClusterAutoscaler",
			Effect: apiv1.TaintEffectNoSchedule,
			Value:  fmt.Sprint(time.Now().Unix()),
		}}

		// Only pods on node2 and node3 survive (pod1 was evicted from node1)
		podsDefault := makePods(tscDefault)
		survivingPods := []*apiv1.Pod{podsDefault[1], podsDefault[2]}

		snapshot := testsnapshot.NewTestSnapshotOrDie(t)
		allNodes := []*apiv1.Node{taintedNode1, node2, node3}
		clustersnapshot.InitializeClusterSnapshotOrDie(t, snapshot, allNodes, survivingPods)

		// Replacement pod created by the RS controller after eviction
		replacementPod := BuildTestPod("pod1-replacement", 100, 100000)
		replacementPod.Labels = map[string]string{"app": "topo-app"}
		replacementPod.Spec.TopologySpreadConstraints = []apiv1.TopologySpreadConstraint{tscDefault}

		_, schedErr := snapshot.SchedulePodOnAnyNodeMatching(replacementPod, func(ni *framework.NodeInfo) bool {
			return true
		})
		assert.NotNil(t, schedErr,
			"replacement pod should be UNSCHEDULABLE with default nodeTaintsPolicy=Ignore: "+
				"tainted node1 is a phantom domain with 0 pods, inflating skew on all real candidates")
	})

	// -----------------------------------------------------------------------
	// Post-taint state: scheduling SUCCEEDS with nodeTaintsPolicy=Honor.
	//
	// When the TSC has nodeTaintsPolicy=Honor, PodTopologySpread excludes
	// node1 (untolerated NoSchedule taint) from domain counting entirely.
	// Only 2 domains remain: node2(1), node3(1).
	// Scheduling on node2: skew = (1+1) - 1 = 1 ≤ maxSkew(1) → ACCEPT
	//
	// This matches CAS's simulation behavior (RemoveNodeInfo), so the
	// scale-down decision and real-world outcome are consistent.
	// -----------------------------------------------------------------------
	t.Run("post_taint_scheduling_succeeds_with_nodeTaintsPolicy_Honor", func(t *testing.T) {
		taintedNode1 := node1.DeepCopy()
		taintedNode1.Spec.Taints = []apiv1.Taint{{
			Key:    "ToBeDeletedByClusterAutoscaler",
			Effect: apiv1.TaintEffectNoSchedule,
			Value:  fmt.Sprint(time.Now().Unix()),
		}}

		podsHonor := makePods(tscHonor)
		survivingPods := []*apiv1.Pod{podsHonor[1], podsHonor[2]}

		snapshot := testsnapshot.NewTestSnapshotOrDie(t)
		allNodes := []*apiv1.Node{taintedNode1, node2, node3}
		clustersnapshot.InitializeClusterSnapshotOrDie(t, snapshot, allNodes, survivingPods)

		replacementPod := BuildTestPod("pod1-replacement", 100, 100000)
		replacementPod.Labels = map[string]string{"app": "topo-app"}
		replacementPod.Spec.TopologySpreadConstraints = []apiv1.TopologySpreadConstraint{tscHonor}

		nodeName, schedErr := snapshot.SchedulePodOnAnyNodeMatching(replacementPod, func(ni *framework.NodeInfo) bool {
			return true
		})
		assert.Nil(t, schedErr,
			"replacement pod should be SCHEDULABLE with nodeTaintsPolicy=Honor: "+
				"tainted node1 is excluded from domain counting, matching CAS simulation")
		assert.NotEmpty(t, nodeName, "should find a node to schedule on")
	})

}
