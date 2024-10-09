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
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"k8s.io/klog/v2"

	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"
)

// CreateOrUpdateVMSS invokes az.VirtualMachineScaleSetsClient.Update().
func (az *Cloud) CreateOrUpdateVMSS(resourceGroupName string, VMScaleSetName string, parameters compute.VirtualMachineScaleSet) *retry.Error {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	// When vmss is being deleted, CreateOrUpdate API would report "the vmss is being deleted" error.
	// Since it is being deleted, we shouldn't send more CreateOrUpdate requests for it.
	klog.V(3).Infof("CreateOrUpdateVMSS: verify the status of the vmss being created or updated")
	vmss, rerr := az.VirtualMachineScaleSetsClient.Get(ctx, resourceGroupName, VMScaleSetName)
	if rerr != nil {
		klog.Errorf("CreateOrUpdateVMSS: error getting vmss(%s): %v", VMScaleSetName, rerr)
		return rerr
	}
	if vmss.ProvisioningState != nil && strings.EqualFold(*vmss.ProvisioningState, consts.ProvisionStateDeleting) {
		klog.V(3).Infof("CreateOrUpdateVMSS: found vmss %s being deleted, skipping", VMScaleSetName)
		return nil
	}

	rerr = az.VirtualMachineScaleSetsClient.CreateOrUpdate(ctx, resourceGroupName, VMScaleSetName, parameters)
	klog.V(10).Infof("UpdateVmssVMWithRetry: VirtualMachineScaleSetsClient.CreateOrUpdate(%s): end", VMScaleSetName)
	if rerr != nil {
		klog.Errorf("CreateOrUpdateVMSS: error CreateOrUpdate vmss(%s): %v", VMScaleSetName, rerr)
		return rerr
	}

	return nil
}
