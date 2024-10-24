/*
Copyright 2023 The Kubernetes Authors.

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

package podsharding

import (
	"testing"

	v1 "k8s.io/api/core/v1"
)

func TestPodSharderWithoutProvReq(t *testing.T) {
	testCases := []shardComputeFunctionTestCase{
		{
			name: "pod without node-selectors",
			pod: v1.Pod{
				Spec: v1.PodSpec{
					NodeSelector: map[string]string{},
					Tolerations:  []v1.Toleration{},
				},
			},
			expectedNodeGroupDescriptor: NodeGroupDescriptor{
				Labels: map[string]string{},
				Taints: []v1.Taint{},
			},
		},
		{
			name: "pod with non-pod-sharding node-selectors",
			pod: v1.Pod{
				Spec: v1.PodSpec{
					NodeSelector: map[string]string{
						"notSharding": "nodeSelector",
					},
					Tolerations: []v1.Toleration{},
				},
			},
			expectedNodeGroupDescriptor: NodeGroupDescriptor{
				Labels: map[string]string{},
				Taints: []v1.Taint{},
			},
		},
		{
			name: "pod with non-pod-sharding node-selector",
			pod: v1.Pod{
				Spec: v1.PodSpec{
					NodeSelector: map[string]string{
						"notSharding": "nodeSelector",
					},
					Tolerations: []v1.Toleration{},
				},
			},
			expectedNodeGroupDescriptor: NodeGroupDescriptor{
				Labels: map[string]string{},
				Taints: []v1.Taint{},
			},
		},
		{
			name: "pod with pod-sharding node-selector",
			pod: v1.Pod{
				Spec: v1.PodSpec{
					NodeSelector: map[string]string{
						"sharding": "nodeSelector",
					},
					Tolerations: []v1.Toleration{},
				},
			},
			expectedNodeGroupDescriptor: NodeGroupDescriptor{
				Labels: map[string]string{
					"sharding": "nodeSelector",
				},
				Taints: []v1.Taint{},
			},
		},
		{
			name: "pod with pod-sharding node-selector",
			pod: v1.Pod{
				Spec: v1.PodSpec{
					NodeSelector: map[string]string{
						"sharding": "nodeSelector",
					},
					Tolerations: []v1.Toleration{},
				},
			},
			expectedNodeGroupDescriptor: NodeGroupDescriptor{
				Labels: map[string]string{
					"sharding": "nodeSelector",
				},
				Taints: []v1.Taint{},
			},
		},
		{
			name: "pod with pod-sharding node-selector and other node-selector",
			pod: v1.Pod{
				Spec: v1.PodSpec{
					NodeSelector: map[string]string{
						"sharding":    "nodeSelector",
						"notSharding": "nodeSelector",
					},
					Tolerations: []v1.Toleration{},
				},
			},
			expectedNodeGroupDescriptor: NodeGroupDescriptor{
				Labels: map[string]string{
					"sharding": "nodeSelector",
				},
				Taints: []v1.Taint{},
			},
		},
		{
			name: "pod with pod-sharding node-selector and tolerations",
			pod: v1.Pod{
				Spec: v1.PodSpec{
					NodeSelector: map[string]string{
						"sharding": "nodeSelector",
					},
					Tolerations: []v1.Toleration{
						{
							Key:      "key",
							Operator: "Exists",
							Value:    "value",
							Effect:   "NoSchedule",
						},
					},
				},
			},
			expectedNodeGroupDescriptor: NodeGroupDescriptor{
				Labels: map[string]string{
					"sharding": "nodeSelector",
				},
				Taints: []v1.Taint{},
			},
		},
	}

	shardingNodeSelectors := map[string]string{
		"sharding":        "nodeSelector",
		"anotherSharding": "nodeSelector",
	}

	testPodSharder(t, NewOssPodSharder(shardingNodeSelectors), testCases)
}

type shardComputeFunctionTestCase struct {
	name                        string
	pod                         v1.Pod
	expectedNodeGroupDescriptor NodeGroupDescriptor
}

func testPodSharder(t *testing.T, sharder PodSharder, testCases []shardComputeFunctionTestCase) {
	t.Helper()
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			shards := sharder.ComputePodShards([]*v1.Pod{&tc.pod})
			if len(shards) != 1 {
				t.Errorf("Expected to get precisely 1 shard, but got %d. Shards: %v", len(shards), shards)
				return
			}
			assertNodeGroupDescriptorEqual(t, tc.expectedNodeGroupDescriptor, shards[0].NodeGroupDescriptor)
		})
	}
}
