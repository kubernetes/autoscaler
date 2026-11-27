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

package clusterapi

import (
	"context"
	"fmt"
	"path"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/tools/clientcmd"
	klog "k8s.io/klog/v2"

	"k8s.io/client-go/informers"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider"
	"sigs.k8s.io/cluster-autoscaler/pkg/cloudprovider/builder"
	coreoptions "sigs.k8s.io/cluster-autoscaler/pkg/core/options"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/errors"
	"sigs.k8s.io/cluster-autoscaler/pkg/utils/gpu"
)

// ProviderName is the cloud provider name for this provider.
const ProviderName = "clusterapi"

func init() {
	builder.RegisterCloudProvider(ProviderName, func(opts *coreoptions.AutoscalerOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter, informerFactory informers.SharedInformerFactory) cloudprovider.CloudProvider {
		return BuildClusterAPI(opts, do, rl)
	})
	builder.SetDefaultCloudProvider(ProviderName)
}

const (
	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "cluster-api/accelerator"
)

var _ cloudprovider.CloudProvider = (*provider)(nil)

type provider struct {
	controller      *machineController
	providerName    string
	resourceLimiter *cloudprovider.ResourceLimiter
	// stop cancels the context the controller informers run with, it is nil for providers built by newProvider.
	stop context.CancelFunc
}

func (p *provider) Name() string {
	return p.providerName
}

func (p *provider) GetResourceLimiter(ctx context.Context) (*cloudprovider.ResourceLimiter, error) {
	return p.resourceLimiter, nil
}

func (p *provider) NodeGroups(ctx context.Context) []cloudprovider.NodeGroup {
	nodegroups, err := p.controller.nodeGroups()
	if err != nil {
		klog.Errorf("error getting node groups: %v", err)
		return nil
	}
	return nodegroups
}

func (p *provider) NodeGroupForNode(ctx context.Context, node *corev1.Node) (cloudprovider.NodeGroup, error) {
	ng, err := p.controller.nodeGroupForNode(node)
	if err != nil {
		return nil, err
	}
	if ng == nil {
		return nil, nil
	}
	return ng, nil
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (p *provider) HasInstance(ctx context.Context, node *corev1.Node) (bool, error) {
	machineID := node.Annotations[machineAnnotationKey]
	ns := node.Annotations[clusterNamespaceAnnotationKey]

	machine, err := p.controller.findMachine(path.Join(ns, machineID))
	if machine != nil {
		return true, nil
	}

	return false, fmt.Errorf("machine not found for node %s: %v", node.Name, err)
}

func (*provider) Pricing(ctx context.Context) (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

func (*provider) GetAvailableMachineTypes(ctx context.Context) ([]string, error) {
	return []string{}, nil
}

func (*provider) NewNodeGroup(
	ctx context.Context,
	machineType string,
	labels map[string]string,
	systemLabels map[string]string,
	taints []corev1.Taint,
	extraResources map[string]resource.Quantity,
) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (p *provider) Cleanup(ctx context.Context) error {
	if p.stop == nil {
		return nil
	}
	p.stop()
	p.controller.shutdown()
	return nil
}

func (p *provider) Refresh(ctx context.Context) error {
	return nil
}

// GetInstanceID gets the instance ID for the specified node.
func (p *provider) GetInstanceID(node *corev1.Node) string {
	return node.Spec.ProviderID
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (p *provider) GetAvailableGPUTypes(ctx context.Context) map[string]struct{} {
	// TODO: implement this
	return nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (p *provider) GPULabel(ctx context.Context) string {
	return GPULabel
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (p *provider) GetNodeGpuConfig(ctx context.Context, node *corev1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(context.TODO(), p, node)
}

func newProvider(
	name string,
	rl *cloudprovider.ResourceLimiter,
	controller *machineController,
) cloudprovider.CloudProvider {
	return &provider{
		providerName:    name,
		resourceLimiter: rl,
		controller:      controller,
	}
}

// BuildClusterAPI builds CloudProvider implementation for machine api.
func BuildClusterAPI(opts *coreoptions.AutoscalerOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	managementKubeconfig := opts.CloudConfig
	if managementKubeconfig == "" && !opts.ClusterAPICloudConfigAuthoritative {
		managementKubeconfig = opts.KubeClientOpts.KubeConfigPath
	}

	managementConfig, err := clientcmd.BuildConfigFromFlags("", managementKubeconfig)
	if err != nil {
		klog.Fatalf("cannot build management cluster config: %v", err)
	}
	managementConfig.QPS = opts.KubeClientOpts.KubeClientQPS
	managementConfig.Burst = opts.KubeClientOpts.KubeClientBurst

	workloadKubeconfig := opts.KubeClientOpts.KubeConfigPath

	workloadConfig, err := clientcmd.BuildConfigFromFlags("", workloadKubeconfig)
	if err != nil {
		klog.Fatalf("cannot build workload cluster config: %v", err)
	}
	workloadConfig.QPS = opts.KubeClientOpts.KubeClientQPS
	workloadConfig.Burst = opts.KubeClientOpts.KubeClientBurst

	p, err := NewProvider(context.Background(), ProviderOptions{
		ManagementConfig: managementConfig,
		WorkloadConfig:   workloadConfig,
		DiscoveryOptions: do,
		ResourceLimiter:  rl,
	})
	if err != nil {
		klog.Fatal(err)
	}

	if err := RegisterProcessors(p, opts.Processors); err != nil {
		klog.Fatalf("unable to register scale down upgrade processor: %v", err)
	}

	return p
}
