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
)

// Interface is the client interface for Private DNS Zones
// Don't forget to run "hack/update-mock-clients.sh" command to generate the mock client.
type Interface interface {

	//Get gets the PrivateDNSZone
	Get(ctx context.Context, resourceGroupName string, privateZoneName string) (result privatedns.PrivateZone, err error)

	// CreateOrUpdate creates or updates a private dns zone.
	CreateOrUpdate(ctx context.Context, resourceGroupName string, privateZoneName string, parameters privatedns.PrivateZone, waitForCompletion bool) error
}
