/*
Copyright 2024 The Kubernetes Authors.

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

package provreq

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/provisioningrequest/provreqclient"
	"k8s.io/utils/clock/testing"
	"k8s.io/utils/lru"
)

// NewFakePodsInjector creates a new instance of ProvisioningRequestPodsInjector with the given client and clock for testing.
func NewFakePodsInjector(client *provreqclient.ProvisioningRequestClient, clock *testing.FakePassiveClock) *ProvisioningRequestPodsInjector {
	return &ProvisioningRequestPodsInjector{initialRetryTime: 1 * time.Minute, maxBackoffTime: 10 * time.Minute, backoffDuration: lru.New(1000), client: client, clock: clock}
}
