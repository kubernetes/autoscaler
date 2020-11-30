/*
Copyright 2016 The Kubernetes Authors.

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
	"fmt"
	"sync"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// OnScaleUpFunc is a function called on node group increase in TestCloudProvider.
// First parameter is the NodeGroup id, second is the increase delta.
type OnScaleUpFunc func(string, int) error

// OnScaleDownFunc is a function called on cluster scale down
type OnScaleDownFunc func(string, string) error

// OnNodeGroupCreateFunc is a function called when a new node group is created.
type OnNodeGroupCreateFunc func(string) error

// OnNodeGroupDeleteFunc is a function called when a node group is deleted.
type OnNodeGroupDeleteFunc func(string) error

// TestCloudProvider is a dummy cloud provider to be used in tests.
type TestCloudProvider struct {
	sync.Mutex
	nodes             map[string]string
	groups            map[string]cloudprovider.NodeGroup
	onScaleUp         func(string, int) error
	onScaleDown       func(string, string) error
	onNodeGroupCreate func(string) error
	onNodeGroupDelete func(string) error
	machineTypes      []string
	machineTemplates  map[string]*schedulerframework.NodeInfo
	priceModel        cloudprovider.PricingModel
	resourceLimiter   *cloudprovider.ResourceLimiter
}

// NewTestCloudProvider builds new TestCloudProvider
func NewTestCloudProvider(onScaleUp OnScaleUpFunc, onScaleDown OnScaleDownFunc) *TestCloudProvider {
	return &TestCloudProvider{
		nodes:           make(map[string]string),
		groups:          make(map[string]cloudprovider.NodeGroup),
		onScaleUp:       onScaleUp,
		onScaleDown:     onScaleDown,
		resourceLimiter: cloudprovider.NewResourceLimiter(make(map[string]int64), make(map[string]int64)),
	}
}

// NewTestAutoprovisioningCloudProvider builds new TestCloudProvider with autoprovisioning support
func NewTestAutoprovisioningCloudProvider(onScaleUp OnScaleUpFunc, onScaleDown OnScaleDownFunc,
	onNodeGroupCreate OnNodeGroupCreateFunc, onNodeGroupDelete OnNodeGroupDeleteFunc,
	machineTypes []string, machineTemplates map[string]*schedulerframework.NodeInfo) *TestCloudProvider {
	return &TestCloudProvider{
		nodes:             make(map[string]string),
		groups:            make(map[string]cloudprovider.NodeGroup),
		onScaleUp:         onScaleUp,
		onScaleDown:       onScaleDown,
		onNodeGroupCreate: onNodeGroupCreate,
		onNodeGroupDelete: onNodeGroupDelete,
		machineTypes:      machineTypes,
		machineTemplates:  machineTemplates,
		resourceLimiter:   cloudprovider.NewResourceLimiter(make(map[string]int64), make(map[string]int64)),
	}
}

// Name returns name of the cloud provider.
func (tcp *TestCloudProvider) Name() string {
	return "TestCloudProvider"
}

// GPULabel returns the label added to nodes with GPU resource.
func (tcp *TestCloudProvider) GPULabel() string {
	return "TestGPULabel/accelerator"
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports
func (tcp *TestCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return map[string]struct{}{
		"nvidia-tesla-k80":  {},
		"nvidia-tesla-p100": {},
		"nvidia-tesla-v100": {},
	}
}

// NodeGroups returns all node groups configured for this cloud provider.
func (tcp *TestCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	tcp.Lock()
	defer tcp.Unlock()

	result := make([]cloudprovider.NodeGroup, 0)
	for _, group := range tcp.groups {
		result = append(result, group)
	}
	return result
}

// GetNodeGroup returns node group with the given name.
func (tcp *TestCloudProvider) GetNodeGroup(name string) cloudprovider.NodeGroup {
	tcp.Lock()
	defer tcp.Unlock()
	return tcp.groups[name]
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred.
func (tcp *TestCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	tcp.Lock()
	defer tcp.Unlock()

	groupName, found := tcp.nodes[node.Name]
	if !found {
		return nil, nil
	}
	group, found := tcp.groups[groupName]
	if !found {
		return nil, nil
	}
	return group, nil
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (tcp *TestCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	if tcp.priceModel == nil {
		return nil, cloudprovider.ErrNotImplemented
	}

	return tcp.priceModel, nil
}

// SetPricingModel set given priceModel to test cloud provider
func (tcp *TestCloudProvider) SetPricingModel(priceModel cloudprovider.PricingModel) {
	tcp.priceModel = priceModel
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
func (tcp *TestCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return tcp.machineTypes, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
func (tcp *TestCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return &TestNodeGroup{
		cloudProvider:   tcp,
		id:              "autoprovisioned-" + machineType,
		minSize:         0,
		maxSize:         1000,
		targetSize:      0,
		exist:           false,
		autoprovisioned: true,
		machineType:     machineType,
		labels:          labels,
		taints:          taints,
	}, nil
}

// NewNodeGroupWithId creates a new node group with custom ID suffix.
func (tcp *TestCloudProvider) NewNodeGroupWithId(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity, id string) (cloudprovider.NodeGroup, error) {
	return &TestNodeGroup{
		cloudProvider:   tcp,
		id:              "autoprovisioned-" + machineType + "-" + id,
		minSize:         0,
		maxSize:         1000,
		targetSize:      0,
		exist:           false,
		autoprovisioned: true,
		machineType:     machineType,
		labels:          labels,
		taints:          taints,
	}, nil
}

// InsertNodeGroup adds already created node group to test cloud provider.
func (tcp *TestCloudProvider) InsertNodeGroup(nodeGroup cloudprovider.NodeGroup) {
	tcp.Lock()
	defer tcp.Unlock()

	tcp.groups[nodeGroup.Id()] = nodeGroup
}

// BuildNodeGroup returns a test node group.
func (tcp *TestCloudProvider) BuildNodeGroup(id string, min, max, size int, autoprovisioned bool, machineType string) *TestNodeGroup {
	return &TestNodeGroup{
		cloudProvider:   tcp,
		id:              id,
		minSize:         min,
		maxSize:         max,
		targetSize:      size,
		exist:           true,
		autoprovisioned: autoprovisioned,
		machineType:     machineType,
	}
}

// AddNodeGroup adds node group to test cloud provider.
func (tcp *TestCloudProvider) AddNodeGroup(id string, min int, max int, size int) {
	nodeGroup := tcp.BuildNodeGroup(id, min, max, size, false, "")
	tcp.InsertNodeGroup(nodeGroup)
}

// AddAutoprovisionedNodeGroup adds node group to test cloud provider.
func (tcp *TestCloudProvider) AddAutoprovisionedNodeGroup(id string, min int, max int, size int, machineType string) *TestNodeGroup {
	nodeGroup := tcp.BuildNodeGroup(id, min, max, size, true, machineType)
	tcp.InsertNodeGroup(nodeGroup)
	return nodeGroup
}

// DeleteNodeGroup removes node group from test cloud provider.
func (tcp *TestCloudProvider) DeleteNodeGroup(id string) {
	tcp.Lock()
	defer tcp.Unlock()

	delete(tcp.groups, id)
}

// AddNode adds the given node to the group.
func (tcp *TestCloudProvider) AddNode(nodeGroupId string, node *apiv1.Node) {
	tcp.Lock()
	defer tcp.Unlock()

	tcp.nodes[node.Name] = nodeGroupId
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (tcp *TestCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return tcp.resourceLimiter, nil
}

// SetResourceLimiter sets resource limiter.
func (tcp *TestCloudProvider) SetResourceLimiter(resourceLimiter *cloudprovider.ResourceLimiter) {
	tcp.resourceLimiter = resourceLimiter
}

// Cleanup this is a function to close resources associated with the cloud provider
func (tcp *TestCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (tcp *TestCloudProvider) Refresh() error {
	return nil
}

// TestNodeGroup is a node group used by TestCloudProvider.
type TestNodeGroup struct {
	sync.Mutex
	cloudProvider   *TestCloudProvider
	id              string
	maxSize         int
	minSize         int
	targetSize      int
	exist           bool
	autoprovisioned bool
	machineType     string
	labels          map[string]string
	taints          []apiv1.Taint
}

// NewTestNodeGroup creates a TestNodeGroup without setting up the realted TestCloudProvider.
// Useful for testing only.
func NewTestNodeGroup(id string, maxSize, minSize, targetSize int, exist, autoprovisioned bool,
	machineType string, labels map[string]string, taints []apiv1.Taint) *TestNodeGroup {
	return &TestNodeGroup{
		id:              id,
		maxSize:         maxSize,
		minSize:         minSize,
		targetSize:      targetSize,
		exist:           exist,
		autoprovisioned: autoprovisioned,
		machineType:     machineType,
		labels:          labels,
		taints:          taints,
	}
}

// MaxSize returns maximum size of the node group.
func (tng *TestNodeGroup) MaxSize() int {
	tng.Lock()
	defer tng.Unlock()

	return tng.maxSize
}

// MinSize returns minimum size of the node group.
func (tng *TestNodeGroup) MinSize() int {
	tng.Lock()
	defer tng.Unlock()

	return tng.minSize
}

// TargetSize returns the current target size of the node group. It is possible that the
// number of nodes in Kubernetes is different at the moment but should be equal
// to Size() once everything stabilizes (new nodes finish startup and registration or
// removed nodes are deleted completely)
func (tng *TestNodeGroup) TargetSize() (int, error) {
	tng.Lock()
	defer tng.Unlock()

	return tng.targetSize, nil
}

// SetTargetSize sets target size for group. Function is used only in tests.
func (tng *TestNodeGroup) SetTargetSize(size int) {
	tng.Lock()
	defer tng.Unlock()
	tng.targetSize = size
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated.
func (tng *TestNodeGroup) IncreaseSize(delta int) error {
	tng.Lock()
	tng.targetSize += delta
	tng.Unlock()

	return tng.cloudProvider.onScaleUp(tng.id, delta)
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
func (tng *TestNodeGroup) Exist() bool {
	tng.Lock()
	defer tng.Unlock()
	return tng.exist
}

// Create creates the node group on the cloud provider side.
func (tng *TestNodeGroup) Create() (cloudprovider.NodeGroup, error) {
	if tng.Exist() {
		return nil, fmt.Errorf("group already exist")
	}
	newNodeGroup := tng.cloudProvider.AddAutoprovisionedNodeGroup(tng.id, tng.minSize, tng.maxSize, 0, tng.machineType)
	return newNodeGroup, tng.cloudProvider.onNodeGroupCreate(tng.id)
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
func (tng *TestNodeGroup) Delete() error {
	err := tng.cloudProvider.onNodeGroupDelete(tng.id)
	if err == nil {
		tng.cloudProvider.DeleteNodeGroup(tng.Id())
	}
	return err
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
func (tng *TestNodeGroup) DecreaseTargetSize(delta int) error {
	tng.Lock()
	tng.targetSize += delta
	tng.Unlock()

	return tng.cloudProvider.onScaleUp(tng.id, delta)
}

// DeleteNodes deletes nodes from this node group. Error is returned either on
// failure or if the given node doesn't belong to this node group. This function
// should wait until node group size is updated.
func (tng *TestNodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	tng.Lock()
	id := tng.id
	tng.targetSize -= len(nodes)
	tng.Unlock()
	for _, node := range nodes {
		err := tng.cloudProvider.onScaleDown(id, node.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

// Id returns an unique identifier of the node group.
func (tng *TestNodeGroup) Id() string {
	tng.Lock()
	defer tng.Unlock()

	return tng.id
}

// Debug returns a string containing all information regarding this node group.
func (tng *TestNodeGroup) Debug() string {
	tng.Lock()
	defer tng.Unlock()

	return fmt.Sprintf("%s target:%d min:%d max:%d", tng.id, tng.targetSize, tng.minSize, tng.maxSize)
}

// Nodes returns a list of all nodes that belong to this node group.
func (tng *TestNodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	tng.Lock()
	defer tng.Unlock()

	instances := make([]cloudprovider.Instance, 0)
	for node, nodegroup := range tng.cloudProvider.nodes {
		if nodegroup == tng.id {
			instances = append(instances, cloudprovider.Instance{Id: node})
		}
	}
	return instances, nil
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (tng *TestNodeGroup) Autoprovisioned() bool {
	return tng.autoprovisioned
}

// TemplateNodeInfo returns a node template for this node group.
func (tng *TestNodeGroup) TemplateNodeInfo() (*schedulerframework.NodeInfo, error) {
	if tng.cloudProvider.machineTemplates == nil {
		return nil, cloudprovider.ErrNotImplemented
	}
	if tng.autoprovisioned {
		template, found := tng.cloudProvider.machineTemplates[tng.machineType]
		if !found {
			return nil, fmt.Errorf("no template declared for %s", tng.machineType)
		}
		return template, nil
	}
	template, found := tng.cloudProvider.machineTemplates[tng.id]
	if !found {
		return nil, fmt.Errorf("no template declared for %s", tng.id)
	}
	return template, nil
}

// Labels returns labels passed to the test node group when it was created.
func (tng *TestNodeGroup) Labels() map[string]string {
	return tng.labels
}

// Taints returns taintspassed to the test node group when it was created.
func (tng *TestNodeGroup) Taints() []apiv1.Taint {
	return tng.taints
}

// MachineType returns machine type passed to the test node group when it was created.
func (tng *TestNodeGroup) MachineType() string {
	return tng.machineType
}
