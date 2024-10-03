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

package wrapper

import (
	"context"
	"fmt"
	"reflect"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/externalgrpc/protos"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	klog "k8s.io/klog/v2"
)

// Wrapper implements protos.CloudProviderServer.
type Wrapper struct {
	protos.UnimplementedCloudProviderServer

	provider cloudprovider.CloudProvider
}

// NewCloudProviderGrpcWrapper creates a grpc wrapper for a cloud provider implementation.
func NewCloudProviderGrpcWrapper(provider cloudprovider.CloudProvider) *Wrapper {
	return &Wrapper{
		provider: provider,
	}
}

// apiv1Node converts a protos.ExternalGrpcNode to a apiv1.Node.
func apiv1Node(pbNode *protos.ExternalGrpcNode) *apiv1.Node {
	apiv1Node := &apiv1.Node{}
	apiv1Node.ObjectMeta = metav1.ObjectMeta{
		Name:        pbNode.GetName(),
		Annotations: pbNode.GetAnnotations(),
		Labels:      pbNode.GetLabels(),
	}
	apiv1Node.Spec = apiv1.NodeSpec{
		ProviderID: pbNode.GetProviderID(),
	}
	return apiv1Node
}

// apiv1Node converts an apiv1.Node to a protos.ExternalGrpcNode.
func pbNodeGroup(ng cloudprovider.NodeGroup) *protos.NodeGroup {
	return &protos.NodeGroup{
		Id:      ng.Id(),
		MaxSize: int32(ng.MaxSize()),
		MinSize: int32(ng.MinSize()),
		Debug:   ng.Debug(),
	}
}

func debug(req fmt.Stringer) {
	klog.V(10).Infof("got gRPC request: %T %s", req, req)
}

// NodeGroups is the wrapper for the cloud provider NodeGroups method.
func (w *Wrapper) NodeGroups(_ context.Context, req *protos.NodeGroupsRequest) (*protos.NodeGroupsResponse, error) {
	debug(req)

	pbNgs := make([]*protos.NodeGroup, 0)
	for _, ng := range w.provider.NodeGroups() {
		pbNgs = append(pbNgs, pbNodeGroup(ng))
	}
	return &protos.NodeGroupsResponse{
		NodeGroups: pbNgs,
	}, nil
}

// NodeGroupForNode is the wrapper for the cloud provider NodeGroupForNode method.
func (w *Wrapper) NodeGroupForNode(_ context.Context, req *protos.NodeGroupForNodeRequest) (*protos.NodeGroupForNodeResponse, error) {
	debug(req)

	pbNode := req.GetNode()
	if pbNode == nil {
		return nil, fmt.Errorf("request fields were nil")
	}
	node := apiv1Node(pbNode)
	ng, err := w.provider.NodeGroupForNode(node)
	if err != nil {
		return nil, err
	}
	// Checks if ng is nil interface or contains nil value
	if ng == nil || reflect.ValueOf(ng).IsNil() {
		return &protos.NodeGroupForNodeResponse{
			NodeGroup: &protos.NodeGroup{}, //NodeGroup with id = "", meaning the node should not be processed by cluster autoscaler
		}, nil
	}
	return &protos.NodeGroupForNodeResponse{
		NodeGroup: pbNodeGroup(ng),
	}, nil
}

// PricingNodePrice is the wrapper for the cloud provider Pricing NodePrice method.
func (w *Wrapper) PricingNodePrice(_ context.Context, req *protos.PricingNodePriceRequest) (*protos.PricingNodePriceResponse, error) {
	debug(req)

	model, err := w.provider.Pricing()
	if err != nil {
		if err == cloudprovider.ErrNotImplemented {
			return nil, status.Error(codes.Unimplemented, err.Error())
		}
		return nil, err
	}
	reqNode := req.GetNode()
	reqStartTime := req.GetStartTime()
	reqEndTime := req.GetEndTime()
	if reqNode == nil || reqStartTime == nil || reqEndTime == nil {
		return nil, fmt.Errorf("request fields were nil")
	}
	price, nodePriceErr := model.NodePrice(apiv1Node(reqNode), reqStartTime.Time, reqEndTime.Time)
	if nodePriceErr != nil {
		return nil, nodePriceErr
	}
	return &protos.PricingNodePriceResponse{
		Price: price,
	}, nil
}

// PricingPodPrice is the wrapper for the cloud provider Pricing PodPrice method.
func (w *Wrapper) PricingPodPrice(_ context.Context, req *protos.PricingPodPriceRequest) (*protos.PricingPodPriceResponse, error) {
	debug(req)

	model, err := w.provider.Pricing()
	if err != nil {
		if err == cloudprovider.ErrNotImplemented {
			return nil, status.Error(codes.Unimplemented, err.Error())
		}
		return nil, err
	}
	reqPod := req.GetPod()
	reqStartTime := req.GetStartTime()
	reqEndTime := req.GetEndTime()
	if reqPod == nil || reqStartTime == nil || reqEndTime == nil {
		return nil, fmt.Errorf("request fields were nil")
	}
	price, podPriceErr := model.PodPrice(reqPod, reqStartTime.Time, reqEndTime.Time)
	if podPriceErr != nil {
		return nil, podPriceErr
	}
	return &protos.PricingPodPriceResponse{
		Price: price,
	}, nil
}

// GPULabel is the wrapper for the cloud provider GPULabel method.
func (w *Wrapper) GPULabel(_ context.Context, req *protos.GPULabelRequest) (*protos.GPULabelResponse, error) {
	debug(req)

	label := w.provider.GPULabel()
	return &protos.GPULabelResponse{
		Label: label,
	}, nil
}

// GetAvailableGPUTypes is the wrapper for the cloud provider GetAvailableGPUTypes method.
func (w *Wrapper) GetAvailableGPUTypes(_ context.Context, req *protos.GetAvailableGPUTypesRequest) (*protos.GetAvailableGPUTypesResponse, error) {
	debug(req)

	types := w.provider.GetAvailableGPUTypes()
	pbGpuTypes := make(map[string]*anypb.Any)
	for t := range types {
		pbGpuTypes[t] = nil
	}
	return &protos.GetAvailableGPUTypesResponse{
		GpuTypes: pbGpuTypes,
	}, nil
}

// Cleanup is the wrapper for the cloud provider Cleanup method.
func (w *Wrapper) Cleanup(_ context.Context, req *protos.CleanupRequest) (*protos.CleanupResponse, error) {
	debug(req)

	err := w.provider.Cleanup()
	return &protos.CleanupResponse{}, err
}

// Refresh is the wrapper for the cloud provider Refresh method.
func (w *Wrapper) Refresh(_ context.Context, req *protos.RefreshRequest) (*protos.RefreshResponse, error) {
	debug(req)

	err := w.provider.Refresh()
	return &protos.RefreshResponse{}, err
}

// getNodeGroup retrieves the NodeGroup giving its id.
func (w *Wrapper) getNodeGroup(id string) cloudprovider.NodeGroup {
	for _, n := range w.provider.NodeGroups() {
		if n.Id() == id {
			return n
		}
	}
	return nil
}

// NodeGroupTargetSize is the wrapper for the cloud provider NodeGroup TargetSize method.
func (w *Wrapper) NodeGroupTargetSize(_ context.Context, req *protos.NodeGroupTargetSizeRequest) (*protos.NodeGroupTargetSizeResponse, error) {
	debug(req)

	id := req.GetId()
	ng := w.getNodeGroup(id)
	if ng == nil {
		return nil, fmt.Errorf("NodeGroup %q, not found", id)
	}
	size, err := ng.TargetSize()
	if err != nil {
		return nil, err
	}
	return &protos.NodeGroupTargetSizeResponse{
		TargetSize: int32(size),
	}, nil
}

// NodeGroupIncreaseSize is the wrapper for the cloud provider NodeGroup IncreaseSize method.
func (w *Wrapper) NodeGroupIncreaseSize(_ context.Context, req *protos.NodeGroupIncreaseSizeRequest) (*protos.NodeGroupIncreaseSizeResponse, error) {
	debug(req)

	id := req.GetId()
	ng := w.getNodeGroup(id)
	if ng == nil {
		return nil, fmt.Errorf("NodeGroup %q, not found", id)
	}
	err := ng.IncreaseSize(int(req.GetDelta()))
	if err != nil {
		return nil, err
	}
	return &protos.NodeGroupIncreaseSizeResponse{}, nil
}

// NodeGroupDeleteNodes is the wrapper for the cloud provider NodeGroup DeleteNodes method.
func (w *Wrapper) NodeGroupDeleteNodes(_ context.Context, req *protos.NodeGroupDeleteNodesRequest) (*protos.NodeGroupDeleteNodesResponse, error) {
	debug(req)

	id := req.GetId()
	ng := w.getNodeGroup(id)
	if ng == nil {
		return nil, fmt.Errorf("NodeGroup %q, not found", id)
	}
	nodes := make([]*apiv1.Node, 0)
	for _, n := range req.GetNodes() {
		nodes = append(nodes, apiv1Node(n))
	}
	err := ng.DeleteNodes(nodes)
	if err != nil {
		return nil, err
	}
	return &protos.NodeGroupDeleteNodesResponse{}, nil
}

// NodeGroupDecreaseTargetSize is the wrapper for the cloud provider NodeGroup DecreaseTargetSize method.
func (w *Wrapper) NodeGroupDecreaseTargetSize(_ context.Context, req *protos.NodeGroupDecreaseTargetSizeRequest) (*protos.NodeGroupDecreaseTargetSizeResponse, error) {
	debug(req)

	id := req.GetId()
	ng := w.getNodeGroup(id)
	if ng == nil {
		return nil, fmt.Errorf("NodeGroup %q, not found", id)
	}
	err := ng.DecreaseTargetSize(int(req.GetDelta()))
	if err != nil {
		return nil, err
	}
	return &protos.NodeGroupDecreaseTargetSizeResponse{}, nil
}

// NodeGroupNodes is the wrapper for the cloud provider NodeGroup Nodes method.
func (w *Wrapper) NodeGroupNodes(_ context.Context, req *protos.NodeGroupNodesRequest) (*protos.NodeGroupNodesResponse, error) {
	debug(req)

	id := req.GetId()
	ng := w.getNodeGroup(id)
	if ng == nil {
		return nil, fmt.Errorf("NodeGroup %q, not found", id)
	}
	instances, err := ng.Nodes()
	if err != nil {
		return nil, err
	}
	pbInstances := make([]*protos.Instance, 0)
	for _, i := range instances {
		pbInstance := new(protos.Instance)
		pbInstance.Id = i.Id
		if i.Status == nil {
			pbInstance.Status = &protos.InstanceStatus{
				InstanceState: protos.InstanceStatus_unspecified,
				ErrorInfo:     &protos.InstanceErrorInfo{},
			}
		} else {
			pbInstance.Status = new(protos.InstanceStatus)
			pbInstance.Status.InstanceState = protos.InstanceStatus_InstanceState(i.Status.State)
			if i.Status.ErrorInfo == nil {
				pbInstance.Status.ErrorInfo = &protos.InstanceErrorInfo{}
			} else {
				pbInstance.Status.ErrorInfo = &protos.InstanceErrorInfo{
					ErrorCode:          i.Status.ErrorInfo.ErrorCode,
					ErrorMessage:       i.Status.ErrorInfo.ErrorMessage,
					InstanceErrorClass: int32(i.Status.ErrorInfo.ErrorClass),
				}
			}
		}
		pbInstances = append(pbInstances, pbInstance)
	}
	return &protos.NodeGroupNodesResponse{
		Instances: pbInstances,
	}, nil
}

// NodeGroupTemplateNodeInfo is the wrapper for the cloud provider NodeGroup TemplateNodeInfo method.
func (w *Wrapper) NodeGroupTemplateNodeInfo(_ context.Context, req *protos.NodeGroupTemplateNodeInfoRequest) (*protos.NodeGroupTemplateNodeInfoResponse, error) {
	debug(req)

	id := req.GetId()
	ng := w.getNodeGroup(id)
	if ng == nil {
		return nil, fmt.Errorf("NodeGroup %q, not found", id)
	}
	info, err := ng.TemplateNodeInfo()
	if err != nil {
		if err == cloudprovider.ErrNotImplemented {
			return nil, status.Error(codes.Unimplemented, err.Error())
		}
		return nil, err
	}
	return &protos.NodeGroupTemplateNodeInfoResponse{
		NodeInfo: info.Node(),
	}, nil
}

// NodeGroupGetOptions is the wrapper for the cloud provider NodeGroup GetOptions method.
func (w *Wrapper) NodeGroupGetOptions(_ context.Context, req *protos.NodeGroupAutoscalingOptionsRequest) (*protos.NodeGroupAutoscalingOptionsResponse, error) {
	debug(req)

	id := req.GetId()
	ng := w.getNodeGroup(id)
	if ng == nil {
		return nil, fmt.Errorf("NodeGroup %q, not found", id)
	}
	pbDefaults := req.GetDefaults()
	if pbDefaults == nil {
		return nil, fmt.Errorf("request fields were nil")
	}
	defaults := config.NodeGroupAutoscalingOptions{
		ScaleDownUtilizationThreshold:    pbDefaults.GetScaleDownGpuUtilizationThreshold(),
		ScaleDownGpuUtilizationThreshold: pbDefaults.GetScaleDownGpuUtilizationThreshold(),
		ScaleDownUnneededTime:            pbDefaults.GetScaleDownUnneededTime().Duration,
		ScaleDownUnreadyTime:             pbDefaults.GetScaleDownUnneededTime().Duration,
		MaxNodeProvisionTime:             pbDefaults.GetMaxNodeProvisionTime().Duration,
	}
	opts, err := ng.GetOptions(defaults)
	if err != nil {
		if err == cloudprovider.ErrNotImplemented {
			return nil, status.Error(codes.Unimplemented, err.Error())
		}
		return nil, err
	}
	if opts == nil {
		return nil, fmt.Errorf("GetOptions not implemented") //make this explicitly so that grpc response is discarded
	}
	return &protos.NodeGroupAutoscalingOptionsResponse{
		NodeGroupAutoscalingOptions: &protos.NodeGroupAutoscalingOptions{
			ScaleDownUtilizationThreshold:    opts.ScaleDownUtilizationThreshold,
			ScaleDownGpuUtilizationThreshold: opts.ScaleDownGpuUtilizationThreshold,
			ScaleDownUnneededTime: &metav1.Duration{
				Duration: opts.ScaleDownUnneededTime,
			},
			ScaleDownUnreadyTime: &metav1.Duration{
				Duration: opts.ScaleDownUnreadyTime,
			},
			MaxNodeProvisionTime: &metav1.Duration{
				Duration: opts.MaxNodeProvisionTime,
			},
		},
	}, nil
}
