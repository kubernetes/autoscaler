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
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown"
	"k8s.io/autoscaler/cluster-autoscaler/debuggingsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	processor_callbacks "k8s.io/autoscaler/cluster-autoscaler/processors/callbacks"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	kube_client "k8s.io/client-go/kubernetes"
	kube_record "k8s.io/client-go/tools/record"
	klog "k8s.io/klog/v2"
)

// AutoscalingContext contains user-configurable constant and configuration-related objects passed to
// scale up/scale down functions.
type AutoscalingContext struct {
	// Options to customize how autoscaling works
	config.AutoscalingOptions
	// Kubernetes API clients.
	AutoscalingKubeClients
	// CloudProvider used in CA.
	CloudProvider cloudprovider.CloudProvider
	// TODO(kgolab) - move away too as it's not config
	// PredicateChecker to check if a pod can fit into a node.
	PredicateChecker predicatechecker.PredicateChecker
	// ClusterSnapshot denotes cluster snapshot used for predicate checking.
	ClusterSnapshot clustersnapshot.ClusterSnapshot
	// ExpanderStrategy is the strategy used to choose which node group to expand when scaling up
	ExpanderStrategy expander.Strategy
	// EstimatorBuilder is the builder function for node count estimator to be used.
	EstimatorBuilder estimator.EstimatorBuilder
	// ProcessorCallbacks is interface defining extra callback methods which can be called by processors used in extension points.
	ProcessorCallbacks processor_callbacks.ProcessorCallbacks
	// DebuggingSnapshotter is the interface for capturing the debugging snapshot
	DebuggingSnapshotter debuggingsnapshot.DebuggingSnapshotter
	// ScaleDownActuator is the interface for draining and deleting nodes
	ScaleDownActuator scaledown.Actuator
	// MaxNodes the maximum number of nodes in the whole cluster when MaxNodesTotal is zero
	MaxNodes int
}

// AutoscalingKubeClients contains all Kubernetes API clients,
// including listers and event recorders.
type AutoscalingKubeClients struct {
	// Listers.
	kube_util.ListerRegistry
	// ClientSet interface.
	ClientSet kube_client.Interface
	// Recorder for recording events.
	Recorder kube_record.EventRecorder
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
func NewAutoscalingContext(
	options config.AutoscalingOptions,
	predicateChecker predicatechecker.PredicateChecker,
	clusterSnapshot clustersnapshot.ClusterSnapshot,
	autoscalingKubeClients *AutoscalingKubeClients,
	cloudProvider cloudprovider.CloudProvider,
	expanderStrategy expander.Strategy,
	estimatorBuilder estimator.EstimatorBuilder,
	processorCallbacks processor_callbacks.ProcessorCallbacks,
	debuggingSnapshotter debuggingsnapshot.DebuggingSnapshotter) *AutoscalingContext {
	return &AutoscalingContext{
		AutoscalingOptions:     options,
		CloudProvider:          cloudProvider,
		AutoscalingKubeClients: *autoscalingKubeClients,
		PredicateChecker:       predicateChecker,
		ClusterSnapshot:        clusterSnapshot,
		ExpanderStrategy:       expanderStrategy,
		EstimatorBuilder:       estimatorBuilder,
		ProcessorCallbacks:     processorCallbacks,
		DebuggingSnapshotter:   debuggingSnapshotter,
	}
}

// NewAutoscalingKubeClients builds AutoscalingKubeClients out of basic client.
func NewAutoscalingKubeClients(opts config.AutoscalingOptions, kubeClient, eventsKubeClient kube_client.Interface) *AutoscalingKubeClients {
	listerRegistryStopChannel := make(chan struct{})
	listerRegistry := kube_util.NewListerRegistryWithDefaultListers(kubeClient, listerRegistryStopChannel)
	kubeEventRecorder := kube_util.CreateEventRecorder(eventsKubeClient, opts.RecordDuplicatedEvents)
	logRecorder, err := utils.NewStatusMapRecorder(kubeClient, opts.ConfigNamespace, kubeEventRecorder, opts.WriteStatusConfigMap, opts.StatusConfigMapName)
	if err != nil {
		klog.Error("Failed to initialize status configmap, unable to write status events")
		// Get a dummy, so we can at least safely call the methods
		// TODO(maciekpytel): recover from this after successful status configmap update?
		logRecorder, _ = utils.NewStatusMapRecorder(eventsKubeClient, opts.ConfigNamespace, kubeEventRecorder, false, opts.StatusConfigMapName)
	}

	return &AutoscalingKubeClients{
		ListerRegistry: listerRegistry,
		ClientSet:      kubeClient,
		Recorder:       kubeEventRecorder,
		LogRecorder:    logRecorder,
	}
}
