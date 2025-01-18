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
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/storageaccountclient/mockstorageaccountclient"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmclient/mockvmclient"
	providerazureconsts "sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/retry"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2021-09-01/storage"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	rerrTooManyReqs       = retry.Error{HTTPStatusCode: http.StatusTooManyRequests}
	rerrInternalErr       = retry.Error{HTTPStatusCode: http.StatusInternalServerError}
	testValidProviderID0  = "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/as-vm-0"
	testValidProviderID1  = "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/as-vm-1"
	testInvalidProviderID = "/subscriptions/sub/resourceGroups/rg/providers/provider/virtualMachines/as-vm-0/"
)

func newTestAgentPool(manager *AzureManager, name string) *AgentPool {
	return &AgentPool{
		azureRef: azureRef{
			Name: name,
		},
		manager:    manager,
		minSize:    1,
		maxSize:    5,
		parameters: make(map[string]interface{}),
		template:   make(map[string]interface{}),
	}
}

func getExpectedVMs() []compute.VirtualMachine {
	expectedVMs := []compute.VirtualMachine{
		{
			Name: to.StringPtr("000-0-00000000-0"),
			ID:   to.StringPtr("/subscriptions/sub/resourceGroups/rg/providers/provider/0"),
			Tags: map[string]*string{"poolName": to.StringPtr("as")},
			VirtualMachineProperties: &compute.VirtualMachineProperties{
				StorageProfile: &compute.StorageProfile{
					OsDisk: &compute.OSDisk{
						OsType: compute.OperatingSystemTypesLinux,
						Vhd:    &compute.VirtualHardDisk{URI: to.StringPtr("https://foo.blob/vhds/bar.vhd")},
					},
				},
				NetworkProfile: &compute.NetworkProfile{
					NetworkInterfaces: &[]compute.NetworkInterfaceReference{
						{},
					},
				},
			},
		},
		{
			Name: to.StringPtr("00000000001"),
			ID:   to.StringPtr("/subscriptions/sub/resourceGroups/rg/providers/provider/0"),
			Tags: map[string]*string{"poolName": to.StringPtr("as")},
			VirtualMachineProperties: &compute.VirtualMachineProperties{
				StorageProfile: &compute.StorageProfile{
					OsDisk: &compute.OSDisk{OsType: compute.OperatingSystemTypesWindows},
				},
			},
		},
	}

	return expectedVMs
}

func TestInitialize(t *testing.T) {
	as := newTestAgentPool(newTestAzureManager(t), "as")

	err := as.initialize()
	assert.NoError(t, err)
	assert.NotNil(t, as.template)
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
		testAS.manager.azClient.deploymentClient = &DeploymentClientMock{
			FakeStore: test.deployments,
		}

		err := testAS.deleteOutdatedDeployments()
		assert.Equal(t, test.expectedErr, err, test.desc)
		existedDeployments, _ := testAS.manager.azClient.deploymentClient.List(context.Background(), "")
		existedDeploymentsNames := make(map[string]bool)
		for _, deployment := range existedDeployments {
			existedDeploymentsNames[*deployment.Name] = true
		}
		assert.Equal(t, test.expectedDeploymentsNames, existedDeploymentsNames, test.desc)
	}
}

func TestAgentPoolGetVMsFromCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testAS := newTestAgentPool(newTestAzureManager(t), "testAS")
	expectedVMs := []compute.VirtualMachine{
		{
			Tags: map[string]*string{"poolName": to.StringPtr("testAS")},
		},
	}

	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	testAS.manager.azClient.virtualMachinesClient = mockVMClient
	mockVMClient.EXPECT().List(gomock.Any(), testAS.manager.config.ResourceGroup).Return(expectedVMs, nil)
	testAS.manager.config.VMType = providerazureconsts.VMTypeStandard
	ac, err := newAzureCache(testAS.manager.azClient, refreshInterval, *testAS.manager.config)
	assert.NoError(t, err)
	testAS.manager.azureCache = ac

	vms, err := testAS.getVMsFromCache()
	assert.Equal(t, 1, len(vms))
	assert.NoError(t, err)
}

func TestGetVMIndexes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	as := newTestAgentPool(newTestAzureManager(t), "as")
	expectedVMs := getExpectedVMs()
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	as.manager.azClient.virtualMachinesClient = mockVMClient
	mockVMClient.EXPECT().List(gomock.Any(), as.manager.config.ResourceGroup).Return(expectedVMs, nil)
	as.manager.config.VMType = providerazureconsts.VMTypeStandard
	ac, err := newAzureCache(as.manager.azClient, refreshInterval, *as.manager.config)
	assert.NoError(t, err)
	as.manager.azureCache = ac

	sortedIndexes, indexToVM, err := as.GetVMIndexes()
	assert.NoError(t, err)
	assert.Equal(t, 2, len(sortedIndexes))
	assert.Equal(t, 2, len(indexToVM))

	expectedVMs[0].ID = to.StringPtr("foo")
	mockVMClient.EXPECT().List(gomock.Any(), as.manager.config.ResourceGroup).Return(expectedVMs, nil)
	err = as.manager.forceRefresh()
	assert.NoError(t, err)
	sortedIndexes, indexToVM, err = as.GetVMIndexes()
	expectedErr := fmt.Errorf("\"azure://foo\" isn't in Azure resource ID format")
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, sortedIndexes)
	assert.Nil(t, indexToVM)

	expectedVMs[0].Name = to.StringPtr("foo")
	mockVMClient.EXPECT().List(gomock.Any(), as.manager.config.ResourceGroup).Return(expectedVMs, nil)
	err = as.manager.forceRefresh()
	sortedIndexes, indexToVM, err = as.GetVMIndexes()
	expectedErr = fmt.Errorf("resource name was missing from identifier")
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, sortedIndexes)
	assert.Nil(t, indexToVM)
}

func TestGetCurSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	as := newTestAgentPool(newTestAzureManager(t), "as")
	as.curSize = 1
	expectedVMs := getExpectedVMs()
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	as.manager.azClient.virtualMachinesClient = mockVMClient
	mockVMClient.EXPECT().List(gomock.Any(), as.manager.config.ResourceGroup).Return(expectedVMs, nil)
	as.manager.config.VMType = providerazureconsts.VMTypeStandard
	ac, err := newAzureCache(as.manager.azClient, refreshInterval, *as.manager.config)
	assert.NoError(t, err)
	as.manager.azureCache = ac

	as.lastRefresh = time.Now()
	curSize, err := as.getCurSize()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), curSize)

	as.lastRefresh = time.Now().Add(-1 * 3 * time.Minute)
	curSize, err = as.getCurSize()
	assert.NoError(t, err)
	assert.Equal(t, int64(2), curSize)
}

func TestAgentPoolTargetSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	as := newTestAgentPool(newTestAzureManager(t), "as")
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	as.manager.azClient.virtualMachinesClient = mockVMClient
	expectedVMs := getExpectedVMs()
	mockVMClient.EXPECT().List(gomock.Any(), as.manager.config.ResourceGroup).Return(expectedVMs, nil)
	as.manager.config.VMType = providerazureconsts.VMTypeStandard
	ac, err := newAzureCache(as.manager.azClient, refreshInterval, *as.manager.config)
	assert.NoError(t, err)
	as.manager.azureCache = ac

	as.lastRefresh = time.Now().Add(-1 * 15 * time.Second)
	size, err := as.getCurSize()
	assert.NoError(t, err)
	assert.Equal(t, int64(2), size)
}

func TestAgentPoolIncreaseSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	as := newTestAgentPool(newTestAzureManager(t), "as")
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	as.manager.azClient.virtualMachinesClient = mockVMClient
	expectedVMs := getExpectedVMs()
	mockVMClient.EXPECT().List(gomock.Any(), as.manager.config.ResourceGroup).Return(expectedVMs, nil).MaxTimes(2)
	as.manager.config.VMType = providerazureconsts.VMTypeStandard
	ac, err := newAzureCache(as.manager.azClient, refreshInterval, *as.manager.config)
	assert.NoError(t, err)
	as.manager.azureCache = ac

	err = as.IncreaseSize(-1)
	expectedErr := fmt.Errorf("size increase must be positive")
	assert.Equal(t, expectedErr, err)

	mockVMClient.EXPECT().List(gomock.Any(), as.manager.config.ResourceGroup).Return(expectedVMs, nil).MaxTimes(2)
	err = as.manager.Refresh()
	assert.NoError(t, err)
	err = as.IncreaseSize(4)
	expectedErr = fmt.Errorf("size increase too large - desired:6 max:5")

	err = as.IncreaseSize(2)
	assert.NoError(t, err)
}

func TestAgentPoolDecreaseTargetSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	as := newTestAgentPool(newTestAzureManager(t), "as")
	as.curSize = 3
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	as.manager.azClient.virtualMachinesClient = mockVMClient
	expectedVMs := getExpectedVMs()
	mockVMClient.EXPECT().List(gomock.Any(), as.manager.config.ResourceGroup).Return(expectedVMs, nil).MaxTimes(3)
	as.manager.config.VMType = providerazureconsts.VMTypeStandard
	ac, err := newAzureCache(as.manager.azClient, refreshInterval, *as.manager.config)
	assert.NoError(t, err)
	as.manager.azureCache = ac

	err = as.DecreaseTargetSize(-1)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), as.curSize)

	mockVMClient.EXPECT().List(gomock.Any(), as.manager.config.ResourceGroup).Return(expectedVMs, nil).MaxTimes(2)
	err = as.manager.Refresh()
	assert.NoError(t, err)
	err = as.DecreaseTargetSize(-1)
	expectedErr := fmt.Errorf("attempt to delete existing nodes targetSize:2 delta:-1 existingNodes: 2")
	assert.Equal(t, expectedErr, err)
}

func TestAgentPoolBelongs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	as := newTestAgentPool(newTestAzureManager(t), "as")
	as.manager.azureCache.instanceToNodeGroup[azureRef{Name: testValidProviderID0}] = as

	flag, err := as.Belongs(&apiv1.Node{Spec: apiv1.NodeSpec{ProviderID: testValidProviderID0}})
	assert.NoError(t, err)
	assert.True(t, flag)

	flag, err = as.Belongs(&apiv1.Node{
		Spec:       apiv1.NodeSpec{ProviderID: testInvalidProviderID},
		ObjectMeta: v1.ObjectMeta{Name: "node"},
	})
	expectedErr := fmt.Errorf("node doesn't belong to a known agent pool")
	assert.Equal(t, expectedErr, err)
	assert.False(t, flag)

	as1 := newTestAgentPool(newTestAzureManager(t), "as1")
	as1.manager.azureCache.instanceToNodeGroup[azureRef{Name: testValidProviderID0}] = as
	flag, err = as1.Belongs(&apiv1.Node{Spec: apiv1.NodeSpec{ProviderID: testValidProviderID0}})
	assert.NoError(t, err)
	assert.False(t, flag)
}

func TestDeleteInstances(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	as := newTestAgentPool(newTestAzureManager(t), "as")
	as1 := newTestAgentPool(newTestAzureManager(t), "as1")
	as.manager.azureCache.instanceToNodeGroup[azureRef{Name: testValidProviderID0}] = as
	as.manager.azureCache.instanceToNodeGroup[azureRef{Name: testValidProviderID1}] = as1
	as.manager.azureCache.instanceToNodeGroup[azureRef{Name: testInvalidProviderID}] = as

	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	as.manager.azClient.virtualMachinesClient = mockVMClient

	mockSAClient := mockstorageaccountclient.NewMockInterface(ctrl)
	as.manager.azClient.storageAccountsClient = mockSAClient

	err := as.DeleteInstances([]*azureRef{})
	assert.NoError(t, err)

	instances := []*azureRef{
		{Name: "foo"},
	}
	err = as.DeleteInstances(instances)
	expectedErr := fmt.Errorf("\"foo\" isn't in Azure resource ID format")
	assert.Equal(t, expectedErr, err)

	instances = []*azureRef{
		{Name: testValidProviderID0},
		{Name: "foo"},
	}
	err = as.DeleteInstances(instances)
	assert.Equal(t, expectedErr, err)

	instances = []*azureRef{
		{Name: testInvalidProviderID},
	}
	err = as.DeleteInstances(instances)
	expectedErr = fmt.Errorf("resource name was missing from identifier")
	assert.Equal(t, expectedErr, err)

	instances = []*azureRef{
		{Name: testValidProviderID0},
		{Name: testValidProviderID1},
	}

	err = as.DeleteInstances(instances)
	expectedErr = fmt.Errorf("cannot delete instance (%s) which don't belong to the same node pool (\"as\")", testValidProviderID1)
	assert.Equal(t, expectedErr, err)

	instances = []*azureRef{
		{Name: testValidProviderID0},
	}
	mockVMClient.EXPECT().Get(gomock.Any(), as.manager.config.ResourceGroup, "as-vm-0", gomock.Any()).Return(getExpectedVMs()[0], nil)
	mockVMClient.EXPECT().Delete(gomock.Any(), as.manager.config.ResourceGroup, "as-vm-0").Return(nil)
	mockSAClient.EXPECT().ListKeys(gomock.Any(), as.manager.config.SubscriptionID, as.manager.config.ResourceGroup, "foo").Return(storage.AccountListKeysResult{
		Keys: &[]storage.AccountKey{
			{Value: to.StringPtr("dmFsdWUK")},
		},
	}, nil)
	err = as.DeleteInstances(instances)
	expectedErrStr := "The specified account is disabled."
	assert.True(t, strings.Contains(err.Error(), expectedErrStr))
}

func TestAgentPoolDeleteNodes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	as := newTestAgentPool(newTestAzureManager(t), "as")
	as.manager.azureCache.instanceToNodeGroup[azureRef{Name: testValidProviderID0}] = as
	expectedVMs := getExpectedVMs()
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	as.manager.azClient.virtualMachinesClient = mockVMClient
	mockSAClient := mockstorageaccountclient.NewMockInterface(ctrl)
	as.manager.azClient.storageAccountsClient = mockSAClient
	mockVMClient.EXPECT().List(gomock.Any(), as.manager.config.ResourceGroup).Return(expectedVMs, nil)
	as.manager.config.VMType = providerazureconsts.VMTypeStandard
	ac, err := newAzureCache(as.manager.azClient, refreshInterval, *as.manager.config)
	as.manager.config.VMType = providerazureconsts.VMTypeVMSS
	assert.NoError(t, err)
	as.manager.azureCache = ac

	err = as.DeleteNodes([]*apiv1.Node{
		{
			Spec:       apiv1.NodeSpec{ProviderID: testInvalidProviderID},
			ObjectMeta: v1.ObjectMeta{Name: "node"},
		},
	})
	expectedErr := fmt.Errorf("node doesn't belong to a known agent pool")
	assert.Equal(t, expectedErr, err)

	as1 := newTestAgentPool(newTestAzureManager(t), "as1")
	as.manager.azureCache.instanceToNodeGroup[azureRef{Name: testValidProviderID0}] = as1
	err = as.DeleteNodes([]*apiv1.Node{
		{
			Spec:       apiv1.NodeSpec{ProviderID: testValidProviderID0},
			ObjectMeta: v1.ObjectMeta{Name: "node"},
		},
	})
	expectedErr = fmt.Errorf("node belongs to a different asg than as")
	assert.Equal(t, expectedErr, err)

	as.manager.azureCache.instanceToNodeGroup[azureRef{Name: testValidProviderID0}] = as
	mockVMClient.EXPECT().Get(gomock.Any(), as.manager.config.ResourceGroup, "as-vm-0", gomock.Any()).Return(getExpectedVMs()[0], nil)
	mockVMClient.EXPECT().Delete(gomock.Any(), as.manager.config.ResourceGroup, "as-vm-0").Return(nil)
	mockSAClient.EXPECT().ListKeys(gomock.Any(), as.manager.config.SubscriptionID, as.manager.config.ResourceGroup, "foo").Return(storage.AccountListKeysResult{
		Keys: &[]storage.AccountKey{
			{Value: to.StringPtr("dmFsdWUK")},
		},
	}, nil)
	err = as.DeleteNodes([]*apiv1.Node{
		{
			Spec:       apiv1.NodeSpec{ProviderID: testValidProviderID0},
			ObjectMeta: v1.ObjectMeta{Name: "node"},
		},
	})
	expectedErrStr := "The specified account is disabled."
	assert.True(t, strings.Contains(err.Error(), expectedErrStr))

	as.minSize = 3
	err = as.DeleteNodes([]*apiv1.Node{})
	expectedErr = fmt.Errorf("min size reached, nodes will not be deleted")
	assert.Equal(t, expectedErr, err)
}

func TestAgentPoolNodes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	as := newTestAgentPool(newTestAzureManager(t), "as")
	expectedVMs := []compute.VirtualMachine{
		{
			Tags: map[string]*string{"poolName": to.StringPtr("as")},
			ID:   to.StringPtr(""),
		},
		{
			Tags: map[string]*string{"poolName": to.StringPtr("as")},
			ID:   &testValidProviderID0,
		},
	}

	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	as.manager.azClient.virtualMachinesClient = mockVMClient
	mockVMClient.EXPECT().List(gomock.Any(), as.manager.config.ResourceGroup).Return(expectedVMs, nil)
	as.manager.config.VMType = providerazureconsts.VMTypeStandard
	ac, err := newAzureCache(as.manager.azClient, refreshInterval, *as.manager.config)
	assert.NoError(t, err)
	as.manager.azureCache = ac

	nodes, err := as.Nodes()
	assert.NoError(t, err)
	assert.Equal(t, 1, len(nodes))

	expectedVMs = []compute.VirtualMachine{
		{
			Tags: map[string]*string{"poolName": to.StringPtr("as")},
			ID:   to.StringPtr("foo"),
		},
	}
	mockVMClient.EXPECT().List(gomock.Any(), as.manager.config.ResourceGroup).Return(expectedVMs, nil)
	err = as.manager.forceRefresh()
	assert.NoError(t, err)
	nodes, err = as.Nodes()
	expectedErr := fmt.Errorf("\"azure://foo\" isn't in Azure resource ID format")
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, nodes)
}
