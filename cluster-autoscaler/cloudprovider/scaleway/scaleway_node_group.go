/*
Copyright 2022 The Kubernetes Authors.

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

package scaleway

import (
	"context"
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/scaleway/scalewaygo"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/klog/v2"
)

// NodeGroup contains configuration info and functions to control a set
// of nodes that have the same capacity and set of labels.
type NodeGroup struct {
	scalewaygo.Client

	nodes map[string]*scalewaygo.Node
	pool  scalewaygo.Pool
}

// MaxSize returns maximum size of the node group.
func (ng *NodeGroup) MaxSize() int {
	klog.V(6).Info("MaxSize,called")
	return ng.pool.MaxSize
}

// MinSize returns minimum size of the node group.
func (ng *NodeGroup) MinSize() int {
	klog.V(6).Info("MinSize,called")
	return ng.pool.MinSize
}

// TargetSize returns the current target size of the node group. It is possible that the
// number of nodes in Kubernetes is different at the moment but should be equal
// to Size() once everything stabilizes (new nodes finish startup and registration or
// removed nodes are deleted completely). Implementation required.
func (ng *NodeGroup) TargetSize() (int, error) {
	klog.V(6).Info("TargetSize,called")
	return ng.pool.Size, nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (ng *NodeGroup) IncreaseSize(delta int) error {
	klog.V(4).Infof("IncreaseSize,ClusterID=%s,delta=%d", ng.pool.ClusterID, delta)

	if delta <= 0 {
		return fmt.Errorf("delta must be strictly positive, have: %d", delta)
	}

	targetSize := ng.pool.Size + delta

	if targetSize < 0 {
		return fmt.Errorf("size cannot be negative. current: %d delta: %d", ng.pool.Size, delta)
	}

	if targetSize > ng.MaxSize() {
		return fmt.Errorf("size increase is too large. current: %d desired: %d max: %d", ng.pool.Size, targetSize, ng.MaxSize())
	}

	updatedPool, err := ng.UpdatePool(context.Background(), ng.pool.ID, targetSize)
	if err != nil {
		return err
	}

	ng.pool.Size = updatedPool.Size
	ng.pool.Status = updatedPool.Status
	return nil
}

// AtomicIncreaseSize tries to increase the size of the node group atomically.
// It returns error if requesting the entire delta fails. The method doesn't wait until the new instances appear.
// Implementation is optional. Implementation of this method generally requires external cloud provider support
// for atomically requesting multiple instances. If implemented, CA will take advantage of the method while scaling up
// BestEffortAtomicScaleUp ProvisioningClass, guaranteeing that all instances required for such a
// ProvisioningRequest are provisioned atomically.
func (ng *NodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes nodes from this node group. Error is returned either on
// failure or if the given node doesn't belong to this node group. This function
// should wait until node group size is updated. Implementation required.
func (ng *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	ctx := context.Background()
	klog.V(4).Infof("DeleteNodes: %d nodes to reclaim", len(nodes))
	for _, n := range nodes {
		node, ok := ng.nodes[n.Spec.ProviderID]
		if !ok {
			klog.Errorf("DeleteNodes,ProviderID=%s,PoolID=%s,node marked for deletion not found in pool", n.Spec.ProviderID, ng.pool.ID)
			continue
		}

		deletedNode, err := ng.DeleteNode(ctx, node.ID)
		if err != nil {
			return err
		}

		ng.pool.Size--
		ng.nodes[n.Spec.ProviderID].Status = deletedNode.Status
	}

	return nil
}

// ForceDeleteNodes deletes nodes from this node group, without checking for
// constraints like minimal size validation etc. Error is returned either on
// failure or if the given node doesn't belong to this node group. This function
// should wait until node group size is updated.
func (ng *NodeGroup) ForceDeleteNodes(nodes []*apiv1.Node) error {
	return cloudprovider.ErrNotImplemented
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (ng *NodeGroup) DecreaseTargetSize(delta int) error {
	klog.V(4).Infof("DecreaseTargetSize,ClusterID=%s,delta=%d", ng.pool.ClusterID, delta)

	if delta >= 0 {
		return fmt.Errorf("delta must be strictly negative, have: %d", delta)
	}

	targetSize := ng.pool.Size + delta

	if targetSize < 0 {
		return fmt.Errorf("size cannot be negative. current: %d delta: %d", ng.pool.Size, delta)
	}

	if targetSize < ng.MinSize() {
		return fmt.Errorf("size decrease is too large. current: %d desired: %d min: %d", ng.pool.Size, targetSize, ng.MinSize())
	}

	ctx := context.Background()
	updatedNode, err := ng.UpdatePool(ctx, ng.pool.ID, targetSize)
	if err != nil {
		return err
	}

	ng.pool.Size = updatedNode.Size
	ng.pool.Status = scalewaygo.PoolStatusScaling

	return nil
}

// Id returns an unique identifier of the node group.
func (ng *NodeGroup) Id() string {
	return ng.pool.ID
}

// Debug returns a string containing all information regarding this node group.
func (ng *NodeGroup) Debug() string {
	klog.V(4).Info("Debug,called")
	return fmt.Sprintf("id:%s,status:%s,version:%s,autoscaling:%t,size:%d,min_size:%d,max_size:%d", ng.Id(), ng.pool.Status, ng.pool.Version, ng.pool.Autoscaling, ng.pool.Size, ng.MinSize(), ng.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
// It is required that Instance objects returned by this method have Id field set.
// Other fields are optional.
// This list should include also instances that might have not become a kubernetes node yet.
func (ng *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	klog.V(4).Infof("Nodes,PoolID=%s", ng.pool.ID)

	nodes := make([]cloudprovider.Instance, 0, len(ng.nodes))
	for _, node := range ng.nodes {
		nodes = append(nodes, cloudprovider.Instance{
			Id:     node.ProviderID,
			Status: fromScwStatus(node.Status),
		})
	}

	return nodes, nil
}

// TemplateNodeInfo returns a framework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The returned
// NodeInfo is expected to have a fully populated Node object, with all of the labels,
// capacity and allocatable information as well as all pods that are started on
// the node by default, using manifest (most likely only kube-proxy). Implementation optional.
func (ng *NodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	klog.V(4).Infof("TemplateNodeInfo,PoolID=%s", ng.pool.ID)
	node := apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ng.pool.Labels[apiv1.LabelHostname],
			Labels: ng.pool.Labels,
		},
		Status: apiv1.NodeStatus{
			Capacity:    apiv1.ResourceList{},
			Allocatable: apiv1.ResourceList{},
		},
	}

	for capacityName, capacityValue := range ng.pool.Capacity {
		node.Status.Capacity[apiv1.ResourceName(capacityName)] = *resource.NewQuantity(capacityValue, resource.DecimalSI)
	}

	for allocatableName, allocatableValue := range ng.pool.Allocatable {
		node.Status.Allocatable[apiv1.ResourceName(allocatableName)] = *resource.NewQuantity(allocatableValue, resource.DecimalSI)
	}

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	node.Spec.Taints = parseTaints(ng.pool.Taints)

	nodeInfo := framework.NewNodeInfo(&node, nil, framework.NewPodInfo(cloudprovider.BuildKubeProxy(ng.pool.Name), nil))

	return nodeInfo, nil
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one. Implementation required.
func (ng *NodeGroup) Exist() bool {
	klog.V(4).Infof("Exist,PoolID=%s", ng.pool.ID)
	return true
}

// Create creates the node group on the cloud provider side. Implementation optional.
func (ng *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
// Implementation optional.
func (ng *NodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An autoprovisioned group
// was created by CA and can be deleted when scaled to 0.
func (ng *NodeGroup) Autoprovisioned() bool {
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
// Implementation optional. Callers MUST handle `cloudprovider.ErrNotImplemented`.
func (ng *NodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func fromScwStatus(status scalewaygo.NodeStatus) *cloudprovider.InstanceStatus {
	st := &cloudprovider.InstanceStatus{}
	switch status {
	case scalewaygo.NodeStatusReady:
		st.State = cloudprovider.InstanceRunning
	case scalewaygo.NodeStatusCreating, scalewaygo.NodeStatusStarting,
		scalewaygo.NodeStatusRegistering, scalewaygo.NodeStatusNotReady,
		scalewaygo.NodeStatusUpgrading, scalewaygo.NodeStatusRebooting:
		st.State = cloudprovider.InstanceCreating
	case scalewaygo.NodeStatusDeleting:
		st.State = cloudprovider.InstanceDeleting
	case scalewaygo.NodeStatusCreationError:
		st.ErrorInfo = &cloudprovider.InstanceErrorInfo{
			ErrorCode:    string(scalewaygo.NodeStatusCreationError),
			ErrorMessage: "scaleway node could not be created",
		}
	case scalewaygo.NodeStatusDeleted:
		st.ErrorInfo = &cloudprovider.InstanceErrorInfo{
			ErrorCode:    string(scalewaygo.NodeStatusDeleted),
			ErrorMessage: "node has already been deleted",
		}
	case scalewaygo.NodeStatusLocked:
		st.ErrorInfo = &cloudprovider.InstanceErrorInfo{
			ErrorCode:    string(scalewaygo.NodeStatusLocked),
			ErrorMessage: "node is locked for legal reasons",
		}
	default:
		st.ErrorInfo = &cloudprovider.InstanceErrorInfo{
			ErrorCode: string(status),
		}
	}

	return st
}

func parseTaints(taints map[string]string) []apiv1.Taint {
	k8sTaints := make([]apiv1.Taint, 0, len(taints))
	for key, valueEffect := range taints {
		splittedValueEffect := strings.Split(valueEffect, ":")
		var taint apiv1.Taint

		switch apiv1.TaintEffect(splittedValueEffect[len(splittedValueEffect)-1]) {
		case apiv1.TaintEffectNoExecute:
			taint.Effect = apiv1.TaintEffectNoExecute
		case apiv1.TaintEffectNoSchedule:
			taint.Effect = apiv1.TaintEffectNoSchedule
		case apiv1.TaintEffectPreferNoSchedule:
			taint.Effect = apiv1.TaintEffectPreferNoSchedule
		default:
			continue
		}
		if len(splittedValueEffect) == 2 {
			taint.Value = splittedValueEffect[0]
		}
		taint.Key = key

		k8sTaints = append(k8sTaints, taint)
	}

	return k8sTaints
}
