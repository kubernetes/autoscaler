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

package hetzner

import (
	"context"
	"errors"
	"fmt"
	"github.com/hetznercloud/hcloud-go/hcloud"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
	"strconv"
	"strings"
	"sync"
)

const (
	hcloudLabelNamespace = "hetzner.cloud"
	nodeIDLabel        = hcloudLabelNamespace + "/node-id"
)

var (
	// ErrNodePoolNotExist is return if no node pool exists for a given cluster ID
	ErrNodePoolNotExist = errors.New("node pool does not exist")
)

// NodeGroup implements cloudprovider.NodeGroup interface. NodeGroup contains
// configuration info and functions to control a set of nodes that have the
// same capacity and set of labels.
type NodeGroup struct {
	id      string
	manager *Manager
	size    int

	sync.Mutex
}

// MaxSize returns maximum size of the node group.
func (n *NodeGroup) MaxSize() int {
	return 100
}

// MinSize returns minimum size of the node group.
func (n *NodeGroup) MinSize() int {
	return 0
}

// TargetSize returns the current target size of the node group. It is possible
// that the number of nodes in Kubernetes is different at the moment but should
// be equal to Size() once everything stabilizes (new nodes finish startup and
// registration or removed nodes are deleted completely). Implementation
// required.
func (n *NodeGroup) TargetSize() (int, error) {
	return n.size, nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (n *NodeGroup) IncreaseSize(delta int) error {
	return nil
}

// DeleteNodes deletes nodes from this node group (and also increasing the size
// of the node group with that). Error is returned either on failure or if the
// given node doesn't belong to this node group. This function should wait
// until node group size is updated. Implementation required.
func (n *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	var instanceIDs []*hcloud.Server
	for _, node := range nodes {
		nodeID := node.Spec.ProviderID
		instanceIDs = append(instanceIDs, toServer(nodeID))
	}

	ctx := context.Background()

	n.Lock()
	defer n.Unlock()

	klog.V(4).Infof("Evicting servers: %v", instanceIDs)


	for _, node := range instanceIDs {
		_, err := n.manager.client.Server.Delete(ctx, node)
		klog.Fatalf("Failed to delete server ID %d error: %v", node.ID, err)
	}

	n.size = n.size - len(instanceIDs)
	return nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (n *NodeGroup) DecreaseTargetSize(delta int) error {
	n.size = n.size + delta
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
	listOptions := hcloud.ListOpts{
		PerPage: 50,
		LabelSelector: nodeIDLabel,
	}
	requestOptions := hcloud.ServerListOpts{ListOpts: listOptions}
	servers, err := n.manager.client.Server.AllWithOpts(context.Background(), requestOptions)
	if err != nil {
		klog.Fatalf("Failed to get servers for Hcloud: %v", err)
	}

	instances := make([]cloudprovider.Instance, 0, len(servers))
	for _, vm := range servers {
		instances = append(instances, toInstance(vm))
	}

	return instances, nil
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

func toInstance(vm *hcloud.Server) cloudprovider.Instance {
	return cloudprovider.Instance{
		Id:     toProviderID(vm.ID),
		Status: toInstanceStatus(vm.Status),
	}
}

func toProviderID(nodeID int) string {
	return fmt.Sprintf("%s%d", doProviderIDPrefix, nodeID)
}

func toServer(providerID string) *hcloud.Server {
	i, err := strconv.Atoi(strings.TrimPrefix(providerID, doProviderIDPrefix))
	if err != nil {
		klog.Fatalf("Failed to convert server ID %s error: %v", providerID, err)
	}
	return &hcloud.Server{ID: i}
}

func toInstanceStatus(status hcloud.ServerStatus) *cloudprovider.InstanceStatus {
	if status == "" {
		return nil
	}

	st := &cloudprovider.InstanceStatus{}
	switch status {
	case hcloud.ServerStatusInitializing:
	case hcloud.ServerStatusStarting:
		st.State = cloudprovider.InstanceCreating
	case hcloud.ServerStatusRunning:
		st.State = cloudprovider.InstanceRunning
	case hcloud.ServerStatusOff:
	case hcloud.ServerStatusDeleting:
	case hcloud.ServerStatusStopping:
		st.State = cloudprovider.InstanceDeleting
	default:
		st.ErrorInfo = &cloudprovider.InstanceErrorInfo{
			ErrorClass:   cloudprovider.OtherErrorClass,
			ErrorCode:    "no-code-hcloud",
			ErrorMessage: "error",
		}
	}

	return st
}
