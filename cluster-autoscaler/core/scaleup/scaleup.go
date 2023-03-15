package scaleup

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/autoscaler/cluster-autoscaler/utils/taints"
	schedulerframework "k8s.io/kubernetes/pkg/scheduler/framework"
)

// ManagerFactory is a component that creates a new instance of the scale up manager.
type ManagerFactory interface {
	// NewManager builds a new instance of the scale up manager.
	NewManager(
		autoscalingContext *context.AutoscalingContext,
		processors *ca_processors.AutoscalingProcessors,
		clusterStateRegistry *clusterstate.ClusterStateRegistry,
		ignoredTaints taints.TaintKeySet,
	) Manager
}

// Manager is a component that picks the node group to resize and triggers
// creation of needed instances.
type Manager interface {
	// ScaleUp tries to scale the cluster up. Returns appropriate status or error if
	// an unexpected error occurred. Assumes that all nodes in the cluster are ready
	// and in sync with instance groups.
	ScaleUp(
		unschedulablePods []*apiv1.Pod,
		nodes []*apiv1.Node,
		daemonSets []*appsv1.DaemonSet,
		nodeInfos map[string]*schedulerframework.NodeInfo,
	) (*status.ScaleUpStatus, errors.AutoscalerError)
	// ScaleUpToNodeGroupMinSize tries to scale up node groups that have less nodes
	// than the configured min size. The source of truth for the current node group
	// size is the TargetSize queried directly from cloud providers. Returns
	// appropriate status or error if an unexpected error occurred.
	ScaleUpToNodeGroupMinSize(
		nodes []*apiv1.Node,
		nodeInfos map[string]*schedulerframework.NodeInfo,
	) (*status.ScaleUpStatus, errors.AutoscalerError)
}
