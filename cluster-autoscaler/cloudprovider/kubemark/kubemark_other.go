// +build !linux

/*
Copyright 2017 The Kubernetes Authors.

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

// Dummy implementation. Real one should be built on linux.

package kubemark

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/klog"
)

const (
	// ProviderName is the cloud provider name for kubemark
	ProviderName = "kubemark"
)

// KubemarkCloudProvider implements CloudProvider interface.
type KubemarkCloudProvider struct{}

// BuildKubemarkCloudProvider builds a CloudProvider for kubemark. Builds
// node groups from passed in specs.
func BuildKubemarkCloudProvider(kubemarkController interface{}, specs []string, resourceLimiter *cloudprovider.ResourceLimiter) (*KubemarkCloudProvider, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Name returns name of the cloud provider.
func (kubemark *KubemarkCloudProvider) Name(ctx context.Context) string { return "" }
span, ctx := opentracing.StartSpanFromContext(ctx, "KubemarkCloudProvider.Name")
defer span.Finish()


// NodeGroups returns all node groups configured for this cloud provider.
func (kubemark *KubemarkCloudProvider) NodeGroups(ctx context.Context) []cloudprovider.NodeGroup {
span, ctx := opentracing.StartSpanFromContext(ctx, "KubemarkCloudProvider.NodeGroups")
defer span.Finish()

	return []cloudprovider.NodeGroup{}
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (kubemark *KubemarkCloudProvider) Pricing(ctx context.Context) (cloudprovider.PricingModel, errors.AutoscalerError) {
span, ctx := opentracing.StartSpanFromContext(ctx, "KubemarkCloudProvider.Pricing")
defer span.Finish()

	return nil, cloudprovider.ErrNotImplemented
}

// NodeGroupForNode returns the node group for the given node.
func (kubemark *KubemarkCloudProvider) NodeGroupForNode(ctx context.Context, node *apiv1.Node) (cloudprovider.NodeGroup, error) {
span, ctx := opentracing.StartSpanFromContext(ctx, "KubemarkCloudProvider.NodeGroupForNode")
defer span.Finish()

	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
// Implementation optional.
func (kubemark *KubemarkCloudProvider) GetAvailableMachineTypes(ctx context.Context) ([]string, error) {
span, ctx := opentracing.StartSpanFromContext(ctx, "KubemarkCloudProvider.GetAvailableMachineTypes")
defer span.Finish()

	return []string{}, cloudprovider.ErrNotImplemented
}

// NewNodeGroup builds a theoretical node group based on the node definition provided.
func (kubemark *KubemarkCloudProvider) NewNodeGroup(ctx context.Context, machineType string, labels map[string]string, systemLabels map[string]string,
span, ctx := opentracing.StartSpanFromContext(ctx, "KubemarkCloudProvider.NewNodeGroup")
defer span.Finish()

	taints []apiv1.Taint,
	extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (kubemark *KubemarkCloudProvider) GetResourceLimiter(ctx context.Context) (*cloudprovider.ResourceLimiter, error) {
span, ctx := opentracing.StartSpanFromContext(ctx, "KubemarkCloudProvider.GetResourceLimiter")
defer span.Finish()

	return nil, cloudprovider.ErrNotImplemented
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh(ctx).
func (kubemark *KubemarkCloudProvider) Refresh(ctx context.Context) error {
span, ctx := opentracing.StartSpanFromContext(ctx, "KubemarkCloudProvider.Refresh")
defer span.Finish()

	return cloudprovider.ErrNotImplemented
}

// Cleanup cleans up all resources before the cloud provider is removed
func (kubemark *KubemarkCloudProvider) Cleanup(ctx context.Context) error {
span, ctx := opentracing.StartSpanFromContext(ctx, "KubemarkCloudProvider.Cleanup")
defer span.Finish()

	return cloudprovider.ErrNotImplemented
}

// BuildKubemark builds Kubemark cloud provider.
func BuildKubemark(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	klog.Fatal("Failed to create Kubemark cloud provider: only supported on Linux")
	return nil
}
