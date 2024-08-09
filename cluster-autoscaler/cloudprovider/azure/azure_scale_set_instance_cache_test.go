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
	"fmt"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"

	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmssvmclient/mockvmssvmclient"
)

var (
	ctrl                                 *gomock.Controller
	currentTime, expiredTime             time.Time
	provider                             *AzureCloudProvider
	scaleSet                             *ScaleSet
	mockVMSSVMClient                     *mockvmssvmclient.MockInterface
	expectedVMSSVMs                      []compute.VirtualMachineScaleSetVM
	expectedStates                       []cloudprovider.InstanceState
	instanceCache, expectedInstanceCache []cloudprovider.Instance
)

func testGetInstanceCacheWithStates(t *testing.T, vms []compute.VirtualMachineScaleSetVM,
	states []cloudprovider.InstanceState) []cloudprovider.Instance {
	assert.Equal(t, len(vms), len(states))
	var instanceCacheTest []cloudprovider.Instance
	for i := 0; i < len(vms); i++ {
		instanceCacheTest = append(instanceCacheTest, cloudprovider.Instance{
			Id:     azurePrefix + fmt.Sprintf(fakeVirtualMachineScaleSetVMID, i),
			Status: &cloudprovider.InstanceStatus{State: states[i]},
		})
	}
	return instanceCacheTest
}
