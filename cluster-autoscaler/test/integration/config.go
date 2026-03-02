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

package integration

import (
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"time"
)

// DefaultAutoscalingOptions provides the baseline configuration for all tests.
var DefaultAutoscalingOptions = config.AutoscalingOptions{
	NodeGroupDefaults: config.NodeGroupAutoscalingOptions{
		ScaleDownUnneededTime:         time.Second,
		ScaleDownUnreadyTime:          time.Minute,
		ScaleDownUtilizationThreshold: 0.5,
		MaxNodeProvisionTime:          10 * time.Second,
	},
	EstimatorName:              estimator.BinpackingEstimatorName,
	EnforceNodeGroupMinSize:    true,
	ScaleDownSimulationTimeout: 24 * time.Hour,
	ScaleDownDelayAfterAdd:     0,
	ScaleDownDelayAfterDelete:  0,
	ScaleDownDelayAfterFailure: 0,
	ScaleDownDelayTypeLocal:    true,
	ScaleDownEnabled:           true,
	MaxNodesTotal:              10,
	MaxCoresTotal:              10,
	MaxMemoryTotal:             100000,
	ExpanderNames:              "least-waste",
	ScaleUpFromZero:            true,
	FrequentLoopsEnabled:       true,
	ClusterName:                "cluster-test",
	MaxBinpackingTime:          10 * time.Second,
}

// TestConfig is the "blueprint" for a test. It defines the entire
// initial state of the world before the test runs.
type TestConfig struct {
	// BaseOptions can be set to DefaultAutoscalingOptions or a custom base.
	BaseOptions *config.AutoscalingOptions
	// OptionsOverrides allows adding options overrides.
	OptionsOverrides []AutoscalingOptionOverride
}

// NewTestConfig creates a test config pre-populated with DefaultAutoscalingOptions.
func NewTestConfig() *TestConfig {
	return &TestConfig{
		BaseOptions:      &DefaultAutoscalingOptions,
		OptionsOverrides: []AutoscalingOptionOverride{},
	}
}

// AutoscalingOptionOverride is a function that modifies an AutoscalingOptions object.
type AutoscalingOptionOverride func(*config.AutoscalingOptions)

// WithOverrides allows adding options overrides to the config.
func (c *TestConfig) WithOverrides(overrides ...AutoscalingOptionOverride) *TestConfig {
	c.OptionsOverrides = append(c.OptionsOverrides, overrides...)
	return c
}

// ResolveOptions merges the base options with all registered overrides.
func (c *TestConfig) ResolveOptions() config.AutoscalingOptions {
	var opts config.AutoscalingOptions
	if c.BaseOptions != nil {
		opts = *c.BaseOptions
	} else {
		opts = DefaultAutoscalingOptions
	}

	for _, override := range c.OptionsOverrides {
		override(&opts)
	}
	return opts
}

// WithCloudProviderName sets the cloud provider name.
func WithCloudProviderName(name string) AutoscalingOptionOverride {
	return func(o *config.AutoscalingOptions) {
		o.CloudProviderName = name
	}
}

// WithScaleDownUnneededTime sets the scale down unneeded time option.
func WithScaleDownUnneededTime(d time.Duration) AutoscalingOptionOverride {
	return func(o *config.AutoscalingOptions) {
		o.NodeGroupDefaults.ScaleDownUnneededTime = d
	}
}

// WithScaleDownDelayAfterAdd sets the scale down delay after add option.
func WithScaleDownDelayAfterAdd(d time.Duration) AutoscalingOptionOverride {
	return func(o *config.AutoscalingOptions) {
		o.ScaleDownDelayAfterAdd = d
	}
}

// WithScaleDownDelayAfterDelete sets the scale down delay after delete option.
func WithScaleDownDelayAfterDelete(d time.Duration) AutoscalingOptionOverride {
	return func(o *config.AutoscalingOptions) {
		o.ScaleDownDelayAfterDelete = d
	}
}

// WithScaleDownDelayAfterFailure sets the scale down delay after failure option.
func WithScaleDownDelayAfterFailure(d time.Duration) AutoscalingOptionOverride {
	return func(o *config.AutoscalingOptions) {
		o.ScaleDownDelayAfterFailure = d
	}
}

// WithMaxCoresTotal sets the maximum total cores.
func WithMaxCoresTotal(n int64) AutoscalingOptionOverride {
	return func(o *config.AutoscalingOptions) {
		o.MaxCoresTotal = n
	}
}

// WithMaxScaleDownParallelism sets the maximum scale down parallelism.
func WithMaxScaleDownParallelism(sdp int) AutoscalingOptionOverride {
	return func(o *config.AutoscalingOptions) {
		o.MaxScaleDownParallelism = sdp
	}
}

// WithMaxDrainParallelism sets the maximum drain parallelism.
func WithMaxDrainParallelism(dp int) AutoscalingOptionOverride {
	return func(o *config.AutoscalingOptions) {
		o.MaxDrainParallelism = dp
	}
}

// WithExpendablePodsPriorityCutoff sets the priority cutoff to expandable pods.
func WithExpendablePodsPriorityCutoff(dp int) AutoscalingOptionOverride {
	return func(o *config.AutoscalingOptions) {
		o.ExpendablePodsPriorityCutoff = dp
	}
}

// WithMaxTotalUnreadyPercentage sets maximum percentage of unready nodes in the cluster
func WithMaxTotalUnreadyPercentage(maxTotalUnreadyPercentage float64) AutoscalingOptionOverride {
	return func(o *config.AutoscalingOptions) {
		o.MaxTotalUnreadyPercentage = maxTotalUnreadyPercentage
	}
}

// WithForceDeleteFailedNodes sets ForceDeleteFailedNodes.
func WithForceDeleteFailedNodes(b bool) AutoscalingOptionOverride {
	return func(o *config.AutoscalingOptions) {
		o.ForceDeleteFailedNodes = b
	}
}

// WithMaxNodeGroupBinpackingDuration sets MaxNodeGroupBinpackingDuration.
func WithMaxNodeGroupBinpackingDuration(d time.Duration) AutoscalingOptionOverride {
	return func(o *config.AutoscalingOptions) {
		o.MaxNodeGroupBinpackingDuration = d
	}
}

// WithOkTotalUnreadyCount sets OkTotalUnreadyCount.
func WithOkTotalUnreadyCount(cnt int) AutoscalingOptionOverride {
	return func(o *config.AutoscalingOptions) {
		o.OkTotalUnreadyCount = cnt
	}
}

// WithZeroOrMaxNodeScaling sets the zero or max node scaling option.
func WithZeroOrMaxNodeScaling(b bool) AutoscalingOptionOverride {
	return func(o *config.AutoscalingOptions) {
		o.NodeGroupDefaults.ZeroOrMaxNodeScaling = b
	}
}

// WithAllowNonAtomicScaleUpToMax sets the allow non atomic scaleup option.
func WithAllowNonAtomicScaleUpToMax(b bool) AutoscalingOptionOverride {
	return func(o *config.AutoscalingOptions) {
		o.NodeGroupDefaults.AllowNonAtomicScaleUpToMax = b
	}
}
