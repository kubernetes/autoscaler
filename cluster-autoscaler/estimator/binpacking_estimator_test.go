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
	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"
	. "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	"github.com/stretchr/testify/assert"
)

func makePodEquivalenceGroup(cpuPerPod int64, memoryPerPod int64, hostport int32, maxSkew int32, topologySpreadingKey string, podCount int) PodEquivalenceGroup {
	pod := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "estimatee",
			Namespace: "universe",
			Labels: map[string]string{
				"app": "estimatee",
			},
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Resources: apiv1.ResourceRequirements{
						Requests: apiv1.ResourceList{
							apiv1.ResourceCPU:    *resource.NewMilliQuantity(cpuPerPod, resource.DecimalSI),
							apiv1.ResourceMemory: *resource.NewQuantity(memoryPerPod*units.MiB, resource.DecimalSI),
						},
					},
				},
			},
		},
	}
	if hostport > 0 {
		pod.Spec.Containers[0].Ports = []apiv1.ContainerPort{
			{
				HostPort: hostport,
			},
		}
	}
	if maxSkew > 0 {
		pod.Spec.TopologySpreadConstraints = []apiv1.TopologySpreadConstraint{
			{
				MaxSkew:           maxSkew,
				TopologyKey:       topologySpreadingKey,
				WhenUnsatisfiable: "DoNotSchedule",
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "estimatee",
					},
				},
			},
		}
	}
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
	highResourcePodGroup := makePodEquivalenceGroup(500, 1000, 0, 0, "", 10)
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
			name:                 "simple resource-based binpacking",
			millicores:           350*3 - 50,
			memory:               2 * 1000,
			podsEquivalenceGroup: []PodEquivalenceGroup{makePodEquivalenceGroup(350, 1000, 0, 0, "", 10)},
			expectNodeCount:      5,
			expectPodCount:       10,
		},
		{
			name:                 "pods-per-node bound binpacking",
			millicores:           10000,
			memory:               20000,
			podsEquivalenceGroup: []PodEquivalenceGroup{makePodEquivalenceGroup(10, 100, 0, 0, "", 20)},
			expectNodeCount:      2,
			expectPodCount:       20,
		},
		{
			name:                 "hostport conflict forces pod-per-node",
			millicores:           1000,
			memory:               5000,
			podsEquivalenceGroup: []PodEquivalenceGroup{makePodEquivalenceGroup(200, 1000, 5555, 0, "", 8)},
			expectNodeCount:      8,
			expectPodCount:       8,
		},
		{
			name:                 "limiter cuts binpacking",
			millicores:           1000,
			memory:               5000,
			podsEquivalenceGroup: []PodEquivalenceGroup{makePodEquivalenceGroup(500, 1000, 0, 0, "", 20)},
			maxNodes:             5,
			expectNodeCount:      5,
			expectPodCount:       10,
		},
		{
			name:                 "decreasing ordered pods are processed first",
			millicores:           1000,
			memory:               5000,
			podsEquivalenceGroup: append([]PodEquivalenceGroup{makePodEquivalenceGroup(50, 1000, 0, 0, "", 10)}, highResourcePodGroup),
			maxNodes:             5,
			expectNodeCount:      5,
			expectPodCount:       10,
			expectProcessedPods:  highResourcePodGroup.Pods,
		},
		{
			name:                 "hostname topology spreading with maxSkew=2 forces 2 pods/node",
			millicores:           1000,
			memory:               5000,
			podsEquivalenceGroup: []PodEquivalenceGroup{makePodEquivalenceGroup(20, 100, 0, 2, "kubernetes.io/hostname", 8)},
			expectNodeCount:      4,
			expectPodCount:       8,
		},
		{
			name:                 "zonal topology spreading with maxSkew=2 only allows 2 pods to schedule",
			millicores:           1000,
			memory:               5000,
			podsEquivalenceGroup: []PodEquivalenceGroup{makePodEquivalenceGroup(20, 100, 0, 2, "topology.kubernetes.io/zone", 8)},
			expectNodeCount:      1,
			expectPodCount:       2,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			clusterSnapshot := clustersnapshot.NewBasicClusterSnapshot()
			// Add one node in different zone to trigger topology spread constraints
			clusterSnapshot.AddNode(makeNode(100, 100, 10, "oldnode", "zone-jupiter"))

			predicateChecker, err := predicatechecker.NewTestPredicateChecker()
			assert.NoError(t, err)
			limiter := NewThresholdBasedEstimationLimiter([]Threshold{NewStaticThreshold(tc.maxNodes, time.Duration(0))})
			processor := NewDecreasingPodOrderer()
			estimator := NewBinpackingNodeEstimator(predicateChecker, clusterSnapshot, limiter, processor, nil /* EstimationContext */, nil /* EstimationAnalyserFunc */)
			node := makeNode(tc.millicores, tc.memory, 10, "template", "zone-mars")
			nodeInfo := schedulerframework.NewNodeInfo()
			nodeInfo.SetNode(node)

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
	podsEquivalenceGroup := []PodEquivalenceGroup{makePodEquivalenceGroup(50, 100, 0, 0, "", 50000), makePodEquivalenceGroup(95, 190, 0, 0, "", 1000)}

	for i := 0; i < b.N; i++ {
		clusterSnapshot := clustersnapshot.NewBasicClusterSnapshot()
		clusterSnapshot.AddNode(makeNode(100, 100, 10, "oldnode", "zone-jupiter"))

		predicateChecker, err := predicatechecker.NewTestPredicateChecker()
		assert.NoError(b, err)
		limiter := NewThresholdBasedEstimationLimiter([]Threshold{NewStaticThreshold(maxNodes, time.Duration(0))})
		processor := NewDecreasingPodOrderer()
		estimator := NewBinpackingNodeEstimator(predicateChecker, clusterSnapshot, limiter, processor, nil /* EstimationContext */, nil /* EstimationAnalyserFunc */)
		node := makeNode(millicores, memory, podsPerNode, "template", "zone-mars")
		nodeInfo := schedulerframework.NewNodeInfo()
		nodeInfo.SetNode(node)

		estimatedNodes, estimatedPods := estimator.Estimate(podsEquivalenceGroup, nodeInfo, nil)
		assert.Equal(b, expectNodeCount, estimatedNodes)
		assert.Equal(b, expectPodCount, len(estimatedPods))
	}
}
