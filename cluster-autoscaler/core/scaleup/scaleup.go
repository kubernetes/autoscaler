/*
Copyright 2023 The Kubernetes Authors.

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

package scaleup

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
)

// Orchestrator is a component that picks the node group to resize and triggers
// creation of needed instances.
type Orchestrator interface {
	// Initialize initializes the orchestrator object with required fields.
	Initialize(
		autoscalingContext *context.AutoscalingContext,
		processors *ca_processors.AutoscalingProcessors,
		clusterStateRegistry *clusterstate.ClusterStateRegistry,
		estimatorBuilder estimator.EstimatorBuilder,
		taintConfig taints.TaintConfig,
	)
	// ScaleUp tries to scale the cluster up. Returns appropriate status or error if
	// an unexpected error occurred. Assumes that all nodes in the cluster are ready
	// and in sync with instance groups.
	ScaleUp(
		unschedulablePods []*apiv1.Pod,
		nodes []*apiv1.Node,
		daemonSets []*appsv1.DaemonSet,
		nodeInfos map[string]*framework.NodeInfo,
		allOrNothing bool,
	) (*status.ScaleUpStatus, errors.AutoscalerError)
	// ScaleUpToNodeGroupMinSize tries to scale up node groups that have less nodes
	// than the configured min size. The source of truth for the current node group
	// size is the TargetSize queried directly from cloud providers. Returns
	// appropriate status or error if an unexpected error occurred.
	ScaleUpToNodeGroupMinSize(
		nodes []*apiv1.Node,
		nodeInfos map[string]*framework.NodeInfo,
	) (*status.ScaleUpStatus, errors.AutoscalerError)
}
