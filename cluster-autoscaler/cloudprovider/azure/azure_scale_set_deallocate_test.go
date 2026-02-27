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

package azure

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/azure/deallocate"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachineclient/mock_virtualmachineclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachinescalesetclient/mock_virtualmachinescalesetclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachinescalesetvmclient/mock_virtualmachinescalesetvmclient"
)

// fakeSuccessHandler implements runtime.PollingHandler and always reports done with no error.
type fakeSuccessHandler[T any] struct{}

func (f *fakeSuccessHandler[T]) Done() bool                                       { return true }
func (f *fakeSuccessHandler[T]) Poll(ctx context.Context) (*http.Response, error) { return nil, nil }
func (f *fakeSuccessHandler[T]) Result(ctx context.Context, out *T) error         { return nil }

// fakeErrorHandler implements runtime.PollingHandler and always reports done with an error.
type fakeErrorHandler[T any] struct {
	err error
}

func (f *fakeErrorHandler[T]) Done() bool                                       { return true }
func (f *fakeErrorHandler[T]) Poll(ctx context.Context) (*http.Response, error) { return nil, nil }
func (f *fakeErrorHandler[T]) Result(ctx context.Context, out *T) error         { return f.err }

func newFakeStartPoller(handler runtime.PollingHandler[armcompute.VirtualMachineScaleSetsClientStartResponse]) *runtime.Poller[armcompute.VirtualMachineScaleSetsClientStartResponse] {
	resp := &http.Response{Header: map[string][]string{"Fake-Poller-Status": {"Done"}}}
	p, _ := runtime.NewPoller(resp, runtime.Pipeline{},
		&runtime.NewPollerOptions[armcompute.VirtualMachineScaleSetsClientStartResponse]{Handler: handler})
	return p
}

func newFakeDeallocatePoller(handler runtime.PollingHandler[armcompute.VirtualMachineScaleSetsClientDeallocateResponse]) *runtime.Poller[armcompute.VirtualMachineScaleSetsClientDeallocateResponse] {
	resp := &http.Response{Header: map[string][]string{"Fake-Poller-Status": {"Done"}}}
	p, _ := runtime.NewPoller(resp, runtime.Pipeline{},
		&runtime.NewPollerOptions[armcompute.VirtualMachineScaleSetsClientDeallocateResponse]{Handler: handler})
	return p
}

func TestDeallocateNodes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	vmssName := "test-asg"
	var vmssCapacity int64 = 3
	cases := []struct {
		name              string
		orchestrationMode armcompute.OrchestrationMode
		enableForceDelete bool
		scaleDownPolicy   deallocate.ScaleDownPolicy
	}{
		{
			name:              "uniform, force delete enabled, deallocate mode",
			orchestrationMode: armcompute.OrchestrationModeUniform,
			enableForceDelete: true,
			scaleDownPolicy:   deallocate.Deallocate,
		},
		{
			name:              "uniform, force delete disabled, deallocate mode",
			orchestrationMode: armcompute.OrchestrationModeUniform,
			enableForceDelete: false,
			scaleDownPolicy:   deallocate.Deallocate,
		},
		/* Flex + Deallocate is not supported yet
		{
			name:              "flexible, force delete enabled, deallocate mode",
			orchestrationMode: armcompute.OrchestrationModeFlexible,
			enableForceDelete: true,
			scaleDownPolicy:   deallocate.Deallocate,
		},
		{
			name:              "flexible, force delete disabled, deallocate mode",
			orchestrationMode: armcompute.OrchestrationModeFlexible,
			enableForceDelete: false,
			scaleDownPolicy:   deallocate.Deallocate,
		},
		*/
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			orchMode := tc.orchestrationMode
			enableForceDelete := tc.enableForceDelete
			scaleDownPolicy := tc.scaleDownPolicy

			expectedVMSSVMs := newTestVMSSVMList(3)
			expectedVMs := newTestVMList(3)

			manager := newTestAzureManager(t)
			manager.config.EnableForceDelete = enableForceDelete
			expectedScaleSets := newTestVMSSList(vmssCapacity, vmssName, "eastus", orchMode)

			mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
			mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).Times(2)

			mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
			if scaleDownPolicy == deallocate.Delete {
				mockDeleteClient.EXPECT().BeginDeleteInstances(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
			} else {
				mockDeleteClient.EXPECT().BeginDeallocate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)
			}

			manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
			manager.azClient.vmssClientForDelete = mockDeleteClient

			mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
			mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)

			if orchMode == armcompute.OrchestrationModeUniform {
				mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, "test-asg").Return(expectedVMSSVMs, nil).AnyTimes()
				manager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
			} else {
				manager.config.EnableVmssFlexNodes = true
				mockVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedVMs, nil).AnyTimes()
				manager.azClient.virtualMachinesClient = mockVMClient
			}

			err := manager.forceRefresh()
			assert.NoError(t, err)

			resourceLimiter := cloudprovider.NewResourceLimiter(
				map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
				map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
			provider, err := BuildAzureCloudProvider(manager, resourceLimiter)

			assert.NoError(t, err)

			registered := manager.RegisterNodeGroup(
				newTestScaleSet(manager, testASG))
			manager.explicitlyConfigured[testASG] = true

			assert.True(t, registered)
			err = manager.forceRefresh()
			assert.NoError(t, err)

			scaleSet, ok := provider.NodeGroups()[0].(*ScaleSet)
			assert.True(t, ok)

			targetSize, err := scaleSet.TargetSize()
			assert.NoError(t, err)
			assert.Equal(t, 3, targetSize)

			scaleSet.scaleDownPolicy = scaleDownPolicy

			// Perform the delete operation
			nodesToDelete := []*apiv1.Node{
				newApiNode(orchMode, 0),
				newApiNode(orchMode, 2),
			}
			err = scaleSet.DeleteNodes(nodesToDelete)
			assert.NoError(t, err)

			if scaleDownPolicy == deallocate.Delete {
				// create scale set with vmss capacity 1
				expectedScaleSets = newTestVMSSList(1, vmssName, "eastus", orchMode)
			} else {
				// create scale set with vmss capacity 3
				expectedScaleSets = newTestVMSSList(3, vmssName, "eastus", orchMode)
			}

			mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()

			if orchMode == armcompute.OrchestrationModeUniform {
				if scaleDownPolicy == deallocate.Delete {
					expectedVMSSVMs[0].Properties.ProvisioningState = ptr.To(VMProvisioningStateDeleting)
					expectedVMSSVMs[2].Properties.ProvisioningState = ptr.To(VMProvisioningStateDeleting)
				} else {
					// DeleteNodes above waits for results in a goroutine, and deallocate implementation (waitForDeallocateInstancesResult)
					// currently accesses cache and adjusts provisioning state directly, so need to lock the instanceMutex to avoid data races
					// (Locks around scaleSet.lastInstanceRefresh below are added for the same reason)
					scaleSet.instanceMutex.Lock()
					expectedVMSSVMs[0].Properties.ProvisioningState = ptr.To(provisioningStateSucceeded)
					expectedVMSSVMs[2].Properties.ProvisioningState = ptr.To(provisioningStateSucceeded)
					expectedVMSSVMs[0].Properties.InstanceView = &armcompute.VirtualMachineScaleSetVMInstanceView{Statuses: []*armcompute.InstanceViewStatus{{Code: ptr.To(vmPowerStateDeallocating)}}}
					expectedVMSSVMs[2].Properties.InstanceView = &armcompute.VirtualMachineScaleSetVMInstanceView{Statuses: []*armcompute.InstanceViewStatus{{Code: ptr.To(vmPowerStateDeallocating)}}}
					scaleSet.instanceMutex.Unlock()

				}
				mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, "test-asg").Return(expectedVMSSVMs, nil).AnyTimes()
			} else {
				if scaleDownPolicy == deallocate.Delete {
					expectedVMs[0].Properties.ProvisioningState = ptr.To(VMProvisioningStateDeleting)
					expectedVMs[2].Properties.ProvisioningState = ptr.To(VMProvisioningStateDeleting)
				} else {
					// Flex + Deallocate is not supported yet; is not tested here, fail just in case
					assert.Fail(t, "flexible orchestration mode does not support deallocate")
				}
				mockVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedVMs, nil).AnyTimes()
			}

			err = manager.forceRefresh()
			assert.NoError(t, err)

			// Ensure the the cached size has been proactively decremented by 2
			targetSize, err = scaleSet.TargetSize()
			assert.NoError(t, err)
			assert.Equal(t, 1, targetSize)

			// Ensure that the status for the instances is Deleting or Deallocated
			// lastInstanceRefresh is set to time.Now() to avoid resetting instanceCache.
			scaleSet.instanceMutex.Lock()
			scaleSet.lastInstanceRefresh = time.Now()
			scaleSet.instanceMutex.Unlock()

			instance0, found, err := scaleSet.getInstanceByProviderID(nodesToDelete[0].Spec.ProviderID)
			assert.True(t, found, true)
			assert.NoError(t, err)
			if scaleDownPolicy == deallocate.Delete {
				assert.Equal(t, cloudprovider.InstanceDeleting, instance0.Status.State)
			} else {
				assert.Equal(t, cloudprovider.InstanceDeallocating, instance0.Status.State)
			}

			// lastInstanceRefresh is set to time.Now() to avoid resetting instanceCache.
			scaleSet.instanceMutex.Lock()
			scaleSet.lastInstanceRefresh = time.Now()
			scaleSet.instanceMutex.Unlock()

			instance2, found, err := scaleSet.getInstanceByProviderID(nodesToDelete[1].Spec.ProviderID)
			assert.True(t, found, true)
			assert.NoError(t, err)
			if scaleDownPolicy == deallocate.Delete {
				assert.Equal(t, cloudprovider.InstanceDeleting, instance2.Status.State)
			} else {
				assert.Equal(t, cloudprovider.InstanceDeallocating, instance2.Status.State)
			}
		})
	}
}

func TestWaitForStartInstance(t *testing.T) {
	provider := newTestProvider(t)

	expectedVMSSVMs := newTestVMSSVMList(3)
	var instances []cloudprovider.Instance
	for _, vm := range expectedVMSSVMs {
		instances = append(instances, cloudprovider.Instance{
			Id: azurePrefix + *vm.ID,
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceRunning,
			},
		})
	}

	asg := &ScaleSet{
		manager:         provider.azureManager,
		minSize:         1,
		maxSize:         5,
		InstanceCache:   InstanceCache{instanceCache: instances, instancesRefreshPeriod: defaultVmssInstancesRefreshPeriod},
		scaleDownPolicy: deallocate.Deallocate,
	}
	asg.Name = testASG

	t.Run("success: cache stays valid", func(t *testing.T) {
		asg.instanceMutex.Lock()
		asg.lastInstanceRefresh = time.Now()
		asg.instanceMutex.Unlock()

		poller := newFakeStartPoller(&fakeSuccessHandler[armcompute.VirtualMachineScaleSetsClientStartResponse]{})
		asg.waitForStartInstance(poller, "0")

		for _, vm := range asg.instanceCache {
			assert.Equal(t, cloudprovider.InstanceRunning, vm.Status.State)
		}
		// Cache should NOT be invalidated on success
		asg.instanceMutex.Lock()
		assert.False(t, asg.lastInstanceRefresh.IsZero(), "instanceCache should not be invalidated on success")
		asg.instanceMutex.Unlock()
	})

	t.Run("failure: cache is invalidated", func(t *testing.T) {
		asg.instanceMutex.Lock()
		asg.lastInstanceRefresh = time.Now()
		asg.instanceMutex.Unlock()

		timeBeforeCall := time.Now()
		poller := newFakeStartPoller(&fakeErrorHandler[armcompute.VirtualMachineScaleSetsClientStartResponse]{
			err: fmt.Errorf("some start error"),
		})
		asg.waitForStartInstance(poller, "0")

		// On failure, instanceCache should be invalidated (lastInstanceRefresh set to the past)
		asg.instanceMutex.Lock()
		assert.True(t, asg.lastInstanceRefresh.Before(timeBeforeCall), "instanceCache should be invalidated on failure")
		asg.instanceMutex.Unlock()
	})
}

func TestWaitForDeallocateInstances(t *testing.T) {
	provider := newTestProvider(t)

	expectedVMSSVMs := newTestVMSSVMList(3)
	var instances []cloudprovider.Instance
	for _, vm := range expectedVMSSVMs {
		instances = append(instances, cloudprovider.Instance{
			Id: azurePrefix + *vm.ID,
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceDeallocating,
			},
		})
	}

	instanceRefs := []*azureRef{
		{Name: azurePrefix + *expectedVMSSVMs[0].ID},
		{Name: azurePrefix + *expectedVMSSVMs[1].ID},
		{Name: azurePrefix + *expectedVMSSVMs[2].ID},
	}

	asg := &ScaleSet{
		manager:         provider.azureManager,
		minSize:         1,
		maxSize:         5,
		InstanceCache:   InstanceCache{instanceCache: instances, instancesRefreshPeriod: defaultVmssInstancesRefreshPeriod},
		scaleDownPolicy: deallocate.Deallocate,
	}
	asg.Name = testASG

	t.Run("success: instances set to deallocated", func(t *testing.T) {
		// Reset states to Deallocating
		for i := range asg.instanceCache {
			asg.instanceCache[i].Status.State = cloudprovider.InstanceDeallocating
		}
		asg.instanceMutex.Lock()
		asg.lastInstanceRefresh = time.Now()
		asg.instanceMutex.Unlock()

		poller := newFakeDeallocatePoller(&fakeSuccessHandler[armcompute.VirtualMachineScaleSetsClientDeallocateResponse]{})
		asg.waitForDeallocateInstances(poller, instanceRefs, []*string{})

		for _, vm := range asg.instanceCache {
			assert.Equal(t, cloudprovider.InstanceDeallocated, vm.Status.State)
		}
		// Cache should NOT be invalidated on success
		asg.instanceMutex.Lock()
		assert.False(t, asg.lastInstanceRefresh.IsZero(), "instanceCache should not be invalidated on success")
		asg.instanceMutex.Unlock()
	})

	t.Run("failure: cache is invalidated", func(t *testing.T) {
		// Reset states to Deallocating
		for i := range asg.instanceCache {
			asg.instanceCache[i].Status.State = cloudprovider.InstanceDeallocating
		}
		asg.instanceMutex.Lock()
		asg.lastInstanceRefresh = time.Now()
		asg.instanceMutex.Unlock()

		timeBeforeCall := time.Now()
		poller := newFakeDeallocatePoller(&fakeErrorHandler[armcompute.VirtualMachineScaleSetsClientDeallocateResponse]{
			err: fmt.Errorf("some deallocate error"),
		})
		asg.waitForDeallocateInstances(poller, instanceRefs, []*string{})

		// On failure, instanceCache should be invalidated (lastInstanceRefresh set to the past)
		asg.instanceMutex.Lock()
		assert.True(t, asg.lastInstanceRefresh.Before(timeBeforeCall), "instanceCache should be invalidated on failure")
		asg.instanceMutex.Unlock()

		// States should NOT have been changed to Deallocated
		for _, vm := range asg.instanceCache {
			assert.Equal(t, cloudprovider.InstanceDeallocating, vm.Status.State)
		}
	})
}
