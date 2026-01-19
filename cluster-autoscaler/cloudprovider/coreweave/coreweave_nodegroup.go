/*
Copyright 2025 The Kubernetes Authors.

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

package coreweave

import (
	"fmt"
	"maps"
	"math/rand"
	"sync"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog/v2"
)

// CoreWeaveNodeGroup represents a node group in the CoreWeave cloud provider.
type CoreWeaveNodeGroup struct {
	Name     string
	nodepool *CoreWeaveNodePool
	mutex    sync.Mutex
}

// NewCoreWeaveNodeGroup creates a new CoreWeaveNodeGroup instance.
// It takes a CoreWeaveNodePool as an argument and constructs the node group name
func NewCoreWeaveNodeGroup(nodepool *CoreWeaveNodePool) *CoreWeaveNodeGroup {
	return &CoreWeaveNodeGroup{
		Name:     nodepool.GetName(),
		nodepool: nodepool,
	}
}

// MaxSize returns the maximum size of the node group.
func (ng *CoreWeaveNodeGroup) MaxSize() int {
	return ng.nodepool.GetMaxSize()
}

// MinSize returns the minimum size of the node group.
func (ng *CoreWeaveNodeGroup) MinSize() int {
	return ng.nodepool.GetMinSize()
}

// TargetSize returns the current target size of the node group.
func (ng *CoreWeaveNodeGroup) TargetSize() (int, error) {
	return ng.nodepool.GetTargetSize(), nil
}

// IncreaseSize increases the size of the node group by the specified delta.
func (ng *CoreWeaveNodeGroup) IncreaseSize(delta int) error {
	klog.V(4).Infof("Increasing size of node group %s by %d", ng.Name, delta)
	return ng.nodepool.SetSize(ng.nodepool.GetTargetSize() + delta)
}

// AtomicIncreaseSize atomically increases the size of the node group by the specified delta.
// This method is not implemented for CoreWeaveNodeGroup.
func (ng *CoreWeaveNodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes the specified nodes from the node group.
// It verifies that the nodes belong to this node group and that deleting them does not violate the
// minimum size constraint. If the checks pass, it marks the nodes for removal and updates the
// target size of the node group accordingly.
// If any check fails, it returns an error with a descriptive message.
// If the nodes are successfully marked for removal, it returns nil.
func (ng *CoreWeaveNodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	ng.mutex.Lock()
	defer ng.mutex.Unlock()
	// Validate that the nodes belong to this node group
	err := ng.nodepool.ValidateNodes(nodes)
	if err != nil {
		return fmt.Errorf("some nodes do not belong to node group %s: %v", ng.Name, err)
	}
	//update target size
	if err := ng.nodepool.SetSize(ng.nodepool.GetTargetSize() - len(nodes)); err != nil {
		return fmt.Errorf("failed to update target size after marking nodes for removal: %v", err)
	}
	return nil
}

// ForceDeleteNodes is not implemented for CoreWeaveNodeGroup.
// node groups in CoreWeave do not support force deletion of nodes.
func (ng *CoreWeaveNodeGroup) ForceDeleteNodes(nodes []*apiv1.Node) error {
	return cloudprovider.ErrNotImplemented
}

// DecreaseTargetSize decreases the target size of the node group by the specified delta.
func (ng *CoreWeaveNodeGroup) DecreaseTargetSize(delta int) error {
	klog.V(4).Infof("Decreasing target size of node group %s by %d", ng.Name, delta)
	if delta < 0 {
		delta = -delta
	}
	return ng.nodepool.SetSize(ng.nodepool.GetTargetSize() - delta)
}

// Id returns the unique identifier of the node group.
func (ng *CoreWeaveNodeGroup) Id() string {
	return ng.nodepool.GetUID()
}

// Debug returns a debug string for the node group.
func (ng *CoreWeaveNodeGroup) Debug() string {
	return ""
}

// Nodes returns the list of nodes in the node group.
func (ng *CoreWeaveNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	// Return empty slice to avoid "not registered" warnings
	return []cloudprovider.Instance{}, nil
}

// TemplateNodeInfo returns a template NodeInfo for the node group.
// This is used by the autoscaler to simulate what a new node would look like
// when scaling from zero or when no nodes currently exist in the node group.
func (ng *CoreWeaveNodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	instanceTypeName := ng.nodepool.GetInstanceType()
	if instanceTypeName == "" {
		return nil, fmt.Errorf("node pool %s has no instance type defined", ng.Name)
	}

	instanceType, err := GetInstanceType(instanceTypeName)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance type info for %s: %v", instanceTypeName, err)
	}

	node, err := ng.buildNodeFromInstanceType(instanceTypeName, instanceType)
	if err != nil {
		return nil, fmt.Errorf("failed to build node from instance type: %v", err)
	}

	// The second parameter is for ResourceSlices when using DRA. CoreWeave only DRA for rack based instances which are
	// not supported by the Cluster Autoscaler at this time
	nodeInfo := framework.NewNodeInfo(node, nil)

	return nodeInfo, nil
}

// buildNodeFromInstanceType creates a template Node from the instance type and node pool configuration.
func (ng *CoreWeaveNodeGroup) buildNodeFromInstanceType(instanceTypeName string, instanceType *InstanceType) (*apiv1.Node, error) {
	nodeName := fmt.Sprintf("%s-template-%d", ng.Name, rand.Int63())

	capacity := ng.buildResourceList(instanceType)

	labels := ng.buildNodeLabels(nodeName, instanceTypeName, instanceType)

	taints := ng.nodepool.GetNodeTaints()

	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   nodeName,
			Labels: labels,
		},
		Status: apiv1.NodeStatus{
			// Capacity and Allocatable are set to the same value, ignoring system pods
			Capacity:    capacity,
			Allocatable: capacity,
			Conditions:  cloudprovider.BuildReadyConditions(),
		},
		Spec: apiv1.NodeSpec{
			Taints: taints,
		},
	}

	return node, nil
}

// buildResourceList creates a ResourceList from the instance type specifications.
func (ng *CoreWeaveNodeGroup) buildResourceList(instanceType *InstanceType) apiv1.ResourceList {
	resources := apiv1.ResourceList{}

	// CPU
	resources[apiv1.ResourceCPU] = *resource.NewQuantity(instanceType.VCPU, resource.DecimalSI)

	// Memory - stored in kibibytes (Ki), convert to bytes for template
	resources[apiv1.ResourceMemory] = *resource.NewQuantity(instanceType.MemoryKi*1024, resource.BinarySI)

	// Ephemeral storage - stored in kibibytes (Ki), convert to bytes for template
	if instanceType.EphemeralStorageKi > 0 {
		resources[apiv1.ResourceEphemeralStorage] = *resource.NewQuantity(instanceType.EphemeralStorageKi*1024, resource.BinarySI)
	}

	// GPU - use nvidia.com/gpu as the resource name
	if instanceType.GPU > 0 {
		resources[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(instanceType.GPU, resource.DecimalSI)
	}

	// Default to max of 110 pods if not specified (Kubernetes default)
	resources[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)
	if instanceType.MaxPods > 0 {
		resources[apiv1.ResourcePods] = *resource.NewQuantity(instanceType.MaxPods, resource.DecimalSI)
	}

	return resources
}

// buildNodeLabels creates the labels for a template node.
func (ng *CoreWeaveNodeGroup) buildNodeLabels(nodeName, instanceTypeName string, instanceType *InstanceType) map[string]string {
	labels := make(map[string]string)

	labels[apiv1.LabelInstanceTypeStable] = instanceTypeName
	labels[apiv1.LabelArchStable] = cloudprovider.DefaultArch
	if instanceType.Architecture != "" {
		labels[apiv1.LabelArchStable] = instanceType.Architecture
	}
	labels[apiv1.LabelOSStable] = cloudprovider.DefaultOS

	labels[coreWeaveNodePoolUID] = ng.nodepool.GetUID()
	labels[coreWeaveNodePoolName] = ng.nodepool.GetName()

	maps.Copy(labels, ng.nodepool.GetNodeLabels())

	return labels
}

// Exist checks if the node group exists.
func (ng *CoreWeaveNodeGroup) Exist() bool { return true }

// Create is not implemented for CoreWeaveNodeGroup.
func (ng *CoreWeaveNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete is not implemented for CoreWeaveNodeGroup.
func (ng *CoreWeaveNodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns whether the node group is autoprovisioned.
// In CoreWeave, node groups are not autoprovisioned, so it returns false
func (ng *CoreWeaveNodeGroup) Autoprovisioned() bool { return false }

// GetOptions returns the autoscaling options for the node group.
func (ng *CoreWeaveNodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}
