/*
Copyright 2016 The Kubernetes Authors.

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

package clusterapi

import (
	"fmt"
	"k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/scheduler/cache"
	schedulercache "k8s.io/kubernetes/pkg/scheduler/cache"
	"log"
	"sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
)

// ClusterapiNodeGroup implements NodeGroup interface.
type ClusterapiNodeGroup struct {
	machineManager    MachineManager
	machineDeployment *v1alpha1.MachineDeployment
	attrs             *MachineDeploymentAttrs
}

// NewClusterapiNodeGroup creates a ClusterapiNodeGroup
func NewClusterapiNodeGroup(machineManager MachineManager, machineDeployment *v1alpha1.MachineDeployment) *ClusterapiNodeGroup {
	attrs := GetMachineDeploymentAttrs(machineDeployment)
	if nil == attrs {
		// should never happen because MachineManager only hands out MachineDeployment with valid annotations
		log.Panicf("NewClusterapiNodeGroup called w/ attribute-less machineDeployment (%s)", machineDeployment.Name)
	}

	ng := &ClusterapiNodeGroup{
		machineManager:    machineManager,
		machineDeployment: machineDeployment,
		attrs:             attrs,
	}
	return ng
}

// MaxSize returns maximum size of the node group.
func (ng *ClusterapiNodeGroup) MaxSize() int {
	return ng.attrs.maxSize
}

// MinSize returns minimum size of the node group.
func (ng *ClusterapiNodeGroup) MinSize() int {
	return ng.attrs.minSize
}

// TargetSize returns the current target size of the node group. It is possible that the
// number of nodes in Kubernetes is different at the moment but should be equal
// to Size() once everything stabilizes (new nodes finish startup and registration or
// removed nodes are deleted completely).
func (ng *ClusterapiNodeGroup) TargetSize() (int, error) {
	replicas := ng.machineDeployment.Spec.Replicas
	if nil == replicas {
		return 0, fmt.Errorf("replica count unset for MachineDeployment: %s", ng.machineDeployment.Name)
	}
	return int(*replicas), nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated.
func (ng *ClusterapiNodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("ClusterapiNodeGroup size increase size must be positive")
	}
	size, err := ng.TargetSize()
	if err != nil {
		return err
	}
	if size+delta > ng.MaxSize() {
		return fmt.Errorf("ClusterapiNodeGroup size increase too large - desired:%d max:%d", size+delta, ng.MaxSize())
	}
	return ng.machineManager.SetDeploymentSize(ng.machineDeployment, size+delta)
	// TODO interface documentation: "This function should wait until node group size is updated"
	//  have we fulfilled that?
}

// DeleteNodes deletes nodes from this node group. Error is returned either on
// failure or if the given node doesn't belong to this node group. This function
// should wait until node group size is updated.
func (ng *ClusterapiNodeGroup) DeleteNodes([]*v1.Node) error {
	// TODO waiting for https://github.com/kubernetes-sigs/cluster-api/pull/513
	klog.Info("ClusterapiNodeGroup.DeleteNodes not implemented")
	return nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target.
func (ng *ClusterapiNodeGroup) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("ClusterapiNodeGroup size decrease size must be negative")
	}

	size, err := ng.TargetSize()
	if err != nil {
		return err
	}
	nodes := ng.machineManager.NodesForDeployment(ng.machineDeployment)
	if nodes == nil {
		return fmt.Errorf("ClusterapiNodeGroup MachineDeployment not found: %s", ng.machineDeployment.Name)
	}
	if size+delta < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			size, delta, len(nodes))
	}
	return ng.machineManager.SetDeploymentSize(ng.machineDeployment, size+delta)
	// TODO interface documentation: "This function should wait until node group size is updated"
	//  have we fulfilled that?
}

// Id returns an unique identifier of the node group.
func (ng *ClusterapiNodeGroup) Id() string {
	return ng.machineDeployment.Name
}

// Debug returns a string containing all information regarding this node group.
func (ng *ClusterapiNodeGroup) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", ng.Id(), ng.MinSize(), ng.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
func (ng *ClusterapiNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	nodes := ng.machineManager.NodesForDeployment(ng.machineDeployment)
	if nodes == nil {
		klog.Infof("Empty ClusterapiNodeGroup: MachineDeployment %s", ng.machineDeployment.Name)
		return []cloudprovider.Instance{}, nil
	}

	result := make([]cloudprovider.Instance, len(nodes))
	for i, node := range nodes {
		result[i] = cloudprovider.Instance{
			Id: string(node.Spec.ProviderID),
		}
	}
	return result, nil
}

// TemplateNodeInfo returns a schedulercache.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The returned
// NodeInfo is expected to have a fully populated Node object, with all of the labels,
// capacity and allocatable information as well as all pods that are started on
// the node by default, using manifest (most likely only kube-proxy).
func (ng *ClusterapiNodeGroup) TemplateNodeInfo() (*cache.NodeInfo, error) {
	node, err := buildNodeFromOpenstackMachineDeployment(ng.machineDeployment)
	if err != nil {
		return nil, err
	}

	nodeInfo := schedulercache.NewNodeInfo(cloudprovider.BuildKubeProxy(ng.machineDeployment.Name))
	nodeInfo.SetNode(node)
	return nodeInfo, nil
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
func (ng *ClusterapiNodeGroup) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (ng *ClusterapiNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrAlreadyExist
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
func (ng *ClusterapiNodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An autoprovisioned group
// was created by CA and can be deleted when scaled to 0.
func (ng *ClusterapiNodeGroup) Autoprovisioned() bool {
	return false
}
