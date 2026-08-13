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
	gocontext "context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/cluster-autoscaler/pkg/context"
	"sigs.k8s.io/cluster-autoscaler/pkg/processors/customresources"
)

// TrackerFactory builds quota trackers.
type TrackerFactory struct {
	crp            customresources.CustomResourcesProcessor
	quotasProvider Provider
	nodeFilter     NodeFilter
}

// TrackerOptions stores configuration for quota tracking.
type TrackerOptions struct {
	CustomResourcesProcessor customresources.CustomResourcesProcessor
	QuotaProvider            Provider
	NodeFilter               NodeFilter
}

// NewTrackerFactory creates a new TrackerFactory.
func NewTrackerFactory(opts TrackerOptions) *TrackerFactory {
	return &TrackerFactory{
		crp:            opts.CustomResourcesProcessor,
		quotasProvider: opts.QuotaProvider,
		nodeFilter:     opts.NodeFilter,
	}
}

// NewMaxQuotasTracker builds a new Tracker for maximum limits enforcement.
//
// NewMaxQuotasTracker calculates resources used by the nodes for every
// quota returned by the Provider. Then, based on usages and limits it calculates
// how many resources can be still added to the cluster. Returns a Tracker object.
func (f *TrackerFactory) NewMaxQuotasTracker(ctx gocontext.Context, autoscalingCtx *context.AutoscalingContext, nodes []*corev1.Node) (*Tracker, error) {
	return f.newQuotasTracker(ctx, autoscalingCtx, nodes, false /* isMinEnforcement */)
}

// NewQuotasTracker builds a new Tracker for maximum limits enforcement.
//
// Deprecated: Use NewMaxQuotasTracker instead.
func (f *TrackerFactory) NewQuotasTracker(ctx gocontext.Context, autoscalingCtx *context.AutoscalingContext, nodes []*corev1.Node) (*Tracker, error) {
	return f.NewMaxQuotasTracker(ctx, autoscalingCtx, nodes)
}

// NewMinQuotasTracker builds a new Tracker for minimum limits enforcement.
//
// NewMinQuotasTracker calculates resources used by the nodes for every
// quota returned by the Provider. Then, based on usages and limits it calculates
// how many resources can be still removed from the cluster. Returns a Tracker object.
func (f *TrackerFactory) NewMinQuotasTracker(ctx gocontext.Context, autoscalingCtx *context.AutoscalingContext, nodes []*corev1.Node) (*Tracker, error) {
	return f.newQuotasTracker(ctx, autoscalingCtx, nodes, true /* isMinEnforcement */)
}

// newQuotasTracker builds a new Tracker for either minimum or maximum limits enforcement.
// If isMinEnforcement is true, it calculates headroom to the minimum limit (for scale-down).
// Otherwise, it calculates headroom to the maximum limit (for scale-up).
func (f *TrackerFactory) newQuotasTracker(ctx gocontext.Context, autoscalingCtx *context.AutoscalingContext, nodes []*corev1.Node, isMinEnforcement bool) (*Tracker, error) {
	logger := klog.FromContext(ctx)
	quotas, err := f.quotasProvider.Quotas(ctx)
	if err != nil {
		return nil, err
	}
	nc := newNodeResourcesCache(f.crp)
	uc := newUsageCalculator(f.nodeFilter, nc)
	usages, err := uc.calculateUsages(ctx, autoscalingCtx, nodes, quotas)
	if err != nil {
		return nil, err
	}
	var quotaStatuses []*quotaStatus
	for _, rq := range quotas {
		logger.V(5).Info("Logging quota status", "resourceQuotaId", rq.ID(), "limits", rq.Limits(), "usages", usages[rq.ID()])
		limitsLeft := make(resourceList)
		limits := rq.Limits()
		for resourceType, limit := range limits {
			usage := usages[rq.ID()][resourceType]
			if isMinEnforcement {
				limitsLeft[resourceType] = max(0, usage-limit)
			} else {
				limitsLeft[resourceType] = max(0, limit-usage)
			}
		}
		quotaStatuses = append(quotaStatuses, &quotaStatus{
			quota:      rq,
			limitsLeft: limitsLeft,
		})
	}
	tracker := newTracker(quotaStatuses, nc)
	return tracker, nil
}
