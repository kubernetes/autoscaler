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
	"errors"
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/scaleway/scalewaygo"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog/v2"
)

// NodeGroup implements cloudprovider.NodeGroup interface.
// it is used to resize a Scaleway Pool which is a group of nodes with the same capacity.
type NodeGroup struct {
	scalewaygo.Client

	nodes map[string]*scalewaygo.Node
	specs *scalewaygo.GenericNodeSpecs
	p     *scalewaygo.Pool
}

// MaxSize returns maximum size of the node group.
func (ng *NodeGroup) MaxSize() int {
	klog.V(6).Info("MaxSize,called")

	return int(ng.p.MaxSize)
}

// MinSize returns minimum size of the node group.
func (ng *NodeGroup) MinSize() int {
	klog.V(6).Info("MinSize,called")

	return int(ng.p.MinSize)
}

// TargetSize returns the current target size of the node group. It is possible that the
// number of nodes in Kubernetes is different at the moment but should be equal
// to Size() once everything stabilizes (new nodes finish startup and registration or
// removed nodes are deleted completely).
func (ng *NodeGroup) TargetSize() (int, error) {
	klog.V(6).Info("TargetSize,called")
	return int(ng.p.Size), nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated.
func (ng *NodeGroup) IncreaseSize(delta int) error {

	klog.V(4).Infof("IncreaseSize,ClusterID=%s,delta=%d", ng.p.ClusterID, delta)

	if delta <= 0 {
		return fmt.Errorf("delta must be strictly positive, have: %d", delta)
	}

	targetSize := ng.p.Size + uint32(delta)

	if targetSize > uint32(ng.MaxSize()) {
		return fmt.Errorf("size increase is too large. current: %d desired: %d max: %d",
			ng.p.Size, targetSize, ng.MaxSize())
	}

	ctx := context.Background()
	pool, err := ng.UpdatePool(ctx, &scalewaygo.UpdatePoolRequest{
		PoolID: ng.p.ID,
		Size:   &targetSize,
	})
	if err != nil {
		return err
	}

	if pool.Size != targetSize {
		return fmt.Errorf("couldn't increase size to %d. Current size is: %d",
			targetSize, pool.Size)
	}

	ng.p.Size = targetSize
	return nil
}

// AtomicIncreaseSize is not implemented.
func (ng *NodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes nodes from this node group. Error is returned either on
// failure or if the given node doesn't belong to this node group. This function
// should wait until node group size is updated.
func (ng *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	ctx := context.Background()
	klog.V(4).Info("DeleteNodes,", len(nodes), " nodes to reclaim")
	for _, n := range nodes {

		node, ok := ng.nodes[n.Spec.ProviderID]
		if !ok {
			klog.Errorf("DeleteNodes,ProviderID=%s,PoolID=%s,node marked for deletion not found in pool", n.Spec.ProviderID, ng.p.ID)
			continue
		}

		updatedNode, err := ng.DeleteNode(ctx, &scalewaygo.DeleteNodeRequest{
			NodeID: node.ID,
		})
		if err != nil || updatedNode.Status != scalewaygo.NodeStatusDeleting {
			return err
		}

		ng.p.Size--
		ng.nodes[n.Spec.ProviderID].Status = scalewaygo.NodeStatusDeleting
	}

	return nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target.
func (ng *NodeGroup) DecreaseTargetSize(delta int) error {

	klog.V(4).Infof("DecreaseTargetSize,ClusterID=%s,delta=%d", ng.p.ClusterID, delta)

	if delta >= 0 {
		return fmt.Errorf("delta must be strictly negative, have: %d", delta)
	}

	targetSize := ng.p.Size + uint32(delta)
	if int(targetSize) < ng.MinSize() {
		return fmt.Errorf("size decrease is too large. current: %d desired: %d min: %d",
			ng.p.Size, targetSize, ng.MinSize())
	}

	ctx := context.Background()
	pool, err := ng.UpdatePool(ctx, &scalewaygo.UpdatePoolRequest{
		PoolID: ng.p.ID,
		Size:   &targetSize,
	})
	if err != nil {
		return err
	}

	if pool.Size != targetSize {
		return fmt.Errorf("couldn't decrease size to %d. Current size is: %d",
			targetSize, pool.Size)
	}

	ng.p.Size = targetSize
	return nil
}

// Id returns an unique identifier of the node group.
func (ng *NodeGroup) Id() string {
	return ng.p.ID
}

// Debug returns a string containing all information regarding this node group.
func (ng *NodeGroup) Debug() string {
	klog.V(4).Info("Debug,called")
	return fmt.Sprintf("id:%s,status:%s,version:%s,autoscaling:%t,size:%d,min_size:%d,max_size:%d", ng.Id(), ng.p.Status, ng.p.Version, ng.p.Autoscaling, ng.p.Size, ng.MinSize(), ng.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
func (ng *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	var nodes []cloudprovider.Instance

	klog.V(4).Info("Nodes,PoolID=", ng.p.ID)

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
// the node by default, using manifest (most likely only kube-proxy).
func (ng *NodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	klog.V(4).Infof("TemplateNodeInfo,PoolID=%s", ng.p.ID)
	node := apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ng.specs.Labels[apiv1.LabelHostname],
			Labels: ng.specs.Labels,
		},
		Status: apiv1.NodeStatus{
			Capacity:    apiv1.ResourceList{},
			Allocatable: apiv1.ResourceList{},
		},
	}
	node.Status.Capacity[apiv1.ResourceCPU] = *resource.NewQuantity(int64(ng.specs.CpuCapacity), resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceMemory] = *resource.NewQuantity(int64(ng.specs.MemoryCapacity), resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourceEphemeralStorage] = *resource.NewQuantity(int64(ng.specs.LocalStorageCapacity), resource.DecimalSI)
	node.Status.Capacity[apiv1.ResourcePods] = *resource.NewQuantity(int64(ng.specs.MaxPods), resource.DecimalSI)

	node.Status.Allocatable[apiv1.ResourceCPU] = *resource.NewQuantity(int64(ng.specs.CpuAllocatable), resource.DecimalSI)
	node.Status.Allocatable[apiv1.ResourceMemory] = *resource.NewQuantity(int64(ng.specs.MemoryAllocatable), resource.DecimalSI)
	node.Status.Allocatable[apiv1.ResourceEphemeralStorage] = *resource.NewQuantity(int64(ng.specs.LocalStorageAllocatable), resource.DecimalSI)
	node.Status.Allocatable[apiv1.ResourcePods] = *resource.NewQuantity(int64(ng.specs.MaxPods), resource.DecimalSI)

	if ng.specs.Gpu > 0 {
		nbGpu := *resource.NewQuantity(int64(ng.specs.Gpu), resource.DecimalSI)
		node.Status.Capacity[gpu.ResourceNvidiaGPU] = nbGpu
		node.Status.Allocatable[gpu.ResourceNvidiaGPU] = nbGpu
	}

	node.Status.Conditions = cloudprovider.BuildReadyConditions()
	node.Spec.Taints = parseTaints(ng.specs.Taints)

	nodeInfo := framework.NewNodeInfo(&node, nil, &framework.PodInfo{Pod: cloudprovider.BuildKubeProxy(ng.p.Name)})
	return nodeInfo, nil
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

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
func (ng *NodeGroup) Exist() bool {

	klog.V(4).Infof("Exist,PoolID=%s", ng.p.ID)

	_, err := ng.GetPool(context.Background(), &scalewaygo.GetPoolRequest{
		PoolID: ng.p.ID,
	})
	if err != nil && errors.Is(err, scalewaygo.ErrClientSide) {
		return false
	}
	return true

}

// Pool Autoprovision feature is not supported by Scaleway

// Create creates the node group on the cloud provider side.
func (ng *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
func (ng *NodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (ng *NodeGroup) Autoprovisioned() bool {
	return false
}

// GetOptions returns nil which means 'use defaults options'
func (ng *NodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// nodesFromPool returns the nodes associated to a Scaleway Pool
func nodesFromPool(client scalewaygo.Client, p *scalewaygo.Pool) (map[string]*scalewaygo.Node, error) {

	ctx := context.Background()
	resp, err := client.ListNodes(ctx, &scalewaygo.ListNodesRequest{ClusterID: p.ClusterID, PoolID: &p.ID})
	if err != nil {
		return nil, err
	}

	nodes := make(map[string]*scalewaygo.Node)
	for _, node := range resp.Nodes {
		nodes[node.ProviderID] = node
	}

	klog.V(4).Infof("nodesFromPool,PoolID=%s,%d nodes found", p.ID, len(nodes))

	return nodes, nil
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
