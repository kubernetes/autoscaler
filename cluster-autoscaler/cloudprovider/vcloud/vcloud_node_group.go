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

package vcloud

import (
	"context"
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"
)

const (
	vcloudLabelNamespace   = "k8s.io.infra.vnetwork.io"
	machineIDLabel         = vcloudLabelNamespace + "/machine-id"
	vcloudProviderIDPrefix = "vcloud://"
)

// NodeGroup implements cloudprovider.NodeGroup interface for VCloud
type NodeGroup struct {
	id        string
	clusterID string
	client    *VCloudAPIClient
	manager   *EnhancedManager

	minSize    int
	maxSize    int
	targetSize int
}

// MaxSize returns maximum size of the node group
func (n *NodeGroup) MaxSize() int {
	return n.maxSize
}

// MinSize returns minimum size of the node group
func (n *NodeGroup) MinSize() int {
	return n.minSize
}

// TargetSize returns the current target size of the node group
func (n *NodeGroup) TargetSize() (int, error) {
	return n.targetSize, nil
}

// IncreaseSize increases the size of the node group
func (n *NodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("delta must be positive, have: %d", delta)
	}

	currentSize, err := n.TargetSize()
	if err != nil {
		return fmt.Errorf("failed to get current size: %v", err)
	}

	targetSize := currentSize + delta
	if targetSize > n.MaxSize() {
		return fmt.Errorf("size increase is too large. current: %d desired: %d max: %d",
			currentSize, targetSize, n.MaxSize())
	}

	ctx := context.Background()
	err = n.client.ScaleNodePool(ctx, n.id, targetSize)
	if err != nil {
		return fmt.Errorf("failed to create instances: %v", err)
	}

	// Update local state immediately (like DigitalOcean)
	n.targetSize = targetSize
	return nil
}

// AtomicIncreaseSize is not implemented
func (n *NodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes nodes from this node group.
// This implementation follows the common pattern used by other cloud providers
// by deleting individual instances rather than scaling the pool.
func (n *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	ctx := context.Background()

	// Validate minimum size constraint before attempting any deletions
	currentSize, err := n.TargetSize()
	if err != nil {
		return fmt.Errorf("failed to get current size: %v", err)
	}

	newSize := currentSize - len(nodes)
	if newSize < n.MinSize() {
		return fmt.Errorf("cannot delete %d nodes: would violate minimum size constraint (min: %d, current: %d, after deletion: %d)",
			len(nodes), n.MinSize(), currentSize, newSize)
	}

	// Validate that all nodes belong to this node group
	nodeInstances, err := n.Nodes()
	if err != nil {
		return fmt.Errorf("failed to get node group instances: %v", err)
	}

	// Create a map of valid instance IDs for quick lookup
	validInstances := make(map[string]bool)
	for _, instance := range nodeInstances {
		// Extract instance ID from provider ID format
		if instanceID, err := fromProviderID(instance.Id); err == nil {
			validInstances[instanceID] = true
		}
	}

	// Extract and validate instance IDs from nodes
	var instancesToDelete []string
	for _, node := range nodes {
		instanceID, err := n.extractInstanceID(node)
		if err != nil {
			return fmt.Errorf("cannot extract instance ID from node %q: %v", node.Name, err)
		}

		// Verify node belongs to this node group
		if !validInstances[instanceID] {
			return fmt.Errorf("node %q (instance %s) does not belong to node group %s", node.Name, instanceID, n.id)
		}

		instancesToDelete = append(instancesToDelete, instanceID)
	}

	// Delete instances one by one (following common cloud provider pattern)
	var deletedCount int
	for _, instanceID := range instancesToDelete {
		err := n.client.DeleteInstance(ctx, n.id, instanceID)
		if err != nil {
			// If some instances were deleted but this one failed, log the partial success
			if deletedCount > 0 {
				klog.Warningf("Partially deleted %d out of %d instances before error", deletedCount, len(instancesToDelete))
			}
			return fmt.Errorf("failed to delete instance %s: %v", instanceID, err)
		}
		deletedCount++
		klog.V(2).Infof("Successfully deleted instance %s from node group %s", instanceID, n.id)
	}

	// Update local state to reflect deletions (like DigitalOcean)
	newTargetSize := currentSize - deletedCount
	n.targetSize = newTargetSize
	klog.Infof("Successfully deleted %d nodes from node group %s (new target size: %d)", deletedCount, n.id, newTargetSize)

	return nil
}

// ForceDeleteNodes deletes nodes from the group regardless of constraints.
// This implementation follows the common pattern used by other cloud providers.
func (n *NodeGroup) ForceDeleteNodes(nodes []*apiv1.Node) error {
	ctx := context.Background()

	// Extract instance IDs from nodes
	var instancesToDelete []string
	for _, node := range nodes {
		instanceID, err := n.extractInstanceID(node)
		if err != nil {
			return fmt.Errorf("cannot extract instance ID from node %q: %v", node.Name, err)
		}
		instancesToDelete = append(instancesToDelete, instanceID)
	}

	// Delete instances one by one (forced deletion ignores size constraints)
	var deletedCount int
	for _, instanceID := range instancesToDelete {
		err := n.client.DeleteInstance(ctx, n.id, instanceID)
		if err != nil {
			// Log partial success if some instances were deleted
			if deletedCount > 0 {
				klog.Warningf("Force deleted %d out of %d instances before error", deletedCount, len(instancesToDelete))
			}
			return fmt.Errorf("failed to force delete instance %s: %v", instanceID, err)
		}
		deletedCount++
		klog.V(2).Infof("Force deleted instance %s from node group %s", instanceID, n.id)
	}

	// Update local state to reflect forced deletions (like DigitalOcean)
	currentSize, _ := n.TargetSize()
	newTargetSize := currentSize - deletedCount
	n.targetSize = newTargetSize
	klog.Infof("Force deleted %d nodes from node group %s (new target size: %d)", deletedCount, n.id, newTargetSize)

	return nil
}

// DecreaseTargetSize decreases the target size of the node group
func (n *NodeGroup) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("delta must be negative, have: %d", delta)
	}

	currentSize, err := n.TargetSize()
	if err != nil {
		return fmt.Errorf("failed to get current size: %v", err)
	}

	targetSize := currentSize + delta
	if targetSize < n.MinSize() {
		return fmt.Errorf("size decrease is too small. current: %d desired: %d min: %d",
			currentSize, targetSize, n.MinSize())
	}

	// Update local state (like DigitalOcean)
	n.targetSize = targetSize
	return nil
}

// Id returns an unique identifier of the node group
func (n *NodeGroup) Id() string {
	return n.id
}

// Debug returns a string containing all information regarding this node group
func (n *NodeGroup) Debug() string {
	return fmt.Sprintf("vcloud node group %s (cluster: %s, min: %d, max: %d, target: %d)",
		n.id, n.clusterID, n.minSize, n.maxSize, n.targetSize)
}

// Nodes returns a list of all nodes that belong to this node group
func (n *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	var instanceIDs []string

	// Get instance IDs from the machines API
	ctx := context.Background()
	var apiErr error
	instanceIDs, apiErr = n.client.ListNodePoolInstances(ctx, n.id)
	if apiErr != nil {
		klog.V(2).Infof("Failed to get instances from machines API: %v", apiErr)
		instanceIDs = []string{} // Use final fallback
	} else {
		klog.V(4).Infof("Using machines API response: %d instances", len(instanceIDs))
	}

	var result []cloudprovider.Instance

	// If we got actual instance IDs, use them
	if len(instanceIDs) > 0 {
		klog.V(4).Infof("Using %d actual instance IDs for node group %s", len(instanceIDs), n.id)
		for _, instanceID := range instanceIDs {
			result = append(result, cloudprovider.Instance{
				Id:     toProviderID(instanceID),
				Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
			})
		}
	} else {
		// Final fallback: create instances based on target size with generated IDs
		klog.V(4).Infof("Using final fallback for node group %s", n.id)
		for i := 0; i < n.targetSize; i++ {
			instanceID := fmt.Sprintf("%s-instance-%d", n.id, i)
			result = append(result, cloudprovider.Instance{
				Id:     toProviderID(instanceID),
				Status: &cloudprovider.InstanceStatus{State: cloudprovider.InstanceRunning},
			})
		}
	}

	klog.V(4).Infof("Node group %s returning %d instances", n.id, len(result))
	for _, instance := range result {
		klog.V(5).Infof("  Instance: %s", instance.Id)
	}
	return result, nil
}

// TemplateNodeInfo returns a framework.NodeInfo structure of an empty node
func (n *NodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Exist checks if the node group really exists on the cloud provider side
func (n *NodeGroup) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side
func (n *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side
func (n *NodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned
func (n *NodeGroup) Autoprovisioned() bool {
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions for this node group
func (n *NodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	_ = defaults // Unused parameter, but required by interface
	return nil, cloudprovider.ErrNotImplemented
}

// Helper functions

// extractInstanceID extracts the instance ID from a Kubernetes node
func (n *NodeGroup) extractInstanceID(node *apiv1.Node) (string, error) {
	// Try to get instance ID from machine ID label first
	if machineID, ok := node.Labels[machineIDLabel]; ok {
		return machineID, nil
	}

	// Try to extract from provider ID
	if node.Spec.ProviderID != "" {
		return fromProviderID(node.Spec.ProviderID)
	}

	return "", fmt.Errorf("no instance ID found in node labels or provider ID")
}

// toProviderID converts an instance ID to a provider ID format
func toProviderID(instanceID string) string {
	return vcloudProviderIDPrefix + instanceID
}

// fromProviderID extracts instance ID from provider ID format
func fromProviderID(providerID string) (string, error) {
	if !strings.HasPrefix(providerID, vcloudProviderIDPrefix) {
		return "", fmt.Errorf("invalid provider ID format: %s", providerID)
	}
	return strings.TrimPrefix(providerID, vcloudProviderIDPrefix), nil
}

// toInstanceStatus converts VCloud instance state to cloudprovider.InstanceStatus
func toInstanceStatus(state string) *cloudprovider.InstanceStatus {
	status := &cloudprovider.InstanceStatus{}

	switch strings.ToLower(state) {
	case "creating", "pending", "provisioning":
		status.State = cloudprovider.InstanceCreating
	case "running", "active":
		status.State = cloudprovider.InstanceRunning
	case "deleting", "terminating", "destroyed":
		status.State = cloudprovider.InstanceDeleting
	default:
		status.ErrorInfo = &cloudprovider.InstanceErrorInfo{
			ErrorClass:   cloudprovider.OtherErrorClass,
			ErrorCode:    "unknown-state-vcloud",
			ErrorMessage: fmt.Sprintf("unknown instance state: %s", state),
		}
		status.State = cloudprovider.InstanceRunning // Default to running for unknown states
	}

	return status
}
