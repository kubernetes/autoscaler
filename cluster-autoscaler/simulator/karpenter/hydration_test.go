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

package karpenter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/testsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
)

func TestHydrateClusterState_PodAffinity(t *testing.T) {
	snapshot := testsnapshot.NewTestSnapshotOrDie(t)

	// 1. Create a full node with a target pod for affinity
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "full-node",
		},
		Status: apiv1.NodeStatus{
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("4Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
		},
	}

	// This pod fills up the node CPU capacity (requested = allocatable)
	fillingPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "filling-pod",
			Namespace: "default",
			Labels: map[string]string{
				"app": "backend", // Target label for affinity
			},
		},
		Spec: apiv1.PodSpec{
			NodeName: "full-node",
			Containers: []apiv1.Container{
				{
					Name: "container",
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceCPU:    resource.MustParse("2"),
							apiv1.ResourceMemory: resource.MustParse("4Gi"),
						},
					},
				},
			},
		},
	}

	err := snapshot.SetClusterState([]*apiv1.Node{node}, []*apiv1.Pod{fillingPod}, nil, nil, nil, nil, nil)
	assert.NoError(t, err)

	nodeInfo, err := snapshot.GetNodeInfo("full-node")
	assert.NoError(t, err)
	assert.NotNil(t, nodeInfo)

	// 2. Create a pending pod with PodAffinity to "app: backend"
	pendingPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pending-pod",
			Namespace: "default",
		},
		Spec: apiv1.PodSpec{
			Affinity: &apiv1.Affinity{
				PodAffinity: &apiv1.PodAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: []apiv1.PodAffinityTerm{
						{
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"app": "backend",
								},
							},
							TopologyKey: "kubernetes.io/hostname",
						},
					},
				},
			},
			Containers: []apiv1.Container{
				{
					Name: "container",
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceCPU: resource.MustParse("100m"),
						},
					},
				},
			},
		},
	}

	// Run HydrateClusterState
	relevantPods, relevantNodes, err := HydrateClusterState(
		context.Background(),
		snapshot,
		[]*framework.NodeInfo{nodeInfo},
		[]*apiv1.Pod{pendingPod},
	)

	assert.NoError(t, err)

	// Since pendingPod has affinity to fillingPod (on full-node),
	// full-node and fillingPod should NOT be pruned.
	assert.Len(t, relevantNodes, 1, "Full node should not be pruned because it contains a pod matched by PodAffinity")
	assert.Equal(t, "full-node", relevantNodes[0].Name)

	assert.Len(t, relevantPods, 1, "Filling pod should not be pruned because it is matched by PodAffinity")
	assert.Equal(t, "filling-pod", relevantPods[0].Name)
}

func TestHydrateClusterState_HostPortAndCSIVolumesExemption(t *testing.T) {
	snapshot := testsnapshot.NewTestSnapshotOrDie(t)

	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
		Status: apiv1.NodeStatus{
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("8"),
				apiv1.ResourceMemory: resource.MustParse("16Gi"),
			},
		},
	}

	hostPortPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "hostport-pod", Namespace: "default"},
		Spec: apiv1.PodSpec{
			NodeName: "node-1",
			Containers: []apiv1.Container{{
				Name:  "c",
				Ports: []apiv1.ContainerPort{{HostPort: 8080}},
			}},
		},
	}

	csiPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "csi-pod", Namespace: "default"},
		Spec: apiv1.PodSpec{
			NodeName: "node-1",
			Volumes: []apiv1.Volume{{
				Name: "vol",
				VolumeSource: apiv1.VolumeSource{
					PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{ClaimName: "my-pvc"},
				},
			}},
			Containers: []apiv1.Container{{Name: "c"}},
		},
	}

	err := snapshot.SetClusterState([]*apiv1.Node{node}, []*apiv1.Pod{hostPortPod, csiPod}, nil, nil, nil, nil, nil)
	assert.NoError(t, err)

	nodeInfo, err := snapshot.GetNodeInfo("node-1")
	assert.NoError(t, err)

	pendingPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pending-pod", Namespace: "default"},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{{
				Name: "c",
				Resources: apiv1.ResourceRequirements{
					Requests: apiv1.ResourceList{apiv1.ResourceCPU: resource.MustParse("1")},
				},
			}},
		},
	}

	relevantPods, relevantNodes, err := HydrateClusterState(context.Background(), snapshot, []*framework.NodeInfo{nodeInfo}, []*apiv1.Pod{pendingPod})
	assert.NoError(t, err)
	assert.Len(t, relevantNodes, 1)
	assert.Len(t, relevantPods, 2, "HostPort and CSI volume pods must never be pruned")
}

func TestHydrateClusterState_ResourceClamping(t *testing.T) {
	snapshot := testsnapshot.NewTestSnapshotOrDie(t)

	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
		Status: apiv1.NodeStatus{
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("8"),
				apiv1.ResourceMemory: resource.MustParse("16Gi"),
				"nvidia.com/gpu":     resource.MustParse("2"),
			},
		},
	}

	// Irrelevant running pod that will be pruned
	prunedPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pruned-pod", Namespace: "default"},
		Spec: apiv1.PodSpec{
			NodeName: "node-1",
			Containers: []apiv1.Container{{
				Name: "c",
				Resources: apiv1.ResourceRequirements{
					Requests: apiv1.ResourceList{
						apiv1.ResourceCPU:    resource.MustParse("4"),
						apiv1.ResourceMemory: resource.MustParse("8Gi"),
						"nvidia.com/gpu":     resource.MustParse("1"),
					},
				},
			}},
		},
	}

	err := snapshot.SetClusterState([]*apiv1.Node{node}, []*apiv1.Pod{prunedPod}, nil, nil, nil, nil, nil)
	assert.NoError(t, err)

	nodeInfo, err := snapshot.GetNodeInfo("node-1")
	assert.NoError(t, err)

	pendingPod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pending-pod", Namespace: "default"},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{{
				Name: "c",
				Resources: apiv1.ResourceRequirements{
					Requests: apiv1.ResourceList{apiv1.ResourceCPU: resource.MustParse("1")},
				},
			}},
		},
	}

	relevantPods, relevantNodes, err := HydrateClusterState(context.Background(), snapshot, []*framework.NodeInfo{nodeInfo}, []*apiv1.Pod{pendingPod})
	assert.NoError(t, err)
	assert.Len(t, relevantPods, 0, "pruned-pod should be pruned")
	assert.Len(t, relevantNodes, 1)

	clampedNode := relevantNodes[0]
	assert.Equal(t, int64(4000), clampedNode.Status.Allocatable.Cpu().MilliValue(), "CPU should be clamped from 8 to 4")
	mem8Gi := resource.MustParse("8Gi")
	assert.Equal(t, mem8Gi.Value(), clampedNode.Status.Allocatable.Memory().Value(), "Mem should be clamped from 16Gi to 8Gi")
	gpuQty := clampedNode.Status.Allocatable["nvidia.com/gpu"]
	assert.Equal(t, int64(1), gpuQty.Value(), "GPU should be clamped from 2 to 1")
}

func TestHydrateClusterState_DynamicPendingPodThresholds(t *testing.T) {
	// 1. Test Disjoint Pending Pod Resources: pod1 requests only CPU, pod2 requests only Memory
	pendingPodCpuOnly := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-cpu", Namespace: "default"},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{{
				Name:      "c",
				Resources: apiv1.ResourceRequirements{Requests: apiv1.ResourceList{apiv1.ResourceCPU: resource.MustParse("1")}},
			}},
		},
	}
	pendingPodMemOnly := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-mem", Namespace: "default"},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{{
				Name:      "c",
				Resources: apiv1.ResourceRequirements{Requests: apiv1.ResourceList{apiv1.ResourceMemory: resource.MustParse("1Gi")}},
			}},
		},
	}

	// Calculate unique pending pod request shapes to verify disjoint resources resolve to 2 unique shapes
	uniqueReqs := getUniquePendingPodRequests([]*apiv1.Pod{pendingPodCpuOnly, pendingPodMemOnly})
	assert.Len(t, uniqueReqs, 2, "should extract 2 unique pod request shapes")

	// 2. Node with 0 CPU free & 2Gi Mem free (should be retained because pendingPodMemOnly can fit!)
	nodeMemFree := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-mem-free", Labels: map[string]string{apiv1.LabelHostname: "node-mem-free"}},
		Status: apiv1.NodeStatus{
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("0"),
				apiv1.ResourceMemory: resource.MustParse("2Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("0"),
				apiv1.ResourceMemory: resource.MustParse("2Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
		},
	}

	// 3. Node with 2 CPU free & 0 Mem free (should be retained because pendingPodCpuOnly can fit!)
	nodeCpuFree := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-cpu-free", Labels: map[string]string{apiv1.LabelHostname: "node-cpu-free"}},
		Status: apiv1.NodeStatus{
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("0"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("0"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
		},
	}

	// 4. Fully utilized node (0 CPU, 0 Mem) (should be pruned!)
	nodeFull := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-full", Labels: map[string]string{apiv1.LabelHostname: "node-full"}},
		Status: apiv1.NodeStatus{
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("0"),
				apiv1.ResourceMemory: resource.MustParse("0"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("0"),
				apiv1.ResourceMemory: resource.MustParse("0"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
		},
	}

	snapshot := testsnapshot.NewTestSnapshotOrDie(t)
	err := snapshot.SetClusterState([]*apiv1.Node{nodeMemFree, nodeCpuFree, nodeFull}, nil, nil, nil, nil, nil, nil)
	assert.NoError(t, err)

	niMemFree, _ := snapshot.GetNodeInfo("node-mem-free")
	niCpuFree, _ := snapshot.GetNodeInfo("node-cpu-free")
	niFull, _ := snapshot.GetNodeInfo("node-full")

	relevantPods, relevantNodes, err := HydrateClusterState(context.Background(), snapshot, []*framework.NodeInfo{niMemFree, niCpuFree, niFull}, []*apiv1.Pod{pendingPodCpuOnly, pendingPodMemOnly})
	assert.NoError(t, err)
	assert.Len(t, relevantPods, 0)
	assert.Len(t, relevantNodes, 2, "nodeMemFree and nodeCpuFree should be retained, nodeFull should be pruned")

	nodeNames := []string{relevantNodes[0].Name, relevantNodes[1].Name}
	assert.Contains(t, nodeNames, "node-mem-free")
	assert.Contains(t, nodeNames, "node-cpu-free")
	assert.NotContains(t, nodeNames, "node-full")
}

func TestHydrateClusterState_ParetoMinimalPodShapes(t *testing.T) {
	// Pod A: (1 CPU, 3Gi Mem)
	podA := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-a", Namespace: "default"},
		Spec: apiv1.PodSpec{Containers: []apiv1.Container{{
			Resources: apiv1.ResourceRequirements{Requests: apiv1.ResourceList{apiv1.ResourceCPU: resource.MustParse("1"), apiv1.ResourceMemory: resource.MustParse("3Gi")}},
		}}},
	}
	// Pod B: (2 CPU, 2Gi Mem)
	podB := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-b", Namespace: "default"},
		Spec: apiv1.PodSpec{Containers: []apiv1.Container{{
			Resources: apiv1.ResourceRequirements{Requests: apiv1.ResourceList{apiv1.ResourceCPU: resource.MustParse("2"), apiv1.ResourceMemory: resource.MustParse("2Gi")}},
		}}},
	}
	// Pod C: (3 CPU, 1Gi Mem)
	podC := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-c", Namespace: "default"},
		Spec: apiv1.PodSpec{Containers: []apiv1.Container{{
			Resources: apiv1.ResourceRequirements{Requests: apiv1.ResourceList{apiv1.ResourceCPU: resource.MustParse("3"), apiv1.ResourceMemory: resource.MustParse("1Gi")}},
		}}},
	}
	// Pod D (Dominated by Pod A): (2 CPU, 4Gi Mem) -- since 1 <= 2 and 3Gi <= 4Gi
	podDominated := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod-dom", Namespace: "default"},
		Spec: apiv1.PodSpec{Containers: []apiv1.Container{{
			Resources: apiv1.ResourceRequirements{Requests: apiv1.ResourceList{apiv1.ResourceCPU: resource.MustParse("2"), apiv1.ResourceMemory: resource.MustParse("4Gi")}},
		}}},
	}

	// 1. Verify Pareto reduction: Pods A, B, C are mutually incomparable, so ALL 3 are retained. Pod D is dominated, so it is pruned from the minimal set.
	minimalShapes := getUniquePendingPodRequests([]*apiv1.Pod{podA, podB, podC, podDominated})
	assert.Len(t, minimalShapes, 3, "Pod A, B, C should all be retained in minimal set; Pod D should be pruned")

	// 2. Test Node with (2 CPU, 2Gi Mem) free:
	// Cannot fit Pod A (1, 3Gi - needs 3Gi Mem)
	// Cannot fit Pod C (3, 1Gi - needs 3 CPU)
	// Fits Pod B (2, 2Gi)!
	nodeIntermediate := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-2cpu-2gi", Labels: map[string]string{apiv1.LabelHostname: "node-2cpu-2gi"}},
		Status: apiv1.NodeStatus{
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    resource.MustParse("2"),
				apiv1.ResourceMemory: resource.MustParse("2Gi"),
				apiv1.ResourcePods:   resource.MustParse("110"),
			},
		},
	}

	snapshot := testsnapshot.NewTestSnapshotOrDie(t)
	err := snapshot.SetClusterState([]*apiv1.Node{nodeIntermediate}, nil, nil, nil, nil, nil, nil)
	assert.NoError(t, err)

	ni, _ := snapshot.GetNodeInfo("node-2cpu-2gi")
	_, relevantNodes, err := HydrateClusterState(context.Background(), snapshot, []*framework.NodeInfo{ni}, []*apiv1.Pod{podA, podB, podC})
	assert.NoError(t, err)
	assert.Len(t, relevantNodes, 1, "node with (2 CPU, 2Gi Mem) must be retained because Pod B (2 CPU, 2Gi Mem) fits")
}
