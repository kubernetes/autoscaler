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
	"context"
	"time"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
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
		ticker := time.NewTicker(su.updateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				su.updateStatus()
			}
		}
	}()
}

func (su *Updater) updateStatus() {
	// The lease renewal has timeout of the given update interval (10s).
	// If the lease cannot be renewed in the allotted time (for example due to client-side throttling),
	// the vpa-updater will consider that vpa-admission-controller as unhealthy and won't evict Pods in endless loop.
	ctx, cancel := context.WithTimeout(context.Background(), su.updateInterval)
	defer cancel()

	if err := su.client.UpdateStatus(ctx); err != nil {
		klog.Errorf("Status update by %s failed: %v", su.client.holderIdentity, err)
	}
}
