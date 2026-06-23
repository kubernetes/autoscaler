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
	"strings"
	"sync"
	"time"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/rand"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	fakek8s "k8s.io/autoscaler/cluster-autoscaler/utils/fake"
	testutils "k8s.io/autoscaler/cluster-autoscaler/utils/test"
	"k8s.io/autoscaler/cluster-autoscaler/utils/units"
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
	// nodeGroupForNodeWorksForDeletedNodes controls the behavior of CloudProvider.NodeGroupForNode(), so that different behaviors can be tested:
	// - [default] If false, NodeGroupForNode() returns the NodeGroup based on the CloudProvider.nodeToGroup map, which means it doesn't work for deleted Nodes.
	// - If true, NodeGroupForNode() returns the NodeGroup based on Node.ProviderID, which means it works for deleted Nodes.
	nodeGroupForNodeWorksForDeletedNodes bool
	// hasInstanceImplemented controls the behavior of CloudProvider.HasInstance(), so that different behaviors can be tested:
	// - [default] If false, HasInstance() is implemented and responds true for Nodes tracked by the fake CloudProvider until they're deleted via DeleteNodes().
	// - If true, HasInstance() always returns the ErrNotImplemented error. This is supported by CA, HasInstance() is an optional method.
	hasInstanceNotImplemented bool // The field is negated so that the default behavior with the field being false is "HasInstance() is implemented".

	// k8s is thread-safe, can be safely accessed without the CloudProvider mutex.
	k8s *fakek8s.Kubernetes
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
			// TODO: The fake cloud provider shouldn't have its own limits configuration for CPU and memory, it should honor AutoscalingOptions.MaxCoresTotal/MaxMemoryTotal.
			cloudprovider.ResourceNameCores:  1000000,
			cloudprovider.ResourceNameMemory: 1000000 * units.GiB,
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
// The method behaves differently based on the CloudProvider.nodeGroupForNodeWorksForDeletedNodes field,
// so that different behaviors can be tested (see the comment on the field for more details).
func (c *CloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	c.Lock()
	defer c.Unlock()

	if c.nodeGroupForNodeWorksForDeletedNodes {
		groupId, err := groupIdFromNodeProviderId(node.Spec.ProviderID)
		if err != nil {
			return nil, err
		}
		return c.groups[groupId], nil
	}

	groupId, ok := c.nodeToGroup[node.Name]
	if !ok {
		return nil, nil
	}
	return c.groups[groupId], nil
}

// HasInstance returns true if the given node is managed by this cloud provider, until the Node is deleted via DeleteNodes().
// The method behaves differently based on the CloudProvider.nodeGroupForNodeWorksForDeletedNodes field,
// so that different behaviors can be tested (see the comment on the field for more details).
func (c *CloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	c.Lock()
	defer c.Unlock()

	if c.hasInstanceNotImplemented {
		return false, cloudprovider.ErrNotImplemented
	}

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
func WithNode(templateNode *apiv1.Node) NodeGroupOption {
	return func(n *NodeGroup) {
		n.setTemplateFromNodeAndPods(templateNode)
		addedNode := n.addNodeFromTemplate(templateNode.Name)
		if n.provider.k8s != nil {
			n.provider.k8s.AddNode(addedNode)
		}
	}
}

// WithNodes configures the provided Node as the template for the NodeGroup, and adds nodeCount copies of the template to the NodeGroup.
func WithNodes(templateNode *apiv1.Node, nodeCount int) NodeGroupOption {
	return func(n *NodeGroup) {
		n.setTemplateFromNodeAndPods(templateNode)
		for i := range nodeCount {
			addedNode := n.addNodeFromTemplate(fmt.Sprintf("%s-node-%d", n.id, i))
			if n.provider.k8s != nil {
				n.provider.k8s.AddNode(addedNode)
			}
		}
	}
}

// WithNGSize sets the minimum and maximum size of the node group.
func WithNGSize(min, max int) NodeGroupOption {
	return func(n *NodeGroup) {
		n.minSize = min
		n.maxSize = max

	}
}

// WithTemplate sets the node template for the node group.
func WithTemplate(template *framework.NodeInfo) NodeGroupOption {
	return func(n *NodeGroup) {
		n.template = template
	}
}

// WithNodeGarbageCollectionDelay can be used to simulate Node objects hanging around in the K8s API for some time after their VMs are deleted.
func WithNodeGarbageCollectionDelay(delay time.Duration) NodeGroupOption {
	return func(n *NodeGroup) {
		n.nodeGarbageCollectionDelay = delay
	}
}

// WithNodeRegistrationDelay can be used to simulate Node objects taking some time to appear in the K8s API after their VMs are created.
func WithNodeRegistrationDelay(delay time.Duration) NodeGroupOption {
	return func(n *NodeGroup) {
		n.nodeRegistrationDelay = delay
	}
}

// WithNodeReadinessDelay can be used to simulate Node objects taking some time to transition to Ready after they appear in the K8s API. If not used, the Nodes
// will appear as Ready immediately after registration.
func WithNodeReadinessDelay(delay time.Duration) NodeGroupOption {
	return func(n *NodeGroup) {
		n.nodeReadinessDelay = delay
	}
}

// AddNodeGroup is a helper for tests to add a group with its template.
func (c *CloudProvider) AddNodeGroup(id string, opts ...NodeGroupOption) *NodeGroup {
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

	return group
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
}

// SetResourceLimit allows the test to reach in and change the limits.
func (c *CloudProvider) SetResourceLimit(resource string, min, max int64) {
	c.Lock()
	defer c.Unlock()
	c.minLimits[resource] = min
	c.maxLimits[resource] = max
}

// ConfigureNodeGroupForNodeBehavior configures the behavior of CloudProvider.NodeGroupForNode(). See the comment on CloudProvider.nodeGroupForNodeWorksForDeletedNodes for more details.
func (c *CloudProvider) ConfigureNodeGroupForNodeBehavior(worksForDeletedNodes bool) {
	c.Lock()
	defer c.Unlock()
	c.nodeGroupForNodeWorksForDeletedNodes = worksForDeletedNodes
}

// ConfigureHasInstanceBehavior configures the behavior of CloudProvider.HasInstance(). See the comment on CloudProvider.hasInstanceNotImplemented for more details.
func (c *CloudProvider) ConfigureHasInstanceBehavior(notImplemented bool) {
	c.Lock()
	defer c.Unlock()
	// The field is negated so that the default behavior with the field being false is "HasInstance() is implemented". The argument to this method is also negated so that
	// they match 1-1 and the comment on the field works for this method as well.
	c.hasInstanceNotImplemented = notImplemented
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

	// nodeGarbageCollectionDelay can be used to simulate Node objects hanging around in the K8s API for some time after their VMs are deleted.
	nodeGarbageCollectionDelay time.Duration
	// nodeRegistrationDelay can be used to simulate Node objects taking some time to appear in the K8s API after their VMs are created.
	nodeRegistrationDelay time.Duration
	// nodeReadinessDelay can be used to simulate Node objects taking some time to transition to Ready after they appear in the K8s API. If unset,
	// the Nodes will appear as Ready immediately after registration.
	nodeReadinessDelay time.Duration
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
			delete(n.instances, node.Spec.ProviderID)
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

	// Remove the Nodes from the K8s fake if it's configured.
	if n.provider.k8s == nil {
		return nil
	}
	if n.nodeGarbageCollectionDelay == 0 {
		// No delays configured, synchronously delete the Nodes from the K8s fake.
		for _, node := range nodes {
			n.provider.k8s.DeleteNode(node.Name)
		}
	} else {
		// Node garbage collection delay is configured, handle deleting the Nodes from the K8s fake asynchronously.
		// The K8s fake is thread safe, so we don't need the provider lock to access it - should be safe to use in a goroutine.
		go simulateK8sApiNodeDeletion(n.provider.k8s, nodes, n.nodeGarbageCollectionDelay)
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
func (n *NodeGroup) TargetSize() (int, error) {
	n.Lock()
	defer n.Unlock()
	return n.targetSize, nil
}

// IncreaseSize adds nodes to the node group and updates internal instance mapping.
func (n *NodeGroup) IncreaseSize(delta int) error {
	n.Lock()
	defer n.Unlock()

	if n.targetSize+delta > n.maxSize {
		return fmt.Errorf("size too large")
	}
	if n.template == nil || n.template.Node() == nil {
		return fmt.Errorf("node group %s has no template to create new nodes", n.id)
	}

	n.provider.Lock()
	defer n.provider.Unlock()

	var newNodes []*apiv1.Node
	for i := 0; i < delta; i++ {
		nodeName := fmt.Sprintf("%s-node-%s", n.id, rand.String(4))
		newNode := n.addNodeFromTemplate(nodeName)
		newNodes = append(newNodes, newNode)
	}

	// Add the new Nodes to the K8s fake if it's configured.
	if n.provider.k8s == nil {
		return nil
	}
	if n.nodeRegistrationDelay == 0 && n.nodeReadinessDelay == 0 {
		// No delays configured, synchronously add the Nodes as Ready to the K8s fake.
		for _, node := range newNodes {
			testutils.IsReady(true)(node)
			n.provider.k8s.AddNode(node)
		}
	} else {
		// Node registration/readiness delays are configured, handle adding the Nodes to the K8s fake asynchronously.
		// The K8s fake is thread safe, so we don't need the provider lock to access it - should be safe to use in a goroutine.
		go simulateK8sApiNodeRegistration(n.provider.k8s, newNodes, n.nodeRegistrationDelay, n.nodeReadinessDelay)
	}

	return nil
}

// TemplateNodeInfo returns the template node information for this node group.
func (n *NodeGroup) TemplateNodeInfo() (*framework.NodeInfo, error) {
	n.Lock()
	defer n.Unlock()
	if n.template == nil {
		return nil, cloudprovider.ErrNotImplemented
	}
	return n.template, nil
}

// GetTargetSize returns the target size as a raw integer (helper method).
func (n *NodeGroup) GetTargetSize() int {
	n.Lock()
	defer n.Unlock()
	return n.targetSize
}

// setTemplateFromNodeAndPods configures the result of TemplateNodeInfo() based on a copy of the provided Node and the provided Pods.
// Should be called with the NodeGroup mutex already locked, or during NodeGroup object initialization without the lock.
func (n *NodeGroup) setTemplateFromNodeAndPods(node *apiv1.Node, pods ...*apiv1.Pod) {
	templateNode := cloneNodeForNodeGroup(node, n.id, fmt.Sprintf("%s-template", n.id))
	n.template = framework.NewTestNodeInfo(templateNode, pods...)
}

// addNode adds a deep-copy of the provided Node to this NodeGroup and its CloudProvider. The added Node is returned from the method.
// Should be called with the NodeGroup mutex already locked, or during NodeGroup object initialization without the lock.
func (n *NodeGroup) addNodeFromTemplate(nodeName string) *apiv1.Node {
	node := cloneNodeForNodeGroup(n.template.Node(), n.id, nodeName)
	n.instances[node.Spec.ProviderID] = cloudprovider.InstanceRunning
	n.provider.nodeToGroup[node.Name] = n.id
	n.targetSize += 1
	return node
}

// cloneNodeForNodeGroup returns a deep-copy of the provided templateNode, to be tracked by the fake CloudProvider/NodeGroup.
// Any Node used with the fake CloudProvider should go through this function.
func cloneNodeForNodeGroup(templateNode *apiv1.Node, ngName, nodeName string) *apiv1.Node {
	node := templateNode.DeepCopy()
	node.Name = nodeName
	node.Spec.ProviderID = fmt.Sprintf("fake-provider/%s/%s", ngName, node.Name)
	return node
}

// groupIdFromNodeProviderId parses the Spec.ProviderID set by cloneNodeForNodeGroup() and extracts the NodeGroup id from it.
func groupIdFromNodeProviderId(nodeProviderId string) (string, error) {
	parts := strings.Split(nodeProviderId, "/")
	if len(parts) != 3 {
		return "", fmt.Errorf("ngIdFromNodeProviderId(): nodeProviderId should be in the format <provider_name>/<ng_id>/<node_name>, got %s", nodeProviderId)
	}
	return parts[1], nil
}

func simulateK8sApiNodeDeletion(k8s *fakek8s.Kubernetes, nodes []*apiv1.Node, garbageCollectionDelay time.Duration) {
	if garbageCollectionDelay > 0 {
		// If configured, simulate a delay between the VM deletion call and the Nodes disappearing.
		time.Sleep(garbageCollectionDelay)
	}

	for _, node := range nodes {
		k8s.DeleteNode(node.Name)
	}
}

func simulateK8sApiNodeRegistration(k8s *fakek8s.Kubernetes, nodes []*apiv1.Node, registrationDelay, readinessDelay time.Duration) {
	if registrationDelay > 0 {
		// If configured, simulate a delay between the VM creation call and the Node registering in the K8s API.
		time.Sleep(registrationDelay)
	}

	// If no readinessDelay is configured, the Nodes just start as Ready. Otherwise, they start as not-Ready, and are updated to Ready after readinessDelay.
	nodesInitiallyReady := readinessDelay == 0
	for _, node := range nodes {
		testutils.IsReady(nodesInitiallyReady)(node)
		k8s.AddNode(node)
	}

	if readinessDelay == 0 {
		return
	}
	// If configured, simulate a delay between the Node registering in the K8s API and the Node becoming Ready.
	time.Sleep(readinessDelay)
	for _, node := range nodes {
		// We injected the Node object to the API earlier, so something could be reading it asynchronously - we need to deep copy before modifying the readiness.
		nodeCopy := node.DeepCopy()
		testutils.IsReady(true)(nodeCopy)
		k8s.UpdateNode(nodeCopy)
	}
}
