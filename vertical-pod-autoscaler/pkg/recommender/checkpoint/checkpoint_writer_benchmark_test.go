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

package checkpoint

import (
	"context"
	"fmt"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"

	vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_api "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/autoscaling.k8s.io/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
)

// benchmarkCheckpointClient allows updates without api latency, locks and other
// side effects that could inflate the benchmark results.
type benchmarkCheckpointClient struct {
	vpa_api.VerticalPodAutoscalerCheckpointInterface
}

func (c *benchmarkCheckpointClient) VerticalPodAutoscalerCheckpoints(string) vpa_api.VerticalPodAutoscalerCheckpointInterface {
	return c
}

func (*benchmarkCheckpointClient) Patch(context.Context, string, types.PatchType, []byte, metav1.PatchOptions, ...string) (*vpa_types.VerticalPodAutoscalerCheckpoint, error) {
	return nil, nil
}

// BenchmarkStoreCheckpoints measures a complete checkpoint pass, including VPA
// sorting, pod scanning, histogram aggregation, serialization, and worker dispatch.
// Each VPA matches five pods with two containers each. Setup is excluded from timing.
func BenchmarkStoreCheckpoints(b *testing.B) {
	testCases := []struct {
		vpaCount int
		workers  int
	}{
		{vpaCount: 1, workers: 1},
		{vpaCount: 1, workers: 4},
		{vpaCount: 10, workers: 1},
		{vpaCount: 10, workers: 4},
		{vpaCount: 100, workers: 1},
		{vpaCount: 100, workers: 4},
		{vpaCount: 1000, workers: 1},
		{vpaCount: 1000, workers: 4},
	}
	for _, tc := range testCases {
		b.Run(fmt.Sprintf("vpas=%d/workers=%d", tc.vpaCount, tc.workers), func(b *testing.B) {
			cluster := newCheckpointBenchmarkCluster(b, tc.vpaCount)
			containerCount := 0
			for _, pod := range cluster.Pods() {
				containerCount += len(pod.Containers)
			}
			writer := NewCheckpointWriter(cluster, &benchmarkCheckpointClient{})
			ctx := context.Background()

			writer.StoreCheckpoints(ctx, tc.workers)
			for id, vpa := range cluster.VPAs() {
				if vpa.CheckpointWritten.IsZero() {
					b.Fatalf("no checkpoint written for %v", id)
				}
			}

			b.ReportAllocs()
			for b.Loop() {
				writer.StoreCheckpoints(ctx, tc.workers)
			}
			b.ReportMetric(float64(len(cluster.VPAs()))*float64(containerCount), "container-visits/op")
		})
	}
}

func newCheckpointBenchmarkCluster(b *testing.B, vpaCount int) model.ClusterState {
	b.Helper()
	cluster := model.NewClusterState(time.Minute)
	now := time.Now()
	interval := model.GetAggregationsConfig().MemoryAggregationIntervalDuration
	for i := range vpaCount {
		app := fmt.Sprintf("app-%d", i)
		vpa := &vpa_types.VerticalPodAutoscaler{
			ObjectMeta: metav1.ObjectMeta{Namespace: "benchmark", Name: app},
		}
		if err := cluster.AddOrUpdateVpa(vpa, labels.SelectorFromSet(labels.Set{"app": app})); err != nil {
			b.Fatal(err)
		}

		// Create 5 pods per VPA.
		for p := range 5 {
			podID := model.PodID{Namespace: "benchmark", PodName: fmt.Sprintf("%s-pod-%d", app, p)}
			cluster.AddOrUpdatePod(podID, map[string]string{"app": app}, corev1.PodRunning)

			// Create two containers per pod.
			for c := range 2 {
				containerID := model.ContainerID{PodID: podID, ContainerName: fmt.Sprintf("container-%d", c)}
				if err := cluster.AddOrUpdateContainer(containerID, model.Resources{}); err != nil {
					b.Fatal(err)
				}
				// Include completed memory intervals and the current peak, which
				// the checkpoint writer must subtract from the saved histogram.
				for sample := range 8 {
					for _, resource := range []model.ResourceName{model.ResourceCPU, model.ResourceMemory} {
						usage := model.CPUAmountFromCores(0.1 * float64(sample+1))
						if resource == model.ResourceMemory {
							usage = model.MemoryAmountFromBytes(float64(128+sample*16) * 1024 * 1024)
						}
						if !cluster.GetContainer(containerID).AddSample(&model.ContainerUsageSample{
							MeasureStart: now.Add(-time.Duration(7-sample) * interval),
							Usage:        usage,
							Resource:     resource,
						}) {
							b.Fatal("usage sample discarded")
						}
					}
				}
			}
		}
	}
	return cluster
}
