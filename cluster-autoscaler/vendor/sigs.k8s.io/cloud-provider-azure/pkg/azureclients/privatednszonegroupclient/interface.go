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

package privatednszonegroupclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2021-02-01/network"
)

// Interface is the client interface for Private DNS Zone Group.
// Don't forget to run "hack/update-mock-clients.sh" command to generate the mock client.
type Interface interface {

	// Get gets the private dns zone group
	Get(ctx context.Context, resourceGroupName string, privateEndpointName string, privateDNSZoneGroupName string) (result network.PrivateDNSZoneGroup, err error)

	// CreateOrUpdate creates or updates a private dns zone group endpoint.
	CreateOrUpdate(ctx context.Context, resourceGroupName string, privateEndpointName string, privateDNSZoneGroupName string, parameters network.PrivateDNSZoneGroup, waitForCompletion bool) error
}
