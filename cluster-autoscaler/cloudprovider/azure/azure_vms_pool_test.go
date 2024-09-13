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
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/stretchr/testify/assert"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	providerazure "sigs.k8s.io/cloud-provider-azure/pkg/provider"
)

func newTestVMsPool(manager *AzureManager, name string) *VMsPool {
	return &VMsPool{
		azureRef: azureRef{
			Name: name,
		},
		manager: manager,
		minSize: 3,
		maxSize: 10,
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

func TestNewVMsPool(t *testing.T) {
	spec := &dynamic.NodeGroupSpec{
		Name:    "test-nodepool",
		MinSize: 1,
		MaxSize: 5,
	}
	am := &AzureManager{
		config: &Config{
			Config: providerazure.Config{
				ResourceGroup: "test-resource-group",
			},
		},
	}

	nodepool := NewVMsPool(spec, am)

	assert.Equal(t, "test-nodepool", nodepool.azureRef.Name)
	assert.Equal(t, "test-resource-group", nodepool.resourceGroup)
	assert.Equal(t, int64(-1), nodepool.curSize)
	assert.Equal(t, 1, nodepool.minSize)
	assert.Equal(t, 5, nodepool.maxSize)
	assert.Equal(t, am, nodepool.manager)
}

func TestMinSize(t *testing.T) {
	agentPool := &VMsPool{
		minSize: 1,
	}

	assert.Equal(t, 1, agentPool.MinSize())
}

func TestExist(t *testing.T) {
	agentPool := &VMsPool{}

	assert.True(t, agentPool.Exist())
}
func TestCreate(t *testing.T) {
	agentPool := &VMsPool{}

	nodeGroup, err := agentPool.Create()
	assert.Nil(t, nodeGroup)
	assert.Equal(t, cloudprovider.ErrAlreadyExist, err)
}

func TestDelete(t *testing.T) {
	agentPool := &VMsPool{}

	err := agentPool.Delete()
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

func TestAutoprovisioned(t *testing.T) {
	agentPool := &VMsPool{}

	assert.False(t, agentPool.Autoprovisioned())
}

func TestGetOptions(t *testing.T) {
	agentPool := &VMsPool{}
	defaults := config.NodeGroupAutoscalingOptions{}

	options, err := agentPool.GetOptions(defaults)
	assert.Nil(t, options)
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}
func TestMaxSize(t *testing.T) {
	agentPool := &VMsPool{
		maxSize: 10,
	}

	assert.Equal(t, 10, agentPool.MaxSize())
}

func TestTargetSize(t *testing.T) {
	agentPool := &VMsPool{}

	size, err := agentPool.TargetSize()
	assert.Equal(t, -1, size)
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

func TestIncreaseSize(t *testing.T) {
	agentPool := &VMsPool{}

	err := agentPool.IncreaseSize(1)
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

func TestDeleteNodes(t *testing.T) {
	agentPool := &VMsPool{}

	err := agentPool.DeleteNodes(nil)
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

func TestDecreaseTargetSize(t *testing.T) {
	agentPool := &VMsPool{}

	err := agentPool.DecreaseTargetSize(1)
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

func TestId(t *testing.T) {
	agentPool := &VMsPool{
		azureRef: azureRef{
			Name: "test-id",
		},
	}

	assert.Equal(t, "test-id", agentPool.Id())
}

func TestDebug(t *testing.T) {
	agentPool := &VMsPool{
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
	agentPool := &VMsPool{}

	nodeInfo, err := agentPool.TemplateNodeInfo()
	assert.Nil(t, nodeInfo)
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

func TestAtomicIncreaseSize(t *testing.T) {
	agentPool := &VMsPool{}

	err := agentPool.AtomicIncreaseSize(1)
	assert.Equal(t, cloudprovider.ErrNotImplemented, err)
}

// Test cases for getVMsFromCache()
// Test case 1 - when the vms pool is not found in the cache
// Test case 2 - when the vms pool is found in the cache but has no VMs
// Test case 3 - when the vms pool is found in the cache and has VMs
// Test case 4 - when the vms pool is found in the cache and has VMs with no name
func TestGetVMsFromCache(t *testing.T) {
	// Test case 1
	manager := &AzureManager{
		azureCache: &azureCache{
			virtualMachines: make(map[string][]compute.VirtualMachine),
		},
	}
	agentPool := &VMsPool{
		manager: manager,
		azureRef: azureRef{
			Name: "test-vms-pool",
		},
	}

	_, err := agentPool.getVMsFromCache()
	assert.EqualError(t, err, "vms pool test-vms-pool not found in the cache")

	// Test case 2
	manager.azureCache.virtualMachines["test-vms-pool"] = []compute.VirtualMachine{}
	_, err = agentPool.getVMsFromCache()
	assert.NoError(t, err)

	// Test case 3
	manager.azureCache.virtualMachines["test-vms-pool"] = newTestVMsPoolVMList(3)
	vms, err := agentPool.getVMsFromCache()
	assert.NoError(t, err)
	assert.Len(t, vms, 3)

	// Test case 4
	manager.azureCache.virtualMachines["test-vms-pool"] = newTestVMsPoolVMList(3)
	agentPool.azureRef.Name = ""
	_, err = agentPool.getVMsFromCache()
	assert.EqualError(t, err, "vms pool  not found in the cache")
}

// Test cases for Nodes()
// Test case 1 - when there are no VMs in the pool
// Test case 2 - when there are VMs in the pool
// Test case 3 - when there are VMs in the pool with no ID
// Test case 4 - when there is an error converting resource group name
// Test case 5 - when there is an error getting VMs from cache
func TestNodes(t *testing.T) {
	// Test case 1
	manager := &AzureManager{
		azureCache: &azureCache{
			virtualMachines: make(map[string][]compute.VirtualMachine),
		},
	}
	agentPool := &VMsPool{
		manager: manager,
		azureRef: azureRef{
			Name: "test-vms-pool",
		},
	}

	nodes, err := agentPool.Nodes()
	assert.EqualError(t, err, "vms pool test-vms-pool not found in the cache")
	assert.Empty(t, nodes)

	// Test case 2
	manager.azureCache.virtualMachines["test-vms-pool"] = newTestVMsPoolVMList(3)
	nodes, err = agentPool.Nodes()
	assert.NoError(t, err)
	assert.Len(t, nodes, 3)

	// Test case 3
	manager.azureCache.virtualMachines["test-vms-pool"] = newTestVMsPoolVMList(3)
	manager.azureCache.virtualMachines["test-vms-pool"][0].ID = nil
	nodes, err = agentPool.Nodes()
	assert.NoError(t, err)
	assert.Len(t, nodes, 2)
	manager.azureCache.virtualMachines["test-vms-pool"] = newTestVMsPoolVMList(3)
	emptyString := ""
	manager.azureCache.virtualMachines["test-vms-pool"][0].ID = &emptyString
	nodes, err = agentPool.Nodes()
	assert.NoError(t, err)
	assert.Len(t, nodes, 2)

	// Test case 4
	manager.azureCache.virtualMachines["test-vms-pool"] = newTestVMsPoolVMList(3)
	bogusID := "foo"
	manager.azureCache.virtualMachines["test-vms-pool"][0].ID = &bogusID
	nodes, err = agentPool.Nodes()
	assert.Empty(t, nodes)
	assert.Error(t, err)

	// Test case 5
	manager.azureCache.virtualMachines["test-vms-pool"] = newTestVMsPoolVMList(1)
	agentPool.azureRef.Name = ""
	nodes, err = agentPool.Nodes()
	assert.Empty(t, nodes)
	assert.Error(t, err)
}
