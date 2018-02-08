/*
Copyright 2018 The Kubernetes Authors.

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

package aztools

import (
	"fmt"
	"strconv"
	"sync"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/aztools/az"
	cloudBuilder "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/builder"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"

	"github.com/golang/glog"
)

const (
	// ProviderName is the cloud provider name for Azure
	ProviderName = "aztools"

	scaleToZeroSupported = false
)

// OnScaleUpFunc is a function called on node group increase in AzToolsCloudProvider.
// First parameter is the NodeGroup id, second is the increase delta.
type OnScaleUpFunc func(string, int) error

// OnScaleDownFunc is a function called on cluster scale down
type OnScaleDownFunc func(string, string) error

// OnNodeGroupCreateFunc is a fuction called when a new node group is created.
type OnNodeGroupCreateFunc func(string) error

// OnNodeGroupDeleteFunc is a function called when a node group is deleted.
type OnNodeGroupDeleteFunc func(string) error

// AzToolsCloudProvider is a cloud provider to be used with az_tools.
type AzToolsCloudProvider struct {
	sync.Mutex
	nodes             map[string]string
	groups            map[string]cloudprovider.NodeGroup
	onScaleUp         func(string, int) error
	onScaleDown       func(string, string) error
	onNodeGroupCreate func(string) error
	onNodeGroupDelete func(string) error
	machineTypes      []string
	machineTemplates  map[string]*schedulercache.NodeInfo
	resourceLimiter   *cloudprovider.ResourceLimiter
	kubeClient        kube_client.Interface
}

func BuildAzToolsCloudProvider(clusterName string, discoveryOpts cloudprovider.NodeGroupDiscoveryOptions, rl *cloudprovider.ResourceLimiter) (*AzToolsCloudProvider, error) {
	// TODO(harry): Check if `az list` something works
	provider := NewAzToolsCloudProvider(az.OnScaleUp, az.OnScaleDown, rl)

	for i, spec := range discoveryOpts.NodeGroupSpecs {
		if i > 0 {
			return nil, fmt.Errorf("multiple node groups detected: this is not supported for az tools for now")
		}
		s, err := dynamic.SpecFromString(spec, scaleToZeroSupported)
		if err != nil {
			return nil, fmt.Errorf("failed to parse node group spec: %v: %v", spec, err)
		}

		grpID := clusterName + "-" + strconv.Itoa(i)

		// min, max, targetSize
		provider.AddNodeGroup(grpID, s.MinSize, s.MaxSize, 0)

		// Initializing: fetch all nodes in the cluster and add them into group.
		nodes, err := provider.kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		for node := range nodes.Items {
			provider.AddNode(grpID, node.Name)
		}
	}

	return provider, nil
}

func createInClusterKubeClient() kube_client.Interface {
	kubeConfig, err := config.GetKubeClientConfig("")
	if err != nil {
		glog.Fatalf("Failed to build in cluster Kubernetes client configuration: %v", err)
	}

	return kube_client.NewForConfigOrDie(kubeConfig)
}

// NewAzToolsCloudProvider builds new AzToolsCloudProvider
func NewAzToolsCloudProvider(onScaleUp OnScaleUpFunc, onScaleDown OnScaleDownFunc, rl *cloudprovider.ResourceLimiter) *AzToolsCloudProvider {
	return &AzToolsCloudProvider{
		nodes:           make(map[string]string),
		groups:          make(map[string]cloudprovider.NodeGroup),
		onScaleUp:       onScaleUp,
		onScaleDown:     onScaleDown,
		resourceLimiter: rl,
		kubeClient:      createInClusterKubeClient(),
	}
}

// Name returns name of the cloud provider.
func (azcp *AzToolsCloudProvider) Name() string {
	return ProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (azcp *AzToolsCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	azcp.Lock()
	defer azcp.Unlock()

	result := make([]cloudprovider.NodeGroup, 0)
	for _, group := range azcp.groups {
		result = append(result, group)
	}
	return result
}

// GetNodeGroup returns node group with the given name.
func (azcp *AzToolsCloudProvider) GetNodeGroup(name string) cloudprovider.NodeGroup {
	azcp.Lock()
	defer azcp.Unlock()
	return azcp.groups[name]
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred.
func (azcp *AzToolsCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	azcp.Lock()
	defer azcp.Unlock()

	groupName, found := azcp.nodes[node.Name]
	if !found {
		return nil, nil
	}
	group, found := azcp.groups[groupName]
	if !found {
		return nil, nil
	}
	return group, nil
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (azcp *AzToolsCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
func (azcp *AzToolsCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return azcp.machineTypes, nil
}

// AddNodeGroup adds node group to test cloud provider.
func (azcp *AzToolsCloudProvider) AddNodeGroup(id string, min int, max int, size int) {
	azcp.Lock()
	defer azcp.Unlock()

	azcp.groups[id] = &AzToolsNodeGroup{
		cloudProvider:   azcp,
		id:              id,
		minSize:         min,
		maxSize:         max,
		targetSize:      size,
		exist:           true,
		autoprovisioned: false,
	}
}

// AddNode adds the given node to the group.
func (azcp *AzToolsCloudProvider) AddNode(nodeGroupId string, node *apiv1.Node) {
	azcp.Lock()
	defer azcp.Unlock()
	azcp.nodes[node.Name] = nodeGroupId
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (azcp *AzToolsCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return azcp.resourceLimiter, nil
}

// SetResourceLimiter sets resource limiter.
func (azcp *AzToolsCloudProvider) SetResourceLimiter(resourceLimiter *cloudprovider.ResourceLimiter) {
	azcp.resourceLimiter = resourceLimiter
}

// Cleanup this is a function to close resources associated with the cloud provider
func (azcp *AzToolsCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (azcp *AzToolsCloudProvider) Refresh() error {
	azcp.Lock()
	defer azcp.Unlock()

	nodes, err := provider.kubeClient.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Clear all cached nodes in provider
	azcp.nodes = map[string]string{}

	// Add real world nodes into cache
	for node := range nodes.Items {
		azcp.AddNode(grpID, node.Name)
	}
	return nil
}

// AzToolsNodeGroup is a node group used by AzToolsCloudProvider.
type AzToolsNodeGroup struct {
	sync.Mutex
	cloudProvider   *AzToolsCloudProvider
	id              string
	maxSize         int
	minSize         int
	targetSize      int
	exist           bool
	autoprovisioned bool
	machineType     string
}

// MaxSize returns maximum size of the node group.
func (aztng *AzToolsNodeGroup) MaxSize() int {
	aztng.Lock()
	defer aztng.Unlock()

	return aztng.maxSize
}

// MinSize returns minimum size of the node group.
func (aztng *AzToolsNodeGroup) MinSize() int {
	aztng.Lock()
	defer aztng.Unlock()

	return aztng.minSize
}

// TargetSize returns the current target size of the node group. It is possible that the
// number of nodes in Kubernetes is different at the moment but should be equal
// to Size() once everything stabilizes (new nodes finish startup and registration or
// removed nodes are deleted completely)
func (aztng *AzToolsNodeGroup) TargetSize() (int, error) {
	aztng.Lock()
	defer aztng.Unlock()

	return aztng.targetSize, nil
}

// SetTargetSize sets target size for group. Function is used only in tests.
func (aztng *AzToolsNodeGroup) SetTargetSize(size int) {
	aztng.Lock()
	defer aztng.Unlock()
	aztng.targetSize = size
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated.
func (aztng *AzToolsNodeGroup) IncreaseSize(delta int) error {
	aztng.Lock()
	aztng.targetSize += delta
	aztng.Unlock()

	return aztng.cloudProvider.onScaleUp(aztng.id, delta)
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
func (aztng *AzToolsNodeGroup) Exist() bool {
	aztng.Lock()
	defer aztng.Unlock()
	return aztng.exist
}

// Create creates the node group on the cloud provider side.
func (aztng *AzToolsNodeGroup) Create() error {
	return fmt.Errorf("unimplemented")
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
func (aztng *AzToolsNodeGroup) Delete() error {
	return aztng.cloudProvider.onNodeGroupDelete(aztng.id)
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
func (aztng *AzToolsNodeGroup) DecreaseTargetSize(delta int) error {
	aztng.Lock()
	aztng.targetSize += delta
	aztng.Unlock()

	return aztng.cloudProvider.onScaleUp(aztng.id, delta)
}

// DeleteNodes deletes nodes from this node group. Error is returned either on
// failure or if the given node doesn't belong to this node group. This function
// should wait until node group size is updated.
func (aztng *AzToolsNodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	aztng.Lock()
	id := aztng.id
	aztng.targetSize -= len(nodes)
	aztng.Unlock()
	for _, node := range nodes {
		err := aztng.cloudProvider.onScaleDown(id, node.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

// Id returns an unique identifier of the node group.
func (aztng *AzToolsNodeGroup) Id() string {
	aztng.Lock()
	defer aztng.Unlock()

	return aztng.id
}

// Debug returns a string containing all information regarding this node group.
func (aztng *AzToolsNodeGroup) Debug() string {
	aztng.Lock()
	defer aztng.Unlock()

	return fmt.Sprintf("%s target:%d min:%d max:%d", aztng.id, aztng.targetSize, aztng.minSize, aztng.maxSize)
}

// Nodes returns a list of all nodes that belong to this node group.
func (aztng *AzToolsNodeGroup) Nodes() ([]string, error) {
	aztng.Lock()
	defer aztng.Unlock()

	result := make([]string, 0)
	for node, nodegroup := range aztng.cloudProvider.nodes {
		if nodegroup == aztng.id {
			result = append(result, node)
		}
	}
	return result, nil
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (aztng *AzToolsNodeGroup) Autoprovisioned() bool {
	return aztng.autoprovisioned
}

// TemplateNodeInfo returns a node template for this node group.
func (aztng *AzToolsNodeGroup) TemplateNodeInfo() (*schedulercache.NodeInfo, error) {
	if aztng.cloudProvider.machineTemplates == nil {
		return nil, cloudprovider.ErrNotImplemented
	}
	if aztng.autoprovisioned {
		template, found := aztng.cloudProvider.machineTemplates[aztng.machineType]
		if !found {
			return nil, fmt.Errorf("No template declared for %s", aztng.machineType)
		}
		return template, nil
	}
	template, found := aztng.cloudProvider.machineTemplates[aztng.id]
	if !found {
		return nil, fmt.Errorf("No template declared for %s", aztng.id)
	}
	return template, nil
}
