/*
Copyright 2017 The Kubernetes Authors.

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

package metrics

import (
	"math/big"
	"time"

	k8sapiv1 "k8s.io/api/core/v1"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	core "k8s.io/client-go/testing"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	metricsapi "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"k8s.io/metrics/pkg/client/clientset/versioned/fake"
)

type metricsClientTestCase struct {
	snapshotTimestamp    time.Time
	snapshotWindow       time.Duration
	namespace            *v1.Namespace
	pod1Snaps, pod2Snaps []*ContainerMetricsSnapshot
}

func newMetricsClientTestCase() *metricsClientTestCase {
	namespaceName := "test-namespace"

	testCase := &metricsClientTestCase{
		snapshotTimestamp: time.Now(),
		snapshotWindow:    time.Duration(1234),
		namespace:         &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespaceName}},
	}

	id1 := model.ContainerID{PodID: model.PodID{Namespace: namespaceName, PodName: "Pod1"}, ContainerName: "Name1"}
	id2 := model.ContainerID{PodID: model.PodID{Namespace: namespaceName, PodName: "Pod1"}, ContainerName: "Name2"}
	id3 := model.ContainerID{PodID: model.PodID{Namespace: namespaceName, PodName: "Pod2"}, ContainerName: "Name1"}
	id4 := model.ContainerID{PodID: model.PodID{Namespace: namespaceName, PodName: "Pod2"}, ContainerName: "Name2"}

	testCase.pod1Snaps = append(testCase.pod1Snaps, testCase.newContainerMetricsSnapshot(id1, 400, 333))
	testCase.pod1Snaps = append(testCase.pod1Snaps, testCase.newContainerMetricsSnapshot(id2, 800, 666))
	testCase.pod2Snaps = append(testCase.pod2Snaps, testCase.newContainerMetricsSnapshot(id3, 401, 334))
	testCase.pod2Snaps = append(testCase.pod2Snaps, testCase.newContainerMetricsSnapshot(id4, 801, 667))

	return testCase
}

func newEmptyMetricsClientTestCase() *metricsClientTestCase {
	return &metricsClientTestCase{}
}

func (tc *metricsClientTestCase) newContainerMetricsSnapshot(id model.ContainerID, cpuUsage int64, memUsage int64) *ContainerMetricsSnapshot {
	return &ContainerMetricsSnapshot{
		ID:             id,
		SnapshotTime:   tc.snapshotTimestamp,
		SnapshotWindow: tc.snapshotWindow,
		Usage: model.Resources{
			model.ResourceCPU:    model.ResourceAmount(cpuUsage),
			model.ResourceMemory: model.ResourceAmount(memUsage),
		},
	}
}

func (tc *metricsClientTestCase) createFakeMetricsClient() MetricsClient {
	fakeMetricsGetter := &fake.Clientset{}
	fakeMetricsGetter.AddReactor("list", "pods", func(action core.Action) (handled bool, ret runtime.Object, err error) {
		return true, tc.getFakePodMetricsList(), nil
	})
	return NewMetricsClient(fakeMetricsGetter.MetricsV1beta1())
}

func (tc *metricsClientTestCase) getFakePodMetricsList() *metricsapi.PodMetricsList {
	metrics := &metricsapi.PodMetricsList{}
	if tc.pod1Snaps != nil && tc.pod2Snaps != nil {
		metrics.Items = append(metrics.Items, makePodMetrics(tc.pod1Snaps))
		metrics.Items = append(metrics.Items, makePodMetrics(tc.pod2Snaps))
	}
	return metrics
}

func makePodMetrics(snaps []*ContainerMetricsSnapshot) metricsapi.PodMetrics {
	firstSnap := snaps[0]
	podMetrics := metricsapi.PodMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: firstSnap.ID.Namespace,
			Name:      firstSnap.ID.PodName,
		},
		Timestamp:  metav1.Time{Time: firstSnap.SnapshotTime},
		Window:     metav1.Duration{Duration: firstSnap.SnapshotWindow},
		Containers: make([]metricsapi.ContainerMetrics, len(snaps)),
	}

	for i, snap := range snaps {
		resourceList := calculateResourceList(snap.Usage)
		podMetrics.Containers[i] = metricsapi.ContainerMetrics{
			Name:  snap.ID.ContainerName,
			Usage: resourceList,
		}
	}
	return podMetrics
}

func calculateResourceList(usage model.Resources) k8sapiv1.ResourceList {
	cpuCores := big.NewRat(int64(usage[model.ResourceCPU]), 1000)
	cpuQuantityString := cpuCores.FloatString(3)

	memoryBytes := big.NewInt(int64(usage[model.ResourceMemory]))
	memoryQuantityString := memoryBytes.String()

	resourceMap := map[k8sapiv1.ResourceName]resource.Quantity{
		k8sapiv1.ResourceCPU:    resource.MustParse(cpuQuantityString),
		k8sapiv1.ResourceMemory: resource.MustParse(memoryQuantityString),
	}
	return k8sapiv1.ResourceList(resourceMap)
}

func (tc *metricsClientTestCase) getAllSnaps() []*ContainerMetricsSnapshot {
	return append(tc.pod1Snaps, tc.pod2Snaps...)
}
