package resourcelimits

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
	drasnapshot "k8s.io/autoscaler/cluster-autoscaler/simulator/dynamicresources/snapshot"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
)

// fakeProvider is a fake implementation of a resource limit provider.
type fakeProvider struct {
	limiters []Limiter
}

// NewFakeProvider creates a new fakeProvider.
func NewFakeProvider() *fakeProvider {
	return &fakeProvider{}
}

// AddLimiter adds a limiter to the provider.
func (p *fakeProvider) AddLimiter(limiter Limiter) {
	p.limiters = append(p.limiters, limiter)
}

// AllLimiters returns all limiters from the provider.
func (p *fakeProvider) AllLimiters() ([]Limiter, error) {
	return p.limiters, nil
}

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
	NodeResourceTargets         func(*apiv1.Node) []customresources.CustomResourceTarget
	GetNodeResourceTargetsError errors.AutoscalerError
}

func (f *fakeCustomResourcesProcessor) FilterOutNodesWithUnreadyResources(context *context.AutoscalingContext, allNodes, readyNodes []*apiv1.Node, draSnapshot *drasnapshot.Snapshot) ([]*apiv1.Node, []*apiv1.Node) {
	return allNodes, readyNodes
}

func (f *fakeCustomResourcesProcessor) GetNodeResourceTargets(context *context.AutoscalingContext, node *apiv1.Node, nodeGroup cloudprovider.NodeGroup) ([]customresources.CustomResourceTarget, errors.AutoscalerError) {
	if f.GetNodeResourceTargetsError != nil {
		return nil, f.GetNodeResourceTargetsError
	}
	if f.NodeResourceTargets == nil {
		return nil, nil
	}
	return f.NodeResourceTargets(node), nil
}

func (f *fakeCustomResourcesProcessor) CleanUp() {
}

type fakeLimiter struct {
	id          string
	appliesToFn func(*apiv1.Node) bool
	minLimits   resourceList
	maxLimits   resourceList
}

func (f *fakeLimiter) ID() string {
	return f.id
}

func (f *fakeLimiter) AppliesTo(node *apiv1.Node) bool {
	return f.appliesToFn(node)
}

func (f *fakeLimiter) MaxLimits() map[string]int64 {
	return f.maxLimits
}

func (f *fakeLimiter) MinLimits() map[string]int64 {
	return f.minLimits
}
