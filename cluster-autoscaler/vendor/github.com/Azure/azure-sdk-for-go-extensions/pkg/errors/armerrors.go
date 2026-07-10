/*
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

package errors

import (
	"errors"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

// IsResponseError checks if the error is of type *azcore.ResponseError
// and returns the response error or nil if it's not.
func IsResponseError(err error) *azcore.ResponseError {
	var azErr *azcore.ResponseError
	if errors.As(err, &azErr) && err != nil {
		return azErr
	}
	return nil
}

// IsNotFoundErr is used to determine if we are failing to find a resource within azure.
func IsNotFoundErr(err error) bool {
	azErr := IsResponseError(err)
	return azErr != nil && azErr.ErrorCode == ResourceNotFound
}

// ZonalAllocationFailureOccurred communicates if we have failed to allocate a resource in a zone, and should try another zone.
// To learn more about zonal allocation failures, visit: http://aka.ms/allocation-guidance
func ZonalAllocationFailureOccurred(err error) bool {
	azErr := IsResponseError(err)
	return azErr != nil && azErr.ErrorCode == ZoneAllocationFailed
}

// SKUFamilyQuotaHasBeenReached tells us if we have exceeded our Quota.
func SKUFamilyQuotaHasBeenReached(err error) bool {
	azErr := IsResponseError(err)
	return azErr != nil && azErr.ErrorCode == OperationNotAllowed && strings.Contains(azErr.Error(), SKUFamilyQuotaExceededTerm)
}

// SubscriptionQuotaHasBeenReached tells us if we have exceeded our Quota.
func SubscriptionQuotaHasBeenReached(err error) bool {
	azErr := IsResponseError(err)
	return azErr != nil && azErr.ErrorCode == OperationNotAllowed && strings.Contains(azErr.Error(), SubscriptionQuotaExceededTerm)
}


// RegionalQuotaHasBeenReached communicates if we have reached the quota limit for a given region under a specific subscription
func RegionalQuotaHasBeenReached(err error) bool {
	azErr := IsResponseError(err)
	return azErr != nil && azErr.ErrorCode == OperationNotAllowed && strings.Contains(azErr.Error(), RegionalQuotaExceededTerm)
}

// LowPriorityQuotaHasBeenReached communicates if we have reached the quota limit for low priority VMs under a specific subscription 
// Low priority VMs are generally Spot VMs, but can also be low priority VMs created via the Azure CLI or Azure Portal 
func LowPriorityQuotaHasBeenReached(err error) bool {
	azErr := IsResponseError(err)
	return azErr != nil && azErr.ErrorCode == OperationNotAllowed && strings.Contains(azErr.Error(), LowPriorityQuotaExceededTerm)
}
// IsNicReservedForAnotherVM occurs when a NIC is associated with another VM during deletion. See https://aka.ms/deletenic 
func IsNicReservedForAnotherVM(err error) bool {
	azErr := IsResponseError(err)
	return azErr != nil && azErr.ErrorCode == NicReservedForAnotherVM
}

// IsSKUNotAvailable https://aka.ms/azureskunotavailable: either not available for a location or zone, or out of capacity for Spot.
func IsSKUNotAvailable(err error) bool {
	azErr := IsResponseError(err)
	return azErr != nil && azErr.ErrorCode == SKUNotAvailableErrorCode
}
