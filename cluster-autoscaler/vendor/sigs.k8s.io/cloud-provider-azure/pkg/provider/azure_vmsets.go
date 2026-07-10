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

package provider

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v9"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	cloudprovider "k8s.io/cloud-provider"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
)

//go:generate sh -c "mockgen -destination=$GOPATH/src/sigs.k8s.io/cloud-provider-azure/pkg/provider/azure_mock_vmsets.go -source=$GOPATH/src/sigs.k8s.io/cloud-provider-azure/pkg/provider/azure_vmsets.go -package=provider VMSet"

// VMSet defines functions all vmsets (including scale set and availability
// set) should be implemented.
// Don't forget to run the following command to generate the mock client:
// mockgen -destination=$GOPATH/src/sigs.k8s.io/cloud-provider-azure/pkg/provider/azure_mock_vmsets.go -source=$GOPATH/src/sigs.k8s.io/cloud-provider-azure/pkg/provider/azure_vmsets.go -package=provider VMSet
type VMSet interface {
	// GetInstanceIDByNodeName gets the cloud provider ID by node name.
	// It must return ("", cloudprovider.InstanceNotFound) if the instance does
	// not exist or is no longer running.
	GetInstanceIDByNodeName(ctx context.Context, name string) (string, error)
	// GetInstanceTypeByNodeName gets the instance type by node name.
	GetInstanceTypeByNodeName(ctx context.Context, name string) (string, error)
	// GetIPByNodeName gets machine private IP and public IP by node name.
	GetIPByNodeName(ctx context.Context, name string) (string, string, error)
	// GetPrimaryInterface gets machine primary network interface by node name.
	GetPrimaryInterface(ctx context.Context, nodeName string) (*armnetwork.Interface, error)
	// GetNodeNameByProviderID gets the node name by provider ID.
	GetNodeNameByProviderID(ctx context.Context, providerID string) (types.NodeName, error)

	// GetZoneByNodeName gets cloudprovider.Zone by node name.
	GetZoneByNodeName(ctx context.Context, name string) (cloudprovider.Zone, error)

	// GetPrimaryVMSetName returns the VM set name depending on the configured vmType.
	// It returns config.PrimaryScaleSetName for vmss and config.PrimaryAvailabilitySetName for standard vmType.
	GetPrimaryVMSetName() string
	// GetVMSetNames selects all possible availability sets or scale sets
	// (depending vmType configured) for service load balancer, if the service has
	// no loadbalancer mode annotation returns the primary VMSet. If service annotation
	// for loadbalancer exists then return the eligible VMSet.
	GetVMSetNames(ctx context.Context, service *v1.Service, nodes []*v1.Node) (availabilitySetNames []*string, err error)
	// GetNodeVMSetName returns the availability set or vmss name by the node name.
	// It will return empty string when using standalone vms.
	GetNodeVMSetName(ctx context.Context, node *v1.Node) (string, error)
	// EnsureHostsInPool ensures the given Node's primary IP configurations are
	// participating in the specified LoadBalancer Backend Pool.
	EnsureHostsInPool(ctx context.Context, service *v1.Service, nodes []*v1.Node, backendPoolID string, vmSetName string) error
	// EnsureHostInPool ensures the given VM's Primary NIC's Primary IP Configuration is
	// participating in the specified LoadBalancer Backend Pool.
	EnsureHostInPool(ctx context.Context, service *v1.Service, nodeName types.NodeName, backendPoolID string, vmSetName string) (string, string, string, *armcompute.VirtualMachineScaleSetVM, error)
	// EnsureBackendPoolDeleted ensures the loadBalancer backendAddressPools deleted from the specified nodes.
	EnsureBackendPoolDeleted(ctx context.Context, service *v1.Service, backendPoolIDs []string, vmSetName string, backendAddressPools []*armnetwork.BackendAddressPool, deleteFromVMSet bool) (bool, error)
	// EnsureBackendPoolDeletedFromVMSets ensures the loadBalancer backendAddressPools deleted from the specified VMSS/VMAS
	EnsureBackendPoolDeletedFromVMSets(ctx context.Context, vmSetNamesMap map[string]bool, backendPoolIDs []string) error

	// AttachDisk attaches a disk to vm
	AttachDisk(ctx context.Context, nodeName types.NodeName, diskMap map[string]*AttachDiskOptions) error
	// DetachDisk detaches a disk from vm
	DetachDisk(ctx context.Context, nodeName types.NodeName, diskMap map[string]string, forceDetach bool) error
	// WaitForUpdateResult waits for the response of the update request

	// GetDataDisks gets a list of data disks attached to the node.
	GetDataDisks(ctx context.Context, nodeName types.NodeName, crt azcache.AzureCacheReadType) ([]*armcompute.DataDisk, *string, error)

	// UpdateVM updates a vm
	UpdateVM(ctx context.Context, nodeName types.NodeName) error

	// GetPowerStatusByNodeName returns the powerState for the specified node.
	GetPowerStatusByNodeName(ctx context.Context, name string) (string, error)

	// GetProvisioningStateByNodeName returns the provisioningState for the specified node.
	GetProvisioningStateByNodeName(ctx context.Context, name string) (string, error)

	// GetPrivateIPsByNodeName returns a slice of all private ips assigned to node (ipv6 and ipv4)
	GetPrivateIPsByNodeName(ctx context.Context, name string) ([]string, error)

	// GetNodeNameByIPConfigurationID gets the nodeName and vmSetName by IP configuration ID.
	GetNodeNameByIPConfigurationID(ctx context.Context, ipConfigurationID string) (string, string, error)

	// GetNodeCIDRMasksByProviderID returns the node CIDR subnet mask by provider ID.
	GetNodeCIDRMasksByProviderID(ctx context.Context, providerID string) (int, int, error)

	// GetAgentPoolVMSetNames returns all vmSet names according to the nodes
	GetAgentPoolVMSetNames(ctx context.Context, nodes []*v1.Node) ([]*string, error)

	// DeleteCacheForNode removes the node entry from cache.
	DeleteCacheForNode(ctx context.Context, nodeName string) error

	// RefreshCaches invalidates and renew all related caches.
	RefreshCaches() error
}

// AttachDiskOptions attach disk options
type AttachDiskOptions struct {
	CachingMode             armcompute.CachingTypes
	DiskName                string
	DiskEncryptionSetID     string
	WriteAcceleratorEnabled bool
	Lun                     int32
}
