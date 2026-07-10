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

package provider

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"

	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

var (
	vmCacheTTLDefaultInSeconds           = 60
	loadBalancerCacheTTLDefaultInSeconds = 120
	publicIPCacheTTLDefaultInSeconds     = 120

	azureNodeProviderIDRE    = regexp.MustCompile(`^azure:///subscriptions/(?:.*)/resourceGroups/(?:.*)/providers/Microsoft.Compute/(?:.*)`)
	azureResourceGroupNameRE = regexp.MustCompile(`.*/subscriptions/(?:.*)/resourceGroups/(.+)/providers/(?:.*)`)
)

// checkExistsFromError inspects an error and returns a true if err is nil,
// false if error is an autorest.Error with StatusCode=404 and will return the
// error back if error is another status code or another type of error.
func checkResourceExistsFromError(err error) (bool, error) {
	if err == nil {
		return true, nil
	}
	var rerr *azcore.ResponseError
	if errors.As(err, &rerr) {
		if rerr.StatusCode == http.StatusNotFound {
			return false, nil
		}
	}

	return false, err
}

// IsNodeUnmanaged returns true if the node is not managed by Azure cloud provider.
// Those nodes includes on-prem or VMs from other clouds. They will not be added to load balancer
// backends. Azure routes and managed disks are also not supported for them.
func (az *Cloud) IsNodeUnmanaged(nodeName string) (bool, error) {
	unmanagedNodes, err := az.GetUnmanagedNodes()
	if err != nil {
		return false, err
	}

	return unmanagedNodes.Has(nodeName), nil
}

// IsNodeUnmanagedByProviderID returns true if the node is not managed by Azure cloud provider.
// All managed node's providerIDs are in format 'azure:///subscriptions/<id>/resourceGroups/<rg>/providers/Microsoft.Compute/.*'
func (az *Cloud) IsNodeUnmanagedByProviderID(providerID string) bool {
	return !azureNodeProviderIDRE.Match([]byte(providerID))
}

// ConvertResourceGroupNameToLower converts the resource group name in the resource ID to be lowered.
func ConvertResourceGroupNameToLower(resourceID string) (string, error) {
	matches := azureResourceGroupNameRE.FindStringSubmatch(resourceID)
	if len(matches) != 2 {
		return "", fmt.Errorf("%q isn't in Azure resource ID format %q", resourceID, azureResourceGroupNameRE.String())
	}

	resourceGroup := matches[1]
	return strings.Replace(resourceID, resourceGroup, strings.ToLower(resourceGroup), 1), nil
}

func (az *Cloud) useSharedLoadBalancerHealthProbeMode() bool {
	return strings.EqualFold(az.ClusterServiceLoadBalancerHealthProbeMode, consts.ClusterServiceLoadBalancerHealthProbeModeShared)
}
