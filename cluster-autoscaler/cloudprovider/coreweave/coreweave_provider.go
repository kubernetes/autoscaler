/*
Copyright 2025 The Kubernetes Authors.

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

package coreweave

import (
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/klog/v2"
)

// CoreWeaveCloudProvider implements the CloudProvider interface for CoreWeave.
type CoreWeaveCloudProvider struct {
	manager         CoreWeaveManagerInterface
	resourceLimiter *cloudprovider.ResourceLimiter
}

// NewCoreWeaveCloudProvider initializes a new CoreWeave cloud provider instance.
// It creates a CoreWeave client and a dynamic client, then initializes the CoreWeave
// manager with these clients. If any step fails, it logs the error and returns nil.
func NewCoreWeaveCloudProvider(rl *cloudprovider.ResourceLimiter, opts config.AutoscalingOptions) cloudprovider.CloudProvider {
	// Get the shared *rest.Config using the autoscaler's logic
	restConfig := kube_util.GetKubeConfig(opts.KubeClientOpts)
	clientset, err := GetCoreWeaveClient(restConfig)
	if err != nil {
		klog.Errorf("Failed to create CoreWeave client: %v", err)
		return nil
	}
	dynamicClient, err := GetCoreWeaveDynamicClient(restConfig)
	if err != nil {
		klog.Errorf("Failed to create CoreWeave dynamic client: %v", err)
		return nil
	}
	// Create the CoreWeave manager with the dynamic client and clientset
	manager, err := NewCoreWeaveManager(dynamicClient, clientset)
	if err != nil {
		klog.Errorf("Failed to create CoreWeave manager: %v", err)
		return nil
	}
	return &CoreWeaveCloudProvider{
		manager:         manager,
		resourceLimiter: rl,
	}
}

// Name returns the name of the cloud provider.
func (c *CoreWeaveCloudProvider) Name() string {
	return cloudprovider.CoreWeaveProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (c *CoreWeaveCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	// Check if the manager is nil
	if c.manager == nil {
		klog.Error("CoreWeave manager is nil, cannot retrieve node groups")
		return nil
	}
	// Call the manager to update and retrieve node groups
	klog.V(4).Info("Updating node groups in CoreWeave cloud provider")
	nodeGroups, err := c.manager.UpdateNodeGroup()
	if err != nil {
		klog.Errorf("Failed to update node groups: %v", err)
		return nil
	}
	return nodeGroups
}

// NodeGroupForNode returns the node group for the given node.
func (c *CoreWeaveCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	klog.V(4).Infof("Getting node group for node %s", node.Name)
	// Check if the manager is nil before proceeding
	if c.manager == nil {
		return nil, fmt.Errorf("CoreWeave manager is nil")
	}
	// Check if the node has the expected label for coreWeaveNodePoolUID
	if node.Labels == nil || node.Labels[coreWeaveNodePoolUID] == "" {
		klog.V(4).Infof("Node %s has no labels, cannot determine node group", node.Name)
		return nil, nil
	}
	// Log the labels of the node
	klog.V(4).Infof("Node labels: %v", node.Labels)
	return c.manager.GetNodeGroup(node.Labels[coreWeaveNodePoolUID])
}

// HasInstance checks if a given node has a corresponding instance in this cloud provider.
func (c *CoreWeaveCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	// Check if the manager is nil
	if c.manager == nil {
		return false, fmt.Errorf("CoreWeave manager is nil")
	}
	// Check if the node has the expected label for coreWeaveNodePoolUID
	if node.Labels == nil || node.Labels[coreWeaveNodePoolUID] == "" {
		klog.V(4).Infof("Node %s has no labels, cannot determine if it has an instance", node.Name)
		return false, nil
	}
	// Log the labels of the node
	klog.V(4).Infof("Node labels: %v", node.Labels)
	// Get the node group for the node
	ng, err := c.manager.GetNodeGroup(node.Labels[coreWeaveNodePoolUID])
	if err != nil {
		klog.V(4).Infof("Failed to get node group for node %s: %v", node.Name, err)
		return false, nil
	}
	if ng == nil {
		klog.V(4).Infof("Node group for node %s not found", node.Name)
		return false, nil
	}
	klog.V(4).Infof("Node group for node %s found: %s", node.Name, ng.Id())
	return true, nil
}

// Pricing returns the pricing model for this cloud provider.
// This method is not implemented for CoreWeave.
func (c *CoreWeaveCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes returns a list of available machine types for this cloud provider.
// This method is not implemented for CoreWeave.
func (c *CoreWeaveCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// NewNodeGroup creates a new node group with the specified machine type, labels, system labels, taints, and extra resources.
// This method is not implemented for CoreWeave.
func (c *CoreWeaveCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns the resource limiter for this cloud provider.
func (c *CoreWeaveCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return c.resourceLimiter, nil
}

// GPULabel returns the label used to identify GPU nodes.
// This method is not implemented for CoreWeave, so it returns an empty string.
func (c *CoreWeaveCloudProvider) GPULabel() string {
	return ""
}

// GetAvailableGPUTypes returns a map of available GPU types for this cloud provider.
// This method is not implemented for CoreWeave, so it returns nil.
func (c *CoreWeaveCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return nil
}

// GetNodeGpuConfig returns the GPU configuration for a given node.
// This method is not implemented for CoreWeave, so it returns nil.
func (c *CoreWeaveCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return nil
}

// Cleanup performs any necessary cleanup for the cloud provider.
// This method is not implemented for CoreWeave, so it returns nil.
func (c *CoreWeaveCloudProvider) Cleanup() error {
	return nil
}

// Refresh refreshes the state of the cloud provider.
func (c *CoreWeaveCloudProvider) Refresh() error {
	return c.manager.Refresh()
}

// IsNodeCandidateForScaleDown returns whether the node is a good candidate for scaling down.
func (c *CoreWeaveCloudProvider) IsNodeCandidateForScaleDown(node *apiv1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

// BuildCoreWeave builds the CoreWeave cloud provider with the given options and returns it.
func BuildCoreWeave(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	klog.V(4).Infof("Building CoreWeave cloud provider with options: %+v", opts)
	return NewCoreWeaveCloudProvider(rl, opts)
}
