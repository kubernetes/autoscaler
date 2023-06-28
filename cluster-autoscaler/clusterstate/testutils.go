/*
Copyright 2023 The Kubernetes Authors.

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

package clusterstate

import (
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

// NewMockMaxNodeProvisionTimeProvider returns static maxNodeProvisionTimeProvider which returns constant MaxNodeProvisionTime for every NodeGroup.
func NewMockMaxNodeProvisionTimeProvider(maxNodeProvisionTime time.Duration) *staticMockMaxNodeProvisionTimeProvider {
	return &staticMockMaxNodeProvisionTimeProvider{maxNodeProvisionTime}
}

type staticMockMaxNodeProvisionTimeProvider struct {
	staticMaxNodeProvisionTime time.Duration
}

// GetMaxNodeProvisionTime returns constant MaxNodeProvisionTime value that should be used for every NodeGroup.
func (p *staticMockMaxNodeProvisionTimeProvider) GetMaxNodeProvisionTime(cloudprovider.NodeGroup) (time.Duration, error) {
	return p.staticMaxNodeProvisionTime, nil
}
