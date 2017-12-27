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

package azure

import (
	"fmt"
	"strings"

	"github.com/golang/glog"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

const (
	// ProviderName is the cloud provider name for Azure
	ProviderName = "azure"
)

// AzureCloudProvider provides implementation of CloudProvider interface for Azure.
type AzureCloudProvider struct {
	azureManager    *AzureManager
	scaleSets       []*ScaleSet
	resourceLimiter *cloudprovider.ResourceLimiter
}

// BuildAzureCloudProvider creates new AzureCloudProvider
func BuildAzureCloudProvider(azureManager *AzureManager, specs []string, resourceLimiter *cloudprovider.ResourceLimiter) (*AzureCloudProvider, error) {
	azure := &AzureCloudProvider{
		azureManager:    azureManager,
		resourceLimiter: resourceLimiter,
	}
	for _, spec := range specs {
		if err := azure.addNodeGroup(spec); err != nil {
			return nil, err
		}
	}

	return azure, nil
}

// Cleanup stops the go routine that is handling the current view of the ASGs in the form of a cache
func (azure *AzureCloudProvider) Cleanup() error {
	azure.azureManager.Cleanup()
	return nil
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

// NodeGroupForNode returns the node group for the given node.
func (azure *AzureCloudProvider) NodeGroupForNode(node *apiv1.Node) (cloudprovider.NodeGroup, error) {
	glog.V(6).Infof("Searching for node group for the node: %s, %s\n", node.Spec.ExternalID, node.Spec.ProviderID)
	ref := &AzureRef{
		Name: strings.ToLower(node.Spec.ProviderID),
	}

	scaleSet, err := azure.azureManager.GetScaleSetForInstance(ref)

	return scaleSet, err
}

// Pricing returns pricing model for this cloud provider or error if not available.
func (azure *AzureCloudProvider) Pricing() (cloudprovider.PricingModel, errors.AutoscalerError) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetAvailableMachineTypes get all machine types that can be requested from the cloud provider.
func (azure *AzureCloudProvider) GetAvailableMachineTypes() ([]string, error) {
	return []string{}, nil
}

// NewNodeGroup builds a theoretical node group based on the node definition provided. The node group is not automatically
// created on the cloud provider side. The node group is not returned by NodeGroups() until it is created.
func (azure *AzureCloudProvider) NewNodeGroup(machineType string, labels map[string]string, systemLabels map[string]string, extraResources map[string]resource.Quantity) (cloudprovider.NodeGroup, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// GetResourceLimiter returns struct containing limits (max, min) for resources (cores, memory etc.).
func (azure *AzureCloudProvider) GetResourceLimiter() (*cloudprovider.ResourceLimiter, error) {
	return azure.resourceLimiter, nil
}

// Refresh is called before every main loop and can be used to dynamically update cloud provider state.
// In particular the list of node groups returned by NodeGroups can change as a result of CloudProvider.Refresh().
func (azure *AzureCloudProvider) Refresh() error {
	return nil
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
		return nil, fmt.Errorf("Wrong id: expected format azure:///<unique-id>, got %v", id)
	}
	return &AzureRef{
		Name: splitted[len(splitted)-1],
	}, nil
}
