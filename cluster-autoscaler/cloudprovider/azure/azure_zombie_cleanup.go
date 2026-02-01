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
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

type zombieInstance struct {
	vmssName   string
	instanceID string
	hasK8sNode bool
	reason     string
}

// cleanupZombieNodes detects and cleans up zombie VMSS instances that are in non-functional states.
// A zombie instance is one that:
// - Has failed provisioning state
// - Has failed VM extensions
// - Never registered in Kubernetes (age > threshold)
// - Has unreachable taint (age > threshold)
// - Is NotReady with a running VM (age > threshold)
//
// Implementation follows a hybrid approach:
// - VMs that never registered in K8s are directly deleted
// - VMs with K8s nodes (unreachable/NotReady) are logged only (autoscaler handles deletion)
func (m *AzureManager) cleanupZombieNodes() error {
	return m.cleanupZombieNodesWithContext(nil)
}

// cleanupZombieNodesWithContext is similar to cleanupZombieNodes but accepts Kubernetes nodes
// as context to help correlate VMSS instances with K8s nodes.
func (m *AzureManager) cleanupZombieNodesWithContext(nodes []*apiv1.Node) error {
	ctx, cancel := getContextWithTimeout(vmssContextTimeout)
	defer cancel()

	// Get all VMSS in the resource group
	scaleSets, err := m.azClient.virtualMachineScaleSetsClient.List(ctx, m.config.ResourceGroup)
	if err != nil {
		klog.Errorf("Failed to list VMSS: %v", err)
		return fmt.Errorf("failed to list VMSS: %v", err)
	}

	// Build K8s node lookup table
	k8sNodeMap := make(map[string]*apiv1.Node)
	if nodes != nil {
		for _, node := range nodes {
			if node.Spec.ProviderID == "" {
				continue
			}
			// Normalize provider ID
			normalizedID := normalizeProviderID(node.Spec.ProviderID)
			k8sNodeMap[normalizedID] = node
		}
	}

	// Process each VMSS
	zombiesByVMSS := make(map[string][]zombieInstance)
	currentTime := time.Now()
	minAgeDuration := time.Duration(m.config.ZombieMinAgeMinutes) * time.Minute

	for _, scaleSet := range scaleSets {
		vmssName := ptr.Deref(scaleSet.Name, "")
		if vmssName == "" {
			continue
		}

		// Get all VMs in this VMSS with instance view
		vms, err := m.azClient.virtualMachineScaleSetVMsClient.List(
			ctx,
			m.config.ResourceGroup,
			vmssName,
			string(compute.InstanceViewTypesInstanceView),
		)
		if err != nil {
			klog.Warningf("Failed to list VMs for VMSS %s: %v", vmssName, err)
			continue
		}

		// Empty VMSS is normal - scale to zero
		if len(vms) == 0 {
			continue
		}

		// Analyze each VMSS instance for zombie status
		for _, vm := range vms {
			instanceID := ptr.Deref(vm.InstanceID, "")
			if instanceID == "" {
				continue
			}

			// Evaluate zombie conditions
			isZombie, hasK8sNode, reason := m.evaluateZombieStatus(vm, k8sNodeMap, currentTime, minAgeDuration)

			if isZombie {
				klog.V(2).Infof("ZOMBIE DETECTED: Instance %s (VMSS: %s) - %s", instanceID, vmssName, reason)
				zombiesByVMSS[vmssName] = append(zombiesByVMSS[vmssName], zombieInstance{
					vmssName:   vmssName,
					instanceID: instanceID,
					hasK8sNode: hasK8sNode,
					reason:     reason,
				})
			}
		}
	}

	// Cleanup phase
	totalZombies := 0
	for _, zombies := range zombiesByVMSS {
		totalZombies += len(zombies)
	}

	if totalZombies == 0 {
		klog.V(4).Info("No zombies found")
		return nil
	}

	// Separate zombies by whether they have K8s nodes
	// Manually delete unregistered VMs, log registered ones
	unregisteredZombies := make(map[string][]zombieInstance)
	registeredZombieCount := 0

	for vmssName, zombies := range zombiesByVMSS {
		for _, zombie := range zombies {
			if !zombie.hasK8sNode {
				// Safe to delete directly - no K8s state
				unregisteredZombies[vmssName] = append(unregisteredZombies[vmssName], zombie)
			} else {
				// Has K8s node - let autoscaler handle deletion
				registeredZombieCount++
				klog.V(2).Infof("Registered zombie node (will be handled by autoscaler): Instance %s (VMSS: %s) - %s",
					zombie.instanceID, vmssName, zombie.reason)
			}
		}
	}

	unregisteredCount := totalZombies - registeredZombieCount

	// Dry-run or active cleanup for unregistered zombies
	if m.config.ZombieCleanupDryRun {
		klog.Infof("DRY RUN: Would clean up %d unregistered zombie(s), %d registered zombies logged",
			unregisteredCount, registeredZombieCount)
		for vmssName, zombies := range unregisteredZombies {
			for _, zombie := range zombies {
				klog.V(2).Infof("Would remove: Instance %s (VMSS: %s, reason: %s)",
					zombie.instanceID, vmssName, zombie.reason)
			}
		}
		return nil
	}

	if unregisteredCount == 0 {
		klog.V(2).Infof("No unregistered zombies to clean up (%d registered zombies will be handled by autoscaler)",
			registeredZombieCount)
		return nil
	}

	// Actually cleaning up unregistered zombies
	klog.Infof("Cleaning up %d unregistered zombie(s) (%d registered zombies will be handled by autoscaler)...",
		unregisteredCount, registeredZombieCount)

	// Batch delete zombies per VMSS for efficiency
	for vmssName, zombies := range unregisteredZombies {
		instanceIDs := make([]string, len(zombies))
		for i, zombie := range zombies {
			instanceIDs[i] = zombie.instanceID
		}

		vmInstanceIDs := compute.VirtualMachineScaleSetVMInstanceRequiredIDs{
			InstanceIds: &instanceIDs,
		}

		klog.V(2).Infof("Deleting %d unregistered zombie instance(s) from VMSS %s: %v",
			len(instanceIDs), vmssName, instanceIDs)

		// Delete VMSS instances
		_, err := m.azClient.virtualMachineScaleSetsClient.DeleteInstancesAsync(
			ctx,
			m.config.ResourceGroup,
			vmssName,
			vmInstanceIDs,
			false, // forceDelete
		)

		if err != nil {
			klog.Errorf("Failed to delete zombie instances from VMSS %s: %v", vmssName, err)
			continue
		}

		// Log success for each zombie
		for _, zombie := range zombies {
			klog.V(2).Infof("Successfully cleaned zombie: Instance %s (VMSS: %s)", zombie.instanceID, vmssName)
		}
	}

	return nil
}

// evaluateZombieStatus evaluates a VMSS VM to determine if it's a zombie
// Returns (isZombie bool, hasK8sNode bool, reason string)
func (m *AzureManager) evaluateZombieStatus(
	vm compute.VirtualMachineScaleSetVM,
	k8sNodeMap map[string]*apiv1.Node,
	currentTime time.Time,
	minAgeDuration time.Duration,
) (bool, bool, string) {
	provisioningState := ptr.Deref(vm.ProvisioningState, "")
	powerState := ""
	var vmssAge *time.Duration
	if vm.InstanceView != nil && vm.InstanceView.Statuses != nil {
		for _, status := range *vm.InstanceView.Statuses {
			code := ptr.Deref(status.Code, "")

			if strings.HasPrefix(code, "PowerState/") {
				powerState = strings.TrimPrefix(code, "PowerState/")
			}

			if vmssAge == nil && status.Time != nil {
				age := currentTime.Sub(status.Time.Time)
				vmssAge = &age
			}
		}
	}

	// Check VM extension status
	extensionsInstalled := false
	extensionsFailed := false
	if vm.InstanceView != nil && vm.InstanceView.Extensions != nil && len(*vm.InstanceView.Extensions) > 0 {
		extensionsInstalled = true
		for _, ext := range *vm.InstanceView.Extensions {
			if ext.Statuses != nil {
				for _, status := range *ext.Statuses {
					code := ptr.Deref(status.Code, "")
					if status.Level == compute.Error || code == "ProvisioningState/failed" {
						extensionsFailed = true
						break
					}
				}
			}
			if extensionsFailed {
				break
			}
		}
	}

	// Find corresponding K8s node
	var k8sNode *apiv1.Node
	hasK8sNode := false
	normalizedVMID := normalizeProviderID(ptr.Deref(vm.ID, ""))
	if normalizedVMID != "" {
		k8sNode, hasK8sNode = k8sNodeMap[normalizedVMID]
	}

	// Calculate node age if K8s node exists
	var nodeAge *time.Duration
	isReady := false
	if hasK8sNode && k8sNode != nil {
		age := currentTime.Sub(k8sNode.CreationTimestamp.Time)
		nodeAge = &age

		// Check Ready condition
		for _, condition := range k8sNode.Status.Conditions {
			if condition.Type == apiv1.NodeReady && condition.Status == apiv1.ConditionTrue {
				isReady = true
				break
			}
		}
	}

	// SCENARIO 1: Extensions failed or never installed
	if !hasK8sNode && (!extensionsInstalled || extensionsFailed) {
		if vmssAge != nil && *vmssAge > minAgeDuration {
			if extensionsFailed {
				return true, false, fmt.Sprintf("Extensions FAILED, age: %.0f minutes", vmssAge.Minutes())
			}
			return true, false, fmt.Sprintf("Extensions NOT INSTALLED (flapping zombie), age: %.0f minutes", vmssAge.Minutes())
		} else if vmssAge == nil {
			if extensionsFailed {
				return true, false, "Extensions FAILED, no timestamp"
			}
			return true, false, "Extensions NOT INSTALLED, no timestamp"
		}
	}

	// SCENARIO 2: Provisioning failed
	if provisioningState == "Failed" {
		return true, false, "Provisioning FAILED"
	}

	// SCENARIO 3: VMSS succeeded but never registered in K8s
	if !hasK8sNode && provisioningState == "Succeeded" {
		if vmssAge != nil && *vmssAge > minAgeDuration {
			return true, false, fmt.Sprintf("Never registered in K8s (AllocationFailed), age: %.0f minutes", vmssAge.Minutes())
		} else if vmssAge == nil {
			return true, false, "Never registered in K8s (AllocationFailed), no timestamp"
		}
	}

	// SCENARIO 4: Has K8s node but with critical issues
	if hasK8sNode && k8sNode != nil {
		// Sub-check 4a: Unreachable taint
		hasUnreachableTaint := false
		if k8sNode.Spec.Taints != nil {
			for _, taint := range k8sNode.Spec.Taints {
				if taint.Key == "node.kubernetes.io/unreachable" {
					hasUnreachableTaint = true
					break
				}
			}
		}

		if hasUnreachableTaint {
			if nodeAge != nil && *nodeAge > minAgeDuration {
				return true, true, fmt.Sprintf("Node has unreachable taint for %.0f minutes", nodeAge.Minutes())
			}
			// Node is too young to be considered a zombie
			return false, true, ""
		}

		// Sub-check 4b: NotReady status (only if no unreachable taint)
		if !isReady {
			// Special case: Deallocated by autoscaler so healthy state
			if powerState == "deallocated" {
				return false, true, ""
			}

			// Problem case: VM running but node not ready
			if powerState == "running" && nodeAge != nil && *nodeAge > minAgeDuration {
				return true, true, fmt.Sprintf("Running VM but NotReady node for %.0f minutes", nodeAge.Minutes())
			}

			// Node is too young or not running
			return false, true, ""
		}

		// Node is healthy
		return false, true, ""
	}

	// Not a zombie
	return false, false, ""
}

// normalizeProviderID normalizes Azure provider IDs for comparison
// Removes "azure://" prefix and ensures consistent format
func normalizeProviderID(providerID string) string {
	normalized := strings.TrimPrefix(providerID, "azure://")
	normalized = strings.ReplaceAll(normalized, "\\", "/")
	normalized = strings.ToLower(normalized)
	return normalized
}
