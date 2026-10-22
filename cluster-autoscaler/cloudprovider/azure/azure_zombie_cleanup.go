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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

type zombieInstance struct {
	vmssName   string
	instanceID string
	hasK8sNode bool
	reason     string
}

type zombieObservationKey struct {
	nodeGroupID string
	providerID  string
	createdAt   time.Time
}

func (m *AzureManager) resetZombieObservations() {
	m.zombieCleanupMutex.Lock()
	defer m.zombieCleanupMutex.Unlock()
	m.zombieObservations = nil
}

// cleanupZombieNodes detects and cleans up non-functional VMSS instances in autoscaler-managed node groups.
// A zombie instance is one that meets any of the following conditions:
// - Has failed provisioning state, is stopped, and has no K8s node
// - Has failed VM extensions
// - Never registered in Kubernetes (age > threshold)
// - Has unreachable taint beyond the threshold (using taint or Ready-condition timestamps)
// - Is NotReady with a running VM beyond the threshold (using the Ready-condition timestamp)
//
// Implementation follows a hybrid approach:
// - VMs that never registered in K8s are directly deleted
// - VMs with K8s nodes (unreachable/NotReady) are logged only (autoscaler handles deletion)
// If Kubernetes nodes cannot be listed, cleanup is skipped until the next refresh.
// Registered-node checks skip missing or invalid timestamps rather than using node creation time.
// Azure operations use the cache's effective resource group, including the hosted-resource-group override.
// Deallocated and deallocating VMs are excluded, whether or not a Kubernetes node exists.
// Direct deletion is skipped when the VM power state is missing or unrecognized.
// Age-based deletion also requires a full threshold of consecutive eligible observations.
// Registration, transitional states, missing instances, and observation errors reset that grace period.
// Observation history is in memory, so restarting CA grants a fresh grace period.
func (m *AzureManager) cleanupZombieNodes() error {
	if m.kubeClient == nil {
		m.resetZombieObservations()
		return fmt.Errorf("cannot clean up zombie nodes: Kubernetes client is unavailable")
	}

	ctx, cancel := getContextWithTimeout(vmssContextTimeout)
	defer cancel()

	nodeList, err := m.kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		m.resetZombieObservations()
		return fmt.Errorf("cannot clean up zombie nodes: failed to list Kubernetes nodes: %w", err)
	}

	nodes := make([]*apiv1.Node, len(nodeList.Items))
	for i := range nodeList.Items {
		nodes[i] = &nodeList.Items[i]
	}

	return m.cleanupZombieNodesWithContext(nodes, time.Now())
}

// cleanupZombieNodesWithContext is similar to cleanupZombieNodes but accepts Kubernetes nodes
// as context to help correlate VMSS instances with K8s nodes.
func (m *AzureManager) cleanupZombieNodesWithContext(nodes []*apiv1.Node, currentTime time.Time) error {
	m.zombieCleanupMutex.Lock()
	defer m.zombieCleanupMutex.Unlock()

	ctx, cancel := getContextWithTimeout(vmssContextTimeout)
	defer cancel()

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

	// Process only VMSS node groups registered with the autoscaler.
	zombiesByVMSS := make(map[string][]zombieInstance)
	minAgeDuration := time.Duration(m.config.ZombieMinAgeMinutes) * time.Minute
	observations := make(map[zombieObservationKey]time.Time)

	for _, nodeGroup := range m.getNodeGroups() {
		scaleSet, ok := nodeGroup.(*ScaleSet)
		if !ok {
			continue
		}
		vmssName := scaleSet.Name
		if vmssName == "" {
			continue
		}

		// Get all VMs in this VMSS with instance view
		vms, err := m.azClient.virtualMachineScaleSetVMsClient.ListVMInstanceView(
			ctx,
			m.azureCache.resourceGroup,
			vmssName,
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

			if isZombie && !hasK8sNode && ptr.Deref(vm.Properties.ProvisioningState, "") == "Succeeded" {
				key := zombieObservationKey{
					nodeGroupID: vmssName,
					providerID:  normalizeProviderID(ptr.Deref(vm.ID, "")),
					createdAt:   vm.Properties.TimeCreated.UTC(),
				}
				if key.providerID == "" {
					klog.V(4).Infof("Skipping instance %s: no VM identity for registration grace period", instanceID)
					continue
				}
				firstSeen, found := m.zombieObservations[key]
				if !found || firstSeen.After(currentTime) {
					firstSeen = currentTime
				}
				observations[key] = firstSeen
				if currentTime.Sub(firstSeen) <= minAgeDuration {
					klog.V(4).Infof("Skipping instance %s: registration grace period has not elapsed", instanceID)
					continue
				}
			}

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
	// Only observations from this scan survive. Failed VMSS lists, missing VMs,
	// deregistered groups, and recovered instances lose their timers.
	m.zombieObservations = observations

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
		instanceIDs := make([]*string, len(zombies))
		for i, zombie := range zombies {
			instanceIDs[i] = ptr.To(zombie.instanceID)
		}

		vmInstanceIDs := armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs{
			InstanceIDs: instanceIDs,
		}

		klog.V(2).Infof("Deleting %d unregistered zombie instance(s) from VMSS %s: %v",
			len(instanceIDs), vmssName, instanceIDs)

		// Delete VMSS instances
		poller, err := m.azClient.vmssClientForDelete.BeginDeleteInstances(
			ctx,
			m.azureCache.resourceGroup,
			vmssName,
			vmInstanceIDs,
			&armcompute.VirtualMachineScaleSetsClientBeginDeleteInstancesOptions{
				ForceDeletion: ptr.To(false),
			},
		)

		if err != nil {
			klog.Errorf("Failed to delete zombie instances from VMSS %s: %v", vmssName, err)
			continue
		}

		klog.V(2).Infof("Initiated deletion of %d zombie instance(s) from VMSS %s", len(zombies), vmssName)

		// Poll for completion in background
		if poller != nil {
			go func(vmssName string, ids []*string) {
				pollCtx, pollCancel := getContextWithTimeout(asyncContextTimeout)
				defer pollCancel()
				_, pollErr := poller.PollUntilDone(pollCtx, nil)
				if pollErr != nil {
					klog.Errorf("Zombie deletion polling failed for VMSS %s instances %v: %v", vmssName, ids, pollErr)
					return
				}
				klog.V(2).Infof("Zombie deletion completed for VMSS %s instances %v", vmssName, ids)
			}(vmssName, instanceIDs)
		}
	}

	return nil
}

// evaluateZombieStatus evaluates a VMSS VM to determine if it's a zombie
// Returns (isZombie bool, hasK8sNode bool, reason string)
func (m *AzureManager) evaluateZombieStatus(
	vm *armcompute.VirtualMachineScaleSetVM,
	k8sNodeMap map[string]*apiv1.Node,
	currentTime time.Time,
	minAgeDuration time.Duration,
) (bool, bool, string) {
	provisioningState := ""
	if vm.Properties != nil {
		provisioningState = ptr.Deref(vm.Properties.ProvisioningState, "")
	}
	powerState := ""
	var vmssAge *time.Duration

	if vm.Properties != nil && vm.Properties.TimeCreated != nil {
		age := currentTime.Sub(*vm.Properties.TimeCreated)
		vmssAge = &age
	}

	if vm.Properties != nil && vm.Properties.InstanceView != nil && vm.Properties.InstanceView.Statuses != nil {
		for _, status := range vm.Properties.InstanceView.Statuses {
			if status == nil {
				continue
			}
			code := ptr.Deref(status.Code, "")

			if strings.HasPrefix(code, "PowerState/") {
				powerState = strings.TrimPrefix(code, "PowerState/")
			}
		}
	}

	// Check VM extension status
	extensionsInstalled := false
	extensionsFailed := false
	if vm.Properties != nil && vm.Properties.InstanceView != nil && vm.Properties.InstanceView.Extensions != nil && len(vm.Properties.InstanceView.Extensions) > 0 {
		extensionsInstalled = true
		for _, ext := range vm.Properties.InstanceView.Extensions {
			if ext.Statuses != nil {
				for _, status := range ext.Statuses {
					code := ptr.Deref(status.Code, "")
					if (status.Level != nil && *status.Level == armcompute.StatusLevelTypesError) || code == "ProvisioningState/failed" {
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

	// Parked VMs may be intentionally retained for reuse even without a K8s node.
	if powerState == "deallocated" || powerState == "deallocating" {
		return false, hasK8sNode, ""
	}

	if !hasK8sNode {
		switch powerState {
		case "running", "stopped":
		case "starting", "stopping":
			klog.V(4).Infof("Skipping unregistered instance %s: transitional power state %q",
				ptr.Deref(vm.InstanceID, ""), powerState)
			return false, false, ""
		default:
			klog.V(4).Infof("Skipping unregistered instance %s: missing or unrecognized power state %q",
				ptr.Deref(vm.InstanceID, ""), powerState)
			return false, false, ""
		}
	}

	// SCENARIO 1: Extensions failed or never installed
	if !hasK8sNode && provisioningState == "Succeeded" && (!extensionsInstalled || extensionsFailed) {
		if vmssAge == nil {
			klog.V(4).Infof("Skipping instance %s: no TimeCreated yet (extensions installed=%v, failed=%v)",
				ptr.Deref(vm.InstanceID, ""), extensionsInstalled, extensionsFailed)
			return false, false, ""
		}
		if *vmssAge > minAgeDuration {
			if extensionsFailed {
				return true, false, fmt.Sprintf("Extensions FAILED, age: %.0f minutes", vmssAge.Minutes())
			}
			return true, false, fmt.Sprintf("Extensions NOT INSTALLED (flapping zombie), age: %.0f minutes", vmssAge.Minutes())
		}
	}

	// SCENARIO 2: Provisioning failed on an unregistered, stopped VM.
	if !hasK8sNode && provisioningState == "Failed" && powerState == "stopped" {
		return true, false, "Provisioning FAILED and VM is stopped"
	}

	// SCENARIO 3: VMSS succeeded but never registered in K8s
	if !hasK8sNode && provisioningState == "Succeeded" {
		if vmssAge == nil {
			klog.V(4).Infof("Skipping instance %s: no TimeCreated yet (provisioning succeeded, not in K8s)",
				ptr.Deref(vm.InstanceID, ""))
			return false, false, ""
		}
		if *vmssAge > minAgeDuration {
			return true, false, fmt.Sprintf("Never registered in K8s (AllocationFailed), age: %.0f minutes", vmssAge.Minutes())
		}
	}

	// SCENARIO 4: Has K8s node but with critical issues
	if hasK8sNode && k8sNode != nil {
		var readyCondition *apiv1.NodeCondition
		for i := range k8sNode.Status.Conditions {
			if k8sNode.Status.Conditions[i].Type == apiv1.NodeReady {
				readyCondition = &k8sNode.Status.Conditions[i]
				break
			}
		}
		if readyCondition != nil && readyCondition.Status == apiv1.ConditionTrue {
			return false, true, ""
		}

		hasUnreachableTaint := false
		var unhealthySince *metav1.Time
		for _, taint := range k8sNode.Spec.Taints {
			if taint.Key != "node.kubernetes.io/unreachable" {
				continue
			}
			hasUnreachableTaint = true
			if taint.TimeAdded == nil {
				continue
			}
			if taint.TimeAdded.IsZero() || taint.TimeAdded.Time.After(currentTime) {
				klog.V(4).Infof("Skipping registered node %s: unreachable taint timestamp is zero or in the future", k8sNode.Name)
				return false, true, ""
			}
			// Use the most recent timestamp when both unreachable taint effects are present.
			if unhealthySince == nil || taint.TimeAdded.Time.After(unhealthySince.Time) {
				unhealthySince = taint.TimeAdded
			}
		}

		if !hasUnreachableTaint && powerState != "running" {
			return false, true, ""
		}

		reason := "Node has unreachable taint"
		if unhealthySince == nil {
			if readyCondition == nil || (readyCondition.Status != apiv1.ConditionFalse && readyCondition.Status != apiv1.ConditionUnknown) {
				klog.V(4).Infof("Skipping registered node %s: no timestamped unreachable taint or unhealthy Ready condition", k8sNode.Name)
				return false, true, ""
			}
			unhealthySince = &readyCondition.LastTransitionTime
			if hasUnreachableTaint {
				reason = fmt.Sprintf("Node has unreachable taint; Ready condition has been %s", readyCondition.Status)
			} else {
				reason = fmt.Sprintf("Running VM; Ready condition has been %s", readyCondition.Status)
			}
		}
		if unhealthySince.IsZero() || unhealthySince.Time.After(currentTime) {
			klog.V(4).Infof("Skipping registered node %s: Ready-condition timestamp is zero or in the future", k8sNode.Name)
			return false, true, ""
		}
		unhealthyDuration := currentTime.Sub(unhealthySince.Time)
		if unhealthyDuration > minAgeDuration {
			return true, true, fmt.Sprintf("%s for %.0f minutes", reason, unhealthyDuration.Minutes())
		}
		return false, true, ""
	}

	// Not a zombie
	return false, false, ""
}

// normalizeProviderID normalizes Azure provider IDs for comparison
// Removes "azure://" prefix and ensures consistent format
func normalizeProviderID(providerID string) string {
	normalized := strings.ToLower(providerID)
	normalized = strings.TrimPrefix(normalized, "azure://")
	normalized = strings.ReplaceAll(normalized, "\\", "/")
	return normalized
}
