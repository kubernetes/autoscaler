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

package utho

import (
	"fmt"
	"os"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog/v2"
)

var _ cloudprovider.CloudProvider = (*uthoCloudProvider)(nil)

const uthoProviderIDPrefix = "utho://"

type uthoCloudProvider struct {
	manager         *Manager
	resourceLimiter *cloudprovider.ResourceLimiter
}

func newUthoCloudProvider(manager *Manager, rl *cloudprovider.ResourceLimiter) *uthoCloudProvider {
	return &uthoCloudProvider{
		manager:         manager,
		resourceLimiter: rl,
	}
}

// Name returns name of the cloud provider.
func (u *uthoCloudProvider) Name() string {
	return cloudprovider.UthoProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (u *uthoCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	nodeGroups := make([]cloudprovider.NodeGroup, len(u.manager.nodeGroups))
	for i, ng := range u.manager.nodeGroups {
		nodeGroups[i] = ng
	}
	return nodeGroups
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (u *uthoCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	rawID := node.Spec.ProviderID
	if rawID == "" {
		if lbl, exists := node.Labels["node_id"]; exists {
			rawID = lbl
			klog.V(5).Infof("Spec.ProviderID empty, using node_id label: %q", rawID)
		} else {
			klog.Warningf("node %q has no Spec.ProviderID or node_id label; skipping", node.Name)
			return nil, nil
		}
	}

	providerID := normalizeID(rawID)
	klog.V(5).Infof("checking nodegroup for normalized ID: %q", providerID)

	for _, group := range u.manager.nodeGroups {
		klog.V(4).Infof("iterating over node group %q", group.Id())
		nodes, err := group.Nodes()
		if err != nil {
			return nil, fmt.Errorf("failed to list nodes for group %q: %w", group.Id(), err)
		}
		for _, node := range nodes {
			normalized := normalizeID(node.Id)
			klog.V(5).Infof("checking node has: %q want: %q", normalized, providerID)
			if normalized == providerID {
				return group, nil
			}
		}
	}

	// there is no "ErrNotExist" error, so we have to return a nil error
	return nil, nil
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (u *uthoCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

// Pricing returns pricing model for this cloud provider or error if not available.
// Implementation optional.
func (u *uthoCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
// Implementation optional.
func (u *uthoCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition
// provided. The node group is not automatically created on the cloud provider
// side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (u *uthoCloudProvider) NewNodeGroup(
	machineType string,
	labels map[string]string,
	systemLabels map[string]string,
	taints []apiv1.Taint,
	extraResources map[string]resource.Quantity,
) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (u *uthoCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return u.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (u *uthoCloudProvider) GPULabel() string {
	return ""
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (u *uthoCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return nil
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (u *uthoCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(u, node)
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (u *uthoCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (u *uthoCloudProvider) Refresh() error {
	klog.V(4).Info("Refreshing node group cache")
	return u.manager.Refresh()
}

// BuildUtho builds the Utho cloud provider.
func BuildUtho(
	opts config.AutoscalingOptions,
	do cloudprovider.NodeGroupDiscoveryOptions,
	rl *cloudprovider.ResourceLimiter,
) cloudprovider.CloudProvider {
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
		klog.Fatalf("Failed to create Utho manager: %v", err)
	}

	// the cloud provider automatically uses all node pools in Utho.
	// This means we don't use the cloudprovider.NodeGroupDiscoveryOptions
	// flags (which can be set via '--node-group-auto-discovery' or '-nodes')
	return newUthoCloudProvider(manager, rl)
}

// toProviderID returns a provider ID from the given node ID.
func toProviderID(nodeID string) string {
	return fmt.Sprintf("%s%s", uthoProviderIDPrefix, nodeID)
}
