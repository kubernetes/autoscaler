/*
Copyright 2023 The Kubernetes Authors.

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

package deploymentclient

import (
	"context"

	resources "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

// List gets a list of Deployment in the resource group.
func (client *Client) List(ctx context.Context, resourceGroupName string) (result []*resources.DeploymentExtended, rerr error) {
	pager := client.DeploymentsClient.NewListByResourceGroupPager(resourceGroupName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, nextResult.Value...)
	}
	return result, nil
}

// Get gets the Deployment
func (client *Client) Get(ctx context.Context, resourceGroupName string, resourceName string) (result *resources.DeploymentExtended, rerr error) {
	var ops *resources.DeploymentsClientGetOptions

	resp, err := client.DeploymentsClient.Get(ctx, resourceGroupName, resourceName, ops)
	if err != nil {
		return nil, err
	}
	//handle statuscode
	return &resp.DeploymentExtended, nil
}

// CreateOrUpdate creates or updates a Deployment.
func (client *Client) CreateOrUpdate(ctx context.Context, resourceGroupName string, resourceName string, resource resources.Deployment) (*resources.DeploymentExtended, error) {
	resp, err := utils.NewPollerWrapper(client.DeploymentsClient.BeginCreateOrUpdate(ctx, resourceGroupName, resourceName, resource, nil)).WaitforPollerResp(ctx)
	if err != nil {
		return nil, err
	}
	return &resp.DeploymentExtended, nil
}
