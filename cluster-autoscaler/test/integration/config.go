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
