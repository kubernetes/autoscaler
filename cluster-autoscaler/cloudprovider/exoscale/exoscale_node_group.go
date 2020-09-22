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

package exoscale

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/exoscale/egoscale"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/k8s.io/klog"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework/v1alpha1"
)

// NodeGroup implements cloudprovider.NodeGroup interface. NodeGroup contains
// configuration info and functions to control a set of nodes that have the
// same capacity and set of labels.
type NodeGroup struct {
	id           string
	instancePool *egoscale.InstancePool
	manager      *Manager

	sync.Mutex
}

// MaxSize returns maximum size of the node group.
func (n *NodeGroup) MaxSize() int {
	limit, err := n.manager.computeInstanceLimit()
	if err != nil {
		return 0
	}

	return limit
}

// MinSize returns minimum size of the node group.
func (n *NodeGroup) MinSize() int {
	return 1
}

// TargetSize returns the current target size of the node group. It is possible that the
// number of nodes in Kubernetes is different at the moment but should be equal
// to Size() once everything stabilizes (new nodes finish startup and registration or
// removed nodes are deleted completely). Implementation required.
func (n *NodeGroup) TargetSize() (int, error) {
	return n.instancePool.Size, nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (n *NodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("delta must be positive, have: %d", delta)
	}

	targetSize := n.instancePool.Size + delta

	if targetSize > n.MaxSize() {
		return fmt.Errorf("size increase is too large. current: %d desired: %d max: %d",
			n.instancePool.Size, targetSize, n.MaxSize())
	}

	ctx := context.Background()

	klog.V(4).Infof("Scaling Instance Pool %s to %d", n.instancePool.ID, targetSize)

	_, err := n.manager.client.RequestWithContext(ctx, egoscale.ScaleInstancePool{
		ID:     n.instancePool.ID,
		ZoneID: n.instancePool.ZoneID,
		Size:   targetSize,
	})
	if err != nil {
		return err
	}

	if err := n.waitInstancePoolRunning(ctx); err != nil {
		return err
	}

	n.instancePool.Size = targetSize

	return nil
}

// DeleteNodes deletes nodes from this node group. Error is returned either on
// failure or if the given node doesn't belong to this node group. This function
// should wait until node group size is updated. Implementation required.
func (n *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	var instanceIDs []egoscale.UUID
	for _, node := range nodes {
		nodeID := node.Spec.ProviderID

		uuid, err := egoscale.ParseUUID(toNodeID(nodeID))
		if err != nil {
			return err
		}

		instanceIDs = append(instanceIDs, *uuid)
	}

	ctx := context.Background()

	n.Lock()
	defer n.Unlock()

	if err := n.waitInstancePoolRunning(ctx); err != nil {
		return err
	}

	klog.V(4).Infof("Evicting Instance Pool %s members: %v", n.instancePool.ID, instanceIDs)

	err := n.manager.client.BooleanRequest(egoscale.EvictInstancePoolMembers{
		ID:        n.instancePool.ID,
		ZoneID:    n.instancePool.ZoneID,
		MemberIDs: instanceIDs,
	})
	if err != nil {
		return err
	}

	if err := n.waitInstancePoolRunning(ctx); err != nil {
		return err
	}

	n.instancePool.Size = n.instancePool.Size - len(instanceIDs)

	return nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (n *NodeGroup) DecreaseTargetSize(_ int) error {
	// Exoscale Instance Pools don't support down-sizing without deleting members,
	// so it is not possible to implement it according to the documented behavior.
	return nil
}

// Id returns an unique identifier of the node group.
func (n *NodeGroup) Id() string {
	return n.id
}

// Debug returns a string containing all information regarding this node group.
func (n *NodeGroup) Debug() string {
	return fmt.Sprintf("Node group ID: %s (min:%d max:%d)", n.Id(), n.MinSize(), n.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
// It is required that Instance objects returned by this method have Id field set.
// Other fields are optional.
// This list should include also instances that might have not become a kubernetes node yet.
func (n *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	if n.instancePool == nil {
		return nil, errors.New("instance pool instance is not created")
	}

	instances := make([]cloudprovider.Instance, 0, len(n.instancePool.VirtualMachines))
	for _, vm := range n.instancePool.VirtualMachines {
		instances = append(instances, toInstance(vm))
	}

	return instances, nil
}

// TemplateNodeInfo returns a schedulerframework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The returned
// NodeInfo is expected to have a fully populated Node object, with all of the labels,
// capacity and allocatable information as well as all pods that are started on
// the node by default, using manifest (most likely only kube-proxy). Implementation optional.
func (n *NodeGroup) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one. Implementation required.
func (n *NodeGroup) Exist() bool {
	return n.instancePool != nil
}

// Create creates the node group on the cloud provider side. Implementation optional.
func (n *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
// Implementation optional.
func (n *NodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An autoprovisioned group
// was created by CA and can be deleted when scaled to 0.
func (n *NodeGroup) Autoprovisioned() bool {
	return false
}

// toInstance converts the given egoscale.VirtualMachine to a
// cloudprovider.Instance
func toInstance(vm egoscale.VirtualMachine) cloudprovider.Instance {
	return cloudprovider.Instance{
		Id:     toProviderID(vm.ID.String()),
		Status: toInstanceStatus(egoscale.VirtualMachineState(vm.State)),
	}
}

// toInstanceStatus converts the given egoscale.VirtualMachineState to a
// cloudprovider.InstanceStatus
func toInstanceStatus(vmState egoscale.VirtualMachineState) *cloudprovider.InstanceStatus {
	if vmState == "" {
		return nil
	}

	st := &cloudprovider.InstanceStatus{}
	switch vmState {
	case egoscale.VirtualMachineStarting:
		st.State = cloudprovider.InstanceCreating
	case egoscale.VirtualMachineRunning:
		st.State = cloudprovider.InstanceRunning
	case egoscale.VirtualMachineStopping:
		st.State = cloudprovider.InstanceDeleting
	default:
		st.ErrorInfo = &cloudprovider.InstanceErrorInfo{
			ErrorClass:   cloudprovider.OtherErrorClass,
			ErrorCode:    "no-code-exoscale",
			ErrorMessage: "error",
		}
	}

	return st
}

func (n *NodeGroup) waitInstancePoolRunning(ctx context.Context) error {
	err := n.poller(
		ctx,
		egoscale.GetInstancePool{ID: n.instancePool.ID, ZoneID: n.instancePool.ZoneID},
		func(i interface{}, err error) (bool, error) {
			if err != nil {
				return false, err
			}

			if i.(*egoscale.GetInstancePoolResponse).InstancePools[0].State ==
				egoscale.InstancePoolRunning {
				return true, nil
			}

			return false, nil
		},
	)

	return err
}

func (n *NodeGroup) poller(ctx context.Context, req egoscale.Command, callback func(interface{}, error) (bool, error)) error {
	timeout := time.Minute * 10
	c, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for t := time.Tick(time.Second * 10); ; { // nolint: staticcheck
		g, err := n.manager.client.RequestWithContext(c, req)
		ok, err := callback(g, err)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}

		select {
		case <-c.Done():
			return fmt.Errorf("context timeout after: %v", timeout)
		case <-t:
			continue
		}
	}
}
