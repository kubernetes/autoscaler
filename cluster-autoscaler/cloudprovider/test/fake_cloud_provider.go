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

package test

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"sync"
)

// CloudProvider is a fake implementation of the cloudprovider interface for testing.
type CloudProvider struct {
	sync.Mutex
}

// NewCloudProvider creates a new instance of the fake CloudProvider.
func NewCloudProvider() *CloudProvider {
	return &CloudProvider{}
}

// NodeGroups returns all node groups configured in the fake CloudProvider.
func (c *CloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	panic("not implemented")
}

// NodeGroupForNode returns the node group that a given node belongs to.
func (c *CloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	panic("not implemented")
}

// HasInstance returns true if the given node is managed by this cloud provider.
func (c *CloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	panic("not implemented")
}

// GetResourceLimiter generates a NEW limiter based on our current internal maps.
func (c *CloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	panic("not implemented")
}

// GPULabel returns the label used to identify GPU types in this provider.
func (c *CloudProvider) GPULabel() string {
	panic("not implemented")
}

// GetAvailableGPUTypes returns a map of all GPU types available in this provider.
func (c *CloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	panic("not implemented")
}

// GetNodeGpuConfig returns the GPU configuration for a specific node.
func (c *CloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	panic("not implemented")
}

// Cleanup performs any necessary teardown of the CloudProvider.
func (c *CloudProvider) Cleanup() error {
	panic("not implemented")
}

// Refresh updates the internal state of the CloudProvider.
func (c *CloudProvider) Refresh() error { panic("not implemented") }

// Name returns the name of the cloud provider.
func (c *CloudProvider) Name() string { panic("not implemented") }

// Pricing returns the pricing model associated with the provider.
func (c *CloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	panic("not implemented")
}

// GetAvailableMachineTypes returns the machine types supported by the provider.
func (c *CloudProvider) GetAvailableMachineTypes() ([]string, error) {
	panic("not implemented")
}

// NewNodeGroup creates a new node group based on the provided specifications.
func (c *CloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	panic("not implemented")
}
