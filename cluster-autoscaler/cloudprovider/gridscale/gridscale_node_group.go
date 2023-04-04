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

package gridscale

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gridscale/gsclient-go/v3"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

var (
	// ErrNodePoolNotExist is return if no node pool exists for a given cluster ID
	ErrNodePoolNotExist = errors.New("node pool does not exist")
)

// NodeGroup implements cloudprovider.NodeGroup interface. NodeGroup contains
// configuration info and functions to control a set of nodes that have the
// same capacity and set of labels.
type NodeGroup struct {
	id          string
	clusterUUID string
	client      nodeGroupClient
	nodeCount   int
	minSize     int
	maxSize     int
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
	return n.nodeCount, nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (n *NodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("delta must be positive, have: %d", delta)
	}

	targetSize := n.nodeCount + delta

	if targetSize > n.MaxSize() {
		return fmt.Errorf("size increase is too large. current: %d desired: %d max: %d",
			n.nodeCount, targetSize, n.MaxSize())
	}

	ctx := context.Background()
	updateRequestBody := gsclient.PaaSServiceUpdateRequest{
		Parameters: map[string]interface{}{
			"k8s_worker_node_count": targetSize,
		},
	}
	err := n.client.UpdatePaaSService(ctx, n.clusterUUID, updateRequestBody)
	if err != nil {
		return err
	}
	// update internal cache
	n.nodeCount = targetSize
	return nil
}

// DeleteNodes deletes nodes from this node group (and also increasing the size
// of the node group with that). Error is returned either on failure or if the
// given node doesn't belong to this node group. This function should wait
// until node group size is updated. Implementation required.
func (n *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	for _, node := range nodes {
		klog.V(4).Infof("Deleting node %s from node group", node.Name)
		klog.V(4).Infof("                   Node data: %v ", node)
	}

	targetSize := n.nodeCount - len(nodes)
	ctx := context.Background()
	updateRequestBody := gsclient.PaaSServiceUpdateRequest{
		Parameters: map[string]interface{}{
			"k8s_worker_node_count": targetSize,
		},
	}
	err := n.client.UpdatePaaSService(ctx, n.clusterUUID, updateRequestBody)
	if err != nil {
		return err
	}
	// update internal cache
	n.nodeCount = targetSize
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

	targetSize := n.nodeCount + delta
	if targetSize < n.MinSize() {
		return fmt.Errorf("size decrease is too small. current: %d desired: %d min: %d",
			n.nodeCount, targetSize, n.MinSize())
	}

	ctx := context.Background()
	updateRequestBody := gsclient.PaaSServiceUpdateRequest{
		Parameters: map[string]interface{}{
			"k8s_worker_node_count": targetSize,
		},
	}
	err := n.client.UpdatePaaSService(ctx, n.clusterUUID, updateRequestBody)
	if err != nil {
		return err
	}
	// update internal cache
	n.nodeCount = targetSize
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
	//TODO(arslan): after increasing a node pool, the number of nodes is not
	//anymore equal to the cache here. We should return a placeholder node for
	//that. As an example PR check this out:
	//https://github.com/kubernetes/autoscaler/pull/2235

	serverList, err := n.client.GetServerList(context.Background())
	if err != nil {
		return nil, err
	}
	var gskNodeList []gsclient.Server
	for _, server := range serverList {
		// skip master node
		if strings.Contains(server.Properties.Name, "master") {
			continue
		}
		// append nodes that have the label
		// #gsk#<clusterUUID>
		for _, label := range server.Properties.Labels {
			if label == fmt.Sprintf("#gsk#%s", n.clusterUUID) {
				gskNodeList = append(gskNodeList, server)
			}
		}
	}
	return toInstances(gskNodeList), nil
}

// TemplateNodeInfo returns a schedulerframework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The
// returned NodeInfo is expected to have a fully populated Node object, with
// all of the labels, capacity and allocatable information as well as all pods
// that are started on the node by default, using manifest (most likely only
// kube-proxy). Implementation optional.
func (n *NodeGroup) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Exist checks if the node group really exists on the cloud provider side.
// Allows to tell the theoretical node group from the real one. Implementation
// required.
func (n *NodeGroup) Exist() bool {
	return true
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

// toInstances converts a slice of gsclient.Server to
// cloudprovider.Instance
func toInstances(nodes []gsclient.Server) []cloudprovider.Instance {
	instances := make([]cloudprovider.Instance, 0, len(nodes))
	for _, nd := range nodes {
		instances = append(instances, toInstance(nd))
	}
	return instances
}

// toInstance converts the given gsclient.Server to a
// cloudprovider.Instance
func toInstance(node gsclient.Server) cloudprovider.Instance {
	return cloudprovider.Instance{
		Id:     toProviderID(node.Properties.ObjectUUID),
		Status: toInstanceStatus(node.Properties.Status),
	}
}

// toInstanceStatus converts the given node status to a
// cloudprovider.InstanceStatus
func toInstanceStatus(nodeStatus string) *cloudprovider.InstanceStatus {
	if nodeStatus == "" {
		return nil
	}

	st := &cloudprovider.InstanceStatus{}
	switch nodeStatus {
	case "provisioning":
		st.State = cloudprovider.InstanceCreating
	case "active":
		st.State = cloudprovider.InstanceRunning
	default:
		st.ErrorInfo = &cloudprovider.InstanceErrorInfo{
			ErrorClass:   cloudprovider.OtherErrorClass,
			ErrorCode:    "no-code-gridscale",
			ErrorMessage: fmt.Sprintf("unknown node status: %s", nodeStatus),
		}
	}

	return st
}
