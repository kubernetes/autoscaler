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
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	fakecloudprovider "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/test"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	fakek8s "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes/fake"
	"k8s.io/client-go/kubernetes"
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
	MaxBinpackingTime:          5 * time.Minute,
}

// Config is the "blueprint" for a test. It defines the entire
// initial state of the world before the test runs.
type Config struct {
	AutoscalingOptions *config.AutoscalingOptions
	CloudProvider      cloudprovider.CloudProvider
	KubeClient         kubernetes.Interface
}

// Builder uses the Builder Pattern to create a test Config.
type Builder struct {
	defaultOptions  config.AutoscalingOptions
	optionOverrides []AutoscalingOptionOverride
	cloudProvider   cloudprovider.CloudProvider
	kubeClient      kubernetes.Interface
}

// AutoscalingOptionOverride is a function that modifies an AutoscalingOptions object.
type AutoscalingOptionOverride func(*config.AutoscalingOptions)

// NewBuilder creates a new builder, pre-filled with all the standard autoscaler defaults.
func NewBuilder() *Builder {
	return &Builder{
		defaultOptions:  DefaultAutoscalingOptions,
		optionOverrides: []AutoscalingOptionOverride{},
	}
}

// WithCloudProvider allows injecting specific implementations of cloud provider.
func (b *Builder) WithCloudProvider(cp cloudprovider.CloudProvider) *Builder {
	b.cloudProvider = cp
	return b
}

// WithKubeClient allows injecting specific implementations of kube client.
func (b *Builder) WithKubeClient(kc kubernetes.Interface) *Builder {
	b.kubeClient = kc
	return b
}

// WithAutoscalingOptions now takes the base and the specific overrides.
func (b *Builder) WithAutoscalingOptions(base config.AutoscalingOptions, overrides ...AutoscalingOptionOverride) *Builder {
	b.defaultOptions = base
	b.optionOverrides = append(b.optionOverrides, overrides...)
	return b
}

// WithCloudProviderName sets the cloud provider name.
func WithCloudProviderName(name string) AutoscalingOptionOverride {
	return func(o *config.AutoscalingOptions) {
		o.CloudProviderName = name
	}
}

func WithScaleDownUnneededTime(d time.Duration) AutoscalingOptionOverride {
	return func(o *config.AutoscalingOptions) {
		o.NodeGroupDefaults.ScaleDownUnneededTime = d
	}
}

func WithScaleDownDelayAfterAdd(d time.Duration) AutoscalingOptionOverride {
	return func(o *config.AutoscalingOptions) {
		o.ScaleDownDelayAfterAdd = d
	}
}

// Build returns the final, constructed Config.
func (b *Builder) Build() *Config {
	opts := b.defaultOptions

	for _, override := range b.optionOverrides {
		override(&opts)
	}

	cp := b.cloudProvider
	if cp == nil {
		cp = fakecloudprovider.NewCloudProvider()
	}

	kc := b.kubeClient
	if kc == nil {
		kc = fakek8s.NewKubernetes().KubeClient
	}
	return &Config{
		AutoscalingOptions: &opts,
		CloudProvider:      cp,
		KubeClient:         kc,
	}
}
