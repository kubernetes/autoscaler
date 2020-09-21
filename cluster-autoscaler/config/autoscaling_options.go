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

// AutoscalingOptions contain various options to customize how autoscaling works
type AutoscalingOptions struct {
	// MaxEmptyBulkDelete is a number of empty nodes that can be removed at the same time.
	MaxEmptyBulkDelete int
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
	// ExpanderName sets the type of node group expander to be used in scale up
	ExpanderName string
	// IgnoreDaemonSetsUtilization is whether CA will ignore DaemonSet pods when calculating resource utilization for scaling down
	IgnoreDaemonSetsUtilization bool
	// IgnoreMirrorPodsUtilization is whether CA will ignore Mirror pods when calculating resource utilization for scaling down
	IgnoreMirrorPodsUtilization bool
	// MaxGracefulTerminationSec is maximum number of seconds scale down waits for pods to terminate before
	// removing the node from cloud provider.
	MaxGracefulTerminationSec int
	//  Maximum time CA waits for node to be provisioned
	MaxNodeProvisionTime time.Duration
	// MaxTotalUnreadyPercentage is the maximum percentage of unready nodes after which CA halts operations
	MaxTotalUnreadyPercentage float64
	// OkTotalUnreadyCount is the number of allowed unready nodes, irrespective of max-total-unready-percentage
	OkTotalUnreadyCount int
	// ScaleUpFromZero defines if CA should scale up when there 0 ready nodes.
	ScaleUpFromZero bool
	// CloudConfig is the path to the cloud provider configuration file. Empty string for no configuration file.
	CloudConfig string
	// CloudProviderName sets the type of the cloud provider CA is about to run in. Allowed values: gce, aws
	CloudProviderName string
	// NodeGroups is the list of node groups a.k.a autoscaling targets
	NodeGroups []string
	// ScaleDownEnabled is used to allow CA to scale down the cluster
	ScaleDownEnabled bool
	// ScaleDownDelayAfterAdd sets the duration from the last scale up to the time when CA starts to check scale down options
	ScaleDownDelayAfterAdd time.Duration
	// ScaleDownDelayAfterDelete sets the duration between scale down attempts if scale down removes one or more nodes
	ScaleDownDelayAfterDelete time.Duration
	// ScaleDownDelayAfterFailure sets the duration before the next scale down attempt if scale down results in an error
	ScaleDownDelayAfterFailure time.Duration
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
	// NodeDeletionDelayTimeout is maximum time CA waits for removing delay-deletion.cluster-autoscaler.kubernetes.io/ annotations before deleting the node.
	NodeDeletionDelayTimeout time.Duration
	// WriteStatusConfigMap tells if the status information should be written to a ConfigMap
	WriteStatusConfigMap bool
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
	// IgnoredTaints is a list of taints to ignore when considering a node template for scheduling.
	IgnoredTaints []string
	// BalancingExtraIgnoredLabels is a list of labels to additionally ignore when comparing if two node groups are similar.
	// Labels in BasicIgnoredLabels and the cloud provider-specific ignored labels are always ignored.
	BalancingExtraIgnoredLabels []string
	// AWSUseStaticInstanceList tells if AWS cloud provider use static instance type list or dynamically fetch from remote APIs.
	AWSUseStaticInstanceList bool
	// Path to kube configuration if available
	KubeConfigPath string
	// ClusterAPICloudConfigAuthoritative tells the Cluster API provider to treat the CloudConfig option as authoritative and
	// not use KubeConfigPath as a fallback when it is not provided.
	ClusterAPICloudConfigAuthoritative bool
}
