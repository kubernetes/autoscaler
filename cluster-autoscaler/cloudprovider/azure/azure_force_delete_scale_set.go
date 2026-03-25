/*
Copyright 2022 The Kubernetes Authors.

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
	"context"
	"strings"

	azerrors "github.com/Azure/azure-sdk-for-go-extensions/pkg/errors"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"

	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
)

// When Azure Dedicated Host is enabled or using isolated vm skus, force deleting a VMSS fails with the following error:
//
// "predominantErrorDetail": {
//   "innererror": {
//     "internalErrorCode": "OperationNotAllowedOnResourceThatManagesUpdatesWithMaintenanceControl"
//   },
//   "code": "OperationNotAllowed",
//   "message": "Operation 'ForceDelete' is not allowed on resource 'aks-newnp-11436513-vmss' since it manages updates using maintenance control."
// },
//
// A programmatically way to determine if a VM size is isolated or not has not been found. The isolated VM documentation:
//     https://docs.microsoft.com/en-us/azure/virtual-machines/isolation
// has the current list of isolated VM sizes, but new isolated VM size could be introduced in the future.
//
// As a result of not being able to find out if a VM size is isolated or not, we'll do the following:
// - if scaleSet has isolated vm size or dedicated host, disable forDelete
// - else use forceDelete
//   - if new isolated sku were added or dedicatedHost was not updated properly, this forceDelete call will fail with above error.
//     In that case, call normal delete (fall-back)

var isolatedVMSizes = map[string]bool{
	strings.ToLower("Standard_E80ids_v4"):   true,
	strings.ToLower("Standard_E80is_v4"):    true,
	strings.ToLower("Standard_E104i_v5"):    true,
	strings.ToLower("Standard_E104is_v5"):   true,
	strings.ToLower("Standard_E104id_v5"):   true,
	strings.ToLower("Standard_E104ids_v5"):  true,
	strings.ToLower("Standard_M192is_v2"):   true,
	strings.ToLower("Standard_M192ims_v2"):  true,
	strings.ToLower("Standard_M192ids_v2"):  true,
	strings.ToLower("Standard_M192idms_v2"): true,
	strings.ToLower("Standard_F72s_v2"):     true,
	strings.ToLower("Standard_M128ms"):      true,
}

func (scaleSet *ScaleSet) deleteInstances(ctx context.Context, requiredIds *armcompute.VirtualMachineScaleSetVMInstanceRequiredIDs, commonAsgId string) (*runtime.Poller[armcompute.VirtualMachineScaleSetsClientDeleteInstancesResponse], error) {
	scaleSet.instanceMutex.Lock()
	defer scaleSet.instanceMutex.Unlock()

	skuName := scaleSet.getSKU()
	resourceGroup := scaleSet.manager.config.ResourceGroup
	forceDelete := shouldForceDelete(skuName, scaleSet)

	poller, err := scaleSet.manager.azClient.vmssClientForDelete.BeginDeleteInstances(ctx, resourceGroup, commonAsgId, *requiredIds, &armcompute.VirtualMachineScaleSetsClientBeginDeleteInstancesOptions{
		ForceDeletion: &forceDelete,
	})
	if forceDelete && isOperationNotAllowed(err) {
		klog.Infof("falling back to normal delete for instances %v for %s", requiredIds.InstanceIDs, scaleSet.Name)
		return scaleSet.manager.azClient.vmssClientForDelete.BeginDeleteInstances(ctx, resourceGroup, commonAsgId, *requiredIds, &armcompute.VirtualMachineScaleSetsClientBeginDeleteInstancesOptions{
			ForceDeletion: ptr.To(false),
		})
	}
	return poller, err
}

func shouldForceDelete(skuName string, scaleSet *ScaleSet) bool {
	return scaleSet.enableForceDelete && !isolatedVMSizes[strings.ToLower(skuName)] && !scaleSet.dedicatedHost
}

// isOperationNotAllowed checks if `error` is an OperationNotAllowed error.
func isOperationNotAllowed(err error) bool {
	if err == nil {
		return false
	}
	if azerr := azerrors.IsResponseError(err); azerr != nil {
		return azerr.ErrorCode == azerrors.OperationNotAllowed
	}
	return strings.Contains(err.Error(), azerrors.OperationNotAllowed)
}
