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
	"context"
	"fmt"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/stretchr/testify/mock"
)

const (
	fakeVirtualMachineScaleSetVMID = "/subscriptions/test-subscription-id/resourceGroups/test-asg/providers/Microsoft.Compute/virtualMachineScaleSets/agents/virtualMachines/%d"
	fakeVirtualMachineVMID         = "/subscriptions/test-subscription-id/resourceGroups/test-asg/providers/Microsoft.Compute/virtualMachines/%d"
)

// DeploymentClientMock mocks for DeploymentsClient.
type DeploymentClientMock struct {
	mock.Mock

	mutex         sync.Mutex
	FakeStore     map[string]armresources.DeploymentExtended
	TemplateStore map[string]map[string]interface{}
}

// Get gets the DeploymentExtended by deploymentName.
func (m *DeploymentClientMock) Get(ctx context.Context, resourceGroupName string, deploymentName string) (*armresources.DeploymentExtended, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	deploy, ok := m.FakeStore[deploymentName]
	if !ok {
		return nil, fmt.Errorf("deployment not found")
	}

	return &deploy, nil
}

// ExportTemplate exports the deployment's template.
func (m *DeploymentClientMock) ExportTemplate(ctx context.Context, resourceGroupName string, deploymentName string) (*armresources.DeploymentExportResult, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	_, ok := m.FakeStore[deploymentName]
	if !ok {
		return nil, fmt.Errorf("deployment not found")
	}

	template := m.TemplateStore[deploymentName]
	if template == nil {
		template = map[string]interface{}{"resources": []interface{}{}}
	}

	return &armresources.DeploymentExportResult{
		Template: template,
	}, nil
}

// CreateOrUpdate creates or updates the Deployment.
func (m *DeploymentClientMock) CreateOrUpdate(ctx context.Context, resourceGroupName string, deploymentName string, parameters armresources.Deployment) (*armresources.DeploymentExtended, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	deploy, ok := m.FakeStore[deploymentName]
	if !ok {
		deploy = armresources.DeploymentExtended{
			Properties: &armresources.DeploymentPropertiesExtended{},
		}
	}

	if parameters.Properties != nil {
		deploy.Properties.Parameters = parameters.Properties.Parameters
		deploy.Properties.TemplateLink = parameters.Properties.TemplateLink
	}
	deploy.Name = &deploymentName
	m.FakeStore[deploymentName] = deploy
	return &deploy, nil
}

// List gets all the deployments for a resource group.
func (m *DeploymentClientMock) List(ctx context.Context, resourceGroupName string) ([]*armresources.DeploymentExtended, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	result := make([]*armresources.DeploymentExtended, 0)
	for i := range m.FakeStore {
		deployment := m.FakeStore[i]
		result = append(result, &deployment)
	}

	return result, nil
}

// Delete deletes the given deployment
func (m *DeploymentClientMock) Delete(ctx context.Context, resourceGroupName, deploymentName string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.FakeStore[deploymentName]; !ok {
		return fmt.Errorf("there is no such a deployment with name %s", deploymentName)
	}

	delete(m.FakeStore, deploymentName)

	return nil
}

func fakeVMSSWithTags(vmssName string, tags map[string]*string) armcompute.VirtualMachineScaleSet {
	skuName := "Standard_D4_v2"
	var vmssCapacity int64 = 3

	return armcompute.VirtualMachineScaleSet{
		Name: &vmssName,
		SKU: &armcompute.SKU{
			Capacity: &vmssCapacity,
			Name:     &skuName,
		},
		Tags: tags,
	}

}
