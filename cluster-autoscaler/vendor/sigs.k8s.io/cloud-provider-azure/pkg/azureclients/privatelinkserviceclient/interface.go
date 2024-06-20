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

package privatelinkserviceclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"

	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

const (
	// APIVersion is the API version for network.
	APIVersion = "2022-07-01"
	// AzureStackCloudAPIVersion is the API version for Azure Stack
	AzureStackCloudAPIVersion = "2019-03-01"
	// AzureStackCloudName is the cloud name of Azure Stack
	AzureStackCloudName = "AZURESTACKCLOUD"
)

// Interface is the client interface for Private Link Services.
// Don't forget to run "hack/update-mock-clients.sh" command to generate the mock client.
type Interface interface {

	// Get gets the private link service
	Get(ctx context.Context, resourceGroupName string, privateLinkServiceName string, expand string) (result network.PrivateLinkService, rerr *retry.Error)

	// List gets a list of PrivateLinkService in the resource group.
	List(ctx context.Context, resourceGroupName string) (result []network.PrivateLinkService, rerr *retry.Error)

	// CreateOrUpdate creates or updates a private link service.
	CreateOrUpdate(ctx context.Context, resourceGroupName string, privateLinkServiceName string, privateLinkService network.PrivateLinkService, etag string) *retry.Error

	// Delete deletes a private link service by name.
	Delete(ctx context.Context, resourceGroupName string, privateLinkServiceName string) *retry.Error

	// Delete deletes a private endpoint connection to the private link service by name
	DeletePEConnection(ctx context.Context, resourceGroupName string, privateLinkServiceName string, privateEndpointConnectionName string) *retry.Error
}
