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

package externalgrpc

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/externalgrpc/protos"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	klog "k8s.io/klog/v2"
)

// NodeGroup implements cloudprovider.NodeGroup interface. NodeGroup contains
// configuration info and functions to control a set of nodes that have the
// same capacity and set of labels.
type NodeGroup struct {
	id          string // this must be a stable identifier
	minSize     int    // cached value
	maxSize     int    // cached value
	debug       string // cached value
	client      protos.CloudProviderClient
	grpcTimeout time.Duration

	mutex    sync.Mutex
	nodeInfo **framework.NodeInfo // used to cache NodeGroupTemplateNodeInfo() grpc calls
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
	ctx, cancel := context.WithTimeout(context.Background(), n.grpcTimeout)
	defer cancel()
	klog.V(5).Infof("Performing gRPC call NodeGroupTargetSize for node group %v", n.id)
	res, err := n.client.NodeGroupTargetSize(ctx, &protos.NodeGroupTargetSizeRequest{
		Id: n.id,
	})
	if err != nil {
		klog.V(1).Infof("Error on gRPC call NodeGroupTargetSize: %v", err)
		return 0, err
	}
	return int(res.TargetSize), nil
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (n *NodeGroup) IncreaseSize(delta int) error {
	ctx, cancel := context.WithTimeout(context.Background(), n.grpcTimeout)
	defer cancel()
	klog.V(5).Infof("Performing gRPC call NodeGroupIncreaseSize for node group %v", n.id)
	_, err := n.client.NodeGroupIncreaseSize(ctx, &protos.NodeGroupIncreaseSizeRequest{
		Id:    n.id,
		Delta: int32(delta),
	})
	if err != nil {
		klog.V(1).Infof("Error on gRPC call NodeGroupIncreaseSize: %v", err)
		return err
	}
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
	pbNodes := make([]*protos.ExternalGrpcNode, 0)
	for _, n := range nodes {
		pbNodes = append(pbNodes, externalGrpcNode(n))
	}
	ctx, cancel := context.WithTimeout(context.Background(), n.grpcTimeout)
	defer cancel()
	klog.V(5).Infof("Performing gRPC call NodeGroupDeleteNodes for node group %v", n.id)
	_, err := n.client.NodeGroupDeleteNodes(ctx, &protos.NodeGroupDeleteNodesRequest{
		Id:    n.id,
		Nodes: pbNodes,
	})
	if err != nil {
		klog.V(1).Infof("Error on gRPC call NodeGroupDeleteNodes: %v", err)
		return err
	}
	return nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target. Implementation required.
func (n *NodeGroup) DecreaseTargetSize(delta int) error {
	ctx, cancel := context.WithTimeout(context.Background(), n.grpcTimeout)
	defer cancel()
	klog.V(5).Infof("Performing gRPC call NodeGroupDecreaseTargetSize for node group %v", n.id)
	_, err := n.client.NodeGroupDecreaseTargetSize(ctx, &protos.NodeGroupDecreaseTargetSizeRequest{
		Id:    n.id,
		Delta: int32(delta),
	})
	if err != nil {
		klog.V(1).Infof("Error on gRPC call NodeGroupDecreaseTargetSize: %v", err)
		return err
	}
	return nil
}

// Id returns an unique identifier of the node group.
func (n *NodeGroup) Id() string {
	return n.id
}

// Debug returns a string containing all information regarding this node group.
func (n *NodeGroup) Debug() string {
	return n.debug
}

// Nodes returns a list of all nodes that belong to this node group. It is
// required that Instance objects returned by this method have Id field set.
// Other fields are optional.
func (n *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), n.grpcTimeout)
	defer cancel()
	klog.V(5).Infof("Performing gRPC call NodeGroupNodes for node group %v", n.id)
	res, err := n.client.NodeGroupNodes(ctx, &protos.NodeGroupNodesRequest{
		Id: n.id,
	})
	if err != nil {
		klog.V(1).Infof("Error on gRPC call NodeGroupNodes: %v", err)
		return nil, err
	}
	instances := make([]cloudprovider.Instance, 0)
	for _, pbInstance := range res.GetInstances() {
		var instance cloudprovider.Instance
		instance.Id = pbInstance.GetId()
		pbStatus := pbInstance.GetStatus()
		if pbStatus.GetInstanceState() != protos.InstanceStatus_unspecified {
			instance.Status = new(cloudprovider.InstanceStatus)
			instance.Status.State = cloudprovider.InstanceState(pbStatus.GetInstanceState())
			pbErrorInfo := pbStatus.GetErrorInfo()
			if pbErrorInfo.GetErrorCode() != "" {
				instance.Status.ErrorInfo = &cloudprovider.InstanceErrorInfo{
					ErrorClass:   cloudprovider.InstanceErrorClass(pbErrorInfo.GetInstanceErrorClass()),
					ErrorCode:    pbErrorInfo.GetErrorCode(),
					ErrorMessage: pbErrorInfo.GetErrorMessage(),
				}
			}
		}
		instances = append(instances, instance)
	}
	return instances, nil
}

// TemplateNodeInfo returns a framework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The
// returned NodeInfo is expected to have a fully populated Node object, with
// all of the labels, capacity and allocatable information as well as all pods
// that are started on the node by default, using manifest (most likely only
// kube-proxy). Implementation optional.
//
// The definition of a generic `NodeInfo` for each potential provider is a pretty
// complex approach and does not cover all the scenarios. For the sake of simplicity,
// the `nodeInfo` is defined as a Kubernetes `k8s.io.api.core.v1.Node` type
// where the system could still extract certain info about the node.
func (n *NodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	if n.nodeInfo != nil {
		klog.V(5).Infof("Returning cached nodeInfo for node group %v", n.id)
		return *n.nodeInfo, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), n.grpcTimeout)
	defer cancel()
	klog.V(5).Infof("Performing gRPC call NodeGroupTemplateNodeInfo for node group %v", n.id)
	res, err := n.client.NodeGroupTemplateNodeInfo(ctx, &protos.NodeGroupTemplateNodeInfoRequest{
		Id: n.id,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.Unimplemented {
			return nil, cloudprovider.ErrNotImplemented
		}
		klog.V(1).Infof("Error on gRPC call NodeGroupTemplateNodeInfo: %v", err)
		return nil, err
	}
	pbNodeInfo := res.GetNodeInfo()
	if pbNodeInfo == nil {
		n.nodeInfo = new(*framework.NodeInfo)
		return nil, nil
	}
	nodeInfo := framework.NewNodeInfo(pbNodeInfo, nil)
	n.nodeInfo = &nodeInfo
	return nodeInfo, nil
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
	ctx, cancel := context.WithTimeout(context.Background(), n.grpcTimeout)
	defer cancel()
	klog.V(5).Infof("Performing gRPC call NodeGroupGetOptions for node group %v", n.id)
	res, err := n.client.NodeGroupGetOptions(ctx, &protos.NodeGroupAutoscalingOptionsRequest{
		Id: n.id,
		Defaults: &protos.NodeGroupAutoscalingOptions{
			ScaleDownUtilizationThreshold:    defaults.ScaleDownUtilizationThreshold,
			ScaleDownGpuUtilizationThreshold: defaults.ScaleDownGpuUtilizationThreshold,
			ScaleDownUnneededTime: &metav1.Duration{
				Duration: defaults.ScaleDownUnneededTime,
			},
			ScaleDownUnreadyTime: &metav1.Duration{
				Duration: defaults.ScaleDownUnreadyTime,
			},
			MaxNodeProvisionTime: &metav1.Duration{
				Duration: defaults.MaxNodeProvisionTime,
			},
		},
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.Unimplemented {
			return nil, cloudprovider.ErrNotImplemented
		}
		klog.V(1).Infof("Error on gRPC call NodeGroupGetOptions: %v", err)
		return nil, err
	}
	pbOpts := res.GetNodeGroupAutoscalingOptions()
	if pbOpts == nil {
		return nil, nil
	}
	opts := &config.NodeGroupAutoscalingOptions{
		ScaleDownUtilizationThreshold:    pbOpts.GetScaleDownUtilizationThreshold(),
		ScaleDownGpuUtilizationThreshold: pbOpts.GetScaleDownGpuUtilizationThreshold(),
		ScaleDownUnneededTime:            pbOpts.GetScaleDownUnneededTime().Duration,
		ScaleDownUnreadyTime:             pbOpts.GetScaleDownUnreadyTime().Duration,
		MaxNodeProvisionTime:             pbOpts.GetMaxNodeProvisionTime().Duration,
	}
	return opts, nil
}
