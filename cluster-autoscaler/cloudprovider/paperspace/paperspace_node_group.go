/*
Copyright 2019 The Kubernetes Authors.

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

package paperspace

import (
	"context"
	"errors"
	"fmt"

	psgo "github.com/Paperspace/paperspace-go"
	apiv1 "k8s.io/api/core/v1"
	schedulernodeinfo "k8s.io/kubernetes/pkg/scheduler/nodeinfo"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
)

var _ cloudprovider.NodeGroup = (*NodeGroup)(nil)

const (
	psksLabelNamespace = "paperspace.com"
	nodeIDLabel        = psksLabelNamespace + "/node-id"
	poolNameLabel      = psksLabelNamespace + "/pool-name"
	poolTypeLabel      = psksLabelNamespace + "/pool-type"
)

var (
	// ErrNodePoolNotExist is return if no node pool exists for a given cluster ID
	ErrNodePoolNotExist = errors.New("node pool does not exist")
)

// NodeGroup implements cloudprovider.NodeGroup interface. NodeGroup contains
// configuration info and functions to control a set of nodes that have the
// same capacity and set of labels.
type NodeGroup struct {
	id        string
	clusterID string
	manager   *Manager
	asg       psgo.AutoscalingGroup

	minSize int
	maxSize int
}

// MaxSize returns maximum size of the node group.
func (n *NodeGroup) MaxSize() int {
	return n.maxSize
}

// MinSize returns minimum size of the node group.
func (n *NodeGroup) MinSize() int {
	return n.minSize
}

// TargetSize returns the current target size of the node group. It is possible
// that the number of nodes in Kubernetes is different at the moment but should
// be equal to Size() once everything stabilizes (new nodes finish startup and
// registration or removed nodes are deleted completely). Implementation
// required.
func (n *NodeGroup) TargetSize() (int, error) {
	return n.asg.Current, nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (n *NodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("delta must be positive, have: %d", delta)
	}

	targetSize := n.asg.Current + delta

	if targetSize > n.MaxSize() {
		return fmt.Errorf("size increase is too large. current: %d desired: %d max: %d",
			n.asg.Current, targetSize, n.MaxSize())
	}

	ctx := context.Background()
	req := psgo.AutoscalingGroupUpdateParams{
		RequestParams: psgo.RequestParams{Context: ctx},
		Attributes: psgo.AutoscalingGroupUpdateAttributeParams{
			Current: targetSize,
		},
	}

	err := n.manager.client.UpdateAutoscalingGroup(n.id, req)
	if err != nil {
		return err
	}

	// update internal cache
	n.asg.Current = targetSize
	return nil
}

// DeleteNodes deletes nodes from this node group (and also increasing the size
// of the node group with that). Error is returned either on failure or if the
// given node doesn't belong to this node group. This function should wait
// until node group size is updated. Implementation required.
func (n *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	ctx := context.Background()
	for _, node := range nodes {
		nodeID := toNodeID(node.Spec.ProviderID)
		if nodeID == "" {
			// CA creates fake node objects to represent upcoming VMs that
			// haven't registered as nodes yet. We cannot delete the node at
			// this point.
			return fmt.Errorf("cannot delete node %q with provider ID %q on node pool %q: node ID label %q is missing", node.Name, node.Spec.ProviderID, n.id, nodeIDLabel)
		}

		err := n.manager.client.DeleteMachine(nodeID, psgo.MachineDeleteParams{RequestParams: psgo.RequestParams{Context: ctx}})
		if err != nil {
			return fmt.Errorf("deleting node failed for cluster: %q node pool: %q node: %q: %s",
				n.clusterID, n.id, nodeID, err)
		}

		// decrement the count by one  after a successful delete
		n.asg.Current--
	}

	return nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (n *NodeGroup) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("delta must be negative, have: %d", delta)
	}

	targetSize := n.asg.Current + delta
	if targetSize < n.MinSize() {
		return fmt.Errorf("size decrease is too small. current: %d desired: %d min: %d",
			n.asg.Current, targetSize, n.MinSize())
	}

	ctx := context.Background()
	req := psgo.AutoscalingGroupUpdateParams{
		RequestParams: psgo.RequestParams{Context: ctx},
		Attributes: psgo.AutoscalingGroupUpdateAttributeParams{
			Current: targetSize,
		},
	}

	err := n.manager.client.UpdateAutoscalingGroup(n.id, req)
	if err != nil {
		return err
	}

	// update internal cache
	n.asg.Current = targetSize
	return nil
}

// Id returns an unique identifier of the node group.
func (n *NodeGroup) Id() string {
	return n.id
}

// Debug returns a string containing all information regarding this node group.
func (n *NodeGroup) Debug() string {
	return fmt.Sprintf("cluster ID: %s (min:%d max:%d)", n.Id(), n.MinSize(), n.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.  It is
// required that Instance objects returned by this method have Id field set.
// Other fields are optional.
func (n *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	if !n.Exist() {
		return nil, errors.New("node pool instance is not created")
	}

	//TODO(arslan): after increasing a node pool, the number of nodes is not
	//anymore equal to the cache here. We should return a placeholder node for
	//that. As an example PR check this out:
	//https://github.com/kubernetes/autoscaler/pull/2235
	return toInstances(n.asg.Nodes), nil
}

// TemplateNodeInfo returns a schedulerframework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The
// returned NodeInfo is expected to have a fully populated Node object, with
// all of the labels, capacity and allocatable information as well as all pods
// that are started on the node by default, using manifest (most likely only
// kube-proxy). Implementation optional.
func (n *NodeGroup) TemplateNodeInfo() (*schedulernodeinfo.NodeInfo, error) {
	node, err := n.manager.buildNodeFromTemplate(n.asg)
	if err != nil {
		return nil, err
	}

	nodeInfo := schedulernodeinfo.NewNodeInfo(cloudprovider.BuildKubeProxy(n.asg.ID))
	nodeInfo.SetNode(node)
	return nodeInfo, nil
}

// Exist checks if the node group really exists on the cloud provider side.
// Allows to tell the theoretical node group from the real one. Implementation
// required.
func (n *NodeGroup) Exist() bool {
	return n.asg.ID != ""
}

// Create creates the node group on the cloud provider side. Implementation
// optional.
func (n *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.  This will be
// executed only for autoprovisioned node groups, once their size drops to 0.
// Implementation optional.
func (n *NodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An
// autoprovisioned group was created by CA and can be deleted when scaled to 0.
func (n *NodeGroup) Autoprovisioned() bool {
	return false
}

// toInstances converts a slice of *psgo.Machine to
// cloudprovider.Instance
func toInstances(nodes []psgo.Machine) []cloudprovider.Instance {
	instances := make([]cloudprovider.Instance, 0, len(nodes))
	for _, nd := range nodes {
		instances = append(instances, toInstance(nd))
	}
	return instances
}

// toInstance converts the given *godo.KubernetesNode to a
// cloudprovider.Instance
func toInstance(node psgo.Machine) cloudprovider.Instance {
	return cloudprovider.Instance{
		Id:     toProviderID(node.ID),
		Status: toInstanceStatus(node.State),
	}
}

// toInstanceStatus converts the given *godo.KubernetesNodeStatus to a
// cloudprovider.InstanceStatus
func toInstanceStatus(nodeState psgo.MachineState) *cloudprovider.InstanceStatus {
	if nodeState == "" {
		return nil
	}

	st := &cloudprovider.InstanceStatus{}
	switch nodeState {
	case psgo.MachineStateProvisioning:
		st.State = cloudprovider.InstanceCreating
	case psgo.MachineStateRunning:
		st.State = cloudprovider.InstanceRunning
	default:
		st.ErrorInfo = &cloudprovider.InstanceErrorInfo{
			ErrorClass:   cloudprovider.OtherErrorClass,
			ErrorCode:    "no-code-paperspace",
			ErrorMessage: string(nodeState),
		}
	}

	return st
}
