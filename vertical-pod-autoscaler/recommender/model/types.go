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

package model

import (
	"time"

	"github.com/golang/glog"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// MetricName represents the name of the resource monitored by recommender.
type MetricName string

// ResourceAmount represents quantity of a certain resource within a container.
type ResourceAmount int64

// Resources is a map from resource name to the corresponding ResourceAmount.
type Resources map[MetricName]ResourceAmount

const (
	// ResourceCPU represents CPU in millicores (1core = 1000millicores).
	ResourceCPU MetricName = "cpu"
	// ResourceMemory represents memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024).
	ResourceMemory MetricName = "memory"
)

// CPUAmountFromCores converts CPU cores to a ResourceAmount.
func CPUAmountFromCores(cores float64) ResourceAmount {
	return ResourceAmount(cores * 1000.0)
}

// MemoryAmountFromBytes converts memory bytes to a ResourceAmount.
func MemoryAmountFromBytes(bytes float64) ResourceAmount {
	return ResourceAmount(bytes)
}

// ResoucesAsResourceList converts internal Resources representation to ResourcesList.
func ResoucesAsResourceList(resources Resources) apiv1.ResourceList {
	result := make(apiv1.ResourceList)
	for key, value := range resources {
		var newKey apiv1.ResourceName
		switch key {
		case ResourceCPU:
			newKey = apiv1.ResourceCPU
		case ResourceMemory:
			newKey = apiv1.ResourceMemory
		default:
			glog.Errorf("Cannot translate %v resource name", key)
			continue
		}
		result[newKey] = *resource.NewScaledQuantity(int64(value), 0)
	}
	return result
}

// PodID contains information needed to identify a Pod within a cluster.
type PodID struct {
	// Namespaces where the Pod is defined.
	Namespace string
	// PodName is the name of the pod unique within a namespace.
	PodName string
}

// ContainerID contains information needed to identify a Container within a cluster.
type ContainerID struct {
	PodID
	// ContainerName is the name of the container, unique within a pod.
	ContainerName string
}

// VpaID contains information needed to identify a VPA API object within a cluster.
type VpaID struct {
	Namespace string
	VpaName   string
}

// ContainerMetricsSnapshot contains information about usage of certain container within defined time window.
type ContainerMetricsSnapshot struct {
	// ID identifies a specific container those metrics are coming from.
	ID ContainerID
	// End time of the measurement interval.
	SnapshotTime time.Time
	// Duration of the measurement interval, which is [SnapshotTime - SnapshotWindow, SnapshotTime].
	SnapshotWindow time.Duration
	// Actual usage of the resources over the measurement interval.
	Usage Resources
}

// BasicPodSpec contains basic information defining a pod and its containers.
type BasicPodSpec struct {
	// ID identifies a pod within a cluster.
	ID PodID
	// Labels of the pod. It is used to match pods with certain VPA opjects.
	PodLabels map[string]string
	// List of containers within this pod.
	Containers []BasicContainerSpec
}

// BasicContainerSpec contains basic information defining a container.
type BasicContainerSpec struct {
	// ID identifies the container within a cluster.
	ID ContainerID
	// Name of the image running within the container.
	Image string
	// Currently requested resources for this container.
	Request Resources
}
