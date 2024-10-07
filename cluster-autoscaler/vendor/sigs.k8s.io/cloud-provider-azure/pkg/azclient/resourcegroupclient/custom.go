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

package resourcegroupclient

import (
	"context"

	resources "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

func (client *Client) Get(ctx context.Context, resourceGroupName string) (result *resources.ResourceGroup, rerr error) {
	resp, err := client.ResourceGroupsClient.Get(ctx, resourceGroupName, nil)
	if err != nil {
		return nil, err
	}
	//handle statuscode
	return &resp.ResourceGroup, nil
}
func (client *Client) CreateOrUpdate(ctx context.Context, resourceGroupName string, resourceParam resources.ResourceGroup) (*resources.ResourceGroup, error) {
	resp, err := client.ResourceGroupsClient.CreateOrUpdate(ctx, resourceGroupName, resourceParam, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ResourceGroup, nil
}

func (client *Client) Delete(ctx context.Context, resourceGroupName string) error {
	_, err := utils.NewPollerWrapper(client.BeginDelete(ctx, resourceGroupName, nil)).WaitforPollerResp(ctx)
	return err
}

func (client *Client) List(ctx context.Context) (result []*resources.ResourceGroup, rerr error) {
	pager := client.ResourceGroupsClient.NewListPager(nil)
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, nextResult.Value...)
	}
	return result, nil
}
