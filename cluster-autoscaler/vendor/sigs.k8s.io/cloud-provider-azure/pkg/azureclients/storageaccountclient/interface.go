/*
Copyright 2020 The Kubernetes Authors.

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

package storageaccountclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2021-09-01/storage"

	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

const (
	// APIVersion is the API version for network.
	APIVersion = "2021-02-01"
	// AzureStackCloudAPIVersion is the API version for Azure Stack
	AzureStackCloudAPIVersion = "2018-02-01"
	// AzureStackCloudName is the cloud name of Azure Stack
	AzureStackCloudName = "AZURESTACKCLOUD"
)

// Interface is the client interface for StorageAccounts.
// Don't forget to run "hack/update-mock-clients.sh" command to generate the mock client.
type Interface interface {
	// Create creates a StorageAccount.
	Create(ctx context.Context, subsID, resourceGroupName, accountName string, parameters storage.AccountCreateParameters) *retry.Error

	// Update updates a StorageAccount.
	Update(ctx context.Context, subsID, resourceGroupName, accountName string, parameters storage.AccountUpdateParameters) *retry.Error

	// Delete deletes a StorageAccount by name.
	Delete(ctx context.Context, subsID, resourceGroupName, accountName string) *retry.Error

	// ListKeys get a list of storage account keys.
	ListKeys(ctx context.Context, subsID, resourceGroupName, accountName string) (storage.AccountListKeysResult, *retry.Error)

	// ListByResourceGroup get a list storage accounts by resourceGroup.
	ListByResourceGroup(ctx context.Context, subsID, resourceGroupName string) ([]storage.Account, *retry.Error)

	// GetProperties gets properties of the StorageAccount.
	GetProperties(ctx context.Context, subsID, resourceGroupName, accountName string) (result storage.Account, rerr *retry.Error)
}
