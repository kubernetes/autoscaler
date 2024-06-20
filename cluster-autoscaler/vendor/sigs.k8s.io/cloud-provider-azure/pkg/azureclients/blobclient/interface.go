/*
Copyright 2022 The Kubernetes Authors.

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

package blobclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2021-09-01/storage"

	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

const (
	// APIVersion is the API version for storage.
	APIVersion = "2021-09-01"
	// AzureStackCloudAPIVersion is the API version for Azure Stack
	AzureStackCloudAPIVersion = "2019-06-01"
	// AzureStackCloudName is the cloud name of Azure Stack
	AzureStackCloudName = "AZURESTACKCLOUD"
)

// Interface is the client interface for creating file shares, interface for test injection.
// Don't forget to run "hack/update-mock-clients.sh" command to generate the mock client.
type Interface interface {
	CreateContainer(ctx context.Context, subsID, resourceGroupName, accountName, containerName string, parameters storage.BlobContainer) *retry.Error
	DeleteContainer(ctx context.Context, subsID, resourceGroupName, accountName, containerName string) *retry.Error
	GetContainer(ctx context.Context, subsID, resourceGroupName, accountName, containerName string) (storage.BlobContainer, *retry.Error)
	GetServiceProperties(ctx context.Context, subsID, resourceGroupName, accountName string) (storage.BlobServiceProperties, error)
	SetServiceProperties(ctx context.Context, subsID, resourceGroupName, accountName string, parameters storage.BlobServiceProperties) (storage.BlobServiceProperties, error)
}
