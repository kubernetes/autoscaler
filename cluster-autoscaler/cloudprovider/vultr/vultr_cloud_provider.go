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

package vultr

import (
	"fmt"
	"os"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog/v2"
)

var _ cloudprovider.CloudProvider = (*vultrCloudProvider)(nil)

const vultrProviderIDPrefix = "vultr://"

type vultrCloudProvider struct {
	manager         *manager
	resourceLimiter *cloudprovider.ResourceLimiter
}

func newVultrCloudProvider(manager *manager, rl *cloudprovider.ResourceLimiter) *vultrCloudProvider {
	return &vultrCloudProvider{
		manager:         manager,
		resourceLimiter: rl,
	}
}

// Name returns name of the cloud provider.
func (v *vultrCloudProvider) Name() string {
	return cloudprovider.VultrProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (v *vultrCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	nodeGroups := make([]cloudprovider.NodeGroup, len(v.manager.nodeGroups))
	for i, ng := range v.manager.nodeGroups {
		nodeGroups[i] = ng
	}
	return nodeGroups
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (v *vultrCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	providerID := node.Spec.ProviderID

	// we want to find the pool for a specific node
	for _, group := range v.manager.nodeGroups {
		nodes, err := group.Nodes()
		if err != nil {
			return nil, err
		}

		for _, node := range nodes {
			if node.Id == providerID {
				return group, nil
			}
		}
	}
	return nil, nil
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (v *vultrCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

// Pricing returns pricing model for this cloud provider or error if not available.
// Implementation optional.
func (v *vultrCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
// Implementation optional.
func (v *vultrCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition
// provided. The node group is not automatically created on the cloud provider
// side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (v *vultrCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string, taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (v *vultrCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return v.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (v *vultrCloudProvider) GPULabel() string {
	return ""
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (v *vultrCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return nil
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (v *vultrCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(v, node)
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (v *vultrCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (v *vultrCloudProvider) Refresh() error {
	klog.V(4).Info("Refreshing node group cache")
	return v.manager.Refresh()
}

// toProviderID returns a provider ID from the given node ID.
func toProviderID(nodeID string) string {
	return fmt.Sprintf("%s%s", vultrProviderIDPrefix, nodeID)
}

// toNodeID returns a node or droplet ID from the given provider ID.
func toNodeID(providerID string) string {
	return strings.TrimPrefix(providerID, vultrProviderIDPrefix)
}

// BuildVultr builds the Vultr cloud provider.
func BuildVultr(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	if opts.CloudConfig == "" {
		klog.Fatalf("No config file provided, please specify it via the --cloud-config flag")
	}

	configFile, err := os.Open(opts.CloudConfig)
	if err != nil {
		klog.Fatalf("Could not open cloud provider configuration file %q, error: %v", opts.CloudConfig, err)
	}

	defer configFile.Close()

	manager, err := newManager(configFile)
	if err != nil {
		klog.Fatalf("Failed to create Vultr manager: %v", err)
	}

	// the cloud provider automatically uses all node pools in Vultr.
	// This means we don't use the cloudprovider.NodeGroupDiscoveryOptions
	// flags (which can be set via '--node-group-auto-discovery' or '-nodes')
	return newVultrCloudProvider(manager, rl)
}
