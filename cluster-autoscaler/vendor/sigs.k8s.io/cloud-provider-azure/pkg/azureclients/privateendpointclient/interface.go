/*
Copyright 2021 The Kubernetes Authors.

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

package privateendpointclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"

	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

const (
	// APIVersion is the API version for network.
	APIVersion = "2022-07-01"
	// AzureStackCloudName is the cloud name of Azure Stack
	AzureStackCloudName = "AZURESTACKCLOUD"
)

// Interface is the client interface for Private Endpoints.
// Don't forget to run "hack/update-mock-clients.sh" command to generate the mock client.
type Interface interface {

	// Get gets the private endpoint
	Get(ctx context.Context, resourceGroupName string, privateEndpointName string, expand string) (result network.PrivateEndpoint, rerr *retry.Error)

	// CreateOrUpdate creates or updates a private endpoint.
	CreateOrUpdate(ctx context.Context, resourceGroupName string, endpointName string, privateEndpoint network.PrivateEndpoint, etag string, waitForCompletion bool) *retry.Error
}
