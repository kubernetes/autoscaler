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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice/v5"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
		manager:                   manager,
		minSize:                   3,
		maxSize:                   10,
		agentPoolName:             vmsAgentPoolName,
		sku:                       vmSku,
		clusterName:               manager.config.ClusterName,
		resourceGroup:             manager.config.ResourceGroup,
		clusterResourceGroup:      manager.config.ClusterResourceGroup,
		enableDynamicInstanceList: true,
		sizeRefreshPeriod:         30 * time.Second,
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
	assert.Equal(t, "MC_rg", ap.resourceGroup)
	assert.Equal(t, "rg", ap.clusterResourceGroup)
	assert.Equal(t, "mycluster", ap.clusterName)
	assert.Equal(t, 1, ap.minSize)
	assert.Equal(t, 10, ap.maxSize)
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
	mockVMClient.EXPECT().List(gomock.Any(), ap.resourceGroup).Return(expectedVMs, nil)

	agentpool := getTestVMsAgentPool(false)
	fakeAPListPager := getFakeAgentpoolListPager(&agentpool)
	mockAgentpoolclient.EXPECT().NewListPager(gomock.Any(), gomock.Any(), nil).
		Return(fakeAPListPager)

	ac, err := newAzureCache(ap.manager.azClient, refreshInterval, *ap.manager.config)
	assert.NoError(t, err)
	ac.enableVMsAgentPool = true
	ap.manager.azureCache = ac

	vms, err := ap.getVMsFromCache()
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
	mockVMClient.EXPECT().List(gomock.Any(), ap.resourceGroup).Return(expectedVMs, nil)

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

	ap := newTestVMsPool(newTestAzureManager(t))
	ap.curSize = -1 // not initialized
	ap.lastSizeRefresh = time.Now().Add(-1 * time.Second)

	curSize, err := ap.getVMPoolSize()
	assert.Equal(t, int64(-1), curSize)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("VMs agent pool %s not found in cache", vmsAgentPoolName))
}

func TestVMsPoolIncreaseSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	manager := newTestAzureManager(t)

	ap := newTestVMsPool(manager)
	ap.curSize = 3
	ap.lastSizeRefresh = time.Now().Add(-1 * time.Second)
	expectedVMs := newTestVMsPoolVMList(3)

	mockVMClient := mockvmclient.NewMockInterface(ctrl)
	ap.manager.azClient.virtualMachinesClient = mockVMClient
	mockVMClient.EXPECT().List(gomock.Any(), ap.resourceGroup).Return(expectedVMs, nil)

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

func TestDeleteVMsPoolNodes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAgentpoolclient := NewMockAgentPoolsClient(ctrl)
	manager := newTestAzureManager(t)
	manager.azClient.agentPoolClient = mockAgentpoolclient

	node := newVMsNode(0)
	providerID := node.Spec.ProviderID

	ap := newTestVMsPool(manager)
	manager.azureCache = &azureCache{
		vmsPoolMap: map[string]armcontainerservice.AgentPool{
			vmsAgentPoolName: getTestVMsAgentPool(false),
		},
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
		vmsAgentPoolName,
		gomock.Any(), gomock.Any()).Return(fakePoller, nil)

	derr := ap.DeleteNodes([]*apiv1.Node{node})
	assert.NoError(t, derr)
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

func TestTargetSize(t *testing.T) {
	manager := newTestAzureManager(t)
	node := newVMsNode(0)
	providerID := node.Spec.ProviderID

	agentPool := newTestVMsPool(manager)
	manager.azureCache = &azureCache{
		vmsPoolMap: map[string]armcontainerservice.AgentPool{
			vmsAgentPoolName: getTestVMsAgentPool(false),
		},
		instanceToNodeGroup: map[azureRef]cloudprovider.NodeGroup{
			{Name: providerID}: agentPool,
		},
	}

	size, err := agentPool.TargetSize()
	assert.Equal(t, 3, size)
	assert.NoError(t, err)
}

func TestDecreaseTargetSize(t *testing.T) {
	manager := newTestAzureManager(t)
	node := newVMsNode(0)
	providerID := node.Spec.ProviderID

	agentPool := newTestVMsPool(manager)
	manager.azureCache = &azureCache{
		vmsPoolMap: map[string]armcontainerservice.AgentPool{
			vmsAgentPoolName: getTestVMsAgentPool(false),
		},
		instanceToNodeGroup: map[azureRef]cloudprovider.NodeGroup{
			{Name: providerID}: agentPool,
		},
	}

	err := agentPool.DecreaseTargetSize(1)
	assert.NoError(t, err)
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
	manager := newTestAzureManager(t)
	node := newVMsNode(0)
	providerID := node.Spec.ProviderID

	agentPool := newTestVMsPool(manager)
	assert.True(t, agentPool.enableDynamicInstanceList) // default is true

	// The dynamic SKU list ("cache") in the test manager is empty
	manager.azureCache = &azureCache{
		vmsPoolMap: map[string]armcontainerservice.AgentPool{
			vmsAgentPoolName: getTestVMsAgentPool(false),
		},
		instanceToNodeGroup: map[azureRef]cloudprovider.NodeGroup{
			{Name: providerID}: agentPool,
		},
	}
	assert.False(t, manager.azureCache.HasVMSKUs())

	t.Run("Checking fallback to static because dynamic list is empty", func(t *testing.T) {
		nodeInfo, err := agentPool.TemplateNodeInfo()
		assert.NoError(t, err)
		assert.NotNil(t, nodeInfo)
		assert.NotEmpty(t, nodeInfo.Pods)
	})

	t.Run("Checking dynamic workflow", func(t *testing.T) {
		GetInstanceTypeDynamically = func(template NodeTemplate, azCache *azureCache) (InstanceType, error) {
			instanceType := InstanceType{}
			instanceType.VCPU = 1
			instanceType.GPU = 2
			instanceType.MemoryMb = 3
			return instanceType, nil
		}
		nodeInfo, err := agentPool.TemplateNodeInfo()
		assert.Equal(t, *nodeInfo.Node().Status.Capacity.Cpu(), *resource.NewQuantity(1, resource.DecimalSI))
		assert.Equal(t, *nodeInfo.Node().Status.Capacity.Memory(), *resource.NewQuantity(3*1024*1024, resource.DecimalSI))
		assert.NoError(t, err)
		assert.NotNil(t, nodeInfo)
		assert.NotEmpty(t, nodeInfo.Pods)
	})

	t.Run("Checking static workflow if dynamic fails", func(t *testing.T) {
		GetInstanceTypeDynamically = func(template NodeTemplate, azCache *azureCache) (InstanceType, error) {
			return InstanceType{}, fmt.Errorf("dynamic error exists")
		}
		GetInstanceTypeStatically = func(template NodeTemplate) (*InstanceType, error) {
			instanceType := InstanceType{}
			instanceType.VCPU = 1
			instanceType.GPU = 2
			instanceType.MemoryMb = 3
			return &instanceType, nil
		}
		nodeInfo, err := agentPool.TemplateNodeInfo()
		assert.Equal(t, *nodeInfo.Node().Status.Capacity.Cpu(), *resource.NewQuantity(1, resource.DecimalSI))
		assert.Equal(t, *nodeInfo.Node().Status.Capacity.Memory(), *resource.NewQuantity(3*1024*1024, resource.DecimalSI))
		assert.NoError(t, err)
		assert.NotNil(t, nodeInfo)
		assert.NotEmpty(t, nodeInfo.Pods)
	})

	t.Run("Fails to find vm instance information using static and dynamic workflow, instance not supported", func(t *testing.T) {
		GetInstanceTypeDynamically = func(template NodeTemplate, azCache *azureCache) (InstanceType, error) {
			return InstanceType{}, fmt.Errorf("dynamic error exists")
		}
		GetInstanceTypeStatically = func(template NodeTemplate) (*InstanceType, error) {
			return &InstanceType{}, fmt.Errorf("static error exists")
		}
		nodeInfo, err := agentPool.TemplateNodeInfo()
		assert.Empty(t, nodeInfo)
		assert.Equal(t, err, fmt.Errorf("static error exists"))
	})

	t.Run("Checking static-only workflow", func(t *testing.T) {
		agentPool.enableDynamicInstanceList = false

		GetInstanceTypeStatically = func(template NodeTemplate) (*InstanceType, error) {
			instanceType := InstanceType{}
			instanceType.VCPU = 1
			instanceType.GPU = 2
			instanceType.MemoryMb = 3
			return &instanceType, nil
		}
		nodeInfo, err := agentPool.TemplateNodeInfo()
		assert.Equal(t, *nodeInfo.Node().Status.Capacity.Cpu(), *resource.NewQuantity(1, resource.DecimalSI))
		assert.Equal(t, *nodeInfo.Node().Status.Capacity.Memory(), *resource.NewQuantity(3*1024*1024, resource.DecimalSI))
		assert.NoError(t, err)
		assert.NotNil(t, nodeInfo)
		assert.NotEmpty(t, nodeInfo.Pods)
	})

	t.Run("Checking static-only workflow with built-in SKU list", func(t *testing.T) {
		agentPool.enableDynamicInstanceList = false

		nodeInfo, err := agentPool.TemplateNodeInfo()
		assert.NoError(t, err)
		assert.NotNil(t, nodeInfo)
		assert.NotEmpty(t, nodeInfo.Pods)
	})
}

func TestAtomicIncreaseSize(t *testing.T) {
	agentPool := &VMPool{}

	err := agentPool.AtomicIncreaseSize(1)
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

// Test cases for getVMsFromCache()
// Test case 1 - when the vms pool is not found in the cache
// Test case 2 - when the vms pool is found in the cache but has no VMs
// Test case 3 - when the vms pool is found in the cache and has VMs
// Test case 4 - when the vms pool is found in the cache and has VMs with no name
func TestGetVMsFromCache(t *testing.T) {
	manager := newTestAzureManager(t)
	agentPool := newTestVMsPool(manager)

	// Test case 1 - vms agentpool not found in cache
	_, err := agentPool.getVMsFromCache()
	assert.EqualError(t, err, "VMs agent pool test-vms-pool not found in cache")

	// add vms agent pool to vmsPoolMap cache so it will be found
	manager.azureCache.vmsPoolMap = map[string]armcontainerservice.AgentPool{
		vmsAgentPoolName: getTestVMsAgentPool(false),
	}

	// Test case 2 - has no VMs
	manager.azureCache.virtualMachines[vmsAgentPoolName] = []compute.VirtualMachine{}
	vms, err := agentPool.getVMsFromCache()
	assert.NoError(t, err)
	assert.Len(t, vms, 0)

	// Test case 3 - has 3 VMs
	manager.azureCache.virtualMachines[vmsAgentPoolName] = newTestVMsPoolVMList(3)
	vms, err = agentPool.getVMsFromCache()
	assert.NoError(t, err)
	assert.Len(t, vms, 3)

	// Test case 4 - agent pool has no name
	manager.azureCache.virtualMachines[vmsAgentPoolName] = newTestVMsPoolVMList(3)
	agentPool.agentPoolName = ""
	_, err = agentPool.getVMsFromCache()
	assert.EqualError(t, err, "VMs agent pool  not found in cache")
}

// Test cases for Nodes()
// Test case 1 - when there are no VMs in the pool
// Test case 2 - when there are VMs in the pool
// Test case 3 - when there are VMs in the pool with no ID
// Test case 4 - when there is an error converting resource group name
// Test case 5 - when there is an error getting VMs from cache
func TestNodes(t *testing.T) {
	// Test case 1
	manager := newTestAzureManager(t)
	agentPool := newTestVMsPool(manager)

	nodes, err := agentPool.Nodes()
	assert.EqualError(t, err, "VMs agent pool test-vms-pool not found in cache")
	assert.Empty(t, nodes)

	// add vms agent pool to vmsPoolMap cache so it will be found
	manager.azureCache.vmsPoolMap = map[string]armcontainerservice.AgentPool{
		vmsAgentPoolName: getTestVMsAgentPool(false),
	}

	// Test case 2
	manager.azureCache.virtualMachines[vmsAgentPoolName] = newTestVMsPoolVMList(3)
	nodes, err = agentPool.Nodes()
	assert.NoError(t, err)
	assert.Len(t, nodes, 3)

	// Test case 3
	manager.azureCache.virtualMachines[vmsAgentPoolName] = newTestVMsPoolVMList(3)
	manager.azureCache.virtualMachines[vmsAgentPoolName][0].ID = nil
	nodes, err = agentPool.Nodes()
	assert.NoError(t, err)
	assert.Len(t, nodes, 2)
	manager.azureCache.virtualMachines[vmsAgentPoolName] = newTestVMsPoolVMList(3)
	emptyString := ""
	manager.azureCache.virtualMachines[vmsAgentPoolName][0].ID = &emptyString
	nodes, err = agentPool.Nodes()
	assert.NoError(t, err)
	assert.Len(t, nodes, 2)

	// Test case 4
	manager.azureCache.virtualMachines[vmsAgentPoolName] = newTestVMsPoolVMList(3)
	bogusID := "foo"
	manager.azureCache.virtualMachines[vmsAgentPoolName][0].ID = &bogusID
	nodes, err = agentPool.Nodes()
	assert.Empty(t, nodes)
	assert.Error(t, err)

	// Test case 5
	manager.azureCache.virtualMachines[vmsAgentPoolName] = newTestVMsPoolVMList(1)
	agentPool.agentPoolName = ""
	nodes, err = agentPool.Nodes()
	assert.Empty(t, nodes)
	assert.Error(t, err)
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
