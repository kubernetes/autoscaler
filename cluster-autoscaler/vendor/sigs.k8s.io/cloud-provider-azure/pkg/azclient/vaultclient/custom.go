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

package vaultclient

import (
	"context"

	armkeyvault "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/utils"
)

// Delete deletes a Vault by name.
func (client *Client) Delete(ctx context.Context, resourceGroupName string, resourceName string) error {
	_, err := client.VaultsClient.Delete(ctx, resourceGroupName, resourceName, nil)
	return err
}

// CreateOrUpdate creates or updates a Vault.
func (client *Client) CreateOrUpdate(ctx context.Context, resourceGroupName string, resourceName string, resource armkeyvault.VaultCreateOrUpdateParameters) (*armkeyvault.Vault, error) {
	resp, err := utils.NewPollerWrapper(client.VaultsClient.BeginCreateOrUpdate(ctx, resourceGroupName, resourceName, resource, nil)).WaitforPollerResp(ctx)
	if err != nil {
		return nil, err
	}
	if resp != nil {
		return &resp.Vault, nil
	}
	return nil, nil
}

func (client *Client) PurgeDeleted(ctx context.Context, vaultName string, location string) error {
	_, err := utils.NewPollerWrapper(client.BeginPurgeDeleted(ctx, vaultName, location, nil)).WaitforPollerResp(ctx)
	return err
}
