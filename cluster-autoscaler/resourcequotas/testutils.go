/*
Copyright 2025 The Kubernetes Authors.

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

// FakeQuota is a simple implementation of Quota for testing.
type FakeQuota struct {
	Name        string
	AppliesToFn func(*apiv1.Node) bool
	LimitsVal   map[string]int64
}

// ID returns the name of the quota.
func (f *FakeQuota) ID() string {
	return f.Name
}

// AppliesTo checks if a node applies to the quota, which is determined by the result of `AppliesToFn`.
func (f *FakeQuota) AppliesTo(node *apiv1.Node) bool {
	return f.AppliesToFn(node)
}

// Limits returns the limits defined by the quota.
func (f *FakeQuota) Limits() map[string]int64 {
	return f.LimitsVal
}

// MatchEveryNode returns true for every passed node.
func MatchEveryNode(_ *apiv1.Node) bool {
	return true
}

// FakeProvider is a fake implementation of Provider for testing.
type FakeProvider struct {
	quotas []Quota
	err    error
}

// Quotas returns quotas or error explicitly passed to the fake provider.
func (f *FakeProvider) Quotas() ([]Quota, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.quotas, nil
}

// NewFakeProvider returns a new FakeProvider with hard coded quotas.
func NewFakeProvider(quotas []Quota) *FakeProvider {
	return &FakeProvider{quotas: quotas}
}

// NewFailingProvider returns a new FakeProvider with an error.
func NewFailingProvider(err error) *FakeProvider {
	return &FakeProvider{err: err}
}
