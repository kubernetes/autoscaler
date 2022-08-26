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

// Package oapi provides primitives to interact with the openapi HTTP API.
package oapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider/exoscale/internal/github.com/deepmap/oapi-codegen/pkg/runtime"
)

// Defines values for AccessKeyType.
const (
	AccessKeyTypeRestricted AccessKeyType = "restricted"

	AccessKeyTypeUnrestricted AccessKeyType = "unrestricted"
)

// Defines values for AccessKeyVersion.
const (
	AccessKeyVersionV1 AccessKeyVersion = "v1"

	AccessKeyVersionV2 AccessKeyVersion = "v2"
)

// Defines values for AccessKeyResourceDomain.
const (
	AccessKeyResourceDomainPartner AccessKeyResourceDomain = "partner"

	AccessKeyResourceDomainSos AccessKeyResourceDomain = "sos"
)

// Defines values for AccessKeyResourceResourceType.
const (
	AccessKeyResourceResourceTypeBucket AccessKeyResourceResourceType = "bucket"

	AccessKeyResourceResourceTypeProduct AccessKeyResourceResourceType = "product"
)

// Defines values for DbaasNodeStateRole.
const (
	DbaasNodeStateRoleMaster DbaasNodeStateRole = "master"

	DbaasNodeStateRoleReadReplica DbaasNodeStateRole = "read-replica"

	DbaasNodeStateRoleStandby DbaasNodeStateRole = "standby"
)

// Defines values for DbaasNodeStateState.
const (
	DbaasNodeStateStateLeaving DbaasNodeStateState = "leaving"

	DbaasNodeStateStateRunning DbaasNodeStateState = "running"

	DbaasNodeStateStateSettingUpVm DbaasNodeStateState = "setting_up_vm"

	DbaasNodeStateStateSyncingData DbaasNodeStateState = "syncing_data"

	DbaasNodeStateStateUnknown DbaasNodeStateState = "unknown"
)

// Defines values for DbaasNodeStateProgressUpdatePhase.
const (
	DbaasNodeStateProgressUpdatePhaseBasebackup DbaasNodeStateProgressUpdatePhase = "basebackup"

	DbaasNodeStateProgressUpdatePhaseFinalize DbaasNodeStateProgressUpdatePhase = "finalize"

	DbaasNodeStateProgressUpdatePhasePrepare DbaasNodeStateProgressUpdatePhase = "prepare"

	DbaasNodeStateProgressUpdatePhaseStream DbaasNodeStateProgressUpdatePhase = "stream"
)

// Defines values for DbaasServiceMaintenanceDow.
const (
	DbaasServiceMaintenanceDowFriday DbaasServiceMaintenanceDow = "friday"

	DbaasServiceMaintenanceDowMonday DbaasServiceMaintenanceDow = "monday"

	DbaasServiceMaintenanceDowNever DbaasServiceMaintenanceDow = "never"

	DbaasServiceMaintenanceDowSaturday DbaasServiceMaintenanceDow = "saturday"

	DbaasServiceMaintenanceDowSunday DbaasServiceMaintenanceDow = "sunday"

	DbaasServiceMaintenanceDowThursday DbaasServiceMaintenanceDow = "thursday"

	DbaasServiceMaintenanceDowTuesday DbaasServiceMaintenanceDow = "tuesday"

	DbaasServiceMaintenanceDowWednesday DbaasServiceMaintenanceDow = "wednesday"
)

// Defines values for DbaasServiceNotificationLevel.
const (
	DbaasServiceNotificationLevelNotice DbaasServiceNotificationLevel = "notice"

	DbaasServiceNotificationLevelWarning DbaasServiceNotificationLevel = "warning"
)

// Defines values for DbaasServiceNotificationType.
const (
	DbaasServiceNotificationTypeServiceEndOfLife DbaasServiceNotificationType = "service_end_of_life"

	DbaasServiceNotificationTypeServicePoweredOffRemoval DbaasServiceNotificationType = "service_powered_off_removal"
)

// Defines values for DbaasServiceOpensearchIndexPatternsSortingAlgorithm.
const (
	DbaasServiceOpensearchIndexPatternsSortingAlgorithmAlphabetical DbaasServiceOpensearchIndexPatternsSortingAlgorithm = "alphabetical"

	DbaasServiceOpensearchIndexPatternsSortingAlgorithmCreationDate DbaasServiceOpensearchIndexPatternsSortingAlgorithm = "creation_date"
)

// Defines values for DeployTargetType.
const (
	DeployTargetTypeDedicated DeployTargetType = "dedicated"

	DeployTargetTypeEdge DeployTargetType = "edge"
)

// Defines values for DnsDomainRecordType.
const (
	DnsDomainRecordTypeNs DnsDomainRecordType = "ns"

	DnsDomainRecordTypeSoa DnsDomainRecordType = "soa"
)

// Defines values for ElasticIpHealthcheckMode.
const (
	ElasticIpHealthcheckModeHttp ElasticIpHealthcheckMode = "http"

	ElasticIpHealthcheckModeHttps ElasticIpHealthcheckMode = "https"

	ElasticIpHealthcheckModeTcp ElasticIpHealthcheckMode = "tcp"
)

// Defines values for EnumComponentRoute.
const (
	EnumComponentRouteDynamic EnumComponentRoute = "dynamic"

	EnumComponentRoutePrivate EnumComponentRoute = "private"

	EnumComponentRoutePrivatelink EnumComponentRoute = "privatelink"

	EnumComponentRoutePublic EnumComponentRoute = "public"
)

// Defines values for EnumComponentUsage.
const (
	EnumComponentUsagePrimary EnumComponentUsage = "primary"

	EnumComponentUsageReplica EnumComponentUsage = "replica"
)

// Defines values for EnumIntegrationTypes.
const (
	EnumIntegrationTypesDatasource EnumIntegrationTypes = "datasource"

	EnumIntegrationTypesMetrics EnumIntegrationTypes = "metrics"

	EnumIntegrationTypesReadReplica EnumIntegrationTypes = "read_replica"
)

// Defines values for EnumKafkaAclPermissions.
const (
	EnumKafkaAclPermissionsAdmin EnumKafkaAclPermissions = "admin"

	EnumKafkaAclPermissionsRead EnumKafkaAclPermissions = "read"

	EnumKafkaAclPermissionsReadwrite EnumKafkaAclPermissions = "readwrite"

	EnumKafkaAclPermissionsWrite EnumKafkaAclPermissions = "write"
)

// Defines values for EnumKafkaAuthMethod.
const (
	EnumKafkaAuthMethodCertificate EnumKafkaAuthMethod = "certificate"

	EnumKafkaAuthMethodSasl EnumKafkaAuthMethod = "sasl"
)

// Defines values for EnumMasterLinkStatus.
const (
	EnumMasterLinkStatusDown EnumMasterLinkStatus = "down"

	EnumMasterLinkStatusUp EnumMasterLinkStatus = "up"
)

// Defines values for EnumMigrationStatus.
const (
	EnumMigrationStatusDone EnumMigrationStatus = "done"

	EnumMigrationStatusFailed EnumMigrationStatus = "failed"

	EnumMigrationStatusRunning EnumMigrationStatus = "running"

	EnumMigrationStatusSyncing EnumMigrationStatus = "syncing"
)

// Defines values for EnumPgMigrationMethod.
const (
	EnumPgMigrationMethodDump EnumPgMigrationMethod = "dump"

	EnumPgMigrationMethodReplication EnumPgMigrationMethod = "replication"
)

// Defines values for EnumPgPoolMode.
const (
	EnumPgPoolModeSession EnumPgPoolMode = "session"

	EnumPgPoolModeStatement EnumPgPoolMode = "statement"

	EnumPgPoolModeTransaction EnumPgPoolMode = "transaction"
)

// Defines values for EnumPgSynchronousReplication.
const (
	EnumPgSynchronousReplicationOff EnumPgSynchronousReplication = "off"

	EnumPgSynchronousReplicationQuorum EnumPgSynchronousReplication = "quorum"
)

// Defines values for EnumPgVariant.
const (
	EnumPgVariantAiven EnumPgVariant = "aiven"

	EnumPgVariantTimescale EnumPgVariant = "timescale"
)

// Defines values for EnumServiceState.
const (
	EnumServiceStatePoweroff EnumServiceState = "poweroff"

	EnumServiceStateRebalancing EnumServiceState = "rebalancing"

	EnumServiceStateRebuilding EnumServiceState = "rebuilding"

	EnumServiceStateRunning EnumServiceState = "running"
)

// Defines values for EnumSortOrder.
const (
	EnumSortOrderAsc EnumSortOrder = "asc"

	EnumSortOrderDesc EnumSortOrder = "desc"
)

// Defines values for InstanceState.
const (
	InstanceStateDestroyed InstanceState = "destroyed"

	InstanceStateDestroying InstanceState = "destroying"

	InstanceStateError InstanceState = "error"

	InstanceStateExpunging InstanceState = "expunging"

	InstanceStateMigrating InstanceState = "migrating"

	InstanceStateRunning InstanceState = "running"

	InstanceStateStarting InstanceState = "starting"

	InstanceStateStopped InstanceState = "stopped"

	InstanceStateStopping InstanceState = "stopping"
)

// Defines values for InstancePoolState.
const (
	InstancePoolStateCreating InstancePoolState = "creating"

	InstancePoolStateDestroying InstancePoolState = "destroying"

	InstancePoolStateRunning InstancePoolState = "running"

	InstancePoolStateScalingDown InstancePoolState = "scaling-down"

	InstancePoolStateScalingUp InstancePoolState = "scaling-up"

	InstancePoolStateSuspended InstancePoolState = "suspended"
)

// Defines values for InstanceTypeFamily.
const (
	InstanceTypeFamilyColossus InstanceTypeFamily = "colossus"

	InstanceTypeFamilyCpu InstanceTypeFamily = "cpu"

	InstanceTypeFamilyGpu InstanceTypeFamily = "gpu"

	InstanceTypeFamilyGpu2 InstanceTypeFamily = "gpu2"

	InstanceTypeFamilyMemory InstanceTypeFamily = "memory"

	InstanceTypeFamilyStandard InstanceTypeFamily = "standard"

	InstanceTypeFamilyStorage InstanceTypeFamily = "storage"
)

// Defines values for InstanceTypeSize.
const (
	InstanceTypeSizeColossus InstanceTypeSize = "colossus"

	InstanceTypeSizeExtraLarge InstanceTypeSize = "extra-large"

	InstanceTypeSizeHuge InstanceTypeSize = "huge"

	InstanceTypeSizeJumbo InstanceTypeSize = "jumbo"

	InstanceTypeSizeLarge InstanceTypeSize = "large"

	InstanceTypeSizeMedium InstanceTypeSize = "medium"

	InstanceTypeSizeMega InstanceTypeSize = "mega"

	InstanceTypeSizeMicro InstanceTypeSize = "micro"

	InstanceTypeSizeSmall InstanceTypeSize = "small"

	InstanceTypeSizeTiny InstanceTypeSize = "tiny"

	InstanceTypeSizeTitan InstanceTypeSize = "titan"
)

// Defines values for LoadBalancerState.
const (
	LoadBalancerStateCreating LoadBalancerState = "creating"

	LoadBalancerStateDeleting LoadBalancerState = "deleting"

	LoadBalancerStateError LoadBalancerState = "error"

	LoadBalancerStateRunning LoadBalancerState = "running"
)

// Defines values for LoadBalancerServerStatusStatus.
const (
	LoadBalancerServerStatusStatusFailure LoadBalancerServerStatusStatus = "failure"

	LoadBalancerServerStatusStatusSuccess LoadBalancerServerStatusStatus = "success"
)

// Defines values for LoadBalancerServiceProtocol.
const (
	LoadBalancerServiceProtocolTcp LoadBalancerServiceProtocol = "tcp"

	LoadBalancerServiceProtocolUdp LoadBalancerServiceProtocol = "udp"
)

// Defines values for LoadBalancerServiceState.
const (
	LoadBalancerServiceStateCreating LoadBalancerServiceState = "creating"

	LoadBalancerServiceStateDeleting LoadBalancerServiceState = "deleting"

	LoadBalancerServiceStateError LoadBalancerServiceState = "error"

	LoadBalancerServiceStateRunning LoadBalancerServiceState = "running"

	LoadBalancerServiceStateUpdating LoadBalancerServiceState = "updating"
)

// Defines values for LoadBalancerServiceStrategy.
const (
	LoadBalancerServiceStrategyRoundRobin LoadBalancerServiceStrategy = "round-robin"

	LoadBalancerServiceStrategySourceHash LoadBalancerServiceStrategy = "source-hash"
)

// Defines values for LoadBalancerServiceHealthcheckMode.
const (
	LoadBalancerServiceHealthcheckModeHttp LoadBalancerServiceHealthcheckMode = "http"

	LoadBalancerServiceHealthcheckModeHttps LoadBalancerServiceHealthcheckMode = "https"

	LoadBalancerServiceHealthcheckModeTcp LoadBalancerServiceHealthcheckMode = "tcp"
)

// Defines values for ManagerType.
const (
	ManagerTypeInstancePool ManagerType = "instance-pool"

	ManagerTypeSksNodepool ManagerType = "sks-nodepool"
)

// Defines values for OperationReason.
const (
	OperationReasonBusy OperationReason = "busy"

	OperationReasonConflict OperationReason = "conflict"

	OperationReasonFault OperationReason = "fault"

	OperationReasonForbidden OperationReason = "forbidden"

	OperationReasonIncorrect OperationReason = "incorrect"

	OperationReasonInterrupted OperationReason = "interrupted"

	OperationReasonNotFound OperationReason = "not-found"

	OperationReasonPartial OperationReason = "partial"

	OperationReasonUnavailable OperationReason = "unavailable"

	OperationReasonUnknown OperationReason = "unknown"

	OperationReasonUnsupported OperationReason = "unsupported"
)

// Defines values for OperationState.
const (
	OperationStateFailure OperationState = "failure"

	OperationStatePending OperationState = "pending"

	OperationStateSuccess OperationState = "success"

	OperationStateTimeout OperationState = "timeout"
)

// Defines values for SecurityGroupRuleFlowDirection.
const (
	SecurityGroupRuleFlowDirectionEgress SecurityGroupRuleFlowDirection = "egress"

	SecurityGroupRuleFlowDirectionIngress SecurityGroupRuleFlowDirection = "ingress"
)

// Defines values for SecurityGroupRuleProtocol.
const (
	SecurityGroupRuleProtocolAh SecurityGroupRuleProtocol = "ah"

	SecurityGroupRuleProtocolEsp SecurityGroupRuleProtocol = "esp"

	SecurityGroupRuleProtocolGre SecurityGroupRuleProtocol = "gre"

	SecurityGroupRuleProtocolIcmp SecurityGroupRuleProtocol = "icmp"

	SecurityGroupRuleProtocolIcmpv6 SecurityGroupRuleProtocol = "icmpv6"

	SecurityGroupRuleProtocolIpip SecurityGroupRuleProtocol = "ipip"

	SecurityGroupRuleProtocolTcp SecurityGroupRuleProtocol = "tcp"

	SecurityGroupRuleProtocolUdp SecurityGroupRuleProtocol = "udp"
)

// Defines values for SksClusterAddons.
const (
	SksClusterAddonsExoscaleCloudController SksClusterAddons = "exoscale-cloud-controller"

	SksClusterAddonsMetricsServer SksClusterAddons = "metrics-server"
)

// Defines values for SksClusterCni.
const (
	SksClusterCniCalico SksClusterCni = "calico"
)

// Defines values for SksClusterLevel.
const (
	SksClusterLevelPro SksClusterLevel = "pro"

	SksClusterLevelStarter SksClusterLevel = "starter"
)

// Defines values for SksClusterState.
const (
	SksClusterStateCreating SksClusterState = "creating"

	SksClusterStateDeleting SksClusterState = "deleting"

	SksClusterStateError SksClusterState = "error"

	SksClusterStateRotatingCcmCredentials SksClusterState = "rotating-ccm-credentials"

	SksClusterStateRunning SksClusterState = "running"

	SksClusterStateSuspending SksClusterState = "suspending"

	SksClusterStateUpdating SksClusterState = "updating"

	SksClusterStateUpgrading SksClusterState = "upgrading"
)

// Defines values for SksNodepoolAddons.
const (
	SksNodepoolAddonsLinbit SksNodepoolAddons = "linbit"
)

// Defines values for SksNodepoolState.
const (
	SksNodepoolStateCreating SksNodepoolState = "creating"

	SksNodepoolStateDeleting SksNodepoolState = "deleting"

	SksNodepoolStateError SksNodepoolState = "error"

	SksNodepoolStateRenewingToken SksNodepoolState = "renewing-token"

	SksNodepoolStateRunning SksNodepoolState = "running"

	SksNodepoolStateUpdating SksNodepoolState = "updating"
)

// Defines values for SksNodepoolTaintEffect.
const (
	SksNodepoolTaintEffectNoExecute SksNodepoolTaintEffect = "NoExecute"

	SksNodepoolTaintEffectNoSchedule SksNodepoolTaintEffect = "NoSchedule"

	SksNodepoolTaintEffectPreferNoSchedule SksNodepoolTaintEffect = "PreferNoSchedule"
)

// Defines values for SnapshotState.
const (
	SnapshotStateDeleted SnapshotState = "deleted"

	SnapshotStateDeleting SnapshotState = "deleting"

	SnapshotStateError SnapshotState = "error"

	SnapshotStateExported SnapshotState = "exported"

	SnapshotStateExporting SnapshotState = "exporting"

	SnapshotStateReady SnapshotState = "ready"

	SnapshotStateSnapshotting SnapshotState = "snapshotting"
)

// Defines values for TemplateBootMode.
const (
	TemplateBootModeLegacy TemplateBootMode = "legacy"

	TemplateBootModeUefi TemplateBootMode = "uefi"
)

// Defines values for TemplateVisibility.
const (
	TemplateVisibilityPrivate TemplateVisibility = "private"

	TemplateVisibilityPublic TemplateVisibility = "public"
)

// Defines values for ZoneName.
const (
	ZoneNameAtVie1 ZoneName = "at-vie-1"

	ZoneNameBgSof1 ZoneName = "bg-sof-1"

	ZoneNameChDk2 ZoneName = "ch-dk-2"

	ZoneNameChGva2 ZoneName = "ch-gva-2"

	ZoneNameChZrh1 ZoneName = "ch-zrh-1"

	ZoneNameDeFra1 ZoneName = "de-fra-1"

	ZoneNameDeMuc1 ZoneName = "de-muc-1"
)

// IAM Access Key
type AccessKey struct {
	// IAM Access Key
	Key *string `json:"key,omitempty"`

	// IAM Access Key name
	Name *string `json:"name,omitempty"`

	// IAM Access Key operations
	Operations *[]string `json:"operations,omitempty"`

	// IAM Access Key Resources
	Resources *[]AccessKeyResource `json:"resources,omitempty"`

	// IAM Access Key Secret
	Secret *string `json:"secret,omitempty"`

	// IAM Access Key tags
	Tags *[]string `json:"tags,omitempty"`

	// IAM Access Key type
	Type *AccessKeyType `json:"type,omitempty"`

	// IAM Access Key version
	Version *AccessKeyVersion `json:"version,omitempty"`
}

// IAM Access Key type
type AccessKeyType string

// IAM Access Key version
type AccessKeyVersion string

// Access key operation
type AccessKeyOperation struct {
	// Name of the operation
	Operation *string `json:"operation,omitempty"`

	// Tags associated with the operation
	Tags *[]string `json:"tags,omitempty"`
}

// Access key resource
type AccessKeyResource struct {
	// Resource domain
	Domain *AccessKeyResourceDomain `json:"domain,omitempty"`

	// Resource name
	ResourceName *string `json:"resource-name,omitempty"`

	// Resource type
	ResourceType *AccessKeyResourceResourceType `json:"resource-type,omitempty"`
}

// Resource domain
type AccessKeyResourceDomain string

// Resource type
type AccessKeyResourceResourceType string

// Anti-affinity Group
type AntiAffinityGroup struct {
	// Anti-affinity Group description
	Description *string `json:"description,omitempty"`

	// Anti-affinity Group ID
	Id *string `json:"id,omitempty"`

	// Anti-affinity Group instances
	Instances *[]Instance `json:"instances,omitempty"`

	// Anti-affinity Group name
	Name *string `json:"name,omitempty"`
}

// DBaaS plan backup config
type DbaasBackupConfig struct {
	// Interval of taking a frequent backup in service types supporting different backup schedules
	FrequentIntervalMinutes *int64 `json:"frequent-interval-minutes,omitempty"`

	// Maximum age of the oldest frequent backup in service types supporting different backup schedules
	FrequentOldestAgeMinutes *int64 `json:"frequent-oldest-age-minutes,omitempty"`

	// Interval of taking a frequent backup in service types supporting different backup schedules
	InfrequentIntervalMinutes *int64 `json:"infrequent-interval-minutes,omitempty"`

	// Maximum age of the oldest infrequent backup in service types supporting different backup schedules
	InfrequentOldestAgeMinutes *int64 `json:"infrequent-oldest-age-minutes,omitempty"`

	// The interval, in hours, at which backups are generated.
	//                                             For some services, like PostgreSQL, this is the interval
	//                                             at which full snapshots are taken and continuous incremental
	//                                             backup stream is maintained in addition to that.
	Interval *int64 `json:"interval,omitempty"`

	// Maximum number of backups to keep. Zero when no backups are created.
	MaxCount *int64 `json:"max-count,omitempty"`

	// Mechanism how backups can be restored. 'regular'
	//                                             means a backup is restored as is so that the system
	//                                             is restored to the state it was when the backup was generated.
	//                                             'pitr' means point-in-time-recovery, which allows restoring
	//                                             the system to any state since the first available full snapshot.
	RecoveryMode *string `json:"recovery-mode,omitempty"`
}

// DbaasIntegration defines model for dbaas-integration.
type DbaasIntegration struct {
	// Description of the integration
	Description *string `json:"description,omitempty"`

	// Destination service name
	Dest *string `json:"dest,omitempty"`

	// Integration id
	Id *string `json:"id,omitempty"`

	// Whether the integration is active or not
	IsActive *bool `json:"is-active,omitempty"`

	// Whether the integration is enabled or not
	IsEnabled *bool `json:"is-enabled,omitempty"`

	// Integration settings
	Settings *map[string]interface{} `json:"settings,omitempty"`

	// Source service name
	Source *string `json:"source,omitempty"`

	// Integration status
	Status *string `json:"status,omitempty"`

	// Integration type
	Type *string `json:"type,omitempty"`
}

// DbaasMigrationStatus defines model for dbaas-migration-status.
type DbaasMigrationStatus struct {
	// Migration status per database
	Details *[]struct {
		// Migrated db name (PG) or number (Redis)
		Dbname *string `json:"dbname,omitempty"`

		// Error message in case that migration has failed
		Error *string `json:"error,omitempty"`

		// Migration method
		Method *string              `json:"method,omitempty"`
		Status *EnumMigrationStatus `json:"status,omitempty"`
	} `json:"details,omitempty"`

	// Error message in case that migration has failed
	Error *string `json:"error,omitempty"`

	// Redis only: how many seconds since last I/O with redis master
	MasterLastIoSecondsAgo *int64                `json:"master-last-io-seconds-ago,omitempty"`
	MasterLinkStatus       *EnumMasterLinkStatus `json:"master-link-status,omitempty"`

	// Migration method. Empty in case of multiple methods or error
	Method *string `json:"method,omitempty"`

	// Migration status
	Status *string `json:"status,omitempty"`
}

// Automatic maintenance settings
type DbaasNodeState struct {
	// Name of the service node
	Name string `json:"name"`

	// Extra information regarding the progress for current state
	ProgressUpdates *[]DbaasNodeStateProgressUpdate `json:"progress-updates,omitempty"`

	// Role of this node. Only returned for a subset of service types
	Role *DbaasNodeStateRole `json:"role,omitempty"`

	// Current state of the service node
	State DbaasNodeStateState `json:"state"`
}

// Role of this node. Only returned for a subset of service types
type DbaasNodeStateRole string

// Current state of the service node
type DbaasNodeStateState string

// Extra information regarding the progress for current state
type DbaasNodeStateProgressUpdate struct {
	// Indicates whether this phase has been completed or not
	Completed bool `json:"completed"`

	// Current progress for this phase. May be missing or null.
	Current *int64 `json:"current,omitempty"`

	// Maximum progress value for this phase. May be missing or null. May change.
	Max *int64 `json:"max,omitempty"`

	// Minimum progress value for this phase. May be missing or null.
	Min *int64 `json:"min,omitempty"`

	// Key identifying this phase
	Phase DbaasNodeStateProgressUpdatePhase `json:"phase"`

	// Unit for current/min/max values. New units may be added.
	//                         If null should be treated as generic unit
	Unit *string `json:"unit,omitempty"`
}

// Key identifying this phase
type DbaasNodeStateProgressUpdatePhase string

// DBaaS plan
type DbaasPlan struct {
	// Requires authorization or publicly available
	Authorized *bool `json:"authorized,omitempty"`

	// DBaaS plan backup config
	BackupConfig *DbaasBackupConfig `json:"backup-config,omitempty"`

	// DBaaS plan disk space
	DiskSpace *int64 `json:"disk-space,omitempty"`

	// DBaaS plan max memory allocated percentage
	MaxMemoryPercent *int64 `json:"max-memory-percent,omitempty"`

	// DBaaS plan name
	Name *string `json:"name,omitempty"`

	// DBaaS plan node count
	NodeCount *int64 `json:"node-count,omitempty"`

	// DBaaS plan CPU count per node
	NodeCpuCount *int64 `json:"node-cpu-count,omitempty"`

	// DBaaS plan memory count per node
	NodeMemory *int64 `json:"node-memory,omitempty"`
}

// List of backups for the service
type DbaasServiceBackup struct {
	// Internal name of this backup
	BackupName string `json:"backup-name"`

	// Backup timestamp (ISO 8601)
	BackupTime time.Time `json:"backup-time"`

	// Backup's original size before compression
	DataSize int64 `json:"data-size"`
}

// DbaasServiceCommon defines model for dbaas-service-common.
type DbaasServiceCommon struct {
	// Service creation timestamp (ISO 8601)
	CreatedAt *time.Time `json:"created-at,omitempty"`

	// TODO UNIT disk space for data storage
	DiskSize *int64 `json:"disk-size,omitempty"`

	// Service integrations
	Integrations *[]DbaasIntegration `json:"integrations,omitempty"`
	Name         DbaasServiceName    `json:"name"`

	// Number of service nodes in the active plan
	NodeCount *int64 `json:"node-count,omitempty"`

	// Number of CPUs for each node
	NodeCpuCount *int64 `json:"node-cpu-count,omitempty"`

	// TODO UNIT of memory for each node
	NodeMemory *int64 `json:"node-memory,omitempty"`

	// Service notifications
	Notifications *[]DbaasServiceNotification `json:"notifications,omitempty"`

	// Subscription plan
	Plan  string            `json:"plan"`
	State *EnumServiceState `json:"state,omitempty"`

	// Service is protected against termination and powering off
	TerminationProtection *bool                `json:"termination-protection,omitempty"`
	Type                  DbaasServiceTypeName `json:"type"`

	// Service last update timestamp (ISO 8601)
	UpdatedAt *time.Time `json:"updated-at,omitempty"`
}

// DbaasServiceKafka defines model for dbaas-service-kafka.
type DbaasServiceKafka struct {
	// List of Kafka ACL entries
	Acl *[]struct {
		// ID
		Id         *string                 `json:"id,omitempty"`
		Permission EnumKafkaAclPermissions `json:"permission"`

		// Topic name pattern
		Topic string `json:"topic"`

		// Username
		Username string `json:"username"`
	} `json:"acl,omitempty"`

	// Kafka authentication methods
	AuthenticationMethods *struct {
		// Whether certificate/SSL authentication is enabled
		Certificate *bool `json:"certificate,omitempty"`

		// Whether SASL authentication is enabled
		Sasl *bool `json:"sasl,omitempty"`
	} `json:"authentication-methods,omitempty"`

	// List of backups for the service
	Backups *[]DbaasServiceBackup `json:"backups,omitempty"`

	// Service component information objects
	Components *[]struct {
		// Service component name
		Component string `json:"component"`

		// DNS name for connecting to the service component
		Host                      string               `json:"host"`
		KafkaAuthenticationMethod *EnumKafkaAuthMethod `json:"kafka-authentication-method,omitempty"`

		// Port number for connecting to the service component
		Port  int64              `json:"port"`
		Route EnumComponentRoute `json:"route"`
		Usage EnumComponentUsage `json:"usage"`
	} `json:"components,omitempty"`

	// Kafka connection information properties
	ConnectionInfo *struct {
		AccessCert  *string   `json:"access-cert,omitempty"`
		AccessKey   *string   `json:"access-key,omitempty"`
		Nodes       *[]string `json:"nodes,omitempty"`
		RegistryUri *string   `json:"registry-uri,omitempty"`
		RestUri     *string   `json:"rest-uri,omitempty"`
	} `json:"connection-info,omitempty"`

	// Service creation timestamp (ISO 8601)
	CreatedAt *time.Time `json:"created-at,omitempty"`

	// TODO UNIT disk space for data storage
	DiskSize *int64 `json:"disk-size,omitempty"`

	// Service integrations
	Integrations *[]DbaasIntegration `json:"integrations,omitempty"`

	// Allow incoming connections from CIDR address block, e.g. '10.20.0.0/16'
	IpFilter *[]string `json:"ip-filter,omitempty"`

	// Whether Kafka Connect is enabled
	KafkaConnectEnabled *bool `json:"kafka-connect-enabled,omitempty"`

	// Kafka Connect configuration values
	KafkaConnectSettings *map[string]interface{} `json:"kafka-connect-settings,omitempty"`

	// Whether Kafka REST is enabled
	KafkaRestEnabled *bool `json:"kafka-rest-enabled,omitempty"`

	// Kafka REST configuration
	KafkaRestSettings *map[string]interface{} `json:"kafka-rest-settings,omitempty"`

	// Kafka-specific settings
	KafkaSettings *map[string]interface{} `json:"kafka-settings,omitempty"`

	// Automatic maintenance settings
	Maintenance *DbaasServiceMaintenance `json:"maintenance,omitempty"`
	Name        DbaasServiceName         `json:"name"`

	// Number of service nodes in the active plan
	NodeCount *int64 `json:"node-count,omitempty"`

	// Number of CPUs for each node
	NodeCpuCount *int64 `json:"node-cpu-count,omitempty"`

	// TODO UNIT of memory for each node
	NodeMemory *int64 `json:"node-memory,omitempty"`

	// State of individual service nodes
	NodeStates *[]DbaasNodeState `json:"node-states,omitempty"`

	// Service notifications
	Notifications *[]DbaasServiceNotification `json:"notifications,omitempty"`

	// Subscription plan
	Plan string `json:"plan"`

	// Whether Schema-Registry is enabled
	SchemaRegistryEnabled *bool `json:"schema-registry-enabled,omitempty"`

	// Schema Registry configuration
	SchemaRegistrySettings *map[string]interface{} `json:"schema-registry-settings,omitempty"`
	State                  *EnumServiceState       `json:"state,omitempty"`

	// Service is protected against termination and powering off
	TerminationProtection *bool                `json:"termination-protection,omitempty"`
	Type                  DbaasServiceTypeName `json:"type"`

	// Service last update timestamp (ISO 8601)
	UpdatedAt *time.Time `json:"updated-at,omitempty"`

	// URI for connecting to the service (may be absent)
	Uri *string `json:"uri,omitempty"`

	// service_uri parameterized into key-value pairs
	UriParams *map[string]interface{} `json:"uri-params,omitempty"`

	// List of service users
	Users *[]struct {
		AccessCert       *string    `json:"access-cert,omitempty"`
		AccessCertExpiry *time.Time `json:"access-cert-expiry,omitempty"`
		AccessKey        *string    `json:"access-key,omitempty"`
		Password         *string    `json:"password,omitempty"`
		Type             *string    `json:"type,omitempty"`
		Username         *string    `json:"username,omitempty"`
	} `json:"users,omitempty"`

	// Kafka version
	Version *string `json:"version,omitempty"`
}

// DbaasServiceLogs defines model for dbaas-service-logs.
type DbaasServiceLogs struct {
	FirstLogOffset *string `json:"first-log-offset,omitempty"`
	Logs           *[]struct {
		Message *string `json:"message,omitempty"`
		Node    *string `json:"node,omitempty"`
		Time    *string `json:"time,omitempty"`
		Unit    *string `json:"unit,omitempty"`
	} `json:"logs,omitempty"`
	Offset *string `json:"offset,omitempty"`
}

// Automatic maintenance settings
type DbaasServiceMaintenance struct {
	// Day of week for installing updates
	Dow DbaasServiceMaintenanceDow `json:"dow"`

	// Time for installing updates, UTC
	Time string `json:"time"`

	// List of updates waiting to be installed
	Updates []DbaasServiceUpdate `json:"updates"`
}

// Day of week for installing updates
type DbaasServiceMaintenanceDow string

// DbaasServiceMysql defines model for dbaas-service-mysql.
type DbaasServiceMysql struct {
	// Backup schedule
	BackupSchedule *struct {
		// The hour of day (in UTC) when backup for the service is started. New backup is only started if previous backup has already completed.
		BackupHour *int64 `json:"backup-hour,omitempty"`

		// The minute of an hour when backup for the service is started. New backup is only started if previous backup has already completed.
		BackupMinute *int64 `json:"backup-minute,omitempty"`
	} `json:"backup-schedule,omitempty"`

	// List of backups for the service
	Backups *[]DbaasServiceBackup `json:"backups,omitempty"`

	// Service component information objects
	Components *[]struct {
		// Service component name
		Component string `json:"component"`

		// DNS name for connecting to the service component
		Host string `json:"host"`

		// Port number for connecting to the service component
		Port  int64              `json:"port"`
		Route EnumComponentRoute `json:"route"`
		Usage EnumComponentUsage `json:"usage"`
	} `json:"components,omitempty"`

	// MySQL connection information properties
	ConnectionInfo *struct {
		Params *[]struct {
			AdditionalProperties map[string]string `json:"-"`
		} `json:"params,omitempty"`
		Standby *[]string `json:"standby,omitempty"`
		Uri     *[]string `json:"uri,omitempty"`
	} `json:"connection-info,omitempty"`

	// Service creation timestamp (ISO 8601)
	CreatedAt *time.Time `json:"created-at,omitempty"`

	// TODO UNIT disk space for data storage
	DiskSize *int64 `json:"disk-size,omitempty"`

	// Service integrations
	Integrations *[]DbaasIntegration `json:"integrations,omitempty"`

	// Allowed CIDR address blocks for incoming connections
	IpFilter *[]string `json:"ip-filter,omitempty"`

	// Automatic maintenance settings
	Maintenance *DbaasServiceMaintenance `json:"maintenance,omitempty"`

	// MySQL-specific settings
	MysqlSettings *map[string]interface{} `json:"mysql-settings,omitempty"`
	Name          DbaasServiceName        `json:"name"`

	// Number of service nodes in the active plan
	NodeCount *int64 `json:"node-count,omitempty"`

	// Number of CPUs for each node
	NodeCpuCount *int64 `json:"node-cpu-count,omitempty"`

	// TODO UNIT of memory for each node
	NodeMemory *int64 `json:"node-memory,omitempty"`

	// State of individual service nodes
	NodeStates *[]DbaasNodeState `json:"node-states,omitempty"`

	// Service notifications
	Notifications *[]DbaasServiceNotification `json:"notifications,omitempty"`

	// Subscription plan
	Plan  string            `json:"plan"`
	State *EnumServiceState `json:"state,omitempty"`

	// Service is protected against termination and powering off
	TerminationProtection *bool                `json:"termination-protection,omitempty"`
	Type                  DbaasServiceTypeName `json:"type"`

	// Service last update timestamp (ISO 8601)
	UpdatedAt *time.Time `json:"updated-at,omitempty"`

	// URI for connecting to the service (may be absent)
	Uri *string `json:"uri,omitempty"`

	// service_uri parameterized into key-value pairs
	UriParams *map[string]interface{} `json:"uri-params,omitempty"`

	// List of service users
	Users *[]struct {
		Authentication *string `json:"authentication,omitempty"`
		Password       *string `json:"password,omitempty"`
		Type           *string `json:"type,omitempty"`
		Username       *string `json:"username,omitempty"`
	} `json:"users,omitempty"`

	// MySQL version
	Version *string `json:"version,omitempty"`
}

// DbaasServiceName defines model for dbaas-service-name.
type DbaasServiceName string

// Service notifications
type DbaasServiceNotification struct {
	// Notification level
	Level DbaasServiceNotificationLevel `json:"level"`

	// Human notification message
	Message string `json:"message"`

	// Notification type
	Metadata map[string]interface{} `json:"metadata"`

	// Notification type
	Type DbaasServiceNotificationType `json:"type"`
}

// Notification level
type DbaasServiceNotificationLevel string

// Notification type
type DbaasServiceNotificationType string

// DbaasServiceOpensearch defines model for dbaas-service-opensearch.
type DbaasServiceOpensearch struct {
	// List of backups for the service
	Backups *[]DbaasServiceBackup `json:"backups,omitempty"`

	// Service component information objects
	Components *[]struct {
		// Service component name
		Component string `json:"component"`

		// DNS name for connecting to the service component
		Host string `json:"host"`

		// Port number for connecting to the service component
		Port  int64              `json:"port"`
		Route EnumComponentRoute `json:"route"`
		Usage EnumComponentUsage `json:"usage"`
	} `json:"components,omitempty"`

	// Opensearch connection information properties
	ConnectionInfo *struct {
		DashboardUri *string   `json:"dashboard-uri,omitempty"`
		Password     *string   `json:"password,omitempty"`
		Uri          *[]string `json:"uri,omitempty"`
		Username     *string   `json:"username,omitempty"`
	} `json:"connection-info,omitempty"`

	// Service creation timestamp (ISO 8601)
	CreatedAt *time.Time `json:"created-at,omitempty"`

	// DbaaS service description
	Description *string `json:"description,omitempty"`

	// TODO UNIT disk space for data storage
	DiskSize *int64 `json:"disk-size,omitempty"`

	// Allows you to create glob style patterns and set a max number of indexes matching this pattern you want to keep. Creating indexes exceeding this value will cause the oldest one to get deleted. You could for example create a pattern looking like 'logs.?' and then create index logs.1, logs.2 etc, it will delete logs.1 once you create logs.6. Do note 'logs.?' does not apply to logs.10. Note: Setting max_index_count to 0 will do nothing and the pattern gets ignored.
	IndexPatterns *[]struct {
		// Maximum number of indexes to keep
		MaxIndexCount *int64 `json:"max-index-count,omitempty"`

		// fnmatch pattern
		Pattern *string `json:"pattern,omitempty"`

		// Deletion sorting algorithm
		SortingAlgorithm *DbaasServiceOpensearchIndexPatternsSortingAlgorithm `json:"sorting-algorithm,omitempty"`
	} `json:"index-patterns,omitempty"`

	// Template settings for all new indexes
	IndexTemplate *struct {
		// The maximum number of nested JSON objects that a single document can contain across all nested types. This limit helps to prevent out of memory errors when a document contains too many nested objects. Default is 10000.
		MappingNestedObjectsLimit *int64 `json:"mapping-nested-objects-limit,omitempty"`

		// The number of replicas each primary shard has.
		NumberOfReplicas *int64 `json:"number-of-replicas,omitempty"`

		// The number of primary shards that an index should have.
		NumberOfShards *int64 `json:"number-of-shards,omitempty"`
	} `json:"index-template,omitempty"`

	// Service integrations
	Integrations *[]DbaasIntegration `json:"integrations,omitempty"`

	// Allowed CIDR address blocks for incoming connections
	IpFilter *[]string `json:"ip-filter,omitempty"`

	// Aiven automation resets index.refresh_interval to default value for every index to be sure that indices are always visible to search. If it doesn't fit your case, you can disable this by setting up this flag to true.
	KeepIndexRefreshInterval *bool `json:"keep-index-refresh-interval,omitempty"`

	// Automatic maintenance settings
	Maintenance *DbaasServiceMaintenance `json:"maintenance,omitempty"`

	// Maximum number of indexes to keep before deleting the oldest one
	MaxIndexCount *int64           `json:"max-index-count,omitempty"`
	Name          DbaasServiceName `json:"name"`

	// Number of service nodes in the active plan
	NodeCount *int64 `json:"node-count,omitempty"`

	// Number of CPUs for each node
	NodeCpuCount *int64 `json:"node-cpu-count,omitempty"`

	// TODO UNIT of memory for each node
	NodeMemory *int64 `json:"node-memory,omitempty"`

	// State of individual service nodes
	NodeStates *[]DbaasNodeState `json:"node-states,omitempty"`

	// Service notifications
	Notifications *[]DbaasServiceNotification `json:"notifications,omitempty"`

	// OpenSearch Dashboards settings
	OpensearchDashboards *struct {
		// Enable or disable OpenSearch Dashboards (default: true)
		Enabled *bool `json:"enabled,omitempty"`

		// Limits the maximum amount of memory (in MiB) the OpenSearch Dashboards process can use. This sets the max_old_space_size option of the nodejs running the OpenSearch Dashboards. Note: the memory reserved by OpenSearch Dashboards is not available for OpenSearch. (default: 128)
		MaxOldSpaceSize *int64 `json:"max-old-space-size,omitempty"`

		// Timeout in milliseconds for requests made by OpenSearch Dashboards towards OpenSearch (default: 30000)
		OpensearchRequestTimeout *int64 `json:"opensearch-request-timeout,omitempty"`
	} `json:"opensearch-dashboards,omitempty"`

	// OpenSearch-specific settings
	OpensearchSettings *map[string]interface{} `json:"opensearch-settings,omitempty"`

	// Subscription plan
	Plan  string            `json:"plan"`
	State *EnumServiceState `json:"state,omitempty"`

	// Service is protected against termination and powering off
	TerminationProtection *bool                `json:"termination-protection,omitempty"`
	Type                  DbaasServiceTypeName `json:"type"`

	// Service last update timestamp (ISO 8601)
	UpdatedAt *time.Time `json:"updated-at,omitempty"`

	// URI for connecting to the service (may be absent)
	Uri *string `json:"uri,omitempty"`

	// service_uri parameterized into key-value pairs
	UriParams *map[string]interface{} `json:"uri-params,omitempty"`

	// List of service users
	Users *[]struct {
		Password *string `json:"password,omitempty"`
		Type     *string `json:"type,omitempty"`
		Username *string `json:"username,omitempty"`
	} `json:"users,omitempty"`

	// OpenSearch version
	Version *string `json:"version,omitempty"`
}

// Deletion sorting algorithm
type DbaasServiceOpensearchIndexPatternsSortingAlgorithm string

// DbaasServicePg defines model for dbaas-service-pg.
type DbaasServicePg struct {
	// Backup schedule
	BackupSchedule *struct {
		// The hour of day (in UTC) when backup for the service is started. New backup is only started if previous backup has already completed.
		BackupHour *int64 `json:"backup-hour,omitempty"`

		// The minute of an hour when backup for the service is started. New backup is only started if previous backup has already completed.
		BackupMinute *int64 `json:"backup-minute,omitempty"`
	} `json:"backup-schedule,omitempty"`

	// List of backups for the service
	Backups *[]DbaasServiceBackup `json:"backups,omitempty"`

	// Service component information objects
	Components *[]struct {
		// Service component name
		Component string `json:"component"`

		// DNS name for connecting to the service component
		Host string `json:"host"`

		// Port number for connecting to the service component
		Port  int64              `json:"port"`
		Route EnumComponentRoute `json:"route"`
		Usage EnumComponentUsage `json:"usage"`
	} `json:"components,omitempty"`

	// PG connection information properties
	ConnectionInfo *struct {
		Params *[]struct {
			AdditionalProperties map[string]string `json:"-"`
		} `json:"params,omitempty"`
		Standby *[]string                 `json:"standby,omitempty"`
		Syncing *[]map[string]interface{} `json:"syncing,omitempty"`
		Uri     *[]string                 `json:"uri,omitempty"`
	} `json:"connection-info,omitempty"`

	// PostgreSQL PGBouncer connection pools
	ConnectionPools *[]struct {
		// Connection URI for the DB pool
		ConnectionUri string `json:"connection-uri"`

		// Service database name
		Database string         `json:"database"`
		Mode     EnumPgPoolMode `json:"mode"`

		// Connection pool name
		Name string `json:"name"`

		// Size of PGBouncer's PostgreSQL side connection pool
		Size int64 `json:"size"`

		// Pool username
		Username string `json:"username"`
	} `json:"connection-pools,omitempty"`

	// Service creation timestamp (ISO 8601)
	CreatedAt *time.Time `json:"created-at,omitempty"`

	// TODO UNIT disk space for data storage
	DiskSize *int64 `json:"disk-size,omitempty"`

	// Service integrations
	Integrations *[]DbaasIntegration `json:"integrations,omitempty"`

	// Allowed CIDR address blocks for incoming connections
	IpFilter *[]string `json:"ip-filter,omitempty"`

	// Automatic maintenance settings
	Maintenance *DbaasServiceMaintenance `json:"maintenance,omitempty"`
	Name        DbaasServiceName         `json:"name"`

	// Number of service nodes in the active plan
	NodeCount *int64 `json:"node-count,omitempty"`

	// Number of CPUs for each node
	NodeCpuCount *int64 `json:"node-cpu-count,omitempty"`

	// TODO UNIT of memory for each node
	NodeMemory *int64 `json:"node-memory,omitempty"`

	// State of individual service nodes
	NodeStates *[]DbaasNodeState `json:"node-states,omitempty"`

	// Service notifications
	Notifications *[]DbaasServiceNotification `json:"notifications,omitempty"`

	// PostgreSQL-specific settings
	PgSettings *map[string]interface{} `json:"pg-settings,omitempty"`

	// PGBouncer connection pooling settings
	PgbouncerSettings *map[string]interface{} `json:"pgbouncer-settings,omitempty"`

	// PGLookout settings
	PglookoutSettings *map[string]interface{} `json:"pglookout-settings,omitempty"`

	// Subscription plan
	Plan string `json:"plan"`

	// Percentage of total RAM that the database server uses for shared memory buffers. Valid range is 20-60 (float), which corresponds to 20% - 60%. This setting adjusts the shared_buffers configuration value.
	SharedBuffersPercentage *int64                        `json:"shared-buffers-percentage,omitempty"`
	State                   *EnumServiceState             `json:"state,omitempty"`
	SynchronousReplication  *EnumPgSynchronousReplication `json:"synchronous-replication,omitempty"`

	// Service is protected against termination and powering off
	TerminationProtection *bool `json:"termination-protection,omitempty"`

	// TimescaleDB extension configuration values
	TimescaledbSettings *map[string]interface{} `json:"timescaledb-settings,omitempty"`
	Type                DbaasServiceTypeName    `json:"type"`

	// Service last update timestamp (ISO 8601)
	UpdatedAt *time.Time `json:"updated-at,omitempty"`

	// URI for connecting to the service (may be absent)
	Uri *string `json:"uri,omitempty"`

	// service_uri parameterized into key-value pairs
	UriParams *map[string]interface{} `json:"uri-params,omitempty"`

	// List of service users
	Users *[]struct {
		AccessControl *struct {
			AllowReplication *bool `json:"allow-replication,omitempty"`
		} `json:"access-control,omitempty"`
		Password *string `json:"password,omitempty"`
		Type     *string `json:"type,omitempty"`
		Username *string `json:"username,omitempty"`
	} `json:"users,omitempty"`

	// PostgreSQL version
	Version *string `json:"version,omitempty"`

	// Sets the maximum amount of memory to be used by a query operation (such as a sort or hash table) before writing to temporary disk files, in MB. Default is 1MB + 0.075% of total RAM (up to 32MB).
	WorkMem *int64 `json:"work-mem,omitempty"`
}

// DbaasServiceRedis defines model for dbaas-service-redis.
type DbaasServiceRedis struct {
	// List of backups for the service
	Backups *[]DbaasServiceBackup `json:"backups,omitempty"`

	// Service component information objects
	Components *[]struct {
		// Service component name
		Component string `json:"component"`

		// DNS name for connecting to the service component
		Host string `json:"host"`

		// Port number for connecting to the service component
		Port  int64              `json:"port"`
		Route EnumComponentRoute `json:"route"`

		// Whether the endpoint is encrypted or accepts plaintext.
		//              By default endpoints are always encrypted and
		//              this property is only included for service components that may disable encryption.
		Ssl   *bool              `json:"ssl,omitempty"`
		Usage EnumComponentUsage `json:"usage"`
	} `json:"components,omitempty"`

	// Redis connection information properties
	ConnectionInfo *struct {
		Password *string   `json:"password,omitempty"`
		Slave    *[]string `json:"slave,omitempty"`
		Uri      *[]string `json:"uri,omitempty"`
	} `json:"connection-info,omitempty"`

	// Service creation timestamp (ISO 8601)
	CreatedAt *time.Time `json:"created-at,omitempty"`

	// TODO UNIT disk space for data storage
	DiskSize *int64 `json:"disk-size,omitempty"`

	// Service integrations
	Integrations *[]DbaasIntegration `json:"integrations,omitempty"`

	// Allowed CIDR address blocks for incoming connections
	IpFilter *[]string `json:"ip-filter,omitempty"`

	// Automatic maintenance settings
	Maintenance *DbaasServiceMaintenance `json:"maintenance,omitempty"`
	Name        DbaasServiceName         `json:"name"`

	// Number of service nodes in the active plan
	NodeCount *int64 `json:"node-count,omitempty"`

	// Number of CPUs for each node
	NodeCpuCount *int64 `json:"node-cpu-count,omitempty"`

	// TODO UNIT of memory for each node
	NodeMemory *int64 `json:"node-memory,omitempty"`

	// State of individual service nodes
	NodeStates *[]DbaasNodeState `json:"node-states,omitempty"`

	// Service notifications
	Notifications *[]DbaasServiceNotification `json:"notifications,omitempty"`

	// Subscription plan
	Plan string `json:"plan"`

	// Redis-specific settings
	RedisSettings *map[string]interface{} `json:"redis-settings,omitempty"`
	State         *EnumServiceState       `json:"state,omitempty"`

	// Service is protected against termination and powering off
	TerminationProtection *bool                `json:"termination-protection,omitempty"`
	Type                  DbaasServiceTypeName `json:"type"`

	// Service last update timestamp (ISO 8601)
	UpdatedAt *time.Time `json:"updated-at,omitempty"`

	// URI for connecting to the service (may be absent)
	Uri *string `json:"uri,omitempty"`

	// service_uri parameterized into key-value pairs
	UriParams *map[string]interface{} `json:"uri-params,omitempty"`

	// List of service users
	Users *[]struct {
		AccessControl *struct {
			Categories *[]string `json:"categories,omitempty"`
			Channels   *[]string `json:"channels,omitempty"`
			Commands   *[]string `json:"commands,omitempty"`
			Keys       *[]string `json:"keys,omitempty"`
		} `json:"access-control,omitempty"`
		Password *string `json:"password,omitempty"`
		Type     *string `json:"type,omitempty"`
		Username *string `json:"username,omitempty"`
	} `json:"users,omitempty"`

	// Redis version
	Version *string `json:"version,omitempty"`
}

// DBaaS service
type DbaasServiceType struct {
	// DbaaS service available versions
	AvailableVersions *[]string `json:"available-versions,omitempty"`

	// DbaaS service default version
	DefaultVersion *string `json:"default-version,omitempty"`

	// DbaaS service description
	Description *string               `json:"description,omitempty"`
	Name        *DbaasServiceTypeName `json:"name,omitempty"`

	// DbaaS service plans
	Plans *[]DbaasPlan `json:"plans,omitempty"`
}

// DbaasServiceTypeName defines model for dbaas-service-type-name.
type DbaasServiceTypeName string

// Update waiting to be installed
type DbaasServiceUpdate struct {
	// Deadline for installing the update
	Deadline *time.Time `json:"deadline,omitempty"`

	// Description of the update
	Description *string `json:"description,omitempty"`

	// The earliest time the update will be automatically applied
	StartAfter *time.Time `json:"start-after,omitempty"`

	// The time when the update will be automatically applied
	StartAt *time.Time `json:"start-at,omitempty"`
}

// Deploy target
type DeployTarget struct {
	// Deploy Target description
	Description *string `json:"description,omitempty"`

	// Deploy Target ID
	Id *string `json:"id,omitempty"`

	// Deploy Target name
	Name *string `json:"name,omitempty"`

	// Deploy Target type
	Type *DeployTargetType `json:"type,omitempty"`
}

// Deploy Target type
type DeployTargetType string

// DNS domain
type DnsDomain struct {
	// DNS domain creation date
	CreatedAt *time.Time `json:"created-at,omitempty"`

	// DNS domain ID
	Id *string `json:"id,omitempty"`

	// DNS domain unicode name
	UnicodeName *string `json:"unicode-name,omitempty"`
}

// DNS domain record
type DnsDomainRecord struct {
	// DNS domain record content
	Content *string `json:"content,omitempty"`

	// DNS domain record creation date
	CreatedAt *time.Time `json:"created-at,omitempty"`

	// DNS domain record ID
	Id *string `json:"id,omitempty"`

	// DNS domain record name
	Name *string `json:"name,omitempty"`

	// DNS domain record priority
	Priority *int64 `json:"priority,omitempty"`

	// DNS domain record TTL
	Ttl *int64 `json:"ttl,omitempty"`

	// DNS domain record priority
	Type *DnsDomainRecordType `json:"type,omitempty"`

	// DNS domain record update date
	UpdatedAt *time.Time `json:"updated-at,omitempty"`
}

// DNS domain record priority
type DnsDomainRecordType string

// Elastic IP
type ElasticIp struct {
	// Elastic IP description
	Description *string `json:"description,omitempty"`

	// Elastic IP address healthcheck
	Healthcheck *ElasticIpHealthcheck `json:"healthcheck,omitempty"`

	// Elastic IP ID
	Id *string `json:"id,omitempty"`

	// Elastic IP address
	Ip     *string `json:"ip,omitempty"`
	Labels *Labels `json:"labels,omitempty"`
}

// Elastic IP address healthcheck
type ElasticIpHealthcheck struct {
	// Interval between the checks (default: 10)
	Interval *int64 `json:"interval,omitempty"`

	// Healthcheck mode
	Mode ElasticIpHealthcheckMode `json:"mode"`

	// Healthcheck port
	Port int64 `json:"port"`

	// Number of attempts before considering the target unhealthy (default: 3)
	StrikesFail *int64 `json:"strikes-fail,omitempty"`

	// Number of attempts before considering the target healthy (default: 2)
	StrikesOk *int64 `json:"strikes-ok,omitempty"`

	// Healthcheck timeout value (default: 2)
	Timeout *int64 `json:"timeout,omitempty"`

	// Skip TLS verification
	TlsSkipVerify *bool `json:"tls-skip-verify,omitempty"`

	// SNI domain for HTTPS healthchecks
	TlsSni *string `json:"tls-sni,omitempty"`

	// Healthcheck URI
	Uri *string `json:"uri,omitempty"`
}

// Healthcheck mode
type ElasticIpHealthcheckMode string

// EnumComponentRoute defines model for enum-component-route.
type EnumComponentRoute string

// EnumComponentUsage defines model for enum-component-usage.
type EnumComponentUsage string

// EnumIntegrationTypes defines model for enum-integration-types.
type EnumIntegrationTypes string

// EnumKafkaAclPermissions defines model for enum-kafka-acl-permissions.
type EnumKafkaAclPermissions string

// EnumKafkaAuthMethod defines model for enum-kafka-auth-method.
type EnumKafkaAuthMethod string

// EnumMasterLinkStatus defines model for enum-master-link-status.
type EnumMasterLinkStatus string

// EnumMigrationStatus defines model for enum-migration-status.
type EnumMigrationStatus string

// EnumPgMigrationMethod defines model for enum-pg-migration-method.
type EnumPgMigrationMethod string

// EnumPgPoolMode defines model for enum-pg-pool-mode.
type EnumPgPoolMode string

// EnumPgSynchronousReplication defines model for enum-pg-synchronous-replication.
type EnumPgSynchronousReplication string

// EnumPgVariant defines model for enum-pg-variant.
type EnumPgVariant string

// EnumServiceState defines model for enum-service-state.
type EnumServiceState string

// EnumSortOrder defines model for enum-sort-order.
type EnumSortOrder string

// A notable event which happened on the infrastructure
type Event struct {
	// Event payload. This is a free-form map
	Payload *map[string]interface{} `json:"payload,omitempty"`

	// Time at which the event happened, millisecond resolution
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

// Instance
type Instance struct {
	// Instance Anti-affinity Groups
	AntiAffinityGroups *[]AntiAffinityGroup `json:"anti-affinity-groups,omitempty"`

	// Instance creation date
	CreatedAt *time.Time `json:"created-at,omitempty"`

	// Deploy target
	DeployTarget *DeployTarget `json:"deploy-target,omitempty"`

	// Instance disk size in GB
	DiskSize *int64 `json:"disk-size,omitempty"`

	// Instance Elastic IPs
	ElasticIps *[]ElasticIp `json:"elastic-ips,omitempty"`

	// Instance ID
	Id *string `json:"id,omitempty"`

	// Compute instance type
	InstanceType *InstanceType `json:"instance-type,omitempty"`

	// Instance IPv6 address
	Ipv6Address *string `json:"ipv6-address,omitempty"`
	Labels      *Labels `json:"labels,omitempty"`

	// Resource manager
	Manager *Manager `json:"manager,omitempty"`

	// Instance name
	Name *string `json:"name,omitempty"`

	// Instance Private Networks
	PrivateNetworks *[]PrivateNetwork `json:"private-networks,omitempty"`

	// Instance public IPv4 address
	PublicIp *string `json:"public-ip,omitempty"`

	// Instance Security Groups
	SecurityGroups *[]SecurityGroup `json:"security-groups,omitempty"`

	// Instance Snapshots
	Snapshots *[]Snapshot `json:"snapshots,omitempty"`

	// SSH key
	SshKey *SshKey `json:"ssh-key,omitempty"`

	// Instance state
	State *InstanceState `json:"state,omitempty"`

	// Instance template
	Template *Template `json:"template,omitempty"`

	// Instance Cloud-init user-data
	UserData *string `json:"user-data,omitempty"`
}

// Instance state
type InstanceState string

// Instance Pool
type InstancePool struct {
	// Instance Pool Anti-affinity Groups
	AntiAffinityGroups *[]AntiAffinityGroup `json:"anti-affinity-groups,omitempty"`

	// Deploy target
	DeployTarget *DeployTarget `json:"deploy-target,omitempty"`

	// Instance Pool description
	Description *string `json:"description,omitempty"`

	// Instances disk size in GB
	DiskSize *int64 `json:"disk-size,omitempty"`

	// Instances Elastic IPs
	ElasticIps *[]ElasticIp `json:"elastic-ips,omitempty"`

	// Instance Pool ID
	Id *string `json:"id,omitempty"`

	// The instances created by the Instance Pool will be prefixed with this value (default: pool)
	InstancePrefix *string `json:"instance-prefix,omitempty"`

	// Compute instance type
	InstanceType *InstanceType `json:"instance-type,omitempty"`

	// Instances
	Instances *[]Instance `json:"instances,omitempty"`

	// Enable IPv6 for instances
	Ipv6Enabled *bool   `json:"ipv6-enabled,omitempty"`
	Labels      *Labels `json:"labels,omitempty"`

	// Resource manager
	Manager *Manager `json:"manager,omitempty"`

	// Instance Pool name
	Name *string `json:"name,omitempty"`

	// Instance Pool Private Networks
	PrivateNetworks *[]PrivateNetwork `json:"private-networks,omitempty"`

	// Instance Pool Security Groups
	SecurityGroups *[]SecurityGroup `json:"security-groups,omitempty"`

	// Number of instances
	Size *int64 `json:"size,omitempty"`

	// SSH key
	SshKey *SshKey `json:"ssh-key,omitempty"`

	// Instance Pool state
	State *InstancePoolState `json:"state,omitempty"`

	// Instance template
	Template *Template `json:"template,omitempty"`

	// Instances Cloud-init user-data
	UserData *string `json:"user-data,omitempty"`
}

// Instance Pool state
type InstancePoolState string

// Compute instance type
type InstanceType struct {
	// Requires authorization or publicly available
	Authorized *bool `json:"authorized,omitempty"`

	// CPU count
	Cpus *int64 `json:"cpus,omitempty"`

	// Instance type family
	Family *InstanceTypeFamily `json:"family,omitempty"`

	// GPU count
	Gpus *int64 `json:"gpus,omitempty"`

	// Instance type ID
	Id *string `json:"id,omitempty"`

	// Available memory
	Memory *int64 `json:"memory,omitempty"`

	// Instance type size
	Size *InstanceTypeSize `json:"size,omitempty"`

	// Instance Type available zones
	Zones *[]ZoneName `json:"zones,omitempty"`
}

// Instance type family
type InstanceTypeFamily string

// Instance type size
type InstanceTypeSize string

// Labels defines model for labels.
type Labels struct {
	AdditionalProperties map[string]string `json:"-"`
}

// Load Balancer
type LoadBalancer struct {
	// Load Balancer creation date
	CreatedAt *time.Time `json:"created-at,omitempty"`

	// Load Balancer description
	Description *string `json:"description,omitempty"`

	// Load Balancer ID
	Id *string `json:"id,omitempty"`

	// Load Balancer public IP
	Ip     *string `json:"ip,omitempty"`
	Labels *Labels `json:"labels,omitempty"`

	// Load Balancer name
	Name *string `json:"name,omitempty"`

	// Load Balancer Services
	Services *[]LoadBalancerService `json:"services,omitempty"`

	// Load Balancer state
	State *LoadBalancerState `json:"state,omitempty"`
}

// Load Balancer state
type LoadBalancerState string

// Load Balancer Service status
type LoadBalancerServerStatus struct {
	// Backend server public IP
	PublicIp *string `json:"public-ip,omitempty"`

	// Status of the instance's healthcheck
	Status *LoadBalancerServerStatusStatus `json:"status,omitempty"`
}

// Status of the instance's healthcheck
type LoadBalancerServerStatusStatus string

// Load Balancer Service
type LoadBalancerService struct {
	// Load Balancer Service description
	Description *string `json:"description,omitempty"`

	// Load Balancer Service healthcheck
	Healthcheck *LoadBalancerServiceHealthcheck `json:"healthcheck,omitempty"`

	// Healthcheck status per backend server
	HealthcheckStatus *[]LoadBalancerServerStatus `json:"healthcheck-status,omitempty"`

	// Load Balancer Service ID
	Id *string `json:"id,omitempty"`

	// Instance Pool
	InstancePool *InstancePool `json:"instance-pool,omitempty"`

	// Load Balancer Service name
	Name *string `json:"name,omitempty"`

	// Port exposed on the Load Balancer's public IP
	Port *int64 `json:"port,omitempty"`

	// Network traffic protocol
	Protocol *LoadBalancerServiceProtocol `json:"protocol,omitempty"`

	// Load Balancer Service state
	State *LoadBalancerServiceState `json:"state,omitempty"`

	// Load balancing strategy
	Strategy *LoadBalancerServiceStrategy `json:"strategy,omitempty"`

	// Port on which the network traffic will be forwarded to on the receiving instance
	TargetPort *int64 `json:"target-port,omitempty"`
}

// Network traffic protocol
type LoadBalancerServiceProtocol string

// Load Balancer Service state
type LoadBalancerServiceState string

// Load balancing strategy
type LoadBalancerServiceStrategy string

// Load Balancer Service healthcheck
type LoadBalancerServiceHealthcheck struct {
	// Healthcheck interval (default: 10)
	Interval *int64 `json:"interval,omitempty"`

	// Healthcheck mode
	Mode *LoadBalancerServiceHealthcheckMode `json:"mode,omitempty"`

	// Healthcheck port
	Port *int64 `json:"port,omitempty"`

	// Number of retries before considering a Service failed
	Retries *int64 `json:"retries,omitempty"`

	// Healthcheck timeout value (default: 2)
	Timeout *int64 `json:"timeout,omitempty"`

	// SNI domain for HTTPS healthchecks
	TlsSni *string `json:"tls-sni,omitempty"`

	// Healthcheck URI
	Uri *string `json:"uri,omitempty"`
}

// Healthcheck mode
type LoadBalancerServiceHealthcheckMode string

// Resource manager
type Manager struct {
	// Manager ID
	Id *string `json:"id,omitempty"`

	// Manager type
	Type *ManagerType `json:"type,omitempty"`
}

// Manager type
type ManagerType string

// Operation
type Operation struct {
	// Operation ID
	Id *string `json:"id,omitempty"`

	// Operation message
	Message *string `json:"message,omitempty"`

	// Operation failure reason
	Reason *OperationReason `json:"reason,omitempty"`

	// Resource reference
	Reference *Reference `json:"reference,omitempty"`

	// Operation status
	State *OperationState `json:"state,omitempty"`
}

// Operation failure reason
type OperationReason string

// Operation status
type OperationState string

// Private Network
type PrivateNetwork struct {
	// Private Network description
	Description *string `json:"description,omitempty"`

	// Private Network end IP address
	EndIp *string `json:"end-ip,omitempty"`

	// Private Network ID
	Id     *string `json:"id,omitempty"`
	Labels *Labels `json:"labels,omitempty"`

	// Private Network leased IP addresses
	Leases *[]PrivateNetworkLease `json:"leases,omitempty"`

	// Private Network name
	Name *string `json:"name,omitempty"`

	// Private Network netmask
	Netmask *string `json:"netmask,omitempty"`

	// Private Network start IP address
	StartIp *string `json:"start-ip,omitempty"`
}

// Private Network leased IP address
type PrivateNetworkLease struct {
	// Attached instance ID
	InstanceId *string `json:"instance-id,omitempty"`

	// Private Network IP address
	Ip *string `json:"ip,omitempty"`
}

// Organization Quota
type Quota struct {
	// Resource Limit. -1 for Unlimited
	Limit *int64 `json:"limit,omitempty"`

	// Resource Name
	Resource *string `json:"resource,omitempty"`

	// Resource Usage
	Usage *int64 `json:"usage,omitempty"`
}

// Resource reference
type Reference struct {
	// Command name
	Command *string `json:"command,omitempty"`

	// Reference ID
	Id *string `json:"id,omitempty"`

	// Link to the referenced resource
	Link *string `json:"link,omitempty"`
}

// Security Group
type SecurityGroup struct {
	// Security Group description
	Description *string `json:"description,omitempty"`

	// Security Group external sources
	ExternalSources *[]string `json:"external-sources,omitempty"`

	// Security Group ID
	Id *string `json:"id,omitempty"`

	// Security Group name
	Name *string `json:"name,omitempty"`

	// Security Group rules
	Rules *[]SecurityGroupRule `json:"rules,omitempty"`
}

// Security Group
type SecurityGroupResource struct {
	// Security Group ID
	Id string `json:"id"`

	// Security Group name
	Name *string `json:"name,omitempty"`
}

// Security Group rule
type SecurityGroupRule struct {
	// Security Group rule description
	Description *string `json:"description,omitempty"`

	// End port of the range
	EndPort *int64 `json:"end-port,omitempty"`

	// Network flow direction to match
	FlowDirection *SecurityGroupRuleFlowDirection `json:"flow-direction,omitempty"`

	// ICMP details
	Icmp *struct {
		Code *int64 `json:"code,omitempty"`
		Type *int64 `json:"type,omitempty"`
	} `json:"icmp,omitempty"`

	// Security Group rule ID
	Id *string `json:"id,omitempty"`

	// CIDR-formatted network allowed
	Network *string `json:"network,omitempty"`

	// Network protocol
	Protocol *SecurityGroupRuleProtocol `json:"protocol,omitempty"`

	// Security Group
	SecurityGroup *SecurityGroupResource `json:"security-group,omitempty"`

	// Start port of the range
	StartPort *int64 `json:"start-port,omitempty"`
}

// Network flow direction to match
type SecurityGroupRuleFlowDirection string

// Network protocol
type SecurityGroupRuleProtocol string

// SKS Cluster
type SksCluster struct {
	// Cluster addons
	Addons *[]SksClusterAddons `json:"addons,omitempty"`

	// Enable auto upgrade of the control plane to the latest patch version available
	AutoUpgrade *bool `json:"auto-upgrade,omitempty"`

	// Cluster CNI
	Cni *SksClusterCni `json:"cni,omitempty"`

	// Cluster creation date
	CreatedAt *time.Time `json:"created-at,omitempty"`

	// Cluster description
	Description *string `json:"description,omitempty"`

	// Cluster endpoint
	Endpoint *string `json:"endpoint,omitempty"`

	// Cluster ID
	Id     *string `json:"id,omitempty"`
	Labels *Labels `json:"labels,omitempty"`

	// Cluster level
	Level *SksClusterLevel `json:"level,omitempty"`

	// Cluster name
	Name *string `json:"name,omitempty"`

	// Cluster Nodepools
	Nodepools *[]SksNodepool `json:"nodepools,omitempty"`

	// Cluster state
	State *SksClusterState `json:"state,omitempty"`

	// Control plane Kubernetes version
	Version *string `json:"version,omitempty"`
}

// SksClusterAddons defines model for SksCluster.Addons.
type SksClusterAddons string

// Cluster CNI
type SksClusterCni string

// Cluster level
type SksClusterLevel string

// Cluster state
type SksClusterState string

// SksClusterDeprecatedResource defines model for sks-cluster-deprecated-resource.
type SksClusterDeprecatedResource struct {
	AdditionalProperties map[string]string `json:"-"`
}

// Kubeconfig request for a SKS cluster
type SksKubeconfigRequest struct {
	// List of roles. The certificate present in the Kubeconfig will have these roles set in the Org field.
	Groups *[]string `json:"groups,omitempty"`

	// Validity in seconds of the Kubeconfig user certificate (default: 30 days)
	Ttl *int64 `json:"ttl,omitempty"`

	// User name in the generated Kubeconfig. The certificate present in the Kubeconfig will also have this name set for the CN field.
	User *string `json:"user,omitempty"`
}

// SKS Nodepool
type SksNodepool struct {
	// Nodepool addons
	Addons *[]SksNodepoolAddons `json:"addons,omitempty"`

	// Nodepool Anti-affinity Groups
	AntiAffinityGroups *[]AntiAffinityGroup `json:"anti-affinity-groups,omitempty"`

	// Nodepool creation date
	CreatedAt *time.Time `json:"created-at,omitempty"`

	// Deploy target
	DeployTarget *DeployTarget `json:"deploy-target,omitempty"`

	// Nodepool description
	Description *string `json:"description,omitempty"`

	// Nodepool instances disk size in GB
	DiskSize *int64 `json:"disk-size,omitempty"`

	// Nodepool ID
	Id *string `json:"id,omitempty"`

	// Instance Pool
	InstancePool *InstancePool `json:"instance-pool,omitempty"`

	// The instances created by the Nodepool will be prefixed with this value (default: pool)
	InstancePrefix *string `json:"instance-prefix,omitempty"`

	// Compute instance type
	InstanceType *InstanceType `json:"instance-type,omitempty"`
	Labels       *Labels       `json:"labels,omitempty"`

	// Nodepool name
	Name *string `json:"name,omitempty"`

	// Nodepool Private Networks
	PrivateNetworks *[]PrivateNetwork `json:"private-networks,omitempty"`

	// Nodepool Security Groups
	SecurityGroups *[]SecurityGroup `json:"security-groups,omitempty"`

	// Number of instances
	Size *int64 `json:"size,omitempty"`

	// Nodepool state
	State  *SksNodepoolState  `json:"state,omitempty"`
	Taints *SksNodepoolTaints `json:"taints,omitempty"`

	// Instance template
	Template *Template `json:"template,omitempty"`

	// Nodepool version
	Version *string `json:"version,omitempty"`
}

// SksNodepoolAddons defines model for SksNodepool.Addons.
type SksNodepoolAddons string

// Nodepool state
type SksNodepoolState string

// Nodepool taint
type SksNodepoolTaint struct {
	// Nodepool taint effect
	Effect SksNodepoolTaintEffect `json:"effect"`

	// Nodepool taint value
	Value string `json:"value"`
}

// Nodepool taint effect
type SksNodepoolTaintEffect string

// SksNodepoolTaints defines model for sks-nodepool-taints.
type SksNodepoolTaints struct {
	AdditionalProperties map[string]SksNodepoolTaint `json:"-"`
}

// SKS Cluster OpenID config map
type SksOidc struct {
	// OpenID client ID
	ClientId string `json:"client-id"`

	// JWT claim to use as the user's group
	GroupsClaim *string `json:"groups-claim,omitempty"`

	// Prefix prepended to group claims
	GroupsPrefix *string `json:"groups-prefix,omitempty"`

	// OpenID provider URL
	IssuerUrl string `json:"issuer-url"`

	// A key value map that describes a required claim in the ID Token
	RequiredClaim *SksOidc_RequiredClaim `json:"required-claim,omitempty"`

	// JWT claim to use as the user name
	UsernameClaim *string `json:"username-claim,omitempty"`

	// Prefix prepended to username claims
	UsernamePrefix *string `json:"username-prefix,omitempty"`
}

// A key value map that describes a required claim in the ID Token
type SksOidc_RequiredClaim struct {
	AdditionalProperties map[string]string `json:"-"`
}

// Snapshot
type Snapshot struct {
	// Snapshot creation date
	CreatedAt *time.Time `json:"created-at,omitempty"`

	// Exported snapshot information
	Export *struct {
		// Exported snapshot disk file MD5 checksum
		Md5sum *string `json:"md5sum,omitempty"`

		// Exported snapshot disk file pre-signed URL
		PresignedUrl *string `json:"presigned-url,omitempty"`
	} `json:"export,omitempty"`

	// Snapshot ID
	Id *string `json:"id,omitempty"`

	// Instance
	Instance *Instance `json:"instance,omitempty"`

	// Snapshot name
	Name *string `json:"name,omitempty"`

	// Snapshot size in GB
	Size *int64 `json:"size,omitempty"`

	// Snapshot state
	State *SnapshotState `json:"state,omitempty"`
}

// Snapshot state
type SnapshotState string

// SSH key
type SshKey struct {
	// SSH key fingerprint
	Fingerprint *string `json:"fingerprint,omitempty"`

	// SSH key name
	Name *string `json:"name,omitempty"`
}

// Instance template
type Template struct {
	// Boot mode (default: legacy)
	BootMode *TemplateBootMode `json:"boot-mode,omitempty"`

	// Template build
	Build *string `json:"build,omitempty"`

	// Template MD5 checksum
	Checksum *string `json:"checksum,omitempty"`

	// Template creation date
	CreatedAt *time.Time `json:"created-at,omitempty"`

	// Template default user
	DefaultUser *string `json:"default-user,omitempty"`

	// Template description
	Description *string `json:"description,omitempty"`

	// Template family
	Family *string `json:"family,omitempty"`

	// Template ID
	Id *string `json:"id,omitempty"`

	// Template maintainer
	Maintainer *string `json:"maintainer,omitempty"`

	// Template name
	Name *string `json:"name,omitempty"`

	// Enable password-based login
	PasswordEnabled *bool `json:"password-enabled,omitempty"`

	// Template size
	Size *int64 `json:"size,omitempty"`

	// Enable SSH key-based login
	SshKeyEnabled *bool `json:"ssh-key-enabled,omitempty"`

	// Template source URL
	Url *string `json:"url,omitempty"`

	// Template version
	Version *string `json:"version,omitempty"`

	// Template visibility
	Visibility *TemplateVisibility `json:"visibility,omitempty"`
}

// Boot mode (default: legacy)
type TemplateBootMode string

// Template visibility
type TemplateVisibility string

// Zone
type Zone struct {
	Name *ZoneName `json:"name,omitempty"`
}

// ZoneName defines model for zone-name.
type ZoneName string

// CreateAccessKeyJSONBody defines parameters for CreateAccessKey.
type CreateAccessKeyJSONBody struct {
	// IAM Access Key name
	Name *string `json:"name,omitempty"`

	// IAM Access Key operations
	Operations *[]string `json:"operations,omitempty"`

	// IAM Access Key Resources
	Resources *[]AccessKeyResource `json:"resources,omitempty"`

	// IAM Access Key tags
	Tags *[]string `json:"tags,omitempty"`
}

// CreateAntiAffinityGroupJSONBody defines parameters for CreateAntiAffinityGroup.
type CreateAntiAffinityGroupJSONBody struct {
	// Anti-affinity Group description
	Description *string `json:"description,omitempty"`

	// Anti-affinity Group name
	Name string `json:"name"`
}

// CreateDbaasServiceKafkaJSONBody defines parameters for CreateDbaasServiceKafka.
type CreateDbaasServiceKafkaJSONBody struct {
	// Kafka authentication methods
	AuthenticationMethods *struct {
		// Enable certificate/SSL authentication
		Certificate *bool `json:"certificate,omitempty"`

		// Enable SASL authentication
		Sasl *bool `json:"sasl,omitempty"`
	} `json:"authentication-methods,omitempty"`

	// Service integrations to enable for the service. Some integration types affect how a service is created and they must be provided as part of the creation call instead of being defined later.
	Integrations *[]struct {
		DestService *DbaasServiceName `json:"dest-service,omitempty"`

		// Integration settings
		Settings      *map[string]interface{} `json:"settings,omitempty"`
		SourceService *DbaasServiceName       `json:"source-service,omitempty"`
		Type          EnumIntegrationTypes    `json:"type"`
	} `json:"integrations,omitempty"`

	// Allow incoming connections from CIDR address block, e.g. '10.20.0.0/16'
	IpFilter *[]string `json:"ip-filter,omitempty"`

	// Allow clients to connect to kafka_connect from the public internet for service nodes that are in a project VPC or another type of private network
	KafkaConnectEnabled *bool `json:"kafka-connect-enabled,omitempty"`

	// Kafka Connect configuration values
	KafkaConnectSettings *map[string]interface{} `json:"kafka-connect-settings,omitempty"`

	// Enable Kafka-REST service
	KafkaRestEnabled *bool `json:"kafka-rest-enabled,omitempty"`

	// Kafka REST configuration
	KafkaRestSettings *map[string]interface{} `json:"kafka-rest-settings,omitempty"`

	// Kafka-specific settings
	KafkaSettings *map[string]interface{} `json:"kafka-settings,omitempty"`

	// Automatic maintenance settings
	Maintenance *struct {
		// Day of week for installing updates
		Dow CreateDbaasServiceKafkaJSONBodyMaintenanceDow `json:"dow"`

		// Time for installing updates, UTC
		Time string `json:"time"`
	} `json:"maintenance,omitempty"`

	// Subscription plan
	Plan string `json:"plan"`

	// Enable Schema-Registry service
	SchemaRegistryEnabled *bool `json:"schema-registry-enabled,omitempty"`

	// Schema Registry configuration
	SchemaRegistrySettings *map[string]interface{} `json:"schema-registry-settings,omitempty"`

	// Service is protected against termination and powering off
	TerminationProtection *bool `json:"termination-protection,omitempty"`

	// Kafka major version
	Version *string `json:"version,omitempty"`
}

// CreateDbaasServiceKafkaJSONBodyMaintenanceDow defines parameters for CreateDbaasServiceKafka.
type CreateDbaasServiceKafkaJSONBodyMaintenanceDow string

// UpdateDbaasServiceKafkaJSONBody defines parameters for UpdateDbaasServiceKafka.
type UpdateDbaasServiceKafkaJSONBody struct {
	// Kafka authentication methods
	AuthenticationMethods *struct {
		// Enable certificate/SSL authentication
		Certificate *bool `json:"certificate,omitempty"`

		// Enable SASL authentication
		Sasl *bool `json:"sasl,omitempty"`
	} `json:"authentication-methods,omitempty"`

	// Allow incoming connections from CIDR address block, e.g. '10.20.0.0/16'
	IpFilter *[]string `json:"ip-filter,omitempty"`

	// Allow clients to connect to kafka_connect from the public internet for service nodes that are in a project VPC or another type of private network
	KafkaConnectEnabled *bool `json:"kafka-connect-enabled,omitempty"`

	// Kafka Connect configuration values
	KafkaConnectSettings *map[string]interface{} `json:"kafka-connect-settings,omitempty"`

	// Enable Kafka-REST service
	KafkaRestEnabled *bool `json:"kafka-rest-enabled,omitempty"`

	// Kafka REST configuration
	KafkaRestSettings *map[string]interface{} `json:"kafka-rest-settings,omitempty"`

	// Kafka-specific settings
	KafkaSettings *map[string]interface{} `json:"kafka-settings,omitempty"`

	// Automatic maintenance settings
	Maintenance *struct {
		// Day of week for installing updates
		Dow UpdateDbaasServiceKafkaJSONBodyMaintenanceDow `json:"dow"`

		// Time for installing updates, UTC
		Time string `json:"time"`
	} `json:"maintenance,omitempty"`

	// Subscription plan
	Plan *string `json:"plan,omitempty"`

	// Enable Schema-Registry service
	SchemaRegistryEnabled *bool `json:"schema-registry-enabled,omitempty"`

	// Schema Registry configuration
	SchemaRegistrySettings *map[string]interface{} `json:"schema-registry-settings,omitempty"`

	// Service is protected against termination and powering off
	TerminationProtection *bool `json:"termination-protection,omitempty"`
}

// UpdateDbaasServiceKafkaJSONBodyMaintenanceDow defines parameters for UpdateDbaasServiceKafka.
type UpdateDbaasServiceKafkaJSONBodyMaintenanceDow string

// CreateDbaasServiceMysqlJSONBody defines parameters for CreateDbaasServiceMysql.
type CreateDbaasServiceMysqlJSONBody struct {
	// Custom password for admin user. Defaults to random string. This must be set only when a new service is being created.
	AdminPassword *string `json:"admin-password,omitempty"`

	// Custom username for admin user. This must be set only when a new service is being created.
	AdminUsername  *string `json:"admin-username,omitempty"`
	BackupSchedule *struct {
		// The hour of day (in UTC) when backup for the service is started. New backup is only started if previous backup has already completed.
		BackupHour *int64 `json:"backup-hour,omitempty"`

		// The minute of an hour when backup for the service is started. New backup is only started if previous backup has already completed.
		BackupMinute *int64 `json:"backup-minute,omitempty"`
	} `json:"backup-schedule,omitempty"`

	// The minimum amount of time in seconds to keep binlog entries before deletion. This may be extended for services that require binlog entries for longer than the default for example if using the MySQL Debezium Kafka connector.
	BinlogRetentionPeriod *int64            `json:"binlog-retention-period,omitempty"`
	ForkFromService       *DbaasServiceName `json:"fork-from-service,omitempty"`

	// Service integrations to enable for the service. Some integration types affect how a service is created and they must be provided as part of the creation call instead of being defined later.
	Integrations *[]struct {
		DestService *DbaasServiceName `json:"dest-service,omitempty"`

		// Integration settings
		Settings      *map[string]interface{} `json:"settings,omitempty"`
		SourceService *DbaasServiceName       `json:"source-service,omitempty"`
		Type          EnumIntegrationTypes    `json:"type"`
	} `json:"integrations,omitempty"`

	// Allow incoming connections from CIDR address block, e.g. '10.20.0.0/16'
	IpFilter *[]string `json:"ip-filter,omitempty"`

	// Automatic maintenance settings
	Maintenance *struct {
		// Day of week for installing updates
		Dow CreateDbaasServiceMysqlJSONBodyMaintenanceDow `json:"dow"`

		// Time for installing updates, UTC
		Time string `json:"time"`
	} `json:"maintenance,omitempty"`

	// Migrate data from existing server
	Migration *struct {
		// Database name for bootstrapping the initial connection
		Dbname *string `json:"dbname,omitempty"`

		// Hostname or IP address of the server where to migrate data from
		Host string `json:"host"`

		// Comma-separated list of databases, which should be ignored during migration (supported by MySQL only at the moment)
		IgnoreDbs *string                `json:"ignore-dbs,omitempty"`
		Method    *EnumPgMigrationMethod `json:"method,omitempty"`

		// Password for authentication with the server where to migrate data from
		Password *string `json:"password,omitempty"`

		// Port number of the server where to migrate data from
		Port int64 `json:"port"`

		// The server where to migrate data from is secured with SSL
		Ssl *bool `json:"ssl,omitempty"`

		// User name for authentication with the server where to migrate data from
		Username *string `json:"username,omitempty"`
	} `json:"migration,omitempty"`

	// MySQL-specific settings
	MysqlSettings *map[string]interface{} `json:"mysql-settings,omitempty"`

	// Subscription plan
	Plan string `json:"plan"`

	// ISO time of a backup to recover from for services that support arbitrary times
	RecoveryBackupTime *string `json:"recovery-backup-time,omitempty"`

	// Service is protected against termination and powering off
	TerminationProtection *bool `json:"termination-protection,omitempty"`

	// MySQL major version
	Version *string `json:"version,omitempty"`
}

// CreateDbaasServiceMysqlJSONBodyMaintenanceDow defines parameters for CreateDbaasServiceMysql.
type CreateDbaasServiceMysqlJSONBodyMaintenanceDow string

// UpdateDbaasServiceMysqlJSONBody defines parameters for UpdateDbaasServiceMysql.
type UpdateDbaasServiceMysqlJSONBody struct {
	BackupSchedule *struct {
		// The hour of day (in UTC) when backup for the service is started. New backup is only started if previous backup has already completed.
		BackupHour *int64 `json:"backup-hour,omitempty"`

		// The minute of an hour when backup for the service is started. New backup is only started if previous backup has already completed.
		BackupMinute *int64 `json:"backup-minute,omitempty"`
	} `json:"backup-schedule,omitempty"`

	// The minimum amount of time in seconds to keep binlog entries before deletion. This may be extended for services that require binlog entries for longer than the default for example if using the MySQL Debezium Kafka connector.
	BinlogRetentionPeriod *int64 `json:"binlog-retention-period,omitempty"`

	// Allow incoming connections from CIDR address block, e.g. '10.20.0.0/16'
	IpFilter *[]string `json:"ip-filter,omitempty"`

	// Automatic maintenance settings
	Maintenance *struct {
		// Day of week for installing updates
		Dow UpdateDbaasServiceMysqlJSONBodyMaintenanceDow `json:"dow"`

		// Time for installing updates, UTC
		Time string `json:"time"`
	} `json:"maintenance,omitempty"`

	// Migrate data from existing server
	Migration *struct {
		// Database name for bootstrapping the initial connection
		Dbname *string `json:"dbname,omitempty"`

		// Hostname or IP address of the server where to migrate data from
		Host string `json:"host"`

		// Comma-separated list of databases, which should be ignored during migration (supported by MySQL only at the moment)
		IgnoreDbs *string                `json:"ignore-dbs,omitempty"`
		Method    *EnumPgMigrationMethod `json:"method,omitempty"`

		// Password for authentication with the server where to migrate data from
		Password *string `json:"password,omitempty"`

		// Port number of the server where to migrate data from
		Port int64 `json:"port"`

		// The server where to migrate data from is secured with SSL
		Ssl *bool `json:"ssl,omitempty"`

		// User name for authentication with the server where to migrate data from
		Username *string `json:"username,omitempty"`
	} `json:"migration,omitempty"`

	// MySQL-specific settings
	MysqlSettings *map[string]interface{} `json:"mysql-settings,omitempty"`

	// Subscription plan
	Plan *string `json:"plan,omitempty"`

	// Service is protected against termination and powering off
	TerminationProtection *bool `json:"termination-protection,omitempty"`
}

// UpdateDbaasServiceMysqlJSONBodyMaintenanceDow defines parameters for UpdateDbaasServiceMysql.
type UpdateDbaasServiceMysqlJSONBodyMaintenanceDow string

// CreateDbaasServiceOpensearchJSONBody defines parameters for CreateDbaasServiceOpensearch.
type CreateDbaasServiceOpensearchJSONBody struct {
	ForkFromService *DbaasServiceName `json:"fork-from-service,omitempty"`

	// Allows you to create glob style patterns and set a max number of indexes matching this pattern you want to keep. Creating indexes exceeding this value will cause the oldest one to get deleted. You could for example create a pattern looking like 'logs.?' and then create index logs.1, logs.2 etc, it will delete logs.1 once you create logs.6. Do note 'logs.?' does not apply to logs.10. Note: Setting max_index_count to 0 will do nothing and the pattern gets ignored.
	IndexPatterns *[]struct {
		// Maximum number of indexes to keep
		MaxIndexCount *int64 `json:"max-index-count,omitempty"`

		// fnmatch pattern
		Pattern *string `json:"pattern,omitempty"`

		// Deletion sorting algorithm
		SortingAlgorithm *CreateDbaasServiceOpensearchJSONBodyIndexPatternsSortingAlgorithm `json:"sorting-algorithm,omitempty"`
	} `json:"index-patterns,omitempty"`

	// Template settings for all new indexes
	IndexTemplate *struct {
		// The maximum number of nested JSON objects that a single document can contain across all nested types. This limit helps to prevent out of memory errors when a document contains too many nested objects. Default is 10000.
		MappingNestedObjectsLimit *int64 `json:"mapping-nested-objects-limit,omitempty"`

		// The number of replicas each primary shard has.
		NumberOfReplicas *int64 `json:"number-of-replicas,omitempty"`

		// The number of primary shards that an index should have.
		NumberOfShards *int64 `json:"number-of-shards,omitempty"`
	} `json:"index-template,omitempty"`

	// Service integrations to enable for the service. Some integration types affect how a service is created and they must be provided as part of the creation call instead of being defined later.
	Integrations *[]struct {
		DestService *DbaasServiceName `json:"dest-service,omitempty"`

		// Integration settings
		Settings      *map[string]interface{} `json:"settings,omitempty"`
		SourceService *DbaasServiceName       `json:"source-service,omitempty"`
		Type          EnumIntegrationTypes    `json:"type"`
	} `json:"integrations,omitempty"`

	// Allow incoming connections from CIDR address block, e.g. '10.20.0.0/16'
	IpFilter *[]string `json:"ip-filter,omitempty"`

	// Aiven automation resets index.refresh_interval to default value for every index to be sure that indices are always visible to search. If it doesn't fit your case, you can disable this by setting up this flag to true.
	KeepIndexRefreshInterval *bool `json:"keep-index-refresh-interval,omitempty"`

	// Automatic maintenance settings
	Maintenance *struct {
		// Day of week for installing updates
		Dow CreateDbaasServiceOpensearchJSONBodyMaintenanceDow `json:"dow"`

		// Time for installing updates, UTC
		Time string `json:"time"`
	} `json:"maintenance,omitempty"`

	// Maximum number of indexes to keep before deleting the oldest one
	MaxIndexCount *int64 `json:"max-index-count,omitempty"`

	// OpenSearch Dashboards settings
	OpensearchDashboards *struct {
		// Enable or disable OpenSearch Dashboards (default: true)
		Enabled *bool `json:"enabled,omitempty"`

		// Limits the maximum amount of memory (in MiB) the OpenSearch Dashboards process can use. This sets the max_old_space_size option of the nodejs running the OpenSearch Dashboards. Note: the memory reserved by OpenSearch Dashboards is not available for OpenSearch. (default: 128)
		MaxOldSpaceSize *int64 `json:"max-old-space-size,omitempty"`

		// Timeout in milliseconds for requests made by OpenSearch Dashboards towards OpenSearch (default: 30000)
		OpensearchRequestTimeout *int64 `json:"opensearch-request-timeout,omitempty"`
	} `json:"opensearch-dashboards,omitempty"`

	// OpenSearch-specific settings
	OpensearchSettings *map[string]interface{} `json:"opensearch-settings,omitempty"`

	// Subscription plan
	Plan string `json:"plan"`

	// Name of a backup to recover from for services that support backup names
	RecoveryBackupName *string `json:"recovery-backup-name,omitempty"`

	// Service is protected against termination and powering off
	TerminationProtection *bool `json:"termination-protection,omitempty"`

	// OpenSearch major version
	Version *string `json:"version,omitempty"`
}

// CreateDbaasServiceOpensearchJSONBodyIndexPatternsSortingAlgorithm defines parameters for CreateDbaasServiceOpensearch.
type CreateDbaasServiceOpensearchJSONBodyIndexPatternsSortingAlgorithm string

// CreateDbaasServiceOpensearchJSONBodyMaintenanceDow defines parameters for CreateDbaasServiceOpensearch.
type CreateDbaasServiceOpensearchJSONBodyMaintenanceDow string

// UpdateDbaasServiceOpensearchJSONBody defines parameters for UpdateDbaasServiceOpensearch.
type UpdateDbaasServiceOpensearchJSONBody struct {
	// Allows you to create glob style patterns and set a max number of indexes matching this pattern you want to keep. Creating indexes exceeding this value will cause the oldest one to get deleted. You could for example create a pattern looking like 'logs.?' and then create index logs.1, logs.2 etc, it will delete logs.1 once you create logs.6. Do note 'logs.?' does not apply to logs.10. Note: Setting max_index_count to 0 will do nothing and the pattern gets ignored.
	IndexPatterns *[]struct {
		// Maximum number of indexes to keep
		MaxIndexCount *int64 `json:"max-index-count,omitempty"`

		// fnmatch pattern
		Pattern *string `json:"pattern,omitempty"`

		// Deletion sorting algorithm
		SortingAlgorithm *UpdateDbaasServiceOpensearchJSONBodyIndexPatternsSortingAlgorithm `json:"sorting-algorithm,omitempty"`
	} `json:"index-patterns,omitempty"`

	// Template settings for all new indexes
	IndexTemplate *struct {
		// The maximum number of nested JSON objects that a single document can contain across all nested types. This limit helps to prevent out of memory errors when a document contains too many nested objects. Default is 10000.
		MappingNestedObjectsLimit *int64 `json:"mapping-nested-objects-limit,omitempty"`

		// The number of replicas each primary shard has.
		NumberOfReplicas *int64 `json:"number-of-replicas,omitempty"`

		// The number of primary shards that an index should have.
		NumberOfShards *int64 `json:"number-of-shards,omitempty"`
	} `json:"index-template,omitempty"`

	// Allow incoming connections from CIDR address block, e.g. '10.20.0.0/16'
	IpFilter *[]string `json:"ip-filter,omitempty"`

	// Aiven automation resets index.refresh_interval to default value for every index to be sure that indices are always visible to search. If it doesn't fit your case, you can disable this by setting up this flag to true.
	KeepIndexRefreshInterval *bool `json:"keep-index-refresh-interval,omitempty"`

	// Automatic maintenance settings
	Maintenance *struct {
		// Day of week for installing updates
		Dow UpdateDbaasServiceOpensearchJSONBodyMaintenanceDow `json:"dow"`

		// Time for installing updates, UTC
		Time string `json:"time"`
	} `json:"maintenance,omitempty"`

	// Maximum number of indexes to keep before deleting the oldest one
	MaxIndexCount *int64 `json:"max-index-count,omitempty"`

	// OpenSearch Dashboards settings
	OpensearchDashboards *struct {
		// Enable or disable OpenSearch Dashboards (default: true)
		Enabled *bool `json:"enabled,omitempty"`

		// Limits the maximum amount of memory (in MiB) the OpenSearch Dashboards process can use. This sets the max_old_space_size option of the nodejs running the OpenSearch Dashboards. Note: the memory reserved by OpenSearch Dashboards is not available for OpenSearch. (default: 128)
		MaxOldSpaceSize *int64 `json:"max-old-space-size,omitempty"`

		// Timeout in milliseconds for requests made by OpenSearch Dashboards towards OpenSearch (default: 30000)
		OpensearchRequestTimeout *int64 `json:"opensearch-request-timeout,omitempty"`
	} `json:"opensearch-dashboards,omitempty"`

	// OpenSearch-specific settings
	OpensearchSettings *map[string]interface{} `json:"opensearch-settings,omitempty"`

	// Subscription plan
	Plan *string `json:"plan,omitempty"`

	// Service is protected against termination and powering off
	TerminationProtection *bool `json:"termination-protection,omitempty"`
}

// UpdateDbaasServiceOpensearchJSONBodyIndexPatternsSortingAlgorithm defines parameters for UpdateDbaasServiceOpensearch.
type UpdateDbaasServiceOpensearchJSONBodyIndexPatternsSortingAlgorithm string

// UpdateDbaasServiceOpensearchJSONBodyMaintenanceDow defines parameters for UpdateDbaasServiceOpensearch.
type UpdateDbaasServiceOpensearchJSONBodyMaintenanceDow string

// CreateDbaasServicePgJSONBody defines parameters for CreateDbaasServicePg.
type CreateDbaasServicePgJSONBody struct {
	// Custom password for admin user. Defaults to random string. This must be set only when a new service is being created.
	AdminPassword *string `json:"admin-password,omitempty"`

	// Custom username for admin user. This must be set only when a new service is being created.
	AdminUsername  *string `json:"admin-username,omitempty"`
	BackupSchedule *struct {
		// The hour of day (in UTC) when backup for the service is started. New backup is only started if previous backup has already completed.
		BackupHour *int64 `json:"backup-hour,omitempty"`

		// The minute of an hour when backup for the service is started. New backup is only started if previous backup has already completed.
		BackupMinute *int64 `json:"backup-minute,omitempty"`
	} `json:"backup-schedule,omitempty"`
	ForkFromService *DbaasServiceName `json:"fork-from-service,omitempty"`

	// Service integrations to enable for the service. Some integration types affect how a service is created and they must be provided as part of the creation call instead of being defined later.
	Integrations *[]struct {
		DestService *DbaasServiceName `json:"dest-service,omitempty"`

		// Integration settings
		Settings      *map[string]interface{} `json:"settings,omitempty"`
		SourceService *DbaasServiceName       `json:"source-service,omitempty"`
		Type          EnumIntegrationTypes    `json:"type"`
	} `json:"integrations,omitempty"`

	// Allow incoming connections from CIDR address block, e.g. '10.20.0.0/16'
	IpFilter *[]string `json:"ip-filter,omitempty"`

	// Automatic maintenance settings
	Maintenance *struct {
		// Day of week for installing updates
		Dow CreateDbaasServicePgJSONBodyMaintenanceDow `json:"dow"`

		// Time for installing updates, UTC
		Time string `json:"time"`
	} `json:"maintenance,omitempty"`

	// Migrate data from existing server
	Migration *struct {
		// Database name for bootstrapping the initial connection
		Dbname *string `json:"dbname,omitempty"`

		// Hostname or IP address of the server where to migrate data from
		Host string `json:"host"`

		// Comma-separated list of databases, which should be ignored during migration (supported by MySQL only at the moment)
		IgnoreDbs *string                `json:"ignore-dbs,omitempty"`
		Method    *EnumPgMigrationMethod `json:"method,omitempty"`

		// Password for authentication with the server where to migrate data from
		Password *string `json:"password,omitempty"`

		// Port number of the server where to migrate data from
		Port int64 `json:"port"`

		// The server where to migrate data from is secured with SSL
		Ssl *bool `json:"ssl,omitempty"`

		// User name for authentication with the server where to migrate data from
		Username *string `json:"username,omitempty"`
	} `json:"migration,omitempty"`

	// PostgreSQL-specific settings
	PgSettings *map[string]interface{} `json:"pg-settings,omitempty"`

	// PGBouncer connection pooling settings
	PgbouncerSettings *map[string]interface{} `json:"pgbouncer-settings,omitempty"`

	// PGLookout settings
	PglookoutSettings *map[string]interface{} `json:"pglookout-settings,omitempty"`

	// Subscription plan
	Plan string `json:"plan"`

	// ISO time of a backup to recover from for services that support arbitrary times
	RecoveryBackupTime *string `json:"recovery-backup-time,omitempty"`

	// Percentage of total RAM that the database server uses for shared memory buffers. Valid range is 20-60 (float), which corresponds to 20% - 60%. This setting adjusts the shared_buffers configuration value.
	SharedBuffersPercentage *int64                        `json:"shared-buffers-percentage,omitempty"`
	SynchronousReplication  *EnumPgSynchronousReplication `json:"synchronous-replication,omitempty"`

	// Service is protected against termination and powering off
	TerminationProtection *bool `json:"termination-protection,omitempty"`

	// TimescaleDB extension configuration values
	TimescaledbSettings *map[string]interface{} `json:"timescaledb-settings,omitempty"`
	Variant             *EnumPgVariant          `json:"variant,omitempty"`

	// PostgreSQL major version
	Version *string `json:"version,omitempty"`

	// Sets the maximum amount of memory to be used by a query operation (such as a sort or hash table) before writing to temporary disk files, in MB. Default is 1MB + 0.075% of total RAM (up to 32MB).
	WorkMem *int64 `json:"work-mem,omitempty"`
}

// CreateDbaasServicePgJSONBodyMaintenanceDow defines parameters for CreateDbaasServicePg.
type CreateDbaasServicePgJSONBodyMaintenanceDow string

// UpdateDbaasServicePgJSONBody defines parameters for UpdateDbaasServicePg.
type UpdateDbaasServicePgJSONBody struct {
	BackupSchedule *struct {
		// The hour of day (in UTC) when backup for the service is started. New backup is only started if previous backup has already completed.
		BackupHour *int64 `json:"backup-hour,omitempty"`

		// The minute of an hour when backup for the service is started. New backup is only started if previous backup has already completed.
		BackupMinute *int64 `json:"backup-minute,omitempty"`
	} `json:"backup-schedule,omitempty"`

	// Allow incoming connections from CIDR address block, e.g. '10.20.0.0/16'
	IpFilter *[]string `json:"ip-filter,omitempty"`

	// Automatic maintenance settings
	Maintenance *struct {
		// Day of week for installing updates
		Dow UpdateDbaasServicePgJSONBodyMaintenanceDow `json:"dow"`

		// Time for installing updates, UTC
		Time string `json:"time"`
	} `json:"maintenance,omitempty"`

	// Migrate data from existing server
	Migration *struct {
		// Database name for bootstrapping the initial connection
		Dbname *string `json:"dbname,omitempty"`

		// Hostname or IP address of the server where to migrate data from
		Host string `json:"host"`

		// Comma-separated list of databases, which should be ignored during migration (supported by MySQL only at the moment)
		IgnoreDbs *string                `json:"ignore-dbs,omitempty"`
		Method    *EnumPgMigrationMethod `json:"method,omitempty"`

		// Password for authentication with the server where to migrate data from
		Password *string `json:"password,omitempty"`

		// Port number of the server where to migrate data from
		Port int64 `json:"port"`

		// The server where to migrate data from is secured with SSL
		Ssl *bool `json:"ssl,omitempty"`

		// User name for authentication with the server where to migrate data from
		Username *string `json:"username,omitempty"`
	} `json:"migration,omitempty"`

	// PostgreSQL-specific settings
	PgSettings *map[string]interface{} `json:"pg-settings,omitempty"`

	// PGBouncer connection pooling settings
	PgbouncerSettings *map[string]interface{} `json:"pgbouncer-settings,omitempty"`

	// PGLookout settings
	PglookoutSettings *map[string]interface{} `json:"pglookout-settings,omitempty"`

	// Subscription plan
	Plan *string `json:"plan,omitempty"`

	// Percentage of total RAM that the database server uses for shared memory buffers. Valid range is 20-60 (float), which corresponds to 20% - 60%. This setting adjusts the shared_buffers configuration value.
	SharedBuffersPercentage *int64                        `json:"shared-buffers-percentage,omitempty"`
	SynchronousReplication  *EnumPgSynchronousReplication `json:"synchronous-replication,omitempty"`

	// Service is protected against termination and powering off
	TerminationProtection *bool `json:"termination-protection,omitempty"`

	// TimescaleDB extension configuration values
	TimescaledbSettings *map[string]interface{} `json:"timescaledb-settings,omitempty"`
	Variant             *EnumPgVariant          `json:"variant,omitempty"`

	// Sets the maximum amount of memory to be used by a query operation (such as a sort or hash table) before writing to temporary disk files, in MB. Default is 1MB + 0.075% of total RAM (up to 32MB).
	WorkMem *int64 `json:"work-mem,omitempty"`
}

// UpdateDbaasServicePgJSONBodyMaintenanceDow defines parameters for UpdateDbaasServicePg.
type UpdateDbaasServicePgJSONBodyMaintenanceDow string

// CreateDbaasServiceRedisJSONBody defines parameters for CreateDbaasServiceRedis.
type CreateDbaasServiceRedisJSONBody struct {
	ForkFromService *DbaasServiceName `json:"fork-from-service,omitempty"`

	// Service integrations to enable for the service. Some integration types affect how a service is created and they must be provided as part of the creation call instead of being defined later.
	Integrations *[]struct {
		DestService *DbaasServiceName `json:"dest-service,omitempty"`

		// Integration settings
		Settings      *map[string]interface{} `json:"settings,omitempty"`
		SourceService *DbaasServiceName       `json:"source-service,omitempty"`
		Type          EnumIntegrationTypes    `json:"type"`
	} `json:"integrations,omitempty"`

	// Allow incoming connections from CIDR address block, e.g. '10.20.0.0/16'
	IpFilter *[]string `json:"ip-filter,omitempty"`

	// Automatic maintenance settings
	Maintenance *struct {
		// Day of week for installing updates
		Dow CreateDbaasServiceRedisJSONBodyMaintenanceDow `json:"dow"`

		// Time for installing updates, UTC
		Time string `json:"time"`
	} `json:"maintenance,omitempty"`

	// Migrate data from existing server
	Migration *struct {
		// Database name for bootstrapping the initial connection
		Dbname *string `json:"dbname,omitempty"`

		// Hostname or IP address of the server where to migrate data from
		Host string `json:"host"`

		// Comma-separated list of databases, which should be ignored during migration (supported by MySQL only at the moment)
		IgnoreDbs *string                `json:"ignore-dbs,omitempty"`
		Method    *EnumPgMigrationMethod `json:"method,omitempty"`

		// Password for authentication with the server where to migrate data from
		Password *string `json:"password,omitempty"`

		// Port number of the server where to migrate data from
		Port int64 `json:"port"`

		// The server where to migrate data from is secured with SSL
		Ssl *bool `json:"ssl,omitempty"`

		// User name for authentication with the server where to migrate data from
		Username *string `json:"username,omitempty"`
	} `json:"migration,omitempty"`

	// Subscription plan
	Plan string `json:"plan"`

	// Name of a backup to recover from for services that support backup names
	RecoveryBackupName *string `json:"recovery-backup-name,omitempty"`

	// Redis.conf settings
	RedisSettings *map[string]interface{} `json:"redis-settings,omitempty"`

	// Service is protected against termination and powering off
	TerminationProtection *bool `json:"termination-protection,omitempty"`
}

// CreateDbaasServiceRedisJSONBodyMaintenanceDow defines parameters for CreateDbaasServiceRedis.
type CreateDbaasServiceRedisJSONBodyMaintenanceDow string

// UpdateDbaasServiceRedisJSONBody defines parameters for UpdateDbaasServiceRedis.
type UpdateDbaasServiceRedisJSONBody struct {
	// Allow incoming connections from CIDR address block, e.g. '10.20.0.0/16'
	IpFilter *[]string `json:"ip-filter,omitempty"`

	// Automatic maintenance settings
	Maintenance *struct {
		// Day of week for installing updates
		Dow UpdateDbaasServiceRedisJSONBodyMaintenanceDow `json:"dow"`

		// Time for installing updates, UTC
		Time string `json:"time"`
	} `json:"maintenance,omitempty"`

	// Migrate data from existing server
	Migration *struct {
		// Database name for bootstrapping the initial connection
		Dbname *string `json:"dbname,omitempty"`

		// Hostname or IP address of the server where to migrate data from
		Host string `json:"host"`

		// Comma-separated list of databases, which should be ignored during migration (supported by MySQL only at the moment)
		IgnoreDbs *string                `json:"ignore-dbs,omitempty"`
		Method    *EnumPgMigrationMethod `json:"method,omitempty"`

		// Password for authentication with the server where to migrate data from
		Password *string `json:"password,omitempty"`

		// Port number of the server where to migrate data from
		Port int64 `json:"port"`

		// The server where to migrate data from is secured with SSL
		Ssl *bool `json:"ssl,omitempty"`

		// User name for authentication with the server where to migrate data from
		Username *string `json:"username,omitempty"`
	} `json:"migration,omitempty"`

	// Subscription plan
	Plan *string `json:"plan,omitempty"`

	// Redis.conf settings
	RedisSettings *map[string]interface{} `json:"redis-settings,omitempty"`

	// Service is protected against termination and powering off
	TerminationProtection *bool `json:"termination-protection,omitempty"`
}

// UpdateDbaasServiceRedisJSONBodyMaintenanceDow defines parameters for UpdateDbaasServiceRedis.
type UpdateDbaasServiceRedisJSONBodyMaintenanceDow string

// GetDbaasServiceLogsJSONBody defines parameters for GetDbaasServiceLogs.
type GetDbaasServiceLogsJSONBody struct {
	// How many log entries to receive at most, up to 500 (default: 100)
	Limit *int64 `json:"limit,omitempty"`

	// Opaque offset identifier
	Offset    *string        `json:"offset,omitempty"`
	SortOrder *EnumSortOrder `json:"sort-order,omitempty"`
}

// GetDbaasServiceMetricsJSONBody defines parameters for GetDbaasServiceMetrics.
type GetDbaasServiceMetricsJSONBody struct {
	// Metrics time period (default: hour)
	Period *GetDbaasServiceMetricsJSONBodyPeriod `json:"period,omitempty"`
}

// GetDbaasServiceMetricsJSONBodyPeriod defines parameters for GetDbaasServiceMetrics.
type GetDbaasServiceMetricsJSONBodyPeriod string

// CreateElasticIpJSONBody defines parameters for CreateElasticIp.
type CreateElasticIpJSONBody struct {
	// Elastic IP description
	Description *string `json:"description,omitempty"`

	// Elastic IP address healthcheck
	Healthcheck *ElasticIpHealthcheck `json:"healthcheck,omitempty"`
	Labels      *Labels               `json:"labels,omitempty"`
}

// UpdateElasticIpJSONBody defines parameters for UpdateElasticIp.
type UpdateElasticIpJSONBody struct {
	// Elastic IP description
	Description *string `json:"description"`

	// Elastic IP address healthcheck
	Healthcheck *ElasticIpHealthcheck `json:"healthcheck,omitempty"`
	Labels      *Labels               `json:"labels,omitempty"`
}

// ResetElasticIpFieldParamsField defines parameters for ResetElasticIpField.
type ResetElasticIpFieldParamsField string

// AttachInstanceToElasticIpJSONBody defines parameters for AttachInstanceToElasticIp.
type AttachInstanceToElasticIpJSONBody struct {
	// Instance
	Instance Instance `json:"instance"`
}

// DetachInstanceFromElasticIpJSONBody defines parameters for DetachInstanceFromElasticIp.
type DetachInstanceFromElasticIpJSONBody struct {
	// Instance
	Instance Instance `json:"instance"`
}

// ListEventsParams defines parameters for ListEvents.
type ListEventsParams struct {
	From *time.Time `json:"from,omitempty"`
	To   *time.Time `json:"to,omitempty"`
}

// ListInstancesParams defines parameters for ListInstances.
type ListInstancesParams struct {
	ManagerId   *string                         `json:"manager-id,omitempty"`
	ManagerType *ListInstancesParamsManagerType `json:"manager-type,omitempty"`
}

// ListInstancesParamsManagerType defines parameters for ListInstances.
type ListInstancesParamsManagerType string

// CreateInstanceJSONBody defines parameters for CreateInstance.
type CreateInstanceJSONBody struct {
	// Instance Anti-affinity Groups
	AntiAffinityGroups *[]AntiAffinityGroup `json:"anti-affinity-groups,omitempty"`

	// Deploy target
	DeployTarget *DeployTarget `json:"deploy-target,omitempty"`

	// Instance disk size in GB
	DiskSize int64 `json:"disk-size"`

	// Compute instance type
	InstanceType InstanceType `json:"instance-type"`

	// Enable IPv6
	Ipv6Enabled *bool   `json:"ipv6-enabled,omitempty"`
	Labels      *Labels `json:"labels,omitempty"`

	// Instance name
	Name *string `json:"name,omitempty"`

	// Instance Security Groups
	SecurityGroups *[]SecurityGroup `json:"security-groups,omitempty"`

	// SSH key
	SshKey *SshKey `json:"ssh-key,omitempty"`

	// Instance template
	Template Template `json:"template"`

	// Instance Cloud-init user-data
	UserData *string `json:"user-data,omitempty"`
}

// CreateInstancePoolJSONBody defines parameters for CreateInstancePool.
type CreateInstancePoolJSONBody struct {
	// Instance Pool Anti-affinity Groups
	AntiAffinityGroups *[]AntiAffinityGroup `json:"anti-affinity-groups,omitempty"`

	// Deploy target
	DeployTarget *DeployTarget `json:"deploy-target,omitempty"`

	// Instance Pool description
	Description *string `json:"description,omitempty"`

	// Instances disk size in GB
	DiskSize int64 `json:"disk-size"`

	// Instances Elastic IPs
	ElasticIps *[]ElasticIp `json:"elastic-ips,omitempty"`

	// Prefix to apply to instances names (default: pool)
	InstancePrefix *string `json:"instance-prefix,omitempty"`

	// Compute instance type
	InstanceType InstanceType `json:"instance-type"`

	// Enable IPv6 for instances
	Ipv6Enabled *bool   `json:"ipv6-enabled,omitempty"`
	Labels      *Labels `json:"labels,omitempty"`

	// Instance Pool name
	Name string `json:"name"`

	// Instance Pool Private Networks
	PrivateNetworks *[]PrivateNetwork `json:"private-networks,omitempty"`

	// Instance Pool Security Groups
	SecurityGroups *[]SecurityGroup `json:"security-groups,omitempty"`

	// Number of instances
	Size int64 `json:"size"`

	// SSH key
	SshKey *SshKey `json:"ssh-key,omitempty"`

	// Instance template
	Template Template `json:"template"`

	// Instances Cloud-init user-data
	UserData *string `json:"user-data,omitempty"`
}

// UpdateInstancePoolJSONBody defines parameters for UpdateInstancePool.
type UpdateInstancePoolJSONBody struct {
	// Instance Pool Anti-affinity Groups
	AntiAffinityGroups *[]AntiAffinityGroup `json:"anti-affinity-groups,omitempty"`

	// Deploy target
	DeployTarget *DeployTarget `json:"deploy-target,omitempty"`

	// Instance Pool description
	Description *string `json:"description"`

	// Instances disk size in GB
	DiskSize *int64 `json:"disk-size,omitempty"`

	// Instances Elastic IPs
	ElasticIps *[]ElasticIp `json:"elastic-ips,omitempty"`

	// Prefix to apply to instances names (default: pool)
	InstancePrefix *string `json:"instance-prefix,omitempty"`

	// Compute instance type
	InstanceType *InstanceType `json:"instance-type,omitempty"`

	// Enable IPv6 for instances
	Ipv6Enabled *bool   `json:"ipv6-enabled,omitempty"`
	Labels      *Labels `json:"labels,omitempty"`

	// Instance Pool name
	Name *string `json:"name,omitempty"`

	// Instance Pool Private Networks
	PrivateNetworks *[]PrivateNetwork `json:"private-networks,omitempty"`

	// Instance Pool Security Groups
	SecurityGroups *[]SecurityGroup `json:"security-groups,omitempty"`

	// SSH key
	SshKey *SshKey `json:"ssh-key,omitempty"`

	// Instance template
	Template *Template `json:"template,omitempty"`

	// Instances Cloud-init user-data
	UserData *string `json:"user-data"`
}

// ResetInstancePoolFieldParamsField defines parameters for ResetInstancePoolField.
type ResetInstancePoolFieldParamsField string

// EvictInstancePoolMembersJSONBody defines parameters for EvictInstancePoolMembers.
type EvictInstancePoolMembersJSONBody struct {
	Instances *[]string `json:"instances,omitempty"`
}

// ScaleInstancePoolJSONBody defines parameters for ScaleInstancePool.
type ScaleInstancePoolJSONBody struct {
	// Number of managed instances
	Size int64 `json:"size"`
}

// UpdateInstanceJSONBody defines parameters for UpdateInstance.
type UpdateInstanceJSONBody struct {
	Labels *Labels `json:"labels,omitempty"`

	// Instance name
	Name *string `json:"name,omitempty"`

	// Instance Cloud-init user-data
	UserData *string `json:"user-data,omitempty"`
}

// ResetInstanceFieldParamsField defines parameters for ResetInstanceField.
type ResetInstanceFieldParamsField string

// ResetInstanceJSONBody defines parameters for ResetInstance.
type ResetInstanceJSONBody struct {
	// Instance disk size in GB
	DiskSize *int64 `json:"disk-size,omitempty"`

	// Instance template
	Template *Template `json:"template,omitempty"`
}

// ResizeInstanceDiskJSONBody defines parameters for ResizeInstanceDisk.
type ResizeInstanceDiskJSONBody struct {
	// Instance disk size in GB
	DiskSize int64 `json:"disk-size"`
}

// ScaleInstanceJSONBody defines parameters for ScaleInstance.
type ScaleInstanceJSONBody struct {
	// Compute instance type
	InstanceType InstanceType `json:"instance-type"`
}

// StartInstanceJSONBody defines parameters for StartInstance.
type StartInstanceJSONBody struct {
	// Boot in Rescue Mode, using named profile (supported: netboot, netboot-efi)
	RescueProfile *StartInstanceJSONBodyRescueProfile `json:"rescue-profile,omitempty"`
}

// StartInstanceJSONBodyRescueProfile defines parameters for StartInstance.
type StartInstanceJSONBodyRescueProfile string

// RevertInstanceToSnapshotJSONBody defines parameters for RevertInstanceToSnapshot.
type RevertInstanceToSnapshotJSONBody struct {
	// Snapshot ID
	Id string `json:"id"`
}

// CreateLoadBalancerJSONBody defines parameters for CreateLoadBalancer.
type CreateLoadBalancerJSONBody struct {
	// Load Balancer description
	Description *string `json:"description,omitempty"`
	Labels      *Labels `json:"labels,omitempty"`

	// Load Balancer name
	Name string `json:"name"`
}

// UpdateLoadBalancerJSONBody defines parameters for UpdateLoadBalancer.
type UpdateLoadBalancerJSONBody struct {
	Description *string `json:"description"`
	Labels      *Labels `json:"labels,omitempty"`
	Name        *string `json:"name,omitempty"`
}

// AddServiceToLoadBalancerJSONBody defines parameters for AddServiceToLoadBalancer.
type AddServiceToLoadBalancerJSONBody struct {
	// Load Balancer Service description
	Description *string `json:"description,omitempty"`

	// Load Balancer Service healthcheck
	Healthcheck LoadBalancerServiceHealthcheck `json:"healthcheck"`

	// Instance Pool
	InstancePool InstancePool `json:"instance-pool"`

	// Load Balancer Service name
	Name string `json:"name"`

	// Port exposed on the Load Balancer's public IP
	Port int64 `json:"port"`

	// Network traffic protocol
	Protocol AddServiceToLoadBalancerJSONBodyProtocol `json:"protocol"`

	// Load balancing strategy
	Strategy AddServiceToLoadBalancerJSONBodyStrategy `json:"strategy"`

	// Port on which the network traffic will be forwarded to on the receiving instance
	TargetPort int64 `json:"target-port"`
}

// AddServiceToLoadBalancerJSONBodyProtocol defines parameters for AddServiceToLoadBalancer.
type AddServiceToLoadBalancerJSONBodyProtocol string

// AddServiceToLoadBalancerJSONBodyStrategy defines parameters for AddServiceToLoadBalancer.
type AddServiceToLoadBalancerJSONBodyStrategy string

// UpdateLoadBalancerServiceJSONBody defines parameters for UpdateLoadBalancerService.
type UpdateLoadBalancerServiceJSONBody struct {
	// Load Balancer Service description
	Description *string `json:"description"`

	// Load Balancer Service healthcheck
	Healthcheck *LoadBalancerServiceHealthcheck `json:"healthcheck,omitempty"`

	// Load Balancer Service name
	Name *string `json:"name,omitempty"`

	// Port exposed on the Load Balancer's public IP
	Port *int64 `json:"port,omitempty"`

	// Network traffic protocol
	Protocol *UpdateLoadBalancerServiceJSONBodyProtocol `json:"protocol,omitempty"`

	// Load balancing strategy
	Strategy *UpdateLoadBalancerServiceJSONBodyStrategy `json:"strategy,omitempty"`

	// Port on which the network traffic will be forwarded to on the receiving instance
	TargetPort *int64 `json:"target-port,omitempty"`
}

// UpdateLoadBalancerServiceJSONBodyProtocol defines parameters for UpdateLoadBalancerService.
type UpdateLoadBalancerServiceJSONBodyProtocol string

// UpdateLoadBalancerServiceJSONBodyStrategy defines parameters for UpdateLoadBalancerService.
type UpdateLoadBalancerServiceJSONBodyStrategy string

// ResetLoadBalancerServiceFieldParamsField defines parameters for ResetLoadBalancerServiceField.
type ResetLoadBalancerServiceFieldParamsField string

// ResetLoadBalancerFieldParamsField defines parameters for ResetLoadBalancerField.
type ResetLoadBalancerFieldParamsField string

// CreatePrivateNetworkJSONBody defines parameters for CreatePrivateNetwork.
type CreatePrivateNetworkJSONBody struct {
	// Private Network description
	Description *string `json:"description,omitempty"`

	// Private Network end IP address
	EndIp  *string `json:"end-ip,omitempty"`
	Labels *Labels `json:"labels,omitempty"`

	// Private Network name
	Name string `json:"name"`

	// Private Network netmask
	Netmask *string `json:"netmask,omitempty"`

	// Private Network start IP address
	StartIp *string `json:"start-ip,omitempty"`
}

// UpdatePrivateNetworkJSONBody defines parameters for UpdatePrivateNetwork.
type UpdatePrivateNetworkJSONBody struct {
	// Private Network description
	Description *string `json:"description,omitempty"`

	// Private Network end IP address
	EndIp  *string `json:"end-ip,omitempty"`
	Labels *Labels `json:"labels,omitempty"`

	// Private Network name
	Name *string `json:"name,omitempty"`

	// Private Network netmask
	Netmask *string `json:"netmask,omitempty"`

	// Private Network start IP address
	StartIp *string `json:"start-ip,omitempty"`
}

// ResetPrivateNetworkFieldParamsField defines parameters for ResetPrivateNetworkField.
type ResetPrivateNetworkFieldParamsField string

// AttachInstanceToPrivateNetworkJSONBody defines parameters for AttachInstanceToPrivateNetwork.
type AttachInstanceToPrivateNetworkJSONBody struct {
	// Instance
	Instance Instance `json:"instance"`

	// Static IP address lease for the corresponding network interface
	Ip *string `json:"ip,omitempty"`
}

// DetachInstanceFromPrivateNetworkJSONBody defines parameters for DetachInstanceFromPrivateNetwork.
type DetachInstanceFromPrivateNetworkJSONBody struct {
	// Instance
	Instance Instance `json:"instance"`
}

// UpdatePrivateNetworkInstanceIpJSONBody defines parameters for UpdatePrivateNetworkInstanceIp.
type UpdatePrivateNetworkInstanceIpJSONBody struct {
	// Instance
	Instance Instance `json:"instance"`

	// Static IP address lease for the corresponding network interface
	Ip *string `json:"ip,omitempty"`
}

// CreateSecurityGroupJSONBody defines parameters for CreateSecurityGroup.
type CreateSecurityGroupJSONBody struct {
	// Security Group description
	Description *string `json:"description,omitempty"`

	// Security Group name
	Name string `json:"name"`
}

// AddRuleToSecurityGroupJSONBody defines parameters for AddRuleToSecurityGroup.
type AddRuleToSecurityGroupJSONBody struct {
	// Security Group rule description
	Description *string `json:"description,omitempty"`

	// End port of the range
	EndPort *int64 `json:"end-port,omitempty"`

	// Network flow direction to match
	FlowDirection AddRuleToSecurityGroupJSONBodyFlowDirection `json:"flow-direction"`

	// ICMP details (default: -1 (ANY))
	Icmp *struct {
		Code *int64 `json:"code,omitempty"`
		Type *int64 `json:"type,omitempty"`
	} `json:"icmp,omitempty"`

	// CIDR-formatted network allowed
	Network *string `json:"network,omitempty"`

	// Network protocol
	Protocol AddRuleToSecurityGroupJSONBodyProtocol `json:"protocol"`

	// Security Group
	SecurityGroup *SecurityGroupResource `json:"security-group,omitempty"`

	// Start port of the range
	StartPort *int64 `json:"start-port,omitempty"`
}

// AddRuleToSecurityGroupJSONBodyFlowDirection defines parameters for AddRuleToSecurityGroup.
type AddRuleToSecurityGroupJSONBodyFlowDirection string

// AddRuleToSecurityGroupJSONBodyProtocol defines parameters for AddRuleToSecurityGroup.
type AddRuleToSecurityGroupJSONBodyProtocol string

// AddExternalSourceToSecurityGroupJSONBody defines parameters for AddExternalSourceToSecurityGroup.
type AddExternalSourceToSecurityGroupJSONBody struct {
	// CIDR-formatted network to add
	Cidr string `json:"cidr"`
}

// AttachInstanceToSecurityGroupJSONBody defines parameters for AttachInstanceToSecurityGroup.
type AttachInstanceToSecurityGroupJSONBody struct {
	// Instance
	Instance Instance `json:"instance"`
}

// DetachInstanceFromSecurityGroupJSONBody defines parameters for DetachInstanceFromSecurityGroup.
type DetachInstanceFromSecurityGroupJSONBody struct {
	// Instance
	Instance Instance `json:"instance"`
}

// RemoveExternalSourceFromSecurityGroupJSONBody defines parameters for RemoveExternalSourceFromSecurityGroup.
type RemoveExternalSourceFromSecurityGroupJSONBody struct {
	// CIDR-formatted network to remove
	Cidr string `json:"cidr"`
}

// CreateSksClusterJSONBody defines parameters for CreateSksCluster.
type CreateSksClusterJSONBody struct {
	// Cluster addons
	Addons *[]CreateSksClusterJSONBodyAddons `json:"addons,omitempty"`

	// Enable auto upgrade of the control plane to the latest patch version available
	AutoUpgrade *bool `json:"auto-upgrade,omitempty"`

	// Cluster CNI
	Cni *CreateSksClusterJSONBodyCni `json:"cni,omitempty"`

	// Cluster description
	Description *string `json:"description,omitempty"`
	Labels      *Labels `json:"labels,omitempty"`

	// Cluster service level
	Level CreateSksClusterJSONBodyLevel `json:"level"`

	// Cluster name
	Name string `json:"name"`

	// SKS Cluster OpenID config map
	Oidc *SksOidc `json:"oidc,omitempty"`

	// Control plane Kubernetes version
	Version string `json:"version"`
}

// CreateSksClusterJSONBodyAddons defines parameters for CreateSksCluster.
type CreateSksClusterJSONBodyAddons string

// CreateSksClusterJSONBodyCni defines parameters for CreateSksCluster.
type CreateSksClusterJSONBodyCni string

// CreateSksClusterJSONBodyLevel defines parameters for CreateSksCluster.
type CreateSksClusterJSONBodyLevel string

// GenerateSksClusterKubeconfigJSONBody defines parameters for GenerateSksClusterKubeconfig.
type GenerateSksClusterKubeconfigJSONBody SksKubeconfigRequest

// ListSksClusterVersionsParams defines parameters for ListSksClusterVersions.
type ListSksClusterVersionsParams struct {
	IncludeDeprecated *string `json:"include-deprecated,omitempty"`
}

// UpdateSksClusterJSONBody defines parameters for UpdateSksCluster.
type UpdateSksClusterJSONBody struct {
	// Enable auto upgrade of the control plane to the latest patch version available
	AutoUpgrade *bool `json:"auto-upgrade,omitempty"`

	// Cluster description
	Description *string `json:"description"`
	Labels      *Labels `json:"labels,omitempty"`

	// Cluster name
	Name *string `json:"name,omitempty"`
}

// GetSksClusterAuthorityCertParamsAuthority defines parameters for GetSksClusterAuthorityCert.
type GetSksClusterAuthorityCertParamsAuthority string

// CreateSksNodepoolJSONBody defines parameters for CreateSksNodepool.
type CreateSksNodepoolJSONBody struct {
	// Nodepool addons
	Addons *[]CreateSksNodepoolJSONBodyAddons `json:"addons,omitempty"`

	// Nodepool Anti-affinity Groups
	AntiAffinityGroups *[]AntiAffinityGroup `json:"anti-affinity-groups,omitempty"`

	// Deploy target
	DeployTarget *DeployTarget `json:"deploy-target,omitempty"`

	// Nodepool description
	Description *string `json:"description,omitempty"`

	// Nodepool instances disk size in GB
	DiskSize int64 `json:"disk-size"`

	// Prefix to apply to instances names (default: pool)
	InstancePrefix *string `json:"instance-prefix,omitempty"`

	// Compute instance type
	InstanceType InstanceType `json:"instance-type"`
	Labels       *Labels      `json:"labels,omitempty"`

	// Nodepool name
	Name string `json:"name"`

	// Nodepool Private Networks
	PrivateNetworks *[]PrivateNetwork `json:"private-networks,omitempty"`

	// Nodepool Security Groups
	SecurityGroups *[]SecurityGroup `json:"security-groups,omitempty"`

	// Number of instances
	Size   int64              `json:"size"`
	Taints *SksNodepoolTaints `json:"taints,omitempty"`
}

// CreateSksNodepoolJSONBodyAddons defines parameters for CreateSksNodepool.
type CreateSksNodepoolJSONBodyAddons string

// UpdateSksNodepoolJSONBody defines parameters for UpdateSksNodepool.
type UpdateSksNodepoolJSONBody struct {
	// Nodepool Anti-affinity Groups
	AntiAffinityGroups *[]AntiAffinityGroup `json:"anti-affinity-groups,omitempty"`

	// Deploy target
	DeployTarget *DeployTarget `json:"deploy-target,omitempty"`

	// Nodepool description
	Description *string `json:"description"`

	// Nodepool instances disk size in GB
	DiskSize *int64 `json:"disk-size,omitempty"`

	// Prefix to apply to managed instances names (default: pool)
	InstancePrefix *string `json:"instance-prefix,omitempty"`

	// Compute instance type
	InstanceType *InstanceType `json:"instance-type,omitempty"`
	Labels       *Labels       `json:"labels,omitempty"`

	// Nodepool name
	Name *string `json:"name,omitempty"`

	// Nodepool Private Networks
	PrivateNetworks *[]PrivateNetwork `json:"private-networks,omitempty"`

	// Nodepool Security Groups
	SecurityGroups *[]SecurityGroup   `json:"security-groups,omitempty"`
	Taints         *SksNodepoolTaints `json:"taints,omitempty"`
}

// ResetSksNodepoolFieldParamsField defines parameters for ResetSksNodepoolField.
type ResetSksNodepoolFieldParamsField string

// EvictSksNodepoolMembersJSONBody defines parameters for EvictSksNodepoolMembers.
type EvictSksNodepoolMembersJSONBody struct {
	Instances *[]string `json:"instances,omitempty"`
}

// ScaleSksNodepoolJSONBody defines parameters for ScaleSksNodepool.
type ScaleSksNodepoolJSONBody struct {
	// Number of instances
	Size int64 `json:"size"`
}

// UpgradeSksClusterJSONBody defines parameters for UpgradeSksCluster.
type UpgradeSksClusterJSONBody struct {
	// Control plane Kubernetes version
	Version string `json:"version"`
}

// ResetSksClusterFieldParamsField defines parameters for ResetSksClusterField.
type ResetSksClusterFieldParamsField string

// PromoteSnapshotToTemplateJSONBody defines parameters for PromoteSnapshotToTemplate.
type PromoteSnapshotToTemplateJSONBody struct {
	// Template default user
	DefaultUser *string `json:"default-user,omitempty"`

	// Template name
	Name string `json:"name"`

	// Enable password-based login in the template
	PasswordEnabled *bool `json:"password-enabled,omitempty"`

	// Enable SSH key-based login in the template
	SshKeyEnabled *bool `json:"ssh-key-enabled,omitempty"`
}

// GetSosPresignedUrlParams defines parameters for GetSosPresignedUrl.
type GetSosPresignedUrlParams struct {
	Key *string `json:"key,omitempty"`
}

// RegisterSshKeyJSONBody defines parameters for RegisterSshKey.
type RegisterSshKeyJSONBody struct {
	// Private Network name
	Name string `json:"name"`

	// Public key value
	PublicKey string `json:"public-key"`
}

// ListTemplatesParams defines parameters for ListTemplates.
type ListTemplatesParams struct {
	Visibility *ListTemplatesParamsVisibility `json:"visibility,omitempty"`
	Family     *string                        `json:"family,omitempty"`
}

// ListTemplatesParamsVisibility defines parameters for ListTemplates.
type ListTemplatesParamsVisibility string

// RegisterTemplateJSONBody defines parameters for RegisterTemplate.
type RegisterTemplateJSONBody struct {
	// Boot mode (default: legacy)
	BootMode *RegisterTemplateJSONBodyBootMode `json:"boot-mode,omitempty"`

	// Template MD5 checksum
	Checksum string `json:"checksum"`

	// Template default user
	DefaultUser *string `json:"default-user,omitempty"`

	// Template description
	Description *string `json:"description,omitempty"`

	// Template name
	Name string `json:"name"`

	// Enable password-based login
	PasswordEnabled bool `json:"password-enabled"`

	// Template size
	Size *int64 `json:"size,omitempty"`

	// Enable SSH key-based login
	SshKeyEnabled bool `json:"ssh-key-enabled"`

	// Template source URL
	Url string `json:"url"`
}

// RegisterTemplateJSONBodyBootMode defines parameters for RegisterTemplate.
type RegisterTemplateJSONBodyBootMode string

// CopyTemplateJSONBody defines parameters for CopyTemplate.
type CopyTemplateJSONBody struct {
	// Zone
	TargetZone Zone `json:"target-zone"`
}

// UpdateTemplateJSONBody defines parameters for UpdateTemplate.
type UpdateTemplateJSONBody struct {
	// Template Description
	Description *string `json:"description"`

	// Template name
	Name *string `json:"name,omitempty"`
}

// CreateAccessKeyJSONRequestBody defines body for CreateAccessKey for application/json ContentType.
type CreateAccessKeyJSONRequestBody CreateAccessKeyJSONBody

// CreateAntiAffinityGroupJSONRequestBody defines body for CreateAntiAffinityGroup for application/json ContentType.
type CreateAntiAffinityGroupJSONRequestBody CreateAntiAffinityGroupJSONBody

// CreateDbaasServiceKafkaJSONRequestBody defines body for CreateDbaasServiceKafka for application/json ContentType.
type CreateDbaasServiceKafkaJSONRequestBody CreateDbaasServiceKafkaJSONBody

// UpdateDbaasServiceKafkaJSONRequestBody defines body for UpdateDbaasServiceKafka for application/json ContentType.
type UpdateDbaasServiceKafkaJSONRequestBody UpdateDbaasServiceKafkaJSONBody

// CreateDbaasServiceMysqlJSONRequestBody defines body for CreateDbaasServiceMysql for application/json ContentType.
type CreateDbaasServiceMysqlJSONRequestBody CreateDbaasServiceMysqlJSONBody

// UpdateDbaasServiceMysqlJSONRequestBody defines body for UpdateDbaasServiceMysql for application/json ContentType.
type UpdateDbaasServiceMysqlJSONRequestBody UpdateDbaasServiceMysqlJSONBody

// CreateDbaasServiceOpensearchJSONRequestBody defines body for CreateDbaasServiceOpensearch for application/json ContentType.
type CreateDbaasServiceOpensearchJSONRequestBody CreateDbaasServiceOpensearchJSONBody

// UpdateDbaasServiceOpensearchJSONRequestBody defines body for UpdateDbaasServiceOpensearch for application/json ContentType.
type UpdateDbaasServiceOpensearchJSONRequestBody UpdateDbaasServiceOpensearchJSONBody

// CreateDbaasServicePgJSONRequestBody defines body for CreateDbaasServicePg for application/json ContentType.
type CreateDbaasServicePgJSONRequestBody CreateDbaasServicePgJSONBody

// UpdateDbaasServicePgJSONRequestBody defines body for UpdateDbaasServicePg for application/json ContentType.
type UpdateDbaasServicePgJSONRequestBody UpdateDbaasServicePgJSONBody

// CreateDbaasServiceRedisJSONRequestBody defines body for CreateDbaasServiceRedis for application/json ContentType.
type CreateDbaasServiceRedisJSONRequestBody CreateDbaasServiceRedisJSONBody

// UpdateDbaasServiceRedisJSONRequestBody defines body for UpdateDbaasServiceRedis for application/json ContentType.
type UpdateDbaasServiceRedisJSONRequestBody UpdateDbaasServiceRedisJSONBody

// GetDbaasServiceLogsJSONRequestBody defines body for GetDbaasServiceLogs for application/json ContentType.
type GetDbaasServiceLogsJSONRequestBody GetDbaasServiceLogsJSONBody

// GetDbaasServiceMetricsJSONRequestBody defines body for GetDbaasServiceMetrics for application/json ContentType.
type GetDbaasServiceMetricsJSONRequestBody GetDbaasServiceMetricsJSONBody

// CreateElasticIpJSONRequestBody defines body for CreateElasticIp for application/json ContentType.
type CreateElasticIpJSONRequestBody CreateElasticIpJSONBody

// UpdateElasticIpJSONRequestBody defines body for UpdateElasticIp for application/json ContentType.
type UpdateElasticIpJSONRequestBody UpdateElasticIpJSONBody

// AttachInstanceToElasticIpJSONRequestBody defines body for AttachInstanceToElasticIp for application/json ContentType.
type AttachInstanceToElasticIpJSONRequestBody AttachInstanceToElasticIpJSONBody

// DetachInstanceFromElasticIpJSONRequestBody defines body for DetachInstanceFromElasticIp for application/json ContentType.
type DetachInstanceFromElasticIpJSONRequestBody DetachInstanceFromElasticIpJSONBody

// CreateInstanceJSONRequestBody defines body for CreateInstance for application/json ContentType.
type CreateInstanceJSONRequestBody CreateInstanceJSONBody

// CreateInstancePoolJSONRequestBody defines body for CreateInstancePool for application/json ContentType.
type CreateInstancePoolJSONRequestBody CreateInstancePoolJSONBody

// UpdateInstancePoolJSONRequestBody defines body for UpdateInstancePool for application/json ContentType.
type UpdateInstancePoolJSONRequestBody UpdateInstancePoolJSONBody

// EvictInstancePoolMembersJSONRequestBody defines body for EvictInstancePoolMembers for application/json ContentType.
type EvictInstancePoolMembersJSONRequestBody EvictInstancePoolMembersJSONBody

// ScaleInstancePoolJSONRequestBody defines body for ScaleInstancePool for application/json ContentType.
type ScaleInstancePoolJSONRequestBody ScaleInstancePoolJSONBody

// UpdateInstanceJSONRequestBody defines body for UpdateInstance for application/json ContentType.
type UpdateInstanceJSONRequestBody UpdateInstanceJSONBody

// ResetInstanceJSONRequestBody defines body for ResetInstance for application/json ContentType.
type ResetInstanceJSONRequestBody ResetInstanceJSONBody

// ResizeInstanceDiskJSONRequestBody defines body for ResizeInstanceDisk for application/json ContentType.
type ResizeInstanceDiskJSONRequestBody ResizeInstanceDiskJSONBody

// ScaleInstanceJSONRequestBody defines body for ScaleInstance for application/json ContentType.
type ScaleInstanceJSONRequestBody ScaleInstanceJSONBody

// StartInstanceJSONRequestBody defines body for StartInstance for application/json ContentType.
type StartInstanceJSONRequestBody StartInstanceJSONBody

// RevertInstanceToSnapshotJSONRequestBody defines body for RevertInstanceToSnapshot for application/json ContentType.
type RevertInstanceToSnapshotJSONRequestBody RevertInstanceToSnapshotJSONBody

// CreateLoadBalancerJSONRequestBody defines body for CreateLoadBalancer for application/json ContentType.
type CreateLoadBalancerJSONRequestBody CreateLoadBalancerJSONBody

// UpdateLoadBalancerJSONRequestBody defines body for UpdateLoadBalancer for application/json ContentType.
type UpdateLoadBalancerJSONRequestBody UpdateLoadBalancerJSONBody

// AddServiceToLoadBalancerJSONRequestBody defines body for AddServiceToLoadBalancer for application/json ContentType.
type AddServiceToLoadBalancerJSONRequestBody AddServiceToLoadBalancerJSONBody

// UpdateLoadBalancerServiceJSONRequestBody defines body for UpdateLoadBalancerService for application/json ContentType.
type UpdateLoadBalancerServiceJSONRequestBody UpdateLoadBalancerServiceJSONBody

// CreatePrivateNetworkJSONRequestBody defines body for CreatePrivateNetwork for application/json ContentType.
type CreatePrivateNetworkJSONRequestBody CreatePrivateNetworkJSONBody

// UpdatePrivateNetworkJSONRequestBody defines body for UpdatePrivateNetwork for application/json ContentType.
type UpdatePrivateNetworkJSONRequestBody UpdatePrivateNetworkJSONBody

// AttachInstanceToPrivateNetworkJSONRequestBody defines body for AttachInstanceToPrivateNetwork for application/json ContentType.
type AttachInstanceToPrivateNetworkJSONRequestBody AttachInstanceToPrivateNetworkJSONBody

// DetachInstanceFromPrivateNetworkJSONRequestBody defines body for DetachInstanceFromPrivateNetwork for application/json ContentType.
type DetachInstanceFromPrivateNetworkJSONRequestBody DetachInstanceFromPrivateNetworkJSONBody

// UpdatePrivateNetworkInstanceIpJSONRequestBody defines body for UpdatePrivateNetworkInstanceIp for application/json ContentType.
type UpdatePrivateNetworkInstanceIpJSONRequestBody UpdatePrivateNetworkInstanceIpJSONBody

// CreateSecurityGroupJSONRequestBody defines body for CreateSecurityGroup for application/json ContentType.
type CreateSecurityGroupJSONRequestBody CreateSecurityGroupJSONBody

// AddRuleToSecurityGroupJSONRequestBody defines body for AddRuleToSecurityGroup for application/json ContentType.
type AddRuleToSecurityGroupJSONRequestBody AddRuleToSecurityGroupJSONBody

// AddExternalSourceToSecurityGroupJSONRequestBody defines body for AddExternalSourceToSecurityGroup for application/json ContentType.
type AddExternalSourceToSecurityGroupJSONRequestBody AddExternalSourceToSecurityGroupJSONBody

// AttachInstanceToSecurityGroupJSONRequestBody defines body for AttachInstanceToSecurityGroup for application/json ContentType.
type AttachInstanceToSecurityGroupJSONRequestBody AttachInstanceToSecurityGroupJSONBody

// DetachInstanceFromSecurityGroupJSONRequestBody defines body for DetachInstanceFromSecurityGroup for application/json ContentType.
type DetachInstanceFromSecurityGroupJSONRequestBody DetachInstanceFromSecurityGroupJSONBody

// RemoveExternalSourceFromSecurityGroupJSONRequestBody defines body for RemoveExternalSourceFromSecurityGroup for application/json ContentType.
type RemoveExternalSourceFromSecurityGroupJSONRequestBody RemoveExternalSourceFromSecurityGroupJSONBody

// CreateSksClusterJSONRequestBody defines body for CreateSksCluster for application/json ContentType.
type CreateSksClusterJSONRequestBody CreateSksClusterJSONBody

// GenerateSksClusterKubeconfigJSONRequestBody defines body for GenerateSksClusterKubeconfig for application/json ContentType.
type GenerateSksClusterKubeconfigJSONRequestBody GenerateSksClusterKubeconfigJSONBody

// UpdateSksClusterJSONRequestBody defines body for UpdateSksCluster for application/json ContentType.
type UpdateSksClusterJSONRequestBody UpdateSksClusterJSONBody

// CreateSksNodepoolJSONRequestBody defines body for CreateSksNodepool for application/json ContentType.
type CreateSksNodepoolJSONRequestBody CreateSksNodepoolJSONBody

// UpdateSksNodepoolJSONRequestBody defines body for UpdateSksNodepool for application/json ContentType.
type UpdateSksNodepoolJSONRequestBody UpdateSksNodepoolJSONBody

// EvictSksNodepoolMembersJSONRequestBody defines body for EvictSksNodepoolMembers for application/json ContentType.
type EvictSksNodepoolMembersJSONRequestBody EvictSksNodepoolMembersJSONBody

// ScaleSksNodepoolJSONRequestBody defines body for ScaleSksNodepool for application/json ContentType.
type ScaleSksNodepoolJSONRequestBody ScaleSksNodepoolJSONBody

// UpgradeSksClusterJSONRequestBody defines body for UpgradeSksCluster for application/json ContentType.
type UpgradeSksClusterJSONRequestBody UpgradeSksClusterJSONBody

// PromoteSnapshotToTemplateJSONRequestBody defines body for PromoteSnapshotToTemplate for application/json ContentType.
type PromoteSnapshotToTemplateJSONRequestBody PromoteSnapshotToTemplateJSONBody

// RegisterSshKeyJSONRequestBody defines body for RegisterSshKey for application/json ContentType.
type RegisterSshKeyJSONRequestBody RegisterSshKeyJSONBody

// RegisterTemplateJSONRequestBody defines body for RegisterTemplate for application/json ContentType.
type RegisterTemplateJSONRequestBody RegisterTemplateJSONBody

// CopyTemplateJSONRequestBody defines body for CopyTemplate for application/json ContentType.
type CopyTemplateJSONRequestBody CopyTemplateJSONBody

// UpdateTemplateJSONRequestBody defines body for UpdateTemplate for application/json ContentType.
type UpdateTemplateJSONRequestBody UpdateTemplateJSONBody

// Getter for additional properties for Labels. Returns the specified
// element and whether it was found
func (a Labels) Get(fieldName string) (value string, found bool) {
	if a.AdditionalProperties != nil {
		value, found = a.AdditionalProperties[fieldName]
	}
	return
}

// Setter for additional properties for Labels
func (a *Labels) Set(fieldName string, value string) {
	if a.AdditionalProperties == nil {
		a.AdditionalProperties = make(map[string]string)
	}
	a.AdditionalProperties[fieldName] = value
}

// Override default JSON handling for Labels to handle AdditionalProperties
func (a *Labels) UnmarshalJSON(b []byte) error {
	object := make(map[string]json.RawMessage)
	err := json.Unmarshal(b, &object)
	if err != nil {
		return err
	}

	if len(object) != 0 {
		a.AdditionalProperties = make(map[string]string)
		for fieldName, fieldBuf := range object {
			var fieldVal string
			err := json.Unmarshal(fieldBuf, &fieldVal)
			if err != nil {
				return fmt.Errorf("error unmarshaling field %s: %w", fieldName, err)
			}
			a.AdditionalProperties[fieldName] = fieldVal
		}
	}
	return nil
}

// Override default JSON handling for Labels to handle AdditionalProperties
func (a Labels) MarshalJSON() ([]byte, error) {
	var err error
	object := make(map[string]json.RawMessage)

	for fieldName, field := range a.AdditionalProperties {
		object[fieldName], err = json.Marshal(field)
		if err != nil {
			return nil, fmt.Errorf("error marshaling '%s': %w", fieldName, err)
		}
	}
	return json.Marshal(object)
}

// Getter for additional properties for SksClusterDeprecatedResource. Returns the specified
// element and whether it was found
func (a SksClusterDeprecatedResource) Get(fieldName string) (value string, found bool) {
	if a.AdditionalProperties != nil {
		value, found = a.AdditionalProperties[fieldName]
	}
	return
}

// Setter for additional properties for SksClusterDeprecatedResource
func (a *SksClusterDeprecatedResource) Set(fieldName string, value string) {
	if a.AdditionalProperties == nil {
		a.AdditionalProperties = make(map[string]string)
	}
	a.AdditionalProperties[fieldName] = value
}

// Override default JSON handling for SksClusterDeprecatedResource to handle AdditionalProperties
func (a *SksClusterDeprecatedResource) UnmarshalJSON(b []byte) error {
	object := make(map[string]json.RawMessage)
	err := json.Unmarshal(b, &object)
	if err != nil {
		return err
	}

	if len(object) != 0 {
		a.AdditionalProperties = make(map[string]string)
		for fieldName, fieldBuf := range object {
			var fieldVal string
			err := json.Unmarshal(fieldBuf, &fieldVal)
			if err != nil {
				return fmt.Errorf("error unmarshaling field %s: %w", fieldName, err)
			}
			a.AdditionalProperties[fieldName] = fieldVal
		}
	}
	return nil
}

// Override default JSON handling for SksClusterDeprecatedResource to handle AdditionalProperties
func (a SksClusterDeprecatedResource) MarshalJSON() ([]byte, error) {
	var err error
	object := make(map[string]json.RawMessage)

	for fieldName, field := range a.AdditionalProperties {
		object[fieldName], err = json.Marshal(field)
		if err != nil {
			return nil, fmt.Errorf("error marshaling '%s': %w", fieldName, err)
		}
	}
	return json.Marshal(object)
}

// Getter for additional properties for SksNodepoolTaints. Returns the specified
// element and whether it was found
func (a SksNodepoolTaints) Get(fieldName string) (value SksNodepoolTaint, found bool) {
	if a.AdditionalProperties != nil {
		value, found = a.AdditionalProperties[fieldName]
	}
	return
}

// Setter for additional properties for SksNodepoolTaints
func (a *SksNodepoolTaints) Set(fieldName string, value SksNodepoolTaint) {
	if a.AdditionalProperties == nil {
		a.AdditionalProperties = make(map[string]SksNodepoolTaint)
	}
	a.AdditionalProperties[fieldName] = value
}

// Override default JSON handling for SksNodepoolTaints to handle AdditionalProperties
func (a *SksNodepoolTaints) UnmarshalJSON(b []byte) error {
	object := make(map[string]json.RawMessage)
	err := json.Unmarshal(b, &object)
	if err != nil {
		return err
	}

	if len(object) != 0 {
		a.AdditionalProperties = make(map[string]SksNodepoolTaint)
		for fieldName, fieldBuf := range object {
			var fieldVal SksNodepoolTaint
			err := json.Unmarshal(fieldBuf, &fieldVal)
			if err != nil {
				return fmt.Errorf("error unmarshaling field %s: %w", fieldName, err)
			}
			a.AdditionalProperties[fieldName] = fieldVal
		}
	}
	return nil
}

// Override default JSON handling for SksNodepoolTaints to handle AdditionalProperties
func (a SksNodepoolTaints) MarshalJSON() ([]byte, error) {
	var err error
	object := make(map[string]json.RawMessage)

	for fieldName, field := range a.AdditionalProperties {
		object[fieldName], err = json.Marshal(field)
		if err != nil {
			return nil, fmt.Errorf("error marshaling '%s': %w", fieldName, err)
		}
	}
	return json.Marshal(object)
}

// Getter for additional properties for SksOidc_RequiredClaim. Returns the specified
// element and whether it was found
func (a SksOidc_RequiredClaim) Get(fieldName string) (value string, found bool) {
	if a.AdditionalProperties != nil {
		value, found = a.AdditionalProperties[fieldName]
	}
	return
}

// Setter for additional properties for SksOidc_RequiredClaim
func (a *SksOidc_RequiredClaim) Set(fieldName string, value string) {
	if a.AdditionalProperties == nil {
		a.AdditionalProperties = make(map[string]string)
	}
	a.AdditionalProperties[fieldName] = value
}

// Override default JSON handling for SksOidc_RequiredClaim to handle AdditionalProperties
func (a *SksOidc_RequiredClaim) UnmarshalJSON(b []byte) error {
	object := make(map[string]json.RawMessage)
	err := json.Unmarshal(b, &object)
	if err != nil {
		return err
	}

	if len(object) != 0 {
		a.AdditionalProperties = make(map[string]string)
		for fieldName, fieldBuf := range object {
			var fieldVal string
			err := json.Unmarshal(fieldBuf, &fieldVal)
			if err != nil {
				return fmt.Errorf("error unmarshaling field %s: %w", fieldName, err)
			}
			a.AdditionalProperties[fieldName] = fieldVal
		}
	}
	return nil
}

// Override default JSON handling for SksOidc_RequiredClaim to handle AdditionalProperties
func (a SksOidc_RequiredClaim) MarshalJSON() ([]byte, error) {
	var err error
	object := make(map[string]json.RawMessage)

	for fieldName, field := range a.AdditionalProperties {
		object[fieldName], err = json.Marshal(field)
		if err != nil {
			return nil, fmt.Errorf("error marshaling '%s': %w", fieldName, err)
		}
	}
	return json.Marshal(object)
}

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// ListAccessKeys request
	ListAccessKeys(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CreateAccessKey request with any body
	CreateAccessKeyWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CreateAccessKey(ctx context.Context, body CreateAccessKeyJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListAccessKeyKnownOperations request
	ListAccessKeyKnownOperations(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListAccessKeyOperations request
	ListAccessKeyOperations(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// RevokeAccessKey request
	RevokeAccessKey(ctx context.Context, key string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetAccessKey request
	GetAccessKey(ctx context.Context, key string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListAntiAffinityGroups request
	ListAntiAffinityGroups(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CreateAntiAffinityGroup request with any body
	CreateAntiAffinityGroupWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CreateAntiAffinityGroup(ctx context.Context, body CreateAntiAffinityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeleteAntiAffinityGroup request
	DeleteAntiAffinityGroup(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetAntiAffinityGroup request
	GetAntiAffinityGroup(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetDbaasCaCertificate request
	GetDbaasCaCertificate(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetDbaasServiceKafka request
	GetDbaasServiceKafka(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CreateDbaasServiceKafka request with any body
	CreateDbaasServiceKafkaWithBody(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CreateDbaasServiceKafka(ctx context.Context, name DbaasServiceName, body CreateDbaasServiceKafkaJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// UpdateDbaasServiceKafka request with any body
	UpdateDbaasServiceKafkaWithBody(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	UpdateDbaasServiceKafka(ctx context.Context, name DbaasServiceName, body UpdateDbaasServiceKafkaJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetDbaasMigrationStatus request
	GetDbaasMigrationStatus(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetDbaasServiceMysql request
	GetDbaasServiceMysql(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CreateDbaasServiceMysql request with any body
	CreateDbaasServiceMysqlWithBody(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CreateDbaasServiceMysql(ctx context.Context, name DbaasServiceName, body CreateDbaasServiceMysqlJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// UpdateDbaasServiceMysql request with any body
	UpdateDbaasServiceMysqlWithBody(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	UpdateDbaasServiceMysql(ctx context.Context, name DbaasServiceName, body UpdateDbaasServiceMysqlJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetDbaasServiceOpensearch request
	GetDbaasServiceOpensearch(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CreateDbaasServiceOpensearch request with any body
	CreateDbaasServiceOpensearchWithBody(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CreateDbaasServiceOpensearch(ctx context.Context, name DbaasServiceName, body CreateDbaasServiceOpensearchJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// UpdateDbaasServiceOpensearch request with any body
	UpdateDbaasServiceOpensearchWithBody(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	UpdateDbaasServiceOpensearch(ctx context.Context, name DbaasServiceName, body UpdateDbaasServiceOpensearchJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetDbaasServicePg request
	GetDbaasServicePg(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CreateDbaasServicePg request with any body
	CreateDbaasServicePgWithBody(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CreateDbaasServicePg(ctx context.Context, name DbaasServiceName, body CreateDbaasServicePgJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// UpdateDbaasServicePg request with any body
	UpdateDbaasServicePgWithBody(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	UpdateDbaasServicePg(ctx context.Context, name DbaasServiceName, body UpdateDbaasServicePgJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetDbaasServiceRedis request
	GetDbaasServiceRedis(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CreateDbaasServiceRedis request with any body
	CreateDbaasServiceRedisWithBody(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CreateDbaasServiceRedis(ctx context.Context, name DbaasServiceName, body CreateDbaasServiceRedisJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// UpdateDbaasServiceRedis request with any body
	UpdateDbaasServiceRedisWithBody(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	UpdateDbaasServiceRedis(ctx context.Context, name DbaasServiceName, body UpdateDbaasServiceRedisJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListDbaasServices request
	ListDbaasServices(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetDbaasServiceLogs request with any body
	GetDbaasServiceLogsWithBody(ctx context.Context, serviceName string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	GetDbaasServiceLogs(ctx context.Context, serviceName string, body GetDbaasServiceLogsJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetDbaasServiceMetrics request with any body
	GetDbaasServiceMetricsWithBody(ctx context.Context, serviceName string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	GetDbaasServiceMetrics(ctx context.Context, serviceName string, body GetDbaasServiceMetricsJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListDbaasServiceTypes request
	ListDbaasServiceTypes(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetDbaasServiceType request
	GetDbaasServiceType(ctx context.Context, serviceTypeName string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeleteDbaasService request
	DeleteDbaasService(ctx context.Context, name string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetDbaasSettingsKafka request
	GetDbaasSettingsKafka(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetDbaasSettingsMysql request
	GetDbaasSettingsMysql(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetDbaasSettingsOpensearch request
	GetDbaasSettingsOpensearch(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetDbaasSettingsPg request
	GetDbaasSettingsPg(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetDbaasSettingsRedis request
	GetDbaasSettingsRedis(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListDeployTargets request
	ListDeployTargets(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetDeployTarget request
	GetDeployTarget(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListDnsDomains request
	ListDnsDomains(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetDnsDomain request
	GetDnsDomain(ctx context.Context, id int64, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListDnsDomainRecords request
	ListDnsDomainRecords(ctx context.Context, id int64, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetDnsDomainRecord request
	GetDnsDomainRecord(ctx context.Context, id int64, recordId int64, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListElasticIps request
	ListElasticIps(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CreateElasticIp request with any body
	CreateElasticIpWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CreateElasticIp(ctx context.Context, body CreateElasticIpJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeleteElasticIp request
	DeleteElasticIp(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetElasticIp request
	GetElasticIp(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// UpdateElasticIp request with any body
	UpdateElasticIpWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	UpdateElasticIp(ctx context.Context, id string, body UpdateElasticIpJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ResetElasticIpField request
	ResetElasticIpField(ctx context.Context, id string, field ResetElasticIpFieldParamsField, reqEditors ...RequestEditorFn) (*http.Response, error)

	// AttachInstanceToElasticIp request with any body
	AttachInstanceToElasticIpWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	AttachInstanceToElasticIp(ctx context.Context, id string, body AttachInstanceToElasticIpJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DetachInstanceFromElasticIp request with any body
	DetachInstanceFromElasticIpWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	DetachInstanceFromElasticIp(ctx context.Context, id string, body DetachInstanceFromElasticIpJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListEvents request
	ListEvents(ctx context.Context, params *ListEventsParams, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListInstances request
	ListInstances(ctx context.Context, params *ListInstancesParams, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CreateInstance request with any body
	CreateInstanceWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CreateInstance(ctx context.Context, body CreateInstanceJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListInstancePools request
	ListInstancePools(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CreateInstancePool request with any body
	CreateInstancePoolWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CreateInstancePool(ctx context.Context, body CreateInstancePoolJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeleteInstancePool request
	DeleteInstancePool(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetInstancePool request
	GetInstancePool(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// UpdateInstancePool request with any body
	UpdateInstancePoolWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	UpdateInstancePool(ctx context.Context, id string, body UpdateInstancePoolJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ResetInstancePoolField request
	ResetInstancePoolField(ctx context.Context, id string, field ResetInstancePoolFieldParamsField, reqEditors ...RequestEditorFn) (*http.Response, error)

	// EvictInstancePoolMembers request with any body
	EvictInstancePoolMembersWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	EvictInstancePoolMembers(ctx context.Context, id string, body EvictInstancePoolMembersJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ScaleInstancePool request with any body
	ScaleInstancePoolWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	ScaleInstancePool(ctx context.Context, id string, body ScaleInstancePoolJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListInstanceTypes request
	ListInstanceTypes(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetInstanceType request
	GetInstanceType(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeleteInstance request
	DeleteInstance(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetInstance request
	GetInstance(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// UpdateInstance request with any body
	UpdateInstanceWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	UpdateInstance(ctx context.Context, id string, body UpdateInstanceJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ResetInstanceField request
	ResetInstanceField(ctx context.Context, id string, field ResetInstanceFieldParamsField, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CreateSnapshot request
	CreateSnapshot(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// RebootInstance request
	RebootInstance(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ResetInstance request with any body
	ResetInstanceWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	ResetInstance(ctx context.Context, id string, body ResetInstanceJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ResizeInstanceDisk request with any body
	ResizeInstanceDiskWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	ResizeInstanceDisk(ctx context.Context, id string, body ResizeInstanceDiskJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ScaleInstance request with any body
	ScaleInstanceWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	ScaleInstance(ctx context.Context, id string, body ScaleInstanceJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// StartInstance request with any body
	StartInstanceWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	StartInstance(ctx context.Context, id string, body StartInstanceJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// StopInstance request
	StopInstance(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// RevertInstanceToSnapshot request with any body
	RevertInstanceToSnapshotWithBody(ctx context.Context, instanceId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	RevertInstanceToSnapshot(ctx context.Context, instanceId string, body RevertInstanceToSnapshotJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListLoadBalancers request
	ListLoadBalancers(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CreateLoadBalancer request with any body
	CreateLoadBalancerWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CreateLoadBalancer(ctx context.Context, body CreateLoadBalancerJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeleteLoadBalancer request
	DeleteLoadBalancer(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetLoadBalancer request
	GetLoadBalancer(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// UpdateLoadBalancer request with any body
	UpdateLoadBalancerWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	UpdateLoadBalancer(ctx context.Context, id string, body UpdateLoadBalancerJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// AddServiceToLoadBalancer request with any body
	AddServiceToLoadBalancerWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	AddServiceToLoadBalancer(ctx context.Context, id string, body AddServiceToLoadBalancerJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeleteLoadBalancerService request
	DeleteLoadBalancerService(ctx context.Context, id string, serviceId string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetLoadBalancerService request
	GetLoadBalancerService(ctx context.Context, id string, serviceId string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// UpdateLoadBalancerService request with any body
	UpdateLoadBalancerServiceWithBody(ctx context.Context, id string, serviceId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	UpdateLoadBalancerService(ctx context.Context, id string, serviceId string, body UpdateLoadBalancerServiceJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ResetLoadBalancerServiceField request
	ResetLoadBalancerServiceField(ctx context.Context, id string, serviceId string, field ResetLoadBalancerServiceFieldParamsField, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ResetLoadBalancerField request
	ResetLoadBalancerField(ctx context.Context, id string, field ResetLoadBalancerFieldParamsField, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetOperation request
	GetOperation(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListPrivateNetworks request
	ListPrivateNetworks(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CreatePrivateNetwork request with any body
	CreatePrivateNetworkWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CreatePrivateNetwork(ctx context.Context, body CreatePrivateNetworkJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeletePrivateNetwork request
	DeletePrivateNetwork(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetPrivateNetwork request
	GetPrivateNetwork(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// UpdatePrivateNetwork request with any body
	UpdatePrivateNetworkWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	UpdatePrivateNetwork(ctx context.Context, id string, body UpdatePrivateNetworkJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ResetPrivateNetworkField request
	ResetPrivateNetworkField(ctx context.Context, id string, field ResetPrivateNetworkFieldParamsField, reqEditors ...RequestEditorFn) (*http.Response, error)

	// AttachInstanceToPrivateNetwork request with any body
	AttachInstanceToPrivateNetworkWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	AttachInstanceToPrivateNetwork(ctx context.Context, id string, body AttachInstanceToPrivateNetworkJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DetachInstanceFromPrivateNetwork request with any body
	DetachInstanceFromPrivateNetworkWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	DetachInstanceFromPrivateNetwork(ctx context.Context, id string, body DetachInstanceFromPrivateNetworkJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// UpdatePrivateNetworkInstanceIp request with any body
	UpdatePrivateNetworkInstanceIpWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	UpdatePrivateNetworkInstanceIp(ctx context.Context, id string, body UpdatePrivateNetworkInstanceIpJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListQuotas request
	ListQuotas(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetQuota request
	GetQuota(ctx context.Context, entity string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListSecurityGroups request
	ListSecurityGroups(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CreateSecurityGroup request with any body
	CreateSecurityGroupWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CreateSecurityGroup(ctx context.Context, body CreateSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeleteSecurityGroup request
	DeleteSecurityGroup(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetSecurityGroup request
	GetSecurityGroup(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// AddRuleToSecurityGroup request with any body
	AddRuleToSecurityGroupWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	AddRuleToSecurityGroup(ctx context.Context, id string, body AddRuleToSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeleteRuleFromSecurityGroup request
	DeleteRuleFromSecurityGroup(ctx context.Context, id string, ruleId string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// AddExternalSourceToSecurityGroup request with any body
	AddExternalSourceToSecurityGroupWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	AddExternalSourceToSecurityGroup(ctx context.Context, id string, body AddExternalSourceToSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// AttachInstanceToSecurityGroup request with any body
	AttachInstanceToSecurityGroupWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	AttachInstanceToSecurityGroup(ctx context.Context, id string, body AttachInstanceToSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DetachInstanceFromSecurityGroup request with any body
	DetachInstanceFromSecurityGroupWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	DetachInstanceFromSecurityGroup(ctx context.Context, id string, body DetachInstanceFromSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// RemoveExternalSourceFromSecurityGroup request with any body
	RemoveExternalSourceFromSecurityGroupWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	RemoveExternalSourceFromSecurityGroup(ctx context.Context, id string, body RemoveExternalSourceFromSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListSksClusters request
	ListSksClusters(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CreateSksCluster request with any body
	CreateSksClusterWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CreateSksCluster(ctx context.Context, body CreateSksClusterJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListSksClusterDeprecatedResources request
	ListSksClusterDeprecatedResources(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GenerateSksClusterKubeconfig request with any body
	GenerateSksClusterKubeconfigWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	GenerateSksClusterKubeconfig(ctx context.Context, id string, body GenerateSksClusterKubeconfigJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListSksClusterVersions request
	ListSksClusterVersions(ctx context.Context, params *ListSksClusterVersionsParams, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeleteSksCluster request
	DeleteSksCluster(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetSksCluster request
	GetSksCluster(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// UpdateSksCluster request with any body
	UpdateSksClusterWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	UpdateSksCluster(ctx context.Context, id string, body UpdateSksClusterJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetSksClusterAuthorityCert request
	GetSksClusterAuthorityCert(ctx context.Context, id string, authority GetSksClusterAuthorityCertParamsAuthority, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CreateSksNodepool request with any body
	CreateSksNodepoolWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CreateSksNodepool(ctx context.Context, id string, body CreateSksNodepoolJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeleteSksNodepool request
	DeleteSksNodepool(ctx context.Context, id string, sksNodepoolId string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetSksNodepool request
	GetSksNodepool(ctx context.Context, id string, sksNodepoolId string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// UpdateSksNodepool request with any body
	UpdateSksNodepoolWithBody(ctx context.Context, id string, sksNodepoolId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	UpdateSksNodepool(ctx context.Context, id string, sksNodepoolId string, body UpdateSksNodepoolJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ResetSksNodepoolField request
	ResetSksNodepoolField(ctx context.Context, id string, sksNodepoolId string, field ResetSksNodepoolFieldParamsField, reqEditors ...RequestEditorFn) (*http.Response, error)

	// EvictSksNodepoolMembers request with any body
	EvictSksNodepoolMembersWithBody(ctx context.Context, id string, sksNodepoolId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	EvictSksNodepoolMembers(ctx context.Context, id string, sksNodepoolId string, body EvictSksNodepoolMembersJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ScaleSksNodepool request with any body
	ScaleSksNodepoolWithBody(ctx context.Context, id string, sksNodepoolId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	ScaleSksNodepool(ctx context.Context, id string, sksNodepoolId string, body ScaleSksNodepoolJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// RotateSksCcmCredentials request
	RotateSksCcmCredentials(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// RotateSksOperatorsCa request
	RotateSksOperatorsCa(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// UpgradeSksCluster request with any body
	UpgradeSksClusterWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	UpgradeSksCluster(ctx context.Context, id string, body UpgradeSksClusterJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// UpgradeSksClusterServiceLevel request
	UpgradeSksClusterServiceLevel(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ResetSksClusterField request
	ResetSksClusterField(ctx context.Context, id string, field ResetSksClusterFieldParamsField, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListSnapshots request
	ListSnapshots(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeleteSnapshot request
	DeleteSnapshot(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetSnapshot request
	GetSnapshot(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ExportSnapshot request
	ExportSnapshot(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// PromoteSnapshotToTemplate request with any body
	PromoteSnapshotToTemplateWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	PromoteSnapshotToTemplate(ctx context.Context, id string, body PromoteSnapshotToTemplateJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetSosPresignedUrl request
	GetSosPresignedUrl(ctx context.Context, bucket string, params *GetSosPresignedUrlParams, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListSshKeys request
	ListSshKeys(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// RegisterSshKey request with any body
	RegisterSshKeyWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	RegisterSshKey(ctx context.Context, body RegisterSshKeyJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeleteSshKey request
	DeleteSshKey(ctx context.Context, name string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetSshKey request
	GetSshKey(ctx context.Context, name string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListTemplates request
	ListTemplates(ctx context.Context, params *ListTemplatesParams, reqEditors ...RequestEditorFn) (*http.Response, error)

	// RegisterTemplate request with any body
	RegisterTemplateWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	RegisterTemplate(ctx context.Context, body RegisterTemplateJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DeleteTemplate request
	DeleteTemplate(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetTemplate request
	GetTemplate(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// CopyTemplate request with any body
	CopyTemplateWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	CopyTemplate(ctx context.Context, id string, body CopyTemplateJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// UpdateTemplate request with any body
	UpdateTemplateWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	UpdateTemplate(ctx context.Context, id string, body UpdateTemplateJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ListZones request
	ListZones(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) ListAccessKeys(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListAccessKeysRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateAccessKeyWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateAccessKeyRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateAccessKey(ctx context.Context, body CreateAccessKeyJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateAccessKeyRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListAccessKeyKnownOperations(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListAccessKeyKnownOperationsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListAccessKeyOperations(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListAccessKeyOperationsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) RevokeAccessKey(ctx context.Context, key string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewRevokeAccessKeyRequest(c.Server, key)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetAccessKey(ctx context.Context, key string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetAccessKeyRequest(c.Server, key)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListAntiAffinityGroups(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListAntiAffinityGroupsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateAntiAffinityGroupWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateAntiAffinityGroupRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateAntiAffinityGroup(ctx context.Context, body CreateAntiAffinityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateAntiAffinityGroupRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteAntiAffinityGroup(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteAntiAffinityGroupRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetAntiAffinityGroup(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetAntiAffinityGroupRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetDbaasCaCertificate(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetDbaasCaCertificateRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetDbaasServiceKafka(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetDbaasServiceKafkaRequest(c.Server, name)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateDbaasServiceKafkaWithBody(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateDbaasServiceKafkaRequestWithBody(c.Server, name, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateDbaasServiceKafka(ctx context.Context, name DbaasServiceName, body CreateDbaasServiceKafkaJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateDbaasServiceKafkaRequest(c.Server, name, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateDbaasServiceKafkaWithBody(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateDbaasServiceKafkaRequestWithBody(c.Server, name, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateDbaasServiceKafka(ctx context.Context, name DbaasServiceName, body UpdateDbaasServiceKafkaJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateDbaasServiceKafkaRequest(c.Server, name, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetDbaasMigrationStatus(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetDbaasMigrationStatusRequest(c.Server, name)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetDbaasServiceMysql(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetDbaasServiceMysqlRequest(c.Server, name)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateDbaasServiceMysqlWithBody(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateDbaasServiceMysqlRequestWithBody(c.Server, name, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateDbaasServiceMysql(ctx context.Context, name DbaasServiceName, body CreateDbaasServiceMysqlJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateDbaasServiceMysqlRequest(c.Server, name, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateDbaasServiceMysqlWithBody(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateDbaasServiceMysqlRequestWithBody(c.Server, name, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateDbaasServiceMysql(ctx context.Context, name DbaasServiceName, body UpdateDbaasServiceMysqlJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateDbaasServiceMysqlRequest(c.Server, name, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetDbaasServiceOpensearch(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetDbaasServiceOpensearchRequest(c.Server, name)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateDbaasServiceOpensearchWithBody(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateDbaasServiceOpensearchRequestWithBody(c.Server, name, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateDbaasServiceOpensearch(ctx context.Context, name DbaasServiceName, body CreateDbaasServiceOpensearchJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateDbaasServiceOpensearchRequest(c.Server, name, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateDbaasServiceOpensearchWithBody(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateDbaasServiceOpensearchRequestWithBody(c.Server, name, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateDbaasServiceOpensearch(ctx context.Context, name DbaasServiceName, body UpdateDbaasServiceOpensearchJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateDbaasServiceOpensearchRequest(c.Server, name, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetDbaasServicePg(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetDbaasServicePgRequest(c.Server, name)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateDbaasServicePgWithBody(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateDbaasServicePgRequestWithBody(c.Server, name, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateDbaasServicePg(ctx context.Context, name DbaasServiceName, body CreateDbaasServicePgJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateDbaasServicePgRequest(c.Server, name, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateDbaasServicePgWithBody(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateDbaasServicePgRequestWithBody(c.Server, name, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateDbaasServicePg(ctx context.Context, name DbaasServiceName, body UpdateDbaasServicePgJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateDbaasServicePgRequest(c.Server, name, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetDbaasServiceRedis(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetDbaasServiceRedisRequest(c.Server, name)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateDbaasServiceRedisWithBody(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateDbaasServiceRedisRequestWithBody(c.Server, name, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateDbaasServiceRedis(ctx context.Context, name DbaasServiceName, body CreateDbaasServiceRedisJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateDbaasServiceRedisRequest(c.Server, name, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateDbaasServiceRedisWithBody(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateDbaasServiceRedisRequestWithBody(c.Server, name, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateDbaasServiceRedis(ctx context.Context, name DbaasServiceName, body UpdateDbaasServiceRedisJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateDbaasServiceRedisRequest(c.Server, name, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListDbaasServices(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListDbaasServicesRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetDbaasServiceLogsWithBody(ctx context.Context, serviceName string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetDbaasServiceLogsRequestWithBody(c.Server, serviceName, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetDbaasServiceLogs(ctx context.Context, serviceName string, body GetDbaasServiceLogsJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetDbaasServiceLogsRequest(c.Server, serviceName, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetDbaasServiceMetricsWithBody(ctx context.Context, serviceName string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetDbaasServiceMetricsRequestWithBody(c.Server, serviceName, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetDbaasServiceMetrics(ctx context.Context, serviceName string, body GetDbaasServiceMetricsJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetDbaasServiceMetricsRequest(c.Server, serviceName, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListDbaasServiceTypes(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListDbaasServiceTypesRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetDbaasServiceType(ctx context.Context, serviceTypeName string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetDbaasServiceTypeRequest(c.Server, serviceTypeName)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteDbaasService(ctx context.Context, name string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteDbaasServiceRequest(c.Server, name)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetDbaasSettingsKafka(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetDbaasSettingsKafkaRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetDbaasSettingsMysql(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetDbaasSettingsMysqlRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetDbaasSettingsOpensearch(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetDbaasSettingsOpensearchRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetDbaasSettingsPg(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetDbaasSettingsPgRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetDbaasSettingsRedis(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetDbaasSettingsRedisRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListDeployTargets(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListDeployTargetsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetDeployTarget(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetDeployTargetRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListDnsDomains(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListDnsDomainsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetDnsDomain(ctx context.Context, id int64, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetDnsDomainRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListDnsDomainRecords(ctx context.Context, id int64, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListDnsDomainRecordsRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetDnsDomainRecord(ctx context.Context, id int64, recordId int64, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetDnsDomainRecordRequest(c.Server, id, recordId)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListElasticIps(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListElasticIpsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateElasticIpWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateElasticIpRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateElasticIp(ctx context.Context, body CreateElasticIpJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateElasticIpRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteElasticIp(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteElasticIpRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetElasticIp(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetElasticIpRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateElasticIpWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateElasticIpRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateElasticIp(ctx context.Context, id string, body UpdateElasticIpJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateElasticIpRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ResetElasticIpField(ctx context.Context, id string, field ResetElasticIpFieldParamsField, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewResetElasticIpFieldRequest(c.Server, id, field)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) AttachInstanceToElasticIpWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewAttachInstanceToElasticIpRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) AttachInstanceToElasticIp(ctx context.Context, id string, body AttachInstanceToElasticIpJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewAttachInstanceToElasticIpRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DetachInstanceFromElasticIpWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDetachInstanceFromElasticIpRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DetachInstanceFromElasticIp(ctx context.Context, id string, body DetachInstanceFromElasticIpJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDetachInstanceFromElasticIpRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListEvents(ctx context.Context, params *ListEventsParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListEventsRequest(c.Server, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListInstances(ctx context.Context, params *ListInstancesParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListInstancesRequest(c.Server, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateInstanceWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateInstanceRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateInstance(ctx context.Context, body CreateInstanceJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateInstanceRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListInstancePools(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListInstancePoolsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateInstancePoolWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateInstancePoolRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateInstancePool(ctx context.Context, body CreateInstancePoolJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateInstancePoolRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteInstancePool(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteInstancePoolRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetInstancePool(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetInstancePoolRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateInstancePoolWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateInstancePoolRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateInstancePool(ctx context.Context, id string, body UpdateInstancePoolJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateInstancePoolRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ResetInstancePoolField(ctx context.Context, id string, field ResetInstancePoolFieldParamsField, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewResetInstancePoolFieldRequest(c.Server, id, field)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) EvictInstancePoolMembersWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewEvictInstancePoolMembersRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) EvictInstancePoolMembers(ctx context.Context, id string, body EvictInstancePoolMembersJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewEvictInstancePoolMembersRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ScaleInstancePoolWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewScaleInstancePoolRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ScaleInstancePool(ctx context.Context, id string, body ScaleInstancePoolJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewScaleInstancePoolRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListInstanceTypes(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListInstanceTypesRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetInstanceType(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetInstanceTypeRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteInstance(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteInstanceRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetInstance(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetInstanceRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateInstanceWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateInstanceRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateInstance(ctx context.Context, id string, body UpdateInstanceJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateInstanceRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ResetInstanceField(ctx context.Context, id string, field ResetInstanceFieldParamsField, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewResetInstanceFieldRequest(c.Server, id, field)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateSnapshot(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateSnapshotRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) RebootInstance(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewRebootInstanceRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ResetInstanceWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewResetInstanceRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ResetInstance(ctx context.Context, id string, body ResetInstanceJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewResetInstanceRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ResizeInstanceDiskWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewResizeInstanceDiskRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ResizeInstanceDisk(ctx context.Context, id string, body ResizeInstanceDiskJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewResizeInstanceDiskRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ScaleInstanceWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewScaleInstanceRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ScaleInstance(ctx context.Context, id string, body ScaleInstanceJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewScaleInstanceRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) StartInstanceWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewStartInstanceRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) StartInstance(ctx context.Context, id string, body StartInstanceJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewStartInstanceRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) StopInstance(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewStopInstanceRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) RevertInstanceToSnapshotWithBody(ctx context.Context, instanceId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewRevertInstanceToSnapshotRequestWithBody(c.Server, instanceId, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) RevertInstanceToSnapshot(ctx context.Context, instanceId string, body RevertInstanceToSnapshotJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewRevertInstanceToSnapshotRequest(c.Server, instanceId, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListLoadBalancers(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListLoadBalancersRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateLoadBalancerWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateLoadBalancerRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateLoadBalancer(ctx context.Context, body CreateLoadBalancerJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateLoadBalancerRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteLoadBalancer(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteLoadBalancerRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetLoadBalancer(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetLoadBalancerRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateLoadBalancerWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateLoadBalancerRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateLoadBalancer(ctx context.Context, id string, body UpdateLoadBalancerJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateLoadBalancerRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) AddServiceToLoadBalancerWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewAddServiceToLoadBalancerRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) AddServiceToLoadBalancer(ctx context.Context, id string, body AddServiceToLoadBalancerJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewAddServiceToLoadBalancerRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteLoadBalancerService(ctx context.Context, id string, serviceId string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteLoadBalancerServiceRequest(c.Server, id, serviceId)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetLoadBalancerService(ctx context.Context, id string, serviceId string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetLoadBalancerServiceRequest(c.Server, id, serviceId)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateLoadBalancerServiceWithBody(ctx context.Context, id string, serviceId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateLoadBalancerServiceRequestWithBody(c.Server, id, serviceId, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateLoadBalancerService(ctx context.Context, id string, serviceId string, body UpdateLoadBalancerServiceJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateLoadBalancerServiceRequest(c.Server, id, serviceId, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ResetLoadBalancerServiceField(ctx context.Context, id string, serviceId string, field ResetLoadBalancerServiceFieldParamsField, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewResetLoadBalancerServiceFieldRequest(c.Server, id, serviceId, field)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ResetLoadBalancerField(ctx context.Context, id string, field ResetLoadBalancerFieldParamsField, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewResetLoadBalancerFieldRequest(c.Server, id, field)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetOperation(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetOperationRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListPrivateNetworks(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListPrivateNetworksRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreatePrivateNetworkWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreatePrivateNetworkRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreatePrivateNetwork(ctx context.Context, body CreatePrivateNetworkJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreatePrivateNetworkRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeletePrivateNetwork(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeletePrivateNetworkRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetPrivateNetwork(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetPrivateNetworkRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdatePrivateNetworkWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdatePrivateNetworkRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdatePrivateNetwork(ctx context.Context, id string, body UpdatePrivateNetworkJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdatePrivateNetworkRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ResetPrivateNetworkField(ctx context.Context, id string, field ResetPrivateNetworkFieldParamsField, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewResetPrivateNetworkFieldRequest(c.Server, id, field)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) AttachInstanceToPrivateNetworkWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewAttachInstanceToPrivateNetworkRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) AttachInstanceToPrivateNetwork(ctx context.Context, id string, body AttachInstanceToPrivateNetworkJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewAttachInstanceToPrivateNetworkRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DetachInstanceFromPrivateNetworkWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDetachInstanceFromPrivateNetworkRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DetachInstanceFromPrivateNetwork(ctx context.Context, id string, body DetachInstanceFromPrivateNetworkJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDetachInstanceFromPrivateNetworkRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdatePrivateNetworkInstanceIpWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdatePrivateNetworkInstanceIpRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdatePrivateNetworkInstanceIp(ctx context.Context, id string, body UpdatePrivateNetworkInstanceIpJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdatePrivateNetworkInstanceIpRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListQuotas(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListQuotasRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetQuota(ctx context.Context, entity string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetQuotaRequest(c.Server, entity)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListSecurityGroups(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListSecurityGroupsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateSecurityGroupWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateSecurityGroupRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateSecurityGroup(ctx context.Context, body CreateSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateSecurityGroupRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteSecurityGroup(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteSecurityGroupRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetSecurityGroup(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetSecurityGroupRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) AddRuleToSecurityGroupWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewAddRuleToSecurityGroupRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) AddRuleToSecurityGroup(ctx context.Context, id string, body AddRuleToSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewAddRuleToSecurityGroupRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteRuleFromSecurityGroup(ctx context.Context, id string, ruleId string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteRuleFromSecurityGroupRequest(c.Server, id, ruleId)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) AddExternalSourceToSecurityGroupWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewAddExternalSourceToSecurityGroupRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) AddExternalSourceToSecurityGroup(ctx context.Context, id string, body AddExternalSourceToSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewAddExternalSourceToSecurityGroupRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) AttachInstanceToSecurityGroupWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewAttachInstanceToSecurityGroupRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) AttachInstanceToSecurityGroup(ctx context.Context, id string, body AttachInstanceToSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewAttachInstanceToSecurityGroupRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DetachInstanceFromSecurityGroupWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDetachInstanceFromSecurityGroupRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DetachInstanceFromSecurityGroup(ctx context.Context, id string, body DetachInstanceFromSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDetachInstanceFromSecurityGroupRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) RemoveExternalSourceFromSecurityGroupWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewRemoveExternalSourceFromSecurityGroupRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) RemoveExternalSourceFromSecurityGroup(ctx context.Context, id string, body RemoveExternalSourceFromSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewRemoveExternalSourceFromSecurityGroupRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListSksClusters(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListSksClustersRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateSksClusterWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateSksClusterRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateSksCluster(ctx context.Context, body CreateSksClusterJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateSksClusterRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListSksClusterDeprecatedResources(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListSksClusterDeprecatedResourcesRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GenerateSksClusterKubeconfigWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGenerateSksClusterKubeconfigRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GenerateSksClusterKubeconfig(ctx context.Context, id string, body GenerateSksClusterKubeconfigJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGenerateSksClusterKubeconfigRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListSksClusterVersions(ctx context.Context, params *ListSksClusterVersionsParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListSksClusterVersionsRequest(c.Server, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteSksCluster(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteSksClusterRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetSksCluster(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetSksClusterRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateSksClusterWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateSksClusterRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateSksCluster(ctx context.Context, id string, body UpdateSksClusterJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateSksClusterRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetSksClusterAuthorityCert(ctx context.Context, id string, authority GetSksClusterAuthorityCertParamsAuthority, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetSksClusterAuthorityCertRequest(c.Server, id, authority)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateSksNodepoolWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateSksNodepoolRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CreateSksNodepool(ctx context.Context, id string, body CreateSksNodepoolJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCreateSksNodepoolRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteSksNodepool(ctx context.Context, id string, sksNodepoolId string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteSksNodepoolRequest(c.Server, id, sksNodepoolId)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetSksNodepool(ctx context.Context, id string, sksNodepoolId string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetSksNodepoolRequest(c.Server, id, sksNodepoolId)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateSksNodepoolWithBody(ctx context.Context, id string, sksNodepoolId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateSksNodepoolRequestWithBody(c.Server, id, sksNodepoolId, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateSksNodepool(ctx context.Context, id string, sksNodepoolId string, body UpdateSksNodepoolJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateSksNodepoolRequest(c.Server, id, sksNodepoolId, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ResetSksNodepoolField(ctx context.Context, id string, sksNodepoolId string, field ResetSksNodepoolFieldParamsField, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewResetSksNodepoolFieldRequest(c.Server, id, sksNodepoolId, field)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) EvictSksNodepoolMembersWithBody(ctx context.Context, id string, sksNodepoolId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewEvictSksNodepoolMembersRequestWithBody(c.Server, id, sksNodepoolId, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) EvictSksNodepoolMembers(ctx context.Context, id string, sksNodepoolId string, body EvictSksNodepoolMembersJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewEvictSksNodepoolMembersRequest(c.Server, id, sksNodepoolId, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ScaleSksNodepoolWithBody(ctx context.Context, id string, sksNodepoolId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewScaleSksNodepoolRequestWithBody(c.Server, id, sksNodepoolId, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ScaleSksNodepool(ctx context.Context, id string, sksNodepoolId string, body ScaleSksNodepoolJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewScaleSksNodepoolRequest(c.Server, id, sksNodepoolId, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) RotateSksCcmCredentials(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewRotateSksCcmCredentialsRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) RotateSksOperatorsCa(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewRotateSksOperatorsCaRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpgradeSksClusterWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpgradeSksClusterRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpgradeSksCluster(ctx context.Context, id string, body UpgradeSksClusterJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpgradeSksClusterRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpgradeSksClusterServiceLevel(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpgradeSksClusterServiceLevelRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ResetSksClusterField(ctx context.Context, id string, field ResetSksClusterFieldParamsField, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewResetSksClusterFieldRequest(c.Server, id, field)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListSnapshots(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListSnapshotsRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteSnapshot(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteSnapshotRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetSnapshot(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetSnapshotRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ExportSnapshot(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewExportSnapshotRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PromoteSnapshotToTemplateWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPromoteSnapshotToTemplateRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PromoteSnapshotToTemplate(ctx context.Context, id string, body PromoteSnapshotToTemplateJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPromoteSnapshotToTemplateRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetSosPresignedUrl(ctx context.Context, bucket string, params *GetSosPresignedUrlParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetSosPresignedUrlRequest(c.Server, bucket, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListSshKeys(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListSshKeysRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) RegisterSshKeyWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewRegisterSshKeyRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) RegisterSshKey(ctx context.Context, body RegisterSshKeyJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewRegisterSshKeyRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteSshKey(ctx context.Context, name string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteSshKeyRequest(c.Server, name)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetSshKey(ctx context.Context, name string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetSshKeyRequest(c.Server, name)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListTemplates(ctx context.Context, params *ListTemplatesParams, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListTemplatesRequest(c.Server, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) RegisterTemplateWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewRegisterTemplateRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) RegisterTemplate(ctx context.Context, body RegisterTemplateJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewRegisterTemplateRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DeleteTemplate(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDeleteTemplateRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetTemplate(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetTemplateRequest(c.Server, id)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CopyTemplateWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCopyTemplateRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) CopyTemplate(ctx context.Context, id string, body CopyTemplateJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewCopyTemplateRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateTemplateWithBody(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateTemplateRequestWithBody(c.Server, id, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) UpdateTemplate(ctx context.Context, id string, body UpdateTemplateJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewUpdateTemplateRequest(c.Server, id, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ListZones(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewListZonesRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewListAccessKeysRequest generates requests for ListAccessKeys
func NewListAccessKeysRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/access-key")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateAccessKeyRequest calls the generic CreateAccessKey builder with application/json body
func NewCreateAccessKeyRequest(server string, body CreateAccessKeyJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateAccessKeyRequestWithBody(server, "application/json", bodyReader)
}

// NewCreateAccessKeyRequestWithBody generates requests for CreateAccessKey with any type of body
func NewCreateAccessKeyRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/access-key")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewListAccessKeyKnownOperationsRequest generates requests for ListAccessKeyKnownOperations
func NewListAccessKeyKnownOperationsRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/access-key-known-operations")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewListAccessKeyOperationsRequest generates requests for ListAccessKeyOperations
func NewListAccessKeyOperationsRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/access-key-operations")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewRevokeAccessKeyRequest generates requests for RevokeAccessKey
func NewRevokeAccessKeyRequest(server string, key string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "key", runtime.ParamLocationPath, key)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/access-key/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetAccessKeyRequest generates requests for GetAccessKey
func NewGetAccessKeyRequest(server string, key string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "key", runtime.ParamLocationPath, key)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/access-key/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewListAntiAffinityGroupsRequest generates requests for ListAntiAffinityGroups
func NewListAntiAffinityGroupsRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/anti-affinity-group")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateAntiAffinityGroupRequest calls the generic CreateAntiAffinityGroup builder with application/json body
func NewCreateAntiAffinityGroupRequest(server string, body CreateAntiAffinityGroupJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateAntiAffinityGroupRequestWithBody(server, "application/json", bodyReader)
}

// NewCreateAntiAffinityGroupRequestWithBody generates requests for CreateAntiAffinityGroup with any type of body
func NewCreateAntiAffinityGroupRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/anti-affinity-group")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewDeleteAntiAffinityGroupRequest generates requests for DeleteAntiAffinityGroup
func NewDeleteAntiAffinityGroupRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/anti-affinity-group/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetAntiAffinityGroupRequest generates requests for GetAntiAffinityGroup
func NewGetAntiAffinityGroupRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/anti-affinity-group/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetDbaasCaCertificateRequest generates requests for GetDbaasCaCertificate
func NewGetDbaasCaCertificateRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-ca-certificate")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetDbaasServiceKafkaRequest generates requests for GetDbaasServiceKafka
func NewGetDbaasServiceKafkaRequest(server string, name DbaasServiceName) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-kafka/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateDbaasServiceKafkaRequest calls the generic CreateDbaasServiceKafka builder with application/json body
func NewCreateDbaasServiceKafkaRequest(server string, name DbaasServiceName, body CreateDbaasServiceKafkaJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateDbaasServiceKafkaRequestWithBody(server, name, "application/json", bodyReader)
}

// NewCreateDbaasServiceKafkaRequestWithBody generates requests for CreateDbaasServiceKafka with any type of body
func NewCreateDbaasServiceKafkaRequestWithBody(server string, name DbaasServiceName, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-kafka/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewUpdateDbaasServiceKafkaRequest calls the generic UpdateDbaasServiceKafka builder with application/json body
func NewUpdateDbaasServiceKafkaRequest(server string, name DbaasServiceName, body UpdateDbaasServiceKafkaJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUpdateDbaasServiceKafkaRequestWithBody(server, name, "application/json", bodyReader)
}

// NewUpdateDbaasServiceKafkaRequestWithBody generates requests for UpdateDbaasServiceKafka with any type of body
func NewUpdateDbaasServiceKafkaRequestWithBody(server string, name DbaasServiceName, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-kafka/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewGetDbaasMigrationStatusRequest generates requests for GetDbaasMigrationStatus
func NewGetDbaasMigrationStatusRequest(server string, name DbaasServiceName) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-migration-status/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetDbaasServiceMysqlRequest generates requests for GetDbaasServiceMysql
func NewGetDbaasServiceMysqlRequest(server string, name DbaasServiceName) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-mysql/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateDbaasServiceMysqlRequest calls the generic CreateDbaasServiceMysql builder with application/json body
func NewCreateDbaasServiceMysqlRequest(server string, name DbaasServiceName, body CreateDbaasServiceMysqlJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateDbaasServiceMysqlRequestWithBody(server, name, "application/json", bodyReader)
}

// NewCreateDbaasServiceMysqlRequestWithBody generates requests for CreateDbaasServiceMysql with any type of body
func NewCreateDbaasServiceMysqlRequestWithBody(server string, name DbaasServiceName, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-mysql/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewUpdateDbaasServiceMysqlRequest calls the generic UpdateDbaasServiceMysql builder with application/json body
func NewUpdateDbaasServiceMysqlRequest(server string, name DbaasServiceName, body UpdateDbaasServiceMysqlJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUpdateDbaasServiceMysqlRequestWithBody(server, name, "application/json", bodyReader)
}

// NewUpdateDbaasServiceMysqlRequestWithBody generates requests for UpdateDbaasServiceMysql with any type of body
func NewUpdateDbaasServiceMysqlRequestWithBody(server string, name DbaasServiceName, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-mysql/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewGetDbaasServiceOpensearchRequest generates requests for GetDbaasServiceOpensearch
func NewGetDbaasServiceOpensearchRequest(server string, name DbaasServiceName) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-opensearch/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateDbaasServiceOpensearchRequest calls the generic CreateDbaasServiceOpensearch builder with application/json body
func NewCreateDbaasServiceOpensearchRequest(server string, name DbaasServiceName, body CreateDbaasServiceOpensearchJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateDbaasServiceOpensearchRequestWithBody(server, name, "application/json", bodyReader)
}

// NewCreateDbaasServiceOpensearchRequestWithBody generates requests for CreateDbaasServiceOpensearch with any type of body
func NewCreateDbaasServiceOpensearchRequestWithBody(server string, name DbaasServiceName, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-opensearch/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewUpdateDbaasServiceOpensearchRequest calls the generic UpdateDbaasServiceOpensearch builder with application/json body
func NewUpdateDbaasServiceOpensearchRequest(server string, name DbaasServiceName, body UpdateDbaasServiceOpensearchJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUpdateDbaasServiceOpensearchRequestWithBody(server, name, "application/json", bodyReader)
}

// NewUpdateDbaasServiceOpensearchRequestWithBody generates requests for UpdateDbaasServiceOpensearch with any type of body
func NewUpdateDbaasServiceOpensearchRequestWithBody(server string, name DbaasServiceName, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-opensearch/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewGetDbaasServicePgRequest generates requests for GetDbaasServicePg
func NewGetDbaasServicePgRequest(server string, name DbaasServiceName) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-postgres/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateDbaasServicePgRequest calls the generic CreateDbaasServicePg builder with application/json body
func NewCreateDbaasServicePgRequest(server string, name DbaasServiceName, body CreateDbaasServicePgJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateDbaasServicePgRequestWithBody(server, name, "application/json", bodyReader)
}

// NewCreateDbaasServicePgRequestWithBody generates requests for CreateDbaasServicePg with any type of body
func NewCreateDbaasServicePgRequestWithBody(server string, name DbaasServiceName, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-postgres/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewUpdateDbaasServicePgRequest calls the generic UpdateDbaasServicePg builder with application/json body
func NewUpdateDbaasServicePgRequest(server string, name DbaasServiceName, body UpdateDbaasServicePgJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUpdateDbaasServicePgRequestWithBody(server, name, "application/json", bodyReader)
}

// NewUpdateDbaasServicePgRequestWithBody generates requests for UpdateDbaasServicePg with any type of body
func NewUpdateDbaasServicePgRequestWithBody(server string, name DbaasServiceName, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-postgres/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewGetDbaasServiceRedisRequest generates requests for GetDbaasServiceRedis
func NewGetDbaasServiceRedisRequest(server string, name DbaasServiceName) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-redis/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateDbaasServiceRedisRequest calls the generic CreateDbaasServiceRedis builder with application/json body
func NewCreateDbaasServiceRedisRequest(server string, name DbaasServiceName, body CreateDbaasServiceRedisJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateDbaasServiceRedisRequestWithBody(server, name, "application/json", bodyReader)
}

// NewCreateDbaasServiceRedisRequestWithBody generates requests for CreateDbaasServiceRedis with any type of body
func NewCreateDbaasServiceRedisRequestWithBody(server string, name DbaasServiceName, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-redis/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewUpdateDbaasServiceRedisRequest calls the generic UpdateDbaasServiceRedis builder with application/json body
func NewUpdateDbaasServiceRedisRequest(server string, name DbaasServiceName, body UpdateDbaasServiceRedisJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUpdateDbaasServiceRedisRequestWithBody(server, name, "application/json", bodyReader)
}

// NewUpdateDbaasServiceRedisRequestWithBody generates requests for UpdateDbaasServiceRedis with any type of body
func NewUpdateDbaasServiceRedisRequestWithBody(server string, name DbaasServiceName, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-redis/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewListDbaasServicesRequest generates requests for ListDbaasServices
func NewListDbaasServicesRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-service")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetDbaasServiceLogsRequest calls the generic GetDbaasServiceLogs builder with application/json body
func NewGetDbaasServiceLogsRequest(server string, serviceName string, body GetDbaasServiceLogsJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewGetDbaasServiceLogsRequestWithBody(server, serviceName, "application/json", bodyReader)
}

// NewGetDbaasServiceLogsRequestWithBody generates requests for GetDbaasServiceLogs with any type of body
func NewGetDbaasServiceLogsRequestWithBody(server string, serviceName string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "service-name", runtime.ParamLocationPath, serviceName)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-service-logs/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewGetDbaasServiceMetricsRequest calls the generic GetDbaasServiceMetrics builder with application/json body
func NewGetDbaasServiceMetricsRequest(server string, serviceName string, body GetDbaasServiceMetricsJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewGetDbaasServiceMetricsRequestWithBody(server, serviceName, "application/json", bodyReader)
}

// NewGetDbaasServiceMetricsRequestWithBody generates requests for GetDbaasServiceMetrics with any type of body
func NewGetDbaasServiceMetricsRequestWithBody(server string, serviceName string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "service-name", runtime.ParamLocationPath, serviceName)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-service-metrics/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewListDbaasServiceTypesRequest generates requests for ListDbaasServiceTypes
func NewListDbaasServiceTypesRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-service-type")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetDbaasServiceTypeRequest generates requests for GetDbaasServiceType
func NewGetDbaasServiceTypeRequest(server string, serviceTypeName string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "service-type-name", runtime.ParamLocationPath, serviceTypeName)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-service-type/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewDeleteDbaasServiceRequest generates requests for DeleteDbaasService
func NewDeleteDbaasServiceRequest(server string, name string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-service/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetDbaasSettingsKafkaRequest generates requests for GetDbaasSettingsKafka
func NewGetDbaasSettingsKafkaRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-settings-kafka")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetDbaasSettingsMysqlRequest generates requests for GetDbaasSettingsMysql
func NewGetDbaasSettingsMysqlRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-settings-mysql")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetDbaasSettingsOpensearchRequest generates requests for GetDbaasSettingsOpensearch
func NewGetDbaasSettingsOpensearchRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-settings-opensearch")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetDbaasSettingsPgRequest generates requests for GetDbaasSettingsPg
func NewGetDbaasSettingsPgRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-settings-pg")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetDbaasSettingsRedisRequest generates requests for GetDbaasSettingsRedis
func NewGetDbaasSettingsRedisRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dbaas-settings-redis")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewListDeployTargetsRequest generates requests for ListDeployTargets
func NewListDeployTargetsRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/deploy-target")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetDeployTargetRequest generates requests for GetDeployTarget
func NewGetDeployTargetRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/deploy-target/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewListDnsDomainsRequest generates requests for ListDnsDomains
func NewListDnsDomainsRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dns-domain")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetDnsDomainRequest generates requests for GetDnsDomain
func NewGetDnsDomainRequest(server string, id int64) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dns-domain/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewListDnsDomainRecordsRequest generates requests for ListDnsDomainRecords
func NewListDnsDomainRecordsRequest(server string, id int64) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dns-domain/%s/record", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetDnsDomainRecordRequest generates requests for GetDnsDomainRecord
func NewGetDnsDomainRecordRequest(server string, id int64, recordId int64) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "record-id", runtime.ParamLocationPath, recordId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/dns-domain/%s/record/%s", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewListElasticIpsRequest generates requests for ListElasticIps
func NewListElasticIpsRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/elastic-ip")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateElasticIpRequest calls the generic CreateElasticIp builder with application/json body
func NewCreateElasticIpRequest(server string, body CreateElasticIpJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateElasticIpRequestWithBody(server, "application/json", bodyReader)
}

// NewCreateElasticIpRequestWithBody generates requests for CreateElasticIp with any type of body
func NewCreateElasticIpRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/elastic-ip")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewDeleteElasticIpRequest generates requests for DeleteElasticIp
func NewDeleteElasticIpRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/elastic-ip/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetElasticIpRequest generates requests for GetElasticIp
func NewGetElasticIpRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/elastic-ip/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewUpdateElasticIpRequest calls the generic UpdateElasticIp builder with application/json body
func NewUpdateElasticIpRequest(server string, id string, body UpdateElasticIpJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUpdateElasticIpRequestWithBody(server, id, "application/json", bodyReader)
}

// NewUpdateElasticIpRequestWithBody generates requests for UpdateElasticIp with any type of body
func NewUpdateElasticIpRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/elastic-ip/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewResetElasticIpFieldRequest generates requests for ResetElasticIpField
func NewResetElasticIpFieldRequest(server string, id string, field ResetElasticIpFieldParamsField) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "field", runtime.ParamLocationPath, field)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/elastic-ip/%s/%s", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewAttachInstanceToElasticIpRequest calls the generic AttachInstanceToElasticIp builder with application/json body
func NewAttachInstanceToElasticIpRequest(server string, id string, body AttachInstanceToElasticIpJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewAttachInstanceToElasticIpRequestWithBody(server, id, "application/json", bodyReader)
}

// NewAttachInstanceToElasticIpRequestWithBody generates requests for AttachInstanceToElasticIp with any type of body
func NewAttachInstanceToElasticIpRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/elastic-ip/%s:attach", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewDetachInstanceFromElasticIpRequest calls the generic DetachInstanceFromElasticIp builder with application/json body
func NewDetachInstanceFromElasticIpRequest(server string, id string, body DetachInstanceFromElasticIpJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewDetachInstanceFromElasticIpRequestWithBody(server, id, "application/json", bodyReader)
}

// NewDetachInstanceFromElasticIpRequestWithBody generates requests for DetachInstanceFromElasticIp with any type of body
func NewDetachInstanceFromElasticIpRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/elastic-ip/%s:detach", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewListEventsRequest generates requests for ListEvents
func NewListEventsRequest(server string, params *ListEventsParams) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/event")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	queryValues := queryURL.Query()

	if params.From != nil {

		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "from", runtime.ParamLocationQuery, *params.From); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}

	}

	if params.To != nil {

		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "to", runtime.ParamLocationQuery, *params.To); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}

	}

	queryURL.RawQuery = queryValues.Encode()

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewListInstancesRequest generates requests for ListInstances
func NewListInstancesRequest(server string, params *ListInstancesParams) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	queryValues := queryURL.Query()

	if params.ManagerId != nil {

		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "manager-id", runtime.ParamLocationQuery, *params.ManagerId); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}

	}

	if params.ManagerType != nil {

		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "manager-type", runtime.ParamLocationQuery, *params.ManagerType); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}

	}

	queryURL.RawQuery = queryValues.Encode()

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateInstanceRequest calls the generic CreateInstance builder with application/json body
func NewCreateInstanceRequest(server string, body CreateInstanceJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateInstanceRequestWithBody(server, "application/json", bodyReader)
}

// NewCreateInstanceRequestWithBody generates requests for CreateInstance with any type of body
func NewCreateInstanceRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewListInstancePoolsRequest generates requests for ListInstancePools
func NewListInstancePoolsRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance-pool")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateInstancePoolRequest calls the generic CreateInstancePool builder with application/json body
func NewCreateInstancePoolRequest(server string, body CreateInstancePoolJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateInstancePoolRequestWithBody(server, "application/json", bodyReader)
}

// NewCreateInstancePoolRequestWithBody generates requests for CreateInstancePool with any type of body
func NewCreateInstancePoolRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance-pool")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewDeleteInstancePoolRequest generates requests for DeleteInstancePool
func NewDeleteInstancePoolRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance-pool/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetInstancePoolRequest generates requests for GetInstancePool
func NewGetInstancePoolRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance-pool/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewUpdateInstancePoolRequest calls the generic UpdateInstancePool builder with application/json body
func NewUpdateInstancePoolRequest(server string, id string, body UpdateInstancePoolJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUpdateInstancePoolRequestWithBody(server, id, "application/json", bodyReader)
}

// NewUpdateInstancePoolRequestWithBody generates requests for UpdateInstancePool with any type of body
func NewUpdateInstancePoolRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance-pool/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewResetInstancePoolFieldRequest generates requests for ResetInstancePoolField
func NewResetInstancePoolFieldRequest(server string, id string, field ResetInstancePoolFieldParamsField) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "field", runtime.ParamLocationPath, field)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance-pool/%s/%s", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewEvictInstancePoolMembersRequest calls the generic EvictInstancePoolMembers builder with application/json body
func NewEvictInstancePoolMembersRequest(server string, id string, body EvictInstancePoolMembersJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewEvictInstancePoolMembersRequestWithBody(server, id, "application/json", bodyReader)
}

// NewEvictInstancePoolMembersRequestWithBody generates requests for EvictInstancePoolMembers with any type of body
func NewEvictInstancePoolMembersRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance-pool/%s:evict", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewScaleInstancePoolRequest calls the generic ScaleInstancePool builder with application/json body
func NewScaleInstancePoolRequest(server string, id string, body ScaleInstancePoolJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewScaleInstancePoolRequestWithBody(server, id, "application/json", bodyReader)
}

// NewScaleInstancePoolRequestWithBody generates requests for ScaleInstancePool with any type of body
func NewScaleInstancePoolRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance-pool/%s:scale", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewListInstanceTypesRequest generates requests for ListInstanceTypes
func NewListInstanceTypesRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance-type")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetInstanceTypeRequest generates requests for GetInstanceType
func NewGetInstanceTypeRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance-type/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewDeleteInstanceRequest generates requests for DeleteInstance
func NewDeleteInstanceRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetInstanceRequest generates requests for GetInstance
func NewGetInstanceRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewUpdateInstanceRequest calls the generic UpdateInstance builder with application/json body
func NewUpdateInstanceRequest(server string, id string, body UpdateInstanceJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUpdateInstanceRequestWithBody(server, id, "application/json", bodyReader)
}

// NewUpdateInstanceRequestWithBody generates requests for UpdateInstance with any type of body
func NewUpdateInstanceRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewResetInstanceFieldRequest generates requests for ResetInstanceField
func NewResetInstanceFieldRequest(server string, id string, field ResetInstanceFieldParamsField) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "field", runtime.ParamLocationPath, field)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance/%s/%s", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateSnapshotRequest generates requests for CreateSnapshot
func NewCreateSnapshotRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance/%s:create-snapshot", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewRebootInstanceRequest generates requests for RebootInstance
func NewRebootInstanceRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance/%s:reboot", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewResetInstanceRequest calls the generic ResetInstance builder with application/json body
func NewResetInstanceRequest(server string, id string, body ResetInstanceJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewResetInstanceRequestWithBody(server, id, "application/json", bodyReader)
}

// NewResetInstanceRequestWithBody generates requests for ResetInstance with any type of body
func NewResetInstanceRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance/%s:reset", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewResizeInstanceDiskRequest calls the generic ResizeInstanceDisk builder with application/json body
func NewResizeInstanceDiskRequest(server string, id string, body ResizeInstanceDiskJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewResizeInstanceDiskRequestWithBody(server, id, "application/json", bodyReader)
}

// NewResizeInstanceDiskRequestWithBody generates requests for ResizeInstanceDisk with any type of body
func NewResizeInstanceDiskRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance/%s:resize-disk", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewScaleInstanceRequest calls the generic ScaleInstance builder with application/json body
func NewScaleInstanceRequest(server string, id string, body ScaleInstanceJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewScaleInstanceRequestWithBody(server, id, "application/json", bodyReader)
}

// NewScaleInstanceRequestWithBody generates requests for ScaleInstance with any type of body
func NewScaleInstanceRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance/%s:scale", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewStartInstanceRequest calls the generic StartInstance builder with application/json body
func NewStartInstanceRequest(server string, id string, body StartInstanceJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewStartInstanceRequestWithBody(server, id, "application/json", bodyReader)
}

// NewStartInstanceRequestWithBody generates requests for StartInstance with any type of body
func NewStartInstanceRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance/%s:start", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewStopInstanceRequest generates requests for StopInstance
func NewStopInstanceRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance/%s:stop", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewRevertInstanceToSnapshotRequest calls the generic RevertInstanceToSnapshot builder with application/json body
func NewRevertInstanceToSnapshotRequest(server string, instanceId string, body RevertInstanceToSnapshotJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewRevertInstanceToSnapshotRequestWithBody(server, instanceId, "application/json", bodyReader)
}

// NewRevertInstanceToSnapshotRequestWithBody generates requests for RevertInstanceToSnapshot with any type of body
func NewRevertInstanceToSnapshotRequestWithBody(server string, instanceId string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "instance-id", runtime.ParamLocationPath, instanceId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/instance/%s:revert-snapshot", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewListLoadBalancersRequest generates requests for ListLoadBalancers
func NewListLoadBalancersRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/load-balancer")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateLoadBalancerRequest calls the generic CreateLoadBalancer builder with application/json body
func NewCreateLoadBalancerRequest(server string, body CreateLoadBalancerJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateLoadBalancerRequestWithBody(server, "application/json", bodyReader)
}

// NewCreateLoadBalancerRequestWithBody generates requests for CreateLoadBalancer with any type of body
func NewCreateLoadBalancerRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/load-balancer")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewDeleteLoadBalancerRequest generates requests for DeleteLoadBalancer
func NewDeleteLoadBalancerRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/load-balancer/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetLoadBalancerRequest generates requests for GetLoadBalancer
func NewGetLoadBalancerRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/load-balancer/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewUpdateLoadBalancerRequest calls the generic UpdateLoadBalancer builder with application/json body
func NewUpdateLoadBalancerRequest(server string, id string, body UpdateLoadBalancerJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUpdateLoadBalancerRequestWithBody(server, id, "application/json", bodyReader)
}

// NewUpdateLoadBalancerRequestWithBody generates requests for UpdateLoadBalancer with any type of body
func NewUpdateLoadBalancerRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/load-balancer/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewAddServiceToLoadBalancerRequest calls the generic AddServiceToLoadBalancer builder with application/json body
func NewAddServiceToLoadBalancerRequest(server string, id string, body AddServiceToLoadBalancerJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewAddServiceToLoadBalancerRequestWithBody(server, id, "application/json", bodyReader)
}

// NewAddServiceToLoadBalancerRequestWithBody generates requests for AddServiceToLoadBalancer with any type of body
func NewAddServiceToLoadBalancerRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/load-balancer/%s/service", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewDeleteLoadBalancerServiceRequest generates requests for DeleteLoadBalancerService
func NewDeleteLoadBalancerServiceRequest(server string, id string, serviceId string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "service-id", runtime.ParamLocationPath, serviceId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/load-balancer/%s/service/%s", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetLoadBalancerServiceRequest generates requests for GetLoadBalancerService
func NewGetLoadBalancerServiceRequest(server string, id string, serviceId string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "service-id", runtime.ParamLocationPath, serviceId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/load-balancer/%s/service/%s", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewUpdateLoadBalancerServiceRequest calls the generic UpdateLoadBalancerService builder with application/json body
func NewUpdateLoadBalancerServiceRequest(server string, id string, serviceId string, body UpdateLoadBalancerServiceJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUpdateLoadBalancerServiceRequestWithBody(server, id, serviceId, "application/json", bodyReader)
}

// NewUpdateLoadBalancerServiceRequestWithBody generates requests for UpdateLoadBalancerService with any type of body
func NewUpdateLoadBalancerServiceRequestWithBody(server string, id string, serviceId string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "service-id", runtime.ParamLocationPath, serviceId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/load-balancer/%s/service/%s", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewResetLoadBalancerServiceFieldRequest generates requests for ResetLoadBalancerServiceField
func NewResetLoadBalancerServiceFieldRequest(server string, id string, serviceId string, field ResetLoadBalancerServiceFieldParamsField) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "service-id", runtime.ParamLocationPath, serviceId)
	if err != nil {
		return nil, err
	}

	var pathParam2 string

	pathParam2, err = runtime.StyleParamWithLocation("simple", false, "field", runtime.ParamLocationPath, field)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/load-balancer/%s/service/%s/%s", pathParam0, pathParam1, pathParam2)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewResetLoadBalancerFieldRequest generates requests for ResetLoadBalancerField
func NewResetLoadBalancerFieldRequest(server string, id string, field ResetLoadBalancerFieldParamsField) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "field", runtime.ParamLocationPath, field)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/load-balancer/%s/%s", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetOperationRequest generates requests for GetOperation
func NewGetOperationRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/operation/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewListPrivateNetworksRequest generates requests for ListPrivateNetworks
func NewListPrivateNetworksRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/private-network")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreatePrivateNetworkRequest calls the generic CreatePrivateNetwork builder with application/json body
func NewCreatePrivateNetworkRequest(server string, body CreatePrivateNetworkJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreatePrivateNetworkRequestWithBody(server, "application/json", bodyReader)
}

// NewCreatePrivateNetworkRequestWithBody generates requests for CreatePrivateNetwork with any type of body
func NewCreatePrivateNetworkRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/private-network")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewDeletePrivateNetworkRequest generates requests for DeletePrivateNetwork
func NewDeletePrivateNetworkRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/private-network/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetPrivateNetworkRequest generates requests for GetPrivateNetwork
func NewGetPrivateNetworkRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/private-network/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewUpdatePrivateNetworkRequest calls the generic UpdatePrivateNetwork builder with application/json body
func NewUpdatePrivateNetworkRequest(server string, id string, body UpdatePrivateNetworkJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUpdatePrivateNetworkRequestWithBody(server, id, "application/json", bodyReader)
}

// NewUpdatePrivateNetworkRequestWithBody generates requests for UpdatePrivateNetwork with any type of body
func NewUpdatePrivateNetworkRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/private-network/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewResetPrivateNetworkFieldRequest generates requests for ResetPrivateNetworkField
func NewResetPrivateNetworkFieldRequest(server string, id string, field ResetPrivateNetworkFieldParamsField) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "field", runtime.ParamLocationPath, field)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/private-network/%s/%s", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewAttachInstanceToPrivateNetworkRequest calls the generic AttachInstanceToPrivateNetwork builder with application/json body
func NewAttachInstanceToPrivateNetworkRequest(server string, id string, body AttachInstanceToPrivateNetworkJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewAttachInstanceToPrivateNetworkRequestWithBody(server, id, "application/json", bodyReader)
}

// NewAttachInstanceToPrivateNetworkRequestWithBody generates requests for AttachInstanceToPrivateNetwork with any type of body
func NewAttachInstanceToPrivateNetworkRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/private-network/%s:attach", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewDetachInstanceFromPrivateNetworkRequest calls the generic DetachInstanceFromPrivateNetwork builder with application/json body
func NewDetachInstanceFromPrivateNetworkRequest(server string, id string, body DetachInstanceFromPrivateNetworkJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewDetachInstanceFromPrivateNetworkRequestWithBody(server, id, "application/json", bodyReader)
}

// NewDetachInstanceFromPrivateNetworkRequestWithBody generates requests for DetachInstanceFromPrivateNetwork with any type of body
func NewDetachInstanceFromPrivateNetworkRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/private-network/%s:detach", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewUpdatePrivateNetworkInstanceIpRequest calls the generic UpdatePrivateNetworkInstanceIp builder with application/json body
func NewUpdatePrivateNetworkInstanceIpRequest(server string, id string, body UpdatePrivateNetworkInstanceIpJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUpdatePrivateNetworkInstanceIpRequestWithBody(server, id, "application/json", bodyReader)
}

// NewUpdatePrivateNetworkInstanceIpRequestWithBody generates requests for UpdatePrivateNetworkInstanceIp with any type of body
func NewUpdatePrivateNetworkInstanceIpRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/private-network/%s:update-ip", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewListQuotasRequest generates requests for ListQuotas
func NewListQuotasRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/quota")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetQuotaRequest generates requests for GetQuota
func NewGetQuotaRequest(server string, entity string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "entity", runtime.ParamLocationPath, entity)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/quota/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewListSecurityGroupsRequest generates requests for ListSecurityGroups
func NewListSecurityGroupsRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/security-group")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateSecurityGroupRequest calls the generic CreateSecurityGroup builder with application/json body
func NewCreateSecurityGroupRequest(server string, body CreateSecurityGroupJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateSecurityGroupRequestWithBody(server, "application/json", bodyReader)
}

// NewCreateSecurityGroupRequestWithBody generates requests for CreateSecurityGroup with any type of body
func NewCreateSecurityGroupRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/security-group")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewDeleteSecurityGroupRequest generates requests for DeleteSecurityGroup
func NewDeleteSecurityGroupRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/security-group/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetSecurityGroupRequest generates requests for GetSecurityGroup
func NewGetSecurityGroupRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/security-group/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewAddRuleToSecurityGroupRequest calls the generic AddRuleToSecurityGroup builder with application/json body
func NewAddRuleToSecurityGroupRequest(server string, id string, body AddRuleToSecurityGroupJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewAddRuleToSecurityGroupRequestWithBody(server, id, "application/json", bodyReader)
}

// NewAddRuleToSecurityGroupRequestWithBody generates requests for AddRuleToSecurityGroup with any type of body
func NewAddRuleToSecurityGroupRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/security-group/%s/rules", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewDeleteRuleFromSecurityGroupRequest generates requests for DeleteRuleFromSecurityGroup
func NewDeleteRuleFromSecurityGroupRequest(server string, id string, ruleId string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "rule-id", runtime.ParamLocationPath, ruleId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/security-group/%s/rules/%s", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewAddExternalSourceToSecurityGroupRequest calls the generic AddExternalSourceToSecurityGroup builder with application/json body
func NewAddExternalSourceToSecurityGroupRequest(server string, id string, body AddExternalSourceToSecurityGroupJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewAddExternalSourceToSecurityGroupRequestWithBody(server, id, "application/json", bodyReader)
}

// NewAddExternalSourceToSecurityGroupRequestWithBody generates requests for AddExternalSourceToSecurityGroup with any type of body
func NewAddExternalSourceToSecurityGroupRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/security-group/%s:add-source", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewAttachInstanceToSecurityGroupRequest calls the generic AttachInstanceToSecurityGroup builder with application/json body
func NewAttachInstanceToSecurityGroupRequest(server string, id string, body AttachInstanceToSecurityGroupJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewAttachInstanceToSecurityGroupRequestWithBody(server, id, "application/json", bodyReader)
}

// NewAttachInstanceToSecurityGroupRequestWithBody generates requests for AttachInstanceToSecurityGroup with any type of body
func NewAttachInstanceToSecurityGroupRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/security-group/%s:attach", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewDetachInstanceFromSecurityGroupRequest calls the generic DetachInstanceFromSecurityGroup builder with application/json body
func NewDetachInstanceFromSecurityGroupRequest(server string, id string, body DetachInstanceFromSecurityGroupJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewDetachInstanceFromSecurityGroupRequestWithBody(server, id, "application/json", bodyReader)
}

// NewDetachInstanceFromSecurityGroupRequestWithBody generates requests for DetachInstanceFromSecurityGroup with any type of body
func NewDetachInstanceFromSecurityGroupRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/security-group/%s:detach", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewRemoveExternalSourceFromSecurityGroupRequest calls the generic RemoveExternalSourceFromSecurityGroup builder with application/json body
func NewRemoveExternalSourceFromSecurityGroupRequest(server string, id string, body RemoveExternalSourceFromSecurityGroupJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewRemoveExternalSourceFromSecurityGroupRequestWithBody(server, id, "application/json", bodyReader)
}

// NewRemoveExternalSourceFromSecurityGroupRequestWithBody generates requests for RemoveExternalSourceFromSecurityGroup with any type of body
func NewRemoveExternalSourceFromSecurityGroupRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/security-group/%s:remove-source", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewListSksClustersRequest generates requests for ListSksClusters
func NewListSksClustersRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sks-cluster")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateSksClusterRequest calls the generic CreateSksCluster builder with application/json body
func NewCreateSksClusterRequest(server string, body CreateSksClusterJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateSksClusterRequestWithBody(server, "application/json", bodyReader)
}

// NewCreateSksClusterRequestWithBody generates requests for CreateSksCluster with any type of body
func NewCreateSksClusterRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sks-cluster")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewListSksClusterDeprecatedResourcesRequest generates requests for ListSksClusterDeprecatedResources
func NewListSksClusterDeprecatedResourcesRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sks-cluster-deprecated-resources/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGenerateSksClusterKubeconfigRequest calls the generic GenerateSksClusterKubeconfig builder with application/json body
func NewGenerateSksClusterKubeconfigRequest(server string, id string, body GenerateSksClusterKubeconfigJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewGenerateSksClusterKubeconfigRequestWithBody(server, id, "application/json", bodyReader)
}

// NewGenerateSksClusterKubeconfigRequestWithBody generates requests for GenerateSksClusterKubeconfig with any type of body
func NewGenerateSksClusterKubeconfigRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sks-cluster-kubeconfig/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewListSksClusterVersionsRequest generates requests for ListSksClusterVersions
func NewListSksClusterVersionsRequest(server string, params *ListSksClusterVersionsParams) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sks-cluster-version")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	queryValues := queryURL.Query()

	if params.IncludeDeprecated != nil {

		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "include-deprecated", runtime.ParamLocationQuery, *params.IncludeDeprecated); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}

	}

	queryURL.RawQuery = queryValues.Encode()

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewDeleteSksClusterRequest generates requests for DeleteSksCluster
func NewDeleteSksClusterRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sks-cluster/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetSksClusterRequest generates requests for GetSksCluster
func NewGetSksClusterRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sks-cluster/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewUpdateSksClusterRequest calls the generic UpdateSksCluster builder with application/json body
func NewUpdateSksClusterRequest(server string, id string, body UpdateSksClusterJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUpdateSksClusterRequestWithBody(server, id, "application/json", bodyReader)
}

// NewUpdateSksClusterRequestWithBody generates requests for UpdateSksCluster with any type of body
func NewUpdateSksClusterRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sks-cluster/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewGetSksClusterAuthorityCertRequest generates requests for GetSksClusterAuthorityCert
func NewGetSksClusterAuthorityCertRequest(server string, id string, authority GetSksClusterAuthorityCertParamsAuthority) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "authority", runtime.ParamLocationPath, authority)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sks-cluster/%s/authority/%s/cert", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCreateSksNodepoolRequest calls the generic CreateSksNodepool builder with application/json body
func NewCreateSksNodepoolRequest(server string, id string, body CreateSksNodepoolJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCreateSksNodepoolRequestWithBody(server, id, "application/json", bodyReader)
}

// NewCreateSksNodepoolRequestWithBody generates requests for CreateSksNodepool with any type of body
func NewCreateSksNodepoolRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sks-cluster/%s/nodepool", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewDeleteSksNodepoolRequest generates requests for DeleteSksNodepool
func NewDeleteSksNodepoolRequest(server string, id string, sksNodepoolId string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "sks-nodepool-id", runtime.ParamLocationPath, sksNodepoolId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sks-cluster/%s/nodepool/%s", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetSksNodepoolRequest generates requests for GetSksNodepool
func NewGetSksNodepoolRequest(server string, id string, sksNodepoolId string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "sks-nodepool-id", runtime.ParamLocationPath, sksNodepoolId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sks-cluster/%s/nodepool/%s", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewUpdateSksNodepoolRequest calls the generic UpdateSksNodepool builder with application/json body
func NewUpdateSksNodepoolRequest(server string, id string, sksNodepoolId string, body UpdateSksNodepoolJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUpdateSksNodepoolRequestWithBody(server, id, sksNodepoolId, "application/json", bodyReader)
}

// NewUpdateSksNodepoolRequestWithBody generates requests for UpdateSksNodepool with any type of body
func NewUpdateSksNodepoolRequestWithBody(server string, id string, sksNodepoolId string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "sks-nodepool-id", runtime.ParamLocationPath, sksNodepoolId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sks-cluster/%s/nodepool/%s", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewResetSksNodepoolFieldRequest generates requests for ResetSksNodepoolField
func NewResetSksNodepoolFieldRequest(server string, id string, sksNodepoolId string, field ResetSksNodepoolFieldParamsField) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "sks-nodepool-id", runtime.ParamLocationPath, sksNodepoolId)
	if err != nil {
		return nil, err
	}

	var pathParam2 string

	pathParam2, err = runtime.StyleParamWithLocation("simple", false, "field", runtime.ParamLocationPath, field)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sks-cluster/%s/nodepool/%s/%s", pathParam0, pathParam1, pathParam2)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewEvictSksNodepoolMembersRequest calls the generic EvictSksNodepoolMembers builder with application/json body
func NewEvictSksNodepoolMembersRequest(server string, id string, sksNodepoolId string, body EvictSksNodepoolMembersJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewEvictSksNodepoolMembersRequestWithBody(server, id, sksNodepoolId, "application/json", bodyReader)
}

// NewEvictSksNodepoolMembersRequestWithBody generates requests for EvictSksNodepoolMembers with any type of body
func NewEvictSksNodepoolMembersRequestWithBody(server string, id string, sksNodepoolId string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "sks-nodepool-id", runtime.ParamLocationPath, sksNodepoolId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sks-cluster/%s/nodepool/%s:evict", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewScaleSksNodepoolRequest calls the generic ScaleSksNodepool builder with application/json body
func NewScaleSksNodepoolRequest(server string, id string, sksNodepoolId string, body ScaleSksNodepoolJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewScaleSksNodepoolRequestWithBody(server, id, sksNodepoolId, "application/json", bodyReader)
}

// NewScaleSksNodepoolRequestWithBody generates requests for ScaleSksNodepool with any type of body
func NewScaleSksNodepoolRequestWithBody(server string, id string, sksNodepoolId string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "sks-nodepool-id", runtime.ParamLocationPath, sksNodepoolId)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sks-cluster/%s/nodepool/%s:scale", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewRotateSksCcmCredentialsRequest generates requests for RotateSksCcmCredentials
func NewRotateSksCcmCredentialsRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sks-cluster/%s/rotate-ccm-credentials", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewRotateSksOperatorsCaRequest generates requests for RotateSksOperatorsCa
func NewRotateSksOperatorsCaRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sks-cluster/%s/rotate-operators-ca", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewUpgradeSksClusterRequest calls the generic UpgradeSksCluster builder with application/json body
func NewUpgradeSksClusterRequest(server string, id string, body UpgradeSksClusterJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUpgradeSksClusterRequestWithBody(server, id, "application/json", bodyReader)
}

// NewUpgradeSksClusterRequestWithBody generates requests for UpgradeSksCluster with any type of body
func NewUpgradeSksClusterRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sks-cluster/%s/upgrade", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewUpgradeSksClusterServiceLevelRequest generates requests for UpgradeSksClusterServiceLevel
func NewUpgradeSksClusterServiceLevelRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sks-cluster/%s/upgrade-service-level", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewResetSksClusterFieldRequest generates requests for ResetSksClusterField
func NewResetSksClusterFieldRequest(server string, id string, field ResetSksClusterFieldParamsField) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	var pathParam1 string

	pathParam1, err = runtime.StyleParamWithLocation("simple", false, "field", runtime.ParamLocationPath, field)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sks-cluster/%s/%s", pathParam0, pathParam1)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewListSnapshotsRequest generates requests for ListSnapshots
func NewListSnapshotsRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/snapshot")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewDeleteSnapshotRequest generates requests for DeleteSnapshot
func NewDeleteSnapshotRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/snapshot/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetSnapshotRequest generates requests for GetSnapshot
func NewGetSnapshotRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/snapshot/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewExportSnapshotRequest generates requests for ExportSnapshot
func NewExportSnapshotRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/snapshot/%s:export", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewPromoteSnapshotToTemplateRequest calls the generic PromoteSnapshotToTemplate builder with application/json body
func NewPromoteSnapshotToTemplateRequest(server string, id string, body PromoteSnapshotToTemplateJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewPromoteSnapshotToTemplateRequestWithBody(server, id, "application/json", bodyReader)
}

// NewPromoteSnapshotToTemplateRequestWithBody generates requests for PromoteSnapshotToTemplate with any type of body
func NewPromoteSnapshotToTemplateRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/snapshot/%s:promote", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewGetSosPresignedUrlRequest generates requests for GetSosPresignedUrl
func NewGetSosPresignedUrlRequest(server string, bucket string, params *GetSosPresignedUrlParams) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "bucket", runtime.ParamLocationPath, bucket)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/sos/%s/presigned-url", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	queryValues := queryURL.Query()

	if params.Key != nil {

		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "key", runtime.ParamLocationQuery, *params.Key); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}

	}

	queryURL.RawQuery = queryValues.Encode()

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewListSshKeysRequest generates requests for ListSshKeys
func NewListSshKeysRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/ssh-key")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewRegisterSshKeyRequest calls the generic RegisterSshKey builder with application/json body
func NewRegisterSshKeyRequest(server string, body RegisterSshKeyJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewRegisterSshKeyRequestWithBody(server, "application/json", bodyReader)
}

// NewRegisterSshKeyRequestWithBody generates requests for RegisterSshKey with any type of body
func NewRegisterSshKeyRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/ssh-key")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewDeleteSshKeyRequest generates requests for DeleteSshKey
func NewDeleteSshKeyRequest(server string, name string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/ssh-key/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetSshKeyRequest generates requests for GetSshKey
func NewGetSshKeyRequest(server string, name string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "name", runtime.ParamLocationPath, name)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/ssh-key/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewListTemplatesRequest generates requests for ListTemplates
func NewListTemplatesRequest(server string, params *ListTemplatesParams) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/template")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	queryValues := queryURL.Query()

	if params.Visibility != nil {

		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "visibility", runtime.ParamLocationQuery, *params.Visibility); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}

	}

	if params.Family != nil {

		if queryFrag, err := runtime.StyleParamWithLocation("form", true, "family", runtime.ParamLocationQuery, *params.Family); err != nil {
			return nil, err
		} else if parsed, err := url.ParseQuery(queryFrag); err != nil {
			return nil, err
		} else {
			for k, v := range parsed {
				for _, v2 := range v {
					queryValues.Add(k, v2)
				}
			}
		}

	}

	queryURL.RawQuery = queryValues.Encode()

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewRegisterTemplateRequest calls the generic RegisterTemplate builder with application/json body
func NewRegisterTemplateRequest(server string, body RegisterTemplateJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewRegisterTemplateRequestWithBody(server, "application/json", bodyReader)
}

// NewRegisterTemplateRequestWithBody generates requests for RegisterTemplate with any type of body
func NewRegisterTemplateRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/template")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewDeleteTemplateRequest generates requests for DeleteTemplate
func NewDeleteTemplateRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/template/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("DELETE", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewGetTemplateRequest generates requests for GetTemplate
func NewGetTemplateRequest(server string, id string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/template/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// NewCopyTemplateRequest calls the generic CopyTemplate builder with application/json body
func NewCopyTemplateRequest(server string, id string, body CopyTemplateJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewCopyTemplateRequestWithBody(server, id, "application/json", bodyReader)
}

// NewCopyTemplateRequestWithBody generates requests for CopyTemplate with any type of body
func NewCopyTemplateRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/template/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewUpdateTemplateRequest calls the generic UpdateTemplate builder with application/json body
func NewUpdateTemplateRequest(server string, id string, body UpdateTemplateJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewUpdateTemplateRequestWithBody(server, id, "application/json", bodyReader)
}

// NewUpdateTemplateRequestWithBody generates requests for UpdateTemplate with any type of body
func NewUpdateTemplateRequestWithBody(server string, id string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "id", runtime.ParamLocationPath, id)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/template/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewListZonesRequest generates requests for ListZones
func NewListZonesRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/zone")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// ListAccessKeys request
	ListAccessKeysWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListAccessKeysResponse, error)

	// CreateAccessKey request with any body
	CreateAccessKeyWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateAccessKeyResponse, error)

	CreateAccessKeyWithResponse(ctx context.Context, body CreateAccessKeyJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateAccessKeyResponse, error)

	// ListAccessKeyKnownOperations request
	ListAccessKeyKnownOperationsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListAccessKeyKnownOperationsResponse, error)

	// ListAccessKeyOperations request
	ListAccessKeyOperationsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListAccessKeyOperationsResponse, error)

	// RevokeAccessKey request
	RevokeAccessKeyWithResponse(ctx context.Context, key string, reqEditors ...RequestEditorFn) (*RevokeAccessKeyResponse, error)

	// GetAccessKey request
	GetAccessKeyWithResponse(ctx context.Context, key string, reqEditors ...RequestEditorFn) (*GetAccessKeyResponse, error)

	// ListAntiAffinityGroups request
	ListAntiAffinityGroupsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListAntiAffinityGroupsResponse, error)

	// CreateAntiAffinityGroup request with any body
	CreateAntiAffinityGroupWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateAntiAffinityGroupResponse, error)

	CreateAntiAffinityGroupWithResponse(ctx context.Context, body CreateAntiAffinityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateAntiAffinityGroupResponse, error)

	// DeleteAntiAffinityGroup request
	DeleteAntiAffinityGroupWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*DeleteAntiAffinityGroupResponse, error)

	// GetAntiAffinityGroup request
	GetAntiAffinityGroupWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetAntiAffinityGroupResponse, error)

	// GetDbaasCaCertificate request
	GetDbaasCaCertificateWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetDbaasCaCertificateResponse, error)

	// GetDbaasServiceKafka request
	GetDbaasServiceKafkaWithResponse(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*GetDbaasServiceKafkaResponse, error)

	// CreateDbaasServiceKafka request with any body
	CreateDbaasServiceKafkaWithBodyWithResponse(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateDbaasServiceKafkaResponse, error)

	CreateDbaasServiceKafkaWithResponse(ctx context.Context, name DbaasServiceName, body CreateDbaasServiceKafkaJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateDbaasServiceKafkaResponse, error)

	// UpdateDbaasServiceKafka request with any body
	UpdateDbaasServiceKafkaWithBodyWithResponse(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateDbaasServiceKafkaResponse, error)

	UpdateDbaasServiceKafkaWithResponse(ctx context.Context, name DbaasServiceName, body UpdateDbaasServiceKafkaJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateDbaasServiceKafkaResponse, error)

	// GetDbaasMigrationStatus request
	GetDbaasMigrationStatusWithResponse(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*GetDbaasMigrationStatusResponse, error)

	// GetDbaasServiceMysql request
	GetDbaasServiceMysqlWithResponse(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*GetDbaasServiceMysqlResponse, error)

	// CreateDbaasServiceMysql request with any body
	CreateDbaasServiceMysqlWithBodyWithResponse(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateDbaasServiceMysqlResponse, error)

	CreateDbaasServiceMysqlWithResponse(ctx context.Context, name DbaasServiceName, body CreateDbaasServiceMysqlJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateDbaasServiceMysqlResponse, error)

	// UpdateDbaasServiceMysql request with any body
	UpdateDbaasServiceMysqlWithBodyWithResponse(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateDbaasServiceMysqlResponse, error)

	UpdateDbaasServiceMysqlWithResponse(ctx context.Context, name DbaasServiceName, body UpdateDbaasServiceMysqlJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateDbaasServiceMysqlResponse, error)

	// GetDbaasServiceOpensearch request
	GetDbaasServiceOpensearchWithResponse(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*GetDbaasServiceOpensearchResponse, error)

	// CreateDbaasServiceOpensearch request with any body
	CreateDbaasServiceOpensearchWithBodyWithResponse(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateDbaasServiceOpensearchResponse, error)

	CreateDbaasServiceOpensearchWithResponse(ctx context.Context, name DbaasServiceName, body CreateDbaasServiceOpensearchJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateDbaasServiceOpensearchResponse, error)

	// UpdateDbaasServiceOpensearch request with any body
	UpdateDbaasServiceOpensearchWithBodyWithResponse(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateDbaasServiceOpensearchResponse, error)

	UpdateDbaasServiceOpensearchWithResponse(ctx context.Context, name DbaasServiceName, body UpdateDbaasServiceOpensearchJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateDbaasServiceOpensearchResponse, error)

	// GetDbaasServicePg request
	GetDbaasServicePgWithResponse(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*GetDbaasServicePgResponse, error)

	// CreateDbaasServicePg request with any body
	CreateDbaasServicePgWithBodyWithResponse(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateDbaasServicePgResponse, error)

	CreateDbaasServicePgWithResponse(ctx context.Context, name DbaasServiceName, body CreateDbaasServicePgJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateDbaasServicePgResponse, error)

	// UpdateDbaasServicePg request with any body
	UpdateDbaasServicePgWithBodyWithResponse(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateDbaasServicePgResponse, error)

	UpdateDbaasServicePgWithResponse(ctx context.Context, name DbaasServiceName, body UpdateDbaasServicePgJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateDbaasServicePgResponse, error)

	// GetDbaasServiceRedis request
	GetDbaasServiceRedisWithResponse(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*GetDbaasServiceRedisResponse, error)

	// CreateDbaasServiceRedis request with any body
	CreateDbaasServiceRedisWithBodyWithResponse(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateDbaasServiceRedisResponse, error)

	CreateDbaasServiceRedisWithResponse(ctx context.Context, name DbaasServiceName, body CreateDbaasServiceRedisJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateDbaasServiceRedisResponse, error)

	// UpdateDbaasServiceRedis request with any body
	UpdateDbaasServiceRedisWithBodyWithResponse(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateDbaasServiceRedisResponse, error)

	UpdateDbaasServiceRedisWithResponse(ctx context.Context, name DbaasServiceName, body UpdateDbaasServiceRedisJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateDbaasServiceRedisResponse, error)

	// ListDbaasServices request
	ListDbaasServicesWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListDbaasServicesResponse, error)

	// GetDbaasServiceLogs request with any body
	GetDbaasServiceLogsWithBodyWithResponse(ctx context.Context, serviceName string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*GetDbaasServiceLogsResponse, error)

	GetDbaasServiceLogsWithResponse(ctx context.Context, serviceName string, body GetDbaasServiceLogsJSONRequestBody, reqEditors ...RequestEditorFn) (*GetDbaasServiceLogsResponse, error)

	// GetDbaasServiceMetrics request with any body
	GetDbaasServiceMetricsWithBodyWithResponse(ctx context.Context, serviceName string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*GetDbaasServiceMetricsResponse, error)

	GetDbaasServiceMetricsWithResponse(ctx context.Context, serviceName string, body GetDbaasServiceMetricsJSONRequestBody, reqEditors ...RequestEditorFn) (*GetDbaasServiceMetricsResponse, error)

	// ListDbaasServiceTypes request
	ListDbaasServiceTypesWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListDbaasServiceTypesResponse, error)

	// GetDbaasServiceType request
	GetDbaasServiceTypeWithResponse(ctx context.Context, serviceTypeName string, reqEditors ...RequestEditorFn) (*GetDbaasServiceTypeResponse, error)

	// DeleteDbaasService request
	DeleteDbaasServiceWithResponse(ctx context.Context, name string, reqEditors ...RequestEditorFn) (*DeleteDbaasServiceResponse, error)

	// GetDbaasSettingsKafka request
	GetDbaasSettingsKafkaWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetDbaasSettingsKafkaResponse, error)

	// GetDbaasSettingsMysql request
	GetDbaasSettingsMysqlWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetDbaasSettingsMysqlResponse, error)

	// GetDbaasSettingsOpensearch request
	GetDbaasSettingsOpensearchWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetDbaasSettingsOpensearchResponse, error)

	// GetDbaasSettingsPg request
	GetDbaasSettingsPgWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetDbaasSettingsPgResponse, error)

	// GetDbaasSettingsRedis request
	GetDbaasSettingsRedisWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetDbaasSettingsRedisResponse, error)

	// ListDeployTargets request
	ListDeployTargetsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListDeployTargetsResponse, error)

	// GetDeployTarget request
	GetDeployTargetWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetDeployTargetResponse, error)

	// ListDnsDomains request
	ListDnsDomainsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListDnsDomainsResponse, error)

	// GetDnsDomain request
	GetDnsDomainWithResponse(ctx context.Context, id int64, reqEditors ...RequestEditorFn) (*GetDnsDomainResponse, error)

	// ListDnsDomainRecords request
	ListDnsDomainRecordsWithResponse(ctx context.Context, id int64, reqEditors ...RequestEditorFn) (*ListDnsDomainRecordsResponse, error)

	// GetDnsDomainRecord request
	GetDnsDomainRecordWithResponse(ctx context.Context, id int64, recordId int64, reqEditors ...RequestEditorFn) (*GetDnsDomainRecordResponse, error)

	// ListElasticIps request
	ListElasticIpsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListElasticIpsResponse, error)

	// CreateElasticIp request with any body
	CreateElasticIpWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateElasticIpResponse, error)

	CreateElasticIpWithResponse(ctx context.Context, body CreateElasticIpJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateElasticIpResponse, error)

	// DeleteElasticIp request
	DeleteElasticIpWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*DeleteElasticIpResponse, error)

	// GetElasticIp request
	GetElasticIpWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetElasticIpResponse, error)

	// UpdateElasticIp request with any body
	UpdateElasticIpWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateElasticIpResponse, error)

	UpdateElasticIpWithResponse(ctx context.Context, id string, body UpdateElasticIpJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateElasticIpResponse, error)

	// ResetElasticIpField request
	ResetElasticIpFieldWithResponse(ctx context.Context, id string, field ResetElasticIpFieldParamsField, reqEditors ...RequestEditorFn) (*ResetElasticIpFieldResponse, error)

	// AttachInstanceToElasticIp request with any body
	AttachInstanceToElasticIpWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*AttachInstanceToElasticIpResponse, error)

	AttachInstanceToElasticIpWithResponse(ctx context.Context, id string, body AttachInstanceToElasticIpJSONRequestBody, reqEditors ...RequestEditorFn) (*AttachInstanceToElasticIpResponse, error)

	// DetachInstanceFromElasticIp request with any body
	DetachInstanceFromElasticIpWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*DetachInstanceFromElasticIpResponse, error)

	DetachInstanceFromElasticIpWithResponse(ctx context.Context, id string, body DetachInstanceFromElasticIpJSONRequestBody, reqEditors ...RequestEditorFn) (*DetachInstanceFromElasticIpResponse, error)

	// ListEvents request
	ListEventsWithResponse(ctx context.Context, params *ListEventsParams, reqEditors ...RequestEditorFn) (*ListEventsResponse, error)

	// ListInstances request
	ListInstancesWithResponse(ctx context.Context, params *ListInstancesParams, reqEditors ...RequestEditorFn) (*ListInstancesResponse, error)

	// CreateInstance request with any body
	CreateInstanceWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateInstanceResponse, error)

	CreateInstanceWithResponse(ctx context.Context, body CreateInstanceJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateInstanceResponse, error)

	// ListInstancePools request
	ListInstancePoolsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListInstancePoolsResponse, error)

	// CreateInstancePool request with any body
	CreateInstancePoolWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateInstancePoolResponse, error)

	CreateInstancePoolWithResponse(ctx context.Context, body CreateInstancePoolJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateInstancePoolResponse, error)

	// DeleteInstancePool request
	DeleteInstancePoolWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*DeleteInstancePoolResponse, error)

	// GetInstancePool request
	GetInstancePoolWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetInstancePoolResponse, error)

	// UpdateInstancePool request with any body
	UpdateInstancePoolWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateInstancePoolResponse, error)

	UpdateInstancePoolWithResponse(ctx context.Context, id string, body UpdateInstancePoolJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateInstancePoolResponse, error)

	// ResetInstancePoolField request
	ResetInstancePoolFieldWithResponse(ctx context.Context, id string, field ResetInstancePoolFieldParamsField, reqEditors ...RequestEditorFn) (*ResetInstancePoolFieldResponse, error)

	// EvictInstancePoolMembers request with any body
	EvictInstancePoolMembersWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*EvictInstancePoolMembersResponse, error)

	EvictInstancePoolMembersWithResponse(ctx context.Context, id string, body EvictInstancePoolMembersJSONRequestBody, reqEditors ...RequestEditorFn) (*EvictInstancePoolMembersResponse, error)

	// ScaleInstancePool request with any body
	ScaleInstancePoolWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ScaleInstancePoolResponse, error)

	ScaleInstancePoolWithResponse(ctx context.Context, id string, body ScaleInstancePoolJSONRequestBody, reqEditors ...RequestEditorFn) (*ScaleInstancePoolResponse, error)

	// ListInstanceTypes request
	ListInstanceTypesWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListInstanceTypesResponse, error)

	// GetInstanceType request
	GetInstanceTypeWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetInstanceTypeResponse, error)

	// DeleteInstance request
	DeleteInstanceWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*DeleteInstanceResponse, error)

	// GetInstance request
	GetInstanceWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetInstanceResponse, error)

	// UpdateInstance request with any body
	UpdateInstanceWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateInstanceResponse, error)

	UpdateInstanceWithResponse(ctx context.Context, id string, body UpdateInstanceJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateInstanceResponse, error)

	// ResetInstanceField request
	ResetInstanceFieldWithResponse(ctx context.Context, id string, field ResetInstanceFieldParamsField, reqEditors ...RequestEditorFn) (*ResetInstanceFieldResponse, error)

	// CreateSnapshot request
	CreateSnapshotWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*CreateSnapshotResponse, error)

	// RebootInstance request
	RebootInstanceWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*RebootInstanceResponse, error)

	// ResetInstance request with any body
	ResetInstanceWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ResetInstanceResponse, error)

	ResetInstanceWithResponse(ctx context.Context, id string, body ResetInstanceJSONRequestBody, reqEditors ...RequestEditorFn) (*ResetInstanceResponse, error)

	// ResizeInstanceDisk request with any body
	ResizeInstanceDiskWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ResizeInstanceDiskResponse, error)

	ResizeInstanceDiskWithResponse(ctx context.Context, id string, body ResizeInstanceDiskJSONRequestBody, reqEditors ...RequestEditorFn) (*ResizeInstanceDiskResponse, error)

	// ScaleInstance request with any body
	ScaleInstanceWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ScaleInstanceResponse, error)

	ScaleInstanceWithResponse(ctx context.Context, id string, body ScaleInstanceJSONRequestBody, reqEditors ...RequestEditorFn) (*ScaleInstanceResponse, error)

	// StartInstance request with any body
	StartInstanceWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*StartInstanceResponse, error)

	StartInstanceWithResponse(ctx context.Context, id string, body StartInstanceJSONRequestBody, reqEditors ...RequestEditorFn) (*StartInstanceResponse, error)

	// StopInstance request
	StopInstanceWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*StopInstanceResponse, error)

	// RevertInstanceToSnapshot request with any body
	RevertInstanceToSnapshotWithBodyWithResponse(ctx context.Context, instanceId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*RevertInstanceToSnapshotResponse, error)

	RevertInstanceToSnapshotWithResponse(ctx context.Context, instanceId string, body RevertInstanceToSnapshotJSONRequestBody, reqEditors ...RequestEditorFn) (*RevertInstanceToSnapshotResponse, error)

	// ListLoadBalancers request
	ListLoadBalancersWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListLoadBalancersResponse, error)

	// CreateLoadBalancer request with any body
	CreateLoadBalancerWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateLoadBalancerResponse, error)

	CreateLoadBalancerWithResponse(ctx context.Context, body CreateLoadBalancerJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateLoadBalancerResponse, error)

	// DeleteLoadBalancer request
	DeleteLoadBalancerWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*DeleteLoadBalancerResponse, error)

	// GetLoadBalancer request
	GetLoadBalancerWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetLoadBalancerResponse, error)

	// UpdateLoadBalancer request with any body
	UpdateLoadBalancerWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateLoadBalancerResponse, error)

	UpdateLoadBalancerWithResponse(ctx context.Context, id string, body UpdateLoadBalancerJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateLoadBalancerResponse, error)

	// AddServiceToLoadBalancer request with any body
	AddServiceToLoadBalancerWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*AddServiceToLoadBalancerResponse, error)

	AddServiceToLoadBalancerWithResponse(ctx context.Context, id string, body AddServiceToLoadBalancerJSONRequestBody, reqEditors ...RequestEditorFn) (*AddServiceToLoadBalancerResponse, error)

	// DeleteLoadBalancerService request
	DeleteLoadBalancerServiceWithResponse(ctx context.Context, id string, serviceId string, reqEditors ...RequestEditorFn) (*DeleteLoadBalancerServiceResponse, error)

	// GetLoadBalancerService request
	GetLoadBalancerServiceWithResponse(ctx context.Context, id string, serviceId string, reqEditors ...RequestEditorFn) (*GetLoadBalancerServiceResponse, error)

	// UpdateLoadBalancerService request with any body
	UpdateLoadBalancerServiceWithBodyWithResponse(ctx context.Context, id string, serviceId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateLoadBalancerServiceResponse, error)

	UpdateLoadBalancerServiceWithResponse(ctx context.Context, id string, serviceId string, body UpdateLoadBalancerServiceJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateLoadBalancerServiceResponse, error)

	// ResetLoadBalancerServiceField request
	ResetLoadBalancerServiceFieldWithResponse(ctx context.Context, id string, serviceId string, field ResetLoadBalancerServiceFieldParamsField, reqEditors ...RequestEditorFn) (*ResetLoadBalancerServiceFieldResponse, error)

	// ResetLoadBalancerField request
	ResetLoadBalancerFieldWithResponse(ctx context.Context, id string, field ResetLoadBalancerFieldParamsField, reqEditors ...RequestEditorFn) (*ResetLoadBalancerFieldResponse, error)

	// GetOperation request
	GetOperationWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetOperationResponse, error)

	// ListPrivateNetworks request
	ListPrivateNetworksWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListPrivateNetworksResponse, error)

	// CreatePrivateNetwork request with any body
	CreatePrivateNetworkWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreatePrivateNetworkResponse, error)

	CreatePrivateNetworkWithResponse(ctx context.Context, body CreatePrivateNetworkJSONRequestBody, reqEditors ...RequestEditorFn) (*CreatePrivateNetworkResponse, error)

	// DeletePrivateNetwork request
	DeletePrivateNetworkWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*DeletePrivateNetworkResponse, error)

	// GetPrivateNetwork request
	GetPrivateNetworkWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetPrivateNetworkResponse, error)

	// UpdatePrivateNetwork request with any body
	UpdatePrivateNetworkWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdatePrivateNetworkResponse, error)

	UpdatePrivateNetworkWithResponse(ctx context.Context, id string, body UpdatePrivateNetworkJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdatePrivateNetworkResponse, error)

	// ResetPrivateNetworkField request
	ResetPrivateNetworkFieldWithResponse(ctx context.Context, id string, field ResetPrivateNetworkFieldParamsField, reqEditors ...RequestEditorFn) (*ResetPrivateNetworkFieldResponse, error)

	// AttachInstanceToPrivateNetwork request with any body
	AttachInstanceToPrivateNetworkWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*AttachInstanceToPrivateNetworkResponse, error)

	AttachInstanceToPrivateNetworkWithResponse(ctx context.Context, id string, body AttachInstanceToPrivateNetworkJSONRequestBody, reqEditors ...RequestEditorFn) (*AttachInstanceToPrivateNetworkResponse, error)

	// DetachInstanceFromPrivateNetwork request with any body
	DetachInstanceFromPrivateNetworkWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*DetachInstanceFromPrivateNetworkResponse, error)

	DetachInstanceFromPrivateNetworkWithResponse(ctx context.Context, id string, body DetachInstanceFromPrivateNetworkJSONRequestBody, reqEditors ...RequestEditorFn) (*DetachInstanceFromPrivateNetworkResponse, error)

	// UpdatePrivateNetworkInstanceIp request with any body
	UpdatePrivateNetworkInstanceIpWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdatePrivateNetworkInstanceIpResponse, error)

	UpdatePrivateNetworkInstanceIpWithResponse(ctx context.Context, id string, body UpdatePrivateNetworkInstanceIpJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdatePrivateNetworkInstanceIpResponse, error)

	// ListQuotas request
	ListQuotasWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListQuotasResponse, error)

	// GetQuota request
	GetQuotaWithResponse(ctx context.Context, entity string, reqEditors ...RequestEditorFn) (*GetQuotaResponse, error)

	// ListSecurityGroups request
	ListSecurityGroupsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListSecurityGroupsResponse, error)

	// CreateSecurityGroup request with any body
	CreateSecurityGroupWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateSecurityGroupResponse, error)

	CreateSecurityGroupWithResponse(ctx context.Context, body CreateSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateSecurityGroupResponse, error)

	// DeleteSecurityGroup request
	DeleteSecurityGroupWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*DeleteSecurityGroupResponse, error)

	// GetSecurityGroup request
	GetSecurityGroupWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetSecurityGroupResponse, error)

	// AddRuleToSecurityGroup request with any body
	AddRuleToSecurityGroupWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*AddRuleToSecurityGroupResponse, error)

	AddRuleToSecurityGroupWithResponse(ctx context.Context, id string, body AddRuleToSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*AddRuleToSecurityGroupResponse, error)

	// DeleteRuleFromSecurityGroup request
	DeleteRuleFromSecurityGroupWithResponse(ctx context.Context, id string, ruleId string, reqEditors ...RequestEditorFn) (*DeleteRuleFromSecurityGroupResponse, error)

	// AddExternalSourceToSecurityGroup request with any body
	AddExternalSourceToSecurityGroupWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*AddExternalSourceToSecurityGroupResponse, error)

	AddExternalSourceToSecurityGroupWithResponse(ctx context.Context, id string, body AddExternalSourceToSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*AddExternalSourceToSecurityGroupResponse, error)

	// AttachInstanceToSecurityGroup request with any body
	AttachInstanceToSecurityGroupWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*AttachInstanceToSecurityGroupResponse, error)

	AttachInstanceToSecurityGroupWithResponse(ctx context.Context, id string, body AttachInstanceToSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*AttachInstanceToSecurityGroupResponse, error)

	// DetachInstanceFromSecurityGroup request with any body
	DetachInstanceFromSecurityGroupWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*DetachInstanceFromSecurityGroupResponse, error)

	DetachInstanceFromSecurityGroupWithResponse(ctx context.Context, id string, body DetachInstanceFromSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*DetachInstanceFromSecurityGroupResponse, error)

	// RemoveExternalSourceFromSecurityGroup request with any body
	RemoveExternalSourceFromSecurityGroupWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*RemoveExternalSourceFromSecurityGroupResponse, error)

	RemoveExternalSourceFromSecurityGroupWithResponse(ctx context.Context, id string, body RemoveExternalSourceFromSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*RemoveExternalSourceFromSecurityGroupResponse, error)

	// ListSksClusters request
	ListSksClustersWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListSksClustersResponse, error)

	// CreateSksCluster request with any body
	CreateSksClusterWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateSksClusterResponse, error)

	CreateSksClusterWithResponse(ctx context.Context, body CreateSksClusterJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateSksClusterResponse, error)

	// ListSksClusterDeprecatedResources request
	ListSksClusterDeprecatedResourcesWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*ListSksClusterDeprecatedResourcesResponse, error)

	// GenerateSksClusterKubeconfig request with any body
	GenerateSksClusterKubeconfigWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*GenerateSksClusterKubeconfigResponse, error)

	GenerateSksClusterKubeconfigWithResponse(ctx context.Context, id string, body GenerateSksClusterKubeconfigJSONRequestBody, reqEditors ...RequestEditorFn) (*GenerateSksClusterKubeconfigResponse, error)

	// ListSksClusterVersions request
	ListSksClusterVersionsWithResponse(ctx context.Context, params *ListSksClusterVersionsParams, reqEditors ...RequestEditorFn) (*ListSksClusterVersionsResponse, error)

	// DeleteSksCluster request
	DeleteSksClusterWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*DeleteSksClusterResponse, error)

	// GetSksCluster request
	GetSksClusterWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetSksClusterResponse, error)

	// UpdateSksCluster request with any body
	UpdateSksClusterWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateSksClusterResponse, error)

	UpdateSksClusterWithResponse(ctx context.Context, id string, body UpdateSksClusterJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateSksClusterResponse, error)

	// GetSksClusterAuthorityCert request
	GetSksClusterAuthorityCertWithResponse(ctx context.Context, id string, authority GetSksClusterAuthorityCertParamsAuthority, reqEditors ...RequestEditorFn) (*GetSksClusterAuthorityCertResponse, error)

	// CreateSksNodepool request with any body
	CreateSksNodepoolWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateSksNodepoolResponse, error)

	CreateSksNodepoolWithResponse(ctx context.Context, id string, body CreateSksNodepoolJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateSksNodepoolResponse, error)

	// DeleteSksNodepool request
	DeleteSksNodepoolWithResponse(ctx context.Context, id string, sksNodepoolId string, reqEditors ...RequestEditorFn) (*DeleteSksNodepoolResponse, error)

	// GetSksNodepool request
	GetSksNodepoolWithResponse(ctx context.Context, id string, sksNodepoolId string, reqEditors ...RequestEditorFn) (*GetSksNodepoolResponse, error)

	// UpdateSksNodepool request with any body
	UpdateSksNodepoolWithBodyWithResponse(ctx context.Context, id string, sksNodepoolId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateSksNodepoolResponse, error)

	UpdateSksNodepoolWithResponse(ctx context.Context, id string, sksNodepoolId string, body UpdateSksNodepoolJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateSksNodepoolResponse, error)

	// ResetSksNodepoolField request
	ResetSksNodepoolFieldWithResponse(ctx context.Context, id string, sksNodepoolId string, field ResetSksNodepoolFieldParamsField, reqEditors ...RequestEditorFn) (*ResetSksNodepoolFieldResponse, error)

	// EvictSksNodepoolMembers request with any body
	EvictSksNodepoolMembersWithBodyWithResponse(ctx context.Context, id string, sksNodepoolId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*EvictSksNodepoolMembersResponse, error)

	EvictSksNodepoolMembersWithResponse(ctx context.Context, id string, sksNodepoolId string, body EvictSksNodepoolMembersJSONRequestBody, reqEditors ...RequestEditorFn) (*EvictSksNodepoolMembersResponse, error)

	// ScaleSksNodepool request with any body
	ScaleSksNodepoolWithBodyWithResponse(ctx context.Context, id string, sksNodepoolId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ScaleSksNodepoolResponse, error)

	ScaleSksNodepoolWithResponse(ctx context.Context, id string, sksNodepoolId string, body ScaleSksNodepoolJSONRequestBody, reqEditors ...RequestEditorFn) (*ScaleSksNodepoolResponse, error)

	// RotateSksCcmCredentials request
	RotateSksCcmCredentialsWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*RotateSksCcmCredentialsResponse, error)

	// RotateSksOperatorsCa request
	RotateSksOperatorsCaWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*RotateSksOperatorsCaResponse, error)

	// UpgradeSksCluster request with any body
	UpgradeSksClusterWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpgradeSksClusterResponse, error)

	UpgradeSksClusterWithResponse(ctx context.Context, id string, body UpgradeSksClusterJSONRequestBody, reqEditors ...RequestEditorFn) (*UpgradeSksClusterResponse, error)

	// UpgradeSksClusterServiceLevel request
	UpgradeSksClusterServiceLevelWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*UpgradeSksClusterServiceLevelResponse, error)

	// ResetSksClusterField request
	ResetSksClusterFieldWithResponse(ctx context.Context, id string, field ResetSksClusterFieldParamsField, reqEditors ...RequestEditorFn) (*ResetSksClusterFieldResponse, error)

	// ListSnapshots request
	ListSnapshotsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListSnapshotsResponse, error)

	// DeleteSnapshot request
	DeleteSnapshotWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*DeleteSnapshotResponse, error)

	// GetSnapshot request
	GetSnapshotWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetSnapshotResponse, error)

	// ExportSnapshot request
	ExportSnapshotWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*ExportSnapshotResponse, error)

	// PromoteSnapshotToTemplate request with any body
	PromoteSnapshotToTemplateWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PromoteSnapshotToTemplateResponse, error)

	PromoteSnapshotToTemplateWithResponse(ctx context.Context, id string, body PromoteSnapshotToTemplateJSONRequestBody, reqEditors ...RequestEditorFn) (*PromoteSnapshotToTemplateResponse, error)

	// GetSosPresignedUrl request
	GetSosPresignedUrlWithResponse(ctx context.Context, bucket string, params *GetSosPresignedUrlParams, reqEditors ...RequestEditorFn) (*GetSosPresignedUrlResponse, error)

	// ListSshKeys request
	ListSshKeysWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListSshKeysResponse, error)

	// RegisterSshKey request with any body
	RegisterSshKeyWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*RegisterSshKeyResponse, error)

	RegisterSshKeyWithResponse(ctx context.Context, body RegisterSshKeyJSONRequestBody, reqEditors ...RequestEditorFn) (*RegisterSshKeyResponse, error)

	// DeleteSshKey request
	DeleteSshKeyWithResponse(ctx context.Context, name string, reqEditors ...RequestEditorFn) (*DeleteSshKeyResponse, error)

	// GetSshKey request
	GetSshKeyWithResponse(ctx context.Context, name string, reqEditors ...RequestEditorFn) (*GetSshKeyResponse, error)

	// ListTemplates request
	ListTemplatesWithResponse(ctx context.Context, params *ListTemplatesParams, reqEditors ...RequestEditorFn) (*ListTemplatesResponse, error)

	// RegisterTemplate request with any body
	RegisterTemplateWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*RegisterTemplateResponse, error)

	RegisterTemplateWithResponse(ctx context.Context, body RegisterTemplateJSONRequestBody, reqEditors ...RequestEditorFn) (*RegisterTemplateResponse, error)

	// DeleteTemplate request
	DeleteTemplateWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*DeleteTemplateResponse, error)

	// GetTemplate request
	GetTemplateWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetTemplateResponse, error)

	// CopyTemplate request with any body
	CopyTemplateWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CopyTemplateResponse, error)

	CopyTemplateWithResponse(ctx context.Context, id string, body CopyTemplateJSONRequestBody, reqEditors ...RequestEditorFn) (*CopyTemplateResponse, error)

	// UpdateTemplate request with any body
	UpdateTemplateWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateTemplateResponse, error)

	UpdateTemplateWithResponse(ctx context.Context, id string, body UpdateTemplateJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateTemplateResponse, error)

	// ListZones request
	ListZonesWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListZonesResponse, error)
}

type ListAccessKeysResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		AccessKeys *[]AccessKey `json:"access-keys,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListAccessKeysResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListAccessKeysResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateAccessKeyResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *AccessKey
}

// Status returns HTTPResponse.Status
func (r CreateAccessKeyResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateAccessKeyResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListAccessKeyKnownOperationsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		AccessKeyOperations *[]AccessKeyOperation `json:"access-key-operations,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListAccessKeyKnownOperationsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListAccessKeyKnownOperationsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListAccessKeyOperationsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		AccessKeyOperations *[]AccessKeyOperation `json:"access-key-operations,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListAccessKeyOperationsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListAccessKeyOperationsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type RevokeAccessKeyResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r RevokeAccessKeyResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r RevokeAccessKeyResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetAccessKeyResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *AccessKey
}

// Status returns HTTPResponse.Status
func (r GetAccessKeyResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetAccessKeyResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListAntiAffinityGroupsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		AntiAffinityGroups *[]AntiAffinityGroup `json:"anti-affinity-groups,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListAntiAffinityGroupsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListAntiAffinityGroupsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateAntiAffinityGroupResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r CreateAntiAffinityGroupResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateAntiAffinityGroupResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteAntiAffinityGroupResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DeleteAntiAffinityGroupResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteAntiAffinityGroupResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetAntiAffinityGroupResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *AntiAffinityGroup
}

// Status returns HTTPResponse.Status
func (r GetAntiAffinityGroupResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetAntiAffinityGroupResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDbaasCaCertificateResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Certificate *string `json:"certificate,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r GetDbaasCaCertificateResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDbaasCaCertificateResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDbaasServiceKafkaResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DbaasServiceKafka
}

// Status returns HTTPResponse.Status
func (r GetDbaasServiceKafkaResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDbaasServiceKafkaResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateDbaasServiceKafkaResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DbaasServiceKafka
}

// Status returns HTTPResponse.Status
func (r CreateDbaasServiceKafkaResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateDbaasServiceKafkaResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpdateDbaasServiceKafkaResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DbaasServiceKafka
}

// Status returns HTTPResponse.Status
func (r UpdateDbaasServiceKafkaResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpdateDbaasServiceKafkaResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDbaasMigrationStatusResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DbaasMigrationStatus
}

// Status returns HTTPResponse.Status
func (r GetDbaasMigrationStatusResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDbaasMigrationStatusResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDbaasServiceMysqlResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DbaasServiceMysql
}

// Status returns HTTPResponse.Status
func (r GetDbaasServiceMysqlResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDbaasServiceMysqlResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateDbaasServiceMysqlResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DbaasServiceMysql
}

// Status returns HTTPResponse.Status
func (r CreateDbaasServiceMysqlResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateDbaasServiceMysqlResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpdateDbaasServiceMysqlResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DbaasServiceMysql
}

// Status returns HTTPResponse.Status
func (r UpdateDbaasServiceMysqlResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpdateDbaasServiceMysqlResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDbaasServiceOpensearchResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DbaasServiceOpensearch
}

// Status returns HTTPResponse.Status
func (r GetDbaasServiceOpensearchResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDbaasServiceOpensearchResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateDbaasServiceOpensearchResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DbaasServiceOpensearch
}

// Status returns HTTPResponse.Status
func (r CreateDbaasServiceOpensearchResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateDbaasServiceOpensearchResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpdateDbaasServiceOpensearchResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DbaasServiceOpensearch
}

// Status returns HTTPResponse.Status
func (r UpdateDbaasServiceOpensearchResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpdateDbaasServiceOpensearchResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDbaasServicePgResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DbaasServicePg
}

// Status returns HTTPResponse.Status
func (r GetDbaasServicePgResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDbaasServicePgResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateDbaasServicePgResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DbaasServicePg
}

// Status returns HTTPResponse.Status
func (r CreateDbaasServicePgResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateDbaasServicePgResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpdateDbaasServicePgResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DbaasServicePg
}

// Status returns HTTPResponse.Status
func (r UpdateDbaasServicePgResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpdateDbaasServicePgResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDbaasServiceRedisResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DbaasServiceRedis
}

// Status returns HTTPResponse.Status
func (r GetDbaasServiceRedisResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDbaasServiceRedisResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateDbaasServiceRedisResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DbaasServiceRedis
}

// Status returns HTTPResponse.Status
func (r CreateDbaasServiceRedisResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateDbaasServiceRedisResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpdateDbaasServiceRedisResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DbaasServiceRedis
}

// Status returns HTTPResponse.Status
func (r UpdateDbaasServiceRedisResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpdateDbaasServiceRedisResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListDbaasServicesResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		DbaasServices *[]DbaasServiceCommon `json:"dbaas-services,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListDbaasServicesResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListDbaasServicesResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDbaasServiceLogsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DbaasServiceLogs
}

// Status returns HTTPResponse.Status
func (r GetDbaasServiceLogsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDbaasServiceLogsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDbaasServiceMetricsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Metrics *map[string]interface{} `json:"metrics,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r GetDbaasServiceMetricsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDbaasServiceMetricsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListDbaasServiceTypesResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		DbaasServiceTypes *[]DbaasServiceType `json:"dbaas-service-types,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListDbaasServiceTypesResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListDbaasServiceTypesResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDbaasServiceTypeResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DbaasServiceType
}

// Status returns HTTPResponse.Status
func (r GetDbaasServiceTypeResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDbaasServiceTypeResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteDbaasServiceResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DbaasServiceCommon
}

// Status returns HTTPResponse.Status
func (r DeleteDbaasServiceResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteDbaasServiceResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDbaasSettingsKafkaResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Settings *struct {
			// Kafka broker configuration values
			Kafka *struct {
				AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
				Properties           *map[string]interface{} `json:"properties,omitempty"`
				Title                *string                 `json:"title,omitempty"`
				Type                 *string                 `json:"type,omitempty"`
			} `json:"kafka,omitempty"`

			// Kafka Connect configuration values
			KafkaConnect *struct {
				AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
				Properties           *map[string]interface{} `json:"properties,omitempty"`
				Title                *string                 `json:"title,omitempty"`
				Type                 *string                 `json:"type,omitempty"`
			} `json:"kafka-connect,omitempty"`

			// Kafka REST configuration
			KafkaRest *struct {
				AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
				Properties           *map[string]interface{} `json:"properties,omitempty"`
				Title                *string                 `json:"title,omitempty"`
				Type                 *string                 `json:"type,omitempty"`
			} `json:"kafka-rest,omitempty"`

			// Schema Registry configuration
			SchemaRegistry *struct {
				AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
				Properties           *map[string]interface{} `json:"properties,omitempty"`
				Title                *string                 `json:"title,omitempty"`
				Type                 *string                 `json:"type,omitempty"`
			} `json:"schema-registry,omitempty"`
		} `json:"settings,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r GetDbaasSettingsKafkaResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDbaasSettingsKafkaResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDbaasSettingsMysqlResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Settings *struct {
			// mysql.conf configuration values
			Mysql *struct {
				AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
				Properties           *map[string]interface{} `json:"properties,omitempty"`
				Title                *string                 `json:"title,omitempty"`
				Type                 *string                 `json:"type,omitempty"`
			} `json:"mysql,omitempty"`
		} `json:"settings,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r GetDbaasSettingsMysqlResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDbaasSettingsMysqlResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDbaasSettingsOpensearchResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Settings *struct {
			// OpenSearch configuration values
			Opensearch *struct {
				AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
				Properties           *map[string]interface{} `json:"properties,omitempty"`
				Title                *string                 `json:"title,omitempty"`
				Type                 *string                 `json:"type,omitempty"`
			} `json:"opensearch,omitempty"`
		} `json:"settings,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r GetDbaasSettingsOpensearchResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDbaasSettingsOpensearchResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDbaasSettingsPgResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Settings *struct {
			// postgresql.conf configuration values
			Pg *struct {
				AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
				Properties           *map[string]interface{} `json:"properties,omitempty"`
				Title                *string                 `json:"title,omitempty"`
				Type                 *string                 `json:"type,omitempty"`
			} `json:"pg,omitempty"`

			// PGBouncer connection pooling settings
			Pgbouncer *struct {
				AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
				Properties           *map[string]interface{} `json:"properties,omitempty"`
				Title                *string                 `json:"title,omitempty"`
				Type                 *string                 `json:"type,omitempty"`
			} `json:"pgbouncer,omitempty"`

			// PGLookout settings
			Pglookout *struct {
				AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
				Properties           *map[string]interface{} `json:"properties,omitempty"`
				Title                *string                 `json:"title,omitempty"`
				Type                 *string                 `json:"type,omitempty"`
			} `json:"pglookout,omitempty"`

			// TimescaleDB extension configuration values
			Timescaledb *struct {
				AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
				Properties           *map[string]interface{} `json:"properties,omitempty"`
				Title                *string                 `json:"title,omitempty"`
				Type                 *string                 `json:"type,omitempty"`
			} `json:"timescaledb,omitempty"`
		} `json:"settings,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r GetDbaasSettingsPgResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDbaasSettingsPgResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDbaasSettingsRedisResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Settings *struct {
			// Redis configuration values
			Redis *struct {
				AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
				Properties           *map[string]interface{} `json:"properties,omitempty"`
				Title                *string                 `json:"title,omitempty"`
				Type                 *string                 `json:"type,omitempty"`
			} `json:"redis,omitempty"`
		} `json:"settings,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r GetDbaasSettingsRedisResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDbaasSettingsRedisResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListDeployTargetsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		DeployTargets *[]DeployTarget `json:"deploy-targets,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListDeployTargetsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListDeployTargetsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDeployTargetResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DeployTarget
}

// Status returns HTTPResponse.Status
func (r GetDeployTargetResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDeployTargetResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListDnsDomainsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		DnsDomains *[]DnsDomain `json:"dns-domains,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListDnsDomainsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListDnsDomainsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDnsDomainResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DnsDomain
}

// Status returns HTTPResponse.Status
func (r GetDnsDomainResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDnsDomainResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListDnsDomainRecordsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		DnsDomainRecords *[]DnsDomainRecord `json:"dns-domain-records,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListDnsDomainRecordsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListDnsDomainRecordsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetDnsDomainRecordResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DnsDomainRecord
}

// Status returns HTTPResponse.Status
func (r GetDnsDomainRecordResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetDnsDomainRecordResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListElasticIpsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		ElasticIps *[]ElasticIp `json:"elastic-ips,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListElasticIpsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListElasticIpsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateElasticIpResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r CreateElasticIpResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateElasticIpResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteElasticIpResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DeleteElasticIpResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteElasticIpResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetElasticIpResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *ElasticIp
}

// Status returns HTTPResponse.Status
func (r GetElasticIpResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetElasticIpResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpdateElasticIpResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r UpdateElasticIpResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpdateElasticIpResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ResetElasticIpFieldResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r ResetElasticIpFieldResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ResetElasticIpFieldResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type AttachInstanceToElasticIpResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r AttachInstanceToElasticIpResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r AttachInstanceToElasticIpResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DetachInstanceFromElasticIpResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DetachInstanceFromElasticIpResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DetachInstanceFromElasticIpResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListEventsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]Event
}

// Status returns HTTPResponse.Status
func (r ListEventsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListEventsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListInstancesResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Instances *[]Instance `json:"instances,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListInstancesResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListInstancesResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateInstanceResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r CreateInstanceResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateInstanceResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListInstancePoolsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		InstancePools *[]InstancePool `json:"instance-pools,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListInstancePoolsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListInstancePoolsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateInstancePoolResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r CreateInstancePoolResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateInstancePoolResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteInstancePoolResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DeleteInstancePoolResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteInstancePoolResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetInstancePoolResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *InstancePool
}

// Status returns HTTPResponse.Status
func (r GetInstancePoolResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetInstancePoolResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpdateInstancePoolResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r UpdateInstancePoolResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpdateInstancePoolResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ResetInstancePoolFieldResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r ResetInstancePoolFieldResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ResetInstancePoolFieldResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type EvictInstancePoolMembersResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r EvictInstancePoolMembersResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r EvictInstancePoolMembersResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ScaleInstancePoolResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r ScaleInstancePoolResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ScaleInstancePoolResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListInstanceTypesResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		InstanceTypes *[]InstanceType `json:"instance-types,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListInstanceTypesResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListInstanceTypesResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetInstanceTypeResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *InstanceType
}

// Status returns HTTPResponse.Status
func (r GetInstanceTypeResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetInstanceTypeResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteInstanceResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DeleteInstanceResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteInstanceResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetInstanceResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Instance
}

// Status returns HTTPResponse.Status
func (r GetInstanceResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetInstanceResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpdateInstanceResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r UpdateInstanceResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpdateInstanceResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ResetInstanceFieldResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r ResetInstanceFieldResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ResetInstanceFieldResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateSnapshotResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r CreateSnapshotResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateSnapshotResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type RebootInstanceResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r RebootInstanceResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r RebootInstanceResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ResetInstanceResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r ResetInstanceResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ResetInstanceResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ResizeInstanceDiskResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r ResizeInstanceDiskResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ResizeInstanceDiskResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ScaleInstanceResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r ScaleInstanceResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ScaleInstanceResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type StartInstanceResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r StartInstanceResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r StartInstanceResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type StopInstanceResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r StopInstanceResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r StopInstanceResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type RevertInstanceToSnapshotResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r RevertInstanceToSnapshotResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r RevertInstanceToSnapshotResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListLoadBalancersResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		LoadBalancers *[]LoadBalancer `json:"load-balancers,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListLoadBalancersResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListLoadBalancersResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateLoadBalancerResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r CreateLoadBalancerResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateLoadBalancerResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteLoadBalancerResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DeleteLoadBalancerResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteLoadBalancerResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetLoadBalancerResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *LoadBalancer
}

// Status returns HTTPResponse.Status
func (r GetLoadBalancerResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetLoadBalancerResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpdateLoadBalancerResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r UpdateLoadBalancerResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpdateLoadBalancerResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type AddServiceToLoadBalancerResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r AddServiceToLoadBalancerResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r AddServiceToLoadBalancerResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteLoadBalancerServiceResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DeleteLoadBalancerServiceResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteLoadBalancerServiceResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetLoadBalancerServiceResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *LoadBalancerService
}

// Status returns HTTPResponse.Status
func (r GetLoadBalancerServiceResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetLoadBalancerServiceResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpdateLoadBalancerServiceResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r UpdateLoadBalancerServiceResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpdateLoadBalancerServiceResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ResetLoadBalancerServiceFieldResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r ResetLoadBalancerServiceFieldResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ResetLoadBalancerServiceFieldResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ResetLoadBalancerFieldResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r ResetLoadBalancerFieldResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ResetLoadBalancerFieldResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetOperationResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r GetOperationResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetOperationResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListPrivateNetworksResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		PrivateNetworks *[]PrivateNetwork `json:"private-networks,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListPrivateNetworksResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListPrivateNetworksResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreatePrivateNetworkResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r CreatePrivateNetworkResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreatePrivateNetworkResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeletePrivateNetworkResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DeletePrivateNetworkResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeletePrivateNetworkResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetPrivateNetworkResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *PrivateNetwork
}

// Status returns HTTPResponse.Status
func (r GetPrivateNetworkResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetPrivateNetworkResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpdatePrivateNetworkResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r UpdatePrivateNetworkResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpdatePrivateNetworkResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ResetPrivateNetworkFieldResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r ResetPrivateNetworkFieldResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ResetPrivateNetworkFieldResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type AttachInstanceToPrivateNetworkResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r AttachInstanceToPrivateNetworkResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r AttachInstanceToPrivateNetworkResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DetachInstanceFromPrivateNetworkResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DetachInstanceFromPrivateNetworkResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DetachInstanceFromPrivateNetworkResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpdatePrivateNetworkInstanceIpResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r UpdatePrivateNetworkInstanceIpResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpdatePrivateNetworkInstanceIpResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListQuotasResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Quotas *[]Quota `json:"quotas,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListQuotasResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListQuotasResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetQuotaResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Quota
}

// Status returns HTTPResponse.Status
func (r GetQuotaResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetQuotaResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListSecurityGroupsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		SecurityGroups *[]SecurityGroup `json:"security-groups,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListSecurityGroupsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListSecurityGroupsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateSecurityGroupResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r CreateSecurityGroupResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateSecurityGroupResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteSecurityGroupResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DeleteSecurityGroupResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteSecurityGroupResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetSecurityGroupResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *SecurityGroup
}

// Status returns HTTPResponse.Status
func (r GetSecurityGroupResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetSecurityGroupResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type AddRuleToSecurityGroupResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r AddRuleToSecurityGroupResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r AddRuleToSecurityGroupResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteRuleFromSecurityGroupResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DeleteRuleFromSecurityGroupResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteRuleFromSecurityGroupResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type AddExternalSourceToSecurityGroupResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r AddExternalSourceToSecurityGroupResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r AddExternalSourceToSecurityGroupResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type AttachInstanceToSecurityGroupResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r AttachInstanceToSecurityGroupResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r AttachInstanceToSecurityGroupResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DetachInstanceFromSecurityGroupResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DetachInstanceFromSecurityGroupResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DetachInstanceFromSecurityGroupResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type RemoveExternalSourceFromSecurityGroupResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r RemoveExternalSourceFromSecurityGroupResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r RemoveExternalSourceFromSecurityGroupResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListSksClustersResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		SksClusters *[]SksCluster `json:"sks-clusters,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListSksClustersResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListSksClustersResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateSksClusterResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r CreateSksClusterResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateSksClusterResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListSksClusterDeprecatedResourcesResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]SksClusterDeprecatedResource
}

// Status returns HTTPResponse.Status
func (r ListSksClusterDeprecatedResourcesResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListSksClusterDeprecatedResourcesResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GenerateSksClusterKubeconfigResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Kubeconfig *string `json:"kubeconfig,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r GenerateSksClusterKubeconfigResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GenerateSksClusterKubeconfigResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListSksClusterVersionsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		SksClusterVersions *[]string `json:"sks-cluster-versions,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListSksClusterVersionsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListSksClusterVersionsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteSksClusterResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DeleteSksClusterResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteSksClusterResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetSksClusterResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *SksCluster
}

// Status returns HTTPResponse.Status
func (r GetSksClusterResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetSksClusterResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpdateSksClusterResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r UpdateSksClusterResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpdateSksClusterResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetSksClusterAuthorityCertResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Cacert *string `json:"cacert,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r GetSksClusterAuthorityCertResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetSksClusterAuthorityCertResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CreateSksNodepoolResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r CreateSksNodepoolResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CreateSksNodepoolResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteSksNodepoolResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DeleteSksNodepoolResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteSksNodepoolResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetSksNodepoolResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *SksNodepool
}

// Status returns HTTPResponse.Status
func (r GetSksNodepoolResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetSksNodepoolResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpdateSksNodepoolResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r UpdateSksNodepoolResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpdateSksNodepoolResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ResetSksNodepoolFieldResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r ResetSksNodepoolFieldResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ResetSksNodepoolFieldResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type EvictSksNodepoolMembersResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r EvictSksNodepoolMembersResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r EvictSksNodepoolMembersResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ScaleSksNodepoolResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r ScaleSksNodepoolResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ScaleSksNodepoolResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type RotateSksCcmCredentialsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r RotateSksCcmCredentialsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r RotateSksCcmCredentialsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type RotateSksOperatorsCaResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r RotateSksOperatorsCaResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r RotateSksOperatorsCaResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpgradeSksClusterResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r UpgradeSksClusterResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpgradeSksClusterResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpgradeSksClusterServiceLevelResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r UpgradeSksClusterServiceLevelResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpgradeSksClusterServiceLevelResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ResetSksClusterFieldResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r ResetSksClusterFieldResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ResetSksClusterFieldResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListSnapshotsResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Snapshots *[]Snapshot `json:"snapshots,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListSnapshotsResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListSnapshotsResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteSnapshotResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DeleteSnapshotResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteSnapshotResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetSnapshotResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Snapshot
}

// Status returns HTTPResponse.Status
func (r GetSnapshotResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetSnapshotResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ExportSnapshotResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r ExportSnapshotResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ExportSnapshotResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type PromoteSnapshotToTemplateResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r PromoteSnapshotToTemplateResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PromoteSnapshotToTemplateResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetSosPresignedUrlResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Url *string `json:"url,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r GetSosPresignedUrlResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetSosPresignedUrlResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListSshKeysResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		SshKeys *[]SshKey `json:"ssh-keys,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListSshKeysResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListSshKeysResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type RegisterSshKeyResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r RegisterSshKeyResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r RegisterSshKeyResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteSshKeyResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DeleteSshKeyResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteSshKeyResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetSshKeyResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *SshKey
}

// Status returns HTTPResponse.Status
func (r GetSshKeyResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetSshKeyResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListTemplatesResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Templates *[]Template `json:"templates,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListTemplatesResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListTemplatesResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type RegisterTemplateResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r RegisterTemplateResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r RegisterTemplateResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DeleteTemplateResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r DeleteTemplateResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DeleteTemplateResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetTemplateResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Template
}

// Status returns HTTPResponse.Status
func (r GetTemplateResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetTemplateResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type CopyTemplateResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r CopyTemplateResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r CopyTemplateResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type UpdateTemplateResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Operation
}

// Status returns HTTPResponse.Status
func (r UpdateTemplateResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r UpdateTemplateResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ListZonesResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *struct {
		Zones *[]Zone `json:"zones,omitempty"`
	}
}

// Status returns HTTPResponse.Status
func (r ListZonesResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ListZonesResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// ListAccessKeysWithResponse request returning *ListAccessKeysResponse
func (c *ClientWithResponses) ListAccessKeysWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListAccessKeysResponse, error) {
	rsp, err := c.ListAccessKeys(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListAccessKeysResponse(rsp)
}

// CreateAccessKeyWithBodyWithResponse request with arbitrary body returning *CreateAccessKeyResponse
func (c *ClientWithResponses) CreateAccessKeyWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateAccessKeyResponse, error) {
	rsp, err := c.CreateAccessKeyWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateAccessKeyResponse(rsp)
}

func (c *ClientWithResponses) CreateAccessKeyWithResponse(ctx context.Context, body CreateAccessKeyJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateAccessKeyResponse, error) {
	rsp, err := c.CreateAccessKey(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateAccessKeyResponse(rsp)
}

// ListAccessKeyKnownOperationsWithResponse request returning *ListAccessKeyKnownOperationsResponse
func (c *ClientWithResponses) ListAccessKeyKnownOperationsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListAccessKeyKnownOperationsResponse, error) {
	rsp, err := c.ListAccessKeyKnownOperations(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListAccessKeyKnownOperationsResponse(rsp)
}

// ListAccessKeyOperationsWithResponse request returning *ListAccessKeyOperationsResponse
func (c *ClientWithResponses) ListAccessKeyOperationsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListAccessKeyOperationsResponse, error) {
	rsp, err := c.ListAccessKeyOperations(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListAccessKeyOperationsResponse(rsp)
}

// RevokeAccessKeyWithResponse request returning *RevokeAccessKeyResponse
func (c *ClientWithResponses) RevokeAccessKeyWithResponse(ctx context.Context, key string, reqEditors ...RequestEditorFn) (*RevokeAccessKeyResponse, error) {
	rsp, err := c.RevokeAccessKey(ctx, key, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseRevokeAccessKeyResponse(rsp)
}

// GetAccessKeyWithResponse request returning *GetAccessKeyResponse
func (c *ClientWithResponses) GetAccessKeyWithResponse(ctx context.Context, key string, reqEditors ...RequestEditorFn) (*GetAccessKeyResponse, error) {
	rsp, err := c.GetAccessKey(ctx, key, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetAccessKeyResponse(rsp)
}

// ListAntiAffinityGroupsWithResponse request returning *ListAntiAffinityGroupsResponse
func (c *ClientWithResponses) ListAntiAffinityGroupsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListAntiAffinityGroupsResponse, error) {
	rsp, err := c.ListAntiAffinityGroups(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListAntiAffinityGroupsResponse(rsp)
}

// CreateAntiAffinityGroupWithBodyWithResponse request with arbitrary body returning *CreateAntiAffinityGroupResponse
func (c *ClientWithResponses) CreateAntiAffinityGroupWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateAntiAffinityGroupResponse, error) {
	rsp, err := c.CreateAntiAffinityGroupWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateAntiAffinityGroupResponse(rsp)
}

func (c *ClientWithResponses) CreateAntiAffinityGroupWithResponse(ctx context.Context, body CreateAntiAffinityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateAntiAffinityGroupResponse, error) {
	rsp, err := c.CreateAntiAffinityGroup(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateAntiAffinityGroupResponse(rsp)
}

// DeleteAntiAffinityGroupWithResponse request returning *DeleteAntiAffinityGroupResponse
func (c *ClientWithResponses) DeleteAntiAffinityGroupWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*DeleteAntiAffinityGroupResponse, error) {
	rsp, err := c.DeleteAntiAffinityGroup(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteAntiAffinityGroupResponse(rsp)
}

// GetAntiAffinityGroupWithResponse request returning *GetAntiAffinityGroupResponse
func (c *ClientWithResponses) GetAntiAffinityGroupWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetAntiAffinityGroupResponse, error) {
	rsp, err := c.GetAntiAffinityGroup(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetAntiAffinityGroupResponse(rsp)
}

// GetDbaasCaCertificateWithResponse request returning *GetDbaasCaCertificateResponse
func (c *ClientWithResponses) GetDbaasCaCertificateWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetDbaasCaCertificateResponse, error) {
	rsp, err := c.GetDbaasCaCertificate(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetDbaasCaCertificateResponse(rsp)
}

// GetDbaasServiceKafkaWithResponse request returning *GetDbaasServiceKafkaResponse
func (c *ClientWithResponses) GetDbaasServiceKafkaWithResponse(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*GetDbaasServiceKafkaResponse, error) {
	rsp, err := c.GetDbaasServiceKafka(ctx, name, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetDbaasServiceKafkaResponse(rsp)
}

// CreateDbaasServiceKafkaWithBodyWithResponse request with arbitrary body returning *CreateDbaasServiceKafkaResponse
func (c *ClientWithResponses) CreateDbaasServiceKafkaWithBodyWithResponse(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateDbaasServiceKafkaResponse, error) {
	rsp, err := c.CreateDbaasServiceKafkaWithBody(ctx, name, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateDbaasServiceKafkaResponse(rsp)
}

func (c *ClientWithResponses) CreateDbaasServiceKafkaWithResponse(ctx context.Context, name DbaasServiceName, body CreateDbaasServiceKafkaJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateDbaasServiceKafkaResponse, error) {
	rsp, err := c.CreateDbaasServiceKafka(ctx, name, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateDbaasServiceKafkaResponse(rsp)
}

// UpdateDbaasServiceKafkaWithBodyWithResponse request with arbitrary body returning *UpdateDbaasServiceKafkaResponse
func (c *ClientWithResponses) UpdateDbaasServiceKafkaWithBodyWithResponse(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateDbaasServiceKafkaResponse, error) {
	rsp, err := c.UpdateDbaasServiceKafkaWithBody(ctx, name, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateDbaasServiceKafkaResponse(rsp)
}

func (c *ClientWithResponses) UpdateDbaasServiceKafkaWithResponse(ctx context.Context, name DbaasServiceName, body UpdateDbaasServiceKafkaJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateDbaasServiceKafkaResponse, error) {
	rsp, err := c.UpdateDbaasServiceKafka(ctx, name, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateDbaasServiceKafkaResponse(rsp)
}

// GetDbaasMigrationStatusWithResponse request returning *GetDbaasMigrationStatusResponse
func (c *ClientWithResponses) GetDbaasMigrationStatusWithResponse(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*GetDbaasMigrationStatusResponse, error) {
	rsp, err := c.GetDbaasMigrationStatus(ctx, name, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetDbaasMigrationStatusResponse(rsp)
}

// GetDbaasServiceMysqlWithResponse request returning *GetDbaasServiceMysqlResponse
func (c *ClientWithResponses) GetDbaasServiceMysqlWithResponse(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*GetDbaasServiceMysqlResponse, error) {
	rsp, err := c.GetDbaasServiceMysql(ctx, name, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetDbaasServiceMysqlResponse(rsp)
}

// CreateDbaasServiceMysqlWithBodyWithResponse request with arbitrary body returning *CreateDbaasServiceMysqlResponse
func (c *ClientWithResponses) CreateDbaasServiceMysqlWithBodyWithResponse(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateDbaasServiceMysqlResponse, error) {
	rsp, err := c.CreateDbaasServiceMysqlWithBody(ctx, name, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateDbaasServiceMysqlResponse(rsp)
}

func (c *ClientWithResponses) CreateDbaasServiceMysqlWithResponse(ctx context.Context, name DbaasServiceName, body CreateDbaasServiceMysqlJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateDbaasServiceMysqlResponse, error) {
	rsp, err := c.CreateDbaasServiceMysql(ctx, name, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateDbaasServiceMysqlResponse(rsp)
}

// UpdateDbaasServiceMysqlWithBodyWithResponse request with arbitrary body returning *UpdateDbaasServiceMysqlResponse
func (c *ClientWithResponses) UpdateDbaasServiceMysqlWithBodyWithResponse(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateDbaasServiceMysqlResponse, error) {
	rsp, err := c.UpdateDbaasServiceMysqlWithBody(ctx, name, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateDbaasServiceMysqlResponse(rsp)
}

func (c *ClientWithResponses) UpdateDbaasServiceMysqlWithResponse(ctx context.Context, name DbaasServiceName, body UpdateDbaasServiceMysqlJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateDbaasServiceMysqlResponse, error) {
	rsp, err := c.UpdateDbaasServiceMysql(ctx, name, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateDbaasServiceMysqlResponse(rsp)
}

// GetDbaasServiceOpensearchWithResponse request returning *GetDbaasServiceOpensearchResponse
func (c *ClientWithResponses) GetDbaasServiceOpensearchWithResponse(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*GetDbaasServiceOpensearchResponse, error) {
	rsp, err := c.GetDbaasServiceOpensearch(ctx, name, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetDbaasServiceOpensearchResponse(rsp)
}

// CreateDbaasServiceOpensearchWithBodyWithResponse request with arbitrary body returning *CreateDbaasServiceOpensearchResponse
func (c *ClientWithResponses) CreateDbaasServiceOpensearchWithBodyWithResponse(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateDbaasServiceOpensearchResponse, error) {
	rsp, err := c.CreateDbaasServiceOpensearchWithBody(ctx, name, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateDbaasServiceOpensearchResponse(rsp)
}

func (c *ClientWithResponses) CreateDbaasServiceOpensearchWithResponse(ctx context.Context, name DbaasServiceName, body CreateDbaasServiceOpensearchJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateDbaasServiceOpensearchResponse, error) {
	rsp, err := c.CreateDbaasServiceOpensearch(ctx, name, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateDbaasServiceOpensearchResponse(rsp)
}

// UpdateDbaasServiceOpensearchWithBodyWithResponse request with arbitrary body returning *UpdateDbaasServiceOpensearchResponse
func (c *ClientWithResponses) UpdateDbaasServiceOpensearchWithBodyWithResponse(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateDbaasServiceOpensearchResponse, error) {
	rsp, err := c.UpdateDbaasServiceOpensearchWithBody(ctx, name, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateDbaasServiceOpensearchResponse(rsp)
}

func (c *ClientWithResponses) UpdateDbaasServiceOpensearchWithResponse(ctx context.Context, name DbaasServiceName, body UpdateDbaasServiceOpensearchJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateDbaasServiceOpensearchResponse, error) {
	rsp, err := c.UpdateDbaasServiceOpensearch(ctx, name, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateDbaasServiceOpensearchResponse(rsp)
}

// GetDbaasServicePgWithResponse request returning *GetDbaasServicePgResponse
func (c *ClientWithResponses) GetDbaasServicePgWithResponse(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*GetDbaasServicePgResponse, error) {
	rsp, err := c.GetDbaasServicePg(ctx, name, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetDbaasServicePgResponse(rsp)
}

// CreateDbaasServicePgWithBodyWithResponse request with arbitrary body returning *CreateDbaasServicePgResponse
func (c *ClientWithResponses) CreateDbaasServicePgWithBodyWithResponse(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateDbaasServicePgResponse, error) {
	rsp, err := c.CreateDbaasServicePgWithBody(ctx, name, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateDbaasServicePgResponse(rsp)
}

func (c *ClientWithResponses) CreateDbaasServicePgWithResponse(ctx context.Context, name DbaasServiceName, body CreateDbaasServicePgJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateDbaasServicePgResponse, error) {
	rsp, err := c.CreateDbaasServicePg(ctx, name, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateDbaasServicePgResponse(rsp)
}

// UpdateDbaasServicePgWithBodyWithResponse request with arbitrary body returning *UpdateDbaasServicePgResponse
func (c *ClientWithResponses) UpdateDbaasServicePgWithBodyWithResponse(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateDbaasServicePgResponse, error) {
	rsp, err := c.UpdateDbaasServicePgWithBody(ctx, name, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateDbaasServicePgResponse(rsp)
}

func (c *ClientWithResponses) UpdateDbaasServicePgWithResponse(ctx context.Context, name DbaasServiceName, body UpdateDbaasServicePgJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateDbaasServicePgResponse, error) {
	rsp, err := c.UpdateDbaasServicePg(ctx, name, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateDbaasServicePgResponse(rsp)
}

// GetDbaasServiceRedisWithResponse request returning *GetDbaasServiceRedisResponse
func (c *ClientWithResponses) GetDbaasServiceRedisWithResponse(ctx context.Context, name DbaasServiceName, reqEditors ...RequestEditorFn) (*GetDbaasServiceRedisResponse, error) {
	rsp, err := c.GetDbaasServiceRedis(ctx, name, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetDbaasServiceRedisResponse(rsp)
}

// CreateDbaasServiceRedisWithBodyWithResponse request with arbitrary body returning *CreateDbaasServiceRedisResponse
func (c *ClientWithResponses) CreateDbaasServiceRedisWithBodyWithResponse(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateDbaasServiceRedisResponse, error) {
	rsp, err := c.CreateDbaasServiceRedisWithBody(ctx, name, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateDbaasServiceRedisResponse(rsp)
}

func (c *ClientWithResponses) CreateDbaasServiceRedisWithResponse(ctx context.Context, name DbaasServiceName, body CreateDbaasServiceRedisJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateDbaasServiceRedisResponse, error) {
	rsp, err := c.CreateDbaasServiceRedis(ctx, name, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateDbaasServiceRedisResponse(rsp)
}

// UpdateDbaasServiceRedisWithBodyWithResponse request with arbitrary body returning *UpdateDbaasServiceRedisResponse
func (c *ClientWithResponses) UpdateDbaasServiceRedisWithBodyWithResponse(ctx context.Context, name DbaasServiceName, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateDbaasServiceRedisResponse, error) {
	rsp, err := c.UpdateDbaasServiceRedisWithBody(ctx, name, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateDbaasServiceRedisResponse(rsp)
}

func (c *ClientWithResponses) UpdateDbaasServiceRedisWithResponse(ctx context.Context, name DbaasServiceName, body UpdateDbaasServiceRedisJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateDbaasServiceRedisResponse, error) {
	rsp, err := c.UpdateDbaasServiceRedis(ctx, name, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateDbaasServiceRedisResponse(rsp)
}

// ListDbaasServicesWithResponse request returning *ListDbaasServicesResponse
func (c *ClientWithResponses) ListDbaasServicesWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListDbaasServicesResponse, error) {
	rsp, err := c.ListDbaasServices(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListDbaasServicesResponse(rsp)
}

// GetDbaasServiceLogsWithBodyWithResponse request with arbitrary body returning *GetDbaasServiceLogsResponse
func (c *ClientWithResponses) GetDbaasServiceLogsWithBodyWithResponse(ctx context.Context, serviceName string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*GetDbaasServiceLogsResponse, error) {
	rsp, err := c.GetDbaasServiceLogsWithBody(ctx, serviceName, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetDbaasServiceLogsResponse(rsp)
}

func (c *ClientWithResponses) GetDbaasServiceLogsWithResponse(ctx context.Context, serviceName string, body GetDbaasServiceLogsJSONRequestBody, reqEditors ...RequestEditorFn) (*GetDbaasServiceLogsResponse, error) {
	rsp, err := c.GetDbaasServiceLogs(ctx, serviceName, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetDbaasServiceLogsResponse(rsp)
}

// GetDbaasServiceMetricsWithBodyWithResponse request with arbitrary body returning *GetDbaasServiceMetricsResponse
func (c *ClientWithResponses) GetDbaasServiceMetricsWithBodyWithResponse(ctx context.Context, serviceName string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*GetDbaasServiceMetricsResponse, error) {
	rsp, err := c.GetDbaasServiceMetricsWithBody(ctx, serviceName, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetDbaasServiceMetricsResponse(rsp)
}

func (c *ClientWithResponses) GetDbaasServiceMetricsWithResponse(ctx context.Context, serviceName string, body GetDbaasServiceMetricsJSONRequestBody, reqEditors ...RequestEditorFn) (*GetDbaasServiceMetricsResponse, error) {
	rsp, err := c.GetDbaasServiceMetrics(ctx, serviceName, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetDbaasServiceMetricsResponse(rsp)
}

// ListDbaasServiceTypesWithResponse request returning *ListDbaasServiceTypesResponse
func (c *ClientWithResponses) ListDbaasServiceTypesWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListDbaasServiceTypesResponse, error) {
	rsp, err := c.ListDbaasServiceTypes(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListDbaasServiceTypesResponse(rsp)
}

// GetDbaasServiceTypeWithResponse request returning *GetDbaasServiceTypeResponse
func (c *ClientWithResponses) GetDbaasServiceTypeWithResponse(ctx context.Context, serviceTypeName string, reqEditors ...RequestEditorFn) (*GetDbaasServiceTypeResponse, error) {
	rsp, err := c.GetDbaasServiceType(ctx, serviceTypeName, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetDbaasServiceTypeResponse(rsp)
}

// DeleteDbaasServiceWithResponse request returning *DeleteDbaasServiceResponse
func (c *ClientWithResponses) DeleteDbaasServiceWithResponse(ctx context.Context, name string, reqEditors ...RequestEditorFn) (*DeleteDbaasServiceResponse, error) {
	rsp, err := c.DeleteDbaasService(ctx, name, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteDbaasServiceResponse(rsp)
}

// GetDbaasSettingsKafkaWithResponse request returning *GetDbaasSettingsKafkaResponse
func (c *ClientWithResponses) GetDbaasSettingsKafkaWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetDbaasSettingsKafkaResponse, error) {
	rsp, err := c.GetDbaasSettingsKafka(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetDbaasSettingsKafkaResponse(rsp)
}

// GetDbaasSettingsMysqlWithResponse request returning *GetDbaasSettingsMysqlResponse
func (c *ClientWithResponses) GetDbaasSettingsMysqlWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetDbaasSettingsMysqlResponse, error) {
	rsp, err := c.GetDbaasSettingsMysql(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetDbaasSettingsMysqlResponse(rsp)
}

// GetDbaasSettingsOpensearchWithResponse request returning *GetDbaasSettingsOpensearchResponse
func (c *ClientWithResponses) GetDbaasSettingsOpensearchWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetDbaasSettingsOpensearchResponse, error) {
	rsp, err := c.GetDbaasSettingsOpensearch(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetDbaasSettingsOpensearchResponse(rsp)
}

// GetDbaasSettingsPgWithResponse request returning *GetDbaasSettingsPgResponse
func (c *ClientWithResponses) GetDbaasSettingsPgWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetDbaasSettingsPgResponse, error) {
	rsp, err := c.GetDbaasSettingsPg(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetDbaasSettingsPgResponse(rsp)
}

// GetDbaasSettingsRedisWithResponse request returning *GetDbaasSettingsRedisResponse
func (c *ClientWithResponses) GetDbaasSettingsRedisWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetDbaasSettingsRedisResponse, error) {
	rsp, err := c.GetDbaasSettingsRedis(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetDbaasSettingsRedisResponse(rsp)
}

// ListDeployTargetsWithResponse request returning *ListDeployTargetsResponse
func (c *ClientWithResponses) ListDeployTargetsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListDeployTargetsResponse, error) {
	rsp, err := c.ListDeployTargets(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListDeployTargetsResponse(rsp)
}

// GetDeployTargetWithResponse request returning *GetDeployTargetResponse
func (c *ClientWithResponses) GetDeployTargetWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetDeployTargetResponse, error) {
	rsp, err := c.GetDeployTarget(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetDeployTargetResponse(rsp)
}

// ListDnsDomainsWithResponse request returning *ListDnsDomainsResponse
func (c *ClientWithResponses) ListDnsDomainsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListDnsDomainsResponse, error) {
	rsp, err := c.ListDnsDomains(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListDnsDomainsResponse(rsp)
}

// GetDnsDomainWithResponse request returning *GetDnsDomainResponse
func (c *ClientWithResponses) GetDnsDomainWithResponse(ctx context.Context, id int64, reqEditors ...RequestEditorFn) (*GetDnsDomainResponse, error) {
	rsp, err := c.GetDnsDomain(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetDnsDomainResponse(rsp)
}

// ListDnsDomainRecordsWithResponse request returning *ListDnsDomainRecordsResponse
func (c *ClientWithResponses) ListDnsDomainRecordsWithResponse(ctx context.Context, id int64, reqEditors ...RequestEditorFn) (*ListDnsDomainRecordsResponse, error) {
	rsp, err := c.ListDnsDomainRecords(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListDnsDomainRecordsResponse(rsp)
}

// GetDnsDomainRecordWithResponse request returning *GetDnsDomainRecordResponse
func (c *ClientWithResponses) GetDnsDomainRecordWithResponse(ctx context.Context, id int64, recordId int64, reqEditors ...RequestEditorFn) (*GetDnsDomainRecordResponse, error) {
	rsp, err := c.GetDnsDomainRecord(ctx, id, recordId, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetDnsDomainRecordResponse(rsp)
}

// ListElasticIpsWithResponse request returning *ListElasticIpsResponse
func (c *ClientWithResponses) ListElasticIpsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListElasticIpsResponse, error) {
	rsp, err := c.ListElasticIps(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListElasticIpsResponse(rsp)
}

// CreateElasticIpWithBodyWithResponse request with arbitrary body returning *CreateElasticIpResponse
func (c *ClientWithResponses) CreateElasticIpWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateElasticIpResponse, error) {
	rsp, err := c.CreateElasticIpWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateElasticIpResponse(rsp)
}

func (c *ClientWithResponses) CreateElasticIpWithResponse(ctx context.Context, body CreateElasticIpJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateElasticIpResponse, error) {
	rsp, err := c.CreateElasticIp(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateElasticIpResponse(rsp)
}

// DeleteElasticIpWithResponse request returning *DeleteElasticIpResponse
func (c *ClientWithResponses) DeleteElasticIpWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*DeleteElasticIpResponse, error) {
	rsp, err := c.DeleteElasticIp(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteElasticIpResponse(rsp)
}

// GetElasticIpWithResponse request returning *GetElasticIpResponse
func (c *ClientWithResponses) GetElasticIpWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetElasticIpResponse, error) {
	rsp, err := c.GetElasticIp(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetElasticIpResponse(rsp)
}

// UpdateElasticIpWithBodyWithResponse request with arbitrary body returning *UpdateElasticIpResponse
func (c *ClientWithResponses) UpdateElasticIpWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateElasticIpResponse, error) {
	rsp, err := c.UpdateElasticIpWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateElasticIpResponse(rsp)
}

func (c *ClientWithResponses) UpdateElasticIpWithResponse(ctx context.Context, id string, body UpdateElasticIpJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateElasticIpResponse, error) {
	rsp, err := c.UpdateElasticIp(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateElasticIpResponse(rsp)
}

// ResetElasticIpFieldWithResponse request returning *ResetElasticIpFieldResponse
func (c *ClientWithResponses) ResetElasticIpFieldWithResponse(ctx context.Context, id string, field ResetElasticIpFieldParamsField, reqEditors ...RequestEditorFn) (*ResetElasticIpFieldResponse, error) {
	rsp, err := c.ResetElasticIpField(ctx, id, field, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseResetElasticIpFieldResponse(rsp)
}

// AttachInstanceToElasticIpWithBodyWithResponse request with arbitrary body returning *AttachInstanceToElasticIpResponse
func (c *ClientWithResponses) AttachInstanceToElasticIpWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*AttachInstanceToElasticIpResponse, error) {
	rsp, err := c.AttachInstanceToElasticIpWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseAttachInstanceToElasticIpResponse(rsp)
}

func (c *ClientWithResponses) AttachInstanceToElasticIpWithResponse(ctx context.Context, id string, body AttachInstanceToElasticIpJSONRequestBody, reqEditors ...RequestEditorFn) (*AttachInstanceToElasticIpResponse, error) {
	rsp, err := c.AttachInstanceToElasticIp(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseAttachInstanceToElasticIpResponse(rsp)
}

// DetachInstanceFromElasticIpWithBodyWithResponse request with arbitrary body returning *DetachInstanceFromElasticIpResponse
func (c *ClientWithResponses) DetachInstanceFromElasticIpWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*DetachInstanceFromElasticIpResponse, error) {
	rsp, err := c.DetachInstanceFromElasticIpWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDetachInstanceFromElasticIpResponse(rsp)
}

func (c *ClientWithResponses) DetachInstanceFromElasticIpWithResponse(ctx context.Context, id string, body DetachInstanceFromElasticIpJSONRequestBody, reqEditors ...RequestEditorFn) (*DetachInstanceFromElasticIpResponse, error) {
	rsp, err := c.DetachInstanceFromElasticIp(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDetachInstanceFromElasticIpResponse(rsp)
}

// ListEventsWithResponse request returning *ListEventsResponse
func (c *ClientWithResponses) ListEventsWithResponse(ctx context.Context, params *ListEventsParams, reqEditors ...RequestEditorFn) (*ListEventsResponse, error) {
	rsp, err := c.ListEvents(ctx, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListEventsResponse(rsp)
}

// ListInstancesWithResponse request returning *ListInstancesResponse
func (c *ClientWithResponses) ListInstancesWithResponse(ctx context.Context, params *ListInstancesParams, reqEditors ...RequestEditorFn) (*ListInstancesResponse, error) {
	rsp, err := c.ListInstances(ctx, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListInstancesResponse(rsp)
}

// CreateInstanceWithBodyWithResponse request with arbitrary body returning *CreateInstanceResponse
func (c *ClientWithResponses) CreateInstanceWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateInstanceResponse, error) {
	rsp, err := c.CreateInstanceWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateInstanceResponse(rsp)
}

func (c *ClientWithResponses) CreateInstanceWithResponse(ctx context.Context, body CreateInstanceJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateInstanceResponse, error) {
	rsp, err := c.CreateInstance(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateInstanceResponse(rsp)
}

// ListInstancePoolsWithResponse request returning *ListInstancePoolsResponse
func (c *ClientWithResponses) ListInstancePoolsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListInstancePoolsResponse, error) {
	rsp, err := c.ListInstancePools(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListInstancePoolsResponse(rsp)
}

// CreateInstancePoolWithBodyWithResponse request with arbitrary body returning *CreateInstancePoolResponse
func (c *ClientWithResponses) CreateInstancePoolWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateInstancePoolResponse, error) {
	rsp, err := c.CreateInstancePoolWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateInstancePoolResponse(rsp)
}

func (c *ClientWithResponses) CreateInstancePoolWithResponse(ctx context.Context, body CreateInstancePoolJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateInstancePoolResponse, error) {
	rsp, err := c.CreateInstancePool(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateInstancePoolResponse(rsp)
}

// DeleteInstancePoolWithResponse request returning *DeleteInstancePoolResponse
func (c *ClientWithResponses) DeleteInstancePoolWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*DeleteInstancePoolResponse, error) {
	rsp, err := c.DeleteInstancePool(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteInstancePoolResponse(rsp)
}

// GetInstancePoolWithResponse request returning *GetInstancePoolResponse
func (c *ClientWithResponses) GetInstancePoolWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetInstancePoolResponse, error) {
	rsp, err := c.GetInstancePool(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetInstancePoolResponse(rsp)
}

// UpdateInstancePoolWithBodyWithResponse request with arbitrary body returning *UpdateInstancePoolResponse
func (c *ClientWithResponses) UpdateInstancePoolWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateInstancePoolResponse, error) {
	rsp, err := c.UpdateInstancePoolWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateInstancePoolResponse(rsp)
}

func (c *ClientWithResponses) UpdateInstancePoolWithResponse(ctx context.Context, id string, body UpdateInstancePoolJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateInstancePoolResponse, error) {
	rsp, err := c.UpdateInstancePool(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateInstancePoolResponse(rsp)
}

// ResetInstancePoolFieldWithResponse request returning *ResetInstancePoolFieldResponse
func (c *ClientWithResponses) ResetInstancePoolFieldWithResponse(ctx context.Context, id string, field ResetInstancePoolFieldParamsField, reqEditors ...RequestEditorFn) (*ResetInstancePoolFieldResponse, error) {
	rsp, err := c.ResetInstancePoolField(ctx, id, field, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseResetInstancePoolFieldResponse(rsp)
}

// EvictInstancePoolMembersWithBodyWithResponse request with arbitrary body returning *EvictInstancePoolMembersResponse
func (c *ClientWithResponses) EvictInstancePoolMembersWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*EvictInstancePoolMembersResponse, error) {
	rsp, err := c.EvictInstancePoolMembersWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseEvictInstancePoolMembersResponse(rsp)
}

func (c *ClientWithResponses) EvictInstancePoolMembersWithResponse(ctx context.Context, id string, body EvictInstancePoolMembersJSONRequestBody, reqEditors ...RequestEditorFn) (*EvictInstancePoolMembersResponse, error) {
	rsp, err := c.EvictInstancePoolMembers(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseEvictInstancePoolMembersResponse(rsp)
}

// ScaleInstancePoolWithBodyWithResponse request with arbitrary body returning *ScaleInstancePoolResponse
func (c *ClientWithResponses) ScaleInstancePoolWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ScaleInstancePoolResponse, error) {
	rsp, err := c.ScaleInstancePoolWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseScaleInstancePoolResponse(rsp)
}

func (c *ClientWithResponses) ScaleInstancePoolWithResponse(ctx context.Context, id string, body ScaleInstancePoolJSONRequestBody, reqEditors ...RequestEditorFn) (*ScaleInstancePoolResponse, error) {
	rsp, err := c.ScaleInstancePool(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseScaleInstancePoolResponse(rsp)
}

// ListInstanceTypesWithResponse request returning *ListInstanceTypesResponse
func (c *ClientWithResponses) ListInstanceTypesWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListInstanceTypesResponse, error) {
	rsp, err := c.ListInstanceTypes(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListInstanceTypesResponse(rsp)
}

// GetInstanceTypeWithResponse request returning *GetInstanceTypeResponse
func (c *ClientWithResponses) GetInstanceTypeWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetInstanceTypeResponse, error) {
	rsp, err := c.GetInstanceType(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetInstanceTypeResponse(rsp)
}

// DeleteInstanceWithResponse request returning *DeleteInstanceResponse
func (c *ClientWithResponses) DeleteInstanceWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*DeleteInstanceResponse, error) {
	rsp, err := c.DeleteInstance(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteInstanceResponse(rsp)
}

// GetInstanceWithResponse request returning *GetInstanceResponse
func (c *ClientWithResponses) GetInstanceWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetInstanceResponse, error) {
	rsp, err := c.GetInstance(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetInstanceResponse(rsp)
}

// UpdateInstanceWithBodyWithResponse request with arbitrary body returning *UpdateInstanceResponse
func (c *ClientWithResponses) UpdateInstanceWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateInstanceResponse, error) {
	rsp, err := c.UpdateInstanceWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateInstanceResponse(rsp)
}

func (c *ClientWithResponses) UpdateInstanceWithResponse(ctx context.Context, id string, body UpdateInstanceJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateInstanceResponse, error) {
	rsp, err := c.UpdateInstance(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateInstanceResponse(rsp)
}

// ResetInstanceFieldWithResponse request returning *ResetInstanceFieldResponse
func (c *ClientWithResponses) ResetInstanceFieldWithResponse(ctx context.Context, id string, field ResetInstanceFieldParamsField, reqEditors ...RequestEditorFn) (*ResetInstanceFieldResponse, error) {
	rsp, err := c.ResetInstanceField(ctx, id, field, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseResetInstanceFieldResponse(rsp)
}

// CreateSnapshotWithResponse request returning *CreateSnapshotResponse
func (c *ClientWithResponses) CreateSnapshotWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*CreateSnapshotResponse, error) {
	rsp, err := c.CreateSnapshot(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateSnapshotResponse(rsp)
}

// RebootInstanceWithResponse request returning *RebootInstanceResponse
func (c *ClientWithResponses) RebootInstanceWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*RebootInstanceResponse, error) {
	rsp, err := c.RebootInstance(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseRebootInstanceResponse(rsp)
}

// ResetInstanceWithBodyWithResponse request with arbitrary body returning *ResetInstanceResponse
func (c *ClientWithResponses) ResetInstanceWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ResetInstanceResponse, error) {
	rsp, err := c.ResetInstanceWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseResetInstanceResponse(rsp)
}

func (c *ClientWithResponses) ResetInstanceWithResponse(ctx context.Context, id string, body ResetInstanceJSONRequestBody, reqEditors ...RequestEditorFn) (*ResetInstanceResponse, error) {
	rsp, err := c.ResetInstance(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseResetInstanceResponse(rsp)
}

// ResizeInstanceDiskWithBodyWithResponse request with arbitrary body returning *ResizeInstanceDiskResponse
func (c *ClientWithResponses) ResizeInstanceDiskWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ResizeInstanceDiskResponse, error) {
	rsp, err := c.ResizeInstanceDiskWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseResizeInstanceDiskResponse(rsp)
}

func (c *ClientWithResponses) ResizeInstanceDiskWithResponse(ctx context.Context, id string, body ResizeInstanceDiskJSONRequestBody, reqEditors ...RequestEditorFn) (*ResizeInstanceDiskResponse, error) {
	rsp, err := c.ResizeInstanceDisk(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseResizeInstanceDiskResponse(rsp)
}

// ScaleInstanceWithBodyWithResponse request with arbitrary body returning *ScaleInstanceResponse
func (c *ClientWithResponses) ScaleInstanceWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ScaleInstanceResponse, error) {
	rsp, err := c.ScaleInstanceWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseScaleInstanceResponse(rsp)
}

func (c *ClientWithResponses) ScaleInstanceWithResponse(ctx context.Context, id string, body ScaleInstanceJSONRequestBody, reqEditors ...RequestEditorFn) (*ScaleInstanceResponse, error) {
	rsp, err := c.ScaleInstance(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseScaleInstanceResponse(rsp)
}

// StartInstanceWithBodyWithResponse request with arbitrary body returning *StartInstanceResponse
func (c *ClientWithResponses) StartInstanceWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*StartInstanceResponse, error) {
	rsp, err := c.StartInstanceWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseStartInstanceResponse(rsp)
}

func (c *ClientWithResponses) StartInstanceWithResponse(ctx context.Context, id string, body StartInstanceJSONRequestBody, reqEditors ...RequestEditorFn) (*StartInstanceResponse, error) {
	rsp, err := c.StartInstance(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseStartInstanceResponse(rsp)
}

// StopInstanceWithResponse request returning *StopInstanceResponse
func (c *ClientWithResponses) StopInstanceWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*StopInstanceResponse, error) {
	rsp, err := c.StopInstance(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseStopInstanceResponse(rsp)
}

// RevertInstanceToSnapshotWithBodyWithResponse request with arbitrary body returning *RevertInstanceToSnapshotResponse
func (c *ClientWithResponses) RevertInstanceToSnapshotWithBodyWithResponse(ctx context.Context, instanceId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*RevertInstanceToSnapshotResponse, error) {
	rsp, err := c.RevertInstanceToSnapshotWithBody(ctx, instanceId, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseRevertInstanceToSnapshotResponse(rsp)
}

func (c *ClientWithResponses) RevertInstanceToSnapshotWithResponse(ctx context.Context, instanceId string, body RevertInstanceToSnapshotJSONRequestBody, reqEditors ...RequestEditorFn) (*RevertInstanceToSnapshotResponse, error) {
	rsp, err := c.RevertInstanceToSnapshot(ctx, instanceId, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseRevertInstanceToSnapshotResponse(rsp)
}

// ListLoadBalancersWithResponse request returning *ListLoadBalancersResponse
func (c *ClientWithResponses) ListLoadBalancersWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListLoadBalancersResponse, error) {
	rsp, err := c.ListLoadBalancers(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListLoadBalancersResponse(rsp)
}

// CreateLoadBalancerWithBodyWithResponse request with arbitrary body returning *CreateLoadBalancerResponse
func (c *ClientWithResponses) CreateLoadBalancerWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateLoadBalancerResponse, error) {
	rsp, err := c.CreateLoadBalancerWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateLoadBalancerResponse(rsp)
}

func (c *ClientWithResponses) CreateLoadBalancerWithResponse(ctx context.Context, body CreateLoadBalancerJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateLoadBalancerResponse, error) {
	rsp, err := c.CreateLoadBalancer(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateLoadBalancerResponse(rsp)
}

// DeleteLoadBalancerWithResponse request returning *DeleteLoadBalancerResponse
func (c *ClientWithResponses) DeleteLoadBalancerWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*DeleteLoadBalancerResponse, error) {
	rsp, err := c.DeleteLoadBalancer(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteLoadBalancerResponse(rsp)
}

// GetLoadBalancerWithResponse request returning *GetLoadBalancerResponse
func (c *ClientWithResponses) GetLoadBalancerWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetLoadBalancerResponse, error) {
	rsp, err := c.GetLoadBalancer(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetLoadBalancerResponse(rsp)
}

// UpdateLoadBalancerWithBodyWithResponse request with arbitrary body returning *UpdateLoadBalancerResponse
func (c *ClientWithResponses) UpdateLoadBalancerWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateLoadBalancerResponse, error) {
	rsp, err := c.UpdateLoadBalancerWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateLoadBalancerResponse(rsp)
}

func (c *ClientWithResponses) UpdateLoadBalancerWithResponse(ctx context.Context, id string, body UpdateLoadBalancerJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateLoadBalancerResponse, error) {
	rsp, err := c.UpdateLoadBalancer(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateLoadBalancerResponse(rsp)
}

// AddServiceToLoadBalancerWithBodyWithResponse request with arbitrary body returning *AddServiceToLoadBalancerResponse
func (c *ClientWithResponses) AddServiceToLoadBalancerWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*AddServiceToLoadBalancerResponse, error) {
	rsp, err := c.AddServiceToLoadBalancerWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseAddServiceToLoadBalancerResponse(rsp)
}

func (c *ClientWithResponses) AddServiceToLoadBalancerWithResponse(ctx context.Context, id string, body AddServiceToLoadBalancerJSONRequestBody, reqEditors ...RequestEditorFn) (*AddServiceToLoadBalancerResponse, error) {
	rsp, err := c.AddServiceToLoadBalancer(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseAddServiceToLoadBalancerResponse(rsp)
}

// DeleteLoadBalancerServiceWithResponse request returning *DeleteLoadBalancerServiceResponse
func (c *ClientWithResponses) DeleteLoadBalancerServiceWithResponse(ctx context.Context, id string, serviceId string, reqEditors ...RequestEditorFn) (*DeleteLoadBalancerServiceResponse, error) {
	rsp, err := c.DeleteLoadBalancerService(ctx, id, serviceId, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteLoadBalancerServiceResponse(rsp)
}

// GetLoadBalancerServiceWithResponse request returning *GetLoadBalancerServiceResponse
func (c *ClientWithResponses) GetLoadBalancerServiceWithResponse(ctx context.Context, id string, serviceId string, reqEditors ...RequestEditorFn) (*GetLoadBalancerServiceResponse, error) {
	rsp, err := c.GetLoadBalancerService(ctx, id, serviceId, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetLoadBalancerServiceResponse(rsp)
}

// UpdateLoadBalancerServiceWithBodyWithResponse request with arbitrary body returning *UpdateLoadBalancerServiceResponse
func (c *ClientWithResponses) UpdateLoadBalancerServiceWithBodyWithResponse(ctx context.Context, id string, serviceId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateLoadBalancerServiceResponse, error) {
	rsp, err := c.UpdateLoadBalancerServiceWithBody(ctx, id, serviceId, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateLoadBalancerServiceResponse(rsp)
}

func (c *ClientWithResponses) UpdateLoadBalancerServiceWithResponse(ctx context.Context, id string, serviceId string, body UpdateLoadBalancerServiceJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateLoadBalancerServiceResponse, error) {
	rsp, err := c.UpdateLoadBalancerService(ctx, id, serviceId, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateLoadBalancerServiceResponse(rsp)
}

// ResetLoadBalancerServiceFieldWithResponse request returning *ResetLoadBalancerServiceFieldResponse
func (c *ClientWithResponses) ResetLoadBalancerServiceFieldWithResponse(ctx context.Context, id string, serviceId string, field ResetLoadBalancerServiceFieldParamsField, reqEditors ...RequestEditorFn) (*ResetLoadBalancerServiceFieldResponse, error) {
	rsp, err := c.ResetLoadBalancerServiceField(ctx, id, serviceId, field, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseResetLoadBalancerServiceFieldResponse(rsp)
}

// ResetLoadBalancerFieldWithResponse request returning *ResetLoadBalancerFieldResponse
func (c *ClientWithResponses) ResetLoadBalancerFieldWithResponse(ctx context.Context, id string, field ResetLoadBalancerFieldParamsField, reqEditors ...RequestEditorFn) (*ResetLoadBalancerFieldResponse, error) {
	rsp, err := c.ResetLoadBalancerField(ctx, id, field, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseResetLoadBalancerFieldResponse(rsp)
}

// GetOperationWithResponse request returning *GetOperationResponse
func (c *ClientWithResponses) GetOperationWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetOperationResponse, error) {
	rsp, err := c.GetOperation(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetOperationResponse(rsp)
}

// ListPrivateNetworksWithResponse request returning *ListPrivateNetworksResponse
func (c *ClientWithResponses) ListPrivateNetworksWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListPrivateNetworksResponse, error) {
	rsp, err := c.ListPrivateNetworks(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListPrivateNetworksResponse(rsp)
}

// CreatePrivateNetworkWithBodyWithResponse request with arbitrary body returning *CreatePrivateNetworkResponse
func (c *ClientWithResponses) CreatePrivateNetworkWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreatePrivateNetworkResponse, error) {
	rsp, err := c.CreatePrivateNetworkWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreatePrivateNetworkResponse(rsp)
}

func (c *ClientWithResponses) CreatePrivateNetworkWithResponse(ctx context.Context, body CreatePrivateNetworkJSONRequestBody, reqEditors ...RequestEditorFn) (*CreatePrivateNetworkResponse, error) {
	rsp, err := c.CreatePrivateNetwork(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreatePrivateNetworkResponse(rsp)
}

// DeletePrivateNetworkWithResponse request returning *DeletePrivateNetworkResponse
func (c *ClientWithResponses) DeletePrivateNetworkWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*DeletePrivateNetworkResponse, error) {
	rsp, err := c.DeletePrivateNetwork(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeletePrivateNetworkResponse(rsp)
}

// GetPrivateNetworkWithResponse request returning *GetPrivateNetworkResponse
func (c *ClientWithResponses) GetPrivateNetworkWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetPrivateNetworkResponse, error) {
	rsp, err := c.GetPrivateNetwork(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetPrivateNetworkResponse(rsp)
}

// UpdatePrivateNetworkWithBodyWithResponse request with arbitrary body returning *UpdatePrivateNetworkResponse
func (c *ClientWithResponses) UpdatePrivateNetworkWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdatePrivateNetworkResponse, error) {
	rsp, err := c.UpdatePrivateNetworkWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdatePrivateNetworkResponse(rsp)
}

func (c *ClientWithResponses) UpdatePrivateNetworkWithResponse(ctx context.Context, id string, body UpdatePrivateNetworkJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdatePrivateNetworkResponse, error) {
	rsp, err := c.UpdatePrivateNetwork(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdatePrivateNetworkResponse(rsp)
}

// ResetPrivateNetworkFieldWithResponse request returning *ResetPrivateNetworkFieldResponse
func (c *ClientWithResponses) ResetPrivateNetworkFieldWithResponse(ctx context.Context, id string, field ResetPrivateNetworkFieldParamsField, reqEditors ...RequestEditorFn) (*ResetPrivateNetworkFieldResponse, error) {
	rsp, err := c.ResetPrivateNetworkField(ctx, id, field, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseResetPrivateNetworkFieldResponse(rsp)
}

// AttachInstanceToPrivateNetworkWithBodyWithResponse request with arbitrary body returning *AttachInstanceToPrivateNetworkResponse
func (c *ClientWithResponses) AttachInstanceToPrivateNetworkWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*AttachInstanceToPrivateNetworkResponse, error) {
	rsp, err := c.AttachInstanceToPrivateNetworkWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseAttachInstanceToPrivateNetworkResponse(rsp)
}

func (c *ClientWithResponses) AttachInstanceToPrivateNetworkWithResponse(ctx context.Context, id string, body AttachInstanceToPrivateNetworkJSONRequestBody, reqEditors ...RequestEditorFn) (*AttachInstanceToPrivateNetworkResponse, error) {
	rsp, err := c.AttachInstanceToPrivateNetwork(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseAttachInstanceToPrivateNetworkResponse(rsp)
}

// DetachInstanceFromPrivateNetworkWithBodyWithResponse request with arbitrary body returning *DetachInstanceFromPrivateNetworkResponse
func (c *ClientWithResponses) DetachInstanceFromPrivateNetworkWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*DetachInstanceFromPrivateNetworkResponse, error) {
	rsp, err := c.DetachInstanceFromPrivateNetworkWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDetachInstanceFromPrivateNetworkResponse(rsp)
}

func (c *ClientWithResponses) DetachInstanceFromPrivateNetworkWithResponse(ctx context.Context, id string, body DetachInstanceFromPrivateNetworkJSONRequestBody, reqEditors ...RequestEditorFn) (*DetachInstanceFromPrivateNetworkResponse, error) {
	rsp, err := c.DetachInstanceFromPrivateNetwork(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDetachInstanceFromPrivateNetworkResponse(rsp)
}

// UpdatePrivateNetworkInstanceIpWithBodyWithResponse request with arbitrary body returning *UpdatePrivateNetworkInstanceIpResponse
func (c *ClientWithResponses) UpdatePrivateNetworkInstanceIpWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdatePrivateNetworkInstanceIpResponse, error) {
	rsp, err := c.UpdatePrivateNetworkInstanceIpWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdatePrivateNetworkInstanceIpResponse(rsp)
}

func (c *ClientWithResponses) UpdatePrivateNetworkInstanceIpWithResponse(ctx context.Context, id string, body UpdatePrivateNetworkInstanceIpJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdatePrivateNetworkInstanceIpResponse, error) {
	rsp, err := c.UpdatePrivateNetworkInstanceIp(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdatePrivateNetworkInstanceIpResponse(rsp)
}

// ListQuotasWithResponse request returning *ListQuotasResponse
func (c *ClientWithResponses) ListQuotasWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListQuotasResponse, error) {
	rsp, err := c.ListQuotas(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListQuotasResponse(rsp)
}

// GetQuotaWithResponse request returning *GetQuotaResponse
func (c *ClientWithResponses) GetQuotaWithResponse(ctx context.Context, entity string, reqEditors ...RequestEditorFn) (*GetQuotaResponse, error) {
	rsp, err := c.GetQuota(ctx, entity, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetQuotaResponse(rsp)
}

// ListSecurityGroupsWithResponse request returning *ListSecurityGroupsResponse
func (c *ClientWithResponses) ListSecurityGroupsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListSecurityGroupsResponse, error) {
	rsp, err := c.ListSecurityGroups(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListSecurityGroupsResponse(rsp)
}

// CreateSecurityGroupWithBodyWithResponse request with arbitrary body returning *CreateSecurityGroupResponse
func (c *ClientWithResponses) CreateSecurityGroupWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateSecurityGroupResponse, error) {
	rsp, err := c.CreateSecurityGroupWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateSecurityGroupResponse(rsp)
}

func (c *ClientWithResponses) CreateSecurityGroupWithResponse(ctx context.Context, body CreateSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateSecurityGroupResponse, error) {
	rsp, err := c.CreateSecurityGroup(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateSecurityGroupResponse(rsp)
}

// DeleteSecurityGroupWithResponse request returning *DeleteSecurityGroupResponse
func (c *ClientWithResponses) DeleteSecurityGroupWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*DeleteSecurityGroupResponse, error) {
	rsp, err := c.DeleteSecurityGroup(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteSecurityGroupResponse(rsp)
}

// GetSecurityGroupWithResponse request returning *GetSecurityGroupResponse
func (c *ClientWithResponses) GetSecurityGroupWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetSecurityGroupResponse, error) {
	rsp, err := c.GetSecurityGroup(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetSecurityGroupResponse(rsp)
}

// AddRuleToSecurityGroupWithBodyWithResponse request with arbitrary body returning *AddRuleToSecurityGroupResponse
func (c *ClientWithResponses) AddRuleToSecurityGroupWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*AddRuleToSecurityGroupResponse, error) {
	rsp, err := c.AddRuleToSecurityGroupWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseAddRuleToSecurityGroupResponse(rsp)
}

func (c *ClientWithResponses) AddRuleToSecurityGroupWithResponse(ctx context.Context, id string, body AddRuleToSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*AddRuleToSecurityGroupResponse, error) {
	rsp, err := c.AddRuleToSecurityGroup(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseAddRuleToSecurityGroupResponse(rsp)
}

// DeleteRuleFromSecurityGroupWithResponse request returning *DeleteRuleFromSecurityGroupResponse
func (c *ClientWithResponses) DeleteRuleFromSecurityGroupWithResponse(ctx context.Context, id string, ruleId string, reqEditors ...RequestEditorFn) (*DeleteRuleFromSecurityGroupResponse, error) {
	rsp, err := c.DeleteRuleFromSecurityGroup(ctx, id, ruleId, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteRuleFromSecurityGroupResponse(rsp)
}

// AddExternalSourceToSecurityGroupWithBodyWithResponse request with arbitrary body returning *AddExternalSourceToSecurityGroupResponse
func (c *ClientWithResponses) AddExternalSourceToSecurityGroupWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*AddExternalSourceToSecurityGroupResponse, error) {
	rsp, err := c.AddExternalSourceToSecurityGroupWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseAddExternalSourceToSecurityGroupResponse(rsp)
}

func (c *ClientWithResponses) AddExternalSourceToSecurityGroupWithResponse(ctx context.Context, id string, body AddExternalSourceToSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*AddExternalSourceToSecurityGroupResponse, error) {
	rsp, err := c.AddExternalSourceToSecurityGroup(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseAddExternalSourceToSecurityGroupResponse(rsp)
}

// AttachInstanceToSecurityGroupWithBodyWithResponse request with arbitrary body returning *AttachInstanceToSecurityGroupResponse
func (c *ClientWithResponses) AttachInstanceToSecurityGroupWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*AttachInstanceToSecurityGroupResponse, error) {
	rsp, err := c.AttachInstanceToSecurityGroupWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseAttachInstanceToSecurityGroupResponse(rsp)
}

func (c *ClientWithResponses) AttachInstanceToSecurityGroupWithResponse(ctx context.Context, id string, body AttachInstanceToSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*AttachInstanceToSecurityGroupResponse, error) {
	rsp, err := c.AttachInstanceToSecurityGroup(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseAttachInstanceToSecurityGroupResponse(rsp)
}

// DetachInstanceFromSecurityGroupWithBodyWithResponse request with arbitrary body returning *DetachInstanceFromSecurityGroupResponse
func (c *ClientWithResponses) DetachInstanceFromSecurityGroupWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*DetachInstanceFromSecurityGroupResponse, error) {
	rsp, err := c.DetachInstanceFromSecurityGroupWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDetachInstanceFromSecurityGroupResponse(rsp)
}

func (c *ClientWithResponses) DetachInstanceFromSecurityGroupWithResponse(ctx context.Context, id string, body DetachInstanceFromSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*DetachInstanceFromSecurityGroupResponse, error) {
	rsp, err := c.DetachInstanceFromSecurityGroup(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDetachInstanceFromSecurityGroupResponse(rsp)
}

// RemoveExternalSourceFromSecurityGroupWithBodyWithResponse request with arbitrary body returning *RemoveExternalSourceFromSecurityGroupResponse
func (c *ClientWithResponses) RemoveExternalSourceFromSecurityGroupWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*RemoveExternalSourceFromSecurityGroupResponse, error) {
	rsp, err := c.RemoveExternalSourceFromSecurityGroupWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseRemoveExternalSourceFromSecurityGroupResponse(rsp)
}

func (c *ClientWithResponses) RemoveExternalSourceFromSecurityGroupWithResponse(ctx context.Context, id string, body RemoveExternalSourceFromSecurityGroupJSONRequestBody, reqEditors ...RequestEditorFn) (*RemoveExternalSourceFromSecurityGroupResponse, error) {
	rsp, err := c.RemoveExternalSourceFromSecurityGroup(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseRemoveExternalSourceFromSecurityGroupResponse(rsp)
}

// ListSksClustersWithResponse request returning *ListSksClustersResponse
func (c *ClientWithResponses) ListSksClustersWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListSksClustersResponse, error) {
	rsp, err := c.ListSksClusters(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListSksClustersResponse(rsp)
}

// CreateSksClusterWithBodyWithResponse request with arbitrary body returning *CreateSksClusterResponse
func (c *ClientWithResponses) CreateSksClusterWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateSksClusterResponse, error) {
	rsp, err := c.CreateSksClusterWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateSksClusterResponse(rsp)
}

func (c *ClientWithResponses) CreateSksClusterWithResponse(ctx context.Context, body CreateSksClusterJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateSksClusterResponse, error) {
	rsp, err := c.CreateSksCluster(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateSksClusterResponse(rsp)
}

// ListSksClusterDeprecatedResourcesWithResponse request returning *ListSksClusterDeprecatedResourcesResponse
func (c *ClientWithResponses) ListSksClusterDeprecatedResourcesWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*ListSksClusterDeprecatedResourcesResponse, error) {
	rsp, err := c.ListSksClusterDeprecatedResources(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListSksClusterDeprecatedResourcesResponse(rsp)
}

// GenerateSksClusterKubeconfigWithBodyWithResponse request with arbitrary body returning *GenerateSksClusterKubeconfigResponse
func (c *ClientWithResponses) GenerateSksClusterKubeconfigWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*GenerateSksClusterKubeconfigResponse, error) {
	rsp, err := c.GenerateSksClusterKubeconfigWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGenerateSksClusterKubeconfigResponse(rsp)
}

func (c *ClientWithResponses) GenerateSksClusterKubeconfigWithResponse(ctx context.Context, id string, body GenerateSksClusterKubeconfigJSONRequestBody, reqEditors ...RequestEditorFn) (*GenerateSksClusterKubeconfigResponse, error) {
	rsp, err := c.GenerateSksClusterKubeconfig(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGenerateSksClusterKubeconfigResponse(rsp)
}

// ListSksClusterVersionsWithResponse request returning *ListSksClusterVersionsResponse
func (c *ClientWithResponses) ListSksClusterVersionsWithResponse(ctx context.Context, params *ListSksClusterVersionsParams, reqEditors ...RequestEditorFn) (*ListSksClusterVersionsResponse, error) {
	rsp, err := c.ListSksClusterVersions(ctx, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListSksClusterVersionsResponse(rsp)
}

// DeleteSksClusterWithResponse request returning *DeleteSksClusterResponse
func (c *ClientWithResponses) DeleteSksClusterWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*DeleteSksClusterResponse, error) {
	rsp, err := c.DeleteSksCluster(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteSksClusterResponse(rsp)
}

// GetSksClusterWithResponse request returning *GetSksClusterResponse
func (c *ClientWithResponses) GetSksClusterWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetSksClusterResponse, error) {
	rsp, err := c.GetSksCluster(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetSksClusterResponse(rsp)
}

// UpdateSksClusterWithBodyWithResponse request with arbitrary body returning *UpdateSksClusterResponse
func (c *ClientWithResponses) UpdateSksClusterWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateSksClusterResponse, error) {
	rsp, err := c.UpdateSksClusterWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateSksClusterResponse(rsp)
}

func (c *ClientWithResponses) UpdateSksClusterWithResponse(ctx context.Context, id string, body UpdateSksClusterJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateSksClusterResponse, error) {
	rsp, err := c.UpdateSksCluster(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateSksClusterResponse(rsp)
}

// GetSksClusterAuthorityCertWithResponse request returning *GetSksClusterAuthorityCertResponse
func (c *ClientWithResponses) GetSksClusterAuthorityCertWithResponse(ctx context.Context, id string, authority GetSksClusterAuthorityCertParamsAuthority, reqEditors ...RequestEditorFn) (*GetSksClusterAuthorityCertResponse, error) {
	rsp, err := c.GetSksClusterAuthorityCert(ctx, id, authority, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetSksClusterAuthorityCertResponse(rsp)
}

// CreateSksNodepoolWithBodyWithResponse request with arbitrary body returning *CreateSksNodepoolResponse
func (c *ClientWithResponses) CreateSksNodepoolWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CreateSksNodepoolResponse, error) {
	rsp, err := c.CreateSksNodepoolWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateSksNodepoolResponse(rsp)
}

func (c *ClientWithResponses) CreateSksNodepoolWithResponse(ctx context.Context, id string, body CreateSksNodepoolJSONRequestBody, reqEditors ...RequestEditorFn) (*CreateSksNodepoolResponse, error) {
	rsp, err := c.CreateSksNodepool(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCreateSksNodepoolResponse(rsp)
}

// DeleteSksNodepoolWithResponse request returning *DeleteSksNodepoolResponse
func (c *ClientWithResponses) DeleteSksNodepoolWithResponse(ctx context.Context, id string, sksNodepoolId string, reqEditors ...RequestEditorFn) (*DeleteSksNodepoolResponse, error) {
	rsp, err := c.DeleteSksNodepool(ctx, id, sksNodepoolId, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteSksNodepoolResponse(rsp)
}

// GetSksNodepoolWithResponse request returning *GetSksNodepoolResponse
func (c *ClientWithResponses) GetSksNodepoolWithResponse(ctx context.Context, id string, sksNodepoolId string, reqEditors ...RequestEditorFn) (*GetSksNodepoolResponse, error) {
	rsp, err := c.GetSksNodepool(ctx, id, sksNodepoolId, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetSksNodepoolResponse(rsp)
}

// UpdateSksNodepoolWithBodyWithResponse request with arbitrary body returning *UpdateSksNodepoolResponse
func (c *ClientWithResponses) UpdateSksNodepoolWithBodyWithResponse(ctx context.Context, id string, sksNodepoolId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateSksNodepoolResponse, error) {
	rsp, err := c.UpdateSksNodepoolWithBody(ctx, id, sksNodepoolId, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateSksNodepoolResponse(rsp)
}

func (c *ClientWithResponses) UpdateSksNodepoolWithResponse(ctx context.Context, id string, sksNodepoolId string, body UpdateSksNodepoolJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateSksNodepoolResponse, error) {
	rsp, err := c.UpdateSksNodepool(ctx, id, sksNodepoolId, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateSksNodepoolResponse(rsp)
}

// ResetSksNodepoolFieldWithResponse request returning *ResetSksNodepoolFieldResponse
func (c *ClientWithResponses) ResetSksNodepoolFieldWithResponse(ctx context.Context, id string, sksNodepoolId string, field ResetSksNodepoolFieldParamsField, reqEditors ...RequestEditorFn) (*ResetSksNodepoolFieldResponse, error) {
	rsp, err := c.ResetSksNodepoolField(ctx, id, sksNodepoolId, field, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseResetSksNodepoolFieldResponse(rsp)
}

// EvictSksNodepoolMembersWithBodyWithResponse request with arbitrary body returning *EvictSksNodepoolMembersResponse
func (c *ClientWithResponses) EvictSksNodepoolMembersWithBodyWithResponse(ctx context.Context, id string, sksNodepoolId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*EvictSksNodepoolMembersResponse, error) {
	rsp, err := c.EvictSksNodepoolMembersWithBody(ctx, id, sksNodepoolId, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseEvictSksNodepoolMembersResponse(rsp)
}

func (c *ClientWithResponses) EvictSksNodepoolMembersWithResponse(ctx context.Context, id string, sksNodepoolId string, body EvictSksNodepoolMembersJSONRequestBody, reqEditors ...RequestEditorFn) (*EvictSksNodepoolMembersResponse, error) {
	rsp, err := c.EvictSksNodepoolMembers(ctx, id, sksNodepoolId, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseEvictSksNodepoolMembersResponse(rsp)
}

// ScaleSksNodepoolWithBodyWithResponse request with arbitrary body returning *ScaleSksNodepoolResponse
func (c *ClientWithResponses) ScaleSksNodepoolWithBodyWithResponse(ctx context.Context, id string, sksNodepoolId string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ScaleSksNodepoolResponse, error) {
	rsp, err := c.ScaleSksNodepoolWithBody(ctx, id, sksNodepoolId, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseScaleSksNodepoolResponse(rsp)
}

func (c *ClientWithResponses) ScaleSksNodepoolWithResponse(ctx context.Context, id string, sksNodepoolId string, body ScaleSksNodepoolJSONRequestBody, reqEditors ...RequestEditorFn) (*ScaleSksNodepoolResponse, error) {
	rsp, err := c.ScaleSksNodepool(ctx, id, sksNodepoolId, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseScaleSksNodepoolResponse(rsp)
}

// RotateSksCcmCredentialsWithResponse request returning *RotateSksCcmCredentialsResponse
func (c *ClientWithResponses) RotateSksCcmCredentialsWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*RotateSksCcmCredentialsResponse, error) {
	rsp, err := c.RotateSksCcmCredentials(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseRotateSksCcmCredentialsResponse(rsp)
}

// RotateSksOperatorsCaWithResponse request returning *RotateSksOperatorsCaResponse
func (c *ClientWithResponses) RotateSksOperatorsCaWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*RotateSksOperatorsCaResponse, error) {
	rsp, err := c.RotateSksOperatorsCa(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseRotateSksOperatorsCaResponse(rsp)
}

// UpgradeSksClusterWithBodyWithResponse request with arbitrary body returning *UpgradeSksClusterResponse
func (c *ClientWithResponses) UpgradeSksClusterWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpgradeSksClusterResponse, error) {
	rsp, err := c.UpgradeSksClusterWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpgradeSksClusterResponse(rsp)
}

func (c *ClientWithResponses) UpgradeSksClusterWithResponse(ctx context.Context, id string, body UpgradeSksClusterJSONRequestBody, reqEditors ...RequestEditorFn) (*UpgradeSksClusterResponse, error) {
	rsp, err := c.UpgradeSksCluster(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpgradeSksClusterResponse(rsp)
}

// UpgradeSksClusterServiceLevelWithResponse request returning *UpgradeSksClusterServiceLevelResponse
func (c *ClientWithResponses) UpgradeSksClusterServiceLevelWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*UpgradeSksClusterServiceLevelResponse, error) {
	rsp, err := c.UpgradeSksClusterServiceLevel(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpgradeSksClusterServiceLevelResponse(rsp)
}

// ResetSksClusterFieldWithResponse request returning *ResetSksClusterFieldResponse
func (c *ClientWithResponses) ResetSksClusterFieldWithResponse(ctx context.Context, id string, field ResetSksClusterFieldParamsField, reqEditors ...RequestEditorFn) (*ResetSksClusterFieldResponse, error) {
	rsp, err := c.ResetSksClusterField(ctx, id, field, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseResetSksClusterFieldResponse(rsp)
}

// ListSnapshotsWithResponse request returning *ListSnapshotsResponse
func (c *ClientWithResponses) ListSnapshotsWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListSnapshotsResponse, error) {
	rsp, err := c.ListSnapshots(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListSnapshotsResponse(rsp)
}

// DeleteSnapshotWithResponse request returning *DeleteSnapshotResponse
func (c *ClientWithResponses) DeleteSnapshotWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*DeleteSnapshotResponse, error) {
	rsp, err := c.DeleteSnapshot(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteSnapshotResponse(rsp)
}

// GetSnapshotWithResponse request returning *GetSnapshotResponse
func (c *ClientWithResponses) GetSnapshotWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetSnapshotResponse, error) {
	rsp, err := c.GetSnapshot(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetSnapshotResponse(rsp)
}

// ExportSnapshotWithResponse request returning *ExportSnapshotResponse
func (c *ClientWithResponses) ExportSnapshotWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*ExportSnapshotResponse, error) {
	rsp, err := c.ExportSnapshot(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseExportSnapshotResponse(rsp)
}

// PromoteSnapshotToTemplateWithBodyWithResponse request with arbitrary body returning *PromoteSnapshotToTemplateResponse
func (c *ClientWithResponses) PromoteSnapshotToTemplateWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PromoteSnapshotToTemplateResponse, error) {
	rsp, err := c.PromoteSnapshotToTemplateWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePromoteSnapshotToTemplateResponse(rsp)
}

func (c *ClientWithResponses) PromoteSnapshotToTemplateWithResponse(ctx context.Context, id string, body PromoteSnapshotToTemplateJSONRequestBody, reqEditors ...RequestEditorFn) (*PromoteSnapshotToTemplateResponse, error) {
	rsp, err := c.PromoteSnapshotToTemplate(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePromoteSnapshotToTemplateResponse(rsp)
}

// GetSosPresignedUrlWithResponse request returning *GetSosPresignedUrlResponse
func (c *ClientWithResponses) GetSosPresignedUrlWithResponse(ctx context.Context, bucket string, params *GetSosPresignedUrlParams, reqEditors ...RequestEditorFn) (*GetSosPresignedUrlResponse, error) {
	rsp, err := c.GetSosPresignedUrl(ctx, bucket, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetSosPresignedUrlResponse(rsp)
}

// ListSshKeysWithResponse request returning *ListSshKeysResponse
func (c *ClientWithResponses) ListSshKeysWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListSshKeysResponse, error) {
	rsp, err := c.ListSshKeys(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListSshKeysResponse(rsp)
}

// RegisterSshKeyWithBodyWithResponse request with arbitrary body returning *RegisterSshKeyResponse
func (c *ClientWithResponses) RegisterSshKeyWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*RegisterSshKeyResponse, error) {
	rsp, err := c.RegisterSshKeyWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseRegisterSshKeyResponse(rsp)
}

func (c *ClientWithResponses) RegisterSshKeyWithResponse(ctx context.Context, body RegisterSshKeyJSONRequestBody, reqEditors ...RequestEditorFn) (*RegisterSshKeyResponse, error) {
	rsp, err := c.RegisterSshKey(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseRegisterSshKeyResponse(rsp)
}

// DeleteSshKeyWithResponse request returning *DeleteSshKeyResponse
func (c *ClientWithResponses) DeleteSshKeyWithResponse(ctx context.Context, name string, reqEditors ...RequestEditorFn) (*DeleteSshKeyResponse, error) {
	rsp, err := c.DeleteSshKey(ctx, name, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteSshKeyResponse(rsp)
}

// GetSshKeyWithResponse request returning *GetSshKeyResponse
func (c *ClientWithResponses) GetSshKeyWithResponse(ctx context.Context, name string, reqEditors ...RequestEditorFn) (*GetSshKeyResponse, error) {
	rsp, err := c.GetSshKey(ctx, name, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetSshKeyResponse(rsp)
}

// ListTemplatesWithResponse request returning *ListTemplatesResponse
func (c *ClientWithResponses) ListTemplatesWithResponse(ctx context.Context, params *ListTemplatesParams, reqEditors ...RequestEditorFn) (*ListTemplatesResponse, error) {
	rsp, err := c.ListTemplates(ctx, params, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListTemplatesResponse(rsp)
}

// RegisterTemplateWithBodyWithResponse request with arbitrary body returning *RegisterTemplateResponse
func (c *ClientWithResponses) RegisterTemplateWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*RegisterTemplateResponse, error) {
	rsp, err := c.RegisterTemplateWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseRegisterTemplateResponse(rsp)
}

func (c *ClientWithResponses) RegisterTemplateWithResponse(ctx context.Context, body RegisterTemplateJSONRequestBody, reqEditors ...RequestEditorFn) (*RegisterTemplateResponse, error) {
	rsp, err := c.RegisterTemplate(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseRegisterTemplateResponse(rsp)
}

// DeleteTemplateWithResponse request returning *DeleteTemplateResponse
func (c *ClientWithResponses) DeleteTemplateWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*DeleteTemplateResponse, error) {
	rsp, err := c.DeleteTemplate(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDeleteTemplateResponse(rsp)
}

// GetTemplateWithResponse request returning *GetTemplateResponse
func (c *ClientWithResponses) GetTemplateWithResponse(ctx context.Context, id string, reqEditors ...RequestEditorFn) (*GetTemplateResponse, error) {
	rsp, err := c.GetTemplate(ctx, id, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetTemplateResponse(rsp)
}

// CopyTemplateWithBodyWithResponse request with arbitrary body returning *CopyTemplateResponse
func (c *ClientWithResponses) CopyTemplateWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*CopyTemplateResponse, error) {
	rsp, err := c.CopyTemplateWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCopyTemplateResponse(rsp)
}

func (c *ClientWithResponses) CopyTemplateWithResponse(ctx context.Context, id string, body CopyTemplateJSONRequestBody, reqEditors ...RequestEditorFn) (*CopyTemplateResponse, error) {
	rsp, err := c.CopyTemplate(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseCopyTemplateResponse(rsp)
}

// UpdateTemplateWithBodyWithResponse request with arbitrary body returning *UpdateTemplateResponse
func (c *ClientWithResponses) UpdateTemplateWithBodyWithResponse(ctx context.Context, id string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*UpdateTemplateResponse, error) {
	rsp, err := c.UpdateTemplateWithBody(ctx, id, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateTemplateResponse(rsp)
}

func (c *ClientWithResponses) UpdateTemplateWithResponse(ctx context.Context, id string, body UpdateTemplateJSONRequestBody, reqEditors ...RequestEditorFn) (*UpdateTemplateResponse, error) {
	rsp, err := c.UpdateTemplate(ctx, id, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseUpdateTemplateResponse(rsp)
}

// ListZonesWithResponse request returning *ListZonesResponse
func (c *ClientWithResponses) ListZonesWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ListZonesResponse, error) {
	rsp, err := c.ListZones(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseListZonesResponse(rsp)
}

// ParseListAccessKeysResponse parses an HTTP response from a ListAccessKeysWithResponse call
func ParseListAccessKeysResponse(rsp *http.Response) (*ListAccessKeysResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListAccessKeysResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			AccessKeys *[]AccessKey `json:"access-keys,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateAccessKeyResponse parses an HTTP response from a CreateAccessKeyWithResponse call
func ParseCreateAccessKeyResponse(rsp *http.Response) (*CreateAccessKeyResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreateAccessKeyResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest AccessKey
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListAccessKeyKnownOperationsResponse parses an HTTP response from a ListAccessKeyKnownOperationsWithResponse call
func ParseListAccessKeyKnownOperationsResponse(rsp *http.Response) (*ListAccessKeyKnownOperationsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListAccessKeyKnownOperationsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			AccessKeyOperations *[]AccessKeyOperation `json:"access-key-operations,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListAccessKeyOperationsResponse parses an HTTP response from a ListAccessKeyOperationsWithResponse call
func ParseListAccessKeyOperationsResponse(rsp *http.Response) (*ListAccessKeyOperationsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListAccessKeyOperationsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			AccessKeyOperations *[]AccessKeyOperation `json:"access-key-operations,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseRevokeAccessKeyResponse parses an HTTP response from a RevokeAccessKeyWithResponse call
func ParseRevokeAccessKeyResponse(rsp *http.Response) (*RevokeAccessKeyResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &RevokeAccessKeyResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetAccessKeyResponse parses an HTTP response from a GetAccessKeyWithResponse call
func ParseGetAccessKeyResponse(rsp *http.Response) (*GetAccessKeyResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetAccessKeyResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest AccessKey
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListAntiAffinityGroupsResponse parses an HTTP response from a ListAntiAffinityGroupsWithResponse call
func ParseListAntiAffinityGroupsResponse(rsp *http.Response) (*ListAntiAffinityGroupsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListAntiAffinityGroupsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			AntiAffinityGroups *[]AntiAffinityGroup `json:"anti-affinity-groups,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateAntiAffinityGroupResponse parses an HTTP response from a CreateAntiAffinityGroupWithResponse call
func ParseCreateAntiAffinityGroupResponse(rsp *http.Response) (*CreateAntiAffinityGroupResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreateAntiAffinityGroupResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeleteAntiAffinityGroupResponse parses an HTTP response from a DeleteAntiAffinityGroupWithResponse call
func ParseDeleteAntiAffinityGroupResponse(rsp *http.Response) (*DeleteAntiAffinityGroupResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteAntiAffinityGroupResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetAntiAffinityGroupResponse parses an HTTP response from a GetAntiAffinityGroupWithResponse call
func ParseGetAntiAffinityGroupResponse(rsp *http.Response) (*GetAntiAffinityGroupResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetAntiAffinityGroupResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest AntiAffinityGroup
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetDbaasCaCertificateResponse parses an HTTP response from a GetDbaasCaCertificateWithResponse call
func ParseGetDbaasCaCertificateResponse(rsp *http.Response) (*GetDbaasCaCertificateResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetDbaasCaCertificateResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Certificate *string `json:"certificate,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetDbaasServiceKafkaResponse parses an HTTP response from a GetDbaasServiceKafkaWithResponse call
func ParseGetDbaasServiceKafkaResponse(rsp *http.Response) (*GetDbaasServiceKafkaResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetDbaasServiceKafkaResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DbaasServiceKafka
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateDbaasServiceKafkaResponse parses an HTTP response from a CreateDbaasServiceKafkaWithResponse call
func ParseCreateDbaasServiceKafkaResponse(rsp *http.Response) (*CreateDbaasServiceKafkaResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreateDbaasServiceKafkaResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DbaasServiceKafka
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseUpdateDbaasServiceKafkaResponse parses an HTTP response from a UpdateDbaasServiceKafkaWithResponse call
func ParseUpdateDbaasServiceKafkaResponse(rsp *http.Response) (*UpdateDbaasServiceKafkaResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &UpdateDbaasServiceKafkaResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DbaasServiceKafka
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetDbaasMigrationStatusResponse parses an HTTP response from a GetDbaasMigrationStatusWithResponse call
func ParseGetDbaasMigrationStatusResponse(rsp *http.Response) (*GetDbaasMigrationStatusResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetDbaasMigrationStatusResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DbaasMigrationStatus
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetDbaasServiceMysqlResponse parses an HTTP response from a GetDbaasServiceMysqlWithResponse call
func ParseGetDbaasServiceMysqlResponse(rsp *http.Response) (*GetDbaasServiceMysqlResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetDbaasServiceMysqlResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DbaasServiceMysql
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateDbaasServiceMysqlResponse parses an HTTP response from a CreateDbaasServiceMysqlWithResponse call
func ParseCreateDbaasServiceMysqlResponse(rsp *http.Response) (*CreateDbaasServiceMysqlResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreateDbaasServiceMysqlResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DbaasServiceMysql
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseUpdateDbaasServiceMysqlResponse parses an HTTP response from a UpdateDbaasServiceMysqlWithResponse call
func ParseUpdateDbaasServiceMysqlResponse(rsp *http.Response) (*UpdateDbaasServiceMysqlResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &UpdateDbaasServiceMysqlResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DbaasServiceMysql
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetDbaasServiceOpensearchResponse parses an HTTP response from a GetDbaasServiceOpensearchWithResponse call
func ParseGetDbaasServiceOpensearchResponse(rsp *http.Response) (*GetDbaasServiceOpensearchResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetDbaasServiceOpensearchResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DbaasServiceOpensearch
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateDbaasServiceOpensearchResponse parses an HTTP response from a CreateDbaasServiceOpensearchWithResponse call
func ParseCreateDbaasServiceOpensearchResponse(rsp *http.Response) (*CreateDbaasServiceOpensearchResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreateDbaasServiceOpensearchResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DbaasServiceOpensearch
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseUpdateDbaasServiceOpensearchResponse parses an HTTP response from a UpdateDbaasServiceOpensearchWithResponse call
func ParseUpdateDbaasServiceOpensearchResponse(rsp *http.Response) (*UpdateDbaasServiceOpensearchResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &UpdateDbaasServiceOpensearchResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DbaasServiceOpensearch
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetDbaasServicePgResponse parses an HTTP response from a GetDbaasServicePgWithResponse call
func ParseGetDbaasServicePgResponse(rsp *http.Response) (*GetDbaasServicePgResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetDbaasServicePgResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DbaasServicePg
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateDbaasServicePgResponse parses an HTTP response from a CreateDbaasServicePgWithResponse call
func ParseCreateDbaasServicePgResponse(rsp *http.Response) (*CreateDbaasServicePgResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreateDbaasServicePgResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DbaasServicePg
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseUpdateDbaasServicePgResponse parses an HTTP response from a UpdateDbaasServicePgWithResponse call
func ParseUpdateDbaasServicePgResponse(rsp *http.Response) (*UpdateDbaasServicePgResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &UpdateDbaasServicePgResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DbaasServicePg
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetDbaasServiceRedisResponse parses an HTTP response from a GetDbaasServiceRedisWithResponse call
func ParseGetDbaasServiceRedisResponse(rsp *http.Response) (*GetDbaasServiceRedisResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetDbaasServiceRedisResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DbaasServiceRedis
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateDbaasServiceRedisResponse parses an HTTP response from a CreateDbaasServiceRedisWithResponse call
func ParseCreateDbaasServiceRedisResponse(rsp *http.Response) (*CreateDbaasServiceRedisResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreateDbaasServiceRedisResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DbaasServiceRedis
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseUpdateDbaasServiceRedisResponse parses an HTTP response from a UpdateDbaasServiceRedisWithResponse call
func ParseUpdateDbaasServiceRedisResponse(rsp *http.Response) (*UpdateDbaasServiceRedisResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &UpdateDbaasServiceRedisResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DbaasServiceRedis
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListDbaasServicesResponse parses an HTTP response from a ListDbaasServicesWithResponse call
func ParseListDbaasServicesResponse(rsp *http.Response) (*ListDbaasServicesResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListDbaasServicesResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			DbaasServices *[]DbaasServiceCommon `json:"dbaas-services,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetDbaasServiceLogsResponse parses an HTTP response from a GetDbaasServiceLogsWithResponse call
func ParseGetDbaasServiceLogsResponse(rsp *http.Response) (*GetDbaasServiceLogsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetDbaasServiceLogsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DbaasServiceLogs
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetDbaasServiceMetricsResponse parses an HTTP response from a GetDbaasServiceMetricsWithResponse call
func ParseGetDbaasServiceMetricsResponse(rsp *http.Response) (*GetDbaasServiceMetricsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetDbaasServiceMetricsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Metrics *map[string]interface{} `json:"metrics,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListDbaasServiceTypesResponse parses an HTTP response from a ListDbaasServiceTypesWithResponse call
func ParseListDbaasServiceTypesResponse(rsp *http.Response) (*ListDbaasServiceTypesResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListDbaasServiceTypesResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			DbaasServiceTypes *[]DbaasServiceType `json:"dbaas-service-types,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetDbaasServiceTypeResponse parses an HTTP response from a GetDbaasServiceTypeWithResponse call
func ParseGetDbaasServiceTypeResponse(rsp *http.Response) (*GetDbaasServiceTypeResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetDbaasServiceTypeResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DbaasServiceType
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeleteDbaasServiceResponse parses an HTTP response from a DeleteDbaasServiceWithResponse call
func ParseDeleteDbaasServiceResponse(rsp *http.Response) (*DeleteDbaasServiceResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteDbaasServiceResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DbaasServiceCommon
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetDbaasSettingsKafkaResponse parses an HTTP response from a GetDbaasSettingsKafkaWithResponse call
func ParseGetDbaasSettingsKafkaResponse(rsp *http.Response) (*GetDbaasSettingsKafkaResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetDbaasSettingsKafkaResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Settings *struct {
				// Kafka broker configuration values
				Kafka *struct {
					AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
					Properties           *map[string]interface{} `json:"properties,omitempty"`
					Title                *string                 `json:"title,omitempty"`
					Type                 *string                 `json:"type,omitempty"`
				} `json:"kafka,omitempty"`

				// Kafka Connect configuration values
				KafkaConnect *struct {
					AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
					Properties           *map[string]interface{} `json:"properties,omitempty"`
					Title                *string                 `json:"title,omitempty"`
					Type                 *string                 `json:"type,omitempty"`
				} `json:"kafka-connect,omitempty"`

				// Kafka REST configuration
				KafkaRest *struct {
					AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
					Properties           *map[string]interface{} `json:"properties,omitempty"`
					Title                *string                 `json:"title,omitempty"`
					Type                 *string                 `json:"type,omitempty"`
				} `json:"kafka-rest,omitempty"`

				// Schema Registry configuration
				SchemaRegistry *struct {
					AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
					Properties           *map[string]interface{} `json:"properties,omitempty"`
					Title                *string                 `json:"title,omitempty"`
					Type                 *string                 `json:"type,omitempty"`
				} `json:"schema-registry,omitempty"`
			} `json:"settings,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetDbaasSettingsMysqlResponse parses an HTTP response from a GetDbaasSettingsMysqlWithResponse call
func ParseGetDbaasSettingsMysqlResponse(rsp *http.Response) (*GetDbaasSettingsMysqlResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetDbaasSettingsMysqlResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Settings *struct {
				// mysql.conf configuration values
				Mysql *struct {
					AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
					Properties           *map[string]interface{} `json:"properties,omitempty"`
					Title                *string                 `json:"title,omitempty"`
					Type                 *string                 `json:"type,omitempty"`
				} `json:"mysql,omitempty"`
			} `json:"settings,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetDbaasSettingsOpensearchResponse parses an HTTP response from a GetDbaasSettingsOpensearchWithResponse call
func ParseGetDbaasSettingsOpensearchResponse(rsp *http.Response) (*GetDbaasSettingsOpensearchResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetDbaasSettingsOpensearchResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Settings *struct {
				// OpenSearch configuration values
				Opensearch *struct {
					AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
					Properties           *map[string]interface{} `json:"properties,omitempty"`
					Title                *string                 `json:"title,omitempty"`
					Type                 *string                 `json:"type,omitempty"`
				} `json:"opensearch,omitempty"`
			} `json:"settings,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetDbaasSettingsPgResponse parses an HTTP response from a GetDbaasSettingsPgWithResponse call
func ParseGetDbaasSettingsPgResponse(rsp *http.Response) (*GetDbaasSettingsPgResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetDbaasSettingsPgResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Settings *struct {
				// postgresql.conf configuration values
				Pg *struct {
					AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
					Properties           *map[string]interface{} `json:"properties,omitempty"`
					Title                *string                 `json:"title,omitempty"`
					Type                 *string                 `json:"type,omitempty"`
				} `json:"pg,omitempty"`

				// PGBouncer connection pooling settings
				Pgbouncer *struct {
					AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
					Properties           *map[string]interface{} `json:"properties,omitempty"`
					Title                *string                 `json:"title,omitempty"`
					Type                 *string                 `json:"type,omitempty"`
				} `json:"pgbouncer,omitempty"`

				// PGLookout settings
				Pglookout *struct {
					AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
					Properties           *map[string]interface{} `json:"properties,omitempty"`
					Title                *string                 `json:"title,omitempty"`
					Type                 *string                 `json:"type,omitempty"`
				} `json:"pglookout,omitempty"`

				// TimescaleDB extension configuration values
				Timescaledb *struct {
					AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
					Properties           *map[string]interface{} `json:"properties,omitempty"`
					Title                *string                 `json:"title,omitempty"`
					Type                 *string                 `json:"type,omitempty"`
				} `json:"timescaledb,omitempty"`
			} `json:"settings,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetDbaasSettingsRedisResponse parses an HTTP response from a GetDbaasSettingsRedisWithResponse call
func ParseGetDbaasSettingsRedisResponse(rsp *http.Response) (*GetDbaasSettingsRedisResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetDbaasSettingsRedisResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Settings *struct {
				// Redis configuration values
				Redis *struct {
					AdditionalProperties *bool                   `json:"additionalProperties,omitempty"`
					Properties           *map[string]interface{} `json:"properties,omitempty"`
					Title                *string                 `json:"title,omitempty"`
					Type                 *string                 `json:"type,omitempty"`
				} `json:"redis,omitempty"`
			} `json:"settings,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListDeployTargetsResponse parses an HTTP response from a ListDeployTargetsWithResponse call
func ParseListDeployTargetsResponse(rsp *http.Response) (*ListDeployTargetsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListDeployTargetsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			DeployTargets *[]DeployTarget `json:"deploy-targets,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetDeployTargetResponse parses an HTTP response from a GetDeployTargetWithResponse call
func ParseGetDeployTargetResponse(rsp *http.Response) (*GetDeployTargetResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetDeployTargetResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DeployTarget
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListDnsDomainsResponse parses an HTTP response from a ListDnsDomainsWithResponse call
func ParseListDnsDomainsResponse(rsp *http.Response) (*ListDnsDomainsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListDnsDomainsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			DnsDomains *[]DnsDomain `json:"dns-domains,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetDnsDomainResponse parses an HTTP response from a GetDnsDomainWithResponse call
func ParseGetDnsDomainResponse(rsp *http.Response) (*GetDnsDomainResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetDnsDomainResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DnsDomain
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListDnsDomainRecordsResponse parses an HTTP response from a ListDnsDomainRecordsWithResponse call
func ParseListDnsDomainRecordsResponse(rsp *http.Response) (*ListDnsDomainRecordsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListDnsDomainRecordsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			DnsDomainRecords *[]DnsDomainRecord `json:"dns-domain-records,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetDnsDomainRecordResponse parses an HTTP response from a GetDnsDomainRecordWithResponse call
func ParseGetDnsDomainRecordResponse(rsp *http.Response) (*GetDnsDomainRecordResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetDnsDomainRecordResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DnsDomainRecord
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListElasticIpsResponse parses an HTTP response from a ListElasticIpsWithResponse call
func ParseListElasticIpsResponse(rsp *http.Response) (*ListElasticIpsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListElasticIpsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			ElasticIps *[]ElasticIp `json:"elastic-ips,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateElasticIpResponse parses an HTTP response from a CreateElasticIpWithResponse call
func ParseCreateElasticIpResponse(rsp *http.Response) (*CreateElasticIpResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreateElasticIpResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeleteElasticIpResponse parses an HTTP response from a DeleteElasticIpWithResponse call
func ParseDeleteElasticIpResponse(rsp *http.Response) (*DeleteElasticIpResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteElasticIpResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetElasticIpResponse parses an HTTP response from a GetElasticIpWithResponse call
func ParseGetElasticIpResponse(rsp *http.Response) (*GetElasticIpResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetElasticIpResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest ElasticIp
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseUpdateElasticIpResponse parses an HTTP response from a UpdateElasticIpWithResponse call
func ParseUpdateElasticIpResponse(rsp *http.Response) (*UpdateElasticIpResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &UpdateElasticIpResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseResetElasticIpFieldResponse parses an HTTP response from a ResetElasticIpFieldWithResponse call
func ParseResetElasticIpFieldResponse(rsp *http.Response) (*ResetElasticIpFieldResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ResetElasticIpFieldResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseAttachInstanceToElasticIpResponse parses an HTTP response from a AttachInstanceToElasticIpWithResponse call
func ParseAttachInstanceToElasticIpResponse(rsp *http.Response) (*AttachInstanceToElasticIpResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &AttachInstanceToElasticIpResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDetachInstanceFromElasticIpResponse parses an HTTP response from a DetachInstanceFromElasticIpWithResponse call
func ParseDetachInstanceFromElasticIpResponse(rsp *http.Response) (*DetachInstanceFromElasticIpResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DetachInstanceFromElasticIpResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListEventsResponse parses an HTTP response from a ListEventsWithResponse call
func ParseListEventsResponse(rsp *http.Response) (*ListEventsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListEventsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []Event
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListInstancesResponse parses an HTTP response from a ListInstancesWithResponse call
func ParseListInstancesResponse(rsp *http.Response) (*ListInstancesResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListInstancesResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Instances *[]Instance `json:"instances,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateInstanceResponse parses an HTTP response from a CreateInstanceWithResponse call
func ParseCreateInstanceResponse(rsp *http.Response) (*CreateInstanceResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreateInstanceResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListInstancePoolsResponse parses an HTTP response from a ListInstancePoolsWithResponse call
func ParseListInstancePoolsResponse(rsp *http.Response) (*ListInstancePoolsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListInstancePoolsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			InstancePools *[]InstancePool `json:"instance-pools,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateInstancePoolResponse parses an HTTP response from a CreateInstancePoolWithResponse call
func ParseCreateInstancePoolResponse(rsp *http.Response) (*CreateInstancePoolResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreateInstancePoolResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeleteInstancePoolResponse parses an HTTP response from a DeleteInstancePoolWithResponse call
func ParseDeleteInstancePoolResponse(rsp *http.Response) (*DeleteInstancePoolResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteInstancePoolResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetInstancePoolResponse parses an HTTP response from a GetInstancePoolWithResponse call
func ParseGetInstancePoolResponse(rsp *http.Response) (*GetInstancePoolResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetInstancePoolResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest InstancePool
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseUpdateInstancePoolResponse parses an HTTP response from a UpdateInstancePoolWithResponse call
func ParseUpdateInstancePoolResponse(rsp *http.Response) (*UpdateInstancePoolResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &UpdateInstancePoolResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseResetInstancePoolFieldResponse parses an HTTP response from a ResetInstancePoolFieldWithResponse call
func ParseResetInstancePoolFieldResponse(rsp *http.Response) (*ResetInstancePoolFieldResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ResetInstancePoolFieldResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseEvictInstancePoolMembersResponse parses an HTTP response from a EvictInstancePoolMembersWithResponse call
func ParseEvictInstancePoolMembersResponse(rsp *http.Response) (*EvictInstancePoolMembersResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &EvictInstancePoolMembersResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseScaleInstancePoolResponse parses an HTTP response from a ScaleInstancePoolWithResponse call
func ParseScaleInstancePoolResponse(rsp *http.Response) (*ScaleInstancePoolResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ScaleInstancePoolResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListInstanceTypesResponse parses an HTTP response from a ListInstanceTypesWithResponse call
func ParseListInstanceTypesResponse(rsp *http.Response) (*ListInstanceTypesResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListInstanceTypesResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			InstanceTypes *[]InstanceType `json:"instance-types,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetInstanceTypeResponse parses an HTTP response from a GetInstanceTypeWithResponse call
func ParseGetInstanceTypeResponse(rsp *http.Response) (*GetInstanceTypeResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetInstanceTypeResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest InstanceType
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeleteInstanceResponse parses an HTTP response from a DeleteInstanceWithResponse call
func ParseDeleteInstanceResponse(rsp *http.Response) (*DeleteInstanceResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteInstanceResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetInstanceResponse parses an HTTP response from a GetInstanceWithResponse call
func ParseGetInstanceResponse(rsp *http.Response) (*GetInstanceResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetInstanceResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Instance
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseUpdateInstanceResponse parses an HTTP response from a UpdateInstanceWithResponse call
func ParseUpdateInstanceResponse(rsp *http.Response) (*UpdateInstanceResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &UpdateInstanceResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseResetInstanceFieldResponse parses an HTTP response from a ResetInstanceFieldWithResponse call
func ParseResetInstanceFieldResponse(rsp *http.Response) (*ResetInstanceFieldResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ResetInstanceFieldResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateSnapshotResponse parses an HTTP response from a CreateSnapshotWithResponse call
func ParseCreateSnapshotResponse(rsp *http.Response) (*CreateSnapshotResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreateSnapshotResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseRebootInstanceResponse parses an HTTP response from a RebootInstanceWithResponse call
func ParseRebootInstanceResponse(rsp *http.Response) (*RebootInstanceResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &RebootInstanceResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseResetInstanceResponse parses an HTTP response from a ResetInstanceWithResponse call
func ParseResetInstanceResponse(rsp *http.Response) (*ResetInstanceResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ResetInstanceResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseResizeInstanceDiskResponse parses an HTTP response from a ResizeInstanceDiskWithResponse call
func ParseResizeInstanceDiskResponse(rsp *http.Response) (*ResizeInstanceDiskResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ResizeInstanceDiskResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseScaleInstanceResponse parses an HTTP response from a ScaleInstanceWithResponse call
func ParseScaleInstanceResponse(rsp *http.Response) (*ScaleInstanceResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ScaleInstanceResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseStartInstanceResponse parses an HTTP response from a StartInstanceWithResponse call
func ParseStartInstanceResponse(rsp *http.Response) (*StartInstanceResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &StartInstanceResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseStopInstanceResponse parses an HTTP response from a StopInstanceWithResponse call
func ParseStopInstanceResponse(rsp *http.Response) (*StopInstanceResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &StopInstanceResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseRevertInstanceToSnapshotResponse parses an HTTP response from a RevertInstanceToSnapshotWithResponse call
func ParseRevertInstanceToSnapshotResponse(rsp *http.Response) (*RevertInstanceToSnapshotResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &RevertInstanceToSnapshotResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListLoadBalancersResponse parses an HTTP response from a ListLoadBalancersWithResponse call
func ParseListLoadBalancersResponse(rsp *http.Response) (*ListLoadBalancersResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListLoadBalancersResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			LoadBalancers *[]LoadBalancer `json:"load-balancers,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateLoadBalancerResponse parses an HTTP response from a CreateLoadBalancerWithResponse call
func ParseCreateLoadBalancerResponse(rsp *http.Response) (*CreateLoadBalancerResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreateLoadBalancerResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeleteLoadBalancerResponse parses an HTTP response from a DeleteLoadBalancerWithResponse call
func ParseDeleteLoadBalancerResponse(rsp *http.Response) (*DeleteLoadBalancerResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteLoadBalancerResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetLoadBalancerResponse parses an HTTP response from a GetLoadBalancerWithResponse call
func ParseGetLoadBalancerResponse(rsp *http.Response) (*GetLoadBalancerResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetLoadBalancerResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest LoadBalancer
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseUpdateLoadBalancerResponse parses an HTTP response from a UpdateLoadBalancerWithResponse call
func ParseUpdateLoadBalancerResponse(rsp *http.Response) (*UpdateLoadBalancerResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &UpdateLoadBalancerResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseAddServiceToLoadBalancerResponse parses an HTTP response from a AddServiceToLoadBalancerWithResponse call
func ParseAddServiceToLoadBalancerResponse(rsp *http.Response) (*AddServiceToLoadBalancerResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &AddServiceToLoadBalancerResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeleteLoadBalancerServiceResponse parses an HTTP response from a DeleteLoadBalancerServiceWithResponse call
func ParseDeleteLoadBalancerServiceResponse(rsp *http.Response) (*DeleteLoadBalancerServiceResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteLoadBalancerServiceResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetLoadBalancerServiceResponse parses an HTTP response from a GetLoadBalancerServiceWithResponse call
func ParseGetLoadBalancerServiceResponse(rsp *http.Response) (*GetLoadBalancerServiceResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetLoadBalancerServiceResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest LoadBalancerService
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseUpdateLoadBalancerServiceResponse parses an HTTP response from a UpdateLoadBalancerServiceWithResponse call
func ParseUpdateLoadBalancerServiceResponse(rsp *http.Response) (*UpdateLoadBalancerServiceResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &UpdateLoadBalancerServiceResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseResetLoadBalancerServiceFieldResponse parses an HTTP response from a ResetLoadBalancerServiceFieldWithResponse call
func ParseResetLoadBalancerServiceFieldResponse(rsp *http.Response) (*ResetLoadBalancerServiceFieldResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ResetLoadBalancerServiceFieldResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseResetLoadBalancerFieldResponse parses an HTTP response from a ResetLoadBalancerFieldWithResponse call
func ParseResetLoadBalancerFieldResponse(rsp *http.Response) (*ResetLoadBalancerFieldResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ResetLoadBalancerFieldResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetOperationResponse parses an HTTP response from a GetOperationWithResponse call
func ParseGetOperationResponse(rsp *http.Response) (*GetOperationResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetOperationResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListPrivateNetworksResponse parses an HTTP response from a ListPrivateNetworksWithResponse call
func ParseListPrivateNetworksResponse(rsp *http.Response) (*ListPrivateNetworksResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListPrivateNetworksResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			PrivateNetworks *[]PrivateNetwork `json:"private-networks,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreatePrivateNetworkResponse parses an HTTP response from a CreatePrivateNetworkWithResponse call
func ParseCreatePrivateNetworkResponse(rsp *http.Response) (*CreatePrivateNetworkResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreatePrivateNetworkResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeletePrivateNetworkResponse parses an HTTP response from a DeletePrivateNetworkWithResponse call
func ParseDeletePrivateNetworkResponse(rsp *http.Response) (*DeletePrivateNetworkResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeletePrivateNetworkResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetPrivateNetworkResponse parses an HTTP response from a GetPrivateNetworkWithResponse call
func ParseGetPrivateNetworkResponse(rsp *http.Response) (*GetPrivateNetworkResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetPrivateNetworkResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest PrivateNetwork
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseUpdatePrivateNetworkResponse parses an HTTP response from a UpdatePrivateNetworkWithResponse call
func ParseUpdatePrivateNetworkResponse(rsp *http.Response) (*UpdatePrivateNetworkResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &UpdatePrivateNetworkResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseResetPrivateNetworkFieldResponse parses an HTTP response from a ResetPrivateNetworkFieldWithResponse call
func ParseResetPrivateNetworkFieldResponse(rsp *http.Response) (*ResetPrivateNetworkFieldResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ResetPrivateNetworkFieldResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseAttachInstanceToPrivateNetworkResponse parses an HTTP response from a AttachInstanceToPrivateNetworkWithResponse call
func ParseAttachInstanceToPrivateNetworkResponse(rsp *http.Response) (*AttachInstanceToPrivateNetworkResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &AttachInstanceToPrivateNetworkResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDetachInstanceFromPrivateNetworkResponse parses an HTTP response from a DetachInstanceFromPrivateNetworkWithResponse call
func ParseDetachInstanceFromPrivateNetworkResponse(rsp *http.Response) (*DetachInstanceFromPrivateNetworkResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DetachInstanceFromPrivateNetworkResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseUpdatePrivateNetworkInstanceIpResponse parses an HTTP response from a UpdatePrivateNetworkInstanceIpWithResponse call
func ParseUpdatePrivateNetworkInstanceIpResponse(rsp *http.Response) (*UpdatePrivateNetworkInstanceIpResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &UpdatePrivateNetworkInstanceIpResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListQuotasResponse parses an HTTP response from a ListQuotasWithResponse call
func ParseListQuotasResponse(rsp *http.Response) (*ListQuotasResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListQuotasResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Quotas *[]Quota `json:"quotas,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetQuotaResponse parses an HTTP response from a GetQuotaWithResponse call
func ParseGetQuotaResponse(rsp *http.Response) (*GetQuotaResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetQuotaResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Quota
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListSecurityGroupsResponse parses an HTTP response from a ListSecurityGroupsWithResponse call
func ParseListSecurityGroupsResponse(rsp *http.Response) (*ListSecurityGroupsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListSecurityGroupsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			SecurityGroups *[]SecurityGroup `json:"security-groups,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateSecurityGroupResponse parses an HTTP response from a CreateSecurityGroupWithResponse call
func ParseCreateSecurityGroupResponse(rsp *http.Response) (*CreateSecurityGroupResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreateSecurityGroupResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeleteSecurityGroupResponse parses an HTTP response from a DeleteSecurityGroupWithResponse call
func ParseDeleteSecurityGroupResponse(rsp *http.Response) (*DeleteSecurityGroupResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteSecurityGroupResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetSecurityGroupResponse parses an HTTP response from a GetSecurityGroupWithResponse call
func ParseGetSecurityGroupResponse(rsp *http.Response) (*GetSecurityGroupResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetSecurityGroupResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest SecurityGroup
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseAddRuleToSecurityGroupResponse parses an HTTP response from a AddRuleToSecurityGroupWithResponse call
func ParseAddRuleToSecurityGroupResponse(rsp *http.Response) (*AddRuleToSecurityGroupResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &AddRuleToSecurityGroupResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeleteRuleFromSecurityGroupResponse parses an HTTP response from a DeleteRuleFromSecurityGroupWithResponse call
func ParseDeleteRuleFromSecurityGroupResponse(rsp *http.Response) (*DeleteRuleFromSecurityGroupResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteRuleFromSecurityGroupResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseAddExternalSourceToSecurityGroupResponse parses an HTTP response from a AddExternalSourceToSecurityGroupWithResponse call
func ParseAddExternalSourceToSecurityGroupResponse(rsp *http.Response) (*AddExternalSourceToSecurityGroupResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &AddExternalSourceToSecurityGroupResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseAttachInstanceToSecurityGroupResponse parses an HTTP response from a AttachInstanceToSecurityGroupWithResponse call
func ParseAttachInstanceToSecurityGroupResponse(rsp *http.Response) (*AttachInstanceToSecurityGroupResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &AttachInstanceToSecurityGroupResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDetachInstanceFromSecurityGroupResponse parses an HTTP response from a DetachInstanceFromSecurityGroupWithResponse call
func ParseDetachInstanceFromSecurityGroupResponse(rsp *http.Response) (*DetachInstanceFromSecurityGroupResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DetachInstanceFromSecurityGroupResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseRemoveExternalSourceFromSecurityGroupResponse parses an HTTP response from a RemoveExternalSourceFromSecurityGroupWithResponse call
func ParseRemoveExternalSourceFromSecurityGroupResponse(rsp *http.Response) (*RemoveExternalSourceFromSecurityGroupResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &RemoveExternalSourceFromSecurityGroupResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListSksClustersResponse parses an HTTP response from a ListSksClustersWithResponse call
func ParseListSksClustersResponse(rsp *http.Response) (*ListSksClustersResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListSksClustersResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			SksClusters *[]SksCluster `json:"sks-clusters,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateSksClusterResponse parses an HTTP response from a CreateSksClusterWithResponse call
func ParseCreateSksClusterResponse(rsp *http.Response) (*CreateSksClusterResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreateSksClusterResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListSksClusterDeprecatedResourcesResponse parses an HTTP response from a ListSksClusterDeprecatedResourcesWithResponse call
func ParseListSksClusterDeprecatedResourcesResponse(rsp *http.Response) (*ListSksClusterDeprecatedResourcesResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListSksClusterDeprecatedResourcesResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []SksClusterDeprecatedResource
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGenerateSksClusterKubeconfigResponse parses an HTTP response from a GenerateSksClusterKubeconfigWithResponse call
func ParseGenerateSksClusterKubeconfigResponse(rsp *http.Response) (*GenerateSksClusterKubeconfigResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GenerateSksClusterKubeconfigResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Kubeconfig *string `json:"kubeconfig,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListSksClusterVersionsResponse parses an HTTP response from a ListSksClusterVersionsWithResponse call
func ParseListSksClusterVersionsResponse(rsp *http.Response) (*ListSksClusterVersionsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListSksClusterVersionsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			SksClusterVersions *[]string `json:"sks-cluster-versions,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeleteSksClusterResponse parses an HTTP response from a DeleteSksClusterWithResponse call
func ParseDeleteSksClusterResponse(rsp *http.Response) (*DeleteSksClusterResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteSksClusterResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetSksClusterResponse parses an HTTP response from a GetSksClusterWithResponse call
func ParseGetSksClusterResponse(rsp *http.Response) (*GetSksClusterResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetSksClusterResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest SksCluster
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseUpdateSksClusterResponse parses an HTTP response from a UpdateSksClusterWithResponse call
func ParseUpdateSksClusterResponse(rsp *http.Response) (*UpdateSksClusterResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &UpdateSksClusterResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetSksClusterAuthorityCertResponse parses an HTTP response from a GetSksClusterAuthorityCertWithResponse call
func ParseGetSksClusterAuthorityCertResponse(rsp *http.Response) (*GetSksClusterAuthorityCertResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetSksClusterAuthorityCertResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Cacert *string `json:"cacert,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCreateSksNodepoolResponse parses an HTTP response from a CreateSksNodepoolWithResponse call
func ParseCreateSksNodepoolResponse(rsp *http.Response) (*CreateSksNodepoolResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CreateSksNodepoolResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeleteSksNodepoolResponse parses an HTTP response from a DeleteSksNodepoolWithResponse call
func ParseDeleteSksNodepoolResponse(rsp *http.Response) (*DeleteSksNodepoolResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteSksNodepoolResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetSksNodepoolResponse parses an HTTP response from a GetSksNodepoolWithResponse call
func ParseGetSksNodepoolResponse(rsp *http.Response) (*GetSksNodepoolResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetSksNodepoolResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest SksNodepool
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseUpdateSksNodepoolResponse parses an HTTP response from a UpdateSksNodepoolWithResponse call
func ParseUpdateSksNodepoolResponse(rsp *http.Response) (*UpdateSksNodepoolResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &UpdateSksNodepoolResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseResetSksNodepoolFieldResponse parses an HTTP response from a ResetSksNodepoolFieldWithResponse call
func ParseResetSksNodepoolFieldResponse(rsp *http.Response) (*ResetSksNodepoolFieldResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ResetSksNodepoolFieldResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseEvictSksNodepoolMembersResponse parses an HTTP response from a EvictSksNodepoolMembersWithResponse call
func ParseEvictSksNodepoolMembersResponse(rsp *http.Response) (*EvictSksNodepoolMembersResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &EvictSksNodepoolMembersResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseScaleSksNodepoolResponse parses an HTTP response from a ScaleSksNodepoolWithResponse call
func ParseScaleSksNodepoolResponse(rsp *http.Response) (*ScaleSksNodepoolResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ScaleSksNodepoolResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseRotateSksCcmCredentialsResponse parses an HTTP response from a RotateSksCcmCredentialsWithResponse call
func ParseRotateSksCcmCredentialsResponse(rsp *http.Response) (*RotateSksCcmCredentialsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &RotateSksCcmCredentialsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseRotateSksOperatorsCaResponse parses an HTTP response from a RotateSksOperatorsCaWithResponse call
func ParseRotateSksOperatorsCaResponse(rsp *http.Response) (*RotateSksOperatorsCaResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &RotateSksOperatorsCaResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseUpgradeSksClusterResponse parses an HTTP response from a UpgradeSksClusterWithResponse call
func ParseUpgradeSksClusterResponse(rsp *http.Response) (*UpgradeSksClusterResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &UpgradeSksClusterResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseUpgradeSksClusterServiceLevelResponse parses an HTTP response from a UpgradeSksClusterServiceLevelWithResponse call
func ParseUpgradeSksClusterServiceLevelResponse(rsp *http.Response) (*UpgradeSksClusterServiceLevelResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &UpgradeSksClusterServiceLevelResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseResetSksClusterFieldResponse parses an HTTP response from a ResetSksClusterFieldWithResponse call
func ParseResetSksClusterFieldResponse(rsp *http.Response) (*ResetSksClusterFieldResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ResetSksClusterFieldResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListSnapshotsResponse parses an HTTP response from a ListSnapshotsWithResponse call
func ParseListSnapshotsResponse(rsp *http.Response) (*ListSnapshotsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListSnapshotsResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Snapshots *[]Snapshot `json:"snapshots,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeleteSnapshotResponse parses an HTTP response from a DeleteSnapshotWithResponse call
func ParseDeleteSnapshotResponse(rsp *http.Response) (*DeleteSnapshotResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteSnapshotResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetSnapshotResponse parses an HTTP response from a GetSnapshotWithResponse call
func ParseGetSnapshotResponse(rsp *http.Response) (*GetSnapshotResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetSnapshotResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Snapshot
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseExportSnapshotResponse parses an HTTP response from a ExportSnapshotWithResponse call
func ParseExportSnapshotResponse(rsp *http.Response) (*ExportSnapshotResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ExportSnapshotResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParsePromoteSnapshotToTemplateResponse parses an HTTP response from a PromoteSnapshotToTemplateWithResponse call
func ParsePromoteSnapshotToTemplateResponse(rsp *http.Response) (*PromoteSnapshotToTemplateResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PromoteSnapshotToTemplateResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetSosPresignedUrlResponse parses an HTTP response from a GetSosPresignedUrlWithResponse call
func ParseGetSosPresignedUrlResponse(rsp *http.Response) (*GetSosPresignedUrlResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetSosPresignedUrlResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Url *string `json:"url,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListSshKeysResponse parses an HTTP response from a ListSshKeysWithResponse call
func ParseListSshKeysResponse(rsp *http.Response) (*ListSshKeysResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListSshKeysResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			SshKeys *[]SshKey `json:"ssh-keys,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseRegisterSshKeyResponse parses an HTTP response from a RegisterSshKeyWithResponse call
func ParseRegisterSshKeyResponse(rsp *http.Response) (*RegisterSshKeyResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &RegisterSshKeyResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeleteSshKeyResponse parses an HTTP response from a DeleteSshKeyWithResponse call
func ParseDeleteSshKeyResponse(rsp *http.Response) (*DeleteSshKeyResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteSshKeyResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetSshKeyResponse parses an HTTP response from a GetSshKeyWithResponse call
func ParseGetSshKeyResponse(rsp *http.Response) (*GetSshKeyResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetSshKeyResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest SshKey
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListTemplatesResponse parses an HTTP response from a ListTemplatesWithResponse call
func ParseListTemplatesResponse(rsp *http.Response) (*ListTemplatesResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListTemplatesResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Templates *[]Template `json:"templates,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseRegisterTemplateResponse parses an HTTP response from a RegisterTemplateWithResponse call
func ParseRegisterTemplateResponse(rsp *http.Response) (*RegisterTemplateResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &RegisterTemplateResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDeleteTemplateResponse parses an HTTP response from a DeleteTemplateWithResponse call
func ParseDeleteTemplateResponse(rsp *http.Response) (*DeleteTemplateResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DeleteTemplateResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseGetTemplateResponse parses an HTTP response from a GetTemplateWithResponse call
func ParseGetTemplateResponse(rsp *http.Response) (*GetTemplateResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetTemplateResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Template
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseCopyTemplateResponse parses an HTTP response from a CopyTemplateWithResponse call
func ParseCopyTemplateResponse(rsp *http.Response) (*CopyTemplateResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &CopyTemplateResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseUpdateTemplateResponse parses an HTTP response from a UpdateTemplateWithResponse call
func ParseUpdateTemplateResponse(rsp *http.Response) (*UpdateTemplateResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &UpdateTemplateResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Operation
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseListZonesResponse parses an HTTP response from a ListZonesWithResponse call
func ParseListZonesResponse(rsp *http.Response) (*ListZonesResponse, error) {
	bodyBytes, err := ioutil.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ListZonesResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest struct {
			Zones *[]Zone `json:"zones,omitempty"`
		}
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}
