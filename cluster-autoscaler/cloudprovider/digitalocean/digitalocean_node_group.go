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

package digitalocean

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/digitalocean/godo"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	utilerrors "k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/kubernetes/pkg/util/taints"
)

const (
	doksLabelNamespace = "doks.digitalocean.com"
	nodeIDLabel        = doksLabelNamespace + "/node-id"
)

var (
	// ErrNodePoolNotExist is return if no node pool exists for a given cluster ID
	ErrNodePoolNotExist = errors.New("node pool does not exist")
)

// NodeGroup implements cloudprovider.NodeGroup interface. NodeGroup contains
// configuration info and functions to control a set of nodes that have the
// same capacity and set of labels.
type NodeGroup struct {
	id               string
	clusterID        string
	client           nodeGroupClient
	nodePool         *godo.KubernetesNodePool
	nodePoolTemplate *godo.KubernetesNodePoolTemplate
	minSize          int
	maxSize          int
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
	return n.nodePool.Count, nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (n *NodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("delta must be positive, have: %d", delta)
	}

	targetSize := n.nodePool.Count + delta

	if targetSize > n.MaxSize() {
		return fmt.Errorf("size increase is too large. current: %d desired: %d max: %d",
			n.nodePool.Count, targetSize, n.MaxSize())
	}

	req := &godo.KubernetesNodePoolUpdateRequest{
		Count: &targetSize,
	}

	ctx := context.Background()
	updatedNodePool, _, err := n.client.UpdateNodePool(ctx, n.clusterID, n.id, req)
	if err != nil {
		return err
	}

	if updatedNodePool.Count != targetSize {
		return fmt.Errorf("couldn't increase size to %d (delta: %d). Current size is: %d",
			targetSize, delta, updatedNodePool.Count)
	}

	// update internal cache
	n.nodePool.Count = targetSize
	return nil
}

// AtomicIncreaseSize is not implemented.
func (n *NodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes nodes from this node group (and also increasing the size
// of the node group with that). Error is returned either on failure or if the
// given node doesn't belong to this node group. This function should wait
// until node group size is updated. Implementation required.
func (n *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	ctx := context.Background()
	for _, node := range nodes {
		nodeID, ok := node.Labels[nodeIDLabel]
		if !ok {
			// CA creates fake node objects to represent upcoming VMs that
			// haven't registered as nodes yet. We cannot delete the node at
			// this point.
			return fmt.Errorf("cannot delete node %q with provider ID %q on node pool %q: node ID label %q is missing", node.Name, node.Spec.ProviderID, n.id, nodeIDLabel)
		}

		_, err := n.client.DeleteNode(ctx, n.clusterID, n.id, nodeID, nil)
		if err != nil {
			return fmt.Errorf("deleting node failed for cluster: %q node pool: %q node: %q: %s",
				n.clusterID, n.id, nodeID, err)
		}

		// decrement the count by one  after a successful delete
		n.nodePool.Count--
	}

	return nil
}

// ForceDeleteNodes deletes nodes from the group regardless of constraints.
func (n *NodeGroup) ForceDeleteNodes(nodes []*apiv1.Node) error {
	return cloudprovider.ErrNotImplemented
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

	targetSize := n.nodePool.Count + delta
	if targetSize < n.MinSize() {
		return fmt.Errorf("size decrease is too small. current: %d desired: %d min: %d",
			n.nodePool.Count, targetSize, n.MinSize())
	}

	req := &godo.KubernetesNodePoolUpdateRequest{
		Count: &targetSize,
	}

	ctx := context.Background()
	updatedNodePool, _, err := n.client.UpdateNodePool(ctx, n.clusterID, n.id, req)
	if err != nil {
		return err
	}

	if updatedNodePool.Count != targetSize {
		return fmt.Errorf("couldn't increase size to %d (delta: %d). Current size is: %d",
			targetSize, delta, updatedNodePool.Count)
	}

	// update internal cache
	n.nodePool.Count = targetSize
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
	if n.nodePool == nil {
		return nil, errors.New("node pool instance is not created")
	}

	//TODO(arslan): after increasing a node pool, the number of nodes is not
	//anymore equal to the cache here. We should return a placeholder node for
	//that. As an example PR check this out:
	//https://github.com/kubernetes/autoscaler/pull/2235
	return toInstances(n.nodePool.Nodes), nil
}

// TemplateNodeInfo returns a framework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The
// returned NodeInfo is expected to have a fully populated Node object, with
// all of the labels, capacity and allocatable information as well as all pods
// that are started on the node by default, using manifest (most likely only
// kube-proxy). Implementation optional.
func (n *NodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	if n.nodePoolTemplate != nil {
		// Template has already been populated from cache - convert to node info and return
		tni, err := toNodeInfoTemplate(n.nodePoolTemplate)
		if err != nil {
			return nil, utilerrors.NewAutoscalerError(utilerrors.InternalError, err.Error())
		}
		return tni, nil
	}

	// No template present in cache - attempt to fetch from API
	resp, _, err := n.client.GetNodePoolTemplate(context.TODO(), n.clusterID, n.nodePool.Name)
	if err != nil {
		return nil, utilerrors.NewAutoscalerError(utilerrors.InternalError, err.Error())
	}
	return toNodeInfoTemplate(resp)
}

// Exist checks if the node group really exists on the cloud provider side.
// Allows to tell the theoretical node group from the real one. Implementation
// required.
func (n *NodeGroup) Exist() bool {
	return n.nodePool != nil
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

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
func (n *NodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// toInstances converts a slice of *godo.KubernetesNode to
// cloudprovider.Instance
func toInstances(nodes []*godo.KubernetesNode) []cloudprovider.Instance {
	instances := make([]cloudprovider.Instance, 0, len(nodes))
	for _, nd := range nodes {
		instances = append(instances, toInstance(nd))
	}
	return instances
}

// toInstance converts the given *godo.KubernetesNode to a
// cloudprovider.Instance
func toInstance(node *godo.KubernetesNode) cloudprovider.Instance {
	return cloudprovider.Instance{
		Id:     toProviderID(node.DropletID),
		Status: toInstanceStatus(node.Status),
	}
}

// toInstanceStatus converts the given *godo.KubernetesNodeStatus to a
// cloudprovider.InstanceStatus
func toInstanceStatus(nodeState *godo.KubernetesNodeStatus) *cloudprovider.InstanceStatus {
	if nodeState == nil {
		return nil
	}

	st := &cloudprovider.InstanceStatus{}
	switch nodeState.State {
	case "provisioning":
		st.State = cloudprovider.InstanceCreating
	case "running":
		st.State = cloudprovider.InstanceRunning
	case "draining", "deleting":
		st.State = cloudprovider.InstanceDeleting
	default:
		st.ErrorInfo = &cloudprovider.InstanceErrorInfo{
			ErrorClass:   cloudprovider.OtherErrorClass,
			ErrorCode:    "no-code-digitalocean",
			ErrorMessage: nodeState.Message,
		}
	}

	return st
}

func toNodeInfoTemplate(resp *godo.KubernetesNodePoolTemplate) (*framework.NodeInfo, error) {
	allocatable, err := parseToQuanitity(resp.Template.Allocatable.CPU, resp.Template.Allocatable.Pods, resp.Template.Allocatable.Memory)
	if err != nil {
		return nil, fmt.Errorf("failed to create allocatable resources - %s", err)
	}
	capacity, err := parseToQuanitity(resp.Template.Capacity.CPU, resp.Template.Capacity.Pods, resp.Template.Capacity.Memory)
	if err != nil {
		return nil, fmt.Errorf("failed to create capacity resources - %s", err)
	}
	addedTaints, _, err := taints.ParseTaints(resp.Template.Taints)
	if err != nil {
		return nil, fmt.Errorf("failed to parse taints from template - %s", err)
	}
	l := map[string]string{
		apiv1.LabelOSStable:   cloudprovider.DefaultOS,
		apiv1.LabelArchStable: cloudprovider.DefaultArch,
	}

	l = cloudprovider.JoinStringMaps(l, resp.Template.Labels)
	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   generateWorkerName(resp.Template.Name, rand.Int63()),
			Labels: l,
		},
		Spec: apiv1.NodeSpec{
			Taints: addedTaints,
		},
		Status: apiv1.NodeStatus{
			Allocatable: allocatable,
			Capacity:    capacity,
			Phase:       apiv1.NodeRunning,
			Conditions:  cloudprovider.BuildReadyConditions(),
		},
	}
	return framework.NewNodeInfo(node, nil), nil
}

func parseToQuanitity(cpu int64, pods int64, memory string) (apiv1.ResourceList, error) {
	c := resource.NewQuantity(cpu, resource.DecimalSI)
	p := resource.NewQuantity(pods, resource.DecimalSI)
	m, err := resource.ParseQuantity(memory)
	if err != nil {
		return nil, err
	}
	return apiv1.ResourceList{
		apiv1.ResourceCPU:    *c,
		apiv1.ResourceMemory: m,
		apiv1.ResourcePods:   *p,
	}, nil
}

func generateWorkerName(poolName string, workerID int64) string {
	var customAlphabet string = "n38uc7mqfyxojrbwgea6tl2ps5kh4ivd01z9"
	var customAlphabetSize int64 = int64(len(customAlphabet))
	s := ""
	for ; workerID > 0; workerID = workerID / customAlphabetSize {
		s = string(customAlphabet[workerID%customAlphabetSize]) + s
	}
	return fmt.Sprintf("%s-%s", poolName, s)
}
