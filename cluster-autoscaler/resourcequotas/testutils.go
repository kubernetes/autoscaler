package resourcequotas

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
	drasnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

type fakeNodeFilter struct {
	NodeFilterFn func(*apiv1.Node) bool
}

func (f *fakeNodeFilter) ExcludeFromTracking(node *apiv1.Node) bool {
	if f.NodeFilterFn == nil {
		return false
	}
	return f.NodeFilterFn(node)
}

type fakeCustomResourcesProcessor struct {
	NodeResourceTargets func(*apiv1.Node) []customresources.CustomResourceTarget
}

func (f *fakeCustomResourcesProcessor) FilterOutNodesWithUnreadyResources(context *context.AutoscalingContext, allNodes, readyNodes []*apiv1.Node, draSnapshot *drasnapshot.Snapshot) ([]*apiv1.Node, []*apiv1.Node) {
	return allNodes, readyNodes
}

func (f *fakeCustomResourcesProcessor) GetNodeResourceTargets(context *context.AutoscalingContext, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup) ([]customresources.CustomResourceTarget, errors.AutoscalerError) {
	if f.NodeResourceTargets == nil {
		return nil, nil
	}
	return f.NodeResourceTargets(node), nil
}

func (f *fakeCustomResourcesProcessor) CleanUp() {
}

type fakeQuota struct {
	id          string
	appliesToFn func(*apiv1.Node) bool
	limits      resourceList
}

func (f *fakeQuota) ID() string {
	return f.id
}

func (f *fakeQuota) AppliesTo(node *apiv1.Node) bool {
	return f.appliesToFn(node)
}

func (f *fakeQuota) Limits() map[string]int64 {
	return f.limits
}
