/*
Copyright 2016 The Kubernetes Authors.

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

package linode

import (
	"fmt"
	"io"
	"os"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	klog "k8s.io/klog/v2"
)

// linodeCloudProvider implements CloudProvider interface.
type linodeCloudProvider struct {
	manager         *manager
	resourceLimiter *cloudprovider.ResourceLimiter
}

// Name returns name of the cloud provider.
func (l *linodeCloudProvider) Name() string {
	return cloudprovider.LinodeProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (l *linodeCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	nodeGroups := make([]cloudprovider.NodeGroup, len(l.manager.nodeGroups))
	i := 0
	for _, ng := range l.manager.nodeGroups {
		nodeGroups[i] = ng
		i++
	}
	return nodeGroups
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (l *linodeCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	for _, ng := range l.manager.nodeGroups {
		pool, err := ng.findLKEPoolForNode(node)
		if err != nil {
			return nil, err
		}
		if pool != nil {
			return ng, nil
		}
	}
	return nil, nil
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (l *linodeCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

// Pricing returns pricing model for this cloud provider or error if not available.
// Implementation optional.
func (l *linodeCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
// Implementation optional.
func (l *linodeCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, cloudprovider.ErrNotImplemented
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (l *linodeCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (l *linodeCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return l.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (l *linodeCloudProvider) GPULabel() string {
	return ""
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (l *linodeCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return nil
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (l *linodeCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(l, node)
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (l *linodeCloudProvider) Cleanup() error {
	return nil
}

// BuildLinode builds the BuildLinode cloud provider.
func BuildLinode(
	opts config.AutoscalingOptions,
	do cloudprovider.NodeGroupDiscoveryOptions,
	rl *cloudprovider.ResourceLimiter,
) cloudprovider.CloudProvider {

	// the cloud provider automatically uses all node pools in linode.
	// This means we don't use the cloudprovider.NodeGroupDiscoveryOptions
	// flags (which can be set via '--node-group-auto-discovery' or '-nodes')

	if opts.CloudConfig == "" {
		klog.Fatalf("No config file provided, please specify it via the --cloud-config flag")
	}
	configFile, err := os.Open(opts.CloudConfig)
	if err != nil {
		klog.Fatalf("Could not open cloud provider configuration file %q, error: %v", opts.CloudConfig, err)
	}
	defer configFile.Close()
	lcp, err := newLinodeCloudProvider(configFile, rl)
	if err != nil {
		klog.Fatalf("Could not create linode cloud provider: %v", err)
	}
	return lcp
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (l *linodeCloudProvider) Refresh() error {
	return l.manager.refresh()
}

func newLinodeCloudProvider(config io.Reader, rl *cloudprovider.ResourceLimiter) (cloudprovider.CloudProvider, error) {
	m, err := newManager(config)
	if err != nil {
		return nil, fmt.Errorf("could not create linode manager: %v", err)
	}

	err = m.refresh()
	if err != nil {
		klog.V(1).Infof("Error on first import of LKE node pools: %v", err)
	}
	klog.V(1).Infof("First import of existing LKE node pools ended")
	if len(m.nodeGroups) == 0 {
		klog.V(1).Infof("Could not import any LKE node pool in any node group")
	} else {
		klog.V(1).Infof("imported LKE node pools:")
		for _, ng := range m.nodeGroups {
			klog.V(1).Infof("%s", ng.extendedDebug())
		}
	}

	return &linodeCloudProvider{
		manager:         m,
		resourceLimiter: rl,
	}, nil
}
