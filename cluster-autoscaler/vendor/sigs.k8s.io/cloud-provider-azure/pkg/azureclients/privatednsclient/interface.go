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

package privatednsclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/privatedns/mgmt/2018-09-01/privatedns"

	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

const (
	// APIVersion is the API version.
	APIVersion = "2018-09-01"
	// AzureStackCloudName is the cloud name of Azure Stack
	AzureStackCloudName = "AZURESTACKCLOUD"
)

// Interface is the client interface for Private DNS Zones
// Don't forget to run "hack/update-mock-clients.sh" command to generate the mock client.
type Interface interface {
	// Get gets a private DNS zone
	Get(ctx context.Context, resourceGroupName, privateZoneName string) (privatedns.PrivateZone, *retry.Error)

	// CreateOrUpdate creates or updates a private DNS zone.
	CreateOrUpdate(ctx context.Context, resourceGroupName, privateZoneName string, parameters privatedns.PrivateZone, etag string, waitForCompletion bool) *retry.Error
}
