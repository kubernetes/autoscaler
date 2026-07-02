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

package azure

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v6"
	"github.com/stretchr/testify/assert"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/azure/deallocate"
)

// TestInstanceStatusFromVMErrorCodeByScaleDownModeSelfHosted verifies error codes
// for self-hosted clusters with enableFastDeleteOnFailedProvisioning=false.
// In this case, only deallocate mode generates error codes; delete mode returns InstanceRunning.
func TestInstanceStatusFromVMErrorCodeByScaleDownModeSelfHosted(t *testing.T) {
	tests := map[string]struct {
		scaleDownPolicy   deallocate.ScaleDownPolicy
		expectedState     cloudprovider.InstanceState
		expectedErrorCode string
	}{
		"deallocate mode still gets start-deallocated-failed without fast delete": {
			scaleDownPolicy:   deallocate.Deallocate,
			expectedState:     cloudprovider.InstanceCreating,
			expectedErrorCode: "start-deallocated-failed",
		},
		"delete mode falls back to InstanceRunning without fast delete": {
			scaleDownPolicy:   deallocate.Delete,
			expectedState:     cloudprovider.InstanceRunning,
			expectedErrorCode: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			provider := newTestProvider(t)
			scaleSet := &ScaleSet{
				azureRef:                             azureRef{Name: "testScaleSet"},
				manager:                              provider.azureManager,
				minSize:                              1,
				maxSize:                              5,
				enableFastDeleteOnFailedProvisioning: false, // self-hosted: fast delete disabled
				scaleDownPolicy:                      tt.scaleDownPolicy,
			}

			vm := newVMObjectWithState(string(armcompute.GalleryProvisioningStateFailed), vmPowerStateStopped)
			status := scaleSet.instanceStatusFromVM(vm)

			assert.NotNil(t, status)
			assert.Equal(t, tt.expectedState, status.State)
			if tt.expectedErrorCode != "" {
				assert.NotNil(t, status.ErrorInfo)
				assert.Equal(t, cloudprovider.OutOfResourcesErrorClass, status.ErrorInfo.ErrorClass)
				assert.Equal(t, tt.expectedErrorCode, status.ErrorInfo.ErrorCode)
			} else {
				// In delete mode with fast delete off, no error info.
				assert.Nil(t, status.ErrorInfo)
			}
		})
	}
}
