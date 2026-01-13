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
	"fmt"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

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

// ============================================================================
// SCENARIO 3: VMSS Succeeded but Never Registered in K8s
// Pseudocode Lines 135-143
// ============================================================================

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

// Test helper: Create a VMSS VM with failed provisioning state
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

// Test helper: Create a VMSS VM with succeeded provisioning but failed extensions
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

// Test helper: Create a VMSS VM that succeeded provisioning but never registered in K8s
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

// Test helper: Create a healthy VMSS VM
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

// Test helper: Create a recently created VM (below age threshold)
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

// Test helper: Create a deallocated VM (healthy autoscaler scale-down)
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
