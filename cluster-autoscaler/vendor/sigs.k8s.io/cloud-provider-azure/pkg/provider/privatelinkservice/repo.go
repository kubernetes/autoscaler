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

package privatelinkservice

// Generate mocks for the repository interface
//go:generate mockgen -destination=./mock_repo.go -package=privatelinkservice -copyright_file ../../../hack/boilerplate/boilerplate.generatego.txt -source=repo.go Repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v9"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/privatelinkserviceclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/cache"
)

type Repository interface {
	Get(ctx context.Context, resourceGroup, frontendIPConfigID string, crt cache.AzureCacheReadType) (*armnetwork.PrivateLinkService, error)
	List(ctx context.Context, resourceGroup string) ([]*armnetwork.PrivateLinkService, error)
	CreateOrUpdate(ctx context.Context, resourceGroup string, pls armnetwork.PrivateLinkService) (*armnetwork.PrivateLinkService, error)
	Delete(ctx context.Context, resourceGroup, plsName, lbFrontendID string) error
	DeletePEConnection(ctx context.Context, resourceGroup, plsName, peConnName string) error
}

type repo struct {
	client privatelinkserviceclient.Interface
	cache  cache.Resource
}

func NewRepo(
	client privatelinkserviceclient.Interface,
	cacheTTL time.Duration,
	disableAPICallCache bool,
) (Repository, error) {
	c, err := NewCache(client, cacheTTL, disableAPICallCache)
	if err != nil {
		return nil, fmt.Errorf("new PLS cache: %w", err)
	}

	return &repo{
		client: client,
		cache:  c,
	}, nil
}

var (
	ErrMissingPLSName                               = fmt.Errorf("missing PLS name")
	ErrLoadBalancerFrontendIPConfigurationsNotFound = fmt.Errorf("load balancer frontend IP configurations not found")
)

func (r *repo) Get(ctx context.Context, resourceGroup, frontendIPConfigID string, crt cache.AzureCacheReadType) (*armnetwork.PrivateLinkService, error) {
	cacheKey := getPLSCacheKey(resourceGroup, frontendIPConfigID)

	cachedPLS, err := r.cache.GetWithDeepCopy(ctx, cacheKey, crt)
	if err != nil {
		return nil, fmt.Errorf("get PLS: %w", err)
	}

	return cachedPLS.(*armnetwork.PrivateLinkService), nil
}

func (r *repo) List(ctx context.Context, resourceGroup string) ([]*armnetwork.PrivateLinkService, error) {
	return r.client.List(ctx, resourceGroup)
}

func (r *repo) CreateOrUpdate(ctx context.Context, resourceGroup string, pls armnetwork.PrivateLinkService) (*armnetwork.PrivateLinkService, error) {
	if pls.Name == nil {
		return nil, ErrMissingPLSName
	}
	if pls.Properties == nil ||
		len(pls.Properties.LoadBalancerFrontendIPConfigurations) == 0 ||
		pls.Properties.LoadBalancerFrontendIPConfigurations[0].ID == nil {
		return nil, ErrLoadBalancerFrontendIPConfigurationsNotFound
	}
	cacheKey := getPLSCacheKey(resourceGroup, *pls.Properties.LoadBalancerFrontendIPConfigurations[0].ID)

	resp, err := r.client.CreateOrUpdate(ctx, resourceGroup, *pls.Name, pls)
	if err != nil {
		return nil, fmt.Errorf("create or update PLS: %w", err)
	}
	// clear cache
	_ = r.cache.Delete(cacheKey)

	return resp, nil
}

func (r *repo) Delete(ctx context.Context, resourceGroup, plsName, lbFrontendID string) error {
	if plsName == "" {
		return ErrMissingPLSName
	}
	cacheKey := getPLSCacheKey(resourceGroup, lbFrontendID)

	if err := r.client.Delete(ctx, resourceGroup, plsName); err != nil {
		return fmt.Errorf("delete PLS: %w", err)
	}
	// clear cache
	_ = r.cache.Delete(cacheKey)

	return nil
}

func (r *repo) DeletePEConnection(ctx context.Context, resourceGroup, plsName, peConnName string) error {
	if plsName == "" {
		return ErrMissingPLSName
	}

	if err := r.client.DeletePrivateEndpointConnection(ctx, resourceGroup, plsName, peConnName); err != nil {
		return fmt.Errorf("delete PLS PE connection: %w", err)
	}

	return nil
}
