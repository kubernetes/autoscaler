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
	"time"

	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/context"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	"k8s.io/autoscaler/cluster-autoscaler/utils/pods"
	kube_client "k8s.io/client-go/kubernetes"
	kube_record "k8s.io/client-go/tools/record"
)

// AutoscalerOptions is the whole set of options for configuring an autoscaler
type AutoscalerOptions struct {
	context.AutoscalingOptions
	dynamic.ConfigFetcherOptions
	KubeClient        kube_client.Interface
	KubeEventRecorder kube_record.EventRecorder
	PredicateChecker  *simulator.PredicateChecker
	ListerRegistry    kube_util.ListerRegistry
	PodListProcessor  pods.PodListProcessor
}

// Autoscaler is the main component of CA which scales up/down node groups according to its configuration
// The configuration can be injected at the creation of an autoscaler
type Autoscaler interface {
	// RunOnce represents an iteration in the control-loop of CA
	RunOnce(currentTime time.Time) errors.AutoscalerError
	// ExitCleanUp is a clean-up performed just before process termination.
	ExitCleanUp()
}

func initializeDefaultOptions(opts *AutoscalerOptions) error {
	if opts.PodListProcessor == nil {
		opts.PodListProcessor = pods.NewDefaultPodListProcessor()
	}
	return nil
}

// NewAutoscaler creates an autoscaler of an appropriate type according to the parameters
func NewAutoscaler(opts AutoscalerOptions) (Autoscaler, errors.AutoscalerError) {
	err := initializeDefaultOptions(&opts)
	if err != nil {
		return nil, errors.ToAutoscalerError(errors.InternalError, err)
	}
	autoscalerBuilder := NewAutoscalerBuilder(opts.AutoscalingOptions, opts.PredicateChecker, opts.KubeClient, opts.KubeEventRecorder, opts.ListerRegistry, opts.PodListProcessor)
	return autoscalerBuilder.Build()
}
