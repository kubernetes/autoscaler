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
	"errors"
	"fmt"
	"os"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	cloudprovider "k8s.io/cloud-provider"

	azcache "sigs.k8s.io/cloud-provider-azure/pkg/cache"
	"sigs.k8s.io/cloud-provider-azure/pkg/consts"
	"sigs.k8s.io/cloud-provider-azure/pkg/log"
)

var _ cloudprovider.Instances = (*Cloud)(nil)

const (
	// nodeNameEnvironmentName is the environment variable name for getting node name.
	// It is only used for out-of-tree cloud provider.
	nodeNameEnvironmentName = "NODE_NAME"
)

var (
	errNodeNotInitialized = fmt.Errorf("providerID is empty, the node is not initialized yet")
)

func (az *Cloud) addressGetter(ctx context.Context, nodeName types.NodeName) ([]v1.NodeAddress, error) {
	logger := log.FromContextOrBackground(ctx).WithName("addressGetter")
	ip, publicIP, err := az.getIPForMachine(ctx, nodeName)
	if err != nil {
		logger.V(2).Info("NodeAddresses abort backoff", "nodeName", nodeName, "error", err)
		return nil, err
	}

	addresses := []v1.NodeAddress{
		{Type: v1.NodeInternalIP, Address: ip},
		{Type: v1.NodeHostName, Address: string(nodeName)},
	}
	if len(publicIP) > 0 {
		addresses = append(addresses, v1.NodeAddress{
			Type:    v1.NodeExternalIP,
			Address: publicIP,
		})
	}
	return addresses, nil
}

// NodeAddresses returns the addresses of the specified instance.
func (az *Cloud) NodeAddresses(ctx context.Context, name types.NodeName) ([]v1.NodeAddress, error) {
	logger := log.FromContextOrBackground(ctx).WithName("NodeAddresses")
	// Returns nil for unmanaged nodes because azure cloud provider couldn't fetch information for them.
	unmanaged, err := az.IsNodeUnmanaged(string(name))
	if err != nil {
		return nil, err
	}
	if unmanaged {
		logger.V(4).Info("omitting unmanaged node", "nodeName", name)
		return nil, nil
	}

	if az.UseInstanceMetadata {
		metadata, err := az.Metadata.GetMetadata(ctx, azcache.CacheReadTypeDefault)
		if err != nil {
			return nil, err
		}

		if metadata.Compute == nil || metadata.Network == nil {
			return nil, fmt.Errorf("failure of getting instance metadata")
		}

		isLocalInstance, err := az.isCurrentInstance(name, metadata.Compute.Name)
		if err != nil {
			return nil, err
		}

		// Not local instance, get addresses from Azure ARM API.
		if !isLocalInstance {
			if az.VMSet != nil {
				return az.addressGetter(ctx, name)
			}

			// vmSet == nil indicates credentials are not provided.
			return nil, fmt.Errorf("no credentials provided for Azure cloud provider")
		}

		return az.getLocalInstanceNodeAddresses(metadata.Network.Interface, string(name))
	}

	return az.addressGetter(ctx, name)
}

func (az *Cloud) getLocalInstanceNodeAddresses(netInterfaces []*NetworkInterface, nodeName string) ([]v1.NodeAddress, error) {
	if len(netInterfaces) == 0 {
		return nil, fmt.Errorf("no interface is found for the instance")
	}

	// Use ip address got from instance metadata.
	netInterface := netInterfaces[0]
	addresses := []v1.NodeAddress{
		{Type: v1.NodeHostName, Address: nodeName},
	}
	if len(netInterface.IPV4.IPAddress) > 0 && len(netInterface.IPV4.IPAddress[0].PrivateIP) > 0 {
		address := netInterface.IPV4.IPAddress[0]
		addresses = append(addresses, v1.NodeAddress{
			Type:    v1.NodeInternalIP,
			Address: address.PrivateIP,
		})
		if len(address.PublicIP) > 0 {
			addresses = append(addresses, v1.NodeAddress{
				Type:    v1.NodeExternalIP,
				Address: address.PublicIP,
			})
		}
	}
	if len(netInterface.IPV6.IPAddress) > 0 && len(netInterface.IPV6.IPAddress[0].PrivateIP) > 0 {
		address := netInterface.IPV6.IPAddress[0]
		addresses = append(addresses, v1.NodeAddress{
			Type:    v1.NodeInternalIP,
			Address: address.PrivateIP,
		})
		if len(address.PublicIP) > 0 {
			addresses = append(addresses, v1.NodeAddress{
				Type:    v1.NodeExternalIP,
				Address: address.PublicIP,
			})
		}
	}

	if len(addresses) == 1 {
		// No IP addresses is got from instance metadata service, clean up cache and report errors.
		_ = az.Metadata.imsCache.Delete(consts.MetadataCacheKey)
		return nil, fmt.Errorf("get empty IP addresses from instance metadata service")
	}
	return addresses, nil
}

// NodeAddressesByProviderID returns the node addresses of an instances with the specified unique providerID
// This method will not be called from the node that is requesting this ID. i.e. metadata service
// and other local methods cannot be used here
func (az *Cloud) NodeAddressesByProviderID(ctx context.Context, providerID string) ([]v1.NodeAddress, error) {
	logger := log.FromContextOrBackground(ctx).WithName("NodeAddressesByProviderID")
	if providerID == "" {
		return nil, errNodeNotInitialized
	}

	// Returns nil for unmanaged nodes because azure cloud provider couldn't fetch information for them.
	if az.IsNodeUnmanagedByProviderID(providerID) {
		logger.V(4).Info("omitting unmanaged node", "providerID", providerID)
		return nil, nil
	}

	if az.VMSet == nil {
		// vmSet == nil indicates credentials are not provided.
		return nil, fmt.Errorf("no credentials provided for Azure cloud provider")
	}

	name, err := az.VMSet.GetNodeNameByProviderID(ctx, providerID)
	if err != nil {
		return nil, err
	}

	return az.NodeAddresses(ctx, name)
}

// InstanceExistsByProviderID returns true if the instance with the given provider id still exists and is running.
// If false is returned with no error, the instance will be immediately deleted by the cloud controller manager.
func (az *Cloud) InstanceExistsByProviderID(ctx context.Context, providerID string) (bool, error) {
	logger := log.FromContextOrBackground(ctx).WithName("InstanceExistsByProviderID")
	if providerID == "" {
		return false, errNodeNotInitialized
	}

	// Returns true for unmanaged nodes because azure cloud provider always assumes them exists.
	if az.IsNodeUnmanagedByProviderID(providerID) {
		logger.V(4).Info("assuming unmanaged node exists", "providerID", providerID)
		return true, nil
	}

	if az.VMSet == nil {
		// vmSet == nil indicates credentials are not provided.
		return false, fmt.Errorf("no credentials provided for Azure cloud provider")
	}

	name, err := az.VMSet.GetNodeNameByProviderID(ctx, providerID)
	if err != nil {
		if errors.Is(err, cloudprovider.InstanceNotFound) {
			return false, nil
		}
		return false, err
	}

	_, err = az.InstanceID(ctx, name)
	if err != nil {
		if errors.Is(err, cloudprovider.InstanceNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// InstanceShutdownByProviderID returns true if the instance is in safe state to detach volumes
func (az *Cloud) InstanceShutdownByProviderID(ctx context.Context, providerID string) (bool, error) {
	logger := log.FromContextOrBackground(ctx).WithName("InstanceShutdownByProviderID")
	if providerID == "" {
		return false, nil
	}
	if az.VMSet == nil {
		// vmSet == nil indicates credentials are not provided.
		return false, fmt.Errorf("no credentials provided for Azure cloud provider")
	}

	nodeName, err := az.VMSet.GetNodeNameByProviderID(ctx, providerID)
	if err != nil {
		// Returns false, so the controller manager will continue to check InstanceExistsByProviderID().
		if errors.Is(err, cloudprovider.InstanceNotFound) {
			return false, nil
		}

		return false, err
	}

	powerStatus, err := az.VMSet.GetPowerStatusByNodeName(ctx, string(nodeName))
	if err != nil {
		// Returns false, so the controller manager will continue to check InstanceExistsByProviderID().
		if errors.Is(err, cloudprovider.InstanceNotFound) {
			return false, nil
		}

		return false, err
	}
	logger.V(3).Info("gets power status for node", "powerStatus", powerStatus, "nodeName", nodeName)

	provisioningState, err := az.VMSet.GetProvisioningStateByNodeName(ctx, string(nodeName))
	if err != nil {
		// Returns false, so the controller manager will continue to check InstanceExistsByProviderID().
		if errors.Is(err, cloudprovider.InstanceNotFound) {
			return false, nil
		}

		return false, err
	}
	logger.V(3).Info("gets provisioning state for node", "provisioningState", provisioningState, "nodeName", nodeName)

	status := strings.ToLower(powerStatus)
	provisioningSucceeded := strings.EqualFold(strings.ToLower(provisioningState), strings.ToLower(string(consts.ProvisioningStateSucceeded)))
	return provisioningSucceeded && (status == consts.VMPowerStateStopped || status == consts.VMPowerStateDeallocated || status == consts.VMPowerStateDeallocating), nil
}

func (az *Cloud) isCurrentInstance(name types.NodeName, metadataVMName string) (bool, error) {
	var err error
	nodeName := mapNodeNameToVMName(name)

	// VMSS vmName is not same with hostname, use hostname instead.
	if az.VMType == consts.VMTypeVMSS {
		metadataVMName, err = os.Hostname()
		if err != nil {
			return false, err
		}

		// Use name from env variable "NODE_NAME" if it is set.
		nodeNameEnv := os.Getenv(nodeNameEnvironmentName)
		if nodeNameEnv != "" {
			metadataVMName = nodeNameEnv
		}
	}

	metadataVMName = strings.ToLower(metadataVMName)
	return metadataVMName == nodeName, nil
}

// InstanceID returns the cloud provider ID of the specified instance.
// Note that if the instance does not exist or is no longer running, we must return ("", cloudprovider.InstanceNotFound)
func (az *Cloud) InstanceID(ctx context.Context, name types.NodeName) (string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("InstanceID")
	nodeName := mapNodeNameToVMName(name)
	unmanaged, err := az.IsNodeUnmanaged(nodeName)
	if err != nil {
		return "", err
	}
	if unmanaged {
		// InstanceID is same with nodeName for unmanaged nodes.
		logger.V(4).Info("getting ID for unmanaged node", "id", name, "unmanaged", name)
		return nodeName, nil
	}

	if az.UseInstanceMetadata {
		metadata, err := az.Metadata.GetMetadata(ctx, azcache.CacheReadTypeDefault)
		if err != nil {
			return "", err
		}

		if metadata.Compute == nil {
			return "", fmt.Errorf("failure of getting instance metadata")
		}

		isLocalInstance, err := az.isCurrentInstance(name, metadata.Compute.Name)
		if err != nil {
			return "", err
		}

		// Not local instance, get instanceID from Azure ARM API.
		if !isLocalInstance {
			if az.VMSet != nil {
				return az.VMSet.GetInstanceIDByNodeName(ctx, nodeName)
			}

			// vmSet == nil indicates credentials are not provided.
			return "", fmt.Errorf("no credentials provided for Azure cloud provider")
		}
		return az.getLocalInstanceProviderID(metadata, nodeName)
	}

	return az.VMSet.GetInstanceIDByNodeName(ctx, nodeName)
}

func (az *Cloud) getLocalInstanceProviderID(metadata *InstanceMetadata, _ string) (string, error) {
	// Get resource group name and subscription ID.
	resourceGroup := strings.ToLower(metadata.Compute.ResourceGroup)
	subscriptionID := strings.ToLower(metadata.Compute.SubscriptionID)

	if metadata.Compute.ResourceID == "" {
		// No ResourceID is got from instance metadata service, clean up cache and report errors.
		_ = az.Metadata.imsCache.Delete(consts.MetadataCacheKey)
		return "", fmt.Errorf("get empty ResoureceID from instance metadata service")
	}

	providerID := strings.ReplaceAll(metadata.Compute.ResourceID, metadata.Compute.SubscriptionID, subscriptionID)
	providerID = strings.ReplaceAll(providerID, metadata.Compute.ResourceGroup, resourceGroup)

	return providerID, nil
}

// InstanceTypeByProviderID returns the cloudprovider instance type of the node with the specified unique providerID
// This method will not be called from the node that is requesting this ID. i.e. metadata service
// and other local methods cannot be used here
func (az *Cloud) InstanceTypeByProviderID(ctx context.Context, providerID string) (string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("InstanceTypeByProviderID")
	if providerID == "" {
		return "", errNodeNotInitialized
	}

	// Returns "" for unmanaged nodes because azure cloud provider couldn't fetch information for them.
	if az.IsNodeUnmanagedByProviderID(providerID) {
		logger.V(4).Info("omitting unmanaged node", "providerID", providerID)
		return "", nil
	}

	if az.VMSet == nil {
		// vmSet == nil indicates credentials are not provided.
		return "", fmt.Errorf("no credentials provided for Azure cloud provider")
	}

	name, err := az.VMSet.GetNodeNameByProviderID(ctx, providerID)
	if err != nil {
		return "", err
	}

	return az.InstanceType(ctx, name)
}

// InstanceType returns the type of the specified instance.
// Note that if the instance does not exist or is no longer running, we must return ("", cloudprovider.InstanceNotFound)
// (Implementer Note): This is used by kubelet. Kubelet will label the node. Real log from kubelet:
// Adding node label from cloud provider: beta.kubernetes.io/instance-type=[value]
func (az *Cloud) InstanceType(ctx context.Context, name types.NodeName) (string, error) {
	logger := log.FromContextOrBackground(ctx).WithName("InstanceType")
	// Returns "" for unmanaged nodes because azure cloud provider couldn't fetch information for them.
	unmanaged, err := az.IsNodeUnmanaged(string(name))
	if err != nil {
		return "", err
	}
	if unmanaged {
		logger.V(4).Info("omitting unmanaged node", "nodeName", name)
		return "", nil
	}

	if az.UseInstanceMetadata {
		metadata, err := az.Metadata.GetMetadata(ctx, azcache.CacheReadTypeDefault)
		if err != nil {
			return "", err
		}

		if metadata.Compute == nil {
			return "", fmt.Errorf("failure of getting instance metadata")
		}

		isLocalInstance, err := az.isCurrentInstance(name, metadata.Compute.Name)
		if err != nil {
			return "", err
		}
		if !isLocalInstance {
			if az.VMSet != nil {
				return az.VMSet.GetInstanceTypeByNodeName(ctx, string(name))
			}

			// vmSet == nil indicates credentials are not provided.
			return "", fmt.Errorf("no credentials provided for Azure cloud provider")
		}

		if metadata.Compute.VMSize != "" {
			return metadata.Compute.VMSize, nil
		}
	}

	if az.VMSet == nil {
		// vmSet == nil indicates credentials are not provided.
		return "", fmt.Errorf("no credentials provided for Azure cloud provider")
	}

	return az.VMSet.GetInstanceTypeByNodeName(ctx, string(name))
}

// AddSSHKeyToAllInstances adds an SSH public key as a legal identity for all instances
// expected format for the key is standard ssh-keygen format: <protocol> <blob>
func (az *Cloud) AddSSHKeyToAllInstances(_ context.Context, _ string, _ []byte) error {
	return cloudprovider.NotImplemented
}

// CurrentNodeName returns the name of the node we are currently running on.
// On Azure this is the hostname, so we just return the hostname.
func (az *Cloud) CurrentNodeName(_ context.Context, hostname string) (types.NodeName, error) {
	return types.NodeName(hostname), nil
}

// mapNodeNameToVMName maps a k8s NodeName to an Azure VM Name
// This is a simple string cast.
func mapNodeNameToVMName(nodeName types.NodeName) string {
	return string(nodeName)
}
