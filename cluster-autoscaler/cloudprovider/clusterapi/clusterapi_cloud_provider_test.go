/*
Copyright 2016 The Kubernetes Authors.

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

package clusterapi

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"testing"
)

func newTestMachineManager(t *testing.T) *MachineManagerMock {
	manager := new(MachineManagerMock)

	return manager
}
func newTestProvider(t *testing.T) *ClusterapiCloudProvider {
	manager := newTestMachineManager(t)
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})

	return &ClusterapiCloudProvider{
		resourceLimiter: resourceLimiter,
		machineManager:  manager,
	}
}

func TestBuildClusterapiCloudProvider(t *testing.T) {
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	machineManager := newTestMachineManager(t)
	_, err := BuildClusterapiCloudProvider(machineManager, resourceLimiter)
	assert.NoError(t, err)
}

func TestName(t *testing.T) {
	provider := newTestProvider(t)
	assert.Equal(t, provider.Name(), "clusterapi")
}
