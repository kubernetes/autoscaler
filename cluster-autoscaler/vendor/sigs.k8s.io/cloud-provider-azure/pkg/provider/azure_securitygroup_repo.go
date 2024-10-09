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

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

// CreateOrUpdateSecurityGroup invokes az.SecurityGroupsClient.CreateOrUpdate with exponential backoff retry
func (az *Cloud) CreateOrUpdateSecurityGroup(sg network.SecurityGroup) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	rerr := az.SecurityGroupsClient.CreateOrUpdate(ctx, az.SecurityGroupResourceGroup, *sg.Name, sg, pointer.StringDeref(sg.Etag, ""))
	klog.V(10).Infof("SecurityGroupsClient.CreateOrUpdate(%s): end", *sg.Name)
	if rerr == nil {
		// Invalidate the cache right after updating
		_ = az.nsgCache.Delete(*sg.Name)
		return nil
	}

	nsgJSON, _ := json.Marshal(sg)
	klog.Warningf("CreateOrUpdateSecurityGroup(%s) failed: %v, NSG request: %s", pointer.StringDeref(sg.Name, ""), rerr.Error(), string(nsgJSON))

	// Invalidate the cache because ETAG precondition mismatch.
	if rerr.HTTPStatusCode == http.StatusPreconditionFailed {
		klog.V(3).Infof("SecurityGroup cache for %s is cleanup because of http.StatusPreconditionFailed", *sg.Name)
		_ = az.nsgCache.Delete(*sg.Name)
	}

	// Invalidate the cache because another new operation has canceled the current request.
	if strings.Contains(strings.ToLower(rerr.Error().Error()), consts.OperationCanceledErrorMessage) {
		klog.V(3).Infof("SecurityGroup cache for %s is cleanup because CreateOrUpdateSecurityGroup is canceled by another operation", *sg.Name)
		_ = az.nsgCache.Delete(*sg.Name)
	}

	return rerr.Error()
}

func (az *Cloud) newNSGCache() (azcache.Resource, error) {
	getter := func(key string) (interface{}, error) {
		ctx, cancel := getContextWithCancel()
		defer cancel()
		nsg, err := az.SecurityGroupsClient.Get(ctx, az.SecurityGroupResourceGroup, key, "")
		exists, rerr := checkResourceExistsFromError(err)
		if rerr != nil {
			return nil, rerr.Error()
		}

		if !exists {
			klog.V(2).Infof("Security group %q not found", key)
			return nil, nil
		}

		return &nsg, nil
	}

	if az.NsgCacheTTLInSeconds == 0 {
		az.NsgCacheTTLInSeconds = nsgCacheTTLDefaultInSeconds
	}
	return azcache.NewTimedCache(time.Duration(az.NsgCacheTTLInSeconds)*time.Second, getter, az.Config.DisableAPICallCache)
}

func (az *Cloud) getSecurityGroup(crt azcache.AzureCacheReadType) (network.SecurityGroup, error) {
	nsg := network.SecurityGroup{}
	if az.SecurityGroupName == "" {
		return nsg, fmt.Errorf("securityGroupName is not configured")
	}

	securityGroup, err := az.nsgCache.GetWithDeepCopy(az.SecurityGroupName, crt)
	if err != nil {
		return nsg, err
	}

	if securityGroup == nil {
		return nsg, fmt.Errorf("nsg %q not found", az.SecurityGroupName)
	}

	return *(securityGroup.(*network.SecurityGroup)), nil
}
