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
	"time"

	k8sapiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ContainerUtilizationSnapshot represents utilization of resources for a single container, in a given (short) period of time
type ContainerUtilizationSnapshot struct {
	// Metadata identifying container
	ID containerID

	// Time when the snapshot was taken
	SnapshotTime time.Time
	// duration, which this snapshot represents
	SnapshotWindow time.Duration

	// Labels of the pod container belongs to. It is used to match pods with certain VPA opjects.
	PodLabels map[string]string
	// When the container was created
	CreationTime time.Time
	// Image name running in the container
	Image string

	// Currently requested resources for this container
	Request k8sapiv1.ResourceList
	// Actual usage of resources, during the SnapshotWindow.
	Usage k8sapiv1.ResourceList
}

// NewContainerUtilizationSnapshot merges containerUsageSnapshot and containerSpec into single ContainerUtilizationSnapshot.
// Both snap and spec need to have the same container name, pod name and namespace; otherwise error will be returned.
func NewContainerUtilizationSnapshot(snap *containerUsageSnapshot, spec *containerSpec) (*ContainerUtilizationSnapshot, error) {
	if snap.ID.PodName != spec.ID.PodName || snap.ID.ContainerName != spec.ID.ContainerName || snap.ID.Namespace != spec.ID.Namespace {
		return nil, errors.New("spec and snap are from different containers!")
	}
	return &ContainerUtilizationSnapshot{
		ID:             spec.ID,
		CreationTime:   spec.CreationTime.Time,
		Image:          spec.Image,
		SnapshotTime:   snap.SnapshotTime.Time,
		SnapshotWindow: snap.SnapshotWindow.Duration,
		Request:        spec.Request,
		Usage:          snap.Usage,
		PodLabels:      spec.PodLabels,
	}, nil
}

// information about usage of certain container withing defined time window
type containerUsageSnapshot struct {
	ID containerID

	SnapshotTime   metav1.Time
	SnapshotWindow metav1.Duration

	Usage k8sapiv1.ResourceList
}

// Basic info about container specification
type containerSpec struct {
	ID containerID

	CreationTime metav1.Time
	Image        string
	PodLabels    map[string]string

	Request k8sapiv1.ResourceList
}

type containerID struct {
	Namespace     string
	PodName       string
	ContainerName string
}
