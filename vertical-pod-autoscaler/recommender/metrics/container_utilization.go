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
	"errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientapiv1 "k8s.io/client-go/pkg/api/v1"
	k8sapiv1 "k8s.io/kubernetes/pkg/api/v1"
	"time"
)

type ContainerUtilizationSnapshot struct {
	Id containerId

	SnapshotTime   time.Time
	SnapshotWindow time.Duration

	CreationTime time.Time
	Image        string

	Request clientapiv1.ResourceList
	Usage   clientapiv1.ResourceList
}

func NewContainerUtilizationSnapshot(snap *containerUsageSnapshot, spec *containerSpecification) (*ContainerUtilizationSnapshot, error) {
	if snap.Id.PodName != spec.Id.PodName || snap.Id.ContainerName != spec.Id.ContainerName || snap.Id.Namespace != spec.Id.Namespace {
		return nil, errors.New("Specification and Snapshot are comming from different containers!")
	}
	return &ContainerUtilizationSnapshot{
		Id:             spec.Id,
		CreationTime:   spec.CreationTime.Time,
		Image:          spec.Image,
		SnapshotTime:   snap.SnapshotTime.Time,
		SnapshotWindow: snap.SnapshotWindow.Duration,
		Request:        convertToClientApi(spec.Request),
		Usage:          snap.Usage,
	}, nil
}

// type conversion is needed, since ResourceList is coming from different packages in containerUsageSnapshot and containerSpecification
func convertToClientApi(input k8sapiv1.ResourceList) clientapiv1.ResourceList {
	output := make(clientapiv1.ResourceList, len(input))
	for name, quantity := range input {
		var newName clientapiv1.ResourceName = clientapiv1.ResourceName(name.String())
		output[newName] = quantity
	}
	return output
}

func convertToKubernetesApi(input clientapiv1.ResourceList) k8sapiv1.ResourceList {
	output := make(k8sapiv1.ResourceList, len(input))
	for name, quantity := range input {
		var newName k8sapiv1.ResourceName = k8sapiv1.ResourceName(name.String())
		output[newName] = quantity
	}
	return output
}

type containerUsageSnapshot struct {
	Id containerId

	SnapshotTime   metav1.Time
	SnapshotWindow metav1.Duration

	Usage clientapiv1.ResourceList
}

type containerSpecification struct {
	Id           containerId
	CreationTime metav1.Time
	Image        string

	Request k8sapiv1.ResourceList
}

type containerId struct {
	Namespace     string
	PodName       string
	ContainerName string
}
