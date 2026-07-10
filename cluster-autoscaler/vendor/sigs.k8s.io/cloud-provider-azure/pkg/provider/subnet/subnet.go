/*
Copyright 2024 The Kubernetes Authors.

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

package subnet

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v9"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/subnetclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/log"
)

type Repository interface {
	CreateOrUpdate(ctx context.Context, rg string, vnetName string, subnetName string, subnet armnetwork.Subnet) error
	Get(ctx context.Context, rg string, vnetName string, subnetName string) (*armnetwork.Subnet, error)
}

type repo struct {
	SubnetsClient subnetclient.Interface
}

func NewRepo(subnetsClient subnetclient.Interface) (Repository, error) {
	return &repo{
		SubnetsClient: subnetsClient,
	}, nil
}

// CreateOrUpdateSubnet invokes az.SubnetClient.CreateOrUpdate with exponential backoff retry
func (az *repo) CreateOrUpdate(ctx context.Context, rg string, vnetName string, subnetName string, subnet armnetwork.Subnet) error {
	logger := log.FromContextOrBackground(ctx).WithName("SubnetsClient.CreateOrUpdate")
	_, rerr := az.SubnetsClient.CreateOrUpdate(ctx, rg, vnetName, subnetName, subnet)
	logger.V(10).Info("end", "subnetName", subnetName)
	if rerr != nil {
		logger.Error(rerr, "SubnetClient.CreateOrUpdate failed", "subnetName", subnetName)
		return rerr
	}

	return nil
}

func (az *repo) Get(ctx context.Context, rg string, vnetName string, subnetName string) (*armnetwork.Subnet, error) {
	logger := log.FromContextOrBackground(ctx).WithName("SubnetsClient.Get")
	subnet, err := az.SubnetsClient.Get(ctx, rg, vnetName, subnetName, nil)
	if err != nil {
		logger.Error(err, "SubnetClient.Get failed", "subnetName", subnetName)
		return nil, err
	}
	return subnet, nil
}
