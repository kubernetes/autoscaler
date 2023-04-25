/*
Copyright 2021 The Kubernetes Authors.

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

package tencentcloud

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
	"k8s.io/klog/v2"
)

const (
	// ProviderName is the cloud provider name for Tencentcloud
	ProviderName = "tencentcloud"

	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "cloud.tencent.com/tke-accelerator"
)

var (
	availableGPUTypes = map[string]struct{}{
		"nvidia-tesla-k80":  {},
		"nvidia-tesla-p100": {},
		"nvidia-tesla-v100": {},
		"nvidia-tesla-p4":   {},
		"nvidia-tesla-t4":   {},
	}
)

// tencentCloudProvider implements CloudProvider interface.
type tencentCloudProvider struct {
	tencentcloudManager TencentcloudManager
	resourceLimiter     *cloudprovider.ResourceLimiter
}

// BuildTencentCloudProvider builds CloudProvider implementation for Tencentcloud.
func BuildTencentCloudProvider(tencentcloudManager TencentcloudManager, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions, resourceLimiter *cloudprovider.ResourceLimiter) (cloudprovider.CloudProvider, error) {
	if discoveryOpts.StaticDiscoverySpecified() {
		return buildStaticallyDiscoveringProvider(tencentcloudManager, discoveryOpts.NodeGroupSpecs, resourceLimiter)
	}

	return nil, fmt.Errorf("failed to build an tencentcloud cloud provider (no node group specs and no node group auto discovery)")
}

func buildStaticallyDiscoveringProvider(tencentcloudManager TencentcloudManager, specs []string, resourceLimiter *cloudprovider.ResourceLimiter) (*tencentCloudProvider, error) {
	return &tencentCloudProvider{
		tencentcloudManager: tencentcloudManager,
		resourceLimiter:     resourceLimiter,
	}, nil
}

// Cleanup ...
func (tencentcloud *tencentCloudProvider) Cleanup() error {
	tencentcloud.tencentcloudManager.Cleanup()
	return nil
}

// Name returns name of the cloud provider.
func (tencentcloud *tencentCloudProvider) Name() string {
	return ProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (tencentcloud *tencentCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	asgs := tencentcloud.tencentcloudManager.GetAsgs()
	result := make([]cloudprovider.NodeGroup, 0, len(asgs))
	for _, asg := range asgs {
		result = append(result, asg)
	}
	return result
}

// NodeGroupForNode returns the node group for the given node.
func (tencentcloud *tencentCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	if node.Spec.ProviderID == "" {
		return nil, nil
	}
	ref, err := TcRefFromProviderID(node.Spec.ProviderID)
	if err != nil {
		return nil, err
	}

	asg, err := tencentcloud.tencentcloudManager.GetAsgForInstance(ref)
	klog.V(6).Infof("node: %v, asg: %v", ref, asg)
	if err != nil {
		return nil, err
	}

	return asg, nil
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (tencentcloud *tencentCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

// GPULabel returns the label added to nodes with GPU resource.
func (tencentcloud *tencentCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes returns all available GPU types cloud provider supports.
func (tencentcloud *tencentCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return availableGPUTypes
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (tencentcloud *tencentCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(tencentcloud, node)
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (tencentcloud *tencentCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
func (tencentcloud *tencentCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
func (tencentcloud *tencentCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (tencentcloud *tencentCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	resourceLimiter, err := tencentcloud.tencentcloudManager.GetResourceLimiter()
	if err != nil {
		return nil, err
	}
	if resourceLimiter != nil {
		return resourceLimiter, nil
	}
	return tencentcloud.resourceLimiter, nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (tencentcloud *tencentCloudProvider) Refresh() error {

	klog.V(4).Infof("Refresh loop")

	if tencentcloud.tencentcloudManager.GetAsgs() == nil {
		return fmt.Errorf("Refresh tencentcloud tencentcloudManager asg is nil")
	}

	return tencentcloud.tencentcloudManager.Refresh()
}

// BuildTencentcloud returns tencentcloud provider
func BuildTencentcloud(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	var config io.ReadCloser
	if opts.CloudConfig != "" {
		var err error
		config, err = os.Open(opts.CloudConfig)
		if err != nil {
			klog.Fatalf("Couldn't open cloud provider configuration %s: %#v", opts.CloudConfig, err)
		}
		defer config.Close()
	}

	manager, err := CreateTencentcloudManager(config, do, opts.Regional)
	if err != nil {
		klog.Fatalf("Failed to create Tencentcloud Manager: %v", err)
	}

	cloudProvider, err := BuildTencentCloudProvider(manager, do, rl)
	if err != nil {
		klog.Fatalf("Failed to create Tencentcloud cloud provider: %v", err)
	}
	return cloudProvider
}
