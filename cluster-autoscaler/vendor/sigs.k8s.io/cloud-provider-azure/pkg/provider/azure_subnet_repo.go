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
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// CreateOrUpdateSubnet invokes az.SubnetClient.CreateOrUpdate with exponential backoff retry
func (az *Cloud) CreateOrUpdateSubnet(service *v1.Service, subnet network.Subnet) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	var rg string
	if len(az.VnetResourceGroup) > 0 {
		rg = az.VnetResourceGroup
	} else {
		rg = az.ResourceGroup
	}

	rerr := az.SubnetsClient.CreateOrUpdate(ctx, rg, az.VnetName, *subnet.Name, subnet)
	klog.V(10).Infof("SubnetClient.CreateOrUpdate(%s): end", *subnet.Name)
	if rerr != nil {
		klog.Errorf("SubnetClient.CreateOrUpdate(%s) failed: %s", *subnet.Name, rerr.Error().Error())
		az.Event(service, v1.EventTypeWarning, "CreateOrUpdateSubnet", rerr.Error().Error())
		return rerr.Error()
	}

	return nil
}

func (az *Cloud) getSubnet(virtualNetworkName string, subnetName string) (network.Subnet, bool, error) {
	var rg string
	if len(az.VnetResourceGroup) > 0 {
		rg = az.VnetResourceGroup
	} else {
		rg = az.ResourceGroup
	}

	ctx, cancel := getContextWithCancel()
	defer cancel()
	subnet, err := az.SubnetsClient.Get(ctx, rg, virtualNetworkName, subnetName, "")
	exists, rerr := checkResourceExistsFromError(err)
	if rerr != nil {
		return subnet, false, rerr.Error()
	}

	if !exists {
		klog.V(2).Infof("Subnet %q not found", subnetName)
	}
	return subnet, exists, nil
}
