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

package routetable

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v9"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/routetableclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/cache"
)

// Generate mocks for the repository interface
//go:generate mockgen -destination=./mock_repo.go -package=routetable -copyright_file ../../../hack/boilerplate/boilerplate.generatego.txt -source=repo.go Repository

var (
	ErrMissingRouteTableName = fmt.Errorf("missing RouteTable name")
)

type Repository interface {
	Get(ctx context.Context, routeTableName string, crt cache.AzureCacheReadType) (*armnetwork.RouteTable, error)
	CreateOrUpdate(ctx context.Context, routeTable armnetwork.RouteTable) (*armnetwork.RouteTable, error)
}

type repo struct {
	resourceGroup string
	client        routetableclient.Interface
	cache         cache.Resource
}

func NewRepo(
	client routetableclient.Interface,
	resourceGroup string,
	cacheTTL time.Duration,
	disableAPICallCache bool,
) (Repository, error) {
	c, err := NewCache(client, resourceGroup, cacheTTL, disableAPICallCache)
	if err != nil {
		return nil, fmt.Errorf("new RouteTable cache: %w", err)
	}

	return &repo{
		resourceGroup: resourceGroup,
		client:        client,
		cache:         c,
	}, nil
}

func (r *repo) Get(ctx context.Context, routeTableName string, crt cache.AzureCacheReadType) (*armnetwork.RouteTable, error) {
	rt, err := r.cache.GetWithDeepCopy(ctx, routeTableName, crt)
	if err != nil {
		return nil, fmt.Errorf("get RouteTable: %w", err)
	}
	if rt == nil {
		return nil, nil
	}
	routeTable, ok := rt.(*armnetwork.RouteTable)
	if !ok {
		return nil, fmt.Errorf("unexpected type for RouteTable: got %T, want *armnetwork.RouteTable", rt)
	}
	return routeTable, nil
}

func (r *repo) CreateOrUpdate(ctx context.Context, routeTable armnetwork.RouteTable) (*armnetwork.RouteTable, error) {
	if routeTable.Name == nil {
		return nil, ErrMissingRouteTableName
	}

	rv, err := r.client.CreateOrUpdate(ctx, r.resourceGroup, *routeTable.Name, routeTable)
	if err != nil {
		return nil, fmt.Errorf("create or update RouteTable: %w", err)
	}
	_ = r.cache.Delete(*routeTable.Name)

	return rv, nil
}
