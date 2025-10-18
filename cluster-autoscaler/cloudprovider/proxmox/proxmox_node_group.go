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
	"sync"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
	vmIDStart   int
	vmIDEnd     int
	vmConfig    VMConfig

	// cache for VM instances
	instances []cloudprovider.Instance
	mutex     sync.RWMutex
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
	ng.mutex.RLock()
	defer ng.mutex.RUnlock()
	return len(ng.instances), nil
}

// IncreaseSize increases the size of the node group by creating new VMs.
func (ng *NodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("delta must be positive, have: %d", delta)
	}

	ng.mutex.Lock()
	defer ng.mutex.Unlock()

	currentSize := len(ng.instances)
	targetSize := currentSize + delta

	if targetSize > ng.MaxSize() {
		return fmt.Errorf("size increase is too large. current: %d desired: %d max: %d",
			currentSize, targetSize, ng.MaxSize())
	}

	// Create new VMs
	for i := 0; i < delta; i++ {
		vmID, err := ng.getNextAvailableVMID()
		if err != nil {
			return fmt.Errorf("failed to get available VM ID: %v", err)
		}

		err = ng.manager.client.CreateVM(context.Background(), ng.proxmoxNode, ng.templateID, vmID, ng.vmConfig)
		if err != nil {
			return fmt.Errorf("failed to create VM %d: %v", vmID, err)
		}

		// Start the VM
		err = ng.manager.client.StartVM(context.Background(), ng.proxmoxNode, vmID)
		if err != nil {
			klog.Warningf("Failed to start VM %d: %v", vmID, err)
		}

		// Add to instances cache
		instance := cloudprovider.Instance{
			Id: toProviderID(fmt.Sprintf("%d", vmID)),
			Status: &cloudprovider.InstanceStatus{
				State: cloudprovider.InstanceCreating,
			},
		}
		ng.instances = append(ng.instances, instance)

		klog.V(4).Infof("Created VM %d in node group %s", vmID, ng.id)
	}

	return nil
}

// AtomicIncreaseSize increases the size of the node group atomically.
func (ng *NodeGroup) AtomicIncreaseSize(delta int) error {
	return ng.IncreaseSize(delta)
}

// DeleteNodes deletes nodes from this node group.
func (ng *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	ng.mutex.Lock()
	defer ng.mutex.Unlock()

	for _, node := range nodes {
		nodeID := toNodeID(node.Spec.ProviderID)

		// Extract VM ID from node ID (assuming format like "100")
		vmID := 0
		if _, err := fmt.Sscanf(nodeID, "%d", &vmID); err != nil {
			return fmt.Errorf("failed to parse VM ID from node ID %s: %v", nodeID, err)
		}

		// Verify VM belongs to this node group
		if !ng.isVMInGroup(vmID) {
			return fmt.Errorf("VM %d does not belong to node group %s", vmID, ng.id)
		}

		// Stop and delete VM
		err := ng.manager.client.StopVM(context.Background(), ng.proxmoxNode, vmID)
		if err != nil {
			klog.Warningf("Failed to stop VM %d: %v", vmID, err)
		}

		err = ng.manager.client.DeleteVM(context.Background(), ng.proxmoxNode, vmID)
		if err != nil {
			return fmt.Errorf("failed to delete VM %d: %v", vmID, err)
		}

		// Remove from instances cache
		for i, instance := range ng.instances {
			if instance.Id == node.Spec.ProviderID {
				ng.instances = append(ng.instances[:i], ng.instances[i+1:]...)
				break
			}
		}

		klog.V(4).Infof("Deleted VM %d from node group %s", vmID, ng.id)
	}

	return nil
}

// ForceDeleteNodes deletes nodes from this node group without constraints.
func (ng *NodeGroup) ForceDeleteNodes(nodes []*apiv1.Node) error {
	return ng.DeleteNodes(nodes)
}

// DecreaseTargetSize decreases the target size of the node group.
func (ng *NodeGroup) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("delta must be negative, have: %d", delta)
	}

	ng.mutex.Lock()
	defer ng.mutex.Unlock()

	currentSize := len(ng.instances)
	targetSize := currentSize + delta // delta is negative

	if targetSize < ng.MinSize() {
		return fmt.Errorf("size decrease is too large. current: %d desired: %d min: %d",
			currentSize, targetSize, ng.MinSize())
	}

	// For simplicity, we don't actually delete VMs here, just update the cache
	// The actual deletion should be handled by DeleteNodes
	klog.V(4).Infof("Target size for node group %s decreased to %d", ng.id, targetSize)
	return nil
}

// Id returns an unique identifier of the node group.
func (ng *NodeGroup) Id() string {
	return ng.id
}

// Debug returns a string containing all information regarding this node group.
func (ng *NodeGroup) Debug() string {
	ng.mutex.RLock()
	defer ng.mutex.RUnlock()

	return fmt.Sprintf("NodeGroup{id=%s, minSize=%d, maxSize=%d, currentSize=%d, proxmoxNode=%s, templateID=%d}",
		ng.id, ng.minSize, ng.maxSize, len(ng.instances), ng.proxmoxNode, ng.templateID)
}

// Nodes returns a list of all nodes that belong to this node group.
func (ng *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	ng.mutex.RLock()
	defer ng.mutex.RUnlock()

	// Return a copy of the instances slice
	instances := make([]cloudprovider.Instance, len(ng.instances))
	copy(instances, ng.instances)
	return instances, nil
}

// TemplateNodeInfo returns a framework.NodeInfo structure of an empty node.
func (ng *NodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	// Create a basic node template based on VM configuration
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("proxmox-template-%s", ng.id),
			Labels: map[string]string{
				"kubernetes.io/os":                 "linux",
				"kubernetes.io/arch":               "amd64",
				"node.kubernetes.io/instance-type": fmt.Sprintf("proxmox-%dcpu-%dmb", ng.vmConfig.Cores, ng.vmConfig.Memory),
				"cloud.proxmox.com/node-group":     ng.id,
				"cloud.proxmox.com/proxmox-node":   ng.proxmoxNode,
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

// Exist checks if the node group really exists on the cloud provider side.
func (ng *NodeGroup) Exist() bool {
	// For now, assume all configured node groups exist
	return true
}

// Create creates the node group on the cloud provider side.
func (ng *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	// Proxmox node groups are configuration-based, no need to create them explicitly
	return ng, nil
}

// Delete deletes the node group on the cloud provider side.
func (ng *NodeGroup) Delete() error {
	ng.mutex.Lock()
	defer ng.mutex.Unlock()

	// Delete all VMs in the node group
	for _, instance := range ng.instances {
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

	ng.instances = nil
	return nil
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (ng *NodeGroup) Autoprovisioned() bool {
	// For now, all Proxmox node groups are manually configured
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular NodeGroup.
func (ng *NodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// refresh updates the node group's instance cache from Proxmox
func (ng *NodeGroup) refresh() error {
	ng.mutex.Lock()
	defer ng.mutex.Unlock()

	vms, err := ng.manager.client.GetVMs(context.Background(), ng.proxmoxNode)
	if err != nil {
		return fmt.Errorf("failed to get VMs for node %s: %v", ng.proxmoxNode, err)
	}

	// Filter VMs that belong to this node group
	var instances []cloudprovider.Instance
	for _, vm := range vms {
		if ng.isVMInGroup(vm.ID) {
			// Check if VM has the node group tag
			tags := parseVMTags(vm.Tags)
			if nodeGroup, exists := tags["node-group"]; exists && nodeGroup == ng.id {
				state := cloudprovider.InstanceRunning
				if vm.Status != "running" {
					state = cloudprovider.InstanceCreating
				}

				instance := cloudprovider.Instance{
					Id: toProviderID(fmt.Sprintf("%d", vm.ID)),
					Status: &cloudprovider.InstanceStatus{
						State: state,
					},
				}
				instances = append(instances, instance)
			}
		}
	}

	ng.instances = instances
	klog.V(4).Infof("Refreshed node group %s: found %d instances", ng.id, len(instances))
	return nil
}

// Helper function to create resource.Quantity from int64
func newIntQuantity(value int64) *resource.Quantity {
	q := resource.NewQuantity(value, resource.DecimalSI)
	return q
}
