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
	"k8s.io/autoscaler/cluster-autoscaler/config/dynamic"
	"k8s.io/autoscaler/cluster-autoscaler/simulator"
	kube_util "k8s.io/autoscaler/cluster-autoscaler/utils/kubernetes"
	kube_record "k8s.io/client-go/tools/record"
	kube_client "k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
)

// AutoscalerBuilder builds an instance of Autoscaler which is the core of CA
type AutoscalerBuilder interface {
	SetDynamicConfig(config dynamic.Config) AutoscalerBuilder
	Build() (Autoscaler, error)
}

// AutoscalerBuilderImpl builds new autoscalers from its state including initial `AutoscalingOptions` given at startup and
// `dynamic.Config` read on demand from the configmap
type AutoscalerBuilderImpl struct {
	autoscalingOptions AutoscalingOptions
	dynamicConfig      *dynamic.Config
	kubeClient         kube_client.Interface
	kubeEventRecorder  kube_record.EventRecorder
	predicateChecker   *simulator.PredicateChecker
	listerRegistry     kube_util.ListerRegistry
}

// NewAutoscalerBuilder builds an AutoscalerBuilder from required parameters
func NewAutoscalerBuilder(autoscalingOptions AutoscalingOptions, predicateChecker *simulator.PredicateChecker,
	kubeClient kube_client.Interface, kubeEventRecorder kube_record.EventRecorder, listerRegistry kube_util.ListerRegistry) *AutoscalerBuilderImpl {
	return &AutoscalerBuilderImpl{
		autoscalingOptions: autoscalingOptions,
		kubeClient:         kubeClient,
		kubeEventRecorder:  kubeEventRecorder,
		predicateChecker:   predicateChecker,
		listerRegistry:     listerRegistry,
	}
}

// SetDynamicConfig sets an instance of dynamic.Config read from a configmap so that
// the new autoscaler built afterwards reflect the latest configuration contained in the configmap
func (b *AutoscalerBuilderImpl) SetDynamicConfig(config dynamic.Config) AutoscalerBuilder {
	b.dynamicConfig = &config
	return b
}

// Build an autoscaler according to the builder's state
func (b *AutoscalerBuilderImpl) Build() (Autoscaler, error) {
	options := b.autoscalingOptions
	if b.dynamicConfig != nil {
		c := *(b.dynamicConfig)
		options.NodeGroups = c.NodeGroupSpecStrings()
	}
	return NewStaticAutoscaler(options, b.predicateChecker, b.kubeClient, b.kubeEventRecorder, b.listerRegistry)
}
