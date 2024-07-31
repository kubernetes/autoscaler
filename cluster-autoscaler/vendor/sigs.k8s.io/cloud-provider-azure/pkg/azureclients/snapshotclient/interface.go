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

package snapshotclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"

	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

const (
	// APIVersion is the API version for compute.
	APIVersion = "2022-03-02"
	// AzureStackCloudAPIVersion is the API version for Azure Stack
	AzureStackCloudAPIVersion = "2019-03-01"
	// AzureStackCloudName is the cloud name of Azure Stack
	AzureStackCloudName = "AZURESTACKCLOUD"
)

// Interface is the client interface for Snapshots.
// Don't forget to run "hack/update-mock-clients.sh" command to generate the mock client.
type Interface interface {
	// Get gets a Snapshot.
	Get(ctx context.Context, subsID, resourceGroupName, snapshotName string) (compute.Snapshot, *retry.Error)

	// Delete deletes a Snapshot by name.
	Delete(ctx context.Context, subsID, resourceGroupName, snapshotName string) *retry.Error

	// ListByResourceGroup get a list snapshots by resourceGroup.
	ListByResourceGroup(ctx context.Context, subsID, resourceGroupName string) ([]compute.Snapshot, *retry.Error)

	// CreateOrUpdate creates or updates a Snapshot.
	CreateOrUpdate(ctx context.Context, subsID, resourceGroupName, snapshotName string, snapshot compute.Snapshot) *retry.Error
}
