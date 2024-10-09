/*
Copyright 2020 The Kubernetes Authors.

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

package ovhcloud

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/ovhcloud/sdk"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
)

const providerIDPrefix = "openstack:///"

// NodeGroup implements cloudprovider.NodeGroup interface.
type NodeGroup struct {
	sdk.NodePool

	Manager     *OvhCloudManager
	CurrentSize int
	mutex       sync.Mutex
}

// MaxSize returns maximum size of the node pool.
func (ng *NodeGroup) MaxSize() int {
	return int(ng.MaxNodes)
}

// MinSize returns minimum size of the node pool.
func (ng *NodeGroup) MinSize() int {
	return int(ng.MinNodes)
}

// TargetSize returns the current TARGET size of the node pool. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (ng *NodeGroup) TargetSize() (int, error) {
	// By default, fetch the API desired nodes before using target size from autoscaler
	if ng.CurrentSize == -1 {
		return int(ng.DesiredNodes), nil
	}

	return ng.CurrentSize, nil
}

// IncreaseSize increases node pool size.
func (ng *NodeGroup) IncreaseSize(delta int) error {
	// Do not use node group which does not support autoscaling
	if !ng.Autoscale {
		return nil
	}

	klog.V(4).Infof("Increasing NodeGroup size by %d node(s)", delta)

	// First, verify the NodeGroup can be increased
	if delta <= 0 {
		return fmt.Errorf("increase size node group delta must be positive")
	}

	size, err := ng.TargetSize()
	if err != nil {
		return fmt.Errorf("failed to get NodeGroup target size")
	}

	if size+delta > ng.MaxSize() {
		return fmt.Errorf("node group size would be above minimum size - desired: %d, max: %d", size+delta, ng.MaxSize())
	}

	// Then, forge current size and parameters
	ng.CurrentSize = size + delta

	desired := uint32(ng.CurrentSize)
	opts := sdk.UpdateNodePoolOpts{
		DesiredNodes: &desired,
	}
	klog.V(4).Infof("Upscaling node pool %s to %d desired nodes", ng.ID, desired)

	// Call API to increase desired nodes number, automatically creating new nodes
	resp, err := ng.Manager.Client.UpdateNodePool(context.Background(), ng.Manager.ProjectID, ng.Manager.ClusterID, ng.ID, &opts)
	if err != nil {
		return fmt.Errorf("failed to increase node pool desired size: %w", err)
	}
	ng.Status = resp.Status

	return nil
}

// AtomicIncreaseSize is not implemented.
func (ng *NodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes the nodes from the group.
func (ng *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	// DeleteNodes is called in goroutine so it can run in parallel
	// Goroutines created in: ScaleDown.scheduleDeleteEmptyNodes()
	// Adding mutex to ensure CurrentSize attribute keeps consistency
	ng.mutex.Lock()
	defer ng.mutex.Unlock()

	// Do not use node group which does not support autoscaling
	if !ng.Autoscale {
		return nil
	}

	klog.V(4).Infof("Deleting %d node(s)", len(nodes))

	// First, verify the NodeGroup can be decreased
	size, err := ng.TargetSize()
	if err != nil {
		return fmt.Errorf("failed to get NodeGroup target size")
	}

	if size-len(nodes) < ng.MinSize() {
		return fmt.Errorf("node group size would be below minimum size - desired: %d, max: %d", size-len(nodes), ng.MinSize())
	}

	nodeProviderIds := make([]string, 0)
	for _, node := range nodes {
		nodeProviderIds = append(nodeProviderIds, node.Spec.ProviderID)
	}

	desired := uint32(size - len(nodes))
	opts := sdk.UpdateNodePoolOpts{
		DesiredNodes:  &desired,
		NodesToRemove: nodeProviderIds,
	}
	klog.V(4).Infof("Downscaling node pool %s to %d desired nodes by deleting the following nodes: %s", ng.ID, desired, nodeProviderIds)

	// Call API to remove nodes from a NodeGroup
	resp, err := ng.Manager.Client.UpdateNodePool(context.Background(), ng.Manager.ProjectID, ng.Manager.ClusterID, ng.ID, &opts)
	if err != nil {
		return fmt.Errorf("failed to delete node pool nodes: %w", err)
	}

	// Update the node group
	ng.Status = resp.Status
	ng.CurrentSize = size - len(nodes)

	return nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes if the size
// when there is an option to just decrease the target.
func (ng *NodeGroup) DecreaseTargetSize(delta int) error {
	// Cancellation of node provisioning is not supported yet
	return cloudprovider.ErrNotImplemented
}

// Id returns node pool id.
func (ng *NodeGroup) Id() string {
	return ng.Name
}

// Debug returns a debug string for the NodeGroup.
func (ng *NodeGroup) Debug() string {
	// Printing name (target size - min size - max size)
	return fmt.Sprintf("%s (%d:%d:%d)", ng.Id(), ng.CurrentSize, ng.MinSize(), ng.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
func (ng *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	// Fetch all nodes contained in the node group
	nodes, err := ng.Manager.Client.ListNodePoolNodes(context.Background(), ng.Manager.ProjectID, ng.Manager.ClusterID, ng.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list node pool nodes: %w", err)
	}

	klog.V(4).Infof("%d nodes are listed in node pool %s", len(nodes), ng.ID)

	// Cast all API nodes into instance interface
	instances := make([]cloudprovider.Instance, 0)
	for _, node := range nodes {
		instance := cloudprovider.Instance{
			Id:     fmt.Sprintf("%s%s", providerIDPrefix, node.InstanceID),
			Status: toInstanceStatus(node.Status),
		}

		instances = append(instances, instance)

		// Store the associated node group in cache for future reference
		ng.Manager.setNodeGroupPerProviderID(instance.Id, ng)
	}

	return instances, nil
}

// TemplateNodeInfo returns a node template for this node group.
func (ng *NodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	// Forge node template in a node group
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%s-node-%d", ng.Id(), rand.Int63()),
			Labels:      ng.Template.Metadata.Labels,
			Annotations: ng.Template.Metadata.Annotations,
			Finalizers:  ng.Template.Metadata.Finalizers,
		},
		Spec: apiv1.NodeSpec{
			Taints: ng.Template.Spec.Taints,
		},
		Status: apiv1.NodeStatus{
			Capacity:   apiv1.ResourceList{},
			Conditions: cloudprovider.BuildReadyConditions(),
		},
	}

	// Add the nodepool label
	if node.ObjectMeta.Labels == nil {
		node.ObjectMeta.Labels = make(map[string]string)
	}
	node.ObjectMeta.Labels[NodePoolLabel] = ng.Id()

	flavor, err := ng.Manager.getFlavorByName(ng.Flavor)
	if err != nil {
		return nil, fmt.Errorf("failed to get specs for flavor %q: %w", ng.Flavor, err)
	}

	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(110, resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(int64(flavor.VCPUs), resource.DecimalSI)
	node.Status.Capacity[gpu.ResourceNvidiaGPU] = *resource.NewQuantity(int64(flavor.GPUs), resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(int64(flavor.RAM)*int64(math.Pow(1024, 3)), resource.DecimalSI)

	node.Status.Allocatable = node.Status.Capacity

	// Setup node info template
	nodeInfo := framework.NewNodeInfo(node, nil, &framework.PodInfo{Pod: cloudprovider.BuildKubeProxy(ng.Id())})
	return nodeInfo, nil
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
func (ng *NodeGroup) Exist() bool {
	return ng.Id() != ""
}

// Create creates the node group on the cloud provider side.
func (ng *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	klog.V(4).Info("Creating a new NodeGroup")

	// Forge create node pool parameters (defaulting b2-7 for now)
	name := ng.Id()
	size := uint32(ng.CurrentSize)
	min := uint32(ng.MinSize())
	max := uint32(ng.MaxSize())

	opts := sdk.CreateNodePoolOpts{
		FlavorName:   "b2-7",
		Name:         &name,
		DesiredNodes: &size,
		MinNodes:     &min,
		MaxNodes:     &max,
		Autoscale:    true,
	}

	// Call API to add a node pool in the project/cluster
	np, err := ng.Manager.Client.CreateNodePool(context.Background(), ng.Manager.ProjectID, ng.Manager.ClusterID, &opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create node pool: %w", err)
	}

	// Forge a node group interface given the API response
	return &NodeGroup{
		NodePool:    *np,
		Manager:     ng.Manager,
		CurrentSize: int(ng.DesiredNodes),
	}, nil
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
func (ng *NodeGroup) Delete() error {
	klog.V(4).Infof("Deleting NodeGroup %s", ng.Id())

	// Call API to delete the node pool given its project and cluster
	_, err := ng.Manager.Client.DeleteNodePool(context.Background(), ng.Manager.ProjectID, ng.Manager.ClusterID, ng.ID)
	if err != nil {
		return fmt.Errorf("failed to delete node pool: %w", err)
	}

	return nil
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (ng *NodeGroup) Autoprovisioned() bool {
	// This is not handled yet.
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
func (ng *NodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	// If node group autoscaling options nil, return defaults
	if ng.Autoscaling == nil {
		return nil, nil
	}

	// Forge autoscaling configuration from node pool
	cfg := &config.NodeGroupAutoscalingOptions{
		ScaleDownUnneededTime: time.Duration(ng.Autoscaling.ScaleDownUnneededTimeSeconds) * time.Second,
		ScaleDownUnreadyTime:  time.Duration(ng.Autoscaling.ScaleDownUnreadyTimeSeconds) * time.Second,
	}

	// Switch utilization threshold from defaults given flavor type
	if ng.isGpu() {
		cfg.ScaleDownUtilizationThreshold = defaults.ScaleDownUtilizationThreshold
		cfg.ScaleDownGpuUtilizationThreshold = float64(ng.Autoscaling.ScaleDownUtilizationThreshold) // Use this one
	} else {
		cfg.ScaleDownUtilizationThreshold = float64(ng.Autoscaling.ScaleDownUtilizationThreshold) // Use this one
		cfg.ScaleDownGpuUtilizationThreshold = defaults.ScaleDownGpuUtilizationThreshold
	}

	return cfg, nil
}

// isGpu checks if a node group is using GPU machines
func (ng *NodeGroup) isGpu() bool {
	flavor, err := ng.Manager.getFlavorByName(ng.Flavor)
	if err != nil {
		// Fallback when we are unable to get the flavor: refer to the only category
		// known to be a GPU flavor category
		return strings.HasPrefix(ng.Flavor, GPUMachineCategory)
	}

	return flavor.GPUs > 0
}

// toInstanceStatus casts a node status into an instance status
func toInstanceStatus(status string) *cloudprovider.InstanceStatus {
	state := &cloudprovider.InstanceStatus{}

	switch status {
	case "INSTALLING", "REDEPLOYING":
		state.State = cloudprovider.InstanceCreating
	case "DELETING":
		state.State = cloudprovider.InstanceDeleting
	case "READY":
		state.State = cloudprovider.InstanceRunning
	default:
		state.ErrorInfo = &cloudprovider.InstanceErrorInfo{
			ErrorClass:   cloudprovider.OtherErrorClass,
			ErrorCode:    status,
			ErrorMessage: "error",
		}
	}

	return state
}
