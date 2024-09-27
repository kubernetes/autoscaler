//go:build linux
// +build linux

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

// Version to be compiled in the linux environment. May cause compilation issues on
// other OS.

package kubemark

import (
	"fmt"
	"os"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/client-go/informers"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubernetes/pkg/kubemark"

	klog "k8s.io/klog/v2"
)

const (
	// ProviderName is the cloud provider name for kubemark
	ProviderName = "kubemark"

	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = "cloud.google.com/gke-accelerator"
)

var (
	availableGPUTypes = map[string]struct{}{
		"nvidia-tesla-k80":  {},
		"nvidia-tesla-p100": {},
		"nvidia-tesla-v100": {},
	}
)

// KubemarkCloudProvider implements CloudProvider interface for kubemark
type KubemarkCloudProvider struct {
	kubemarkController *kubemark.KubemarkController
	nodeGroups         []*NodeGroup
	resourceLimiter    *cloudprovider.ResourceLimiter
}

// BuildKubemarkCloudProvider builds a CloudProvider for kubemark. Builds
// node groups from passed in specs.
func BuildKubemarkCloudProvider(kubemarkController *kubemark.KubemarkController, specs []string, resourceLimiter *cloudprovider.ResourceLimiter) (*KubemarkCloudProvider, error) {
	kubemark := &KubemarkCloudProvider{
		kubemarkController: kubemarkController,
		nodeGroups:         make([]*NodeGroup, 0),
		resourceLimiter:    resourceLimiter,
	}
	for _, spec := range specs {
		if err := kubemark.addNodeGroup(spec); err != nil {
			return nil, err
		}
	}
	return kubemark, nil
}

func (kubemark *KubemarkCloudProvider) addNodeGroup(spec string) error {
	nodeGroup, err := buildNodeGroup(spec, kubemark.kubemarkController)
	if err != nil {
		return err
	}
	klog.V(2).Infof("adding node group: %s", nodeGroup.Name)
	kubemark.nodeGroups = append(kubemark.nodeGroups, nodeGroup)
	return nil
}

// Name returns name of the cloud provider.
func (kubemark *KubemarkCloudProvider) Name() string {
	return ProviderName
}

// GPULabel returns the label added to nodes with GPU resource.
func (kubemark *KubemarkCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports
func (kubemark *KubemarkCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return availableGPUTypes
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (kubemark *KubemarkCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(kubemark, node)
}

// NodeGroups returns all node groups configured for this cloud provider.
func (kubemark *KubemarkCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	result := make([]cloudprovider.NodeGroup, 0, len(kubemark.nodeGroups))
	for _, nodegroup := range kubemark.nodeGroups {
		result = append(result, nodegroup)
	}
	return result
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (kubemark *KubemarkCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// NodeGroupForNode returns the node group for the given node.
func (kubemark *KubemarkCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	// Skip nodes that are not managed by Kubemark Cloud Provider.
	if !strings.HasPrefix(node.Spec.ProviderID, ProviderName) {
		return nil, nil
	}
	nodeGroupName, err := kubemark.kubemarkController.GetNodeGroupForNode(node.ObjectMeta.Name)
	if err != nil {
		return nil, err
	}
	for _, nodeGroup := range kubemark.nodeGroups {
		if nodeGroup.Name == nodeGroupName {
			return nodeGroup, nil
		}
	}
	return nil, nil
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (kubemark *KubemarkCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
// Implementation optional.
func (kubemark *KubemarkCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, cloudprovider.ErrNotImplemented
}

// NewNodeGroup builds a theoretical node group based on the node definition provided.
func (kubemark *KubemarkCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint,
	extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (kubemark *KubemarkCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return kubemark.resourceLimiter, nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (kubemark *KubemarkCloudProvider) Refresh() error {
	return nil
}

// Cleanup cleans up all resources before the cloud provider is removed
func (kubemark *KubemarkCloudProvider) Cleanup() error {
	return nil
}

// NodeGroup implements NodeGroup interface.
type NodeGroup struct {
	Name               string
	kubemarkController *kubemark.KubemarkController
	minSize            int
	maxSize            int
}

// Id returns nodegroup name.
func (nodeGroup *NodeGroup) Id() string {
	return nodeGroup.Name
}

// MinSize returns minimum size of the node group.
func (nodeGroup *NodeGroup) MinSize() int {
	return nodeGroup.minSize
}

// MaxSize returns maximum size of the node group.
func (nodeGroup *NodeGroup) MaxSize() int {
	return nodeGroup.maxSize
}

// Debug returns a debug string for the nodegroup.
func (nodeGroup *NodeGroup) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", nodeGroup.Id(), nodeGroup.MinSize(), nodeGroup.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
func (nodeGroup *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	instances := make([]cloudprovider.Instance, 0)
	nodes, err := nodeGroup.kubemarkController.GetNodeNamesForNodeGroup(nodeGroup.Name)
	if err != nil {
		return instances, err
	}
	for _, node := range nodes {
		instances = append(instances, cloudprovider.Instance{Id: "kubemark://" + node})
	}
	return instances, nil
}

// DeleteNodes deletes the specified nodes from the node group.
func (nodeGroup *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	size, err := nodeGroup.kubemarkController.GetNodeGroupTargetSize(nodeGroup.Name)
	if err != nil {
		return err
	}
	if size <= nodeGroup.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}
	for _, node := range nodes {
		if err := nodeGroup.kubemarkController.RemoveNodeFromNodeGroup(nodeGroup.Name, node.ObjectMeta.Name); err != nil {
			return err
		}
	}
	return nil
}

// IncreaseSize increases NodeGroup size.
func (nodeGroup *NodeGroup) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}
	size, err := nodeGroup.kubemarkController.GetNodeGroupTargetSize(nodeGroup.Name)
	if err != nil {
		return err
	}
	newSize := int(size) + delta
	if newSize > nodeGroup.MaxSize() {
		return fmt.Errorf("size increase too large, desired: %d max: %d", newSize, nodeGroup.MaxSize())
	}
	return nodeGroup.kubemarkController.SetNodeGroupSize(nodeGroup.Name, newSize)
}

// AtomicIncreaseSize is not implemented.
func (nodeGroup *NodeGroup) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (nodeGroup *NodeGroup) TargetSize() (int, error) {
	size, err := nodeGroup.kubemarkController.GetNodeGroupTargetSize(nodeGroup.Name)
	return int(size), err
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
func (nodeGroup *NodeGroup) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("size decrease must be negative")
	}
	size, err := nodeGroup.kubemarkController.GetNodeGroupTargetSize(nodeGroup.Name)
	if err != nil {
		return err
	}
	nodes, err := nodeGroup.kubemarkController.GetNodeNamesForNodeGroup(nodeGroup.Name)
	if err != nil {
		return err
	}
	newSize := int(size) + delta
	if newSize < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes, targetSize: %d delta: %d existingNodes: %d",
			size, delta, len(nodes))
	}
	return nodeGroup.kubemarkController.SetNodeGroupSize(nodeGroup.Name, newSize)
}

// TemplateNodeInfo returns a node template for this node group.
func (nodeGroup *NodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Exist checks if the node group really exists on the cloud provider side.
func (nodeGroup *NodeGroup) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (nodeGroup *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
func (nodeGroup *NodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (nodeGroup *NodeGroup) Autoprovisioned() bool {
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
func (nodeGroup *NodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func buildNodeGroup(value string, kubemarkController *kubemark.KubemarkController) (*NodeGroup, error) {
	spec, err := dynamic.SpecFromString(value, true)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node group spec: %v", err)
	}

	nodeGroup := &NodeGroup{
		Name:               spec.Name,
		kubemarkController: kubemarkController,
		minSize:            spec.MinSize,
		maxSize:            spec.MaxSize,
	}

	return nodeGroup, nil
}

// BuildKubemark builds Kubemark cloud provider.
func BuildKubemark(opts config.AutoscalingOptions, do cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) cloudprovider.CloudProvider {
	externalConfig, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatalf("Failed to get kubeclient config for external cluster: %v", err)
	}

	// Use provided kubeconfig or fallback to InClusterConfig
	kubemarkConfig := externalConfig
	kubemarkConfigPath := "/kubeconfig/cluster_autoscaler.kubeconfig"
	if _, err := os.Stat(kubemarkConfigPath); !os.IsNotExist(err) {
		if kubemarkConfig, err = clientcmd.BuildConfigFromFlags("", kubemarkConfigPath); err != nil {
			klog.Fatalf("Failed to get kubeclient config for kubemark cluster: %v", err)
		}
	}

	stop := make(chan struct{})

	externalClient := kubeclient.NewForConfigOrDie(externalConfig)
	kubemarkClient := kubeclient.NewForConfigOrDie(kubemarkConfig)

	externalInformerFactory := informers.NewSharedInformerFactory(externalClient, 0)
	kubemarkInformerFactory := informers.NewSharedInformerFactory(kubemarkClient, 0)
	kubemarkNodeInformer := kubemarkInformerFactory.Core().V1().Nodes()
	go kubemarkNodeInformer.Informer().Run(stop)

	kubemarkController, err := kubemark.NewKubemarkController(externalClient, externalInformerFactory,
		kubemarkClient, kubemarkNodeInformer)
	if err != nil {
		klog.Fatalf("Failed to create Kubemark cloud provider: %v", err)
	}

	externalInformerFactory.Start(stop)
	if !kubemarkController.WaitForCacheSync(stop) {
		klog.Fatalf("Failed to sync caches for kubemark controller")
	}
	go kubemarkController.Run(stop)

	provider, err := BuildKubemarkCloudProvider(kubemarkController, do.NodeGroupSpecs, rl)
	if err != nil {
		klog.Fatalf("Failed to create Kubemark cloud provider: %v", err)
	}
	return provider
}
