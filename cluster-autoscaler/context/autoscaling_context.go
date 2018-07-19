/*
Copyright 2016 The Kubernetes Authors.

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

package context

import (
	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	"k8s.io/autoscaler/cluster-autoscaler/clusterstate/utils"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/expander/factory"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	kube_client "k8s.io/client-go/kubernetes"
	kube_record "k8s.io/client-go/tools/record"
)

// AutoscalingContext contains user-configurable constant and configuration-related objects passed to
// scale up/scale down functions.
type AutoscalingContext struct {
	// Options to customize how autoscaling works
	config.AutoscalingOptions
	// CloudProvider used in CA.
	CloudProvider cloudprovider.CloudProvider
	// ClientSet interface.
	ClientSet kube_client.Interface
	// Recorder for recording events.
	Recorder kube_record.EventRecorder
	// TODO(kgolab) - move away too as it's not config
	// PredicateChecker to check if a pod can fit into a node.
	PredicateChecker *simulator.PredicateChecker
	// ExpanderStrategy is the strategy used to choose which node group to expand when scaling up
	ExpanderStrategy expander.Strategy
	// LogRecorder can be used to collect log messages to expose via Events on some central object.
	LogRecorder *utils.LogEventRecorder
}

// NewResourceLimiterFromAutoscalingOptions creates new instance of cloudprovider.ResourceLimiter
// reading limits from AutoscalingOptions struct.
func NewResourceLimiterFromAutoscalingOptions(options config.AutoscalingOptions) *cloudprovider.ResourceLimiter {
	// build min/max maps for resources limits
	minResources := make(map[string]int64)
	maxResources := make(map[string]int64)

	minResources[cloudprovider.ResourceNameCores] = options.MinCoresTotal
	minResources[cloudprovider.ResourceNameMemory] = options.MinMemoryTotal
	maxResources[cloudprovider.ResourceNameCores] = options.MaxCoresTotal
	maxResources[cloudprovider.ResourceNameMemory] = options.MaxMemoryTotal

	for _, gpuLimits := range options.GpuTotal {
		minResources[gpuLimits.GpuType] = gpuLimits.Min
		maxResources[gpuLimits.GpuType] = gpuLimits.Max
	}
	return cloudprovider.NewResourceLimiter(minResources, maxResources)
}

// NewAutoscalingContext returns an autoscaling context from all the necessary parameters passed via arguments
func NewAutoscalingContext(options config.AutoscalingOptions, predicateChecker *simulator.PredicateChecker,
	kubeClient kube_client.Interface, kubeEventRecorder kube_record.EventRecorder,
	logEventRecorder *utils.LogEventRecorder, listerRegistry kube_util.ListerRegistry,
	cloudProvider cloudprovider.CloudProvider) (*AutoscalingContext, errors.AutoscalerError) {
	expanderStrategy, err := factory.ExpanderStrategyFromString(options.ExpanderName,
		cloudProvider, listerRegistry.AllNodeLister())
	if err != nil {
		return nil, err
	}

	autoscalingContext := AutoscalingContext{
		AutoscalingOptions: options,
		CloudProvider:      cloudProvider,
		ClientSet:          kubeClient,
		Recorder:           kubeEventRecorder,
		PredicateChecker:   predicateChecker,
		ExpanderStrategy:   expanderStrategy,
		LogRecorder:        logEventRecorder,
	}

	return &autoscalingContext, nil
}
