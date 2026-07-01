/*
Copyright 2017 The Kubernetes Authors.

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
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachineclient/mock_virtualmachineclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachinescalesetclient/mock_virtualmachinescalesetclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/virtualmachinescalesetvmclient/mock_virtualmachinescalesetvmclient"
)

const (
	testLocation = "eastus"
)

func newTestScaleSet(manager *AzureManager, name string) *ScaleSet {
	return &ScaleSet{
		azureRef: azureRef{
			Name: name,
		},
		manager:           manager,
		minSize:           1,
		maxSize:           5,
		enableForceDelete: manager.config.EnableForceDelete,
	}
}

func newTestScaleSetMinSizeZero(manager *AzureManager, name string) *ScaleSet {
	return &ScaleSet{
		azureRef: azureRef{
			Name: name,
		},
		manager:           manager,
		minSize:           0,
		maxSize:           5,
		enableForceDelete: manager.config.EnableForceDelete,
	}
}

func newTestScaleSetWithFastDelete(manager *AzureManager, name string) *ScaleSet {
	return &ScaleSet{
		azureRef: azureRef{
			Name: name,
		},
		manager:                              manager,
		minSize:                              1,
		maxSize:                              5,
		enableForceDelete:                    manager.config.EnableForceDelete,
		enableFastDeleteOnFailedProvisioning: true,
	}
}

func newTestVMSSList(cap int64, name, loc string, orchmode armcompute.OrchestrationMode) []*armcompute.VirtualMachineScaleSet {
	return []*armcompute.VirtualMachineScaleSet{
		{
			Name: ptr.To(name),
			SKU: &armcompute.SKU{
				Capacity: ptr.To(cap),
				Name:     ptr.To("Standard_D4_v2"),
			},
			Properties: &armcompute.VirtualMachineScaleSetProperties{
				OrchestrationMode: &orchmode,
			},
			Location: ptr.To(loc),
			ID:       ptr.To(name),
		},
	}
}

func newTestVMSSListForEdgeZones(capacity int64, name string) *armcompute.VirtualMachineScaleSet {
	return &armcompute.VirtualMachineScaleSet{
		Name: ptr.To(name),
		SKU: &armcompute.SKU{
			Capacity: ptr.To(capacity),
			Name:     ptr.To("Standard_D4_v2"),
		},
		Properties: &armcompute.VirtualMachineScaleSetProperties{},
		Location:   ptr.To(testLocation),
		ExtendedLocation: &armcompute.ExtendedLocation{
			Name: ptr.To("losangeles"),
			Type: ptr.To(armcompute.ExtendedLocationTypesEdgeZone),
		},
	}
}

func newTestVMSSVMList(count int) []*armcompute.VirtualMachineScaleSetVM {
	var vmssVMList []*armcompute.VirtualMachineScaleSetVM
	for i := 0; i < count; i++ {
		vmssVM := &armcompute.VirtualMachineScaleSetVM{
			ID:         ptr.To(fmt.Sprintf(fakeVirtualMachineScaleSetVMID, i)),
			InstanceID: ptr.To(fmt.Sprintf("%d", i)),
			Properties: &armcompute.VirtualMachineScaleSetVMProperties{
				VMID: ptr.To(fmt.Sprintf("123E4567-E89B-12D3-A456-426655440000-%d", i)),
			},
		}
		vmssVMList = append(vmssVMList, vmssVM)
	}
	return vmssVMList
}

func newTestVMList(count int) []*armcompute.VirtualMachine {
	var vmssVMList []*armcompute.VirtualMachine
	for i := 0; i < count; i++ {
		vmssVM := &armcompute.VirtualMachine{
			ID: ptr.To(fmt.Sprintf(fakeVirtualMachineVMID, i)),
			Properties: &armcompute.VirtualMachineProperties{
				VMID: ptr.To(fmt.Sprintf("123E4567-E89B-12D3-A456-426655440000-%d", i)),
			},
		}
		vmssVMList = append(vmssVMList, vmssVM)
	}
	return vmssVMList
}

func newApiNode(orchmode armcompute.OrchestrationMode, vmID int64) *apiv1.Node {
	providerId := fakeVirtualMachineScaleSetVMID

	if orchmode == armcompute.OrchestrationModeFlexible {
		providerId = fakeVirtualMachineVMID
	}

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: azurePrefix + fmt.Sprintf(providerId, vmID),
		},
	}
	return node
}
func TestScaleSetMaxSize(t *testing.T) {
	provider := newTestProvider(t)
	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)
	assert.Equal(t, provider.NodeGroups()[0].MaxSize(), 5)
}

func TestScaleSetMinSize(t *testing.T) {
	provider := newTestProvider(t)
	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)
	assert.Equal(t, provider.NodeGroups()[0].MinSize(), 1)
}

func TestScaleSetMinSizeZero(t *testing.T) {
	provider := newTestProvider(t)
	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSetMinSizeZero(provider.azureManager, testASG))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)
	assert.Equal(t, provider.NodeGroups()[0].MinSize(), 0)
}

func TestScaleSetTargetSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orchestrationModes := [2]armcompute.OrchestrationMode{armcompute.OrchestrationModeUniform, armcompute.OrchestrationModeFlexible}
	expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus", armcompute.OrchestrationModeUniform)
	spotScaleSet := newTestVMSSList(5, "spot-vmss", "eastus", armcompute.OrchestrationModeUniform)[0]
	priority := armcompute.VirtualMachinePriorityTypesSpot
	spotScaleSet.Properties.VirtualMachineProfile = &armcompute.VirtualMachineScaleSetVMProfile{Priority: &priority}
	expectedScaleSets = append(expectedScaleSets, spotScaleSet)

	expectedVMSSVMs := newTestVMSSVMList(3)
	expectedVMs := newTestVMList(3)

	for _, orchMode := range orchestrationModes {
		provider := newTestProvider(t)
		mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
		mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
		provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
		mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)
		mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return([]*armcompute.VirtualMachine{}, nil).AnyTimes()
		provider.azureManager.azClient.virtualMachinesClient = mockVMClient

		// return a different capacity from GET API
		spotScaleSet.SKU.Capacity = ptr.To[int64](1)
		mockVMSSClient.EXPECT().Get(gomock.Any(), provider.azureManager.config.ResourceGroup, "spot-vmss", nil).Return(spotScaleSet, nil).Times(1)
		provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
		mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)

		mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), provider.azureManager.config.ResourceGroup, "test-asg").Return(expectedVMSSVMs, nil).AnyTimes()
		provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
		err := provider.azureManager.forceRefresh()
		assert.NoError(t, err)

		if orchMode == armcompute.OrchestrationModeUniform {
			mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
			mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), provider.azureManager.config.ResourceGroup, "test-asg").Return(expectedVMSSVMs, nil).AnyTimes()
			provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
		} else {
			provider.azureManager.config.EnableVmssFlexNodes = true
			mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedVMs, nil).AnyTimes()
			mockVMClient.EXPECT().ListVmssFlexVMsWithOutInstanceView(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedVMs, nil).AnyTimes()
		}

		err = provider.azureManager.forceRefresh()
		assert.NoError(t, err)

		registered := provider.azureManager.RegisterNodeGroup(
			newTestScaleSet(provider.azureManager, testASG))
		assert.True(t, registered)
		assert.Equal(t, len(provider.NodeGroups()), 1)

		targetSize, err := provider.NodeGroups()[0].TargetSize()
		assert.NoError(t, err)
		assert.Equal(t, 3, targetSize)

		targetSize, err = provider.NodeGroups()[0].TargetSize()
		assert.NoError(t, err)
		assert.Equal(t, 3, targetSize)

		// With a spot nodegroup
		spotNodeGroup := newTestScaleSet(provider.azureManager, "spot-vmss")
		registered = provider.azureManager.RegisterNodeGroup(spotNodeGroup)
		assert.True(t, registered)
		assert.Equal(t, len(provider.NodeGroups()), 2)

		targetSize, err = provider.NodeGroups()[1].TargetSize()
		assert.NoError(t, err)
		assert.Equal(t, 1, targetSize)
	}
}

func TestScaleSetTargetSizeReturnsErrorForCachedNegativeSize(t *testing.T) {
	provider := newTestProvider(t)
	err := provider.azureManager.forceRefresh()
	assert.NoError(t, err)

	scaleSet := newTestScaleSet(provider.azureManager, testASG)
	scaleSet.curSize = -1
	scaleSet.lastSizeRefresh = time.Now()
	scaleSet.sizeRefreshPeriod = time.Hour

	size, err := scaleSet.TargetSize()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cached size is -1 without provider error")
	assert.Equal(t, -1, size)
}

func TestScaleSetIncreaseSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orchestrationModes := [2]armcompute.OrchestrationMode{armcompute.OrchestrationModeUniform, armcompute.OrchestrationModeFlexible}

	for _, orchMode := range orchestrationModes {

		expectedScaleSets := newTestVMSSList(3, testASG, "eastus", orchMode)
		expectedVMSSVMs := newTestVMSSVMList(3)
		expectedVMs := newTestVMList(3)

		// Include Edge Zone scenario here, testing scale from 3 to 5 and scale from zero cases.
		expectedEdgeZoneScaleSets := newTestVMSSListForEdgeZones(3, "edgezone-vmss")
		expectedEdgeZoneMinZeroScaleSets := newTestVMSSListForEdgeZones(0, "edgezone-minzero-vmss")
		expectedScaleSets = append(expectedScaleSets, expectedEdgeZoneScaleSets, expectedEdgeZoneMinZeroScaleSets)

		provider := newTestProvider(t)
		// expectedScaleSets := newTestVMSSList(3, testASG, testLocation, orchMode)

		// // Include Edge Zone scenario here, testing scale from 3 to 5 and scale from zero cases.
		// expectedEdgeZoneScaleSets := newTestVMSSListForEdgeZones(3, "edgezone-vmss")
		// expectedEdgeZoneMinZeroScaleSets := newTestVMSSListForEdgeZones(0, "edgezone-minzero-vmss")
		// expectedScaleSets = append(expectedScaleSets, *expectedEdgeZoneScaleSets, *expectedEdgeZoneMinZeroScaleSets)

		mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
		mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
		// This should be Anytimes() because the parent function of this call - updateVMSSCapacity() is a goroutine
		// and this test doesn't wait on goroutine, hence, it is difficult to write exact expected number (which is 3 here)
		// before we return from this this.
		// This is a future TODO: sync.WaitGroup should be used in actual code and make code easily testable
		provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

		// Mock the vmssClientForDelete for async CreateOrUpdate calls
		mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
		mockDeleteClient.EXPECT().BeginCreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		mockDeleteClient.EXPECT().BeginDeleteInstances(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		provider.azureManager.azClient.vmssClientForDelete = mockDeleteClient

		mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)
		mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return([]*armcompute.VirtualMachine{}, nil).AnyTimes()
		provider.azureManager.azClient.virtualMachinesClient = mockVMClient

		if orchMode == armcompute.OrchestrationModeUniform {
			mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
			mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), provider.azureManager.config.ResourceGroup, "test-asg").Return(expectedVMSSVMs, nil).AnyTimes()
			provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
		} else {
			provider.azureManager.config.EnableVmssFlexNodes = true
			mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedVMs, nil).AnyTimes()
			mockVMClient.EXPECT().ListVmssFlexVMsWithOutInstanceView(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedVMs, nil).AnyTimes()
		}
		err := provider.azureManager.forceRefresh()
		assert.NoError(t, err)

		ss := newTestScaleSet(provider.azureManager, "test-asg-doesnt-exist")
		err = ss.IncreaseSize(100)
		expectedErr := fmt.Errorf("could not find vmss: test-asg-doesnt-exist")
		assert.Equal(t, expectedErr, err)

		registered := provider.azureManager.RegisterNodeGroup(
			newTestScaleSet(provider.azureManager, testASG))
		assert.True(t, registered)
		assert.Equal(t, len(provider.NodeGroups()), 1)

		// Current target size is 3.
		targetSize, err := provider.NodeGroups()[0].TargetSize()
		assert.NoError(t, err)
		assert.Equal(t, 3, targetSize)

		// Increase 2 nodes.
		err = provider.NodeGroups()[0].IncreaseSize(2)
		assert.NoError(t, err)

		// New target size should be 5.
		targetSize, err = provider.NodeGroups()[0].TargetSize()
		assert.NoError(t, err)
		assert.Equal(t, 5, targetSize)

		// Testing Edge Zone scenario. Scale from 3 to 5.
		registeredForEdgeZone := provider.azureManager.RegisterNodeGroup(
			newTestScaleSet(provider.azureManager, "edgezone-vmss"))
		assert.True(t, registeredForEdgeZone)
		assert.Equal(t, len(provider.NodeGroups()), 2)

		targetSizeForEdgeZone, err := provider.NodeGroups()[1].TargetSize()
		assert.NoError(t, err)
		assert.Equal(t, 3, targetSizeForEdgeZone)

		mockVMSSClient.EXPECT().CreateOrUpdate(gomock.Any(), provider.azureManager.config.ResourceGroup,
			"edgezone-vmss", gomock.Any()).Return(nil, nil).AnyTimes()
		err = provider.NodeGroups()[1].IncreaseSize(2)
		assert.NoError(t, err)

		targetSizeForEdgeZone, err = provider.NodeGroups()[1].TargetSize()
		assert.NoError(t, err)
		assert.Equal(t, 5, targetSizeForEdgeZone)

		// Testing Edge Zone scenario scaleFromZero case. Scale from 0 to 2.
		registeredForEdgeZoneMinZero := provider.azureManager.RegisterNodeGroup(
			newTestScaleSetMinSizeZero(provider.azureManager, "edgezone-minzero-vmss"))
		assert.True(t, registeredForEdgeZoneMinZero)
		assert.Equal(t, len(provider.NodeGroups()), 3)

		// Current target size is 0.
		targetSizeForEdgeZoneMinZero, err := provider.NodeGroups()[2].TargetSize()
		assert.NoError(t, err)
		assert.Equal(t, 0, targetSizeForEdgeZoneMinZero)

		mockVMSSClient.EXPECT().CreateOrUpdate(gomock.Any(), provider.azureManager.config.ResourceGroup,
			"edgezone-minzero-vmss", gomock.Any()).Return(nil, nil).AnyTimes()
		err = provider.NodeGroups()[2].IncreaseSize(2)
		assert.NoError(t, err)

		// New target size should be 2.
		targetSizeForEdgeZoneMinZero, err = provider.NodeGroups()[2].TargetSize()
		assert.NoError(t, err)
		assert.Equal(t, 2, targetSizeForEdgeZoneMinZero)
	}
}

// TestScaleSetIncreaseSizeRaceCondition reproduces a data race between the
// non-atomic createOrUpdateInstances code path and concurrent readers.
//
// This test must be run with `-race` to detect the race. Without -race it is
// a no-op smoke test. `make test-unit` runs the suite with -race.
func TestScaleSetIncreaseSizeRaceCondition(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedScaleSets := newTestVMSSList(3, testASG, testLocation, armcompute.OrchestrationModeUniform)
	expectedVMSSVMs := newTestVMSSVMList(3)

	provider := newTestProvider(t)

	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

	// BeginCreateOrUpdate signals that the writer is parked inside the call
	// (sizeMutex still held by initCreateOrUpdate's deferred Unlock), then
	// blocks until the test releases it. This lets us start reader goroutines
	// while the lock is held so they will be racing against the unprotected
	// writes that follow once initCreateOrUpdate returns.
	started := make(chan struct{})
	release := make(chan struct{})
	mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
	mockDeleteClient.EXPECT().BeginCreateOrUpdate(
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).DoAndReturn(func(_ context.Context, _, _ string, _ armcompute.VirtualMachineScaleSet, _ *armcompute.VirtualMachineScaleSetsClientBeginCreateOrUpdateOptions) (*runtime.Poller[armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse], error) {
		close(started)
		<-release
		return nil, nil
	})
	provider.azureManager.azClient.vmssClientForDelete = mockDeleteClient

	mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)
	mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return([]*armcompute.VirtualMachine{}, nil).AnyTimes()
	provider.azureManager.azClient.virtualMachinesClient = mockVMClient

	mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), provider.azureManager.config.ResourceGroup, testASG).Return(expectedVMSSVMs, nil).AnyTimes()
	provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient

	err := provider.azureManager.forceRefresh()
	assert.NoError(t, err)

	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, testASG))
	assert.True(t, registered)

	ng := provider.NodeGroups()[0]

	// Kick off the writer (non-atomic IncreaseSize -> createOrUpdateInstances).
	increaseDone := make(chan error, 1)
	go func() {
		increaseDone <- ng.IncreaseSize(2)
	}()

	// Wait until createOrUpdateInstances is parked inside BeginCreateOrUpdate
	// with sizeMutex held.
	<-started

	// Spin concurrent readers that touch curSize / lastSizeRefresh under
	// sizeMutex. Once we release the writer, it will exit BeginCreateOrUpdate,
	// drop sizeMutex, and perform the (unprotected) writes to curSize /
	// lastSizeRefresh. The race detector will flag the unsynchronized access.
	var wg sync.WaitGroup
	stop := make(chan struct{})
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					_, _ = ng.TargetSize()
				}
			}
		}()
	}

	// Give readers a moment to start contending for sizeMutex, then release
	// the writer so the unprotected writes execute concurrently with reads.
	time.Sleep(10 * time.Millisecond)
	close(release)

	err = <-increaseDone
	assert.NoError(t, err)

	// Let the readers run a bit longer past the unprotected writes to widen
	// the race window, then stop them.
	time.Sleep(10 * time.Millisecond)
	close(stop)
	wg.Wait()
}

func TestScaleSetAtomicIncreaseSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedScaleSets := newTestVMSSList(3, testASG, "eastus", armcompute.OrchestrationModeUniform)
	expectedVMSSVMs := newTestVMSSVMList(3)

	provider := newTestProvider(t)

	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

	// BeginCreateOrUpdate returns nil poller — simulates immediate success.
	// Verify the request payload contains the expected new capacity (3 + 2 = 5).
	mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
	mockDeleteClient.EXPECT().BeginCreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Cond(func(x any) bool {
			vmss, ok := x.(armcompute.VirtualMachineScaleSet)
			return ok && vmss.SKU != nil && vmss.SKU.Capacity != nil && *vmss.SKU.Capacity == 5
		}),
		gomock.Any()).Return(nil, nil)
	provider.azureManager.azClient.vmssClientForDelete = mockDeleteClient

	mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)
	mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return([]*armcompute.VirtualMachine{}, nil).AnyTimes()
	provider.azureManager.azClient.virtualMachinesClient = mockVMClient

	mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), provider.azureManager.config.ResourceGroup, testASG).Return(expectedVMSSVMs, nil).AnyTimes()
	provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient

	err := provider.azureManager.forceRefresh()
	assert.NoError(t, err)

	// Error: non-existent scale set
	ss := newTestScaleSet(provider.azureManager, "test-asg-doesnt-exist")
	err = ss.AtomicIncreaseSize(1)
	expectedErr := fmt.Errorf("could not find vmss: test-asg-doesnt-exist")
	assert.Equal(t, expectedErr, err)

	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, testASG))
	assert.True(t, registered)

	// Error: negative delta
	err = provider.NodeGroups()[0].AtomicIncreaseSize(-1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "size increase must be positive")

	// Error: zero delta
	err = provider.NodeGroups()[0].AtomicIncreaseSize(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "size increase must be positive")

	// Error: exceeds max size (max is 5, current is 3, delta 3 = 6 > 5)
	err = provider.NodeGroups()[0].AtomicIncreaseSize(3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "size increase too large")

	// Current target size is 3.
	targetSize, err := provider.NodeGroups()[0].TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 3, targetSize)

	// Success: atomic increase by 2 (blocks until complete, nil poller = immediate).
	err = provider.NodeGroups()[0].AtomicIncreaseSize(2)
	assert.NoError(t, err)

	// New target size should be 5.
	targetSize, err = provider.NodeGroups()[0].TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 5, targetSize)
}

func TestScaleSetAtomicIncreaseSizeFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedScaleSets := newTestVMSSList(3, testASG, "eastus", armcompute.OrchestrationModeUniform)
	expectedVMSSVMs := newTestVMSSVMList(3)

	provider := newTestProvider(t)

	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

	// BeginCreateOrUpdate returns an error — simulates Azure rejecting the request.
	// Verify the request payload contains the expected new capacity (3 + 2 = 5).
	mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
	mockDeleteClient.EXPECT().BeginCreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Cond(func(x any) bool {
			vmss, ok := x.(armcompute.VirtualMachineScaleSet)
			return ok && vmss.SKU != nil && vmss.SKU.Capacity != nil && *vmss.SKU.Capacity == 5
		}),
		gomock.Any()).
		Return(nil, fmt.Errorf("azure capacity unavailable"))
	provider.azureManager.azClient.vmssClientForDelete = mockDeleteClient

	mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)
	mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return([]*armcompute.VirtualMachine{}, nil).AnyTimes()
	provider.azureManager.azClient.virtualMachinesClient = mockVMClient

	mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), provider.azureManager.config.ResourceGroup, testASG).Return(expectedVMSSVMs, nil).AnyTimes()
	provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient

	err := provider.azureManager.forceRefresh()
	assert.NoError(t, err)

	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, testASG))
	assert.True(t, registered)

	// BeginCreateOrUpdate fails — size should NOT be updated.
	err = provider.NodeGroups()[0].AtomicIncreaseSize(2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "azure capacity unavailable")

	// Target size should remain 3 (not updated on failure).
	targetSize, err := provider.NodeGroups()[0].TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 3, targetSize)
}

// fakeErrorPollerHandler simulates a poller that fails during Poll.
type fakeErrorPollerHandler[T any] struct {
	pollErr error
}

func (f *fakeErrorPollerHandler[T]) Done() bool {
	return false
}

func (f *fakeErrorPollerHandler[T]) Poll(ctx context.Context) (*http.Response, error) {
	return nil, f.pollErr
}

func (f *fakeErrorPollerHandler[T]) Result(ctx context.Context, out *T) error {
	return nil
}

func TestScaleSetAtomicIncreaseSizePollerFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedScaleSets := newTestVMSSList(3, testASG, "eastus", armcompute.OrchestrationModeUniform)
	expectedVMSSVMs := newTestVMSSVMList(3)

	provider := newTestProvider(t)

	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

	// Create a poller that will fail during PollUntilDone
	pollerResp := &http.Response{
		Header: http.Header{},
	}
	failingPoller, pollerErr := runtime.NewPoller(pollerResp, runtime.Pipeline{},
		&runtime.NewPollerOptions[armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse]{
			Handler: &fakeErrorPollerHandler[armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse]{
				pollErr: fmt.Errorf("long running operation failed"),
			},
		})
	assert.NoError(t, pollerErr)

	// BeginCreateOrUpdate succeeds (returns a poller), but PollUntilDone will fail.
	// Verify the request payload contains the expected new capacity (3 + 2 = 5).
	mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
	mockDeleteClient.EXPECT().BeginCreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Cond(func(x any) bool {
			vmss, ok := x.(armcompute.VirtualMachineScaleSet)
			return ok && vmss.SKU != nil && vmss.SKU.Capacity != nil && *vmss.SKU.Capacity == 5
		}),
		gomock.Any()).
		Return(failingPoller, nil)
	provider.azureManager.azClient.vmssClientForDelete = mockDeleteClient

	mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)
	mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return([]*armcompute.VirtualMachine{}, nil).AnyTimes()
	provider.azureManager.azClient.virtualMachinesClient = mockVMClient

	mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), provider.azureManager.config.ResourceGroup, testASG).Return(expectedVMSSVMs, nil).AnyTimes()
	provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient

	err := provider.azureManager.forceRefresh()
	assert.NoError(t, err)

	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, testASG))
	assert.True(t, registered)

	// Current target size is 3.
	targetSize, err := provider.NodeGroups()[0].TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 3, targetSize)

	// PollUntilDone fails — size should NOT be updated.
	err = provider.NodeGroups()[0].AtomicIncreaseSize(2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "long running operation failed")

	// Target size should remain 3 (not updated on poller failure).
	targetSize, err = provider.NodeGroups()[0].TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 3, targetSize)
}

// TestIncreaseSizeOnVMProvisioningFailed has been tweeked only for Uniform Orchestration mode.
// If ProvisioningState == failed and power state is not running, Status.State == InstanceCreating with errorInfo populated.
func TestScaleSetIncreaseSizeOnVMProvisioningFailed(t *testing.T) {
	testCases := map[string]struct {
		expectInstanceRunning    bool
		isMissingInstanceView    bool
		statuses                 []*armcompute.InstanceViewStatus
		expectErrorInfoPopulated bool
	}{
		"out of resources when no power state exists": {
			expectErrorInfoPopulated: false,
		},
		"out of resources when VM is stopped": {
			statuses:                 []*armcompute.InstanceViewStatus{{Code: ptr.To(vmPowerStateStopped)}},
			expectErrorInfoPopulated: false,
		},
		"out of resources when VM reports invalid power state": {
			statuses:                 []*armcompute.InstanceViewStatus{{Code: ptr.To("PowerState/invalid")}},
			expectErrorInfoPopulated: false,
		},
		"instance running when power state is running": {
			expectInstanceRunning:    true,
			statuses:                 []*armcompute.InstanceViewStatus{{Code: ptr.To(vmPowerStateRunning)}},
			expectErrorInfoPopulated: false,
		},
		"instance running if instance view cannot be retrieved": {
			expectInstanceRunning:    true,
			isMissingInstanceView:    true,
			expectErrorInfoPopulated: false,
		},
	}
	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			manager := newTestAzureManager(t)
			vmssName := "vmss-failed-upscale"

			expectedScaleSets := newTestVMSSList(3, "vmss-failed-upscale", "eastus", armcompute.OrchestrationModeUniform)
			expectedVMSSVMs := newTestVMSSVMList(3)
			// The failed state is important line of code here
			expectedVMs := newTestVMList(3)
			expectedVMSSVMs[2].Properties.ProvisioningState = ptr.To(VMProvisioningStateFailed)
			if !testCase.isMissingInstanceView {
				expectedVMSSVMs[2].Properties.InstanceView = &armcompute.VirtualMachineScaleSetVMInstanceView{Statuses: testCase.statuses}
			}

			mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
			mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil)
			manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

			// Mock the vmssClientForDelete for async CreateOrUpdate calls
			mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
			mockDeleteClient.EXPECT().BeginCreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
			mockDeleteClient.EXPECT().BeginDeleteInstances(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
			manager.azClient.vmssClientForDelete = mockDeleteClient

			mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
			mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, "vmss-failed-upscale").Return(expectedVMSSVMs, nil).AnyTimes()
			manager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient

			mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)
			mockVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedVMs, nil).AnyTimes()
			manager.azClient.virtualMachinesClient = mockVMClient

			manager.explicitlyConfigured["vmss-failed-upscale"] = true
			registered := manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))
			assert.True(t, registered)
			manager.Refresh()

			provider, err := BuildAzureCloudProvider(manager, nil)
			assert.NoError(t, err)

			scaleSet, ok := provider.NodeGroups()[0].(*ScaleSet)
			assert.True(t, ok)

			// Increase size by one, but the new node fails provisioning
			err = scaleSet.IncreaseSize(1)
			assert.NoError(t, err)

			nodes, err := scaleSet.Nodes()
			assert.NoError(t, err)

			assert.Equal(t, 3, len(nodes))

			assert.Equal(t, testCase.expectErrorInfoPopulated, nodes[2].Status.ErrorInfo != nil)
			if testCase.expectErrorInfoPopulated {
				assert.Equal(t, cloudprovider.InstanceCreating, nodes[2].Status.State)
			} else {
				assert.Equal(t, cloudprovider.InstanceRunning, nodes[2].Status.State)
			}
		})
	}
}

func TestIncreaseSizeOnVMProvisioningFailedWithFastDelete(t *testing.T) {
	testCases := map[string]struct {
		expectInstanceRunning    bool
		isMissingInstanceView    bool
		statuses                 []*armcompute.InstanceViewStatus
		expectErrorInfoPopulated bool
	}{
		"out of resources when no power state exists": {
			expectErrorInfoPopulated: true,
		},
		"out of resources when VM is stopped": {
			statuses:                 []*armcompute.InstanceViewStatus{{Code: ptr.To(vmPowerStateStopped)}},
			expectErrorInfoPopulated: true,
		},
		"out of resources when VM reports invalid power state": {
			statuses:                 []*armcompute.InstanceViewStatus{{Code: ptr.To("PowerState/invalid")}},
			expectErrorInfoPopulated: true,
		},
		"instance running when power state is running": {
			expectInstanceRunning:    true,
			statuses:                 []*armcompute.InstanceViewStatus{{Code: ptr.To(vmPowerStateRunning)}},
			expectErrorInfoPopulated: false,
		},
		"instance running if instance view cannot be retrieved": {
			expectInstanceRunning:    true,
			isMissingInstanceView:    true,
			expectErrorInfoPopulated: false,
		},
	}
	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			manager := newTestAzureManager(t)
			vmssName := "vmss-failed-upscale"

			expectedScaleSets := newTestVMSSList(3, "vmss-failed-upscale", "eastus", armcompute.OrchestrationModeUniform)
			expectedVMSSVMs := newTestVMSSVMList(3)
			// The failed state is important line of code here
			expectedVMs := newTestVMList(3)
			expectedVMSSVMs[2].Properties.ProvisioningState = ptr.To(VMProvisioningStateFailed)
			if !testCase.isMissingInstanceView {
				expectedVMSSVMs[2].Properties.InstanceView = &armcompute.VirtualMachineScaleSetVMInstanceView{Statuses: testCase.statuses}
			}

			mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
			mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil)
			manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

			// Mock the vmssClientForDelete for async CreateOrUpdate calls
			mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
			mockDeleteClient.EXPECT().BeginCreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
			mockDeleteClient.EXPECT().BeginDeleteInstances(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
			manager.azClient.vmssClientForDelete = mockDeleteClient

			mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
			mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, "vmss-failed-upscale").Return(expectedVMSSVMs, nil).AnyTimes()
			manager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient

			mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)
			mockVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedVMs, nil).AnyTimes()
			manager.azClient.virtualMachinesClient = mockVMClient

			manager.explicitlyConfigured["vmss-failed-upscale"] = true
			registered := manager.RegisterNodeGroup(newTestScaleSetWithFastDelete(manager, vmssName))
			assert.True(t, registered)
			manager.Refresh()

			provider, err := BuildAzureCloudProvider(manager, nil)
			assert.NoError(t, err)

			scaleSet, ok := provider.NodeGroups()[0].(*ScaleSet)
			assert.True(t, ok)

			// Increase size by one, but the new node fails provisioning
			err = scaleSet.IncreaseSize(1)
			assert.NoError(t, err)

			nodes, err := scaleSet.Nodes()
			assert.NoError(t, err)

			assert.Equal(t, 3, len(nodes))

			assert.Equal(t, testCase.expectErrorInfoPopulated, nodes[2].Status.ErrorInfo != nil)
			if testCase.expectErrorInfoPopulated {
				assert.Equal(t, cloudprovider.InstanceCreating, nodes[2].Status.State)
			} else {
				assert.Equal(t, cloudprovider.InstanceRunning, nodes[2].Status.State)
			}
		})
	}
}

func TestScaleSetIncreaseSizeOnVMSSUpdating(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := newTestAzureManager(t)
	vmssName := "vmss-updating"
	var vmssCapacity int64 = 3
	orchMode := armcompute.OrchestrationModeUniform

	expectedScaleSets := []*armcompute.VirtualMachineScaleSet{
		{
			Name: &vmssName,
			SKU: &armcompute.SKU{
				Capacity: &vmssCapacity,
			},
			Properties: &armcompute.VirtualMachineScaleSetProperties{
				ProvisioningState: ptr.To("Updating"),
				OrchestrationMode: &orchMode,
			},
		},
	}
	expectedVMSSVMs := newTestVMSSVMList(3)

	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil)
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

	// Mock the vmssClientForDelete for async CreateOrUpdate calls
	mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
	mockDeleteClient.EXPECT().BeginCreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	mockDeleteClient.EXPECT().BeginDeleteInstances(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	manager.azClient.vmssClientForDelete = mockDeleteClient

	mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, "vmss-updating").Return(expectedVMSSVMs, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
	manager.explicitlyConfigured["vmss-updating"] = true
	registered := manager.RegisterNodeGroup(newTestScaleSet(manager, vmssName))
	assert.True(t, registered)

	err := manager.Refresh()
	assert.NoError(t, err)

	provider, err := BuildAzureCloudProvider(manager, nil)
	assert.NoError(t, err)

	// Scaling should continue even VMSS is under updating.
	scaleSet, ok := provider.NodeGroups()[0].(*ScaleSet)
	assert.True(t, ok)
	err = scaleSet.IncreaseSize(1)
	assert.NoError(t, err)
}

func TestScaleSetBelongs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orchestrationModes := [2]armcompute.OrchestrationMode{armcompute.OrchestrationModeUniform, armcompute.OrchestrationModeFlexible}
	expectedVMSSVMs := newTestVMSSVMList(3)
	expectedVMs := newTestVMList(3)

	for _, orchMode := range orchestrationModes {
		expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus", orchMode)
		provider := newTestProvider(t)
		mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
		mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
		provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
		mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)
		provider.azureManager.azClient.virtualMachinesClient = mockVMClient

		if orchMode == armcompute.OrchestrationModeUniform {
			mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
			mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), provider.azureManager.config.ResourceGroup, "test-asg").Return(expectedVMSSVMs, nil).AnyTimes()
			provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
			mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return([]*armcompute.VirtualMachine{}, nil).AnyTimes()
		} else {
			provider.azureManager.config.EnableVmssFlexNodes = true
			mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedVMs, nil).AnyTimes()
			mockVMClient.EXPECT().ListVmssFlexVMsWithOutInstanceView(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedVMs, nil).AnyTimes()
		}

		registered := provider.azureManager.RegisterNodeGroup(
			newTestScaleSet(provider.azureManager, testASG))
		assert.True(t, registered)

		scaleSet, ok := provider.NodeGroups()[0].(*ScaleSet)
		assert.True(t, ok)
		provider.azureManager.explicitlyConfigured["test-asg"] = true
		err := provider.azureManager.forceRefresh()
		assert.NoError(t, err)

		invalidNode := &apiv1.Node{
			Spec: apiv1.NodeSpec{
				ProviderID: azurePrefix + "/subscriptions/test-subscrition-id/resourcegroups/invalid-asg/providers/microsoft.compute/virtualmachinescalesets/agents/virtualmachines/0",
			},
		}
		_, err = scaleSet.Belongs(invalidNode)
		assert.Error(t, err)

		validNode := newApiNode(orchMode, 0)
		belongs, err := scaleSet.Belongs(validNode)
		assert.Equal(t, true, belongs)
		assert.NoError(t, err)
	}
}

func TestScaleSetDeleteNodes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	vmssName := "test-asg"
	var vmssCapacity int64 = 3
	cases := []struct {
		name              string
		orchestrationMode armcompute.OrchestrationMode
		enableForceDelete bool
	}{
		{
			name:              "uniform, force delete enabled",
			orchestrationMode: armcompute.OrchestrationModeUniform,
			enableForceDelete: true,
		},
		{
			name:              "uniform, force delete disabled",
			orchestrationMode: armcompute.OrchestrationModeUniform,
			enableForceDelete: false,
		},
		{
			name:              "flexible, force delete enabled",
			orchestrationMode: armcompute.OrchestrationModeFlexible,
			enableForceDelete: true,
		},
		{
			name:              "flexible, force delete disabled",
			orchestrationMode: armcompute.OrchestrationModeFlexible,
			enableForceDelete: false,
		},
	}

	for _, tc := range cases {
		orchMode := tc.orchestrationMode
		enableForceDelete := tc.enableForceDelete

		expectedVMSSVMs := newTestVMSSVMList(3)
		expectedVMs := newTestVMList(3)

		manager := newTestAzureManager(t)
		manager.config.EnableForceDelete = enableForceDelete
		expectedScaleSets := newTestVMSSList(vmssCapacity, vmssName, "eastus", orchMode)
		fmt.Printf("orchMode: %s, enableForceDelete: %t\n", orchMode, enableForceDelete)

		mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
		mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).Times(2)
		manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

		// Mock the delete client to accept BeginDeleteInstances calls
		mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
		mockDeleteClient.EXPECT().BeginDeleteInstances(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		manager.azClient.vmssClientForDelete = mockDeleteClient

		mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
		mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)
		manager.azClient.virtualMachinesClient = mockVMClient
		mockVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedVMs, nil).AnyTimes()

		if orchMode == armcompute.OrchestrationModeUniform {
			mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, "test-asg").Return(expectedVMSSVMs, nil).AnyTimes()
			manager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
		} else {
			manager.config.EnableVmssFlexNodes = true
			mockVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedVMs, nil).AnyTimes()
			mockVMClient.EXPECT().ListVmssFlexVMsWithOutInstanceView(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedVMs, nil).AnyTimes()
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

		// Perform the delete operation
		nodesToDelete := []*apiv1.Node{
			newApiNode(orchMode, 0),
			newApiNode(orchMode, 2),
		}
		err = scaleSet.DeleteNodes(nodesToDelete)
		assert.NoError(t, err)

		// create scale set with vmss capacity 1
		expectedScaleSets = newTestVMSSList(1, vmssName, "eastus", orchMode)

		mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()

		if orchMode == armcompute.OrchestrationModeUniform {
			expectedVMSSVMs[0].Properties.ProvisioningState = ptr.To(VMProvisioningStateDeleting)
			expectedVMSSVMs[2].Properties.ProvisioningState = ptr.To(VMProvisioningStateDeleting)
			mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, "test-asg").Return(expectedVMSSVMs, nil).AnyTimes()
		} else {
			expectedVMs[0].Properties.ProvisioningState = ptr.To(VMProvisioningStateDeleting)
			expectedVMs[2].Properties.ProvisioningState = ptr.To(VMProvisioningStateDeleting)
			mockVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedVMs, nil).AnyTimes()
		}

		err = manager.forceRefresh()
		assert.NoError(t, err)

		// Ensure the the cached size has been proactively decremented by 2
		targetSize, err = scaleSet.TargetSize()
		assert.NoError(t, err)
		assert.Equal(t, 1, targetSize)

		// Ensure that the status for the instances is Deleting
		// lastInstanceRefresh is set to time.Now() to avoid resetting instanceCache.
		scaleSet.lastInstanceRefresh = time.Now()
		instance0, found, err := scaleSet.getInstanceByProviderID(nodesToDelete[0].Spec.ProviderID)
		assert.True(t, found, true)
		assert.NoError(t, err)
		assert.Equal(t, instance0.Status.State, cloudprovider.InstanceDeleting)

		// lastInstanceRefresh is set to time.Now() to avoid resetting instanceCache.
		scaleSet.lastInstanceRefresh = time.Now()
		instance2, found, err := scaleSet.getInstanceByProviderID(nodesToDelete[1].Spec.ProviderID)
		assert.True(t, found, true)
		assert.NoError(t, err)
		assert.Equal(t, instance2.Status.State, cloudprovider.InstanceDeleting)
	}
}

func TestScaleSetForceDeleteNodesDoesNotPublishNegativeCachedSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	vmssName := testASG
	var vmssCapacity int64 = 3
	orchMode := armcompute.OrchestrationModeUniform
	expectedScaleSets := newTestVMSSList(vmssCapacity, vmssName, "eastus", orchMode)
	expectedVMSSVMs := newTestVMSSVMList(3)
	expectedVMs := newTestVMList(3)

	manager := newTestAzureManager(t)

	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

	mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
	mockDeleteClient.EXPECT().BeginDeleteInstances(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	manager.azClient.vmssClientForDelete = mockDeleteClient

	mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, testASG).Return(expectedVMSSVMs, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient

	mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)
	mockVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedVMs, nil).AnyTimes()
	manager.azClient.virtualMachinesClient = mockVMClient

	err := manager.forceRefresh()
	assert.NoError(t, err)

	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	provider, err := BuildAzureCloudProvider(manager, resourceLimiter)
	assert.NoError(t, err)

	registered := manager.RegisterNodeGroup(newTestScaleSet(manager, testASG))
	manager.explicitlyConfigured[testASG] = true
	assert.True(t, registered)
	err = manager.forceRefresh()
	assert.NoError(t, err)

	scaleSet, ok := provider.NodeGroups()[0].(*ScaleSet)
	assert.True(t, ok)

	targetSize, err := scaleSet.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 3, targetSize)

	scaleSet.curSize = 1
	scaleSet.lastSizeRefresh = time.Now()
	scaleSet.sizeRefreshPeriod = time.Hour

	nodesToDelete := []*apiv1.Node{
		newApiNode(orchMode, 0),
		newApiNode(orchMode, 2),
	}
	err = scaleSet.ForceDeleteNodes(nodesToDelete)
	assert.NoError(t, err)

	scaleSet.sizeMutex.Lock()
	curSize := scaleSet.curSize
	lastSizeRefresh := scaleSet.lastSizeRefresh
	scaleSet.sizeMutex.Unlock()
	assert.Equal(t, int64(1), curSize)
	assert.True(t, lastSizeRefresh.IsZero())

	targetSize, err = scaleSet.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 3, targetSize)
}

func TestScaleSetDeleteNodeUnregistered(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	vmssName := "test-asg"
	var vmssCapacity int64 = 2

	cases := []struct {
		name              string
		orchestrationMode armcompute.OrchestrationMode
		enableForceDelete bool
	}{
		{
			name:              "uniform, force delete enabled",
			orchestrationMode: armcompute.OrchestrationModeUniform,
			enableForceDelete: true,
		},
		{
			name:              "uniform, force delete disabled",
			orchestrationMode: armcompute.OrchestrationModeUniform,
			enableForceDelete: false,
		},
		{
			name:              "flexible, force delete enabled",
			orchestrationMode: armcompute.OrchestrationModeFlexible,
			enableForceDelete: true,
		},
		{
			name:              "flexible, force delete disabled",
			orchestrationMode: armcompute.OrchestrationModeFlexible,
			enableForceDelete: false,
		},
	}

	for _, tc := range cases {
		orchMode := tc.orchestrationMode
		enableForceDelete := tc.enableForceDelete
		expectedVMSSVMs := newTestVMSSVMList(2)
		expectedVMs := newTestVMList(2)

		manager := newTestAzureManager(t)
		manager.config.EnableForceDelete = enableForceDelete
		expectedScaleSets := newTestVMSSList(vmssCapacity, vmssName, "eastus", orchMode)

		mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
		mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).Times(2)
		manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

		// Mock the delete client to accept BeginDeleteInstances calls
		mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
		mockDeleteClient.EXPECT().BeginDeleteInstances(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
		manager.azClient.vmssClientForDelete = mockDeleteClient

		mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)
		mockVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedVMs, nil).AnyTimes()
		manager.azClient.virtualMachinesClient = mockVMClient

		if orchMode == armcompute.OrchestrationModeUniform {
			mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
			mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, "test-asg").Return(expectedVMSSVMs, nil).AnyTimes()
			manager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
		} else {
			manager.config.EnableVmssFlexNodes = true
			mockVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedVMs, nil).AnyTimes()
			mockVMClient.EXPECT().ListVmssFlexVMsWithOutInstanceView(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedVMs, nil).AnyTimes()
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
		scaleSet.instancesRefreshPeriod = defaultVmssInstancesRefreshPeriod

		targetSize, err := scaleSet.TargetSize()
		assert.NoError(t, err)
		assert.Equal(t, 2, targetSize)

		// annotate node with unregistered annotation
		annotations := make(map[string]string)
		annotations[cloudprovider.FakeNodeReasonAnnotation] = cloudprovider.FakeNodeUnregistered

		nodesToDelete := []*apiv1.Node{
			newApiNode(orchMode, 0),
		}
		nodesToDelete[0].ObjectMeta.Annotations = annotations

		err = scaleSet.DeleteNodes(nodesToDelete)
		assert.NoError(t, err)

		// Ensure the the cached size has NOT been proactively decremented
		targetSize, err = scaleSet.TargetSize()
		assert.NoError(t, err)
		assert.Equal(t, 2, targetSize)

		// Ensure that the status for the instances is Deleting
		// lastInstanceRefresh is set to time.Now() to avoid resetting instanceCache.
		scaleSet.lastInstanceRefresh = time.Now()
		instance0, found, err := scaleSet.getInstanceByProviderID(nodesToDelete[0].Spec.ProviderID)
		assert.True(t, found, true)
		assert.NoError(t, err)
		assert.Equal(t, cloudprovider.InstanceDeleting, instance0.Status.State)
	}
}

func TestScaleSetDeleteInstancesWithForceDeleteEnabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := newTestAzureManager(t)
	// enabling forceDelete
	manager.config.EnableForceDelete = true

	vmssName := "test-asg"
	var vmssCapacity int64 = 3
	//hostGroupId := "test-hostGroup"
	//hostGroup := &compute.SubResource{
	//	ID: &hostGroupId,
	//}
	orchMode := armcompute.OrchestrationModeUniform

	expectedScaleSets := []*armcompute.VirtualMachineScaleSet{
		{
			Name: &vmssName,
			SKU: &armcompute.SKU{
				Capacity: &vmssCapacity,
			},
			Properties: &armcompute.VirtualMachineScaleSetProperties{
				OrchestrationMode: &orchMode,
			},
		},
	}
	expectedVMSSVMs := newTestVMSSVMList(3)

	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).Times(2)
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

	// Mock the delete client to accept BeginDeleteInstances calls
	mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
	mockDeleteClient.EXPECT().BeginDeleteInstances(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	manager.azClient.vmssClientForDelete = mockDeleteClient

	mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, "test-asg").Return(expectedVMSSVMs, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
	err := manager.forceRefresh()
	assert.NoError(t, err)

	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	provider, err := BuildAzureCloudProvider(manager, resourceLimiter)
	assert.NoError(t, err)

	registered := manager.RegisterNodeGroup(
		newTestScaleSet(manager, "test-asg"))
	manager.explicitlyConfigured["test-asg"] = true
	assert.True(t, registered)
	err = manager.forceRefresh()
	assert.NoError(t, err)

	scaleSet, ok := provider.NodeGroups()[0].(*ScaleSet)
	assert.True(t, ok)

	targetSize, err := scaleSet.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 3, targetSize)

	// Perform the delete operation
	nodesToDelete := []*apiv1.Node{
		{
			Spec: apiv1.NodeSpec{
				ProviderID: azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 0),
			},
		},
		{
			Spec: apiv1.NodeSpec{
				ProviderID: azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 2),
			},
		},
	}
	err = scaleSet.DeleteNodes(nodesToDelete)
	assert.NoError(t, err)
	vmssCapacity = 1
	orchMode = armcompute.OrchestrationModeUniform
	expectedScaleSets = []*armcompute.VirtualMachineScaleSet{
		{
			Name: &vmssName,
			SKU: &armcompute.SKU{
				Capacity: &vmssCapacity,
			},
			Properties: &armcompute.VirtualMachineScaleSetProperties{
				OrchestrationMode: &orchMode,
			},
		},
	}
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	expectedVMSSVMs[0].Properties.ProvisioningState = ptr.To(VMProvisioningStateDeleting)
	expectedVMSSVMs[2].Properties.ProvisioningState = ptr.To(VMProvisioningStateDeleting)
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, "test-asg").Return(expectedVMSSVMs, nil).AnyTimes()
	err = manager.forceRefresh()
	assert.NoError(t, err)

	// Ensure the the cached size has been proactively decremented by 2
	targetSize, err = scaleSet.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 1, targetSize)

	// Ensure that the status for the instances is Deleting
	// lastInstanceRefresh is set to time.Now() to avoid resetting instanceCache.
	scaleSet.lastInstanceRefresh = time.Now()
	instance0, found, err := scaleSet.getInstanceByProviderID(azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 0))
	assert.True(t, found, true)
	assert.NoError(t, err)
	assert.Equal(t, instance0.Status.State, cloudprovider.InstanceDeleting)

	// lastInstanceRefresh is set to time.Now() to avoid resetting instanceCache.
	scaleSet.lastInstanceRefresh = time.Now()
	instance2, found, err := scaleSet.getInstanceByProviderID(azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 2))
	assert.True(t, found, true)
	assert.NoError(t, err)
	assert.Equal(t, instance2.Status.State, cloudprovider.InstanceDeleting)

}

func TestScaleSetDeleteNoConflictRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	vmssName := "test-asg"
	var vmssCapacity int64 = 3

	manager := newTestAzureManager(t)

	expectedVMSSVMs := []*armcompute.VirtualMachineScaleSetVM{
		{
			ID:         ptr.To(fakeVirtualMachineScaleSetVMID),
			InstanceID: ptr.To("0"),
			Properties: &armcompute.VirtualMachineScaleSetVMProperties{
				VMID:              ptr.To("123E4567-E89B-12D3-A456-426655440000"),
				ProvisioningState: ptr.To(VMProvisioningStateDeleting),
			},
		},
	}
	orchMode := armcompute.OrchestrationModeUniform
	expectedScaleSets := []*armcompute.VirtualMachineScaleSet{
		{
			Name: &vmssName,
			SKU: &armcompute.SKU{
				Capacity: &vmssCapacity,
			},
			Properties: &armcompute.VirtualMachineScaleSetProperties{
				OrchestrationMode: &orchMode,
			},
		},
	}

	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
	mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, "test-asg").Return(expectedVMSSVMs, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient

	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	provider, err := BuildAzureCloudProvider(manager, resourceLimiter)
	assert.NoError(t, err)

	registered := manager.RegisterNodeGroup(newTestScaleSet(manager, "test-asg"))
	manager.explicitlyConfigured["test-asg"] = true
	assert.True(t, registered)
	manager.Refresh()

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 0),
		},
	}

	scaleSet, ok := provider.NodeGroups()[0].(*ScaleSet)
	assert.True(t, ok)

	err = scaleSet.DeleteNodes([]*apiv1.Node{node})
}

func TestScaleSetId(t *testing.T) {
	provider := newTestProvider(t)
	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)
	assert.Equal(t, provider.NodeGroups()[0].Id(), "test-asg")
}

func TestAgentPoolDebug(t *testing.T) {
	asg := ScaleSet{
		manager: newTestAzureManager(t),
		minSize: 5,
		maxSize: 55,
	}
	asg.Name = "test-scale-set"
	assert.Equal(t, asg.Debug(), "test-scale-set (5:55)")
}

func TestScaleSetNodes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	orchestrationModes := [2]armcompute.OrchestrationMode{armcompute.OrchestrationModeUniform, armcompute.OrchestrationModeFlexible}

	expectedVMSSVMs := newTestVMSSVMList(3)
	expectedVMs := newTestVMList(3)

	for _, orchMode := range orchestrationModes {
		expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus", orchMode)
		provider := newTestProvider(t)

		mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
		mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
		provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
		mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)
		provider.azureManager.azClient.virtualMachinesClient = mockVMClient

		if orchMode == armcompute.OrchestrationModeUniform {
			mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
			mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), provider.azureManager.config.ResourceGroup, "test-asg").Return(expectedVMSSVMs, nil).AnyTimes()
			provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
			mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return([]*armcompute.VirtualMachine{}, nil).AnyTimes()

		} else {
			provider.azureManager.config.EnableVmssFlexNodes = true
			mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedVMs, nil).AnyTimes()
			mockVMClient.EXPECT().ListVmssFlexVMsWithOutInstanceView(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedVMs, nil).AnyTimes()
		}

		registered := provider.azureManager.RegisterNodeGroup(
			newTestScaleSet(provider.azureManager, "test-asg"))
		provider.azureManager.explicitlyConfigured["test-asg"] = true
		err := provider.azureManager.forceRefresh()
		assert.NoError(t, err)

		assert.True(t, registered)
		assert.Equal(t, len(provider.NodeGroups()), 1)

		node := newApiNode(orchMode, 0)
		group, err := provider.NodeGroupForNode(node)
		assert.NoError(t, err)
		assert.NotNil(t, group, "Group should not be nil")
		assert.Equal(t, group.Id(), testASG)
		assert.Equal(t, group.MinSize(), 1)
		assert.Equal(t, group.MaxSize(), 5)

		ss, ok := group.(*ScaleSet)
		ss.lastInstanceRefresh = time.Now()
		assert.True(t, ok)
		assert.NotNil(t, ss)
		instances, err := group.Nodes()
		assert.NoError(t, err)
		assert.Equal(t, len(instances), 3)

		if orchMode == armcompute.OrchestrationModeUniform {

			assert.Equal(t, instances[0], cloudprovider.Instance{Id: azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 0)})
			assert.Equal(t, instances[1], cloudprovider.Instance{Id: azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 1)})
			assert.Equal(t, instances[2], cloudprovider.Instance{Id: azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 2)})
		} else {

			assert.Equal(t, instances[0], cloudprovider.Instance{Id: azurePrefix + fmt.Sprintf(fakeVirtualMachineVMID, 0)})
			assert.Equal(t, instances[1], cloudprovider.Instance{Id: azurePrefix + fmt.Sprintf(fakeVirtualMachineVMID, 1)})
			assert.Equal(t, instances[2], cloudprovider.Instance{Id: azurePrefix + fmt.Sprintf(fakeVirtualMachineVMID, 2)})
		}
	}

}

func TestScaleSetEnableVmssFlexNodesFlag(t *testing.T) {

	// flag set to false
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedVMs := newTestVMList(3)
	expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus", armcompute.OrchestrationModeFlexible)

	provider := newTestProvider(t)
	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	provider.azureManager.config.EnableVmssFlexNodes = false
	provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
	mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)

	mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedVMs, nil).AnyTimes()
	mockVMClient.EXPECT().ListVmssFlexVMsWithOutInstanceView(gomock.Any(), gomock.Any(), gomock.Any()).Return(expectedVMs, nil).AnyTimes()
	provider.azureManager.azClient.virtualMachinesClient = mockVMClient

	provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, testASG))
	provider.azureManager.explicitlyConfigured[testASG] = true
	err := provider.azureManager.Refresh()
	assert.Error(t, err, "vmss - \"test-asg\" with Flexible orchestration detected but 'enbaleVmssFlex' feature flag is turned off")

	// flag set to true
	provider.azureManager.config.EnableVmssFlexNodes = true
	err = provider.azureManager.Refresh()
	assert.NoError(t, err)
}

func TestScaleSetTemplateNodeInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus", armcompute.OrchestrationModeUniform)

	provider := newTestProvider(t)
	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
	err := provider.azureManager.forceRefresh()
	assert.NoError(t, err)

	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)

	asg := ScaleSet{
		manager: provider.azureManager,
		minSize: 1,
		maxSize: 5,
	}
	asg.Name = "test-asg"

	// The dynamic SKU list ("cache") in the test provider is empty
	// (initialized with cfg.EnableDynamicInstanceList = false).
	assert.False(t, provider.azureManager.azureCache.HasVMSKUs())

	t.Run("Checking fallback to static because dynamic list is empty", func(t *testing.T) {
		asg.enableDynamicInstanceList = true

		nodeInfo, err := asg.TemplateNodeInfo()
		assert.NoError(t, err)
		assert.NotNil(t, nodeInfo)
		assert.NotEmpty(t, nodeInfo.Pods())
	})

	// Properly testing dynamic SKU list through skewer is not possible,
	// because there are no Resource API mocks included yet.
	// Instead, the rest of the (consumer side) tests here
	// override GetInstanceTypeDynamically and GetInstanceTypeStatically functions.

	t.Run("Checking dynamic workflow", func(t *testing.T) {
		asg.enableDynamicInstanceList = true

		GetInstanceTypeDynamically = func(template NodeTemplate, azCache *azureCache) (InstanceType, error) {
			vmssType := InstanceType{}
			vmssType.VCPU = 1
			vmssType.GPU = 2
			vmssType.MemoryMb = 3
			return vmssType, nil
		}
		nodeInfo, err := asg.TemplateNodeInfo()
		assert.Equal(t, *nodeInfo.Node().Status.Capacity.Cpu(), *resource.NewQuantity(1, resource.DecimalSI))
		assert.Equal(t, *nodeInfo.Node().Status.Capacity.Memory(), *resource.NewQuantity(3*1024*1024, resource.DecimalSI))
		assert.NoError(t, err)
		assert.NotNil(t, nodeInfo)
		assert.NotEmpty(t, nodeInfo.Pods())
	})

	t.Run("Checking static workflow if dynamic fails", func(t *testing.T) {
		asg.enableDynamicInstanceList = true

		GetInstanceTypeDynamically = func(template NodeTemplate, azCache *azureCache) (InstanceType, error) {
			return InstanceType{}, fmt.Errorf("dynamic error exists")
		}
		GetInstanceTypeStatically = func(template NodeTemplate) (*InstanceType, error) {
			vmssType := InstanceType{}
			vmssType.VCPU = 1
			vmssType.GPU = 2
			vmssType.MemoryMb = 3
			return &vmssType, nil
		}
		nodeInfo, err := asg.TemplateNodeInfo()
		assert.Equal(t, *nodeInfo.Node().Status.Capacity.Cpu(), *resource.NewQuantity(1, resource.DecimalSI))
		assert.Equal(t, *nodeInfo.Node().Status.Capacity.Memory(), *resource.NewQuantity(3*1024*1024, resource.DecimalSI))
		assert.NoError(t, err)
		assert.NotNil(t, nodeInfo)
		assert.NotEmpty(t, nodeInfo.Pods())
	})

	t.Run("Fails to find vmss instance information using static and dynamic workflow, instance not supported", func(t *testing.T) {
		asg.enableDynamicInstanceList = true

		GetInstanceTypeDynamically = func(template NodeTemplate, azCache *azureCache) (InstanceType, error) {
			return InstanceType{}, fmt.Errorf("dynamic error exists")
		}
		GetInstanceTypeStatically = func(template NodeTemplate) (*InstanceType, error) {
			return &InstanceType{}, fmt.Errorf("static error exists")
		}
		nodeInfo, err := asg.TemplateNodeInfo()
		assert.Empty(t, nodeInfo)
		assert.Equal(t, err, fmt.Errorf("static error exists"))
	})

	// Note: static-only workflow tests can be removed once support for dynamic is always on

	t.Run("Checking static-only workflow", func(t *testing.T) {
		asg.enableDynamicInstanceList = false

		GetInstanceTypeStatically = func(template NodeTemplate) (*InstanceType, error) {
			vmssType := InstanceType{}
			vmssType.VCPU = 1
			vmssType.GPU = 2
			vmssType.MemoryMb = 3
			return &vmssType, nil
		}
		nodeInfo, err := asg.TemplateNodeInfo()
		assert.Equal(t, *nodeInfo.Node().Status.Capacity.Cpu(), *resource.NewQuantity(1, resource.DecimalSI))
		assert.Equal(t, *nodeInfo.Node().Status.Capacity.Memory(), *resource.NewQuantity(3*1024*1024, resource.DecimalSI))
		assert.NoError(t, err)
		assert.NotNil(t, nodeInfo)
		assert.NotEmpty(t, nodeInfo.Pods())
	})

	t.Run("Checking static-only workflow with built-in SKU list", func(t *testing.T) {
		asg.enableDynamicInstanceList = false

		nodeInfo, err := asg.TemplateNodeInfo()
		assert.NoError(t, err)
		assert.NotNil(t, nodeInfo)
		assert.NotEmpty(t, nodeInfo.Pods())
	})

}
func TestScaleSetCseErrors(t *testing.T) {
	errorMessage := ptr.To("Error Message Test")
	vmssVMs := armcompute.VirtualMachineScaleSetVM{
		Name:       ptr.To("vmTest"),
		ID:         ptr.To(fakeVirtualMachineScaleSetVMID),
		InstanceID: ptr.To("0"),
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			VMID:              ptr.To("123E4567-E89B-12D3-A456-426655440000"),
			ProvisioningState: ptr.To("Succeeded"),
			InstanceView: &armcompute.VirtualMachineScaleSetVMInstanceView{
				Extensions: []*armcompute.VirtualMachineExtensionInstanceView{
					{
						Statuses: []*armcompute.InstanceViewStatus{
							{
								Level:   ptr.To(armcompute.StatusLevelTypesError),
								Message: errorMessage,
							},
						},
					},
				},
			},
		},
	}

	manager := newTestAzureManager(t)
	resourceLimiter := cloudprovider.NewResourceLimiter(
		map[string]int64{cloudprovider.ResourceNameCores: 1, cloudprovider.ResourceNameMemory: 10000000},
		map[string]int64{cloudprovider.ResourceNameCores: 10, cloudprovider.ResourceNameMemory: 100000000})
	provider, _ := BuildAzureCloudProvider(manager, resourceLimiter)
	manager.RegisterNodeGroup(
		newTestScaleSet(manager, "test-asg"))
	manager.explicitlyConfigured["test-asg"] = true
	scaleSet, _ := provider.NodeGroups()[0].(*ScaleSet)

	t.Run("getCSEErrorMessages test with CSE error in VM extensions", func(t *testing.T) {
		expectedCSEWErrorMessage := "Error Message Test"
		vmssVMs.Properties.InstanceView.Extensions[0].Name = ptr.To(vmssCSEExtensionName)
		actualCSEErrorMessage, actualCSEFailureBool := scaleSet.cseErrors(vmssVMs.Properties.InstanceView.Extensions)
		assert.True(t, actualCSEFailureBool)
		assert.Equal(t, []string{expectedCSEWErrorMessage}, actualCSEErrorMessage)
	})
	t.Run("getCSEErrorMessages test with no CSE error in VM extensions", func(t *testing.T) {
		vmssVMs.Properties.InstanceView.Extensions[0].Name = ptr.To("notCSEExtension")
		actualCSEErrorMessage, actualCSEFailureBool := scaleSet.cseErrors(vmssVMs.Properties.InstanceView.Extensions)
		assert.False(t, actualCSEFailureBool)
		assert.Equal(t, []string(nil), actualCSEErrorMessage)
	})
}

func newVMObjectWithState(provisioningState string, powerState string) *armcompute.VirtualMachineScaleSetVM {
	return &armcompute.VirtualMachineScaleSetVM{
		ID: ptr.To("1"), // Beware; refactor if needed
		Properties: &armcompute.VirtualMachineScaleSetVMProperties{
			ProvisioningState: ptr.To(provisioningState),
			InstanceView: &armcompute.VirtualMachineScaleSetVMInstanceView{
				Statuses: []*armcompute.InstanceViewStatus{
					{Code: ptr.To(powerState)},
				},
			},
		},
	}
}

// Suggestion: could populate all combinations, should reunify with TestInstanceStatusFromVM
func TestInstanceStatusFromProvisioningStateAndPowerState(t *testing.T) {
	t.Run("fast delete enablement = false", func(t *testing.T) {
		t.Run("provisioning state = failed, power state = starting", func(t *testing.T) {
			status := instanceStatusFromProvisioningStateAndPowerState("1", ptr.To(string(armcompute.GalleryProvisioningStateFailed)), vmPowerStateStarting, disableFastDeleteOnFailure)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})

		t.Run("provisioning state = failed, power state = running", func(t *testing.T) {
			status := instanceStatusFromProvisioningStateAndPowerState("1", ptr.To(string(armcompute.GalleryProvisioningStateFailed)), vmPowerStateRunning, disableFastDeleteOnFailure)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})

		t.Run("provisioning state = failed, power state = stopping", func(t *testing.T) {
			status := instanceStatusFromProvisioningStateAndPowerState("1", ptr.To(string(armcompute.GalleryProvisioningStateFailed)), vmPowerStateStopping, disableFastDeleteOnFailure)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})

		t.Run("provisioning state = failed, power state = stopped", func(t *testing.T) {

			status := instanceStatusFromProvisioningStateAndPowerState("1", ptr.To(string(armcompute.GalleryProvisioningStateFailed)), vmPowerStateStopped, disableFastDeleteOnFailure)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})

		t.Run("provisioning state = failed, power state = deallocated", func(t *testing.T) {
			status := instanceStatusFromProvisioningStateAndPowerState("1", ptr.To(string(armcompute.GalleryProvisioningStateFailed)), vmPowerStateDeallocated, disableFastDeleteOnFailure)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})

		t.Run("provisioning state = failed, power state = unknown", func(t *testing.T) {
			status := instanceStatusFromProvisioningStateAndPowerState("1", ptr.To(string(armcompute.GalleryProvisioningStateFailed)), vmPowerStateUnknown, disableFastDeleteOnFailure)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})
	})

	t.Run("fast delete enablement = true", func(t *testing.T) {
		t.Run("provisioning state = failed, power state = starting", func(t *testing.T) {
			status := instanceStatusFromProvisioningStateAndPowerState("1", ptr.To(string(armcompute.GalleryProvisioningStateFailed)), vmPowerStateStarting, enableFastDeleteOnFailure)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})

		t.Run("provisioning state = failed, power state = running", func(t *testing.T) {
			status := instanceStatusFromProvisioningStateAndPowerState("1", ptr.To(string(armcompute.GalleryProvisioningStateFailed)), vmPowerStateRunning, enableFastDeleteOnFailure)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceRunning, status.State)
		})

		t.Run("provisioning state = failed, power state = stopping", func(t *testing.T) {
			status := instanceStatusFromProvisioningStateAndPowerState("1", ptr.To(string(armcompute.GalleryProvisioningStateFailed)), vmPowerStateStopping, enableFastDeleteOnFailure)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceCreating, status.State)
			assert.NotNil(t, status.ErrorInfo)
		})

		t.Run("provisioning state = failed, power state = stopped", func(t *testing.T) {
			status := instanceStatusFromProvisioningStateAndPowerState("1", ptr.To(string(armcompute.GalleryProvisioningStateFailed)), vmPowerStateStopped, enableFastDeleteOnFailure)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceCreating, status.State)
			assert.NotNil(t, status.ErrorInfo)
		})

		t.Run("provisioning state = failed, power state = deallocated", func(t *testing.T) {
			status := instanceStatusFromProvisioningStateAndPowerState("1", ptr.To(string(armcompute.GalleryProvisioningStateFailed)), vmPowerStateDeallocated, enableFastDeleteOnFailure)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceCreating, status.State)
			assert.NotNil(t, status.ErrorInfo)
		})

		t.Run("provisioning state = failed, power state = unknown", func(t *testing.T) {
			status := instanceStatusFromProvisioningStateAndPowerState("1", ptr.To(string(armcompute.GalleryProvisioningStateFailed)), vmPowerStateUnknown, enableFastDeleteOnFailure)

			assert.NotNil(t, status)
			assert.Equal(t, cloudprovider.InstanceCreating, status.State)
			assert.NotNil(t, status.ErrorInfo)
		})
	})
}

// testPollingHandler implements runtime.PollingHandler[T] for testing.
// It returns the configured error on Poll() and reports done after the error is returned.
type testPollingHandler[T any] struct {
	err    error
	polled bool
}

func (f *testPollingHandler[T]) Done() bool {
	return f.polled && f.err == nil
}

func (f *testPollingHandler[T]) Poll(_ context.Context) (*http.Response, error) {
	f.polled = true
	return &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: http.NoBody}, f.err
}

func (f *testPollingHandler[T]) Result(_ context.Context, _ *T) error {
	return f.err
}

// newTestPollerWithError creates a runtime.Poller that returns the given error on PollUntilDone.
func newTestPollerWithError(err error) *runtime.Poller[armcompute.VirtualMachineScaleSetsClientDeleteInstancesResponse] {
	resp := &http.Response{
		StatusCode: http.StatusAccepted,
		Header:     http.Header{},
		Body:       http.NoBody,
	}
	pl := runtime.NewPipeline("test", "v0.0.0", runtime.PipelineOptions{}, nil)
	poller, pollerErr := runtime.NewPoller[armcompute.VirtualMachineScaleSetsClientDeleteInstancesResponse](resp, pl, &runtime.NewPollerOptions[armcompute.VirtualMachineScaleSetsClientDeleteInstancesResponse]{
		Handler: &testPollingHandler[armcompute.VirtualMachineScaleSetsClientDeleteInstancesResponse]{err: err},
	})
	if pollerErr != nil {
		panic(pollerErr)
	}
	return poller
}

func TestWaitForDeleteInstancesWithOperationPreemptedRetry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := newTestAzureManager(t)

	vmssName := "test-asg"
	var vmssCapacity int64 = 3
	orchMode := armcompute.OrchestrationModeUniform

	expectedScaleSets := []*armcompute.VirtualMachineScaleSet{
		{
			Name: &vmssName,
			SKU: &armcompute.SKU{
				Capacity: &vmssCapacity,
			},
			Properties: &armcompute.VirtualMachineScaleSetProperties{
				OrchestrationMode: &orchMode,
			},
		},
	}
	expectedVMSSVMs := newTestVMSSVMList(3)

	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

	mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, "test-asg").Return(expectedVMSSVMs, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient

	// Override the default delete client mock: expect exactly 1 retry call to BeginDeleteInstances
	mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
	mockDeleteClient.EXPECT().BeginDeleteInstances(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).Times(1)
	manager.azClient.vmssClientForDelete = mockDeleteClient

	err := manager.forceRefresh()
	assert.NoError(t, err)

	registered := manager.RegisterNodeGroup(newTestScaleSet(manager, "test-asg"))
	manager.explicitlyConfigured["test-asg"] = true
	assert.True(t, registered)
	err = manager.forceRefresh()
	assert.NoError(t, err)

	requiredIds := &armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs{
		InstanceIDs: []*string{ptr.To("0")},
	}

	scaleSet := newTestScaleSet(manager, "test-asg")

	// lastInstanceRefresh starts at zero value for a new ScaleSet
	assert.True(t, scaleSet.lastInstanceRefresh.IsZero())

	// Create a poller that returns an OperationPreempted error on PollUntilDone
	preemptedPoller := newTestPollerWithError(&azcore.ResponseError{ErrorCode: "OperationPreempted"})

	// Act: the initial PollUntilDone fails with OperationPreempted, triggering a retry.
	// The retry calls deleteInstances which returns (nil, nil) from the mock — retryPoller is nil,
	// so retryErr stays nil, which counts as success. The function should invalidate cache and return.
	scaleSet.waitForDeleteInstances(preemptedPoller, requiredIds)

	// Assert: BeginDeleteInstances was called exactly once (the retry). gomock.Times(1) enforces this.
	// Assert: instance cache was invalidated (lastInstanceRefresh changed from zero)
	assert.False(t, scaleSet.lastInstanceRefresh.IsZero(),
		"expected instance cache to be invalidated after preempted retry success")
}

func TestWaitForDeleteInstancesNoRetryOnOtherErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := newTestAzureManager(t)

	vmssName := "test-asg"
	var vmssCapacity int64 = 3
	orchMode := armcompute.OrchestrationModeUniform

	expectedScaleSets := []*armcompute.VirtualMachineScaleSet{
		{
			Name: &vmssName,
			SKU: &armcompute.SKU{
				Capacity: &vmssCapacity,
			},
			Properties: &armcompute.VirtualMachineScaleSetProperties{
				OrchestrationMode: &orchMode,
			},
		},
	}
	expectedVMSSVMs := newTestVMSSVMList(3)

	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

	mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, "test-asg").Return(expectedVMSSVMs, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient

	// Override the default delete client mock: expect exactly 0 calls to BeginDeleteInstances.
	// If the retry path is accidentally triggered, gomock will fail the test.
	mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
	mockDeleteClient.EXPECT().BeginDeleteInstances(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	manager.azClient.vmssClientForDelete = mockDeleteClient

	err := manager.forceRefresh()
	assert.NoError(t, err)

	registered := manager.RegisterNodeGroup(newTestScaleSet(manager, "test-asg"))
	manager.explicitlyConfigured["test-asg"] = true
	assert.True(t, registered)
	err = manager.forceRefresh()
	assert.NoError(t, err)

	requiredIds := &armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs{
		InstanceIDs: []*string{ptr.To("0")},
	}

	for _, strictCacheUpdates := range []bool{false, true} {
		t.Run(fmt.Sprintf("strict cache updates %t", strictCacheUpdates), func(t *testing.T) {
			manager.config.StrictCacheUpdates = strictCacheUpdates
			scaleSet := newTestScaleSet(manager, "test-asg")
			scaleSet.curSize = 1
			scaleSet.lastSizeRefresh = time.Now()
			scaleSet.sizeRefreshPeriod = time.Hour

			// Create a poller that returns a non-preempted error
			otherErrorPoller := newTestPollerWithError(errors.New("InternalServerError: something went wrong"))

			// Act: PollUntilDone fails with a non-preempted error — should NOT trigger retry
			scaleSet.waitForDeleteInstances(otherErrorPoller, requiredIds)

			// Assert: BeginDeleteInstances was never called. gomock.Times(0) enforces this.
			// Assert: caches were invalidated so TargetSize refreshes from VMSS capacity.
			assert.False(t, scaleSet.lastInstanceRefresh.IsZero())
			targetSize, err := scaleSet.TargetSize()
			assert.NoError(t, err)
			assert.Equal(t, 3, targetSize)
		})
	}
}

func TestWaitForDeleteInstancesFailureRefreshesAzureCacheInstanceState(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := newTestAzureManager(t)
	manager.config.StrictCacheUpdates = false

	vmssName := "test-asg"
	var vmssCapacity int64 = 3
	orchMode := armcompute.OrchestrationModeUniform

	expectedScaleSets := []*armcompute.VirtualMachineScaleSet{
		{
			Name: &vmssName,
			SKU: &armcompute.SKU{
				Capacity: &vmssCapacity,
			},
			Properties: &armcompute.VirtualMachineScaleSetProperties{
				OrchestrationMode: &orchMode,
			},
		},
	}
	expectedVMSSVMs := newTestVMSSVMList(3)

	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

	mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).Return(expectedVMSSVMs, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient

	scaleSet := newTestScaleSet(manager, vmssName)
	registered := manager.RegisterNodeGroup(scaleSet)
	manager.explicitlyConfigured[vmssName] = true
	assert.True(t, registered)
	assert.NoError(t, manager.forceRefresh())

	deletedNode := newApiNode(armcompute.OrchestrationModeUniform, 0)
	manager.azureCache.setInstanceStateByProviderID(deletedNode.Spec.ProviderID, cloudprovider.InstanceDeleting)

	// Simulate stale proactive deleting state before the delete operation result is known.
	hasInstance, err := manager.azureCache.HasInstance(deletedNode.Spec.ProviderID)
	assert.False(t, hasInstance)
	assert.NoError(t, err)

	requiredIds := &armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs{
		InstanceIDs: []*string{ptr.To("0")},
	}
	otherErrorPoller := newTestPollerWithError(errors.New("InternalServerError: something went wrong"))

	// A failed waiter should trigger a refresh and reconcile stale deleting state.
	scaleSet.waitForDeleteInstances(otherErrorPoller, requiredIds)

	hasInstance, err = manager.azureCache.HasInstance(deletedNode.Spec.ProviderID)
	assert.True(t, hasInstance)
	assert.NoError(t, err)
}

func TestWaitForDeleteInstancesRetryFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := newTestAzureManager(t)

	vmssName := "test-asg"
	var vmssCapacity int64 = 3
	orchMode := armcompute.OrchestrationModeUniform

	expectedScaleSets := []*armcompute.VirtualMachineScaleSet{
		{
			Name: &vmssName,
			SKU: &armcompute.SKU{
				Capacity: &vmssCapacity,
			},
			Properties: &armcompute.VirtualMachineScaleSetProperties{
				OrchestrationMode: &orchMode,
			},
		},
	}
	expectedVMSSVMs := newTestVMSSVMList(3)

	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

	mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, "test-asg").Return(expectedVMSSVMs, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient

	// Override the default delete client mock: the retry call to deleteInstances itself fails
	mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
	mockDeleteClient.EXPECT().BeginDeleteInstances(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, errors.New("ResourceGroupNotFound")).Times(1)
	manager.azClient.vmssClientForDelete = mockDeleteClient

	err := manager.forceRefresh()
	assert.NoError(t, err)

	registered := manager.RegisterNodeGroup(newTestScaleSet(manager, "test-asg"))
	manager.explicitlyConfigured["test-asg"] = true
	assert.True(t, registered)
	err = manager.forceRefresh()
	assert.NoError(t, err)

	requiredIds := &armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs{
		InstanceIDs: []*string{ptr.To("0")},
	}

	scaleSet := newTestScaleSet(manager, "test-asg")

	// lastInstanceRefresh starts at zero value for a new ScaleSet
	assert.True(t, scaleSet.lastInstanceRefresh.IsZero())

	// Create a poller that returns OperationPreempted
	preemptedPoller := newTestPollerWithError(&azcore.ResponseError{ErrorCode: "OperationPreempted"})

	// Act: PollUntilDone fails with OperationPreempted, retry is attempted but deleteInstances fails.
	// Should fall through to the error path and invalidate cache.
	scaleSet.waitForDeleteInstances(preemptedPoller, requiredIds)

	// Assert: retry was attempted exactly once (gomock.Times(1) enforces this)
	// Assert: cache was still invalidated despite failure (the !StrictCacheUpdates path)
	assert.False(t, scaleSet.lastInstanceRefresh.IsZero(),
		"expected instance cache to be invalidated after retry failure")
}

func TestScaleSetIncreaseSizeWithETag(t *testing.T) {
	t.Parallel()
	const cachedEtag = `W/"abc"`
	cases := map[string]struct {
		useEtag         bool
		cachedEtag      *string
		expectIfMatch   *string
		expectFinalSize int64
	}{
		"flag off: no IfMatch even with cached ETag": {
			useEtag:         false,
			cachedEtag:      ptr.To(cachedEtag),
			expectIfMatch:   nil,
			expectFinalSize: 4,
		},
		"flag on: IfMatch sent from cached ETag": {
			useEtag:         true,
			cachedEtag:      ptr.To(cachedEtag),
			expectIfMatch:   ptr.To(cachedEtag),
			expectFinalSize: 4,
		},
		"flag on but no cached ETag: no IfMatch": {
			useEtag:         true,
			cachedEtag:      nil,
			expectIfMatch:   nil,
			expectFinalSize: 4,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			manager := newTestAzureManager(t)
			manager.config.EnableVMSSEtag = tc.useEtag

			vmssName := "vmss-etag"
			orchMode := armcompute.OrchestrationModeUniform
			vmss := &armcompute.VirtualMachineScaleSet{
				Name: ptr.To(vmssName),
				SKU:  &armcompute.SKU{Capacity: ptr.To[int64](3)},
				Properties: &armcompute.VirtualMachineScaleSetProperties{
					OrchestrationMode: &orchMode,
				},
				Etag: tc.cachedEtag,
			}

			mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
			mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).
				Return([]*armcompute.VirtualMachineScaleSet{vmss}, nil).AnyTimes()
			manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

			mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
			mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).
				Return(newTestVMSSVMList(3), nil).AnyTimes()
			manager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient

			mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)
			mockVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).
				Return([]*armcompute.VirtualMachine{}, nil).AnyTimes()
			manager.azClient.virtualMachinesClient = mockVMClient

			var capturedOpts *armcompute.VirtualMachineScaleSetsClientBeginCreateOrUpdateOptions
			mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
			mockDeleteClient.EXPECT().
				BeginCreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, _ string, _ armcompute.VirtualMachineScaleSet,
					opts *armcompute.VirtualMachineScaleSetsClientBeginCreateOrUpdateOptions,
				) (*runtime.Poller[armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse], error) {
					capturedOpts = opts
					return nil, nil
				}).Times(1)
			mockDeleteClient.EXPECT().BeginDeleteInstances(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, nil).AnyTimes()
			manager.azClient.vmssClientForDelete = mockDeleteClient

			manager.explicitlyConfigured[vmssName] = true
			ss := newTestScaleSet(manager, vmssName)
			assert.True(t, manager.RegisterNodeGroup(ss))
			assert.NoError(t, manager.Refresh())

			provider, err := BuildAzureCloudProvider(manager, nil)
			assert.NoError(t, err)
			scaleSet := provider.NodeGroups()[0].(*ScaleSet)

			err = scaleSet.IncreaseSize(1)
			assert.NoError(t, err)

			if tc.expectIfMatch == nil {
				assert.True(t, capturedOpts == nil || capturedOpts.IfMatch == nil,
					"expected no IfMatch but got %v", capturedOpts)
			} else {
				if assert.NotNil(t, capturedOpts) && assert.NotNil(t, capturedOpts.IfMatch) {
					assert.Equal(t, *tc.expectIfMatch, *capturedOpts.IfMatch)
				}
			}

			assert.Equal(t, tc.expectFinalSize, scaleSet.curSize)
		})
	}
}

// TestScaleSetETagPreconditionFailureInvalidatesSizeCache verifies that when a
// 412 ETag precondition failure persists across the in-provider refresh-and-retry,
// the size cache is invalidated, so a subsequent TargetSize call picks up an
// out-of-band VMSS capacity change rather than returning the stale cached size.
func TestScaleSetETagPreconditionFailureInvalidatesSizeCache(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := newTestAzureManager(t)
	manager.config.EnableVMSSEtag = true

	vmssName := "vmss-etag-stale"
	orchMode := armcompute.OrchestrationModeUniform

	manager.azureCache.setScaleSet(vmssName, &armcompute.VirtualMachineScaleSet{
		Name: ptr.To(vmssName),
		SKU:  &armcompute.SKU{Capacity: ptr.To[int64](3)},
		Properties: &armcompute.VirtualMachineScaleSetProperties{
			OrchestrationMode: &orchMode,
		},
		Etag: ptr.To(`W/"old"`),
	})

	// The refresh GET returns a fresh ETag but the capacity is still below target,
	// so the retried PUT is attempted and rejected with another 412.
	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().Get(gomock.Any(), manager.config.ResourceGroup, vmssName, gomock.Any()).
		Return(&armcompute.VirtualMachineScaleSet{
			Name: ptr.To(vmssName),
			SKU:  &armcompute.SKU{Capacity: ptr.To[int64](3)},
			Properties: &armcompute.VirtualMachineScaleSetProperties{
				OrchestrationMode: &orchMode,
			},
			Etag: ptr.To(`W/"refreshed"`),
		}, nil).AnyTimes()
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).
		Return([]*armcompute.VirtualMachineScaleSet{}, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

	mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
	mockDeleteClient.EXPECT().
		BeginCreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, &azcore.ResponseError{StatusCode: http.StatusPreconditionFailed}).AnyTimes()
	manager.azClient.vmssClientForDelete = mockDeleteClient

	scaleSet := newTestScaleSet(manager, vmssName)
	scaleSet.sizeRefreshPeriod = manager.azureCache.refreshInterval

	err := scaleSet.IncreaseSize(1)
	assert.Error(t, err)

	// Simulate an out-of-band capacity change observed on the next cache fetch.
	manager.azureCache.setScaleSet(vmssName, &armcompute.VirtualMachineScaleSet{
		Name: ptr.To(vmssName),
		SKU:  &armcompute.SKU{Capacity: ptr.To[int64](5)},
		Properties: &armcompute.VirtualMachineScaleSetProperties{
			OrchestrationMode: &orchMode,
		},
		Etag: ptr.To(`W/"new"`),
	})

	target, err := scaleSet.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 5, target)
}

// TestScaleSetETagPreconditionFailureRollsBackCapacity verifies that when a capacity
// update is rejected (412) and the in-provider retry also fails, the eager capacity
// mutation on the cached VMSS object is rolled back, so TargetSize reports the prior
// size rather than the rejected desired size.
func TestScaleSetETagPreconditionFailureRollsBackCapacity(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := newTestAzureManager(t)
	manager.config.EnableVMSSEtag = true

	vmssName := "vmss-etag-rollback"
	orchMode := armcompute.OrchestrationModeUniform

	manager.azureCache.setScaleSet(vmssName, &armcompute.VirtualMachineScaleSet{
		Name: ptr.To(vmssName),
		SKU:  &armcompute.SKU{Capacity: ptr.To[int64](3)},
		Properties: &armcompute.VirtualMachineScaleSetProperties{
			OrchestrationMode: &orchMode,
		},
		Etag: ptr.To(`W/"old"`),
	})

	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().Get(gomock.Any(), manager.config.ResourceGroup, vmssName, gomock.Any()).
		Return(&armcompute.VirtualMachineScaleSet{
			Name: ptr.To(vmssName),
			SKU:  &armcompute.SKU{Capacity: ptr.To[int64](3)},
			Properties: &armcompute.VirtualMachineScaleSetProperties{
				OrchestrationMode: &orchMode,
			},
			Etag: ptr.To(`W/"refreshed"`),
		}, nil).AnyTimes()
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).
		Return([]*armcompute.VirtualMachineScaleSet{}, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

	mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
	mockDeleteClient.EXPECT().
		BeginCreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, &azcore.ResponseError{StatusCode: http.StatusPreconditionFailed}).AnyTimes()
	manager.azClient.vmssClientForDelete = mockDeleteClient

	scaleSet := newTestScaleSet(manager, vmssName)
	scaleSet.sizeRefreshPeriod = manager.azureCache.refreshInterval

	err := scaleSet.IncreaseSize(1)
	assert.Error(t, err)

	// Without the rejected mutation, TargetSize must still report the prior size (3),
	// not the desired size (4) that was never accepted.
	target, err := scaleSet.TargetSize()
	assert.NoError(t, err)
	assert.Equal(t, 3, target)
}

// TestScaleSetETagRetrySucceedsAfterPreconditionFailure verifies that a single 412
// is transparently recovered: the provider refreshes the ETag via GET and re-issues
// the capacity update once with the fresh If-Match, so IncreaseSize succeeds without
// surfacing an error (which would otherwise back off the node group).
func TestScaleSetETagRetrySucceedsAfterPreconditionFailure(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := newTestAzureManager(t)
	manager.config.EnableVMSSEtag = true

	vmssName := "vmss-etag-retry"
	orchMode := armcompute.OrchestrationModeUniform
	const staleEtag = `W/"stale"`
	const freshEtag = `W/"fresh"`

	manager.azureCache.setScaleSet(vmssName, &armcompute.VirtualMachineScaleSet{
		Name: ptr.To(vmssName),
		SKU:  &armcompute.SKU{Capacity: ptr.To[int64](3)},
		Properties: &armcompute.VirtualMachineScaleSetProperties{
			OrchestrationMode: &orchMode,
		},
		Etag: ptr.To(staleEtag),
	})

	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().Get(gomock.Any(), manager.config.ResourceGroup, vmssName, gomock.Any()).
		Return(&armcompute.VirtualMachineScaleSet{
			Name: ptr.To(vmssName),
			SKU:  &armcompute.SKU{Capacity: ptr.To[int64](3)},
			Properties: &armcompute.VirtualMachineScaleSetProperties{
				OrchestrationMode: &orchMode,
			},
			Etag: ptr.To(freshEtag),
		}, nil).Times(1)
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).
		Return([]*armcompute.VirtualMachineScaleSet{}, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

	var ifMatches []*string
	mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
	mockDeleteClient.EXPECT().
		BeginCreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, _ string, _ armcompute.VirtualMachineScaleSet,
			opts *armcompute.VirtualMachineScaleSetsClientBeginCreateOrUpdateOptions,
		) (*runtime.Poller[armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse], error) {
			ifMatches = append(ifMatches, opts.IfMatch)
			if len(ifMatches) == 1 {
				return nil, &azcore.ResponseError{StatusCode: http.StatusPreconditionFailed}
			}
			return newTestCreateOrUpdatePoller(t, ptr.To(freshEtag)), nil
		}).Times(2)
	manager.azClient.vmssClientForDelete = mockDeleteClient

	scaleSet := newTestScaleSet(manager, vmssName)
	scaleSet.sizeRefreshPeriod = manager.azureCache.refreshInterval

	err := scaleSet.IncreaseSize(1)
	assert.NoError(t, err)

	if assert.Len(t, ifMatches, 2) {
		// First PUT carries the stale cached ETag and is rejected.
		if assert.NotNil(t, ifMatches[0]) {
			assert.Equal(t, staleEtag, *ifMatches[0])
		}
		// The retried PUT carries the freshly-fetched ETag.
		if assert.NotNil(t, ifMatches[1]) {
			assert.Equal(t, freshEtag, *ifMatches[1])
		}
	}
}

// TestScaleSetETagRetrySkippedWhenAlreadyAtTarget verifies that if the refresh GET
// after a 412 shows the VMSS already at (or above) the desired capacity, the provider
// treats the scale-up as satisfied and does not issue a second PUT that would shrink it.
func TestScaleSetETagRetrySkippedWhenAlreadyAtTarget(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := newTestAzureManager(t)
	manager.config.EnableVMSSEtag = true

	vmssName := "vmss-etag-skip"
	orchMode := armcompute.OrchestrationModeUniform

	manager.azureCache.setScaleSet(vmssName, &armcompute.VirtualMachineScaleSet{
		Name: ptr.To(vmssName),
		SKU:  &armcompute.SKU{Capacity: ptr.To[int64](3)},
		Properties: &armcompute.VirtualMachineScaleSetProperties{
			OrchestrationMode: &orchMode,
		},
		Etag: ptr.To(`W/"stale"`),
	})

	// A concurrent writer already grew the VMSS to the desired size (4).
	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().Get(gomock.Any(), manager.config.ResourceGroup, vmssName, gomock.Any()).
		Return(&armcompute.VirtualMachineScaleSet{
			Name: ptr.To(vmssName),
			SKU:  &armcompute.SKU{Capacity: ptr.To[int64](4)},
			Properties: &armcompute.VirtualMachineScaleSetProperties{
				OrchestrationMode: &orchMode,
			},
			Etag: ptr.To(`W/"fresh"`),
		}, nil).Times(1)
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).
		Return([]*armcompute.VirtualMachineScaleSet{}, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

	var putCalls int
	mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
	mockDeleteClient.EXPECT().
		BeginCreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, _ string, _ armcompute.VirtualMachineScaleSet,
			_ *armcompute.VirtualMachineScaleSetsClientBeginCreateOrUpdateOptions,
		) (*runtime.Poller[armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse], error) {
			putCalls++
			return nil, &azcore.ResponseError{StatusCode: http.StatusPreconditionFailed}
		}).Times(1)
	manager.azClient.vmssClientForDelete = mockDeleteClient

	scaleSet := newTestScaleSet(manager, vmssName)
	scaleSet.sizeRefreshPeriod = manager.azureCache.refreshInterval

	err := scaleSet.IncreaseSize(1)
	assert.NoError(t, err)
	assert.Equal(t, 1, putCalls, "expected no second PUT when VMSS already at target")
}

// TestScaleSetETagReconcilesConcurrentWriter exercises the core scenario ETag mode is
// meant to protect: another writer mutates the VMSS between CA's read and its write.
// Using a mock that enforces optimistic concurrency like the real API, CA's first PUT
// carries its stale If-Match and is rejected (412) rather than clobbering the concurrent
// change; CA then refreshes the ETag via GET and re-issues the PUT once with the
// up-to-date If-Match, so the scale-up still lands on top of the concurrent change.
func TestScaleSetETagReconcilesConcurrentWriter(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := newTestAzureManager(t)
	manager.config.EnableVMSSEtag = true

	vmssName := "vmss-etag-concurrent"
	orchMode := armcompute.OrchestrationModeUniform

	// CA's cached view: ETag E1, capacity 3.
	const caCachedEtag = `W/"E1"`
	manager.azureCache.setScaleSet(vmssName, &armcompute.VirtualMachineScaleSet{
		Name:       ptr.To(vmssName),
		SKU:        &armcompute.SKU{Capacity: ptr.To[int64](3)},
		Properties: &armcompute.VirtualMachineScaleSetProperties{OrchestrationMode: &orchMode},
		Etag:       ptr.To(caCachedEtag),
	})

	// Simulated server state: a concurrent writer already advanced the ETag to E2
	// (e.g. it retagged the VMSS) while leaving capacity at 3.
	var (
		mu             sync.Mutex
		serverEtag     = `W/"E2"`
		serverCapacity = int64(3)
	)

	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().Get(gomock.Any(), manager.config.ResourceGroup, vmssName, gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, _ string, _ *armcompute.ExpandTypesForGetVMScaleSets,
		) (*armcompute.VirtualMachineScaleSet, error) {
			mu.Lock()
			defer mu.Unlock()
			return &armcompute.VirtualMachineScaleSet{
				Name:       ptr.To(vmssName),
				SKU:        &armcompute.SKU{Capacity: ptr.To(serverCapacity)},
				Properties: &armcompute.VirtualMachineScaleSetProperties{OrchestrationMode: &orchMode},
				Etag:       ptr.To(serverEtag),
			}, nil
		}).Times(1)
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).
		Return([]*armcompute.VirtualMachineScaleSet{}, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

	var ifMatches []*string
	mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
	mockDeleteClient.EXPECT().
		BeginCreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, _ string, op armcompute.VirtualMachineScaleSet,
			opts *armcompute.VirtualMachineScaleSetsClientBeginCreateOrUpdateOptions,
		) (*runtime.Poller[armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse], error) {
			mu.Lock()
			defer mu.Unlock()
			ifMatches = append(ifMatches, opts.IfMatch)
			// Enforce optimistic concurrency like the real API: reject a stale If-Match.
			if opts.IfMatch != nil && *opts.IfMatch != serverEtag {
				return nil, &azcore.ResponseError{StatusCode: http.StatusPreconditionFailed}
			}
			// Accept the write: advance capacity and roll the ETag forward.
			serverCapacity = *op.SKU.Capacity
			serverEtag = `W/"E3"`
			return newTestCreateOrUpdatePoller(t, ptr.To(serverEtag)), nil
		}).Times(2)
	manager.azClient.vmssClientForDelete = mockDeleteClient

	scaleSet := newTestScaleSet(manager, vmssName)
	scaleSet.sizeRefreshPeriod = manager.azureCache.refreshInterval

	err := scaleSet.IncreaseSize(1)
	assert.NoError(t, err)

	if assert.Len(t, ifMatches, 2) {
		// First write used CA's stale ETag and was rejected: the concurrent change is not overwritten.
		if assert.NotNil(t, ifMatches[0]) {
			assert.Equal(t, caCachedEtag, *ifMatches[0])
		}
		// Second write used the refreshed server ETag and was accepted.
		if assert.NotNil(t, ifMatches[1]) {
			assert.Equal(t, `W/"E2"`, *ifMatches[1])
		}
	}

	mu.Lock()
	assert.Equal(t, int64(4), serverCapacity, "scale-up should land after reconciling the concurrent writer")
	mu.Unlock()
}

// TestScaleSetETagRetryTracksRefreshedObject verifies that after an ETag retry the
// object the caller tracks (and that the async completion updates) is the freshly
// fetched VMSS now living in the cache, not the stale pre-retry copy. Without this the
// final LRO ETag and capacity would land on a detached object, leaving the cache with a
// stale ETag (forcing a needless 412 on the next PUT) and a stale TargetSize.
func TestScaleSetETagRetryTracksRefreshedObject(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := newTestAzureManager(t)
	manager.config.EnableVMSSEtag = true

	vmssName := "vmss-etag-track"
	orchMode := armcompute.OrchestrationModeUniform
	const (
		staleEtag = `W/"E1"` // CA's cached ETag
		freshEtag = `W/"E2"` // server ETag returned by the refresh GET
		finalEtag = `W/"E3"` // ETag returned by the successful re-PUT (LRO result)
	)

	staleObj := &armcompute.VirtualMachineScaleSet{
		Name:       ptr.To(vmssName),
		SKU:        &armcompute.SKU{Capacity: ptr.To[int64](3)},
		Properties: &armcompute.VirtualMachineScaleSetProperties{OrchestrationMode: &orchMode},
		Etag:       ptr.To(staleEtag),
	}
	manager.azureCache.setScaleSet(vmssName, staleObj)

	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().Get(gomock.Any(), manager.config.ResourceGroup, vmssName, gomock.Any()).
		Return(&armcompute.VirtualMachineScaleSet{
			Name:       ptr.To(vmssName),
			SKU:        &armcompute.SKU{Capacity: ptr.To[int64](3)},
			Properties: &armcompute.VirtualMachineScaleSetProperties{OrchestrationMode: &orchMode},
			Etag:       ptr.To(freshEtag),
		}, nil).Times(1)
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).
		Return([]*armcompute.VirtualMachineScaleSet{}, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

	var puts int
	mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
	mockDeleteClient.EXPECT().
		BeginCreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, _ string, _ armcompute.VirtualMachineScaleSet,
			_ *armcompute.VirtualMachineScaleSetsClientBeginCreateOrUpdateOptions,
		) (*runtime.Poller[armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse], error) {
			puts++
			if puts == 1 {
				return nil, &azcore.ResponseError{StatusCode: http.StatusPreconditionFailed}
			}
			return newTestCreateOrUpdatePoller(t, ptr.To(finalEtag)), nil
		}).Times(2)
	manager.azClient.vmssClientForDelete = mockDeleteClient

	scaleSet := newTestScaleSet(manager, vmssName)
	scaleSet.sizeRefreshPeriod = manager.azureCache.refreshInterval

	ctx, cancel := getContextWithTimeout(asyncContextTimeout)
	defer cancel()

	effectiveVMSS, poller, err := scaleSet.initCreateOrUpdate(ctx, staleObj, 4)
	assert.NoError(t, err)
	assert.NotNil(t, poller)

	// The tracked object must be the freshly-fetched VMSS now in the cache, not the
	// stale pre-retry copy.
	cached := manager.azureCache.getScaleSets()[vmssName]
	assert.Same(t, cached, effectiveVMSS, "caller should track the refreshed cached object")
	assert.NotSame(t, staleObj, effectiveVMSS, "caller must not keep tracking the stale object")
	if assert.NotNil(t, effectiveVMSS.Etag) {
		assert.Equal(t, freshEtag, *effectiveVMSS.Etag)
	}

	// Drive completion synchronously so the final ETag deterministically lands on the
	// cached object.
	scaleSet.waitForCreateOrUpdateInstances(poller, effectiveVMSS)

	cached = manager.azureCache.getScaleSets()[vmssName]
	if assert.NotNil(t, cached.Etag) {
		assert.Equal(t, finalEtag, *cached.Etag, "final LRO ETag should land on the cached object")
	}
}

// TestScaleSetETagRetrySkipPublishesFreshCapacity verifies that when the refresh GET
// after a 412 shows a concurrent writer scaled the VMSS ABOVE our target, the skip
// publishes the actual (higher) capacity rather than our stale, lower target.
func TestScaleSetETagRetrySkipPublishesFreshCapacity(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := newTestAzureManager(t)
	manager.config.EnableVMSSEtag = true

	vmssName := "vmss-etag-skip-higher"
	orchMode := armcompute.OrchestrationModeUniform

	manager.azureCache.setScaleSet(vmssName, &armcompute.VirtualMachineScaleSet{
		Name:       ptr.To(vmssName),
		SKU:        &armcompute.SKU{Capacity: ptr.To[int64](3)},
		Properties: &armcompute.VirtualMachineScaleSetProperties{OrchestrationMode: &orchMode},
		Etag:       ptr.To(`W/"stale"`),
	})

	// A concurrent writer grew the VMSS to 5, above our desired target of 4.
	mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().Get(gomock.Any(), manager.config.ResourceGroup, vmssName, gomock.Any()).
		Return(&armcompute.VirtualMachineScaleSet{
			Name:       ptr.To(vmssName),
			SKU:        &armcompute.SKU{Capacity: ptr.To[int64](5)},
			Properties: &armcompute.VirtualMachineScaleSetProperties{OrchestrationMode: &orchMode},
			Etag:       ptr.To(`W/"fresh"`),
		}, nil).Times(1)
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).
		Return([]*armcompute.VirtualMachineScaleSet{}, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

	var putCalls int
	mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
	mockDeleteClient.EXPECT().
		BeginCreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, _ string, _ armcompute.VirtualMachineScaleSet,
			_ *armcompute.VirtualMachineScaleSetsClientBeginCreateOrUpdateOptions,
		) (*runtime.Poller[armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse], error) {
			putCalls++
			return nil, &azcore.ResponseError{StatusCode: http.StatusPreconditionFailed}
		}).Times(1)
	manager.azClient.vmssClientForDelete = mockDeleteClient

	scaleSet := newTestScaleSet(manager, vmssName)
	scaleSet.sizeRefreshPeriod = manager.azureCache.refreshInterval

	err := scaleSet.IncreaseSize(1)
	assert.NoError(t, err)
	assert.Equal(t, 1, putCalls, "expected no second PUT when VMSS already above target")

	scaleSet.sizeMutex.Lock()
	curSize := scaleSet.curSize
	scaleSet.sizeMutex.Unlock()
	assert.Equal(t, int64(5), curSize, "skip should publish the concurrent writer's actual capacity")
}

func TestWaitForCreateOrUpdateInstancesRefreshesETag(t *testing.T) {
	t.Parallel()
	const oldEtag = `W/"old"`
	const newEtag = `W/"new"`
	cases := map[string]struct {
		useEtag  bool
		wantEtag *string
	}{
		"flag on: adopts new ETag from completed operation": {useEtag: true, wantEtag: ptr.To(newEtag)},
		"flag off: leaves cached ETag untouched":            {useEtag: false, wantEtag: ptr.To(oldEtag)},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			manager := newTestAzureManager(t)
			manager.config.EnableVMSSEtag = tc.useEtag
			ss := newTestScaleSet(manager, "vmss-etag")

			vmssInfo := &armcompute.VirtualMachineScaleSet{
				Name: ptr.To("vmss-etag"),
				SKU:  &armcompute.SKU{Capacity: ptr.To[int64](3)},
				Etag: ptr.To(oldEtag),
			}

			ss.waitForCreateOrUpdateInstances(newTestCreateOrUpdatePoller(t, ptr.To(newEtag)), vmssInfo)

			if assert.NotNil(t, vmssInfo.Etag) {
				assert.Equal(t, *tc.wantEtag, *vmssInfo.Etag)
			}
		})
	}
}

func TestAtomicIncreaseSizeWithETag(t *testing.T) {
	t.Parallel()
	const cachedEtag = `W/"abc"`
	const newEtag = `W/"def"`
	cases := map[string]struct {
		useEtag         bool
		expectIfMatch   *string
		expectFinalSize int64
		expectEtag      *string
	}{
		"flag off: no IfMatch, ETag untouched": {
			useEtag:         false,
			expectIfMatch:   nil,
			expectFinalSize: 4,
			expectEtag:      ptr.To(cachedEtag),
		},
		"flag on: IfMatch sent and new ETag adopted on success": {
			useEtag:         true,
			expectIfMatch:   ptr.To(cachedEtag),
			expectFinalSize: 4,
			expectEtag:      ptr.To(newEtag),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			manager := newTestAzureManager(t)
			manager.config.EnableVMSSEtag = tc.useEtag

			vmssName := "vmss-etag-atomic"
			orchMode := armcompute.OrchestrationModeUniform
			vmss := &armcompute.VirtualMachineScaleSet{
				Name: ptr.To(vmssName),
				SKU:  &armcompute.SKU{Capacity: ptr.To[int64](3)},
				Properties: &armcompute.VirtualMachineScaleSetProperties{
					OrchestrationMode: &orchMode,
				},
				Etag: ptr.To(cachedEtag),
			}

			mockVMSSClient := mock_virtualmachinescalesetclient.NewMockInterface(ctrl)
			mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).
				Return([]*armcompute.VirtualMachineScaleSet{vmss}, nil).AnyTimes()
			manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

			mockVMSSVMClient := mock_virtualmachinescalesetvmclient.NewMockInterface(ctrl)
			mockVMSSVMClient.EXPECT().ListVMInstanceView(gomock.Any(), manager.config.ResourceGroup, vmssName).
				Return(newTestVMSSVMList(3), nil).AnyTimes()
			manager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient

			mockVMClient := mock_virtualmachineclient.NewMockInterface(ctrl)
			mockVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).
				Return([]*armcompute.VirtualMachine{}, nil).AnyTimes()
			manager.azClient.virtualMachinesClient = mockVMClient

			var capturedOpts *armcompute.VirtualMachineScaleSetsClientBeginCreateOrUpdateOptions
			mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
			mockDeleteClient.EXPECT().
				BeginCreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, _ string, _ string, _ armcompute.VirtualMachineScaleSet,
					opts *armcompute.VirtualMachineScaleSetsClientBeginCreateOrUpdateOptions,
				) (*runtime.Poller[armcompute.VirtualMachineScaleSetsClientCreateOrUpdateResponse], error) {
					capturedOpts = opts
					return newTestCreateOrUpdatePoller(t, ptr.To(newEtag)), nil
				}).Times(1)
			manager.azClient.vmssClientForDelete = mockDeleteClient

			manager.explicitlyConfigured[vmssName] = true
			ss := newTestScaleSet(manager, vmssName)
			assert.True(t, manager.RegisterNodeGroup(ss))
			assert.NoError(t, manager.Refresh())

			provider, err := BuildAzureCloudProvider(manager, nil)
			assert.NoError(t, err)
			scaleSet := provider.NodeGroups()[0].(*ScaleSet)

			err = scaleSet.AtomicIncreaseSize(1)
			assert.NoError(t, err)

			if tc.expectIfMatch == nil {
				assert.True(t, capturedOpts == nil || capturedOpts.IfMatch == nil,
					"expected no IfMatch but got %v", capturedOpts)
			} else {
				if assert.NotNil(t, capturedOpts) && assert.NotNil(t, capturedOpts.IfMatch) {
					assert.Equal(t, *tc.expectIfMatch, *capturedOpts.IfMatch)
				}
			}

			assert.Equal(t, tc.expectFinalSize, scaleSet.curSize)

			cached := manager.azureCache.getScaleSets()[vmssName]
			if assert.NotNil(t, cached) && assert.NotNil(t, cached.Etag) {
				assert.Equal(t, *tc.expectEtag, *cached.Etag)
			}
		})
	}
}

// TestETagConcurrentAccessNoDataRace overlaps the ETag reader (initCreateOrUpdate)
// with the ETag writer (waitForCreateOrUpdateInstances) on the same shared cached
// VMSS object. It guards against regressing the ETag read back to sizeMutex (the
// writer uses vmssSizeMutex); run under -race to detect a cross-lock data race.
func TestETagConcurrentAccessNoDataRace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := newTestAzureManager(t)
	manager.config.EnableVMSSEtag = true
	ss := newTestScaleSet(manager, "vmss-etag-race")

	vmssInfo := &armcompute.VirtualMachineScaleSet{
		Name:     ptr.To("vmss-etag-race"),
		Location: ptr.To(testLocation),
		SKU:      &armcompute.SKU{Name: ptr.To("Standard_D2_v2"), Capacity: ptr.To[int64](3)},
		Etag:     ptr.To(`W/"r0"`),
	}

	mockDeleteClient := NewMockVMSSDeleteClient(ctrl)
	mockDeleteClient.EXPECT().
		BeginCreateOrUpdate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil).AnyTimes()
	manager.azClient.vmssClientForDelete = mockDeleteClient

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			ctx, cancel := getContextWithTimeout(vmssContextTimeout)
			defer cancel()
			_, _, _ = ss.initCreateOrUpdate(ctx, vmssInfo, 4)
		}()
		go func(n int) {
			defer wg.Done()
			ss.waitForCreateOrUpdateInstances(newTestCreateOrUpdatePoller(t, ptr.To(fmt.Sprintf(`W/"r%d"`, n))), vmssInfo)
		}(i)
	}
	wg.Wait()
}
