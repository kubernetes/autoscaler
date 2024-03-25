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
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
)

func TestDrainable(t *testing.T) {
	testTime := time.Date(2024, time.March, 9, 0, 0, 0, 0, time.UTC)
	testCases := map[string]struct {
		pod   *apiv1.Pod
		nodes []*apiv1.Node
		want  drainability.Status
	}{
		"regular pod": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod",
					Namespace: "ns",
				},
			},
			want: drainability.NewUndefinedStatus(),
		},
		"pod on non force-scale-down node": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod",
					Namespace: "ns",
				},
				Spec: apiv1.PodSpec{
					NodeName: "node",
				},
			},
			nodes: []*apiv1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
					},
					Spec: apiv1.NodeSpec{
						Taints: []apiv1.Taint{},
					},
				},
			},
			want: drainability.NewUndefinedStatus(),
		},
		"pod on force-scale-down node without deadline": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod",
					Namespace: "ns",
				},
				Spec: apiv1.PodSpec{
					NodeName: "node",
				},
			},
			nodes: []*apiv1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
					},
					Spec: apiv1.NodeSpec{
						Taints: []apiv1.Taint{
							{
								Key:    taints.ForceScaleDownTaint,
								Effect: apiv1.TaintEffectNoSchedule,
							},
						},
					},
				},
			},
			want: drainability.NewDrainableStatus(),
		},
		"pod on force-scale-down node before deadline": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod",
					Namespace: "ns",
				},
				Spec: apiv1.PodSpec{
					NodeName: "node",
				},
			},
			nodes: []*apiv1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
					},
					Spec: apiv1.NodeSpec{
						Taints: []apiv1.Taint{
							{
								Key:       taints.ForceScaleDownTaint,
								Effect:    apiv1.TaintEffectNoSchedule,
								Value:     "30",
								TimeAdded: &metav1.Time{Time: testTime},
							},
						},
					},
				},
			},
			want: drainability.NewDrainableStatus(),
		},
		"pod on force-scale-down node after deadline": {
			pod: &apiv1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "pod",
					Namespace: "ns",
				},
				Spec: apiv1.PodSpec{
					NodeName: "node",
				},
			},
			nodes: []*apiv1.Node{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
					},
					Spec: apiv1.NodeSpec{
						Taints: []apiv1.Taint{
							{
								Key:       taints.ForceScaleDownTaint,
								Effect:    apiv1.TaintEffectNoSchedule,
								Value:     "30",
								TimeAdded: &metav1.Time{Time: testTime.Add(-60 * time.Minute)},
							},
						},
					},
				},
			},
			want: drainability.NewSkipStatus(),
		},
	}

	for desc, tc := range testCases {
		tc := tc
		t.Run(desc, func(t *testing.T) {
			t.Parallel()
			ctx := drainability.DrainContext{
				Listers:   newMockListerRegistry(tc.nodes),
				Timestamp: testTime,
			}
			got := New().Drainable(&ctx, tc.pod)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("Rule.Drainable(%v): got status diff (-want +got):\n%s", tc.pod.Name, diff)
			}
		})
	}
}

type mockListerRegistry struct {
	kube_util.ListerRegistry
	nodes []*apiv1.Node
}

func newMockListerRegistry(nodes []*apiv1.Node) *mockListerRegistry {
	return &mockListerRegistry{
		nodes: nodes,
	}
}

func (mlr mockListerRegistry) AllNodeLister() kube_util.NodeLister {
	return &mockNodeLister{nodes: mlr.nodes}
}

type mockNodeLister struct {
	nodes []*apiv1.Node
}

func (mnl *mockNodeLister) List() ([]*apiv1.Node, error) {
	return mnl.nodes, nil
}
func (mnl *mockNodeLister) Get(name string) (*apiv1.Node, error) {
	for _, node := range mnl.nodes {
		if node.Name == name {
			return node, nil
		}
	}
	return nil, fmt.Errorf("node %s not found", name)
}
