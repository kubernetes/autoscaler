/*
Copyright 2024 The Kubernetes Authors.

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

package forcescaledown

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

func TestBackfillTaintTimeAddedIfEmtpy(t *testing.T) {
	defaultTimeAdded := &metav1.Time{Time: time.Date(2024, time.March, 9, 0, 0, 0, 0, time.UTC)}
	testCases := []struct {
		name               string
		nodes              []*apiv1.Node
		nodesWithTimeAdded []string
	}{
		{
			name: "no nodes to process",
		},
		{
			name: "all nodes already have taint with TimeAdded",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
			},
			nodesWithTimeAdded: []string{"n1"},
		},
		{
			name: "1 out of 2 nodes need to be backfilled",
			nodes: []*apiv1.Node{
				BuildTestNode("n1", 1000, 10),
				BuildTestNode("n2", 1000, 10),
			},
			nodesWithTimeAdded: []string{"n1"},
		},
	}

	for index := range testCases {
		tc := testCases[index]
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			backfilledNodeNames := map[string]bool{}
			for _, name := range tc.nodesWithTimeAdded {
				backfilledNodeNames[name] = true
			}
			for index := range tc.nodes {
				if backfilledNodeNames[tc.nodes[index].Name] {
					taint := apiv1.Taint{
						Key:       taints.ForceScaleDownTaint,
						Effect:    apiv1.TaintEffectNoSchedule,
						TimeAdded: defaultTimeAdded,
					}
					tc.nodes[index].Spec.Taints = append(tc.nodes[index].Spec.Taints, taint)
				}
			}

			fakeClient := fake.NewSimpleClientset()
			for index := range tc.nodes {
				_, err := fakeClient.CoreV1().Nodes().Create(context.TODO(), tc.nodes[index], metav1.CreateOptions{})
				assert.NoError(t, err)
			}
			informerFactory := informers.NewSharedInformerFactory(fakeClient, 0)
			stop := make(chan struct{})
			informerFactory.Start(stop)

			processor := NewForceScaleDownNodeTaintsObserver(fakeClient, informerFactory)
			err := processor.backfillTaintTimeAddedIfEmtpy(context.TODO())
			assert.NoError(t, err)
			nodeList, err := fakeClient.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			assert.NoError(t, err)
			assert.Equal(t, len(tc.nodes), len(nodeList.Items))
			for _, node := range nodeList.Items {
				for _, taint := range node.Spec.Taints {
					assert.NotNil(t, taint.TimeAdded, "Node %s taint %s should have non-nil TimeAdded", node.Name, taint.Key)
				}
			}
		})
	}
}
