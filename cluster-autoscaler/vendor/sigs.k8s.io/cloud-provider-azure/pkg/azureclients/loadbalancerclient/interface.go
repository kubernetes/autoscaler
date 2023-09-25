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

package loadbalancerclient

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"

	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

const (
	// APIVersion is the API version for network.
	APIVersion = "2022-07-01"
	// AzureStackCloudAPIVersion is the API version for Azure Stack
	AzureStackCloudAPIVersion = "2018-11-01"
	// AzureStackCloudName is the cloud name of Azure Stack
	AzureStackCloudName = "AZURESTACKCLOUD"
)

// Interface is the client interface for LoadBalancer.
// Don't forget to run "hack/update-mock-clients.sh" command to generate the mock client.
type Interface interface {
	// Get gets a LoadBalancer.
	Get(ctx context.Context, resourceGroupName string, loadBalancerName string, expand string) (result network.LoadBalancer, rerr *retry.Error)

	// List gets a list of LoadBalancer in the resource group.
	List(ctx context.Context, resourceGroupName string) (result []network.LoadBalancer, rerr *retry.Error)

	// CreateOrUpdate creates or updates a LoadBalancer.
	CreateOrUpdate(ctx context.Context, resourceGroupName string, loadBalancerName string, parameters network.LoadBalancer, etag string) *retry.Error

	// CreateOrUpdateBackendPools creates or updates loadbalancer's backend address pool.
	CreateOrUpdateBackendPools(ctx context.Context, resourceGroupName string, loadBalancerName string, backendPoolName string, parameters network.BackendAddressPool, etag string) *retry.Error

	// Delete deletes a LoadBalancer by name.
	Delete(ctx context.Context, resourceGroupName string, loadBalancerName string) *retry.Error

	// GetLBBackendPool gets a LoadBalancer backend pool.
	GetLBBackendPool(ctx context.Context, resourceGroupName string, loadBalancerName string, backendPoolName string, expand string) (network.BackendAddressPool, *retry.Error)

	// DeleteLBBackendPool deletes a LoadBalancer backend pool by name.
	DeleteLBBackendPool(ctx context.Context, resourceGroupName, loadBalancerName, backendPoolName string) *retry.Error

	// MigrateToIPBasedBackendPool migrates a NIC-based backend pool to IP-based.
	MigrateToIPBasedBackendPool(ctx context.Context, resourceGroupName string, loadBalancerName string, backendPoolNames []string) *retry.Error
}
