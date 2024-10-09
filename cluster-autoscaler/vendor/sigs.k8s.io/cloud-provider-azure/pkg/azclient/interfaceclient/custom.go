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

package interfaceclient

import (
	"context"

	armnetwork "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v4"
)

// List gets a list of IPGroup in the resource group.
func (client *Client) ListVirtualMachineScaleSetNetworkInterfaces(ctx context.Context, resourceGroupName string, virtualMachineScaleSetName string) ([]*armnetwork.Interface, error) {
	pager := client.InterfacesClient.NewListVirtualMachineScaleSetNetworkInterfacesPager(resourceGroupName, virtualMachineScaleSetName, nil)
	var result []*armnetwork.Interface
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		result = append(result, nextResult.Value...)
	}
	return result, nil
}

func (client *Client) GetVirtualMachineScaleSetNetworkInterface(ctx context.Context, resourceGroupName string, virtualMachineScaleSetName string, virtualmachineIndex string, networkInterfaceName string) (*armnetwork.Interface, error) {
	resp, err := client.InterfacesClient.GetVirtualMachineScaleSetNetworkInterface(ctx, resourceGroupName, virtualMachineScaleSetName, virtualmachineIndex, networkInterfaceName, nil)
	if err != nil {
		return nil, err
	}
	//handle statuscode
	return &resp.Interface, nil
}
