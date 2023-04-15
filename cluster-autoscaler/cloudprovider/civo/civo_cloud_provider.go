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

package civo

import (
	"fmt"
	"io"
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

var _ cloudprovider.CloudProvider = (*civoCloudProvider)(nil)

const (
	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "civo.com/gpu-node"

	civoProviderIDPrefix = "civo://"
)

// civoCloudProvider implements CloudProvider interface.
type civoCloudProvider struct {
	manager         *Manager
	resourceLimiter *cloudprovider.ResourceLimiter
}

func newCivoCloudProvider(manager *Manager, rl *cloudprovider.ResourceLimiter) (*civoCloudProvider, error) {
	if err := manager.Refresh(); err != nil {
		return nil, err
	}

	return &civoCloudProvider{
		manager:         manager,
		resourceLimiter: rl,
	}, nil
}

// Name returns name of the cloud provider.
func (d *civoCloudProvider) Name() string {
	return cloudprovider.CivoProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (d *civoCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	nodeGroups := make([]cloudprovider.NodeGroup, len(d.manager.nodeGroups))
	for i, ng := range d.manager.nodeGroups {
		nodeGroups[i] = ng
	}
	return nodeGroups
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (d *civoCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	providerID := node.Spec.ProviderID
	nodeID := toNodeID(providerID)

	klog.V(5).Infof("checking nodegroup for node ID: %q", nodeID)
	for _, group := range d.manager.nodeGroups {
		klog.V(5).Infof("iterating over node group %q", group.Id())
		nodes, err := group.Nodes()
		if err != nil {
			return nil, err
		}

		for _, node := range nodes {
			klog.V(6).Infof("checking node has: %q want: %q", node.Id, providerID)
			if node.Id != providerID {
				continue
			}

			return group, nil
		}
	}

	klog.V(5).Infof("not including node in the workers node group because not present in Civo API: %q", node.Name)
	return nil, nil
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (d *civoCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

// Pricing returns pricing model for this cloud provider or error if not
// available. Implementation optional.
func (d *civoCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from
// the cloud provider. Implementation optional.
func (d *civoCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition
// provided. The node group is not automatically created on the cloud provider
// side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (d *civoCloudProvider) NewNodeGroup(
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
func (d *civoCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return d.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (d *civoCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (d *civoCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return nil
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (d *civoCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(d, node)
}

// Cleanup cleans up open resources before the cloud provider is destroyed,
// i.e. go routines etc.
func (d *civoCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop and can be used to dynamically
// update cloud provider state. In particular the list of node groups returned
// by NodeGroups() can change as a result of CloudProvider.Refresh().
func (d *civoCloudProvider) Refresh() error {
	klog.V(4).Info("Refreshing node group cache")
	return d.manager.Refresh()
}

// BuildCivo builds the Civo cloud provider.
func BuildCivo(
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
		defer func() {
			err := configFile.Close()
			if err == nil {
				klog.Fatalf("Failed to close Civo cloud provider cloud config file: %v", err)
			}
		}()
	}

	manager, err := newManager(configFile, do)
	if err != nil {
		klog.Fatalf("Failed to create Civo manager: %v", err)
	}

	provider, err := newCivoCloudProvider(manager, rl)
	if err != nil {
		klog.Fatalf("Failed to create Civo cloud provider: %v", err)
	}

	return provider
}

// toProviderID returns a provider ID from the given node ID.
func toProviderID(nodeID string) string {
	return fmt.Sprintf("%s%s", civoProviderIDPrefix, nodeID)
}

// toNodeID returns a node ID from the given provider ID.
func toNodeID(providerID string) string {
	return strings.TrimPrefix(providerID, civoProviderIDPrefix)
}
