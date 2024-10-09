/*
Copyright 2016 The Kubernetes Authors.

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

package estimator

import (
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"

	"github.com/stretchr/testify/assert"
)

func makePodEquivalenceGroup(pod *apiv1.Pod, podCount int) PodEquivalenceGroup {
	pods := []*apiv1.Pod{}
	for i := 0; i < podCount; i++ {
		pods = append(pods, pod)
	}
	return PodEquivalenceGroup{
		Pods: pods,
	}
}

func makeNode(cpu, mem, podCount int64, name string, zone string) *apiv1.Node {
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"kubernetes.io/hostname":      name,
				"topology.kubernetes.io/zone": zone,
			},
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    *resource.NewMilliQuantity(cpu, resource.DecimalSI),
				apiv1.ResourceMemory: *resource.NewQuantity(mem*units.MiB, resource.DecimalSI),
				apiv1.ResourcePods:   *resource.NewQuantity(podCount, resource.DecimalSI),
			},
		},
	}
	node.Status.Allocatable = node.Status.Capacity
	SetNodeReadyState(node, true, time.Time{})
	return node
}

func TestBinpackingEstimate(t *testing.T) {
	highResourcePodGroup := makePodEquivalenceGroup(
		BuildTestPod(
			"estimatee",
			500,
			1000,
			WithNamespace("universe"),
			WithLabels(map[string]string{
				"app": "estimatee",
			}),
		),
		10,
	)
	testCases := []struct {
		name                 string
		millicores           int64
		memory               int64
		maxNodes             int
		podsEquivalenceGroup []PodEquivalenceGroup
		topologySpreadingKey string
		expectNodeCount      int
		expectPodCount       int
		expectProcessedPods  []*apiv1.Pod
	}{
		{
			name:       "simple resource-based binpacking",
			millicores: 350*3 - 50,
			memory:     2 * 1000,
			podsEquivalenceGroup: []PodEquivalenceGroup{makePodEquivalenceGroup(
				BuildTestPod(
					"estimatee",
					350,
					1000,
					WithNamespace("universe"),
					WithLabels(map[string]string{
						"app": "estimatee",
					})), 10)},
			expectNodeCount: 5,
			expectPodCount:  10,
		},
		{
			name:       "pods-per-node bound binpacking",
			millicores: 10000,
			memory:     20000,
			podsEquivalenceGroup: []PodEquivalenceGroup{makePodEquivalenceGroup(
				BuildTestPod(
					"estimatee",
					10,
					100,
					WithNamespace("universe"),
					WithLabels(map[string]string{
						"app": "estimatee",
					})), 20)},
			expectNodeCount: 2,
			expectPodCount:  20,
		},
		{
			name:       "hostport conflict forces pod-per-node",
			millicores: 1000,
			memory:     5000,
			podsEquivalenceGroup: []PodEquivalenceGroup{makePodEquivalenceGroup(
				BuildTestPod(
					"estimatee",
					200,
					1000,
					WithNamespace("universe"),
					WithLabels(map[string]string{
						"app": "estimatee",
					}),
					WithHostPort(5555)), 8)},
			expectNodeCount: 8,
			expectPodCount:  8,
		},
		{
			name:       "limiter cuts binpacking",
			millicores: 1000,
			memory:     5000,
			podsEquivalenceGroup: []PodEquivalenceGroup{makePodEquivalenceGroup(
				BuildTestPod(
					"estimatee",
					500,
					1000,
					WithNamespace("universe"),
					WithLabels(map[string]string{
						"app": "estimatee",
					})), 20)},
			maxNodes:        5,
			expectNodeCount: 5,
			expectPodCount:  10,
		},
		{
			name:       "decreasing ordered pods are processed first",
			millicores: 1000,
			memory:     5000,
			podsEquivalenceGroup: append([]PodEquivalenceGroup{makePodEquivalenceGroup(
				BuildTestPod(
					"estimatee",
					50,
					1000,
					WithNamespace("universe"),
					WithLabels(map[string]string{
						"app": "estimatee",
					})), 10)}, highResourcePodGroup),
			maxNodes:            5,
			expectNodeCount:     5,
			expectPodCount:      10,
			expectProcessedPods: highResourcePodGroup.Pods,
		},
		{
			name:       "hostname topology spreading with maxSkew=2 forces 2 pods/node",
			millicores: 1000,
			memory:     5000,
			podsEquivalenceGroup: []PodEquivalenceGroup{makePodEquivalenceGroup(
				BuildTestPod(
					"estimatee",
					20,
					100,
					WithNamespace("universe"),
					WithLabels(map[string]string{
						"app": "estimatee",
					}),
					WithMaxSkew(2, "kubernetes.io/hostname")), 8)},
			expectNodeCount: 4,
			expectPodCount:  8,
		},
		{
			name:       "zonal topology spreading with maxSkew=2 only allows 2 pods to schedule",
			millicores: 1000,
			memory:     5000,
			podsEquivalenceGroup: []PodEquivalenceGroup{makePodEquivalenceGroup(
				BuildTestPod(
					"estimatee",
					20,
					100,
					WithNamespace("universe"),
					WithLabels(map[string]string{
						"app": "estimatee",
					}),
					WithMaxSkew(2, "topology.kubernetes.io/zone")), 8)},
			expectNodeCount: 1,
			expectPodCount:  2,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fwHandle := framework.TestFrameworkHandleOrDie(t)
			clusterSnapshot := clustersnapshot.NewBasicClusterSnapshot(fwHandle, true)
			// Add one node in different zone to trigger topology spread constraints
			err := clusterSnapshot.AddNodeInfo(framework.NewTestNodeInfo(makeNode(100, 100, 10, "oldnode", "zone-jupiter")))
			assert.NoError(t, err)

			predicateChecker := predicatechecker.NewSchedulerBasedPredicateChecker(fwHandle)
			limiter := NewThresholdBasedEstimationLimiter([]Threshold{NewStaticThreshold(tc.maxNodes, time.Duration(0))})
			processor := NewDecreasingPodOrderer()
			estimator := NewBinpackingNodeEstimator(predicateChecker, clusterSnapshot, limiter, processor, nil /* EstimationContext */, nil /* EstimationAnalyserFunc */)
			node := makeNode(tc.millicores, tc.memory, 10, "template", "zone-mars")
			nodeInfo := framework.NewNodeInfo(node, nil)

			estimatedNodes, estimatedPods := estimator.Estimate(tc.podsEquivalenceGroup, nodeInfo, nil)
			assert.Equal(t, tc.expectNodeCount, estimatedNodes)
			assert.Equal(t, tc.expectPodCount, len(estimatedPods))
			if tc.expectProcessedPods != nil {
				assert.Equal(t, tc.expectProcessedPods, estimatedPods)
			}
		})
	}
}

func BenchmarkBinpackingEstimate(b *testing.B) {
	millicores := int64(1000)
	memory := int64(5000)
	podsPerNode := int64(100)
	maxNodes := 3000
	expectNodeCount := 2595
	expectPodCount := 51000
	podsEquivalenceGroup := []PodEquivalenceGroup{
		makePodEquivalenceGroup(
			BuildTestPod(
				"estimatee",
				50,
				100,
				WithNamespace("universe"),
				WithLabels(map[string]string{
					"app": "estimatee",
				})),
			50000,
		),
		makePodEquivalenceGroup(
			BuildTestPod(
				"estimatee",
				95,
				190,
				WithNamespace("universe"),
				WithLabels(map[string]string{
					"app": "estimatee",
				})),
			1000,
		),
	}

	for i := 0; i < b.N; i++ {
		fwHandle, err := framework.TestFrameworkHandle()
		assert.NoError(b, err)
		clusterSnapshot := clustersnapshot.NewBasicClusterSnapshot(fwHandle, true)
		err = clusterSnapshot.AddNodeInfo(framework.NewTestNodeInfo(makeNode(100, 100, 10, "oldnode", "zone-jupiter")))
		assert.NoError(b, err)

		predicateChecker := predicatechecker.NewSchedulerBasedPredicateChecker(fwHandle)
		limiter := NewThresholdBasedEstimationLimiter([]Threshold{NewStaticThreshold(maxNodes, time.Duration(0))})
		processor := NewDecreasingPodOrderer()
		estimator := NewBinpackingNodeEstimator(predicateChecker, clusterSnapshot, limiter, processor, nil /* EstimationContext */, nil /* EstimationAnalyserFunc */)
		node := makeNode(millicores, memory, podsPerNode, "template", "zone-mars")
		nodeInfo := framework.NewNodeInfo(node, nil)

		estimatedNodes, estimatedPods := estimator.Estimate(podsEquivalenceGroup, nodeInfo, nil)
		assert.Equal(b, expectNodeCount, estimatedNodes)
		assert.Equal(b, expectPodCount, len(estimatedPods))
	}
}
