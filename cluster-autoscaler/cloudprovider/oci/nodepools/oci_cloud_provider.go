/*
Copyright 2020-2023 Oracle and/or its affiliates.
*/

package nodepools

import (
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	ocicommon "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/oci/common"
	caerrors "k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	klog "k8s.io/klog/v2"
)

// OciCloudProvider creates a cloud provider object that is compatible with node pools
type OciCloudProvider struct {
	rl      *cloudprovider.ResourceLimiter
	manager NodePoolManager
}

// NewOciCloudProvider creates a new CloudProvider implementation that supports OKE Node Pools
func NewOciCloudProvider(manager NodePoolManager, rl *cloudprovider.ResourceLimiter) *OciCloudProvider {
	return &OciCloudProvider{
		manager: manager,
		rl:      rl,
	}
}

// Name returns name of the cloud provider.
func (ocp *OciCloudProvider) Name() string {
	return cloudprovider.OracleCloudProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (ocp *OciCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	nodePools := ocp.manager.GetNodePools()
	result := make([]cloudprovider.NodeGroup, 0, len(nodePools))
	for _, nodePool := range nodePools {
		result = append(result, nodePool)
	}
	return result
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (ocp *OciCloudProvider) NodeGroupForNode(n *apiv1.Node) (cloudprovider.NodeGroup, error) {

	ociRef, err := ocicommon.NodeToOciRef(n)
	if err != nil {
		return nil, err
	}

	ng, err := ocp.manager.GetNodePoolForInstance(ociRef)

	// this instance may be part of a node pool that the autoscaler does not handle
	if errors.Cause(err) == errInstanceNodePoolNotFound {
		return nil, nil
	}
	return ng, err
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (ocp *OciCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(ocp, node)
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (ocp *OciCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	instance, err := ocicommon.NodeToOciRef(node)
	if err != nil {
		return true, err
	}
	np, err := ocp.manager.GetNodePoolForInstance(instance)
	if err != nil {
		return true, err
	}
	nodes, err := ocp.manager.GetNodePoolNodes(np)
	if err != nil {
		return true, err
	}
	for _, n := range nodes {
		if n.Id == instance.InstanceID {
			return true, nil
		}
	}
	return false, nil
}

// Pricing returns pricing model for this cloud provider or error if not available.
// Implementation optional.
func (ocp *OciCloudProvider) Pricing() (cloudprovider.PricingModel, caerrors.AutoscalerError) {
	klog.Info("Pricing called")
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
// Implementation optional.
func (ocp *OciCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	klog.Info("GetAvailableMachineTypes called")
	return nil, cloudprovider.ErrNotImplemented
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (ocp *OciCloudProvider) NewNodeGroup(machineType string,
	labels map[string]string,
	systemLabels map[string]string,
	taints []apiv1.Taint,
	extraResources map[string]resource.Quantity,
) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (ocp *OciCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return ocp.rl, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (ocp *OciCloudProvider) GPULabel() string {
	return ""
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (ocp *OciCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return map[string]struct{}{}
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (ocp *OciCloudProvider) Cleanup() error {
	return ocp.manager.Cleanup()
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (ocp *OciCloudProvider) Refresh() error {
	return ocp.manager.Refresh()
}
