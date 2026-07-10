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

package errutils

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
)

func CheckResourceExistsFromAzcoreError(err error) (bool, error) {
	if err == nil {
		return true, nil
	}
	var respError *azcore.ResponseError
	if errors.As(err, &respError) && respError != nil {
		if respError.StatusCode == http.StatusNotFound {
			return false, nil
		}
	}
	return false, err
}

// HasStatusForbiddenOrIgnoredError return true if the given error code is part of the error message
// This should only be used when trying to delete resources
func HasStatusForbiddenOrIgnoredError(err error) bool {
	if err == nil {
		return false
	}
	var respError *azcore.ResponseError
	if !errors.As(err, &respError) {
		return false
	}
	if respError == nil {
		return false
	}

	if respError.StatusCode == http.StatusNotFound {
		return true
	}

	if respError.StatusCode == http.StatusForbidden {
		return true
	}
	return false
}

// GetVMSSMetadataByRawError gets the vmss name by parsing the error message
func GetVMSSMetadataByRawError(err error) (string, string, error) {
	if err == nil || !isErrorLoadBalancerInUseByVirtualMachineScaleSet(err.Error()) {
		return "", "", nil
	}

	reg := regexp.MustCompile(`.*/subscriptions/(?:.*)/resourceGroups/(.*)/providers/Microsoft.Compute/virtualMachineScaleSets/(.+)\.`)
	matches := reg.FindStringSubmatch(err.Error())
	if len(matches) != 3 {
		return "", "", fmt.Errorf("GetVMSSMetadataByRawError: couldn't find a VMSS resource Id from error message %w", err)
	}

	return matches[1], matches[2], nil
}

// isErrorLoadBalancerInUseByVirtualMachineScaleSet determines if the Error is
// LoadBalancerInUseByVirtualMachineScaleSet
func isErrorLoadBalancerInUseByVirtualMachineScaleSet(rawError string) bool {
	return strings.Contains(rawError, "LoadBalancerInUseByVirtualMachineScaleSet")
}
