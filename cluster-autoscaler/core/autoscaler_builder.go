/*
Copyright 2017 The Kubernetes Authors.

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
	"k8s.io/autoscaler/cluster-autoscaler/context"
	ca_processors "k8s.io/autoscaler/cluster-autoscaler/processors"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	"k8s.io/autoscaler/cluster-autoscaler/utils/errors"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	kube_client "k8s.io/client-go/kubernetes"
	kube_record "k8s.io/client-go/tools/record"
)

// AutoscalerBuilder builds an instance of Autoscaler which is the core of CA
type AutoscalerBuilder interface {
	Build() (Autoscaler, errors.AutoscalerError)
}

// AutoscalerBuilderImpl builds new autoscalers from its state including initial `AutoscalingOptions` given at startup.
type AutoscalerBuilderImpl struct {
	autoscalingOptions context.AutoscalingOptions
	kubeClient         kube_client.Interface
	kubeEventRecorder  kube_record.EventRecorder
	predicateChecker   *simulator.PredicateChecker
	listerRegistry     kube_util.ListerRegistry
	processors         *ca_processors.AutoscalingProcessors
}

// NewAutoscalerBuilder builds an AutoscalerBuilder from required parameters
func NewAutoscalerBuilder(autoscalingOptions context.AutoscalingOptions, predicateChecker *simulator.PredicateChecker,
	kubeClient kube_client.Interface, kubeEventRecorder kube_record.EventRecorder, listerRegistry kube_util.ListerRegistry, processors *ca_processors.AutoscalingProcessors) *AutoscalerBuilderImpl {
	return &AutoscalerBuilderImpl{
		autoscalingOptions: autoscalingOptions,
		kubeClient:         kubeClient,
		kubeEventRecorder:  kubeEventRecorder,
		predicateChecker:   predicateChecker,
		listerRegistry:     listerRegistry,
		processors:         processors,
	}
}

// Build an autoscaler according to the builder's state
func (b *AutoscalerBuilderImpl) Build() (Autoscaler, errors.AutoscalerError) {
	options := b.autoscalingOptions
	return NewStaticAutoscaler(options, b.predicateChecker, b.kubeClient, b.kubeEventRecorder, b.listerRegistry, b.processors)
}
