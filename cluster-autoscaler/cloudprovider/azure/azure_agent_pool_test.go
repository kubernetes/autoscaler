/*
Copyright 2019 The Kubernetes Authors.

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
	"testing"
	"time"

	"k8s.io/legacy-cloud-providers/azure/clients/vmclient/mockvmclient"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func newTestAgentPool(manager *AzureManager, name string) *AgentPool {
	return &AgentPool{
		azureRef: azureRef{
			Name: name,
		},
		manager: manager,
		minSize: 1,
		maxSize: 5,
	}
}

func TestDeleteOutdatedDeployments(t *testing.T) {
	timeLayout := "2006-01-02 15:04:05"
	timeBenchMark, _ := time.Parse(timeLayout, "2000-01-01 00:00:00")

	testCases := []struct {
		deployments              map[string]resources.DeploymentExtended
		expectedDeploymentsNames map[string]bool
		expectedErr              error
		desc                     string
	}{
		{
			deployments: map[string]resources.DeploymentExtended{
				"non-cluster-autoscaler-0000": {
					Name: to.StringPtr("non-cluster-autoscaler-0000"),
					Properties: &resources.DeploymentPropertiesExtended{
						ProvisioningState: to.StringPtr("Succeeded"),
						Timestamp:         &date.Time{Time: timeBenchMark.Add(2 * time.Minute)},
					},
				},
				"cluster-autoscaler-0000": {
					Name: to.StringPtr("cluster-autoscaler-0000"),
					Properties: &resources.DeploymentPropertiesExtended{
						ProvisioningState: to.StringPtr("Succeeded"),
						Timestamp:         &date.Time{Time: timeBenchMark},
					},
				},
				"cluster-autoscaler-0001": {
					Name: to.StringPtr("cluster-autoscaler-0001"),
					Properties: &resources.DeploymentPropertiesExtended{
						ProvisioningState: to.StringPtr("Succeeded"),
						Timestamp:         &date.Time{Time: timeBenchMark.Add(time.Minute)},
					},
				},
				"cluster-autoscaler-0002": {
					Name: to.StringPtr("cluster-autoscaler-0002"),
					Properties: &resources.DeploymentPropertiesExtended{
						ProvisioningState: to.StringPtr("Succeeded"),
						Timestamp:         &date.Time{Time: timeBenchMark.Add(2 * time.Minute)},
					},
				},
			},
			expectedDeploymentsNames: map[string]bool{
				"non-cluster-autoscaler-0000": true,
				"cluster-autoscaler-0001":     true,
				"cluster-autoscaler-0002":     true,
			},
			expectedErr: nil,
			desc:        "cluster autoscaler provider azure should delete outdated deployments created by cluster autoscaler",
		},
	}

	for _, test := range testCases {
		testAS := newTestAgentPool(newTestAzureManager(t), "testAS")
		testAS.manager.azClient.deploymentsClient = &DeploymentsClientMock{
			FakeStore: test.deployments,
		}

		err := testAS.deleteOutdatedDeployments()
		assert.Equal(t, test.expectedErr, err, test.desc)
		existedDeployments, err := testAS.manager.azClient.deploymentsClient.List(context.Background(), "", "", to.Int32Ptr(0))
		existedDeploymentsNames := make(map[string]bool)
		for _, deployment := range existedDeployments {
			existedDeploymentsNames[*deployment.Name] = true
		}
		assert.Equal(t, test.expectedDeploymentsNames, existedDeploymentsNames, test.desc)
	}
}

func TestGetVirtualMachinesFromCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testAS := newTestAgentPool(newTestAzureManager(t), "testAS")
	expectedVMs := []compute.VirtualMachine{
		{
			Tags: map[string]*string{"poolName": to.StringPtr("testAS")},
		},
	}

	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	mockVMClient.EXPECT().List(gomock.Any(), testAS.manager.config.ResourceGroup).Return(expectedVMs, nil)
	testAS.manager.azClient.virtualMachinesClient = mockVMClient

	vms, err := testAS.getVirtualMachinesFromCache()
	assert.Equal(t, 1, len(vms))
	assert.NoError(t, err)
}
