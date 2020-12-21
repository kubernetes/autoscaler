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
	"math/rand"
	"regexp"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/ovhcloud/sdk"
)

// instanceIdRegex defines the expression used for instance's ID
var instanceIdRegex = regexp.MustCompile(`^(.*)/(.*)$`)

// NodeGroup implements cloudprovider.NodeGroup interface.
type NodeGroup struct {
	sdk.NodePool

	Manager     *OvhCloudManager
	CurrentSize int
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

	klog.V(4).Infof("Increasing NodeGroup size with %d nodes", delta)

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

	// Eventually, call API to increase desired nodes number, automatically creating new nodes (wait for the pool to be READY before trying to update)
	if ng.Status == "READY" {
		resp, err := ng.Manager.Client.UpdateNodePool(context.Background(), ng.Manager.ProjectID, ng.Manager.ClusterID, ng.ID, &opts)
		if err != nil {
			return fmt.Errorf("failed to increase node pool desired size: %w", err)
		}

		ng.Status = resp.Status
	}

	return nil
}

// DeleteNodes deletes the nodes from the group.
func (ng *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	// Do not use node group which does not support autoscaling
	if !ng.Autoscale {
		return nil
	}

	klog.V(4).Infof("Deleting %d nodes", len(nodes))

	// First, verify the NodeGroup can be decreased
	size, err := ng.TargetSize()
	if err != nil {
		return fmt.Errorf("failed to get NodeGroup target size")
	}

	if size-len(nodes) < ng.MinSize() {
		return fmt.Errorf("node group size would be below minimum size - desired: %d, max: %d", size-len(nodes), ng.MinSize())
	}

	// Then, fetch node group instances and check that all nodes to remove are linked correctly,
	// otherwise it will return an error
	instances, err := ng.Nodes()
	if err != nil {
		return fmt.Errorf("failed to list current node group instances: %w", err)
	}

	nodeIds, err := extractNodeIds(nodes, instances, ng.Id())
	if err != nil {
		return fmt.Errorf("failed to extract node ids to remove: %w", err)
	}

	// Then, forge current size and parameters
	ng.CurrentSize = size - len(nodes)

	desired := uint32(ng.CurrentSize)
	opts := sdk.UpdateNodePoolOpts{
		DesiredNodes:  &desired,
		NodesToRemove: nodeIds,
	}

	// Eventually, call API to remove nodes from a NodeGroup (wait for the pool to be READY before trying to update)
	if ng.Status == "READY" {
		resp, err := ng.Manager.Client.UpdateNodePool(context.Background(), ng.Manager.ProjectID, ng.Manager.ClusterID, ng.ID, &opts)
		if err != nil {
			return fmt.Errorf("failed to delete node pool nodes: %w", err)
		}

		ng.Status = resp.Status
	}

	return nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes if the size
// when there is an option to just decrease the target.
func (ng *NodeGroup) DecreaseTargetSize(delta int) error {
	// Do not use node group which does not support autoscaling
	if !ng.Autoscale {
		return nil
	}

	klog.V(4).Infof("Decreasing NodeGroup size with %d nodes", delta)

	// First, verify the NodeGroup can be decreased
	if delta >= 0 {
		return fmt.Errorf("decrease size node group delta must be negative")
	}

	size, err := ng.TargetSize()
	if err != nil {
		return fmt.Errorf("failed to get NodeGroup target size")
	}

	if size+delta < ng.MinSize() {
		return fmt.Errorf("node group size would be below minimum size - desired: %d, max: %d", size+delta, ng.MinSize())
	}

	ng.CurrentSize = size + delta

	return nil
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

	// Cast all API nodes into instance interface
	instances := make([]cloudprovider.Instance, 0)
	for _, node := range nodes {
		instance := cloudprovider.Instance{
			Id:     fmt.Sprintf("%s/%s", node.ID, node.Name),
			Status: toInstanceStatus(node.Status),
		}

		instances = append(instances, instance)
	}

	return instances, nil
}

// TemplateNodeInfo returns a node template for this node group.
func (ng *NodeGroup) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	// Forge node template in a node group
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-node-%d", ng.Id(), rand.Int63()),
			Labels: map[string]string{
				NodePoolLabel: ng.Id(),
			},
		},
		Spec: apiv1.NodeSpec{},
		Status: apiv1.NodeStatus{
			Capacity:   apiv1.ResourceList{},
			Conditions: cloudprovider.BuildReadyConditions(),
		},
	}

	// Setup node info template
	nodeInfo := schedulerframework.NewNodeInfo(cloudprovider.BuildKubeProxy(ng.Id()))
	err := nodeInfo.SetNode(node)
	if err != nil {
		return nil, fmt.Errorf("failed to set up node info: %w", err)
	}

	return nodeInfo, nil
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
func (ng *NodeGroup) Exist() bool {
	return ng.Id() != ""
}

// Create creates the node group on the cloud provider side.
func (ng *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	klog.V(4).Infof("Creating a new NodeGroup")

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

// extractNodeIds find in an array of node resource their cloud instances IDs
func extractNodeIds(nodes []*apiv1.Node, instances []cloudprovider.Instance, groupLabel string) ([]string, error) {
	nodeIds := make([]string, 0)

	// Loop through each nodes to find any un-wanted nodes to remove
	for _, node := range nodes {
		// First, check if node resource has correct node pool label
		label, ok := node.Labels[NodePoolLabel]
		if !ok || label != groupLabel {
			return nil, fmt.Errorf("node %s without label %s, current: %s", node.Name, groupLabel, label)
		}

		// Then, loop through each group instances to find if the node is in the list
		found := false
		for _, instance := range instances {
			match := instanceIdRegex.FindStringSubmatch(instance.Id)
			if len(match) < 2 {
				continue
			}

			id := match[1]
			name := match[2]

			if node.Name != name {
				continue
			}

			found = true
			nodeIds = append(nodeIds, id)
			break
		}

		if !found {
			return nil, fmt.Errorf("node %s not found in group instances", node.Name)
		}
	}

	return nodeIds, nil
}

// toInstanceStatus cast a node status into an instance status
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
