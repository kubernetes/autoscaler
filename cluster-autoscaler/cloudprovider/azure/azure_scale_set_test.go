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
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-03-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmclient/mockvmclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmssclient/mockvmssclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmssvmclient/mockvmssvmclient"
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

func newTestVMSSList(cap int64, name, loc string, orchmode compute.OrchestrationMode) []compute.VirtualMachineScaleSet {
	return []compute.VirtualMachineScaleSet{
		{
			Name: to.StringPtr(name),
			Sku: &compute.Sku{
				Capacity: to.Int64Ptr(cap),
				Name:     to.StringPtr("Standard_D4_v2"),
			},
			VirtualMachineScaleSetProperties: &compute.VirtualMachineScaleSetProperties{
				OrchestrationMode: orchmode,
			},
			Location: to.StringPtr(loc),
			ID:       to.StringPtr(name),
		},
	}
}

func newTestVMSSListForEdgeZones(capacity int64, name string) *compute.VirtualMachineScaleSet {
	return &compute.VirtualMachineScaleSet{
		Name: to.StringPtr(name),
		Sku: &compute.Sku{
			Capacity: to.Int64Ptr(capacity),
			Name:     to.StringPtr("Standard_D4_v2"),
		},
		VirtualMachineScaleSetProperties: &compute.VirtualMachineScaleSetProperties{},
		Location:                         to.StringPtr(testLocation),
		ExtendedLocation: &compute.ExtendedLocation{
			Name: to.StringPtr("losangeles"),
			Type: compute.ExtendedLocationTypes("EdgeZone"),
		},
	}
}

func newTestVMSSVMList(count int) []compute.VirtualMachineScaleSetVM {
	var vmssVMList []compute.VirtualMachineScaleSetVM
	for i := 0; i < count; i++ {
		vmssVM := compute.VirtualMachineScaleSetVM{
			ID:         to.StringPtr(fmt.Sprintf(fakeVirtualMachineScaleSetVMID, i)),
			InstanceID: to.StringPtr(fmt.Sprintf("%d", i)),
			VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
				VMID: to.StringPtr(fmt.Sprintf("123E4567-E89B-12D3-A456-426655440000-%d", i)),
			},
		}
		vmssVMList = append(vmssVMList, vmssVM)
	}
	return vmssVMList
}

func newTestVMList(count int) []compute.VirtualMachine {
	var vmssVMList []compute.VirtualMachine
	for i := 0; i < count; i++ {
		vmssVM := compute.VirtualMachine{
			ID: to.StringPtr(fmt.Sprintf(fakeVirtualMachineVMID, i)),
			VirtualMachineProperties: &compute.VirtualMachineProperties{
				VMID: to.StringPtr(fmt.Sprintf("123E4567-E89B-12D3-A456-426655440000-%d", i)),
			},
		}
		vmssVMList = append(vmssVMList, vmssVM)
	}
	return vmssVMList
}

func newApiNode(orchmode compute.OrchestrationMode, vmID int64) *apiv1.Node {
	providerId := fakeVirtualMachineScaleSetVMID

	if orchmode == compute.Flexible {
		providerId = fakeVirtualMachineVMID
	}

	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: azurePrefix + fmt.Sprintf(providerId, vmID),
		},
	}
	return node
}
func TestMaxSize(t *testing.T) {
	provider := newTestProvider(t)
	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)
	assert.Equal(t, provider.NodeGroups()[0].MaxSize(), 5)
}

func TestMinSize(t *testing.T) {
	provider := newTestProvider(t)
	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)
	assert.Equal(t, provider.NodeGroups()[0].MinSize(), 1)
}

func TestMinSizeZero(t *testing.T) {
	provider := newTestProvider(t)
	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSetMinSizeZero(provider.azureManager, testASG))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)
	assert.Equal(t, provider.NodeGroups()[0].MinSize(), 0)
}

func TestTargetSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orchestrationModes := [2]compute.OrchestrationMode{compute.Uniform, compute.Flexible}
	expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus", compute.Uniform)
	spotScaleSet := newTestVMSSList(5, "spot-vmss", "eastus", compute.Uniform)[0]
	spotScaleSet.VirtualMachineProfile = &compute.VirtualMachineScaleSetVMProfile{Priority: compute.Spot}
	expectedScaleSets = append(expectedScaleSets, spotScaleSet)

	expectedVMSSVMs := newTestVMSSVMList(3)
	expectedVMs := newTestVMList(3)

	for _, orchMode := range orchestrationModes {
		provider := newTestProvider(t)
		mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
		mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
		provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
		mockVMClient := mockvmclient.NewMockInterface(ctrl)
		mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return([]compute.VirtualMachine{}, nil).AnyTimes()
		provider.azureManager.azClient.virtualMachinesClient = mockVMClient

		// return a different capacity from GET API
		spotScaleSet.Sku.Capacity = to.Int64Ptr(1)
		mockVMSSClient.EXPECT().Get(gomock.Any(), provider.azureManager.config.ResourceGroup, "spot-vmss").Return(spotScaleSet, nil).Times(1)
		provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
		mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)

		mockVMSSVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup, "test-asg", gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()
		provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
		err := provider.azureManager.forceRefresh()
		assert.NoError(t, err)

		if orchMode == compute.Uniform {
			mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)
			mockVMSSVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup, "test-asg", gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()
			provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
		} else {
			provider.azureManager.config.EnableVmssFlexNodes = true
			mockVMClient.EXPECT().ListVmssFlexVMsWithoutInstanceView(gomock.Any(), "test-asg").Return(expectedVMs, nil).AnyTimes()
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

func TestIncreaseSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orchestrationModes := [2]compute.OrchestrationMode{compute.Uniform, compute.Flexible}

	for _, orchMode := range orchestrationModes {

		expectedScaleSets := newTestVMSSList(3, testASG, "eastus", orchMode)
		expectedVMSSVMs := newTestVMSSVMList(3)
		expectedVMs := newTestVMList(3)

		// Include Edge Zone scenario here, testing scale from 3 to 5 and scale from zero cases.
		expectedEdgeZoneScaleSets := newTestVMSSListForEdgeZones(3, "edgezone-vmss")
		expectedEdgeZoneMinZeroScaleSets := newTestVMSSListForEdgeZones(0, "edgezone-minzero-vmss")
		expectedScaleSets = append(expectedScaleSets, *expectedEdgeZoneScaleSets, *expectedEdgeZoneMinZeroScaleSets)

		provider := newTestProvider(t)
		// expectedScaleSets := newTestVMSSList(3, testASG, testLocation, orchMode)

		// // Include Edge Zone scenario here, testing scale from 3 to 5 and scale from zero cases.
		// expectedEdgeZoneScaleSets := newTestVMSSListForEdgeZones(3, "edgezone-vmss")
		// expectedEdgeZoneMinZeroScaleSets := newTestVMSSListForEdgeZones(0, "edgezone-minzero-vmss")
		// expectedScaleSets = append(expectedScaleSets, *expectedEdgeZoneScaleSets, *expectedEdgeZoneMinZeroScaleSets)

		mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
		mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
		mockVMSSClient.EXPECT().CreateOrUpdateAsync(gomock.Any(), provider.azureManager.config.ResourceGroup, testASG, gomock.Any()).Return(nil, nil)
		// This should be Anytimes() because the parent function of this call - updateVMSSCapacity() is a goroutine
		// and this test doesn't wait on goroutine, hence, it is difficult to write exact expected number (which is 3 here)
		// before we return from this this.
		// This is a future TODO: sync.WaitGroup should be used in actual code and make code easily testable
		mockVMSSClient.EXPECT().WaitForCreateOrUpdateResult(gomock.Any(), gomock.Any(), provider.azureManager.config.ResourceGroup).Return(&http.Response{StatusCode: http.StatusOK}, nil).AnyTimes()
		provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
		mockVMClient := mockvmclient.NewMockInterface(ctrl)
		mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return([]compute.VirtualMachine{}, nil).AnyTimes()
		provider.azureManager.azClient.virtualMachinesClient = mockVMClient

		if orchMode == compute.Uniform {
			mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)
			mockVMSSVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup, "test-asg", gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()
			provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
		} else {
			provider.azureManager.config.EnableVmssFlexNodes = true
			mockVMClient.EXPECT().ListVmssFlexVMsWithoutInstanceView(gomock.Any(), "test-asg").Return(expectedVMs, nil).AnyTimes()
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

		mockVMSSClient.EXPECT().CreateOrUpdateAsync(gomock.Any(), provider.azureManager.config.ResourceGroup,
			"edgezone-vmss", gomock.Any()).Return(nil, nil)
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

		mockVMSSClient.EXPECT().CreateOrUpdateAsync(gomock.Any(), provider.azureManager.config.ResourceGroup,
			"edgezone-minzero-vmss", gomock.Any()).Return(nil, nil)
		err = provider.NodeGroups()[2].IncreaseSize(2)
		assert.NoError(t, err)

		// New target size should be 2.
		targetSizeForEdgeZoneMinZero, err = provider.NodeGroups()[2].TargetSize()
		assert.NoError(t, err)
		assert.Equal(t, 2, targetSizeForEdgeZoneMinZero)
	}
}

func TestIncreaseSizeOnVMSSUpdating(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager := newTestAzureManager(t)
	vmssName := "vmss-updating"
	var vmssCapacity int64 = 3

	expectedScaleSets := []compute.VirtualMachineScaleSet{
		{
			Name: &vmssName,
			Sku: &compute.Sku{
				Capacity: &vmssCapacity,
			},
			VirtualMachineScaleSetProperties: &compute.VirtualMachineScaleSetProperties{
				ProvisioningState: to.StringPtr(string(compute.ProvisioningStateUpdating)),
				OrchestrationMode: compute.Uniform,
			},
		},
	}
	expectedVMSSVMs := newTestVMSSVMList(3)

	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil)
	mockVMSSClient.EXPECT().CreateOrUpdateAsync(gomock.Any(), manager.config.ResourceGroup, vmssName, gomock.Any()).Return(
		nil, nil)
	mockVMSSClient.EXPECT().WaitForCreateOrUpdateResult(gomock.Any(), gomock.Any(), manager.config.ResourceGroup).Return(
		&http.Response{StatusCode: http.StatusOK}, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
	mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, "vmss-updating",
		gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()
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

func TestBelongs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orchestrationModes := [2]compute.OrchestrationMode{compute.Uniform, compute.Flexible}
	expectedVMSSVMs := newTestVMSSVMList(3)
	expectedVMs := newTestVMList(3)

	for _, orchMode := range orchestrationModes {
		expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus", orchMode)
		provider := newTestProvider(t)
		mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
		mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil)
		provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
		mockVMClient := mockvmclient.NewMockInterface(ctrl)
		mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return([]compute.VirtualMachine{}, nil).AnyTimes()
		provider.azureManager.azClient.virtualMachinesClient = mockVMClient

		if orchMode == compute.Uniform {
			mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)
			mockVMSSVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup, "test-asg", gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()
			provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
		} else {
			provider.azureManager.config.EnableVmssFlexNodes = true
			mockVMClient.EXPECT().ListVmssFlexVMsWithoutInstanceView(gomock.Any(), "test-asg").Return(expectedVMs, nil).AnyTimes()
		}

		registered := provider.azureManager.RegisterNodeGroup(
			newTestScaleSet(provider.azureManager, testASG))
		assert.True(t, registered)

		scaleSet, ok := provider.NodeGroups()[0].(*ScaleSet)
		assert.True(t, ok)
		provider.azureManager.explicitlyConfigured["test-asg"] = true
		err := provider.azureManager.Refresh()
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

func TestDeleteNodes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	vmssName := "test-asg"
	var vmssCapacity int64 = 3
	cases := []struct {
		name              string
		orchestrationMode compute.OrchestrationMode
		enableForceDelete bool
	}{
		{
			name:              "uniform, force delete enabled",
			orchestrationMode: compute.Uniform,
			enableForceDelete: true,
		},
		{
			name:              "uniform, force delete disabled",
			orchestrationMode: compute.Uniform,
			enableForceDelete: false,
		},
		{
			name:              "flexible, force delete enabled",
			orchestrationMode: compute.Flexible,
			enableForceDelete: true,
		},
		{
			name:              "flexible, force delete disabled",
			orchestrationMode: compute.Flexible,
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

		mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
		mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).Times(2)
		mockVMSSClient.EXPECT().DeleteInstancesAsync(gomock.Any(), manager.config.ResourceGroup, gomock.Any(), gomock.Any(), enableForceDelete).Return(nil, nil)
		mockVMSSClient.EXPECT().WaitForDeleteInstancesResult(gomock.Any(), gomock.Any(), manager.config.ResourceGroup).Return(&http.Response{StatusCode: http.StatusOK}, nil).AnyTimes()
		manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient

		mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)
		mockVMClient := mockvmclient.NewMockInterface(ctrl)
		manager.azClient.virtualMachinesClient = mockVMClient
		mockVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedVMs, nil).AnyTimes()

		if orchMode == compute.Uniform {
			mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, "test-asg", gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()
			manager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
		} else {
			manager.config.EnableVmssFlexNodes = true
			mockVMClient.EXPECT().ListVmssFlexVMsWithoutInstanceView(gomock.Any(), "test-asg").Return(expectedVMs, nil).AnyTimes()
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

		if orchMode == compute.Uniform {
			expectedVMSSVMs[0].ProvisioningState = to.StringPtr(provisioningStateDeleting)
			expectedVMSSVMs[2].ProvisioningState = to.StringPtr(provisioningStateDeleting)
			mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, "test-asg", gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()
		} else {
			expectedVMs[0].ProvisioningState = to.StringPtr(provisioningStateDeleting)
			expectedVMs[2].ProvisioningState = to.StringPtr(provisioningStateDeleting)
			mockVMClient.EXPECT().ListVmssFlexVMsWithoutInstanceView(gomock.Any(), "test-asg").Return(expectedVMs, nil).AnyTimes()
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

func TestDeleteNodeUnregistered(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	vmssName := "test-asg"
	var vmssCapacity int64 = 2

	cases := []struct {
		name              string
		orchestrationMode compute.OrchestrationMode
		enableForceDelete bool
	}{
		{
			name:              "uniform, force delete enabled",
			orchestrationMode: compute.Uniform,
			enableForceDelete: true,
		},
		{
			name:              "uniform, force delete disabled",
			orchestrationMode: compute.Uniform,
			enableForceDelete: false,
		},
		{
			name:              "flexible, force delete enabled",
			orchestrationMode: compute.Flexible,
			enableForceDelete: true,
		},
		{
			name:              "flexible, force delete disabled",
			orchestrationMode: compute.Flexible,
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

		mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
		mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).Times(2)
		mockVMSSClient.EXPECT().DeleteInstancesAsync(gomock.Any(), manager.config.ResourceGroup, gomock.Any(), gomock.Any(), enableForceDelete).Return(nil, nil)
		mockVMSSClient.EXPECT().WaitForDeleteInstancesResult(gomock.Any(), gomock.Any(), manager.config.ResourceGroup).Return(&http.Response{StatusCode: http.StatusOK}, nil).AnyTimes()
		manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
		mockVMClient := mockvmclient.NewMockInterface(ctrl)
		mockVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedVMs, nil).AnyTimes()
		manager.azClient.virtualMachinesClient = mockVMClient

		if orchMode == compute.Uniform {
			mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)
			mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, "test-asg", gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()
			manager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient
		} else {
			manager.config.EnableVmssFlexNodes = true
			mockVMClient.EXPECT().ListVmssFlexVMsWithoutInstanceView(gomock.Any(), "test-asg").Return(expectedVMs, nil).AnyTimes()
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

func TestDeleteInstancesWithForceDeleteEnabled(t *testing.T) {
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

	expectedScaleSets := []compute.VirtualMachineScaleSet{
		{
			Name: &vmssName,
			Sku: &compute.Sku{
				Capacity: &vmssCapacity,
			},
			VirtualMachineScaleSetProperties: &compute.VirtualMachineScaleSetProperties{
				OrchestrationMode: compute.Uniform,
			},
		},
	}
	expectedVMSSVMs := newTestVMSSVMList(3)

	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).Times(2)
	mockVMSSClient.EXPECT().DeleteInstancesAsync(gomock.Any(), manager.config.ResourceGroup, gomock.Any(), gomock.Any(), true).Return(nil, nil)
	mockVMSSClient.EXPECT().WaitForDeleteInstancesResult(gomock.Any(), gomock.Any(), manager.config.ResourceGroup).Return(&http.Response{StatusCode: http.StatusOK}, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
	mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, "test-asg", gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()
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
	expectedScaleSets = []compute.VirtualMachineScaleSet{
		{
			Name: &vmssName,
			Sku: &compute.Sku{
				Capacity: &vmssCapacity,
			},
			VirtualMachineScaleSetProperties: &compute.VirtualMachineScaleSetProperties{
				OrchestrationMode: compute.Uniform,
			},
		},
	}
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	expectedVMSSVMs[0].ProvisioningState = to.StringPtr(string(compute.ProvisioningStateDeleting))
	expectedVMSSVMs[2].ProvisioningState = to.StringPtr(string(compute.ProvisioningStateDeleting))
	mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, "test-asg", gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()
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

func TestDeleteNoConflictRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	vmssName := "test-asg"
	var vmssCapacity int64 = 3

	manager := newTestAzureManager(t)

	expectedVMSSVMs := []compute.VirtualMachineScaleSetVM{
		{
			ID:         to.StringPtr(fakeVirtualMachineScaleSetVMID),
			InstanceID: to.StringPtr("0"),
			VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
				VMID:              to.StringPtr("123E4567-E89B-12D3-A456-426655440000"),
				ProvisioningState: to.StringPtr("Deleting"),
			},
		},
	}
	expectedScaleSets := []compute.VirtualMachineScaleSet{
		{
			Name: &vmssName,
			Sku: &compute.Sku{
				Capacity: &vmssCapacity,
			},
			VirtualMachineScaleSetProperties: &compute.VirtualMachineScaleSetProperties{
				OrchestrationMode: compute.Uniform,
			},
		},
	}

	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	manager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
	mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)
	mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, "test-asg", gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()
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

func TestId(t *testing.T) {
	provider := newTestProvider(t)
	registered := provider.azureManager.RegisterNodeGroup(
		newTestScaleSet(provider.azureManager, "test-asg"))
	assert.True(t, registered)
	assert.Equal(t, len(provider.NodeGroups()), 1)
	assert.Equal(t, provider.NodeGroups()[0].Id(), "test-asg")
}

func TestDebug(t *testing.T) {
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
	orchestrationModes := [2]compute.OrchestrationMode{compute.Uniform, compute.Flexible}

	expectedVMSSVMs := newTestVMSSVMList(3)
	expectedVMs := newTestVMList(3)

	for _, orchMode := range orchestrationModes {
		expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus", orchMode)
		provider := newTestProvider(t)

		mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
		mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
		provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
		mockVMClient := mockvmclient.NewMockInterface(ctrl)
		mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return([]compute.VirtualMachine{}, nil).AnyTimes()
		provider.azureManager.azClient.virtualMachinesClient = mockVMClient

		if orchMode == compute.Uniform {
			mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)
			mockVMSSVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup, "test-asg", gomock.Any()).Return(expectedVMSSVMs, nil).AnyTimes()
			provider.azureManager.azClient.virtualMachineScaleSetVMsClient = mockVMSSVMClient

		} else {
			provider.azureManager.config.EnableVmssFlexNodes = true
			mockVMClient.EXPECT().ListVmssFlexVMsWithoutInstanceView(gomock.Any(), "test-asg").Return(expectedVMs, nil).AnyTimes()
		}

		registered := provider.azureManager.RegisterNodeGroup(
			newTestScaleSet(provider.azureManager, "test-asg"))
		provider.azureManager.explicitlyConfigured["test-asg"] = true
		err := provider.azureManager.Refresh()
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

		if orchMode == compute.Uniform {

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

func TestEnableVmssFlexNodesFlag(t *testing.T) {

	// flag set to false
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedVMs := newTestVMList(3)
	expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus", compute.Flexible)

	provider := newTestProvider(t)
	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMSSClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedScaleSets, nil).AnyTimes()
	provider.azureManager.config.EnableVmssFlexNodes = false
	provider.azureManager.azClient.virtualMachineScaleSetsClient = mockVMSSClient
	mockVMClient := mockvmclient.NewMockInterface(ctrl)

	mockVMClient.EXPECT().List(gomock.Any(), provider.azureManager.config.ResourceGroup).Return(expectedVMs, nil).AnyTimes()
	mockVMClient.EXPECT().ListVmssFlexVMsWithoutInstanceView(gomock.Any(), testASG).Return(expectedVMs, nil).AnyTimes()
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

func TestTemplateNodeInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedScaleSets := newTestVMSSList(3, "test-asg", "eastus", compute.Uniform)

	provider := newTestProvider(t)
	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
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

	t.Run("Checking fallback to static because dynamic list is empty", func(t *testing.T) {
		asg.enableDynamicInstanceList = true

		nodeInfo, err := asg.TemplateNodeInfo()
		assert.NoError(t, err)
		assert.NotNil(t, nodeInfo)
		assert.NotEmpty(t, nodeInfo.Pods)
	})

	// Properly testing dynamic SKU list through skewer is not possible,
	// because there are no Resource API mocks included yet.
	// Instead, the rest of the (consumer side) tests here
	// override GetVMSSTypeDynamically and GetVMSSTypeStatically functions.

	t.Run("Checking dynamic workflow", func(t *testing.T) {
		asg.enableDynamicInstanceList = true

		GetVMSSTypeDynamically = func(template NodeTemplate, azCache *azureCache) (InstanceType, error) {
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
		assert.NotEmpty(t, nodeInfo.Pods)
	})

	t.Run("Checking static workflow if dynamic fails", func(t *testing.T) {
		asg.enableDynamicInstanceList = true

		GetVMSSTypeDynamically = func(template NodeTemplate, azCache *azureCache) (InstanceType, error) {
			return InstanceType{}, fmt.Errorf("dynamic error exists")
		}
		GetVMSSTypeStatically = func(template NodeTemplate) (*InstanceType, error) {
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
		assert.NotEmpty(t, nodeInfo.Pods)
	})

	t.Run("Fails to find vmss instance information using static and dynamic workflow, instance not supported", func(t *testing.T) {
		asg.enableDynamicInstanceList = true

		GetVMSSTypeDynamically = func(template NodeTemplate, azCache *azureCache) (InstanceType, error) {
			return InstanceType{}, fmt.Errorf("dynamic error exists")
		}
		GetVMSSTypeStatically = func(template NodeTemplate) (*InstanceType, error) {
			return &InstanceType{}, fmt.Errorf("static error exists")
		}
		nodeInfo, err := asg.TemplateNodeInfo()
		assert.Empty(t, nodeInfo)
		assert.Equal(t, err, fmt.Errorf("static error exists"))
	})

	// Note: static-only workflow tests can be removed once support for dynamic is always on

	t.Run("Checking static-only workflow", func(t *testing.T) {
		asg.enableDynamicInstanceList = false

		GetVMSSTypeStatically = func(template NodeTemplate) (*InstanceType, error) {
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
		assert.NotEmpty(t, nodeInfo.Pods)
	})

	t.Run("Checking static-only workflow with built-in SKU list", func(t *testing.T) {
		asg.enableDynamicInstanceList = false

		nodeInfo, err := asg.TemplateNodeInfo()
		assert.NoError(t, err)
		assert.NotNil(t, nodeInfo)
		assert.NotEmpty(t, nodeInfo.Pods)
	})

}
func TestCseErrors(t *testing.T) {
	errorMessage := to.StringPtr("Error Message Test")
	vmssVMs := compute.VirtualMachineScaleSetVM{
		Name:       to.StringPtr("vmTest"),
		ID:         to.StringPtr(fakeVirtualMachineScaleSetVMID),
		InstanceID: to.StringPtr("0"),
		VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
			VMID:              to.StringPtr("123E4567-E89B-12D3-A456-426655440000"),
			ProvisioningState: to.StringPtr("Succeeded"),
			InstanceView: &compute.VirtualMachineScaleSetVMInstanceView{
				Extensions: &[]compute.VirtualMachineExtensionInstanceView{
					{
						Statuses: &[]compute.InstanceViewStatus{
							{
								Level:   "Error",
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
		(*vmssVMs.InstanceView.Extensions)[0].Name = to.StringPtr(vmssCSEExtensionName)
		actualCSEErrorMessage, actualCSEFailureBool := scaleSet.cseErrors(vmssVMs.InstanceView.Extensions)
		assert.True(t, actualCSEFailureBool)
		assert.Equal(t, []string{expectedCSEWErrorMessage}, actualCSEErrorMessage)
	})
	t.Run("getCSEErrorMessages test with no CSE error in VM extensions", func(t *testing.T) {
		(*vmssVMs.InstanceView.Extensions)[0].Name = to.StringPtr("notCSEExtension")
		actualCSEErrorMessage, actualCSEFailureBool := scaleSet.cseErrors(vmssVMs.InstanceView.Extensions)
		assert.False(t, actualCSEFailureBool)
		assert.Equal(t, []string(nil), actualCSEErrorMessage)
	})
}
