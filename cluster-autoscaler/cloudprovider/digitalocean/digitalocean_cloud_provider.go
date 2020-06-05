/*
Copyright 2019 The Kubernetes Authors.

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

package digitalocean

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
	klog "k8s.io/klog/v2"
)

var _ cloudprovider.CloudProvider = (*digitaloceanCloudProvider)(nil)

const (
	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "cloud.digitalocean.com/gpu-node"

	doProviderIDPrefix = "digitalocean://"
)

// digitaloceanCloudProvider implements CloudProvider interface.
type digitaloceanCloudProvider struct {
	manager         *Manager
	resourceLimiter *cloudprovider.ResourceLimiter
}

func newDigitalOceanCloudProvider(manager *Manager, rl *cloudprovider.ResourceLimiter) (*digitaloceanCloudProvider, error) {
	if err := manager.Refresh(); err != nil {
		return nil, err
	}

	return &digitaloceanCloudProvider{
		manager:         manager,
		resourceLimiter: rl,
	}, nil
}

// Name returns name of the cloud provider.
func (d *digitaloceanCloudProvider) Name() string {
	return cloudprovider.DigitalOceanProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (d *digitaloceanCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	nodeGroups := make([]cloudprovider.NodeGroup, len(d.manager.nodeGroups))
	for i, ng := range d.manager.nodeGroups {
		nodeGroups[i] = ng
	}
	return nodeGroups
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (d *digitaloceanCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	providerID := node.Spec.ProviderID
	nodeID := toNodeID(providerID)

	klog.V(5).Infof("checking nodegroup for node ID: %q", nodeID)

	// NOTE(arslan): the number of node groups per cluster is usually very
	// small. So even though this looks like quadratic runtime, it's OK to
	// proceed with this.
	for _, group := range d.manager.nodeGroups {
		klog.V(5).Infof("iterating over node group %q", group.Id())
		nodes, err := group.Nodes()
		if err != nil {
			return nil, err
		}

		for _, node := range nodes {
			klog.V(6).Infof("checking node has: %q want: %q", node.Id, providerID)
			// CA uses node.Spec.ProviderID when looking for (un)registered nodes,
			// so we need to use it here too.
			if node.Id != providerID {
				continue
			}

			return group, nil
		}
	}

	// there is no "ErrNotExist" error, so we have to return a nil error
	return nil, nil
}

// Pricing returns pricing model for this cloud provider or error if not
// available. Implementation optional.
func (d *digitaloceanCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from
// the cloud provider. Implementation optional.
func (d *digitaloceanCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition
// provided. The node group is not automatically created on the cloud provider
// side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (d *digitaloceanCloudProvider) NewNodeGroup(
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
func (d *digitaloceanCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return d.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (d *digitaloceanCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (d *digitaloceanCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return nil
}

// Cleanup cleans up open resources before the cloud provider is destroyed,
// i.e. go routines etc.
func (d *digitaloceanCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop and can be used to dynamically
// update cloud provider state. In particular the list of node groups returned
// by NodeGroups() can change as a result of CloudProvider.Refresh().
func (d *digitaloceanCloudProvider) Refresh() error {
	klog.V(4).Info("Refreshing node group cache")
	return d.manager.Refresh()
}

// BuildDigitalOcean builds the DigitalOcean cloud provider.
func BuildDigitalOcean(
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

	manager, err := newManager(configFile)
	if err != nil {
		klog.Fatalf("Failed to create DigitalOcean manager: %v", err)
	}

	// the cloud provider automatically uses all node pools in DigitalOcean.
	// This means we don't use the cloudprovider.NodeGroupDiscoveryOptions
	// flags (which can be set via '--node-group-auto-discovery' or '-nodes')
	provider, err := newDigitalOceanCloudProvider(manager, rl)
	if err != nil {
		klog.Fatalf("Failed to create DigitalOcean cloud provider: %v", err)
	}

	return provider
}

// toProviderID returns a provider ID from the given node ID.
func toProviderID(nodeID string) string {
	return fmt.Sprintf("%s%s", doProviderIDPrefix, nodeID)
}

// toNodeID returns a node or droplet ID from the given provider ID.
func toNodeID(providerID string) string {
	return strings.TrimPrefix(providerID, doProviderIDPrefix)
}
