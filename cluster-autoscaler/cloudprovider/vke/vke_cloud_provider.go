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

package vke

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/vke/sdk"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
)

const (
	GPULabel = "node.kubernetes.vke.com.tr/gpu"

	// NodePoolLabel is the label added to nodes grouped by node group.
	NodePoolLabel = "nodepool"

	// MachineAvailableState defines the state for available flavors for node resources.
	MachineAvailableState = "available"

	// GPUMachineCategory defines the default instance category for GPU resources.
	GPUMachineCategory = "t"
)

var _ cloudprovider.CloudProvider = (*VKEProvider)(nil)
var _ cloudprovider.NodeGroup = (*NodeGroup)(nil)

type VKEProvider struct {
	manager *VKEManager

	autoscalingOptions config.AutoscalingOptions
	discoveryOptions   cloudprovider.NodeGroupDiscoveryOptions
	resourceLimiter    *cloudprovider.ResourceLimiter
}

func BuildVKE(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	// Open cloud provider folder
	var configFile io.ReadCloser
	if opts.CloudConfig != "" {
		var err error

		configFile, err = os.Open(opts.CloudConfig)
		if err != nil {
			klog.Fatalf("Failed to open cloud provider configuration %s: %v", opts.CloudConfig, err)
		}

		defer configFile.Close()
	}

	// Create a new manager given the cloud config previously loaded
	manager, err := NewManager(configFile)
	if err != nil {
		klog.Fatalf("Failed to create VKE manager: %v", err)
	}

	provider := &VKEProvider{
		manager: manager,

		autoscalingOptions: opts,
		discoveryOptions:   do,
		resourceLimiter:    rl,
	}

	return provider
}

// Name returns name of the cloud provider.
func (provider *VKEProvider) Name() string {
	return cloudprovider.VKEProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (provider *VKEProvider) NodeGroups() []cloudprovider.NodeGroup {
	groups := make([]cloudprovider.NodeGroup, 0)
	// Cast API node pools into CA node groups
	klog.V(5).Infof("Listing node pools to build NodeGroups %v", provider.manager.NodePools)
	for _, pool := range provider.manager.NodePools {
		ng := NodeGroup{
			NodePool:    pool,
			Manager:     provider.manager,
			CurrentSize: -1,
		}

		groups = append(groups, &ng)
	}

	return groups
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (provider *VKEProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	// Empty provider ID means the node is not yet registered with OpenStack.
	if node.Spec.ProviderID == "" || node.Spec.ProviderID == providerIDPrefix {
		return nil, nil
	}

	// Try to retrieve the associated node group from an already built mapping in cache
	if ng := provider.findNodeGroupFromCache(node.Spec.ProviderID); ng != nil {
		return ng, nil
	}

	klog.V(5).Infof("trying to find node group of node %s (provider ID %s) by listing all nodes under autoscaled node pools", node.Name, node.Spec.ProviderID)

	// This should also refresh the cache for the next time
	ng, err := provider.findNodeGroupByListingNodes(node)
	if ng == nil {
		klog.Warningf("unable to find which node group the node %s (provider ID %s) belongs to", node.Name, node.Spec.ProviderID)
	}

	return ng, err
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (provider *VKEProvider) HasInstance(node *apiv1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

// findNodeGroupFromCache tries to retrieve the associated node group from an already built mapping in cache
func (provider *VKEProvider) findNodeGroupFromCache(providerID string) cloudprovider.NodeGroup {
	nodeGroup := provider.manager.getNodeGroupPerProviderID(providerID)
	if nodeGroup != nil {
		return nodeGroup
	}
	return nil // To avoid returning a (*cloudprovider.NodeGroup)(nil), which is different from nil
}

// findNodeGroupByListingNodes finds the associated node group from by listing all nodes under autoscaled node pools
func (provider *VKEProvider) findNodeGroupByListingNodes(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	for _, ng := range provider.NodeGroups() {
		instances, err := ng.Nodes()
		klog.V(5).Infof("Listing nodes in node group %v", instances)
		if err != nil {
			return nil, fmt.Errorf("failed to list nodes in node group %s: %w", ng.Id(), err)
		}

		for _, instance := range instances {
			klog.V(5).Infof("InstanceID: %s, ProviderID: %s ", instance.Id, node.Spec.ProviderID)
			if instance.Id == node.Spec.ProviderID {
				return ng, nil
			}
		}
	}

	return nil, nil
}

// Pricing returns pricing model for this cloud provider or error if not
// available. Implementation optional.
func (provider *VKEProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	// This is not implemented in API
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from
// the cloud provider. Implementation optional.
func (provider *VKEProvider) GetAvailableMachineTypes() ([]string, error) {
	klog.V(4).Info("Getting available machine types")

	flavorsByName, err := provider.manager.getFlavorsByID()
	if err != nil {
		return nil, fmt.Errorf("failed to get flavors: %w", err)
	}

	// Cast flavors into machine types string array
	machineTypes := make([]string, 0)
	for _, flavor := range flavorsByName {
		if flavor.State == MachineAvailableState {
			machineTypes = append(machineTypes, flavor.Id)
		}
	}

	return machineTypes, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition
// provided. The node group is not automatically created on the cloud provider
// side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (provider *VKEProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string, taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	ng := &NodeGroup{
		NodePool: sdk.NodePool{
			Name:     fmt.Sprintf("%s-%d", machineType, rand.Int63()),
			Flavor:   machineType,
			MinNodes: 0,
			MaxNodes: 100,
		},
		Manager:     provider.manager,
		CurrentSize: -1,
	}

	return ng, nil
}

// GetResourceLimiter returns struct containing limits (max, min) for
// resources (cores, memory etc.).
func (provider *VKEProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return provider.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (provider *VKEProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (provider *VKEProvider) GetAvailableGPUTypes() map[string]struct{} {
	klog.V(4).Info("Getting available GPU types")

	flavorsByName, err := provider.manager.getFlavorsByID()
	if err != nil {
		klog.Errorf("Failed to get flavors: %v", err)
		return nil
	}

	// Cast flavors into gpu types string array
	gpuTypes := make(map[string]struct{}, 0)
	for _, flavor := range flavorsByName {
		if flavor.State == MachineAvailableState && flavor.GPUs > 0 {
			gpuTypes[flavor.Id] = struct{}{}
		}
	}

	return gpuTypes
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (provider *VKEProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(provider, node)
}

// Cleanup cleans up open resources before the cloud provider is destroyed,
// i.e. go routines etc.
func (provider *VKEProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop and can be used to dynamically
// update cloud provider state. In particular the list of node groups returned
// by NodeGroups() can change as a result of CloudProvider.Refresh().
func (provider *VKEProvider) Refresh() error {
	klog.V(4).Info("Listing node pools to refresh NodeGroups")

	// Check if OpenStack keystone token need to be revoke and re-create
	err := provider.manager.ReAuthenticate()
	if err != nil {
		return fmt.Errorf("failed to re-authenticate client: %w", err)
	}
	klog.V(4).Infof("ClusterId: %s", provider.manager.ClusterID)
	pools, err := provider.manager.Client.ListNodePools(context.Background(), provider.manager.ClusterID)
	klog.V(5).Infof("Pools: %v", pools)
	if err != nil {
		return fmt.Errorf("failed to refresh node pool list: %w", err)
	}

	// Update the node pools cache
	provider.manager.NodePools = pools
	klog.V(5).Infof("Node pools refreshed: %v", pools)

	return nil
}
