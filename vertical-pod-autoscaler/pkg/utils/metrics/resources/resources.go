/*
Copyright 2025 The Kubernetes Authors.

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

// Package resources contains metrics for the VPA resource helper functions.
package resources

import (
	"github.com/prometheus/client_golang/prometheus"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/metrics"
)

const (
	metricsNamespace = metrics.TopMetricsNamespace + "resources"
)

type resourcesSource string

const (
	// ContainerStatus indicates that resources were fetched from the
	// containerStatuses pod field.
	ContainerStatus resourcesSource = "Pod.Status.ContainerStatuses"
	// InitContainerStatus indicates that resources were fetched from the
	// initContainerStatuses pod field.
	InitContainerStatus resourcesSource = "Pod.Status.InitContainerStatuses"
	// PodSpecContainer indicates that resources were fetched from the
	// containers field in the pod spec.
	PodSpecContainer resourcesSource = "Pod.Spec.Containers"
	// PodSpecInitContainer indicates that resources were fetched from the
	// initContainers field in the pod spec.
	PodSpecInitContainer resourcesSource = "Pod.Spec.InitContainers"
)

var (
	getResourcesCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricsNamespace,
			Name:      "get_resources_count",
			Help:      "Count of calls to get the resources of a pod or container.",
		}, []string{"source"},
	)
)

// Register initializes all metrics for VPA resources
func Register() {
	_ = prometheus.Register(getResourcesCount)
}

// RecordGetResourcesCount records how many times VPA requested the resources (
// CPU/memory requests and limits) of a pod or container by the data source.
func RecordGetResourcesCount(source resourcesSource) {
	getResourcesCount.WithLabelValues(string(source)).Inc()
}
