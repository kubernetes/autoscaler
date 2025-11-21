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
	"net/http"
	"sync"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources/v2"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

const (
	fakeVirtualMachineScaleSetVMID = "/subscriptions/test-subscription-id/resourceGroups/test-asg/providers/Microsoft.Compute/virtualMachineScaleSets/agents/virtualMachines/%d"
	fakeVirtualMachineVMID         = "/subscriptions/test-subscription-id/resourceGroups/test-asg/providers/Microsoft.Compute/virtualMachines/%d"
)

// DeploymentClientMock mocks for DeploymentsClient.
type DeploymentClientMock struct {
	mock.Mock

	mutex     sync.Mutex
	FakeStore map[string]armresources.DeploymentExtended
	// Store templates separately since DeploymentPropertiesExtended doesn't have Template field in SDK v2
	Templates map[string]interface{}
}

// Get gets the DeploymentExtended by deploymentName.
func (m *DeploymentClientMock) Get(ctx context.Context, resourceGroupName string, deploymentName string, options *armresources.DeploymentsClientGetOptions) (armresources.DeploymentsClientGetResponse, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	deploy, ok := m.FakeStore[deploymentName]
	if !ok {
		return armresources.DeploymentsClientGetResponse{}, fmt.Errorf("deployment not found")
	}

	return armresources.DeploymentsClientGetResponse{
		DeploymentExtended: deploy,
	}, nil
}

// ExportTemplate exports the deployment's template.
func (m *DeploymentClientMock) ExportTemplate(ctx context.Context, resourceGroupName string, deploymentName string, options *armresources.DeploymentsClientExportTemplateOptions) (armresources.DeploymentsClientExportTemplateResponse, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	_, ok := m.FakeStore[deploymentName]
	if !ok {
		return armresources.DeploymentsClientExportTemplateResponse{}, fmt.Errorf("deployment not found")
	}

	template, templateOk := m.Templates[deploymentName]
	if !templateOk {
		template = make(map[string]interface{})
	}

	return armresources.DeploymentsClientExportTemplateResponse{
		DeploymentExportResult: armresources.DeploymentExportResult{
			Template: template,
		},
	}, nil
}

// BeginCreateOrUpdate creates or updates the Deployment.
func (m *DeploymentClientMock) BeginCreateOrUpdate(ctx context.Context, resourceGroupName string, deploymentName string, parameters armresources.Deployment, options *armresources.DeploymentsClientBeginCreateOrUpdateOptions) (*runtime.Poller[armresources.DeploymentsClientCreateOrUpdateResponse], error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.Templates == nil {
		m.Templates = make(map[string]interface{})
	}

	deploy, ok := m.FakeStore[deploymentName]
	if !ok {
		deploy = armresources.DeploymentExtended{
			Name:       ptr.To(deploymentName),
			Properties: &armresources.DeploymentPropertiesExtended{},
		}
	}

	deploy.Properties.Parameters = parameters.Properties.Parameters
	deploy.Properties.TemplateLink = parameters.Properties.TemplateLink

	// Store the template separately if provided
	if parameters.Properties.Template != nil {
		m.Templates[deploymentName] = parameters.Properties.Template
	}

	m.FakeStore[deploymentName] = deploy

	// Return a fake poller for the create/update operation
	result := armresources.DeploymentsClientCreateOrUpdateResponse{
		DeploymentExtended: deploy,
	}
	handler := &fakePollerHandler[armresources.DeploymentsClientCreateOrUpdateResponse]{
		done:   true,
		result: result,
	}

	return runtime.NewPoller(
		&http.Response{StatusCode: http.StatusAccepted},
		runtime.Pipeline{},
		&runtime.NewPollerOptions[armresources.DeploymentsClientCreateOrUpdateResponse]{
			Handler: handler,
		},
	)
}

// NewListByResourceGroupPager gets all the deployments for a resource group.
func (m *DeploymentClientMock) NewListByResourceGroupPager(resourceGroupName string, options *armresources.DeploymentsClientListByResourceGroupOptions) *runtime.Pager[armresources.DeploymentsClientListByResourceGroupResponse] {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	result := make([]*armresources.DeploymentExtended, 0)
	for i := range m.FakeStore {
		deploy := m.FakeStore[i]
		result = append(result, &deploy)
	}

	// Create a fake pager that returns all deployments
	return runtime.NewPager(runtime.PagingHandler[armresources.DeploymentsClientListByResourceGroupResponse]{
		More: func(page armresources.DeploymentsClientListByResourceGroupResponse) bool {
			return false
		},
		Fetcher: func(ctx context.Context, page *armresources.DeploymentsClientListByResourceGroupResponse) (armresources.DeploymentsClientListByResourceGroupResponse, error) {
			return armresources.DeploymentsClientListByResourceGroupResponse{
				DeploymentListResult: armresources.DeploymentListResult{
					Value: result,
				},
			}, nil
		},
	})
}

// BeginDelete deletes the given deployment
func (m *DeploymentClientMock) BeginDelete(ctx context.Context, resourceGroupName, deploymentName string, options *armresources.DeploymentsClientBeginDeleteOptions) (*runtime.Poller[armresources.DeploymentsClientDeleteResponse], error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.FakeStore[deploymentName]; !ok {
		return nil, fmt.Errorf("there is no such a deployment with name %s", deploymentName)
	}

	delete(m.FakeStore, deploymentName)
	delete(m.Templates, deploymentName)

	// Create a fake poller using NewPoller with a proper handler
	handler := &fakePollerHandler[armresources.DeploymentsClientDeleteResponse]{
		done:   true,
		result: armresources.DeploymentsClientDeleteResponse{},
	}

	return runtime.NewPoller(
		&http.Response{StatusCode: http.StatusAccepted},
		runtime.Pipeline{},
		&runtime.NewPollerOptions[armresources.DeploymentsClientDeleteResponse]{
			Handler: handler,
		},
	)
}

// fakePollerHandler is a fake poller handler for testing
type fakePollerHandler[T any] struct {
	mu     sync.Mutex
	done   bool
	result T
}

func (f *fakePollerHandler[T]) Done() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.done
}

func (f *fakePollerHandler[T]) Poll(ctx context.Context) (*http.Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.done = true
	return &http.Response{StatusCode: http.StatusOK}, nil
}

func (f *fakePollerHandler[T]) Result(ctx context.Context, out *T) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	*out = f.result
	return nil
}

// List is a helper method for tests that returns deployments as a slice.
func (m *DeploymentClientMock) List(ctx context.Context, resourceGroupName string) ([]armresources.DeploymentExtended, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	result := make([]armresources.DeploymentExtended, 0)
	for i := range m.FakeStore {
		result = append(result, m.FakeStore[i])
	}

	return result, nil
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
