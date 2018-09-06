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

package openstack

import (
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	schedulercache "k8s.io/kubernetes/pkg/scheduler/cache"
)

const (
	// ProviderName is the name of OpenStack cloud provider.
	ProviderName = "openstack"
)

// OpenStackCloudProvider implements CloudProvider interface.
type OpenStackCloudProvider struct {
	openstackManager OpenStackManager
	// This resource limiter is used if resource limits are not defined through cloud API.
	resourceLimiterFromFlags *cloudprovider.ResourceLimiter
}

// BuildOpenStackCloudProvider builds CloudProvider implementation for OpenStack.
func BuildOpenStackCloudProvider(openstackManager OpenStackManager, resourceLimiter *cloudprovider.ResourceLimiter) (*OpenStackCloudProvider, error) {
	return &OpenStackCloudProvider{openstackManager: openstackManager, resourceLimiterFromFlags: resourceLimiter}, nil
}

// Cleanup cleans up all resources before the cloud provider is removed
func (openstack *OpenStackCloudProvider) Cleanup() error {
	openstack.openstackManager.Cleanup()
	return nil
}

// Name returns name of the cloud provider.
func (openstack *OpenStackCloudProvider) Name() string {
	return ProviderName
}

// NodeGroups returns all node groups configured for this cloud provider.
func (openstack *OpenStackCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	asgs := openstack.openstackManager.GetASGs()
	result := make([]cloudprovider.NodeGroup, 0, len(asgs))
	for _, asg := range asgs {
		result = append(result, asg.Config)
	}
	return result
}

// NodeGroupForNode returns the node group for the given node.
func (openstack *OpenStackCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	ref, err := OpenStackRefFromProviderId(node.Spec.ProviderID)
	if err != nil {
		return nil, err
	}
	asg, err := openstack.openstackManager.GetASGForInstance(ref)
	return asg, err
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
func (openstack *OpenStackCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
func (openstack *OpenStackCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string,
	taints []apiv1.Taint, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (openstack *OpenStackCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	resourceLimiter, err := openstack.openstackManager.GetResourceLimiter()
	if err != nil {
		return nil, err
	}
	if resourceLimiter != nil {
		return resourceLimiter, nil
	}
	return openstack.resourceLimiterFromFlags, nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (openstack *OpenStackCloudProvider) Refresh() error {
	return openstack.openstackManager.Refresh()
}

// OpenStackRef contains s reference to some entity in OpenStack world.
type OpenStackRef struct {
	Project     string
	RootStack   string
	Stack       string
	Name        string
}

func (ref OpenStackRef) String() string {
	return fmt.Sprintf("%s/%s/%s/%s", ref.Project, ref.RootStack, ref.Stack, ref.Name)
}

// OpenStackRefFromProviderId creates InstanceConfig object
// from provider id which must be in format:
// openstack://<project-id>/<root_stack_id>/<stack_id>/<name>
func OpenStackRefFromProviderId(id string) (*OpenStackRef, error) {
	splitted := strings.Split(id[12:], "/")
	if len(splitted) != 4 {
		return nil, fmt.Errorf("Wrong id: expected format openstack://<project-id>/<root_stack_id>/<stack_id>/<name>, got %v", id)
	}
	return &OpenStackRef{
		Project:    splitted[0],
		RootStack:  splitted[1],
		Stack:      splitted[2],
		Name:       splitted[3],
	}, nil
}

// ASG implements NodeGroup interface.
type ASG interface {
	cloudprovider.NodeGroup

	OpenStackRef() OpenStackRef
}

type openstackASG struct {
	openstackRef OpenStackRef

	openstackManager OpenStackManager
	minSize    int
	maxSize    int
}

// OpenStackRef returns ASG's OpenStackRef
func (asg *openstackASG) OpenStackRef() OpenStackRef {
	return asg.openstackRef
}

// MaxSize returns maximum size of the node group.
func (asg *openstackASG) MaxSize() int {
	return asg.maxSize
}

// MinSize returns minimum size of the node group.
func (asg *openstackASG) MinSize() int {
	return asg.minSize
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kubernetes.
func (asg *openstackASG) TargetSize() (int, error) {
	size, err := asg.openstackManager.GetASGSize(asg)
	return int(size), err
}

// IncreaseSize increases ASG size
func (asg *openstackASG) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}
	size, err := asg.openstackManager.GetASGSize(asg)
	if err != nil {
		return err
	}
	if int(size)+delta > asg.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", int(size)+delta, asg.MaxSize())
	}
	return asg.openstackManager.SetASGSize(asg, size+int64(delta))
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
func (asg *openstackASG) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("size decrease must be negative")
	}
	size, err := asg.openstackManager.GetASGSize(asg)
	if err != nil {
		return err
	}
	nodes, err := asg.openstackManager.GetASGNodes(asg)
	if err != nil {
		return err
	}
	if int(size)+delta < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			size, delta, len(nodes))
	}
	return asg.openstackManager.SetASGSize(asg, size+int64(delta))
}

// Belongs returns true if the given node belongs to the NodeGroup.
func (asg *openstackASG) Belongs(node *apiv1.Node) (bool, error) {
	ref, err := OpenStackRefFromProviderId(node.Spec.ProviderID)
	if err != nil {
		return false, err
	}
	targetASG, err := asg.openstackManager.GetASGForInstance(ref)
	if err != nil {
		return false, err
	}
	if targetASG == nil {
		return false, fmt.Errorf("%s doesn't belong to a known asg", node.Name)
	}
	if targetASG.Id() != asg.Id() {
		return false, nil
	}
	return true, nil
}

// DeleteNodes deletes the nodes from the group.
func (asg *openstackASG) DeleteNodes(nodes []*apiv1.Node) error {
	size, err := asg.openstackManager.GetASGSize(asg)
	if err != nil {
		return err
	}
	if int(size) <= asg.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}
	refs := make([]*OpenStackRef, 0, len(nodes))
	for _, node := range nodes {

		belongs, err := asg.Belongs(node)
		if err != nil {
			return err
		}
		if !belongs {
			return fmt.Errorf("%s belong to a different asg than %s", node.Name, asg.Id())
		}
		openstackref, err := OpenStackRefFromProviderId(node.Spec.ProviderID)
		if err != nil {
			return err
		}
		refs = append(refs, openstackref)
	}
	return asg.openstackManager.DeleteInstances(refs)
}

// Id returns asg url.
func (asg *openstackASG) Id() string {
	return GenerateASGUrl(asg.openstackRef)
}

// Debug returns a debug string for the ASG.
func (asg *openstackASG) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", asg.Id(), asg.MinSize(), asg.MaxSize())
}

// Nodes returns a list of all nodes that belong to this node group.
func (asg *openstackASG) Nodes() ([]string, error) {
	return asg.openstackManager.GetASGNodes(asg)
}

// Exist checks if the node group really exists on the cloud provider side.
func (asg *openstackASG) Exist() bool {
	return true
}

// Create creates the node group on the cloud provider side.
func (asg *openstackASG) Create() (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Delete deletes the node group on the cloud provider side.
func (asg *openstackASG) Delete() error {
	return cloudprovider.ErrNotImplemented
}

// Autoprovisioned returns true if the node group is autoprovisioned.
func (asg *openstackASG) Autoprovisioned() bool {
	return false
}

// TemplateNodeInfo returns a node template for this node group.
func (asg *openstackASG) TemplateNodeInfo() (*schedulercache.NodeInfo, error) {
	node, err := asg.openstackManager.GetASGTemplateNode(asg)
	if err != nil {
		return nil, err
	}
	nodeInfo := schedulercache.NewNodeInfo(cloudprovider.BuildKubeProxy(asg.Id()))
	nodeInfo.SetNode(node)
	return nodeInfo, nil
}
