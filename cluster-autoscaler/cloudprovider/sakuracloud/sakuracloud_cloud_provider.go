/*
Copyright The Kubernetes Authors.

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

package sakuracloud

import (
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/informers"
	"k8s.io/klog/v2"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider/builder"
	coreoptions "sigs.k8s.io/cluster-autoscaler/pkg/core/options"
	autoscalerErrors "sigs.k8s.io/cluster-autoscaler/pkg/utils/errors"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/gpu"
)

// ProviderName is the cloud provider name for SAKURA cloud.
const ProviderName = "sakuracloud"

var _ cloudprovider.CloudProvider = (*sakuracloudCloudProvider)(nil)

func init() {
	builder.RegisterCloudProvider(ProviderName, func(opts *coreoptions.AutoscalerOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter, informerFactory informers.SharedInformerFactory) cloudprovider.CloudProvider {
		return BuildSakuraCloud(opts, do, rl)
	})
	builder.SetDefaultCloudProvider(ProviderName)
}

const (
	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "sakuracloud/gpu-node"

	providerIDPrefix = "sakuracloud://"
)

// providerIDForServer builds "sakuracloud://<zone>/<serverName>". The node's
// kubelet must be started with the matching --provider-id (the bootstrap
// startup note derives it from the hostname, which equals the server name).
func providerIDForServer(zone, serverName string) string {
	return fmt.Sprintf("%s%s/%s", providerIDPrefix, zone, serverName)
}

// serverNameFromProviderID extracts the server name from a providerID.
func serverNameFromProviderID(providerID string) (string, error) {
	rest, ok := strings.CutPrefix(providerID, providerIDPrefix)
	if !ok {
		return "", fmt.Errorf("providerID %q does not have prefix %q", providerID, providerIDPrefix)
	}
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) != 2 || parts[1] == "" {
		return "", fmt.Errorf("providerID %q: expected format %szone/serverName", providerID, providerIDPrefix)
	}
	return parts[1], nil
}

// sakuracloudCloudProvider implements cloudprovider.CloudProvider for
// SAKURA cloud (sakura.ad.jp).
type sakuracloudCloudProvider struct {
	manager         *sakuracloudManager
	resourceLimiter *cloudprovider.ResourceLimiter
}

// Name returns the name of the cloud provider.
func (d *sakuracloudCloudProvider) Name() string {
	return ProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (d *sakuracloudCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	groups := make([]cloudprovider.NodeGroup, 0, len(d.manager.nodeGroups))
	for _, group := range d.manager.nodeGroups {
		groups = append(groups, group)
	}
	return groups
}

// NodeGroupForNode returns the node group for the given node. Nodes whose
// providerID is not sakuracloud:// (a mixed-provider cluster: control plane,
// other clouds) are not managed by this provider and yield nil, per the
// CloudProvider contract.
func (d *sakuracloudCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	serverName, err := serverNameFromProviderID(node.Spec.ProviderID)
	if err != nil {
		klog.V(4).Infof("sakuracloud: node %s has foreign providerID %q, treating as unmanaged", node.Name, node.Spec.ProviderID)
		return nil, nil
	}
	server := d.manager.serverByName(serverName)
	if server == nil {
		return nil, nil
	}
	groupName, ok := server.groupName()
	if !ok {
		return nil, nil
	}
	group, exists := d.manager.nodeGroups[groupName]
	if !exists {
		return nil, nil
	}
	return group, nil
}

// HasInstance returns whether a given node has a corresponding instance.
func (d *sakuracloudCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

// Pricing is not implemented.
func (d *sakuracloudCloudProvider) Pricing() (cloudprovider.PricingModel, autoscalerErrors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes is not implemented.
func (d *sakuracloudCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup is not implemented: node groups are statically configured.
func (d *sakuracloudCloudProvider) NewNodeGroup(
	machineType string,
	labels map[string]string,
	systemLabels map[string]string,
	taints []apiv1.Taint,
	extraResources map[string]resource.Quantity,
) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns resource constraints for the cluster.
func (d *sakuracloudCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return d.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (d *sakuracloudCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes returns all available GPU types.
func (d *sakuracloudCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return nil
}

// GetNodeGpuConfig returns the GPU config of the given node.
func (d *sakuracloudCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(d, node)
}

// Cleanup closes open resources.
func (d *sakuracloudCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop iteration.
func (d *sakuracloudCloudProvider) Refresh() error {
	if err := d.manager.refreshServers(); err != nil {
		return err
	}
	for _, group := range d.manager.nodeGroups {
		group.syncTargetSize()
	}
	return nil
}

// BuildSakuraCloud builds the SAKURA cloud provider.
func BuildSakuraCloud(_ *coreoptions.AutoscalerOptions, _ cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	manager, err := newManager()
	if err != nil {
		klog.Fatalf("Failed to create SAKURA cloud manager: %v", err)
	}
	return &sakuracloudCloudProvider{
		manager:         manager,
		resourceLimiter: rl,
	}
}
