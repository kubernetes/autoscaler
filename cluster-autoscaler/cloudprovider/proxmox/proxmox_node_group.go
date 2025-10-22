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

package proxmox

import (
	"context"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"
)

// NodeGroup implements cloudprovider.NodeGroup interface for Proxmox VMs
type NodeGroup struct {
	id          string
	manager     *Manager
	minSize     int
	maxSize     int
	proxmoxNode string
	templateID  int
	vmConfig    VMConfig

	// cache for VM instances
	instanceCache *InstanceCache
	// track VMs we've created (VM ID -> Provider ID mapping)
	createdVMs map[int]string
}

// MaxSize returns maximum size of the node group.
func (ng *NodeGroup) MaxSize() int {
	return ng.maxSize
}

// MinSize returns minimum size of the node group.
func (ng *NodeGroup) MinSize() int {
	return ng.minSize
}

// TargetSize returns the current target size of the node group.
func (ng *NodeGroup) TargetSize() (int, error) {
	return ng.instanceCache.Len(), nil
}

// IncreaseSize increases the size of the node group by creating new VMs.
func (ng *NodeGroup) IncreaseSize(delta int) error {
	klog.V(4).Infof("🚀 IncreaseSize called for node group %s with delta %d", ng.id, delta)

	if delta <= 0 {
		return fmt.Errorf("delta must be positive, have: %d", delta)
	}
	ctx := context.Background()
	currentSize := ng.instanceCache.Len()
	targetSize := currentSize + delta

	if targetSize > ng.MaxSize() {
		return fmt.Errorf("size increase is too large. current: %d desired: %d max: %d",
			currentSize, targetSize, ng.MaxSize())
	}

	// Create new VMs
	for range delta {
		vmID, err := ng.manager.client.GetNextFreeVMID(ctx)
		if err != nil {
			return fmt.Errorf("failed to get available VM ID: %v", err)
		}

		// Prepare VM config and populate cloud-init from secret if configured
		vmConfig := ng.vmConfig

		if err := ng.manager.client.CreateVM(ctx, ng.proxmoxNode, ng.templateID, vmID, vmConfig, ng.id); err != nil {
			return fmt.Errorf("failed to create VM %d: %v", vmID, err)
		}

		// Add to instances cache
		vm, err := ng.manager.client.GetVM(ctx, ng.proxmoxNode, vmID)
		if err != nil {
			return fmt.Errorf("failed to get VM %d: %v", vmID, err)
		}

		instance := cloudprovider.Instance{
			Id: toProviderID(vm.UUID),
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
			},
		}
		ng.instanceCache.Add(instance)

		klog.V(4).Infof("Created VM %d in node group %s", vmID, ng.id)
	}

	return nil
}

// AtomicIncreaseSize increases the size of the node group atomically.
func (ng *NodeGroup) AtomicIncreaseSize(delta int) error {
	klog.V(4).Infof("🔬 AtomicIncreaseSize called for node group %s with delta %d - returning ErrNotImplemented", ng.id, delta)
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes nodes from this node group.
func (ng *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	vms, err := ng.manager.client.GetVMs(context.Background(), ng.proxmoxNode)
	if err != nil {
		return fmt.Errorf("failed to get VMs for deletion: %v", err)
	}

	for _, node := range nodes {
		nodeUUID := toNodeID(node.Spec.ProviderID)

		// Get all VMs to find the one with matching UUID

		var targetVM *VM
		for _, vm := range vms {
			// Check if this VM has the matching UUID and belongs to our node group
			if vm.UUID == nodeUUID {
				tags := parseVMTags(vm.Tags)
				if nodeGroup, exists := tags["node-group"]; exists && nodeGroup == ng.id {
					targetVM = &vm
					break
				}
			}
		}

		if targetVM == nil {
			return fmt.Errorf("VM with UUID %s not found in node group %s", nodeUUID, ng.id)
		}

		// Stop and delete VM
		err = ng.manager.client.StopVM(context.Background(), ng.proxmoxNode, targetVM.ID)
		if err != nil {
			klog.Warningf("Failed to stop VM %d: %v", targetVM.ID, err)
		}

		klog.V(4).Infof("🗑️  Deleting VM %d (UUID: %s)", targetVM.ID, targetVM.UUID)
		err = ng.manager.client.DeleteVM(context.Background(), ng.proxmoxNode, targetVM.ID)
		if err != nil {
			return fmt.Errorf("failed to delete VM %d: %v", targetVM.ID, err)
		}

		ng.instanceCache.Remove(node.Spec.ProviderID)
	}

	return nil
}

// ForceDeleteNodes deletes nodes from this node group without constraints.
func (ng *NodeGroup) ForceDeleteNodes(nodes []*apiv1.Node) error {
	return nil
}

// DecreaseTargetSize decreases the target size of the node group.
func (ng *NodeGroup) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("delta must be negative, have: %d", delta)
	}

	currentSize := ng.instanceCache.Len()
	targetSize := currentSize + delta // delta is negative

	if targetSize < ng.MinSize() {
		klog.V(2).Infof("❌ Scale-down rejected: target size %d would be below minimum %d", targetSize, ng.MinSize())
		return fmt.Errorf("size decrease is too large. current: %d desired: %d min: %d",
			currentSize, targetSize, ng.MinSize())
	}

	removedInstance := ng.instanceCache.Pop()
	if removedInstance != nil {
		klog.V(4).Infof("remove node: %s from cache", removedInstance.Id)
	}

	return nil
}

func (ng *NodeGroup) Id() string {
	return ng.id
}

func (ng *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	instances := make([]cloudprovider.Instance, 0, ng.instanceCache.Len())
	instances = append(instances, ng.instanceCache.Items()...)
	return instances, nil
}

func (ng *NodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	// Create a basic node template based on VM configuration
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("proxmox-template-%s", ng.id),
			Labels: map[string]string{
				"kubernetes.io/os":               "linux",
				"kubernetes.io/arch":             "amd64",
				"cloud.proxmox.com/node-group":   ng.id,
				"cloud.proxmox.com/proxmox-node": ng.proxmoxNode,
				"cloud.proxmox.com/managed":      "true",
			},
		},
		Spec: apiv1.NodeSpec{
			ProviderID: toProviderID("template"),
		},
		Status: apiv1.NodeStatus{
			Capacity: apiv1.ResourceList{
				apiv1.ResourceCPU:    *newIntQuantity(int64(ng.vmConfig.Cores)),
				apiv1.ResourceMemory: *newIntQuantity(int64(ng.vmConfig.Memory) * 1024 * 1024), // Convert MB to bytes
				apiv1.ResourcePods:   *newIntQuantity(110),                                     // Default pod capacity
			},
			Allocatable: apiv1.ResourceList{
				apiv1.ResourceCPU:    *newIntQuantity(int64(ng.vmConfig.Cores)),
				apiv1.ResourceMemory: *newIntQuantity(int64(ng.vmConfig.Memory) * 1024 * 1024), // Convert MB to bytes
				apiv1.ResourcePods:   *newIntQuantity(110),                                     // Default pod capacity
			},
			Conditions: []apiv1.NodeCondition{
				{
					Type:   apiv1.NodeReady,
					Status: apiv1.ConditionTrue,
				},
			},
		},
	}

	nodeInfo := framework.NewNodeInfo(node, nil)
	return nodeInfo, nil
}

func (ng *NodeGroup) Exist() bool {
	// TODO: IMplement this, check with nodes vs nodes on pmox
	return true
}

// Create creates the node group on the cloud provider side.
func (ng *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	// Proxmox node groups are configuration-based, no need to create them explicitly
	return nil, nil
}

// Delete deletes the node group on the cloud provider side.
func (ng *NodeGroup) Delete() error {
	// Delete all VMs in the node group
	for _, instance := range ng.instanceCache.Items() {
		nodeID := toNodeID(instance.Id)
		vmID := 0
		if _, err := fmt.Sscanf(nodeID, "%d", &vmID); err != nil {
			klog.Warningf("Failed to parse VM ID from node ID %s: %v", nodeID, err)
			continue
		}

		err := ng.manager.client.StopVM(context.Background(), ng.proxmoxNode, vmID)
		if err != nil {
			klog.Warningf("Failed to stop VM %d: %v", vmID, err)
		}

		err = ng.manager.client.DeleteVM(context.Background(), ng.proxmoxNode, vmID)
		if err != nil {
			klog.Warningf("Failed to delete VM %d: %v", vmID, err)
		}
	}

	ng.instanceCache.Clear()
	return nil
}

func (ng *NodeGroup) Autoprovisioned() bool {
	// For now, all Proxmox node groups are manually configured
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular NodeGroup.
func (ng *NodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, nil
}

func (ng *NodeGroup) Debug() string {
	return fmt.Sprintf("node group id=%s, targetSize=%d, minSize=%d, maxSize=%d", ng.id, ng.instanceCache.Len(), ng.MinSize(), ng.MaxSize())
}

// refresh updates the node group's instance cache from Proxmox
func (ng *NodeGroup) refresh() error {
	klog.V(4).Infof("🔄 Refreshing node group %s on Proxmox node %s", ng.id, ng.proxmoxNode)
	ctx := context.Background()

	instances, err := ng.getCurrentInstances(ctx)
	if err != nil {
		return err
	}

	ng.instanceCache.Clear()
	ng.instanceCache.Add(instances...)
	return nil
}

func (ng *NodeGroup) getCurrentInstances(ctx context.Context) ([]cloudprovider.Instance, error) {
	vms, err := ng.manager.client.GetVMs(ctx, ng.proxmoxNode)
	if err != nil {
		return nil, fmt.Errorf("failed to get VMs for node %s: %v", ng.proxmoxNode, err)
	}

	// Find VMs that belong to this node group based on tags
	var instances []cloudprovider.Instance
	for _, vm := range vms {
		tags := parseVMTags(vm.Tags)
		nodeGroup, exists := tags["node-group"]
		if !exists || nodeGroup != ng.id {
			continue
		}

		state := cloudprovider.InstanceCreating

		providerID := toProviderID(vm.UUID)

		if vm.Status == "running" {
			state = cloudprovider.InstanceRunning
		}

		instance := cloudprovider.Instance{
			Id: providerID,
			Status: &cloudprovider.InstanceStatus{
				State: state,
			},
		}
		instances = append(instances, instance)
	}

	return instances, nil
}
