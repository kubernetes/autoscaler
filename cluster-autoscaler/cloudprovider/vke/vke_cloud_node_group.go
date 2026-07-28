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

package vke

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"sync"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/vke/sdk"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

const providerIDPrefix = "openstack:///"

// NodeGroup implements cloudprovider.NodeGroup interface.
type NodeGroup struct {
	sdk.NodePool

	Manager     *VKEManager
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
	// Prefer the in-memory target size set by scale operations; otherwise use API current nodes.
	if ng.CurrentSize == -1 {
		klog.V(4).Infof("NodeGroup %s has a target size of %d", ng.ID, ng.CurrentNodes)
		return ng.CurrentNodes, nil
	}

	klog.V(4).Infof("NodeGroup %s has a target size of %d", ng.ID, ng.CurrentSize)
	return ng.CurrentSize, nil
}

// IncreaseSize increases node pool size.
func (ng *NodeGroup) IncreaseSize(delta int) error {
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
		return fmt.Errorf("node group size would be above maximum size - desired: %d, max: %d", size+delta, ng.MaxSize())
	}

	klog.V(4).Infof("Upscaling node pool %s to %d current nodes", ng.ID, size+delta)
	for i := 0; i < delta; i++ {
		_, err := ng.Manager.Client.AddNode(context.Background(), ng.Manager.ClusterID, ng.ID)
		if err != nil {
			return fmt.Errorf("failed to increase node pool desired size: %w", err)
		}
		ng.CurrentSize = size + i + 1
	}

	return nil
}

// DeleteNodes deletes the nodes from the group.
func (ng *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	// DeleteNodes is called in goroutine so it can run in parallel
	// Goroutines created in: ScaleDown.scheduleDeleteEmptyNodes()
	// Adding mutex to ensure CurrentSize attribute keeps consistency
	ng.mutex.Lock()
	defer ng.mutex.Unlock()

	klog.V(4).Infof("Deleting %d node(s)", len(nodes))

	// First, verify the NodeGroup can be decreased
	size, err := ng.TargetSize()
	if err != nil {
		return fmt.Errorf("failed to get NodeGroup target size")
	}

	if size-len(nodes) < ng.MinSize() {
		return fmt.Errorf("node group size would be below minimum size - desired: %d, min: %d", size-len(nodes), ng.MinSize())
	}

	for _, node := range nodes {
		shortID := strings.TrimPrefix(node.Spec.ProviderID, providerIDPrefix)
		err = ng.Manager.Client.DeleteNode(context.Background(), ng.Manager.ClusterID, ng.ID, shortID)
		if err != nil {
			return fmt.Errorf("failed to delete node %s: %w", shortID, err)
		}
	}

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
	nodes, err := ng.Manager.Client.ListNodePoolNodes(context.Background(), ng.Manager.ClusterID, ng.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list node pool nodes: %w", err)
	}

	klog.V(4).Infof("%d nodes are listed in node pool %s", len(nodes), ng.ID)

	// Cast all API nodes into instance interface
	instances := make([]cloudprovider.Instance, 0)
	for _, node := range nodes {
		instanceID := strings.TrimPrefix(node.Id, providerIDPrefix)
		instance := cloudprovider.Instance{
			Id:     providerIDPrefix + instanceID,
			Status: toInstanceStatus(node.Status),
		}

		instances = append(instances, instance)

		// Store the associated node group in cache for future reference
		ng.Manager.setNodeGroupPerProviderID(instance.Id, ng)
	}

	return instances, nil
}

// TemplateNodeInfo returns a node template for this node group.
func (ng *NodeGroup) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	// Forge node template in a node group
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("%v-node-%d", ng.Name, rand.Int63()),
			Labels:      map[string]string{},
			Annotations: map[string]string{},
			Finalizers:  []string{},
		},
		Spec: apiv1.NodeSpec{
			Taints: []apiv1.Taint{},
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
	nodeInfo := schedulerframework.NewNodeInfo(cloudprovider.BuildKubeProxy(ng.Id()))
	nodeInfo.SetNode(node)

	return nodeInfo, nil
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
func (ng *NodeGroup) Exist() bool {
	return ng.Id() != ""
}

// Create creates the node group on the cloud provider side.
func (ng *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	// Autoprovisioning is not supported for VKE yet.
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
func (ng *NodeGroup) Delete() error {
	// Autoprovisioning is not supported for VKE yet.
	return cloudprovider.ErrNotImplemented
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
	return nil, nil
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
