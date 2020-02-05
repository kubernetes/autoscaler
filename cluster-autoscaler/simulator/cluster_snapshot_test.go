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
	"k8s.io/apimachinery/pkg/labels"
)

var snapshots = map[string]func() ClusterSnapshot{
	"basic": func() ClusterSnapshot { return NewBasicClusterSnapshot() },
	"delta": func() ClusterSnapshot { return NewDeltaClusterSnapshot() },
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
				assert.Equal(t, nodeCount+len(extraNodes), len(forkNodes))

				err = clusterSnapshot.Revert()
				assert.NoError(t, err)

				baseNodes, err := clusterSnapshot.NodeInfos().List()
				assert.NoError(t, err)
				assert.Equal(t, nodeCount, len(baseNodes))
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

				err = clusterSnapshot.Revert()
				assert.NoError(t, err)

				basePods, err := clusterSnapshot.Pods().List(labels.Everything())
				assert.NoError(t, err)
				assert.Equal(t, 0, len(basePods))
			})
	}
}
