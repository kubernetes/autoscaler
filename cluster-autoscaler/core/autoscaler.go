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

package core

import (
	"strings"
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/cloudprovider"
	cloudBuilder "k8s.io/autoscaler/cluster-autoscaler/cloudprovider/builder"
	"k8s.io/autoscaler/cluster-autoscaler/config"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaledown/pdb"
	"k8s.io/autoscaler/cluster-autoscaler/core/scaleup"
	"k8s.io/autoscaler/cluster-autoscaler/debuggingsnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/estimator"
	"k8s.io/autoscaler/cluster-autoscaler/expander"
	"k8s.io/autoscaler/cluster-autoscaler/expander/factory"
	"k8s.io/autoscaler/cluster-autoscaler/observers/loopstart"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/clustersnapshot/predicate"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/drainability/rules"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/framework"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/options"
	"k8s.io/autoscaler/cluster-autoscaler/simulator/predicatechecker"
	"k8s.io/autoscaler/cluster-autoscaler/utils/backoff"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	"k8s.io/client-go/informers"
	kube_client "k8s.io/client-go/kubernetes"
)

// AutoscalerOptions is the whole set of options for configuring an autoscaler
type AutoscalerOptions struct {
	config.AutoscalingOptions
	KubeClient             kube_client.Interface
	InformerFactory        informers.SharedInformerFactory
	AutoscalingKubeClients *context.AutoscalingKubeClients
	CloudProvider          cloudprovider.CloudProvider
	FrameworkHandle        *framework.Handle
	PredicateChecker       predicatechecker.PredicateChecker
	ClusterSnapshot        clustersnapshot.ClusterSnapshot
	ExpanderStrategy       expander.Strategy
	EstimatorBuilder       estimator.EstimatorBuilder
	Processors             *ca_processors.AutoscalingProcessors
	LoopStartNotifier      *loopstart.ObserversList
	Backoff                backoff.Backoff
	DebuggingSnapshotter   debuggingsnapshot.DebuggingSnapshotter
	RemainingPdbTracker    pdb.RemainingPdbTracker
	ScaleUpOrchestrator    scaleup.Orchestrator
	DeleteOptions          options.NodeDeleteOptions
	DrainabilityRules      rules.Rules
}

// Autoscaler is the main component of CA which scales up/down node groups according to its configuration
// The configuration can be injected at the creation of an autoscaler
type Autoscaler interface {
	// Start starts components running in background.
	Start() error
	// RunOnce represents an iteration in the control-loop of CA
	RunOnce(currentTime time.Time) errors.AutoscalerError
	// ExitCleanUp is a clean-up performed just before process termination.
	ExitCleanUp()
	// LastScaleUpTime is a time of the last scale up
	LastScaleUpTime() time.Time
	// LastScaleUpTime is a time of the last scale down
	LastScaleDownDeleteTime() time.Time
}

// NewAutoscaler creates an autoscaler of an appropriate type according to the parameters
func NewAutoscaler(opts AutoscalerOptions, informerFactory informers.SharedInformerFactory) (Autoscaler, errors.AutoscalerError) {
	err := initializeDefaultOptions(&opts, informerFactory)
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}
	return NewStaticAutoscaler(
		opts.AutoscalingOptions,
		opts.FrameworkHandle,
		opts.PredicateChecker,
		opts.ClusterSnapshot,
		opts.AutoscalingKubeClients,
		opts.Processors,
		opts.LoopStartNotifier,
		opts.CloudProvider,
		opts.ExpanderStrategy,
		opts.EstimatorBuilder,
		opts.Backoff,
		opts.DebuggingSnapshotter,
		opts.RemainingPdbTracker,
		opts.ScaleUpOrchestrator,
		opts.DeleteOptions,
		opts.DrainabilityRules,
	), nil
}

// Initialize default options if not provided.
func initializeDefaultOptions(opts *AutoscalerOptions, informerFactory informers.SharedInformerFactory) error {
	if opts.Processors == nil {
		opts.Processors = ca_processors.DefaultProcessors(opts.AutoscalingOptions)
	}
	if opts.LoopStartNotifier == nil {
		opts.LoopStartNotifier = loopstart.NewObserversList(nil)
	}
	if opts.AutoscalingKubeClients == nil {
		opts.AutoscalingKubeClients = context.NewAutoscalingKubeClients(opts.AutoscalingOptions, opts.KubeClient, opts.InformerFactory)
	}
	if opts.FrameworkHandle == nil {
		fwHandle, err := framework.NewHandle(opts.InformerFactory, opts.SchedulerConfig)
		if err != nil {
			return err
		}
		opts.FrameworkHandle = fwHandle
	}
	if opts.ClusterSnapshot == nil {
		opts.ClusterSnapshot = predicate.NewPredicateSnapshot(clustersnapshot.NewBasicClusterSnapshot(), opts.FrameworkHandle)
	}
	if opts.RemainingPdbTracker == nil {
		opts.RemainingPdbTracker = pdb.NewBasicRemainingPdbTracker()
	}
	if opts.CloudProvider == nil {
		opts.CloudProvider = cloudBuilder.NewCloudProvider(opts.AutoscalingOptions, informerFactory)
	}
	if opts.ExpanderStrategy == nil {
		expanderFactory := factory.NewFactory()
		expanderFactory.RegisterDefaultExpanders(opts.CloudProvider, opts.AutoscalingKubeClients, opts.KubeClient, opts.ConfigNamespace, opts.GRPCExpanderCert, opts.GRPCExpanderURL)
		expanderStrategy, err := expanderFactory.Build(strings.Split(opts.ExpanderNames, ","))
		if err != nil {
			return err
		}
		opts.ExpanderStrategy = expanderStrategy
	}
	if opts.EstimatorBuilder == nil {
		thresholds := []estimator.Threshold{
			estimator.NewStaticThreshold(opts.MaxNodesPerScaleUp, opts.MaxNodeGroupBinpackingDuration),
			estimator.NewSngCapacityThreshold(),
			estimator.NewClusterCapacityThreshold(),
		}
		estimatorBuilder, err := estimator.NewEstimatorBuilder(
			opts.EstimatorName,
			estimator.NewThresholdBasedEstimationLimiter(thresholds),
			estimator.NewDecreasingPodOrderer(),
			/* EstimationAnalyserFunc */ nil,
		)
		if err != nil {
			return err
		}
		opts.EstimatorBuilder = estimatorBuilder
	}
	if opts.Backoff == nil {
		opts.Backoff =
			backoff.NewIdBasedExponentialBackoff(opts.InitialNodeGroupBackoffDuration, opts.MaxNodeGroupBackoffDuration, opts.NodeGroupBackoffResetTimeout)
	}
	if opts.DrainabilityRules == nil {
		opts.DrainabilityRules = rules.Default(opts.DeleteOptions)
	}

	return nil
}
