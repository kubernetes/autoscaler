/*
Copyright 2019 The Kubernetes Authors.

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

package verda

import (
	"encoding/json"
	"strconv"
	"time"
)

// InstanceCPU represents CPU information
type InstanceCPU struct {
	Description   string `json:"description"`
	NumberOfCores int    `json:"number_of_cores"`
}

// InstanceGPU represents GPU information
type InstanceGPU struct {
	Description  string `json:"description"`
	NumberOfGPUs int    `json:"number_of_gpus"`
}

// InstanceMemory represents memory information
type InstanceMemory struct {
	Description     string `json:"description"`
	SizeInGigabytes int    `json:"size_in_gigabytes"`
}

// InstanceStorage represents storage information
type InstanceStorage struct {
	Description string `json:"description"`
}

// Instance represents a Verda instance
type Instance struct {
	ID              string          `json:"id"`
	InstanceType    string          `json:"instance_type"`
	Image           string          `json:"image"`
	PricePerHour    FlexibleFloat   `json:"price_per_hour"`
	Hostname        string          `json:"hostname"`
	Description     string          `json:"description"`
	IP              *string         `json:"ip"`
	Status          string          `json:"status"`
	CreatedAt       time.Time       `json:"created_at"`
	SSHKeyIDs       []string        `json:"ssh_key_ids"`
	CPU             InstanceCPU     `json:"cpu"`
	GPU             InstanceGPU     `json:"gpu"`
	Memory          InstanceMemory  `json:"memory"`
	Storage         InstanceStorage `json:"storage"`
	OSVolumeID      *string         `json:"os_volume_id"`
	GPUMemory       InstanceMemory  `json:"gpu_memory"`
	Location        string          `json:"location"`
	IsSpot          bool            `json:"is_spot"`
	OSName          string          `json:"os_name"`
	StartupScriptID *string         `json:"startup_script_id"`
	JupyterToken    *string         `json:"jupyter_token"`
	Contract        string          `json:"contract"`
	Pricing         string          `json:"pricing"`
}

// CreateInstanceRequest represents the request to create an instance
type CreateInstanceRequest struct {
	InstanceType    string                 `json:"instance_type"`
	Image           string                 `json:"image"`
	Hostname        string                 `json:"hostname"`
	Description     string                 `json:"description"`
	SSHKeyIDs       []string               `json:"ssh_key_ids,omitempty"`
	LocationCode    string                 `json:"location_code,omitempty"`
	Contract        string                 `json:"contract,omitempty"`
	Pricing         string                 `json:"pricing,omitempty"`
	StartupScriptID *string                `json:"startup_script_id,omitempty"`
	Volumes         []VolumeCreateRequest  `json:"volumes,omitempty"`
	ExistingVolumes []string               `json:"existing_volumes,omitempty"`
	OSVolume        *OSVolumeCreateRequest `json:"os_volume,omitempty"`
	IsSpot          bool                   `json:"is_spot,omitempty"`
	Coupon          *string                `json:"coupon,omitempty"`
}

// VolumeCreateRequest represents a volume to be created
type VolumeCreateRequest struct {
	Size         int    `json:"size"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	LocationCode string `json:"location_code,omitempty"`
}

// VolumeAttachRequest represents a request to attach a volume to an instance
type VolumeAttachRequest struct {
	InstanceID string `json:"instance_id"`
}

// VolumeDetachRequest represents a request to detach a volume from an instance
type VolumeDetachRequest struct {
	InstanceID string `json:"instance_id"`
}

// VolumeCloneRequest represents a request to clone a volume
type VolumeCloneRequest struct {
	Name         string `json:"name"`
	LocationCode string `json:"location_code,omitempty"`
}

// VolumeResizeRequest represents a request to resize a volume
type VolumeResizeRequest struct {
	Size int `json:"size"`
}

// VolumeRenameRequest represents a request to rename a volume
type VolumeRenameRequest struct {
	Name string `json:"name"`
}

// OSVolumeCreateRequest represents OS volume configuration
type OSVolumeCreateRequest struct {
	Name string `json:"name"`
	Size int    `json:"size"`
}

// InstanceActionRequest represents an action to perform on instances
type InstanceActionRequest struct {
	Action    string   `json:"action"`
	ID        []string `json:"id"`
	VolumeIDs []string `json:"volume_ids,omitempty"`
}

// InstanceAvailability represents instance availability information
type InstanceAvailability struct {
	LocationCode   string   `json:"location_code"`
	Availabilities []string `json:"availabilities"`
}

// LocationAvailability represents instance type availability by location code
type LocationAvailability struct {
	LocationCode   string   `json:"location_code"`
	Availabilities []string `json:"availabilities"`
}

// Volume represents a Verda volume
type Volume struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Size       int       `json:"size"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	InstanceID *string   `json:"instance_id"`
}

// VolumeType represents available volume type specifications
type VolumeType struct {
	Type                 string          `json:"type"`
	Price                VolumeTypePrice `json:"price"`
	IsSharedFS           bool            `json:"is_shared_fs"`
	BurstBandwidth       float64         `json:"burst_bandwidth"`
	ContinuousBandwidth  float64         `json:"continuous_bandwidth"`
	InternalNetworkSpeed float64         `json:"internal_network_speed"`
	IOPS                 string          `json:"iops"`
}

// VolumeTypePrice represents the pricing structure for a volume type
type VolumeTypePrice struct {
	MonthlyPerGB float64 `json:"monthly_per_gb"`
	Currency     string  `json:"currency"`
}

// SSHKey represents an SSH key
type SSHKey struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	PublicKey   string    `json:"key"`
	Fingerprint string    `json:"fingerprint"`
	CreatedAt   time.Time `json:"created_at"`
}

// StartupScript represents a startup script
type StartupScript struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Script    string    `json:"script"`
	CreatedAt time.Time `json:"created_at"`
}

// Location represents a datacenter location
type Location struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	CountryCode string `json:"country_code"`
}

// Balance represents account balance information
type Balance struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

// Image represents an OS image for instances
type Image struct {
	ID        string   `json:"id"`
	ImageType string   `json:"image_type"`
	Name      string   `json:"name"`
	IsDefault bool     `json:"is_default"`
	IsCluster bool     `json:"is_cluster"`
	Details   []string `json:"details"`
	Category  string   `json:"category"`
}

// ContainerType represents a serverless container compute resource option
type ContainerType struct {
	ID                  string         `json:"id"`
	Model               string         `json:"model"`
	Name                string         `json:"name"`
	InstanceType        string         `json:"instance_type"`
	CPU                 InstanceCPU    `json:"cpu"`
	GPU                 InstanceGPU    `json:"gpu"`
	GPUMemory           InstanceMemory `json:"gpu_memory"`
	Memory              InstanceMemory `json:"memory"`
	ServerlessPrice     FlexibleFloat  `json:"serverless_price"`
	ServerlessSpotPrice FlexibleFloat  `json:"serverless_spot_price"`
	Currency            string         `json:"currency"`
	Manufacturer        string         `json:"manufacturer"`
}

// InstanceTypeInfo represents detailed instance type information with pricing
type InstanceTypeInfo struct {
	ID              string          `json:"id"`
	InstanceType    string          `json:"instance_type"`
	Model           string          `json:"model"`
	Name            string          `json:"name"`
	CPU             InstanceCPU     `json:"cpu"`
	GPU             InstanceGPU     `json:"gpu"`
	GPUMemory       InstanceMemory  `json:"gpu_memory"`
	Memory          InstanceMemory  `json:"memory"`
	PricePerHour    FlexibleFloat   `json:"price_per_hour"`
	SpotPrice       FlexibleFloat   `json:"spot_price"`
	DynamicPrice    FlexibleFloat   `json:"dynamic_price"`
	MaxDynamicPrice FlexibleFloat   `json:"max_dynamic_price"`
	Storage         InstanceStorage `json:"storage"`
	Currency        string          `json:"currency"`
	Manufacturer    string          `json:"manufacturer"`
	BestFor         []string        `json:"best_for"`
	Description     string          `json:"description"`
}

// PriceHistoryRecord represents a single price record in the price history
type PriceHistoryRecord struct {
	Date                string        `json:"date"`
	FixedPricePerHour   FlexibleFloat `json:"fixed_price_per_hour"`
	DynamicPricePerHour FlexibleFloat `json:"dynamic_price_per_hour"`
	Currency            string        `json:"currency"`
}

// InstanceTypePriceHistory maps instance type names to their price history records
type InstanceTypePriceHistory map[string][]PriceHistoryRecord

// LongTermPeriod represents a long-term rental period option
type LongTermPeriod struct {
	Code               string  `json:"code"`
	Name               string  `json:"name"`
	IsEnabled          bool    `json:"is_enabled"`
	UnitName           string  `json:"unit_name"`
	UnitValue          int     `json:"unit_value"`
	DiscountPercentage float64 `json:"discount_percentage"`
}

// Action constants
const (
	ActionBoot          = "boot"
	ActionStart         = "start"
	ActionShutdown      = "shutdown"
	ActionDelete        = "delete"
	ActionDiscontinue   = "discontinue"
	ActionHibernate     = "hibernate"
	ActionConfigureSpot = "configure_spot"
	ActionForceShutdown = "force_shutdown"
	ActionDeleteStuck   = "delete_stuck"
	ActionDeploy        = "deploy"
	ActionTransfer      = "transfer"
)

// Instance status constants
const (
	StatusNew          = "new"
	StatusOrdered      = "ordered"
	StatusProvisioning = "provisioning"
	StatusValidating   = "validating"
	StatusRunning      = "running"
	StatusOffline      = "offline"
	StatusPending      = "pending"
	StatusDiscontinued = "discontinued"
	StatusUnknown      = "unknown"
	StatusNotFound     = "notfound"
	StatusError        = "error"
	StatusDeleting     = "deleting"
	StatusNoCapacity   = "no_capacity"
)

// Contract type constants
const (
	// ContractTypePayAsYouGo represents pay-as-you-go billing.
	ContractTypePayAsYouGo = "PAY_AS_YOU_GO"
	// ContractTypeLongTerm represents long-term contract billing.
	ContractTypeLongTerm = "LONG_TERM"
	ContractTypeSPOT     = "SPOT"
)

// Pricing type constants
const (
	PricingTypeFIXED   = "FIXED_PRICE"
	PricingTypeDYNAMIC = "DYNAMIC" // deprecated
)

// Default location
const (
	LocationFIN01 = "FIN-01"
)

// Volume type constants
const (
	VolumeTypeHDD               = "HDD"
	VolumeTypeNVMe              = "NVMe"
	VolumeTypeHDDShared         = "HDD_Shared"
	VolumeTypeNVMeShared        = "NVMe_Shared"
	VolumeTypeNVMeLocalStorage  = "NVMe_Local_Storage"
	VolumeTypeNVMeSharedCluster = "NVMe_Shared_Cluster"
	VolumeTypeNVMeOSCluster     = "NVMe_OS_Cluster"
)

// Volume status constants - these match the actual API values
const (
	VolumeStatusOrdered   = "ordered"
	VolumeStatusAttached  = "attached"
	VolumeStatusAttaching = "attaching"
	VolumeStatusDetached  = "detached"
	VolumeStatusDeleted   = "deleted"
	VolumeStatusCloning   = "cloning"
	VolumeStatusDetaching = "detaching"
	VolumeStatusDeleting  = "deleting"
	VolumeStatusRestoring = "restoring"
	VolumeStatusCreated   = "created"
	VolumeStatusExported  = "exported"
	VolumeStatusCanceled  = "canceled"
	VolumeStatusCanceling = "canceling"
)

// Cluster represents a Verda cluster
type Cluster struct {
	ID              string          `json:"id"`
	ClusterType     string          `json:"cluster_type"`
	Image           string          `json:"image"`
	PricePerHour    FlexibleFloat   `json:"price_per_hour"`
	Hostname        string          `json:"hostname"`
	Description     string          `json:"description"`
	IP              *string         `json:"ip"`
	Status          string          `json:"status"`
	CreatedAt       time.Time       `json:"created_at"`
	SSHKeyIDs       []string        `json:"ssh_key_ids"`
	CPU             InstanceCPU     `json:"cpu"`
	GPU             InstanceGPU     `json:"gpu"`
	Memory          InstanceMemory  `json:"memory"`
	Storage         InstanceStorage `json:"storage"`
	GPUMemory       InstanceMemory  `json:"gpu_memory"`
	Location        string          `json:"location"`
	OSName          string          `json:"os_name"`
	StartupScriptID *string         `json:"startup_script_id"`
	Contract        string          `json:"contract"`
	Pricing         string          `json:"pricing"`
}

// CreateClusterRequest represents the request to create a cluster
type CreateClusterRequest struct {
	ClusterType     string   `json:"cluster_type"`
	Image           string   `json:"image"`
	Hostname        string   `json:"hostname"`
	Description     string   `json:"description,omitempty"`
	SSHKeyIDs       []string `json:"ssh_key_ids"`
	LocationCode    string   `json:"location_code,omitempty"`
	Contract        string   `json:"contract,omitempty"`
	Pricing         string   `json:"pricing,omitempty"`
	StartupScriptID *string  `json:"startup_script_id,omitempty"`
	SharedVolumes   []string `json:"shared_volumes,omitempty"`
	ExistingVolumes []string `json:"existing_volumes,omitempty"`
	Coupon          *string  `json:"coupon,omitempty"`
}

// CreateClusterResponse represents the response from creating a cluster
type CreateClusterResponse struct {
	ID string `json:"id"`
}

// ClusterActionRequest represents an action to perform on clusters
type ClusterActionRequest struct {
	IDList any    `json:"id_list"` // Can be string or []string
	Action string `json:"action"`
}

// ClusterAvailability represents cluster availability information
type ClusterAvailability struct {
	ClusterType  string `json:"cluster_type"`
	LocationCode string `json:"location_code"`
	Available    bool   `json:"available"`
}

// ClusterType represents a cluster configuration type
type ClusterType struct {
	ClusterType  string          `json:"cluster_type"`
	Description  string          `json:"description"`
	PricePerHour FlexibleFloat   `json:"price_per_hour"`
	CPU          InstanceCPU     `json:"cpu"`
	GPU          InstanceGPU     `json:"gpu"`
	Memory       InstanceMemory  `json:"memory"`
	Storage      InstanceStorage `json:"storage"`
	GPUMemory    InstanceMemory  `json:"gpu_memory"`
	Manufacturer string          `json:"manufacturer"`
	Available    bool            `json:"available"`
}

// ClusterImage represents an OS image for clusters
type ClusterImage struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Available   bool   `json:"available"`
}

// Cluster action constants
const (
	ClusterActionDiscontinue = "discontinue"
)

// Container Deployment Types

// ContainerDeployment represents a serverless container deployment
type ContainerDeployment struct {
	Name                      string                     `json:"name"`
	Containers                []DeploymentContainer      `json:"containers"`
	EndpointBaseURL           string                     `json:"endpoint_base_url"`
	CreatedAt                 time.Time                  `json:"created_at"`
	Compute                   *ContainerCompute          `json:"compute,omitempty"`
	ContainerRegistrySettings *ContainerRegistrySettings `json:"container_registry_settings,omitempty"`
	IsSpot                    bool                       `json:"is_spot"`
}

// TargetNode represents the compute node/GPU configuration
type TargetNode struct {
	Name string `json:"name"` // e.g., "RTX 4500 Ada", "H100"
	Size int    `json:"size"` // Number of GPUs
}

// ContainerRegistrySettings represents registry authentication settings
type ContainerRegistrySettings struct {
	IsPrivate   bool                    `json:"is_private"`
	Credentials *RegistryCredentialsRef `json:"credentials,omitempty"`
}

// RegistryCredentialsRef references registry credentials by name
type RegistryCredentialsRef struct {
	Name string `json:"name"`
}

// DeploymentContainer represents a container configuration in a deployment response
type DeploymentContainer struct {
	Image               ContainerImage                `json:"image"`
	Name                string                        `json:"name,omitempty"`
	ExposedPort         int                           `json:"exposed_port,omitempty"`
	Healthcheck         *ContainerHealthcheck         `json:"healthcheck,omitempty"`
	EntrypointOverrides *ContainerEntrypointOverrides `json:"entrypoint_overrides,omitempty"`
	Env                 []ContainerEnvVar             `json:"env,omitempty"`
	VolumeMounts        []ContainerVolumeMount        `json:"volume_mounts,omitempty"`
	AutoUpdate          *ContainerAutoUpdate          `json:"autoupdate,omitempty"`
}

// ContainerImage represents a container image reference.
type ContainerImage struct {
	Image         string    `json:"image"`
	LastUpdatedAt time.Time `json:"last_updated_at,omitempty"`
}

// ContainerEnvVar represents an environment variable with type
type ContainerEnvVar struct {
	Type                     string `json:"type"` // "plain" or "secret"
	Name                     string `json:"name"`
	ValueOrReferenceToSecret string `json:"value_or_reference_to_secret,omitempty"`
}

// DeploymentScalingOptions represents scaling configuration for container deployment
type DeploymentScalingOptions struct {
	DeadlineSeconds        int `json:"deadline_seconds,omitempty"`
	MaxReplicaCount        int `json:"max_replica_count"`
	QueueMessageTTLSeconds int `json:"queue_message_ttl_seconds,omitempty"`
}

// ContainerCompute represents compute resources for deployments
type ContainerCompute struct {
	Name string `json:"name"` // e.g., "H100", "A100"
	Size int    `json:"size"` // Number of GPUs
}

// CreateDeploymentRequest represents a request to create a new deployment
type CreateDeploymentRequest struct {
	Name                      string                      `json:"name"`
	IsSpot                    bool                        `json:"is_spot"`
	Compute                   ContainerCompute            `json:"compute"`
	ContainerRegistrySettings ContainerRegistrySettings   `json:"container_registry_settings"`
	Scaling                   ContainerScalingOptions     `json:"scaling"`
	Containers                []CreateDeploymentContainer `json:"containers"`
}

// CreateDeploymentContainer represents a container configuration for create/update requests
// Note: In requests, image is a string; in responses, image is an object
type CreateDeploymentContainer struct {
	Image               string                        `json:"image"`
	ExposedPort         int                           `json:"exposed_port,omitempty"`
	Healthcheck         *ContainerHealthcheck         `json:"healthcheck,omitempty"`
	EntrypointOverrides *ContainerEntrypointOverrides `json:"entrypoint_overrides,omitempty"`
	Env                 []ContainerEnvVar             `json:"env,omitempty"`
	VolumeMounts        []ContainerVolumeMount        `json:"volume_mounts,omitempty"`
	AutoUpdate          *ContainerAutoUpdate          `json:"autoupdate,omitempty"`
}

// UpdateDeploymentRequest represents a request to update a deployment
type UpdateDeploymentRequest struct {
	IsSpot                    *bool                       `json:"is_spot,omitempty"`
	Compute                   *ContainerCompute           `json:"compute,omitempty"`
	ContainerRegistrySettings *ContainerRegistrySettings  `json:"container_registry_settings,omitempty"`
	Scaling                   *ContainerScalingOptions    `json:"scaling,omitempty"`
	Containers                []CreateDeploymentContainer `json:"containers,omitempty"`
}

// DeploymentStatus represents the status of a deployment
type DeploymentStatus struct {
	Status            string `json:"status"`
	DesiredReplicas   int    `json:"desired_replicas,omitempty"`
	CurrentReplicas   int    `json:"current_replicas,omitempty"`
	AvailableReplicas int    `json:"available_replicas,omitempty"`
	UpdatedAt         string `json:"updated_at,omitempty"`
}

// ContainerScalingOptions represents scaling configuration for container deployments
// Used by both GET and PATCH /container-deployments/{name}/scaling
type ContainerScalingOptions struct {
	MinReplicaCount              int              `json:"min_replica_count,omitempty"`
	MaxReplicaCount              int              `json:"max_replica_count,omitempty"`
	ScaleDownPolicy              *ScalingPolicy   `json:"scale_down_policy,omitempty"`
	ScaleUpPolicy                *ScalingPolicy   `json:"scale_up_policy,omitempty"`
	QueueMessageTTLSeconds       int              `json:"queue_message_ttl_seconds,omitempty"`
	ConcurrentRequestsPerReplica int              `json:"concurrent_requests_per_replica,omitempty"`
	ScalingTriggers              *ScalingTriggers `json:"scaling_triggers,omitempty"`
}

// ScalingPolicy represents scale up/down policy configuration
type ScalingPolicy struct {
	DelaySeconds int `json:"delay_seconds,omitempty"`
}

// ScalingTriggers represents the various scaling triggers
type ScalingTriggers struct {
	QueueLoad      *QueueLoadTrigger   `json:"queue_load,omitempty"`
	CPUUtilization *UtilizationTrigger `json:"cpu_utilization,omitempty"`
	GPUUtilization *UtilizationTrigger `json:"gpu_utilization,omitempty"`
}

// QueueLoadTrigger represents queue load based scaling trigger
type QueueLoadTrigger struct {
	Threshold float64 `json:"threshold,omitempty"`
}

// UtilizationTrigger represents CPU/GPU utilization based scaling trigger
type UtilizationTrigger struct {
	Enabled   bool `json:"enabled,omitempty"`
	Threshold int  `json:"threshold,omitempty"`
}

// UpdateScalingOptionsRequest is an alias for ContainerScalingOptions used for PATCH requests
// All fields are optional for partial updates
type UpdateScalingOptionsRequest = ContainerScalingOptions

// DeploymentReplicas represents replica information for a deployment
type DeploymentReplicas struct {
	Replicas []ReplicaInfo `json:"replicas"`
}

// ReplicaInfo represents information about a single replica
type ReplicaInfo struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Node   string `json:"node,omitempty"`
}

// EnvironmentVariablesRequest represents a request to add/update environment variables
// Used by POST and PATCH /container-deployments/{name}/environment-variables
type EnvironmentVariablesRequest struct {
	ContainerName string            `json:"container_name"`
	Env           []ContainerEnvVar `json:"env"`
}

// ContainerHealthcheck represents container healthcheck
type ContainerHealthcheck struct {
	Enabled bool   `json:"enabled"`
	Port    int    `json:"port,omitempty"`
	Path    string `json:"path,omitempty"`
}

// ContainerEntrypointOverrides includes functionality for overriding container entrypoint
type ContainerEntrypointOverrides struct {
	Enabled    bool     `json:"enabled"`
	Entrypoint []string `json:"entrypoint,omitempty"`
	Cmd        []string `json:"cmd,omitempty"`
}

// ContainerVolumeMount represents the container volume mount
type ContainerVolumeMount struct {
	Type       string `json:"type"` // "scratch", "secret", etc.
	MountPath  string `json:"mount_path"`
	SecretName string `json:"secret_name,omitempty"`
	SizeInMB   int    `json:"size_in_mb,omitempty"`
	VolumeID   string `json:"volume_id,omitempty"`
}

// ContainerAutoUpdate has automatic update instructions that can be used when updating existing deployment
type ContainerAutoUpdate struct {
	Enabled   bool   `json:"enabled"`
	Mode      string `json:"mode"`
	TagFilter string `json:"tag_filter,omitempty"`
}

// DeleteEnvironmentVariablesRequest represents a request to delete environment variables
// Uses the same format as add/update but only the Name field is required
type DeleteEnvironmentVariablesRequest struct {
	ContainerName string            `json:"container_name"`
	Env           []ContainerEnvVar `json:"env"`
}

// ComputeResource represents available compute resources
type ComputeResource struct {
	Name        string `json:"name"`
	Size        string `json:"size"`
	IsAvailable bool   `json:"is_available"`
}

// Secret represents a secret used in deployments
type Secret struct {
	Name       string `json:"name"`
	CreatedAt  string `json:"created_at"`
	SecretType string `json:"secret_type"`
}

// CreateSecretRequest represents a request to create a new secret
type CreateSecretRequest struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// FileSecret represents a fileset secret
type FileSecret struct {
	Name       string   `json:"name"`
	CreatedAt  string   `json:"created_at"`
	SecretType string   `json:"secret_type"`
	FileNames  []string `json:"file_names,omitempty"`
}

// CreateFileSecretRequest represents a request to create a fileset secret
type CreateFileSecretRequest struct {
	Name  string           `json:"name"`
	Files []FileSecretFile `json:"files"`
}

// FileSecretFile represents a file in a fileset secret.
type FileSecretFile struct {
	Name          string `json:"file_name"`
	Base64Content string `json:"base64_content"`
}

// RegistryCredentials represents container registry credentials
type RegistryCredentials struct {
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// CreateRegistryCredentialsRequest represents a request to create registry credentials
// Type can be: "dockerhub", "gcr", "ghcr", "ecr", "scaleway", etc.
type CreateRegistryCredentialsRequest struct {
	Name              string `json:"name"`
	Type              string `json:"type"`
	Username          string `json:"username,omitempty"`
	AccessToken       string `json:"access_token,omitempty"`
	ServiceAccountKey string `json:"service_account_key,omitempty"`
	DockerConfigJson  string `json:"docker_config_json,omitempty"`
	AccessKeyID       string `json:"access_key_id,omitempty"`
	SecretAccessKey   string `json:"secret_access_key,omitempty"`
	Region            string `json:"region,omitempty"`
	EcrRepo           string `json:"ecr_repo,omitempty"`
	ScalewayDomain    string `json:"scaleway_domain,omitempty"`
	ScalewayUUID      string `json:"scaleway_uuid,omitempty"`
}

// Serverless Jobs Types

// JobDeploymentShortInfo represents summary information about a job deployment
type JobDeploymentShortInfo struct {
	Name      string            `json:"name"`
	CreatedAt string            `json:"created_at"`
	Compute   *ContainerCompute `json:"compute,omitempty"`
}

// JobDeployment represents a complete serverless job deployment
// Shares types with ContainerDeployment for consistency
type JobDeployment struct {
	Name                      string                     `json:"name"`
	Containers                []DeploymentContainer      `json:"containers"`
	EndpointBaseURL           string                     `json:"endpoint_base_url,omitempty"`
	CreatedAt                 string                     `json:"created_at,omitempty"`
	Compute                   *ContainerCompute          `json:"compute,omitempty"`
	ContainerRegistrySettings *ContainerRegistrySettings `json:"container_registry_settings,omitempty"`
	Scaling                   *JobScalingOptions         `json:"scaling,omitempty"`
}

// JobContainerRegistrySettings represents registry settings for job deployments
type JobContainerRegistrySettings struct {
	Credentials *JobRegistryCredentials `json:"credentials,omitempty"`
}

// JobRegistryCredentials references registry credentials by name
type JobRegistryCredentials struct {
	Name string `json:"name"`
}

// JobContainer represents a container configuration in a job deployment
type JobContainer struct {
	Image               string                  `json:"image"`
	ExposedPort         int                     `json:"exposed_port,omitempty"`
	Healthcheck         *JobHealthcheck         `json:"healthcheck,omitempty"`
	EntrypointOverrides *JobEntrypointOverrides `json:"entrypoint_overrides,omitempty"`
	Env                 []JobEnvVar             `json:"env,omitempty"`
	VolumeMounts        []JobVolumeMount        `json:"volume_mounts,omitempty"`
}

// JobHealthcheck represents health check configuration for jobs
type JobHealthcheck struct {
	Enabled bool   `json:"enabled"`
	Port    int    `json:"port,omitempty"`
	Path    string `json:"path,omitempty"`
}

// JobEntrypointOverrides allows overriding container entrypoint and cmd for jobs
type JobEntrypointOverrides struct {
	Enabled    bool     `json:"enabled"`
	Entrypoint []string `json:"entrypoint,omitempty"`
	Cmd        []string `json:"cmd,omitempty"`
}

// JobEnvVar represents an environment variable for job containers
type JobEnvVar struct {
	Name                     string `json:"name"`
	ValueOrReferenceToSecret string `json:"value_or_reference_to_secret"`
	Type                     string `json:"type"` // "plain" or "secret"
}

// JobVolumeMount represents a volume mount for job containers
type JobVolumeMount struct {
	Type       string `json:"type"` // "scratch", "secret", etc.
	MountPath  string `json:"mount_path"`
	SecretName string `json:"secret_name,omitempty"`
	SizeInMB   int    `json:"size_in_mb,omitempty"`
	VolumeID   string `json:"volumeId,omitempty"`
}

// JobCompute represents compute resources for job deployments
type JobCompute struct {
	Name string `json:"name"` // e.g., "H100", "A100"
	Size int    `json:"size"` // Number of GPUs
}

// CreateJobDeploymentRequest represents a request to create a new job deployment
// Shares container, compute, and scaling types with container deployments
type CreateJobDeploymentRequest struct {
	Name                      string                      `json:"name"`
	ContainerRegistrySettings *ContainerRegistrySettings  `json:"container_registry_settings,omitempty"`
	Containers                []CreateDeploymentContainer `json:"containers"`
	Compute                   *ContainerCompute           `json:"compute,omitempty"`
	Scaling                   *ContainerScalingOptions    `json:"scaling,omitempty"`
}

// UpdateJobDeploymentRequest represents a request to update a job deployment
// Shares container, compute, and scaling types with container deployments
type UpdateJobDeploymentRequest struct {
	ContainerRegistrySettings *ContainerRegistrySettings  `json:"container_registry_settings,omitempty"`
	Containers                []CreateDeploymentContainer `json:"containers,omitempty"`
	Compute                   *ContainerCompute           `json:"compute,omitempty"`
	Scaling                   *ContainerScalingOptions    `json:"scaling,omitempty"`
}

// JobScalingOptions represents scaling configuration for a job deployment
type JobScalingOptions struct {
	MaxReplicaCount        int `json:"max_replica_count"`
	QueueMessageTTLSeconds int `json:"queue_message_ttl_seconds,omitempty"`
	DeadlineSeconds        int `json:"deadline_seconds,omitempty"`
}

// JobDeploymentStatus represents the status of a job deployment
type JobDeploymentStatus struct {
	Status        string `json:"status"`
	ActiveJobs    int    `json:"active_jobs,omitempty"`
	SucceededJobs int    `json:"succeeded_jobs,omitempty"`
	FailedJobs    int    `json:"failed_jobs,omitempty"`
}

// FlexibleFloat is a custom type that can unmarshal both string and float64 values
type FlexibleFloat float64

// UnmarshalJSON implements json.Unmarshaler to handle both string and float64 inputs
func (f *FlexibleFloat) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as float64 first
	var floatVal float64
	if err := json.Unmarshal(data, &floatVal); err == nil {
		*f = FlexibleFloat(floatVal)
		return nil
	}

	// Try to unmarshal as string
	var strVal string
	if err := json.Unmarshal(data, &strVal); err != nil {
		return err
	}

	// Convert string to float64
	floatVal, err := strconv.ParseFloat(strVal, 64)
	if err != nil {
		return err
	}

	*f = FlexibleFloat(floatVal)
	return nil
}

// MarshalJSON implements json.Marshaler to always marshal as float64
func (f FlexibleFloat) MarshalJSON() ([]byte, error) {
	return json.Marshal(float64(f))
}

// Float64 returns the float64 value
func (f FlexibleFloat) Float64() float64 {
	return float64(f)
}
