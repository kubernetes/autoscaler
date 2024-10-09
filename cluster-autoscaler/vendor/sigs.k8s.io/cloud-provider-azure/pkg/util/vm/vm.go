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

package vm

import (
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"

	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"

	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	stringutils "sigs.k8s.io/cloud-provider-azure/pkg/util/string"
)

// GetVMPowerState returns the power state of the VM
func GetVMPowerState(vmName string, vmStatuses *[]compute.InstanceViewStatus) string {
	logger := klog.Background().WithName("getVMSSVMPowerState").WithValues("vmName", vmName)
	if vmStatuses != nil {
		for _, status := range *vmStatuses {
			state := ptr.Deref(status.Code, "")
			if stringutils.HasPrefixCaseInsensitive(state, consts.VMPowerStatePrefix) {
				return strings.TrimPrefix(state, consts.VMPowerStatePrefix)
			}
		}
	}
	logger.V(3).Info("vm status is nil in the instance view or there is no power state in the status")
	return consts.VMPowerStateUnknown
}

// IsNotActiveVMState checks if the VM is in the active states
func IsNotActiveVMState(provisioningState, powerState string) bool {
	return strings.EqualFold(provisioningState, consts.ProvisioningStateDeleting) ||
		strings.EqualFold(provisioningState, consts.ProvisioningStateUnknown) ||
		strings.EqualFold(powerState, consts.VMPowerStateUnknown) ||
		strings.EqualFold(powerState, consts.VMPowerStateStopped) ||
		strings.EqualFold(powerState, consts.VMPowerStateStopping) ||
		strings.EqualFold(powerState, consts.VMPowerStateDeallocated) ||
		strings.EqualFold(powerState, consts.VMPowerStateDeallocating)
}
