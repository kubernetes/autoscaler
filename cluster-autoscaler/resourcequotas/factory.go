package resourcequotas

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
)

// TrackerFactory builds trackers.
type TrackerFactory struct {
	crp             customresources.CustomResourcesProcessor
	quotaProviders  []Provider
	usageCalculator *usageCalculator
}

type TrackerOptions struct {
	CRP        customresources.CustomResourcesProcessor
	Providers  []Provider
	NodeFilter NodeFilter
}

// NewTrackerFactory creates a new TrackerFactory.
func NewTrackerFactory(opts TrackerOptions) *TrackerFactory {
	uc := newUsageCalculator(opts.CRP, opts.NodeFilter)
	return &TrackerFactory{
		crp:             opts.CRP,
		quotaProviders:  opts.Providers,
		usageCalculator: uc,
	}
}

// NewQuotasTracker builds a new Tracker.
func (f *TrackerFactory) NewQuotasTracker(ctx *context.AutoscalingContext, nodes []*corev1.Node) (*Tracker, error) {
	quotas, err := f.quotas()
	if err != nil {
		return nil, err
	}
	usages, err := f.usageCalculator.calculateUsages(ctx, nodes, quotas)
	if err != nil {
		return nil, err
	}
	limitsLeft := make(map[string]resourceList)
	for _, rq := range quotas {
		limitsLeft[rq.ID()] = make(resourceList)
		limits := rq.Limits()
		for resourceType, limit := range limits {
			usage := usages[rq.ID()][resourceType]
			limitsLeft[rq.ID()][resourceType] = max(0, limit-usage)
		}
	}
	tracker := newTracker(f.crp, quotas, limitsLeft)
	return tracker, nil
}

func (f *TrackerFactory) quotas() ([]Quota, error) {
	var quotas []Quota
	for _, provider := range f.quotaProviders {
		provQuotas, err := provider.Quotas()
		if err != nil {
			return nil, fmt.Errorf("failed to get quotas from provider: %w", err)
		}
		for _, rq := range provQuotas {
			quotas = append(quotas, rq)
		}
	}
	return quotas, nil
}
