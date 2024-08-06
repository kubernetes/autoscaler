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
	"fmt"
	"path"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/scale"
	"k8s.io/client-go/tools/clientcmd"
	klog "k8s.io/klog/v2"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
)

const (
	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "cluster-api/accelerator"
)

var _ cloudprovider.CloudProvider = (*provider)(nil)

type provider struct {
	controller      *machineController
	providerName    string
	resourceLimiter *cloudprovider.ResourceLimiter
}

func (p *provider) Name() string {
	return p.providerName
}

func (p *provider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return p.resourceLimiter, nil
}

func (p *provider) NodeGroups() []cloudprovider.NodeGroup {
	nodegroups, err := p.controller.nodeGroups()
	if err != nil {
		klog.Errorf("error getting node groups: %v", err)
		return nil
	}
	return nodegroups
}

func (p *provider) NodeGroupForNode(node *corev1.Node) (cloudprovider.NodeGroup, error) {
	ng, err := p.controller.nodeGroupForNode(node)
	if err != nil {
		return nil, err
	}
	if ng == nil || reflect.ValueOf(ng).IsNil() {
		return nil, nil
	}
	return ng, nil
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (p *provider) HasInstance(node *corev1.Node) (bool, error) {
	machineID := node.Annotations[machineAnnotationKey]
	ns := node.Annotations[clusterNamespaceAnnotationKey]

	machine, err := p.controller.findMachine(path.Join(ns, machineID))
	if machine != nil {
		return true, nil
	}

	return false, fmt.Errorf("machine not found for node %s: %v", node.Name, err)
}

func (*provider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

func (*provider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

func (*provider) NewNodeGroup(
	machineType string,
	labels map[string]string,
	systemLabels map[string]string,
	taints []corev1.Taint,
	extraResources map[string]resource.Quantity,
) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (*provider) Cleanup() error {
	return nil
}

func (p *provider) Refresh() error {
	return nil
}

// GetInstanceID gets the instance ID for the specified node.
func (p *provider) GetInstanceID(node *corev1.Node) string {
	return node.Spec.ProviderID
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (p *provider) GetAvailableGPUTypes() map[string]struct{} {
	// TODO: implement this
	return nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (p *provider) GPULabel() string {
	return GPULabel
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (p *provider) GetNodeGpuConfig(node *corev1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(p, node)
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
func BuildClusterAPI(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
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

	// Grab a dynamic interface that we can create informers from
	managementClient, err := dynamic.NewForConfig(managementConfig)
	if err != nil {
		klog.Fatalf("could not generate dynamic client for config")
	}
	managementClientSet, err := kubernetes.NewForConfig(managementConfig)
	if err != nil {
		klog.Fatalf("could not generate kube clientset from mgmt config")
	}

	workloadClient, err := kubernetes.NewForConfig(workloadConfig)
	if err != nil {
		klog.Fatalf("create kube clientset failed: %v", err)
	}

	managementDiscoveryClient, err := discovery.NewDiscoveryClientForConfig(managementConfig)
	if err != nil {
		klog.Fatalf("create discovery client failed: %v", err)
	}

	cachedDiscovery := memory.NewMemCacheClient(managementDiscoveryClient)
	managementScaleClient, err := scale.NewForConfig(
		managementConfig,
		restmapper.NewDeferredDiscoveryRESTMapper(cachedDiscovery),
		dynamic.LegacyAPIPathResolverFunc,
		scale.NewDiscoveryScaleKindResolver(managementDiscoveryClient))
	if err != nil {
		klog.Fatalf("create scale client failed: %v", err)
	}

	// Ideally this would be passed in but the builder is not
	// currently organised to do so.
	stopCh := make(chan struct{})

	controller, err := newMachineController(managementClient, managementClientSet, workloadClient, managementDiscoveryClient, managementScaleClient, do, stopCh)
	if err != nil {
		klog.Fatal(err)
	}

	if err := controller.run(); err != nil {
		klog.Fatal(err)
	}

	return newProvider(cloudprovider.ClusterAPIProviderName, rl, controller)
}
