/*
Copyright 2020 The Kubernetes Authors.

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

package logic

import (
	"time"

	"k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/status"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

const (
	updateInterval = 10 * time.Second
)

// StatusUpdater periodically updates Admission Controller status.
type StatusUpdater struct {
	client *status.Client
}

// NewStatusUpdater returns a new status updater.
func NewStatusUpdater(c clientset.Interface, holderIdentity string) *StatusUpdater {
	return &StatusUpdater{
		client: status.NewClient(
			c,
			status.AdmissionControllerStatusName,
			status.AdmissionControllerStatusNamespace,
			updateInterval,
			holderIdentity,
		),
	}
}

// Run starts status updates.
func (su *StatusUpdater) Run(stopCh <-chan struct{}) {
	go func() {
		for {
			select {
			case <-stopCh:
				return
			case <-time.After(updateInterval):
				if err := su.client.UpdateStatus(); err != nil {
					klog.Errorf("Admission Controller status update failed: %v", err)
				}
			}
		}
	}()
}
