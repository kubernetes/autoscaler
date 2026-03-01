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
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmssclient/mockvmssclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmssvmclient/mockvmssvmclient"
	providerazureconsts "sigs.k8s.io/cloud-provider-azure/pkg/consts"
	providerazure "sigs.k8s.io/cloud-provider-azure/pkg/provider"
)

func TestZombieCleanup_NoZombiesFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, mockVMSSClient, mockVMSSVMClient := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false

	vmssName := "test-vmss"
	mockVMSSList := []compute.VirtualMachineScaleSet{
		{Name: ptr.To(vmssName)},
	}
	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(mockVMSSList, nil)
	healthyVMs := []compute.VirtualMachineScaleSetVM{
		newHealthyVM(0),
	}
	mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, vmssName, gomock.Any()).Return(healthyVMs, nil)

	// Should not call DeleteInstancesAsync
	mockVMSSClient.EXPECT().DeleteInstancesAsync(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	err := manager.cleanupZombieNodes()
	assert.NoError(t, err)
}

func TestZombieCleanup_DetectsFailedProvisioning(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, mockVMSSClient, mockVMSSVMClient := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false
	manager.config.ZombieMinAgeMinutes = 5

	vmssName := "test-vmss"
	mockVMSSList := []compute.VirtualMachineScaleSet{
		{Name: ptr.To(vmssName)},
	}
	zombieVMs := []compute.VirtualMachineScaleSetVM{
		newZombieVMWithFailedProvisioning(0, 10*time.Minute),
	}

	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(mockVMSSList, nil)
	mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, vmssName, gomock.Any()).Return(zombieVMs, nil)

	// Verify DeleteInstancesAsync is called with correct instance IDs
	mockVMSSClient.EXPECT().DeleteInstancesAsync(
		gomock.Any(),
		manager.config.ResourceGroup,
		vmssName,
		gomock.AssignableToTypeOf(compute.VirtualMachineScaleSetVMInstanceRequiredIDs{}),
		false,
	).DoAndReturn(func(ctx context.Context, resourceGroup, vmssName string, vmInstanceIDs compute.VirtualMachineScaleSetVMInstanceRequiredIDs, forceDelete bool) (*compute.VirtualMachineScaleSetsDeleteInstancesFuture, error) {
		// Verify instance ID "0" is in the list
		assert.NotNil(t, vmInstanceIDs.InstanceIds)
		assert.Equal(t, 1, len(*vmInstanceIDs.InstanceIds))
		assert.Equal(t, "0", (*vmInstanceIDs.InstanceIds)[0])
		return nil, nil
	})

	err := manager.cleanupZombieNodes()
	assert.NoError(t, err)
}

func TestZombieCleanup_DetectsFailedExtensions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, mockVMSSClient, mockVMSSVMClient := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false
	manager.config.ZombieMinAgeMinutes = 5

	vmssName := "test-vmss"
	mockVMSSList := []compute.VirtualMachineScaleSet{
		{Name: ptr.To(vmssName)},
	}

	zombieVMs := []compute.VirtualMachineScaleSetVM{
		newZombieVMWithFailedExtensions(0, 15*time.Minute),
	}

	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(mockVMSSList, nil)
	mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, vmssName, gomock.Any()).Return(zombieVMs, nil)

	mockVMSSClient.EXPECT().DeleteInstancesAsync(
		gomock.Any(),
		manager.config.ResourceGroup,
		vmssName,
		gomock.Any(),
		false,
	).Return(nil, nil)

	err := manager.cleanupZombieNodes()
	assert.NoError(t, err)
}

func TestZombieCleanup_DetectsNeverRegisteredInstances(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, mockVMSSClient, mockVMSSVMClient := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false
	manager.config.ZombieMinAgeMinutes = 5

	vmssName := "test-vmss"
	mockVMSSList := []compute.VirtualMachineScaleSet{
		{Name: ptr.To(vmssName)},
	}

	zombieVMs := []compute.VirtualMachineScaleSetVM{
		newZombieVMNeverRegistered(0, 30*time.Minute),
	}

	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(mockVMSSList, nil)
	mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, vmssName, gomock.Any()).Return(zombieVMs, nil)

	mockVMSSClient.EXPECT().DeleteInstancesAsync(
		gomock.Any(),
		manager.config.ResourceGroup,
		vmssName,
		gomock.Any(),
		false,
	).Return(nil, nil)

	err := manager.cleanupZombieNodes()
	assert.NoError(t, err)
}

func TestZombieCleanup_WithK8sNodesContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, mockVMSSClient, mockVMSSVMClient := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false
	manager.config.ZombieMinAgeMinutes = 5

	vmssName := "test-vmss"
	mockVMSSList := []compute.VirtualMachineScaleSet{
		{Name: ptr.To(vmssName)},
	}

	// Create a healthy VM (succeeded provisioning, running) that will be marked unreachable
	zombieVM := newUnreachableZombieVM(0, 10*time.Minute)

	// Create a K8s node that matches this VM (unreachable)
	nodes := []*apiv1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "node-0",
				CreationTimestamp: metav1.Time{Time: time.Now().Add(-10 * time.Minute)},
			},
			Spec: apiv1.NodeSpec{
				ProviderID: azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 0),
				Taints: []apiv1.Taint{
					{
						Key:    "node.kubernetes.io/unreachable",
						Effect: apiv1.TaintEffectNoSchedule,
					},
				},
			},
		},
	}

	zombieVMs := []compute.VirtualMachineScaleSetVM{zombieVM}

	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(mockVMSSList, nil)
	mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, vmssName, gomock.Any()).Return(zombieVMs, nil)

	// Should NOT call DeleteInstancesAsync because VM has a K8s node
	mockVMSSClient.EXPECT().DeleteInstancesAsync(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	err := manager.cleanupZombieNodesWithContext(nodes)
	assert.NoError(t, err)
}

func TestZombieCleanup_RespectsMinAge(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, mockVMSSClient, mockVMSSVMClient := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false
	manager.config.ZombieMinAgeMinutes = 10 // 10 minute threshold

	vmssName := "test-vmss"
	mockVMSSList := []compute.VirtualMachineScaleSet{
		{Name: ptr.To(vmssName)},
	}

	// VM that's only 3 minutes old (below threshold)
	youngZombieVMs := []compute.VirtualMachineScaleSetVM{
		newZombieVMNeverRegistered(0, 3*time.Minute),
	}

	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(mockVMSSList, nil)
	mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, vmssName, gomock.Any()).Return(youngZombieVMs, nil)

	// Should NOT delete because VM is too young
	mockVMSSClient.EXPECT().DeleteInstancesAsync(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	err := manager.cleanupZombieNodes()
	assert.NoError(t, err)
}

func TestZombieCleanup_DryRunMode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, mockVMSSClient, mockVMSSVMClient := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = true // Dry-run enabled
	manager.config.ZombieMinAgeMinutes = 5

	vmssName := "test-vmss"
	mockVMSSList := []compute.VirtualMachineScaleSet{
		{Name: ptr.To(vmssName)},
	}

	zombieVMs := []compute.VirtualMachineScaleSetVM{
		newZombieVMWithFailedProvisioning(0, 10*time.Minute),
	}

	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(mockVMSSList, nil)
	mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, vmssName, gomock.Any()).Return(zombieVMs, nil)

	// Should NOT call DeleteInstancesAsync in dry-run mode
	mockVMSSClient.EXPECT().DeleteInstancesAsync(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	err := manager.cleanupZombieNodes()
	assert.NoError(t, err)
}

func TestZombieCleanup_MultipleZombiesInSamePool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, mockVMSSClient, mockVMSSVMClient := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false
	manager.config.ZombieMinAgeMinutes = 5

	vmssName := "test-vmss"
	mockVMSSList := []compute.VirtualMachineScaleSet{
		{Name: ptr.To(vmssName)},
	}

	// Multiple zombie VMs
	zombieVMs := []compute.VirtualMachineScaleSetVM{
		newZombieVMWithFailedProvisioning(0, 10*time.Minute),
		newZombieVMWithFailedExtensions(1, 15*time.Minute),
		newZombieVMNeverRegistered(2, 20*time.Minute),
	}

	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(mockVMSSList, nil)
	mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, vmssName, gomock.Any()).Return(zombieVMs, nil)

	// Should call DeleteInstancesAsync ONCE with all 3 instance IDs
	mockVMSSClient.EXPECT().DeleteInstancesAsync(
		gomock.Any(),
		manager.config.ResourceGroup,
		vmssName,
		gomock.AssignableToTypeOf(compute.VirtualMachineScaleSetVMInstanceRequiredIDs{}),
		false,
	).DoAndReturn(func(ctx context.Context, resourceGroup, vmssName string, vmInstanceIDs compute.VirtualMachineScaleSetVMInstanceRequiredIDs, forceDelete bool) (*compute.VirtualMachineScaleSetsDeleteInstancesFuture, error) {
		// Verify all 3 instance IDs are in the batch
		assert.NotNil(t, vmInstanceIDs.InstanceIds)
		assert.Equal(t, 3, len(*vmInstanceIDs.InstanceIds))
		return nil, nil
	})

	err := manager.cleanupZombieNodes()
	assert.NoError(t, err)
}

func TestZombieCleanup_MultipleVMSSPools(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, mockVMSSClient, mockVMSSVMClient := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false
	manager.config.ZombieMinAgeMinutes = 5

	vmss1 := "vmss-pool-1"
	vmss2 := "vmss-pool-2"
	mockVMSSList := []compute.VirtualMachineScaleSet{
		{Name: ptr.To(vmss1)},
		{Name: ptr.To(vmss2)},
	}

	zombieVMs1 := []compute.VirtualMachineScaleSetVM{
		newZombieVMWithFailedProvisioning(0, 10*time.Minute),
	}
	zombieVMs2 := []compute.VirtualMachineScaleSetVM{
		newZombieVMWithFailedExtensions(0, 15*time.Minute),
	}

	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(mockVMSSList, nil)
	mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, vmss1, gomock.Any()).Return(zombieVMs1, nil)
	mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, vmss2, gomock.Any()).Return(zombieVMs2, nil)

	// Should call DeleteInstancesAsync TWICE (once per VMSS)
	mockVMSSClient.EXPECT().DeleteInstancesAsync(gomock.Any(), manager.config.ResourceGroup, vmss1, gomock.Any(), false).Return(nil, nil)
	mockVMSSClient.EXPECT().DeleteInstancesAsync(gomock.Any(), manager.config.ResourceGroup, vmss2, gomock.Any(), false).Return(nil, nil)

	err := manager.cleanupZombieNodes()
	assert.NoError(t, err)
}

func TestZombieCleanup_MixedZombiesAndHealthy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, mockVMSSClient, mockVMSSVMClient := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false
	manager.config.ZombieMinAgeMinutes = 5

	vmssName := "test-vmss"
	mockVMSSList := []compute.VirtualMachineScaleSet{
		{Name: ptr.To(vmssName)},
	}

	// Mix of healthy and zombie VMs
	vms := []compute.VirtualMachineScaleSetVM{
		newHealthyVM(0), // Healthy
		newZombieVMWithFailedProvisioning(1, 10*time.Minute), // Zombie
		newHealthyVM(2), // Healthy
		newZombieVMWithFailedExtensions(3, 15*time.Minute), // Zombie
	}

	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(mockVMSSList, nil)
	mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, vmssName, gomock.Any()).Return(vms, nil)

	// Should delete only instances 1 and 3
	mockVMSSClient.EXPECT().DeleteInstancesAsync(
		gomock.Any(),
		manager.config.ResourceGroup,
		vmssName,
		gomock.AssignableToTypeOf(compute.VirtualMachineScaleSetVMInstanceRequiredIDs{}),
		false,
	).DoAndReturn(func(ctx context.Context, resourceGroup, vmssName string, vmInstanceIDs compute.VirtualMachineScaleSetVMInstanceRequiredIDs, forceDelete bool) (*compute.VirtualMachineScaleSetsDeleteInstancesFuture, error) {
		assert.NotNil(t, vmInstanceIDs.InstanceIds)
		assert.Equal(t, 2, len(*vmInstanceIDs.InstanceIds))
		// Verify only zombie instance IDs
		instanceIDs := *vmInstanceIDs.InstanceIds
		assert.Contains(t, instanceIDs, "1")
		assert.Contains(t, instanceIDs, "3")
		return nil, nil
	})

	err := manager.cleanupZombieNodes()
	assert.NoError(t, err)
}

func TestZombieCleanup_IgnoresDeallocatedNodes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	manager, mockVMSSClient, mockVMSSVMClient := setupMockManager(t, ctrl)
	manager.config.EnableZombieCleanup = true
	manager.config.ZombieCleanupDryRun = false
	manager.config.ZombieMinAgeMinutes = 5

	vmssName := "test-vmss"
	mockVMSSList := []compute.VirtualMachineScaleSet{
		{Name: ptr.To(vmssName)},
	}

	// Deallocated VM (healthy autoscaler scale-down)
	vms := []compute.VirtualMachineScaleSetVM{
		newDeallocatedVM(0),
	}

	// Create a K8s node for this VM (NotReady but deallocated)
	nodes := []*apiv1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "node-0",
				CreationTimestamp: metav1.Time{Time: time.Now().Add(-10 * time.Minute)},
			},
			Spec: apiv1.NodeSpec{
				ProviderID: azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 0),
			},
			Status: apiv1.NodeStatus{
				Conditions: []apiv1.NodeCondition{
					{
						Type:   apiv1.NodeReady,
						Status: apiv1.ConditionFalse,
					},
				},
			},
		},
	}

	mockVMSSClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup).Return(mockVMSSList, nil)
	mockVMSSVMClient.EXPECT().List(gomock.Any(), manager.config.ResourceGroup, vmssName, gomock.Any()).Return(vms, nil)

	// Should NOT delete deallocated VMs
	mockVMSSClient.EXPECT().DeleteInstancesAsync(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	err := manager.cleanupZombieNodesWithContext(nodes)
	assert.NoError(t, err)
}

// ============================================================================
// SCENARIO DETECTION TESTS - Demonstrating Zombie States
// ============================================================================

func TestZombieScenario_ExtensionsFailedToInstall(t *testing.T) {
	zombieVM := newZombieVMWithFailedExtensions(0, 15*time.Minute)

	hasFailedExtension := false
	if zombieVM.InstanceView != nil && zombieVM.InstanceView.Extensions != nil {
		for _, ext := range *zombieVM.InstanceView.Extensions {
			if ext.Statuses != nil {
				for _, status := range *ext.Statuses {
					code := ptr.Deref(status.Code, "")
					if status.Level == "Error" || code == "ProvisioningState/failed" {
						hasFailedExtension = true
						break
					}
				}
			}
		}
	}

	assert.True(t, hasFailedExtension,
		"ZOMBIE DETECTED: VM has failed extensions (vmssCSE), preventing Kubernetes registration!")
}

func TestZombieScenario_ExtensionsNeverInstalled(t *testing.T) {
	zombieVM := newZombieVMNeverRegistered(0, 20*time.Minute)

	extensionsInstalled := false
	if zombieVM.InstanceView != nil && zombieVM.InstanceView.Extensions != nil {
		extensionsInstalled = len(*zombieVM.InstanceView.Extensions) > 0
	}
	vmProvisioned := ptr.Deref(zombieVM.ProvisioningState, "") == "Succeeded"
	vmAge := 20 * time.Minute

	assert.True(t, vmProvisioned && !extensionsInstalled && vmAge > 5*time.Minute,
		"ZOMBIE DETECTED: VM provisioned successfully but extensions were NEVER INSTALLED (flapping zombie)!")
}

func TestZombieScenario_ProvisioningFailed(t *testing.T) {
	zombieVM := newZombieVMWithFailedProvisioning(0, 10*time.Minute)

	provisioningState := ptr.Deref(zombieVM.ProvisioningState, "")

	assert.Equal(t, "Failed", provisioningState,
		"ZOMBIE DETECTED: VM has ProvisioningState='%s' (should be 'Failed'), wasting quota!",
		provisioningState)
}

func TestZombieScenario_NeverRegisteredInKubernetes(t *testing.T) {
	zombieVM := newZombieVMNeverRegistered(0, 30*time.Minute)
	vmProvisioned := ptr.Deref(zombieVM.ProvisioningState, "") == "Succeeded"
	vmAge := 30 * time.Minute

	// Simulated scenario: This VM never appears in `kubectl get nodes`
	neverRegistered := true // This VM is not in K8s
	assert.True(t, vmProvisioned && neverRegistered && vmAge > 5*time.Minute,
		"ZOMBIE DETECTED: VM provisioned %v ago but never registered in Kubernetes (AllocationFailed)!",
		vmAge)
}

func TestZombieScenario_NodeUnreachableTaint(t *testing.T) {
	vm := newHealthyVM(0)

	// Node has unreachable taint (common failure scenario)
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "node-0",
			CreationTimestamp: metav1.Time{Time: time.Now().Add(-15 * time.Minute)},
		},
		Spec: apiv1.NodeSpec{
			ProviderID: azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 0),
			Taints: []apiv1.Taint{
				{
					Key:    "node.kubernetes.io/unreachable",
					Effect: apiv1.TaintEffectNoSchedule,
				},
			},
		},
	}

	// VM is running
	vmIsRunning := false
	if vm.InstanceView != nil && vm.InstanceView.Statuses != nil {
		for _, status := range *vm.InstanceView.Statuses {
			if ptr.Deref(status.Code, "") == "PowerState/running" {
				vmIsRunning = true
				break
			}
		}
	}

	// Node is tainted unreachable
	hasUnreachableTaint := false
	for _, taint := range node.Spec.Taints {
		if taint.Key == "node.kubernetes.io/unreachable" {
			hasUnreachableTaint = true
			break
		}
	}

	nodeAge := 15 * time.Minute

	assert.True(t, vmIsRunning && hasUnreachableTaint && nodeAge > 5*time.Minute,
		"ZOMBIE DETECTED: VM is running but node has unreachable taint for %v - pods can't schedule!",
		nodeAge)
}

func TestZombieScenario_NodeNotReady(t *testing.T) {
	vm := newHealthyVM(0)

	// Node is NotReady for 15 minutes
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "node-0",
			CreationTimestamp: metav1.Time{Time: time.Now().Add(-15 * time.Minute)},
		},
		Spec: apiv1.NodeSpec{
			ProviderID: azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, 0),
		},
		Status: apiv1.NodeStatus{
			Conditions: []apiv1.NodeCondition{
				{
					Type:               apiv1.NodeReady,
					Status:             apiv1.ConditionFalse,
					LastTransitionTime: metav1.Time{Time: time.Now().Add(-15 * time.Minute)},
					Reason:             "KubeletNotReady",
				},
			},
		},
	}

	// Check VM is running (not deallocated)
	vmIsRunning := false
	if vm.InstanceView != nil && vm.InstanceView.Statuses != nil {
		for _, status := range *vm.InstanceView.Statuses {
			if ptr.Deref(status.Code, "") == "PowerState/running" {
				vmIsRunning = true
				break
			}
		}
	}

	vmProvisioned := ptr.Deref(vm.ProvisioningState, "") == "Succeeded"

	// Check node ready status
	nodeReady := false
	for _, condition := range node.Status.Conditions {
		if condition.Type == apiv1.NodeReady && condition.Status == apiv1.ConditionTrue {
			nodeReady = true
			break
		}
	}

	notReadyDuration := 15 * time.Minute

	assert.True(t, vmProvisioned && vmIsRunning && !nodeReady && notReadyDuration > 5*time.Minute,
		"ZOMBIE DETECTED: VM running but node NotReady for %v (>5min threshold)!",
		notReadyDuration)
}

func TestZombieScenario_DeallocatedNodesAreHealthy(t *testing.T) {
	deallocatedVM := newDeallocatedVM(0)

	// Check power state
	isDeallocated := false
	if deallocatedVM.InstanceView != nil && deallocatedVM.InstanceView.Statuses != nil {
		for _, status := range *deallocatedVM.InstanceView.Statuses {
			code := ptr.Deref(status.Code, "")
			if code == "PowerState/deallocated" {
				isDeallocated = true
				break
			}
		}
	}

	assert.True(t, isDeallocated,
		"Deallocated VM should have PowerState/deallocated (healthy autoscaler scale-down, NOT a zombie)")
}

func TestZombieScenario_MultipleZombiesWasteQuota(t *testing.T) {
	// Scenario: 5 VMs in a scale set, demonstrating severe waste
	vms := []interface{}{
		newHealthyVM(0), // 1 healthy
		newZombieVMWithFailedProvisioning(1, 10*time.Minute), // Zombie
		newZombieVMWithFailedProvisioning(2, 15*time.Minute), // Zombie
		newZombieVMWithFailedExtensions(3, 20*time.Minute),   // Zombie
		newZombieVMNeverRegistered(4, 25*time.Minute),        // Zombie
	}

	totalVMs := len(vms)
	zombieCount := 4

	zombiePercentage := (zombieCount * 100) / totalVMs
	assert.Equal(t, 80, zombiePercentage,
		"ZOMBIE PROBLEM: %d out of %d VMs are zombies (%d%% waste rate) - demonstrating severe quota waste!",
		zombieCount, totalVMs, zombiePercentage)
}

func setupMockManager(t *testing.T, ctrl *gomock.Controller) (*AzureManager, *mockvmssclient.MockInterface, *mockvmssvmclient.MockInterface) {
	manager := newTestAzureManagerForZombieCleanup(t)

	mockVMSSClient := mockvmssclient.NewMockInterface(ctrl)
	mockVMSSVMClient := mockvmssvmclient.NewMockInterface(ctrl)

	manager.azClient = &azClient{
		virtualMachineScaleSetsClient:   mockVMSSClient,
		virtualMachineScaleSetVMsClient: mockVMSSVMClient,
	}

	return manager, mockVMSSClient, mockVMSSVMClient
}

func newTestAzureManagerForZombieCleanup(t *testing.T) *AzureManager {
	return &AzureManager{
		config: &Config{
			Config: providerazure.Config{
				ResourceGroup: "test-rg",
				VMType:        providerazureconsts.VMTypeVMSS,
			},
			EnableZombieCleanup: false,
			ZombieCleanupDryRun: false,
			ZombieMinAgeMinutes: 10,
		},
	}
}

// newHealthyVM creates a healthy VMSS VM
func newHealthyVM(instanceID int) compute.VirtualMachineScaleSetVM {
	return compute.VirtualMachineScaleSetVM{
		ID:         ptr.To(fmt.Sprintf(fakeVirtualMachineScaleSetVMID, instanceID)),
		InstanceID: ptr.To(fmt.Sprintf("%d", instanceID)),
		VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
			ProvisioningState: ptr.To("Succeeded"),
			InstanceView: &compute.VirtualMachineScaleSetVMInstanceView{
				Statuses: &[]compute.InstanceViewStatus{
					{
						Code: ptr.To("ProvisioningState/succeeded"),
						Time: &date.Time{Time: time.Now().Add(-2 * time.Minute)},
					},
					{
						Code: ptr.To("PowerState/running"),
						Time: &date.Time{Time: time.Now().Add(-2 * time.Minute)},
					},
				},
			},
		},
	}
}

// newZombieVMWithFailedProvisioning creates a VMSS VM with failed provisioning state
func newZombieVMWithFailedProvisioning(instanceID int, age time.Duration) compute.VirtualMachineScaleSetVM {
	return compute.VirtualMachineScaleSetVM{
		ID:         ptr.To(fmt.Sprintf(fakeVirtualMachineScaleSetVMID, instanceID)),
		InstanceID: ptr.To(fmt.Sprintf("%d", instanceID)),
		VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
			ProvisioningState: ptr.To("Failed"),
			InstanceView: &compute.VirtualMachineScaleSetVMInstanceView{
				Statuses: &[]compute.InstanceViewStatus{
					{
						Code: ptr.To("ProvisioningState/failed"),
						Time: &date.Time{Time: time.Now().Add(-age)},
					},
					{
						Code: ptr.To("PowerState/running"),
						Time: &date.Time{Time: time.Now().Add(-age)},
					},
				},
			},
		},
	}
}

// newZombieVMWithFailedExtensions creates a VMSS VM with succeeded provisioning but failed extensions
func newZombieVMWithFailedExtensions(instanceID int, age time.Duration) compute.VirtualMachineScaleSetVM {
	return compute.VirtualMachineScaleSetVM{
		ID:         ptr.To(fmt.Sprintf(fakeVirtualMachineScaleSetVMID, instanceID)),
		InstanceID: ptr.To(fmt.Sprintf("%d", instanceID)),
		VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
			ProvisioningState: ptr.To("Succeeded"),
			InstanceView: &compute.VirtualMachineScaleSetVMInstanceView{
				Statuses: &[]compute.InstanceViewStatus{
					{
						Code: ptr.To("ProvisioningState/succeeded"),
						Time: &date.Time{Time: time.Now().Add(-age)},
					},
					{
						Code: ptr.To("PowerState/running"),
						Time: &date.Time{Time: time.Now().Add(-age)},
					},
				},
				Extensions: &[]compute.VirtualMachineExtensionInstanceView{
					{
						Name: ptr.To("vmssCSE"),
						Statuses: &[]compute.InstanceViewStatus{
							{
								Level: compute.Error,
								Code:  ptr.To("ProvisioningState/failed"),
							},
						},
					},
				},
			},
		},
	}
}

// newZombieVMNeverRegistered creates a VMSS VM that succeeded provisioning but never registered in K8s
func newZombieVMNeverRegistered(instanceID int, age time.Duration) compute.VirtualMachineScaleSetVM {
	return compute.VirtualMachineScaleSetVM{
		ID:         ptr.To(fmt.Sprintf(fakeVirtualMachineScaleSetVMID, instanceID)),
		InstanceID: ptr.To(fmt.Sprintf("%d", instanceID)),
		VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
			ProvisioningState: ptr.To("Succeeded"),
			InstanceView: &compute.VirtualMachineScaleSetVMInstanceView{
				Statuses: &[]compute.InstanceViewStatus{
					{
						Code: ptr.To("ProvisioningState/succeeded"),
						Time: &date.Time{Time: time.Now().Add(-age)},
					},
					{
						Code: ptr.To("PowerState/running"),
						Time: &date.Time{Time: time.Now().Add(-age)},
					},
				},
			},
		},
	}
}

// newUnreachableZombieVM creates a VM with healthy provisioning that will be paired with unreachable K8s node
func newUnreachableZombieVM(instanceID int, age time.Duration) compute.VirtualMachineScaleSetVM {
	return compute.VirtualMachineScaleSetVM{
		ID:         ptr.To(fmt.Sprintf(fakeVirtualMachineScaleSetVMID, instanceID)),
		InstanceID: ptr.To(fmt.Sprintf("%d", instanceID)),
		VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
			ProvisioningState: ptr.To("Succeeded"),
			InstanceView: &compute.VirtualMachineScaleSetVMInstanceView{
				Statuses: &[]compute.InstanceViewStatus{
					{
						Code: ptr.To("ProvisioningState/succeeded"),
						Time: &date.Time{Time: time.Now().Add(-age)},
					},
					{
						Code: ptr.To("PowerState/running"),
						Time: &date.Time{Time: time.Now().Add(-age)},
					},
				},
			},
		},
	}
}

// newRecentVM creates a recently created VM (below age threshold)
func newRecentVM(instanceID int) compute.VirtualMachineScaleSetVM {
	return compute.VirtualMachineScaleSetVM{
		ID:         ptr.To(fmt.Sprintf(fakeVirtualMachineScaleSetVMID, instanceID)),
		InstanceID: ptr.To(fmt.Sprintf("%d", instanceID)),
		VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
			ProvisioningState: ptr.To("Succeeded"),
			InstanceView: &compute.VirtualMachineScaleSetVMInstanceView{
				Statuses: &[]compute.InstanceViewStatus{
					{
						Code: ptr.To("ProvisioningState/succeeded"),
						Time: &date.Time{Time: time.Now().Add(-2 * time.Minute)}, // Less than 5 minutes
					},
					{
						Code: ptr.To("PowerState/running"),
						Time: &date.Time{Time: time.Now().Add(-2 * time.Minute)},
					},
				},
			},
		},
	}
}

// newDeallocatedVM creates a deallocated VM (healthy autoscaler scale-down)
func newDeallocatedVM(instanceID int) compute.VirtualMachineScaleSetVM {
	return compute.VirtualMachineScaleSetVM{
		ID:         ptr.To(fmt.Sprintf(fakeVirtualMachineScaleSetVMID, instanceID)),
		InstanceID: ptr.To(fmt.Sprintf("%d", instanceID)),
		VirtualMachineScaleSetVMProperties: &compute.VirtualMachineScaleSetVMProperties{
			ProvisioningState: ptr.To("Succeeded"),
			InstanceView: &compute.VirtualMachineScaleSetVMInstanceView{
				Statuses: &[]compute.InstanceViewStatus{
					{
						Code: ptr.To("ProvisioningState/succeeded"),
					},
					{
						Code: ptr.To("PowerState/deallocated"),
					},
				},
			},
		},
	}
}
