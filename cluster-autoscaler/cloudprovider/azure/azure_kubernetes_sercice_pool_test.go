/*
Copyright 2020 The Kubernetes Authors.

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
	"net/http"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-04-01/containerservice"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/legacy-cloud-providers/azure/clients/containerserviceclient/mockcontainerserviceclient"
	"k8s.io/legacy-cloud-providers/azure/clients/interfaceclient/mockinterfaceclient"
	"k8s.io/legacy-cloud-providers/azure/clients/vmclient/mockvmclient"
	"k8s.io/legacy-cloud-providers/azure/retry"
)

const (
	testAKSPoolName = "aks"
)

var (
	errInternal    = &retry.Error{HTTPStatusCode: http.StatusInternalServerError}
	errInternalRaw = fmt.Errorf("Retriable: false, RetryAfter: 0s, HTTPStatusCode: 500, RawError: <nil>")
)

func getTestAKSPool(manager *AzureManager, name string) *AKSAgentPool {
	return &AKSAgentPool{
		azureRef: azureRef{
			Name: name,
		},
		manager:           manager,
		minSize:           1,
		maxSize:           10,
		resourceGroup:     "rg",
		nodeResourceGroup: "rg",
		clusterName:       "cluster",
		curSize:           5,
		util: &AzUtil{
			manager: manager,
		},
	}
}

func getExpectedManagedCluster() containerservice.ManagedCluster {
	return containerservice.ManagedCluster{
		Name: to.StringPtr("cluster"),
		ManagedClusterProperties: &containerservice.ManagedClusterProperties{
			AgentPoolProfiles: &[]containerservice.ManagedClusterAgentPoolProfile{
				{
					Name:  to.StringPtr(testAKSPoolName),
					Count: to.Int32Ptr(1),
				},
			},
		},
	}
}

func TestSetNodeCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	aksPool := getTestAKSPool(newTestAzureManager(t), testAKSPoolName)
	mockAKSClient := mockcontainerserviceclient.NewMockInterface(ctrl)
	mockAKSClient.EXPECT().Get(
		gomock.Any(),
		aksPool.resourceGroup,
		aksPool.clusterName).Return(getExpectedManagedCluster(), nil)
	mockAKSClient.EXPECT().CreateOrUpdate(
		gomock.Any(),
		aksPool.resourceGroup,
		aksPool.clusterName,
		gomock.Any(),
		gomock.Any()).Return(nil)
	aksPool.manager.azClient.managedKubernetesServicesClient = mockAKSClient

	err := aksPool.SetNodeCount(3)
	assert.NoError(t, err)

	mockAKSClient.EXPECT().Get(
		gomock.Any(),
		aksPool.resourceGroup,
		aksPool.clusterName).Return(getExpectedManagedCluster(), errInternal)
	err = aksPool.SetNodeCount(3)
	expectedErr := errInternalRaw
	assert.Equal(t, expectedErr, err)

	mockAKSClient.EXPECT().Get(
		gomock.Any(),
		aksPool.resourceGroup,
		aksPool.clusterName).Return(containerservice.ManagedCluster{
		ManagedClusterProperties: &containerservice.ManagedClusterProperties{
			AgentPoolProfiles: &[]containerservice.ManagedClusterAgentPoolProfile{},
		},
	}, nil)
	err = aksPool.SetNodeCount(3)
	expectedErr = fmt.Errorf("could not find pool with name: {aks}")
	assert.Equal(t, expectedErr, err)

	mockAKSClient.EXPECT().Get(
		gomock.Any(),
		aksPool.resourceGroup,
		aksPool.clusterName).Return(getExpectedManagedCluster(), nil)
	mockAKSClient.EXPECT().CreateOrUpdate(
		gomock.Any(),
		aksPool.resourceGroup,
		aksPool.clusterName,
		gomock.Any(),
		gomock.Any()).Return(errInternal)
	err = aksPool.SetNodeCount(3)
	expectedErr = errInternalRaw
	assert.Equal(t, expectedErr, err)
}

func TestGetNodeCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	aksPool := getTestAKSPool(newTestAzureManager(t), testAKSPoolName)
	mockAKSClient := mockcontainerserviceclient.NewMockInterface(ctrl)
	mockAKSClient.EXPECT().Get(
		gomock.Any(),
		aksPool.resourceGroup,
		aksPool.clusterName).Return(getExpectedManagedCluster(), nil)
	aksPool.manager.azClient.managedKubernetesServicesClient = mockAKSClient

	count, err := aksPool.GetNodeCount()
	assert.Equal(t, 1, count)
	assert.NoError(t, err)

	mockAKSClient.EXPECT().Get(gomock.Any(),
		aksPool.resourceGroup,
		aksPool.clusterName).Return(getExpectedManagedCluster(), errInternal)
	count, err = aksPool.GetNodeCount()
	expectedErr := errInternalRaw
	assert.Equal(t, -1, count)
	assert.Equal(t, expectedErr, err)

	mockAKSClient.EXPECT().Get(
		gomock.Any(),
		aksPool.resourceGroup,
		aksPool.clusterName).Return(containerservice.ManagedCluster{
		ManagedClusterProperties: &containerservice.ManagedClusterProperties{
			AgentPoolProfiles: &[]containerservice.ManagedClusterAgentPoolProfile{},
		},
	}, nil)
	count, err = aksPool.GetNodeCount()
	expectedErr = fmt.Errorf("could not find pool with name: {aks}")
	assert.Equal(t, -1, count)
	assert.Equal(t, expectedErr, err)

	mockAKSClient.EXPECT().Get(
		gomock.Any(),
		aksPool.resourceGroup,
		aksPool.clusterName).Return(containerservice.ManagedCluster{}, errInternal)
	count, err = aksPool.GetNodeCount()
	assert.Equal(t, -1, count)
	assert.Equal(t, err, errInternalRaw)
}

func TestAKSTargetSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	aksPool := getTestAKSPool(newTestAzureManager(t), testAKSPoolName)
	mockAKSClient := mockcontainerserviceclient.NewMockInterface(ctrl)
	mockAKSClient.EXPECT().Get(gomock.Any(), aksPool.resourceGroup, aksPool.clusterName).Return(getExpectedManagedCluster(), nil)
	aksPool.manager.azClient.managedKubernetesServicesClient = mockAKSClient

	aksPool.lastRefresh = time.Now()
	size, err := aksPool.TargetSize()
	assert.Equal(t, 5, size)
	assert.NoError(t, err)

	aksPool.lastRefresh = time.Now().Add(-1 * 20 * time.Second)
	size, err = aksPool.TargetSize()
	assert.Equal(t, 1, size)
	assert.NoError(t, err)
}

func TestAKSSetSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		targetSize    int
		isScalingDown bool
		expectedSize  int
		expectedErr   error
	}{
		{
			targetSize:    0,
			isScalingDown: true,
			expectedSize:  5,
			expectedErr:   fmt.Errorf("size-decreasing request of 0 is smaller than min size 1"),
		},
		{
			targetSize:   3,
			expectedSize: 3,
		},
	}

	for _, test := range tests {
		aksPool := getTestAKSPool(newTestAzureManager(t), testAKSPoolName)
		mockAKSClient := mockcontainerserviceclient.NewMockInterface(ctrl)
		mockAKSClient.EXPECT().Get(
			gomock.Any(),
			aksPool.resourceGroup,
			aksPool.clusterName).Return(getExpectedManagedCluster(), nil).MaxTimes(1)
		mockAKSClient.EXPECT().CreateOrUpdate(
			gomock.Any(),
			aksPool.resourceGroup,
			aksPool.clusterName,
			gomock.Any(),
			gomock.Any()).Return(nil).MaxTimes(1)
		aksPool.manager.azClient.managedKubernetesServicesClient = mockAKSClient

		err := aksPool.SetSize(test.targetSize, test.isScalingDown)
		assert.Equal(t, test.expectedErr, err)
		assert.Equal(t, test.expectedSize, aksPool.curSize)
	}
}

func TestAKSIncreaseSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	aksPool := getTestAKSPool(newTestAzureManager(t), testAKSPoolName)
	aksPool.lastRefresh = time.Now()

	err := aksPool.IncreaseSize(-1)
	assert.Equal(t, fmt.Errorf("size increase must be +ve"), err)

	err = aksPool.IncreaseSize(6)
	assert.Equal(t, fmt.Errorf("size-increasing request of 11 is bigger than max size 10"), err)

	mockAKSClient := mockcontainerserviceclient.NewMockInterface(ctrl)
	mockAKSClient.EXPECT().Get(gomock.Any(), aksPool.resourceGroup, aksPool.clusterName).Return(getExpectedManagedCluster(), nil)
	mockAKSClient.EXPECT().CreateOrUpdate(gomock.Any(), aksPool.resourceGroup, aksPool.clusterName, gomock.Any(), gomock.Any()).Return(nil)
	aksPool.manager.azClient.managedKubernetesServicesClient = mockAKSClient

	err = aksPool.IncreaseSize(3)
	assert.NoError(t, err)
}

func TestIsAKSNode(t *testing.T) {
	aksPool := getTestAKSPool(newTestAzureManager(t), testAKSPoolName)
	tags := map[string]*string{"poolName": to.StringPtr(testAKSPoolName)}
	isAKSNode := aksPool.IsAKSNode(tags)
	assert.True(t, isAKSNode)

	tags = map[string]*string{"poolName": to.StringPtr("fake")}
	isAKSNode = aksPool.IsAKSNode(tags)
	assert.False(t, isAKSNode)
}

func TestDeleteNodesAKS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedVMs := []compute.VirtualMachine{
		{
			ID:   to.StringPtr("id1"),
			Name: to.StringPtr("vm1"),
			VirtualMachineProperties: &compute.VirtualMachineProperties{
				NetworkProfile: &compute.NetworkProfile{
					NetworkInterfaces: &[]compute.NetworkInterfaceReference{
						{ID: to.StringPtr("id")},
					},
				},
				StorageProfile: &compute.StorageProfile{
					OsDisk: &compute.OSDisk{
						ManagedDisk: &compute.ManagedDiskParameters{},
					},
				},
			},
		},
	}

	aksPool := getTestAKSPool(newTestAzureManager(t), testAKSPoolName)
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	mockVMClient.EXPECT().List(gomock.Any(), aksPool.nodeResourceGroup).Return(expectedVMs, nil)
	mockVMClient.EXPECT().Get(
		gomock.Any(),
		aksPool.nodeResourceGroup,
		*expectedVMs[0].Name,
		gomock.Any()).Return(expectedVMs[0], nil)
	mockVMClient.EXPECT().Delete(gomock.Any(), aksPool.nodeResourceGroup, *expectedVMs[0].Name)
	aksPool.manager.azClient.virtualMachinesClient = mockVMClient
	mockAKSClient := mockcontainerserviceclient.NewMockInterface(ctrl)
	mockAKSClient.EXPECT().Get(
		gomock.Any(),
		aksPool.resourceGroup,
		aksPool.clusterName).Return(getExpectedManagedCluster(), nil)
	mockAKSClient.EXPECT().CreateOrUpdate(
		gomock.Any(),
		aksPool.resourceGroup,
		aksPool.clusterName,
		gomock.Any(),
		gomock.Any())
	aksPool.manager.azClient.managedKubernetesServicesClient = mockAKSClient
	mockNICClient := mockinterfaceclient.NewMockInterface(ctrl)
	mockNICClient.EXPECT().Delete(
		gomock.Any(),
		aksPool.resourceGroup,
		"id").Return(nil)
	aksPool.manager.azClient.interfacesClient = mockNICClient

	nodes := []*apiv1.Node{
		{
			Spec: apiv1.NodeSpec{
				ProviderID: "id1",
			},
		},
	}
	err := aksPool.DeleteNodes(nodes)
	assert.Equal(t, 4, aksPool.curSize)
	assert.NoError(t, err)
}

func TestAKSNodes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedVMs := []compute.VirtualMachine{
		{
			Name: to.StringPtr("name"),
			ID:   to.StringPtr("/subscriptions/sub/resourceGroups/rg/providers/provider/vm1"),
			Tags: map[string]*string{"poolName": to.StringPtr(testAKSPoolName)},
		},
	}

	aksPool := getTestAKSPool(newTestAzureManager(t), testAKSPoolName)
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	mockVMClient.EXPECT().List(gomock.Any(), aksPool.nodeResourceGroup).Return(expectedVMs, nil)
	aksPool.manager.azClient.virtualMachinesClient = mockVMClient

	instances, err := aksPool.Nodes()
	assert.Equal(t, 1, len(instances))
	assert.NoError(t, err)

	mockVMClient.EXPECT().List(gomock.Any(), aksPool.nodeResourceGroup).Return([]compute.VirtualMachine{}, errInternal)
	instances, err = aksPool.Nodes()
	expectedErr := errInternalRaw
	assert.Nil(t, instances)
	assert.Equal(t, expectedErr, err)

	expectedVMs[0].ID = to.StringPtr("fakeID")
	mockVMClient.EXPECT().List(gomock.Any(), aksPool.nodeResourceGroup).Return(expectedVMs, nil)
	instances, err = aksPool.Nodes()
	assert.Equal(t, 0, len(instances))
	assert.NoError(t, err)
}

func TestAKSDecreaseTargetSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	aksPool := getTestAKSPool(newTestAzureManager(t), testAKSPoolName)

	err := aksPool.DecreaseTargetSize(1)
	expectedErr := fmt.Errorf("size decrease must be negative")
	assert.Equal(t, expectedErr, err)

	aksPool.lastRefresh = time.Now().Add(-1 * 20 * time.Second)

	mockAKSClient := mockcontainerserviceclient.NewMockInterface(ctrl)
	expectedMC := getExpectedManagedCluster()
	(*expectedMC.AgentPoolProfiles)[0].Count = to.Int32Ptr(2)
	mockAKSClient.EXPECT().Get(gomock.Any(), aksPool.resourceGroup, aksPool.clusterName).Return(expectedMC, nil)
	aksPool.manager.azClient.managedKubernetesServicesClient = mockAKSClient

	expectedVMs := []compute.VirtualMachine{
		{
			Name: to.StringPtr("name"),
			ID:   to.StringPtr("/subscriptions/sub/resourceGroups/rg/providers/provider/vm1"),
			Tags: map[string]*string{"poolName": to.StringPtr(testAKSPoolName)},
		},
	}
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	mockVMClient.EXPECT().List(gomock.Any(), aksPool.nodeResourceGroup).Return(expectedVMs, nil)
	aksPool.manager.azClient.virtualMachinesClient = mockVMClient

	mockAKSClient.EXPECT().Get(
		gomock.Any(),
		aksPool.resourceGroup,
		aksPool.clusterName).Return(getExpectedManagedCluster(), nil)
	mockAKSClient.EXPECT().CreateOrUpdate(
		gomock.Any(),
		aksPool.resourceGroup,
		aksPool.clusterName,
		gomock.Any(),
		gomock.Any()).Return(nil)
	aksPool.manager.azClient.managedKubernetesServicesClient = mockAKSClient

	err = aksPool.DecreaseTargetSize(-1)
	assert.Equal(t, 1, aksPool.curSize)
	assert.NoError(t, err)
}
