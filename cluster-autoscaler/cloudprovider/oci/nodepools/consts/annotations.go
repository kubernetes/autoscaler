/*
Copyright 2020-2023 Oracle and/or its affiliates.
*/

package consts

import "time"

const (
	// ArmArch is the Arm Architecture string used by kubernetes
	ArmArch string = "arm64"

	// OkeUseInstancePrincipalEnvVar is deprecated and should not be used anymore. There's a matching OCI env var,
	// but we keep this one around for backwards compatibility
	OkeUseInstancePrincipalEnvVar = "OKE_USE_INSTANCE_PRINCIPAL"
	// OkeHostOverrideEnvVar is a hidden flag that allows CA to hit another containerengine endpoint
	OkeHostOverrideEnvVar = "OKE_HOST_OVERRIDE"

	// DefaultRefreshInterval is the interval to refresh the node pool instances information
	DefaultRefreshInterval = 1 * time.Minute

	// OciNodePoolResourceIdent is the string identifier in the ocid that indicates the resource is a node pool
	OciNodePoolResourceIdent = "nodepool"

	// ToBeDeletedByClusterAutoscaler is the taint used to ensure that after a node has been called to be deleted
	// no more pods will schedule onto it
	ToBeDeletedByClusterAutoscaler = "ignore-taint.cluster-autoscaler.kubernetes.io/oke-impending-node-termination"

	// EphemeralStorageSize is the freeform tag key that would be used to determine the ephemeral-storage size of the node
	EphemeralStorageSize = "cluster-autoscaler/node-ephemeral-storage"
)
