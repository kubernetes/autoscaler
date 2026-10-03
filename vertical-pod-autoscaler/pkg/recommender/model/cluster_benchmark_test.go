/*
Copyright The Kubernetes Authors.

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

package model

import (
	"fmt"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/test"
)

const benchmarkNamespace = "bench-namespace"

// newBenchmarkClusterState returns a clusterState tracking vpaCount VPAs, each
// selecting pods with the label app=app-<i>.
func newBenchmarkClusterState(b *testing.B, vpaCount int) *clusterState {
	b.Helper()
	cluster := NewClusterState(testGcPeriod)
	for i := range vpaCount {
		vpaObject := test.VerticalPodAutoscaler().
			WithNamespace(benchmarkNamespace).
			WithName(fmt.Sprintf("vpa-%d", i)).
			WithContainer("container-1").
			Get()
		selector := labels.SelectorFromSet(labels.Set{"app": fmt.Sprintf("app-%d", i)})
		if err := cluster.AddOrUpdateVpa(vpaObject, selector); err != nil {
			b.Fatalf("failed to add VPA: %v", err)
		}
	}
	return cluster
}

// BenchmarkClusterStateAddOrUpdatePod measures the cost of feeding pods into
// the cluster state. The cluster state feeder calls AddOrUpdatePod for every
// pod in the cluster on every recommender loop iteration:
//   - existing-pod: re-adding an already tracked pod with unchanged labels
//     (the steady-state resync path).
//   - new-pod: adding a previously unknown pod and deleting it again (the pod
//     churn path, e.g. after updater evictions). Its cost scales with the
//     number of VPAs, because new pods are matched against every VPA.
func BenchmarkClusterStateAddOrUpdatePod(b *testing.B) {
	podID := PodID{Namespace: benchmarkNamespace, PodName: "pod-1"}
	podLabels := labels.Set{"app": "app-0"}

	b.Run("existing-pod", func(b *testing.B) {
		cluster := newBenchmarkClusterState(b, 100)
		cluster.AddOrUpdatePod(podID, podLabels, corev1.PodRunning)
		b.ReportAllocs()
		for b.Loop() {
			cluster.AddOrUpdatePod(podID, podLabels, corev1.PodRunning)
		}
	})

	for _, vpaCount := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("new-pod/vpas=%d", vpaCount), func(b *testing.B) {
			cluster := newBenchmarkClusterState(b, vpaCount)
			b.ReportAllocs()
			for b.Loop() {
				cluster.AddOrUpdatePod(podID, podLabels, corev1.PodRunning)
				cluster.DeletePod(podID)
			}
		})
	}
}

// BenchmarkClusterStateAddSample measures the cost of aggregating a single
// container usage sample. The cluster state feeder calls AddSample for every
// container in the cluster on every metrics resync.
func BenchmarkClusterStateAddSample(b *testing.B) {
	for _, resourceName := range []ResourceName{ResourceCPU, ResourceMemory} {
		b.Run(string(resourceName), func(b *testing.B) {
			cluster := newBenchmarkClusterState(b, 1)
			podID := PodID{Namespace: benchmarkNamespace, PodName: "pod-1"}
			containerID := ContainerID{PodID: podID, ContainerName: "container-1"}
			cluster.AddOrUpdatePod(podID, labels.Set{"app": "app-0"}, corev1.PodRunning)
			if err := cluster.AddOrUpdateContainer(containerID, testRequest); err != nil {
				b.Fatalf("failed to add container: %v", err)
			}
			usage := CPUAmountFromCores(0.5)
			if resourceName == ResourceMemory {
				usage = MemoryAmountFromBytes(512 * 1024 * 1024)
			}
			timestamp := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
			b.ReportAllocs()
			for b.Loop() {
				timestamp = timestamp.Add(time.Second)
				err := cluster.AddSample(&ContainerUsageSampleWithKey{
					ContainerUsageSample: ContainerUsageSample{
						MeasureStart: timestamp,
						Usage:        usage,
						Resource:     resourceName,
					},
					Container: containerID,
				})
				if err != nil {
					b.Fatalf("failed to add sample: %v", err)
				}
			}
		})
	}
}

// BenchmarkVpaAggregateStateByContainerName measures the cost of merging the
// aggregations contributing to a VPA into per-container-name state. The
// recommender performs this merge for every VPA on every loop iteration,
// before computing the recommendation. Pods with distinct label sets (e.g.
// different pod-template-hash values across rolling updates) each contribute
// a separate aggregation that must be merged.
func BenchmarkVpaAggregateStateByContainerName(b *testing.B) {
	// One hour of per-minute usage samples per aggregation.
	const samplesPerAggregation = 60
	for _, labelSetCount := range []int{1, 10, 50} {
		b.Run(fmt.Sprintf("aggregations=%d", labelSetCount), func(b *testing.B) {
			cluster := NewClusterState(testGcPeriod)
			vpaObject := test.VerticalPodAutoscaler().
				WithNamespace(benchmarkNamespace).
				WithName("bench-vpa").
				WithContainer("container-1").
				Get()
			selector := labels.SelectorFromSet(labels.Set{"app": "bench"})
			if err := cluster.AddOrUpdateVpa(vpaObject, selector); err != nil {
				b.Fatalf("failed to add VPA: %v", err)
			}
			for i := range labelSetCount {
				podID := PodID{Namespace: benchmarkNamespace, PodName: fmt.Sprintf("pod-%d", i)}
				podLabels := labels.Set{"app": "bench", "pod-template-hash": fmt.Sprintf("hash-%d", i)}
				cluster.AddOrUpdatePod(podID, podLabels, corev1.PodRunning)
				containerID := ContainerID{PodID: podID, ContainerName: "container-1"}
				if err := cluster.AddOrUpdateContainer(containerID, testRequest); err != nil {
					b.Fatalf("failed to add container: %v", err)
				}
				timestamp := testTimestamp
				for j := range samplesPerAggregation {
					timestamp = timestamp.Add(time.Minute)
					err := cluster.AddSample(&ContainerUsageSampleWithKey{
						ContainerUsageSample: ContainerUsageSample{
							MeasureStart: timestamp,
							Usage:        CPUAmountFromCores(0.1 + 0.01*float64(j%10)),
							Resource:     ResourceCPU,
						},
						Container: containerID,
					})
					if err != nil {
						b.Fatalf("failed to add sample: %v", err)
					}
				}
			}
			vpa := cluster.vpas[VpaID{Namespace: benchmarkNamespace, VpaName: "bench-vpa"}]
			if got := len(vpa.aggregateContainerStates); got != labelSetCount {
				b.Fatalf("expected %d aggregations, got %d", labelSetCount, got)
			}
			b.ReportAllocs()
			for b.Loop() {
				merged := vpa.AggregateStateByContainerName()
				if len(merged) != 1 {
					b.Fatalf("expected aggregate state for 1 container name, got %d", len(merged))
				}
			}
		})
	}
}
