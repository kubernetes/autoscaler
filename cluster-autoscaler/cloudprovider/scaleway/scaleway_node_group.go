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
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/scaleway/scalewaygo"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog/v2"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
	"strings"
)

var (
	ErrNodeNotInPool = errors.New("node marked for deletion and not found in pool")
)

type NodeGroup struct {
	scalewaygo.Client

	nodes map[string]*scalewaygo.Node
	specs *scalewaygo.GenericNodeSpecs
	p     *scalewaygo.Pool
}

func (ng *NodeGroup) MaxSize() int {
	klog.V(6).Info("MaxSize,called")

	return int(ng.p.MaxSize)
}

func (ng *NodeGroup) MinSize() int {
	klog.V(6).Info("MinSize,called")

	return int(ng.p.MinSize)
}

func (ng *NodeGroup) TargetSize() (int, error) {
	klog.V(6).Info("TargetSize,called")
	return int(ng.p.Size), nil
}

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

func (ng *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	ctx := context.Background()
	klog.V(4).Info("DeleteNodes,", len(nodes), " nodes to reclaim")
	for _, n := range nodes {

		node, ok := ng.nodes[n.Spec.ProviderID]
		if !ok {
			klog.Errorf("DeleteNodes,ProviderID=%s,PoolID=%s,%s", n.Spec.ProviderID, ng.p.ID, ErrNodeNotInPool)
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

func (ng *NodeGroup) Id() string {
	return ng.p.ID
}

func (ng *NodeGroup) Debug() string {
	klog.V(4).Info("Debug,called")
	return fmt.Sprintf("id:%s,status:%s,version:%s,autoscaling:%t,size:%d,min_size:%d,max_size:%d", ng.Id(), ng.p.Status, ng.p.Version, ng.p.Autoscaling, ng.p.Size, ng.MinSize(), ng.MaxSize())
}

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

// TemplateNodeInfo returns a schedulerframework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The returned
// NodeInfo is expected to have a fully populated Node object, with all of the labels,
// capacity and allocatable information as well as all pods that are started on
// the node by default, using manifest (most likely only kube-proxy).
func (ng *NodeGroup) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
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

	nodeInfo := schedulerframework.NewNodeInfo(cloudprovider.BuildKubeProxy(ng.p.Name))
	nodeInfo.SetNode(&node)
	return nodeInfo, nil
}

func parseTaints(taints map[string]string) []apiv1.Taint {
	k8s_taints := make([]apiv1.Taint, 0, len(taints))

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

		k8s_taints = append(k8s_taints, taint)
	}
	return k8s_taints
}

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

func (ng *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (ng *NodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

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
	case scalewaygo.NodeStatusCreating, scalewaygo.NodeStatusNotReady,
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
