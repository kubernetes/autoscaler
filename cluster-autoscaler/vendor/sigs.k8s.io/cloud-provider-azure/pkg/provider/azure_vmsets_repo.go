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

package provider

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-07-01/network"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	cloudprovider "k8s.io/cloud-provider"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
)

// GetVirtualMachineWithRetry invokes az.getVirtualMachine with exponential backoff retry
func (az *Cloud) GetVirtualMachineWithRetry(name types.NodeName, crt azcache.AzureCacheReadType) (compute.VirtualMachine, error) {
	var machine compute.VirtualMachine
	var retryErr error
	err := wait.ExponentialBackoff(az.RequestBackoff(), func() (bool, error) {
		machine, retryErr = az.getVirtualMachine(name, crt)
		if errors.Is(retryErr, cloudprovider.InstanceNotFound) {
			return true, cloudprovider.InstanceNotFound
		}
		if retryErr != nil {
			klog.Errorf("GetVirtualMachineWithRetry(%s): backoff failure, will retry, err=%v", name, retryErr)
			return false, nil
		}
		klog.V(2).Infof("GetVirtualMachineWithRetry(%s): backoff success", name)
		return true, nil
	})
	if errors.Is(err, wait.ErrWaitTimeout) {
		err = retryErr
	}
	return machine, err
}

// ListVirtualMachines invokes az.VirtualMachinesClient.List with exponential backoff retry
func (az *Cloud) ListVirtualMachines(resourceGroup string) ([]compute.VirtualMachine, error) {
	ctx, cancel := getContextWithCancel()
	defer cancel()

	allNodes, rerr := az.VirtualMachinesClient.List(ctx, resourceGroup)
	if rerr != nil {
		klog.Errorf("VirtualMachinesClient.List(%v) failure with err=%v", resourceGroup, rerr)
		return nil, rerr.Error()
	}
	klog.V(6).Infof("VirtualMachinesClient.List(%v) success", resourceGroup)
	return allNodes, nil
}

// getPrivateIPsForMachine is wrapper for optional backoff getting private ips
// list of a node by name
func (az *Cloud) getPrivateIPsForMachine(nodeName types.NodeName) ([]string, error) {
	return az.getPrivateIPsForMachineWithRetry(nodeName)
}

func (az *Cloud) getPrivateIPsForMachineWithRetry(nodeName types.NodeName) ([]string, error) {
	var privateIPs []string
	err := wait.ExponentialBackoff(az.RequestBackoff(), func() (bool, error) {
		var retryErr error
		privateIPs, retryErr = az.VMSet.GetPrivateIPsByNodeName(string(nodeName))
		if retryErr != nil {
			// won't retry since the instance doesn't exist on Azure.
			if errors.Is(retryErr, cloudprovider.InstanceNotFound) {
				return true, retryErr
			}
			klog.Errorf("GetPrivateIPsByNodeName(%s): backoff failure, will retry,err=%v", nodeName, retryErr)
			return false, nil
		}
		klog.V(3).Infof("GetPrivateIPsByNodeName(%s): backoff success", nodeName)
		return true, nil
	})
	return privateIPs, err
}

func (az *Cloud) getIPForMachine(nodeName types.NodeName) (string, string, error) {
	return az.GetIPForMachineWithRetry(nodeName)
}

// GetIPForMachineWithRetry invokes az.getIPForMachine with exponential backoff retry
func (az *Cloud) GetIPForMachineWithRetry(name types.NodeName) (string, string, error) {
	var ip, publicIP string
	err := wait.ExponentialBackoff(az.RequestBackoff(), func() (bool, error) {
		var retryErr error
		ip, publicIP, retryErr = az.VMSet.GetIPByNodeName(string(name))
		if retryErr != nil {
			klog.Errorf("GetIPForMachineWithRetry(%s): backoff failure, will retry,err=%v", name, retryErr)
			return false, nil
		}
		klog.V(3).Infof("GetIPForMachineWithRetry(%s): backoff success", name)
		return true, nil
	})
	return ip, publicIP, err
}

func (az *Cloud) newVMCache() (azcache.Resource, error) {
	getter := func(key string) (interface{}, error) {
		// Currently InstanceView request are used by azure_zones, while the calls come after non-InstanceView
		// request. If we first send an InstanceView request and then a non InstanceView request, the second
		// request will still hit throttling. This is what happens now for cloud controller manager: In this
		// case we do get instance view every time to fulfill the azure_zones requirement without hitting
		// throttling.
		// Consider adding separate parameter for controlling 'InstanceView' once node update issue #56276 is fixed
		ctx, cancel := getContextWithCancel()
		defer cancel()

		resourceGroup, err := az.GetNodeResourceGroup(key)
		if err != nil {
			return nil, err
		}

		vm, verr := az.VirtualMachinesClient.Get(ctx, resourceGroup, key, compute.InstanceViewTypesInstanceView)
		exists, rerr := checkResourceExistsFromError(verr)
		if rerr != nil {
			return nil, rerr.Error()
		}

		if !exists {
			klog.V(2).Infof("Virtual machine %q not found", key)
			return nil, nil
		}

		if vm.VirtualMachineProperties != nil &&
			strings.EqualFold(pointer.StringDeref(vm.VirtualMachineProperties.ProvisioningState, ""), string(consts.ProvisioningStateDeleting)) {
			klog.V(2).Infof("Virtual machine %q is under deleting", key)
			return nil, nil
		}

		return &vm, nil
	}

	if az.VMCacheTTLInSeconds == 0 {
		az.VMCacheTTLInSeconds = vmCacheTTLDefaultInSeconds
	}
	return azcache.NewTimedCache(time.Duration(az.VMCacheTTLInSeconds)*time.Second, getter, az.Config.DisableAPICallCache)
}

// getVirtualMachine calls 'VirtualMachinesClient.Get' with a timed cache
// The service side has throttling control that delays responses if there are multiple requests onto certain vm
// resource request in short period.
func (az *Cloud) getVirtualMachine(nodeName types.NodeName, crt azcache.AzureCacheReadType) (vm compute.VirtualMachine, err error) {
	vmName := string(nodeName)
	cachedVM, err := az.vmCache.Get(vmName, crt)
	if err != nil {
		return vm, err
	}

	if cachedVM == nil {
		klog.Warningf("Unable to find node %s: %v", nodeName, cloudprovider.InstanceNotFound)
		return vm, cloudprovider.InstanceNotFound
	}

	return *(cachedVM.(*compute.VirtualMachine)), nil
}

func (az *Cloud) getRouteTable(crt azcache.AzureCacheReadType) (routeTable network.RouteTable, exists bool, err error) {
	if len(az.RouteTableName) == 0 {
		return routeTable, false, fmt.Errorf("route table name is not configured")
	}

	cachedRt, err := az.rtCache.GetWithDeepCopy(az.RouteTableName, crt)
	if err != nil {
		return routeTable, false, err
	}

	if cachedRt == nil {
		return routeTable, false, nil
	}

	return *(cachedRt.(*network.RouteTable)), true, nil
}
