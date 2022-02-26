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

// ClusterAutoscalerConditionType is the type of ClusterAutoscalerCondition.
type ClusterAutoscalerConditionType string

const (
	// ClusterAutoscalerHealth - is a condition that explains what is the current health
	// of ClusterAutoscaler or its node groups.
	ClusterAutoscalerHealth ClusterAutoscalerConditionType = "Health"
	// ClusterAutoscalerScaleDown is a condition that explains what is the current status
	// of a node group with regard to scale down activities.
	ClusterAutoscalerScaleDown ClusterAutoscalerConditionType = "ScaleDown"
	// ClusterAutoscalerScaleUp is a condition that explains what is the current status
	// of a node group with regard to scale up activities.
	ClusterAutoscalerScaleUp ClusterAutoscalerConditionType = "ScaleUp"
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
)

// ClusterAutoscalerCondition describes some aspect of ClusterAutoscaler work.
type ClusterAutoscalerCondition struct {
	// Type defines the aspect that the condition describes. For example, it can be Health or ScaleUp/Down activity.
	Type ClusterAutoscalerConditionType `json:"type,omitempty"`
	// Status of the condition.
	Status ClusterAutoscalerConditionStatus `json:"status,omitempty"`
	// Message is a free text extra information about the condition. It may contain some
	// extra debugging data, like why the cluster is unhealthy.
	Message string `json:"message,omitempty"`
	// Reason is a unique, one-word, CamelCase reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// LastProbeTime is the last time we probed the condition.
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`
	// LastTransitionTime is the time since when the condition was in the given state.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// ClusterAutoscalerStatus contains ClusterAutoscaler status.
type ClusterAutoscalerStatus struct {
	// NodeGroupStatuses contains status information of individual node groups on which CA works.
	NodeGroupStatuses []NodeGroupStatus `json:"nodeGroupStatuses,omitempty"`
	// ClusterwideConditions contains conditions that apply to the whole autoscaler.
	ClusterwideConditions []ClusterAutoscalerCondition `json:"clusterwideConditions,omitempty"`
}

// NodeGroupStatus contains status of a group of nodes controlled by ClusterAutoscaler.
type NodeGroupStatus struct {
	// ProviderID is the cloud-provider-specific name of the node group. On GCE it will be equal
	// to MIG url, on AWS it will be ASG name, etc.

	// ProviderID string `json:"providerID,omitempty"`

	// Conditions is a list of conditions that describe the state of the node group.
	Conditions []ClusterAutoscalerCondition `json:"conditions,omitempty"`
}
