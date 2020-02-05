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

package simulator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"
)

var snapshots = map[string]func() ClusterSnapshot{
	"basic": func() ClusterSnapshot { return NewBasicClusterSnapshot() },
	"delta": func() ClusterSnapshot { return NewDeltaClusterSnapshot() },
}

func nodeNames(nodes []*apiv1.Node) []string {
	names := make([]string, len(nodes), len(nodes))
	for i, node := range nodes {
		names[i] = node.Name
	}
	return names
}

func nodeInfoNames(nodeInfos []*schedulernodeinfo.NodeInfo) []string {
	names := make([]string, len(nodeInfos), len(nodeInfos))
	for i, node := range nodeInfos {
		names[i] = node.Node().Name
	}
	return names
}

func nodeInfoPods(nodeInfos []*schedulernodeinfo.NodeInfo) []*apiv1.Pod {
	pods := []*apiv1.Pod{}
	for _, node := range nodeInfos {
		pods = append(pods, node.Pods()...)
	}
	return pods
}

func TestForkAddNode(t *testing.T) {
	nodeCount := 3

	nodes := createTestNodes(nodeCount)
	extraNodes := createTestNodesWithPrefix("tmp", 2)

	for name, snapshotFactory := range snapshots {
		t.Run(fmt.Sprintf("%s: fork should not affect base data: adding nodes", name),
			func(t *testing.T) {
				clusterSnapshot := snapshotFactory()
				err := clusterSnapshot.AddNodes(nodes)
				assert.NoError(t, err)

				err = clusterSnapshot.Fork()
				assert.NoError(t, err)

				for _, node := range extraNodes {
					err = clusterSnapshot.AddNode(node)
					assert.NoError(t, err)
				}
				forkNodes, err := clusterSnapshot.NodeInfos().List()
				assert.NoError(t, err)
				assert.ElementsMatch(t, append(nodeNames(nodes), nodeNames(extraNodes)...), nodeInfoNames(forkNodes))

				err = clusterSnapshot.Revert()
				assert.NoError(t, err)

				baseNodes, err := clusterSnapshot.NodeInfos().List()
				assert.NoError(t, err)
				assert.ElementsMatch(t, nodeNames(nodes), nodeInfoNames(baseNodes))
			})
	}
}

func TestForkAddPods(t *testing.T) {
	nodeCount := 3
	podCount := 90

	nodes := createTestNodes(nodeCount)
	pods := createTestPods(podCount)
	assignPodsToNodes(pods, nodes)

	for name, snapshotFactory := range snapshots {
		t.Run(fmt.Sprintf("%s: fork should not affect base data: adding pods", name),
			func(t *testing.T) {
				clusterSnapshot := snapshotFactory()
				err := clusterSnapshot.AddNodes(nodes)
				assert.NoError(t, err)

				err = clusterSnapshot.Fork()
				assert.NoError(t, err)

				for _, pod := range pods {
					err = clusterSnapshot.AddPod(pod, pod.Spec.NodeName)
					assert.NoError(t, err)
				}
				forkPods, err := clusterSnapshot.Pods().List(labels.Everything())
				assert.NoError(t, err)
				assert.ElementsMatch(t, pods, forkPods)
				forkNodes, err := clusterSnapshot.NodeInfos().List()
				assert.NoError(t, err)
				assert.ElementsMatch(t, nodeNames(nodes), nodeInfoNames(forkNodes))

				err = clusterSnapshot.Revert()
				assert.NoError(t, err)

				basePods, err := clusterSnapshot.Pods().List(labels.Everything())
				assert.NoError(t, err)
				assert.Equal(t, 0, len(basePods))
				baseNodes, err := clusterSnapshot.NodeInfos().List()
				assert.NoError(t, err)
				assert.ElementsMatch(t, nodeNames(nodes), nodeInfoNames(baseNodes))
			})
	}
}

func TestForkRemovePods(t *testing.T) {
	nodeCount := 3
	podCount := 90
	deletedPodCount := 10

	nodes := createTestNodes(nodeCount)
	pods := createTestPods(podCount)
	assignPodsToNodes(pods, nodes)

	for name, snapshotFactory := range snapshots {
		t.Run(fmt.Sprintf("%s: fork should not affect base data: removing pods", name),
			func(t *testing.T) {
				clusterSnapshot := snapshotFactory()
				err := clusterSnapshot.AddNodes(nodes)
				assert.NoError(t, err)

				for _, pod := range pods {
					err = clusterSnapshot.AddPod(pod, pod.Spec.NodeName)
					assert.NoError(t, err)
				}

				err = clusterSnapshot.Fork()
				assert.NoError(t, err)

				for _, pod := range pods[:deletedPodCount] {
					err = clusterSnapshot.RemovePod(pod.Namespace, pod.Name)
					assert.NoError(t, err)
				}

				forkPods, err := clusterSnapshot.Pods().List(labels.Everything())
				assert.NoError(t, err)
				assert.ElementsMatch(t, pods[deletedPodCount:], forkPods)
				forkNodes, err := clusterSnapshot.NodeInfos().List()
				assert.NoError(t, err)
				assert.ElementsMatch(t, nodeNames(nodes), nodeInfoNames(forkNodes))
				assert.ElementsMatch(t, pods[deletedPodCount:], nodeInfoPods(forkNodes))

				err = clusterSnapshot.Revert()
				assert.NoError(t, err)

				basePods, err := clusterSnapshot.Pods().List(labels.Everything())
				assert.NoError(t, err)
				assert.ElementsMatch(t, pods, basePods)
				baseNodes, err := clusterSnapshot.NodeInfos().List()
				assert.NoError(t, err)
				assert.ElementsMatch(t, nodeNames(nodes), nodeInfoNames(baseNodes))
				assert.ElementsMatch(t, pods, nodeInfoPods(baseNodes))
			})
	}
}
