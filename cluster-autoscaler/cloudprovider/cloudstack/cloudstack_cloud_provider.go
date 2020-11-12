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

package cloudstack

import (
	"fmt"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"

	v1 "k8s.io/api/core/v1"
	klog "k8s.io/klog/v2"
)

const (
	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "accelerator"
)

var (
	availableGPUTypes = map[string]struct{}{
		"nvidia-tesla-k80":  {},
		"nvidia-tesla-p100": {},
		"nvidia-tesla-v100": {},
	}
	availableMachineTypes = []string{}
)

// cloudStackCloudProvider implements CloudProvider interface.
type cloudStackCloudProvider struct {
	manager         *manager
	resourceLimiter *cloudprovider.ResourceLimiter
}

// Name returns name of the cloud provider.
func (provider *cloudStackCloudProvider) Name() string {
	return cloudprovider.CloudStackProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (provider *cloudStackCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	asg := provider.manager.asg
	return []cloudprovider.NodeGroup{asg}
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred.
func (provider *cloudStackCloudProvider) NodeGroupForNode(node *v1.Node) (cloudprovider.NodeGroup, error) {
	return provider.manager.clusterForNode(node)
}

// Cleanup cleans up open resources before the cloud provider is destroyed, i.e. go routines etc.
func (provider *cloudStackCloudProvider) Cleanup() error {
	return provider.manager.cleanup()
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
func (provider *cloudStackCloudProvider) Refresh() error {
	return provider.manager.refresh()
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
func (provider *cloudStackCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return availableMachineTypes, nil
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (provider *cloudStackCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return provider.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (provider *cloudStackCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (provider *cloudStackCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return availableGPUTypes
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (provider *cloudStackCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
func (provider *cloudStackCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string, taints []v1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func createClusterConfig(opts config.AutoscalingOptions) (*clusterConfig, error) {
	if len(opts.NodeGroups) == 0 {
		return nil, fmt.Errorf("Cluster details not present. Please use the --nodes=<min>:<max>:<id>")
	}

	clusterDetails := strings.Split(opts.NodeGroups[0], ":")
	if len(clusterDetails) != 3 {
		return nil, fmt.Errorf("Cluster details not present. Please use the --nodes=<min>:<max>:<id>")
	}

	minClusterSize, err := strconv.Atoi(clusterDetails[0])
	if err != nil {
		return nil, fmt.Errorf("Invalid value for min cluster size %s : %v", clusterDetails[0], err)
	}

	maxClusterSize, err := strconv.Atoi(clusterDetails[1])
	if err != nil {
		return nil, fmt.Errorf("Invalid value for max cluster size %s : %v", clusterDetails[1], err)
	}

	return &clusterConfig{
		clusterID: clusterDetails[2],
		minSize:   minClusterSize,
		maxSize:   maxClusterSize,
	}, nil
}

// BuildCloudStack builds CloudProvider implementation for CloudStack
func BuildCloudStack(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {

	config, err := createClusterConfig(opts)
	if err != nil {
		klog.Fatal(err)
	}

	manager, err := newManager(config, withConfigFile(opts.CloudConfig))
	if err != nil {
		klog.Fatalf("Failed to create CloudStack Manager: %v", err)
	}

	return &cloudStackCloudProvider{
		manager:         manager,
		resourceLimiter: rl,
	}
}
