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
	"net/http"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"k8s.io/utils/ptr"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/azure/deallocate"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachinescalesetclient/mock_virtualmachinescalesetclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachinescalesetvmclient/mock_virtualmachinescalesetvmclient"
)

func testGetInstanceCacheWithStates(t *testing.T, vms []*armcompute.VirtualMachineScaleSetVM,
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
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateStarting)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})

		t.Run("provisioning state = failed, power state = running", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateRunning)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})

		t.Run("provisioning state = failed, power state = stopping", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateStopping)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})

		t.Run("provisioning state = failed, power state = stopped", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateStopped)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})

		t.Run("provisioning state = failed, power state = deallocated", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateDeallocated)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})

		t.Run("provisioning state = failed, power state = unknown", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateUnknown)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})
	})

	t.Run("fast delete enablement = true", func(t *testing.T) {
		provider := newTestProvider(t)
		scaleSet := newTestScaleSetWithFastDelete(provider.azureManager, "testScaleSet")

		t.Run("provisioning state = failed, power state = starting", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateStarting)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})

		t.Run("provisioning state = failed, power state = running", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateRunning)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})

		t.Run("provisioning state = failed, power state = stopping", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateStopping)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceCreating, status.State)
			assert.NotNil(t, status.ErrorInfo)
		})

		t.Run("provisioning state = failed, power state = stopped", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateStopped)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceCreating, status.State)
			assert.NotNil(t, status.ErrorInfo)
		})

		t.Run("provisioning state = failed, power state = deallocated", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateDeallocated)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceCreating, status.State)
			assert.NotNil(t, status.ErrorInfo)
		})

		t.Run("provisioning state = failed, power state = unknown", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateUnknown)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceCreating, status.State)
			assert.NotNil(t, status.ErrorInfo)
		})
	})
}

// Suggestion: simplify all unit tests below here; looks like it is trying to test multiple things (e.g., state handling, cache, deallocate mode) at the same time
var (
	ctrl                                 *gomock.Controller
	currentTime, expiredTime             time.Time
	provider                             *AzureCloudProvider
	scaleSet                             *ScaleSet
	mockVMSSVMClient                     *mock_virtualmachinescalesetvmclient.MockInterface
	expectedVMSSVMs                      []*armcompute.VirtualMachineScaleSetVM
	expectedStates                       []cloudprovider.InstanceState
	instanceCache, expectedInstanceCache []cloudprovider.Instance
)

func newTestScaleSetDeallocateMode(manager *AzureManager, name string) *ScaleSet {
	return &ScaleSet{
		azureRef: azureRef{
			Name: name,
		},
		manager:           manager,
		minSize:           1,
		maxSize:           5,
		enableForceDelete: manager.config.EnableForceDelete,
		scaleDownPolicy:   deallocate.Deallocate,
	}
}

func newTestScaleSetDeallocateModeWithFastDelete(manager *AzureManager, name string) *ScaleSet {
	return &ScaleSet{
		azureRef: azureRef{
			Name: name,
		},
		manager:           manager,
		minSize:           1,
		maxSize:           5,
		enableForceDelete: manager.config.EnableForceDelete,
		scaleDownPolicy:   deallocate.Deallocate,
	}
}

func TestInvalidateInstanceCache(t *testing.T) {
	TestBeforeEachNoInstanceCacheResetNeededHelper(t)

	scaleSet.invalidateInstanceCache()
	assert.Lessf(t, scaleSet.lastInstanceRefresh, currentTime, "lastInstanceRefresh should be less than current"+
		"time as instanceCache is invalidated")
}

// TestValidateInstanceCache tests only with orchestrationMode = Uniform
func TestValidateInstanceCache(t *testing.T) {
	TestBeforeEachNoInstanceCacheResetNeededHelper(t)

	// t1 - expect no update to instanceCache because timer has not yet expired
	err := scaleSet.validateInstanceCache()
	assert.NoErrorf(t, err, "err is not expected when validating instanceCache with a fresh cache")

	TestBeforeEachInstanceCacheResetNeededHelper(t)

	// t2 - expect update happens because instanceCache is invalidated without deallocate mode and enableCSE
	err = scaleSet.validateInstanceCache()
	assert.NoError(t, err)
	assert.Equalf(t, len(expectedVMSSVMs), len(scaleSet.instanceCache), "instanceCache must be updated")
	assert.Greaterf(t, scaleSet.lastInstanceRefresh, expiredTime, "after refresh, instanceCache should have updated"+
		"lastInstanceRefresh")

	// t3 - throttling on get call - cache is not update but refresh time is updated
	instanceCacheLen := len(scaleSet.instanceCache)
	expiredTime = scaleSet.lastInstanceRefresh.Add(-1 * scaleSet.instancesRefreshPeriod)
	scaleSet.lastInstanceRefresh = expiredTime
	throttledError := &azcore.ResponseError{
		StatusCode:  http.StatusTooManyRequests,
		RawResponse: &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{}},
	}
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), provider.azureManager.config.ResourceGroup, testASG).Return(
		[]*armcompute.VirtualMachineScaleSetVM{}, throttledError).Times(1)
	err = scaleSet.validateInstanceCache()
	assert.NoError(t, err)
	assert.Equalf(t, instanceCacheLen, len(scaleSet.instanceCache), "instanceCache must not be updated")
	assert.Greaterf(t, scaleSet.lastInstanceRefresh, expiredTime, "lastInstanceRefresh must be updated even if the "+
		"instanceCache is not updated")
}

func TestGetInstanceByProviderID(t *testing.T) {
	TestBeforeEachNoInstanceCacheResetNeededHelper(t)

	// t1 - cache is not stale - instance exists in the instanceCache
	reqProviderID := azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 0)
	actualInstance, found, err := scaleSet.getInstanceByProviderID(reqProviderID)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, instanceCache[0], actualInstance)

	TestBeforeEachInstanceCacheResetNeededHelper(t)

	// t2 - cache is stale - instance exists in the instanceCache
	reqProviderID = azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 1)
	actualInstance, found, err = scaleSet.getInstanceByProviderID(reqProviderID)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, expectedInstanceCache[1].Id, actualInstance.Id)
	assert.Equal(t, expectedInstanceCache[1].Status.State, actualInstance.Status.State)

	// t3 - cache is not stale and instance doesn't exist in the instanceCache
	scaleSet.lastInstanceRefresh = time.Now()
	reqProviderID = azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 2)
	actualInstance, found, err = scaleSet.getInstanceByProviderID(reqProviderID)
	assert.NoError(t, err)
	assert.False(t, found)
	assert.Empty(t, actualInstance)
}

func TestGetInstancesByState(t *testing.T) {
	TestBeforeEachNoInstanceCacheResetNeededHelper(t)
	scaleSet.scaleDownPolicy = deallocate.Deallocate // Suggestion: (also apply to similar entires) use TestScaleSetDellocateMode; right now that is abstracted behind TestBeforeEachNoInstanceCacheResetNeededHelper

	// t1 cache is not stale - instance with given state exists in the instanceCache
	actualInstances, err := scaleSet.getInstancesByState(cloudprovider.InstanceDeallocated)
	assert.NoError(t, err)
	assert.Equal(t, instanceCache, actualInstances)

	TestBeforeEachInstanceCacheResetNeededHelper(t)
	scaleSet.scaleDownPolicy = deallocate.Deallocate

	// t2 - cache is stale - instance with given state exists in the instanceCache
	actualInstances, err = scaleSet.getInstancesByState(cloudprovider.InstanceFailed)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(actualInstances)) // there should be only one instance with failed State
	assert.Equal(t, expectedInstanceCache[0].Id, actualInstances[0].Id)
	assert.Equal(t, expectedInstanceCache[0].Status.State, actualInstances[0].Status.State)

	// t3 - cache is not stale and instance with given state doesn't exist in the instanceCache
	scaleSet.lastInstanceRefresh = time.Now()
	actualInstances, err = scaleSet.getInstancesByState(cloudprovider.InstanceDeallocating)
	assert.NoError(t, err)
	assert.Empty(t, actualInstances)
}

func TestGetInstanceCacheSize(t *testing.T) {
	TestBeforeEachNoInstanceCacheResetNeededHelper(t)

	// t1 - cache is not stale - cache will not update and return size of 1
	actualSize, err := scaleSet.getInstanceCacheSize()
	assert.NoError(t, err)
	assert.Equal(t, 1, int(actualSize))

	TestBeforeEachInstanceCacheResetNeededHelper(t)

	// t2 - cache is stale - update will return size of 2
	actualSize, err = scaleSet.getInstanceCacheSize()
	assert.NoError(t, err)
	assert.Equal(t, len(expectedInstanceCache), int(actualSize))
}

func TestSetInstanceStatusByProviderID(t *testing.T) {
	TestBeforeEachNoInstanceCacheResetNeededHelper(t)
	scaleSet.scaleDownPolicy = deallocate.Deallocate

	// t1 - cache is not stale - instance exists in the instanceCache
	providerID := azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 0)
	status := cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning}
	scaleSet.setInstanceStatusByProviderID(providerID, status)
	actualInstance, found, err := scaleSet.getInstanceByProviderID(providerID)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, cloudprovider.InstanceRunning, actualInstance.Status.State)

	TestBeforeEachInstanceCacheResetNeededHelper(t)
	scaleSet.scaleDownPolicy = deallocate.Deallocate

	// t2 - cache is stale - expectInstanceCache update, set for providerID=2 will not be added to the instanceCache because
	// it doesn't exist in the cache. GetScaleSetVms() will have not introduced instance with providerID=2
	providerID = azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 2)
	status = cloudprovider.InstanceStatus{State: cloudprovider.InstanceFailed}
	scaleSet.setInstanceStatusByProviderID(providerID, status) // it will not set for providerID=2 as it is not already present in the cache
	actualInstances, err := scaleSet.getInstancesByState(cloudprovider.InstanceFailed)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(actualInstances))
}

// beforeEachNoInstanceCacheResetNeededHelper has 1 instance in the instanceCache with state = deallocated.
func TestBeforeEachNoInstanceCacheResetNeededHelper(t *testing.T) {
	ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	provider = newTestProvider(t)
	scaleSet = newTestScaleSet(provider.azureManager, testASG)
	scaleSet.instancesRefreshPeriod = defaultVmssInstancesRefreshPeriod

	expectedScaleSets := newTestVMSSList(3, testASG, "eastus", armcompute.OrchestrationModeUniform)
	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
	mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()

	registered := provider.azureManager.RegisterNodeGroup(
		scaleSet)
	provider.azureManager.explicitlyConfigured[testASG] = true
	assert.True(t, registered)
	err := provider.azureManager.forceRefresh()
	assert.NoError(t, err)

	currentTime = time.Now()
	scaleSet.lastInstanceRefresh = currentTime
	expectedVMSSVMs = newTestVMSSVMList(1)
	expectedStates = []cloudprovider.InstanceState{cloudprovider.InstanceDeallocated}
	instanceCache = testGetInstanceCacheWithStates(t, expectedVMSSVMs, expectedStates)
	scaleSet.instanceCache = instanceCache
}

// beforeEachInstanceCacheResetNeededHelper has 1 instance in the instanceCache with state = failed, deallocated
func TestBeforeEachInstanceCacheResetNeededHelper(t *testing.T) {
	expiredTime = scaleSet.lastInstanceRefresh.Add(-1 * scaleSet.instancesRefreshPeriod)
	scaleSet.lastInstanceRefresh = expiredTime
	mockVMSSVMClient = mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
	provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
	expectedVMSSVMs = newTestVMSSVMList(2)
	expectedVMSSVMs[0].Properties.ProvisioningState = ptr.To(string(armcompute.GalleryProvisioningStateFailed))
	expectedVMSSVMs[1].Properties.ProvisioningState = ptr.To(string(armcompute.GalleryProvisioningStateSucceeded))
	expectedVMSSVMs[1].Properties.InstanceView = &armcompute.VirtualMachineScaleSetVMInstanceView{
		Statuses: []*armcompute.InstanceViewStatus{
			{Code: ptr.To(vmPowerStateDeallocated)},
		},
	}
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), provider.azureManager.config.ResourceGroup, testASG).Return(
		expectedVMSSVMs, nil)
	expectedStates = []cloudprovider.InstanceState{cloudprovider.InstanceFailed, cloudprovider.InstanceDeallocated}
	expectedInstanceCache = testGetInstanceCacheWithStates(t, expectedVMSSVMs, expectedStates)
}

func TestInstanceStatusFromVMEnableFastDeleteOnFailedProvisioning(t *testing.T) {
	TestBeforeEachNoInstanceCacheResetNeededHelper(t)

	scaleSet.scaleDownPolicy = deallocate.Deallocate
	// Disabled EnableFastDelete, deallocate mode, running power state
	vm := &armcompute.VirtualMachineScaleSetVM{
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			ProvisioningState: ptr.To(string(armcompute.GalleryProvisioningStateFailed)),
			InstanceView: &armcompute.VirtualMachineScaleSetVMInstanceView{
				Statuses: []*armcompute.InstanceViewStatus{
					{Code: ptr.To(vmPowerStateRunning)},
				},
			},
		},
	}
	status := scaleSet.instanceStatusFromVM(vm)
	assert.NotNil(t, status)
	assert.Equal(t, cloudprovider.InstanceFailed, status.State)

	// Enabled EnableFastDelete, deallocate mode, running power state
	scaleSet.enableFastDeleteOnFailedProvisioning = true
	vm = &armcompute.VirtualMachineScaleSetVM{
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			ProvisioningState: ptr.To(string(armcompute.GalleryProvisioningStateFailed)),
			InstanceView: &armcompute.VirtualMachineScaleSetVMInstanceView{
				Statuses: []*armcompute.InstanceViewStatus{
					{Code: ptr.To(vmPowerStateRunning)},
				},
			},
		},
	}
	status = scaleSet.instanceStatusFromVM(vm)
	assert.NotNil(t, status)
	assert.Equal(t, cloudprovider.InstanceFailed, status.State)
	scaleSet.enableFastDeleteOnFailedProvisioning = false

	// Enabled EnableFastDelete, deallocate mode, not running power state
	scaleSet.enableFastDeleteOnFailedProvisioning = true
	vm = &armcompute.VirtualMachineScaleSetVM{
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			ProvisioningState: ptr.To(string(armcompute.GalleryProvisioningStateFailed)),
			InstanceView: &armcompute.VirtualMachineScaleSetVMInstanceView{
				Statuses: []*armcompute.InstanceViewStatus{
					{Code: ptr.To(vmPowerStateStopped)},
				},
			},
		},
	}
	status = scaleSet.instanceStatusFromVM(vm)
	assert.NotNil(t, status)
	assert.Equal(t, cloudprovider.InstanceFailed, status.State)
	scaleSet.enableFastDeleteOnFailedProvisioning = false

	scaleSet.scaleDownPolicy = deallocate.Delete
	// Disabled EnableFastDelete, delete mode, running power state
	vm = &armcompute.VirtualMachineScaleSetVM{
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			ProvisioningState: ptr.To(string(armcompute.GalleryProvisioningStateFailed)),
			InstanceView: &armcompute.VirtualMachineScaleSetVMInstanceView{
				Statuses: []*armcompute.InstanceViewStatus{
					{Code: ptr.To(vmPowerStateRunning)},
				},
			},
		},
	}
	status = scaleSet.instanceStatusFromVM(vm)
	assert.NotNil(t, status)
	assert.Equal(t, cloudprovider.InstanceRunning, status.State)

	// Enabled EnableFastDelete, delete mode, running power state
	scaleSet.enableFastDeleteOnFailedProvisioning = true
	vm = &armcompute.VirtualMachineScaleSetVM{
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			ProvisioningState: ptr.To(string(armcompute.GalleryProvisioningStateFailed)),
			InstanceView: &armcompute.VirtualMachineScaleSetVMInstanceView{
				Statuses: []*armcompute.InstanceViewStatus{
					{Code: ptr.To(vmPowerStateRunning)},
				},
			},
		},
	}
	status = scaleSet.instanceStatusFromVM(vm)
	assert.NotNil(t, status)
	assert.Equal(t, cloudprovider.InstanceRunning, status.State)
	scaleSet.enableFastDeleteOnFailedProvisioning = false

	// Enabled EnableFastDelete, delete mode, not running power state
	scaleSet.enableFastDeleteOnFailedProvisioning = true
	vm = &armcompute.VirtualMachineScaleSetVM{
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			ProvisioningState: ptr.To(string(armcompute.GalleryProvisioningStateFailed)),
			InstanceView: &armcompute.VirtualMachineScaleSetVMInstanceView{
				Statuses: []*armcompute.InstanceViewStatus{
					{Code: ptr.To(vmPowerStateStopped)},
				},
			},
		},
	}
	status = scaleSet.instanceStatusFromVM(vm)
	assert.NotNil(t, status)
	assert.Equal(t, cloudprovider.InstanceCreating, status.State)
	assert.NotEmpty(t, status.ErrorInfo)
	scaleSet.enableFastDeleteOnFailedProvisioning = false
}

func TestInstanceStatusFromVMEnableFastDeleteOnCSEFailure(t *testing.T) {
	TestBeforeEachNoInstanceCacheResetNeededHelper(t)

	// Enabled EnableFastDelete, delete mode, not running power state
	scaleSet.enableFastDeleteOnFailedProvisioning = true
	scaleSet.enableDetailedCSEMessage = true
	vm := &armcompute.VirtualMachineScaleSetVM{
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			ProvisioningState: ptr.To(string(armcompute.GalleryProvisioningStateFailed)),
			InstanceView: &armcompute.VirtualMachineScaleSetVMInstanceView{
				Statuses: []*armcompute.InstanceViewStatus{
					{Code: ptr.To(vmPowerStateStopped)},
				},
				Extensions: []*armcompute.VirtualMachineExtensionInstanceView{
					{
						Name: ptr.To(vmssCSEExtensionName),
						Statuses: []*armcompute.InstanceViewStatus{
							{
								Level:   ptr.To(armcompute.StatusLevelTypesError),
								Code:    ptr.To(vmssExtensionProvisioningFailed),
								Message: ptr.To("Custom Script Extension failed to provision"),
							},
						},
					},
				},
			},
		},
	}
	status := scaleSet.instanceStatusFromVM(vm)
	assert.NotNil(t, status)
	assert.Equal(t, cloudprovider.InstanceCreating, status.State)
	assert.NotEmpty(t, status.ErrorInfo)
	assert.Equal(t, cloudprovider.OtherErrorClass, status.ErrorInfo.ErrorClass)
	assert.Equal(t, vmssExtensionProvisioningFailed, status.ErrorInfo.ErrorCode)
	scaleSet.enableFastDeleteOnFailedProvisioning = false
}

func TestInstanceStatusFromVMDeallocateMode(t *testing.T) {
	t.Run("fast delete enablement = false", func(t *testing.T) {
		provider := newTestProvider(t)
		scaleSet := newTestScaleSetDeallocateMode(provider.azureManager, "testScaleSet")

		t.Run("provisioning state = failed, power state = starting", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateStarting)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceFailed, status.State)
		})

		t.Run("provisioning state = failed, power state = running", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateRunning)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceFailed, status.State)
		})

		t.Run("provisioning state = failed, power state = stopping", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateStopping)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceFailed, status.State)
		})

		t.Run("provisioning state = failed, power state = stopped", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateStopped)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceFailed, status.State)
		})

		t.Run("provisioning state = failed, power state = deallocated", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateDeallocated)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceFailed, status.State)
		})

		t.Run("provisioning state = failed, power state = unknown", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateUnknown)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceFailed, status.State)
		})
	})

	t.Run("fast delete enablement = true", func(t *testing.T) {
		provider := newTestProvider(t)
		scaleSet := newTestScaleSetDeallocateModeWithFastDelete(provider.azureManager, "testScaleSet")

		t.Run("provisioning state = failed, power state = starting", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateStarting)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceFailed, status.State)
		})

		t.Run("provisioning state = failed, power state = running", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateRunning)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceFailed, status.State)
		})

		t.Run("provisioning state = failed, power state = stopping", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateStopping)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceFailed, status.State)
		})

		t.Run("provisioning state = failed, power state = stopped", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateStopped)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceFailed, status.State)
		})

		t.Run("provisioning state = failed, power state = deallocated", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateDeallocated)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceFailed, status.State)
		})

		t.Run("provisioning state = failed, power state = unknown", func(t *testing.T) {
			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateUnknown)

			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceFailed, status.State)
		})
	})
}
