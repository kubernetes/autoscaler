/*
Copyright 2018 The Kubernetes Authors.

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

package config

import (
	"time"

	gce_localssdsize "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/gce/localssdsize"
	kubelet_config "k8s.io/kubernetes/pkg/kubelet/apis/config"
	scheduler_config "k8s.io/kubernetes/pkg/scheduler/apis/config"
)

// GpuLimits define lower and upper bound on GPU instances of given type in cluster
type GpuLimits struct {
	// Type of the GPU (e.g. nvidia-tesla-k80)
	GpuType string
	// Lower bound on number of GPUs of given type in cluster
	Min int64
	// Upper bound on number of GPUs of given type in cluster
	Max int64
}

// NodeGroupAutoscalingOptions contain various options to customize how autoscaling of
// a given NodeGroup works. Different options can be used for each NodeGroup.
type NodeGroupAutoscalingOptions struct {
	// ScaleDownUtilizationThreshold sets threshold for nodes to be considered for scale down if cpu or memory utilization is over threshold.
	// Well-utilized nodes are not touched.
	ScaleDownUtilizationThreshold float64
	// ScaleDownGpuUtilizationThreshold sets threshold for gpu nodes to be considered for scale down if gpu utilization is over threshold.
	// Well-utilized nodes are not touched.
	ScaleDownGpuUtilizationThreshold float64
	// ScaleDownUnneededTime sets the duration CA expects a node to be unneeded/eligible for removal
	// before scaling down the node.
	ScaleDownUnneededTime time.Duration
	// ScaleDownUnreadyTime represents how long an unready node should be unneeded before it is eligible for scale down
	ScaleDownUnreadyTime time.Duration
	// Maximum time CA waits for node to be provisioned
	MaxNodeProvisionTime time.Duration
	// ZeroOrMaxNodeScaling means that a node group should be scaled up to maximum size or down to zero nodes all at once instead of one-by-one.
	ZeroOrMaxNodeScaling bool
	// IgnoreDaemonSetsUtilization sets if daemonsets utilization should be considered during node scale-down
	IgnoreDaemonSetsUtilization bool
}

// GCEOptions contain autoscaling options specific to GCE cloud provider.
type GCEOptions struct {
	// ConcurrentRefreshes is the maximum number of concurrently refreshed instance groups or instance templates or zones with mig instances
	ConcurrentRefreshes int
	// MigInstancesMinRefreshWaitTime is the minimum time which needs to pass before GCE MIG instances from a given MIG can be refreshed.
	MigInstancesMinRefreshWaitTime time.Duration
	// DomainUrl is the GCE url used to make calls to GCE API.
	DomainUrl string
	// LocalSSDDiskSizeProvider provides local ssd disk size based on machine type
	LocalSSDDiskSizeProvider gce_localssdsize.LocalSSDSizeProvider
	// BulkMigInstancesListingEnabled means that cluster instances should be listed in bulk instead of per mig.
	// Instances of migs having instances in creating or deleting state are re-fetched using igm.ListInstances. Inconsistencies are handled by re-fetching using igm.ListInstances
	BulkMigInstancesListingEnabled bool
}

const (
	// DefaultMaxAllocatableDifferenceRatio describes how Node.Status.Allocatable can differ between groups in the same NodeGroupSet
	DefaultMaxAllocatableDifferenceRatio = 0.05
	// DefaultMaxFreeDifferenceRatio describes how free resources (allocatable - daemon and system pods)
	DefaultMaxFreeDifferenceRatio = 0.05
	// DefaultMaxCapacityMemoryDifferenceRatio describes how Node.Status.Capacity.Memory
	DefaultMaxCapacityMemoryDifferenceRatio = 0.015
)

// NodeGroupDifferenceRatios contains various ratios used to determine if two NodeGroups are similar and makes scaling decisions
type NodeGroupDifferenceRatios struct {
	// MaxAllocatableDifferenceRatio describes how Node.Status.Allocatable can differ between groups in the same NodeGroupSet
	MaxAllocatableDifferenceRatio float64
	// MaxFreeDifferenceRatio describes how free resources (allocatable - daemon and system pods) can differ between groups in the same NodeGroupSet
	MaxFreeDifferenceRatio float64
	// MaxCapacityMemoryDifferenceRatio describes how Node.Status.Capacity.Memory can differ between groups in the same NodeGroupSetAutoscalingOptions
	MaxCapacityMemoryDifferenceRatio float64
}

// NewDefaultNodeGroupDifferenceRatios returns default NodeGroupDifferenceRatios values
func NewDefaultNodeGroupDifferenceRatios() NodeGroupDifferenceRatios {
	return NodeGroupDifferenceRatios{
		MaxAllocatableDifferenceRatio:    DefaultMaxAllocatableDifferenceRatio,
		MaxFreeDifferenceRatio:           DefaultMaxFreeDifferenceRatio,
		MaxCapacityMemoryDifferenceRatio: DefaultMaxCapacityMemoryDifferenceRatio,
	}
}

// AutoscalingOptions contain various options to customize how autoscaling works
type AutoscalingOptions struct {
	// NodeGroupDefaults are default values for per NodeGroup options.
	// They will be used any time a specific value is not provided for a given NodeGroup.
	NodeGroupDefaults NodeGroupAutoscalingOptions
	// MaxEmptyBulkDelete is a number of empty nodes that can be removed at the same time.
	MaxEmptyBulkDelete int
	// MaxNodesTotal sets the maximum number of nodes in the whole cluster
	MaxNodesTotal int
	// MaxCoresTotal sets the maximum number of cores in the whole cluster
	MaxCoresTotal int64
	// MinCoresTotal sets the minimum number of cores in the whole cluster
	MinCoresTotal int64
	// MaxMemoryTotal sets the maximum memory (in bytes) in the whole cluster
	MaxMemoryTotal int64
	// MinMemoryTotal sets the maximum memory (in bytes) in the whole cluster
	MinMemoryTotal int64
	// GpuTotal is a list of strings with configuration of min/max limits for different GPUs.
	GpuTotal []GpuLimits
	// NodeGroupAutoDiscovery represents one or more definition(s) of node group auto-discovery
	NodeGroupAutoDiscovery []string
	// EstimatorName is the estimator used to estimate the number of needed nodes in scale up.
	EstimatorName string
	// ExpanderNames sets the chain of node group expanders to be used in scale up
	ExpanderNames string
	// GRPCExpanderCert is the location of the cert passed to the gRPC server for TLS when using the gRPC expander
	GRPCExpanderCert string
	// GRPCExpanderURL is the url of the gRPC server when using the gRPC expander
	GRPCExpanderURL string
	// IgnoreMirrorPodsUtilization is whether CA will ignore Mirror pods when calculating resource utilization for scaling down
	IgnoreMirrorPodsUtilization bool
	// MaxGracefulTerminationSec is maximum number of seconds scale down waits for pods to terminate before
	// removing the node from cloud provider.
	// DrainPriorityConfig takes higher precedence and MaxGracefulTerminationSec will not be applicable when the DrainPriorityConfig is set.
	MaxGracefulTerminationSec int
	// DrainPriorityConfig is a list of ShutdownGracePeriodByPodPriority.
	// This field is optional and could be nil.
	// DrainPriorityConfig takes higher precedence and MaxGracefulTerminationSec will not be applicable when the DrainPriorityConfig is set.
	DrainPriorityConfig []kubelet_config.ShutdownGracePeriodByPodPriority
	// MaxTotalUnreadyPercentage is the maximum percentage of unready nodes after which CA halts operations
	MaxTotalUnreadyPercentage float64
	// OkTotalUnreadyCount is the number of allowed unready nodes, irrespective of max-total-unready-percentage
	OkTotalUnreadyCount int
	// ScaleUpFromZero defines if CA should scale up when there 0 ready nodes.
	ScaleUpFromZero bool
	// ParallelScaleUp defines whether CA can scale up node groups in parallel.
	ParallelScaleUp bool
	// CloudConfig is the path to the cloud provider configuration file. Empty string for no configuration file.
	CloudConfig string
	// CloudProviderName sets the type of the cloud provider CA is about to run in. Allowed values: gce, aws
	CloudProviderName string
	// NodeGroups is the list of node groups a.k.a autoscaling targets
	NodeGroups []string
	// EnforceNodeGroupMinSize is used to allow CA to scale up the node group to the configured min size if needed.
	EnforceNodeGroupMinSize bool
	// ScaleDownEnabled is used to allow CA to scale down the cluster
	ScaleDownEnabled bool
	// ScaleDownUnreadyEnabled is used to allow CA to scale down unready nodes of the cluster
	ScaleDownUnreadyEnabled bool
	// ScaleDownDelayAfterAdd sets the duration from the last scale up to the time when CA starts to check scale down options
	ScaleDownDelayAfterAdd time.Duration
	// ScaleDownDelayAfterDelete sets the duration between scale down attempts if scale down removes one or more nodes
	ScaleDownDelayAfterDelete time.Duration
	// ScaleDownDelayAfterFailure sets the duration before the next scale down attempt if scale down results in an error
	ScaleDownDelayAfterFailure time.Duration
	// ScaleDownDelayTypeLocal sets if the --scale-down-delay-after-* flags should be applied locally per nodegroup
	// or globally across all nodegroups
	ScaleDownDelayTypeLocal bool
	// ScaleDownNonEmptyCandidatesCount is the maximum number of non empty nodes
	// considered at once as candidates for scale down.
	ScaleDownNonEmptyCandidatesCount int
	// ScaleDownCandidatesPoolRatio is a ratio of nodes that are considered
	// as additional non empty candidates for scale down when some candidates from
	// previous iteration are no longer valid.
	ScaleDownCandidatesPoolRatio float64
	// ScaleDownCandidatesPoolMinCount is the minimum number of nodes that are
	// considered as additional non empty candidates for scale down when some
	// candidates from previous iteration are no longer valid.
	// The formula to calculate additional candidates number is following:
	// max(#nodes * ScaleDownCandidatesPoolRatio, ScaleDownCandidatesPoolMinCount)
	ScaleDownCandidatesPoolMinCount int
	// ScaleDownSimulationTimeout defines the maximum time that can be
	// spent on scale down simulation.
	ScaleDownSimulationTimeout time.Duration
	// SchedulerConfig allows changing configuration of in-tree
	// scheduler plugins acting on PreFilter and Filter extension points
	SchedulerConfig *scheduler_config.KubeSchedulerConfiguration
	// NodeDeletionDelayTimeout is maximum time CA waits for removing delay-deletion.cluster-autoscaler.kubernetes.io/ annotations before deleting the node.
	NodeDeletionDelayTimeout time.Duration
	// WriteStatusConfigMap tells if the status information should be written to a ConfigMap
	WriteStatusConfigMap bool
	// StaticConfigMapName
	StatusConfigMapName string
	// BalanceSimilarNodeGroups enables logic that identifies node groups with similar machines and tries to balance node count between them.
	BalanceSimilarNodeGroups bool
	// ConfigNamespace is the namespace cluster-autoscaler is running in and all related configmaps live in
	ConfigNamespace string
	// ClusterName if available
	ClusterName string
	// NodeAutoprovisioningEnabled tells whether the node auto-provisioning is enabled for this cluster.
	NodeAutoprovisioningEnabled bool
	// MaxAutoprovisionedNodeGroupCount is the maximum number of autoprovisioned groups in the cluster.
	MaxAutoprovisionedNodeGroupCount int
	// UnremovableNodeRecheckTimeout is the timeout before we check again a node that couldn't be removed before
	UnremovableNodeRecheckTimeout time.Duration
	// Pods with priority below cutoff are expendable. They can be killed without any consideration during scale down and they don't cause scale-up.
	// Pods with null priority (PodPriority disabled) are non-expendable.
	ExpendablePodsPriorityCutoff int
	// Regional tells whether the cluster is regional.
	Regional bool
	// Pods newer than this will not be considered as unschedulable for scale-up.
	NewPodScaleUpDelay time.Duration
	// MaxBulkSoftTaint sets the maximum number of nodes that can be (un)tainted PreferNoSchedule during single scaling down run.
	// Value of 0 turns turn off such tainting.
	MaxBulkSoftTaintCount int
	// MaxBulkSoftTaintTime sets the maximum duration of single run of PreferNoSchedule tainting.
	MaxBulkSoftTaintTime time.Duration
	// MaxPodEvictionTime sets the maximum time CA tries to evict a pod before giving up.
	MaxPodEvictionTime time.Duration
	// StartupTaints is a list of taints CA considers to reflect transient node
	// status that should be removed when creating a node template for scheduling.
	// startup taints are expected to appear during node startup.
	StartupTaints []string
	// StatusTaints is a list of taints CA considers to reflect transient node
	// status that should be removed when creating a node template for scheduling.
	// The status taints are expected to appear during node lifetime, after startup.
	StatusTaints []string
	// BalancingExtraIgnoredLabels is a list of labels to additionally ignore when comparing if two node groups are similar.
	// Labels in BasicIgnoredLabels and the cloud provider-specific ignored labels are always ignored.
	BalancingExtraIgnoredLabels []string
	// BalancingLabels is a list of labels to use when comparing if two node groups are similar.
	// If this is set, only labels are used to compare node groups. It is mutually exclusive with BalancingExtraIgnoredLabels.
	BalancingLabels []string
	// AWSUseStaticInstanceList tells if AWS cloud provider use static instance type list or dynamically fetch from remote APIs.
	AWSUseStaticInstanceList bool
	// GCEOptions contain autoscaling options specific to GCE cloud provider.
	GCEOptions GCEOptions
	// KubeClientOpts specify options for kube client
	KubeClientOpts KubeClientOptions
	// ClusterAPICloudConfigAuthoritative tells the Cluster API provider to treat the CloudConfig option as authoritative and
	// not use KubeConfigPath as a fallback when it is not provided.
	ClusterAPICloudConfigAuthoritative bool
	// Enable or disable cordon nodes functionality before terminating the node during downscale process
	CordonNodeBeforeTerminate bool
	// DaemonSetEvictionForEmptyNodes is whether CA will gracefully terminate DaemonSet pods from empty nodes.
	DaemonSetEvictionForEmptyNodes bool
	// DaemonSetEvictionForOccupiedNodes is whether CA will gracefully terminate DaemonSet pods from non-empty nodes.
	DaemonSetEvictionForOccupiedNodes bool
	// User agent to use for HTTP calls.
	UserAgent string
	// InitialNodeGroupBackoffDuration is the duration of first backoff after a new node failed to start
	InitialNodeGroupBackoffDuration time.Duration
	// MaxNodeGroupBackoffDuration is the maximum backoff duration for a NodeGroup after new nodes failed to start.
	MaxNodeGroupBackoffDuration time.Duration
	// NodeGroupBackoffResetTimeout is the time after last failed scale-up when the backoff duration is reset.
	NodeGroupBackoffResetTimeout time.Duration
	// MaxScaleDownParallelism is the maximum number of nodes (both empty and needing drain) that can be deleted in parallel.
	MaxScaleDownParallelism int
	// MaxDrainParallelism is the maximum number of nodes needing drain, that can be drained and deleted in parallel.
	MaxDrainParallelism int
	// RecordDuplicatedEvents controls whether events should be duplicated within a 5 minute window.
	RecordDuplicatedEvents bool
	// MaxNodesPerScaleUp controls how many nodes can be added in a single scale-up.
	// Note that this is strictly a performance optimization aimed at limiting binpacking time, not a tool to rate-limit
	// scale-up. There is nothing stopping CA from adding MaxNodesPerScaleUp every loop.
	MaxNodesPerScaleUp int
	// MaxNodeGroupBinpackingDuration is a maximum time that can be spent binpacking a single NodeGroup. If the threshold
	// is exceeded binpacking will be cut short and a partial scale-up will be performed.
	MaxNodeGroupBinpackingDuration time.Duration
	// MaxBinpackingTime is the maximum time spend on binpacking for a single scale-up.
	// If binpacking is limited by this, scale-up will continue with the already calculated scale-up options.
	MaxBinpackingTime time.Duration
	// NodeDeletionBatcherInterval is a time for how long CA ScaleDown gather nodes to delete them in batch.
	NodeDeletionBatcherInterval time.Duration
	// SkipNodesWithSystemPods tells if nodes with pods from kube-system should be deleted (except for DaemonSet or mirror pods)
	SkipNodesWithSystemPods bool
	// SkipNodesWithLocalStorage tells if nodes with pods with local storage, e.g. EmptyDir or HostPath, should be deleted
	SkipNodesWithLocalStorage bool
	// SkipNodesWithCustomControllerPods tells if nodes with custom-controller owned pods should be skipped from deletion (skip if 'true')
	SkipNodesWithCustomControllerPods bool
	// MinReplicaCount controls the minimum number of replicas that a replica set or replication controller should have
	// to allow their pods deletion in scale down
	MinReplicaCount int
	// NodeDeleteDelayAfterTaint is the duration to wait before deleting a node after tainting it
	NodeDeleteDelayAfterTaint time.Duration
	// NodeGroupSetRatio is a collection of ratios used by CA used to make scaling decisions.
	NodeGroupSetRatios NodeGroupDifferenceRatios
	// dynamicNodeDeleteDelayAfterTaintEnabled is used to enable/disable dynamic adjustment of NodeDeleteDelayAfterTaint
	// based on the latency between the CA and the api-server
	DynamicNodeDeleteDelayAfterTaintEnabled bool
	// BypassedSchedulers are used to specify which schedulers to bypass their processing
	BypassedSchedulers map[string]bool
	// ProvisioningRequestEnabled tells if CA processes ProvisioningRequest.
	ProvisioningRequestEnabled bool
	// AsyncNodeGroupsEnabled tells if CA creates/deletes node groups asynchronously.
	AsyncNodeGroupsEnabled bool
	// ProvisioningRequestInitialBackoffTime is the initial time for ProvisioningRequest be considered by CA after failed ScaleUp request.
	ProvisioningRequestInitialBackoffTime time.Duration
	// ProvisioningRequestMaxBackoffTime is the max time for ProvisioningRequest be considered by CA after failed ScaleUp request.
	ProvisioningRequestMaxBackoffTime time.Duration
	// ProvisioningRequestMaxCacheSize is the max size for ProvisioningRequest cache that is stored for retry backoff.
	ProvisioningRequestMaxBackoffCacheSize int
	// EnableDynamicResources configures whether logic for handling DRA objects is enabled.
	EnableDynamicResources bool
}

// KubeClientOptions specify options for kube client
type KubeClientOptions struct {
	// Master specifies master location.
	Master string
	// Path to kube configuration if available
	KubeConfigPath string
	// APIContentType specifies type of requests sent to APIServer.
	APIContentType string
	// Burst setting for kubernetes client
	KubeClientBurst int
	// QPS setting for kubernetes client
	KubeClientQPS float32
}
