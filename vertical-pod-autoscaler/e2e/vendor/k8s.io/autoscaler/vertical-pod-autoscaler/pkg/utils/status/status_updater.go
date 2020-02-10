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

package status

import (
	"time"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

// Updater periodically updates status object.
type Updater struct {
	client         *Client
	updateInterval time.Duration
}

// NewUpdater returns a new status updater.
func NewUpdater(c clientset.Interface, statusName, statusNamespace string,
	updateInterval time.Duration, holderIdentity string) *Updater {
	return &Updater{
		client: NewClient(
			c,
			statusName,
			statusNamespace,
			updateInterval,
			holderIdentity,
		),
		updateInterval: updateInterval,
	}
}

// Run starts status updates.
func (su *Updater) Run(stopCh <-chan struct{}) {
	go func() {
		for {
			select {
			case <-stopCh:
				return
			case <-time.After(su.updateInterval):
				if err := su.client.UpdateStatus(); err != nil {
					klog.Errorf("Status update by %s failed: %v", su.client.holderIdentity, err)
				}
			}
		}
	}()
}
