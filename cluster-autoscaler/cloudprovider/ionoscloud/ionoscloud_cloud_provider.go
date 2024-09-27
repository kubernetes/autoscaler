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

package ionoscloud

import (
	"errors"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	caerrors "k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/gpu"
	"k8s.io/klog/v2"
)

const (
	// GPULabel is the label added to nodes with GPU resource.
	GPULabel = ""
)

type nodePool struct {
	manager IonosCloudManager

	id  string
	min int
	max int
}

var _ cloudprovider.NodeGroup = &nodePool{}

// MaxSize returns maximum size of the node group.
func (n *nodePool) MaxSize() int {
	return n.max
}

// MinSize returns minimum size of the node group.
func (n *nodePool) MinSize() int {
	return n.min
}

// TargetSize returns the current target size of the node group. It is possible that the
// number of nodes in Kubernetes is different at the moment but should be equal
// to Size() once everything stabilizes (new nodes finish startup and registration or
// removed nodes are deleted completely). Implementation required.
func (n *nodePool) TargetSize() (int, error) {
	size, err := n.manager.GetNodeGroupTargetSize(n)
	return size, err
}

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated. Implementation required.
func (n *nodePool) IncreaseSize(delta int) error {
	if delta <= 0 {
		return errors.New("size increase must be positive")
	}
	size, err := n.manager.GetNodeGroupSize(n)
	if err != nil {
		return err
	}
	targetSize := size + delta
	if targetSize > n.max {
		return fmt.Errorf("size increase exceeds upper bound of %d", n.max)
	}
	return n.manager.SetNodeGroupSize(n, targetSize)
}

// AtomicIncreaseSize is not implemented.
func (n *nodePool) AtomicIncreaseSize(delta int) error {
	return cloudprovider.ErrNotImplemented
}

// DeleteNodes deletes nodes from this node group (and also decreasing the size
// of the node group with that). Error is returned either on failure or if the
// given node doesn't belong to this node group. This function should wait
// until node group size is updated. Implementation required.
func (n *nodePool) DeleteNodes(nodes []*apiv1.Node) error {
	if acquired := n.manager.TryLockNodeGroup(n); !acquired {
		return errors.New("node deletion already in progress")
	}
	defer n.manager.UnlockNodeGroup(n)

	for _, node := range nodes {
		nodeID := convertToNodeID(node.Spec.ProviderID)
		if err := n.manager.DeleteNode(n, nodeID); err != nil {
			return err
		}
	}
	return nil
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes when there
// is an option to just decrease the target size. Implementation required.
func (n *nodePool) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return errors.New("size decrease must be negative")
	}
	size, err := n.manager.GetNodeGroupTargetSize(n)
	if err != nil {
		return err
	}
	if size+delta < n.min {
		return fmt.Errorf("size decrease exceeds lower bound of %d", n.min)
	}
	// IonosCloud does not allow modification of the target size while nodes are being provisioned.
	return errors.New("currently not supported behavior")
}

// Id returns an unique identifier of the node group.
func (n *nodePool) Id() string {
	return n.id
}

// Debug returns a string containing all information regarding this node group.
func (n *nodePool) Debug() string {
	return fmt.Sprintf("ID=%s, Min=%d, Max=%d", n.id, n.min, n.max)
}

// Nodes returns a list of all nodes that belong to this node group.
// It is required that Instance objects returned by this method have Id field set.
// Other fields are optional.
// This list should include also instances that might have not become a kubernetes node yet.
func (n *nodePool) Nodes() ([]cloudprovider.Instance, error) {
	return n.manager.GetInstancesForNodeGroup(n)
}

// TemplateNodeInfo returns a framework.NodeInfo structure of an empty
// (as if just started) node. This will be used in scale-up simulations to
// predict what would a new node look like if a node group was expanded. The
// returned NodeInfo is expected to have a fully populated Node object, with
// all of the labels, capacity and allocatable information as well as all pods
// that are started on the node by default, using manifest (most likely only
// kube-proxy). Implementation optional.
func (n *nodePool) TemplateNodeInfo() (*framework.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one. Implementation required.
func (n *nodePool) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side. Implementation optional.
func (n *nodePool) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
// Implementation optional.
func (n *nodePool) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned. An autoprovisioned group
// was created by CA and can be deleted when scaled to 0.
func (n *nodePool) Autoprovisioned() bool {
	return false
}

// GetOptions returns NodeGroupAutoscalingOptions that should be used for this particular
// NodeGroup. Returning a nil will result in using default options.
func (n *nodePool) GetOptions(defaults config.NodeGroupAutoscalingOptions) (*config.NodeGroupAutoscalingOptions, error) {
	return nil, nil
}

// IonosCloudCloudProvider implements cloudprovider.CloudProvider.
type IonosCloudCloudProvider struct {
	manager         IonosCloudManager
	resourceLimiter *cloudprovider.ResourceLimiter
}

var _ cloudprovider.CloudProvider = &IonosCloudCloudProvider{}

// BuildIonosCloudCloudProvider builds CloudProvider implementation for Ionos Cloud.
func BuildIonosCloudCloudProvider(manager IonosCloudManager, rl *cloudprovider.ResourceLimiter) *IonosCloudCloudProvider {
	return &IonosCloudCloudProvider{manager: manager, resourceLimiter: rl}
}

// Name returns name of the cloud provider.
func (ic *IonosCloudCloudProvider) Name() string {
	return cloudprovider.IonoscloudProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (ic *IonosCloudCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	return ic.manager.GetNodeGroups()
}

// NodeGroupForNode returns the node group for the given node, nil if the node
// should not be processed by cluster autoscaler, or non-nil error if such
// occurred. Must be implemented.
func (ic *IonosCloudCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	providerID := node.Spec.ProviderID
	if nodeGroup := ic.manager.GetNodeGroupForNode(node); nodeGroup != nil {
		klog.V(5).Infof("Found cached node group entry %s for node %s", nodeGroup.Id(), providerID)
		return nodeGroup, nil
	}
	klog.V(4).Infof("No cached node group entry found for node %s", providerID)

	for _, nodeGroup := range ic.manager.GetNodeGroups() {
		klog.V(5).Infof("Checking node group %s", nodeGroup.Id())
		nodes, err := nodeGroup.Nodes()
		if err != nil {
			return nil, err
		}

		for _, node := range nodes {
			if node.Id != providerID {
				continue
			}
			klog.V(4).Infof("Found node group %s for node %s after refresh", nodeGroup.Id(), providerID)
			return nodeGroup, nil
		}
	}

	// there is no "ErrNotExist" error, so we have to return a nil error
	return nil, nil
}

// HasInstance returns whether a given node has a corresponding instance in this cloud provider
func (ic *IonosCloudCloudProvider) HasInstance(node *apiv1.Node) (bool, error) {
	return true, cloudprovider.ErrNotImplemented
}

// Pricing returns pricing model for this cloud provider or error if not
// available. Implementation optional.
func (ic *IonosCloudCloudProvider) Pricing() (cloudprovider.PricingModel, caerrors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from
// the cloud provider. Implementation optional.
func (ic *IonosCloudCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// NewNodeGroup builds a theoretical node group based on the node definition
// provided. The node group is not automatically created on the cloud provider
// side. The node group is not returned by NodeGroups() until it is created.
// Implementation optional.
func (ic *IonosCloudCloudProvider) NewNodeGroup(
	machineType string,
	labels map[string]string,
	systemLabels map[string]string,
	taints []apiv1.Taint,
	extraResources map[string]resource.Quantity,
) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for
// resources (cores, memory etc.).
func (ic *IonosCloudCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return ic.resourceLimiter, nil
}

// GPULabel returns the label added to nodes with GPU resource.
func (ic *IonosCloudCloudProvider) GPULabel() string {
	return GPULabel
}

// GetAvailableGPUTypes return all available GPU types cloud provider supports.
func (ic *IonosCloudCloudProvider) GetAvailableGPUTypes() map[string]struct{} {
	return nil
}

// GetNodeGpuConfig returns the label, type and resource name for the GPU added to node. If node doesn't have
// any GPUs, it returns nil.
func (ic *IonosCloudCloudProvider) GetNodeGpuConfig(node *apiv1.Node) *cloudprovider.GpuConfig {
	return gpu.GetNodeGPUFromCloudProvider(ic, node)
}

// Cleanup cleans up read resources before the cloud provider is destroyed,
// i.e. go routines etc.
func (ic *IonosCloudCloudProvider) Cleanup() error {
	return nil
}

// Refresh is called before every main loop and can be used to dynamically
// update cloud provider state. In particular the list of node groups returned
// by NodeGroups() can change as a result of CloudProvider.Refresh().
func (ic *IonosCloudCloudProvider) Refresh() error {
	// Currently only static node groups are supported.
	return nil
}

// BuildIonosCloud builds the IonosCloud cloud provider.
func BuildIonosCloud(
	opts config.AutoscalingOptions,
	_ cloudprovider.NodeGroupDiscoveryOptions,
	rl *cloudprovider.ResourceLimiter,
) cloudprovider.CloudProvider {
	manager, err := CreateIonosCloudManager(opts.NodeGroups, opts.UserAgent)
	if err != nil {
		klog.Fatalf("Failed to create IonosCloud cloud provider: %v", err)
	}

	provider := BuildIonosCloudCloudProvider(manager, rl)
	RegisterMetrics()
	return provider
}
