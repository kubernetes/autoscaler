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
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"
)

// TestCloudProvider is a dummy cloud provider to be used in tests.
type TestCloudProvider struct {
	sync.Mutex
	nodes      map[string]string
	groups     map[string]cloudprovider.NodeGroup
	onIncrease func(string, int) error
	onDelete   func(string, string) error
}

// OnIncreaseFunc is a function called on node group increase in TestCloudProvider.
// First parameter is the NodeGroup id, second is the increase delta.
type OnIncreaseFunc func(string, int) error

// OnDeleteFunc is a function called on cluster
type OnDeleteFunc func(string, string) error

// NewTestCloudProvider builds new TestCloudProvider
func NewTestCloudProvider(onIncrease OnIncreaseFunc, onDelete OnDeleteFunc) *TestCloudProvider {
	return &TestCloudProvider{
		nodes:      make(map[string]string),
		groups:     make(map[string]cloudprovider.NodeGroup),
		onIncrease: onIncrease,
		onDelete:   onDelete,
	}
}

// Name returns name of the cloud provider.
func (tcp *TestCloudProvider) Name() string {
	return "TestCloudProvider"
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
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvilableMachineTypes get all machine types that can be requested from the cloud provider.
func (tcp *TestCloudProvider) GetAvilableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
func (tcp *TestCloudProvider) NewNodeGroup(name string, machineType string, labels map[string]string, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// AddNodeGroup adds node group to test cloud provider.
func (tcp *TestCloudProvider) AddNodeGroup(id string, min int, max int, size int) {
	tcp.Lock()
	defer tcp.Unlock()

	tcp.groups[id] = &TestNodeGroup{
		cloudProvider: tcp,
		id:            id,
		minSize:       min,
		maxSize:       max,
		targetSize:    size,
	}
}

// AddNode adds the given node to the group.
func (tcp *TestCloudProvider) AddNode(nodeGroupId string, node *apiv1.Node) {
	tcp.Lock()
	defer tcp.Unlock()
	tcp.nodes[node.Name] = nodeGroupId
}

// TestNodeGroup is a node group used by TestCloudProvider.
type TestNodeGroup struct {
	sync.Mutex
	cloudProvider *TestCloudProvider
	id            string
	maxSize       int
	minSize       int
	targetSize    int
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

// IncreaseSize increases the size of the node group. To delete a node you need
// to explicitly name it and use DeleteNode. This function should wait until
// node group size is updated.
func (tng *TestNodeGroup) IncreaseSize(delta int) error {
	tng.Lock()
	tng.targetSize += delta
	tng.Unlock()

	return tng.cloudProvider.onIncrease(tng.id, delta)
}

// Exist checks if the node group really exists on the cloud provider side. Allows to tell the
// theoretical node group from the real one.
func (tng *TestNodeGroup) Exist() (bool, error) {
	return true, nil
}

// Create creates the node group on the cloud provider side.
func (tng *TestNodeGroup) Create() error {
	return cloudprovider.ErrAlreadyExist
}

// Delete deletes the node group on the cloud provider side.
// This will be executed only for autoprovisioned node groups, once their size drops to 0.
func (tng *TestNodeGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
func (tng *TestNodeGroup) DecreaseTargetSize(delta int) error {
	tng.Lock()
	tng.targetSize += delta
	tng.Unlock()

	return tng.cloudProvider.onIncrease(tng.id, delta)
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
		err := tng.cloudProvider.onDelete(id, node.Name)
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
func (tng *TestNodeGroup) Nodes() ([]string, error) {
	tng.Lock()
	defer tng.Unlock()

	result := make([]string, 0)
	for node, nodegroup := range tng.cloudProvider.nodes {
		if nodegroup == tng.id {
			result = append(result, node)
		}
	}
	return result, nil
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (tng *TestNodeGroup) Autoprovisioned() bool {
	return false
}

// TemplateNodeInfo returns a node template for this node group.
func (tng *TestNodeGroup) TemplateNodeInfo() (*schedulercache.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}
