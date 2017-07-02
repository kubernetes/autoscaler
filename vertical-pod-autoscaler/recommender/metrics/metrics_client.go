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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1lister "k8s.io/kubernetes/pkg/client/listers/core/v1"
	resourceclient "k8s.io/metrics/pkg/client/clientset_generated/clientset/typed/metrics/v1alpha1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1alpha1"
)

type MetricsClient interface {
	GetContainersUtilization() ([]*ContainerUtilizationSnapshot, error)
}

type metricsClient struct {
	metricsGetter   resourceclient.PodMetricsesGetter
	podLister       v1lister.PodLister
	namespaceLister v1lister.NamespaceLister
}

func NewMetricsClient(metricsGetter resourceclient.PodMetricsesGetter, podLister v1lister.PodLister, namespaceLister v1lister.NamespaceLister) MetricsClient {
	return &metricsClient{
		metricsGetter:   metricsGetter,
		podLister:       podLister,
		namespaceLister: namespaceLister,
	}
}

func (client *metricsClient) GetContainersUtilization() ([]*ContainerUtilizationSnapshot, error) {

	usageSnapshots, err := client.getContainersUsage()
	containerSpecs, err := client.getContainersSpec()
	if err != nil {
		return nil, err
	}

	utilizationSnapshots, err := mergeIntoUtilization(usageSnapshots, containerSpecs)
	if err != nil {
		return nil, err
	}
	return utilizationSnapshots, nil
}

func (client *metricsClient) getContainersSpec() ([]*containerSpecification, error) {

	var containerSpecs []*containerSpecification

	pods, err := client.podLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			containerSpec := newContainerSpec(container, pod)
			containerSpecs = append(containerSpecs, containerSpec)
		}
	}

	return containerSpecs, nil

}

func (client *metricsClient) getContainersUsage() ([]*containerUsageSnapshot, error) {

	var usageSnapshots []*containerUsageSnapshot

	namespaces, err := client.getAllNamespaces()
	if err != nil {
		return nil, err
	}
	for _, namespace := range namespaces {
		podMetricsList, err := client.metricsGetter.PodMetricses(namespace).List(metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for _, podMetrics := range podMetricsList.Items {
			containerSnapshots := createContainerUsageSnapshots(podMetrics)
			usageSnapshots = append(usageSnapshots, containerSnapshots...)
		}
	}

	return usageSnapshots, nil
}

func createContainerUsageSnapshots(podMetrics v1alpha1.PodMetrics) []*containerUsageSnapshot {
	snapshots := make([]*containerUsageSnapshot, len(podMetrics.Containers))
	for i, containerMetrics := range podMetrics.Containers {
		snapshots[i] = newContainerUsageSnapshot(containerMetrics, podMetrics)
	}
	return snapshots
}

func mergeIntoUtilization(snapshots []*containerUsageSnapshot, specifications []*containerSpecification) ([]*ContainerUtilizationSnapshot, error) {
	specsMap := make(map[containerId]*containerSpecification, len(specifications))
	for _, spec := range specifications {
		specsMap[spec.Id] = spec
	}

	result := make([]*ContainerUtilizationSnapshot, len(snapshots))

	for i, snap := range snapshots {
		spec := specsMap[snap.Id]
		utilizationSnapshot, err := NewContainerUtilizationSnapshot(snap, spec)
		if err == nil {
			result[i] = utilizationSnapshot
		} else {
			return nil, err
		}
	}

	return result, nil
}

func newContainerSpec(container v1.Container, pod *v1.Pod) *containerSpecification {
	return &containerSpecification{
		Id: containerId{
			PodName:       pod.Name,
			Namespace:     pod.Namespace,
			ContainerName: container.Name,
		},

		CreationTime: pod.CreationTimestamp,
		Image:        container.Image,
		Request:      container.Resources.Requests,
	}
}

func newContainerUsageSnapshot(containerMetrics v1alpha1.ContainerMetrics, podMetrics v1alpha1.PodMetrics) *containerUsageSnapshot {
	return &containerUsageSnapshot{
		Id: containerId{
			ContainerName: containerMetrics.Name,
			Namespace:     podMetrics.Namespace,
			PodName:       podMetrics.Name,
		},
		Usage:          containerMetrics.Usage,
		SnapshotTime:   podMetrics.Timestamp,
		SnapshotWindow: podMetrics.Window,
	}
}

func (client *metricsClient) getAllNamespaces() ([]string, error) {
	namespaces, err := client.namespaceLister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	count := len(namespaces)
	result := make([]string, count)

	for i, namespace := range namespaces {
		result[i] = namespace.Name
	}

	return result, nil
}
