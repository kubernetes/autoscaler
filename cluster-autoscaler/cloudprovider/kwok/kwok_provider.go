/*
Copyright 2023 The Kubernetes Authors.

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

package kwok

import (
	"context"
	"fmt"
	"os"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// Name returns name of the cloud provider.
func (kwok *KwokCloudProvider) Name() string {
	return ProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (kwok *KwokCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	result := make([]cloudprovider.NodeGroup, 0, len(kwok.nodeGroups))
	for _, nodegroup := range kwok.nodeGroups {
		result = append(result, nodegroup)
	}
	return result
}

// NodeGroupForNode returns the node group for the given node.
func (kwok *KwokCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	// Skip nodes that are not managed by kwok cloud provider.
	if !strings.HasPrefix(node.Spec.ProviderID, ProviderName) {
		klog.V(2).Infof("ignoring node '%s' because it is not managed by kwok", node.GetName())
		return nil, nil
	}

	for _, nodeGroup := range kwok.nodeGroups {
		if nodeGroup.name == getNGName(node, kwok.config) {
			klog.V(5).Infof("found nodegroup '%s' for node '%s'", nodeGroup.name, node.GetName())
			return nodeGroup, nil
		}
	}
	return nil, nil
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
// Since there is no underlying cloud provider instance, return true
func (kwok *KwokCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	return true, nil
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (kwok *KwokCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
// Implementation optional.
func (kwok *KwokCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, cloudprovider.ErrNotImplemented
}

// NewNodeGroup builds a theoretical node group based on the node definition provided.
func (kwok *KwokCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint,
	extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (kwok *KwokCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return kwok.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (kwok *KwokCloudProvider) GPULabel() string {
	// GPULabel() might get called before the config is loaded
	if kwok.config == nil || kwok.config.status == nil {
		return ""
	}
	return kwok.config.status.gpuLabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports
func (kwok *KwokCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	// GetAvailableGPUTypes() might get called before the config is loaded
	if kwok.config == nil || kwok.config.status == nil {
		return map[string]struct{}{}
	}
	return kwok.config.status.availableGPUTypes
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (kwok *KwokCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(kwok, node)
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (kwok *KwokCloudProvider) Refresh() error {

	allNodes, err := kwok.allNodesLister.List(labels.Everything())
	if err != nil {
		klog.ErrorS(err, "failed to list all nodes from lister")
		return err
	}

	targetSizeInCluster := make(map[string]int)

	for _, node := range allNodes {
		ngName := getNGName(node, kwok.config)
		if ngName == "" {
			continue
		}

		targetSizeInCluster[ngName] += 1
	}

	for _, ng := range kwok.nodeGroups {
		ng.targetSize = targetSizeInCluster[ng.Id()]
	}

	return nil
}

// Cleanup cleans up all resources before the cloud provider is removed
func (kwok *KwokCloudProvider) Cleanup() error {
	for _, ng := range kwok.nodeGroups {
		nodeNames, err := ng.getNodeNamesForNodeGroup()
		if err != nil {
			return fmt.Errorf("error cleaning up: %v", err)
		}

		for _, node := range nodeNames {
			err := kwok.kubeClient.CoreV1().Nodes().Delete(context.Background(), node, v1.DeleteOptions{})
			if err != nil {
				klog.Errorf("error cleaning up kwok provider nodes '%v'", node)
			}
		}
	}

	return nil
}

// BuildKwok builds kwok cloud provider.
func BuildKwok(opts config.AutoscalingOptions,
	do cloudprovider.NodeGroupDiscoveryOptions,
	rl *cloudprovider.ResourceLimiter,
	informerFactory informers.SharedInformerFactory) cloudprovider.CloudProvider {

	var restConfig *rest.Config
	var err error
	if os.Getenv("KWOK_PROVIDER_MODE") == "local" {
		// Check and load kubeconfig from the path set
		// in KUBECONFIG env variable (if not use default path of ~/.kube/config)
		apiConfig, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
		if err != nil {
			klog.Fatal(err)
		}

		// Create rest config from kubeconfig
		restConfig, err = clientcmd.NewDefaultClientConfig(*apiConfig, &clientcmd.ConfigOverrides{}).ClientConfig()
		if err != nil {
			klog.Fatal(err)
		}
	} else {
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			klog.Fatalf("failed to get kubeclient config for cluster: %v", err)
		}
	}

	// TODO: switch to using the same kube/rest config as the core CA after
	// https://github.com/kubernetes/autoscaler/pull/6180/files is merged
	kubeClient := kubeclient.NewForConfigOrDie(restConfig)

	p, err := BuildKwokProvider(&kwokOptions{
		kubeClient:      kubeClient,
		autoscalingOpts: &opts,
		discoveryOpts:   &do,
		resourceLimiter: rl,
		ngNodeListerFn:  kube_util.NewNodeLister,
		allNodesLister:  informerFactory.Core().V1().Nodes().Lister()})

	if err != nil {
		klog.Fatal(err)
	}

	return p
}

// BuildKwokProvider builds the kwok provider
func BuildKwokProvider(ko *kwokOptions) (*KwokCloudProvider, error) {

	kwokConfig, err := LoadConfigFile(ko.kubeClient)
	if err != nil {
		return nil, fmt.Errorf("failed to load kwok provider config: %v", err)
	}

	var nodegroups []*NodeGroup
	var nodeTemplates []*apiv1.Node
	switch kwokConfig.ReadNodesFrom {
	case nodeTemplatesFromConfigMap:
		if nodeTemplates, err = LoadNodeTemplatesFromConfigMap(kwokConfig.ConfigMap.Name, ko.kubeClient); err != nil {
			return nil, err
		}
	case nodeTemplatesFromCluster:
		if nodeTemplates, err = loadNodeTemplatesFromCluster(kwokConfig, ko.kubeClient, nil); err != nil {
			return nil, err
		}
	}

	if !kwokConfig.Nodes.SkipTaint {
		for _, no := range nodeTemplates {
			no.Spec.Taints = append(no.Spec.Taints, kwokProviderTaint())
		}
	}

	nodegroups = createNodegroups(nodeTemplates, ko.kubeClient, kwokConfig, ko.ngNodeListerFn, ko.allNodesLister)

	return &KwokCloudProvider{
		nodeGroups:      nodegroups,
		kubeClient:      ko.kubeClient,
		resourceLimiter: ko.resourceLimiter,
		config:          kwokConfig,
		allNodesLister:  ko.allNodesLister,
	}, nil
}

func kwokProviderTaint() apiv1.Taint {
	return apiv1.Taint{
		Key:    "kwok-provider",
		Value:  "true",
		Effect: apiv1.TaintEffectNoSchedule,
	}
}
