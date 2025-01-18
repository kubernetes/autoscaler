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
	"context"
	"time"

	k8sapiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/recommender/model"
	recommender_metrics "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics/recommender"
	"k8s.io/klog/v2"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// ContainerMetricsSnapshot contains information about usage of certain container within defined time window.
type ContainerMetricsSnapshot struct {
	// ID identifies a specific container those metrics are coming from.
	ID model.ContainerID
	// End time of the measurement interval.
	SnapshotTime time.Time
	// Duration of the measurement interval, which is [SnapshotTime - SnapshotWindow, SnapshotTime].
	SnapshotWindow time.Duration
	// Actual usage of the resources over the measurement interval.
	Usage model.Resources
}

// MetricsClient provides simple metrics on resources usage on container level.
type MetricsClient interface {
	// GetContainersMetrics returns an array of ContainerMetricsSnapshots,
	// representing resource usage for every running container in the cluster
	GetContainersMetrics(map[model.PodID]bool, bool) ([]*ContainerMetricsSnapshot, error)
}

type metricsClient struct {
	source     PodMetricsLister
	namespace  string
	clientName string
}

// NewMetricsClient creates new instance of MetricsClient, which is used by recommender.
// namespace limits queries to particular namespace, use k8sapiv1.NamespaceAll to select all namespaces.
func NewMetricsClient(source PodMetricsLister, namespace, clientName string) MetricsClient {
	return &metricsClient{
		source:     source,
		namespace:  namespace,
		clientName: clientName,
	}
}

func (c *metricsClient) GetContainersMetrics(podList map[model.PodID]bool, memorySaveMode bool) ([]*ContainerMetricsSnapshot, error) {
	var metricsSnapshots []*ContainerMetricsSnapshot

	podMetricsList, err := c.source.List(context.TODO(), c.namespace, metav1.ListOptions{})
	recommender_metrics.RecordMetricsServerResponse(err, c.clientName)
	if err != nil {
		return nil, err
	}
	klog.V(3).InfoS("podMetrics retrieved for all namespaces", "podMetrics", len(podMetricsList.Items))
	for _, podMetrics := range podMetricsList.Items {
		if !memorySaveMode {
			metricsSnapshotsForPod := createContainerMetricsSnapshots(podMetrics)
			metricsSnapshots = append(metricsSnapshots, metricsSnapshotsForPod...)
			continue
		}

		// only snapshot metrics for pod that has VPA enabled.
		podID := model.PodID{
			PodName:   podMetrics.ObjectMeta.Name,
			Namespace: podMetrics.ObjectMeta.Namespace,
		}
		if _, ok := podList[podID]; ok {
			metricsSnapshotsForPod := createContainerMetricsSnapshots(podMetrics)
			metricsSnapshots = append(metricsSnapshots, metricsSnapshotsForPod...)
		}
	}
	return metricsSnapshots, nil
}

func createContainerMetricsSnapshots(podMetrics v1beta1.PodMetrics) []*ContainerMetricsSnapshot {
	snapshots := make([]*ContainerMetricsSnapshot, len(podMetrics.Containers))
	for i, containerMetrics := range podMetrics.Containers {
		snapshots[i] = newContainerMetricsSnapshot(containerMetrics, podMetrics)
	}
	return snapshots
}

func newContainerMetricsSnapshot(containerMetrics v1beta1.ContainerMetrics, podMetrics v1beta1.PodMetrics) *ContainerMetricsSnapshot {
	usage := calculateUsage(containerMetrics.Usage)

	return &ContainerMetricsSnapshot{
		ID: model.ContainerID{
			ContainerName: containerMetrics.Name,
			PodID: model.PodID{
				Namespace: podMetrics.Namespace,
				PodName:   podMetrics.Name,
			},
		},
		Usage:          usage,
		SnapshotTime:   podMetrics.Timestamp.Time,
		SnapshotWindow: podMetrics.Window.Duration,
	}
}

func calculateUsage(containerUsage k8sapiv1.ResourceList) model.Resources {
	cpuQuantity := containerUsage[k8sapiv1.ResourceCPU]
	cpuMillicores := cpuQuantity.MilliValue()

	memoryQuantity := containerUsage[k8sapiv1.ResourceMemory]
	memoryBytes := memoryQuantity.Value()

	return model.Resources{
		model.ResourceCPU:    model.ResourceAmount(cpuMillicores),
		model.ResourceMemory: model.ResourceAmount(memoryBytes),
	}
}
