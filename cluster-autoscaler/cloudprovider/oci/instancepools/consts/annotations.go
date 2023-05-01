/*
Copyright 2021-2023 Oracle and/or its affiliates.
*/

package consts

import (
	v1 "k8s.io/api/core/v1"
	"time"
)

const (
	// OciUseWorkloadIdentityEnvVar is an env var that indicates whether to use the workload identity provider
	OciUseWorkloadIdentityEnvVar = "OCI_USE_WORKLOAD_IDENTITY"
	// OciUseInstancePrincipalEnvVar is an env var that indicates whether to use an instance principal
	OciUseInstancePrincipalEnvVar = "OCI_USE_INSTANCE_PRINCIPAL"
	// OciUseNonPoolMemberAnnotationEnvVar is an env var indicating that non-members of instance pools will get a special annotation
	OciUseNonPoolMemberAnnotationEnvVar = "OCI_USE_NON_POOL_MEMBER_ANNOTATION"
	// OciCompartmentEnvVar indicates to only use instance pools in this specific compartment
	OciCompartmentEnvVar = "OCI_COMPARTMENT_ID"
	// OciRegionEnvVar indicates to use a specific region
	OciRegionEnvVar = "OCI_REGION"
	// OciRefreshInterval indicates the rate at which the internal instance pool cache will be refreshed (by querying compute)
	OciRefreshInterval = "OCI_REFRESH_INTERVAL"
	// DefaultRefreshInterval is the default rate to refresh the cache
	DefaultRefreshInterval = 5 * time.Minute
	// ResourceGPU is the GPU resource type
	ResourceGPU v1.ResourceName = "nvidia.com/gpu"

	// OciAnnotationCompartmentID the well known annotation string for compartment ids
	OciAnnotationCompartmentID = "oci.oraclecloud.com/compartment-id"
	// OciInstanceIDAnnotation the well known annotation string for instance ids
	OciInstanceIDAnnotation = "oci.oraclecloud.com/instance-id"
	// InstanceIDLabelPrefix the prefix of the instance ocid
	InstanceIDLabelPrefix = "instance-id_prefix"
	// InstanceIDLabelSuffix the suffix of the instance ocid
	InstanceIDLabelSuffix = "instance-id_suffix"
	// OciInstancePoolIDAnnotation the well known annotation string for the instance pool ocid
	OciInstancePoolIDAnnotation = "oci.oraclecloud.com/instancepool-id"
	// InstancePoolIDLabelPrefix the prefix of the instance pool ocid
	InstancePoolIDLabelPrefix = "instancepool-id_prefix"
	// InstancePoolIDLabelSuffix the suffix of the instance pool ocid
	InstancePoolIDLabelSuffix = "instancepool-id_suffix"

	// OciInstancePoolResourceIdent resource identifier in the ocid
	OciInstancePoolResourceIdent = "instancepool"
	// OciInstancePoolLaunchOp is an instance pools operation type
	OciInstancePoolLaunchOp = "LaunchInstancesInPool"
	// InstanceStateUnfulfilled is a status indicating that the instance pool was unable to fulfill the operation
	InstanceStateUnfulfilled = "Unfulfilled"
	// InstanceIDUnfulfilled is the generic placeholder name for upcoming instances
	InstanceIDUnfulfilled = "instance_placeholder"

	// OciInstancePoolIDNonPoolMember indicates a kubernetes node doesn't belong to any OCI Instance Pool.
	OciInstancePoolIDNonPoolMember = "non_pool_member"
)
