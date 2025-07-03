package coreweave

import (
    "k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
    apiv1 "k8s.io/api/core/v1"
    "k8s.io/autoscaler/cluster-autoscaler/config"
   // "k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
    "k8s.io/autoscaler/cluster-autoscaler/utils/errors"
    "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"
)

type CoreWeaveCloudProvider struct{}

func NewCoreWeaveCloudProvider() cloudprovider.CloudProvider {
    return &CoreWeaveCloudProvider{}
}

func (c *CoreWeaveCloudProvider) Name() string {
    return cloudprovider.CoreWeaveProviderName
}

// Implement other methods as needed, returning cloudprovider.ErrNotImplemented for now

func (c *CoreWeaveCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
    return nil
}

func (c *CoreWeaveCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
    return nil, cloudprovider.ErrNotImplemented
}

func (c *CoreWeaveCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
    return false, cloudprovider.ErrNotImplemented
}

func (c *CoreWeaveCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
    return nil, cloudprovider.ErrNotImplemented
}

func (c *CoreWeaveCloudProvider) GetAvailableMachineTypes() ([]string, error) {
    return nil, cloudprovider.ErrNotImplemented
}

func (c *CoreWeaveCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
    taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
    return nil, cloudprovider.ErrNotImplemented
}

func (c *CoreWeaveCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
    return nil, cloudprovider.ErrNotImplemented
}

func (c *CoreWeaveCloudProvider) GPULabel() string {
    return ""
}

func (c *CoreWeaveCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
    return nil
}

func (c *CoreWeaveCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
    return nil
}

func (c *CoreWeaveCloudProvider) Cleanup() error {
    return nil
}

func (c *CoreWeaveCloudProvider) Refresh() error {
    return nil
}

func BuildCoreWeave(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	klog.V(4).Infof("Building CoreWeave cloud provider with options: %+v", opts)
	return &CoreWeaveCloudProvider{}
}