/*
Copyright 2024 The Kubernetes Authors.

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

package nebius

import (
	"fmt"
	"io"
	"os"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	coreoptions "k8s.io/autoscaler/cluster-autoscaler/core/options"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog/v2"
)

var _ cloudprovider.CloudProvider = (*nebiusCloudProvider)(nil)

const (
	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "nebius.com/gpu-node"

	// nebiusProviderIDPrefix is the prefix for Nebius provider IDs.
	nebiusProviderIDPrefix = "nebius://"

	// nodeGroupIDLabel is the label on Kubernetes nodes that identifies the node group.
	nodeGroupIDLabel = "nebius.com/node-group-id"
)

// nebiusCloudProvider implements CloudProvider interface.
type nebiusCloudProvider struct {
	manager         *Manager
	resourceLimiter *cloudprovider.ResourceLimiter
}

func newNebiusCloudProvider(manager *Manager, rl *cloudprovider.ResourceLimiter) (*nebiusCloudProvider, error) {
	if err := manager.Refresh(); err != nil {
		return nil, err
	}

	return &nebiusCloudProvider{
		manager:         manager,
		resourceLimiter: rl,
	}, nil
}

// Name returns name of the cloud provider.
func (d *nebiusCloudProvider) Name() string {
	return cloudprovider.NebiusProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (d *nebiusCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	nodeGroups := make([]cloudprovider.NodeGroup, len(d.manager.nodeGroups))
	for i, ng := range d.manager.nodeGroups {
		nodeGroups[i] = ng
	}
	return nodeGroups
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (d *nebiusCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	// Check node labels for node group ID.
	if nodeGroupID, ok := node.Labels[nodeGroupIDLabel]; ok {
		for _, ng := range d.manager.nodeGroups {
			if ng.id == nodeGroupID {
				return ng, nil
			}
		}
	}

	// Fall back to checking providerID against cached instances.
	providerID := node.Spec.ProviderID
	if providerID == "" {
		return nil, nil
	}

	for _, ng := range d.manager.nodeGroups {
		if ng.hasInstance(providerID) {
			return ng, nil
		}
	}

	klog.V(5).Infof("no node group found for node %q (providerID: %q)", node.Name, providerID)
	return nil, nil
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (d *nebiusCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

// Pricing returns pricing model for this cloud provider or error if not
// available. Implementation optional.
func (d *nebiusCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from
// the cloud provider. Implementation optional.
func (d *nebiusCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition
// provided. The node group is not automatically created on the cloud provider
// side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (d *nebiusCloudProvider) NewNodeGroup(
	machineType string,
	labels map[string]string,
	systemLabels map[string]string,
	taints []apiv1.Taint,
	extraResources map[string]resource.Quantity,
) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for
// resources (cores, memory etc.).
func (d *nebiusCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return d.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (d *nebiusCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (d *nebiusCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return nil
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (d *nebiusCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(d, node)
}

// Cleanup cleans up open resources before the cloud provider is destroyed,
// i.e. go routines etc.
func (d *nebiusCloudProvider) Cleanup() error {
	return d.manager.Cleanup()
}

// Refresh is called before every main loop and can be used to dynamically
// update cloud provider state. In particular the list of node groups returned
// by NodeGroups() can change as a result of CloudProvider.Refresh().
func (d *nebiusCloudProvider) Refresh() error {
	klog.V(4).Info("Refreshing Nebius node group cache")
	return d.manager.Refresh()
}

// BuildNebius builds the Nebius cloud provider.
func BuildNebius(
	opts *coreoptions.AutoscalerOptions,
	do cloudprovider.NodeGroupDiscoveryOptions,
	rl *cloudprovider.ResourceLimiter,
) cloudprovider.CloudProvider {
	var configFile io.ReadCloser
	if opts.CloudConfig != "" {
		var err error
		configFile, err = os.Open(opts.CloudConfig)
		if err != nil {
			klog.Fatalf("Couldn't open cloud provider configuration %s: %#v", opts.CloudConfig, err)
		}
		defer configFile.Close()
	}

	manager, err := newManager(configFile)
	if err != nil {
		klog.Fatalf("Failed to create Nebius manager: %v", err)
	}

	provider, err := newNebiusCloudProvider(manager, rl)
	if err != nil {
		klog.Fatalf("Failed to create Nebius cloud provider: %v", err)
	}

	return provider
}

// toProviderID returns a provider ID from the given instance ID.
func toProviderID(instanceID string) string {
	return fmt.Sprintf("%s%s", nebiusProviderIDPrefix, instanceID)
}
