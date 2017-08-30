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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientapiv1 "k8s.io/client-go/pkg/api/v1"
	k8sapiv1 "k8s.io/kubernetes/pkg/api/v1"
)

// Utilization of resources for a single container, in a given (short) period of time
type ContainerUtilizationSnapshot struct {
	// Metadata identifying container
	ID containerID

	// Time when the snapshot was taken
	SnapshotTime time.Time
	// duration, which this snapshot represents
	SnapshotWindow time.Duration

	// When the container was created
	CreationTime time.Time
	// Image name running in the container
	Image string

	// Currently requested resources for this container
	Request clientapiv1.ResourceList
	// Actual usage of resources, during the SnapshotWindow.
	Usage clientapiv1.ResourceList
}

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
		Request:        convertResourceListToClientAPIType(spec.Request),
		Usage:          snap.Usage,
	}, nil
}

type containerUsageSnapshot struct {
	ID containerID

	SnapshotTime   metav1.Time
	SnapshotWindow metav1.Duration

	Usage clientapiv1.ResourceList
}

type containerSpec struct {
	ID           containerID
	CreationTime metav1.Time
	Image        string

	Request k8sapiv1.ResourceList
}

type containerID struct {
	Namespace     string
	PodName       string
	ContainerName string
}
