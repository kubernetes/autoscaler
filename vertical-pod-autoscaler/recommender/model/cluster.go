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
	labels "k8s.io/apimachinery/pkg/labels"
)

// Cluster holds all runtime information about resources in the cluster
// required for VPA operations.
type Cluster struct {
	// Pods in the cluster.
	Pods map[PodID]*Pod
}

// Pod holds runtime information about a single Pod.
type Pod struct {
	// Unique id of the Pod.
	ID PodID
	// Set of labels attached to the Pod.
	Labels labels.Set
	// Containers that belong to the Pod, keyed by the container name.
	Containers map[string]*ContainerStats
}

// NewCluster returns a new Cluster with no pods.
func NewCluster() *Cluster {
	return &Cluster{
		make(map[PodID]*Pod), // empty pods map.
	}
}

// ContainerUsageSampleWithKey holds a ContainerUsageSample together with the
// ID of the container it belongs to.
type ContainerUsageSampleWithKey struct {
	ContainerUsageSample
	Container ContainerID
}

// AddOrUpdatePod udpates the state of the pod with a given PodID, if it is
// present in the cluster object. Otherwise a new pod is created and added to
// the Cluster object.
func (cluster *Cluster) AddOrUpdatePod(podID PodID, labels labels.Set) {
	if _, podExists := cluster.Pods[podID]; !podExists {
		cluster.Pods[podID] = newPod(podID)
	}
	cluster.Pods[podID].Labels = labels
}

// AddOrUpdateContainer creates a new container with the given ContainerID and
// adds it to the parent pod in the Cluster object, if not yet present.
// Requires the pod to be added to the Cluster first. Otherwise an error is
// returned.
func (cluster *Cluster) AddOrUpdateContainer(containerID ContainerID) error {
	pod, podExists := cluster.Pods[containerID.PodID]
	if !podExists {
		return NewKeyError(containerID.PodID)
	}
	if _, containerExists := pod.Containers[containerID.ContainerName]; !containerExists {
		pod.Containers[containerID.ContainerName] = NewContainerStats()
	}
	return nil
}

// AddSample adds a new usage sample to the proper container in the Cluster
// object. Requires the container as well as the parent pod to be added to the
// Cluster first. Otherwise an error is returned.
func (cluster *Cluster) AddSample(sample *ContainerUsageSampleWithKey) error {
	pod, podExists := cluster.Pods[sample.Container.PodID]
	if !podExists {
		return NewKeyError(sample.Container.PodID)
	}
	containerStats, containerExists := pod.Containers[sample.Container.ContainerName]
	if !containerExists {
		return NewKeyError(sample.Container)
	}
	containerStats.AddSample(&sample.ContainerUsageSample)
	return nil
}

func newPod(id PodID) *Pod {
	return &Pod{
		id,
		make(map[string]string),          // empty labels.
		make(map[string]*ContainerStats), // empty containers.
	}
}
