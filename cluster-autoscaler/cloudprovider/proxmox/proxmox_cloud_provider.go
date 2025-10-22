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

package proxmox

import (
	"context"
	"io"
	"os"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/klog/v2"
)

var _ cloudprovider.CloudProvider = (*proxmoxCloudProvider)(nil)

const (
	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "cloud.proxmox.com/gpu-node"

	proxmoxProviderIDPrefix = "proxmox://"
)

// proxmoxCloudProvider implements CloudProvider interface.
type proxmoxCloudProvider struct {
	manager         *Manager
	resourceLimiter *cloudprovider.ResourceLimiter
}

func newProxmoxCloudProvider(manager *Manager, rl *cloudprovider.ResourceLimiter) *proxmoxCloudProvider {
	return &proxmoxCloudProvider{
		manager:         manager,
		resourceLimiter: rl,
	}
}

// Name returns name of the cloud provider.
func (p *proxmoxCloudProvider) Name() string {
	return cloudprovider.ProxmoxProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (p *proxmoxCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	nodeGroups := make([]cloudprovider.NodeGroup, len(p.manager.nodeGroups))
	for i, ng := range p.manager.nodeGroups {
		nodeGroups[i] = ng
	}
	return nodeGroups
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (p *proxmoxCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	providerID := node.Spec.ProviderID
	for _, group := range p.manager.nodeGroups {
		nodes, err := group.Nodes()
		if err != nil {
			klog.Warningf("Failed to get nodes for group %s: %v", group.Id(), err)
			return nil, err
		}

		for _, instance := range nodes {
			if instance.Id == providerID {
				return group, nil
			}
		}
	}

	return nil, nil
}

func (p *proxmoxCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	nodeUUID := toNodeID(node.Spec.ProviderID)
	if nodeUUID == "" {
		klog.V(4).Infof("❌ Node %s has no valid UUID in providerID", node.Name)
		return false, nil
	}

	// Check all Proxmox nodes for this VM UUID - we need to check if the VM exists anywhere
	// regardless of whether it belongs to our managed node groups
	for _, nodeGroup := range p.manager.nodeGroups {
		vms, err := p.manager.client.GetVMs(context.Background(), nodeGroup.proxmoxNode)
		if err != nil {
			return false, err
		}

		for _, vm := range vms {
			if vm.UUID == nodeUUID {
				return true, nil
			}
		}
	}

	return false, nil
}

// Pricing returns pricing model for this cloud provider or error if not
// available. Implementation optional.
func (p *proxmoxCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from
// the cloud provider. Implementation optional.
func (p *proxmoxCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition
// provided. The node group is not automatically created on the cloud provider
// side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (p *proxmoxCloudProvider) NewNodeGroup(
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
func (p *proxmoxCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return p.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (p *proxmoxCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (p *proxmoxCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return nil
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (p *proxmoxCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return nil
}

// Cleanup cleans up open resources before the cloud provider is destroyed,
// i.e. go routines etc.
func (p *proxmoxCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop and can be used to dynamically
// update cloud provider state. In particular the list of node groups returned
// by NodeGroups() can change as a result of CloudProvider.Refresh().
func (p *proxmoxCloudProvider) Refresh() error {
	return p.manager.Refresh()
}

// BuildProxmox builds the Proxmox cloud provider.
func BuildProxmox(
	opts config.AutoscalingOptions,
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

	manager, err := newManager(configFile, do)
	if err != nil {
		klog.Fatalf("Failed to create Proxmox manager: %v", err)
	}

	return newProxmoxCloudProvider(manager, rl)
}
