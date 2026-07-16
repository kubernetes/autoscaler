/*
Copyright 2023 The Kubernetes Authors.

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

package securitygroup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v9"

	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"

	"sigs.k8s.io/cloud-provider-azure/pkg/azclient/securitygroupclient"
	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/log"
	"sigs.k8s.io/cloud-provider-azure/pkg/util/errutils"
)

const (
	nsgCacheTTLDefaultInSeconds = 120
)

type Repository interface {
	GetSecurityGroup(ctx context.Context) (*armnetwork.SecurityGroup, error)
	CreateOrUpdateSecurityGroup(ctx context.Context, sg *armnetwork.SecurityGroup) error
}

type securityGroupRepo struct {
	securityGroupResourceGroup string
	securityGroupName          string
	nsgCacheTTLInSeconds       int
	securigyGroupClient        securitygroupclient.Interface
	nsgCache                   azcache.Resource
}

func NewSecurityGroupRepo(securityGroupResourceGroup string, securityGroupName string, nsgCacheTTLInSeconds int, disableAPICallCache bool, securityGroupClient securitygroupclient.Interface) (Repository, error) {
	getter := func(ctx context.Context, key string) (interface{}, error) {
		logger := log.FromContextOrBackground(ctx).WithName("NewSecurityGroupRepo.getter")
		nsg, err := securityGroupClient.Get(ctx, securityGroupResourceGroup, key)
		exists, rerr := errutils.CheckResourceExistsFromAzcoreError(err)
		if rerr != nil {
			return nil, err
		}

		if !exists {
			logger.V(2).Info("Security group not found", "securityGroup", key)
			return nil, nil
		}

		return nsg, nil
	}

	if nsgCacheTTLInSeconds == 0 {
		nsgCacheTTLInSeconds = nsgCacheTTLDefaultInSeconds
	}
	cache, err := azcache.NewTimedCache(time.Duration(nsgCacheTTLInSeconds)*time.Second, getter, disableAPICallCache)
	if err != nil {
		logger := log.Background().WithName("NewSecurityGroupRepo")
		logger.Error(err, "Failed to create cache for security group", "securityGroupName", securityGroupName)
		return nil, err
	}

	return &securityGroupRepo{
		securityGroupResourceGroup: securityGroupResourceGroup,
		securityGroupName:          securityGroupName,
		nsgCacheTTLInSeconds:       nsgCacheTTLDefaultInSeconds,
		securigyGroupClient:        securityGroupClient,
		nsgCache:                   cache,
	}, nil
}

// CreateOrUpdateSecurityGroup invokes az.SecurityGroupsClient.CreateOrUpdate with exponential backoff retry
func (az *securityGroupRepo) CreateOrUpdateSecurityGroup(ctx context.Context, sg *armnetwork.SecurityGroup) error {
	logger := log.FromContextOrBackground(ctx).WithName("CreateOrUpdateSecurityGroup")
	_, rerr := az.securigyGroupClient.CreateOrUpdate(ctx, az.securityGroupResourceGroup, *sg.Name, *sg)
	logger.V(10).Info("SecurityGroupsClient.CreateOrUpdate: end", "securityGroupName", *sg.Name)
	if rerr == nil {
		// Invalidate the cache right after updating
		_ = az.nsgCache.Delete(*sg.Name)
		return nil
	}
	var respError *azcore.ResponseError
	if errors.As(rerr, &respError) && respError != nil {
		nsgJSON, _ := json.Marshal(sg)
		klog.Warningf("CreateOrUpdateSecurityGroup(%s) failed: %v, NSG request: %s", ptr.Deref(sg.Name, ""), rerr.Error(), string(nsgJSON))

		// Invalidate the cache because ETAG precondition mismatch.
		if respError.StatusCode == http.StatusPreconditionFailed {
			logger.V(3).Info("SecurityGroup cache is cleanup because of http.StatusPreconditionFailed", "securityGroupName", *sg.Name)
			_ = az.nsgCache.Delete(*sg.Name)
		}

		// Invalidate the cache because another new operation has canceled the current request.
		if strings.Contains(strings.ToLower(respError.Error()), consts.OperationCanceledErrorMessage) {
			logger.V(3).Info("SecurityGroup cache is cleanup because CreateOrUpdateSecurityGroup is canceled by another operation", "securityGroupName", *sg.Name)
			_ = az.nsgCache.Delete(*sg.Name)
		}
	}
	return rerr
}

func (az *securityGroupRepo) GetSecurityGroup(ctx context.Context) (*armnetwork.SecurityGroup, error) {
	nsg := &armnetwork.SecurityGroup{}
	if az.securityGroupName == "" {
		return nsg, fmt.Errorf("securityGroupName is not configured")
	}

	securityGroup, err := az.nsgCache.GetWithDeepCopy(ctx, az.securityGroupName, azcache.CacheReadTypeDefault)
	if err != nil {
		return nsg, err
	}

	if securityGroup == nil {
		return nsg, fmt.Errorf("nsg %q not found", az.securityGroupName)
	}

	return securityGroup.(*armnetwork.SecurityGroup), nil
}
