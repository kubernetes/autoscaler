/*
SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors

SPDX-License-Identifier: Apache-2.0
*/

// Package api defined the schema of the Azure Provider Spec
package api

const (
	// AzureClientID is a constant for a key name that is part of the Azure cloud credentials.
	AzureClientID string = "azureClientId"
	// AzureClientSecret is a constant for a key name that is part of the Azure cloud credentials.
	AzureClientSecret string = "azureClientSecret"
	// AzureSubscriptionID is a constant for a key name that is part of the Azure cloud credentials.
	AzureSubscriptionID string = "azureSubscriptionId"
	// AzureTenantID is a constant for a key name that is part of the Azure cloud credentials.
	AzureTenantID string = "azureTenantId"

	// AzureAlternativeClientID is a constant for a key name of a secret containing the Azure credentials (client id).
	AzureAlternativeClientID = "clientID"
	// AzureAlternativeClientSecret is a constant for a key name of a secret containing the Azure credentials (client
	// secret).
	AzureAlternativeClientSecret = "clientSecret"
	// AzureAlternativeSubscriptionID is a constant for a key name of a secret containing the Azure credentials
	// (subscription id).
	AzureAlternativeSubscriptionID = "subscriptionID"
	// AzureAlternativeTenantID is a constant for a key name of a secret containing the Azure credentials (tenant id).
	AzureAlternativeTenantID = "tenantID"

	// MachineSetKindAvailabilitySet is the machine set kind for AvailabilitySet
	MachineSetKindAvailabilitySet string = "availabilityset"
	// MachineSetKindVMO is the machine set kind for VirtualMachineScaleSet Orchestration Mode VM (VMO)
	MachineSetKindVMO string = "vmo"
)

// AzureProviderSpec is the spec to be used while parsing the calls.
type AzureProviderSpec struct {
	Location      string                        `json:"location,omitempty"`
	Tags          map[string]string             `json:"tags,omitempty"`
	Properties    AzureVirtualMachineProperties `json:"properties,omitempty"`
	ResourceGroup string                        `json:"resourceGroup,omitempty"`
	SubnetInfo    AzureSubnetInfo               `json:"subnetInfo,omitempty"`
}

// AzureVirtualMachineProperties is describes the properties of a Virtual Machine.
type AzureVirtualMachineProperties struct {
	HardwareProfile AzureHardwareProfile   `json:"hardwareProfile,omitempty"`
	StorageProfile  AzureStorageProfile    `json:"storageProfile,omitempty"`
	OsProfile       AzureOSProfile         `json:"osProfile,omitempty"`
	NetworkProfile  AzureNetworkProfile    `json:"networkProfile,omitempty"`
	AvailabilitySet *AzureSubResource      `json:"availabilitySet,omitempty"`
	IdentityID      *string                `json:"identityID,omitempty"`
	Zone            *int                   `json:"zone,omitempty"`
	MachineSet      *AzureMachineSetConfig `json:"machineSet,omitempty"`
}

// AzureHardwareProfile is specifies the hardware settings for the virtual machine.
// Refer github.com/Azure/azure-sdk-for-go/arm/compute/models.go for VMSizes
type AzureHardwareProfile struct {
	VMSize string `json:"vmSize,omitempty"`
}

// AzureMachineSetConfig contains the information about the machine set
type AzureMachineSetConfig struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

// AzureStorageProfile is specifies the storage settings for the virtual machine disks.
type AzureStorageProfile struct {
	ImageReference AzureImageReference `json:"imageReference,omitempty"`
	OsDisk         AzureOSDisk         `json:"osDisk,omitempty"`
	DataDisks      []AzureDataDisk     `json:"dataDisks,omitempty"`
}

// AzureImageReference is specifies information about the image to use. You can specify information about platform images,
// marketplace images, community images or virtual machine images. This element is required when you want to use a platform image,
// marketplace image, community image or virtual machine image, but is not used in other creation operations.
type AzureImageReference struct {
	ID string `json:"id,omitempty"`
	// Uniform Resource Name of the OS image to be used , it has the format 'publisher:offer:sku:version'
	URN *string `json:"urn,omitempty"`
	// CommunityGalleryImageID is the id of the OS image to be used, hosted within an Azure Community Image Gallery.
	CommunityGalleryImageID *string `json:"communityGalleryImageID,omitempty"`
}

// AzureOSDisk is specifies information about the operating system disk used by the virtual machine. <br><br> For more
// information about disks, see [About disks and VHDs for Azure virtual
// machines](https://docs.microsoft.com/azure/virtual-machines/virtual-machines-windows-about-disks-vhds?toc=%2fazure%2fvirtual-machines%2fwindows%2ftoc.json).
type AzureOSDisk struct {
	Name         string                     `json:"name,omitempty"`
	Caching      string                     `json:"caching,omitempty"`
	ManagedDisk  AzureManagedDiskParameters `json:"managedDisk,omitempty"`
	DiskSizeGB   int32                      `json:"diskSizeGB,omitempty"`
	CreateOption string                     `json:"createOption,omitempty"`
}

// AzureDataDisk specifies information about the data disk used by the virtual machine.
type AzureDataDisk struct {
	Name               string `json:"name,omitempty"`
	Lun                *int32 `json:"lun,omitempty"`
	Caching            string `json:"caching,omitempty"`
	StorageAccountType string `json:"storageAccountType,omitempty"`
	DiskSizeGB         int32  `json:"diskSizeGB,omitempty"`
}

// AzureManagedDiskParameters is the parameters of a managed disk.
type AzureManagedDiskParameters struct {
	ID                 string `json:"id,omitempty"`
	StorageAccountType string `json:"storageAccountType,omitempty"`
}

// AzureOSProfile is specifies the operating system settings for the virtual machine.
type AzureOSProfile struct {
	ComputerName       string                  `json:"computerName,omitempty"`
	AdminUsername      string                  `json:"adminUsername,omitempty"`
	AdminPassword      string                  `json:"adminPassword,omitempty"`
	CustomData         string                  `json:"customData,omitempty"`
	LinuxConfiguration AzureLinuxConfiguration `json:"linuxConfiguration,omitempty"`
}

// AzureLinuxConfiguration is specifies the Linux operating system settings on the virtual machine. <br><br>For a list of
// supported Linux distributions, see [Linux on Azure-Endorsed
// Distributions](https://docs.microsoft.com/azure/virtual-machines/virtual-machines-linux-endorsed-distros?toc=%2fazure%2fvirtual-machines%2flinux%2ftoc.json)
// <br><br> For running non-endorsed distributions, see [Information for Non-Endorsed
// Distributions](https://docs.microsoft.com/azure/virtual-machines/virtual-machines-linux-create-upload-generic?toc=%2fazure%2fvirtual-machines%2flinux%2ftoc.json).
type AzureLinuxConfiguration struct {
	DisablePasswordAuthentication bool                  `json:"disablePasswordAuthentication,omitempty"`
	SSH                           AzureSSHConfiguration `json:"ssh,omitempty"`
}

// AzureSSHConfiguration is SSH configuration for Linux based VMs running on Azure
type AzureSSHConfiguration struct {
	PublicKeys AzureSSHPublicKey `json:"publicKeys,omitempty"`
}

// AzureSSHPublicKey is contains information about SSH certificate public key and the path on the Linux VM where the public
// key is placed.
type AzureSSHPublicKey struct {
	Path    string `json:"path,omitempty"`
	KeyData string `json:"keyData,omitempty"`
}

// AzureNetworkProfile is specifies the network interfaces of the virtual machine.
type AzureNetworkProfile struct {
	NetworkInterfaces     AzureNetworkInterfaceReference `json:"networkInterfaces,omitempty"`
	AcceleratedNetworking *bool                          `json:"acceleratedNetworking,omitempty"`
}

// AzureNetworkInterfaceReference is describes a network interface reference.
type AzureNetworkInterfaceReference struct {
	ID                                        string `json:"id,omitempty"`
	*AzureNetworkInterfaceReferenceProperties `json:"properties,omitempty"`
}

// AzureNetworkInterfaceReferenceProperties is describes a network interface reference properties.
type AzureNetworkInterfaceReferenceProperties struct {
	Primary bool `json:"primary,omitempty"`
}

// AzureSubResource is the Sub Resource definition.
type AzureSubResource struct {
	ID string `json:"id,omitempty"`
}

// AzureSubnetInfo is the information containing the subnet details
type AzureSubnetInfo struct {
	VnetName          string  `json:"vnetName,omitempty"`
	VnetResourceGroup *string `json:"vnetResourceGroup,omitempty"`
	SubnetName        string  `json:"subnetName,omitempty"`
}
