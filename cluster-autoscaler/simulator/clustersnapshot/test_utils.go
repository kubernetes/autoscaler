/*
Copyright 2020 The Kubernetes Authors.

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

package clustersnapshot

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"
	drasnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/test"
)

// InitializeClusterSnapshotOrDie clears cluster snapshot and then initializes it with given set of nodes and pods.
// Both Spec.NodeName and Status.NominatedNodeName are used when simulating scheduling pods.
func InitializeClusterSnapshotOrDie(
	t *testing.T,
	snapshot ClusterSnapshot,
	nodes []*apiv1.Node,
	pods []*apiv1.Pod) {
	var err error

	assert.NoError(t, snapshot.SetClusterState(nil, nil, drasnapshot.Snapshot{}))

	for _, node := range nodes {
		err = snapshot.AddNodeInfo(framework.NewTestNodeInfo(node))
		assert.NoError(t, err, "error while adding node %s", node.Name)
	}

	for _, pod := range pods {
		if pod.Spec.NodeName != "" {
			err = snapshot.ForceAddPod(pod, pod.Spec.NodeName)
			assert.NoError(t, err, "error while adding pod %s/%s to node %s", pod.Namespace, pod.Name, pod.Spec.NodeName)
		} else if pod.Status.NominatedNodeName != "" {
			err = snapshot.ForceAddPod(pod, pod.Status.NominatedNodeName)
			assert.NoError(t, err, "error while adding pod %s/%s to nominated node %s", pod.Namespace, pod.Name, pod.Status.NominatedNodeName)
		} else {
			assert.Fail(t, "pod %s/%s does not have Spec.NodeName nor Status.NominatedNodeName set", pod.Namespace, pod.Name)
		}
	}
}

// CreateTestNodesWithPrefix creates n test Nodes with the given name prefix.
func CreateTestNodesWithPrefix(prefix string, n int) []*apiv1.Node {
	nodes := make([]*apiv1.Node, n, n)
	for i := 0; i < n; i++ {
		nodes[i] = test.BuildTestNode(fmt.Sprintf("%s-%d", prefix, i), math.MaxInt, math.MaxInt)
		test.SetNodeReadyState(nodes[i], true, time.Time{})
	}
	return nodes
}

// CreateTestNodes creates n test Nodes.
func CreateTestNodes(n int) []*apiv1.Node {
	return CreateTestNodesWithPrefix("n", n)
}

// CreateTestPodsWithPrefix creates n test Pods with the given name prefix.
func CreateTestPodsWithPrefix(prefix string, n int) []*apiv1.Pod {
	pods := make([]*apiv1.Pod, n, n)
	for i := 0; i < n; i++ {
		pods[i] = test.BuildTestPod(fmt.Sprintf("%s-%d", prefix, i), 1, 1)
	}
	return pods
}

// CreateTestPods creates n test Pods.
func CreateTestPods(n int) []*apiv1.Pod {
	return CreateTestPodsWithPrefix("p", n)
}

// AssignTestPodsToNodes distributes test pods evenly across test nodes, and returns a map of the distribution.
func AssignTestPodsToNodes(pods []*apiv1.Pod, nodes []*apiv1.Node) map[string][]*apiv1.Pod {
	if len(nodes) == 0 {
		return nil
	}
	podsByNode := map[string][]*apiv1.Pod{}
	for i := 0; i < len(pods); i++ {
		pod := pods[i]
		nodeName := nodes[i%len(nodes)].Name

		pod.Spec.NodeName = nodeName
		podsByNode[nodeName] = append(podsByNode[nodeName], pod)
	}
	return podsByNode
}
