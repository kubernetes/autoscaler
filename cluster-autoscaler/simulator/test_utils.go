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
	"testing"

	apiv1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
)

// InitializeClusterSnapshotOrDie clears cluster snapshot and then initializes it with given set of nodes and pods.
// Both Spec.NodeName and Status.NominatedNodeName are used when simulating scheduling pods.
func InitializeClusterSnapshotOrDie(
	t *testing.T,
	snapshot ClusterSnapshot,
	nodes []*apiv1.Node,
	pods []*apiv1.Pod) {
	var err error

	snapshot.Clear()

	for _, node := range nodes {
		err = snapshot.AddNode(node)
		assert.NoError(t, err, "error while adding node %s", node.Name)
	}

	for _, pod := range pods {
		if pod.Spec.NodeName != "" {
			err = snapshot.AddPod(pod, pod.Spec.NodeName)
			assert.NoError(t, err, "error while adding pod %s/%s to node %s", pod.Namespace, pod.Name, pod.Spec.NodeName)
		} else if pod.Status.NominatedNodeName != "" {
			err = snapshot.AddPod(pod, pod.Status.NominatedNodeName)
			assert.NoError(t, err, "error while adding pod %s/%s to nominated node %s", pod.Namespace, pod.Name, pod.Status.NominatedNodeName)
		} else {
			assert.Fail(t, "pod %s/%s does not have Spec.NodeName nor Status.NominatedNodeName set", pod.Namespace, pod.Name)
		}
	}
}
