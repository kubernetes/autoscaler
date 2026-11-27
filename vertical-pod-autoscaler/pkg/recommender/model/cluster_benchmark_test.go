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
	"context"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func BenchmarkGetContributiveAggregateStateKeysActivePods(b *testing.B) {
	podCounts := []int{1, 10, 100, 1_000, 10_000}
	for _, podCount := range podCounts {
		b.Run(fmt.Sprintf("pods=%d", podCount), func(b *testing.B) {
			cluster := NewClusterState(testGcPeriod)
			addVpa(cluster, testVpaID, nil, "c-1=p-1", testTargetRef)
			for i := range podCount {
				podID := PodID{Namespace: testVpaID.Namespace, PodName: fmt.Sprintf("pod-%d", i)}
				podLabels := labels.Set{"c-1": "p-1", "pod": podID.PodName}
				cluster.AddOrUpdatePod(podID, podLabels, corev1.PodRunning)
				if err := cluster.AddOrUpdateContainer(ContainerID{PodID: podID, ContainerName: "c-1"}, testRequest); err != nil {
					b.Fatal(err)
				}
			}

			ctx := context.Background()
			if got := len(cluster.getContributiveAggregateStateKeys(ctx, testControllerFetcher)); got != podCount {
				b.Fatalf("got %d aggregate state keys, want %d", got, podCount)
			}
			b.ReportAllocs()
			for b.Loop() {
				cluster.getContributiveAggregateStateKeys(ctx, testControllerFetcher)
			}
		})
	}
}

func BenchmarkAddOrUpdatePodUnchangedLabels(b *testing.B) {
	cluster := NewClusterState(testGcPeriod)
	podLabels := labels.Set{}
	for i := range 12 {
		podLabels[fmt.Sprintf("label-%d", i)] = fmt.Sprintf("value-%d", i)
	}
	cluster.AddOrUpdatePod(testPodID, podLabels, corev1.PodRunning)

	b.ReportAllocs()
	for b.Loop() {
		cluster.AddOrUpdatePod(testPodID, podLabels, corev1.PodRunning)
	}
}
