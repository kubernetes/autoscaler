/*
Copyright 2021 The Kubernetes Authors.

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

package pods

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/datadog/common"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
)

func TestTransformDataNodesProcess(t *testing.T) {
	tests := []struct {
		name     string
		node     *corev1.Node
		expected *corev1.Node
	}{

		{
			"Resource is added to fresh nodes having local-data label",
			buildTestNode("a", NodeReadyGraceDelay/2, true, false),
			buildTestNode("a", NodeReadyGraceDelay/2, true, true),
		},

		{
			"Resource is not added to old nodes having local-data label",
			buildTestNode("b", 2*NodeReadyGraceDelay, true, false),
			buildTestNode("b", 2*NodeReadyGraceDelay, true, false),
		},

		{
			"Resource is not added to new nodes without local-data label",
			buildTestNode("c", NodeReadyGraceDelay/2, false, false),
			buildTestNode("c", NodeReadyGraceDelay/2, false, false),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clusterSnapshot := clustersnapshot.NewBasicClusterSnapshot()
			err := clusterSnapshot.AddNode(tt.node)
			assert.NoError(t, err)

			ctx := &context.AutoscalingContext{
				ClusterSnapshot: clusterSnapshot,
			}

			proc := NewTransformDataNodes()
			_, err = proc.Process(ctx, []*corev1.Pod{})
			assert.NoError(t, err)

			actual, err := ctx.ClusterSnapshot.NodeInfos().Get(tt.node.GetName())
			assert.NoError(t, err)

			assert.Equal(t, tt.expected.Status.Capacity, actual.Node().Status.Capacity)
			assert.Equal(t, tt.expected.Status.Allocatable, actual.Node().Status.Allocatable)
		})
	}

}

func buildTestNode(name string, age time.Duration, localDataLabel, localDataResource bool) *corev1.Node {
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			SelfLink:          fmt.Sprintf("/api/v1/nodes/%s", name),
			Labels:            map[string]string{},
			CreationTimestamp: metav1.NewTime(time.Now().Add(-age)),
		},
		Status: corev1.NodeStatus{
			Capacity:    corev1.ResourceList{},
			Allocatable: corev1.ResourceList{},
			Conditions: []corev1.NodeCondition{
				{
					Type:               corev1.NodeReady,
					Status:             corev1.ConditionTrue,
					LastTransitionTime: metav1.NewTime(time.Now().Add(-age)),
				},
			},
		},
	}

	if localDataLabel {
		node.ObjectMeta.Labels[common.DatadogLocalStorageLabel] = "true"
	}

	if localDataResource {
		node.Status.Capacity[common.DatadogLocalDataResource] = common.DatadogLocalDataQuantity.DeepCopy()
		node.Status.Allocatable[common.DatadogLocalDataResource] = common.DatadogLocalDataQuantity.DeepCopy()
	}

	return node
}
