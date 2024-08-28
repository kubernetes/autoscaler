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

package accountclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

func (client *Client) Create(ctx context.Context, resourceGroupName string, resourceName string, resource *armstorage.AccountCreateParameters) (*armstorage.Account, error) {
	if resource == nil {
		resource = &armstorage.AccountCreateParameters{}
	}
	resp, err := utils.NewPollerWrapper(client.AccountsClient.BeginCreate(ctx, resourceGroupName, resourceName, *resource, nil)).WaitforPollerResp(ctx)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		return &resp.Account, nil
	}
	return nil, nil
}

func (client *Client) GetProperties(ctx context.Context, resourceGroupName string, accountName string, options *armstorage.AccountsClientGetPropertiesOptions) (*armstorage.Account, error) {
	resp, err := client.AccountsClient.GetProperties(ctx, resourceGroupName, accountName, options)
	if err != nil {
		return nil, err
	}
	//handle statuscode
	return &resp.Account, nil
}

// Delete deletes a Interface by name.
func (client *Client) Delete(ctx context.Context, resourceGroupName string, resourceName string) error {
	_, err := client.AccountsClient.Delete(ctx, resourceGroupName, resourceName, nil)
	return err
}

func (client *Client) ListKeys(ctx context.Context, resourceGroupName string, accountName string) ([]*armstorage.AccountKey, error) {
	resp, err := client.AccountsClient.ListKeys(ctx, resourceGroupName, accountName, nil)
	if err != nil {
		return nil, err
	}
	return resp.Keys, nil
}
