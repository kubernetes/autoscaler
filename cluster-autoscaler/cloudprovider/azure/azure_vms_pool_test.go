/*
Copyright 2017 The Kubernetes Authors.

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

package azure

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-03-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	apiv1 "k8s.io/api/core/v1"
)

func newTestVMsPool(manager *AzureManager, name string) *VMsPool {
	return &VMsPool{
		azureRef: azureRef{
			Name: name,
		},
		manager: manager,
		minSize: 3,
		maxSize: 10,
	}
}

const (
	fakeVMsPoolVMID = "/subscriptions/test-subscription-id/resourceGroups/test-rg/providers/Microsoft.Compute/virtualMachines/%d"
)

func newTestVMsPoolVMList(count int) []compute.VirtualMachine {
	var vmList []compute.VirtualMachine
	for i := 0; i < count; i++ {
		vm := compute.VirtualMachine{
			ID: to.StringPtr(fmt.Sprintf(fakeVMsPoolVMID, i)),
			VirtualMachineProperties: &compute.VirtualMachineProperties{
				VMID: to.StringPtr(fmt.Sprintf("123E4567-E89B-12D3-A456-426655440000-%d", i)),
			},
			Tags: map[string]*string{
				agentpoolTypeTag: to.StringPtr("VirtualMachines"),
				agentpoolNameTag: to.StringPtr("test-vms-pool"),
			},
		}
		vmList = append(vmList, vm)
	}
	return vmList
}

func newVMsNode(vmID int64) *apiv1.Node {
	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "azure://" + fmt.Sprintf(fakeVMsPoolVMID, vmID),
		},
	}
	return node
}
