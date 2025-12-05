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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
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

// NewQuotasTracker builds a new Tracker.
//
// NewQuotasTracker calculates resources used by the nodes for every
// quota returned by the Provider. Then, based on usages and limits it calculates
// how many resources can be still added to the cluster. Returns a Tracker object.
func (f *TrackerFactory) NewQuotasTracker(autoscalingCtx *context.AutoscalingContext, nodes []*corev1.Node) (*Tracker, error) {
	quotas, err := f.quotasProvider.Quotas()
	if err != nil {
		return nil, err
	}
	nc := newNodeResourcesCache(f.crp)
	uc := newUsageCalculator(f.nodeFilter, nc)
	usages, err := uc.calculateUsages(autoscalingCtx, nodes, quotas)
	if err != nil {
		return nil, err
	}
	var quotaStatuses []*quotaStatus
	for _, rq := range quotas {
		limitsLeft := make(resourceList)
		limits := rq.Limits()
		for resourceType, limit := range limits {
			usage := usages[rq.ID()][resourceType]
			limitsLeft[resourceType] = max(0, limit-usage)
		}
		quotaStatuses = append(quotaStatuses, &quotaStatus{
			quota:      rq,
			limitsLeft: limitsLeft,
		})
	}
	tracker := newTracker(quotaStatuses, nc)
	return tracker, nil
}
