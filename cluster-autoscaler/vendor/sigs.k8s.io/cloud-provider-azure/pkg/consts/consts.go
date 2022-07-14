/*
Copyright 2021 The Kubernetes Authors.

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

package consts

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2021-02-01/storage"

	"k8s.io/component-base/featuregate"
)

const (
	// VMTypeVMSS is the vmss vm type
	VMTypeVMSS = "vmss"
	// VMTypeStandard is the vmas vm type
	VMTypeStandard = "standard"

	// ExternalResourceGroupLabel is the label representing the node is in a different
	// resource group from other cloud provider components
	ExternalResourceGroupLabel = "kubernetes.azure.com/resource-group"
	// ManagedByAzureLabel is the label representing the node is managed by cloud provider azure
	ManagedByAzureLabel = "kubernetes.azure.com/managed"
	// NotManagedByAzureLabelValue is the label value representing the node is not managed by cloud provider azure
	NotManagedByAzureLabelValue = "false"

	// LabelFailureDomainBetaZone refer to https://github.com/kubernetes/api/blob/8519c5ea46199d57724725d5b969c5e8e0533692/core/v1/well_known_labels.go#L22-L23
	LabelFailureDomainBetaZone = "failure-domain.beta.kubernetes.io/zone"

	// LabelFailureDomainBetaRegion failure-domain region label
	LabelFailureDomainBetaRegion = "failure-domain.beta.kubernetes.io/region"

	// ADFSIdentitySystem is the override value for tenantID on Azure Stack clouds.
	ADFSIdentitySystem = "adfs"

	// AzureMetricsNamespace is the namespace of the azure metrics
	AzureMetricsNamespace = "cloudprovider_azure"

	// VhdContainerName is the vhd container name
	VhdContainerName = "vhds"
	// UseHTTPSForBlobBasedDisk determines if we use the https for the blob based disk
	UseHTTPSForBlobBasedDisk = true
	// BlobServiceName is the name of the blob service
	BlobServiceName = "blob"

	// MetadataCacheTTL is the TTL of the metadata service
	MetadataCacheTTL = time.Minute
	// MetadataCacheKey is the metadata cache key
	MetadataCacheKey = "InstanceMetadata"
	// MetadataURL is the metadata service endpoint
	MetadataURL = "http://169.254.169.254/metadata/instance"

	// DefaultDiskIOPSReadWrite is the default IOPS Caps & Throughput Cap (MBps)
	// per https://docs.microsoft.com/en-us/azure/virtual-machines/linux/disks-ultra-ssd
	DefaultDiskIOPSReadWrite = 500
	// DefaultDiskMBpsReadWrite is the default disk MBps read write
	DefaultDiskMBpsReadWrite = 100

	DiskEncryptionSetIDFormat = "/subscriptions/{subs-id}/resourceGroups/{rg-name}/providers/Microsoft.Compute/diskEncryptionSets/{diskEncryptionSet-name}"

	// IPv6DualStack is here to avoid having to import features pkg
	// and violate import rules
	IPv6DualStack featuregate.Feature = "IPv6DualStack"

	// MachineIDTemplate is the template of the virtual machine
	MachineIDTemplate = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/%s"
	// AvailabilitySetIDTemplate is the template of the availabilitySet ID
	AvailabilitySetIDTemplate = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/availabilitySets/%s"

	// NodeLabelRole specifies the role of a node
	NodeLabelRole = "kubernetes.io/role"
	// MasterNodeRoleLabel specifies is the master node label for a node
	MasterNodeRoleLabel = "node-role.kubernetes.io/master"
	// ControlPlaneNodeRoleLabel specifies is the control-plane node label for a node
	ControlPlaneNodeRoleLabel = "node-role.kubernetes.io/control-plane"

	// NicFailedState is the failed state of a nic
	NicFailedState = "Failed"

	// StorageAccountNameMaxLength is the max length of a storage name
	StorageAccountNameMaxLength = 24

	// DefaultStorageAccountType is the default storage account type
	DefaultStorageAccountType = string(storage.SkuNameStandardLRS)
	// DefaultStorageAccountKind is the default storage account kind
	DefaultStorageAccountKind = storage.KindStorageV2
	// FileShareAccountNamePrefix is the file share account name prefix
	FileShareAccountNamePrefix = "f"
	// SharedDiskAccountNamePrefix is the shared disk account name prefix
	SharedDiskAccountNamePrefix = "ds"
	// DedicatedDiskAccountNamePrefix is the dedicated disk account name prefix
	DedicatedDiskAccountNamePrefix = "dd"

	// RetryAfterHeaderKey is the retry-after header key in ARM responses.
	RetryAfterHeaderKey = "Retry-After"

	// StrRawVersion is the raw version string
	StrRawVersion string = "raw"

	// VirtualMachineScaleSetsDeallocating indicates VMSS instances are in Deallocating state.
	VirtualMachineScaleSetsDeallocating = "Deallocating"
	// VmssMachineIDTemplate is the vmss manchine ID template
	VmssMachineIDTemplate = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachineScaleSets/%s/virtualMachines/%s"
	// VMSetCIDRIPV4TagKey specifies the node ipv4 CIDR mask of the instances on the VMSS or VMAS
	VMSetCIDRIPV4TagKey = "kubernetesNodeCIDRMaskIPV4"
	// VMSetCIDRIPV6TagKey specifies the node ipv6 CIDR mask of the instances on the VMSS or VMAS
	VMSetCIDRIPV6TagKey = "kubernetesNodeCIDRMaskIPV6"

	// TagsDelimiter is the delimiter of tags
	TagsDelimiter = ","
	// TagKeyValueDelimiter is the delimiter between keys and values in tagas
	TagKeyValueDelimiter = "="
	// VMSetNamesSharingPrimarySLBDelimiter is the delimiter of vmSet names sharing the primary SLB
	VMSetNamesSharingPrimarySLBDelimiter = ","
)

// cache
const (
	// VMSSNameSeparator is the separator of the vmss names
	VMSSNameSeparator = "_"
	// VMSSKey is the key when querying vmss cache
	VMSSKey = "k8svmssKey"
	// VMASKey is the key when querying vmss cache
	VMASKey = "k8svmasKey"
	// AvailabilitySetNodesKey is the availability set nodes key
	AvailabilitySetNodesKey = "k8sAvailabilitySetNodesKey"
	// AvailabilitySetNodesCacheTTLDefaultInSeconds is the TTL of the availabilitySet node cache
	AvailabilitySetNodesCacheTTLDefaultInSeconds = 900
	// VMSSCacheTTLDefaultInSeconds is the TTL of the vmss cache
	VMSSCacheTTLDefaultInSeconds = 600
	// VMSSVirtualMachinesCacheTTLDefaultInSeconds is the TTL of the vmss vm cache
	VMSSVirtualMachinesCacheTTLDefaultInSeconds = 600
	// VMASCacheTTLDefaultInSeconds is the TTL of the vmas cache
	VMASCacheTTLDefaultInSeconds = 600

	// ZoneFetchingInterval defines the interval of performing zoneClient.GetZones
	ZoneFetchingInterval = 30 * time.Minute
)

// azure cloud config
const (
	// CloudProviderName is the value used for the --cloud-provider flag
	CloudProviderName = "azure"
	// AzureStackCloudName is the cloud name of Azure Stack
	AzureStackCloudName = "AZURESTACKCLOUD"
	// RateLimitQPSDefault is the default value of the rate limit qps
	RateLimitQPSDefault = 1.0
	// RateLimitBucketDefault is the default value of rate limit bucket
	RateLimitBucketDefault = 5
	// BackoffRetriesDefault is the default backoff retry count
	BackoffRetriesDefault = 6
	// BackoffExponentDefault is the default value of the backoff exponent
	BackoffExponentDefault = 1.5
	// BackoffDurationDefault is the default value of the backoff duration
	BackoffDurationDefault = 5 // in seconds
	// BackoffJitterDefault is the default value of the backoff jitter
	BackoffJitterDefault = 1.0
)

// load balancer
const (
	// PreConfiguredBackendPoolLoadBalancerTypesNone means that the load balancers are not pre-configured
	PreConfiguredBackendPoolLoadBalancerTypesNone = ""
	// PreConfiguredBackendPoolLoadBalancerTypesInternal means that the `internal` load balancers are pre-configured
	PreConfiguredBackendPoolLoadBalancerTypesInternal = "internal"
	// PreConfiguredBackendPoolLoadBalancerTypesExternal means that the `external` load balancers are pre-configured
	PreConfiguredBackendPoolLoadBalancerTypesExternal = "external"
	// PreConfiguredBackendPoolLoadBalancerTypesAll means that all load balancers are pre-configured
	PreConfiguredBackendPoolLoadBalancerTypesAll = "all"

	// MaximumLoadBalancerRuleCount is the maximum number of load balancer rules
	// ref: https://docs.microsoft.com/en-us/azure/azure-subscription-service-limits#load-balancer.
	MaximumLoadBalancerRuleCount = 250

	// LoadBalancerSkuBasic is the load balancer basic sku
	LoadBalancerSkuBasic = "basic"
	// LoadBalancerSkuStandard is the load balancer standard sku
	LoadBalancerSkuStandard = "standard"

	// ServiceAnnotationLoadBalancerInternal is the annotation used on the service
	ServiceAnnotationLoadBalancerInternal = "service.beta.kubernetes.io/azure-load-balancer-internal"

	// ServiceAnnotationLoadBalancerInternalSubnet is the annotation used on the service
	// to specify what subnet it is exposed on
	ServiceAnnotationLoadBalancerInternalSubnet = "service.beta.kubernetes.io/azure-load-balancer-internal-subnet"

	// ServiceAnnotationLoadBalancerMode is the annotation used on the service to specify
	// which load balancer should be associated with the service. This is valid when using the basic
	// load balancer or turn on the multiple standard load balancers mode, or it would be ignored.
	// 1. Default mode - service has no annotation ("service.beta.kubernetes.io/azure-load-balancer-mode")
	//	  In this case the Loadbalancer of the primary VMSS/VMAS is selected.
	// 2. "__auto__" mode - service is annotated with __auto__ value, this when loadbalancer from any VMSS/VMAS
	//    is selected which has the minimum rules associated with it.
	// 3. "name" mode - this is when the load balancer from the specified VMSS/VMAS that has the
	//    minimum rules associated with it is selected.
	ServiceAnnotationLoadBalancerMode = "service.beta.kubernetes.io/azure-load-balancer-mode"

	// ServiceAnnotationLoadBalancerAutoModeValue is the annotation used on the service to specify the
	// Azure load balancer auto selection from the availability sets
	ServiceAnnotationLoadBalancerAutoModeValue = "__auto__"

	// ServiceAnnotationDNSLabelName is the annotation used on the service
	// to specify the DNS label name for the service.
	ServiceAnnotationDNSLabelName = "service.beta.kubernetes.io/azure-dns-label-name"

	// ServiceAnnotationSharedSecurityRule is the annotation used on the service
	// to specify that the service should be exposed using an Azure security rule
	// that may be shared with other service, trading specificity of rules for an
	// increase in the number of services that can be exposed. This relies on the
	// Azure "augmented security rules" feature.
	ServiceAnnotationSharedSecurityRule = "service.beta.kubernetes.io/azure-shared-securityrule"

	// ServiceAnnotationLoadBalancerResourceGroup is the annotation used on the service
	// to specify the resource group of load balancer objects that are not in the same resource group as the cluster.
	ServiceAnnotationLoadBalancerResourceGroup = "service.beta.kubernetes.io/azure-load-balancer-resource-group"

	// ServiceAnnotationPIPName specifies the pip that will be applied to load balancer
	ServiceAnnotationPIPName = "service.beta.kubernetes.io/azure-pip-name"

	// ServiceAnnotationIPTagsForPublicIP specifies the iptags used when dynamically creating a public ip
	ServiceAnnotationIPTagsForPublicIP = "service.beta.kubernetes.io/azure-pip-ip-tags"

	// ServiceAnnotationAllowedServiceTag is the annotation used on the service
	// to specify a list of allowed service tags separated by comma
	// Refer https://docs.microsoft.com/en-us/azure/virtual-network/security-overview#service-tags for all supported service tags.
	ServiceAnnotationAllowedServiceTag = "service.beta.kubernetes.io/azure-allowed-service-tags"

	// ServiceAnnotationDenyAllExceptLoadBalancerSourceRanges  denies all traffic to the load balancer except those
	// within the service.Spec.LoadBalancerSourceRanges. Ref: https://github.com/kubernetes-sigs/cloud-provider-azure/issues/374.
	ServiceAnnotationDenyAllExceptLoadBalancerSourceRanges = "service.beta.kubernetes.io/azure-deny-all-except-load-balancer-source-ranges"

	// ServiceAnnotationLoadBalancerIdleTimeout is the annotation used on the service
	// to specify the idle timeout for connections on the load balancer in minutes.
	ServiceAnnotationLoadBalancerIdleTimeout = "service.beta.kubernetes.io/azure-load-balancer-tcp-idle-timeout"

	// ServiceAnnotationLoadBalancerEnableHighAvailabilityPorts is the annotation used on the service
	// to enable the high availability ports on the standard internal load balancer.
	ServiceAnnotationLoadBalancerEnableHighAvailabilityPorts = "service.beta.kubernetes.io/azure-load-balancer-enable-high-availability-ports"

	// ServiceAnnotationLoadBalancerHealthProbeProtocol determines the network protocol that the load balancer health probe use.
	// If not set, the local service would use the HTTP and the cluster service would use the TCP by default.
	ServiceAnnotationLoadBalancerHealthProbeProtocol = "service.beta.kubernetes.io/azure-load-balancer-health-probe-protocol"

	// ServiceAnnotationLoadBalancerHealthProbeRequestPath determines the request path of the load balancer health probe.
	// This is only useful for the HTTP and HTTPS, and would be ignored when using TCP. If not set,
	// `/healthz` would be configured by default.
	ServiceAnnotationLoadBalancerHealthProbeRequestPath = "service.beta.kubernetes.io/azure-load-balancer-health-probe-request-path"

	// ServiceAnnotationAzurePIPTags determines what tags should be applied to the public IP of the service. The cluster name
	// and service names tags (which is managed by controller manager itself) would keep unchanged. The supported format
	// is `a=b,c=d,...`. After updated, the old user-assigned tags would not be replaced by the new ones.
	ServiceAnnotationAzurePIPTags = "service.beta.kubernetes.io/azure-pip-tags"

	// ServiceAnnotationAzurePIPTags sets the additional Public IPs (split by comma) besides the service's Public IP configured on LoadBalancer.
	// These additional Public IPs would be consumed by kube-proxy to configure the iptables rules on each node. Note they would not be configured
	// automatically on Azure LoadBalancer. Instead, they need to be configured manually (e.g. on Azure cross-region LoadBalancer by another operator).
	ServiceAnnotationAdditionalPublicIPs = "service.beta.kubernetes.io/azure-additional-public-ips"

	// ServiceTagKey is the service key applied for public IP tags.
	ServiceTagKey = "service"
	// ClusterNameKey is the cluster name key applied for public IP tags.
	ClusterNameKey = "kubernetes-cluster-name"
	// ServiceUsingDNSKey is the service name consuming the DNS label on the public IP
	ServiceUsingDNSKey = "kubernetes-dns-label-service"

	// DefaultLoadBalancerSourceRanges is the default value of the load balancer source ranges
	DefaultLoadBalancerSourceRanges = "0.0.0.0/0"

	// TrueAnnotationValue is the true annotation value
	TrueAnnotationValue = "true"

	// LoadBalancerMinimumPriority is the minimum priority
	LoadBalancerMinimumPriority = 500
	// LoadBalancerMaximumPriority is the maximum priority
	LoadBalancerMaximumPriority = 4096

	// FrontendIPConfigIDTemplate is the template of the frontend IP configuration
	FrontendIPConfigIDTemplate = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/loadBalancers/%s/frontendIPConfigurations/%s"
	// BackendPoolIDTemplate is the template of the backend pool
	BackendPoolIDTemplate = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/loadBalancers/%s/backendAddressPools/%s"
	// LoadBalancerProbeIDTemplate is the template of the load balancer probe
	LoadBalancerProbeIDTemplate = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Network/loadBalancers/%s/probes/%s"

	// InternalLoadBalancerNameSuffix is load balancer suffix
	InternalLoadBalancerNameSuffix = "-internal"

	// FrontendIPConfigNameMaxLength is the max length of the frontend IP configuration
	FrontendIPConfigNameMaxLength = 80
	// LoadBalancerRuleNameMaxLength is the max length of the load balancing rule
	LoadBalancerRuleNameMaxLength = 80
)

// error messages
const (
	// VmssVMNotActiveErrorMessage not active means the instance is under deleting from Azure VMSS.
	VmssVMNotActiveErrorMessage = "not an active Virtual Machine Scale Set VM instanceId"
	// OperationCanceledErrorMessage means the operation is canceled by another new operation.
	OperationCanceledErrorMessage = "canceledandsupersededduetoanotheroperation"
	// CannotDeletePublicIPErrorMessageCode means the public IP cannot be deleted
	CannotDeletePublicIPErrorMessageCode = "PublicIPAddressCannotBeDeleted"
	// ReferencedResourceNotProvisionedMessageCode means the referenced resource has not been provisioned
	ReferencedResourceNotProvisionedMessageCode = "ReferencedResourceNotProvisioned"
)

// node ipam controller
const (
	// DefaultNodeMaskCIDRIPv4 is default mask size for IPv4 node cidr
	DefaultNodeMaskCIDRIPv4 = 24
	// DefaultNodeMaskCIDRIPv6 is default mask size for IPv6 node cidr
	DefaultNodeMaskCIDRIPv6 = 64
	// DefaultNodeCIDRMaskSize is the default mask size for node cidr
	DefaultNodeCIDRMaskSize = 24
)

// metadata service
const (
	// ImdsInstanceAPIVersion is the imds instance api version
	ImdsInstanceAPIVersion = "2019-03-11"
	// ImdsLoadBalancerAPIVersion is the imds load balancer api version
	ImdsLoadBalancerAPIVersion = "2020-10-01"
	// ImdsServer is the imds server endpoint
	ImdsServer = "http://169.254.169.254"
	// ImdsInstanceURI is the imds instance uri
	ImdsInstanceURI = "/metadata/instance"
	// ImdsLoadBalancerURI is the imds load balancer uri
	ImdsLoadBalancerURI = "/metadata/loadbalancer"
)

// routes
const (
	RouteNameFmt       = "%s____%s"
	RouteNameSeparator = "____"
)

// cloud provider config secret
const (
	DefaultCloudProviderConfigSecName      = "azure-cloud-provider"
	DefaultCloudProviderConfigSecNamespace = "kube-system"
	DefaultCloudProviderConfigSecKey       = "cloud-config"
)

// RateLimited error string
const RateLimited = "rate limited"
