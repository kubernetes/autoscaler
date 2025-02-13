/*
Copyright 2017 The Kubernetes Authors.

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

package api

// TODO: Remove this once Cluster Autoscaler api is approved.

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterAutoscalerStatusCondition is the status of the cluster autoscaler.
type ClusterAutoscalerStatusCondition string

const (
	// ClusterAutoscalerRunning status means that the cluster autoscaler has been initialized and running.
	ClusterAutoscalerRunning ClusterAutoscalerStatusCondition = "Running"
	// ClusterAutoscalerInitializing status means that cluster autoscaler is currently being initialized.
	ClusterAutoscalerInitializing ClusterAutoscalerStatusCondition = "Initializing"
)

// ClusterAutoscalerConditionStatus is a status of ClusterAutoscalerCondition.
type ClusterAutoscalerConditionStatus string

const (
	// Statuses for Health condition type.

	// ClusterAutoscalerHealthy status means that the cluster is in a good shape.
	ClusterAutoscalerHealthy ClusterAutoscalerConditionStatus = "Healthy"
	// ClusterAutoscalerUnhealthy status means that the cluster is in a bad shape.
	ClusterAutoscalerUnhealthy ClusterAutoscalerConditionStatus = "Unhealthy"

	// Statuses for ScaleDown condition type.

	//ClusterAutoscalerCandidatesPresent status means that there are candidates for scale down.
	ClusterAutoscalerCandidatesPresent ClusterAutoscalerConditionStatus = "CandidatesPresent"
	//ClusterAutoscalerNoCandidates status means that there are no candidates for scale down.
	ClusterAutoscalerNoCandidates ClusterAutoscalerConditionStatus = "NoCandidates"

	// Statuses for ScaleUp condition type.

	// ClusterAutoscalerNeeded status means that scale up is needed.
	ClusterAutoscalerNeeded ClusterAutoscalerConditionStatus = "Needed"
	// ClusterAutoscalerNotNeeded status means that scale up is not needed.
	ClusterAutoscalerNotNeeded ClusterAutoscalerConditionStatus = "NotNeeded"
	// ClusterAutoscalerInProgress status means that scale up is in progress.
	ClusterAutoscalerInProgress ClusterAutoscalerConditionStatus = "InProgress"
	// ClusterAutoscalerNoActivity status means that there has been no scale up activity recently.
	ClusterAutoscalerNoActivity ClusterAutoscalerConditionStatus = "NoActivity"
	// ClusterAutoscalerBackoff status means that due to a recently failed scale-up no further scale-ups attempts will be made for some time.
	ClusterAutoscalerBackoff ClusterAutoscalerConditionStatus = "Backoff"
	// ClusterAutoscalerHasDynamicHeadroom status means that one or more nodes in a node group has available headroom for vertical scaling which should be preferred over adding more nodes.
	ClusterAutoscalerHasDynamicHeadroom ClusterAutoscalerConditionStatus = "HasDynamicHeadroom"
)

// RegisteredUnreadyNodeCount contains node counts of registered but unready nodes.
type RegisteredUnreadyNodeCount struct {
	// Total number of registered but unready nodes.
	Total int `json:"total" yaml:"total"`
	// ResourceUnready is the number of registered but unready nodes due to a missing resource (e.g. GPU).
	ResourceUnready int `json:"resourceUnready" yaml:"resourceUnready"`
}

// RegisteredNodeCount contains node counts of registered nodes.
type RegisteredNodeCount struct {
	Total      int `json:"total" yaml:"total"`
	Ready      int `json:"ready" yaml:"ready"`
	NotStarted int `json:"notStarted" yaml:"notStarted"`
	// Number of nodes that are being currently deleted. They exist in K8S but are not included in NodeGroup.TargetSize().
	BeingDeleted int                        `json:"beingDeleted,omitempty" yaml:"beingDeleted,omitempty"`
	Unready      RegisteredUnreadyNodeCount `json:"unready,omitempty" yaml:"unready,omitempty"`
}

// NodeCount contains number of nodes that satisfy different criteria.
type NodeCount struct {
	Registered       RegisteredNodeCount `json:"registered,omitempty" yaml:"registered,omitempty"`
	LongUnregistered int                 `json:"longUnregistered" yaml:"longUnregistered"`
	Unregistered     int                 `json:"unregistered" yaml:"unregistered"`
}

// ClusterHealthCondition contains information about health condition for the whole cluster.
type ClusterHealthCondition struct {
	// Status of cluster health.
	Status ClusterAutoscalerConditionStatus `json:"status,omitempty" yaml:"status,omitempty"`
	// NodeCounts contains number of nodes that satisfy different criteria in the cluster.
	NodeCounts NodeCount `json:"nodeCounts,omitempty" yaml:"nodeCounts,omitempty"`
	// LastProbeTime is the last time we probed the condition.
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty" yaml:"lastProbeTime,omitempty"`
	// LastTransitionTime is the time since when the condition was in the given state.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" yaml:"lastTransitionTime,omitempty"`
}

// NodeGroupHealthCondition contains information about health condition for a node group.
type NodeGroupHealthCondition struct {
	// Status of node group health.
	Status ClusterAutoscalerConditionStatus `json:"status,omitempty" yaml:"status,omitempty"`
	// NodeCounts contains number of nodes that satisfy different criteria in the node group.
	NodeCounts NodeCount `json:"nodeCounts,omitempty" yaml:"nodeCounts,omitempty"`
	// CloudProviderTarget is the target size set by cloud provider.
	CloudProviderTarget int `json:"cloudProviderTarget" yaml:"cloudProviderTarget"`
	// MinSize is the CA max size of a node group.
	MinSize int `json:"minSize" yaml:"minSize"`
	// MaxSize is the CA max size of a node group.
	MaxSize int `json:"maxSize" yaml:"maxSize"`
	// LastProbeTime is the last time we probed the condition.
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty" yaml:"lastProbeTime,omitempty"`
	// LastTransitionTime is the time since when the condition was in the given state.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" yaml:"lastTransitionTime,omitempty"`
}

// ClusterScaleUpCondition contains information about scale up condition for the whole cluster.
type ClusterScaleUpCondition struct {
	// Status of the scale up.
	Status ClusterAutoscalerConditionStatus `json:"status,omitempty" yaml:"status,omitempty"`
	// LastProbeTime is the last time we probed the condition.
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty" yaml:"lastProbeTime,omitempty"`
	// LastTransitionTime is the time since when the condition was in the given state.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" yaml:"lastTransitionTime,omitempty"`
}

// BackoffInfo contains error information that caused the backoff.
type BackoffInfo struct {
	// ErrorCode is a specific error code for error condition
	ErrorCode string `json:"errorCode,omitempty" yaml:"errorCode,omitempty"`
	// ErrorMessage is human readable description of error condition
	ErrorMessage string `json:"errorMessage,omitempty" yaml:"errorMessage,omitempty"`
}

// NodeGroupScaleUpCondition contains information about scale up condition for a node group.
type NodeGroupScaleUpCondition struct {
	// Status of the scale up.
	Status ClusterAutoscalerConditionStatus `json:"status,omitempty" yaml:"status,omitempty"`
	// LastProbeTime is the last time we probed the condition.
	BackoffInfo BackoffInfo `json:"backoffInfo,omitempty" yaml:"backoffInfo,omitempty"`
	// LastProbeTime is the last time we probed the condition.
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty" yaml:"lastProbeTime,omitempty"`
	// LastTransitionTime is the time since when the condition was in the given state.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" yaml:"lastTransitionTime,omitempty"`
}

// ScaleDownCondition contains information about scale down condition for a node group or the whole cluster.
type ScaleDownCondition struct {
	// Status of the scale down.
	Status ClusterAutoscalerConditionStatus `json:"status,omitempty" yaml:"status,omitempty"`
	// Candidates number for the scale down.
	Candidates int `json:"candidates,omitempty" yaml:"candidates,omitempty"`
	// LastProbeTime is the last time we probed the condition.
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty" yaml:"lastProbeTime,omitempty"`
	// LastTransitionTime is the time since when the condition was in the given state.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" yaml:"lastTransitionTime,omitempty"`
}

// ClusterWideStatus contains status that apply to the whole cluster.
type ClusterWideStatus struct {
	// Health contains information about health condition of the cluster.
	Health ClusterHealthCondition `json:"health,omitempty" yaml:"health,omitempty"`
	// ScaleUp contains information about scale up condition of the cluster.
	ScaleUp ClusterScaleUpCondition `json:"scaleUp,omitempty" yaml:"scaleUp,omitempty"`
	// ScaleDown contains information about scale down condition of the node group.
	ScaleDown ScaleDownCondition `json:"scaleDown,omitempty" yaml:"scaleDown,omitempty"`
}

// NodeGroupStatus contains status of an individual node group on which CA works..
type NodeGroupStatus struct {
	// Name of the node group.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Health contains information about health condition of the node group.
	Health NodeGroupHealthCondition `json:"health,omitempty" yaml:"health,omitempty"`
	// ScaleUp contains information about scale up condition of the node group.
	ScaleUp NodeGroupScaleUpCondition `json:"scaleUp,omitempty" yaml:"scaleUp,omitempty"`
	// ScaleDown contains information about scale down condition of the node group.
	ScaleDown ScaleDownCondition `json:"scaleDown,omitempty" yaml:"scaleDown,omitempty"`
}

// ClusterAutoscalerStatus contains ClusterAutoscaler status.
type ClusterAutoscalerStatus struct {
	// Time of the cluster autoscaler status.
	Time string `json:"time,omitempty" yaml:"time,omitempty"`
	// AutoscalerStatus contains status of ClusterAutoscaler (e.g. 'Initializing' & 'Running').
	AutoscalerStatus ClusterAutoscalerStatusCondition `json:"autoscalerStatus,omitempty" yaml:"autoscalerStatus,omitempty"`
	// Message contains extra information about the status.
	Message string `json:"message,omitempty" yaml:"message,omitempty"`
	// ClusterWide contains conditions that apply to the whole cluster.
	ClusterWide ClusterWideStatus `json:"clusterWide,omitempty" yaml:"clusterWide,omitempty"`
	// NodeGroups contains status information of individual node groups on which CA works.
	NodeGroups []NodeGroupStatus `json:"nodeGroups,omitempty" yaml:"nodeGroups,omitempty"`
}
