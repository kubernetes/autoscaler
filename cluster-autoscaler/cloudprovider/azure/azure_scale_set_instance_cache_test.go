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

package azure

import (
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/stretchr/testify/assert"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

func testGetInstanceCacheWithStates(t *testing.T, vms []compute.VirtualMachineScaleSetVM,
	states []cloudprovider.InstanceState) []cloudprovider.Instance {
	assert.Equal(t, len(vms), len(states))
	var instanceCacheTest []cloudprovider.Instance
	for i := 0; i < len(vms); i++ {
		instanceCacheTest = append(instanceCacheTest, cloudprovider.Instance{
			Id:     azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, i),
			Status: &cloudprovider.InstanceStatus{State: states[i]},
		})
	}
	return instanceCacheTest
}

// Suggestion: could populate all combinations, should reunify with TestInstanceStatusFromProvisioningStateAndPowerState
func TestInstanceStatusFromVM(t *testing.T) {
	t.Run("fast delete enablement = false", func(t *testing.T) {
		provider := newTestProvider(t)
		scaleSet := newTestScaleSet(provider.azureManager, "testScaleSet")

		t.Run("provisioning state = failed, power state = starting", func(t *testing.T) {
			vm := newVMObjectWithState(string(compute.GalleryProvisioningStateFailed), vmPowerStateStarting)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})

		t.Run("provisioning state = failed, power state = running", func(t *testing.T) {
			vm := newVMObjectWithState(string(compute.GalleryProvisioningStateFailed), vmPowerStateRunning)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})

		t.Run("provisioning state = failed, power state = stopping", func(t *testing.T) {
			vm := newVMObjectWithState(string(compute.GalleryProvisioningStateFailed), vmPowerStateStopping)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})

		t.Run("provisioning state = failed, power state = stopped", func(t *testing.T) {
			vm := newVMObjectWithState(string(compute.GalleryProvisioningStateFailed), vmPowerStateStopped)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})

		t.Run("provisioning state = failed, power state = deallocated", func(t *testing.T) {
			vm := newVMObjectWithState(string(compute.GalleryProvisioningStateFailed), vmPowerStateDeallocated)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})

		t.Run("provisioning state = failed, power state = unknown", func(t *testing.T) {
			vm := newVMObjectWithState(string(compute.GalleryProvisioningStateFailed), vmPowerStateUnknown)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})
	})

	t.Run("fast delete enablement = true", func(t *testing.T) {
		provider := newTestProvider(t)
		scaleSet := newTestScaleSetWithFastDelete(provider.azureManager, "testScaleSet")

		t.Run("provisioning state = failed, power state = starting", func(t *testing.T) {
			vm := newVMObjectWithState(string(compute.GalleryProvisioningStateFailed), vmPowerStateStarting)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})

		t.Run("provisioning state = failed, power state = running", func(t *testing.T) {
			vm := newVMObjectWithState(string(compute.GalleryProvisioningStateFailed), vmPowerStateRunning)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})

		t.Run("provisioning state = failed, power state = stopping", func(t *testing.T) {
			vm := newVMObjectWithState(string(compute.GalleryProvisioningStateFailed), vmPowerStateStopping)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceCreating, status.State)
			assert.NotNil(t, status.ErrorInfo)
		})

		t.Run("provisioning state = failed, power state = stopped", func(t *testing.T) {
			vm := newVMObjectWithState(string(compute.GalleryProvisioningStateFailed), vmPowerStateStopped)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceCreating, status.State)
			assert.NotNil(t, status.ErrorInfo)
		})

		t.Run("provisioning state = failed, power state = deallocated", func(t *testing.T) {
			vm := newVMObjectWithState(string(compute.GalleryProvisioningStateFailed), vmPowerStateDeallocated)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceCreating, status.State)
			assert.NotNil(t, status.ErrorInfo)
		})

		t.Run("provisioning state = failed, power state = unknown", func(t *testing.T) {
			vm := newVMObjectWithState(string(compute.GalleryProvisioningStateFailed), vmPowerStateUnknown)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceCreating, status.State)
			assert.NotNil(t, status.ErrorInfo)
		})
	})
}
