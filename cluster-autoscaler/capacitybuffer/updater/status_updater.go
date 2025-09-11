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

package updater

import (
	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1"
	client "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/client/clientset/versioned"
	common "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/common"
)

// StatusUpdater updates the buffer status bassed
type StatusUpdater struct {
	client client.Interface
}

// NewStatusUpdater creates an instance of StatusUpdater.
func NewStatusUpdater(client client.Interface) *StatusUpdater {
	return &StatusUpdater{
		client: client,
	}
}

// Update updates the buffer status with pod capacity
func (u *StatusUpdater) Update(buffers []*v1.CapacityBuffer) []error {
	var errors []error
	for _, buffer := range buffers {
		err := common.UpdateBufferStatus(u.client, buffer)
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

// CleanUp cleans up the updater's internal structures.
func (u *StatusUpdater) CleanUp() {
}
