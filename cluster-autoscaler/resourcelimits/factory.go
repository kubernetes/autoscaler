package resourcelimits

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/processors/customresources"
)

// TrackerFactory builds trackers.
type TrackerFactory struct {
	crp             customresources.CustomResourcesProcessor
	limitProviders  []Provider
	usageCalculator *usageCalculator
}

// NewTrackerFactory creates a new TrackerFactory.
func NewTrackerFactory(opts TrackerOptions) *TrackerFactory {
	uc := newUsageCalculator(opts.CRP, opts.NodeFilter)
	return &TrackerFactory{
		crp:             opts.CRP,
		limitProviders:  opts.Providers,
		usageCalculator: uc,
	}
}

// NewMaxLimitsTracker builds a new Tracker for max limits.
func (f *TrackerFactory) NewMaxLimitsTracker(ctx *context.AutoscalingContext, nodes []*corev1.Node) (*Tracker, error) {
	return f.newLimitsTracker(ctx, nodes, &maxLimitsStrategy{})
}

// NewMinLimitsTracker builds a new Tracker for min limits.
func (f *TrackerFactory) NewMinLimitsTracker(ctx *context.AutoscalingContext, nodes []*corev1.Node) (*Tracker, error) {
	return f.newLimitsTracker(ctx, nodes, &minLimitsStrategy{})
}

func (f *TrackerFactory) newLimitsTracker(ctx *context.AutoscalingContext, nodes []*corev1.Node, strategy limitStrategy) (*Tracker, error) {
	limiters, err := f.limiters()
	if err != nil {
		return nil, err
	}
	usages, err := f.usageCalculator.calculateUsages(ctx, nodes, limiters)
	if err != nil {
		return nil, err
	}
	limitsLeft := make(map[string]resourceList)
	for _, rl := range limiters {
		limitsLeft[rl.ID()] = make(resourceList)
		limits := strategy.GetLimits(rl)
		for resourceType, limit := range limits {
			usage := usages[rl.ID()][resourceType]
			limitsLeft[rl.ID()][resourceType] = strategy.CalculateLimitsLeft(limit, usage)
		}
	}
	tracker := newTracker(f.crp, limiters, limitsLeft)
	return tracker, nil
}

func (f *TrackerFactory) limiters() ([]Limiter, error) {
	var limiters []Limiter
	for _, provider := range f.limitProviders {
		provLimiters, err := provider.AllLimiters()
		if err != nil {
			return nil, fmt.Errorf("failed to get limiters from provider: %w", err)
		}
		for _, limiter := range provLimiters {
			limiters = append(limiters, limiter)
		}
	}
	return limiters, nil
}

// limitStrategy is an interface for defining limit calculation strategies.
type limitStrategy interface {
	GetLimits(rl Limiter) resourceList
	CalculateLimitsLeft(limit, usage int64) int64
}

// maxLimitsStrategy is a strategy for max limits.
type maxLimitsStrategy struct{}

// GetLimits returns max limits.
func (s *maxLimitsStrategy) GetLimits(rl Limiter) resourceList {
	return rl.MaxLimits()
}

// CalculateLimitsLeft calculates the remaining limits for max limits.
func (s *maxLimitsStrategy) CalculateLimitsLeft(limit, usage int64) int64 {
	return max(0, limit-usage)
}

// minLimitsStrategy is a strategy for min limits.
type minLimitsStrategy struct{}

// GetLimits returns min limits.
func (s *minLimitsStrategy) GetLimits(rl Limiter) resourceList {
	return rl.MinLimits()
}

// CalculateLimitsLeft calculates the remaining limits for min limits.
func (s *minLimitsStrategy) CalculateLimitsLeft(limit, usage int64) int64 {
	return max(0, usage-limit)
}
