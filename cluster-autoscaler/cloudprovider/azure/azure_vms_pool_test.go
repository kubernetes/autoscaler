/*
Copyright 2017 The Kubernetes Authors.

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
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v4"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/golang/mock/gomock"
	ubergomock "go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmclient/mockvmclient"
)

func newTestVMsPool(manager *AzureManager, name string) *VMsPool {
	return &VMsPool{
		azureRef: azureRef{
			Name: name,
		},
		manager:              manager,
		minSize:              3,
		maxSize:              10,
		clusterName:          manager.config.ClusterName,
		resourceGroup:        manager.config.ResourceGroup,
		clusterResourceGroup: manager.config.ClusterResourceGroup,
	}
}

const (
	fakeVMsPoolVMID = "/subscriptions/test-subscription-id/resourceGroups/test-rg/providers/Microsoft.Compute/virtualMachines/%d"
)

func newTestVMsPoolVMList(count int) []compute.VirtualMachine {
	var vmList []compute.VirtualMachine
	for i := 0; i < count; i++ {
		vm := compute.VirtualMachine{
			ID: to.StringPtr(fmt.Sprintf(fakeVMsPoolVMID, i)),
			VirtualMachineProperties: &compute.VirtualMachineProperties{
				VMID: to.StringPtr(fmt.Sprintf("123E4567-E89B-12D3-A456-426655440000-%d", i)),
			},
			Tags: map[string]*string{
				agentpoolTypeTag: to.StringPtr("VirtualMachines"),
				agentpoolNameTag: to.StringPtr("test-vms-pool"),
			},
		}
		vmList = append(vmList, vm)
	}
	return vmList
}

func newVMsNode(vmID int64) *apiv1.Node {
	node := &apiv1.Node{
		Spec: apiv1.NodeSpec{
			ProviderID: "azure://" + fmt.Sprintf(fakeVMsPoolVMID, vmID),
		},
	}
	return node
}

func getTestAgentPool(agentpoolName string, isSystemPool bool) armcontainerservice.AgentPool {
	mode := armcontainerservice.AgentPoolModeUser
	if isSystemPool {
		mode = armcontainerservice.AgentPoolModeSystem
	}
	return armcontainerservice.AgentPool{
		Name: &agentpoolName,
		Properties: &armcontainerservice.ManagedClusterAgentPoolProfileProperties{
			Mode: &mode,
			VirtualMachinesProfile: &armcontainerservice.VirtualMachinesProfile{
				Scale: &armcontainerservice.ScaleProfile{
					Manual: []*armcontainerservice.ManualScaleProfile{
						{
							Count: to.Int32Ptr(3),
							Sizes: []*string{to.StringPtr("Standard_D2_v2")},
						},
					},
				},
			},
		},
	}
}

func TestNewVMsPool(t *testing.T) {
	ctrl := ubergomock.NewController(t)
	defer ctrl.Finish()
	mockAgentpoolclient := NewMockAgentPoolsClient(ctrl)
	manager := newTestAzureManager(t)
	manager.azClient.agentPoolClient = mockAgentpoolclient
	manager.config.ResourceGroup = "MC_rg"
	manager.config.ClusterResourceGroup = "rg"
	manager.config.ClusterName = "mycluster"

	agentpoolName := "pool1"
	spec := &dynamic.NodeGroupSpec{
		Name:    agentpoolName,
		MinSize: 1,
		MaxSize: 10,
	}

	ap, err := NewVMsPool(spec, manager)
	assert.NoError(t, err)
	assert.Equal(t, agentpoolName, ap.azureRef.Name)
	assert.Equal(t, "MC_rg", ap.resourceGroup)
	assert.Equal(t, "rg", ap.clusterResourceGroup)
	assert.Equal(t, "mycluster", ap.clusterName)
	assert.Equal(t, 1, ap.minSize)
	assert.Equal(t, 10, ap.maxSize)
}

func TestGetVMsFromCacheForVMsPool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	agentpoolName := "pool1"

	ap := newTestVMsPool(newTestAzureManager(t), agentpoolName)
	expectedVMs := []compute.VirtualMachine{
		{
			Name: to.StringPtr("aks-pool1-13222729-vms0"),
			Tags: map[string]*string{"aks-managed-poolName": &agentpoolName},
		},
		{
			Name: to.StringPtr("aks-pool1-13222729-vms1"),
			Tags: map[string]*string{"aks-managed-poolName": &agentpoolName},
		},
	}

	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	ap.manager.azClient.virtualMachinesClient = mockVMClient
	mockVMClient.EXPECT().List(gomock.Any(), ap.resourceGroup).Return(expectedVMs, nil)

	ac, err := newAzureCache(ap.manager.azClient, refreshInterval, ap.manager.config.ResourceGroup, "", false, "")
	assert.NoError(t, err)
	ap.manager.azureCache = ac

	vms, err := ap.getVMsFromCache()
	assert.Equal(t, 2, len(vms))
	assert.NoError(t, err)
}

func TestNodes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	agentpoolName := "pool1"
	vmssPoolName := "test-vmss-pool"

	ap := newTestVMsPool(newTestAzureManager(t), agentpoolName)
	expectedVMs := []compute.VirtualMachine{
		{
			ID:   to.StringPtr(fmt.Sprintf(fakeVMsPoolVMID, 0)),
			Name: to.StringPtr("aks-pool1-13222729-vms0"),
			Tags: map[string]*string{"aks-managed-poolName": &agentpoolName},
		},
		{
			ID:   to.StringPtr(fmt.Sprintf(fakeVMsPoolVMID, 1)),
			Name: to.StringPtr("aks-pool1-13222729-vms1"),
			Tags: map[string]*string{"aks-managed-poolName": &agentpoolName},
		},
		{
			ID:   to.StringPtr(fmt.Sprint("vmss-", 4)),
			Name: to.StringPtr("aks-vmssnp-38484957-vmss000004"),
			Tags: map[string]*string{"aks-managed-poolName": &vmssPoolName},
		},
	}

	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	ap.manager.azClient.virtualMachinesClient = mockVMClient
	mockVMClient.EXPECT().List(gomock.Any(), ap.resourceGroup).Return(expectedVMs, nil)

	ac, err := newAzureCache(ap.manager.azClient, refreshInterval, ap.manager.config.ResourceGroup, "", false, "")
	assert.NoError(t, err)
	ap.manager.azureCache = ac

	vms, err := ap.Nodes()
	assert.Equal(t, 2, len(vms))
	assert.NoError(t, err)
}

func TestGetCurSizeForVMsPool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	agentpoolName := "pool1"

	ap := newTestVMsPool(newTestAzureManager(t), agentpoolName)
	expectedVMs := []compute.VirtualMachine{
		{
			Name: to.StringPtr("aks-pool1-13222729-vms0"),
			Tags: map[string]*string{"aks-managed-poolName": &agentpoolName},
		},
		{
			Name: to.StringPtr("aks-pool1-13222729-vms1"),
			Tags: map[string]*string{"aks-managed-poolName": &agentpoolName},
		},
		{
			Name: to.StringPtr("aks-pool1-13222729-vms2"),
			Tags: map[string]*string{"aks-managed-poolName": &agentpoolName},
		},
	}

	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	ap.manager.azClient.virtualMachinesClient = mockVMClient
	mockVMClient.EXPECT().List(gomock.Any(), ap.resourceGroup).Return(expectedVMs, nil)

	ac, err := newAzureCache(ap.manager.azClient, refreshInterval, ap.manager.config.ResourceGroup, "", false, "")
	assert.NoError(t, err)
	ap.manager.azureCache = ac
	ap.curSize = -1 // not initialized

	ap.lastSizeRefresh = time.Now()
	curSize, err := ap.getCurSize()
	assert.NoError(t, err)
	assert.Equal(t, int64(-1), curSize)

	ap.lastSizeRefresh = time.Now().Add(-1 * 30 * time.Second)
	curSize, err = ap.getCurSize()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), curSize)
}

func TestGetVMsPoolSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	agentpoolName := "pool1"

	ap := newTestVMsPool(newTestAzureManager(t), agentpoolName)
	ap.curSize = -1 // not initialized
	ap.lastSizeRefresh = time.Now().Add(-1 * time.Second)

	_, err := ap.getVMsPoolSize()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "getVMsPoolSize: size is -1 for vms pool")
}

func TestVMsPoolIncreaseSize(t *testing.T) {
	ctrl := ubergomock.NewController(t)
	defer ctrl.Finish()
	mockAgentpoolclient := NewMockAgentPoolsClient(ctrl)
	manager := newTestAzureManager(t)
	manager.azClient.agentPoolClient = mockAgentpoolclient
	agentpoolName := "pool1"

	ap := newTestVMsPool(manager, agentpoolName)
	ap.curSize = 3
	ap.lastSizeRefresh = time.Now().Add(-1 * time.Second)

	// failure case 1
	err := ap.IncreaseSize(-1)
	expectedErr := fmt.Errorf("size increase must be positive, current delta: -1")
	assert.Equal(t, expectedErr, err)

	// failure case 2
	err = ap.IncreaseSize(8)
	expectedErr = fmt.Errorf("size-increasing request of 11 is bigger than max size 10")
	assert.Equal(t, expectedErr, err)

	// success case 3
	resp := &http.Response{
		Header: map[string][]string{
			"Fake-Poller-Status": {"Done"},
		},
	}

	fakePoller, err := runtime.NewPoller(resp, runtime.Pipeline{},
		&runtime.NewPollerOptions[armcontainerservice.AgentPoolsClientCreateOrUpdateResponse]{
			Handler: &fakehandler[armcontainerservice.AgentPoolsClientCreateOrUpdateResponse]{},
		})

	assert.NoError(t, err)

	agentpool := getTestAgentPool(agentpoolName, true)
	getResp := armcontainerservice.AgentPoolsClientGetResponse{
		AgentPool: agentpool,
	}

	mockAgentpoolclient.EXPECT().Get(gomock.Any(),
		manager.config.ClusterResourceGroup,
		manager.config.ClusterName,
		agentpoolName, nil).Return(getResp, nil)

	mockAgentpoolclient.EXPECT().BeginCreateOrUpdate(
		gomock.Any(), manager.config.ClusterResourceGroup,
		manager.config.ClusterName,
		agentpoolName,
		gomock.Any(), gomock.Any()).Return(fakePoller, nil)

	err = ap.IncreaseSize(1)
	assert.NoError(t, err)
}

func TestDeleteVMsPoolNodes(t *testing.T) {
	ctrl := ubergomock.NewController(t)
	defer ctrl.Finish()
	mockAgentpoolclient := NewMockAgentPoolsClient(ctrl)
	manager := newTestAzureManager(t)
	manager.azClient.agentPoolClient = mockAgentpoolclient
	agentpoolName := "pool1"
	nodeName := "aks-pool1-13222729-vms0"
	providerID := "azure:///subscriptions/feb5b150-60fe-4441-be73-8c02a524f55a/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/aks-pool1-13222729-vms0"

	node := &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: nodeName,
		},
		Spec: apiv1.NodeSpec{
			ProviderID: providerID,
		},
	}

	ap := newTestVMsPool(manager, agentpoolName)

	manager.azureCache = &azureCache{
		vmsPoolSet: map[string]struct{}{agentpoolName: {}},
		instanceToNodeGroup: map[azureRef]cloudprovider.NodeGroup{
			{Name: providerID}: ap,
		},
	}

	// failure case
	ap.curSize = 2
	ap.lastSizeRefresh = time.Now().Add(-1 * time.Second)
	err := ap.DeleteNodes([]*apiv1.Node{node})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "min size 3 reached, nodes will not be deleted")

	// success case
	ap.curSize = 4
	ap.lastSizeRefresh = time.Now().Add(-1 * time.Second)

	resp := &http.Response{
		Header: map[string][]string{
			"Fake-Poller-Status": {"Done"},
		},
	}
	fakePoller, err := runtime.NewPoller(resp, runtime.Pipeline{},
		&runtime.NewPollerOptions[armcontainerservice.AgentPoolsClientDeleteMachinesResponse]{
			Handler: &fakehandler[armcontainerservice.AgentPoolsClientDeleteMachinesResponse]{},
		})
	assert.NoError(t, err)

	mockAgentpoolclient.EXPECT().BeginDeleteMachines(
		gomock.Any(), manager.config.ClusterResourceGroup,
		manager.config.ClusterName,
		agentpoolName,
		gomock.Any(), gomock.Any()).Return(fakePoller, nil)

	derr := ap.DeleteNodes([]*apiv1.Node{node})
	assert.NoError(t, derr)
}

type fakehandler[T any] struct{}

func (f *fakehandler[T]) Done() bool {
	return true
}

func (f *fakehandler[T]) Poll(ctx context.Context) (*http.Response, error) {
	return nil, nil
}

func (f *fakehandler[T]) Result(ctx context.Context, out *T) error {
	return nil
}
