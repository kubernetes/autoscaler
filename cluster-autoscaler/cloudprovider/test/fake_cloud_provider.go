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
	"fmt"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	fakek8s "k8s.io/autoscaler/cluster-autoscaler/utils/fake"
	"sync"
)

const (
	defaultMinSize = 0
	defaultMaxSize = 1000
)

// CloudProvider is a fake implementation of the cloudprovider interface for testing.
type CloudProvider struct {
	sync.RWMutex
	groups    map[string]cloudprovider.NodeGroup
	minLimits map[string]int64
	maxLimits map[string]int64
	// nodeToGroup tracks which node name belongs to which group ID.
	nodeToGroup map[string]string
	k8s         *fakek8s.Kubernetes
}

// CloudProviderOption defines a function to configure the CloudProvider.
type CloudProviderOption func(*CloudProvider)

// NewCloudProvider creates a new instance of the fake CloudProvider.
func NewCloudProvider(k8s *fakek8s.Kubernetes) *CloudProvider {
	return &CloudProvider{
		groups:      make(map[string]cloudprovider.NodeGroup),
		nodeToGroup: make(map[string]string),
		minLimits: map[string]int64{
			cloudprovider.ResourceNameCores:  0,
			cloudprovider.ResourceNameMemory: 0,
		},
		maxLimits: map[string]int64{
			// Set to a effectively infinite number for tests.
			cloudprovider.ResourceNameCores:  1000000,
			cloudprovider.ResourceNameMemory: 1000000,
		},
		k8s: k8s,
	}
}

// NodeGroups returns all node groups configured in the fake CloudProvider.
func (c *CloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	c.Lock()
	defer c.Unlock()
	var res []cloudprovider.NodeGroup
	for _, g := range c.groups {
		res = append(res, g)
	}
	return res
}

// NodeGroupForNode returns the node group that a given node belongs to.
func (c *CloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	c.Lock()
	defer c.Unlock()
	groupId, ok := c.nodeToGroup[node.Name]
	if !ok {
		return nil, nil
	}
	return c.groups[groupId], nil
}

// HasInstance returns true if the given node is managed by this cloud provider.
func (c *CloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	c.Lock()
	defer c.Unlock()
	_, found := c.nodeToGroup[node.Name]
	return found, nil
}

// GetResourceLimiter generates a new limiter based on our current internal maps.
func (c *CloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	c.Lock()
	defer c.Unlock()
	return cloudprovider.NewResourceLimiter(c.minLimits, c.maxLimits), nil
}

// GPULabel returns the label used to identify GPU types in this provider.
func (c *CloudProvider) GPULabel() string { return "gpu-label" }

// GetAvailableGPUTypes returns a map of all GPU types available in this provider.
func (c *CloudProvider) GetAvailableGPUTypes() map[string]struct{} { return nil }

// GetNodeGpuConfig returns the GPU configuration for a specific node.
func (c *CloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig { return nil }

// Cleanup performs any necessary teardown of the CloudProvider.
func (c *CloudProvider) Cleanup() error { return nil }

// Refresh updates the internal state of the CloudProvider.
func (c *CloudProvider) Refresh() error { return nil }

// Name returns the name of the cloud provider.
func (c *CloudProvider) Name() string { return "Provider" }

// Pricing returns the pricing model associated with the provider.
func (c *CloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes returns the machine types supported by the provider.
func (c *CloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// NewNodeGroup creates a new node group based on the provided specifications.
func (c *CloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// NodeGroupOption is a function that configures a NodeGroup during creation.
type NodeGroupOption func(*NodeGroup)

// WithNode adds a single initial node to the group and
// automatically sets the group's template based on that node.
func WithNode(node *apiv1.Node) NodeGroupOption {
	return func(n *NodeGroup) {
		n.provider.nodeToGroup[node.Name] = n.id
		n.instances[node.Name] = cloudprovider.InstanceRunning
		n.targetSize = 1
		n.template = framework.NewTestNodeInfo(node.DeepCopy())
		if n.provider.k8s != nil {
			n.provider.k8s.AddNode(node)
		}
	}
}

// WithMinSize sets the minimum size of the node group.
func WithMinSize(min int) NodeGroupOption {
	return func(n *NodeGroup) {
		n.minSize = min
	}
}

// WithMaxSize sets the maximum size of the node group.
func WithMaxSize(max int) NodeGroupOption {
	return func(n *NodeGroup) {
		n.maxSize = max
	}
}

// WithTemplate sets the node template for the node group.
func WithTemplate(template *framework.NodeInfo) NodeGroupOption {
	return func(n *NodeGroup) {
		n.template = template
	}
}

// AddNodeGroup is a helper for tests to add a group with its template.
func (c *CloudProvider) AddNodeGroup(id string, opts ...NodeGroupOption) {
	c.Lock()
	defer c.Unlock()

	group := &NodeGroup{
		id:         id,
		minSize:    defaultMinSize,
		maxSize:    defaultMaxSize,
		targetSize: 0,
		instances:  make(map[string]cloudprovider.InstanceState),
		provider:   c,
	}

	for _, opt := range opts {
		opt(group)
	}
	c.groups[id] = group
}

// GetNodeGroup is a helper for tests to get a node group.
func (c *CloudProvider) GetNodeGroup(id string) cloudprovider.NodeGroup {
	c.Lock()
	defer c.Unlock()
	return c.groups[id]
}

// AddNode connects a node name to a group ID.
func (c *CloudProvider) AddNode(groupId string, node *apiv1.Node) {
	c.Lock()
	defer c.Unlock()
	c.nodeToGroup[node.Name] = groupId
	if g, ok := c.groups[groupId].(*NodeGroup); ok {
		g.Lock()
		defer g.Unlock()
		g.instances[node.Name] = cloudprovider.InstanceRunning
		g.targetSize++
	}
}

// SetResourceLimit allows the test to reach in and change the limits.
func (c *CloudProvider) SetResourceLimit(resource string, min, max int64) {
	c.Lock()
	defer c.Unlock()
	c.minLimits[resource] = min
	c.maxLimits[resource] = max
}

// NodeGroup is a fake implementation of the cloudprovider.NodeGroup interface for testing.
type NodeGroup struct {
	sync.RWMutex
	id         string
	minSize    int
	maxSize    int
	targetSize int
	template   *framework.NodeInfo
	// instances maps instanceID -> state.
	instances map[string]cloudprovider.InstanceState
	provider  *CloudProvider
}

// MaxSize returns the maximum size of the node group.
func (n *NodeGroup) MaxSize() int {
	return n.maxSize
}

// MinSize returns the minimum size of the node group.
func (n *NodeGroup) MinSize() int {
	return n.minSize
}

// AtomicIncreaseSize is a version of IncreaseSize that increases the size of the node group atomically.
func (n *NodeGroup) AtomicIncreaseSize(delta int) error {
	return n.IncreaseSize(delta)
}

// DeleteNodes removes specific nodes from the node group and updates the internal mapping.
func (n *NodeGroup) DeleteNodes(nodes []*apiv1.Node) error {
	n.Lock()
	defer n.Unlock()

	n.provider.Lock()
	defer n.provider.Unlock()

	deletedCount := 0
	for _, node := range nodes {
		if groupId, exists := n.provider.nodeToGroup[node.Name]; exists && groupId == n.id {
			delete(n.provider.nodeToGroup, node.Name)
			delete(n.instances, node.Name)
			if n.provider.k8s != nil {
				n.provider.k8s.DeleteNode(node.Name)
			}
			deletedCount++
		} else {
			fmt.Printf("Warning: node %s not found in group %s or already deleted.", node.Name, n.id)
		}
	}

	if n.targetSize >= deletedCount {
		n.targetSize -= deletedCount
	} else {
		n.targetSize = 0
	}

	return nil
}

// ForceDeleteNodes deletes nodes without checking for specific conditions (fake implementation).
func (n *NodeGroup) ForceDeleteNodes(nodes []*apiv1.Node) error {
	return n.DeleteNodes(nodes)
}

// DecreaseTargetSize reduces the target size of the node group by the specified delta.
func (n *NodeGroup) DecreaseTargetSize(delta int) error {
	n.Lock()
	defer n.Unlock()
	n.targetSize -= delta
	return nil
}

// Id returns the unique identifier of the node group.
func (n *NodeGroup) Id() string {
	return n.id
}

// Debug returns a string representation of the node group's current state.
func (n *NodeGroup) Debug() string {
	return fmt.Sprintf("NodeGroup{id: %s, targetSize: %d}", n.id, n.targetSize)
}

// Nodes returns a list of all instances currently existing in this node group.
func (n *NodeGroup) Nodes() ([]cloudprovider.Instance, error) {
	n.provider.Lock()
	defer n.provider.Unlock()

	var instances []cloudprovider.Instance
	for id, state := range n.instances {
		instances = append(instances, cloudprovider.Instance{
			Id: id,
			Status: &cloudprovider.InstanceStatus{
				State: state,
			},
		})
	}
	return instances, nil
}

// Exist returns true if the node group currently exists in the cloud provider.
func (n *NodeGroup) Exist() bool {
	return true
}

// Create creates the node group in the cloud provider (not implemented).
func (n *NodeGroup) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group from the cloud provider (not implemented).
func (n *NodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (n *NodeGroup) Autoprovisioned() bool {
	return false
}

// GetOptions returns autoscaling options specific to this node group.
func (n *NodeGroup) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, nil
}

// TargetSize returns the current target size of the node group.
func (n *NodeGroup) TargetSize() (int, error) { return n.targetSize, nil }

// IncreaseSize adds nodes to the node group and updates internal instance mapping.
func (n *NodeGroup) IncreaseSize(delta int) error {
	n.Lock()
	defer n.Unlock()
	if n.targetSize+delta > n.maxSize {
		return fmt.Errorf("size too large")
	}

	n.provider.Lock()
	defer n.provider.Unlock()

	for i := 0; i < delta; i++ {
		instanceNum := n.targetSize + i
		instanceId := fmt.Sprintf("%s-node-%d", n.id, instanceNum)

		if n.template == nil || n.template.Node() == nil {
			return fmt.Errorf("node group %s has no template to create new nodes", n.id)
		}
		newNode := n.template.Node().DeepCopy()
		newNode.Name = instanceId
		newNode.Spec.ProviderID = instanceId

		n.instances[instanceId] = cloudprovider.InstanceRunning
		n.provider.nodeToGroup[instanceId] = n.id
		if n.provider.k8s != nil {
			n.provider.k8s.AddNode(newNode)
		}
	}
	n.targetSize += delta
	return nil
}

// TemplateNodeInfo returns the template node information for this node group.
func (n *NodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	if n.template == nil {
		return nil, cloudprovider.ErrNotImplemented
	}
	return n.template, nil
}

// GetTargetSize returns the target size as a raw integer (helper method).
func (n *NodeGroup) GetTargetSize() int { return n.targetSize }
