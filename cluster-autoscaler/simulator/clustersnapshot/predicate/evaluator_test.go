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

package predicate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/store/streaming"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

func TestPredicateEvaluator_AntiAffinity(t *testing.T) {
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node1",
			Labels: map[string]string{
				"topology": "zone1",
			},
		},
	}

	// Case 1: Incoming pod has anti-affinity against existing pod
	t.Run("Incoming pod anti-affinity", func(t *testing.T) {
		store := streaming.NewStreamingSnapshotStore()
		podInformer := store.GetPodInformer()
		nodeInformer := store.GetNodeInformer()
		evaluator := NewPredicateEvaluator(podInformer, nodeInformer)

		store.StoreNodeInfo(framework.NewNodeInfo(node, nil))

		existingPod := &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "p1",
				Namespace: "default",
				Labels:    map[string]string{"app": "foo"},
			},
			Spec: apiv1.PodSpec{
				NodeName: "node1",
			},
		}
		store.StorePodInfo(framework.NewPodInfo(existingPod, nil), "node1")

		incomingPod := &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "p2",
				Namespace: "default",
				Labels:    map[string]string{"app": "bar"},
			},
			Spec: apiv1.PodSpec{
				Affinity: &apiv1.Affinity{
					PodAntiAffinity: &apiv1.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []apiv1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{"app": "foo"},
								},
								TopologyKey: "topology",
							},
						},
					},
				},
			},
		}

		err := evaluator.FastCheckAffinity(evaluator.PreparePod(incomingPod), node)
		assert.Error(t, err) // Should fail because p1 is already there
	})

	// Case 2: Existing pod has anti-affinity against incoming pod
	t.Run("Existing pod anti-affinity (symmetry)", func(t *testing.T) {
		store := streaming.NewStreamingSnapshotStore()
		podInformer := store.GetPodInformer()
		nodeInformer := store.GetNodeInformer()
		evaluator := NewPredicateEvaluator(podInformer, nodeInformer)

		store.StoreNodeInfo(framework.NewNodeInfo(node, nil))

		existingPodWithAntiAffinity := &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "p1",
				Namespace: "default",
				Labels:    map[string]string{"app": "foo"},
			},
			Spec: apiv1.PodSpec{
				NodeName: "node1",
				Affinity: &apiv1.Affinity{
					PodAntiAffinity: &apiv1.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []apiv1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{"app": "bar"},
								},
								TopologyKey: "topology",
							},
						},
					},
				},
			},
		}
		store.StorePodInfo(framework.NewPodInfo(existingPodWithAntiAffinity, nil), "node1")

		incomingPod := &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "p2",
				Namespace: "default",
				Labels:    map[string]string{"app": "bar"},
			},
		}


		err := evaluator.FastCheckAffinity(evaluator.PreparePod(incomingPod), node)
		assert.Error(t, err) // Should fail because existingPodWithAntiAffinity rejects bar
	})

	// Case 3: Namespace handling
	t.Run("Namespace handling", func(t *testing.T) {
		store := streaming.NewStreamingSnapshotStore()
		podInformer := store.GetPodInformer()
		nodeInformer := store.GetNodeInformer()
		evaluator := NewPredicateEvaluator(podInformer, nodeInformer)

		store.StoreNodeInfo(framework.NewNodeInfo(node, nil))

		existingPod := &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "p1",
				Namespace: "other-ns",
				Labels:    map[string]string{"app": "foo"},
			},
			Spec: apiv1.PodSpec{
				NodeName: "node1",
			},
		}
		store.StorePodInfo(framework.NewPodInfo(existingPod, nil), "node1")

		// Incoming pod with anti-affinity but ONLY in "default" namespace
		incomingPod := &apiv1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "p2",
				Namespace: "default",
				Labels:    map[string]string{"app": "bar"},
			},
			Spec: apiv1.PodSpec{
				Affinity: &apiv1.Affinity{
					PodAntiAffinity: &apiv1.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []apiv1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{"app": "foo"},
								},
								TopologyKey: "topology",
								Namespaces:  []string{"default"},
							},
						},
					},
				},
			},
		}

		err := evaluator.FastCheckAffinity(evaluator.PreparePod(incomingPod), node)
		assert.NoError(t, err) // Should pass because existing pod is in "other-ns"

		// Change incoming pod to include "other-ns"
		incomingPod.Spec.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution[0].Namespaces = []string{"other-ns"}
		err = evaluator.FastCheckAffinity(evaluator.PreparePod(incomingPod), node)
		assert.Error(t, err) // Should fail now
	})
}
