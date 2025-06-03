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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v5"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"

	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"sigs.k8s.io/cloud-provider-azure/pkg/azureclients/vmclient/mockvmclient"
)

const (
	vmSku            = "Standard_D2_v2"
	vmsAgentPoolName = "test-vms-pool"
	vmsNodeGroupName = vmsAgentPoolName + "/" + vmSku
	fakeVMsNodeName  = "aks-" + vmsAgentPoolName + "-13222729-vms%d"
	fakeVMsPoolVMID  = "/subscriptions/test-subscription-id/resourceGroups/test-rg/providers/Microsoft.Compute/virtualMachines/" + fakeVMsNodeName
)

func newTestVMsPool(manager *AzureManager) *VMPool {
	return &VMPool{
		azureRef: azureRef{
			Name: vmsNodeGroupName,
		},
		manager:       manager,
		minSize:       3,
		maxSize:       10,
		agentPoolName: vmsAgentPoolName,
		sku:           vmSku,
	}
}

func newTestVMsPoolVMList(count int) []compute.VirtualMachine {
	var vmList []compute.VirtualMachine

	for i := 0; i < count; i++ {
		vm := compute.VirtualMachine{
			ID: to.StringPtr(fmt.Sprintf(fakeVMsPoolVMID, i)),
			VirtualMachineProperties: &compute.VirtualMachineProperties{
				VMID: to.StringPtr(fmt.Sprintf("123E4567-E89B-12D3-A456-426655440000-%d", i)),
				HardwareProfile: &compute.HardwareProfile{
					VMSize: compute.VirtualMachineSizeTypes(vmSku),
				},
				ProvisioningState: to.StringPtr("Succeeded"),
			},
			Tags: map[string]*string{
				agentpoolTypeTag: to.StringPtr("VirtualMachines"),
				agentpoolNameTag: to.StringPtr(vmsAgentPoolName),
			},
		}
		vmList = append(vmList, vm)
	}
	return vmList
}

func newVMsNode(vmIdx int64) *apiv1.Node {
	return &apiv1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf(fakeVMsNodeName, vmIdx),
		},
		Spec: apiv1.NodeSpec{
			ProviderID: "azure://" + fmt.Sprintf(fakeVMsPoolVMID, vmIdx),
		},
	}
}

func getTestVMsAgentPool(isSystemPool bool) armcontainerservice.AgentPool {
	mode := armcontainerservice.AgentPoolModeUser
	if isSystemPool {
		mode = armcontainerservice.AgentPoolModeSystem
	}
	vmsPoolType := armcontainerservice.AgentPoolTypeVirtualMachines
	return armcontainerservice.AgentPool{
		Name: to.StringPtr(vmsAgentPoolName),
		Properties: &armcontainerservice.ManagedClusterAgentPoolProfileProperties{
			Type: &vmsPoolType,
			Mode: &mode,
			VirtualMachinesProfile: &armcontainerservice.VirtualMachinesProfile{
				Scale: &armcontainerservice.ScaleProfile{
					Manual: []*armcontainerservice.ManualScaleProfile{
						{
							Count: to.Int32Ptr(3),
							Sizes: []*string{to.StringPtr(vmSku)},
						},
					},
				},
			},
			VirtualMachineNodesStatus: []*armcontainerservice.VirtualMachineNodes{
				{
					Count: to.Int32Ptr(3),
					Size:  to.StringPtr(vmSku),
				},
			},
		},
	}
}

func TestNewVMsPool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAgentpoolclient := NewMockAgentPoolsClient(ctrl)
	manager := newTestAzureManager(t)
	manager.azClient.agentPoolClient = mockAgentpoolclient
	manager.config.ResourceGroup = "MC_rg"
	manager.config.ClusterResourceGroup = "rg"
	manager.config.ClusterName = "mycluster"

	spec := &dynamic.NodeGroupSpec{
		Name:    vmsAgentPoolName,
		MinSize: 1,
		MaxSize: 10,
	}

	ap, err := NewVMPool(spec, manager, vmsAgentPoolName, vmSku)
	assert.NoError(t, err)
	assert.Equal(t, vmsAgentPoolName, ap.azureRef.Name)
	assert.Equal(t, 1, ap.minSize)
	assert.Equal(t, 10, ap.maxSize)
}

func TestMinSize(t *testing.T) {
	agentPool := &VMPool{
		minSize: 1,
	}

	assert.Equal(t, 1, agentPool.MinSize())
}

func TestExist(t *testing.T) {
	agentPool := &VMPool{}

	assert.True(t, agentPool.Exist())
}
func TestCreate(t *testing.T) {
	agentPool := &VMPool{}

	nodeGroup, err := agentPool.Create()
	assert.Nil(t, nodeGroup)
	assert.Equal(t, cloudprovider.ErrAlreadyExist, err)
}

func TestDelete(t *testing.T) {
	agentPool := &VMPool{}

	err := agentPool.Delete()
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

func TestAutoprovisioned(t *testing.T) {
	agentPool := &VMPool{}

	assert.False(t, agentPool.Autoprovisioned())
}

func TestGetOptions(t *testing.T) {
	agentPool := &VMPool{}
	defaults := config.NodeGroupAutoscalingOptions{}

	options, err := agentPool.GetOptions(defaults)
	assert.Nil(t, options)
	assert.Nil(t, err)
}
func TestMaxSize(t *testing.T) {
	agentPool := &VMPool{
		maxSize: 10,
	}

	assert.Equal(t, 10, agentPool.MaxSize())
}

func TestDecreaseTargetSize(t *testing.T) {
	agentPool := newTestVMsPool(newTestAzureManager(t))

	err := agentPool.DecreaseTargetSize(1)
	assert.Nil(t, err)
}

func TestId(t *testing.T) {
	agentPool := &VMPool{
		azureRef: azureRef{
			Name: "test-id",
		},
	}

	assert.Equal(t, "test-id", agentPool.Id())
}

func TestDebug(t *testing.T) {
	agentPool := &VMPool{
		azureRef: azureRef{
			Name: "test-debug",
		},
		minSize: 1,
		maxSize: 5,
	}

	expectedDebugString := "test-debug (1:5)"
	assert.Equal(t, expectedDebugString, agentPool.Debug())
}
func TestTemplateNodeInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ap := newTestVMsPool(newTestAzureManager(t))
	ap.manager.config.EnableVMsAgentPool = true
	mockAgentpoolclient := NewMockAgentPoolsClient(ctrl)
	ap.manager.azClient.agentPoolClient = mockAgentpoolclient
	agentpool := getTestVMsAgentPool(false)
	fakeAPListPager := getFakeAgentpoolListPager(&agentpool)
	mockAgentpoolclient.EXPECT().NewListPager(gomock.Any(), gomock.Any(), nil).
		Return(fakeAPListPager)

	ac, err := newAzureCache(ap.manager.azClient, refreshInterval, *ap.manager.config)
	assert.NoError(t, err)
	ap.manager.azureCache = ac

	nodeInfo, err := ap.TemplateNodeInfo()
	assert.NotNil(t, nodeInfo)
	assert.Nil(t, err)
}

func TestAtomicIncreaseSize(t *testing.T) {
	agentPool := &VMPool{}

	err := agentPool.AtomicIncreaseSize(1)
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

func TestGetVMsFromCache(t *testing.T) {
	manager := &AzureManager{
		azureCache: &azureCache{
			virtualMachines: make(map[string][]compute.VirtualMachine),
			vmsPoolMap:      make(map[string]armcontainerservice.AgentPool),
		},
	}
	agentPool := &VMPool{
		manager:       manager,
		agentPoolName: vmsAgentPoolName,
		sku:           vmSku,
	}

	// Test case 1 - when the vms pool is not found in the cache
	vms, err := agentPool.getVMsFromCache(skipOption{})
	assert.Nil(t, err)
	assert.Len(t, vms, 0)

	// Test case 2 - when the vms pool is found in the cache but has no VMs
	manager.azureCache.virtualMachines[vmsAgentPoolName] = []compute.VirtualMachine{}
	vms, err = agentPool.getVMsFromCache(skipOption{})
	assert.NoError(t, err)
	assert.Len(t, vms, 0)

	// Test case 3 - when the vms pool is found in the cache and has VMs
	manager.azureCache.virtualMachines[vmsAgentPoolName] = newTestVMsPoolVMList(3)
	vms, err = agentPool.getVMsFromCache(skipOption{})
	assert.NoError(t, err)
	assert.Len(t, vms, 3)

	// Test case 4 - should skip failed VMs
	vmList := newTestVMsPoolVMList(3)
	vmList[0].VirtualMachineProperties.ProvisioningState = to.StringPtr("Failed")
	manager.azureCache.virtualMachines[vmsAgentPoolName] = vmList
	vms, err = agentPool.getVMsFromCache(skipOption{skipFailed: true})
	assert.NoError(t, err)
	assert.Len(t, vms, 2)

	// Test case 5 - should skip deleting VMs
	vmList = newTestVMsPoolVMList(3)
	vmList[0].VirtualMachineProperties.ProvisioningState = to.StringPtr("Deleting")
	manager.azureCache.virtualMachines[vmsAgentPoolName] = vmList
	vms, err = agentPool.getVMsFromCache(skipOption{skipDeleting: true})
	assert.NoError(t, err)
	assert.Len(t, vms, 2)

	// Test case 6 - should not skip deleting VMs
	vmList = newTestVMsPoolVMList(3)
	vmList[0].VirtualMachineProperties.ProvisioningState = to.StringPtr("Deleting")
	manager.azureCache.virtualMachines[vmsAgentPoolName] = vmList
	vms, err = agentPool.getVMsFromCache(skipOption{skipFailed: true})
	assert.NoError(t, err)
	assert.Len(t, vms, 3)

	// Test case 7 - when the vms pool is found in the cache and has VMs with no name
	manager.azureCache.virtualMachines[vmsAgentPoolName] = newTestVMsPoolVMList(3)
	agentPool.agentPoolName = ""
	vms, err = agentPool.getVMsFromCache(skipOption{})
	assert.NoError(t, err)
	assert.Len(t, vms, 0)
}

func TestGetVMsFromCacheForVMsPool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ap := newTestVMsPool(newTestAzureManager(t))

	expectedVMs := newTestVMsPoolVMList(2)
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	ap.manager.azClient.virtualMachinesClient = mockVMClient
	ap.manager.config.EnableVMsAgentPool = true
	mockAgentpoolclient := NewMockAgentPoolsClient(ctrl)
	ap.manager.azClient.agentPoolClient = mockAgentpoolclient
	mockVMClient.EXPECT().List(gomock.Any(), ap.manager.config.ResourceGroup).Return(expectedVMs, nil)

	agentpool := getTestVMsAgentPool(false)
	fakeAPListPager := getFakeAgentpoolListPager(&agentpool)
	mockAgentpoolclient.EXPECT().NewListPager(gomock.Any(), gomock.Any(), nil).
		Return(fakeAPListPager)

	ac, err := newAzureCache(ap.manager.azClient, refreshInterval, *ap.manager.config)
	assert.NoError(t, err)
	ac.enableVMsAgentPool = true
	ap.manager.azureCache = ac

	vms, err := ap.getVMsFromCache(skipOption{})
	assert.Equal(t, 2, len(vms))
	assert.NoError(t, err)
}

func TestNodes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ap := newTestVMsPool(newTestAzureManager(t))
	expectedVMs := newTestVMsPoolVMList(2)

	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	ap.manager.azClient.virtualMachinesClient = mockVMClient
	mockVMClient.EXPECT().List(gomock.Any(), ap.manager.config.ResourceGroup).Return(expectedVMs, nil)

	ap.manager.config.EnableVMsAgentPool = true
	mockAgentpoolclient := NewMockAgentPoolsClient(ctrl)
	ap.manager.azClient.agentPoolClient = mockAgentpoolclient
	agentpool := getTestVMsAgentPool(false)
	fakeAPListPager := getFakeAgentpoolListPager(&agentpool)
	mockAgentpoolclient.EXPECT().NewListPager(gomock.Any(), gomock.Any(), nil).
		Return(fakeAPListPager)

	ac, err := newAzureCache(ap.manager.azClient, refreshInterval, *ap.manager.config)
	assert.NoError(t, err)
	ap.manager.azureCache = ac

	vms, err := ap.Nodes()
	assert.Equal(t, 2, len(vms))
	assert.NoError(t, err)
}

func TestGetCurSizeForVMsPool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ap := newTestVMsPool(newTestAzureManager(t))
	expectedVMs := newTestVMsPoolVMList(3)

	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	ap.manager.azClient.virtualMachinesClient = mockVMClient
	mockVMClient.EXPECT().List(gomock.Any(), ap.manager.config.ResourceGroup).Return(expectedVMs, nil)

	ap.manager.config.EnableVMsAgentPool = true
	mockAgentpoolclient := NewMockAgentPoolsClient(ctrl)
	ap.manager.azClient.agentPoolClient = mockAgentpoolclient
	agentpool := getTestVMsAgentPool(false)
	fakeAPListPager := getFakeAgentpoolListPager(&agentpool)
	mockAgentpoolclient.EXPECT().NewListPager(gomock.Any(), gomock.Any(), nil).
		Return(fakeAPListPager)

	ac, err := newAzureCache(ap.manager.azClient, refreshInterval, *ap.manager.config)
	assert.NoError(t, err)
	ap.manager.azureCache = ac

	curSize, err := ap.getCurSize(skipOption{})
	assert.NoError(t, err)
	assert.Equal(t, int32(3), curSize)
}

func TestVMsPoolIncreaseSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := newTestAzureManager(t)

	ap := newTestVMsPool(manager)
	expectedVMs := newTestVMsPoolVMList(3)

	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	ap.manager.azClient.virtualMachinesClient = mockVMClient
	mockVMClient.EXPECT().List(gomock.Any(), ap.manager.config.ResourceGroup).Return(expectedVMs, nil)

	ap.manager.config.EnableVMsAgentPool = true
	mockAgentpoolclient := NewMockAgentPoolsClient(ctrl)
	ap.manager.azClient.agentPoolClient = mockAgentpoolclient
	agentpool := getTestVMsAgentPool(false)
	fakeAPListPager := getFakeAgentpoolListPager(&agentpool)
	mockAgentpoolclient.EXPECT().NewListPager(gomock.Any(), gomock.Any(), nil).
		Return(fakeAPListPager)

	ac, err := newAzureCache(ap.manager.azClient, refreshInterval, *ap.manager.config)
	assert.NoError(t, err)
	ap.manager.azureCache = ac

	// failure case 1
	err1 := ap.IncreaseSize(-1)
	expectedErr := fmt.Errorf("size increase must be positive, current delta: -1")
	assert.Equal(t, expectedErr, err1)

	// failure case 2
	err2 := ap.IncreaseSize(8)
	expectedErr = fmt.Errorf("size-increasing request of 11 is bigger than max size 10")
	assert.Equal(t, expectedErr, err2)

	// success case 3
	resp := &http.Response{
		Header: map[string][]string{
			"Fake-Poller-Status": {"Done"},
		},
	}

	fakePoller, pollerErr := runtime.NewPoller(resp, runtime.Pipeline{},
		&runtime.NewPollerOptions[armcontainerservice.AgentPoolsClientCreateOrUpdateResponse]{
			Handler: &fakehandler[armcontainerservice.AgentPoolsClientCreateOrUpdateResponse]{},
		})

	assert.NoError(t, pollerErr)

	mockAgentpoolclient.EXPECT().BeginCreateOrUpdate(
		gomock.Any(), manager.config.ClusterResourceGroup,
		manager.config.ClusterName,
		vmsAgentPoolName,
		gomock.Any(), gomock.Any()).Return(fakePoller, nil)

	err3 := ap.IncreaseSize(1)
	assert.NoError(t, err3)
}

func TestDeleteVMsPoolNodes_Failed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ap := newTestVMsPool(newTestAzureManager(t))
	node := newVMsNode(0)

	expectedVMs := newTestVMsPoolVMList(3)
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	ap.manager.azClient.virtualMachinesClient = mockVMClient
	ap.manager.config.EnableVMsAgentPool = true
	mockAgentpoolclient := NewMockAgentPoolsClient(ctrl)
	agentpool := getTestVMsAgentPool(false)
	ap.manager.azClient.agentPoolClient = mockAgentpoolclient
	fakeAPListPager := getFakeAgentpoolListPager(&agentpool)
	mockAgentpoolclient.EXPECT().NewListPager(gomock.Any(), gomock.Any(), nil).Return(fakeAPListPager)
	mockVMClient.EXPECT().List(gomock.Any(), ap.manager.config.ResourceGroup).Return(expectedVMs, nil)

	ap.manager.azureCache.enableVMsAgentPool = true
	registered := ap.manager.RegisterNodeGroup(ap)
	assert.True(t, registered)

	ap.manager.explicitlyConfigured[vmsNodeGroupName] = true
	ap.manager.forceRefresh()

	// failure case
	deleteErr := ap.DeleteNodes([]*apiv1.Node{node})
	assert.Error(t, deleteErr)
	assert.Contains(t, deleteErr.Error(), "cannot delete nodes as minimum size of 3 has been reached")
}

func TestDeleteVMsPoolNodes_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ap := newTestVMsPool(newTestAzureManager(t))

	expectedVMs := newTestVMsPoolVMList(5)
	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	ap.manager.azClient.virtualMachinesClient = mockVMClient
	ap.manager.config.EnableVMsAgentPool = true
	mockAgentpoolclient := NewMockAgentPoolsClient(ctrl)
	agentpool := getTestVMsAgentPool(false)
	ap.manager.azClient.agentPoolClient = mockAgentpoolclient
	fakeAPListPager := getFakeAgentpoolListPager(&agentpool)
	mockAgentpoolclient.EXPECT().NewListPager(gomock.Any(), gomock.Any(), nil).Return(fakeAPListPager)
	mockVMClient.EXPECT().List(gomock.Any(), ap.manager.config.ResourceGroup).Return(expectedVMs, nil)

	ap.manager.azureCache.enableVMsAgentPool = true
	registered := ap.manager.RegisterNodeGroup(ap)
	assert.True(t, registered)

	ap.manager.explicitlyConfigured[vmsNodeGroupName] = true
	ap.manager.forceRefresh()

	// success case
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
		gomock.Any(), ap.manager.config.ClusterResourceGroup,
		ap.manager.config.ClusterName,
		vmsAgentPoolName,
		gomock.Any(), gomock.Any()).Return(fakePoller, nil)
	node := newVMsNode(0)
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

func getFakeAgentpoolListPager(agentpool ...*armcontainerservice.AgentPool) *runtime.Pager[armcontainerservice.AgentPoolsClientListResponse] {
	fakeFetcher := func(ctx context.Context, response *armcontainerservice.AgentPoolsClientListResponse) (armcontainerservice.AgentPoolsClientListResponse, error) {
		return armcontainerservice.AgentPoolsClientListResponse{
			AgentPoolListResult: armcontainerservice.AgentPoolListResult{
				Value: agentpool,
			},
		}, nil
	}

	return runtime.NewPager(runtime.PagingHandler[armcontainerservice.AgentPoolsClientListResponse]{
		More: func(response armcontainerservice.AgentPoolsClientListResponse) bool {
			return false
		},
		Fetcher: fakeFetcher,
	})
}
