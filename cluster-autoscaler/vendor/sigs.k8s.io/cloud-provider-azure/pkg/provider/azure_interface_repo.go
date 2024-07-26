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

// CreateOrUpdateInterface invokes az.InterfacesClient.CreateOrUpdate with exponential backoff retry
func (az *Cloud) CreateOrUpdateInterface(service *v1.Service, nic network.Interface) error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	rerr := az.InterfacesClient.CreateOrUpdate(ctx, az.ResourceGroup, *nic.Name, nic)
	klog.V(10).Infof("InterfacesClient.CreateOrUpdate(%s): end", *nic.Name)
	if rerr != nil {
		klog.Errorf("InterfacesClient.CreateOrUpdate(%s) failed: %s", *nic.Name, rerr.Error().Error())
		az.Event(service, v1.EventTypeWarning, "CreateOrUpdateInterface", rerr.Error().Error())
		return rerr.Error()
	}

	return nil
}
