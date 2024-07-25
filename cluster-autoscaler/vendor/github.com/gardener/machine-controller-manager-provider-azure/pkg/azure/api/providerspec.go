/*
SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors

SPDX-License-Identifier: Apache-2.0
*/

// Package api defined the schema of the Azure Provider Spec
package api

const (
	// AzureClientID is a constant for a key name that is part of the Azure cloud credentials.
	// Deprecated: Use ClientID instead.
	AzureClientID string = "azureClientId"
	// AzureClientSecret is a constant for a key name that is part of the Azure cloud credentials.
	// Deprecated: Use ClientSecret instead
	AzureClientSecret string = "azureClientSecret"
	// AzureSubscriptionID is a constant for a key name that is part of the Azure cloud credentials.
	// Deprecated: Use SubscriptionID instead
	AzureSubscriptionID string = "azureSubscriptionId"
	// AzureTenantID is a constant for a key name that is part of the Azure cloud credentials.
	// Deprecated: Use TenantID instead
	AzureTenantID string = "azureTenantId"

	// AzureAlternativeClientID is a constant for a key name of a secret containing the Azure credentials (access id).
	// Deprecated: Use ClientID instead.
	AzureAlternativeClientID = "clientID"
	// AzureAlternativeClientSecret is a constant for a key name of a secret containing the Azure credentials (access
	// secret).
	// Deprecated: Use ClientSecret instead
	AzureAlternativeClientSecret = "clientSecret"
	// AzureAlternativeSubscriptionID is a constant for a key name of a secret containing the Azure credentials
	// (subscription id).
	// Deprecated: Use ClientID instead.
	AzureAlternativeSubscriptionID = "subscriptionID"
	// AzureAlternativeTenantID is a constant for a key name of a secret containing the Azure credentials (tenant id).
	// Deprecated: Use TenantID instead
	AzureAlternativeTenantID = "tenantID"

	// ClientID is a constant for a key name that is part of the Azure cloud credentials.
	ClientID string = "clientID"
	// ClientSecret is a constant for a key name that is part of the Azure cloud credentials.
	ClientSecret string = "clientSecret"
	// SubscriptionID is a constant for a key name that is part of the Azure cloud credentials.
	SubscriptionID string = "subscriptionID"
	// TenantID is a constant for a key name that is part of the Azure cloud credentials.
	TenantID string = "tenantID"
	// UserData is a constant for a key name that is part of the secret passed to Driver methods.
	// This contains a base64 encoded custom script that is run upon start of a VM.
	UserData string = "userData"

	// MachineSetKindAvailabilitySet is the machine set kind for AvailabilitySet.
	// Deprecated: Use AzureVirtualMachineProperties.AvailabilitySet instead.
	MachineSetKindAvailabilitySet string = "availabilityset"
	// MachineSetKindVMO is the machine set kind for VirtualMachineScaleSet Orchestration Mode VM (VMO).
	// Deprecated: Use AzureVirtualMachineProperties.VirtualMachineScaleSet instead.
	MachineSetKindVMO string = "vmo"
)

// AzureProviderSpec is the spec to be used while parsing the calls.
type AzureProviderSpec struct {
	// Location is the name of the region where resources will be created.
	Location string `json:"location,omitempty"`
	// Tags is a map of key-value pairs that will be set on resources. Currently, the tags are shared across VM, NIC, Disks.
	// This is not ideal and will change with https://github.com/gardener/machine-controller-manager/blob/master/docs/proposals/hotupdate-instances.md
	Tags map[string]string `json:"tags,omitempty"`
	// Properties defines configuration properties for different profiles (hardware, os, network, storage, availability/virtual-machine-scale-set etc.)
	Properties AzureVirtualMachineProperties `json:"properties,omitempty"`
	// ResourceGroup is a container that holds related resources for an azure solution. See [https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/overview#resource-groups].
	ResourceGroup string `json:"resourceGroup,omitempty"`
	// SubnetInfo contains the configuration for an existing subnet.
	SubnetInfo AzureSubnetInfo `json:"subnetInfo,omitempty"`
	// CloudConfiguration contains config that controls which cloud to connect to
	CloudConfiguration *CloudConfiguration `json:"cloudConfiguration,omitempty"`
}

// AzureVirtualMachineProperties describes the properties of a Virtual Machine.
type AzureVirtualMachineProperties struct {
	// HardwareProfile specifies the hardware settings for the virtual machine. Currently only VMSize is supported.
	HardwareProfile AzureHardwareProfile `json:"hardwareProfile,omitempty"`
	// StorageProfile specifies the storage settings for the virtual machine.
	StorageProfile AzureStorageProfile `json:"storageProfile,omitempty"`
	// OsProfile specifies the operating system settings used when the virtual machine is created.
	OsProfile AzureOSProfile `json:"osProfile,omitempty"`
	// NetworkProfile specifies the network interfaces for the virtual machine.
	NetworkProfile AzureNetworkProfile `json:"networkProfile,omitempty"`
	// AvailabilitySet specifies the availability set to be associated with the virtual machine.
	// For additional information see: [https://learn.microsoft.com/en-us/azure/virtual-machines/availability-set-overview]
	// Points to note:
	// 1. A VM can only be added to availability set at creation time.
	// 2. The availability set to which the VM is being added should be under the same resource group as the availability set resource.
	// 3. Either of AvailabilitySet or VirtualMachineScaleSet should be specified but not both.
	AvailabilitySet *AzureSubResource `json:"availabilitySet,omitempty"`
	// IdentityID is the managed identity that is associated to the virtual machine.
	// NOTE: Currently only user assigned managed identity is supported.
	// For additional information see the following links:
	// 1. [https://learn.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview]
	// 2: [https://learn.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/qs-configure-portal-windows-vm]
	IdentityID *string `json:"identityID,omitempty"`
	// Zone is an availability zone where the virtual machine will be created.
	Zone *int `json:"zone,omitempty"`
	// VirtualMachineScaleSet specifies the virtual machine scale set to be associated with the virtual machine.
	// For additional information see: [https://learn.microsoft.com/en-us/azure/virtual-machine-scale-sets/]
	// Points to note:
	// 1. A VM can only be added to availability set at creation time.
	// 2. Either of AvailabilitySet or VirtualMachineScaleSet should be specified but not both.
	// 3. Only `Flexible` variant of VMSS is currently supported. It is strongly recommended that consumers turn-off any
	// autoscaling capabilities as it interferes with the lifecycle management of MCM and auto-scaling capabilities offered by Cluster-Autoscaler.
	VirtualMachineScaleSet *AzureSubResource `json:"virtualMachineScaleSet,omitempty"`
	// DiagnosticsProfile specifies if boot metrics are enabled and where they are stored
	// For additional information see: [https://learn.microsoft.com/en-us/azure/virtual-machines/boot-diagnostics]
	DiagnosticsProfile *AzureDiagnosticsProfile `json:"diagnosticsProfile,omitempty"`
	// Deprecated: Use either AvailabilitySet or VirtualMachineScaleSet instead
	MachineSet *AzureMachineSetConfig `json:"machineSet,omitempty"`
	// SecurityProfile specifies the security profile to be used for the virtual machine.
	SecurityProfile *AzureSecurityProfile `json:"securityProfile,omitempty"`
}

// AzureSecurityProfile specifies the security profile to be used for the virtual machine.
type AzureSecurityProfile struct {
	// SecurityType specifies the SecurityType attribute of the virtual machine.
	SecurityType *string `json:"securityType,omitempty"`
	// UefiSettings controls the UEFI parameters for the virtual machine.
	UefiSettings *AzureUefiSettings `json:"uefiSettings,omitempty"`
}

// AzureUefiSettings controls the UEFI parameters for the virtual machine.
type AzureUefiSettings struct {
	// VTpmEnabled enables vTPM for the virtual machine.
	// See https://learn.microsoft.com/en-us/azure/virtual-machines/trusted-launch#vtpm
	VTpmEnabled *bool `json:"vtpmEnabled,omitempty"`
	// SecureBootEnabled enables the use of Secure Boot for the virtual machine.
	SecureBootEnabled *bool `json:"secureBootEnabled,omitempty"`
}

// AzureHardwareProfile specifies the hardware settings for the virtual machine.
// Refer to the [azure-sdk-for-go repository](https://github.com/Azure/azure-sdk-for-go/blob/main/sdk/resourcemanager/compute/armcompute/models.go) for VMSizes.
type AzureHardwareProfile struct {
	// VMSize is an alias for different machine sizes supported by the provider.
	// See [https://docs.microsoft.com/azure/virtual-machines/sizes].The available VM sizes depend on region and availability set.
	VMSize string `json:"vmSize,omitempty"`
}

// AzureMachineSetConfig contains the information about the machine set.
// Deprecated: This type should not be used to differentiate between VirtualMachineScaleSet and AvailabilitySet as
// there are now dedicated struct fields for these.
type AzureMachineSetConfig struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

// AzureStorageProfile specifies the storage settings for the virtual machine disks.
type AzureStorageProfile struct {
	// ImageReference specifies information about the image to use. One can specify information about platform images, marketplace images, or virtual machine images.
	ImageReference AzureImageReference `json:"imageReference,omitempty"`
	// OsDisk contains the information about the operating system disk used by the VM.
	// See [https://learn.microsoft.com/en-us/azure/virtual-machines/managed-disks-overview#os-disk].
	OsDisk AzureOSDisk `json:"osDisk,omitempty"`
	// DataDisks contains the information about disks that can be added as data-disks to a VM.
	// See [https://learn.microsoft.com/en-us/azure/virtual-machines/managed-disks-overview#data-disk]
	DataDisks []AzureDataDisk `json:"dataDisks,omitempty"`
}

// AzureImageReference specifies information about the image to use. You can specify information about platform images,
// marketplace images, community images, shared gallery images or virtual machine images. This element is required when you want to use a platform image,
// marketplace image, community image, shared gallery image or virtual machine image, but is not used in other creation operations.
type AzureImageReference struct {
	ID string `json:"id,omitempty"`
	// URN Uniform Resource Name of the OS image to be used, it has the format 'publisher:offer:sku:version'
	// This is a marketplace image. For marketplace images there needs to be a purchase plan and an agreement. The agreement needs to be accepted.
	URN *string `json:"urn,omitempty"`
	// SkipMarketplaceAgreement will prevent the extension from checking the license agreement for marketplace images.
	SkipMarketplaceAgreement bool `json:"skipMarketplaceAgreement,omitempty"`
	// CommunityGalleryImageID is the id of the OS image to be used, hosted within an Azure Community Image Gallery.
	CommunityGalleryImageID *string `json:"communityGalleryImageID,omitempty"`
	// SharedGalleryImageID is the id of the OS image to be used, hosted within an Azure Shared Image Gallery.
	SharedGalleryImageID *string `json:"sharedGalleryImageID,omitempty"`
}

// AzureOSDisk specifies information about the operating system disk used by the virtual machine.
// For more information about disks, see [https://learn.microsoft.com/en-us/azure/virtual-machines/managed-disks-overview].
type AzureOSDisk struct {
	// Name is the name of the OSDisk
	Name string `json:"name,omitempty"`
	// Caching specifies the caching requirements. Possible values are: None, ReadOnly, ReadWrite.
	Caching string `json:"caching,omitempty"`
	// ManagedDisk specifies the managed disk parameters.
	ManagedDisk AzureManagedDiskParameters `json:"managedDisk,omitempty"`
	// DiskSizeGB is the size of an empty disk in gigabytes.
	DiskSizeGB int32 `json:"diskSizeGB,omitempty"`
	// CreateOption Specifies how the virtual machine should be created. Possible values are: [Attach, FromImage].
	// Attach: This value is used when a specialized disk is used to create the virtual machine.
	// FromImage: This value is used when an image is used to create the virtual machine.
	CreateOption string `json:"createOption,omitempty"`
}

// AzureDataDisk specifies information about the data disk used by the virtual machine.
type AzureDataDisk struct {
	// Name is the name of the disk.
	Name string `json:"name,omitempty"`
	// Lun specifies the logical unit number of the data disk. This value is used to identify data disks within the VM and
	// therefore must be unique for each data disk attached to a VM.
	Lun int32 `json:"lun"`
	// Caching specifies the caching requirements. Possible values are: None, ReadOnly, ReadWrite.
	Caching string `json:"caching,omitempty"`
	// StorageAccountType is the storage account type for a managed disk.
	StorageAccountType string `json:"storageAccountType,omitempty"`
	// DiskSizeGB is the size of an empty disk in gigabytes.
	DiskSizeGB int32 `json:"diskSizeGB,omitempty"`
}

// AzureManagedDiskParameters is the parameters of a managed disk.
type AzureManagedDiskParameters struct {
	// ID is a unique resource ID.
	ID string `json:"id,omitempty"`
	// StorageAccountType is the storage account type for a managed disk.
	StorageAccountType string `json:"storageAccountType,omitempty"`
	// SecurityProfile are the parameters of the encryption of the OS disk.
	SecurityProfile *AzureDiskSecurityProfile `json:"securityProfile,omitempty"`
}

// AzureDiskSecurityProfile are the parameters of the encryption of the OS disk.
type AzureDiskSecurityProfile struct {
	// Specifies the EncryptionType of the managed disk. It is set to DiskWithVMGuestState for encryption of the managed disk
	// along with VMGuestState blob, and VMGuestStateOnly for encryption of just the
	// VMGuestState blob. Note: It can be set only Confidential VMs.
	SecurityEncryptionType *string `json:"securityEncryptionType,omitempty"`
}

// AzureOSProfile specifies the operating system settings for the virtual machine.
type AzureOSProfile struct {
	// ComputerName is the host OS name of the virtual machine in azure. However, in mcm-provider-azure this is set to the name of the VM.
	ComputerName string `json:"computerName,omitempty"`
	// AdminUsername is the name of the administrator account.
	AdminUsername string `json:"adminUsername,omitempty"`
	// AdminPassword specifies the password for the administrator account.
	// WARNING: Currently, this property is never used while creating a VM.
	AdminPassword string `json:"adminPassword,omitempty"`
	// CustomData is the base64 encoded string of custom data. The base-64 encoded string is decoded to a binary array that is saved
	// as a file on the Virtual Machine. See [https://azure.microsoft.com/en-us/blog/custom-data-and-cloud-init-on-windows-azure/].
	CustomData string `json:"customData,omitempty"`
	// LinuxConfiguration specifies the linux OS settings on the VM.
	LinuxConfiguration AzureLinuxConfiguration `json:"linuxConfiguration,omitempty"`
}

// AzureLinuxConfiguration specifies the Linux operating system settings on the virtual machine.
// For a list of supported Linux distributions, see [Linux on Azure-Endorsed Distributions](https://learn.microsoft.com/en-us/azure/virtual-machines/linux/endorsed-distros).
type AzureLinuxConfiguration struct {
	// DisablePasswordAuthentication specifies if the password authentication should be disabled.
	DisablePasswordAuthentication bool `json:"disablePasswordAuthentication,omitempty"`
	// SSH specifies the ssh key configurations for a Linux OS.
	SSH AzureSSHConfiguration `json:"ssh,omitempty"`
}

// AzureSSHConfiguration is SSH configuration for Linux based VMs running on Azure.
type AzureSSHConfiguration struct {
	// PublicKeys specifies a list of SSH public keys used to authenticate with linux based VMs.
	PublicKeys AzureSSHPublicKey `json:"publicKeys,omitempty"`
}

// AzureSSHPublicKey contains information about SSH certificate public key and the path on the Linux VM where the public
// key is placed.
type AzureSSHPublicKey struct {
	// Path specifies the full path on the created VM where ssh public key is stored.
	Path string `json:"path,omitempty"`
	// KeyData is the SSH public key certificate used to authenticate with the VM through ssh.
	// The key needs to be at least 2048-bit and in ssh-rsa format.
	KeyData string `json:"keyData,omitempty"`
}

// AzureNetworkProfile specifies the network interfaces of the virtual machine.
type AzureNetworkProfile struct {
	// NetworkInterfaces Deprecated: This field is currently not used and will be removed in later versions of the API.
	NetworkInterfaces AzureNetworkInterfaceReference `json:"networkInterfaces,omitempty"`
	// AcceleratedNetworking specifies whether the network interface is accelerated networking-enabled.
	AcceleratedNetworking *bool `json:"acceleratedNetworking,omitempty"`
}

// AzureNetworkInterfaceReference describes a network interface reference.
type AzureNetworkInterfaceReference struct {
	ID                                        string `json:"id,omitempty"`
	*AzureNetworkInterfaceReferenceProperties `json:"properties,omitempty"`
}

// AzureNetworkInterfaceReferenceProperties describes a network interface reference properties.
type AzureNetworkInterfaceReferenceProperties struct {
	Primary bool `json:"primary,omitempty"`
}

// AzureSubResource is the Sub Resource definition.
type AzureSubResource struct {
	// ID is the resource id.
	ID string `json:"id,omitempty"`
}

// AzureSubnetInfo is the information containing the subnet details.
type AzureSubnetInfo struct {
	// VnetName is the virtual network name. See [https://learn.microsoft.com/en-us/azure/virtual-network/virtual-networks-overview].
	VnetName string `json:"vnetName,omitempty"`
	// VnetResourceGroup is the resource group within which a virtual network is created. This is optional. If it is not specified then
	// AzureProviderSpec.ResourceGroup is used instead.
	VnetResourceGroup *string `json:"vnetResourceGroup,omitempty"`
	// SubnetName is the name of the subnet which is unique within a resource group.
	SubnetName string `json:"subnetName,omitempty"`
}

// AzureDiagnosticsProfile specifies boot diagnostic options
type AzureDiagnosticsProfile struct {
	// Enabled configures boot diagnostics to be stored or not
	Enabled bool `json:"enabled,omitempty"`
	// StorageURI is the URI of the storage account to use for storing console output and screenshot.
	// If not specified azure managed storage will be used.
	StorageURI *string `json:"storageURI,omitempty"`
}

// The (currently) supported values for the names of clouds to use in the CloudConfiguration.
const (
	CloudNameChina  string = "AzureChina"
	CloudNameGov    string = "AzureGovernment"
	CloudNamePublic string = "AzurePublic"
)

// CloudConfiguration contains detailed config for the cloud to connect to. Currently we only support selection of well-
// known Azure-instances by name, but this could be extended in future to support private clouds.
type CloudConfiguration struct {
	// Name is the name of the cloud to connect to, e.g. "AzurePublic" or "AzureChina".
	Name string `json:"name"`
}
