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

package azure

import (
	"fmt"
	"strconv"
	"strings"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	apiv1 "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"
)

// AzureCloudProvider provides implementation of CloudProvider interface for Azure.
type AzureCloudProvider struct {
	azureManager *AzureManager
	scaleSets    []*ScaleSet
}

// BuildAzureCloudProvider creates new AzureCloudProvider
func BuildAzureCloudProvider(azureManager *AzureManager, specs []string) (*AzureCloudProvider, error) {
	azure := &AzureCloudProvider{
		azureManager: azureManager,
	}
	for _, spec := range specs {
		if err := azure.addNodeGroup(spec); err != nil {
			return nil, err
		}
	}

	return azure, nil
}

// addNodeGroup adds node group defined in string spec. Format:
// minNodes:maxNodes:scaleSetName
func (azure *AzureCloudProvider) addNodeGroup(spec string) error {
	scaleSet, err := buildScaleSet(spec, azure.azureManager)
	if err != nil {
		return err
	}
	azure.scaleSets = append(azure.scaleSets, scaleSet)
	azure.azureManager.RegisterScaleSet(scaleSet)
	return nil
}

// Name returns name of the cloud provider.
func (azure *AzureCloudProvider) Name() string {
	return "azure"
}

// NodeGroups returns all node groups configured for this cloud provider.
func (azure *AzureCloudProvider) NodeGroups() []cloudprovider.NodeGroup {
	result := make([]cloudprovider.NodeGroup, 0, len(azure.scaleSets))
	for _, scaleSet := range azure.scaleSets {
		result = append(result, scaleSet)
	}
	return result
}

// Cleanup cleans up all resources before the cloud provider is removed
func (azure *AzureCloudProvider) Cleanup() error {}

// NodeGroupForNode returns the node group for the given node.
func (azure *AzureCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	fmt.Printf("Searching for node group for the node: %s, %s\n", node.Spec.ExternalID, node.Spec.ProviderID)
	ref := &AzureRef{
		Name: node.Spec.ProviderID,
	}

	scaleSet, err := azure.azureManager.GetScaleSetForInstance(ref)

	return scaleSet, err
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (azure *AzureCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// AzureRef contains a reference to some entity in Azure world.
type AzureRef struct {
	Name string
}

// GetKey returns key of the given azure reference.
func (m *AzureRef) GetKey() string {
	return m.Name
}

// AzureRefFromProviderId creates InstanceConfig object from provider id which
// must be in format: azure:///resourceGroupName/name
func AzureRefFromProviderId(id string) (*AzureRef, error) {
	splitted := strings.Split(id[9:], "/")
	if len(splitted) != 2 {
		return nil, fmt.Errorf("Wrong id: expected format azure:////<unique-id>, got %v", id)
	}
	return &AzureRef{
		Name: splitted[len(splitted)-1],
	}, nil
}

// ScaleSet implements NodeGroup interface.
type ScaleSet struct {
	AzureRef

	azureManager *AzureManager
	minSize      int
	maxSize      int
}

// MinSize returns minimum size of the node group.
func (scaleSet *ScaleSet) MinSize() int {
	return scaleSet.minSize
}

// MaxSize returns maximum size of the node group.
func (scaleSet *ScaleSet) MaxSize() int {
	return scaleSet.maxSize
}

// TargetSize returns the current TARGET size of the node group. It is possible that the
// number is different from the number of nodes registered in Kuberentes.
func (scaleSet *ScaleSet) TargetSize() (int, error) {
	size, err := scaleSet.azureManager.GetScaleSetSize(scaleSet)
	return int(size), err
}

// IncreaseSize increases Scale Set size
func (scaleSet *ScaleSet) IncreaseSize(delta int) error {
	if delta <= 0 {
		return fmt.Errorf("size increase must be positive")
	}
	size, err := scaleSet.azureManager.GetScaleSetSize(scaleSet)
	if err != nil {
		return err
	}
	if int(size)+delta > scaleSet.MaxSize() {
		return fmt.Errorf("size increase too large - desired:%d max:%d", int(size)+delta, scaleSet.MaxSize())
	}
	return scaleSet.azureManager.SetScaleSetSize(scaleSet, size+int64(delta))
}

// DecreaseTargetSize decreases the target size of the node group. This function
// doesn't permit to delete any existing node and can be used only to reduce the
// request for new nodes that have not been yet fulfilled. Delta should be negative.
// It is assumed that cloud provider will not delete the existing nodes if the size
// when there is an option to just decrease the target.
func (scaleSet *ScaleSet) DecreaseTargetSize(delta int) error {
	if delta >= 0 {
		return fmt.Errorf("size decrease size must be negative")
	}
	size, err := scaleSet.azureManager.GetScaleSetSize(scaleSet)
	if err != nil {
		return err
	}
	nodes, err := scaleSet.azureManager.GetScaleSetVms(scaleSet)
	if err != nil {
		return err
	}
	if int(size)+delta < len(nodes) {
		return fmt.Errorf("attempt to delete existing nodes targetSize:%d delta:%d existingNodes: %d",
			size, delta, len(nodes))
	}
	return scaleSet.azureManager.SetScaleSetSize(scaleSet, size+int64(delta))
}

// Belongs returns true if the given node belongs to the NodeGroup.
func (scaleSet *ScaleSet) Belongs(node *apiv1.Node) (bool, error) {
	fmt.Printf("Check if node belongs to this scale set: scaleset:%v, node:%v\n", scaleSet, node)

	ref := &AzureRef{
		Name: node.Spec.ProviderID,
	}

	targetAsg, err := scaleSet.azureManager.GetScaleSetForInstance(ref)
	if err != nil {
		return false, err
	}
	if targetAsg == nil {
		return false, fmt.Errorf("%s doesn't belong to a known scale set", node.Name)
	}
	if targetAsg.Id() != scaleSet.Id() {
		return false, nil
	}
	return true, nil
}

// DeleteNodes deletes the nodes from the group.
func (scaleSet *ScaleSet) DeleteNodes(nodes []*apiv1.Node) error {
	fmt.Printf("Delete nodes requested: %v\n", nodes)
	size, err := scaleSet.azureManager.GetScaleSetSize(scaleSet)
	if err != nil {
		return err
	}
	if int(size) <= scaleSet.MinSize() {
		return fmt.Errorf("min size reached, nodes will not be deleted")
	}
	refs := make([]*AzureRef, 0, len(nodes))
	for _, node := range nodes {
		belongs, err := scaleSet.Belongs(node)
		if err != nil {
			return err
		}
		if belongs != true {
			return fmt.Errorf("%s belongs to a different asg than %s", node.Name, scaleSet.Id())
		}
		azureRef := &AzureRef{
			Name: node.Spec.ProviderID,
		}
		refs = append(refs, azureRef)
	}
	return scaleSet.azureManager.DeleteInstances(refs)
}

// Id returns ScaleSet id.
func (scaleSet *ScaleSet) Id() string {
	return scaleSet.Name
}

// Debug returns a debug string for the Scale Set.
func (scaleSet *ScaleSet) Debug() string {
	return fmt.Sprintf("%s (%d:%d)", scaleSet.Id(), scaleSet.MinSize(), scaleSet.MaxSize())
}

// TemplateNodeInfo returns a node template for this scale set.
func (scaleSet *ScaleSet) TemplateNodeInfo() (*schedulercache.NodeInfo, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// Create ScaleSet from provided spec.
// spec is in the following format: min-size:max-size:scale-set-name.
func buildScaleSet(spec string, azureManager *AzureManager) (*ScaleSet, error) {
	tokens := strings.SplitN(spec, ":", 3)
	if len(tokens) != 3 {
		return nil, fmt.Errorf("wrong nodes configuration: %s", spec)
	}

	scaleSet := ScaleSet{
		azureManager: azureManager,
	}
	if size, err := strconv.Atoi(tokens[0]); err == nil {
		if size <= 0 {
			return nil, fmt.Errorf("min size must be >= 1, got: %d", size)
		}
		scaleSet.minSize = size
	} else {
		return nil, fmt.Errorf("failed to set min size: %s, expected integer", tokens[0])
	}

	if size, err := strconv.Atoi(tokens[1]); err == nil {
		if size < scaleSet.minSize {
			return nil, fmt.Errorf("max size must be greater or equal to min size")
		}
		scaleSet.maxSize = size
	} else {
		return nil, fmt.Errorf("failed to set max size: %s, expected integer", tokens[1])
	}

	if tokens[2] == "" {
		return nil, fmt.Errorf("scale set name must not be blank, got spec: %s", spec)
	}

	scaleSet.Name = tokens[2]
	return &scaleSet, nil
}

// Nodes returns a list of all nodes that belong to this node group.
func (scaleSet *ScaleSet) Nodes() ([]string, error) {
	return scaleSet.azureManager.GetScaleSetVms(scaleSet)
}
