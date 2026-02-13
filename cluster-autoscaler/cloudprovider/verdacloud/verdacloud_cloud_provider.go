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

package verdacloud

import (
	"io"
	"os"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	coreoptions "k8s.io/autoscaler/cluster-autoscaler/core/options"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/klog/v2"
)

// Constants for Verdacloud node labels and resources.
const (
	// NodeGroupLabelKey is the label key used to identify node groups.
	NodeGroupLabelKey = "verda.com/node-group"
	// AcceleratorLabel is the label key for GPU accelerator type.
	AcceleratorLabel = "verda.com/accelerator"
	// ResourceNvidiaGPU is the resource name for NVIDIA GPUs.
	ResourceNvidiaGPU = "nvidia.com/gpu"
)

// VerdacloudCloudProvider implements the cloudprovider.CloudProvider interface for Verdacloud.
type VerdacloudCloudProvider struct {
	manager         *VerdacloudManager
	resourceLimiter *cloudprovider.ResourceLimiter
}

// VerdacloudAsgSpec represents the specification for an Auto Scaling Group.
type VerdacloudAsgSpec struct {
	minSize        int
	maxSize        int
	instanceType   string
	name           string
	hostnamePrefix string
}

func newVerdacloudCloudProvider(manager *VerdacloudManager, rl *cloudprovider.ResourceLimiter) (*VerdacloudCloudProvider, error) {
	return &VerdacloudCloudProvider{
		manager:         manager,
		resourceLimiter: rl,
	}, nil
}

// Name returns the name of the cloud provider.
func (d *VerdacloudCloudProvider) Name() string {
	return cloudprovider.VerdacloudProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (d *VerdacloudCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	asgs := d.manager.getAsgs()
	groups := make([]cloudprovider.NodeGroup, 0, len(asgs))
	for _, asg := range asgs {
		groups = append(groups, &VerdacloudNodeGroup{asg: asg, manager: d.manager})
	}
	return groups
}

// NodeGroupForNode returns the node group for the given node.
func (d *VerdacloudCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	instanceRef, err := instanceRefFromProviderId(node.Spec.ProviderID)
	if err != nil {
		return nil, nil
	}

	asg, err := d.manager.GetAsgForInstance(instanceRef)
	if err != nil {
		return nil, err
	}
	if asg == nil {
		return nil, nil
	}

	return &VerdacloudNodeGroup{asg: asg, manager: d.manager}, nil
}

// HasInstance returns true if the node has an instance in the cloud provider.
func (d *VerdacloudCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	instanceRef, err := instanceRefFromProviderId(node.Spec.ProviderID)
	if err != nil {
		return false, nil
	}

	asg, err := d.manager.GetAsgForInstance(instanceRef)
	if err != nil {
		return false, err
	}

	return asg != nil, nil
}

// Pricing returns the pricing model for this cloud provider.
func (d *VerdacloudCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes returns a list of available machine types.
func (d *VerdacloudCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return d.manager.GetAvailableMachineTypes()
}

// NewNodeGroup creates a new node group with the given parameters.
func (d *VerdacloudCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns the resource limiter for this cloud provider.
func (d *VerdacloudCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return d.resourceLimiter, nil
}

// GPULabel returns the label key for GPU accelerator type.
func (d *VerdacloudCloudProvider) GPULabel() string {
	return AcceleratorLabel
}

// GetAvailableGPUTypes returns a map of available GPU types.
func (d *VerdacloudCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return d.manager.GetAvailableGPUTypes()
}

// GetNodeGpuConfig returns the GPU configuration for the given node.
func (d *VerdacloudCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	gpuLabel := d.GPULabel()
	_, hasGpuLabel := node.Labels[gpuLabel]
	gpuAllocatable, hasGpuAllocatable := node.Status.Allocatable[ResourceNvidiaGPU]
	if hasGpuLabel || (hasGpuAllocatable && !gpuAllocatable.IsZero()) {
		return &cloudprovider.GpuConfig{
			Label:                gpuLabel,
			Type:                 node.Labels[gpuLabel],
			ExtendedResourceName: ResourceNvidiaGPU,
		}
	}
	return nil
}

// Cleanup cleans up the cloud provider.
func (d *VerdacloudCloudProvider) Cleanup() error {
	return nil
}

// Refresh refreshes the state of the cloud provider.
func (d *VerdacloudCloudProvider) Refresh() error {
	return d.manager.Refresh()
}

// BuildVerdacloud builds a Verdacloud cloud provider from the given options.
func BuildVerdacloud(
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

	manager, err := createVerdacloudManager(configFile, do)
	if err != nil {
		klog.Fatalf("Failed to create VerdaCloud manager: %v", err)
	}

	provider, err := newVerdacloudCloudProvider(manager, rl)
	if err != nil {
		klog.Fatalf("Failed to create VerdaCloud cloud provider: %v", err)
	}

	return provider
}
