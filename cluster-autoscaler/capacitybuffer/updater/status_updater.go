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
	"time"

	v1 "k8s.io/autoscaler/cluster-autoscaler/apis/capacitybuffer/autoscaling.x-k8s.io/v1beta1"
	cbclient "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/client"
	cbmetrics "k8s.io/autoscaler/cluster-autoscaler/capacitybuffer/metrics"
	"k8s.io/utils/clock"
)

// StatusUpdater updates the buffer status bassed
type StatusUpdater struct {
	client *cbclient.CapacityBufferClient
	clock  clock.Clock
	processedBuffers *cbmetrics.ProcessingCache
}

// NewStatusUpdater creates an instance of StatusUpdater.
func NewStatusUpdater(client *cbclient.CapacityBufferClient, clock clock.Clock, processedBuffers *cbmetrics.ProcessingCache) *StatusUpdater {
	return &StatusUpdater{
		client: client,
		clock:  clock,
		processedBuffers: processedBuffers,
	}
}

// Update updates the buffer status with pod capacity
func (u *StatusUpdater) Update(buffers []*v1.CapacityBuffer) []error {
	var errors []error
	buffersUpdatedTime := map[string]time.Time{}

	for _, buffer := range buffers {
		updatedBuffer, err := u.client.UpdateCapacityBuffer(buffer)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		if updatedBuffer != nil {
			buffersUpdatedTime[string(updatedBuffer.UID)] = u.clock.Now()
		}
	}
	u.processedBuffers.Update(buffersUpdatedTime)

	return errors
}

// CleanUp cleans up the updater's internal structures.
func (u *StatusUpdater) CleanUp() {
}
