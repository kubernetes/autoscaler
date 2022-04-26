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

package ovhcloud

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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/ovhcloud/sdk"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

const (
	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "node.kubernetes.ovhcloud.com/gpu"

	// NodePoolLabel is the label added to nodes grouped by node group.
	// Should be soon prepend with `node.kubernetes.ovhcloud.com/`
	NodePoolLabel = "nodepool"

	// MachineAvailableState defines the state for available flavors for node resources.
	MachineAvailableState = "available"

	// GPUMachineCategory defines the default instance category for GPU resources.
	GPUMachineCategory = "t"
)

// OVHCloudProvider implements CloudProvider interface.
type OVHCloudProvider struct {
	manager *OvhCloudManager

	autoscalingOptions config.AutoscalingOptions
	discoveryOptions   cloudprovider.NodeGroupDiscoveryOptions
	resourceLimiter    *cloudprovider.ResourceLimiter
}

// BuildOVHcloud builds the OVHcloud provider.
func BuildOVHcloud(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	// Open cloud provider folder
	var configFile io.ReadCloser
	if opts.CloudConfig != "" {
		var err error

		configFile, err = os.Open(opts.CloudConfig)
		if err != nil {
			klog.Fatalf("Failed to open cloud provider configuration %s: %v", opts.CloudConfig)
		}

		defer configFile.Close()
	}

	// Create a new manager given the cloud config previously loaded
	manager, err := NewManager(configFile)
	if err != nil {
		klog.Fatalf("Failed to create OVHcloud manager: %v", err)
	}

	provider := &OVHCloudProvider{
		manager: manager,

		autoscalingOptions: opts,
		discoveryOptions:   do,
		resourceLimiter:    rl,
	}

	return provider
}

// Name returns name of the cloud provider.
func (provider *OVHCloudProvider) Name() string {
	return cloudprovider.OVHcloudProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (provider *OVHCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	groups := make([]cloudprovider.NodeGroup, 0)

	// Cast API node pools into CA node groups
	// Do not add in array node pool which does not support autoscaling
	for _, pool := range provider.manager.NodePools {
		if !pool.Autoscale {
			continue
		}

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
func (provider *OVHCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	// Fetch node pool label on nodes
	labels := node.GetLabels()
	label, exists := labels[NodePoolLabel]
	if !exists {
		return nil, nil
	}

	// Find in node groups stored in cache which one is linked to the node
	for _, pool := range provider.manager.NodePools {
		// Do not check node group which does not support auto-scaling
		if !pool.Autoscale {
			continue
		}

		if pool.Name == label {
			return &NodeGroup{
				NodePool:    pool,
				Manager:     provider.manager,
				CurrentSize: -1,
			}, nil
		}
	}

	return nil, nil
}

// Pricing returns pricing model for this cloud provider or error if not
// available. Implementation optional.
func (provider *OVHCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	// This is not implemented in API
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from
// the cloud provider. Implementation optional.
func (provider *OVHCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	// Fetch all flavors in API
	flavors, err := provider.manager.Client.ListFlavors(context.Background(), provider.manager.ProjectID, provider.manager.ClusterID)
	if err != nil {
		return nil, fmt.Errorf("failed to list available flavors: %w", err)
	}

	// Cast flavors into machine types string array
	machineTypes := make([]string, 0)
	for _, flavor := range flavors {
		if flavor.State == MachineAvailableState {
			machineTypes = append(machineTypes, flavor.Name)
		}
	}

	return machineTypes, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition
// provided. The node group is not automatically created on the cloud provider
// side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (provider *OVHCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string, taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
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
func (provider *OVHCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return provider.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (provider *OVHCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (provider *OVHCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	// Fetch all flavors in API
	flavors, err := provider.manager.Client.ListFlavors(context.Background(), provider.manager.ProjectID, provider.manager.ClusterID)
	if err != nil {
		return nil
	}

	// Cast flavors into gpu types string array
	gpuTypes := make(map[string]struct{}, 0)
	for _, flavor := range flavors {
		if flavor.State == MachineAvailableState && flavor.Category == GPUMachineCategory {
			gpuTypes[flavor.Name] = struct{}{}
		}
	}

	return gpuTypes
}

// Cleanup cleans up open resources before the cloud provider is destroyed,
// i.e. go routines etc.
func (provider *OVHCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop and can be used to dynamically
// update cloud provider state. In particular the list of node groups returned
// by NodeGroups() can change as a result of CloudProvider.Refresh().
func (provider *OVHCloudProvider) Refresh() error {
	klog.V(4).Info("Refreshing NodeGroups")

	// Check if OpenStack keystone token need to be revoke and re-create
	err := provider.manager.ReAuthenticate()
	if err != nil {
		return fmt.Errorf("failed to re-authenticate client: %w", err)
	}

	// Fetch node pools via OVHcloud API
	pools, err := provider.manager.Client.ListNodePools(context.Background(), provider.manager.ProjectID, provider.manager.ClusterID)
	if err != nil {
		return fmt.Errorf("failed to refresh node pool list: %w", err)
	}

	// Do not append node pool which does not support autoscaling
	provider.manager.NodePools = []sdk.NodePool{}
	for _, pool := range pools {
		if !pool.Autoscale {
			continue
		}

		provider.manager.NodePools = append(provider.manager.NodePools, pool)
	}

	return nil
}
